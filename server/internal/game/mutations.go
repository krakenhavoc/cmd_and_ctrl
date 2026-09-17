package game

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

// mutations.go holds the higher-level game-state mutation methods that
// the S03 action protocol dispatches against. Each method is a thin
// composition of the primitive zone/player operations in zone.go and
// player.go. None of these methods enforce real rules: they accept
// the mutation at face value, which is the Option B sandbox posture
// documented in PLAN.md §2.1. Rules enforcement grows in the S13+
// B→C graft track.

// ZoneRef identifies a zone on the game, for mutation APIs that move
// cards between arbitrary zones. Owner is uuid.Nil for shared zones
// (Battlefield, Stack, Exile) and the player ID for per-player zones
// (Library, Hand, Graveyard, Command).
type ZoneRef struct {
	Kind  ZoneKind
	Owner uuid.UUID
}

// zoneFromRefLocked resolves a ZoneRef to a concrete *Zone on the
// game. Returns nil if the ref is malformed or the target player is
// not seated. Must be called with g.mu held.
func (g *Game) zoneFromRefLocked(ref ZoneRef) *Zone {
	if ref.Owner == uuid.Nil {
		switch ref.Kind {
		case ZoneBattlefield:
			return g.Battlefield
		case ZoneStack:
			return g.Stack
		case ZoneExile:
			return g.Exile
		}
		return nil
	}
	p := g.playerByIDLocked(ref.Owner)
	if p == nil {
		return nil
	}
	switch ref.Kind {
	case ZoneLibrary:
		return p.Library
	case ZoneHand:
		return p.Hand
	case ZoneGraveyard:
		return p.Graveyard
	case ZoneCommand:
		return p.Command
	}
	return nil
}

// findCardZoneLocked returns the Zone currently holding the card with
// the given instance ID, or nil if the card is not in any zone. Must
// be called with g.mu held.
func (g *Game) findCardZoneLocked(cardID uuid.UUID) *Zone {
	if g.Battlefield.Contains(cardID) {
		return g.Battlefield
	}
	if g.Stack.Contains(cardID) {
		return g.Stack
	}
	if g.Exile.Contains(cardID) {
		return g.Exile
	}
	for _, p := range g.Seats {
		if p.Library.Contains(cardID) {
			return p.Library
		}
		if p.Hand.Contains(cardID) {
			return p.Hand
		}
		if p.Graveyard.Contains(cardID) {
			return p.Graveyard
		}
		if p.Command.Contains(cardID) {
			return p.Command
		}
	}
	return nil
}

// DrawCard moves the top card of the given player's library into
// their hand. Returns ErrZoneEmpty if the library is empty (the
// player would normally lose on the next state-based action check;
// S03 doesn't enforce that).
//
// S13: gated to no-op during the active player's StepDraw — that
// step's auto-action has already drawn for them, and a manual
// dispatch on top would draw twice. Outside StepDraw the manual
// action is honoured (sandbox / replay support).
func (g *Game) DrawCard(playerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.Turn.Step == StepDraw && g.activeSeatIDLocked() == playerID {
		return nil
	}
	if err := g.drawCardLocked(playerID); err != nil {
		return err
	}
	// Sandbox draw is a special action; the drawer keeps priority
	// and CR 117.5 puts SBAs + the trigger drain here (Smothering
	// Tithe, Consecrated Sphinx watch draws).
	g.runStateChecksLocked()
	return nil
}

// drawCardLocked is the unlocked draw used both by the public
// DrawCard action and by the StepDraw auto-action in
// runStepEntryHooksLocked. S13.1: an empty library marks the player
// for elimination at the next SBA check (CR 704.5b) and surfaces
// ErrZoneEmpty so the manual DrawCard path keeps its existing wire
// behaviour. The auto-fire path in the step entry hook swallows
// ErrZoneEmpty so the cursor still moves.
//
// Caller must hold g.mu.
func (g *Game) drawCardLocked(playerID uuid.UUID) error {
	// S17 sub-PR 2: route through the replacement pipeline so
	// draw-replacement effects ("if you would draw, mill instead",
	// "if you would draw, opponent draws instead", etc.) fire
	// pre-event. Sub-PR 2 registers zero catalog draw-replacements,
	// so applyReplacementsLocked short-circuits with no gathered
	// effects and behavior is byte-for-byte identical to pre-S17.
	ev := &ReplacementEvent{
		Kind:       RepEventDraw,
		Actor:      playerID,
		DrawPlayer: playerID,
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// CR 616 prompt queued; client will submit an order. The
		// resume path in ResolveReplacementOrder re-enters the
		// pipeline and runs the underlying draw. Return nil so the
		// caller (public DrawCard or step-draw auto-action) sees
		// the draw as "in flight" — no ErrZoneEmpty propagation.
		return nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		// Draw canceled by replacement.
		return nil
	}
	return g.actuallyDrawCardLocked(out.DrawPlayer)
}

// actuallyDrawCardLocked is the post-replacement draw body —
// pops the library, pushes to hand, marks known, emits the event.
// Extracted from drawCardLocked in S17 sub-PR 3 so the CR 616
// resume path (ResolveReplacementOrder → applyResolvedReplacementEventLocked)
// runs the same logic as the inline non-paused path. Caller must
// hold g.mu.
func (g *Game) actuallyDrawCardLocked(playerID uuid.UUID) error {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	c, err := p.Library.PopTop()
	if err != nil {
		if err == ErrZoneEmpty {
			p.AttemptedEmptyDraw = true
		}
		return err
	}
	p.Hand.PushTop(c)
	// S13.5: drawn cards become known to the owner. Cards drawn from
	// scry-positioned tops keep any pre-existing scry knowledge via
	// the sticky map; the AddKnower call is idempotent.
	g.markCardKnownInZoneLocked(p.Hand, c.InstanceID)
	// "Cards drawn this turn" bookkeeping (Sylvan Library). Recorded
	// here rather than in DrawNForEffect because this is the one place
	// a card actually crosses from library to hand as a draw — the
	// replacement pipeline above can redirect or cancel the draw, and
	// only the draws that happened should be listed.
	if g.DrawnThisTurn == nil {
		g.DrawnThisTurn = make(map[uuid.UUID][]uuid.UUID)
	}
	g.DrawnThisTurn[playerID] = append(g.DrawnThisTurn[playerID], c.InstanceID)
	g.EmitEvent(Event{
		Kind:    EventDrawCard,
		Actor:   playerID,
		CardID:  c.InstanceID,
		OldZone: ZoneLibrary,
		NewZone: ZoneHand,
	})
	return nil
}

// activeSeatIDLocked returns the player ID at the active seat, or
// uuid.Nil if no active player can be resolved (lobby state, empty
// seats, out-of-range index). Caller must hold g.mu.
func (g *Game) activeSeatIDLocked() uuid.UUID {
	if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
		return uuid.Nil
	}
	p := g.Seats[g.Turn.ActiveSeat]
	if p == nil {
		return uuid.Nil
	}
	return p.ID
}

// PlayCard moves a card from the player's hand to the shared
// battlefield. The card's controller is set to the player (already
// the case for cards entering from your own hand, but stored
// explicitly for clarity).
//
// S13.1: kept as the sandbox / admin direct-drop verb for token
// creation, replay restore, and "fix wedged state" cases. Normal
// play uses CastSpell, which routes through the stack.
func (g *Game) PlayCard(playerID, cardID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	// The card being played is in a hand, so its own types are
	// printed either way — but a land's enters-tapped replacement
	// asks about the BATTLEFIELD ("unless you control a Swamp"),
	// and under Urborg the answer is a layer answer. Fast-path
	// no-op when nothing changed.
	g.RecomputeLayersIfStaleLocked()
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	c, err := MoveCard(p.Hand, g.Battlefield, cardID)
	if err != nil {
		return err
	}
	// Stamp controller explicitly; owner is unchanged.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == c.InstanceID {
			g.Battlefield.Cards[i].Controller = playerID
			break
		}
	}
	// S13.5: arriving at a public zone makes the card known to all.
	g.markCardKnownInZoneLocked(g.Battlefield, c.InstanceID)
	return nil
}

// CastSpellParams carries the announce-time choices that flow into a
// new StackItem. All fields are optional; sensible zero values mean
// "no targets, no modes, no X, no distribution, default flags."
//
// FromZone defaults to "hand" when empty. The only other supported
// value today is "command" for casting a commander out of the
// command zone (S13.1 commander tax + zone-replacement work in
// sub-PR 8 reads this). Other zones (graveyard, library, exile,
// stack) come with S29's alt-cast-paths sprint.
//
// Targets / Modes / XValue / Distribution land in sub-PRs 3 & 4 —
// the params struct accepts them now so the action wire format stays
// stable across the sprint, but CastSpell itself ignores all but
// FromZone, HoldPriority, and SplitSecond at this checkpoint.
type CastSpellParams struct {
	FromZone     string
	Targets      []TargetRef
	Modes        []int
	XValue       int
	Distribution map[uuid.UUID]int
	HoldPriority bool
	SplitSecond  bool

	// DiscardIDs names the cards paid to an additional cost of the
	// form "As an additional cost to cast this spell, discard a
	// card" (CR 601.2f). Validated at announce against the card's
	// AdditionalCost and paid once the spell is on the stack, so
	// discard payoffs trigger above it. Empty for every card
	// without such a cost — and non-empty for one is rejected.
	// Added in S21 sub-PR 5.
	DiscardIDs []uuid.UUID

	// SacrificeIDs names the permanent paid to an additional cost of
	// the form "As an additional cost to cast this spell, sacrifice a
	// creature" (Village Rites, Deadly Dispute). Same discipline as
	// DiscardIDs: validated at announce, paid with the spell already
	// on the stack so the dies-triggers resolve above it, and
	// rejected rather than ignored on a card that charges no such
	// cost. Added in S21 sub-PR 6.
	SacrificeIDs []uuid.UUID

	// TapIDs names the untapped permanents the caster is tapping to
	// help pay for the spell — convoke's "your creatures can help
	// cast this spell", waterbend's "you can tap your artifacts and
	// creatures to help". Each one pays for {1}, or (convoke only)
	// for one mana of that permanent's colour.
	//
	// Same discipline as DiscardIDs and SacrificeIDs: validated at
	// announce, paid with the spell already on the stack, and
	// rejected rather than ignored on a card that offers no such
	// cost. Unlike those two it is OPTIONAL — "you MAY tap any
	// number", so an empty list is always a legal answer and the
	// caster simply pays the whole cost with mana. Added in S22.
	TapIDs []uuid.UUID

	// AlternativeCost names the cost the caster is paying INSTEAD of
	// the mana cost (CR 118.9) — the Key of one of the card's
	// declared game.AlternativeCost entries, "overload" / "evoke" /
	// "cleave". Empty is the ordinary case: pay the printed cost.
	//
	// Unlike DiscardIDs and SacrificeIDs this is a CHOICE, not a
	// demand — the caster may always decline and pay the printed
	// cost instead. What it is not is a discount on top of the
	// additional costs: those are charged either way. A key the
	// card doesn't offer is rejected rather than ignored, because
	// silently charging full price for a cast the player meant to
	// overload is the worst available failure. Added in S22.
	AlternativeCost string

	// AltCostIDs names the cards paid to the NON-MANA half of the
	// claimed alternative cost — Force of Will's "exile a blue card
	// from your hand", Daze's "return an Island you control to its
	// owner's hand", Solitude's evoke pitch. Exactly one entry when
	// the claimed cost has such a component, none otherwise, and a
	// non-empty list on a cost that charges no cards is rejected
	// rather than ignored.
	//
	// A separate slice from DiscardIDs and SacrificeIDs because it
	// pays a different cost: those are ADDITIONAL costs, charged
	// alongside the mana cost and charged whichever cost the caster
	// chose. This one is part of the alternative cost itself and
	// vanishes when the caster declines the offer. Folding them into
	// one list would make "I pitched a blue card" and "I discarded a
	// card" indistinguishable on the wire. Added in S28.
	AltCostIDs []uuid.UUID

	// Face names which printed face of a multi-face card is being
	// cast or played (ADR 0034). Zero — the front face — is the
	// answer for every single-faced card in the game and for every
	// client that has never heard of faces, which is what makes the
	// field safe to add.
	//
	// It is an announce-time PARAMETER, not a PendingChoice, for the
	// same reason the alternative cost, the modes, X, the additional
	// cost and the targets are: every other announce decision is a
	// client-side prompt whose answer rides the cast_spell action.
	// The PendingChoice machinery resumes replacement, search and
	// trigger frames; it has no frame for a half-validated cast and
	// should not grow one.
	//
	// A face the card does not offer is REJECTED (ErrInvalidFace)
	// rather than clamped to the front. Silently casting the wrong
	// half of a modal DFC is the worst available failure: the player
	// meant to play a land and got a seven-mana sorcery, or vice
	// versa.
	Face int

	// Strict enables the S15 mana-cost gate. When set, the server
	// parses the card's ManaCost into an effective cost (plus
	// commander tax for casts from the command zone), checks the
	// caster's ManaPool, rejects with ErrInsufficientMana when the
	// pool can't cover it, and deducts on success. Sourced from the
	// client's `gameplay.strictMana` setting — the action payload
	// carries it per-cast rather than round-tripping through a
	// server-side preference. Default false (permissive).
	Strict bool

	// ForceCast overrides the Strict gate for this one cast. Fired
	// by the client's "Cast anyway" toast after an insufficient-mana
	// error. Implies Strict (the user is explicitly overriding a
	// strict gate that just rejected them), but the server treats
	// it as a permissive proceed: no mana is deducted and an
	// EventCostWarning is emitted. Default false.
	ForceCast bool

	// AutoTap asks the server to plan and execute a mana-source
	// tap-and-fill before the strict-mode cost check. Implies
	// Strict — the auto-tapper exists to make a strict-gated cast
	// succeed without manually clicking each land. The server
	// runs `AutoTapForCostExcluding(controller, cost, x, locked)`
	// against the caller's untapped permanents, taps each card in
	// the returned plan, drops the produced mana into the pool
	// (with greedy color-picking against the cost requirements),
	// then proceeds to the normal CanPay/SpendMana flow — all
	// atomically under the same write lock as the cast. Failure
	// to satisfy the cost (no plan exists, or budget exceeded)
	// returns an *InsufficientManaError before any card taps,
	// keeping the operation all-or-nothing. Added in S15 sub-PR 5.
	AutoTap bool

	// LockedSources are permanent IDs the auto-tapper must NOT
	// consider when planning — the player has reserved them for a
	// later cast via the lock-tap UI. Only consulted when AutoTap
	// is true. Cards already tapped are skipped naturally by the
	// gather pass; LockedSources is for *untapped* permanents the
	// player wants kept available. Added in S15 sub-PR 5.
	LockedSources []uuid.UUID
}

// InsufficientManaError is returned by CastSpell when the Strict
// gate is engaged and the caller's ManaPool can't cover the
// effective cost. Carries the list of missing symbols so the
// client's "Cast anyway" toast can render exactly what's short
// ("{R}{R}" vs "{1}"). Wraps the sentinel error so callers doing
// errors.Is(err, ErrInsufficientMana) still match.
type InsufficientManaError struct {
	Missing []string
}

func (e *InsufficientManaError) Error() string {
	return "game: insufficient mana"
}

// Unwrap lets errors.Is(err, ErrInsufficientMana) still match a
// *InsufficientManaError in the dispatcher / protocol layer.
func (e *InsufficientManaError) Unwrap() error { return ErrInsufficientMana }

// CastSpell is the canonical "play a card from hand" verb (CR 601).
// Lands route directly to the battlefield — they're a special action
// that doesn't use the stack (CR 305). Every other card type goes to
// the stack with a fresh StackMeta entry capturing announce-time
// choices, and the caster RETAINS priority (CR 117.3c) — they pass
// explicitly via PassPriority once they're done.
//
// Sorcery-speed gate: sorceries (and sub-PR 8's commander casts and
// sub-PR 6's loyalty abilities) require main-phase + stack-empty +
// caller-is-active-player, per CR 307.1. Instants honour the
// caller-holds-priority gate at the action layer (requirePriorityHolder
// in the dispatcher), so no extra speed gate is needed here for
// them.
//
// Split-second blocks all casts and activations except mana abilities
// and special actions (CR 702.61). The flag is mirrored on
// Game.SplitSecondActive for fast lookup; recomputed every time the
// stack changes.
//
// On success: the card is in the stack zone (for non-lands) with a
// StackMeta entry, or on the battlefield (for lands). PriorityHolder
// is unchanged — the caster retains priority. SBA loop and pending-
// trigger drain are sub-PR 7 / 6 territory.
func (g *Game) CastSpell(playerID, cardID uuid.UUID, params CastSpellParams) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	// CR 601.2c's announce-time target check runs against the
	// battlefield through predicates that read effective types, so
	// the engine has to be caught up before a Doom Blade is told
	// whether that land is a creature. Fast-path no-op when nothing
	// changed.
	g.RecomputeLayersIfStaleLocked()
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if g.SplitSecondActive {
		return ErrSplitSecondActive
	}
	src, err := g.castSourceZoneLocked(p, params.FromZone)
	if err != nil {
		return err
	}
	// Find the card in the source zone so we can inspect its type
	// before moving anything. Pre-S13.1 PlayCard moved first then
	// stamped — for cast we need the type *before* deciding the
	// destination, so look up first.
	var card Card
	found := false
	for _, c := range src.Cards {
		if c.InstanceID == cardID {
			card = c
			found = true
			break
		}
	}
	if !found {
		return ErrCardNotFound
	}
	// ADR 0034, and the single highest-leverage line in the whole
	// multi-face model. `card` is a VALUE COPY taken out of the
	// source zone, and everything below reads that copy ten more
	// times before anything moves: IsLand(), validateAlternativeCost,
	// TargetModeFor / TargetSpecFor, ModeSpecFor, castTargetSpec,
	// AdditionalCostFor, TapPermanentsCostFor, the sorcery-speed
	// gate, the land branch, and printedCostLocked. Materialising
	// the chosen face HERE makes every one of them face-correct for
	// free.
	//
	// Concretely, for Sea Gate Restoration // Sea Gate, Reborn:
	// face 0 stops passing IsLand() (its type line is "Sorcery", not
	// "Sorcery // Land"), so the land branch no longer fires and the
	// cost gate finally sees {4}{U}{U}{U} instead of "" — which is
	// both halves of #289 and all of #265.
	//
	// S32: the per-instance exile grant is read BEFORE the face gate,
	// because a grant can name a face the card's own layout does not
	// offer — "exile it, then cast it transformed" is a cast of a
	// `transform` card's back face, which CastableFaces refuses on
	// principle. ExilePlay is a property of the card instance and not
	// of any face, so reading it before SetFace is safe and is the
	// only order in which the gate can consult it. See
	// faceForCastLocked for why a face-naming grant SETS the face
	// rather than merely permitting it.
	exileGrant := ExilePlayPermission{}
	if src.Kind == ZoneExile {
		exileGrant = card.ExilePlay
	}
	face, ok := faceForCastLocked(card, params.Face, exileGrant, playerID, g.Turn.Number)
	if !ok {
		slog.Warn("cast_spell rejected: face not offered by this card",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"layout", card.Layout,
			"face_requested", params.Face,
			"faces", card.FaceCount(),
		)
		return ErrInvalidFace
	}
	// params.Face is the settled face from here down — the stack push
	// and the land branch both re-read it to stamp the card in its
	// source zone, and they must agree with the copy the announce
	// gates were judged against.
	params.Face = face
	card.SetFace(params.Face)
	// S21 sub-PR 6: casting out of exile needs a live impulse-exile
	// grant naming this player. Checked before every other gate
	// because it's the one that decides whether the card is yours to
	// touch at all.
	//
	// S29 softened the "no grant, no cast" half by exactly one case:
	// a card whose own text declares ZoneExile castable does not need
	// an instance grant. That declaration is a CARD-level permission,
	// so it opens exile for every copy of the card at any time, and
	// no catalog card declares it today. Suspend and foretell are NOT
	// this shape: CR 702.62a allows a suspended card's cast only while
	// its last-time-counter trigger resolves, and CR 702.143 makes
	// foretold status belong to the exiled instance. Both want a
	// per-instance ExilePlayPermission, the way cascade's free cast
	// works. hasExileGrant rides down to validateCastPathLocked so the
	// zone-cost rule can tell the two apart.
	hasExileGrant := false
	if src.Kind == ZoneExile {
		hasExileGrant = exileGrant.Active(playerID, g.Turn.Number)
		if !hasExileGrant && !CardCastableFromZone(CatalogKey(card), ZoneExile) {
			return ErrNoPlayPermission
		}
		// "You may CAST that card" (Ragavan) does not let you play a
		// land: playing a land is a special action, not a cast
		// (CR 305.1, 116.2a).
		if hasExileGrant && exileGrant.CastOnly && card.IsLand() {
			return ErrNoPlayPermission
		}
	}
	// S22: an alternative cost is claimed at announce and replaces
	// the mana cost (CR 118.9). Resolved before every targeting gate
	// below, because overload and cleave rewrite the target clause —
	// the spell's legality has to be judged under the cost actually
	// being paid, not under the printed one.
	alt, err := validateAlternativeCost(CatalogKey(card), params.AlternativeCost, params.Targets)
	if err != nil {
		slog.Warn("cast_spell rejected: bad alternative cost claim",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"alternative_cost", params.AlternativeCost,
			"targets_received", len(params.Targets),
		)
		return err
	}
	// S29: the cast PATH — zone and price together. Runs here, right
	// after the claim is known to be an offer the card makes and
	// before any targeting work, because the rewrite an alternative
	// cost applies to the target clause is only legitimate if the
	// cost could be claimed from this zone at all.
	if err := g.validateCastPathLocked(card, src.Kind, alt, hasExileGrant); err != nil {
		slog.Warn("cast_spell rejected: illegal cast path",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"from_zone", src.Kind,
			"alternative_cost", params.AlternativeCost,
			"err", err,
		)
		return err
	}
	// CR 118.6: no mana cost is an unpayable cost, and paying it is
	// illegal, so a cast that would pay it is refused here. Checked
	// once the claimed alternative cost and the exile grant are both
	// known, because an alternative cost applied to the unpayable
	// cost may be paid (CR 118.6a).
	// Mode-independent, like the unparseable-cost refusal: there is
	// no cost to have paid on paper either.
	if HasNoManaCost(card) && castPaysPrintedCost(alt, exileGrant, hasExileGrant) {
		slog.Warn("cast_spell rejected: no mana cost to pay",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"from_zone", src.Kind,
		)
		return ErrNoManaCost
	}
	// S28: the non-mana half of the claimed offer — the Condition
	// ("if you control a Swamp"), the life payment, and the card
	// pitched or bounced to pay it. Validated here, next to the claim
	// it belongs to; paid further down with the spell already on the
	// stack, so a Blood Artist watching the pitch triggers above it.
	if err := g.validateAlternativeCostPaymentLocked(playerID, cardID, alt, params.AltCostIDs); err != nil {
		slog.Warn("cast_spell rejected: bad alternative cost payment",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"alternative_cost", params.AlternativeCost,
			"payment_received", len(params.AltCostIDs),
			"err", err,
		)
		return err
	}
	// S17 sub-PR 6 follow-up: if the catalog declares a target_mode
	// for this card, a cast without any target is a client bug (the
	// targeting UI should have opened before firing cast_spell).
	// Rejecting here turns the silent "spell resolves with no effect"
	// failure into a visible ErrInvalidParam that the client's error
	// toast surfaces. Non-catalog cards (empty TargetMode) pass
	// through unchanged. S20: cards with a structured TargetSpec are
	// counted by validateTargetsLocked below instead — an "up to N"
	// clause legitimately arrives with none. S22: an overloaded spell
	// has no target clause left to satisfy.
	if mode := TargetModeFor(CatalogKey(card)); mode != "" && len(params.Targets) == 0 &&
		TargetSpecFor(CatalogKey(card)) == nil && !alt.Clears() {
		slog.Warn("cast_spell rejected: targeted card arrived without targets",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"required_mode", mode,
			"targets_received", len(params.Targets),
		)
		return ErrInvalidParam
	}
	// S20 sub-PR 3: X is a non-negative announce-time choice (CR
	// 601.2b). The cost gate multiplies it into the generic demand;
	// a negative would let a caster be refunded mana.
	if params.XValue < 0 {
		return ErrInvalidParam
	}
	// S20 sub-PR 4: modal spells — the chosen modes must be distinct,
	// in range and the right count (CR 601.2b, 700.2).
	modeSpec := ModeSpecFor(CatalogKey(card))
	if err := validateModes(modeSpec, params.Modes); err != nil {
		slog.Warn("cast_spell rejected: bad mode choice",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"modes_received", params.Modes,
		)
		return err
	}
	// S20: structured targeting. Cards with a TargetSpec — declared
	// on the card, or on the chosen mode of a modal card — get their
	// announce-time targets validated against it (CR 601.2c): zone,
	// count, and predicate. Cards without one keep the S13.1
	// free-form behaviour (any ID the client sent is accepted),
	// except that a modal card whose chosen modes take no target
	// must arrive with none.
	spec, err := castTargetSpec(CatalogKey(card), params.Modes)
	if err != nil {
		return err
	}
	// S22: the alternative cost gets the last word on the clause —
	// overload deletes it, cleave swaps a wider one in. Applied after
	// the modal derivation so a modal card with an alternative cost
	// would compose rather than conflict.
	spec = TargetSpecUnderAlternativeCost(spec, alt)
	// S22: "Exile X target creatures you control" — the clause's
	// count is the X announced at 601.2b, so resolve it into a
	// concrete Min / Max before anything validates against it. The
	// copy (rather than a mutation) matters: the catalog's TargetSpec
	// is shared by every cast of the card.
	if spec != nil && spec.CountFromX {
		// Max 0 means "unbounded" everywhere else in TargetSpec, so
		// an X of zero has to be rejected here rather than left to
		// the count check below — otherwise announcing X=0 would buy
		// an unbounded clause for free, which is the exact shape of
		// the bug this field exists to close.
		if n := countRealTargets(params.Targets); n != params.XValue {
			slog.Warn("cast_spell rejected: X-defined target count mismatch",
				"card_name", card.Name,
				"oracle_id", card.OracleID,
				"x_value", params.XValue,
				"targets_received", n,
			)
			return ErrInvalidParam
		}
		resolved := *spec
		resolved.Min, resolved.Max = params.XValue, params.XValue
		resolved.CountFromX = false
		spec = &resolved
	}
	if spec == nil && modeSpec != nil && len(params.Targets) > 0 {
		return ErrInvalidParam
	}
	if spec != nil {
		if err := g.validateTargetsLocked(playerID, spec, params.Targets); err != nil {
			slog.Warn("cast_spell rejected: illegal target",
				"card_name", card.Name,
				"oracle_id", card.OracleID,
				"targets_received", len(params.Targets),
				"err", err,
			)
			return err
		}
	}
	// S21 sub-PR 5: additional costs (CR 601.2f). Validated here,
	// with the rest of the announce-time choices, and paid further
	// down once the spell is on the stack — validate-all-then-pay,
	// so a rejected cast never leaves a card in the graveyard.
	addCost := AdditionalCostFor(CatalogKey(card))
	if err := g.validateAdditionalCostLocked(playerID, cardID, addCost, params.DiscardIDs, params.SacrificeIDs, params.XValue); err != nil {
		slog.Warn("cast_spell rejected: bad additional cost payment",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"discards_received", len(params.DiscardIDs),
			"sacrifices_received", len(params.SacrificeIDs),
			"err", err,
		)
		return err
	}
	// S22: tap-permanents-as-a-cost — convoke and waterbend. Checked
	// here with the other announce-time choices and paid further
	// down once the spell is on the stack, same validate-all-then-pay
	// discipline the additional cost uses. The budget is measured
	// against the cost the cast owes BEFORE any tapping, so a caster
	// can't tap five creatures at a three-mana spell.
	tapCost := TapPermanentsCostFor(CatalogKey(card))
	if !tapCost.Empty() || len(params.TapIDs) > 0 {
		budget := 0
		if base, berr := g.printedCostLocked(p, card, params); berr == nil {
			budget = tapPermanentsBudget(tapCost, base, params.XValue)
		}
		if err := g.validateTapPermanentsCostLocked(playerID, tapCost, params.TapIDs, budget); err != nil {
			slog.Warn("cast_spell rejected: bad tap-permanents cost payment",
				"card_name", card.Name,
				"oracle_id", card.OracleID,
				"taps_received", len(params.TapIDs),
				"budget", budget,
				"err", err,
			)
			return err
		}
	}
	// Sorcery-speed gate. Lands are special-action-fast (CR 305 is
	// "you may play a land during your main phase if the stack is
	// empty"); they're handled implicitly by the same gate below.
	// Instants bypass the gate entirely. Flash (CR 702.8) lets any
	// card be cast as though it had the timing of an instant — so
	// off-battlefield HasKeyword (backed by CatalogPrintedKeywords
	// from S18 sub-PR 2) lifts the sorcery-speed restriction for
	// hand-resident Ambush Viper and the like. Anything else
	// (sorceries, permanents that aren't instants and don't have
	// flash) needs sorcery speed.
	requiresSorcerySpeed := !card.IsInstant() && !card.IsLand() && !HasKeyword(&card, "flash")
	if card.IsLand() || requiresSorcerySpeed {
		if !g.sorcerySpeedOpenLocked(playerID) {
			return ErrSorcerySpeedRequired
		}
	}
	// Lands skip the stack entirely (CR 305). Move the card to the
	// battlefield and stamp the controller — same shape as PlayCard.
	if card.IsLand() {
		// ADR 0034: the chosen face has to exist on the card IN THE
		// SOURCE ZONE, not just on the local copy, before the
		// replacement pipeline runs.
		//
		// The pipeline resolves the entering card by ID through
		// LookupCardForEffect to find its own self-replacement
		// (replacements.go:531) — that is how a land finds its
		// "as this enters, you may pay N life" clause, and an MDFC
		// land back finds it under CatalogKey "<oracle>#1" only if
		// its ActiveFace is already 1. The entry-choice resume path
		// re-enters later with nothing but the card ID, so the face
		// has to survive the pause too.
		//
		// Restored on the failure and cancel paths below: a cast
		// that does not happen must not leave a card in hand wearing
		// its back face.
		wasFace := setFaceInZoneLocked(src, cardID, params.Face)
		// S17 sub-PR 4: run the CR 614 replacement pipeline so
		// enters-tapped replacements (Kismet) + enters-with-counters
		// effects fire before the land's ETB event. Pipeline runs
		// pre-push so a future destination-rewriter replacement
		// (e.g. a hypothetical "lands go to graveyard instead")
		// would redirect cleanly.
		ev := &ReplacementEvent{
			Kind:    RepEventMove,
			CardID:  cardID,
			OldZone: src.Kind,
			NewZone: ZoneBattlefield,
			Actor:   playerID,
			// A land's entry can now pause on a prompt (the
			// shockland's "pay 2 life"), and this branch returns to
			// the client when it does. Flag the event so the resume
			// path knows it may finish the push on this branch's
			// behalf — see executeEntryToBattlefieldLocked.
			entryResumable: true,
		}
		out, err := g.applyReplacementsLocked(ev)
		if errors.Is(err, errReplacementPending) {
			return nil
		}
		if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
			g.clearReplacementEventLocked(ev.ID)
			setFaceInZoneLocked(src, cardID, wasFace)
			return err
		}
		defer g.clearReplacementEventLocked(ev.ID)
		if out == nil || out.Canceled {
			setFaceInZoneLocked(src, cardID, wasFace)
			return nil
		}

		moved, err := MoveCard(src, g.Battlefield, cardID)
		if err != nil {
			setFaceInZoneLocked(src, cardID, wasFace)
			return err
		}
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == moved.InstanceID {
				g.Battlefield.Cards[i].Controller = playerID
				if out.EntersTapped {
					g.Battlefield.Cards[i].Tapped = true
				}
				// The impulse grant is spent. Zeroing it here means a
				// later effect that exiles this card again can't
				// inherit a permission it never granted.
				g.Battlefield.Cards[i].ExilePlay = ExilePlayPermission{}
			}
		}
		g.markCardKnownInZoneLocked(g.Battlefield, moved.InstanceID)
		for name, n := range out.EntersWithCounters {
			_ = g.AddCounterForEffect(moved.InstanceID, name, n)
		}
		// S31 sub-PR 1: per-turn land-drop tally for the legal-move
		// enumerator. Bookkeeping only — the engine still doesn't
		// refuse a second land (sandbox posture).
		if g.LandsPlayedThisTurn == nil {
			g.LandsPlayedThisTurn = make(map[uuid.UUID]int)
		}
		g.LandsPlayedThisTurn[playerID]++
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			Actor:   playerID,
			CardID:  moved.InstanceID,
			OldZone: src.Kind,
			NewZone: ZoneBattlefield,
		})
		g.EmitEvent(Event{
			Kind:   EventETB,
			Actor:  playerID,
			CardID: moved.InstanceID,
		})
		g.fireETBHookLocked(moved.InstanceID, CatalogKey(moved))
		// Playing a land is a special action (CR 116.2a); the player
		// keeps priority and CR 117.5 drains any landfall-style
		// triggers onto the stack here rather than at the next wrap.
		g.runStateChecksLocked()
		return nil
	}
	// S15 sub-PR 5 auto-tap. When AutoTap is set, plan a tap of the
	// caller's untapped permanents and materialise the produced mana
	// into the pool BEFORE the strict-mode cost gate runs. Failure
	// to satisfy the cost surfaces as *InsufficientManaError without
	// having tapped anything (the planner is read-only; only a
	// successful plan triggers materialisation). The strict gate
	// below then sees a freshly-funded pool and either spends it or
	// — if AutoTap somehow returned a short plan — rejects with the
	// same structured error.
	if params.AutoTap {
		if err := g.applyAutoTapLocked(p, card, params); err != nil {
			return err
		}
	}
	// S15 strict-mode cost gate. Only engaged for non-land casts —
	// lands have no mana cost, and the land-cast branch above has
	// already returned. Strict + payable deducts the cost from the
	// pool; strict + not payable + not forced rejects with a
	// structured InsufficientManaError; permissive or forced emits
	// an EventCostWarning and proceeds without touching the pool
	// (sandbox posture — paper tracking remains valid). A cost the
	// parser can't read rejects in every mode (#289).
	if err := g.applyCastCostLocked(p, card, params, cardID); err != nil {
		return err
	}
	// Non-land: route through the stack. The card lives in
	// Game.Stack; the announce-time choices live in StackMeta.
	if _, err := MoveCard(src, g.Stack, cardID); err != nil {
		return err
	}
	// The impulse grant is spent — see the land branch above.
	for i := range g.Stack.Cards {
		if g.Stack.Cards[i].InstanceID == cardID {
			g.Stack.Cards[i].ExilePlay = ExilePlayPermission{}
			// ADR 0034: the spell on the stack IS the chosen face.
			// Everything downstream of here reads the card out of
			// the stack zone rather than from the local copy — the
			// resolution target re-check, the effect resolver, the
			// graveyard route, and the client's stack overlay — so
			// the face has to be stamped on the real card, not just
			// the copy the announce gates were judged against.
			g.Stack.Cards[i].SetFace(params.Face)
		}
	}
	// S13.5: cast spells are public on the stack.
	g.markCardKnownInZoneLocked(g.Stack, cardID)
	if g.StackMeta == nil {
		g.StackMeta = make(map[uuid.UUID]*StackItem)
	}
	g.StackMeta[cardID] = &StackItem{
		ID:           cardID,
		Kind:         StackItemSpell,
		Controller:   playerID,
		Owner:        card.Owner,
		SourceCardID: cardID,
		Targets:      append([]TargetRef(nil), params.Targets...),
		Modes:        append([]int(nil), params.Modes...),
		XValue:       params.XValue,
		Distribution: cloneDistributionLocked(params.Distribution),
		HoldPriority: params.HoldPriority,
		SplitSecond:  params.SplitSecond,
		AltCost:      params.AlternativeCost,
		CastFromZone: src.Kind,
		Seq:          g.nextStackSeqLocked(),
		// S20: remember the clause the targets were validated under so
		// the resolution re-check and per-slot effect checks use it.
		targetSpec: spec,
	}
	// CR 601.2h: pay the costs. The mana component was charged
	// above (pre-move, as S15 wrote it); the additional cost is
	// paid HERE, with the spell already on the stack, because a
	// discard payoff — or an aristocrats payoff watching the
	// sacrifice — that triggers off it must resolve before the
	// spell does. Validation happened at announce, so a failure
	// past this point is an engine bug rather than a bad request.
	payLife := 0
	if addCost != nil && addCost.PayLifeX {
		payLife = params.XValue
	}
	if err := g.payAdditionalCostLocked(playerID, params.DiscardIDs, params.SacrificeIDs, payLife); err != nil {
		slog.Error("cast_spell: additional cost failed after validation",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"err", err,
		)
		return err
	}
	// S28: the alternative cost's own non-mana components, in the
	// same window and for the same reason — pitching a Force of Will
	// is a card leaving hand while the counterspell is on the stack,
	// and a Daze returns its Island before the spell it is answering
	// has resolved.
	if err := g.payAlternativeCostLocked(playerID, alt, params.AltCostIDs); err != nil {
		slog.Error("cast_spell: alternative cost failed after validation",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"err", err,
		)
		return err
	}
	// S22: the convoke / waterbend taps are a cost too, and land in
	// the same window and for the same reason — with the spell
	// already on the stack, so anything watching the taps triggers
	// above it. The mana side of the payment was already folded into
	// the cost gate above; this is the board half.
	g.payTapPermanentsCostLocked(playerID, params.TapIDs)
	if params.SplitSecond {
		g.SplitSecondActive = true
	}
	// Commander tax bookkeeping (CR 903.8). Increment AFTER the
	// card has reached the stack so a failed cast doesn't bump the
	// counter. Sandbox: the engine doesn't enforce the +{2}
	// surcharge — players track mana in their head; the counter is
	// the affordance.
	if params.FromZone == "command" {
		if p.CommanderCasts == nil {
			p.CommanderCasts = make(map[uuid.UUID]int)
		}
		p.CommanderCasts[cardID]++
	}
	// S19 sub-PR 6: per-turn cast tally, bumped BEFORE EventCast so
	// "first noncreature spell each turn" predicates see this spell
	// counted (Noncreature == 1 means "this is the first").
	if g.SpellsCastThisTurn == nil {
		g.SpellsCastThisTurn = make(map[uuid.UUID]CastTally)
	}
	tally := g.SpellsCastThisTurn[playerID]
	tally.Total++
	if !card.IsCreature() {
		tally.Noncreature++
	}
	g.SpellsCastThisTurn[playerID] = tally
	// S22 airbend: OldZone stamps where the spell was cast FROM.
	// CR 601.2a moves the card to the stack and nothing on the card
	// remembers the zone it left, so "whenever you cast a spell from
	// exile" (Appa) has no other way to ask. The StackItem half of
	// this fact landed with #257 (CastFromZone); this is the event
	// half, and it reuses the ZoneMove-shaped fields rather than
	// growing a new one, because a cast IS a zone move — hand (or
	// exile, or the command zone) to the stack.
	g.EmitEvent(Event{
		Kind:    EventCast,
		Actor:   playerID,
		Source:  cardID,
		CardID:  cardID,
		OldZone: src.Kind,
		NewZone: ZoneStack,
	})
	// CR 115.7: the objects named at 601.2c have now become targets.
	// Emitted after EventCast so a "becomes the target" trigger and
	// a "whenever a player casts a spell" trigger queue in printed
	// order.
	g.emitBecameTargetLocked(playerID, cardID, cardID, params.Targets)
	// The caster receives priority right after casting (CR 117.3c),
	// and CR 603.3 puts any cast-triggered abilities (Rhystic Study,
	// Beast Whisperer) on the stack at that moment — above the
	// spell, so they resolve first.
	g.runStateChecksLocked()
	return nil
}

// applyCastCostLocked enforces the S15 strict-mode mana-cost gate.
// Caller must hold g.mu and have already validated the card lives
// in the source zone. The function decides one of three outcomes:
//
//  1. **Strict, payable** — deduct the effective cost from the
//     caster's ManaPool and emit no warning. The cast proceeds
//     normally with the deducted pool.
//  2. **Strict, not payable, not ForceCast** — return an
//     *InsufficientManaError carrying the missing-symbols slice.
//     Pool is unchanged; the cast is rejected.
//  3. **Permissive OR ForceCast** — emit EventCostWarning and
//     leave the pool alone. Lets the player track mana on paper
//     and lets the strict-mode override toast bypass the gate
//     for one cast without consuming mana the caller may not
//     have actually paid.
//
// The "effective cost" parses the printed ManaCost and adds
// {2}-per-prior-cast for casts from the command zone (CR 903.8).
// ParseCost treats an EMPTY ManaCost as the zero cost, matching land
// behaviour. A non-land spell never gets here on an empty cost it is
// paying: CastSpell refuses that earlier with ErrNoManaCost (CR
// 118.6). An UNPARSEABLE cost rejects the cast outright, before any
// of the three outcomes above. Caller must hold g.mu.
func (g *Game) applyCastCostLocked(p *Player, card Card, params CastSpellParams, cardID uuid.UUID) error {
	cost, err := g.effectiveCostLocked(p, card, params)
	if err != nil {
		// #289: this used to return nil, silently making the card
		// FREE. Split and adventure cards import a joined cost
		// ("{1}{R} // {1}{U}") that ParseCost rightly rejects, so
		// ~1,047 cards cast for nothing with no error on the wire.
		//
		// Refusing is the honest answer, and it is deliberately
		// mode-independent. Permissive mode's bargain is "the
		// engine knows the cost, you pay it on paper" — void when
		// the engine cannot read the cost at all. ForceCast
		// overrides the strict-mana GATE, not the parser: there is
		// no cost for the player to have paid. The card stays in
		// hand and the player sees why.
		g.EmitEvent(Event{
			Kind:     EventCostWarning,
			Actor:    p.ID,
			Source:   cardID,
			ErrorMsg: err.Error(),
		})
		return fmt.Errorf("%w for %s: %w", ErrUnparseableCost, card.Name, err)
	}
	if !params.Strict || params.ForceCast {
		// Permissive default OR strict-mode override. Don't touch
		// the pool; just emit a warning so the client can render
		// the toast / event-log breadcrumb.
		g.EmitEvent(Event{
			Kind:   EventCostWarning,
			Actor:  p.ID,
			Source: cardID,
		})
		return nil
	}
	// S32 (#352): the spend context is what lets restricted mana pay
	// — and what stops it paying for the wrong thing. Ancient
	// Ziggurat's {G} funds a creature spell here and is invisible to
	// a Lightning Bolt.
	spendCtx := ManaSpendForCast(card)
	if !p.ManaPool.CanPayFor(cost, params.XValue, spendCtx) {
		return &InsufficientManaError{Missing: p.ManaPool.MissingFor(cost, params.XValue, spendCtx)}
	}
	// Payable under strict mode — commit the spend.
	p.ManaPool.SpendManaFor(cost, params.XValue, spendCtx)
	g.EmitEvent(Event{
		Kind:   EventManaSpent,
		Actor:  p.ID,
		Source: cardID,
	})
	return nil
}

// applyAutoTapLocked plans + executes the S15 sub-PR 5 auto-tap
// step. Called from CastSpell when params.AutoTap is set, before
// the strict-mode cost gate. Atomic: a planning failure returns
// *InsufficientManaError before any cards tap.
//
// Flow:
//  1. Parse the effective cost (printed cost + commander tax).
//     Unparseable cost short-circuits to nil WITHOUT tapping
//     anything — applyCastCostLocked runs next and rejects the
//     cast with ErrUnparseableCost, so tapping here would strand
//     the caster's lands for a cast that never happens.
//  2. If the pool already covers the cost, skip — auto-tap is
//     idempotent on a funded pool.
//  3. Build the excluded set from LockedSources.
//  4. Run autoTapLocked to get a plan; if no plan exists, return
//     a structured InsufficientManaError keyed off the current
//     pool's missing list (the auto-tapper itself doesn't carry
//     a missing-symbols breakdown).
//  5. Materialise the plan: for each card, tap it and drop its
//     produced mana into the pool with greedy color-picking
//     against the still-unsatisfied cost requirements.
//
// Caller must hold g.mu (CastSpell holds the write lock).
func (g *Game) applyAutoTapLocked(p *Player, card Card, params CastSpellParams) error {
	cost, err := g.effectiveCostLocked(p, card, params)
	if err != nil {
		return nil
	}
	// Same context applyCastCostLocked will pay under, so the
	// "already funded, skip planning" shortcut can't be fooled by
	// restricted mana this cast cannot legally spend.
	if p.ManaPool.CanPayFor(cost, params.XValue, ManaSpendForCast(card)) {
		return nil
	}
	excluded := make(map[uuid.UUID]bool, len(params.LockedSources)+len(params.TapIDs))
	for _, id := range params.LockedSources {
		excluded[id] = true
	}
	// S22: a permanent tapped for convoke or waterbend is already
	// spent. Without this the auto-tapper would happily plan a mana
	// tap of the same Birds of Paradise the caster just convoked,
	// and the two payments would race for one untapped creature.
	for _, id := range params.TapIDs {
		excluded[id] = true
	}
	plan, ok := g.autoTapLocked(p.ID, cost, params.XValue, excluded)
	if !ok {
		return &InsufficientManaError{Missing: p.ManaPool.MissingFor(cost, params.XValue, ManaSpendForCast(card))}
	}
	g.materializePlanLocked(p, plan, cost)
	return nil
}

// materializePlanLocked taps each card in `plan` and drops the
// produced mana into the controller's pool. Multi-option slots
// (Birds of Paradise, Arcane Signet) get greedy color-picking:
// the picker walks the still-unsatisfied colored requirements and
// picks an option that consumes one. When no requirement matches,
// the slot drops its first option as generic-eligible mana.
//
// Skips PendingChoiceMana entirely — the auto-tapper's contract
// is "no further player decisions required". Caller must hold
// g.mu and have validated the plan via autoTapLocked.
func (g *Game) materializePlanLocked(p *Player, plan []uuid.UUID, cost ParsedCost) {
	pending := append([]ColorRequirement(nil), cost.Required...)
	identity := commanderIdentityFor(p)
	for _, cardID := range plan {
		var card *Card
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cardID {
				card = &g.Battlefield.Cards[i]
				break
			}
		}
		if card == nil || card.Tapped {
			continue
		}
		ab := autoTapAbilityFor(ManaAbilitiesForCard(*card))
		if ab == nil {
			continue
		}
		// S32 (#352): the gate and the derived/scaled output are
		// re-evaluated here rather than carried over from planning,
		// so the executor and the planner can never disagree about
		// what a source produces. A gate that has stopped holding
		// since the plan was made drops the source silently, exactly
		// as a source that got tapped in between does.
		if ab.Condition != nil && !ab.Condition(g, p.ID, cardID) {
			continue
		}
		producedStr := ab.Produced
		if ab.ProducedFunc != nil {
			producedStr = ab.ProducedFunc(g, p.ID, cardID)
		}
		slots, err := ParseProducedMana(producedStr)
		if err != nil {
			continue
		}
		card.Tapped = true
		g.EmitEvent(Event{Kind: EventTapCard, Actor: p.ID, CardID: cardID})
		g.EmitEvent(Event{Kind: EventManaAbilityActivated, Actor: p.ID, Source: cardID})
		for _, slot := range slots {
			options := slot.Options
			if len(options) > 1 && len(identity) > 0 && !ab.IgnoreCommanderIdentity {
				// Mirror ActivateManaAbility's narrowing, including
				// its no-overlap fallback: an intersection that came
				// back empty means the identity says nothing useful
				// about this slot, and dropping the mana on the floor
				// would silently short the plan the solver just
				// validated against the raw option set.
				if narrowed := intersectColors(options, identity); len(narrowed) > 0 {
					options = narrowed
				}
			}
			color := pickColorForSlot(options, &pending)
			if color == "" {
				continue
			}
			// #742: a one-colour-N-mana slot adds all N of the picked
			// colour. The planner never selects such a source
			// (gatherTapSources skips it), so this is only reached if
			// that changes — and then it must still mint the amount
			// the printed card does, not one.
			for k := 0; k < slot.AmountFor(color); k++ {
				// Restrictions ride here too. autoTapAbilityFor
				// already refuses restricted abilities, so this is
				// belt-and-braces — but "the auto-tapper is the one
				// path that mints unrestricted copies of restricted
				// mana" is precisely the bug #259 warns about, and one
				// line is cheaper than trusting a filter two files
				// away.
				p.ManaPool.AddMana(ManaToken{
					Color:        color,
					Source:       cardID,
					Restrictions: restrictionsFor(g, ab, p.ID, cardID),
				})
				g.EmitEvent(Event{Kind: EventManaAdded, Actor: p.ID, Source: cardID})
			}
		}
	}
}

// pickColorForSlot consumes one entry from `pending` if any of
// `options` matches a still-unsatisfied requirement, returning
// that color. Otherwise returns the first option (generic-eligible
// fall-through). Empty options returns "".
//
// EVERY slot books its requirement, single-option slots included.
// Issue #273: a one-option slot used to short-circuit straight to
// its colour without ticking the requirement off, so the {W} a
// Plains had just paid stayed on the pending list and the next
// multi-option slot — Command Tower, the only blue source on the
// board — spent itself re-paying it. Teferi's {U} never arrived and
// the strict gate refused a cast the solver had already proved
// payable.
//
// Among the requirements this slot can satisfy, the most restrictive
// one (fewest legal colours) wins. A source that can pay a hybrid
// {W/U} and a plain {U} should take the {U}, leaving the hybrid for
// whatever comes next — same restriction-first instinct the solver
// itself uses when it picks sources.
func pickColorForSlot(options []string, pending *[]ColorRequirement) string {
	if len(options) == 0 {
		return ""
	}
	best, bestOpt := -1, ""
	for i := range *pending {
		req := (*pending)[i]
		for _, opt := range options {
			if !matchColor(opt, req.Options) {
				continue
			}
			if best < 0 || len(req.Options) < len((*pending)[best].Options) {
				best, bestOpt = i, opt
			}
			break
		}
	}
	if best >= 0 {
		*pending = append((*pending)[:best], (*pending)[best+1:]...)
		return bestOpt
	}
	return options[0]
}

// effectiveCostLocked parses the cost the cast actually owes — the
// card's printed ManaCost, or the alternative cost claimed at
// announce (S22) — and adds the commander tax surcharge for casts
// from the command zone
// (CR 903.8 — each previous cast of THIS commander adds {2} to the
// cost). The CommanderCasts counter is incremented AFTER CastSpell
// reaches the stack, so reading it here returns the prior-cast
// count: first cast pays cost+0, second pays cost+2, third pays
// cost+4. Non-command casts return the raw parsed cost. Errors on
// an unparseable ManaCost.
func (g *Game) effectiveCostLocked(p *Player, card Card, params CastSpellParams) (ParsedCost, error) {
	cost, err := g.printedCostLocked(p, card, params)
	if err != nil {
		return ParsedCost{}, err
	}
	// S28: cost modifiers (CR 601.2f) — increases, then reductions,
	// then Trinisphere-style cost-setting effects. Layered AFTER the
	// alternative-cost swap and the commander tax because both of
	// those settle what the spell "would cost", which is the number
	// every modifier is written against: Thalia taxes an overloaded
	// spell's overload cost, and Trinisphere looks at the taxed
	// commander's total rather than the corner of the card.
	//
	// Applied BEFORE the convoke subtraction below for the same
	// reason the tax is: tapping creatures is a way of PAYING the
	// total cost, and CR 601.2f settles the total before anything
	// is paid against it.
	//
	// The zone comes from castZoneFromWire, the same mapping the
	// cast path resolved its source pile with. A private copy of that
	// switch used to live here and knew only hand, command and exile,
	// so a flashback or escape cast reached every modifier as a HAND
	// cast. CastSpell has already refused an unknown string by the
	// time it prices anything; the error below is for a caller that
	// skipped that step.
	fromZone, ok := castZoneFromWire(params.FromZone)
	if !ok {
		return ParsedCost{}, ErrZoneNotFound
	}
	//
	// The targets ride along because CR 601.2c announces them before
	// 601.2f totals the cost, and CastSpell has validated them by now.
	// Only a modifier that declares ReadsTargets ever sees them
	// (ADR 0048 addendum §13).
	cost, err = g.applyCostModifiersLocked(cost, CostQuery{
		Game:       g,
		Card:       card,
		Controller: p.ID,
		FromZone:   fromZone,
		XValue:     params.XValue,
		Targets:    params.Targets,
	})
	if err != nil {
		return ParsedCost{}, err
	}
	// S22: convoke / waterbend. Applied LAST, because it is the only
	// component that spends against the cost rather than adding to
	// it — the tax, the any-colour fold and the alternative-cost swap
	// all have to have settled before we know what the tapped
	// permanents are paying for. A card with no such cost, or a cast
	// that tapped nothing, gets the cost back unchanged.
	tapCost := TapPermanentsCostFor(CatalogKey(card))
	if !tapCost.Empty() {
		cost = tapPermanentsAdjusted(cost, tapCost, g.tapPermanentsPayersLocked(params.TapIDs), params.XValue)
	}
	return cost, nil
}

// printedCostLocked is effectiveCostLocked minus the tap-permanents
// component: the mana this cast owes before convoke or waterbend
// spends anything against it. Split out because the announce-time
// validator needs that number to compute the tap budget, and asking
// effectiveCostLocked for it would be circular — the budget decides
// how many permanents may be tapped, and the tapping is what
// effectiveCostLocked subtracts.
//
// Caller must hold g.mu.
func (g *Game) printedCostLocked(p *Player, card Card, params CastSpellParams) (ParsedCost, error) {
	// S22: an alternative cost replaces the printed one outright
	// (CR 118.9). The commander tax below is layered on top of
	// whichever cost was chosen, because CR 903.8 taxes the cost
	// being paid, not the cost printed in the corner.
	costString := alternativeCostString(card, params.AlternativeCost)
	// S22 airbend: an exile-play grant can carry its own "rather than
	// its mana cost" price ({2}), which belongs to the exiled
	// INSTANCE rather than to the card, so it can't come from the
	// oracle-ID-keyed AlternativeCost catalog above. It wins over the
	// printed cost and is layered BEFORE the commander tax for the
	// same reason the alternative cost is: CR 903.8 taxes whatever
	// cost is actually being paid.
	//
	// The grant prices a cast out of EXILE and nothing else. MoveCard
	// clears it when the card leaves exile (CR 400.7); this check is
	// the second half, so a grant still sitting on a card can't
	// reprice a cast from any other zone.
	fromExile := params.FromZone == string(ZoneExile)
	if ov := card.ExilePlay.CostOverride; ov != "" && fromExile && card.ExilePlay.Active(p.ID, g.Turn.Number) {
		costString = ov
	}
	cost, err := ParseCost(costString)
	if err != nil {
		return ParsedCost{}, err
	}
	if params.FromZone == "command" {
		tax := p.CommanderCasts[card.InstanceID]
		cost.Generic += tax * 2
	}
	// S21 sub-PR 6: "you may spend mana as though it were mana of any
	// color to cast those spells" (Breeches, Brazen Plunderer). Folds
	// the colored slots into the generic demand, which is exactly
	// equivalent for the solver.
	if card.ExilePlay.AnyColor && fromExile && card.ExilePlay.Active(p.ID, g.Turn.Number) {
		cost = asAnyColorCost(cost)
	}
	return cost, nil
}

// castSourceZoneLocked resolves the FromZone string to the zone
// object the card should leave from. A string that names no zone is
// ErrZoneNotFound.
//
// This lookup grants NOTHING on its own — it answers "which pile",
// not "may you". S29's validateCastPathLocked is the permission
// half, and for exile the per-instance ExilePlay grant is checked in
// CastSpell before anything moves.
//
// Graveyard and hand resolve to the CALLER's own zone, which is
// most of what flashback's "from your graveyard" needs: a player
// naming a card in someone else's graveyard simply won't find it.
func (g *Game) castSourceZoneLocked(p *Player, fromZone string) (*Zone, error) {
	kind, ok := castZoneFromWire(fromZone)
	if !ok {
		return nil, ErrZoneNotFound
	}
	switch kind {
	case ZoneHand:
		return p.Hand, nil
	case ZoneCommand:
		return p.Command, nil
	case ZoneGraveyard:
		// S29: the graveyard became castable at all with flashback
		// and escape. Per-player, so the zone lookup already scopes
		// the cast to "your graveyard".
		return p.Graveyard, nil
	case ZoneExile:
		// S21 sub-PR 6: impulse exile. Exile is a SHARED zone, so
		// unlike hand, command and graveyard the zone lookup does not
		// even scope the cast to this player — CastSpell checks the
		// per-card permission before it will move anything.
		return g.Exile, nil
	default:
		return nil, ErrZoneNotFound
	}
}

// sorcerySpeedOpenLocked reports whether the sorcery-speed gate is
// currently open for the given player: caller is the active seat,
// the cursor is on a main phase, and the stack is empty (CR 307.1).
// Caller must hold g.mu.
func (g *Game) sorcerySpeedOpenLocked(playerID uuid.UUID) bool {
	if g.Turn.Step != StepPrecombatMain && g.Turn.Step != StepPostcombatMain {
		return false
	}
	if g.Stack != nil && len(g.Stack.Cards) > 0 {
		return false
	}
	if g.activeSeatIDLocked() != playerID {
		return false
	}
	return true
}

// nextStackSeqLocked mints the insertion sequence stamped onto a
// StackItem as it lands in StackMeta. One above the current maximum
// rather than a persistent counter: the scan is O(items-on-stack)
// (single digits in practice) and survives Clone / RestoreFrom /
// snapshot decode without any extra bookkeeping. Caller must hold
// g.mu.
func (g *Game) nextStackSeqLocked() uint64 {
	var maxSeq uint64
	for _, item := range g.StackMeta {
		if item != nil && item.Seq > maxSeq {
			maxSeq = item.Seq
		}
	}
	return maxSeq + 1
}

// cloneDistributionLocked deep-copies the announce-time distribution
// map so the caller's slice / map can't be mutated through StackMeta.
// Returns nil for an empty input.
func cloneDistributionLocked(in map[uuid.UUID]int) map[uuid.UUID]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// resolveTopOfStackLocked pops the topmost stack item and resolves
// it. For spells, permanents route to the battlefield with their
// recorded controller; instants and sorceries route to the owner's
// graveyard (CR 608.2n). For activated / triggered abilities, the
// item simply ceases to exist (CR 608.2n) — there is no source-card
// movement because the ability's source is a separate card that
// stays put.
//
// S13.1 sub-PR 2 lands the basic resolution path (no target re-check
// yet — that arrives in sub-PR 3; no SBA loop yet — sub-PR 7). The
// caller is responsible for calling this only when the stack is
// non-empty.
//
// Caller must hold g.mu.
func (g *Game) resolveTopOfStackLocked() error {
	if g.Stack == nil || len(g.Stack.Cards) == 0 {
		// No spell on the stack — but there could still be ability
		// items in StackMeta. Find the most recent and resolve it.
		g.resolveTopAbilityLocked()
		return nil
	}
	// Top of the stack is the last card in the slice (LIFO).
	top := g.Stack.Cards[len(g.Stack.Cards)-1]
	item, ok := g.StackMeta[top.InstanceID]
	if !ok || item == nil {
		// Defensive: a card on the stack without a meta entry should
		// not happen — but if it does, route to graveyard so the
		// stack doesn't wedge. No meta means no record of the cost
		// that was paid, so no flashback replacement either.
		return g.routeStackCardToGraveyardLocked(top, "")
	}
	// CR 608.1: spells and abilities share one LIFO stack. An ability
	// item stamped with a higher Seq than the top spell was added
	// later (activated or triggered in response) and must resolve
	// first. Ties — including legacy zero-Seq items from snapshots
	// predating the field — keep the old spell-first behaviour.
	if ab := g.topAbilityLocked(); ab != nil && ab.Seq > item.Seq {
		g.resolveTopAbilityLocked()
		return nil
	}
	delete(g.StackMeta, top.InstanceID)
	defer g.recomputeSplitSecondLocked()

	// Target re-check (CR 608.2b). If the spell declared at least one
	// target (player or card kind) and EVERY target is now illegal
	// (referenced player no longer seated / conceded, referenced card
	// no longer in a zone the engine tracks), the spell is "countered
	// by game rules" — it resolves by going to the owner's graveyard
	// without effect. If only *some* targets have become illegal, the
	// spell still resolves: the engine records what's still valid via
	// the surviving subset, and players resolve the effect manually
	// (sandbox — partial-target effects aren't automated).
	//
	// Self / none targets don't re-check (self is the caster; none
	// has no referent) and count as always-legal for the all-illegal
	// short-circuit.
	if spellAllTargetsIllegalLocked(g, item, castTargetSpecForItem(CatalogKey(top), item)) {
		// "Countered by game rules" — permanents and non-permanents
		// alike go to the owner's graveyard (CR 608.2b). The
		// announce-time choices on StackMeta are discarded along
		// with the item.
		g.EmitEvent(Event{
			Kind:   EventFizzle,
			Actor:  item.Controller,
			Source: top.InstanceID,
			CardID: top.InstanceID,
		})
		// A COPY has no way out of the stack at all: CR 707.10 says
		// it is not a card, so "countered by game rules" leaves it
		// nowhere to go. Checked BEFORE the flashback branch below
		// because a copy of a flashed-back spell is still not a card
		// — the exile replacement has no object to act on.
		if item.IsCopy {
			g.ceaseToExistLocked(top.InstanceID)
			return nil
		}
		// S29: a flashed-back spell that fizzles is still exiled —
		// CR 702.34a replaces every way out of the stack, not just
		// the resolution.
		return g.routeStackCardToGraveyardLocked(top, item.AltCost)
	}
	g.EmitEvent(Event{
		Kind:   EventResolve,
		Actor:  item.Controller,
		Source: top.InstanceID,
		CardID: top.InstanceID,
	})
	// S14: run the catalog's OnResolve hook between target re-check
	// and zone routing. Non-catalog cards return nil (no-op); catalog
	// spells fire their effect here. Errors emit EventEffectError
	// via fireEffectResolverLocked and do not wedge resolution.
	g.fireEffectResolverLocked(item, CatalogKey(top), top.InstanceID)
	// CR 608.3f / 707.10f: a resolving copy of a PERMANENT spell becomes a
	// token. This engine has no token-from-stack-item path, and
	// letting the copy fall through to the battlefield branch below
	// would be worse than doing nothing — it would put a second
	// card-shaped object carrying the original's oracle ID into
	// play, which a bounce spell then duplicates into a hand. Every
	// S30 copy card targets an instant or sorcery, so this is
	// unreachable today; it is written out because it is where the
	// token rule lands (#666).
	if item.IsCopy && top.IsPermanent() {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    item.Controller,
			CardID:   top.InstanceID,
			ErrorMsg: "copying a permanent spell is not implemented (CR 608.3f token)",
		})
		g.ceaseToExistLocked(top.InstanceID)
		return nil
	}
	if top.IsPermanent() {
		// ADR 0034: settle which face the PERMANENT keeps before the
		// replacement pipeline runs, so the entering card's
		// self-replacement is looked up under the right catalog key.
		//
		// An MDFC keeps the face that was cast — the other one never
		// returns — and since S32 so does a `transform` card, which is
		// how a defeated Siege's back face becomes the permanent
		// instead of the battle re-entering the battlefield. CR 712.11's
		// "always cast as its front face" is enforced by CastableFaces
		// refusing to offer the back, so a non-zero transform face here
		// can only have come from an effect that said "cast it
		// TRANSFORMED". An adventure still resolves front-up: its
		// creature half is the permanent no matter which half was cast.
		// See faceOnResolve.
		setFaceInZoneLocked(g.Stack, top.InstanceID, faceOnResolve(top.Layout, top.ActiveFace))
		top.SetFace(faceOnResolve(top.Layout, top.ActiveFace))
		// Permanents resolve to the battlefield with the announce-time
		// controller (which may differ from owner — e.g. cast via a
		// "play this from exile" effect that change controller).
		// S17 sub-PR 4: CR 614 replacement pipeline for enters-tapped
		// (Kismet) + enters-with-counters. Pipeline runs pre-push;
		// canceled permanents stay on the stack (rare in practice —
		// "if X would enter, instead..." effects are edge cases).
		ev := &ReplacementEvent{
			Kind:    RepEventMove,
			CardID:  top.InstanceID,
			OldZone: ZoneStack,
			NewZone: ZoneBattlefield,
			Actor:   item.Controller,
			// S16.5: a resolving permanent's entry can now pause on
			// a prompt — Clone's "choose what to copy" is the first
			// one — and this branch returns to the client when it
			// does. The resume finishes the push on this branch's
			// behalf (executeEntryToBattlefieldLocked).
			//
			// It is flagged resumable only because `stackItem`
			// carries the two things the generic push could not
			// reproduce: the Aura's attach target and the alternative
			// cost the item was paid with. Before that, the pause
			// with no resume was the reason this site was left
			// unflagged — and the reason a CR 616 ordering prompt on
			// a permanent spell's entry dropped the permanent
			// entirely.
			entryResumable: true,
			stackItem:      item,
		}
		// S29: "this creature escapes with a +1/+1 counter on it"
		// (CR 702.138c). Seeded onto the event BEFORE the pipeline
		// runs, so the counters are part of the entry every other
		// replacement gets to see and modify — Doubling Season
		// doubles them — rather than an afterthought stapled on once
		// the permanent has landed.
		g.applyAltCostEntryCountersLocked(ev, top, item)
		out, err := g.applyReplacementsLocked(ev)
		if errors.Is(err, errReplacementPending) {
			return nil
		}
		if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
			g.clearReplacementEventLocked(ev.ID)
			return err
		}
		defer g.clearReplacementEventLocked(ev.ID)
		if out == nil || out.Canceled {
			return nil
		}

		moved, err := MoveCard(g.Stack, g.Battlefield, top.InstanceID)
		if err != nil {
			return err
		}
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == moved.InstanceID {
				g.Battlefield.Cards[i].Controller = item.Controller
				if out.EntersTapped {
					g.Battlefield.Cards[i].Tapped = true
				}
				break
			}
		}
		g.markCardKnownInZoneLocked(g.Battlefield, moved.InstanceID)
		// CR 707.2: a permanent entering as a copy is that copy from
		// the moment it enters, so the values land before the counters
		// (Spark Double's extra +1/+1 goes on the copy) and before any
		// event fires. `moved` is re-taken because it is a pre-copy
		// snapshot and CatalogKey(moved) below would otherwise fire
		// the Clone's own ETB hook rather than the copied card's.
		if copied, ok := g.applyEntersAsCopyLocked(out, moved.InstanceID); ok {
			moved = copied
		}
		for name, n := range out.EntersWithCounters {
			_ = g.AddCounterForEffect(moved.InstanceID, name, n)
		}
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			Actor:   item.Controller,
			CardID:  moved.InstanceID,
			OldZone: ZoneStack,
			NewZone: ZoneBattlefield,
		})
		// S24 / ADR 0036 decision 5: an Aura was cast targeting
		// (CR 303.4a) and enters ATTACHED to what it targeted. This
		// is the only place in the resolution path holding both the
		// landed permanent and the StackItem whose target it was.
		// Between the ZoneMove and the ETB so an ETB trigger already
		// sees the attachment.
		g.attachResolvedAuraLocked(moved.InstanceID, item)
		g.EmitEvent(Event{
			Kind:   EventETB,
			Actor:  item.Controller,
			CardID: moved.InstanceID,
		})
		g.fireETBHookLocked(moved.InstanceID, CatalogKey(moved))
		// S22: evoke's "it's sacrificed when it enters" (CR 702.74a).
		// Queued here because this is the last moment the StackItem —
		// and so the cost that was actually paid — is still reachable.
		g.queueAltCostEntryTriggerLocked(moved, item)
		return nil
	}
	// CR 707.10 — a COPY is not a card, so it has no graveyard to go
	// to and no flashback exile to be caught by either. It ceases to
	// exist, having already run its effect above. See spell_copy.go
	// for why this branch is load-bearing rather than cosmetic.
	if item.IsCopy {
		g.ceaseToExistLocked(top.InstanceID)
		return nil
	}
	// Instants / sorceries: resolve to the owner's graveyard — or to
	// exile, when the flashback cost was paid (CR 702.34a).
	return g.routeStackCardToGraveyardLocked(top, item.AltCost)
}

// spellAllTargetsIllegalLocked reports whether a resolved stack item
// has at least one targeted slot (player or card) and every one of
// those targets is now illegal per CR 608.2b. With a TargetSpec
// (S20) "illegal" is the full predicate — a Doom Blade target that
// became black, a creature that stopped being a creature — checked
// from the item's controller's point of view; without one it's the
// S13.1 existence check. A slot with Kind Self or None is always
// legal. Items with no targets at all return false (nothing to
// re-check).
//
// Caller must hold g.mu.
func spellAllTargetsIllegalLocked(g *Game, item *StackItem, spec *TargetSpec) bool {
	if item == nil || len(item.Targets) == 0 {
		return false
	}
	hadTargeted := false
	anyLegal := false
	for _, t := range item.Targets {
		switch t.Kind {
		case TargetSelf, TargetNone:
			// These aren't "targets" for the re-check — they're fixed
			// references. Treat as always-legal and skip the "had any
			// targeted slot" signal.
			continue
		case TargetPlayer, TargetCard:
			hadTargeted = true
			legal := false
			if spec != nil {
				legal = g.targetLegalLocked(item.Controller, spec, t)
			} else {
				legal = targetStillExistsLocked(g, t)
			}
			if legal {
				anyLegal = true
			}
		}
	}
	if !hadTargeted {
		return false
	}
	return !anyLegal
}

// targetStillExistsLocked performs the CR 608.2b existence check for
// a single TargetRef: the referenced player is still seated and
// non-eliminated, or the referenced card is still in a zone the
// engine tracks (findCardZoneLocked walks every zone). Caller must
// hold g.mu.
func targetStillExistsLocked(g *Game, t TargetRef) bool {
	switch t.Kind {
	case TargetPlayer:
		p := g.playerByIDLocked(t.ID)
		return p != nil && !p.Eliminated
	case TargetCard:
		return g.findCardZoneLocked(t.ID) != nil
	default:
		return true
	}
}

// topAbilityLocked returns the ability item in StackMeta with the
// highest insertion Seq — the most recently added one — or nil when
// no activated / triggered items are pending. Caller must hold g.mu.
func (g *Game) topAbilityLocked() *StackItem {
	var top *StackItem
	for _, item := range g.StackMeta {
		if item == nil || (item.Kind != StackItemActivated && item.Kind != StackItemTriggered) {
			continue
		}
		if top == nil || item.Seq > top.Seq {
			top = item
		}
	}
	return top
}

// resolveTopAbilityLocked resolves the most-recently-added ability
// item in StackMeta (no underlying card on Game.Stack). "Most
// recently added" is the highest insertion Seq — LIFO per CR 608.1,
// deterministic regardless of map-iteration order. Returns nil if
// there are no abilities to resolve. Caller must hold g.mu.
//
// Resolution mirrors the spell path in resolveTopOfStackLocked:
// the item leaves StackMeta, the CR 608.2b target re-check runs
// (every targeted slot illegal → "countered by game rules",
// EventFizzle, no effect), then EventResolve is emitted and the
// item's Effect callback — if any — runs. Errors from Effect
// surface as EventEffectError and do not wedge the stack; the
// ability has ceased to exist either way (CR 608.2n).
func (g *Game) resolveTopAbilityLocked() {
	top := g.topAbilityLocked()
	if top == nil {
		return
	}
	delete(g.StackMeta, top.ID)
	g.recomputeSplitSecondLocked()
	if spellAllTargetsIllegalLocked(g, top, top.targetSpec) {
		g.EmitEvent(Event{
			Kind:   EventFizzle,
			Actor:  top.Controller,
			Source: top.SourceCardID,
			Label:  top.Label,
		})
		return
	}
	g.EmitEvent(Event{
		Kind:   EventResolve,
		Actor:  top.Controller,
		Source: top.SourceCardID,
		Label:  top.Label,
	})
	if top.Effect != nil {
		if err := top.Effect(g, top); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				Actor:    top.Controller,
				Source:   top.SourceCardID,
				ErrorMsg: err.Error(),
			})
		}
	}
}

// routeStackCardToGraveyardLocked moves a card off Game.Stack and
// into its owner's graveyard. Used by resolveTopOfStackLocked for
// instants / sorceries and by the "countered by game rules" path
// (sub-PR 3) when every target is illegal on resolve.
//
// `altCost` is the key of the cost the spell was cast for, because
// S29's flashback replaces this destination: "exile this card
// instead of putting it anywhere else any time it would leave the
// stack" (CR 702.34a). Pass "" for a spell cast for its printed
// cost, and for the defensive no-StackMeta path where there is no
// cost to read.
//
// Caller must hold g.mu.
func (g *Game) routeStackCardToGraveyardLocked(c Card, altCost string) error {
	// S29 flashback. Checked before the owner lookup because it does
	// not depend on one: the destination is the shared exile zone
	// either way, which is also where a card whose owner has left
	// the game already goes.
	//
	// #529: both destinations go through the shared exit primitive,
	// so a commander that fizzles ("countered by game rules", CR
	// 608.2b) or resolves to a graveyard gets the CR 903.9 choice.
	// A queued prompt leaves the card on the stack until the owner
	// answers — priority cannot pass while a choice is outstanding,
	// so nothing resolves on top of it in the meantime.
	r := zoneRoute{CardID: c.InstanceID, Dst: ZoneGraveyard, DstOwner: c.Owner}
	if altCostExilesFromStack(c, altCost) {
		r.Dst, r.DstOwner, r.Actor = ZoneExile, uuid.Nil, c.Owner
	}
	_, err := g.routeCardToZoneLocked(r)
	return err
}

// AbilityParams carries the announce-time choices that flow into an
// activated or triggered ability's stack item. Same shape as
// CastSpellParams minus FromZone (abilities don't move a card) and
// SplitSecond (only spells / activations get the modifier; the
// flag would round-trip on the wire if a future card needed it).
type AbilityParams struct {
	Label        string
	Targets      []TargetRef
	Modes        []int
	XValue       int
	Distribution map[uuid.UUID]int
}

// ActivateAbility creates an activated-ability stack item linked to
// the source card. Implements CR 602: announce → push to stack →
// caller retains priority. Mana abilities are NOT modeled this way
// (CR 605 — they don't use the stack); see the package commentary.
//
// No card moves. The source card stays in its origin zone; the
// stack item carries its own synthetic ID. Resolution removes the
// item (CR 608.2n).
//
// SourceCardID must reference a card that exists in some zone; an
// unknown ID returns ErrCardNotFound.
//
// Caller-gated to priority holder via the action layer; this
// method does not re-check that.
//
// Caller must NOT hold g.mu — this method takes the write lock.
//
// S13.1.
func (g *Game) ActivateAbility(playerID, sourceCardID uuid.UUID, params AbilityParams) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	// Activation legality reads the source's effective types (the
	// CR 302.6 summoning-sickness gate only applies to a creature)
	// and the target predicates read every candidate's. Both are
	// layer-dependent since the type predicates were rerouted
	// through Effective(); fast-path no-op when nothing changed.
	g.RecomputeLayersIfStaleLocked()
	if g.SplitSecondActive {
		return ErrSplitSecondActive
	}
	if g.playerByIDLocked(playerID) == nil {
		return ErrPlayerNotFound
	}
	if g.findCardZoneLocked(sourceCardID) == nil {
		return ErrCardNotFound
	}
	id := uuid.New()
	if g.StackMeta == nil {
		g.StackMeta = make(map[uuid.UUID]*StackItem)
	}
	g.StackMeta[id] = &StackItem{
		ID:           id,
		Kind:         StackItemActivated,
		Controller:   playerID,
		Owner:        playerID,
		SourceCardID: sourceCardID,
		Label:        params.Label,
		Targets:      append([]TargetRef(nil), params.Targets...),
		Modes:        append([]int(nil), params.Modes...),
		XValue:       params.XValue,
		Distribution: cloneDistributionLocked(params.Distribution),
		Seq:          g.nextStackSeqLocked(),
	}
	return nil
}

// ActivateLoyalty applies a planeswalker's loyalty ability by hand.
// Sandbox shape: no stack item, no effect — the loyalty delta is
// applied immediately and the once-per-turn flag is set, and the
// players work out what the ability did between themselves.
//
// This is NOT the path a catalog planeswalker takes. A loyalty
// ability the catalog knows about is an ordinary CR 602 activation
// with an AbilityCost.Loyalty component (see activated.go): it goes
// on the stack, it can target, and it runs an Effect. This action
// survives for the ~thousand planeswalkers with no catalog entry,
// the same way manual `tap` survives for cards whose abilities the
// engine can't express — which is the engine's non-catalog promise
// (ADR 0032).
//
// `delta` is the loyalty change announced by the ability:
// +1 / +2 / -3 / etc, applied to the planeswalker's "loyalty"
// counter.
//
// Returns:
//   - ErrCardNotFound if planeswalkerID is not on the battlefield.
//   - ErrNotAPlaneswalker if it is on the battlefield but isn't one
//     (CR 606.2). Before #329 this action would hand loyalty
//     counters to a Mountain.
//   - ErrCardCallerMismatch if the activator doesn't control it.
//   - ErrInsufficientLoyalty if a negative delta would remove more
//     counters than the planeswalker has (CR 606.6). Paying down to
//     exactly zero is legal; the 704.5i SBA takes it from there.
//   - ErrSorcerySpeedRequired if the gate is closed.
//   - ErrLoyaltyAlreadyActivated if the planeswalker has already
//     activated a loyalty ability this turn (CR 606.3).
//
// Caller must NOT hold g.mu — this method takes the write lock.
//
// S13.1; gates tightened in S27 (#329, #334).
func (g *Game) ActivateLoyalty(playerID, planeswalkerID uuid.UUID, label string, delta int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	// Gated on IsPlaneswalker, which is now a layer read — a
	// creature-land that a Layer-4 effect has turned into a
	// planeswalker has loyalty abilities and one that hasn't
	// doesn't. Fast-path no-op when nothing changed.
	g.RecomputeLayersIfStaleLocked()
	if g.SplitSecondActive {
		return ErrSplitSecondActive
	}
	if g.playerByIDLocked(playerID) == nil {
		return ErrPlayerNotFound
	}
	if !g.sorcerySpeedOpenLocked(playerID) {
		return ErrSorcerySpeedRequired
	}
	if g.LoyaltyActivatedThisTurn[planeswalkerID] {
		return ErrLoyaltyAlreadyActivated
	}
	// Find the planeswalker on the battlefield.
	var pw *Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == planeswalkerID {
			pw = &g.Battlefield.Cards[i]
			break
		}
	}
	if pw == nil {
		return ErrCardNotFound
	}
	// CR 606.2 / 606.3: a loyalty ability belongs to a planeswalker,
	// and only its controller may activate it. Neither was checked
	// before #329, so this action was a counter faucet pointed at
	// any card on the table.
	if !pw.IsPlaneswalker() {
		return ErrNotAPlaneswalker
	}
	// CR 606.1 again, from the other side: a loyalty ability IS an
	// activated ability, so "its activated abilities can't be
	// activated" stops one. Faith's Fetters enchants a permanent,
	// not a creature, and a Fettered planeswalker is the case that
	// makes that clause worth having. S24; see restrictions.go.
	if !CanActivateAbilities(pw) {
		return ErrCantActivate
	}
	if pw.Controller != playerID {
		return ErrCardCallerMismatch
	}
	// CR 606.6: can't remove more loyalty than is there. The old
	// comment here argued for letting the counter go negative so the
	// SBA could see the intent, but the SBA reads "loyalty <= 0" and
	// an activation the rules forbid should be refused at announce,
	// not paid and then cleaned up.
	if delta < 0 && pw.Counters[CounterLoyalty] < -delta {
		return ErrInsufficientLoyalty
	}
	// applyCounterLocked deletes the key at zero and emits the
	// counter event, which is what every other counter mutation in
	// the engine does; the hand-rolled map write here predated it.
	// Paying a cost is not an effect (CR 121.1), so this
	// deliberately bypasses the CR 614 counter-replacement pipeline.
	if err := g.applyCounterLocked(planeswalkerID, CounterLoyalty, delta); err != nil {
		return err
	}
	if g.LoyaltyActivatedThisTurn == nil {
		g.LoyaltyActivatedThisTurn = make(map[uuid.UUID]bool)
	}
	g.LoyaltyActivatedThisTurn[planeswalkerID] = true
	if label != "" {
		// Record a no-card stack item briefly so the wire surfaces
		// the label for the duration of the activation, then drop it.
		// Future "loyalty as a real stack item" would persist this
		// past the action.
		_ = label
	}
	return nil
}

// AnnounceTrigger queues a triggered ability for APNAP-ordered drain
// onto the stack. Per CR 603.3b, all triggers waiting at a priority-
// grant boundary are placed on the stack in active-player-non-active-
// player order, with each affected player choosing the relative
// order of their own simultaneous triggers (here: the order they
// announce them).
//
// Sandbox: the player whose card has a triggered ability clicks
// "trigger" on the card, optionally provides a label and target
// list, and the engine queues an item. The drain happens in
// drainPendingTriggersAPNAPLocked() — called from PassPriority and
// every other priority-grant path.
//
// Caller must NOT hold g.mu — this method takes the write lock.
//
// S13.1.
func (g *Game) AnnounceTrigger(playerID, sourceCardID uuid.UUID, params AbilityParams) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.playerByIDLocked(playerID) == nil {
		return ErrPlayerNotFound
	}
	if g.findCardZoneLocked(sourceCardID) == nil {
		return ErrCardNotFound
	}
	id := uuid.New()
	g.PendingTriggers = append(g.PendingTriggers, &StackItem{
		ID:           id,
		Kind:         StackItemTriggered,
		Controller:   playerID,
		Owner:        playerID,
		SourceCardID: sourceCardID,
		Label:        params.Label,
		Targets:      append([]TargetRef(nil), params.Targets...),
		Modes:        append([]int(nil), params.Modes...),
		XValue:       params.XValue,
		Distribution: cloneDistributionLocked(params.Distribution),
	})
	g.EmitEvent(Event{
		Kind:   EventTrigger,
		Actor:  playerID,
		Source: sourceCardID,
		Label:  params.Label,
	})
	// CR 603.3d: a manually-announced trigger chooses its targets as
	// it goes on the stack, same as the harvested kind.
	g.emitBecameTargetLocked(playerID, sourceCardID, id, params.Targets)
	return nil
}

// isPublicZone reports whether a card sitting in a zone of this
// kind is visible to every seated player (CR 400.2). Used by the
// S13.5 KnownBy machinery to decide whether a zone-move should
// grant knowledge to everyone.
func isPublicZone(k ZoneKind) bool {
	switch k {
	case ZoneBattlefield, ZoneStack, ZoneExile, ZoneGraveyard, ZoneCommand:
		return true
	}
	return false
}

// markCardKnownInZoneLocked is the canonical post-move hook: after
// a card lands in `zone`, mark the appropriate viewers as knowers.
// Public zones grant knowledge to every seated player; the hand
// grants knowledge only to its owner; library grants nothing
// (cards in the library have no knowers until they're drawn or
// scryd into knowledge).
//
// Caller must hold g.mu.
func (g *Game) markCardKnownInZoneLocked(zone *Zone, cardID uuid.UUID) {
	if zone == nil {
		return
	}
	var idsToAdd []uuid.UUID
	switch {
	case isPublicZone(zone.Kind):
		idsToAdd = make([]uuid.UUID, 0, len(g.Seats))
		for _, p := range g.Seats {
			idsToAdd = append(idsToAdd, p.ID)
		}
	case zone.Kind == ZoneHand:
		// Only the hand's owner knows newly-arrived cards.
		idsToAdd = []uuid.UUID{zone.Owner}
	default:
		return
	}
	for i := range zone.Cards {
		if zone.Cards[i].InstanceID == cardID {
			zone.Cards[i].AddKnowersAll(idsToAdd)
			return
		}
	}
}

// clearKnownInZoneLocked drops the KnownBy set on every card in the
// zone. Used by ShuffleLibrary to reset the per-card knowledge
// when the order is no longer determinable.
func clearKnownInZoneLocked(zone *Zone) {
	if zone == nil {
		return
	}
	for i := range zone.Cards {
		zone.Cards[i].ClearKnown()
	}
}

// stateBasedActionsLocked runs one pass of state-based actions per
// CR 704.5. Returns true if any SBA fired so the caller can re-run
// the loop (CR 704.3 — SBAs repeat until none fire).
//
// Player-loss SBAs (S13.1):
//   - 704.5a: a player at 0 or less life loses
//   - 704.5b: a player who tried to draw from an empty library loses
//   - 704.6c / 903.10a: 21 commander damage from a single source
//
// Player-loss SBAs (S13.2):
//   - 704.5c: a player with ≥ 10 poison counters loses
//
// Permanent SBAs (S13.1):
//   - 704.5f: a creature with 0 or less toughness is destroyed
//   - 704.5g: a creature with damage marked >= toughness is destroyed
//
// Counter SBAs (S13.2):
//   - 704.5i: a planeswalker with 0 loyalty counters is moved to its
//     owner's graveyard
//   - 704.5v: a battle with 0 defense counters is moved to its
//     owner's graveyard
//   - 704.5q: +1/+1 and -1/-1 counters on the same creature
//     cancel out — remove min(N, M) of each
//   - 704.5s: a saga whose final-chapter lore counter is set is
//     sacrificed by its controller (the SBA half; the lore-counter
//     advance trigger lands in S14+ with the effect catalog)
//
// Counter ordering (CR 704.3): the +1/+1 / -1/-1 cancel runs BEFORE
// the lethal-damage check so a 2/2 with one +1/+1 and one -1/-1 +
// 1 marked damage doesn't die — the counters cancel first, leaving
// it a 2/2 with 1 damage. The implementation enforces this by
// running the counter cancel pass before destruction collection.
//
// Caller must hold g.mu.
func (g *Game) stateBasedActionsLocked() bool {
	if g.State != StateActive {
		return false
	}
	// S16: refresh effective characteristics before any toughness /
	// loyalty / battle-defense check. Counter mutations + zone moves
	// from prior SBA iterations bump layerVersion; this fast-paths
	// when nothing's changed. Without it, the lethal-damage SBA
	// would see printed toughness instead of post-anthem effective
	// (a 2/2 + Glorious Anthem under 3 marked damage would die
	// because CurrentToughness reads Effective().Toughness == 3).
	g.RecomputeLayersIfStaleLocked()
	fired := false

	// S24 / CR 704.5m + 704.5n: an Equipment attached to something
	// that is no longer a creature becomes unattached; an Aura
	// attached to something it can no longer legally enchant is put
	// into its owner's graveyard.
	//
	// Runs AFTER the recompute (the legality test reads effective
	// types) and BEFORE the destruction pre-pass, so a creature that
	// becomes lethally damaged because its +2/+2 Aura fell off dies
	// in the same settling rather than surviving a round.
	if g.attachmentSBALocked() {
		fired = true
	}

	// Counter cancel (704.5q). Must run before destruction so the
	// post-cancel state is what the lethal-damage SBA sees.
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.IsCreature() || c.Counters == nil {
			continue
		}
		plus := c.Counters["+1/+1"]
		minus := c.Counters["-1/-1"]
		if plus > 0 && minus > 0 {
			cancel := plus
			if minus < cancel {
				cancel = minus
			}
			c.Counters["+1/+1"] -= cancel
			c.Counters["-1/-1"] -= cancel
			if c.Counters["+1/+1"] <= 0 {
				delete(c.Counters, "+1/+1")
			}
			if c.Counters["-1/-1"] <= 0 {
				delete(c.Counters, "-1/-1")
			}
			if len(c.Counters) == 0 {
				c.Counters = nil
				c.LostLastCounter = true
			}
			fired = true
		}
	}

	// Player-loss SBAs.
	for _, p := range g.Seats {
		if p.Eliminated {
			continue
		}
		if p.Life <= 0 || p.AttemptedEmptyDraw || p.IsDeadByCommanderDamage() {
			g.eliminatePlayerLocked(p)
			fired = true
			continue
		}
		// 704.5c: poison ≥ 10. Player.Counters is the S13.2 map;
		// the legacy single-int Player.Poison field stays in sync
		// via SetPoison so old action paths keep working.
		poison := 0
		if p.Counters != nil {
			poison = p.Counters[CounterPoison]
		}
		if poison < p.Poison {
			poison = p.Poison
		}
		if poison >= PoisonLethal {
			g.eliminatePlayerLocked(p)
			fired = true
		}
	}

	// Permanent + counter destruction SBAs. Collect doomed instance
	// IDs in a pre-pass to avoid mutating the slice while iterating.
	//
	// Printed-0 creatures (Toughness == 0 and no counters applied)
	// are skipped: that's the placeholder / unparseable-stats
	// convention documented on Card.Power — the SBA would else
	// destroy every demo seed card. A creature that has LOST its
	// counters is not skipped (Card.LostLastCounter, #683): a 0/0
	// whose last +1/+1 counter was removed or cancelled has a real 0
	// toughness and dies (CR 704.5f).
	// Unless its printed toughness is not a number
	// (Card.VariableToughness): a `*` creature's 0 is the import
	// stand-in whether or not it once had counters, so it stays
	// skipped.
	var doomed []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		// #605: a permanent whose exit is already paused on a player
		// prompt is still HERE, with whatever doomed it intact — a
		// commander at zero toughness keeps its zero toughness while
		// its owner is asked about the command zone. Dooming it again
		// queues a second prompt for a move that is already in flight,
		// and since destroying it counts as fired, runStateChecksLocked
		// goes round again and does it thirty more times. Every one of
		// those siblings becomes unanswerable the moment the first is
		// answered. The move is already asked; wait for the answer.
		if g.zoneChangePausedLocked(c.InstanceID) {
			continue
		}
		if c.IsCreature() {
			if c.Toughness == 0 && len(c.Counters) == 0 && (!c.LostLastCounter || c.VariableToughness) {
				continue
			}
			curT := c.CurrentToughness()
			// CR 704.5f — toughness 0 or less PUTS the creature into
			// its owner's graveyard; it does not destroy it, so
			// indestructible deliberately does not save it here. A
			// 2/2 with indestructible under two -1/-1 counters dies.
			if curT <= 0 {
				doomed = append(doomed, c.InstanceID)
				continue
			}
			// S25 (#77): the two damage-driven creature SBAs below
			// are both DESTRUCTION (CR 704.5g, CR 704.5h), so
			// indestructible switches both off (CR 702.12b). The
			// damage stays marked either way — see indestructible.go.
			indestructible := IsIndestructible(&c)
			if c.DamageMarked >= curT && !indestructible {
				doomed = append(doomed, c.InstanceID)
				continue
			}
			// S18 sub-PR 3: CR 702.2c — a creature hit by any nonzero
			// damage from a deathtouch source is destroyed at the
			// next SBA regardless of toughness. The flag stays set
			// until cleanup (or the card's destruction, via the
			// zone-move listener) so a subsequent SBA pass on the
			// same event cycle doesn't "un-doom" the creature.
			if c.MarkedLethalByDeathtouch && !indestructible {
				doomed = append(doomed, c.InstanceID)
			}
			continue
		}
		// 704.5i — planeswalker with 0 loyalty counters.
		if c.IsPlaneswalker() {
			if c.Counters == nil || c.Counters[CounterLoyalty] <= 0 {
				doomed = append(doomed, c.InstanceID)
			}
			continue
		}
		// 704.5v — battle with 0 defense counters.
		if c.IsBattle() {
			if c.Counters == nil || c.Counters[CounterDefense] <= 0 {
				doomed = append(doomed, c.InstanceID)
			}
			continue
		}
	}
	// S27 / CR 310.12b: a battle at zero defense is DEFEATED, and the
	// defeated trigger has to see it on the battlefield. Announced
	// here — after the doomed set is collected and before any of it
	// moves — so the harvester finds a live source, and so a Siege's
	// "exile it, then cast it transformed" reads a card that still
	// exists. The battle is in `doomed` already via the 704.5v arm
	// above; this only adds the announcement.
	for _, c := range g.Battlefield.Cards {
		if c.IsBattle() && c.Counters[CounterDefense] <= 0 {
			g.emitBattleDefeatedLocked(c.InstanceID)
		}
	}
	// S23: one SBA pass is ONE event (CR 704.3 — all applicable
	// state-based actions are performed simultaneously as a single
	// event). The collection above already worked off a single
	// consistent board; routing the executions through the
	// simultaneous-exit batch makes the TRIGGERS agree, so a Blood
	// Artist that Pyroclasm or Toxic Deluge killed alongside the rest
	// of the board still sees every one of those deaths. See
	// simultaneous.go.
	if g.destroyPermanentsLocked(doomed) > 0 {
		fired = true
	}

	// 704.5s (S27) — a Saga at or past its final chapter, with no
	// chapter ability of its own still on the stack, is SACRIFICED by
	// its controller. Separate from the `doomed` loop above because
	// sacrifice is not destruction: it emits EventSacrifice (which
	// aristocrats payoffs watch) and it ignores indestructible.
	//
	// Runs after the destruction pass so a Saga that was also going
	// to die for another reason has already gone, and the "chapter
	// still on the stack" check sees the settled queue.
	for _, id := range g.sagasReadyToSacrificeLocked() {
		// Same re-entry guard as the doomed sweep above: a Saga
		// commander waiting on the CR 903.9 prompt is still on the
		// battlefield at its final chapter, and sacrificing it a
		// second time would queue a second prompt and emit a second
		// EventSacrifice for one sacrifice.
		if g.zoneChangePausedLocked(id) {
			continue
		}
		if err := g.sacrificePermanentLocked(id); err == nil {
			fired = true
		}
	}

	// 704.5j (S27) — the legend rule. Last, because it is the one
	// state-based action whose outcome is a CHOICE rather than a
	// consequence: running it after the destruction and sacrifice
	// passes means a player is never asked to pick between two
	// legends when one of them was about to leave anyway. See
	// legend_rule.go.
	if g.queueLegendRuleChoicesLocked() {
		fired = true
	}

	return fired
}

// runStateChecksLocked runs the SBA + APNAP-trigger-drain loop until
// the game is quiet (no SBAs fire AND no triggers are pending).
// CR 704.3 + 603.3b — both checks are paired at every priority-grant
// boundary.
//
// Bounded at 32 iterations as a safety belt against an unintended
// SBA / trigger ping-pong; in practice the loop terminates after at
// most a handful of passes (one creature destroyed → one trigger
// drained onto the stack → one quiet re-check).
//
// Each iteration drains PendingTriggers onto the stack via the
// APNAP drain — the drain empties the queue, so the loop terminates
// once SBAs go quiet. (An earlier revision recursed on itself here
// instead of draining; with nothing emptying the queue, a manually
// announced trigger followed by any state check overflowed the
// stack.)
//
// Caller must hold g.mu.
func (g *Game) runStateChecksLocked() {
	const maxIter = 32
	for i := 0; i < maxIter; i++ {
		fired := g.stateBasedActionsLocked()
		hasPending := len(g.PendingTriggers) > 0
		if !fired && !hasPending {
			return
		}
		if hasPending && !g.drainPendingTriggersAPNAPLocked() && !fired {
			// Held behind a CR 603.3b ordering prompt and SBAs are
			// quiet — nothing more to do until the chooser answers.
			return
		}
	}
}

// eliminatePlayerLocked transitions a seated player to eliminated
// state, cleans up their stack items + pending triggers, walks the
// cursor past them if they were the active seat, and runs the
// game-end check. Used by Concede AND by the SBA loop. Caller must
// hold g.mu.
func (g *Game) eliminatePlayerLocked(p *Player) {
	if p.Eliminated {
		return
	}
	p.Eliminated = true
	p.AttemptedEmptyDraw = false
	g.cleanupStackForEliminatedLocked(p.ID)
	g.advancePastEliminatedLocked()
	g.EmitEvent(Event{
		Kind:  EventPlayerEliminated,
		Actor: p.ID,
	})
	// Game-end check.
	survivors := 0
	for _, s := range g.Seats {
		if !s.Eliminated {
			survivors++
		}
	}
	if survivors <= 1 {
		g.State = StateEnded
	}
}

// cleanupStackForEliminatedLocked implements CR 800.4a — when a
// player leaves the game, every spell and ability they control on
// the stack ceases to exist. Spell items have their card removed
// from Game.Stack to exile (closest analogue to "cease to exist").
// Ability items are just deleted from StackMeta. Pending triggers
// controlled by the eliminated player are dropped from the queue.
//
// Targets on remaining stack items pointing at the eliminated
// player are NOT scrubbed here — the existing target re-check at
// resolution time (CR 608.2b, see spellAllTargetsIllegalLocked)
// already turns those slots illegal, so the spell either resolves
// partially or is countered by game rules at the moment the
// mechanic actually matters.
//
// Caller must hold g.mu.
//
// S13.1.
func (g *Game) cleanupStackForEliminatedLocked(playerID uuid.UUID) {
	if len(g.StackMeta) > 0 {
		toRemove := make([]uuid.UUID, 0, len(g.StackMeta))
		for id, item := range g.StackMeta {
			if item != nil && item.Controller == playerID {
				toRemove = append(toRemove, id)
			}
		}
		for _, id := range toRemove {
			item := g.StackMeta[id]
			delete(g.StackMeta, id)
			if item != nil && item.Kind == StackItemSpell && g.Stack != nil && g.Stack.Contains(id) {
				_, _ = MoveCard(g.Stack, g.Exile, id)
			}
		}
	}
	if len(g.PendingTriggers) > 0 {
		kept := g.PendingTriggers[:0]
		for _, t := range g.PendingTriggers {
			if t == nil || t.Controller != playerID {
				kept = append(kept, t)
			}
		}
		g.PendingTriggers = kept
	}
	// A choice owed by a player who has left the game can never be
	// answered, and while it sits in the queue every other seat is
	// blocked behind it (pass_priority is refused client-side and the
	// bot enumerator offers nothing while a choice is open). Drop the
	// eliminated player's prompts — their objects are gone with them
	// (CR 800.4a), so a damage assignment, target pick or scry they
	// owed has nothing left to act on. Same for a cleanup-discard
	// pause in their name. Found by the S31 bot fuzzer.
	if len(g.PendingChoices) > 0 {
		kept := g.PendingChoices[:0]
		for _, c := range g.PendingChoices {
			if c == nil || c.Chooser != playerID {
				kept = append(kept, c)
			}
		}
		g.PendingChoices = kept
		if len(g.PendingChoices) == 0 {
			g.PendingChoices = nil
		}
	}
	if g.DiscardPending != nil {
		delete(g.DiscardPending, playerID)
		if len(g.DiscardPending) == 0 {
			g.DiscardPending = nil
		}
	}
	g.recomputeSplitSecondLocked()
}

// routeBattlefieldCardToOwnerGraveyardLocked moves a battlefield
// card to its owner's graveyard, clearing battlefield-only state
// (combat declarations and position are zeroed by MoveCard's CR 400.7
// cleanup; DamageMarked is cleared by executeBattlefieldLeaveLocked,
// since MoveCard predates the field). Used by SBAs that destroy
// creatures, by every effect destroy, and — because sacrifice is also
// a battlefield exit, though it is not destruction — by sacrifice.go.
//
// #708: THE DAMAGE IS NOT CLEARED HERE. It used to be, on the line
// that found the owner, which is BEFORE the CR 614 window below has
// had a chance to replace the exit. Two things were wrong with that.
// A replacement that keeps the permanent on the battlefield left it
// standing with its damage erased — and damage stays marked until the
// cleanup step (CR 514.2), not until something tried to destroy it —
// so the CR 704.5g lethal-damage check would not see it again. And a
// replacement that wants to READ how much damage is on the permanent
// ("if it would be destroyed, instead …") was handed a zero.
//
// So the clear moved to the LANDED outcome, executeBattlefieldLeaveLocked,
// which is the same terminal-outcome shape damage_tail.go and
// life_tail.go use: the mutation happens where the event actually
// resolves, and the paused and unpaused paths reach it through one
// function. Regeneration, when it ships, removes the damage in its own
// replacement (CR 701.15a says the shield does it), not as a side
// effect of the destroy path.
//
// If the owner is no longer seated, the card lands in exile so the
// engine doesn't carry a stale reference. Caller must hold g.mu.
func (g *Game) routeBattlefieldCardToOwnerGraveyardLocked(cardID uuid.UUID) error {
	var owner *Player
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			owner = g.playerByIDLocked(g.Battlefield.Cards[i].Owner)
			break
		}
	}

	// S17 sub-PR 6: route through the CR 614 replacement pipeline
	// so the CR 903.9 commander-zone built-in can fire for dies-to-
	// damage + wrath + SBA destroys. The built-in is Optional, so
	// when the dying card is a commander the pipeline queues a
	// yes/no prompt for the owner; the physical move waits for
	// their answer via ResolveOptionalReplacement.
	var defaultDest ZoneKind
	var defaultOwner uuid.UUID
	if owner == nil {
		defaultDest = ZoneExile
	} else {
		defaultDest = ZoneGraveyard
		defaultOwner = owner.ID
	}
	ev := &ReplacementEvent{
		Kind:         RepEventMove,
		CardID:       cardID,
		OldZone:      ZoneBattlefield,
		NewZone:      defaultDest,
		NewZoneOwner: defaultOwner,
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// Prompt queued; resume runs the move after the owner
		// answers. Return nil so SBA caller doesn't report failure.
		return nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		return nil
	}
	return g.executeBattlefieldLeaveLocked(cardID, out.NewZone, out.NewZoneOwner, owner)
}

// executeBattlefieldLeaveLocked runs the physical zone move after
// the replacement pipeline has settled on a destination. Factored
// out of routeBattlefieldCardToOwnerGraveyardLocked so the resume
// path (ResolveOptionalReplacement → applyResolvedReplacementEventLocked)
// shares the same implementation.
//
// #708: this is also where DamageMarked is cleared, because this is
// the one place the permanent actually LEAVES. A destruction that a
// replacement turned into something else never gets here, and a
// permanent that is still on the battlefield keeps its damage until
// the cleanup step (CR 514.2) like every other damaged permanent.
//
// Caller must hold g.mu.
func (g *Game) executeBattlefieldLeaveLocked(cardID uuid.UUID, dest ZoneKind, destOwner uuid.UUID, owner *Player) error {
	var destZone *Zone
	var actor uuid.UUID
	switch dest {
	case ZoneCommand:
		// CR 903.9 commander-zone replacement landed. Find the
		// owner via the card — defaultOwner may have been empty
		// when the original owner had left the game.
		card, ok := g.LookupCardForEffect(cardID)
		if !ok {
			return ErrCardNotFound
		}
		cmdOwner := g.playerByIDLocked(card.Owner)
		if cmdOwner == nil {
			// Owner gone — fall back to exile.
			destZone = g.Exile
			dest = ZoneExile
		} else {
			destZone = cmdOwner.Command
			actor = cmdOwner.ID
		}
	case ZoneExile:
		destZone = g.Exile
	case ZoneGraveyard:
		if owner != nil {
			destZone = owner.Graveyard
			actor = owner.ID
		} else {
			p := g.playerByIDLocked(destOwner)
			if p == nil {
				destZone = g.Exile
				dest = ZoneExile
			} else {
				destZone = p.Graveyard
				actor = p.ID
			}
		}
	case ZoneHand:
		if owner != nil {
			destZone = owner.Hand
			actor = owner.ID
		} else {
			p := g.playerByIDLocked(destOwner)
			if p == nil {
				return ErrPlayerNotFound
			}
			destZone = p.Hand
			actor = p.ID
		}
	case ZoneLibrary:
		if owner != nil {
			destZone = owner.Library
			actor = owner.ID
		} else {
			p := g.playerByIDLocked(destOwner)
			if p == nil {
				return ErrPlayerNotFound
			}
			destZone = p.Library
			actor = p.ID
		}
	default:
		return ErrZoneNotFound
	}
	// #708: the damage goes now, and only now. MoveCard's CR 400.7
	// cleanup handles tapped, counters, position and the combat
	// declarations; DamageMarked postdates it and is cleared here.
	//
	// Before the LKI snapshot, so the last-known information a
	// dies-trigger reads is byte-for-byte what it was before this
	// moved — the clear used to happen even earlier (at the top of
	// routeBattlefieldCardToOwnerGraveyardLocked) and nothing has ever
	// seen a dying permanent's damage. Moving WHERE it is cleared is
	// this fix; changing what LKI shows is not, and would be its own
	// decision.
	g.clearBattlefieldDamageLocked(cardID)
	g.snapshotLKILocked(cardID)
	if _, err := MoveCard(g.Battlefield, destZone, cardID); err != nil {
		return err
	}
	g.markCardKnownInZoneLocked(destZone, cardID)
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		Actor:   actor,
		CardID:  cardID,
		OldZone: ZoneBattlefield,
		NewZone: dest,
	})
	g.EmitEvent(Event{Kind: EventLTB, CardID: cardID, Actor: actor, NewZone: dest})
	// A permanent leaving the battlefield is the one event that can
	// invalidate a queued "sacrifice a creature of your choice" prompt
	// (Grave Pact), so re-check them here rather than in
	// runStateChecksLocked — a queued choice stops priority from
	// passing, so the state-check loop is exactly what does NOT run
	// while such a prompt is outstanding. No-op when none is queued.
	g.pruneSacrificeChoicesLocked()
	// #605: and any sibling prompt still asking about a move of THIS
	// card off the battlefield it has now left is unanswerable, for
	// the same reason and at the same moment.
	g.pruneStaleZoneChangeChoicesLocked()
	return nil
}

// MarkDamage adjusts the damage noted on a creature on the
// battlefield by `delta` (positive to add damage, negative to
// remove). Drives the lethal-damage SBA (CR 704.5g). Caller-gated
// to the controller in the action layer for the additive case;
// admins / spectators may apply negative deltas to undo.
//
// Returns ErrCardNotFound if the cardID isn't on the battlefield.
// SBA loop fires after the mutation so a lethal mark applies
// immediately.
//
// S13.1.
func (g *Game) MarkDamage(cardID uuid.UUID, delta int) error {
	return g.markDamageWithKind(uuid.Nil, cardID, delta, false)
}

// MarkCombatDamage is the combat-damage entry point used by the
// combat resolver. Flags IsCombatDamage on the replacement event
// so Fog-class effects (cancel combat damage this turn) key off
// the right subset without intercepting spell damage too. Source
// is the attacker/blocker whose damage is being dealt.
//
// S17 sub-PR 2.
func (g *Game) MarkCombatDamage(source, cardID uuid.UUID, delta int) error {
	return g.markDamageWithKind(source, cardID, delta, true)
}

// markDamageWithKind is the shared body for MarkDamage and
// MarkCombatDamage. Takes g.mu; routes through the replacement
// pipeline with ev.IsCombatDamage set from the caller. Sub-PR 2
// registers zero damage-replacement effects, so behavior is byte-
// for-byte identical to pre-S17.
func (g *Game) markDamageWithKind(source, cardID uuid.UUID, delta int, isCombat bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	ev := &ReplacementEvent{
		Kind:           RepEventDamage,
		Source:         source,
		DamageSource:   source,
		DamageTarget:   cardID,
		DamageAmount:   delta,
		IsCombatDamage: isCombat,
		// #694: the manual mark keeps its own tail shape — a signed
		// delta straight onto DamageMarked, no CR 120.3 split, no
		// combat riders. It is the sandbox verb, and a negative delta
		// (undo a mark) is a legitimate use of it.
		damageTail: &damageTail{kind: damageTailManualMark},
	}
	paused, err := g.damageThroughReplacementsLocked(ev)
	if err != nil {
		return err
	}
	if paused {
		// CR 616 ordering prompt queued; the mark lands from the
		// resume, which runs its own sweep.
		return nil
	}
	// The sandbox verb is its own action boundary, so it owes the SBA
	// sweep the tail deliberately does not run (see damage_tail.go).
	g.runStateChecksLocked()
	return nil
}

// drainPendingTriggersAPNAPLocked moves every queued triggered
// ability onto the stack in APNAP order: active-player's first,
// then turn-order clockwise around the table. Within a single
// player's batch, the queue order is preserved (the caller chose
// the order via the order they called AnnounceTrigger).
//
// Called from PassPriority and other priority-grant boundaries.
// Caller must hold g.mu.
//
// Returns false when the queue is being held behind a CR 603.3b
// ordering prompt (S19 sub-PR 8) — the caller's loop should stop
// spinning until ResolveTriggerOrder re-runs the drain. Returns
// true when the queue is empty or was drained.
//
// S13.1.
func (g *Game) drainPendingTriggersAPNAPLocked() bool {
	if len(g.PendingTriggers) == 0 {
		return true
	}
	numSeats := len(g.Seats)
	if numSeats == 0 {
		return true
	}
	// Bucket triggers by controller seat so we can drain in seat
	// order. Preserves per-controller queue order via stable
	// iteration.
	bySeat := make(map[int][]*StackItem)
	for _, t := range g.PendingTriggers {
		seat := -1
		for i, p := range g.Seats {
			if p.ID == t.Controller {
				seat = i
				break
			}
		}
		if seat == -1 {
			// Controller no longer seated — drop the trigger.
			continue
		}
		bySeat[seat] = append(bySeat[seat], t)
	}
	// CR 603.3b: a player with two or more differing simultaneous
	// triggers chooses their relative order. Ask each such seat
	// (once — items already Ordered don't re-prompt) and hold the
	// WHOLE queue until every ordering prompt is answered, so the
	// APNAP placement below still sees all seats at once.
	held := false
	for seat, items := range bySeat {
		if !seatNeedsTriggerOrder(items) {
			continue
		}
		held = true
		p := g.Seats[seat]
		if g.hasTriggerOrderPromptLocked(p.ID) {
			continue
		}
		ids := make([]uuid.UUID, len(items))
		for i, t := range items {
			ids[i] = t.ID
		}
		g.QueueChoiceForEffect(PendingChoice{
			Kind:            PendingChoiceTriggerOrder,
			Chooser:         p.ID,
			Count:           len(ids),
			Reason:          "Order your triggers",
			TriggerOrderIDs: ids,
		})
	}
	if held {
		return false
	}
	g.PendingTriggers = nil
	if g.StackMeta == nil {
		g.StackMeta = make(map[uuid.UUID]*StackItem)
	}
	// APNAP: active player first, then clockwise. Seq increases in
	// drain order, so the active player's triggers (placed first)
	// resolve last — CR 603.3b stacking order. One scan mints the
	// starting Seq; incrementing locally keeps the drain O(triggers)
	// instead of rescanning StackMeta per placement.
	seq := g.nextStackSeqLocked()
	for offset := 0; offset < numSeats; offset++ {
		seat := (g.Turn.ActiveSeat + offset) % numSeats
		for _, t := range bySeat[seat] {
			t.Seq = seq
			seq++
			g.StackMeta[t.ID] = t
		}
	}
	g.recomputeSplitSecondLocked()
	return true
}

// seatNeedsTriggerOrder reports whether a seat's batch of pending
// triggers needs a CR 603.3b ordering prompt: at least two items,
// not all identical (same source card + same label — two Bident
// draws are interchangeable and asking would be noise), and at
// least one not yet Ordered by an answered prompt.
func seatNeedsTriggerOrder(items []*StackItem) bool {
	if len(items) < 2 {
		return false
	}
	allOrdered := true
	allSame := true
	for _, t := range items {
		if !t.Ordered {
			allOrdered = false
		}
		if t.SourceCardID != items[0].SourceCardID || t.Label != items[0].Label {
			allSame = false
		}
	}
	return !allOrdered && !allSame
}

// hasTriggerOrderPromptLocked reports whether chooser already has a
// PendingChoiceTriggerOrder waiting. Caller must hold g.mu.
func (g *Game) hasTriggerOrderPromptLocked(chooser uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceTriggerOrder && c.Chooser == chooser {
			return true
		}
	}
	return false
}

// DiscardSelection processes the active player's interactive
// cleanup-step discard (S13.4). Validates that the caller is in the
// pending map, that the supplied card IDs all live in the caller's
// hand, and that the count exactly matches the over-max amount the
// engine recorded at cleanup entry. On success, moves each card to
// the caller's graveyard and clears their pending entry. When the
// pending map drains, re-fires the cleanup hook so the cursor
// resumes its auto-advance.
//
// Returns ErrInvalidParam for wrong count, ErrCardNotFound for IDs
// not in the caller's hand. Returns nil and a no-op for callers not
// in the pending map (idempotent — clients can dismiss-without-
// dispatching).
//
// Caller must NOT hold g.mu — this method takes the write lock.
//
// S13.4.
func (g *Game) DiscardSelection(playerID uuid.UUID, cardIDs []uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	want, owed := g.DiscardPending[playerID]
	if !owed {
		// Not in the pending map — caller has nothing to do; treat
		// as no-op so a stale dismiss doesn't error.
		return nil
	}
	if len(cardIDs) != want {
		return ErrInvalidParam
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	// Pre-validate: every supplied card must be in the caller's hand.
	for _, id := range cardIDs {
		if !p.Hand.Contains(id) {
			return ErrCardNotFound
		}
	}
	for _, id := range cardIDs {
		if _, err := MoveCard(p.Hand, p.Graveyard, id); err != nil {
			return err
		}
		g.markCardKnownInZoneLocked(p.Graveyard, id)
		g.EmitEvent(Event{
			Kind:    EventDiscardCard,
			Actor:   playerID,
			CardID:  id,
			OldZone: ZoneHand,
			NewZone: ZoneGraveyard,
		})
	}
	delete(g.DiscardPending, playerID)
	if len(g.DiscardPending) == 0 {
		g.DiscardPending = nil
	}
	// Resume the cleanup auto-advance if the pending map is now
	// empty. Re-fires runStepEntryHooksLocked which rechecks
	// DiscardPending; with the map empty, the auto-advance branch
	// runs and the cursor walks on to the next seat's Untap.
	if len(g.DiscardPending) == 0 && g.Turn.Step == StepCleanup {
		g.runStepEntryHooksLocked()
	}
	return nil
}

// SetMaxHandSize updates the named player's per-player hand-size
// cap. Sandbox-only — used by the S14+ effect catalog for
// Reliquary Tower / Thought Vessel / Library of Leng / Spellbook /
// Null Profusion / Venser's Journal style cards. Until the catalog
// lands, this is also exposed as a manual sandbox helper for
// playgroup adjustments.
//
// `value` is clamped to NoMaxHandSize (-1) for "no cap"; any other
// negative value is rejected with ErrInvalidParam. Doesn't fire
// the SBA loop — the cap only matters at cleanup-step entry, which
// has its own re-check via populateDiscardPendingLocked.
//
// Caller must NOT hold g.mu — this method takes the write lock.
//
// S13.4.
func (g *Game) SetMaxHandSize(playerID uuid.UUID, value int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if value < NoMaxHandSize {
		return ErrInvalidParam
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	p.MaxHandSize = value
	return nil
}

// CounterSpell removes a spell from the stack and routes its card to
// `dst` (defaulting to the spell's owner's graveyard). Implements
// the Counterspell / Hinder / Remand / Spell Crumple shape — the
// caller picks a destination (graveyard, hand, library, exile) at
// announce time.
//
// Battlefield is rejected as a destination: a counter that "puts the
// spell onto the battlefield" would be a different effect entirely
// (and there's no MTG card that does it the way a generic counter
// does). Stack is also rejected — the counter MUST move it off.
//
// Caller-gated to priority holder via the action layer; this method
// does not re-check that.
//
// Caller must NOT hold g.mu — this method takes the write lock.
//
// S13.1.
func (g *Game) CounterSpell(spellID uuid.UUID, dst *ZoneRef) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	// #529: one body, shared with CounterTargetForEffect. This used
	// to be a near-copy of counterSpellLocked that had drifted on
	// flashback handling; see the note there.
	return g.counterSpellLocked(spellID, dst)
}

// CounterAbility removes an activated / triggered ability from the
// stack. Abilities cease to exist on resolution (CR 608.2n); a
// counter is the same destinationless removal. Returns
// ErrCardNotOnStack if the ID doesn't reference an ability item.
//
// Caller must NOT hold g.mu — this method takes the write lock.
//
// S13.1.
func (g *Game) CounterAbility(abilityID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	item, ok := g.StackMeta[abilityID]
	if !ok || item == nil {
		return ErrCardNotOnStack
	}
	if item.Kind != StackItemActivated && item.Kind != StackItemTriggered {
		return ErrCardNotOnStack
	}
	source := item.SourceCardID
	delete(g.StackMeta, abilityID)
	g.recomputeSplitSecondLocked()
	g.EmitEvent(Event{
		Kind:   EventCounterSpell,
		Source: source,
		Target: abilityID,
	})
	return nil
}

// recomputeSplitSecondLocked walks StackMeta and pending triggers
// and refreshes the SplitSecondActive cache. Called after every
// stack mutation. Caller must hold g.mu.
func (g *Game) recomputeSplitSecondLocked() {
	for _, item := range g.StackMeta {
		if item != nil && item.SplitSecond {
			g.SplitSecondActive = true
			return
		}
	}
	for _, item := range g.PendingTriggers {
		if item != nil && item.SplitSecond {
			g.SplitSecondActive = true
			return
		}
	}
	g.SplitSecondActive = false
}

// MoveCardByID moves a card from one zone to another, identified by
// ZoneRef. This is the general-purpose zone mutation used by the
// move_card action; higher-level actions (DrawCard, PlayCard) wrap
// zone-specific MoveCard calls directly for clarity.
//
// A move where src resolves to the same Zone as dst is a no-op:
// MoveCard would Remove the card and re-Push it, and if src.Kind is
// battlefield it would also clear tapped state and counters — which
// would silently wipe a battlefield card's counters on a redundant
// client-issued no-op move. Detect the same-zone case up front.
//
// S13.1: when asCommander is true and the card is a commander
// being moved to graveyard, exile, hand, or library, the
// destination is rewritten to the owner's command zone (CR 903.9
// — commander zone replacement, exercised here as an explicit
// player choice rather than an automatic engine transform).
func (g *Game) MoveCardByID(src, dst ZoneRef, cardID uuid.UUID) error {
	return g.MoveCardByIDAsCommander(src, dst, cardID, false)
}

// MoveCardByIDAsCommander is the S13.1 extended form. asCommander
// is the player's "yes, route this commander back to command zone
// instead" choice that the move_card action exposes via a
// per-request flag. The default-false form preserves the historic
// MoveCardByID behaviour for non-commander moves.
func (g *Game) MoveCardByIDAsCommander(src, dst ZoneRef, cardID uuid.UUID, asCommander bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	srcZone := g.zoneFromRefLocked(src)
	if srcZone == nil {
		return ErrZoneNotFound
	}

	// S17 sub-PR 2: route through the replacement pipeline. The
	// commander-zone built-in (commanderZoneReplacement) gates on
	// card.IsCommander && an eligible destination — NOT on
	// ev.asCommanderMove, which stopped being a gate in #171; the
	// flag survives here only as a routing flavor on the manual
	// move_card action. When the built-in fires, ev.NewZone /
	// ev.NewZoneOwner are rewritten to the owner's command zone.
	//
	// This is one of two callers that predate the shared exit
	// primitive in zone_route.go. It keeps its own pipeline call
	// because it is the admin/sandbox move: it accepts an arbitrary
	// source AND destination including the battlefield, which the
	// exit primitive deliberately does not.
	ev := &ReplacementEvent{
		Kind:            RepEventMove,
		CardID:          cardID,
		OldZone:         srcZone.Kind,
		NewZone:         dst.Kind,
		NewZoneOwner:    dst.Owner,
		asCommanderMove: asCommander,
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// CR 616 prompt queued; resume path will re-enter.
		return nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		return nil
	}
	// Resolve the (possibly rewritten) destination.
	dst = ZoneRef{Kind: out.NewZone, Owner: out.NewZoneOwner}

	dstZone := g.zoneFromRefLocked(dst)
	if dstZone == nil {
		return ErrZoneNotFound
	}
	if srcZone == dstZone {
		// Verify the card is actually present so the caller still
		// sees ErrCardNotFound for a bogus instance ID.
		if !srcZone.Contains(cardID) {
			return ErrCardNotFound
		}
		return nil
	}
	if srcZone.Kind == ZoneBattlefield {
		g.snapshotLKILocked(cardID)
	}
	if _, err := MoveCard(srcZone, dstZone, cardID); err != nil {
		return err
	}
	g.markCardKnownInZoneLocked(dstZone, cardID)

	// S17 sub-PR 2: apply enters-tapped / enters-with-counters
	// from ev BEFORE EventETB fires so listeners + the client see
	// a consistent "entered with counters / tapped" state. Sub-PR 2
	// registers zero catalog replacements that set these, so both
	// branches are dead code paths today — they're here so sub-PR 4
	// can ship Kismet + Hangarback Walker without further plumbing.
	if dstZone.Kind == ZoneBattlefield {
		if out.EntersTapped {
			for i := range dstZone.Cards {
				if dstZone.Cards[i].InstanceID == cardID {
					dstZone.Cards[i].Tapped = true
					break
				}
			}
		}
		for name, n := range out.EntersWithCounters {
			_ = g.AddCounterForEffect(cardID, name, n)
		}
	}

	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		CardID:  cardID,
		OldZone: srcZone.Kind,
		NewZone: dstZone.Kind,
	})
	if srcZone.Kind == ZoneBattlefield {
		g.EmitEvent(Event{Kind: EventLTB, CardID: cardID, NewZone: dstZone.Kind})
	}
	if dstZone.Kind == ZoneBattlefield {
		g.EmitEvent(Event{Kind: EventETB, CardID: cardID})
		// ETB hook for admin direct-drop onto battlefield (and the
		// commander zone-replacement destination). Find the card in
		// the destination zone to pull its Scryfall ID.
		for _, c := range dstZone.Cards {
			if c.InstanceID == cardID {
				g.fireETBHookLocked(cardID, CatalogKey(c))
				break
			}
		}
	}
	// A sandbox move is a special action: the mover keeps priority
	// afterwards (CR 116.3), and CR 117.5 puts SBAs + the APNAP
	// trigger drain at that boundary. Without this, a catalog
	// creature dropped straight onto the battlefield would leave its
	// ETB trigger stranded in PendingTriggers until the next pass /
	// step — and an empty-stack wrap would advance the step first.
	g.runStateChecksLocked()
	return nil
}

// TapCard sets the tapped state of a card on the battlefield. Returns
// ErrCardNotFound if the card is not currently on the battlefield —
// tapping a card in any other zone is meaningless.
//
// Shares setTapStateLocked with the TapTarget / UntapTarget effect
// primitives, which is what makes a sandbox untap fire "whenever a
// permanent becomes untapped" the same way an effect's does.
func (g *Game) TapCard(cardID uuid.UUID, tapped bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	// The summoning-sickness gate below only binds creatures, and
	// whether this permanent is one is a layer answer. Fast-path
	// no-op when nothing changed.
	g.RecomputeLayersIfStaleLocked()
	return g.setTapStateLocked(cardID, tapped)
}

// ActivateManaAbility fires the `abilityIdx`-th mana ability on a
// battlefield permanent. Single code path for both catalog-declared
// abilities (Sol Ring, Arcane Signet, Birds of Paradise — looked up
// via the CatalogManaAbilities hook) and synthetic basic-land
// abilities (Forest → "{G}") the engine derives from TypeLine when
// the catalog has nothing to say.
//
// Validates the card is on the battlefield, controlled by playerID,
// and — when the ability has a tap cost — currently untapped. Taps
// the card as part of the cost. Then parses Produced into one slot
// per brace: single-color slots drop straight into the controller's
// pool as one ManaToken; multi-option slots (pipe syntax, "{W|U|B|R|G}")
// queue a PendingChoiceMana for the controller to pick from.
//
// Mana abilities don't use the stack (CR 605.3) — everything is
// synchronous under the write lock. Event log gets
// EventManaAbilityActivated + EventTapCard (if tap cost) + one
// EventManaAdded per single-color slot.
//
// Returns an error when the card isn't on the battlefield
// (ErrCardNotFound), the caller doesn't control it (ErrNotController),
// the tap cost can't be paid because the card is already tapped
// (ErrAlreadyTapped), or the ability index is out of bounds
// (ErrInvalidParam). Commander-identity filtering for Arcane Signet
// happens inline: the engine intersects the pipe set with the
// controller's commander's color identity before queuing the pick —
// unless the ability set IgnoreCommanderIdentity, which City of Brass
// and the painland duals do because their printed text names their
// colours outright.
//
// S22 adds the last two pieces the painland / Talisman / Ancient Tomb
// / Mana Confluence batch needed:
//
//   - ManaAbilityShape.LifeCost, a "Pay N life" cost component,
//     validated alongside the tap and sacrifice components before any
//     of them is paid (so a rejected activation never leaves the
//     source tapped) and paid between the tap and the sacrifice;
//   - ManaAbilityShape.Rider, everything the oracle text says after
//     the "Add …" clause, run once the produced mana is in the pool.
//
// Any activation that pays life, sacrifices, or fires a rider runs a
// state-based-action pass on the way out, so a player who taps
// Ancient Tomb at 2 life loses here rather than at the next priority
// boundary.
//
// Added in S15 sub-PR 2.
// ManaAbilityParams carries the choices a mana ability's cost needs
// from the activator. Empty for the common case — a bare "{T}: Add
// {C}" needs nothing.
//
// Mana abilities don't use the stack (CR 605.3b), so unlike
// ActivateAbilityParams there is no target list here: a mana ability
// that targeted would have to resolve, and none does.
type ManaAbilityParams struct {
	// SacrificeIDs names the permanents paying a SacrificeOther
	// component. Exactly one for a single-permanent clause; empty
	// when the ability has no such cost.
	SacrificeIDs []uuid.UUID
}

func (g *Game) ActivateManaAbility(playerID, cardID uuid.UUID, abilityIdx int, params ManaAbilityParams) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	// The ability list this resolves an index against is partly
	// type-derived (CR 305.6), so it moves when a Layer-4 static
	// does: under Urborg every land gains "{T}: Add {B}" and loses
	// it again when Urborg does. Catch the engine up before
	// indexing or a player's {B} click lands on a stale list.
	// Fast-path no-op when nothing has changed.
	g.RecomputeLayersIfStaleLocked()
	// Locate the card on the battlefield.
	var card *Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			card = &g.Battlefield.Cards[i]
			break
		}
	}
	if card == nil {
		return ErrCardNotFound
	}
	if card.Controller != playerID {
		return ErrCardCallerMismatch
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	// Resolve the ability shape — catalog first, synthetic basic-land
	// fallback second. A catalog spec with ManaAbilities overrides
	// the synthetic path wholesale (Dryad Arbor, if it ever lands,
	// would declare its own; basic Forest just uses the synthetic).
	// CR 602.5: an effect that stops this permanent's activated
	// abilities being activated stops its mana abilities too, when
	// it says so. Arrest does; Faith's Fetters explicitly does not
	// ("unless they're mana abilities"), which is why the two are
	// separate bits — see restrictions.go.
	if !CanActivateManaAbilities(card) {
		return ErrCantActivate
	}
	abilities := ManaAbilitiesForCard(*card)
	if abilityIdx < 0 || abilityIdx >= len(abilities) {
		return ErrInvalidParam
	}
	ab := abilities[abilityIdx]
	// --- gate ----------------------------------------------------
	//
	// "Activate only if you control five or more lands" (Temple of
	// the False God), "…three or more artifacts" (Mox Opal). CR
	// 602.5: an activation restriction is checked before anything
	// is paid, so a failed gate costs the player nothing.
	if ab.Condition != nil && !ab.Condition(g, playerID, cardID) {
		return ErrConditionNotMet
	}
	// --- validate every cost before paying any ------------------
	//
	// Same discipline as ActivateCatalogAbility: a half-paid cost
	// must never strand the permanent. Ashnod's Altar with a tapped
	// source has to fail WITHOUT eating the creature.
	if ab.TapCost {
		if card.Tapped {
			return ErrAlreadyTapped
		}
		// CR 302.6: a creature's {T} ability needs it to have been
		// under your control since your most recent turn began.
		// Birds of Paradise, Llanowar Merfolk and Palladium Myr all
		// live here; before this check they tapped for mana the turn
		// they landed, which is simply wrong. Non-creature sources
		// (Sol Ring, a Treasure, a land) are never sick, and haste
		// exempts a creature — since #530 HasSummoningSickness
		// really does handle both, and the IsCreature prefix here is
		// redundant. It stays as documentation: this is the rule
		// about creatures.
		if card.IsCreature() && HasSummoningSickness(card) {
			return ErrSummoningSick
		}
	}
	sacrifices, err := g.validateSacrificeCostLocked(playerID, cardID, AbilityCost{
		SacrificeSelf:  ab.SacrificeCost,
		SacrificeOther: ab.SacrificeOther,
	}, params.SacrificeIDs)
	if err != nil {
		return err
	}
	if ab.LifeCost > 0 && p.Life < ab.LifeCost {
		// CR 119.4: you can't pay more life than you have. Paying
		// down to exactly 0 is legal and the SBA loop ends the game
		// after. Same gate ActivateCatalogAbility applies to
		// AbilityCost.Life.
		return ErrInvalidParam
	}
	// A mana component in the cost — the Signet cycle's "{1}, {T}",
	// Cabal Coffers' "{2}, {T}". Parsed and checked here, spent
	// below with everything else, so an unaffordable Signet fails
	// with the source still untapped.
	//
	// Deliberately no auto-tap. A mana ability resolves with no
	// priority window (CR 605.3b), and tapping three lands to feed a
	// Signet is a decision with consequences the planner cannot
	// weigh — the player floats the {1} first, which is how the card
	// is played on paper anyway.
	var manaCost ParsedCost
	if ab.ManaCost != "" {
		parsed, perr := ParseCost(ab.ManaCost)
		if perr != nil {
			return ErrInvalidParam
		}
		manaCost = parsed
		// The mana pays an ACTIVATION (CR 602.2b), and the source
		// permanent's own characteristics are what a restricted
		// token is tested against — Eldrazi Temple mana can fund a
		// colorless Eldrazi's ability, not a Signet's.
		spendCtx := ManaSpendForAbility(*card)
		if !p.ManaPool.CanPayFor(manaCost, 0, spendCtx) {
			return &InsufficientManaError{Missing: p.ManaPool.MissingFor(manaCost, 0, spendCtx)}
		}
	}

	// needStateChecks is set by any component of this activation
	// that can kill a player or a permanent — a sacrifice, a life
	// payment, a damage rider. A single deferred pass covers all of
	// them, so a painland activated at 1 life loses the game on the
	// way out of this call rather than at some later boundary.
	needStateChecks := false

	// --- pay ----------------------------------------------------
	//
	// Mana first, then tap, then life, then sacrifice — the
	// component order ActivateCatalogAbility pays an AbilityCost in.
	// Tap before sacrifice: the tap has to happen while the
	// permanent is still on the battlefield, and sacrifices move
	// cards, which invalidates `card`.
	if ab.ManaCost != "" {
		// Same context the validation above used. Spending under a
		// different context than the check would let a restricted
		// token pay for something it was never cleared for.
		if !p.ManaPool.SpendManaFor(manaCost, 0, ManaSpendForAbility(*card)) {
			return &InsufficientManaError{Missing: []string{ab.ManaCost}}
		}
		g.EmitEvent(Event{Kind: EventManaSpent, Actor: playerID, Source: cardID})
	}
	if ab.TapCost {
		card.Tapped = true
		g.EmitEvent(Event{Kind: EventTapCard, Actor: playerID, CardID: cardID})
	}
	// Life after tap, before sacrifice — the same component order
	// ActivateCatalogAbility pays an AbilityCost in (mana → tap →
	// life → sacrifice). It matters only for the event log, since
	// every component was validated above.
	//
	// #793: the cost path. A mana ability resolves immediately and
	// without the stack (CR 605.3b), so Mana Confluence's life cannot
	// be left waiting on a CR 616 prompt with the mana already in the
	// pool. The CR 614 window still runs — paying life is losing life
	// (CR 119.4) — it just settles in one step.
	if ab.LifeCost > 0 {
		if err := g.PayLifeForEffect(cardID, playerID, ab.LifeCost); err != nil {
			return err
		}
		needStateChecks = true
	}
	// S21 sub-PR 1 made sacrifice-self costs real (Treasure, Eldrazi
	// Spawn, Lotus Petal); the mana-cost pass adds sacrifice-another
	// (Ashnod's Altar, Phyrexian Altar). The mana still lands in the
	// pool below — CR 605.3b: a mana ability resolves immediately,
	// without the stack, so the sacrifices and the mana are one
	// atomic step.
	if len(sacrifices) > 0 {
		// One payment, one simultaneous exit (#747): a Blood Artist
		// paid in alongside another creature sees both deaths.
		if err := g.payCostSacrificesLocked(sacrifices); err != nil {
			return err
		}
		// A sacrificed source leaves `card` dangling. Nothing below
		// touches it (the produced-mana path reads `ab`).
		card = nil
		// The sacrifice can queue dies- / sacrifice-triggers (Blood
		// Artist feeding off an Altar). Drain them on the way out, so
		// the mana is already in the pool when they resolve — which
		// is what makes an Altar plus a payoff a real engine.
		needStateChecks = true
	}
	defer func() {
		if needStateChecks {
			g.runStateChecksLocked()
		}
	}()
	g.EmitEvent(Event{
		Kind:   EventManaAbilityActivated,
		Actor:  playerID,
		Source: cardID,
	})
	// Materialise the produced mana. An ability with a ProducedFunc
	// computes its output now, from the board as it stands AFTER the
	// cost was paid — which is what CR 605.3b's "resolves
	// immediately" means, and what makes Cabal Coffers count the
	// Swamps that are still there.
	produced := ab.Produced
	if ab.ProducedFunc != nil {
		produced = ab.ProducedFunc(g, playerID, cardID)
	}
	slots, err := ParseProducedMana(produced)
	_ = card // card is deliberately nil after a sacrifice; keep the intent explicit
	if err != nil {
		// Mal-formed produced string: emit an effect-error event and
		// stop short of adding mana. The cost has already been paid —
		// the player just didn't get anything for it, which is correct
		// for a broken ability declaration.
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Actor:    playerID,
			Source:   cardID,
			ErrorMsg: err.Error(),
		})
		return nil
	}
	for _, slot := range slots {
		options := slot.Options
		if len(options) == 0 {
			continue
		}
		if len(options) == 1 {
			// Single-color slot — straight into the pool, carrying
			// the ability's spend restrictions (Eldrazi Temple's
			// "colorless Eldrazi only"). Copy the slice: the token
			// outlives this call and clone.go deep-copies it, so
			// aliasing the catalog's backing array would let an undo
			// reach a shared one.
			p.ManaPool.AddMana(ManaToken{
				Color:        options[0],
				Source:       cardID,
				Restrictions: restrictionsFor(g, &ab, playerID, cardID),
			})
			g.EmitEvent(Event{Kind: EventManaAdded, Actor: playerID, Source: cardID})
			continue
		}
		// Multi-option slot — intersect with the controller's commander
		// identity (Arcane Signet). When no commander exists or the
		// identity is empty (placeholder commanders from the demo seed),
		// fall through to the raw option set so Birds of Paradise still
		// offers the full five colors.
		//
		// S22: an ability whose printed text does NOT mention the
		// commander's identity opts out — City of Brass says "any
		// color", a painland names two specific colors, and neither
		// should shrink because of who's in the command zone.
		filtered := options
		if !ab.IgnoreCommanderIdentity {
			filtered = filterPipeByCommanderIdentity(options, p)
		}
		// Queue the pick. The restrictions ride ON THE CHOICE, not
		// just on the ability: the token is minted later, in
		// ResolveManaChoice, and without this a Delighted Halfling
		// pick would land in the pool unrestricted — the #259
		// direction, and the easiest place in this whole seam to
		// leak it.
		g.QueueChoiceForEffect(PendingChoice{
			Kind:             PendingChoiceMana,
			Chooser:          playerID,
			FromPlayer:       playerID,
			Count:            1,
			Source:           cardID,
			Reason:           ab.Label,
			ColorOptions:     filtered,
			ManaRestrictions: restrictionsFor(g, &ab, playerID, cardID),
			// #742: "N mana of any one color" — one pick, N tokens.
			ManaAmounts: copyManaAmounts(slot.Amounts),
		})
	}
	// --- rider --------------------------------------------------
	//
	// CR 605.3b: a mana ability resolves the instant it's activated,
	// so everything after the "Add …" clause — the painland cycle's
	// "This land deals 1 damage to you", Ancient Tomb's 2 — happens
	// here, with the mana already in the pool and no priority window
	// in between. A pipe slot that queued a PendingChoiceMana above
	// is no exception: the colour is still unpicked, but the damage
	// is not waiting on it.
	//
	// The rider runs even when the produced mana went nowhere, which
	// is what the printed cards say: "This land deals 1 damage to
	// you" is not conditional on the mana being useful.
	if ab.Rider != nil {
		needStateChecks = true
		if err := ab.Rider(g, playerID, cardID); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				Actor:    playerID,
				Source:   cardID,
				ErrorMsg: err.Error(),
			})
		}
	}
	return nil
}

// ManaAbilitiesForCard returns the abilities available on a card:
// the ones carried on the instance, else the catalog's, PLUS the
// CR 305.6 intrinsic abilities the card's effective land types give
// it for free. The protocol layer calls this to stamp
// CardView.ManaAbilities; the dispatcher calls it to resolve an
// incoming ability index.
//
// Caller must hold g.mu, and — for the intrinsic half to be
// current — must have let the layer engine catch up first
// (ReadSnapshot does; write paths call RecomputeLayersIfStaleLocked).
//
// The intrinsic half is where Urborg, Tomb of Yawgmoth becomes a
// real card. #258 declined to catalogue it because its Layer-4
// static "would apply cleanly and do nothing" — the synthetic mana
// ability read the printed TypeLine, so a Mountain that the layer
// engine had made a Swamp still only tapped for {R}.
func ManaAbilitiesForCard(c Card) []ManaAbilityShape {
	var declared []ManaAbilityShape
	switch {
	// S24 layer 6: an ability-removing effect takes the DECLARED
	// half and leaves the INTRINSIC half, and the split is the whole
	// of CR 305.7. "Enchanted permanent is a colorless Forest land"
	// removes the abilities the permanent's rules text generated —
	// Sol Ring's "{T}: Add {C}{C}", a Signet's filter — and grants
	// the mana ability that comes with the new land type. That
	// second half is not printed on the card and is not in the
	// catalog: it is derived from the effective subtypes, below, by
	// the same effect that did the removing, so it survives on the
	// other side of this switch rather than being re-granted.
	case c.HasLostAllAbilities():
		declared = nil
	// S21 sub-PR 1: instance abilities win — a token has no oracle
	// ID for the catalog to key on.
	case len(c.ManaAbilities) > 0:
		declared = c.ManaAbilities
	// CatalogAbilityKey, not c.OracleID: an MDFC back face keys on
	// "<oracle_id>#N" (#357). A bare OracleID here would silently
	// resolve a back face to face 0's spec.
	case CatalogManaAbilities != nil:
		declared = CatalogManaAbilities(CatalogAbilityKey(c))
	}
	intrinsic := intrinsicLandManaAbilities(c)
	if len(intrinsic) == 0 {
		return declared
	}
	if len(declared) == 0 {
		return intrinsic
	}
	// A declared ability and an intrinsic one can name the same
	// colour — every catalog dual land ("{T}: Add {B} or {G}") is
	// printed with the land types that would have produced the
	// same mana, and doubling it up would put two ways to make {B}
	// in the client's ability row. Keep the declared shape (it may
	// carry a rider or a pipe) and add only colours it cannot make.
	covered := producibleColors(declared)
	out := make([]ManaAbilityShape, 0, len(declared)+len(intrinsic))
	out = append(out, declared...)
	for _, ab := range intrinsic {
		if covered[landTypeColorOf(ab)] {
			continue
		}
		out = append(out, ab)
	}
	return out
}

// landTypeMana lists the five basic land types (CR 305.6) with the
// mana their intrinsic ability produces. Order is only a tie-break
// for cards whose effective subtypes are unordered; the real
// ordering comes from the subtype list itself.
var landTypeMana = [...]struct{ Subtype, Color string }{
	{"Plains", "W"},
	{"Island", "U"},
	{"Swamp", "B"},
	{"Mountain", "R"},
	{"Forest", "G"},
}

// intrinsicLandManaAbilities builds the CR 305.6 abilities a land's
// EFFECTIVE subtypes grant it: "{T}: Add {W}" for Plains, "{T}: Add
// {U}" for Island, and so on, one per distinct basic land type.
//
// Two things changed here versus the S15 basicLandColor it
// replaces, and they are the same change seen twice:
//
//   - It reads effective subtypes, so a Layer-4 type-granting
//     static (Urborg) is load-bearing rather than cosmetic.
//   - It keys off the basic land TYPE, not the Basic SUPERTYPE.
//     CR 305.6 has never mentioned the supertype; requiring it was
//     an S15 shortcut that left every printed dual — Bayou,
//     Overgrown Tomb, a Triome — producing nothing at all unless
//     someone hand-wrote a catalog entry for it. battle_lands.go
//     documents that shortcut as the reason its cycle declares a
//     pipe ability the printed card puts in reminder text.
//
// Emitted in the order the subtypes appear, so a land with no
// static on it keeps the exact ability list (and therefore the
// exact ability indices, and the exact auto-tapper first choice)
// it had before.
func intrinsicLandManaAbilities(c Card) []ManaAbilityShape {
	if !c.IsLand() {
		return nil
	}
	var subtypes []string
	if c.effective != nil {
		subtypes = c.effective.Subtypes
	} else {
		_, _, subtypes = ParseTypeLine(c.TypeLine)
	}
	if len(subtypes) == 0 {
		return nil
	}
	var out []ManaAbilityShape
	// One bit per entry in landTypeMana rather than a map: this runs
	// once per land per card view, and five lands is already a
	// thousand map allocations a minute at snapshot rates.
	var seen uint8
	for _, sub := range subtypes {
		for i, lt := range landTypeMana {
			if !equalFoldASCII(sub, lt.Subtype) || seen&(1<<i) != 0 {
				continue
			}
			seen |= 1 << i
			out = append(out, ManaAbilityShape{
				TapCost:  true,
				Produced: "{" + lt.Color + "}",
				Label:    "Add {" + lt.Color + "}",
			})
		}
	}
	return out
}

// landTypeColorOf returns the single colour letter an intrinsic
// land ability produces. Only ever called on shapes this file
// built, so the "{X}" form is guaranteed.
func landTypeColorOf(ab ManaAbilityShape) string {
	if len(ab.Produced) != 3 {
		return ""
	}
	return ab.Produced[1:2]
}

// producibleColors is the set of colour letters a declared ability
// list can make, pipes included — City of Brass's "{W|U|B|R|G}"
// covers all five. Unparseable declarations contribute nothing
// rather than failing the merge.
func producibleColors(abilities []ManaAbilityShape) map[string]bool {
	out := map[string]bool{}
	for _, ab := range abilities {
		slots, err := ParseProducedMana(ab.Produced)
		if err != nil {
			continue
		}
		for _, slot := range slots {
			for _, opt := range slot.Options {
				out[opt] = true
			}
		}
	}
	return out
}

// filterPipeByCommanderIdentity narrows `options` to just the colors
// in the active player's commander's color identity. Used by the
// `"{W|U|B|R|G}"` → Arcane Signet path. If the player has no
// commander (or the commander has no color identity), returns the
// raw options unchanged — Birds of Paradise falls through this path
// with the full 5-color set intact.
func filterPipeByCommanderIdentity(options []string, p *Player) []string {
	// Find the player's commander in the command zone. Since #276
	// the colour identity really is on game.Card directly, copied
	// at deck-import time from cards.Card.ColorIdentity — this
	// comment described that code for several sprints before it
	// existed, which is why the gap went unnoticed.
	// For S15 we scan the command zone only — post-move commanders
	// on the battlefield still carry identity, but CR 903.4 keys
	// identity off the printed card, so either source would work.
	identity := commanderIdentityFor(p)
	if len(identity) == 0 {
		return options
	}
	keep := make([]string, 0, len(options))
	for _, o := range options {
		for _, id := range identity {
			if o == id {
				keep = append(keep, o)
				break
			}
		}
	}
	if len(keep) == 0 {
		// No overlap — degenerate case; fall back to raw so the
		// player isn't stuck with an empty picker. Logged via the
		// effect-error event for visibility.
		return options
	}
	return keep
}

// commanderIdentityFor returns the player's commander color identity
// as uppercase single-character strings. Empty when no commander is
// present. Sandbox: picks the first card in the command zone — the
// partner-pair union is S20 polish territory.
//
// Three sources, in order:
//
//  1. **Card.ColorIdentity** — Scryfall's own `color_identity`,
//     copied at deck import. This is the only one of the three that
//     is actually CR 903.4: it folds in mana symbols in rules text
//     and BOTH faces of a double-faced card. Issue #276: a
//     transform / modal-DFC commander has a null top-level
//     mana_cost and colors (the real ones live on card_faces[0]),
//     so sources 2 and 3 both returned empty and every "any colour
//     in your commander's identity" pipe skipped narrowing —
//     Command Tower offered all five colours to an Azorius deck.
//  2. **Effective().Colors** (S16) — the commander's post-layer
//     colours. Not identity, but a good proxy for any commander
//     whose identity is fully captured by their mana cost, and it
//     picks up Layer-5 colour-change effects (Painter's Servant on
//     a commander) automatically.
//  3. **distinctColorsInManaCost** — the original S15 proxy.
//     Covers placeholder commanders with no stamped colours (the
//     demo seed) and the lobby-time pre-effects-init path.
//
// Note the fallbacks can only ever be narrower than the truth, and
// every call site uses the result to NARROW a mana pipe, so the
// pre-#276 behaviour was permissive (offering colours that don't
// exist) rather than restrictive.
func commanderIdentityFor(p *Player) []string {
	if p == nil || p.Command == nil {
		return nil
	}
	for _, c := range p.Command.Cards {
		if !c.IsCommander {
			continue
		}
		if len(c.ColorIdentity) > 0 {
			return c.ColorIdentity
		}
		eff := c.Effective()
		if len(eff.Colors) > 0 {
			return eff.Colors
		}
		return distinctColorsInManaCost(c.ManaCost)
	}
	return nil
}

// distinctColorsInManaCost extracts the unique WUBRG letters that
// appear in a mana-cost string. Used as the S15 proxy for the
// commander's color identity. "{1}{W}{U}" → ["W", "U"].
func distinctColorsInManaCost(cost string) []string {
	seen := map[string]bool{}
	out := []string{}
	for i := 0; i < len(cost); i++ {
		b := cost[i]
		if b >= 'a' && b <= 'z' {
			b -= 'a' - 'A'
		}
		if b == 'W' || b == 'U' || b == 'B' || b == 'R' || b == 'G' {
			k := string(b)
			if !seen[k] {
				seen[k] = true
				out = append(out, k)
			}
		}
	}
	return out
}

// SetBattlefieldPosition stamps a normalised (x, y) position on a
// card on the battlefield. x and y are clamped to [0, 1] — the client
// sends fractions of the battlefield area so the server's stored
// coordinate survives resolution changes on the rendering side.
// Returns ErrCardNotFound when the card is not on the battlefield;
// positions on cards in other zones are meaningless and ignored.
func (g *Game) SetBattlefieldPosition(cardID uuid.UUID, x, y float64) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	x = clampUnit(x)
	y = clampUnit(y)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			g.Battlefield.Cards[i].BattleX = x
			g.Battlefield.Cards[i].BattleY = y
			return nil
		}
	}
	return ErrCardNotFound
}

func clampUnit(v float64) float64 {
	// `!(v >= 0)` intentionally treats NaN the same as a negative
	// input: the comparison is false for NaN, and `!false` pulls us
	// into the zero branch. Storing NaN would poison every subsequent
	// snapshot — Go's encoding/json errors on NaN rather than
	// serialising it, so a single bad write would break broadcasts
	// for every viewer.
	if !(v >= 0) {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// UntapAll untaps every card on the battlefield controlled by the
// given player. Pre-S13 this was the manual untap step; S13's
// StepUntap auto-action now drives normal play. The action remains
// dispatchable (sandbox / replay support) but is gated as a no-op
// during the active player's StepUntap to avoid double-firing on top
// of the auto-action.
//
// The untap itself lives in untap.go — see untapAllForLocked, and
// performUntapStepLocked for the turn-based action this is NOT.
func (g *Game) UntapAll(playerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.playerByIDLocked(playerID) == nil {
		return ErrPlayerNotFound
	}
	if g.Turn.Step == StepUntap && g.activeSeatIDLocked() == playerID {
		return nil
	}
	g.untapAllForLocked(seatOfPlayerLocked(g, playerID))
	return nil
}

// seatOfPlayerLocked returns the seat index of the player with the
// given ID, or -1 if not seated. Caller must hold g.mu.
func seatOfPlayerLocked(g *Game, id uuid.UUID) int {
	for i, p := range g.Seats {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// PassPriority rotates priority to the next non-eliminated seat. When
// priority would pass back to the active seat (every other live seat
// has passed in succession with nothing on the stack), the step auto-
// advances and PriorityHolder is reset to the new ActiveSeat —
// mirroring real MTG rules where priority passing around in
// succession ends the step.
//
// S13: returns ErrNoPriority during the Untap and Cleanup steps,
// which don't grant priority (CR 502.4 / 514.3). Eliminated seats are
// skipped during the rotation so a 4-player game with one dead seat
// still terminates the priority loop on the survivors' wrap.
//
// Stack-aware semantics (priority resets to active seat whenever a
// spell resolves) are deferred to the S13+ rules graft.
func (g *Game) PassPriority() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	// #730: an unanswered prompt gates the table. Every pass is
	// refused, not only the one that would wrap into a step advance —
	// a pass that resolves the top of the stack moves the game on
	// just as surely, and the enumerator has always modelled the rule
	// this way. See choice_gate.go.
	if c := g.blockingChoiceLocked(); c != nil {
		return choicePendingErrorLocked(c)
	}
	numSeats := len(g.Seats)
	if numSeats == 0 {
		return ErrGameNotActive
	}
	if g.Turn.PriorityHolder == NoPriority {
		return ErrNoPriority
	}
	// Walk forward to the next non-eliminated seat. Bounded by
	// numSeats iterations so a fully-eliminated table can't infinite-
	// loop (the surrounding game-end check in Concede flips State to
	// StateEnded in that case; the early ErrGameNotActive guard above
	// catches subsequent calls).
	next := g.Turn.PriorityHolder
	for i := 0; i < numSeats; i++ {
		next = (next + 1) % numSeats
		if next == g.Turn.ActiveSeat {
			break
		}
		if !g.Seats[next].Eliminated {
			g.Turn.PriorityHolder = next
			return nil
		}
	}
	// Wrapped (or only the active seat is alive). Two cases:
	//
	// S13.1: if the stack has any items (spells in Game.Stack OR
	// abilities in StackMeta), resolve the top one and reset
	// priority to the new active seat — the post-resolution priority
	// boundary per CR 117.3b / 608.1. The step does NOT advance; the
	// engine returns to the priority loop at the same step until the
	// stack drains.
	//
	// Empty stack: advance the step (existing behavior). Turn.advance
	// resets PriorityHolder to NoPriority (Untap/Cleanup) or the new
	// ActiveSeat. Run the same per-step entry hooks AdvanceStep does
	// so auto turn-based actions, combat damage resolution, and
	// combat-clear all fire regardless of whether the step changed
	// via a priority-wrap or an explicit advance_step click.
	if g.stackHasItemsLocked() {
		if err := g.resolveTopOfStackLocked(); err != nil {
			return err
		}
		// Drain any pending APNAP triggers onto the stack now that
		// we've crossed a priority-grant boundary (CR 603.3b).
		g.runStateChecksLocked()
		// Priority returns to the active player after a resolution
		// (CR 117.3b). The step doesn't change.
		g.Turn.PriorityHolder = g.Turn.ActiveSeat
		return nil
	}
	g.advanceCursorLocked()
	g.runStepEntryHooksLocked()
	g.drainPendingTriggersAPNAPLocked()
	return nil
}

// stackHasItemsLocked reports whether anything (a spell card or an
// ability item in StackMeta) is currently on the stack. Used by the
// priority-wrap branch to decide between "resolve top" and "advance
// step." Caller must hold g.mu.
func (g *Game) stackHasItemsLocked() bool {
	if g.Stack != nil && len(g.Stack.Cards) > 0 {
		return true
	}
	for _, item := range g.StackMeta {
		if item != nil {
			return true
		}
	}
	return false
}

// DeclareAttacker marks a battlefield card as attacking the target
// player. Gated by the declare_attackers step (MTG: attackers can
// only be declared during the corresponding step) and by the
// attacker being a creature (MTG: only creatures attack).
//
// Returns ErrWrongStep outside the declare_attackers step,
// ErrNotACreature for non-creature cards, ErrPlayerNotFound for an
// unknown target player, ErrCardNotFound for an unknown attacker
// card, ErrSummoningSick for a creature that entered this turn
// without haste (CR 302.6, 702.10), and ErrDefender for a defender
// creature (CR 702.3). Re-declaring the same attacker against a
// different target overwrites the previous target.
//
// Caller authorization (was-it-the-controller) is intentionally
// NOT enforced — sandbox flexibility for casual play. Summoning
// sickness and defender are enforced as of S18.
//
// S22: a creature's FIRST successful declaration emits EventAttack
// (one per attacking creature — see events.go) and then runs the
// state checks, so "whenever ~ attacks" triggers reach the stack
// immediately, inside the declare-attackers step and ahead of
// blockers. Re-declaring an already-attacking creature against a
// different target still overwrites the target but emits nothing.
func (g *Game) DeclareAttacker(attackerID, targetPlayerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.Turn.Step != StepDeclareAttackers {
		return ErrWrongStep
	}
	// Layers must be fresh so HasKeyword reads the current effective
	// characteristic (e.g. a creature granted haste via a Lightning
	// Greaves equip this turn should be attackable) — and, since S27,
	// so the attack-target classifier reads effective TYPES: whether
	// a permanent is a planeswalker right now is a layer question.
	g.RecomputeLayersIfStaleLocked()
	// S27: the target may be a player, a planeswalker or a battle
	// (CR 506.2, 508.1d). The parameter keeps its historical name;
	// its DOMAIN is what widened. Classification happens before the
	// attacker lookup so an unknown target is rejected the same way
	// whichever creature was named.
	if g.classifyAttackTargetLocked(targetPlayerID) == AttackTargetNone {
		// A seat id that resolves to nothing is still
		// ErrPlayerNotFound, which is what every existing caller and
		// test expects; anything else is the new error.
		return ErrPlayerNotFound
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == attackerID {
			card := &g.Battlefield.Cards[i]
			if !card.IsCreature() {
				return ErrNotACreature
			}
			if err := g.canAttackTargetLocked(card.Controller, targetPlayerID); err != nil {
				return err
			}
			if HasKeyword(card, "defender") {
				return ErrDefender
			}
			if HasSummoningSickness(card) {
				return ErrSummoningSick
			}
			// CR 508.1c: a creature under a "can't attack"
			// restriction may not be declared, and unlike the
			// tapped / already-declared checks this one is NOT
			// relaxed for the sandbox's hand-forcing. An effect that
			// says a creature can't attack is the card doing its
			// whole job; a verb that shrugged it off would make
			// Pacifism a blank. See restrictions.go.
			if !CanAttack(card) {
				return ErrCantAttack
			}
			// S22: a creature is declared as an attacker once (CR
			// 508.1). The sandbox additionally lets a player re-point
			// an already-attacking creature at a different defender;
			// that is a correction, not a second attack, so it must
			// not fire attack triggers again.
			firstDeclaration := card.AttackingTarget == uuid.Nil
			controller := card.Controller
			card.AttackingTarget = targetPlayerID
			// CR 508.1f: declaring an attacker taps it, unless the
			// attacker has vigilance (CR 702.20). Vigilance is the
			// one keyword whose job is specifically to skip this
			// tap — without it, the creature is tapped as a cost of
			// attacking.
			if !HasKeyword(card, "vigilance") {
				card.Tapped = true
			}
			// A card declared as attacker can't simultaneously be a
			// blocker — clearing the other field keeps the per-card
			// combat state coherent.
			card.BlockingTarget = uuid.Nil
			if !firstDeclaration {
				return nil
			}
			// S22: "whenever ~ attacks" triggers. One event per
			// attacking creature; `card` must not be read past this
			// point, because the state checks below can reallocate the
			// battlefield slice out from under the pointer.
			g.EmitEvent(Event{
				Kind:   EventAttack,
				Actor:  controller,
				CardID: attackerID,
				Target: targetPlayerID,
			})
			// CR 508.2 / 117.5: attackers are declared as a turn-based
			// action, after which the active player receives priority —
			// the boundary where SBAs run and harvested triggers go on
			// the stack. Without this drain the trigger would sit in
			// PendingTriggers until the next pass and land a step late,
			// after blockers. Same reasoning as the MoveCardByID /
			// CastSpell drains (ADR 0018 §3).
			g.runStateChecksLocked()
			return nil
		}
	}
	return ErrCardNotFound
}

// AttackDeclaration is one (attacker, defending player) pair inside a
// bulk DeclareAttackers submission. The set as a whole is the
// attacking player's declaration for the turn (CR 508.1).
type AttackDeclaration struct {
	Attacker uuid.UUID
	Target   uuid.UUID
}

// DeclareAttackers declares an entire attacking set in ONE mutation.
// It exists because the per-creature DeclareAttacker is the wrong
// granularity for a wide board: the client's "attack with all" only
// has the per-creature verb, so a 12-creature alpha strike becomes 12
// room.Apply calls — 12 full-game clones on the undo stack, 12
// snapshot broadcasts, and an "undo" that needs 12 presses against a
// per-turn budget of one. One action means one undo entry, so undo is
// the exact inverse of "attack with everything" (#318).
//
// It is also closer to the rules than the loop is. CR 508.1 declares
// attackers simultaneously and CR 508.2 gives the active player
// priority afterwards, at which point every "whenever ~ attacks"
// trigger goes on the stack together in APNAP order. Declaring one at
// a time instead fires each trigger through its own state-check drain,
// interleaving resolutions between declarations. Here every card is
// mutated first, then every EventAttack is emitted, then a single
// runStateChecksLocked drains the batch.
//
// Eligibility is STRICT and silent, unlike DeclareAttacker, which is
// deliberately lax so the sandbox can force odd board states by hand.
// A bulk "attack with everything" must never turn one ineligible
// creature into a failed alpha strike, so entries that are unknown,
// not creatures, tapped, summoning-sick (CR 302.6), defenders
// (CR 702.3), under a "can't attack" restriction (CR 508.1c),
// already declared this combat, or pointed at a nonexistent /
// eliminated / self seat are skipped without error.
// Callers that need the lax behaviour keep using DeclareAttacker.
//
// Returns the instance IDs actually declared, in submission order.
// Returns ErrNoLegalAttackers when the whole batch was skipped, so
// the room layer neither records an undo entry nor broadcasts a
// snapshot for a no-op. Caller authorization (does this seat control
// these creatures) is enforced one layer up, in actions.Dispatch.
func (g *Game) DeclareAttackers(decls []AttackDeclaration) ([]uuid.UUID, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return nil, ErrGameNotActive
	}
	if g.Turn.Step != StepDeclareAttackers {
		return nil, ErrWrongStep
	}
	// Layers must be fresh so HasKeyword reads current effective
	// characteristics — a creature handed haste or vigilance this turn
	// has to be judged on the granted keyword, not the printed one.
	g.RecomputeLayersIfStaleLocked()

	// attackEvent captures what EmitEvent needs by value. Card
	// pointers must not survive the mutation loop: emitting events
	// and running state checks can reallocate the battlefield slice.
	type attackEvent struct {
		attacker   uuid.UUID
		controller uuid.UUID
		target     uuid.UUID
	}
	declared := make([]uuid.UUID, 0, len(decls))
	events := make([]attackEvent, 0, len(decls))

	for _, d := range decls {
		card := findBattlefieldCard(g, d.Attacker)
		if card == nil {
			continue
		}
		// S24: the whole attacker-side eligibility list — creature,
		// untapped, not already declared, no defender, not summoning
		// sick, no "can't attack" restriction — is AttackerEligible,
		// and internal/legal's enumerator calls the same function.
		// It used to be spelled out here and spelled out again over
		// there, which is the shape #544 wedged a table with. `seat`
		// is the card's own controller because authorization is
		// enforced one layer up in actions.Dispatch.
		if !AttackerEligible(card, card.Controller) {
			continue
		}
		// S27: the target may be a player, a planeswalker or a
		// battle. canAttackTargetLocked folds in "not yourself", "not
		// a planeswalker you control" and "not a battle you protect",
		// which is what the bare controller comparison used to cover
		// for the player-only case.
		if g.canAttackTargetLocked(card.Controller, d.Target) != nil {
			continue
		}
		card.AttackingTarget = d.Target
		// CR 508.1f: declaring an attacker taps it unless it has
		// vigilance (CR 702.20).
		if !HasKeyword(card, "vigilance") {
			card.Tapped = true
		}
		// Attacking and blocking are mutually exclusive per card.
		card.BlockingTarget = uuid.Nil
		declared = append(declared, d.Attacker)
		events = append(events, attackEvent{
			attacker:   d.Attacker,
			controller: card.Controller,
			target:     d.Target,
		})
	}
	if len(declared) == 0 {
		return nil, ErrNoLegalAttackers
	}
	for _, ev := range events {
		g.EmitEvent(Event{
			Kind:   EventAttack,
			Actor:  ev.controller,
			CardID: ev.attacker,
			Target: ev.target,
		})
	}
	// One drain for the whole declaration — see the CR 508.2 note in
	// the doc comment.
	g.runStateChecksLocked()
	return declared, nil
}

// DeclareBlocker marks a battlefield card as blocking a specific
// declared attacker. Gated by the declare_blockers step. Both IDs
// must exist on the battlefield, and the blocker must be a
// creature.
//
// Returns ErrWrongStep outside the declare_blockers step,
// ErrNotACreature for a non-creature blocker, and ErrCardNotFound
// when either card is missing from the battlefield. Idempotent on
// the same pair.
//
// The attacker need not currently have AttackingTarget set — the
// sandbox accepts pre-emptive blocker declarations.
func (g *Game) DeclareBlocker(blockerID, attackerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.Turn.Step != StepDeclareBlockers {
		return ErrWrongStep
	}
	// Layers must be fresh so HasKeyword reads the current effective
	// characteristic (flying granted by an anthem this turn has to
	// be visible to CanBlock below).
	g.RecomputeLayersIfStaleLocked()
	// Verify the attacker exists on the battlefield. Without this the
	// blocker would silently point at a non-existent attacker ID.
	var attacker *Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == attackerID {
			attacker = &g.Battlefield.Cards[i]
			break
		}
	}
	if attacker == nil {
		return ErrCardNotFound
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == blockerID {
			blocker := &g.Battlefield.Cards[i]
			if !blocker.IsCreature() {
				return ErrNotACreature
			}
			// CR 509.1b: evasion keywords (flying, menace, fear,
			// shadow, …) restrict which creatures can be declared
			// as blockers. CanBlock is the single helper that
			// consolidates all current S18 evasion rules; future
			// keywords (protection, #662) land there.
			if !CanBlock(attacker, blocker) {
				return ErrIllegalBlock
			}
			// S31 sub-PR 0: announce the declaration for the public
			// game log, but only when the pairing is NEW. The sandbox
			// lets a defender re-point a blocker at a different
			// attacker; that is one decision being revised, not two
			// blocks, and EventAttack draws the same line.
			announce := blocker.BlockingTarget != attackerID
			blocker.BlockingTarget = attackerID
			blocker.AttackingTarget = uuid.Nil
			if announce {
				g.EmitEvent(Event{
					Kind:   EventBlock,
					Actor:  blocker.Controller,
					Source: blocker.InstanceID,
					CardID: blocker.InstanceID,
					Target: attackerID,
				})
			}
			return nil
		}
	}
	return ErrCardNotFound
}

// ResolveCombatDamage applies damage from declared attackers that
// went unblocked. For each card with AttackingTarget set:
//
//   - If no other battlefield card is currently blocking it
//     (BlockingTarget == this attacker's instance ID), the
//     attacker's CurrentPower is subtracted from the target
//     player's life — recorded via the same path as ChangeLife so
//     the change shows up in LifeHistory.
//   - If at least one blocker is declared, the attacker is left
//     alone here; creature-vs-creature damage assignment requires
//     full rules support and remains a manual step until S13+.
//
// AttackingTarget / BlockingTarget are intentionally NOT cleared
// here — they persist through the combat_damage step so the client
// can keep its combat-arrow overlay drawn while damage shows up on
// the affected player headers. AdvanceStep calls clearCombatLocked
// when the cursor moves on to end_combat, which is what actually
// retracts the arrows.
//
// Trample, deathtouch, double strike, lifelink, and other combat
// keywords are NOT modeled — those land with rules enforcement.
//
// Auto-invoked by AdvanceStep when entering the combat_damage step,
// so under normal play this never needs to be called explicitly.
func (g *Game) ResolveCombatDamage() {
	// internal helper — caller holds g.mu via AdvanceStep, or we
	// take it ourselves below.
	g.resolveCombatDamageLocked()
}

// resolveCombatDamageLocked runs combat damage as two substeps
// (CR 510.4: a combat with a first-strike or double-strike creature has
// two combat damage steps). First-strike and double-strike creatures
// assign in the first substep; SBA fires between; regular +
// double-strike creatures assign in the second. A creature that died in
// the first substep does not participate in the second —
// collectCombatants re-reads live battlefield state each call.
//
// Combat step tag (#187, ADR 0053 Decision 1): when the first-strike
// substep runs, every combat damage event of this step is stamped with
// Event.CombatStep — CombatStepFirstStrike for the first pass,
// CombatStepRegular for the second. When it does not run, the step is
// "" and nothing is tagged, so the tag's presence alone says there
// were two beats. The value is passed down as an argument, never kept
// on Game.
//
// Each substep delegates to assignAndDealCombatDamageLocked which
// handles unblocked straight-to-player damage, single-blocker damage
// with trample overflow, multi-blocker via CR 510.1c
// damage-assignment prompt, menace close-out validation, and the
// lifelink / deathtouch hooks.
func (g *Game) resolveCombatDamageLocked() {
	if g.State != StateActive {
		return
	}
	// Ensure post-layer effective P/T + Abilities are current before
	// reading CurrentPower / HasKeyword. The fast-path no-ops when
	// no relevant event has fired since the last recompute.
	g.RecomputeLayersIfStaleLocked()

	// Untagged unless the first-strike substep runs.
	regularStep := ""

	// Substep 1 — first-strike damage (CR 510.4). Creatures with
	// first strike or double strike participate.
	if g.hasAnyFirstStrikeCombatants() {
		g.assignAndDealCombatDamageLocked(CombatStepFirstStrike)
		// SBA + trigger drain between substeps so creatures that
		// died to first-strike damage exit before the regular pass.
		g.runStateChecksLocked()
		// Re-recompute in case something died that had a static
		// ability (anthem off the field, etc.) and its absence
		// affects the regular-substep attackers/blockers.
		g.RecomputeLayersIfStaleLocked()
		regularStep = CombatStepRegular
	}

	// Substep 2 — regular damage (CR 510.4). Creatures with
	// double strike (re-hit) OR without first strike.
	g.assignAndDealCombatDamageLocked(regularStep)
	g.runStateChecksLocked()

	// Combat state is intentionally left in place. AdvanceStep's
	// transition into end_combat invokes clearCombatLocked, which
	// clears AttackingTarget / BlockingTarget. Deferring the clear
	// lets the client keep combat arrows drawn for the full duration
	// of the combat_damage step.
}

// hasAnyFirstStrikeCombatants reports whether at least one attacker
// or blocker currently on the battlefield has first strike or
// double strike. When zero, the first-strike substep is skipped.
// Caller must hold g.mu.
func (g *Game) hasAnyFirstStrikeCombatants() bool {
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.AttackingTarget == uuid.Nil && c.BlockingTarget == uuid.Nil {
			continue
		}
		if HasKeyword(c, "first strike") || HasKeyword(c, "double strike") {
			return true
		}
	}
	return false
}

// participatesInSubstep reports whether a combatant deals damage in
// the given substep. Per CR 702.4 / 702.7:
//   - First-strike substep: creatures with first strike OR double strike
//   - Regular substep: creatures with double strike OR without first strike
func participatesInSubstep(c *Card, firstStrike bool) bool {
	fs := HasKeyword(c, "first strike")
	ds := HasKeyword(c, "double strike")
	if firstStrike {
		return fs || ds
	}
	return ds || !fs
}

// assignAndDealCombatDamageLocked assigns and applies damage for one
// substep. step is the substep's Event.CombatStep value:
// CombatStepFirstStrike selects the first-strike substep, and anything
// else ("" or CombatStepRegular) is the regular substep, tagged only
// when a first-strike substep ran before it. Builds per-attacker
// blocker lists (filtered to substep-participating blockers),
// snapshots power, then iterates attackers:
//
//   - Unblocked → straight to AttackingTarget via
//     markCombatDamageToPlayerLocked.
//   - Single blocker → full power on the blocker; trample overflow
//     spills to AttackingTarget if (i) the attacker has trample and
//     (ii) the blocker was assigned at-least-lethal.
//   - Multi-blocker → queue a PendingChoiceDamageAssignment prompt;
//     the attacker's damage lands only after resolve.
//
// Blockers deal their power back to the attacker simultaneously
// (CR 510.1d). Lifelink and deathtouch are dispatched inside the
// damage-marking helpers so they fire uniformly via the replacement
// pipeline.
//
// Menace enforcement (CR 702.111): a single blocker on a menace
// attacker is silently reverted — clear BlockingTarget, leaving the
// attacker unblocked. Done at the top of this function so the
// subsequent blocker list reflects the final legal state.
//
// Caller must hold g.mu.
func (g *Game) assignAndDealCombatDamageLocked(step string) {
	firstStrike := step == CombatStepFirstStrike

	// Menace close-out: scan attackers, if any has menace and
	// exactly one blocker assigned, revert that blocker.
	menaceReversions := 0
	blockersByAttacker := make(map[uuid.UUID][]int, len(g.Battlefield.Cards))
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.BlockingTarget != uuid.Nil {
			blockersByAttacker[c.BlockingTarget] = append(blockersByAttacker[c.BlockingTarget], i)
		}
	}
	for atkID, blkIdxs := range blockersByAttacker {
		atk := findBattlefieldCard(g, atkID)
		if atk == nil {
			continue
		}
		if !HasKeyword(atk, "menace") {
			continue
		}
		if len(blkIdxs) == 1 {
			g.Battlefield.Cards[blkIdxs[0]].BlockingTarget = uuid.Nil
			menaceReversions++
		}
	}
	if menaceReversions > 0 {
		// Rebuild blocker map after reversions.
		blockersByAttacker = make(map[uuid.UUID][]int, len(g.Battlefield.Cards))
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.BlockingTarget != uuid.Nil {
				blockersByAttacker[c.BlockingTarget] = append(blockersByAttacker[c.BlockingTarget], i)
			}
		}
	}

	// Snapshot per-card power BEFORE marking any damage so
	// simultaneous resolution (CR 510.1c/d) sees consistent inputs.
	power := make(map[uuid.UUID]int, len(g.Battlefield.Cards))
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.AttackingTarget != uuid.Nil || c.BlockingTarget != uuid.Nil {
			power[c.InstanceID] = c.CurrentPower()
		}
	}

	// Collect attackers first so we don't re-iterate a slice we may
	// mutate via the pending-prompt path (Battlefield stays stable,
	// but keep the loop over an indexed snapshot for clarity).
	attackerIDs := make([]uuid.UUID, 0, len(g.Battlefield.Cards))
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].AttackingTarget != uuid.Nil {
			attackerIDs = append(attackerIDs, g.Battlefield.Cards[i].InstanceID)
		}
	}

	for _, atkID := range attackerIDs {
		atk := findBattlefieldCard(g, atkID)
		if atk == nil {
			continue // died during this substep's iteration
		}
		atkParticipates := participatesInSubstep(atk, firstStrike)
		atkPower := power[atkID]
		// Blocker list for this attacker — taken as the live (not
		// reverted) set. Per-substep participation is checked later
		// per-blocker so a first-strike blocker can still hit a
		// vanilla attacker in the first-strike substep even when the
		// attacker itself doesn't participate (CR 510.2; blocker
		// damage is independent of the attacker's keywords).
		blkIdxs := blockersByAttacker[atkID]
		liveBlockers := make([]uuid.UUID, 0, len(blkIdxs))
		for _, bi := range blkIdxs {
			blk := &g.Battlefield.Cards[bi]
			if blk.BlockingTarget != atkID {
				continue // reverted earlier in this substep
			}
			liveBlockers = append(liveBlockers, blk.InstanceID)
		}

		// Attacker's damage — gated on the attacker participating
		// in this substep. When it doesn't (e.g. a vanilla attacker
		// in the first-strike substep), skip the whole attacker-
		// side branch but fall through to blocker damage below.
		if atkParticipates {
			if len(liveBlockers) == 0 {
				if atkPower > 0 {
					g.dealCombatDamageToAttackTargetLocked(atk.AttackingTarget, atkID, atkPower, step)
				}
			} else if len(liveBlockers) == 1 {
				// Single blocker: attacker assigns all power. Trample
				// overflows to the defending player if the blocker
				// gets at-least-lethal (CR 702.19b).
				blkID := liveBlockers[0]
				blk := findBattlefieldCard(g, blkID)
				if blk != nil && atkPower > 0 {
					lethalThreshold := blk.CurrentToughness() - blk.DamageMarked
					if HasKeyword(atk, "deathtouch") && lethalThreshold > 1 {
						lethalThreshold = 1
					}
					if lethalThreshold < 0 {
						lethalThreshold = 0
					}
					toBlocker := atkPower
					toPlayer := 0
					if HasKeyword(atk, "trample") && atkPower > lethalThreshold {
						toBlocker = lethalThreshold
						toPlayer = atkPower - lethalThreshold
					}
					if toBlocker > 0 {
						g.markCombatDamageOnCardLocked(blkID, toBlocker, atkID, step)
					}
					if toPlayer > 0 {
						g.dealCombatDamageToAttackTargetLocked(atk.AttackingTarget, atkID, toPlayer, step)
					}
				}
			} else {
				// Multi-blocker → queue damage-assignment prompt.
				// Server pauses here; the resume path fires the actual
				// marks. Until resume, the attacker's damage is NOT
				// applied. Blocker damage still flows (simultaneous)
				// because blockers deal power back regardless of
				// attacker's assignment (CR 510.1d).
				g.queueDamageAssignmentPromptLocked(atk, liveBlockers, atkPower, step)
			}
		}

		// Blockers assign damage to the attacker simultaneously
		// (CR 510.1d). Each blocker's damage flows independently of
		// the attacker's split — and independently of whether the
		// attacker itself participates in this substep, so a first-
		// strike blocker can still hit a vanilla attacker in the
		// first-strike substep (CR 510.2).
		for _, blkID := range liveBlockers {
			blk := findBattlefieldCard(g, blkID)
			if blk == nil || !participatesInSubstep(blk, firstStrike) {
				continue
			}
			blkPower := power[blkID]
			if blkPower <= 0 {
				continue
			}
			g.markCombatDamageOnCardLocked(atkID, blkPower, blkID, step)
		}
	}
}

// queueDamageAssignmentPromptLocked queues a CR 510.1c prompt for
// multi-blocker damage assignment. The attacker's controller is the
// chooser; the client renders a drag-reorder + per-blocker damage
// input panel and returns
// {assignments: [{blocker_id, amount}], trample_to_player}. The
// ResolveDamageAssignment resume path dispatches the damage through
// the same pipeline helpers so lifelink / deathtouch / Fog-style
// replacement all fire uniformly.
//
// step is the queuing substep's Event.CombatStep value. It is stored on
// the frame as-is ("" when no first-strike substep ran) so the resume
// tags the assigned damage the way the same substep's direct paths are
// tagged; FirstStrike is kept alongside it for #702.
//
// Caller must hold g.mu.
func (g *Game) queueDamageAssignmentPromptLocked(atk *Card, blockerIDs []uuid.UUID, atkPower int, step string) {
	frame := &DamageAssignmentFrame{
		AttackerID:        atk.InstanceID,
		BlockerIDs:        append([]uuid.UUID(nil), blockerIDs...),
		AttackerPower:     atkPower,
		AllowTrample:      HasKeyword(atk, "trample"),
		HasDeathtouch:     HasKeyword(atk, "deathtouch"),
		FirstStrike:       step == CombatStepFirstStrike,
		CombatStep:        step,
		SourceLifelink:    HasKeyword(atk, "lifelink"),
		SourceController:  atk.Controller,
		SourceIsCommander: atk.IsCommander,
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:             PendingChoiceDamageAssignment,
		Chooser:          atk.Controller,
		Count:            len(blockerIDs),
		Source:           atk.InstanceID,
		Reason:           "Assign combat damage",
		DamageAssignment: frame,
	})
}

// markCombatDamageOnCardLocked adds damage to a battlefield card and
// emits an EventDealDamage. Caller must hold g.mu and have validated
// the cardID is on the battlefield. No SBA fires between calls — the
// caller is responsible for invoking runStateChecksLocked once after
// all combat damage is marked, so simultaneous-resolution semantics
// (CR 510.1c) hold.
//
// S17 sub-PR 5: routes through the CR 614 replacement pipeline with
// IsCombatDamage=true so Fog-class prevention effects intercept.
// Damage can be modified (reduced) or canceled entirely.
//
// step is the substep's Event.CombatStep value (#187). It rides the
// damage tail, so a CR 616 pause keeps it.
func (g *Game) markCombatDamageOnCardLocked(cardID uuid.UUID, amount int, source uuid.UUID, step string) {
	if amount <= 0 {
		return
	}
	ev := &ReplacementEvent{
		Kind:           RepEventDamage,
		Source:         source,
		DamageSource:   source,
		DamageTarget:   cardID,
		DamageAmount:   amount,
		IsCombatDamage: true,
		// #694: the source's CR 702.2c deathtouch, CR 702.15 lifelink
		// and CR 903.10a commander status are snapshotted HERE, before
		// the pipeline can pause. A CR 616 prompt is answered after
		// the blockers' damage has already landed and been swept, so
		// the attacker may be in a graveyard by the time this event
		// resumes — reading the keywords afterwards would silently
		// lose both riders. Same reasoning the DamageAssignmentFrame
		// paths have always used.
		damageTail: g.combatDamageTailLocked(damageTailPermanent, source, step),
	}
	// S27 (#406): what the damage DOES depends on what the permanent
	// is — marked on a creature, loyalty off a planeswalker, defense
	// off a battle (CR 120.3). See permanent_damage.go. No SBA here:
	// all combat damage is simultaneous (CR 510.2) and the resolver
	// sweeps once after the whole step.
	//
	// A CR 616 ordering prompt lands the damage from the resume,
	// through the same tail this call would have run (#694) — before
	// that it was dropped on the floor.
	_, _ = g.damageThroughReplacementsLocked(ev)
}

// markCombatDamageFromFrameLocked is the damage-to-creature router
// for the CR 510.1c damage-assignment prompt's resume path. Same
// shape as markCombatDamageOnCardLocked but sources the deathtouch
// and lifelink keyword state from the prompt's DamageAssignmentFrame
// instead of looking up the attacker via findBattlefieldCard —
// critical because the attacker may have died to blocker damage
// that resolved in the same substep before the prompt fires.
//
// Caller must hold g.mu.
func (g *Game) markCombatDamageFromFrameLocked(cardID uuid.UUID, amount int, frame *DamageAssignmentFrame) {
	if amount <= 0 || frame == nil {
		return
	}
	ev := &ReplacementEvent{
		Kind:           RepEventDamage,
		Source:         frame.AttackerID,
		DamageSource:   frame.AttackerID,
		DamageTarget:   cardID,
		DamageAmount:   amount,
		IsCombatDamage: true,
		// #694: the frame IS the snapshot, so the tail is built from
		// it rather than from a battlefield lookup.
		damageTail: damageTailFromFrame(damageTailPermanent, frame),
	}
	// S27 (#406): same CR 120.3 split the direct path uses. A CR 616
	// ordering prompt lands the damage through the same tail when it
	// is answered (#694).
	_, _ = g.damageThroughReplacementsLocked(ev)
}

// markCombatDamageToPlayerFromFrameLocked is the damage-to-player
// counterpart of markCombatDamageFromFrameLocked — used by the
// damage-assignment resume path for trample overflow to the
// defending player. Sources lifelink state from the frame.
//
// Caller must hold g.mu.
func (g *Game) markCombatDamageToPlayerFromFrameLocked(playerID uuid.UUID, amount int, frame *DamageAssignmentFrame) {
	if amount <= 0 || frame == nil {
		return
	}
	ev := &ReplacementEvent{
		Kind:           RepEventDamage,
		Source:         frame.AttackerID,
		DamageSource:   frame.AttackerID,
		DamageTarget:   playerID,
		DamageAmount:   amount,
		IsCombatDamage: true,
		// #694: the frame IS the snapshot. CR 903.10a trample overflow
		// from a commander counts toward the 21-damage SBA, and the
		// tally is keyed on frame.AttackerID (the commander's instance
		// ID, the CommanderDamage key since S25 / #77) because the
		// attacker may have died to blocker damage before the prompt
		// resolved.
		damageTail: damageTailFromFrame(damageTailPlayer, frame),
	}
	// A CR 616 ordering prompt lands the damage through the same tail
	// when it is answered (#694).
	_, _ = g.damageThroughReplacementsLocked(ev)
}

// markCombatDamageToPlayerLocked applies combat damage to a player
// via the replacement pipeline (Fog-class prevention + future
// lifelink / redirect hooks). Zeroes out if canceled. Caller must
// hold g.mu. Added in S17 sub-PR 5.
//
// step is the substep's Event.CombatStep value (#187). It rides the
// damage tail, so a CR 616 pause keeps it.
func (g *Game) markCombatDamageToPlayerLocked(playerID, source uuid.UUID, amount int, step string) {
	if amount <= 0 {
		return
	}
	ev := &ReplacementEvent{
		Kind:           RepEventDamage,
		Source:         source,
		DamageSource:   source,
		DamageTarget:   playerID,
		DamageAmount:   amount,
		IsCombatDamage: true,
		// #694: snapshot the source before the pipeline can pause —
		// CR 903.10a commander damage (keyed on the commander's own
		// instance ID since S25 / #77) and CR 702.15 lifelink both
		// ride on it, and a CR 616 prompt is answered after the rest
		// of the combat damage has landed.
		damageTail: g.combatDamageTailLocked(damageTailPlayer, source, step),
	}
	// A CR 616 ordering prompt lands the damage through the same tail
	// when it is answered (#694). Before that fix the pause dropped the
	// life loss entirely.
	_, _ = g.damageThroughReplacementsLocked(ev)
}

// ClearCombat resets every card on the battlefield to "not attacking
// and not blocking". Called by the active player at end of combat
// (or by anyone, really — the sandbox doesn't gate it). Cheap O(n)
// pass over the battlefield.
func (g *Game) ClearCombat() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	g.clearCombatLocked()
	return nil
}

// clearCombatLocked is the internal mutator behind both ClearCombat
// (which takes the lock) and AdvanceStep (which already holds it).
// Caller must hold g.mu.
func (g *Game) clearCombatLocked() {
	for i := range g.Battlefield.Cards {
		g.Battlefield.Cards[i].AttackingTarget = uuid.Nil
		g.Battlefield.Cards[i].BlockingTarget = uuid.Nil
	}
}

// Concede marks the given player as eliminated. If exactly one
// non-eliminated player remains after the mutation, the game's State
// transitions to StateEnded — derived state, no separate Winner field
// is stored (the surviving seat is the implicit winner).
//
// Idempotent on already-eliminated players returns ErrPlayerEliminated
// rather than silently swallowing — clients should disable the
// concede button after the first press, and a duplicate frame from a
// stale tab is worth surfacing.
//
// If Concede is called from the lobby state (no game started yet),
// returns ErrGameNotActive — there is nothing to lose. Calling it
// after the game has already ended is also rejected.
func (g *Game) Concede(playerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if p.Eliminated {
		return ErrPlayerEliminated
	}
	g.EmitEvent(Event{Kind: EventConcede, Actor: playerID})
	// S13.1: delegate to the unified elimination path so concede
	// fires the same stack cleanup + cursor advance + game-end
	// check as an SBA-driven loss.
	g.eliminatePlayerLocked(p)
	return nil
}

// advancePastEliminatedLocked moves the turn cursor forward to the
// next non-eliminated seat if the current ActiveSeat is eliminated.
// Called from Concede (and intended for future state-based action
// elimination paths). Caller must hold g.mu.
func (g *Game) advancePastEliminatedLocked() {
	numSeats := len(g.Seats)
	if numSeats == 0 {
		return
	}
	if !g.Seats[g.Turn.ActiveSeat].Eliminated {
		// Active seat still standing — nothing to advance.
		// But priority might be on an eliminated seat; reset it to
		// the active seat (priority always defaults to active at
		// step boundaries). NoPriority means no one holds priority
		// (Untap/Cleanup) — leave it alone.
		ph := g.Turn.PriorityHolder
		if ph >= 0 && ph < numSeats && g.Seats[ph].Eliminated {
			g.Turn.PriorityHolder = g.Turn.ActiveSeat
		}
		return
	}
	// Active seat eliminated — find the next survivor in turn order.
	// Bounded by numSeats to terminate even if every seat is
	// eliminated (the surrounding Concede caller transitions to
	// StateEnded in that case; the cursor doesn't matter post-end
	// but should not infinite-loop here).
	for i := 0; i < numSeats; i++ {
		next := (g.Turn.ActiveSeat + 1) % numSeats
		nextNumber := g.Turn.Number
		if next == 0 {
			nextNumber++
		}
		g.Turn = Turn{
			Number:         nextNumber,
			ActiveSeat:     next,
			PriorityHolder: initialPriorityHolder(StepUntap, next),
			Phase:          PhaseOf(StepUntap),
			Step:           StepUntap,
		}
		if !g.Seats[next].Eliminated {
			// New active seat starts on Untap; run the entry hook so
			// the auto-untap fires and the cursor advances out of the
			// no-priority step, matching the rest of the engine.
			g.runStepEntryHooksLocked()
			return
		}
	}
}

// PassTurn skips to the next player's untap step, regardless of
// whatever step the current turn is in. Useful for forfeiting a turn
// or when all steps are uneventful. Lands on Untap with NoPriority
// (S13); the entry hook auto-untaps and walks the cursor on to
// Upkeep, matching the normal-flow behaviour of priority wraps and
// AdvanceStep so callers always end at a priority-granting step.
func (g *Game) PassTurn() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	// #730: gated for the same reason advance_step is, and more so —
	// this verb walks the cursor through every remaining step of the
	// turn. A prompt left open behind it is unanswerable in practice.
	// See choice_gate.go.
	if c := g.blockingChoiceLocked(); c != nil {
		return choicePendingErrorLocked(c)
	}
	// Jump to this turn's cleanup and wrap through the shared seam so
	// eliminated seats are skipped and per-turn caches clear.
	g.Turn.Step = StepCleanup
	g.advanceCursorLocked()
	// Refresh per-turn budgets (undo, future per-turn counters) on
	// the new active seat — same hook AdvanceStep / PassPriority's
	// wrap branch run when stepping into untap. The hook also auto-
	// untaps and advances past Untap (no priority) so the cursor
	// lands at Upkeep.
	g.runStepEntryHooksLocked()
	// CR 117.5 / 704.3: new priority grant → run SBAs.
	g.runStateChecksLocked()
	return nil
}

// Mulligan shuffles the player's entire hand back into their library
// and draws newHandSize cards. This is the simplified "London
// mulligan" shape without the card-to-bottom penalty — S08 keeps the
// penalty out of scope; rules enforcement arrives in S13+.
//
// Mulligan increments MulligansTaken and resets HandKept to false:
// taking a mulligan is a fresh decision, so the player must commit
// again afterward via KeepHand.
func (g *Game) Mulligan(playerID uuid.UUID, newHandSize int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if newHandSize < 0 {
		return ErrInvalidParam
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	for _, c := range p.Hand.Cards {
		p.Library.PushTop(c)
	}
	p.Hand.Cards = nil
	p.Library.Shuffle(g.randForLocked(rngStream{kind: rngStreamShuffle, player: p.ID}))
	// S13.5: shuffle wipes per-card knowledge across hand + library
	// (the hand cards are now indistinguishable from the rest of the
	// shuffled pile from the opponent's perspective, and the owner
	// no longer knows the new opening hand until they redraw it).
	clearKnownInZoneLocked(p.Library)
	for range newHandSize {
		if p.Library.Size() == 0 {
			break
		}
		c, _ := p.Library.PopTop()
		p.Hand.PushTop(c)
		// Owner immediately knows their newly-drawn opening hand.
		g.markCardKnownInZoneLocked(p.Hand, c.InstanceID)
	}
	p.MulligansTaken++
	p.HandKept = false
	return nil
}

// KeepHand marks the player as having committed to their current
// opening hand. Idempotent — calling it twice on a player who has
// already kept is a no-op (avoids race conditions where two stale
// client tabs both press keep). Once every seated, non-eliminated
// player has KeptHand, Game.MulligansOpen flips false and the
// "real" game UI takes over.
//
// Returns ErrGameNotActive in lobby/ended state and
// ErrPlayerEliminated for an eliminated player (a defensive guard;
// the client shouldn't surface the keep button in that case anyway).
func (g *Game) KeepHand(playerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if p.Eliminated {
		return ErrPlayerEliminated
	}
	if p.HandKept {
		return nil
	}
	p.HandKept = true
	// Close the mulligan window once all non-eliminated seats have
	// committed. Eliminated seats (improbable here — elimination
	// during the mulligan window would be unusual but possible if
	// the conceded flow is exercised mid-decision) don't gate the
	// transition.
	allKept := true
	for _, seat := range g.Seats {
		if seat.Eliminated {
			continue
		}
		if !seat.HandKept {
			allKept = false
			break
		}
	}
	if allKept {
		g.MulligansOpen = false
		// First-step entry happens here, not at Start: the cursor has
		// been parked on Untap with NoPriority since Start, waiting
		// for everyone to commit. Run the hook now so seat 0's auto-
		// untap fires and the cursor advances to Upkeep, matching the
		// shape of every subsequent step transition. (S13.)
		g.runStepEntryHooksLocked()
	}
	return nil
}

// ShuffleLibrary reshuffles the given player's library in place,
// drawing from the player's shuffle stream (rng.go), so seeded test
// runs stay deterministic and an undone shuffle redoes identically. S13.5: clears KnownBy on every library card —
// any prior scry / top-of-library knowledge dissolves with the
// shuffle (CR 701.24 + the per-instance KnownBy invariant).
func (g *Game) ShuffleLibrary(playerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	p.Library.Shuffle(g.randForLocked(rngStream{kind: rngStreamShuffle, player: p.ID}))
	clearKnownInZoneLocked(p.Library)
	g.EmitEvent(Event{Kind: EventSearchLibrary, Actor: playerID, Label: "shuffle"})
	return nil
}

// ChangePlayerLife adjusts a player's life total by delta (positive
// for gain, negative for loss) and returns the new total.
//
// The sandbox verb: a player dragging their own life counter. It has
// run the CR 614 window since S17 sub-PR 2, and since #482 so does
// every other writer of a life total — all of them land in the one
// tail in life_tail.go, so a hand-typed life change and a catalog
// GainLife can no longer disagree about which replacements apply.
//
// A CR 616 ordering prompt returns (0, nil): the change lands from the
// resume when the affected player answers, and there is no new total
// to report yet.
func (g *Game) ChangePlayerLife(playerID uuid.UUID, delta int) (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return 0, ErrGameNotActive
	}
	ev := &ReplacementEvent{
		Kind:       RepEventLife,
		LifePlayer: playerID,
		LifeDelta:  delta,
	}
	paused, err := g.changeLifeThroughReplacementsLocked(ev)
	if err != nil {
		return 0, err
	}
	if paused {
		return 0, nil
	}
	// Read back off the event rather than the argument: a replacement
	// is free to have redirected the change to someone else.
	p := g.playerByIDLocked(ev.LifePlayer)
	if p == nil {
		return 0, ErrPlayerNotFound
	}
	return p.Life, nil
}

// AddCounter modifies a named counter on a card by delta. Creates the
// counter entry if it's not present. If the resulting count is zero
// or negative, the counter is removed entirely to keep the map clean.
// A delta of 0 is a no-op and does not allocate a counter map.
// The card may be in any zone; rules that restrict counter types to
// specific zones (loyalty only on planeswalkers, etc.) are deferred.
func (g *Game) AddCounter(cardID uuid.UUID, name string, delta int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if name == "" {
		return ErrInvalidParam
	}
	// S17 sub-PR 2: counter-placement replacements (Doubling Season
	// doubles, Hardened Scales adds 1) will fire from here starting
	// in sub-PR 3. For sub-PR 2 zero catalog counter-replacements
	// are registered, so this is byte-for-byte identical to pre-S17.
	ev := &ReplacementEvent{
		Kind:          RepEventCounter,
		CounterTarget: cardID,
		CounterName:   name,
		CounterDelta:  delta,
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		return nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		return nil
	}
	return g.applyCounterLocked(out.CounterTarget, out.CounterName, out.CounterDelta)
}

// applyCounterLocked is the actual counter-map mutation, extracted
// from AddCounter so the replacement pipeline and the AddCounterForEffect
// helper can share the body. Caller must hold g.mu.
func (g *Game) applyCounterLocked(cardID uuid.UUID, name string, delta int) error {
	if name == "" {
		return ErrInvalidParam
	}
	if delta == 0 {
		return nil
	}
	z := g.findCardZoneLocked(cardID)
	if z == nil {
		return ErrCardNotFound
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID == cardID {
			hadCounters := len(z.Cards[i].Counters) > 0
			if z.Cards[i].Counters == nil {
				z.Cards[i].Counters = make(map[string]int)
			}
			z.Cards[i].Counters[name] += delta
			newAmount := z.Cards[i].Counters[name]
			if z.Cards[i].Counters[name] <= 0 {
				delete(z.Cards[i].Counters, name)
				if len(z.Cards[i].Counters) == 0 {
					z.Cards[i].Counters = nil
					// #683: not a placeholder any more — see
					// Card.LostLastCounter and the toughness SBA.
					// Only when there was a counter to lose:
					// removing one from a card with none leaves
					// whatever it was before.
					if hadCounters {
						z.Cards[i].LostLastCounter = true
					}
				}
			}
			g.EmitEvent(Event{
				Kind:   EventCounterPlaced,
				Target: cardID,
				Label:  name,
				Amount: newAmount,
			})
			return nil
		}
	}
	return ErrCardNotFound
}

// SetCommanderDamage sets the total damage one commander has dealt
// to the target player. This is a set operation, not an increment —
// the caller computes and sends the new total. It is the sandbox
// override behind the commander-damage UX, for the cases the engine
// cannot see (a manually resolved effect, a misclick to undo).
//
// `from` is a commander's card INSTANCE ID as of S25 (#77); it was a
// player ID before the rekey. It is looked up across every tracked
// zone rather than on the battlefield alone, because a commander
// sitting in the command zone after a lethal swing is exactly when a
// player reaches for this control. `to` is a seated player.
//
// ErrCardNotFound for an ID that names no card or names a card that
// is not a commander; ErrPlayerNotFound for an unseated `to`. Both
// beat silently polluting the target's damage map with a key nothing
// will ever render — the client keys its per-commander rows off
// instance IDs, so a bogus key is invisible rather than merely wrong.
func (g *Game) SetCommanderDamage(from, to uuid.UUID, amount int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	src, ok := g.LookupCardForEffect(from)
	if !ok || !src.IsCommander {
		return ErrCardNotFound
	}
	target := g.playerByIDLocked(to)
	if target == nil {
		return ErrPlayerNotFound
	}
	if amount < 0 {
		amount = 0
	}
	target.CommanderDamage[from] = amount
	return nil
}

// SetMonarch designates the given player as the monarch (Conspiracy
// mechanic — the monarch draws an extra card at the end of their turn
// in MTG, and any opponent who deals combat damage to them becomes the
// new monarch). Pass uuid.Nil to clear (no current monarch — the rare
// case where a card explicitly removes monarchy).
//
// Sandbox-only: the must-attack constraint and the combat-damage
// transfer rule are not enforced. The marker is the affordance; the
// rules graft track will hook these up if/when the playgroup wants
// them auto-handled.
//
// Returns ErrPlayerNotFound if playerID isn't seated, or
// ErrGameNotActive in lobby/ended state.
func (g *Game) SetMonarch(playerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if playerID != uuid.Nil && g.playerByIDLocked(playerID) == nil {
		return ErrPlayerNotFound
	}
	g.Monarch = playerID
	return nil
}

// SetInitiative designates the given player as having the initiative
// (Commander Legends: Battle for Baldur's Gate — venture into the
// Undercity at the start of upkeep; combat damage transfers initiative).
// Pass uuid.Nil to clear. Same sandbox / non-enforcement posture as
// SetMonarch — the marker is what's surfaced.
func (g *Game) SetInitiative(playerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if playerID != uuid.Nil && g.playerByIDLocked(playerID) == nil {
		return ErrPlayerNotFound
	}
	g.Initiative = playerID
	return nil
}

// SetGoaded marks a battlefield creature as goaded by the given player.
// Pass uuid.Nil for `by` to clear the goad. The card must be on the
// battlefield; goading a card in any other zone is meaningless.
//
// Sandbox-only: the must-attack-and-not-the-goader constraint is not
// enforced — the marker is the affordance for players to remember.
func (g *Game) SetGoaded(cardID, by uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if by != uuid.Nil && g.playerByIDLocked(by) == nil {
		return ErrPlayerNotFound
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			g.Battlefield.Cards[i].GoadedBy = by
			return nil
		}
	}
	return ErrCardNotFound
}

// SetPoison sets a player's poison counter total. 10 is loss in MTG
// (state-based action, deferred to S13+). amount is clamped at 0 from
// below; there's no upper clamp because some cards / formats deal
// arbitrary poison. Replaces (not increments) — clients send the new
// total so two stale tabs don't double-count.
// SetDiscordIdentity stamps the S12.5 OAuth identity metadata
// onto the given seat. Called by the lobby's JoinWithIdentity
// immediately after AddPlayer so the fields are in place before
// the next snapshot is built. Valid in both lobby and active
// state (a "link Discord" flow lands here mid-game without a
// state restriction). Unknown player → ErrPlayerNotFound.
func (g *Game) SetDiscordIdentity(playerID uuid.UUID, discordID, avatarHash, displayName string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	p.DiscordID = discordID
	p.DiscordAvatarHash = avatarHash
	p.DisplayName = displayName
	return nil
}

// SetBot marks a seat as bot-driven with the named policy tier and
// the curated deck it was seated with (deckID may be empty when the
// caller supplied a raw decklist). Lobby state only — a seat cannot
// change hands mid-game. Added in S31 sub-PR 4.
func (g *Game) SetBot(playerID uuid.UUID, tier, deckID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateLobby {
		return ErrGameNotInLobby
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	p.IsBot = true
	p.BotTier = tier
	p.BotDeck = deckID
	return nil
}

// RemovePlayer unseats a player and closes the gap in seat numbers.
// Lobby state only: once a game has started a seat is permanent (a
// player leaves by conceding). Added in S31 sub-PR 4 so a bot can be
// removed from an unstarted table.
func (g *Game) RemovePlayer(playerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateLobby {
		return ErrGameNotInLobby
	}
	idx := -1
	for i, p := range g.Seats {
		if p != nil && p.ID == playerID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrPlayerNotFound
	}
	g.Seats = append(g.Seats[:idx], g.Seats[idx+1:]...)
	for i, p := range g.Seats {
		p.Seat = i
	}
	return nil
}

func (g *Game) SetPoison(playerID uuid.UUID, amount int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if amount < 0 {
		amount = 0
	}
	p.Poison = amount
	// S13.2: keep the unified Counters map in sync so the SBA loop
	// reads the same value the legacy SetPoison action wrote.
	setPlayerCounterLocked(p, CounterPoison, amount)
	g.runStateChecksLocked()
	return nil
}

// AddPlayerCounter modifies a named player-level counter by `delta`
// (positive to add, negative to remove). Negative deltas that would
// drive the count below zero clamp at zero. Empty / unknown names
// are accepted (Sandbox: homebrew counters are fine — the registry
// in counter_types.go is for the engine and the iconography, not
// validation).
//
// Special-cased identifiers stay synchronised with their legacy int
// fields:
//   - poison ↔ Player.Poison
//   - energy ↔ Player.Energy
//
// Drives the SBA loop after the mutation so 10+ poison or 0 life
// (via energy-cost cards in the future) immediately apply.
//
// Caller must NOT hold g.mu — this method takes the write lock.
//
// Added in S13.2.
func (g *Game) AddPlayerCounter(playerID uuid.UUID, name string, delta int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if name == "" {
		return ErrInvalidParam
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if p.Counters == nil {
		p.Counters = make(map[string]int)
	}
	cur := p.Counters[name]
	next := cur + delta
	if next < 0 {
		next = 0
	}
	setPlayerCounterLocked(p, name, next)
	// Mirror to legacy single-int fields so SetPoison / SetEnergy
	// callers continue to read consistent state.
	switch name {
	case CounterPoison:
		p.Poison = next
	case CounterEnergy:
		p.Energy = next
	}
	g.runStateChecksLocked()
	return nil
}

// setPlayerCounterLocked writes a counter value, removing the key
// when the value drops to zero so the map stays sparse on the wire.
// Caller must hold g.mu.
func setPlayerCounterLocked(p *Player, name string, value int) {
	if p.Counters == nil {
		p.Counters = make(map[string]int)
	}
	if value <= 0 {
		delete(p.Counters, name)
		if len(p.Counters) == 0 {
			p.Counters = nil
		}
		return
	}
	p.Counters[name] = value
}

// SetUndoLimit sets the per-player per-turn undo budget. Refreshes
// every seated player's UndosRemaining to the new limit immediately
// (so a mid-turn raise is usable right away by the active player).
// Clamped at 0 from below — passing a negative is treated as "no
// undos allowed".
//
// Sandbox: any seated player or admin may call it. The intent is
// "the table agreed to relax the limit for this game"; in casual
// play that's social, not enforced.
func (g *Game) SetUndoLimit(limit int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if limit < 0 {
		limit = 0
	}
	g.UndoLimit = limit
	for _, p := range g.Seats {
		p.UndosRemaining = limit
	}
	return nil
}

// SpendUndo decrements the named player's UndosRemaining if they
// have any to spend; returns ErrNoUndosRemaining when the budget is
// exhausted. Called from the room layer immediately before a successful
// RestoreFrom so the budget is only debited on a successful undo.
//
// Returns ErrPlayerNotFound for an unseated playerID. Admin / spectator
// undos (callerID uuid.Nil) bypass this entirely — the room layer
// short-circuits to never call SpendUndo for those.
func (g *Game) SpendUndo(playerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if p.UndosRemaining <= 0 {
		return ErrNoUndosRemaining
	}
	p.UndosRemaining--
	return nil
}

// SetEnergy sets a player's energy counter total. Replaces (not
// increments) for the same reason as SetPoison. Clamped at 0 from
// below. No game-loss condition is tied to energy — it's a pure
// resource counter.
func (g *Game) SetEnergy(playerID uuid.UUID, amount int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if amount < 0 {
		amount = 0
	}
	p.Energy = amount
	// S13.2: keep the unified Counters map in sync.
	setPlayerCounterLocked(p, CounterEnergy, amount)
	return nil
}

// controllerOfBattlefieldCardLocked returns the controller of the
// battlefield card with the given ID, or uuid.Nil if it isn't
// there. Used to stamp Actor on combat-damage events so "its
// controller may draw" triggers (Edric) don't have to re-find a
// creature that may have died to simultaneous combat damage by the
// time their prompt is answered. Caller must hold g.mu.
func (g *Game) controllerOfBattlefieldCardLocked(cardID uuid.UUID) uuid.UUID {
	if c := findBattlefieldCard(g, cardID); c != nil {
		return c.Controller
	}
	return uuid.Nil
}

// MoveCardByIDToBottom is MoveCardByIDAsCommander with the card
// seated at the BOTTOM of the destination zone instead of the top.
// It backs the move_card action's `to_bottom` flag, which the admin
// context menu (#170) needs for "put this on the bottom of your
// library" — the one destination MoveCard (which always PushTops)
// can't express.
//
// Implemented as a post-move reorder rather than a second copy of the
// replacement pipeline: the move runs exactly as it always does
// (replacements, LKI snapshot, LTB / ETB events, state-based checks),
// then the card is pulled back out of the destination and re-pushed
// at the bottom. If a replacement rewrote the destination — a
// commander routed to the command zone by CR 903.9 — the card isn't
// in `dst` at all, Remove fails, and the reorder is a no-op. That is
// the behaviour we want: the replacement's destination wins.
//
// The two-phase locking (the move takes g.mu and releases it, then
// this retakes it) is safe because ws.Room.Apply holds the room lock
// across the whole of actions.Dispatch, so no other action can
// interleave between the move and the reorder.
func (g *Game) MoveCardByIDToBottom(src, dst ZoneRef, cardID uuid.UUID, asCommander bool) error {
	if err := g.MoveCardByIDAsCommander(src, dst, cardID, asCommander); err != nil {
		return err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	z := g.zoneFromRefLocked(dst)
	if z == nil {
		return nil
	}
	c, err := z.Remove(cardID)
	if err != nil {
		return nil
	}
	z.PushBottom(c)
	return nil
}
