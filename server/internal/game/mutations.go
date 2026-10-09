package game

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

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
//
// #1479: answered by the card index (card_index.go) rather than by
// walking every zone. The zones it covers — the battlefield, the
// stack, exile and each seat's library, hand, graveyard and command
// zone, not PhasedOut — are unchanged.
func (g *Game) findCardZoneLocked(cardID uuid.UUID) *Zone {
	z, _ := g.locateCardLocked(cardID)
	return z
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
	// S17 sub-PR 2: the draw goes through the replacement pipeline so
	// draw-replacement effects ("if you would draw, mill instead",
	// "if you would draw, opponent draws instead", dredge) fire
	// pre-event. A prompt queued by the window pauses the draw: the
	// resume path re-enters the pipeline and finishes it, and returning
	// nil tells the caller the draw is "in flight" (no ErrZoneEmpty).
	// draw_instead.go.
	return g.drawRunLocked(playerID, 1, DrawThen{})
}

// actuallyDrawCardsLocked performs the N individual card draws a
// settled RepEventDraw asks for (CR 121.2: "if a player is instructed
// to draw multiple cards, that player performs that many individual
// card draws"). N is one for every draw in the game today and more
// only under a draw-amount replacement — Thought Reflection,
// Alhammarret's Archive (#1222).
//
// One at a time is not decoration. Every per-card payoff in the
// catalog reads EventDrawCard (Nekusar, Sheoldred, Consecrated Sphinx,
// Fate Unraveler), so a doubled draw has to emit one per card or they
// all fire once for two cards. The CR 614 window is NOT re-opened for
// the extra cards: they are the same event, and its once-per-event
// tracking (CR 614.5) covers all of them — which is also what stops a
// doubler from doubling its own output forever.
//
// Stops on the first empty library, returning ErrZoneEmpty with
// AttemptedEmptyDraw already set, exactly as one draw does: DrawCard's
// wire behaviour and DrawNForEffect's partial-draw contract both key
// on that error.
//
// A non-positive count draws one. The honest spelling of "no draw" is
// ev.Cancel(), which the pipeline above has already handled, so a zero
// here is a hand-built event's zero value rather than anybody's
// decision — and a draw that silently vanished would be the worse
// answer.
//
// Caller must hold g.mu.
func (g *Game) actuallyDrawCardsLocked(playerID uuid.UUID, n int) error {
	if n <= 0 {
		n = 1
	}
	for i := 0; i < n; i++ {
		if err := g.actuallyDrawCardLocked(playerID); err != nil {
			return err
		}
	}
	return nil
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
	// SplitSecond is the S13.1 sandbox flag: "treat this spell as
	// having split second". Since #1519 a spell that PRINTS split
	// second has it without the flag (castHasSplitSecond reads the
	// card's keywords); the flag survives for a card the catalog and
	// the importer know nothing about, and is ignored on a face-down
	// cast.
	SplitSecond bool

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

	// OptionalCosts names the optional additional costs the caster is
	// choosing to pay (CR 601.2b) — kicker, multikicker, buyback — as
	// POSITIONS in the card's OptionalCosts slice. Paying one N times
	// is naming its index N times, which is how multikicker announces
	// its count (CR 702.33d) without a second field.
	//
	// Announced with the modes and before the targets, because a
	// kicked spell may legally choose different targets from an
	// unkicked one. The mana half joins the total at CR 601.2f; the
	// card and permanent halves ride DiscardIDs and SacrificeIDs
	// after the mandatory cost's own, in index order. The whole list
	// lands on StackItem.Paid.OptionalCosts.
	//
	// Empty for every card that offers nothing — and non-empty for
	// one is rejected, not ignored. Added in ADR 0073 (#664).
	OptionalCosts []int

	// CostBranch names which branch of an either/or additional cost the
	// caster is paying (CR 601.2b, ADR 0100 §2) — an index into the
	// card's AdditionalCost.Either: Demand Answers' "sacrifice an
	// artifact or discard a card". REQUIRED on a branched card and
	// refused on any other; never defaulted to 0, because silently
	// paying a branch the player did not choose is the worst failure
	// available. The branch is then the plan's mandatory entry, so its
	// cards ride DiscardIDs / SacrificeIDs / BlightIDs as any mandatory
	// cost's do, and its mana joins the total at CR 601.2f. Lands on
	// StackItem.Paid.CostBranch.
	CostBranch *int

	// GiftOpponent is the opponent the caster chooses while paying a
	// gift cost (CR 702.174a) — the other half of announcing one.
	// Required exactly when OptionalCosts names the card's gift cost,
	// and rejected, not ignored, when it does not: a recipient sent
	// with no promise is a client that thought it was promising one.
	// Must be a player still in the game other than the caster
	// (GiftOpponentsLocked). Lands on StackItem.Paid.GiftOpponent.
	// Added in ADR 0089 (#1267).
	GiftOpponent uuid.UUID

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

	// TeamworkIDs names the untapped creatures tapped to pay an
	// announced teamwork cost (CR 702.194a): any number of them whose
	// total effective power reaches the teamwork number. BlightIDs
	// names the ONE creature an announced blight cost puts its -1/-1
	// counters on (CR 701.68a). Same discipline as SacrificeIDs:
	// validated at announce, paid with the spell on the stack, and
	// refused rather than ignored when the announcement pays no such
	// cost. Added for #1703.
	TeamworkIDs []uuid.UUID
	BlightIDs   []uuid.UUID

	// RevealIDs names the ONE card an announced reveal / behold cost
	// shows (CR 701.20, ADR 0100 amendment 2026-10-07): a matching card
	// in the caster's hand, or — to behold — a matching permanent they
	// control. Refused rather than ignored when the announcement pays no
	// such cost.
	RevealIDs []uuid.UUID

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

	// DelveIDs names the cards in the caster's graveyard exiled to
	// delve the spell (CR 702.66a, ADR 0100 §1): each one pays for
	// {1} of the generic mana in the total cost, in the order named.
	// Empty is always legal — pay the whole cost with mana. At most
	// CastPrice.DelveBudget of them; a card with no delve that arrives
	// with any is refused rather than ignored.
	//
	// A list of its own rather than AltCostIDs, which escape's
	// graveyard exile rides: delve is not an alternative or an
	// additional cost (CR 702.66b), and a permanent linked to "cards
	// exiled with it" (CR 607.2q) must count these and only these.
	DelveIDs []uuid.UUID

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

	// Fuse announces a FUSED split spell (CR 702.102a, ADR 0103): both
	// halves of a split card with fuse, cast together from hand, for
	// both halves' mana costs (CR 702.102c). Face must be 0 with it.
	// Refused, not ignored, for a card without fuse, from any zone but
	// the hand, with a claimed alternative cost, or for a card whose
	// halves declare modes or additional costs a fused announcement
	// cannot carry.
	Fuse bool

	// PhyrexianLife is how many of the cost's Phyrexian symbols the
	// caster is paying with life instead of mana — 2 life each
	// (CR 107.4f, which covers the ten hybrid Phyrexian symbols
	// {W/U/P}…{G/U/P} too). Zero is the ordinary answer: pay every symbol
	// with its coloured half.
	//
	// A COUNT rather than a list of symbols because the count is the
	// whole of the player's decision — the engine strikes out the
	// symbols a life payment can actually save first (see
	// phyrexian_mana.go), and choosing between two symbols the pool
	// can both pay changes nothing but which colour is left floating.
	//
	// An announce-time parameter for the same reason Face is: CR
	// 601.2b makes "how do you intend to pay each hybrid and
	// Phyrexian symbol" part of announcing the spell. More than the
	// cost prints, or more life than the caster has (CR 119.4), is
	// REJECTED rather than clamped — silently casting for a different
	// price than the player asked for is the worst available failure.
	// Added in S44 (#787).
	//
	// CR 602.2b asks the same question of an activated ability, and
	// ActivateAbilityParams.PhyrexianLife is the same field under the
	// same wire name — one strike-and-pay helper serves both (#917).
	PhyrexianLife int

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
	// plans what the floating pool is missing (ADR 0118 §1,
	// autoTapTopUpLocked) against the caller's untapped permanents
	// other than the locked ones, taps each card in
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

	// commanderAnswers are the CR 903.9 answers the owners of the
	// commanders this cast's costs move gave before the payment began
	// (#1397, cost_commander_choice.go). Unexported: only the parked
	// announcement's resume sets it, so no payload answers for
	// another player's commander.
	commanderAnswers map[uuid.UUID]bool
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
	if err := g.castSpellLocked(playerID, cardID, params); err != nil {
		return err
	}
	// #2275 / CR 117.3c: casting (or playing a land, CR 116.2a, which
	// this verb also carries) is an action, and the passes before it
	// no longer count. A spell restarts the succession by landing on
	// the stack anyway; a land play uses no stack and needs saying.
	g.noteActionTakenLocked(playerID)
	return nil
}

// castSpellLocked is CastSpell's body, split out so a parked cast
// (#1397) can be made again from the CR 903.9 answer's resume. Caller
// must hold g.mu.
func (g *Game) castSpellLocked(playerID, cardID uuid.UUID, params CastSpellParams) error {
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
	src, err := g.castSourceZoneLocked(p, cardID, params.FromZone)
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
	// #1474: a card an EFFECT has paused on its way out cannot be
	// cast or played. CR 601.2a moves the spell before any cost is
	// paid, so it is in none of the cost lists the gate is asked about
	// further down; it is asked here instead, before anything is
	// judged, moved or paid, and for every source zone and the land
	// play alike. See refusePausedCostCardsLocked.
	if err := g.refusePausedCostCardsLocked([]uuid.UUID{cardID}); err != nil {
		return err
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
	// S32: the granted permission is read BEFORE the face gate,
	// because a grant can name a face the card's own layout does not
	// offer — "exile it, then cast it transformed" is a cast of a
	// `transform` card's back face, which CastableFaces refuses on
	// principle. A permission belongs to the card INSTANCE and not to
	// any face, so reading it before SetFace is safe and is the only
	// order in which the gate can consult it. See faceForCastLocked
	// for why a face-naming grant SETS the face rather than merely
	// permitting it.
	//
	// ADR 0066: one lookup for all three granted zones. It answers
	// nil for hand and the command zone, and nil for a card whose own
	// text already opens the zone — Gravecrawler and a printed
	// flashback need no permission and must not be repriced by one.
	grant := g.CastPermissionForClaimLocked(playerID, card, src.Kind, params.AlternativeCost)
	// CR 702.143c, #658. Asked here, of the card as it sits in its
	// source zone, because the answer stops being readable the moment
	// the card moves: CR 406.3a turns a foretold card face up as it is
	// cast, which ADR 0069 decision 5 makes MoveCard's business.
	foretold := CardIsForetold(card)
	// CR 702.62a (#659): a permanent cast from a suspended card has
	// haste. It is the PERMISSION that says so, not the card, so it is
	// read here where the permission is consumed and registered
	// against the object the cast produces. See
	// grantHasteForCastLocked for the layer-6 grant and its declared
	// duration simplification.
	grantsHaste := grant != nil && grant.GrantsHaste
	face, ok := faceForCastLocked(card, params.Face, grant, playerID, params.AlternativeCost)
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
	// ADR 0103, CR 702.102a: a FUSED cast announces both halves of a
	// split card with fuse, from hand, and the spell is both at once
	// (CR 702.102b). Materialised here, on the copy every gate below
	// reads, exactly as the face is: the cost gate sees both halves'
	// costs (CR 702.102c), the target gate both halves' clauses, and
	// the timing gate both halves' types.
	if params.Fuse {
		if err := fusedCastAllowed(card, src.Kind, params, grant); err != nil {
			slog.Warn("cast_spell rejected: fuse not allowed",
				"card_name", card.Name,
				"oracle_id", card.OracleID,
				"from_zone", src.Kind,
				"err", err,
			)
			return err
		}
		card.materialiseFused()
	}
	// S21 sub-PR 6: casting out of exile needs a live permission
	// naming this player. Checked before every other gate because
	// it's the one that decides whether the card is yours to touch at
	// all.
	//
	// #659 RETIRED the one exception S29 carved out here. A card whose
	// own text declared ZoneExile used to need no permission, on the
	// theory that suspend and foretell would use that shape. They do
	// not, and could not: a CARD-level declaration opens exile for
	// every copy of the card, at any time, however the copy got there
	// — so a Path to Exile'd Rift Bolt would be castable for free and
	// a Bojuka Bog'd Saw It Coming would be castable for its foretell
	// cost. Both keywords are per-INSTANCE permissions (ADR 0066), and
	// so is every other way a card gets cast out of exile in this
	// engine: impulse exile, airbend, warp, cascade, a defeated
	// Siege's back face. No catalog card ever declared ZoneExile, so
	// nothing shipped through the hole.
	//
	// The graveyard and the library reach the same question through
	// validateCastPathLocked below, which is handed `grant`; exile
	// keeps its own sentinel because ErrNoPlayPermission is what its
	// clients (the impulse button, the zone browser) already read.
	if src.Kind == ZoneExile && grant == nil {
		return ErrNoPlayPermission
	}
	// "You may CAST that card" (Ragavan) does not let you play a
	// land: playing a land is a special action, not a cast
	// (CR 305.1, 116.2a). True of a graveyard or library permission
	// too — Realmwalker casts creature spells and plays no lands.
	if grant != nil && grant.CastOnly && card.IsLand() {
		return ErrNoPlayPermission
	}
	// S22: an alternative cost is claimed at announce and replaces
	// the mana cost (CR 118.9). Resolved before every targeting gate
	// below, because overload and cleave rewrite the target clause —
	// the spell's legality has to be judged under the cost actually
	// being paid, not under the printed one.
	alt, err := g.resolveAlternativeCostLocked(playerID, card, src.Kind, grant, params.AlternativeCost, params.Targets)
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
	if err := g.validateCastPathLocked(card, src.Kind, alt, grant); err != nil {
		slog.Warn("cast_spell rejected: illegal cast path",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"from_zone", src.Kind,
			"alternative_cost", params.AlternativeCost,
			"err", err,
		)
		return err
	}
	// #1665: a HAND permission (miracle) opens one claim and nothing
	// else. From here down — the price, the X rule, the timing gate —
	// a cast that pays the printed cost reads no permission at all, so
	// a live miracle grant does not make a hard-cast sorcery an
	// instant. A no-op for every other zone. See CastPermission.ForClaim.
	grant = grant.ForClaim(alt)
	// #1729: a permission good for a set number of spells ("you may
	// cast A spell from among ...") is spent by a cast it is the reason
	// for. Decided here, of the card as it sits in its source zone and
	// under the claim just settled, and spent once the cast is made.
	spendsGrant := grant != nil && grant.CastsLeft > 0 && g.castUsesGrantLocked(card, src.Kind, alt)
	// #2173: the follow-up the permission carries runs when the cast is
	// made through it, decided on the same question as the spend above.
	followUp := CastFollowUpKey("")
	if grant != nil && grant.FollowUp != "" && g.castUsesGrantLocked(card, src.Kind, alt) {
		followUp = grant.FollowUp
	}
	grantSource := uuid.Nil
	if grant != nil {
		grantSource = grant.Source
	}
	grantCard := card
	// CR 708.4, ADR 0082 decision 2: the whole of "casting a card
	// face down" is this line, and where it sits is the decision.
	//
	// ABOVE it, two gates have to read the REAL card: the claim
	// resolution (does this card offer this key at all) and the path
	// validation (may the key be claimed from this zone). Stamping
	// before them would erase the very offer being claimed.
	//
	// BELOW it, every gate reads the CR 708.2 OBJECT, because
	// CatalogKey has gone silent for a face-down permanent (ADR 0069
	// decision 4) — no target clause, no modes, no additional or
	// optional costs, no tap cost, no cost modifiers of its own. And
	// `card` is a 2/2 creature that is not an instant, so the
	// sorcery-speed gate below demands sorcery timing, the cast tally
	// counts a creature spell, and the cast gate judges a colourless
	// creature spell with no name. CR 601.2b relative to 601.2c-f,
	// and nothing in this function forks on the fact.
	faceDown := FaceDownNone
	if alt != nil && alt.FaceDown != nil {
		faceDown = alt.FaceDown.Kind
		card.SetFaceDown(faceDown)
	}
	// CR 118.6: no mana cost is an unpayable cost, and paying it is
	// illegal, so a cast that would pay it is refused here. Checked
	// once the claimed alternative cost and the exile grant are both
	// known, because an alternative cost applied to the unpayable
	// cost may be paid (CR 118.6a).
	// Mode-independent, like the unparseable-cost refusal: there is
	// no cost to have paid on paper either.
	if HasNoManaCost(card) && castPaysPrintedCost(alt, grant) {
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
	// CR 107.3b: a spell with {X} in its mana cost, cast while paying
	// neither that cost nor an alternative cost that includes X, has
	// exactly one legal choice for X, and it is 0. Cascade's "{0}"
	// grant, a Siege's free cast and a free alternative cost are the
	// same answer to one question, asked here because 601.2b is where
	// X is announced and because the claimed alternative cost and the
	// exile grant are both settled by now.
	//
	// Refused rather than silently clamped, exactly as the
	// X-defined target count above is: a client that announces X=5
	// on a free Stroke of Genius is wrong about what it is casting,
	// and quietly casting a different spell hides that from whoever
	// has to debug it.
	if params.XValue != 0 &&
		CastCostFor(card, alt, grant).LocksXAtZero() {
		slog.Warn("cast_spell rejected: X must be 0 when the mana cost isn't paid",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"from_zone", src.Kind,
			"alternative_cost", params.AlternativeCost,
			"x_value", params.XValue,
		)
		return ErrInvalidParam
	}
	// ADR 0073, CR 601.2b: the optional additional costs the caster
	// chooses to pay — kicker, multikicker, buyback. Announced HERE,
	// with the modes and before the targets, for two reasons that
	// both matter: CR 601.2c comes after 601.2b and a kicked spell
	// may legally choose different targets from an unkicked one, and
	// every price this cast is about to be quoted (the tap budget,
	// the auto-tap plan, the strict-mana gate) has to include the
	// kicker or none of them agree.
	//
	// Only the CHOICE is checked here — in range, and no cost named
	// more times than it may be paid. The card and permanent halves
	// of the payment are validated with the mandatory additional
	// cost below, against one plan, by one validator.
	optionalCosts := OptionalCostsFor(CatalogKey(card))
	if err := validateOptionalCostChoice(optionalCosts, params.OptionalCosts); err != nil {
		slog.Warn("cast_spell rejected: bad optional additional cost choice",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"optional_costs_received", params.OptionalCosts,
			"optional_costs_offered", len(optionalCosts),
		)
		return err
	}
	// ADR 0089, CR 702.174a: a gift cost is paid by CHOOSING an
	// opponent, so its announcement is the index above plus a
	// recipient. Checked here with the rest of CR 601.2b, before the
	// targets, because the promise can change the target clause
	// (CR 702.174m).
	if err := g.validateGiftChoiceLocked(playerID, optionalCosts, params.OptionalCosts, params.GiftOpponent); err != nil {
		slog.Warn("cast_spell rejected: bad gift choice",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"gift_opponent", params.GiftOpponent,
			"optional_costs", params.OptionalCosts,
		)
		return err
	}
	// ADR 0100 §2, CR 601.2b: an either/or additional cost announces
	// WHICH branch is paid, here with the optional costs — before any
	// price is quoted, because the branch's mana is part of the total
	// (CR 601.2f). Required on a branched card, refused on any other,
	// and refused for a branch the caster cannot pay (CR 601.2h:
	// "Unpayable costs can't be paid"; CR 118.3), which is the same
	// predicate the view's `payable` stamp and the bot enumerator ask.
	addCost, err := ChosenAdditionalCost(AdditionalCostFor(CatalogKey(card)), params.CostBranch)
	if err == nil && params.CostBranch != nil && !g.AdditionalCostBranchPayableLocked(playerID, card, *params.CostBranch) {
		err = ErrCostBranch
	}
	if err != nil {
		slog.Warn("cast_spell rejected: bad either/or cost branch",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"cost_branch", params.CostBranch,
		)
		return err
	}
	// S20 sub-PR 4: modal spells — the chosen modes must be distinct,
	// in range and the right count (CR 601.2b, 700.2). #1590: the
	// count's bounds are read HERE, "as you cast this spell", for a
	// conditional mode count (Jeska's Will's "if you control a
	// commander … choose both instead") — and only here: the choice
	// lands on StackItem.Modes and nothing re-asks the condition, so a
	// commander that leaves in response changes nothing.
	//
	// #1655: AFTER the optional-cost choice above, because CR 601.2b
	// announces both in one step and "if this spell was kicked, choose
	// any number instead" (Inscription of Ruin) reads the kicker the
	// caster is announcing alongside the modes.
	modeSpec := ModeSpecFor(CatalogKey(card))
	modeMin, modeMax := g.modeBoundsLocked(modeSpec, ModeCountQuery{
		Chooser:       playerID,
		OracleID:      CatalogKey(card),
		OptionalCosts: params.OptionalCosts,
	})
	if err := validateModes(modeSpec, modeMin, modeMax, params.Modes); err != nil {
		slog.Warn("cast_spell rejected: bad mode choice",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"modes_received", params.Modes,
			"mode_min", modeMin,
			"mode_max", modeMax,
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
	spec, _ := castClauseSources(CatalogKey(card))
	// S22: the alternative cost gets the last word on the clause —
	// overload deletes it, cleave swaps a wider one in. Applied to
	// the card-level clause list; no card offers an alternative cost
	// AND modes, and the two would compose rather than conflict.
	spec = TargetSpecUnderAlternativeCost(spec, alt)
	// ADR 0089 §3: and a paid optional cost may swap it again — the
	// gift's "if the gift was promised, instead … target …" (CR
	// 702.174m). After the alternative cost, which no gift card
	// offers.
	spec = TargetSpecUnderOptionalCosts(spec, optionalCosts, params.OptionalCosts)
	// #764: the announcement's target STEPS — one per clause of the
	// card-level list, or one per clause of each chosen mode
	// occurrence (CR 608.2c). A card with neither keeps the S13.1
	// free-form behaviour; a modal card whose chosen modes take no
	// target must arrive with none.
	steps := AnnouncedClauses(spec, modeSpec, params.Modes)
	if len(steps) == 0 && modeSpec != nil && len(params.Targets) > 0 {
		return ErrInvalidParam
	}
	// ADR 0089 §3: a card whose ONLY clause is the one an unpaid
	// optional cost would add (Valley Rally's "if the gift was
	// promised, target creature you control") has structured
	// targeting all the same — the S13.1 free-form fallback is for
	// cards that declared nothing — so an unpromised cast that names
	// a target is refused (CR 702.174m).
	if len(steps) == 0 && len(params.Targets) > 0 && optionalCostsDeclareTargets(optionalCosts) {
		return ErrInvalidParam
	}
	// S22: "Exile X target creatures you control" — the clause's
	// count is the X announced at 601.2b, so resolve it into a
	// concrete Min / Max before anything validates against it. Over
	// the STEPS rather than the card-level spec because the clause
	// may belong to a MODE (Heliod's Intervention); the steps hold
	// clause copies, so nothing mutates the shared catalog entry.
	xSteps := resolveStepCountsFromX(steps, params.XValue)
	// #1559: "with mana value X or less" — X is announced before
	// targets (CR 601.2b / 602.2b), so the bound is known here.
	bindStepsX(steps, params.XValue)
	// #1657, CR 601.2d: a divided amount read off the board or off the
	// claimed alternative cost ("X if its madness cost was paid") is
	// fixed here, with the targets, and never re-read.
	g.bindDivideAmountsLocked(steps, DivideAmountArgs{
		Controller: playerID, Source: cardID, AltCost: params.AlternativeCost,
	})
	params.Targets = assignAnnouncedSlots(steps, params.Targets)
	for _, i := range xSteps {
		// Max 0 reads as "unbounded" to the ordinary count check, so
		// an X-counted step is checked for an EXACT count of X here.
		if n, bad := xCountMismatch(steps[i], params.Targets, params.XValue); bad {
			slog.Warn("cast_spell rejected: X-defined target count mismatch",
				"card_name", card.Name,
				"oracle_id", card.OracleID,
				"x_value", params.XValue,
				"targets_received", n,
			)
			return ErrInvalidParam
		}
	}
	// #760, ADR 0073 §7, CR 101.2: the one announce-time cast gate,
	// at the point ADR 0066 named. Every CR 601.2b choice is settled
	// by here — the face, the source zone, the permission, the
	// claimed alternative cost, X, the modes and the optional costs —
	// and nothing has been PAID, so a refused cast leaves the card
	// exactly where it was and costs nothing. That is also what lets
	// CR 601.3a work: a choice made while proposing the spell can
	// lift a ban, and every such choice is on `params` by now.
	//
	// Before the CR 601.2c target check rather than after it, because
	// a banned cast should be refused for the ban rather than for
	// whatever the targeting gate would have said about a spell that
	// was never going to be cast.
	//
	// "Can't beats may" needs no rule of its own here. Cascade, a
	// granted permission and an impulse grant all reach CastSpell, so
	// a free cast passes through this gate like any other.
	//
	// #1439: a LAND is not gated here at all. Playing a land is a
	// special action (CR 305.1, CR 116.2a), not a cast, and every
	// clause CastGateLocked enforces is written about casting (its
	// own doc says so). The land branch further down runs its own
	// CR 305 checks; asking this gate first made Rule of Law and
	// Grafdigger's Cage — both spell-only restrictions — refuse a
	// land drop from hand or from the graveyard.
	if !card.IsLand() {
		if err := g.CastGateLocked(playerID, card, src.Kind, params); err != nil {
			slog.Warn("cast_spell rejected: an effect prevents this cast",
				"card_name", card.Name,
				"oracle_id", card.OracleID,
				"from_zone", src.Kind,
				"err", err,
			)
			return err
		}
	}
	// CR 702.16b: the source of a SPELL is the spell itself, so the
	// quality protection is tested against is the card's own colour
	// and type — not its caster's (#662).
	// CR 202.3e: the X announced with the cast counts in the spell's mana
	// value, which "protection from mana value N or less" reads (#2181).
	castSrc := SourceObject(playerID, &card)
	castSrc.X = params.XValue
	if err := g.validateAnnouncedTargetsLocked(castSrc, steps, params.Targets); err != nil {
		slog.Warn("cast_spell rejected: illegal target",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"targets_received", len(params.Targets),
			"x_value", params.XValue,
			"err", err,
		)
		return err
	}
	// #1563, CR 601.2d: the division is announced with the targets,
	// against the amount the announced X gives. The settled copy — a
	// lone target handed the whole amount — is what the item stores.
	dist, err := settleDistribution(steps, params.Targets, params.Distribution, params.XValue)
	if err != nil {
		slog.Warn("cast_spell rejected: bad division",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"x_value", params.XValue,
			"err", err,
		)
		return err
	}
	params.Distribution = dist
	// S21 sub-PR 5: additional costs (CR 601.2f). Validated here,
	// with the rest of the announce-time choices, and paid further
	// down once the spell is on the stack — validate-all-then-pay,
	// so a rejected cast never leaves a card in the graveyard.
	//
	// ADR 0073: the mandatory cost and the announced optional ones
	// are ONE plan — mandatory first, then each chosen optional cost
	// in index order, once per payment — so there is still exactly
	// one validator and one payer, and the flat discard / sacrifice
	// lists are walked in an order the client can reproduce.
	//
	// ADR 0100 §2: `addCost` is the branch the announcement chose for
	// an either/or cost (settled above), so the validator and the payer
	// below never learn that the card had a choice at all.
	costPlan := castCostPayments(addCost, optionalCosts, params.OptionalCosts)
	// CR 702.120a: escalate's non-mana payments, one per mode beyond the
	// first, join the same plan so the one validator and the one payer
	// see them (#2126).
	costPlan = append(costPlan, escalatePayments(modeSpec, params.Modes)...)
	if err := g.validateAdditionalCostLocked(playerID, cardID, costPlan, params.DiscardIDs, params.SacrificeIDs, params.XValue); err != nil {
		slog.Warn("cast_spell rejected: bad additional cost payment",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"discards_received", len(params.DiscardIDs),
			"sacrifices_received", len(params.SacrificeIDs),
			"optional_costs", params.OptionalCosts,
			"err", err,
		)
		return err
	}
	// #1727, CR 118.3: one permanent pays one sacrifice. The
	// alternative cost's sacrifice (AltCostIDs) and the additional
	// cost's (SacrificeIDs) are validated separately above, each
	// against its own clause, so a creature named to both would pass
	// both and then be sacrificed once and found missing the second
	// time — after the spell was already on the stack.
	if alt != nil && alt.Sacrifice != nil && sharesAnID(params.AltCostIDs, params.SacrificeIDs) {
		slog.Warn("cast_spell rejected: one permanent named to two sacrifice costs",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
		)
		return ErrInvalidParam
	}
	// ADR 0135 §1, CR 118.3 and 701.26a: one permanent is tapped once. A
	// creature tapped for a tap alternative cost (AltCostIDs) can't also
	// be tapped for convoke or waterbend (TapIDs) or for teamwork or
	// escalate (TeamworkIDs): each list is validated on its own as
	// untapped, so a creature named to two would pass both.
	if alt != nil && alt.TapOthers != nil && (sharesAnID(params.AltCostIDs, params.TapIDs) || sharesAnID(params.AltCostIDs, params.TeamworkIDs)) {
		slog.Warn("cast_spell rejected: one permanent named to two tap costs",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
		)
		return ErrInvalidParam
	}
	// ADR 0129 §5, CR 118.3: the energy the whole cast pays — a claimed
	// alternative cost's and the plan's (replicate, once per payment) —
	// against the caster's total, summed so the two cannot each pass
	// against all of it. The alternative cost's own share was checked
	// with its offer above. Never waived (ADR 0129 §4).
	energyOwed := planEnergy(costPlan)
	if alt != nil {
		energyOwed += alt.Energy
	}
	if err := EnergyShortfall(p, energyOwed); err != nil {
		slog.Warn("cast_spell rejected: not enough energy",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"energy_owed", energyOwed,
			"err", err,
		)
		return err
	}
	// #1703: the two components whose payment names creatures on
	// the board — teamwork's taps (CR 702.194a) and blight's one
	// creature (CR 701.68a). Same plan, same validate-all-then-pay.
	if err := g.validateTeamworkLocked(playerID, costPlan, params.TeamworkIDs, params.TapIDs); err != nil {
		slog.Warn("cast_spell rejected: bad teamwork payment",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"teamwork_received", len(params.TeamworkIDs),
			"err", err,
		)
		return err
	}
	if err := g.validateBlightLocked(playerID, costPlan, params.BlightIDs, params.XValue); err != nil {
		slog.Warn("cast_spell rejected: bad blight payment",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"blight_received", len(params.BlightIDs),
			"err", err,
		)
		return err
	}
	if err := g.validateRevealLocked(playerID, cardID, costPlan, params.RevealIDs); err != nil {
		slog.Warn("cast_spell rejected: bad reveal payment",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"reveal_received", len(params.RevealIDs),
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
		if base, _, berr := g.printedCostLocked(p, card, params); berr == nil {
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
	// ADR 0100 §1, CR 702.66a: the graveyard cards named to delve.
	// Checked with the other announce-time choices and exiled further
	// down once the spell is on the stack. The budget is the generic
	// the cast still owes after the modifiers and the taps — the same
	// number CastPrice.DelveBudget reports to the preview and the bot.
	if len(params.DelveIDs) > 0 {
		budget := g.delveBudgetLocked(p, card, params)
		if err := g.validateDelveLocked(playerID, cardID, card, params.DelveIDs, params.AltCostIDs, budget); err != nil {
			slog.Warn("cast_spell rejected: bad delve payment",
				"card_name", card.Name,
				"oracle_id", card.OracleID,
				"delve_received", len(params.DelveIDs),
				"budget", budget,
				"err", err,
			)
			return err
		}
	}
	// Timing (CR 307.1). ONE read, in cast_timing.go, shared with the
	// bot enumerator and the view so the three cannot disagree about
	// when a cast is open (#1195, ADR 0066's 2026-09-22 amendment).
	//
	// It folds four things in CR 101.2's order: the card's own timing
	// (instant, or flash — CR 702.8 — which off-battlefield
	// HasKeyword answers from CatalogPrintedKeywords), the granted
	// permission's override (ADR 0066 Decision 6: Bolas's Citadel
	// does NOT make a sorcery on top of your library castable on an
	// opponent's turn; madness's TimingFlash does), the per-player
	// grants (Vedalken Orrery, Leyline of Anticipation, Emergence
	// Zone) and, last, the per-player restrictions (Teferi, Time
	// Raveler), because "can't" beats "can".
	//
	// LANDS KEEP THEIR OWN BRANCH. Playing a land is a special action
	// (CR 116.2a) and not a cast, so no timing STATEMENT reaches it —
	// an Orrery does not open a land drop and a Dosan does not close
	// one. CR 305's "during your main phase, when the stack is empty"
	// is the whole of its window.
	if card.IsLand() {
		if !g.SorcerySpeedOpenLocked(playerID) {
			return ErrSorcerySpeedRequired
		}
	} else if !g.CastTimingForOfferOpenLocked(playerID, card, src.Kind, grant, alt) {
		return ErrSorcerySpeedRequired
	}
	// Lands skip the stack entirely (CR 305). Move the card to the
	// battlefield and stamp the controller — same shape as PlayCard.
	if card.IsLand() {
		// #500: CR 305.2's land-play allowance, enforced. Owner
		// decision: the sandbox posture on land drops is over — the
		// tally has existed since S31 sub-PR 1 and only the
		// legal-move enumerator ever read it, so any client that did
		// not consult the enumerator could play the whole hand as
		// lands.
		//
		// The gate sits here rather than in the enumerator (which
		// keeps its own check, so a bot is never offered a move the
		// engine will refuse) and before the face is stamped into the
		// source zone, so a refused play leaves the card in hand
		// exactly as it was. The paused-entry path
		// (executeEntryToBattlefieldLocked, the shockland's "pay 2
		// life") is the tail of a play that already passed this gate,
		// so it needs no second check.
		//
		// The allowance is not a literal 1 — see land_drops.go.
		//
		// ADR 0109 §4, CR 101.2: "can't" beats "can", so a "players
		// can't play lands" effect is asked BEFORE the drop count, and an
		// extra drop does not lift it. The same gate the enumerator and
		// the view ask (land_play_gate.go).
		if err := g.LandPlayGateLocked(playerID, card, src.Kind); err != nil {
			slog.Warn("cast_spell rejected: an effect forbids playing this land",
				"card_name", card.Name,
				"oracle_id", card.OracleID,
				"reason", err.Error(),
			)
			return err
		}
		if g.LandDropsRemainingLocked(playerID) <= 0 {
			slog.Warn("cast_spell rejected: no land plays left this turn",
				"card_name", card.Name,
				"oracle_id", card.OracleID,
				"lands_played", g.LandsPlayedThisTurnFor(playerID),
				"allowance", g.EffectiveLandDropsLocked(p),
			)
			return ErrLandDropUnavailable
		}
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
			// CR 305.2: this is the one entry that spends the turn's
			// land drop, and since #478 it says so rather than letting
			// the resume infer it from "a land with no stack item" —
			// a fetched, reanimated or blinked land can pause there now
			// and was never PLAYED.
			landPlay: true,
			// ADR 0102: whose land drop this spends, kept apart from
			// Actor, which an entry-controller effect may rewrite.
			landPlayer: playerID,
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
			g.restoreEntryControllerLocked(ev)
			setFaceInZoneLocked(src, cardID, wasFace)
			return nil
		}

		// #1326 (following #653's precedent at the stack-resolution
		// site above): THE push, rather than a second copy of it.
		// This branch used to reproduce executeEntryToBattlefieldLocked
		// inline — the move, the controller, the tapped stamp, the
		// counters, the land-drop tally, the zone move and the ETB
		// hook — while the RESUME of the very same event (the
		// shockland's "pay 2 life") went through the finisher. Two
		// copies of "a land enters the battlefield" is how the
		// land-drop tally (#478) and now the entry's Played marker
		// (#1326, CR 305.4) could apply to a PAUSED land and silently
		// not to an ordinary one — the common case. ev.landPlay is
		// still true on `out` (applyReplacementsLocked only settles
		// the replacement chain; it never clears the caller's own
		// fields), so the finisher bumps LandsPlayedThisTurn and
		// stamps Event.Played exactly as the resume path does.
		if _, err := g.executeEntryToBattlefieldLocked(out); err != nil {
			setFaceInZoneLocked(src, cardID, wasFace)
			return err
		}
		// Playing a land is a special action (CR 116.2a); the player
		// keeps priority and CR 117.5 drains any landfall-style
		// triggers onto the stack here rather than at the next wrap.
		g.runStateChecksLocked()
		return nil
	}
	// #1397: every card this cast's costs are about to move — the
	// additional cost's discards and sacrifices, the alternative
	// cost's pitched, returned or escaped cards — asked about BEFORE
	// anything is tapped or paid. A commander among them whose owner
	// has not answered CR 903.9 parks the cast on that question, with
	// the card still where it was; the answer casts it again. See
	// cost_commander_choice.go.
	moving := append(append(append([]uuid.UUID(nil), params.DiscardIDs...), params.SacrificeIDs...), params.AltCostIDs...)
	// ADR 0100: and the graveyard cards delve exiles — a delved
	// commander is asked about CR 903.9 before anything is paid.
	moving = append(moving, params.DelveIDs...)
	// #1445 / #1427: a card an EFFECT has already paused on its way
	// out cannot pay — moved, or tapped to convoke / waterbend
	// (TapIDs), tapped to teamwork or blighted (#1703). See
	// refusePausedCostCardsLocked.
	if err := g.refusePausedCostCardsLocked(moving, params.TapIDs, params.TeamworkIDs, params.BlightIDs); err != nil {
		return err
	}
	// ADR 0115: of those, only a card the payment puts into a hand or
	// a library is asked first (CR 903.9b), and the one cast cost that
	// does is the alternative cost's return to hand (Daze, Gush). A
	// discarded, sacrificed, pitched, escaped or delved commander is
	// paid like any other card and offered the command zone afterwards
	// by the CR 903.9a state-based action.
	var asking []costCommanderMove
	if alt != nil && alt.ReturnToHand != nil {
		asking = costCommanderMovesTo(ZoneHand, params.AltCostIDs...)
	}
	asked, answers := g.askCostCommanderLocked(playerID, asking, params.commanderAnswers, card.Name,
		func(g *Game, answers map[uuid.UUID]bool) error {
			again := params
			again.commanderAnswers = answers
			return g.castSpellLocked(playerID, cardID, again)
		})
	if asked {
		return nil
	}
	params.commanderAnswers = answers

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
	paid, err := g.applyCastCostLocked(p, card, params, cardID)
	if err != nil {
		return err
	}
	// Non-land: route through the stack. The card lives in
	// Game.Stack; the announce-time choices live in StackMeta.
	if _, err := MoveCard(src, g.Stack, cardID); err != nil {
		return err
	}
	for i := range g.Stack.Cards {
		if g.Stack.Cards[i].InstanceID == cardID {
			// ADR 0034: the spell on the stack IS the chosen face.
			// Everything downstream of here reads the card out of
			// the stack zone rather than from the local copy — the
			// resolution target re-check, the effect resolver, the
			// graveyard route, and the client's stack overlay — so
			// the face has to be stamped on the real card, not just
			// the copy the announce gates were judged against.
			g.Stack.Cards[i].SetFace(params.Face)
			// ADR 0103, CR 702.102b: a fused spell is both halves at
			// once, with their combined characteristics.
			if params.Fuse {
				g.Stack.Cards[i].materialiseFused()
			}
			// CR 601.2a: the player who casts a spell becomes its
			// controller — and the CARD on the stack says so, not just
			// the StackItem (ADR 0104, #1745 finding 1). A card carries
			// its owner as its controller from the deck load, so a
			// spell cast off another player's card (Gonti, a Ragavan
			// impulse, Wrexial) used to answer "target spell you don't
			// control" by its owner. The face-down viewers rule
			// (CR 708.5) reads it too, and needs it before the landing
			// below asks.
			g.Stack.Cards[i].Controller = playerID
		}
	}
	if faceDown != FaceDownNone {
		// CR 708.4: a spell cast face down is a CR 708.2 object on
		// the stack — a 2/2 creature spell with no name and no text —
		// and its CONTROLLER is the only player who may look at it
		// (CR 708.5). This REPLACES the public marking below rather
		// than adding to it: the stack is a public zone and this is
		// not a public object, which is the one line that separates
		// the two. Same writer as the face-down exile route and the
		// face-down battlefield entry (ADR 0069 decision 2).
		g.applyFaceDownLandingLocked(g.Stack, cardID, faceDown, nil)
	} else {
		// S13.5: cast spells are public on the stack.
		g.markCardKnownInZoneLocked(g.Stack, cardID)
	}
	if g.StackMeta == nil {
		g.StackMeta = make(map[uuid.UUID]*StackItem)
	}
	// CR 702.61a (#1519): the spell's own split second, or the
	// sandbox flag — one writer for both, so the cache below and
	// recomputeSplitSecondLocked can never disagree about where the
	// fact came from. Read off the announce copy, which carries the
	// chosen face and, for a face-down cast, has no text at all.
	splitSecond := castHasSplitSecond(&card, faceDown, params.SplitSecond)
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
		SplitSecond:  splitSecond,
		AltCost:      params.AlternativeCost,
		// CR 702.143c, #658: a spell cast from a foretold card is a
		// foretold spell, whatever cost paid for it. Read off the
		// value copy taken out of the source zone above, which is the
		// last object that still carries the face-down state —
		// CR 406.3a turns the card face up as it is cast, and
		// MoveCard has already done so by here.
		Foretold: foretold,
		// CR 708.4, ADR 0082 decision 3: the permanent this spell
		// becomes enters FACE DOWN. Carried on the item because the
		// permanent is a new object and MoveCard clears the state on
		// every zone change — stack resolution seeds it onto the
		// entry event from here.
		FaceDown: faceDown,
		// ADR 0090, CR 722.3c / 707.12: the prepare spell is cast as a
		// COPY, and a copy is not a card — so it takes CR 707.10's exits
		// off the stack (it ceases to exist rather than landing in a
		// graveyard), which is the path Twincast's copies already use.
		IsCopy: card.PrepareCopy,
		// CR 702.34a / CR 400.7g, ADR 0066. The fact travels with the
		// stack object because the catalog cannot answer for it: a
		// card given flashback by Snapcaster was cast for a cost the
		// catalog has never heard of, and the permission that granted
		// it may be gone by the time the spell leaves the stack.
		// ADR 0103, CR 702.127a: an aftermath half cast from a
		// graveyard is exiled instead of going anywhere else as it
		// leaves the stack, however it leaves — the same replacement
		// flashback's cost carries.
		AltCostExiles: (alt != nil && alt.ExileOnLeavingStack) || (src.Kind == ZoneGraveyard && isAftermathHalf(card)),
		CastFromZone:  src.Kind,
		// #761: the mana that actually paid, or the fact that the
		// engine waived the charge. Stamped here for the reason
		// XValue is: by resolution the tokens are gone from the pool
		// and the Treasure that made one may be in a graveyard, so
		// nothing downstream could recompute it.
		//
		// #664: and which optional additional costs were paid, for
		// exactly the same reason — by resolution the mana is spent,
		// the sacrificed land is in a graveyard, and the catalog
		// cannot say whether a choice was taken. Normalised to
		// ascending order by castCostPayments' own walk, so two
		// clients that announce the same multikicker in different
		// orders produce the same record.
		// #1213: and how many permanents the additional cost
		// sacrificed, for the third time the same reason — by
		// resolution they are in graveyards. Since ADR 0100 §3 a
		// cast's clause may be variable ("sacrifice any number of
		// creatures", "sacrifice X lands"), and this record is the
		// only place the count lives: Vicious Betrayal's "+2/+2 for
		// each creature sacrificed this way" reads it through
		// ctx.Sacrificed(). ADR 0113 §1 (#2072): and WHICH ones, named
		// here while they are still on the battlefield (the payment
		// below moves them) — Fling's "the sacrificed creature's power"
		// reads them through ctx.SacrificedPermanents().
		//
		// ADR 0100 §2: and which either/or branch was paid and which
		// cards the additional cost discarded — Grab the Prize's "if
		// the discarded card wasn't a land card".
		//
		// ADR 0135 §4: and the objects the alternative cost's card
		// component pays with, named while they are still where they were
		// — Adipose Offspring's "the sacrificed creature's toughness".
		Paid: paidWithAltCostObjects(paidWithBranchAndDiscards(
			paidWithGift(paidWithSacrifices(paidWithOptionalCosts(paid, costPlan), g.sacrificeRefsLocked(params.SacrificeIDs)), params.GiftOpponent),
			params.CostBranch, params.DiscardIDs), g.sacrificeRefsLocked(params.AltCostIDs)),
		Seq: g.nextStackSeqLocked(),
		// S20: remember the clause the targets were validated under so
		// the resolution re-check and per-slot effect checks use it.
		// #764: and the ModeSpec, so a per-mode target group can be
		// resolved back to the clause it answered.
		targetSpec: spec,
		modeSpec:   modeSpec,
	}
	// #1547, CR 601.2h: the mana is spent, so what it does when it is
	// spent happens now — Cavern of Souls' "that spell can't be
	// countered", Pyromancer's Goggles' copy trigger. Against the SAME
	// spend context applyCastCostLocked solved the payment under, so a
	// rider's filter and the restriction that let the token pay can
	// never disagree about what the spell is. A trigger queued here is
	// placed by the runStateChecksLocked at the bottom of the cast,
	// above the spell.
	g.applyManaSpendRidersLocked(g.StackMeta[cardID], ManaSpendForCastFrom(card, src.Kind), card)
	// CR 702.62a (#659): the permanent this cast produces has haste.
	// Registered here rather than at resolution because the grant that
	// says so has been consumed by now — the card has left exile and
	// the permission no longer covers the object. The static matches
	// the instance and its controller, so it does nothing while the
	// spell is on the stack and everything the moment it lands.
	if grantsHaste {
		g.grantHasteForCastLocked(playerID, cardID)
	}
	// CR 601.2h: pay the costs. The mana component was charged
	// above (pre-move, as S15 wrote it); the additional cost is
	// paid HERE, with the spell already on the stack, because a
	// discard payoff — or an aristocrats payoff watching the
	// sacrifice — that triggers off it must resolve before the
	// spell does. Validation happened at announce, so a failure
	// past this point is an engine bug rather than a bad request.
	// The life the whole plan pays — a "pay X life" at the announced X,
	// and a chosen branch's fixed "pay 3 life" (ADR 0100 §2) — the same
	// sum the validator checked against CR 119.4.
	payLife := planLife(costPlan, params.XValue)
	if err := g.payAdditionalCostLocked(playerID, params.DiscardIDs, params.SacrificeIDs, payLife, params.commanderAnswers); err != nil {
		slog.Error("cast_spell: additional cost failed after validation",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"err", err,
		)
		return err
	}
	// ADR 0129 §5: the plan's energy (replicate, once per payment),
	// through the one path that pays energy, checked above with the
	// alternative cost's against the caster's total.
	if err := g.payEnergyLocked(playerID, planEnergy(costPlan), cardID); err != nil {
		slog.Error("cast_spell: additional cost energy failed after validation",
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
	if err := g.payAlternativeCostLocked(playerID, cardID, alt, params.AltCostIDs, params.commanderAnswers); err != nil {
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
	// ADR 0100 §1: delve's exiles, in the same window and for the same
	// reason as escape's — with the spell already on the stack, so a
	// "whenever a card leaves your graveyard" watcher triggers above
	// it. The record names the objects that landed in exile, which is
	// what CR 607.2q links the permanent to.
	if len(params.DelveIDs) > 0 {
		delved, err := g.payDelveLocked(params.DelveIDs, params.commanderAnswers)
		if item := g.StackMeta[cardID]; item != nil {
			item.Paid.Delved = delved
		}
		if err != nil {
			slog.Error("cast_spell: delve failed after validation",
				"card_name", card.Name,
				"oracle_id", card.OracleID,
				"err", err,
			)
			return err
		}
	}
	// #1703: teamwork's taps and blight's counters, in the same
	// window and for the same reason. A creature the blight kills
	// dies at the closing state-based check, with the cost paid.
	g.payTeamworkLocked(playerID, params.TeamworkIDs)
	if err := g.payBlightLocked(playerID, costPlan, params.BlightIDs, params.XValue); err != nil {
		slog.Error("cast_spell: blight cost failed after validation",
			"card_name", card.Name,
			"oracle_id", card.OracleID,
			"err", err,
		)
		return err
	}
	// ADR 0100 amendment 2026-10-07: a revealed card is shown with the
	// spell on the stack, so the table sees the Elf the Vanquisher paid
	// with. Nothing moves.
	g.payRevealLocked(playerID, cardID, costPlan, params.RevealIDs)
	if splitSecond {
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
	// ADR 0108 §9: the colours of the instants and sorceries cast this
	// turn (Refraction Trap), read off the spell as it is on the stack.
	for i := range g.Stack.Cards {
		if sc := &g.Stack.Cards[i]; sc.InstanceID == cardID {
			ch := SourceCharacteristics(sc)
			if characteristicHasType(ch, "Instant") || characteristicHasType(ch, "Sorcery") {
				tally = tally.withInstantSorceryColors(ch.Colors)
			}
			// Artificer Class's "first artifact spell you cast each turn".
			if characteristicHasType(ch, "Artifact") {
				tally.Artifact++
			}
			// #2743: the running mana value, with the X this spell was
			// cast with. SourceCharacteristics already priced it that
			// way, face down included.
			if mv, ok := g.castManaValueLocked(sc); ok {
				tally.ManaValue += mv
			}
			break
		}
	}
	g.SpellsCastThisTurn[playerID] = tally
	// CR 722.3c / 601.2i (ADR 0090): "that permanent loses the prepared
	// designation at the time the spell becomes cast" — here, with
	// every cost paid and just before EventCast announces the cast.
	// Read off the value copy taken out of exile, which still names the
	// permanent; MoveCard dropped the link on the copy itself.
	if card.PrepareCopy {
		if perm := findBattlefieldCard(g, card.PreparedBy.ID); perm != nil && perm.ObjectEpoch == card.PreparedBy.Epoch {
			g.unprepareLocked(perm)
		}
	}
	// CR 601.2i (ADR 0106 §4 decision 3): the spell becomes cast here,
	// which is the moment "the next spell you cast this turn can't be
	// countered" is decided. Every live promise of the caster's that
	// this spell matches is spent on it and becomes a mark on its item.
	g.spendCounterShieldPromisesLocked(playerID, cardID)
	// #1852: the same moment decides "the next <kind> spell you cast"
	// promises about flash, price, an extra counter and uncounterability.
	promiseFollowUps := g.spendNextSpellPromisesLocked(playerID, cardID)
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
		// ADR 0118 §2: a forced cast is announced as unpaid, so the
		// log tells the table. ForceCast waives the whole mana gate
		// (applyCastCostLocked), so the flag alone decides it.
		Unpaid: params.ForceCast,
	})
	// CR 115.3: the objects named at 601.2c have now become targets.
	// Emitted after EventCast so a "becomes the target" trigger and
	// a "whenever a player casts a spell" trigger queue in printed
	// order.
	g.emitBecameTargetLocked(playerID, cardID, cardID, params.Targets)
	// ADR 0099 §5: a cast that used a pass-closed grant (discover,
	// cascade) spends it, so the caster's next pass does not lapse a
	// card that is already on the stack — and a discover grant's cast
	// is what completes the discover (CR 701.57b), announced here,
	// after EventCast, so "whenever you discover" goes on the stack
	// above the discovered spell. `card` is still the exile object the
	// grant named: the copy was taken before the move bumped its epoch.
	if grant != nil && grant.LapseOnPass != "" {
		g.consumePassClosedGrantLocked(playerID, card)
	}
	if spendsGrant {
		g.consumeLimitedGrantLocked(playerID, grantCard, grant)
	}
	for _, key := range promiseFollowUps {
		if err := g.runCastFollowUpLocked(key, CastFollowUp{Player: playerID, Spell: cardID}); err != nil {
			return err
		}
	}
	if followUp != "" {
		f := CastFollowUp{Player: playerID, Source: grantSource, Spell: cardID}
		if err := g.runCastFollowUpLocked(followUp, f); err != nil {
			return err
		}
	}
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
func (g *Game) applyCastCostLocked(p *Player, card Card, params CastSpellParams, cardID uuid.UUID) (PaidCost, error) {
	var paid PaidCost
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
		return paid, fmt.Errorf("%w for %s: %w", ErrUnparseableCost, card.Name, err)
	}
	// #1600, CR 609.4b: "you may spend mana as though it were mana of
	// any color" (Chromatic Orrery) widens what may pay the cost — read
	// here, after convoke and delve have taken their share and before
	// the Phyrexian strike, exactly as applyAutoTapLocked reads it.
	cost = g.costAsPaidByLocked(p.ID, ManaSpendForCastParams(card, params), cost, params.XValue)
	// CR 107.4 / CR 601.2b: the Phyrexian symbols the caster announced
	// they are paying with life leave the mana cost here, and the life
	// is paid below — after the mana half is known to be payable, so a
	// rejected cast never costs a point. Validated in every mode,
	// because an over-claim is a malformed announce rather than a
	// mana-gate failure.
	cost, phyrexianLife, err := g.strikePhyrexianLifeLocked(p, card.Name, cost, ManaSpendForCastParams(card, params), params.PhyrexianLife, &paid)
	if err != nil {
		return paid, err
	}
	if !params.Strict || params.ForceCast {
		// Permissive default OR strict-mode override. Don't touch
		// the pool; just emit a warning so the client can render
		// the toast / event-log breadcrumb.
		//
		// The LIFE half is still paid: a life total is engine state
		// in every mode, and permissive mode's bargain is only about
		// the mana the player tracks on paper.
		g.EmitEvent(Event{
			Kind:   EventCostWarning,
			Actor:  p.ID,
			Source: cardID,
		})
		// #761: the engine did not charge, so it has no record of
		// WHAT was paid — and says so, rather than leaving an empty
		// record that reads as "nothing was spent". Every reader
		// treats OnPaper as the weaker-than-printed answer: a Vexing
		// Bauble does not counter this spell, and a converge spell
		// counts no colours.
		paid.OnPaper = true
		return paid, g.payPhyrexianLifeLocked(cardID, p.ID, phyrexianLife)
	}
	// S32 (#352): the spend context is what lets restricted mana pay
	// — and what stops it paying for the wrong thing. Ancient
	// Ziggurat's {G} funds a creature spell here and is invisible to
	// a Lightning Bolt.
	spendCtx := ManaSpendForCastParams(card, params)
	if !p.ManaPool.CanPayFor(cost, params.XValue, spendCtx) {
		return paid, &InsufficientManaError{Missing: p.ManaPool.MissingFor(cost, params.XValue, spendCtx)}
	}
	// CR 601.2h pays every component together, so the order here is
	// an engine-safety choice, not a rules one: the FALLIBLE half
	// goes first. PayLifeForEffect can still refuse after the life
	// total was checked — a replacement that stops the player losing
	// life at all (CR 119.8, #808) — and refusing after the pool was
	// emptied would charge for a cast that did not happen. The
	// reverse cannot bite: nothing a life-loss replacement may do
	// under mustSettleNow reaches a mana pool, so the CanPayFor above
	// still holds when SpendManaFor runs.
	if err := g.payPhyrexianLifeLocked(cardID, p.ID, phyrexianLife); err != nil {
		return paid, err
	}
	// Payable under strict mode — commit the spend.
	//
	// #761: which tokens pay is the spell's business when it reads
	// them back. A converge or sunburst card declares
	// WantsDistinctColors and the solver spreads the generic half
	// across colours it has not spent; everything else keeps the
	// colourless-first order that saves coloured mana for the next
	// cast. The strategy can never change whether the cost is
	// payable — see attemptSpend.
	spent, _ := p.ManaPool.SpendManaForWith(cost, params.XValue, spendCtx, spendStrategyForCast(card))
	paid.Mana = spent
	g.EmitEvent(manaSpentEvent(p.ID, cardID, spent))
	return paid, nil
}

// spendStrategyForCast asks the catalog whether this spell reads the
// colours that paid for it (#761). Converge (CR 702.86) and sunburst
// (CR 702.44) do; everything else keeps the default order.
//
// A DECLARATION on the spec rather than something inferred from the
// oracle text, for the reason DerivesFromOtherSources is one: the
// alternative is a text scan that quietly stops matching when a card
// words the clause differently.
func spendStrategyForCast(card Card) ManaSpendStrategy {
	if CatalogWantsDistinctColors == nil {
		return SpendPreserveColors
	}
	if CatalogWantsDistinctColors(CatalogKey(card)) {
		return SpendDistinctColors
	}
	return SpendPreserveColors
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
//  4. Run autoTapTopUpLocked to get a plan for what the pool is
//     missing (ADR 0118 §1); if no plan exists, return a
//     structured InsufficientManaError keyed off the current
//     pool's missing list (the auto-tapper itself doesn't carry
//     a missing-symbols breakdown).
//  5. Materialise the plan: for each card, tap it and drop its
//     produced mana into the pool with greedy color-picking
//     against the shortfall the plan was built to pay.
//
// Caller must hold g.mu (CastSpell holds the write lock).
func (g *Game) applyAutoTapLocked(p *Player, card Card, params CastSpellParams) error {
	cost, err := g.effectiveCostLocked(p, card, params)
	if err != nil {
		return nil
	}
	// #1600: the cost as this caster may pay it — the same widening
	// applyCastCostLocked will pay under, so the plan funds exactly
	// what the payment accepts.
	cost = g.costAsPaidByLocked(p.ID, ManaSpendForCastParams(card, params), cost, params.XValue)
	// The Phyrexian symbols being paid with life are not the
	// auto-tapper's business: tapping a land for a pip the caster
	// announced they would pay with 2 life is exactly the stranding
	// the unparseable-cost short-circuit above avoids. Same
	// reduction applyCastCostLocked will make. A malformed claim
	// short-circuits the same way a bad parse does — it rejects there,
	// with the error, rather than here with lands already tapped.
	//
	// KNOWN LIMIT, declared rather than fixed: the strike order reads
	// the pool, and here the pool is usually empty, so a cost printing
	// two Phyrexian symbols of DIFFERENT colours has them rank equal
	// and the first is struck. If the board could only have funded the
	// other choice the plan fails and the cast reports insufficient
	// mana. Letting the planner choose is a solver change; no printed
	// card has two Phyrexian symbols of different colours, so nothing
	// can reach it today. Once the plan exists the two agree: it funds
	// every symbol it did not strike, so the payment's own pass ranks
	// the struck one unpayable and strikes it again.
	cost, _, err = g.strikePhyrexianLifeLocked(p, card.Name, cost, ManaSpendForCastParams(card, params), params.PhyrexianLife, nil)
	if err != nil {
		return nil
	}
	// Same context applyCastCostLocked will pay under, so the
	// "already funded, skip planning" shortcut can't be fooled by
	// restricted mana this cast cannot legally spend. The top-up below
	// takes the same shortcut; this one keeps the exclusion list from
	// being built for a cast that needs no plan.
	spendCtx := ManaSpendForCastParams(card, params)
	if p.ManaPool.CanPayFor(cost, params.XValue, spendCtx) {
		return nil
	}
	// The lock-tap reservations, and everything this announcement has
	// already spent: S22's convoke / waterbend taps (without them the
	// planner would tap the same Birds of Paradise the caster just
	// convoked) and, #1242, the permanents and cards named to the
	// additional cost — auto-tap runs BEFORE payAdditionalCostLocked,
	// so a plan could otherwise crack the Eldrazi Spawn the caster
	// offered to Village Rites and leave the sacrifice nothing to pay
	// with, after the mana was made. One list, shared with the
	// legal-move enumerator: see CastAutoTapExclusions.
	excluded := g.castAutoTapExclusionsLocked(p.ID, card, params)
	// #1212: the spell's own source wish. A card whose text reads
	// which mana paid for it ("if mana from a Treasure was spent to
	// cast it") prefers a source it can read back — a tiebreak in the
	// planner's ordering and never a filter, so the plan the solver
	// can find is exactly the plan it could find before.
	//
	// ADR 0118 §1: the plan pays only what the floating pool is
	// missing, so mana already in the pool is spent first, and it is
	// colour-picked against that shortfall (autotap_topup.go).
	plan, short, ok := g.autoTapTopUpLocked(p.ID, cost, params.XValue, spendCtx, excluded, WantedManaSourcesFor(card))
	// #2461: and the plan, carried out, must fund the cost — asked on a
	// clone before anything here is tapped, so a refusal leaves no land
	// tapped and no mana floating.
	if !ok || !g.planFundsLocked(p, plan, short, cost, params.XValue, spendCtx) {
		return &InsufficientManaError{Missing: p.ManaPool.MissingFor(cost, params.XValue, spendCtx)}
	}
	g.materializePlanLocked(p, plan, short)
	return nil
}

// materializePlanLocked taps each card in `plan` and drops the
// produced mana into the controller's pool. Multi-option slots
// (Birds of Paradise, Arcane Signet) get greedy color-picking:
// the picker walks the still-unsatisfied colored requirements and
// picks an option that consumes one. When no requirement matches,
// the slot drops its first option as generic-eligible mana.
//
// A "N mana of any one color" source (#779) is the one slot whose
// colour the executor does NOT re-derive: the plan booked one colour
// for all N tokens and plannedTap.OneColor carries it here, because
// one activation of such a source is one pick and a second greedy
// walk could reach a different answer than the solver did.
//
// #1215: a source whose cost eats it (a Treasure, a Lotus Petal) is
// SACRIFICED here, in the component order ActivateManaAbility pays in
// — counters, then the tap, then the sacrifice, then the mana — and
// the sacrifice's legality is re-asked for the same reason the gate,
// the sickness and the counter cost are: a plan can arrive stale. The
// dies-triggers it queues are NOT drained here. Every caller drains
// them the moment a player would next receive priority (CastSpell,
// ActivateCatalogAbility and the special-action verb each end with
// runStateChecksLocked), which is what puts a Blood Artist trigger
// ABOVE the spell the Treasure was cracked to cast rather than under
// it.
//
// #1228: a planned source need not be a PERMANENT any more. A Spirit
// Guide's "Exile this card from your hand: Add {R}" is a CR 605 mana
// ability that functions from a hand (CR 113.6), so the loop re-finds
// each planned card in whatever zone holds it now and takes the
// non-battlefield arm below when that is not the battlefield. Re-found
// rather than carried on the plan, for the reason every other gate
// here is re-asked: a plan can arrive stale, and a zone stamped at
// planning time would be a second opinion about where a card is.
//
// #1242: and a planned PERMANENT need not be tapped. A Gold token or
// an Eldrazi Spawn prints "Sacrifice this: Add …" with no {T}, so the
// tap — like the sacrifice before it — is a component this executor
// pays exactly when the picked ability prints it. Re-derived from
// ab.TapCost rather than carried on the plan, for the reason the zone
// is: the planner and this function read the one ability
// autoTapAbilityFor picks, so there is nothing for a plan bit to add
// but a second opinion.
//
// Skips PendingChoiceMana entirely — the auto-tapper's contract
// is "no further player decisions required". Caller must hold
// g.mu and have validated the plan via autoTapLocked.
func (g *Game) materializePlanLocked(p *Player, plan tapPlan, cost ParsedCost) {
	pending := append([]ColorRequirement(nil), cost.Required...)
	identity := commanderIdentityFor(g, p)
	for _, planned := range plan {
		cardID := planned.CardID
		var card *Card
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cardID {
				card = &g.Battlefield.Cards[i]
				break
			}
		}
		if card == nil {
			// #1228: not on the battlefield — a hand source, or a
			// permanent that has left since the plan was made. The
			// arm below tells the two apart by asking the card's own
			// ability list, and does nothing for the second.
			g.materializeExiledManaSourceLocked(p, planned, identity, &pending)
			continue
		}
		// #1215: control, re-asked here because the plan may be
		// stale. gatherTapSources only ever offers the payer's own
		// permanents, and a source that has changed hands since is
		// not one of them — tapping it would be somebody else's
		// permanent spent on this cast, and CR 701.21a says a
		// sacrifice cost can only eat what you control, so the
		// component this function now pays makes the re-check load-
		// bearing rather than tidy.
		if card.Controller != p.ID {
			continue
		}
		// #2455: a Costed entry re-finds its ability through the costed
		// picker, every other through autoTapAbilityForRef.
		ab := g.plannedAbilityLocked(p.ID, *card, planned)
		if ab == nil {
			continue
		}
		// #1242: tapped-out, asked AFTER the pick and only of an ability
		// that owes a {T} — the planner's question, through the same
		// helper, in the same order. This used to be `if card.Tapped`
		// before anything else, which was right while every planned
		// source tapped; a Gold that something else tapped is still a
		// sacrifice the plan may make.
		if manaSourceTappedOut(card, ab) {
			continue
		}
		// #1445 / #1427: the planner's paused-exit check, re-asked
		// because a plan can arrive stale.
		if (ab.SacrificeCost || ab.TapCost) && g.zoneChangePausedLocked(cardID) {
			continue
		}
		// #540: CR 302.6, enforced here as well as in the planner.
		// This executor taps `card.Tapped = true` directly rather
		// than routing through ActivateManaAbility, so the gate that
		// path applies does not cover it — and a plan can arrive
		// stale (built before the creature entered) or from a caller
		// that built it by hand. A source that has become sick since
		// the plan was made drops silently, exactly as a gate that
		// stopped holding does above.
		if manaTapBlockedBySickness(card, ab) {
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
		// #789: the counter cost, re-asked here for the same reason
		// the gate and the sickness check are — a plan can arrive
		// stale (built before the last charge counter was spent in
		// response), and tapping a Vivid land the executor then
		// cannot charge would be the strand this whole function
		// avoids.
		if !manaCounterCostPlannable(card, ab) {
			continue
		}
		// #1215: the sacrifice-self component, resolved to the
		// concrete list BEFORE anything is paid and by the SAME
		// validator all three cost sites share, so the auto-tapper
		// cannot grow a second opinion about what a sacrifice cost
		// eats. SacrificeOther is nil by construction —
		// autoTapAbilityFor refuses a source that carries one — which
		// is why no SacrificeIDs are named and none can be demanded.
		//
		// The trailing 0 is #1213's announced X, and 0 is what this
		// site owes: the parameter is read only by a clause whose
		// count comes FROM X ("Sacrifice X Treasures"), the clause
		// here is nil by the construction above, and a mana ability
		// announces no X at all. validateSacrificeCostLocked's own
		// doc says it — "pass 0 from a site that has no X to
		// announce".
		sacrifices, serr := g.validateSacrificeCostLocked(p.ID, cardID, AbilityCost{
			SacrificeSelf: ab.SacrificeCost,
		}, nil, 0)
		if serr != nil {
			continue
		}
		counterPaid := PaidCost{}
		if rc := ab.RemoveCounters; rc != nil {
			counterPaid.CountersRemoved = rc.N
		}
		producedStr := manaAbilityProducedLocked(g, p.ID, cardID, ab, counterPaid)
		slots, err := ParseProducedMana(producedStr)
		if err != nil {
			continue
		}
		// Mirror the activation's option list exactly —
		// identity-first order, or the identity narrowing — so a
		// generic-only slot defaults to an identity colour
		// (pickColorForSlot's options[0]) and the executor never mints
		// a colour the plan didn't have. CR 903.4f (#844): a source
		// the narrowing leaves with nothing to add is dropped BEFORE
		// it is tapped, the same way gatherTapSources refuses to plan
		// it — tapping a land for no mana is worse than not tapping
		// it.
		live := slots[:0]
		for _, slot := range slots {
			slot.Options = manaPickOptions(slot.Options, identity, ab.NarrowToCommanderIdentity)
			if len(slot.Options) == 0 {
				continue
			}
			live = append(live, slot)
		}
		slots = live
		if len(slots) == 0 {
			continue
		}
		// #2558: the set of different colours the plan booked, checked
		// BEFORE the tap like the one colour below. The slot becomes one
		// fixed slot per booked colour, the shape the planner priced.
		expanded, ok := expandDifferentColors(slots, planned.DifferentColors)
		if !ok {
			continue
		}
		slots = expanded
		// #779: the planned colour, checked BEFORE the tap for the
		// same reason the CR 903.4f drop above is — a stale plan (the
		// Nyx Lotus's devotion moved in response, the ability changed)
		// must not tap the permanent and then find it has no colour to
		// mint. Two one-colour slots on one ability is a shape the
		// planner declines, so the executor declines it too rather
		// than guessing which one the plan meant.
		oneColorIdx := oneColorSlot(slots)
		if oneColorIdx == oneColorSlotUnplannable {
			continue
		}
		// #1222: a slot that is a one-pick only AFTER the CR 614 window
		// has multiplied it. Under Mana Reflection a Birds of Paradise
		// mints two of the chosen colour, so the planner priced its
		// printed "{W|U|B|R|G}" as a "two mana of any one color" pick
		// and booked a colour for it (plannedTap.OneColor) — and the
		// printed slot list, which is what the executor produces from,
		// cannot see that. Honour the booking rather than re-deriving
		// it greedily, which is exactly what plannedTap.OneColor exists
		// for: one activation of such a source is ONE pick, and the two
		// halves of the tapper must not reach different answers.
		if oneColorIdx == oneColorSlotNone && planned.OneColor != "" {
			oneColorIdx = firstSlotOffering(slots, planned.OneColor)
		}
		if oneColorIdx >= 0 && !colorOffered(slots[oneColorIdx].Options, planned.OneColor) {
			continue
		}
		// #2455: a Costed entry's mana cost (a Signet's {1}), paid FIRST
		// — ActivateManaAbility's component order, mana before the tap —
		// out of the tokens the plan named for it, which earlier entries
		// made. A cost the pool can no longer pay drops the entry before
		// anything of it is paid.
		if planned.Costed {
			spent, ok := g.payPlannedManaCostLocked(p, *card, ab, planned.FundedBy)
			if !ok {
				continue
			}
			if len(spent) > 0 {
				g.EmitEvent(manaSpentEvent(p.ID, cardID, spent))
			}
		}
		// #789: the counters come off as part of the same payment as
		// the tap, and first, so a refusal leaves the land untapped.
		// applyCounterLocked for the same reason the activation path
		// uses it — a counter-doubling replacement applies only to a
		// counter placed by an effect (CR 614.16), so
		// nothing doubles or cancels it.
		if rc := ab.RemoveCounters; rc != nil {
			if err := g.applyCounterLocked(cardID, rc.Counter, -rc.N); err != nil {
				continue
			}
		}
		// #1183: the SECOND write site for a mana activation, and the
		// one with no counterpart on the CR 602 path. This executor
		// taps the permanent and mints its mana directly rather than
		// routing through ActivateManaAbility, so the record that path
		// writes does not cover it — a plan that spent an exhaust
		// ability without recording it would hand the player the
		// ability straight back on the next cast.
		//
		// Written HERE, at the last point before the permanent is
		// committed and after every re-asked gate above: the tap IS
		// the payment, and nothing below this line can decline the
		// activation. autoTapAbilityFor has already refused a spent
		// ability, so this only ever writes a first use.
		g.noteAbilityActivationLocked(g.activationTallyKeyLocked(cardID, ab.Label))
		// #1242: the tap is a COMPONENT of the cost, paid only when the
		// ability prints one — the same `if ab.TapCost` ActivateManaAbility
		// has always had. A Gold token or an Eldrazi Spawn is cracked, not
		// tapped: it never "becomes tapped" (no EventTapCard, so a
		// "whenever this becomes tapped" watcher stays silent), and
		// CR 106.12a's "tapped for mana" is false for it, which is what
		// the fromTap argument and the triggered mana abilities below
		// read.
		tapped := ab.TapCost
		if tapped {
			card.Tapped = true
		}
		// #763: the permanent as it was when it paid for mana — for the
		// triggered mana abilities fired below and, #1212, for the source
		// kinds the mint records. Copied for the same reason
		// ActivateManaAbility copies it, and taken here because a later
		// source in the same plan may change the board. Taken whether or
		// not the source tapped: a cracked Gold still has to be
		// describable after the sacrifice below has ended it.
		tappedForMana := *card
		if tapped {
			g.EmitEvent(Event{Kind: EventTapCard, Actor: p.ID, CardID: cardID})
		}
		// #2392: the life cost (Mana Confluence's "Pay 1 life"), after
		// the tap and before the sacrifice — ActivateManaAbility's
		// component order. The picker re-asked CR 119.4 and CR 119.8
		// above (autoTapAbilityAccepts), so this refuses only when a
		// replacement stops the payment; the source is then tapped and
		// mints nothing, the sacrifice's honest answer below.
		if ab.LifeCost > 0 {
			if err := g.PayLifeForEffect(cardID, p.ID, ab.LifeCost); err != nil {
				continue
			}
		}
		// ADR 0129 §5: the energy tier's payment (Aether Hub), after the
		// life. The planner held the plan to the controller's energy
		// (autoTapBudget), so this refuses only if something spent it
		// since; the source is then tapped and mints nothing.
		if err := g.payEnergyLocked(p.ID, ab.EnergyCost, cardID); err != nil {
			continue
		}
		// #1215: the sacrifice, AFTER the tap and BEFORE the mana —
		// the component order ActivateManaAbility pays in, and for
		// the same reason: the tap has to happen while the permanent
		// is still on the battlefield, and the mana lands in the pool
		// after the whole cost is paid because CR 605.3b makes the
		// payment and the production one atomic step.
		//
		// A failed payment leaves the source tapped and mints nothing.
		// That is the honest answer rather than a tidy one: the tap
		// has already been emitted and a "whenever this becomes
		// tapped" trigger has already seen it, so unwinding it here
		// would be rewriting history a watcher has read.
		//
		// From here on `card` is DANGLING when the cost ate the
		// source — the battlefield slice it points into has been
		// rewritten. Nothing below reads it: the mana half works off
		// cardID, ab and tappedForMana.
		if len(sacrifices) > 0 {
			// #1397: nil answers — the auto-tapper asks nobody, and a
			// commander whose own mana ability sacrifices it is not a
			// source it plans with.
			if err := g.payCostSacrificesLocked(sacrifices, nil); err != nil {
				continue
			}
		}
		// #1184: the same two stamps the CR 602 announcement carries,
		// so "whenever you activate an exhaust ability" sees Loot, the
		// Pathfinder's exhaust MANA ability through the auto-tapper as
		// well as through a click. The auto-tapper really does
		// activate the ability (CR 605.3 — a mana ability is activated
		// like any other), which is why #1183 made it write the
		// record here.
		g.EmitEvent(Event{
			Kind:   EventManaAbilityActivated,
			Actor:  p.ID,
			Source: cardID,
			// #1210: CardID as well as Source, so the two activation
			// kinds carry the SAME stamps and a watcher that looks up
			// the ability's source object does not have to know which
			// kind it is holding. EventActivateAbility has always set
			// both; this one set only Source, which made a source
			// predicate silently match nothing on the mana path.
			CardID:  cardID,
			Label:   ab.Label,
			Exhaust: ab.Exhaust,
		})
		var addedColors []string
		for si, slot := range slots {
			var color string
			if si == oneColorIdx {
				// #779: the plan's pick, not a fresh one. Its
				// requirements are booked here so a later slot — or a
				// later source in the same plan — does not re-pay a
				// pip these N tokens already cover, which is the #273
				// rule applied to a multi-token pick.
				color = planned.OneColor
				for k := 0; k < slot.AmountFor(color); k++ {
					bookColorRequirement(color, &pending)
				}
			} else if ps, ok := plannedSlotColor(planned, si, len(slots), slot.Options); ok {
				// #2455: an explicit plan's colour. Mana that funds a
				// Costed entry books none of the cost being paid.
				color = ps.Color
				if !ps.Funds {
					bookColorRequirement(color, &pending)
				}
			} else {
				color = pickColorForSlot(slot.Options, &pending)
			}
			if color == "" {
				continue
			}
			// #742: a one-colour-N-mana slot adds all N of the picked
			// colour, in one activation. #1222: through the one
			// production body, which opens the CR 106.12b window on the
			// amount and books whatever surplus it adds against the
			// requirements this cast still owes — the plan already
			// priced the same window, so the ledger and the pool agree.
			//
			// Restrictions ride here too. autoTapAbilityFor already
			// refuses restricted abilities, so this is belt-and-braces
			// — but "the auto-tapper is the one path that mints
			// unrestricted copies of restricted mana" is precisely the
			// bug #259 warns about, and one line is cheaper than
			// trusting a filter two files away.
			addedColors = append(addedColors, g.produceManaLocked(
				p, cardID,
				repeatColor(color, slot.AmountFor(color)),
				restrictionsFor(g, ab, p.ID, cardID),
				ab.SpendRiders,
				// #1212: what the source WAS, snapshotted off the copy
				// taken before the tap — the same copy the triggered
				// mana abilities below use, so the two can never
				// disagree about the permanent they describe.
				//
				// #1215 made that copy LOAD-BEARING here rather than
				// merely consistent. This line used to be able to say
				// "the auto-tapper never plans a sacrifice cost, so the
				// permanent is still here"; it no longer can. The
				// plan's own sacrifice runs between the tap and this
				// mint, so by now a planned Treasure is in a graveyard
				// and a Treasure TOKEN has ceased to exist (CR 111.7).
				// Reading `tappedForMana` is what makes "if mana from a
				// Treasure was spent" answerable on the auto-tap path
				// at all — exactly the reason ActivateManaAbility takes
				// its own snapshot before paying.
				//
				// #1222: every mana the CR 614 window ADDS carries the
				// same snapshot, because CR 106.12b replaces how much
				// mana is produced and not what produced it — a doubled
				// Treasure is two Treasure mana.
				manaSourceKindsOf(tappedForMana),
				// CR 106.12a: a tap for mana exactly when the ability
				// tapped the source. #1242 made that a question — a
				// cracked Gold or Eldrazi Spawn was not tapped, so Mana
				// Reflection's "if you tap a permanent for mana" does
				// not double it, which is the printed answer and the
				// one ActivateManaAbility gives (it passes ab.TapCost).
				tapped,
				&pending,
			)...)
		}
		// #763, CR 605.1b / 605.4a: the third and last production
		// site. `pending` goes in, so a trigger whose output is a
		// colour CHOICE (Fertile Ground's "any color") picks greedily
		// against what this cast still owes instead of leaving a
		// prompt open halfway through an auto-tapped cast — the
		// auto-tapper's "no further player decisions" contract, kept
		// by the same function that keeps it for the source's own
		// slots. The extra mana was never PLANNED (ADR 0074 §7), so
		// whatever it does not pay for simply floats.
		//
		// #1242: only for a source that was TAPPED for mana. CR 605.1b's
		// triggered mana abilities watch a permanent tapped for mana,
		// and a cracked Eldrazi Spawn was not — the same `ab.TapCost &&`
		// ActivateManaAbility gates its own firing on.
		// #2392: the declared rider — a painland's "This land deals 1
		// damage to you" — after the mana and before the triggered mana
		// abilities, where ActivateManaAbility runs it. The picker only
		// accepts a rider that declares its damage (RiderSelfDamage), so
		// this is never an opaque closure the planner could not price.
		// Any state-based action it causes is checked by the caller,
		// which runs the state checks before anyone next gets priority.
		if ab.Rider != nil {
			if err := ab.Rider(g, p.ID, cardID); err != nil {
				g.EmitEvent(Event{
					Kind:     EventEffectError,
					Actor:    p.ID,
					Source:   cardID,
					ErrorMsg: err.Error(),
				})
			}
		}
		if tapped && len(addedColors) > 0 {
			g.fireManaTriggersLocked(ManaProduced{
				Source:     tappedForMana,
				Controller: p.ID,
				Colors:     addedColors,
			}, &pending)
		}
	}
}

// materializeExiledManaSourceLocked is materializePlanLocked's
// NON-BATTLEFIELD arm (#1228): one planned source that is a card in a
// zone its mana ability functions from, paid for by exiling itself.
//
// Two Spirit Guides are the whole family, and the shape is the
// battlefield arm with the tap taken out and the exile put in:
//
//	re-find  →  re-ask the picker, the gate and the condition
//	         →  re-derive the slots  →  pay  →  mint
//
// Everything is re-asked rather than carried across from the planner,
// for the reason the battlefield arm re-asks it: a plan can arrive
// stale — the card may have been discarded in response, the ability
// may have stopped functioning — and a source that has changed drops
// silently rather than stranding what the plan had already spent.
//
// Three differences from the battlefield arm, all of them rules:
//
//   - the "you" is the card's OWNER (CR 108.4), not a controller,
//     because a card outside the battlefield and the stack has none;
//   - the production is NOT a tap for mana (CR 106.12a), so
//     produceManaLocked is told so and no triggered mana ability
//     fires (CR 605.1b wants a permanent tapped for mana — Wild
//     Growth has nothing to attach to a card in hand);
//   - the slots are NOT priced through the CR 614 window here, for
//     the reason the battlefield arm does not price either: the real
//     window runs inside produceManaLocked and pricing twice would
//     multiply twice.
//
// Caller must hold g.mu.
func (g *Game) materializeExiledManaSourceLocked(
	p *Player,
	planned plannedTap,
	identity commanderIdentity,
	pending *[]ColorRequirement,
) {
	cardID := planned.CardID
	var (
		card *Card
		zone ZoneKind
	)
	for _, kind := range supportedManaAbilityZones {
		pile := playerManaZone(p, kind)
		if pile == nil {
			continue
		}
		for i := range pile.Cards {
			if pile.Cards[i].InstanceID == cardID {
				card, zone = &pile.Cards[i], kind
				break
			}
		}
		if card != nil {
			break
		}
	}
	// Not in any pile a mana ability may function from: a permanent
	// that left the battlefield since the plan was made, or a card
	// that has already been spent. Either way it is not a source.
	if card == nil || card.Owner != p.ID {
		return
	}
	ab := g.autoManaExileAbilityFor(p.ID, *card, manaAbilitiesOf(card), zone)
	if ab == nil {
		return
	}
	// #1445: the planner's paused-exit check, re-asked for a stale plan.
	if g.zoneChangePausedLocked(cardID) {
		return
	}
	if g.ActivationGateLocked(p.ID, *card, zone, ActivationAbility{Label: ab.Label, Mana: true}) != nil {
		return
	}
	if ab.Condition != nil && !ab.Condition(g, p.ID, cardID) {
		return
	}
	slots, err := ParseProducedMana(manaAbilityProducedLocked(g, p.ID, cardID, ab, PaidCost{}))
	if err != nil {
		return
	}
	live := slots[:0]
	for _, slot := range slots {
		slot.Options = manaPickOptions(slot.Options, identity, ab.NarrowToCommanderIdentity)
		if len(slot.Options) == 0 {
			continue
		}
		live = append(live, slot)
	}
	slots = live
	if len(slots) == 0 {
		// CR 903.4f: nothing to add. Exiling the card for no mana is
		// worse than leaving it in hand — the same posture the
		// battlefield arm takes about tapping a land for nothing.
		return
	}
	// #2558: the booked set of different colours, as the battlefield
	// arm checks it, before the card is spent.
	expanded, ok := expandDifferentColors(slots, planned.DifferentColors)
	if !ok {
		return
	}
	slots = expanded
	// #779: the plan's booked colour, checked BEFORE the card is
	// spent, exactly as the battlefield arm checks it before the tap.
	oneColorIdx := oneColorSlot(slots)
	if oneColorIdx == oneColorSlotUnplannable {
		return
	}
	if oneColorIdx == oneColorSlotNone && planned.OneColor != "" {
		oneColorIdx = firstSlotOffering(slots, planned.OneColor)
	}
	if oneColorIdx >= 0 && !colorOffered(slots[oneColorIdx].Options, planned.OneColor) {
		return
	}
	// #1212: what this card IS, read BEFORE the cost moves it — the
	// same snapshot ActivateManaAbility takes and for the same
	// reason, one zone over: by the time the mana is minted the card
	// is in exile and this pointer is dangling.
	srcKinds := manaSourceKindsOf(*card)
	// #1183: the activation record, written at the last point before
	// the card is committed — the placement and the argument the
	// battlefield arm gives for its own write, with the exile in the
	// tap's place.
	g.noteAbilityActivationLocked(g.activationTallyKeyLocked(cardID, ab.Label))
	// The payment. Through the one exit primitive with MustSettleNow
	// (payExileSelfCostLocked), so the CR 601.2h indivisible step
	// cannot pause. The auto-tapper asks nobody (nil answers, #1397):
	// no commander carries a Spirit Guide's ability, so the CR 903.9
	// question the hand-clicked activation asks first is not asked
	// here. A failure mints nothing and leaves the card where it was;
	// nothing above this line has changed the board.
	if err := g.payExileSelfCostLocked(p.ID, cardID, true, nil); err != nil {
		return
	}
	// `card` is DANGLING from here — the pile it points into has been
	// rewritten. Nothing below reads it: the mint works off cardID,
	// `ab` and the srcKinds snapshot taken above.
	// #1184: the same two stamps a hand-clicked activation carries,
	// so a watcher sees one kind of event whichever route produced
	// the mana.
	g.EmitEvent(Event{
		Kind:    EventManaAbilityActivated,
		Actor:   p.ID,
		Source:  cardID,
		CardID:  cardID,
		Label:   ab.Label,
		Exhaust: ab.Exhaust,
	})
	for si, slot := range slots {
		var color string
		if si == oneColorIdx {
			color = planned.OneColor
			for k := 0; k < slot.AmountFor(color); k++ {
				bookColorRequirement(color, pending)
			}
		} else {
			color = pickColorForSlot(slot.Options, pending)
		}
		if color == "" {
			continue
		}
		g.produceManaLocked(
			p, cardID,
			repeatColor(color, slot.AmountFor(color)),
			restrictionsFor(g, ab, p.ID, cardID),
			ab.SpendRiders,
			srcKinds,
			// CR 106.12a: not a tap for mana. The card was never
			// tapped and was never a permanent.
			false,
			pending,
		)
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
	if best, bestOpt := mostRestrictiveRequirement(options, *pending); best >= 0 {
		*pending = append((*pending)[:best], (*pending)[best+1:]...)
		return bestOpt
	}
	return options[0]
}

// bookColorRequirement ticks off one still-unsatisfied requirement
// that `color` pays, if there is one. The half of pickColorForSlot
// that runs when the colour is already decided: #779's one-colour
// source mints N tokens of a colour the PLAN chose, and each of them
// has to book its pip or the next slot re-pays it (#273).
func bookColorRequirement(color string, pending *[]ColorRequirement) {
	if color == "" {
		return
	}
	if best, _ := mostRestrictiveRequirement([]string{color}, *pending); best >= 0 {
		*pending = append((*pending)[:best], (*pending)[best+1:]...)
	}
}

// mostRestrictiveRequirement finds the still-unsatisfied requirement
// with the fewest legal colours that any of `options` can pay, and
// returns its index plus the option that pays it. (-1, "") when none
// matches. The one copy of the restriction-first instinct
// pickColorForSlot and bookColorRequirement share.
//
// #1600: a widened requirement (AnyMana) admits every option, so it is
// offered the option it PRINTS when the slot has one — Birds of
// Paradise makes the {R} an Orrery-widened {R} asked for, not the {W}
// listed first — and, between two requirements of the same width, the
// one this slot pays in its printed colour is booked first. Neither
// changes which slots can pay; both keep the mana the plan makes the
// mana the card asked for.
func mostRestrictiveRequirement(options []string, pending []ColorRequirement) (int, string) {
	best, bestOpt, bestPrinted := -1, "", false
	for i := range pending {
		req := pending[i]
		opt, printed, ok := preferredOption(options, req)
		if !ok {
			continue
		}
		if best < 0 || req.width() < pending[best].width() ||
			(req.width() == pending[best].width() && printed && !bestPrinted) {
			best, bestOpt, bestPrinted = i, opt, printed
		}
	}
	return best, bestOpt
}

// preferredOption is the option of `options` that `req` should be paid
// with: the first one it prints, else the first one it admits (a
// widened slot's any mana). printed reports which; ok is false when it
// admits none.
func preferredOption(options []string, req ColorRequirement) (opt string, printed, ok bool) {
	for _, o := range options {
		if matchColor(o, req.Options) {
			return o, true, true
		}
	}
	for _, o := range options {
		if req.Admits(o) {
			return o, false, true
		}
	}
	return "", false, false
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
//
// #696: THE one pricer. Every consumer that has to agree with what
// the payment charges — the auto-tapper, the read-only preview
// endpoint and the bot enumerator — reaches this function through
// PriceCast rather than re-deriving any part of it. See cast_cost.go.
func (g *Game) effectiveCostLocked(p *Player, card Card, params CastSpellParams) (ParsedCost, error) {
	cost, _, err := g.printedCostLocked(p, card, params)
	if err != nil {
		return ParsedCost{}, err
	}
	return g.costAfterModifiersLocked(cost, p, card, params)
}

// costAfterModifiersLocked is effectiveCostLocked's tail: the board's
// cost modifiers (CR 601.2f) and then the convoke / waterbend
// subtraction (CR 601.2h), applied to a cost printedCostLocked has
// already settled.
//
// Split out for PriceCast, which needs BOTH numbers out of one walk —
// the pre-tapping total the announce-time tap budget is measured
// against, and the total the payment charges — without pricing the
// cast twice. Caller must hold g.mu.
func (g *Game) costAfterModifiersLocked(cost ParsedCost, p *Player, card Card, params CastSpellParams) (ParsedCost, error) {
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
	cost, err := g.applyCostModifiersLocked(cost, CostQuery{
		Game:       g,
		Card:       card,
		Controller: p.ID,
		FromZone:   fromZone,
		XValue:     params.XValue,
		Targets:    params.Targets,
		// ADR 0100 §3: the announced sacrifice count, which CR 601.2b
		// settles before 601.2f totals the cost (Torgaar's "{2} less
		// for each creature sacrificed this way").
		Sacrificing: len(params.SacrificeIDs),
		// ADR 0135 §4, CR 702.119a: emerge's reduction, read off the
		// permanent the claimed offer names while it is still on the
		// battlefield.
		AltSacrificeManaValue: g.altSacrificeManaValueLocked(p.ID, card, params),
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
	cost = g.costAfterTapsLocked(cost, card, params)
	// ADR 0100 §1, CR 702.66b: delve applies "only after the total
	// cost of the spell with delve is determined", and after the taps,
	// because convoke may pay a coloured symbol and delve only a
	// generic one. A cast that names no graveyard cards gets the cost
	// back unchanged.
	if len(params.DelveIDs) > 0 && g.DelveForLocked(p.ID, card) {
		cost = delveAdjusted(cost, len(params.DelveIDs), params.XValue)
	}
	return cost, nil
}

// costAfterTapsLocked is the convoke / waterbend subtraction on its
// own. Split out of costAfterModifiersLocked so the delve budget can be
// read off the cost the taps leave, which is what CR 702.66a measures
// it against. Caller must hold g.mu.
func (g *Game) costAfterTapsLocked(cost ParsedCost, card Card, params CastSpellParams) ParsedCost {
	tapCost := TapPermanentsCostFor(CatalogKey(card))
	if !tapCost.Empty() {
		cost = tapPermanentsAdjusted(cost, tapCost, g.tapPermanentsPayersLocked(params.TapIDs), params.XValue)
	}
	return cost
}

// delveBudgetLocked is CastPrice.DelveBudget computed on its own, for
// the announce-time validator: the generic the cast still owes after
// the cost modifiers and the announced taps. Zero for a card with no
// delve, and for a cost that cannot be priced (the cast is refused
// for that elsewhere). Caller must hold g.mu.
func (g *Game) delveBudgetLocked(p *Player, card Card, params CastSpellParams) int {
	if !g.DelveForLocked(p.ID, card) {
		return 0
	}
	base, _, err := g.printedCostLocked(p, card, params)
	if err != nil {
		return 0
	}
	noDelve := params
	noDelve.DelveIDs = nil
	pre, err := g.costAfterModifiersLocked(base, p, card, noDelve)
	if err != nil {
		return 0
	}
	return delveBudget(pre, params.XValue)
}

// printedCostLocked is effectiveCostLocked minus the tap-permanents
// component: the mana this cast owes before convoke or waterbend
// spends anything against it. Split out because the announce-time
// validator needs that number to compute the tap budget, and asking
// effectiveCostLocked for it would be circular — the budget decides
// how many permanents may be tapped, and the tapping is what
// effectiveCostLocked subtracts.
//
// The second return is the pair of cost STRINGS the swap chose
// between — what the card prints, and what this cast pays instead.
// Returned rather than recomputed by the callers that need to SHOW a
// price (the preview endpoint's `cost` field) or ask CR 107.3b about
// it, because a second walk of the same precedence is a second chance
// to disagree about which cost this cast is paying.
//
// Caller must hold g.mu.
func (g *Game) printedCostLocked(p *Player, card Card, params CastSpellParams) (ParsedCost, CastCost, error) {
	// S22: an alternative cost replaces the printed one outright
	// (CR 118.9). The commander tax below is layered on top of
	// whichever cost was chosen, because CR 903.8 taxes the cost
	// being paid, not the cost printed in the corner.
	//
	// S22 airbend, ADR 0066: a granted permission can carry its own
	// "rather than its mana cost" price ({2}), which belongs to the
	// INSTANCE rather than to the card, so it can't come from the
	// oracle-ID-keyed AlternativeCost catalog. It wins over the
	// printed cost and is layered BEFORE the commander tax for the
	// same reason the alternative cost is: CR 903.8 taxes whatever
	// cost is actually being paid.
	//
	// The permission is re-derived here from the zone the cast claims
	// rather than passed down, so a permission that has expired, or
	// that covers a different zone, can never reprice this cast
	// (CR 400.7).
	//
	// Both halves live in CastCostFor, because the announce path asks
	// the same question of the same choice — whether the cost being
	// paid still carries the printed {X} (CR 107.3b) — and two copies
	// of the precedence would be two chances to disagree about which
	// cost this cast is paying.
	srcKind, _ := castZoneFromWire(params.FromZone)
	grant := g.CastPermissionForClaimLocked(p.ID, card, srcKind, params.AlternativeCost)
	alt, err := g.resolveAlternativeCostLocked(p.ID, card, srcKind, grant, params.AlternativeCost, nil)
	if err != nil {
		return ParsedCost{}, CastCost{}, err
	}
	// #1665: a hand permission prices only the claim it opens, the same
	// narrowing CastSpell makes. See CastPermission.ForClaim.
	grant = grant.ForClaim(alt)
	chosen := CastCostFor(card, alt, grant)
	cost, err := ParseCost(chosen.Paid)
	if err != nil {
		return ParsedCost{}, chosen, err
	}
	if params.FromZone == "command" {
		tax := p.CommanderCasts[card.InstanceID]
		cost.Generic += tax * 2
	}
	// ADR 0073 §3, CR 601.2f: the mana half of the optional additional
	// costs the caster announced. Added AFTER the alternative-cost
	// swap and the commander tax and BEFORE the cost modifiers, which
	// is where CR 601.2f puts an additional cost — so Thalia taxes a
	// kicked spell once and Trinisphere reads the kicked total.
	//
	// One helper, called from here and from the bot enumerator, so
	// the price a bot is offered and the price the engine charges
	// cannot drift (#544).
	//
	// ADR 0100 §2: and the chosen either/or branch's mana — Lightning
	// Axe's "pay {5}" — at the same point, as the plan's mandatory
	// entry. An announcement that has not named its branch yet (the
	// view's badge, a preview opened before the radio was answered)
	// prices at no branch mana: CastSpell refuses a missing branch
	// before anything is priced, so this is only ever a quote.
	mandatory, berr := ChosenAdditionalCost(AdditionalCostFor(CatalogKey(card)), params.CostBranch)
	if berr != nil {
		mandatory = nil
	}
	cost, err = AdditionalCostMana(cost, mandatory, OptionalCostsFor(CatalogKey(card)), params.OptionalCosts)
	if err != nil {
		return ParsedCost{}, chosen, err
	}
	// ADR 0065's 2026-09-23 amendment, CR 702.172a: Spree's per-mode
	// cost joins the optional costs' mana at the same point, for the
	// same reason — a card with no Cost on any mode option (every
	// modal card before S45) reprices to the identical number.
	cost, err = AddModeCostMana(cost, ModeSpecFor(CatalogKey(card)), params.Modes)
	if err != nil {
		return ParsedCost{}, chosen, err
	}
	// S21 sub-PR 6: "you may spend mana as though it were mana of any
	// color to cast those spells" (Breeches, Brazen Plunderer). Widens
	// the colored slots to any mana. #1573: "mana of any TYPE" (Hostage
	// Taker) widens the {C} slots too — colorless is a type, not a
	// color (CR 106.1b). Here, in the one pricer, so the payment, the
	// auto-tapper, the preview and the view all read the same widening.
	// #1928: a WIDENING, not a fold into Generic — the grant changes
	// how MANA may pay (CR 609.4b), and convoke (CR 702.51a) and delve
	// (CR 702.66a) are not mana, so a stolen Stoke the Flames still
	// needs red creatures for its {R}{R}, exactly as the player static
	// (costAsPaidByLocked) is read after the taps.
	cost = spendAsThoughAny(grant, cost)
	// #2556: the spell's own "Spend only black mana on X" — resolved
	// against THIS card and stamped last, as the ability pricer stamps
	// an ability's (cost_modifier.go), so every copy of the cost made
	// from here on (the modifiers, the taps, the enumerator's repricing
	// of Base) carries it to costAsPaidByLocked, which folds it for the
	// announced X. String() ignores it: the price shown stays printed.
	cost.SpendOnly = SpellSpendOnlyFor(CatalogKey(card)).ResolveFor(card)
	return cost, chosen, nil
}

// castSourceZoneLocked resolves the FromZone string to the zone
// object the card should leave from. A string that names no zone is
// ErrZoneNotFound.
//
// This lookup grants NOTHING on its own — it answers "which pile",
// not "may you". S29's validateCastPathLocked is the permission
// half, and for exile the per-instance permission is checked in
// CastSpell before anything moves.
//
// Hand and the command zone resolve to the CALLER's own zone, which
// is what CR 601.2 and CR 903.4 mean by them. The graveyard and the
// library are the caller's own too, UNLESS a permission the caller
// holds names this card where it sits — see the branch for why
// (#1022, #1035).
func (g *Game) castSourceZoneLocked(p *Player, cardID uuid.UUID, fromZone string) (*Zone, error) {
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
		// and escape, both of which say "from YOUR graveyard", so the
		// caller's own pile answers almost every cast.
		if p.Graveyard != nil && p.Graveyard.Contains(cardID) {
			return p.Graveyard, nil
		}
		// #1022, and the one exception. ADR 0066 makes a permission a
		// statement about an OBJECT — "you may cast that card" — and
		// nothing in CastPermission says the object has to be in the
		// holder's own zone. Wrexial's "you may cast target instant or
		// sorcery card from that player's graveyard" is the printed
		// shape, and until this branch the permission was answerable
		// by CastPermissionForLocked, invisible to the view (#1022)
		// and unreachable by this lookup, which found no card and
		// returned ErrCardNotFound.
		//
		// A PERMISSION is the only key: the card's own text is never
		// one, because flashback, escape and Gravecrawler all say
		// "your graveyard" and a printed declaration opening every
		// copy in every graveyard would be a rules error in the
		// caster's favour. So the scan asks CastPermissionForLocked —
		// the same function CastSpell validates with and the view and
		// the enumerator read — and a card no permission covers stays
		// exactly as unreachable as it was.
		return g.foreignPileForCastLocked(p, cardID, ZoneGraveyard)
	case ZoneLibrary:
		// S42 / CR 401.5: "you may play lands and cast spells from the
		// top of your library" — the caller's own library answers
		// every printed library clause but one.
		//
		// THE TOP CARD is the only one a permission opens, and that is
		// checked by CastPermissionForLocked rather than here: this
		// lookup answers "which pile", not "may you", and the position
		// rule belongs with the permission that names it.
		if p.Library != nil && p.Library.Contains(cardID) {
			return p.Library, nil
		}
		// #1035, and the same exception the graveyard has carried
		// since #1022. Xanathar, Guild Kingpin's "you may play the top
		// card of their library" is a permission over ANOTHER seat's
		// pile, and until the position check learned to follow the
		// card it could not exist — so this branch could not have been
		// reached. It is the same rule and the same guard: a
		// permission the caller holds is the only key, and a card
		// nothing covers is ErrCardNotFound exactly as it was.
		return g.foreignPileForCastLocked(p, cardID, ZoneLibrary)
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

// foreignPileForCastLocked finds the per-seat pile a card is sitting
// in when it is not the caster's own — a graveyard (#1022) or a
// library (#1035) — and only when the caster holds a CastPermission
// over it there.
//
// ONE function for the two zones, because it is one rule: the pile a
// permission reaches is the pile the CARD is in, and a permission the
// caller holds is the only way in. Splitting it per zone is how the
// library came to be missing the branch in the first place.
//
// The permission check is the whole point: without it this would open
// every printed flashback card in every opponent's graveyard and every
// revealed library top at the table, which is the one direction a
// sandbox must never err in. With it, the answer is the same one
// CastPermissionForLocked gives the view and the bot enumerator, so
// the three cannot disagree about whether the cast exists — including
// CR 401.5's "the top card", which lives with the permission rather
// than here.
//
// ErrCardNotFound when no pile holds the card, and when one does but
// nothing lets this player cast it from there — the error the old
// per-player lookup gave for both, so a client that names a card it
// has no business naming sees no change.
//
// Caller must hold g.mu.
func (g *Game) foreignPileForCastLocked(p *Player, cardID uuid.UUID, kind ZoneKind) (*Zone, error) {
	for _, other := range g.Seats {
		if other == nil || other == p {
			continue
		}
		pile := g.permissionZoneLocked(other.ID, kind)
		if pile == nil {
			continue
		}
		for _, card := range pile.Cards {
			if card.InstanceID != cardID {
				continue
			}
			if g.CastPermissionForLocked(p.ID, card, kind) == nil {
				return nil, ErrCardNotFound
			}
			return pile, nil
		}
	}
	return nil, ErrCardNotFound
}

// SorcerySpeedOpenLocked reports whether the sorcery-speed gate is
// currently open for the given player: caller is the active seat,
// the cursor is on a main phase, and the stack is empty (CR 307.1).
//
// THE ONE sorcery-timing read. Every "any time you could cast a
// sorcery" in the engine comes through here — a sorcery or creature
// cast (CastTimingOpenLocked), a land play (CR 305.1), an "activate
// only as a sorcery" or loyalty ability (ActivationTimingOpenLocked),
// the plot and suspend special actions (SpecialActionTimingOKLocked)
// — and the bot enumerator asks it too, so none of them can hold a
// private copy of the rule.
//
// "Empty" means EMPTY (CR 117.1a, CR 405.1): an activated or
// triggered ability is an object on the stack exactly as a spell is.
// Abilities have no card in g.Stack — they live only in g.StackMeta —
// so this asks stackHasItemsLocked, the same question the priority
// wrap asks, rather than looking at the stack zone alone. Before
// #1352 it looked at the zone alone, and a sorcery, an equip or a
// plot went through over a pending upkeep trigger.
//
// Caller must hold g.mu.
func (g *Game) SorcerySpeedOpenLocked(playerID uuid.UUID) bool {
	if g.Turn.Step != StepPrecombatMain && g.Turn.Step != StepPostcombatMain {
		return false
	}
	if g.stackHasItemsLocked() {
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
//
// Every object put on the stack takes its Seq here — a cast, an
// activation, a trigger, a copy — so this is also the one place that
// sees every push, and the passes made before one no longer count
// (#2275, CR 117.4): the object on top is new, and every seat passes
// over it before it resolves. See priority_succession.go.
func (g *Game) nextStackSeqLocked() uint64 {
	g.restartPassSuccessionLocked()
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
	// #829: a resolution is one of the two points where play moves on,
	// so everything this item emits is one occurrence and the next
	// resolution's events are a different one — which is what lets a
	// second bounce trigger Dour Port-Mage again while the first
	// bounce's draw is still on the stack. See event_batch.go.
	g.beginEventBatchLocked()
	if g.Stack == nil || len(g.Stack.Cards) == 0 {
		// No spell on the stack — but there could still be ability
		// items in StackMeta. Find the most recent and resolve it.
		g.resolveTopAbilityLocked()
		return nil
	}
	// ADR 0107 §3: the spell resolves with the abilities the layer pass
	// gives it (CR 613.1f), and `top` below is the copy every exit
	// reads — rebound among them. A grant registered since the last
	// pass (an ability that resolved just above this spell) has to be
	// on the card before it is copied. A no-op when nothing changed.
	g.RecomputeLayersIfStaleLocked()
	// Top of the stack is the last card in the slice (LIFO).
	top := g.Stack.Cards[len(g.Stack.Cards)-1]
	item, ok := g.StackMeta[top.InstanceID]
	if !ok || item == nil {
		// Defensive: a card on the stack without a meta entry should
		// not happen — but if it does, route to graveyard so the
		// stack doesn't wedge. No meta means no record of the cost
		// that was paid, so no flashback replacement either.
		return g.routeStackCardToGraveyardLocked(top, nil, false)
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
	// #920, CR 707.10: the item is off the stack, but a spell may copy
	// ITSELF while it resolves, and the copy is built from this meta
	// and this card. Both are unreachable a moment from now — the meta
	// is gone from the map above, and the card is routed away below
	// while the copy decision is still an open prompt — so they are
	// parked for the rest of this occurrence. See resolving_item.go.
	g.beginResolvingSpellLocked(item, top)
	// #1289, CR 704.3: nothing is checked until this resolution has
	// finished, including any part of it that waits on a prompt. See
	// resolution_pause.go.
	g.beginResolutionLocked()
	defer g.endResolutionLocked()
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
	if spellAllTargetsIllegalLocked(g, item) {
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
		// the resolution. A BOUGHT-BACK one is not returned to hand,
		// for the mirror-image reason: CR 702.27a says "as it
		// resolves", and a spell countered by game rules never
		// resolves. Hence `false`.
		return g.routeStackCardToGraveyardLocked(top, item, false)
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
	// #2696, CR 702.131a: an instant or sorcery with ascend checks the
	// controller's permanents before any of its other instructions, so
	// "if you have the city's blessing" below reads the answer.
	g.ascendSpellLocked(&top, item)
	g.fireEffectResolverLocked(item, CatalogKey(top), top.InstanceID)
	// CR 608.2c / 700.2d: each chosen bullet's own body, in printed
	// order, once per occurrence. A modal card that branches inside
	// its OnResolve on ctx.HasMode declares no ModeOption.Effect and
	// this is a no-op for it (#764).
	g.runChosenModeEffectsLocked(item, ModeSpecFor(CatalogKey(top)))
	// #489, CR 608.2n: the spell may have MOVED ITSELF. Everything
	// below this line routes the object that is still on the stack —
	// to the battlefield, out of existence, or to a graveyard — and a
	// spell whose own effect put it somewhere else has no such object
	// left. See spellMovedItselfLocked for why that is a return and
	// not an error.
	if g.spellMovedItselfLocked(top.InstanceID) {
		return nil
	}
	// CR 608.3f / CR 111.13 (also CR 707.10f): a resolving copy of a
	// PERMANENT spell does not put a permanent card onto the
	// battlefield — it becomes a TOKEN that is a copy of the spell,
	// and the copy ceases to exist. The branch below is for cards;
	// this one is for the object that is not one. See
	// resolvePermanentSpellCopyLocked (#666).
	//
	// BELOW the #489 check above, and the order is the rule: CR 707.10a
	// says a copy in any zone other than the stack has already ceased
	// to exist, so a copy that moved ITSELF never resolves and never
	// becomes a token.
	if item.IsCopy && top.IsPermanent() {
		return g.resolvePermanentSpellCopyLocked(top, item)
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
		// ADR 0103, CR 709.5d: which door the permanent enters
		// unlocked is the half that was CAST, read before the face is
		// reset below.
		castDoors := castDoorsOf(top, item)
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
			// CR 708.4, ADR 0082 decision 3: a spell cast face down
			// resolves into a face-down permanent. Seeded onto the
			// event rather than applied after the push, so the whole
			// CR 614 window — and any replacement that inspects the
			// entry — sees what is arriving, and so the pause-and-
			// resume path carries it with everything else.
			FaceDown: item.FaceDown,
			// ADR 0103, CR 709.5d: the cast door, seeded like
			// the counters below so the pause-and-resume path
			// carries it.
			EntersUnlocked: castDoors,
		}
		// S29: "this creature escapes with a +1/+1 counter on it"
		// (CR 702.138c). Seeded onto the event BEFORE the pipeline
		// runs, so the counters are part of the entry every other
		// replacement gets to see and modify — Doubling Season
		// doubles them — rather than an afterthought stapled on once
		// the permanent has landed.
		g.applyAltCostEntryCountersLocked(ev, top, item)
		// #1002, CR 614.1c: the card's OWN "this permanent enters with
		// X +1/+1 counters on it", read off the same still-reachable
		// StackItem the line above reads. Two clauses, one entry
		// event, one pipeline. See entry_counters.go.
		g.applyCastEntryCountersLocked(ev, top, item)
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
			g.restoreEntryControllerLocked(ev)
			return nil
		}
		// #653: THE push, rather than a second copy of it. This branch
		// used to reproduce executeEntryToBattlefieldLocked inline —
		// the move, the controller, the copy, the counters, the zone
		// move, the Aura attach, the ETB and evoke's trigger — while
		// the RESUME of the very same event went through the finisher.
		// Two copies of "a spell becomes a permanent" is one too many
		// for a fact that has to be written exactly once, and the two
		// had already drifted twice: #478's land-drop tally, and then
		// ADR 0073 §5's kicked stamp, which had to be written into
		// both. The finisher does everything this did, off the same
		// event: ev.Actor is item.Controller and ev.stackItem is item,
		// which is what carries the Aura's target and the costs that
		// were paid across a pause.
		_, err = g.executeEntryToBattlefieldLocked(out)
		return err
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
	// exile, when the flashback cost was paid (CR 702.34a) or the
	// spell went on an Adventure (CR 715.3d), or to the owner's HAND,
	// when the buyback cost was paid (CR 702.27a). This is the one
	// call site that resolves, so it is the one that passes `true`.
	return g.routeStackCardToGraveyardLocked(top, item, true)
}

// spellMovedItselfLocked reports that the spell this resolution is
// about is no longer on the stack, because its own effect moved it
// there and then somewhere else.
//
// #489. "Exile Ascend from Avernus", "shuffle this card into its
// owner's library", Genesis Ultimatum's "exile Genesis Ultimatum" —
// the instruction is part of the spell's own text, so it runs inside
// `OnResolve`, which is BEFORE the resolution frame decides where the
// spell goes next. Everything after that decision point is about an
// object on the stack: the battlefield entry for a permanent, the
// token a resolving copy of a permanent spell becomes (CR 608.3f,
// #967), and CR 608.2n's "as the final part of an instant or sorcery
// spell's resolution, the spell is put into its owner's graveyard". A
// spell that has already left has none of those left to do, and
// CR 608.2n is the rule that says so: the thing it puts into a
// graveyard is the spell ON THE STACK. For the copy the rule is
// CR 707.10a — a copy in any zone other than the stack has already
// ceased to exist, so it never resolves and never becomes a token.
//
// What happens without this check is not an error, which is why the
// check is worth writing down. It used to be one — the frame called
// `MoveCard(g.Stack, …)` and got `ErrCardNotFound`, which
// `PassPriority` returned to the caller with the post-resolution
// state checks and the priority reset skipped, the shape #489 was
// filed for. #529 then put the stack exit on the shared exit
// primitive (`routeStackCardToGraveyardLocked` → `routeCardToZoneLocked`),
// which finds a card's zone BY SCAN rather than assuming the stack.
// So the error went away and took the symptom with it, and the defect
// got quieter and worse: the frame found the card in exile, saw a
// destination that was not exile, and dutifully moved the spell OUT
// of the exile its own effect had just put it in and into the
// graveyard. Ascend from Avernus would have exiled itself and landed
// in the graveyard a microsecond later, with no error anywhere.
//
// ONE CHECK, EVERY EXIT. It sits in the resolution frame rather than
// in `routeStackCardToGraveyardLocked` because all three of the
// frame's post-effect exits share the assumption, not just the
// graveyard one: a PERMANENT spell that moved itself would otherwise
// take `MoveCard(g.Stack, g.Battlefield, …)` and reintroduce the
// original error verbatim. The two exits ABOVE this point — the
// "countered by game rules" fizzle (CR 608.2b) and the no-StackMeta
// fallback — run before any card code and cannot be in this state, so
// `routeStackCardToGraveyardLocked` needs no guard of its own and
// keeps its single responsibility.
//
// NOT COVERED, deliberately: a self-move that PAUSED. The card is
// still on the stack while a CR 614 prompt about its own move is
// open, so this returns false and the frame routes it to the
// graveyard, which prunes the stale prompt (#605). Skipping the route
// instead would leave the spell on the stack with nobody left to
// finish it — a wedge, which is worse than the misordering. It is
// also out of reach in practice: an instant or sorcery is never a
// commander, so the only pause available to one is a CR 616 ordering
// prompt between two replacements that both apply to its own exit.
//
// Caller must hold g.mu.
func (g *Game) spellMovedItselfLocked(cardID uuid.UUID) bool {
	return g.Stack == nil || !g.Stack.Contains(cardID)
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
func spellAllTargetsIllegalLocked(g *Game, item *StackItem) bool {
	if item == nil || len(item.Targets) == 0 {
		return false
	}
	// #2182: "This ability still resolves if its target becomes
	// illegal." The clause the item was announced under says so; the
	// effect then treats an illegal target as unaffected.
	for _, step := range g.itemAnnouncedClauses(item) {
		if step.Clause.ResolvesIfIllegal {
			return false
		}
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
			// #764: each ref is re-checked against the clause it was
			// announced under, which for a multi-clause or modal item
			// is not the item's first clause.
			if g.TargetStillLegalForEffect(item, t) {
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
// engine tracks (findCardZoneLocked covers every one). Caller must
// hold g.mu.
func targetStillExistsLocked(g *Game, t TargetRef) bool {
	switch t.Kind {
	case TargetPlayer:
		p := g.playerByIDLocked(t.ID)
		return p != nil && !p.Eliminated
	case TargetCard:
		// #1211, CR 115.4: a TargetCard ref may name an ABILITY item
		// on the stack, which has no card in any zone — its source
		// permanent stays where it is and the item is a StackMeta
		// entry alone. It exists while it is still on the stack, and
		// stops existing the moment it resolves or is countered
		// (CR 701.6a), which is exactly the question this asks.
		//
		// Only reached for a ref with no announced clause behind it (a
		// free-form S13.1 announcement); a structured clause is
		// re-checked by specMatchLocked, which reads StackMeta itself.
		if item := g.StackMeta[t.ID]; item != nil && item.Kind != StackItemSpell {
			return true
		}
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
	// #920: the same park the spell path does, without a card — an
	// ability has none on the stack. Nothing copies an ability yet
	// (that is the ability-copy seam), but "the item currently
	// resolving" has one answer or the next reader finds a hole.
	g.beginResolvingLocked(top)
	// #1289: the same CR 704.3 hold the spell path takes.
	g.beginResolutionLocked()
	defer g.endResolutionLocked()
	g.recomputeSplitSecondLocked()
	if spellAllTargetsIllegalLocked(g, top) {
		g.EmitEvent(Event{
			Kind:                EventFizzle,
			Actor:               top.Controller,
			Source:              top.SourceCardID,
			Label:               top.Label,
			ResolvedStackItemID: top.ID,
		})
		return
	}
	g.EmitEvent(Event{
		Kind:                EventResolve,
		Actor:               top.Controller,
		Source:              top.SourceCardID,
		Label:               top.Label,
		ResolvedStackItemID: top.ID,
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
	// CR 608.2c: a modal triggered or activated ability resolves its
	// chosen bullets in printed order, after whatever body the item
	// itself carries (#764).
	g.runChosenModeEffectsLocked(top, top.modeSpec)
}

// routeStackCardToGraveyardLocked moves a card off Game.Stack and
// into its owner's graveyard. Used by resolveTopOfStackLocked for
// instants / sorceries and by the "countered by game rules" path
// (sub-PR 3) when every target is illegal on resolve.
//
// THE ONE PLACE a spell leaving the stack chooses a destination.
// Four rules replace the graveyard, and all four decide here rather
// than in the resolution frame, because a spell has exactly one
// destination and a reader should be able to see the whole contest in
// one switch:
//
//   - FLASHBACK — "exile this card instead of putting it anywhere
//     else any time it would leave the stack" (CR 702.34a). Every
//     exit, which is why `resolved` does not gate it.
//   - BUYBACK — "if the buyback cost was paid, put this card into its
//     owner's hand as it resolves instead of putting it into that
//     player's graveyard" (CR 702.27a, ADR 0073 §6). AS IT RESOLVES
//     and no other exit, which is what `resolved` is for: a bought-back
//     Capsize whose only target left in response is countered by game
//     rules, does not resolve, and goes to the graveyard.
//   - ADVENTURE — "exile that card instead of putting it into its
//     owner's graveyard as that spell finishes resolving", CR 715.3d,
//     with CR 715.4's cast permission landing on it there (#719,
//     adventure.go). Resolution only, for the same reason buyback is:
//     CR 715.3d's "instead of" only fires on a resolution, so a
//     countered, fizzled, discarded or milled adventure card is an
//     ordinary card in an ordinary graveyard, and `resolved` is the
//     fact that tells them apart. Before #988 gave
//     this helper that fact, the adventure leg had to live one frame
//     up to get it.
//   - REBOUND — "if this spell was cast from your hand, instead of
//     putting it into your graveyard as it resolves, exile it and, at
//     the beginning of your next upkeep, you may cast this card from
//     exile" (CR 702.88a, rebound.go, #1854). Resolution only, like
//     the two above it, and a fact about where the spell was CAST
//     from, which the item records.
//
// `item` is the spell's stack item, because buyback and flashback are
// facts about what was PAID and rebound is a fact about where it was
// cast from; the Adventure is a fact about the card and reads the face
// instead.
//
// PRECEDENCE, and the one CHOICE. Flashback wins over the rest,
// because CR 702.34a replaces every exit and the others replace one of
// them (and a flashed-back spell was not cast from a hand, so it never
// has rebound to compete with).
//
// Rebound, buyback and the Adventure exile all replace the SAME event,
// "put it into its owner's graveyard as it resolves", so when two of
// them apply CR 616.1 gives the choice to the affected object's
// controller — the spell's — and CR 616.1f then finds the others
// inapplicable, because the card is no longer going to a graveyard.
// That is resolutionExitsLocked and chooseResolutionExitLocked
// (resolution_exits.go, ADR 0107 §3, #1854). No printed card has two
// of them; a granted rebound (Cast Through Time over a buyback or an
// Adventure spell) is the first way to meet the choice. Until it, the
// buyback-and-Adventure pair was settled by a stated judgement call
// (buyback); it is the controller's choice now, as CR 616.1 says. One
// applicable exit is applied with no question.
//
// Pass nil for the defensive no-StackMeta path, where there is no
// cost to read, and `resolved` false with it.
//
// Caller must hold g.mu.
func (g *Game) routeStackCardToGraveyardLocked(c Card, item *StackItem, resolved bool) error {
	// S29 flashback. Checked before the owner lookup because it does
	// not depend on one: the destination is the shared exile zone
	// either way, which is also where a card whose owner has left
	// the game already goes.
	//
	// #529: both destinations go through the shared exit primitive,
	// so every replacement window sees a spell that fizzles
	// ("countered by game rules", CR 608.2b) or resolves to a
	// graveyard. A commander goes to the graveyard like any card and
	// is offered the command zone afterwards (CR 903.9a, ADR 0115). A
	// queued replacement prompt leaves the card on the stack until
	// the owner answers — priority cannot pass while a choice is
	// outstanding, so nothing resolves on top of it in the meantime.
	r := zoneRoute{
		CardID: c.InstanceID, Dst: ZoneGraveyard, DstOwner: c.Owner,
		// #1320: CR 608.2n puts the card away as the last step of its
		// own resolution; no spell or ability is exiling it (a
		// flashed-back or adventure card below goes to exile by a
		// rule, not by an effect).
		Cause: MoveCause{Kind: MoveCauseRule},
	}
	if altCostExilesFromStack(c, item) {
		r.Dst, r.DstOwner, r.Actor = ZoneExile, uuid.Nil, c.Owner
	} else if resolved {
		// The three "as it resolves" exits (resolution_exits.go). Two
		// or more is CR 616.1's choice, asked of the spell's
		// controller; the card waits on the stack for the answer.
		exits := resolutionExitsLocked(c, item)
		if len(exits) > 1 {
			return g.chooseResolutionExitLocked(c, item, r, exits)
		}
		if len(exits) == 1 {
			exits[0].apply(&r)
		}
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
		// CR 400.7, the same stamp the catalog activation takes:
		// which OBJECT the ability came from. See
		// StackItem.SourceEpoch.
		SourceEpoch:  g.cardObjectEpochLocked(sourceCardID),
		SourceObject: g.sourceObjectRefLocked(sourceCardID),
		Label:        params.Label,
		Targets:      append([]TargetRef(nil), params.Targets...),
		Modes:        append([]int(nil), params.Modes...),
		XValue:       params.XValue,
		Distribution: cloneDistributionLocked(params.Distribution),
		Seq:          g.nextStackSeqLocked(),
	}
	// CR 115.3: the targets are chosen as the ability goes on the
	// stack, and becoming a target is an event whether the ability
	// came out of the catalog or off a player's own reading of the
	// card (#968). Same call, same place as the catalog activation
	// (activated.go) and the manual trigger announce below: after the
	// item exists, so a ward or "becomes the target" trigger harvested
	// off it lands on the stack above the thing that targeted.
	g.emitBecameTargetLocked(playerID, sourceCardID, id, params.Targets)
	// And the drain the catalog path runs for the same reason: the
	// activator receives priority right after activating (CR 117.3c),
	// so a trigger harvested off the announce goes on the stack at
	// that boundary (CR 603.3b) — above the ability, where a ward
	// trigger has to be to counter it.
	g.runStateChecksLocked()
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
	// CR 606.3's sorcery window, through the read the catalogued
	// activation path shares (#1208). It moved BELOW the lookup
	// because the read needs the object: a statement about "loyalty
	// abilities of planeswalkers you control" is asked of THIS
	// permanent, and this sandbox verb is the one place a walker the
	// catalog has never heard of can have a loyalty ability
	// activated at all. Nothing else about the order changed — the
	// once-per-turn flag is still the first thing asked, and a
	// missing permanent still reports ErrCardNotFound.
	if !g.ActivationTimingOpenLocked(playerID, *pw, ZoneBattlefield, ActivationAbility{Label: label, Loyalty: true}) {
		return ErrSorcerySpeedRequired
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
	// #1474: a planeswalker an effect has already paused on its way
	// out is gone as far as the rules are concerned, and its loyalty
	// cost is paid by that object. The catalogued path refuses it in
	// activateCatalogAbilityLocked; this sandbox verb asks the same
	// gate.
	if err := g.refusePausedCostCardsLocked([]uuid.UUID{planeswalkerID}); err != nil {
		return err
	}
	// CR 606.6: can't remove more loyalty than is there. The old
	// comment here argued for letting the counter go negative so the
	// SBA could see the intent, but the SBA reads "loyalty <= 0" and
	// an activation the rules forbid should be refused at announce,
	// not paid and then cleaned up.
	if delta < 0 && pw.Counters[CounterLoyalty] < -delta {
		return ErrInsufficientLoyalty
	}
	// payCostCounterLocked (counter_cost.go): a loyalty ability's
	// counter change is a COST (CR 606.4), not an effect, so a
	// replacement that names "an effect" (Doubling Season) still does
	// not apply — but this now opens the CR 614 window with
	// CounterFromCost set, so a replacement that names no effect at all
	// (Vorinclex, Monstrous Raider) does, matching the catalogued
	// activation path in activated.go. This used to call
	// applyCounterLocked directly and bypass the window entirely, which
	// was right about Doubling Season and wrong about Vorinclex — ADR
	// 0073's 2026-09-28 amendment, #1710.
	if _, err := g.payCostCounterLocked(playerID, planeswalkerID, CounterLoyalty, delta); err != nil {
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
	// #2275 / CR 117.3c: a loyalty ability is an activated ability,
	// and activating one is an action — the passes before it no longer
	// count. This sandbox shape puts nothing on the stack, so it says
	// so itself.
	g.noteActionTakenLocked(playerID)
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
// list, and the engine queues an item and drains it —
// drainPendingTriggersAPNAPLocked(), through runStateChecksLocked,
// the same call PassPriority and every other priority-grant path
// makes. The announcer holds priority and keeps it, so the announce
// IS a CR 603.3b placement moment; anything else genuinely waiting
// goes on the stack in the same APNAP batch.
//
// Two placements, not one, and that is #974: the trigger is placed,
// THEN its targets are announced (CR 603.3d), then what THAT
// triggered is placed — above it. A ward trigger harvested off the
// announcement has to sit above the announced trigger or declining
// the payment counters nothing, and one batch could not put it there:
// APNAP orders a batch by SEAT, so a ward on the active player's
// permanent would have gone under a non-active player's announced
// trigger.
//
// What this costs is the batching of several MANUAL announces into
// one APNAP placement: each click now places its own trigger, in
// click order, rather than all of them in seat order at the next
// pass. That batching only ever applied to triggers the players were
// announcing by hand, it was already theirs to sequence, and the
// alternative was a verb that left its trigger queued while the table
// advanced the step past it.
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
		SourceObject: g.sourceObjectRefLocked(sourceCardID),
		Label:        params.Label,
		Targets:      append([]TargetRef(nil), params.Targets...),
		Modes:        append([]int(nil), params.Modes...),
		XValue:       params.XValue,
		Distribution: cloneDistributionLocked(params.Distribution),
		// CR 603.3d / 115.3: a manually-announced trigger chooses its
		// targets as it goes on the stack, same as the harvested kind,
		// so the drain that places it emits the "becomes the target"
		// event (#1539) — after the item, and the rest of its batch,
		// is on the stack, so what the announcement triggers is placed
		// ABOVE it.
		TargetsAnnouncePending: len(params.Targets) > 0,
	})
	g.EmitEvent(Event{
		Kind:   EventTrigger,
		Actor:  playerID,
		Source: sourceCardID,
		Label:  params.Label,
	})
	// CR 603.3b: the queued trigger is put on the stack the next time
	// a player would receive priority, and clicking "trigger" IS that
	// moment — the announcer has priority and keeps it. Draining here
	// places this trigger (APNAP, alongside anything else genuinely
	// waiting) instead of leaving it queued behind whatever the table
	// does next: before #974 a pass round the table found an empty
	// stack and ADVANCED THE STEP, and the announced trigger landed a
	// step late.
	//
	// The same boundary puts what the announcement triggered on the
	// stack: the drain places the trigger and emits its "becomes the
	// target" events, and the loop's next pass places a ward trigger
	// harvested off them above it, which is where it has to be if
	// declining the payment is to counter anything (#974). Before
	// #1539 the event was emitted here, after the drain, and a second
	// sweep placed the ward; a drain held behind an ordering prompt
	// then let the ward join the batch the announced trigger waited in.
	g.runStateChecksLocked()
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
// CR 704.5. Reports whether any SBA fired and whether a player left.
// The caller repeats the checks (CR 704.3) and settles departures only
// after the repeated passes are quiet.
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
//   - 704.5v/w: a battle with 0 defense counters is moved to its
//     owner's graveyard
//   - 704.5q: +1/+1 and -1/-1 counters on the same creature
//     cancel out — remove min(N, M) of each
//   - 704.5s: a saga whose final-chapter lore counter is set is
//     sacrificed by its controller (the SBA half; the lore-counter
//     advance trigger lands in S14+ with the effect catalog)
//
// Supertype SBAs (ADR 0109 §8):
//   - 704.5k: the world rule — of two or more world permanents, all
//     but the most recent entrant go to their owners' graveyards, and
//     all of them on a tie (entry_ordinal.go)
//
// Existence SBAs (#596):
//   - 704.5d: a token in any zone other than the battlefield ceases
//     to exist — see token_existence.go
//
// Counter ordering (CR 704.3): every action in one check is performed
// simultaneously, so the +1/+1 / -1/-1 cancel runs AFTER the doomed set
// is collected and swept. It never changes power or toughness, so the
// order cannot change which creatures die; it decides only that a
// creature dying in this check leaves with both kinds of counter, which
// is what undying and persist read (#2075).
//
// Caller must hold g.mu.
func (g *Game) stateBasedActionsLocked() (fired, left bool) {
	if g.State != StateActive {
		return false, false
	}
	// ADR 0115 decision 1, CR 903.9a / CR 704.6d: a commander put into
	// a graveyard or exile since the last check is offered the command
	// zone. FIRST, so it collects its cards off the same board the rest
	// of the pass reads (CR 704.3); a commander this pass kills is
	// marked by that move and asked on the next pass. Not "fired": a
	// question is not an action performed, and runStateChecksLocked
	// holds the boundary while it is open. See commander_return.go.
	g.commanderReturnSBALocked()
	// #1199 / CR 702.26, ADR 0084: the "phases out until ~ leaves the
	// battlefield" family comes back the moment its source is gone —
	// Oubliette destroyed, Out of Time's last time counter removed.
	//
	// NOT a state-based action; it is a CR 611.2b "for as long as"
	// duration ending, and this pass is simply the one place in the
	// engine that runs after every action with the board settled. It
	// is FIRST so that the recompute below sees the returning
	// permanents and attachmentSBALocked sweeps an Aura that phased in
	// onto a host that died while it was away (CR 702.26i, CR 704.5m)
	// in this same settling rather than a round later.
	//
	// CR 704.3's "state-based actions ignore phased-out permanents"
	// needs no code: a phased-out permanent is not in the battlefield
	// slice every check below walks.
	if g.sweepPhaseInLocksLocked() {
		fired = true
	}
	// S16: refresh effective characteristics before any toughness /
	// loyalty / battle-defense check. Counter mutations + zone moves
	// from prior SBA iterations bump layerVersion; this fast-paths
	// when nothing's changed. Without it, the lethal-damage SBA
	// would see printed toughness instead of post-anthem effective
	// (a 2/2 + Glorious Anthem under 3 marked damage would die
	// because CurrentToughness reads Effective().Toughness == 3).
	g.RecomputeLayersIfStaleLocked()

	// #2696, CR 702.131b: ascend on a permanent is a static ability
	// that gives its controller the city's blessing any time they
	// control ten permanents. Not a state-based action, so it does not
	// count toward `fired`; this pass is simply the one place that runs
	// after every action with the board settled and before a player
	// receives priority. After the recompute so a keyword granted by a
	// layer-6 effect is seen, and followed by another recompute so the
	// statics that read the new designation are in place before the
	// toughness checks below.
	if g.citysBlessingSweepLocked() {
		g.RecomputeLayersIfStaleLocked()
	}

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

	// Player-loss SBAs (ADR 0057 Decision 2). CR 704.3 performs them
	// all at once, so the pass COLLECTS every loser first and only then
	// takes them out of the game, and the game-over check runs once,
	// after the whole batch: CR 104.3f's "win and lose at once is a
	// loss" is this order, and a player counted as the last survivor
	// can no longer be eliminated a line later in the same pass.
	//
	// Every gate is read here, against the board as it is now. A
	// player at 0 life whose Platinum Angel just died loses at this
	// check with no window to respond (the Abyssal Persecutor ruling);
	// life, poison and commander damage are re-read every pass, so
	// nothing has to be stored for them.
	//
	// Play moves on once, after the batch (#766): moving the turn on
	// per player would begin the turn of a seat that is leaving in this
	// same pass. Ending the turn clears marked damage, so
	// runStateChecksLocked waits until repeated SBA passes settle
	// before rotating.
	type lossEntry struct {
		p     *Player
		cause LossCause
	}
	var losers []lossEntry
	for _, p := range g.Seats {
		if p.Eliminated {
			continue
		}
		// CR 704.5b looks only at draws "since the last time
		// state-based actions were checked", so the flag is consumed
		// by every pass, whether or not a gate stops the loss.
		drew := p.AttemptedEmptyDraw
		p.AttemptedEmptyDraw = false
		if cause, ok := g.sbaLossCauseLocked(p, drew); ok {
			losers = append(losers, lossEntry{p: p, cause: cause})
		}
	}
	for _, l := range losers {
		if g.leaveGameLocked(l.p, l.cause, uuid.Nil) {
			left = true
			fired = true
		}
	}
	// ADR 0057 Decision 3: an active player who lost by an effect in
	// the middle of a resolution has already left; the game-over
	// check and the rotation were deferred to this pass.
	if g.ActiveSeatLeftPending {
		g.ActiveSeatLeftPending = false
		left = true
		fired = true
	}
	if left {
		g.checkGameOverLocked()
	}

	// CR 800.4c (#769) — a permanent whose controller has left the
	// game is exiled. The departure itself already swept the board it
	// left behind (leave_game.go); this catches the one case that can
	// only appear LATER, when a control-changing effect ends and hands
	// its permanent back to a player who is no longer there.
	if g.exileGhostControlledLocked() > 0 {
		fired = true
	}

	// Permanent + counter destruction SBAs. Collect doomed instance
	// IDs in a pre-pass to avoid mutating the slice while iterating.
	//
	// A creature whose toughness the engine does not KNOW is skipped
	// whole — Card.ToughnessIsKnown has the rule and the reasons, and
	// it is the only gate CR 704.5f has here. Skipping is skipping the
	// creature, not just its 704.5f check: an object with no toughness
	// has no lethal-damage threshold either.
	// #667: each entry carries whether its rule DESTROYS the permanent
	// (CR 704.5g/h) or merely puts it into a graveyard (CR 704.5f/i/v),
	// because a regeneration shield replaces the first two and not the
	// last three. The set still leaves as one simultaneous event.
	var doomed []doomedPermanent
	var (
		zeroLoyaltyExemption  zeroLoyaltyExempt
		zeroLoyaltyExemptRead bool
	)
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
		// The rules below are independent: a permanent is checked against
		// EVERY one its current types call for, and is doomed if any
		// applies (CR 704.3). A planeswalker that is also a creature (a
		// Gideon until end of turn) answers to the creature rules and to
		// 704.5i, so no arm may `continue` past the others. It is
		// DESTROYED only if every rule that doomed it destroys; one
		// "put into a graveyard" rule (704.5f, 704.5i, 704.5v/w) is
		// something indestructible and a regeneration shield do nothing
		// about, whatever else applies (ADR 0032 amendment of 2026-10-07).
		var (
			doom        bool
			destruction = true
		)
		if c.IsCreature() {
			// Whether the toughness is KNOWN is judged as if this
			// check's CR 704.5q cancel had already happened: a `*`
			// placeholder whose counters cancel to nothing is as
			// unknown as one that never had any (the cancel itself
			// runs after the sweep, below). Only the CREATURE rules
			// are skipped for it: an object with no toughness has no
			// lethal-damage threshold, but its loyalty is still read.
			if afterPlusMinusCancel(c).ToughnessIsKnown() {
				curT := c.CurrentToughness()
				// CR 704.5f — toughness 0 or less PUTS the creature into
				// its owner's graveyard; it does not destroy it, so
				// indestructible deliberately does not save it here. A
				// 2/2 with indestructible under two -1/-1 counters dies.
				if curT <= 0 {
					doom, destruction = true, false
				} else {
					// S25 (#77): the two damage-driven creature SBAs
					// below are both DESTRUCTION (CR 704.5g, CR 704.5h),
					// so indestructible switches both off (CR 702.12b).
					// The damage stays marked either way — see
					// indestructible.go.
					indestructible := IsIndestructible(&c)
					if c.DamageMarked >= curT && !indestructible {
						doom = true
					}
					// S18 sub-PR 3: CR 702.2c — a creature hit by any
					// nonzero damage from a deathtouch source is
					// destroyed at the next SBA regardless of
					// toughness. The flag is consumed by the pass that
					// reads it (consumeDeathtouchMarksLocked, #2319):
					// CR 704.5h only counts damage dealt since the last
					// check, so a creature that was indestructible for
					// that check is not destroyed by the mark at a later
					// one.
					if c.MarkedLethalByDeathtouch && !indestructible {
						doom = true
					}
				}
			}
		}
		// 704.5i — planeswalker with 0 loyalty counters, unless a static
		// says it isn't put into the graveyard for that (Sanctum Lurker,
		// zero_loyalty_exemption.go). Read lazily: nearly every pass has
		// no walker at zero, and the exemption read is a battlefield scan.
		if c.IsPlaneswalker() && (c.Counters == nil || c.Counters[CounterLoyalty] <= 0) {
			if !zeroLoyaltyExemptRead {
				zeroLoyaltyExemption, zeroLoyaltyExemptRead = g.zeroLoyaltyExemptControllersLocked(), true
			}
			if !zeroLoyaltyExemption.covers(c.Controller) {
				doom, destruction = true, false
			}
		}
		// 704.5v/w — battle with 0 defense counters.
		if c.IsBattle() && (c.Counters == nil || c.Counters[CounterDefense] <= 0) {
			doom, destruction = true, false
		}
		if doom {
			doomed = append(doomed, doomedPermanent{id: c.InstanceID, destruction: destruction})
		}
	}
	// 704.5k (ADR 0109 §8) — the world rule. Not a choice, so it joins
	// the doomed set and leaves in the same simultaneous event as
	// everything else this pass puts into a graveyard (CR 704.3): every
	// world permanent but the most recent entrant, and all of them on a
	// tie. It is not destruction, so indestructible and regeneration do
	// nothing to it. A world permanent already doomed above (a world
	// creature with lethal damage) is not listed twice.
	for _, id := range g.worldRuleDoomedLocked() {
		listed := false
		for _, d := range doomed {
			if d.id == id {
				listed = true
				break
			}
		}
		if !listed {
			doomed = append(doomed, doomedPermanent{id: id})
		}
	}
	// S27 / CR 310.12b: a battle at zero defense is DEFEATED, and the
	// defeated trigger has to see it on the battlefield. Announced
	// here — after the doomed set is collected and before any of it
	// moves — so the harvester finds a live source, and so a Siege's
	// "exile it, then cast it transformed" reads a card that still
	// exists. The battle is in `doomed` already via the 704.5v/w arm
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
	g.sweepDoomedPermanentsLocked(doomed)
	// CR 704.5h counts deathtouch damage dealt "since the last time
	// state-based actions were checked", so this pass has consumed the
	// mark (#2319). DamageMarked stays: CR 704.5g counts it to cleanup.
	g.consumeDeathtouchMarksLocked()
	if len(doomed) > 0 {
		// "Did this pass do anything", which is what `fired` means — not
		// "how many were destroyed", which is what destroyPermanentsLocked
		// now returns (#815). The two stopped being the same number the
		// moment a cancelled destruction counted zero: a permanent whose
		// destruction the CR 614 window replaced away is still doomed and
		// still has to be looked at again, and one whose exit is paused on
		// a replacement's prompt is skipped by the collector above on the
		// next pass. Both are answered by the set this pass COLLECTED.
		fired = true
	}

	// Counter cancel (704.5q). AFTER the doomed set is collected and
	// swept, for CR 704.3: every state-based action in one check is
	// performed simultaneously, so a creature this check puts into a
	// graveyard leaves with the counters it had, both kinds included.
	// Its last-known counters are what undying and persist read
	// (#2075, ADR 0113 §4): a 1/1 undying creature with a +1/+1 counter
	// that gets two -1/-1 counters dies with its +1/+1 counter and does
	// not return. Running the cancel first would have removed that
	// counter before the creature died, and brought it back. The
	// cancel removes as many of one kind as of the other, so it never
	// changes power or toughness and cannot change which creatures the
	// pre-pass dooms.
	//
	// #1664: ONLY +1/+1 against -1/-1. Every other P/T counter kind
	// changes power and toughness (PTCounterDelta), but CR 704.5q
	// names these two and no others: a +1/+0 and a -1/-0 on one
	// creature both stay, as do a +1/+1 and a -2/-1.
	if g.cancelPlusMinusCountersLocked() {
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
		// permanent waiting on a replacement's prompt is still on the
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

	// 704.5j (S27) — the legend rule. Last of the choosing SBAs,
	// because it is the one state-based action whose outcome is a
	// CHOICE rather than a consequence: running it after the
	// destruction and sacrifice passes means a player is never asked
	// to pick between two legends when one of them was about to leave
	// anyway. See legend_rule.go.
	if g.queueLegendRuleChoicesLocked() {
		fired = true
	}

	// 704.5d (#596) — a token in any zone but the battlefield ceases
	// to exist. Last of all, and deliberately after the destruction
	// and sacrifice passes above: a token that died in THIS pass has
	// already landed in its owner's graveyard and already had its
	// dies / leaves-the-battlefield triggers harvested off that move
	// (the harvest is synchronous, inside EmitEvent), so sweeping it
	// now costs those triggers nothing and saves the loop a pass.
	// See token_existence.go.
	if g.tokenCeaseToExistSBALocked() {
		fired = true
	}
	// 704.5e, the part the engine keeps outside the stack (ADR 0090):
	// a CR 722.3c prepare copy ceases to exist anywhere but the stack,
	// and in exile once the permanent that kept it there is gone or
	// unprepared. Beside the token sweep, for the token sweep's reason.
	if g.prepareCopySweepSBALocked() {
		fired = true
	}
	// CR 702.179a (ADR 0138): a player with no speed who controls a
	// permanent with start your engines! gets speed 1. See speed.go.
	if g.startYourEnginesSBALocked() {
		fired = true
	}
	// ADR 0091, CR 702.75a: the controller of a hideaway permanent may
	// look at the card it hid. Not a state-based action and never
	// "fired" — it keeps a knower set current as control moves.
	g.hideawayKnowersSweepLocked()

	return fired, left
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
// Reports whether any state-based action was PERFORMED (CR 704.3's
// "as a result of the check"), across every pass of the loop. Almost
// every caller ignores it — the answer only matters where a rule asks
// the question, and today that is CR 514.3a, which grants priority in
// the cleanup step when an SBA fired even if nothing reached the
// stack (cleanupGrantsPriorityLocked, cleanup.go). One body rather
// than a reporting copy, so the two can never drift.
//
// Caller must hold g.mu.
func (g *Game) runStateChecksLocked() (sbaFired bool) {
	// #1289, CR 704.3: a player gets priority only once the resolving
	// item has finished, and a resolution waiting on one of its own
	// prompts has not. Hold the whole boundary (the sweep and the
	// CR 603.3 drain) until the answer finishes it. See
	// resolution_pause.go.
	if g.holdForOpenResolutionLocked() {
		return false
	}
	// #2165, CR 724.1: an effect ended the turn during the resolution
	// this boundary follows. The rest of the process — the 724.1c
	// check, the skip to the cleanup step and that step itself — is
	// owed now, and it IS this boundary. See end_turn.go.
	if g.TurnEndPending {
		return g.finishEndingTheTurnLocked()
	}
	return g.stateChecksLocked(true)
}

// stateChecksLocked is runStateChecksLocked's loop, past the
// resolution hold. `drain` false is CR 724.1c's check (end_turn.go):
// state-based actions are performed until none fires, and the
// triggered abilities they cause wait on PendingTriggers instead of
// going on the stack, because nobody is about to receive priority.
// Every caller but that one passes true.
//
// Caller must hold g.mu.
func (g *Game) stateChecksLocked(drain bool) (sbaFired bool) {
	// ADR 0107 §6, CR 615.5: the instance of damage is over before a
	// player receives priority, so the next-damage shields' "the damage
	// prevented this way" runs now, once per shield with the total —
	// before the state-based actions, so a combat damage step's damage
	// has all been dealt (CR 510.2) and nothing has died of it yet.
	g.flushPreventionFollowUpsLocked()
	// #2183: the batch is over before a player receives priority, so
	// the AtBatchEnd triggers staged in it are asked now, before the
	// state-based actions can take their sources away.
	g.settleBatchEndTriggersLocked()
	// #1729, CR 610.3: an "until" return is created immediately after
	// its event, so it is owed before the state-based actions — "nothing
	// happens between the two events, including state-based actions"
	// (Hostage Taker ruling, 2017-09-29).
	g.resolveUntilReturnsLocked()
	// #830 / CR 509.2a: a player is about to receive priority, so the
	// block declaration is complete. Lock it in first, so the
	// "becomes blocked" and "blocks" triggers it produces are on
	// PendingTriggers before this pass drains them — one harvest, off
	// the final assignment. No-op when nothing is staged, which is
	// every call outside the declare-blockers step.
	g.commitBlockDeclarationLocked()
	// #859 / CR 508.2: the same rule one step earlier. Attackers are
	// declared as one turn-based action and their triggers go on the
	// stack when the active player next receives priority, so the
	// staged declaration is announced here — once per attacker, off
	// the defender it ends the declaration on. No-op when nothing is
	// staged, which is every call outside the declare-attackers step.
	g.commitAttackDeclarationLocked()
	const maxIter = 32
	departuresPending := false
	for i := 0; i < maxIter; i++ {
		// ADR 0115 decision 3, CR 704.3: the CR 903.9a question is part
		// of the state-based actions, so nothing goes on the stack and
		// no further pass runs until every commander_return prompt is
		// answered. ResolveCommanderReturn runs the checks again.
		if g.holdForCommanderReturnLocked() {
			return sbaFired
		}
		// #809 / CR 603.3d: a targeted trigger dispatched from inside a
		// resolving spell's own events froze its legal set while that
		// spell was still on the stack. This is the priority-grant
		// boundary the rules choose targets at, so the frozen set is
		// re-read here — widened, narrowed, or withdrawn outright when
		// nothing legal is left. No-op when no pick_target is open,
		// which is nearly every call. See trigger_target_timing.go.
		g.refreshTargetChoicesLocked()
		// ADR 0107 §1: the sweep is CR 704.3's "single event", so the
		// per-event state-trigger check waits for it to finish, and the
		// check below reads the board it settled on.
		release := g.holdStateTriggersLocked()
		fired, left := g.stateBasedActionsLocked()
		release()
		sbaFired = sbaFired || fired
		// #1729, CR 610.3: the pass can BE the event — a creature
		// that died is an object that left the battlefield, and a
		// player who lost took their objects out of the game and may
		// have handed on the crown (CR 725.4). The return is owed
		// before the next pass and before the turn moves on. It is not
		// a state-based action, so it does not count toward sbaFired
		// (CR 514.3a reads that), but the loop runs again after it.
		if g.resolveUntilReturnsLocked() {
			fired = true
		}
		// CR 603.8 / CR 704.3: state triggers are asked in every pass,
		// after the state-based actions and before the waiting triggers
		// go on the stack. A state no event announced — a continuous
		// effect that began or ended in a layer pass — is caught here.
		g.stateTriggersLocked()
		// #864: belt-and-braces backstop, run every pass so nothing can
		// leave this function about to hand a seat priority while a
		// choice sits pending for a chooser this same pass (or an
		// earlier one) already eliminated. See sweepEliminatedChoicesLocked.
		g.sweepEliminatedChoicesLocked()
		departuresPending = departuresPending || left
		// ADR 0115 decision 3: the pass that asked a CR 903.9a question
		// ends the boundary here. A departure this pass performed still
		// owes its rotation, so it is deferred to the run the answer
		// starts, through the flag a mid-resolution loss already uses.
		if g.holdForCommanderReturnLocked() {
			if departuresPending && g.State == StateActive {
				g.ActiveSeatLeftPending = true
			}
			return sbaFired
		}
		if departuresPending && g.State == StateActive {
			if fired {
				// CR 704.3 repeats the checks before anything gets priority.
				// Keep the old turn's damage until every destruction caused
				// by these departures and the preceding deaths has settled.
				continue
			}
			g.advancePastEliminatedLocked()
			departuresPending = false
			// Cleanup and the next turn's entry hooks can change the board.
			// Check it before draining the waiting triggers into that turn.
			continue
		}
		if !drain {
			// CR 724.1c: repeat until quiet; put nothing on the stack.
			if !fired {
				return sbaFired
			}
			continue
		}
		hasPending := len(g.PendingTriggers) > 0
		if !fired && !hasPending {
			return sbaFired
		}
		if hasPending && !g.drainPendingTriggersAPNAPLocked() && !fired {
			// Held behind a CR 603.3b ordering prompt and SBAs are
			// quiet — nothing more to do until the chooser answers.
			return sbaFired
		}
	}
	return sbaFired
}

// sbaLossCauseLocked is the player half of CR 704.5a–c and CR
// 704.6c: the first state-based loss that holds for p AND that no
// "can't lose the game" gate stops, in CR 704.5 order — life, the
// empty-library draw, poison, commander damage. A player who is both
// at 0 life and at 10 poison while something stops only the life loss
// (Phyrexian Unlife) falls through to poison and loses to it.
//
// `drew` is the CR 704.5b fact, already consumed from the player by
// the caller. Caller must hold g.mu.
func (g *Game) sbaLossCauseLocked(p *Player, drew bool) (LossCause, bool) {
	if p.Life <= 0 && g.canLoseLocked(p, LossLife) {
		return LossLife, true
	}
	if drew && g.canLoseLocked(p, LossEmptyDraw) {
		return LossEmptyDraw, true
	}
	// 704.5c: poison ≥ 10. Player.Counters is the S13.2 map; the
	// legacy single-int Player.Poison field stays in sync via
	// SetPoison so old action paths keep working.
	poison := 0
	if p.Counters != nil {
		poison = p.Counters[CounterPoison]
	}
	if poison < p.Poison {
		poison = p.Poison
	}
	if poison >= PoisonLethal && g.canLoseLocked(p, LossPoison) {
		return LossPoison, true
	}
	if p.IsDeadByCommanderDamage(g.Settings.CommanderDamage) && g.canLoseLocked(p, LossCommanderDamage) {
		return LossCommanderDamage, true
	}
	return "", false
}

// eliminatePlayerLocked is one player leaving the game on their own
// (a concession): leaveGameLocked, then settleDeparturesLocked. The
// SBA settling loop calls the two halves separately so a batch of
// losers moves play on once, after repeated checks settle. Never
// gated — CR 104.3a. Caller must hold g.mu.
func (g *Game) eliminatePlayerLocked(p *Player) {
	if !g.leaveGameLocked(p, LossConcede, uuid.Nil) {
		return
	}
	// #1729, CR 610.3: leaving can be the event an "until" waits for
	// (their Hostage Taker left the battlefield with them; the crown
	// they wore went to an opponent, CR 725.4). The return happens now,
	// in the turn they left in, before the rotation below moves play on.
	if g.survivingSeatsLocked() > 1 {
		g.resolveUntilReturnsLocked()
	}
	g.settleDeparturesLocked()
}

// leaveGameLocked transitions a seated player to eliminated state,
// cleans up their stack items, pending triggers and prompts, emits
// EventPlayerEliminated (Label = the LossCause, Source = the object
// whose effect made them lose, if any), and then takes their objects
// out of the game (CR 800.4a, leave_game.go). It reads no gate — that
// is loseGameLocked's job — and it does not move the turn on or check
// whether the game is over; settleDeparturesLocked does both, once
// per batch. Reports whether the player left (false when they had
// already). Caller must hold g.mu.
//
// The elimination event is emitted BEFORE the objects go, so anything
// watching a player lose the game sees the board they lost with.
func (g *Game) leaveGameLocked(p *Player, cause LossCause, source uuid.UUID) bool {
	if p.Eliminated {
		return false
	}
	p.Eliminated = true
	p.AttemptedEmptyDraw = false
	// #2450: a player leaving the game is loop progress (ADR 0055's
	// 2026-10-07 amendment, option B, owner decision 3a).
	g.noteLoopProgressLocked()
	// #2275: the departure takes their spells and abilities off the
	// stack (CR 800.4a), so what the other seats passed over may not
	// be what is on top now. The succession starts again from whoever
	// holds priority; the departed seat is no longer waited for.
	g.restartPassSuccessionLocked()
	g.cleanupStackForEliminatedLocked(p.ID)
	g.EmitEvent(Event{
		Kind:   EventPlayerEliminated,
		Actor:  p.ID,
		Source: source,
		CardID: source,
		Label:  string(cause),
	})
	// #769 / CR 800.4a: everything they own leaves the game, their
	// control effects end, and anything still controlled by them is
	// exiled. Runs after the stack cleanup above, which is what folds
	// the spell cards that cleanup used to exile into the removal —
	// a card owned by a player who has left does not sit in exile.
	//
	// Except when this departure is the one that ENDS the game. Then
	// the board is not stripped: nothing can observe the objects
	// leaving, no rule reads the table again, and the last board is
	// what the winner, the post-game screen and the replay look at.
	// checkGameOverLocked leaves the turn cursor where it was for
	// exactly the same reason. See ADR 0060 Decision 5.
	if g.survivingSeatsLocked() > 1 {
		g.leaveGameObjectsLocked(p.ID)
	}
	return true
}

// settleDeparturesLocked runs after one or more players have left:
// the game ends when one player or none is left, and otherwise play
// moves on (advancePastEliminatedLocked — the rest of a departed
// active player's turn ends through the rotation seam, #766).
//
// Used where the departures are the whole event (Concede, and an
// effect loss by a player who is not the active one). The SBA
// settling loop calls the two halves separately, because the SBA
// settling must finish before the turn ends — see
// runStateChecksLocked.
//
// Caller must hold g.mu.
func (g *Game) settleDeparturesLocked() {
	if g.checkGameOverLocked() {
		return
	}
	g.advancePastEliminatedLocked()
}

// cleanupStackForEliminatedLocked implements the stack half of
// CR 800.4a — when a player leaves the game, every spell and ability
// they control on the stack ceases to exist. Ability items are just
// deleted from StackMeta; spell items have their card moved out of
// Game.Stack to exile, which is the rule's last clause ("anything
// still controlled by them is exiled") for a card somebody ELSE owns.
// A card the departed player owns does not stop here: leaveGameLocked
// runs leaveGameObjectsLocked immediately afterwards and that card
// leaves the game from exile, along with the rest of what they own
// (#769). Pending triggers controlled by the eliminated player are
// dropped from the queue.
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
	// CR 800.4a, in the rule's order (ADR 0104 §7): the effects that
	// give the departed player control END before anything they still
	// control is removed. A spell they had stolen goes back to the
	// player it was taken from — or to its caster — rather than being
	// exiled for a player who never cast it. The recompute is what
	// hands it back (the stack step of the layer pass). The same drop
	// ends a resolving spell's theft of a PERMANENT (Act of Treason),
	// which leaveGameObjectsLocked's own recompute then hands back.
	//
	// Not on the departure that ENDS the game: that one keeps the
	// final board as it was (ADR 0060 Decision 5).
	if g.survivingSeatsLocked() > 1 && g.endControlEffectsForLocked(playerID) {
		g.RecomputeLayersIfStaleLocked()
	}
	if len(g.StackMeta) > 0 {
		toRemove := make([]uuid.UUID, 0, len(g.StackMeta))
		for id, item := range g.StackMeta {
			// CR 800.4c for the stack: a spell that went back to a
			// player who had ALREADY left has nobody to control it
			// either, and goes the way this player's own does.
			if item == nil {
				continue
			}
			if p := g.playerByIDLocked(item.Controller); item.Controller == playerID || (p != nil && p.Eliminated) {
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
	g.dropPendingTriggersForLocked(playerID)
	// A choice owed by a player who has left the game can never be
	// answered, and while it sits in the queue every other seat is
	// blocked behind it (pass_priority is refused client-side and the
	// bot enumerator offers nothing while a choice is open). Drop the
	// eliminated player's prompts — their objects are gone with them
	// (CR 800.4a), so a damage assignment, target pick or scry they
	// owed has nothing left to act on. Found by the S31 bot fuzzer.
	//
	// #808: a dropped prompt may be holding a paused replacement event,
	// and a life or damage event carries its CALLER's continuation — the
	// rest of a drain, every later opponent's loss, the caster's gain.
	// dropChoicesForPlayerLocked collects those frames; finish them
	// below, once the queue no longer holds the dropped prompts.
	dropped := g.dropChoicesForPlayerLocked(playerID)
	if g.DiscardPending != nil {
		delete(g.DiscardPending, playerID)
		if len(g.DiscardPending) == 0 {
			g.DiscardPending = nil
		}
	}
	g.recomputeSplitSecondLocked()
	for _, frame := range dropped {
		g.finishDroppedReplacementLocked(playerID, frame)
	}
}

// dropChoicesForPlayerLocked settles every PendingChoice owed by
// playerID — the same-name cleanup-discard pause in their name is
// left to the caller — and returns the replacementResume frames the
// choices it DROPPED were holding, for the caller to finish via
// finishDroppedReplacementLocked (#808). Shared by
// cleanupStackForEliminatedLocked (the instant a player leaves) and
// sweepEliminatedChoicesLocked (#864's defensive backstop, below) so
// the two settle a departed chooser's queue identically — the #794
// lesson, one answer and no second list.
//
// #902 / CR 800.4g/h: "settles", not "drops". A prompt an object
// requires and a player still in the game can answer is REASSIGNED
// rather than dropped, and stays in the queue under its new chooser.
// reassignDepartedChoiceLocked (leave_game.go) holds the whole policy
// — which kinds move, and who inherits — and reports false for
// everything else, which is the pre-#902 drop, event and all.
//
// #961 / CR 800.4f: a drop is not always the end of the question. The
// departure table's second column names the DEFAULT ACTION a dropped
// prompt of that kind still has to take — for pay_unless the cost is
// not paid, so the "unless" branch runs; for option_pick the question
// ends but the rest of the card does not, so the continuation runs
// with "nobody chose" (#1006). The actions themselves are
// runChoiceDropActionLocked (pending_choice.go), shared with every
// other withdrawal path so a kind cannot be settled two ways; what is
// local to a DEPARTURE is the gate in front of them
// (departedChoiceActionAllowedLocked, leave_game.go), because only a
// departure can take the material or the card out of the game
// underneath the action.
//
// They run at the bottom of this function, after the queue has been
// rewritten, for the reason #808's replacement frames do: a
// continuation may queue the next prompt, and it must not land in a
// slice this loop is still writing over. A continuation that queues to
// the DEPARTED seat is refused by QueueChoiceForEffect's own guard
// (#864); one that queues to a survivor is a prompt that seat really
// does owe.
//
// Caller must hold g.mu.
func (g *Game) dropChoicesForPlayerLocked(playerID uuid.UUID) []*replacementResumeFrame {
	if len(g.PendingChoices) == 0 {
		return nil
	}
	var dropped []*replacementResumeFrame
	var settle []*PendingChoice
	kept := g.PendingChoices[:0]
	for _, c := range g.PendingChoices {
		if c == nil || c.Chooser != playerID {
			kept = append(kept, c)
			continue
		}
		if g.reassignDepartedChoiceLocked(c) {
			// Same entry, new chooser. It is still owed, so it still
			// blocks the table and the enumerator still offers it —
			// to a seat that can answer.
			kept = append(kept, c)
			continue
		}
		g.EmitEvent(Event{
			Kind:   EventPendingChoiceDropped,
			Actor:  playerID,
			Source: c.Source,
			Label:  string(c.Kind),
		})
		if choiceDepartureDecisions[c.Kind].onDrop != dropDiscard {
			settle = append(settle, c)
		}
		if c.replacementResume != nil && c.replacementResume.ev != nil {
			dropped = append(dropped, c.replacementResume)
		}
	}
	g.PendingChoices = kept
	if len(g.PendingChoices) == 0 {
		g.PendingChoices = nil
	}
	for _, c := range settle {
		// The object gate is re-read here rather than in the loop
		// above: an earlier action may have been the thing that took
		// the next one's object off the table. Which gate depends on
		// the action, and leave_game.go holds that policy.
		if !g.departedChoiceActionAllowedLocked(c) {
			continue
		}
		g.runChoiceDropActionLocked(c)
	}
	return dropped
}

// sweepEliminatedChoicesLocked drops any PendingChoice whose chooser
// has left the game. #864: cleanupStackForEliminatedLocked already
// does this the instant a player leaves (#287/#808), but that sweep
// runs once, at that exact moment — it cannot catch a choice queued
// for the same player afterward. QueueChoiceForEffect's own guard
// (#864) now refuses to queue such a choice going forward, so this
// sweep is the belt to that guard's braces: a second, independent
// layer that self-heals the queue at every state-based-action pass
// (runStateChecksLocked, which is the seam where the table is next
// about to offer a seat priority) regardless of whether some future
// caller ever manages to slip a choice past the queue-time guard, or
// a game restored from a snapshot written before this fix shipped
// carries one already.
//
// Idempotent: a game with no eliminated seat holding a choice costs
// one scan of g.Seats and returns without touching g.PendingChoices.
// Never touches a live chooser's choice — only p.Eliminated seats are
// considered. Caller must hold g.mu.
func (g *Game) sweepEliminatedChoicesLocked() {
	if len(g.PendingChoices) == 0 {
		return
	}
	for _, p := range g.Seats {
		if p == nil || !p.Eliminated {
			continue
		}
		for _, frame := range g.dropChoicesForPlayerLocked(p.ID) {
			g.finishDroppedReplacementLocked(p.ID, frame)
		}
	}
}

// finishDroppedReplacementLocked settles a paused replacement event
// whose prompt was dropped because its chooser left the game.
//
// Before #808 the frame was simply discarded. For an event kind that
// carries no continuation that is still all there is to do, and the
// only cleanup owed is the event's CR 614.5 once-per-event entry,
// which nothing will clear now.
//
// A LIFE or DAMAGE event is different: it carries its caller's
// continuation (lifeTail / damageTail.then), and "each opponent loses 3
// life, you gain life equal to the life lost this way" is sequenced
// through those continuations, so discarding the first leg's frame
// silently dropped every later opponent's loss and the caster's gain.
// Every terminal outcome of those events has to reach the tail; this is
// the one leaving the game adds.
//
// #865: so does a MOVE, for the same reason and with the same
// consequence — a multi-card discard and a sequenced wipe both carry
// the rest of the batch on zoneRoute.then. That one is
// abandonZoneRouteLocked (zone_route.go), the exit's terminal outcome
// for a prompt taken away rather than answered, shared with the two
// prune paths.
//
// Two cases, split on whose event it was:
//
//   - The player who left IS the one the event happens to — the life
//     player, the damaged player, or the controller of the damaged
//     permanent (the CR 616 chooser is always that player). CR 800.4a
//     takes them and their objects out of the game, so nothing lands
//     and the continuation runs with zero.
//   - They were only the chooser of a "may" on somebody
//     ELSE's event. That event still happens. The pipeline is resumed,
//     and the apply-loop's existing gone-chooser escapes decide for the
//     player who left (the "may" is declined, an ordering stands as
//     gathered); the event then lands through the same functions the
//     unpaused path uses, continuation included.
//
// It deliberately lands WITHOUT the state-based sweep the CR 616 resume
// runs: this is reached from eliminatePlayerLocked, which the SBA loop
// itself calls, and the sweep the change owes happens on the loop's
// next pass or at the next priority boundary.
//
// Caller must hold g.mu, and must already have removed the dropped
// prompt from g.PendingChoices.
func (g *Game) finishDroppedReplacementLocked(gone uuid.UUID, frame *replacementResumeFrame) {
	ev := frame.ev
	logErr := func(what string, err error) {
		if err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: what + ": " + err.Error()})
		}
	}
	if ev.Kind != RepEventLife && ev.Kind != RepEventDamage {
		// #865: an EXIT carries its caller's continuation too, and a
		// discard or a wipe sequenced through it stalls when the frame
		// is discarded. abandonZoneRouteLocked is that event kind's
		// terminal outcome, and clears the CR 614.5 entry for every
		// other kind on its way past.
		logErr("route continuation failed", g.abandonZoneRouteLocked(frame))
		return
	}
	runTail := func(amount int) {
		if ev.Kind == RepEventLife {
			logErr("life continuation failed", g.runLifeTailLocked(ev, amount))
			return
		}
		logErr("damage continuation failed", g.runDamageTailLocked(ev, amount))
	}
	if affectedPlayerForEvent(ev, frame.applicable, g) == gone {
		g.clearReplacementEventLocked(ev.ID)
		runTail(0)
		return
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// The affected player has a question of their own to answer;
		// the event resumes from that prompt like any other.
		return
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		logErr("replacement resume failed", err)
		runTail(0)
		return
	}
	if out == nil || out.Canceled {
		runTail(0)
		return
	}
	// Both landing functions run the continuation themselves on every
	// outcome, a vanished player or permanent included; what is left to
	// do with an error is log it.
	if out.Kind == RepEventLife {
		logErr("life change dropped", g.applyResolvedLifeChangeLocked(out))
		return
	}
	logErr("damage dropped", g.applyResolvedDamageLocked(out))
}

// routeBattlefieldCardToOwnerGraveyardLocked moves a battlefield
// card to its owner's graveyard, clearing battlefield-only state
// (combat declarations, position, marked damage and the CR 702.2c
// deathtouch flag are all zeroed by MoveCard's CR 400.7 cleanup).
// Used by SBAs that destroy creatures, by every effect destroy, and —
// because sacrifice is also a battlefield exit, though it is not
// destruction — by sacrifice.go.
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
// So the clear moved to the LANDED outcome, and #816 moved it one step
// further down into MoveCard's battlefield-exit cleanup, which is the
// landed outcome of EVERY exit rather than of this one: same
// terminal-outcome shape damage_tail.go and life_tail.go use, one
// place. Regeneration, when it ships, removes the damage in its own
// replacement (CR 701.15a says the shield does it), not as a side
// effect of the destroy path.
//
// If the owner is no longer seated, the card lands in exile so the
// engine doesn't carry a stale reference. Caller must hold g.mu.
//
// The FIRE-AND-FORGET form. A caller that has to know what the exit
// actually did — "for each creature destroyed this way" — uses
// routeBattlefieldExitThenLocked and reads the board from the
// continuation, because an exit can pause on a replacement's prompt
// (a commander's CR 903.9b offer for a bounce or a tuck; a death is
// no longer one, ADR 0115) (#815).
func (g *Game) routeBattlefieldCardToOwnerGraveyardLocked(cardID uuid.UUID) error {
	return g.routeBattlefieldExitThenLocked(cardID, nil)
}

// routeBattlefieldExitThenLocked is the same exit with a CONTINUATION:
// `then` runs once the move has reached a TERMINAL outcome — landed
// (wherever the window settled it), cancelled, or refused. A pause is
// not terminal; the resume reaches it later, through
// applyResolvedReplacementEventLocked.
//
// It is the destroy / sacrifice / SBA route's half of the idiom
// lifeTail, damageTail and zoneRoute.then already share (ADR 0013 §5b,
// §5c, §5g). The continuation is carried on a zoneRoute like every
// other exit's, flagged ViaBattlefieldLeave so the resume still
// finishes the move through executeBattlefieldLeaveLocked: this route
// is not folded into the shared exit primitive (see zone_route.go),
// only its "then" is.
//
// `then` deliberately takes no outcome argument. What it wants to know
// is where the permanent ended up, and it reads that off the live
// board — the undo-safety contract every continuation in the engine
// follows, and the only reading that is still true after a rewind into
// the open prompt.
//
// Caller must hold g.mu.
func (g *Game) routeBattlefieldExitThenLocked(cardID uuid.UUID, then func(g *Game) error) error {
	return g.routeBattlefieldExitInBatchThenLocked(cardID, battlefieldExitRoute, nil, then)
}

// routeBattlefieldExitInBatchThenLocked is the batched form used by a
// sequenced simultaneous destruction. `batch` rides the route across a
// replacement prompt so finishBattlefieldLeaveLocked can publish it around
// the resumed move itself, before the continuation starts the next leg.
//
// `r` is the route TEMPLATE the verb chose — destroyRoute for a
// destruction, battlefieldExitRoute for a sacrifice or a zero-counter
// state-based action. It is what tells the CR 701.19 regeneration
// built-in whether this exit is a destruction at all, and whether the
// destroying effect said it can't be regenerated (#667). Its CardID,
// simultaneousExit and then are filled in here, exactly as
// routeLegLocked fills them in for every other exit.
//
// Caller must hold g.mu.
func (g *Game) routeBattlefieldExitInBatchThenLocked(cardID uuid.UUID, r zoneRoute, batch []Card, then func(g *Game) error) error {
	var owner *Player
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			owner = g.playerByIDLocked(g.Battlefield.Cards[i].Owner)
			break
		}
	}

	// S17 sub-PR 6: route through the CR 614 replacement pipeline
	// so every replacement sees dies-to-damage + wrath + SBA
	// destroys. (Until ADR 0115 the CR 903.9 commander-zone built-in
	// fired here and asked the owner before the card moved. It is now
	// a hand-and-library replacement, CR 903.9b: a dying commander
	// lands in the graveyard and CR 903.9a's state-based action asks
	// afterwards.)
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
		// #667: what KIND of exit this is, and the rider the
		// destroying effect printed. Carried on the event rather than
		// re-derived downstream, because every exit off the
		// battlefield ends in the same graveyard by the same route
		// and nothing else can tell a destruction from a sacrifice.
		Destruction:       r.Destruction,
		CantBeRegenerated: r.CantBeRegenerated,
		// A sacrifice paid as a cost is an indivisible CR 602.2b
		// payment. Carry the route's posture onto the replacement
		// event just as routeCardToZoneLocked does for non-battlefield
		// moves, so a CR 616 ordering window settles inline.
		mustSettleNow: r.MustSettleNow,
		// #1397: a sacrifice paid as a cost carries its owner's
		// CR 903.9 answer, given before the payment.
		commanderAnswer: r.commanderAnswer,
	}
	if then != nil {
		r.CardID = cardID
		r.ViaBattlefieldLeave = true
		r.simultaneousExit = batch
		r.then = then
		ev.zoneRoute = &r
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// Prompt queued; the resume runs the move after the owner
		// answers, and the continuation goes with it — a pause is not
		// terminal. Return nil so the SBA caller reports no failure.
		return nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		if tailErr := g.runRouteTailLocked(ev.zoneRoute); tailErr != nil {
			g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: tailErr.Error()})
		}
		return err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		// CR 614: the permanent is not going anywhere — indestructible
		// granted mid-window, "it isn't destroyed instead". Terminal,
		// so a caller sequencing a batch through the continuation is
		// told; it reads the board, finds the permanent still on the
		// battlefield, and counts no destruction (#815).
		return g.runRouteTailLocked(ev.zoneRoute)
	}
	return g.finishBattlefieldLeaveLocked(out, owner)
}

// finishBattlefieldLeaveLocked performs a settled destroy / sacrifice /
// SBA exit and then runs the route's continuation, if it is carrying
// one. The two callers are the unpaused path above and the CR 903.9
// resume in applyResolvedReplacementEventLocked, so a destruction that
// paused and one that did not reach the continuation in the same place
// — after the card has moved and its events are out.
//
// The tail runs even when the move failed: a refused move is as
// terminal as a completed one, and a caller waiting on it must not be
// left waiting. #815.
//
// Caller must hold g.mu.
func (g *Game) finishBattlefieldLeaveLocked(ev *ReplacementEvent, owner *Player) error {
	closeBatch := func() {}
	if ev.zoneRoute != nil {
		closeBatch = g.publishSimultaneousExitLocked(ev.zoneRoute.simultaneousExit)
	}
	defer closeBatch()
	moveErr := g.executeBattlefieldLeaveLocked(ev.CardID, ev.NewZone, ev.NewZoneOwner, owner, ev.ShuffleDestinationLibrary, ev.ExiledWith)
	tailErr := g.runRouteTailLocked(ev.zoneRoute)
	if moveErr != nil {
		return moveErr
	}
	return tailErr
}

// executeBattlefieldLeaveLocked runs the physical zone move after
// the replacement pipeline has settled on a destination. Factored
// out of routeBattlefieldCardToOwnerGraveyardLocked so the resume
// path (ResolveOptionalReplacement → applyResolvedReplacementEventLocked)
// shares the same implementation.
//
// #708 / #816: the marked damage and the CR 702.2c deathtouch flag go
// with the move rather than with the destruction, and MoveCard's
// battlefield-exit cleanup is where that happens now — one exit
// cleanup for every route off the battlefield. A destruction that a
// replacement turned into something else never reaches a move at all,
// and a permanent that is still on the battlefield keeps its damage
// until the cleanup step (CR 514.2) like every other damaged
// permanent.
//
// shuffleAfter is ADR 0013 §5ah's ShuffleDestinationLibrary carried
// down as a plain bool, since this function takes the settled
// destination apart from the *ReplacementEvent it came from. True
// only for a replacement that redirected NewZone to ZoneLibrary as a
// genuine shuffle-in (Blightsteel Colossus, the Eldrazi titans)
// rather than a placement; every other caller passes false.
//
// exiledWith is ReplacementEvent.ExiledWith carried down the same way
// (#2530): stamped onto the card as Card.ExiledWith once it has landed
// in exile, and ignored for any other destination. The zero ref —
// every caller but a replacement that declared the link — stamps
// nothing.
//
// Caller must hold g.mu.
func (g *Game) executeBattlefieldLeaveLocked(cardID uuid.UUID, dest ZoneKind, destOwner uuid.UUID, owner *Player, shuffleAfter bool, exiledWith PermissionCardRef) error {
	var destZone *Zone
	var actor uuid.UUID
	switch dest {
	case ZoneCommand:
		// A CR 903.9b commander-zone replacement landed. Find the
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
	// The exit runs here, while the permanent is still on the
	// battlefield with everything that was true of it: the CR 603.10
	// LKI snapshot an LTB trigger is judged on — including the damage
	// that killed it, which MoveCard's exit cleanup zeroes a line
	// later (#816) — and the Game-side forget of what this object did
	// this turn (#630, CR 400.7). See battlefield_exit.go.
	lki := g.battlefieldExitLocked(cardID)
	if _, err := MoveCard(g.Battlefield, destZone, cardID); err != nil {
		return err
	}
	g.markCardKnownInZoneLocked(destZone, cardID)
	// ADR 0013 §5ah: the card has now actually landed in `dest`, so a
	// caller that asked for a shuffle-in gets it here — the one place
	// both this route and executeZoneRouteLocked perform the physical
	// landing. ShuffleLibraryForEffect also clears every card's
	// Known-by in that library (its own #1335 contract), which is what
	// makes this the honest "reveal it, then it's hidden again" rather
	// than a card that keeps the knowledge a battlefield death gave it.
	if shuffleAfter && dest == ZoneLibrary {
		_ = g.ShuffleLibraryForEffect(destZone.Owner)
	}
	stampExiledWithLocked(destZone, cardID, exiledWith)
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		Actor:   actor,
		CardID:  cardID,
		OldZone: ZoneBattlefield,
		NewZone: dest,
	})
	ltb := Event{Kind: EventLTB, CardID: cardID, Actor: actor, NewZone: dest}
	lki.stamp(&ltb)
	g.EmitEvent(ltb)
	// A permanent leaving the battlefield is the one event that can
	// invalidate a queued "sacrifice a creature of your choice" prompt
	// (Grave Pact), so re-check them here rather than in
	// runStateChecksLocked — a queued choice stops priority from
	// passing, so the state-check loop is exactly what does NOT run
	// while such a prompt is outstanding. No-op when none is queued.
	g.pruneSacrificeChoicesLocked()
	// #1045: and any choose-cards prompt that still offers it — a
	// "choose two permanents" whose candidates are leaving one at a
	// time — for the same reason again.
	g.pruneCardSetChoicesLocked()
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
		damageTail: &damageTail{
			kind:      damageTailManualMark,
			sourceLKI: g.damageSourceLKILocked(source),
		},
	}
	// ADR 0108 PR 0: a combat mark joins the combat damage step's
	// instance; any other mark is an instruction of its own.
	if isCombat {
		ev.DamageInstance = g.combatDamageInstanceLocked()
	} else {
		g.stampEffectDamageInstanceLocked(ev, 0)
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
// ordering prompt (S19 sub-PR 8), or behind a trigger of the same
// batch that is still choosing its targets, modes or "you may"
// (#1529) — the caller's loop should stop spinning until the answer
// (ResolveTriggerOrder, ResolvePickTargets, ResolveModePick,
// ResolveTriggerPrompt) re-runs the drain. Returns
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
	// #1529 / CR 603.3b: a triggered ability still being put on the
	// stack (its "you may", its mode pick or its target pick is open)
	// belongs to this batch. Hold the WHOLE queue until it has joined,
	// so its controller orders it with the rest and the APNAP
	// placement below sees every seat's complete batch. Draining
	// around the prompt is what put a targeted trigger above its
	// untargeted siblings every time. See ADR 0018's #1529 amendment.
	if g.triggerAnnouncementOpenLocked() {
		return false
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
		if seat == -1 || g.Seats[seat].Eliminated {
			// Controller no longer seated, or has since left the game
			// (CR 800.4a) — drop the trigger. #864: this specifically
			// includes a trigger that lands here for a player the SBA
			// loop already eliminated earlier in the SAME pass (a
			// creature destroyed after its controller's life hit 0
			// queues a "dies" trigger against the now-eliminated
			// controller). Dropping it here, before it can ever reach
			// seatNeedsTriggerOrder below, matters beyond tidiness:
			// QueueChoiceForEffect now refuses to queue an ordering
			// prompt for an eliminated chooser, but the loop below
			// still marked the WHOLE APNAP drain `held` for that seat
			// before finding that out, which would have wedged every
			// other seat's triggers behind an ordering prompt that
			// could never be created, let alone answered.
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
		p := g.Seats[seat]
		if !seatNeedsTriggerOrder(items, p.TriggerOrder) {
			continue
		}
		held = true
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
	var announce []*StackItem
	for offset := 0; offset < numSeats; offset++ {
		seat := (g.Turn.ActiveSeat + offset) % numSeats
		for _, t := range bySeat[seat] {
			t.Seq = seq
			seq++
			g.StackMeta[t.ID] = t
			if t.TargetsAnnouncePending {
				announce = append(announce, t)
			}
		}
	}
	g.recomputeSplitSecondLocked()
	g.announcePlacedTargetsLocked(announce)
	return true
}

// announcePlacedTargetsLocked emits the CR 115.3 "becomes the target"
// events a placed batch owes (StackItem.TargetsAnnouncePending, #1539),
// in placement order, once EVERY item of the batch is on the stack.
//
// CR 603.3b: "Then the game once again checks for and performs
// state-based actions until none are performed, then abilities that
// triggered during this process go on the stack." A ward or Monk
// Gyatso trigger harvested here lands on a PendingTriggers the drain
// has already emptied, so it is the NEXT batch: the caller's
// runStateChecksLocked loop runs the state-based actions again and
// puts it on the stack above everything placed here — whichever seat
// controls it, and outside the targeting player's ordering prompt.
//
// Caller must hold g.mu.
func (g *Game) announcePlacedTargetsLocked(placed []*StackItem) {
	for _, t := range placed {
		t.TargetsAnnouncePending = false
		g.emitBecameTargetLocked(t.Controller, t.SourceCardID, t.ID, t.Targets)
	}
}

// seatNeedsTriggerOrder reports whether a seat's batch of pending
// triggers needs a CR 603.3b ordering prompt, given the seat's
// TriggerOrderMode (#1968).
//
// TriggerOrderNever: no. The batch goes on the stack in the order it
// was collected, which is the order every skipped batch has always
// used.
//
// TriggerOrderAlways (#1530): any batch of two or more with an item
// not yet Ordered by an answered prompt, the skips below included.
//
// TriggerOrderWhenItMatters, the default: at least two items, at least
// one not yet Ordered, and an order that could change the game. Three
// shapes are known not to:
//
//   - all identical — same source card and same label. Two Bident
//     draws are interchangeable and asking would be noise.
//   - all commutative — every item is a StackItem.Commutes item, so
//     any order leaves the same board (#1511: a board of prowess
//     creatures, which would otherwise ask on every noncreature
//     spell). An item that commutes still counts only while it has
//     no targets and no modes; Commutes is engine-owned and no such
//     item has either today, so that check is a belt, not the rule.
//   - all copies of one source-blind catalog ability (#1968) — every
//     item names the same catalog row, that row is
//     TriggeredAbility.SourceBlind, and no item carries targets, modes
//     or anything else chosen for it alone (copiesOfOneSourceBlindAbility).
//     Two Soul Wardens, or a set of tokens with the same trigger. The
//     rules argument is in ADR 0018's #1968 amendment.
//
// Anything else prompts, including a batch that is all commutative
// items plus ONE other trigger: where that trigger sits among the
// pumps is a real choice whenever it reads what they change. See
// ADR 0018's #1511 amendment.
//
// An auto-ordered batch keeps its queue order, which is harvest
// order; the drain below places it exactly as it places an answered
// prompt.
func seatNeedsTriggerOrder(items []*StackItem, mode TriggerOrderMode) bool {
	if len(items) < 2 || mode == TriggerOrderNever {
		return false
	}
	if mode == TriggerOrderAlways {
		// #1530: the seat opted out of the skips. Only an already
		// answered batch (every item Ordered) stays out of the prompt.
		for _, t := range items {
			if !t.Ordered {
				return true
			}
		}
		return false
	}
	allOrdered := true
	allSame := true
	allCommute := true
	for _, t := range items {
		if !t.Ordered {
			allOrdered = false
		}
		if t.SourceCardID != items[0].SourceCardID || t.Label != items[0].Label {
			allSame = false
		}
		if !commutesForOrdering(t) {
			allCommute = false
		}
	}
	if allOrdered || allSame || allCommute {
		return false
	}
	return !copiesOfOneSourceBlindAbility(items)
}

// commutesForOrdering is the per-item half of the #1511 skip: the
// item declares it commutes, and it carries nothing a CR 603.3b
// order could interact with — no chosen targets and no chosen
// modes.
func commutesForOrdering(t *StackItem) bool {
	return t.Commutes && len(t.Targets) == 0 && len(t.Modes) == 0
}

// copiesOfOneSourceBlindAbility is the #1968 skip: every item is a
// stamped catalog trigger (Body "catalog/triggered") naming the SAME
// row — the same AbilityRef, so the same catalog key, slot, row and
// label — the running catalog still hands that row back under the ref,
// the row is TriggeredAbility.SourceBlind, and no item carries anything
// chosen or recorded for it alone: no targets, modes, payload, X or
// division. Such items differ only in their source object and their
// trigger context, and a source-blind effect reads neither, so they are
// one effect queued several times and every order resolves the same
// sequence of effects. ADR 0018's #1968 amendment has the argument.
func copiesOfOneSourceBlindAbility(items []*StackItem) bool {
	first := items[0].Params.Ability
	if first == nil {
		return false
	}
	for _, t := range items {
		ref := t.Params.Ability
		if t.Body != CatalogTriggeredBodyKey || ref == nil || *ref != *first {
			return false
		}
		if len(t.Targets) > 0 || len(t.Modes) > 0 || len(t.Payload) > 0 ||
			t.XValue != 0 || len(t.Distribution) > 0 {
			return false
		}
	}
	row, _, outcome := resolveTriggeredAbilityRef(*first)
	return outcome == abilityRefMatched && row.SourceBlind
}

// triggerAnnouncementOpenLocked reports whether some triggered
// ability is still on its way onto the stack: a prompt is open that
// the harvest queued for it and whose answer ends with the ability
// joining PendingTriggers. Those are the CR 603.5 optional-trigger
// yes/no (triggerResume), the CR 603.3c mode pick (modePickResume)
// and the CR 603.3d target pick (pickTargetResume). A pick_target for
// a spell COPY (copyResume, CR 707.10c) is not one: it builds no
// triggered ability.
//
// Every such prompt blocks the table (choice_gate.go), so holding the
// drain behind one stops nothing that was not already stopped.
//
// Caller must hold g.mu.
func (g *Game) triggerAnnouncementOpenLocked() bool {
	for _, c := range g.PendingChoices {
		if c != nil && (c.triggerResume != nil || c.pickTargetResume != nil || c.modePickResume != nil) {
			return true
		}
	}
	return false
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
// pending map drains, the cleanup step's turn-based actions are
// finished, so it takes the one cleanup exit (exitCleanupStepLocked,
// cleanup.go) — which either ends the turn or, if this discard put a
// trigger on the queue, gives the active player priority right here
// (CR 514.3a).
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
	// The discard itself is the one discard path (discard.go); the
	// hand-size bookkeeping is this site's own and runs once the whole
	// batch has landed.
	return g.discardCardsLocked(playerID, cardIDs, discardOptions{
		cause: DiscardCauseCleanup,
		then: func(g *Game, _ []uuid.UUID) error {
			delete(g.DiscardPending, playerID)
			if len(g.DiscardPending) == 0 {
				g.DiscardPending = nil
			}
			// The cleanup step's turn-based actions are finished now
			// that the discard has landed, so take the one cleanup
			// exit (cleanup.go): with the pending map empty it runs
			// the CR 514.3a check and then either ends the turn or
			// gives the active player priority here. This is where a
			// trigger watching the hand-size discard gets onto the
			// stack in the turn it belongs to (#661) — before, this
			// re-fired the whole step entry, which advanced straight
			// out of the turn and left the trigger waiting for the
			// next player's upkeep.
			if g.Turn.Step == StepCleanup {
				g.exitCleanupStepLocked()
			}
			return nil
		},
	})
}

// SetMaxHandSize updates the named player's per-player hand-size
// cap. Sandbox-only — used by the S14+ effect catalog for
// Reliquary Tower / Thought Vessel / Library of Leng / Spellbook /
// Null Profusion / Venser's Journal style cards. Until the catalog
// lands, this is also exposed as a manual sandbox helper for
// playgroup adjustments.
//
// `value` is clamped to NoMaxHandSize (-1) for "no cap"; any other
// negative value is rejected with ErrInvalidParam. Stamped like
// SetMaxHandSizeForEffect, so the sandbox value is folded in CR
// 613.11's timestamp order (ADR 0113 §3 decision 3). Doesn't fire
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
	p.MaxHandSizeAt = timeNowUnixNano()
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
	if err := g.counterSpellLocked(spellID, dst); err != nil {
		return err
	}
	// #2275: the sandbox's hand counter stands in for a resolution the
	// table worked out itself, and it changes what is on the stack.
	// Whatever the seats passed over is not there any more, so the
	// succession starts again (CR 117.4).
	g.restartPassSuccessionLocked()
	return nil
}

// CounterAbility removes an activated / triggered ability from the
// stack. Abilities cease to exist on resolution (CR 608.2n); a
// counter is the same destinationless removal (CR 701.6a). Returns
// ErrCardNotOnStack if the ID doesn't reference an ability item.
//
// Caller must NOT hold g.mu — this method takes the write lock.
//
// S13.1. #1211 folded the body into counterAbilityLocked: this was a
// second copy of the deletion and the emitted event, which is the
// #529 shape exactly — two counter paths that had already drifted
// once on flashback. One deletion, one event.
func (g *Game) CounterAbility(abilityID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if err := g.counterAbilityLocked(abilityID); err != nil {
		return err
	}
	// #2275: as CounterSpell — the stack changed under the passes.
	g.restartPassSuccessionLocked()
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
// S13.1, amended by ADR 0115 decision 5: when asCommander is true and
// the card is a commander moved to a graveyard or exile, it lands there
// (it dies, if it came from the battlefield) and is then put into its
// owner's command zone with CR 903.9a's question pre-answered "yes". A
// move to a hand or a library still asks the CR 903.9b replacement.
func (g *Game) MoveCardByID(src, dst ZoneRef, cardID uuid.UUID) error {
	return g.MoveCardByIDAsCommander(src, dst, cardID, false)
}

// MoveCardByIDAsCommander is the S13.1 extended form. asCommander
// is the player's "yes, send this commander back to the command zone"
// choice that the move_card action exposes via a per-request flag: for
// a graveyard or exile destination it answers CR 903.9a in advance
// (ADR 0115 decision 5). The default-false form preserves the historic
// MoveCardByID behaviour for non-commander moves.
func (g *Game) MoveCardByIDAsCommander(src, dst ZoneRef, cardID uuid.UUID, asCommander bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.moveCardByRefLocked(src, dst, cardID, asCommander, false)
}

// moveCardByRefLocked is the body both sandbox move verbs share —
// MoveCardByIDAsCommander and, with toBottom set, MoveCardByIDToBottom.
//
// It splits the move by what it IS (#707). A destination that is not
// the battlefield or the stack is an EXIT, and goes through the shared
// exit primitive (routeCardToZoneLocked, zone_route.go): that opens the
// CR 614 window with a zoneRoute on the event, so when the CR 903.9
// "put your commander in the command zone instead?" prompt pauses the
// move, the primitive's resume finishes it from whatever zone the card
// was in — graveyard, hand, library, exile, the stack or the
// battlefield — and honours what this move asked for (to the bottom of
// the library, retire the stack item) on the far side of the prompt.
// Before #707 this function ran its own pipeline call with no resume
// behind it, so a paused move from anywhere but the battlefield was
// silently dropped: the prompt closed and the card never moved.
//
// A destination of battlefield or stack is an ENTRY, and stays inline
// below: it owes enters-tapped, enters-with-counters and the ETB fire,
// none of which an exit can express, and no CR 903.9 destination is
// among them.
//
// Caller must hold g.mu.
func (g *Game) moveCardByRefLocked(src, dst ZoneRef, cardID uuid.UUID, asCommander, toBottom bool) error {
	if g.State != StateActive {
		return ErrGameNotActive
	}
	srcZone := g.zoneFromRefLocked(src)
	if srcZone == nil {
		return ErrZoneNotFound
	}
	if dstZone := g.zoneFromRefLocked(dst); dstZone == nil {
		return ErrZoneNotFound
	} else if srcZone == dstZone {
		// A no-op move: MoveCard would Remove the card and re-Push it,
		// and from the battlefield that would run the CR 400.7 exit
		// cleanup over a card that never left. The pipeline must not
		// fire for a move that isn't one either, so this returns
		// before the window — but the card still has to be where the
		// caller says it is, so a bogus instance ID is still an error.
		if !srcZone.Contains(cardID) {
			return ErrCardNotFound
		}
		return nil
	}
	// The exit primitive finds the card by scan, so the src ref would
	// otherwise be advisory; a stale client request naming the zone the
	// card has already left must still fail rather than move it out of
	// wherever it ended up.
	if !srcZone.Contains(cardID) {
		return ErrCardNotFound
	}

	if dst.Kind != ZoneBattlefield && dst.Kind != ZoneStack {
		paused, err := g.routeCardToZoneLocked(zoneRoute{
			CardID:   cardID,
			Dst:      dst.Kind,
			DstOwner: dst.Owner,
			ToBottom: toBottom,
			// A card moved off the stack by hand is no longer a spell
			// on the stack; the route retires its StackMeta entry
			// because the source zone is the stack (#1318), which is
			// what this used to have to ask for by hand.
			AsCommander: asCommander,
			// #1320: nobody's spell or ability moved it.
			Cause: MoveCause{Kind: MoveCauseManual},
		})
		if err != nil {
			return err
		}
		if paused {
			// The CR 903.9 prompt (or a CR 616 ordering prompt) owns
			// the move now. Nothing has moved, so there is nothing for
			// the state checks to see; the resume lands the card.
			return nil
		}
		// ADR 0115 decision 5: as_commander pre-answers CR 903.9a's
		// "yes" for a commander this move put into a graveyard or exile.
		// It died (or was exiled) like any other card first.
		if asCommander {
			g.returnCommanderPreAnsweredLocked(cardID)
		}
		// A sandbox move is a special action: the mover keeps priority
		// afterwards (CR 116.3), and CR 117.5 puts SBAs + the APNAP
		// trigger drain at that boundary.
		g.runStateChecksLocked()
		return nil
	}

	// S17 sub-PR 2: route through the replacement pipeline. The
	// commander-zone built-in (commanderZoneReplacement) gates on
	// card.IsCommander && an eligible destination — NOT on
	// ev.asCommanderMove, which stopped being a gate in #171; the
	// flag survives here only as a routing flavor on the manual
	// move_card action. A battlefield / stack destination is not one
	// of CR 903.9's four, so the built-in cannot fire on this half —
	// what can still pause it is a CR 616 ordering prompt between two
	// entry replacements, which no catalog card produces today and
	// which this entry site has never been able to resume
	// (entryResumable is off; see the field's doc comment).
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
		// A replacement rewrote the destination back to where the card
		// already is. Nothing to do.
		return nil
	}
	var lki exitLKI
	if srcZone.Kind == ZoneBattlefield {
		// LKI, and the CR 400.7 forget — battlefield_exit.go. The
		// sandbox move is a battlefield exit like any other: a
		// planeswalker shoved to hand from the context menu is as new
		// an object when it comes back as one Venser bounced.
		lki = g.battlefieldExitLocked(cardID)
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
		g.applyEntryCountersLocked(cardID, out.EntersWithCounters)
		// ADR 0090, CR 722.3a — the sandbox move lands a preparation
		// card prepared exactly as any other entry does.
		g.applyEntersPreparedLocked(cardID, out.EntersPrepared)
	}

	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		CardID:  cardID,
		OldZone: srcZone.Kind,
		NewZone: dstZone.Kind,
	})
	if srcZone.Kind == ZoneBattlefield {
		ltb := Event{Kind: EventLTB, CardID: cardID, NewZone: dstZone.Kind}
		lki.stamp(&ltb)
		g.EmitEvent(ltb)
	}
	if dstZone.Kind == ZoneBattlefield {
		g.EmitEvent(Event{Kind: EventETB, CardID: cardID, EnteredFrom: srcZone.Kind, EnteredFromOwner: srcZone.Owner})
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
	// #1069: the entry side's prune door (battlefield_entry.go). This
	// is the one landing that also serves the STACK, and the prune does
	// not care which of the two it was: the question is whether a card
	// an open choose_cards prompt still offers is in the zone that
	// prompt picks from, and this move has just taken it out of one.
	// The exit half of this function reaches the same prune through
	// executeZoneRouteLocked, which refuses a battlefield or stack
	// destination (routeDestinationLocked), so no move can run it
	// twice.
	g.pruneChoicesAfterArrivalLocked()
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
// happens inline: a pipe set keeps its printed width with the
// controller's commander's colour identity listed first, and only an
// ability that set NarrowToCommanderIdentity (the four cards whose text
// says "in your commander's color identity") is intersected with it —
// see manaPickOptionsFor. CR 903.4f (#844): when that intersection is
// empty — no commander, or a colourless one — the slot adds no mana
// and queues no prompt. The cost is still paid, because the ability is
// still an ability the player may activate; it simply does nothing,
// which is what the rulings on those four cards say. Nothing OFFERS
// such an activation (ManaAbilityAddsNoMana gates the legal-move
// enumerator, the auto-tapper and the client's menu).
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
// #1370 adds Rider's mirror image, ManaAbilityShape.PreRider:
// everything the oracle text says BEFORE the "Add …" clause, run
// once the cost is paid and before the produced string is computed —
// Empowered Autogenerator's "Put a charge counter on this artifact,"
// so its own "X is the number of charge counters" reads the count
// after the placement rather than guessing at it.
//
// Any activation that pays life, sacrifices, or fires a rider (either
// direction) runs a state-based-action pass on the way out, so a
// player who taps Ancient Tomb at 2 life loses here rather than at
// the next priority boundary.
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
	// Ref is the stable ref of the row the activator meant (ADR 0093
	// Decision 5): "own:<i>", "land:<colour>", "grant:<bundle>:<i>:<n>".
	// A mismatch with the row at the index is ErrStaleAbilityRef,
	// before anything is paid. Empty is accepted — the auto-tapper and
	// every client that predates the field send none.
	Ref string

	// SacrificeIDs names the permanents paying a SacrificeOther
	// component. Exactly one for a single-permanent clause; empty
	// when the ability has no such cost.
	SacrificeIDs []uuid.UUID

	// CounterSourceIDs / CounterCounts / CounterKind are the
	// announce-time payment for a RemoveCounters component (#789),
	// with exactly the meanings ActivateAbilityParams gives them —
	// one component, one payment shape, whichever ability kind
	// carries it. Empty for the common self form with a printed
	// kind and a printed count, which is every Vivid land and Ramos.
	CounterSourceIDs []uuid.UUID
	CounterCounts    []int
	CounterKind      string
	CounterKinds     []string

	// TapIDs names the permanents paying a TapOthers component
	// (#758), with exactly the meaning
	// ActivateAbilityParams.TapIDs gives them — one component, one
	// payment shape, whichever ability kind carries it. Empty for
	// every mana ability that does not print the clause, which is
	// all of them but Springleaf Drum's family.
	TapIDs []uuid.UUID

	// DiscardIDs names the cards paying a DiscardCards component
	// (#1213), with exactly the meaning ActivateAbilityParams.DiscardIDs
	// gives them — one component, one payment shape, whichever
	// ability kind carries it. Skirge Familiar's "Discard a card: Add
	// {B}" is the only printed one; every other mana ability sends
	// none.
	DiscardIDs []uuid.UUID

	// ExileIDs names the cards paying an ExileCards component (#1283)
	// — Cadaverous Bloom's "Exile a card from your hand". A separate
	// list from DiscardIDs because the two are separate components
	// with separate exits (game.ExileCost); every other mana ability
	// sends none.
	ExileIDs []uuid.UUID

	// ExilePermanentIDs names the permanents paying an ExilePermanents
	// component (#1600) — Food Chain's "Exile a creature you control",
	// with exactly the meaning ActivateAbilityParams.ExilePermanentIDs
	// gives them. On the wire as `exile_permanent_ids`.
	ExilePermanentIDs []uuid.UUID

	// Colors names, up front, the colour each PICKING slot of the
	// output adds (#1443): one entry per entry of
	// ManaAbilityColorOptions, in output order — a painland's "{R|W}"
	// takes one, a filter land's "{W|U}{W|U}" two. Each must be one the
	// slot offers right now, checked before anything is paid
	// (ErrIllegalManaColor). A named slot is produced straight into
	// the pool and queues no mana_pick.
	//
	// Empty is the ordinary activation, unchanged: every picking slot
	// queues its mana_pick. The auto-tapper and the bot seats never
	// set it.
	Colors []string

	// AutoTap lets the activation pay a MANA component of its cost
	// (Crystal Quarry's "{5}, {T}", a Signet's "{1}, {T}") by tapping
	// the activator's other mana sources for whatever the floating pool
	// is missing (#2215) — the same planner, top-up and executor every
	// other auto-tapped payment uses. CR 605.3a lets a player activate
	// mana abilities while paying for one. The source itself, and every
	// card another component of this cost names, is never spent on it
	// (ManaActivationAutoTapExclusions). Without it, the mana has to be
	// floating already, as before. On the wire as `auto_tap`; the client
	// sends it on every mana activation, and the legal-move enumerator
	// on every one with a mana component.
	AutoTap bool

	// PhyrexianLife is how many of the mana component's symbols the
	// activator pays 2 life each for instead of the mana (CR 107.4f,
	// CR 602.2b, ADR 0131 §2): a printed {B/P} on a mana ability's
	// cost, or a {B} that K'rrik, Son of Yawgmoth lets its controller
	// pay with life — a filter land's "{B}, {T}: Add {B}{B}". The same
	// field and wire name (`phyrexian_life`) as
	// CastSpellParams.PhyrexianLife and ActivateAbilityParams.PhyrexianLife,
	// through the same strike-and-pay helper pair, so the ceiling, the
	// CR 119.4 / 119.8 gate and the choice of which symbols to strike
	// are one piece of code. Claiming one against a mana ability with no
	// mana component, or more than the cost has symbols for, is refused
	// before anything is paid. Zero pays the symbols with mana, as ever;
	// the auto-tap never claims it (CR 601.2b).
	PhyrexianLife int

	// commanderAnswers are the CR 903.9 answers the owners of the
	// commanders this activation's cost moves gave before it began
	// (#1397, cost_commander_choice.go). Unexported: only the parked
	// activation's resume sets it.
	commanderAnswers map[uuid.UUID]bool
}

func (g *Game) ActivateManaAbility(playerID, cardID uuid.UUID, abilityIdx int, params ManaAbilityParams) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.activateManaAbilityLocked(playerID, cardID, abilityIdx, params); err != nil {
		return err
	}
	// #2275 / CR 117.3c: a mana ability is an ability, and the
	// priority holder who activates one has taken an action — the
	// passes before it no longer count. Mana tapped by a player who
	// does NOT hold priority (to answer a "pay {1}" prompt) is not an
	// action in that sense, and noteActionTakenLocked leaves the
	// succession alone.
	g.noteActionTakenLocked(playerID)
	return nil
}

// activateManaAbilityLocked is ActivateManaAbility's body, split out
// so a parked activation (#1397) can be made again from the CR 903.9
// answer's resume. Caller must hold g.mu.
func (g *Game) activateManaAbilityLocked(playerID, cardID uuid.UUID, abilityIdx int, params ManaAbilityParams) error {
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
	// #1228 / CR 113.6: the source is found in WHATEVER zone holds
	// it, not on the battlefield alone — the same move
	// ActivateCatalogAbility made for CR 602 abilities in #660. A
	// Spirit Guide's "Exile this card from your hand: Add {R}" is an
	// ordinary CR 605 mana ability whose ability functions from a
	// hand, not a fork.
	card, srcZone := g.findCardAndZoneLocked(cardID)
	if card == nil {
		return ErrCardNotFound
	}
	if srcZone == ZoneBattlefield {
		if card.Controller != playerID {
			return ErrCardCallerMismatch
		}
		// CR 602.5: an effect that stops this permanent's activated
		// abilities being activated stops its mana abilities too,
		// when it says so. Arrest does; Faith's Fetters explicitly
		// does not ("unless they're mana abilities"), which is why
		// the two are separate bits — see restrictions.go.
		if !CanActivateManaAbilities(card) {
			return ErrCantActivate
		}
	} else if card.Owner != playerID {
		// CR 108.4: a card outside the battlefield and the stack has
		// no controller, so its OWNER is the "you" of the printed
		// text. The same check the CR 602 path makes, asked of the
		// field that means something here.
		//
		// CanActivateManaAbilities is deliberately NOT asked, for the
		// reason CanActivateAbilities is not asked one path over:
		// Arrest and Cursed Totem apply to a permanent, and layer 6
		// has nothing to say about a card in a hand.
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
	abilities, origins := manaAbilityRows(card, true)
	// ADR 0093 Decision 5: the ref names the row the activator meant.
	// A grant appearing or vanishing since the view moved the rows;
	// refuse the stale move before anything is paid (#544).
	if staleAbilityRef(params.Ref, abilityIdx, len(abilities), origins) {
		return ErrStaleAbilityRef
	}
	if abilityIdx < 0 || abilityIdx >= len(abilities) {
		return ErrInvalidParam
	}
	ab := abilities[abilityIdx]
	// CR 113.6 (#1228): the ability has to function from the zone the
	// card is in. After the index lookup, so the index judged is the
	// one the view and the enumerator published, and before
	// everything else, so a Forest's "{T}: Add {G}" fired out of a
	// hand — or a Spirit Guide's ability fired off the battlefield —
	// costs nothing and pays nothing.
	if !ManaAbilityFunctionsFromZone(ab, srcZone) {
		return ErrActivationZoneNotAllowed
	}
	// --- gate ----------------------------------------------------
	//
	// #1210, CR 602.5: the board-wide "can't be activated" gate, the
	// same one ActivateCatalogAbility calls and the same placement —
	// before anything is validated or paid. ActivationAbility.Mana is
	// true here and the RESTRICTION decides what that means: Pithing
	// Needle exempts mana abilities, Cursed Totem does not. See
	// activation_gate.go on why the exemption is not a property of
	// this call site.
	if err := g.ActivationGateLocked(playerID, *card, srcZone, ActivationAbility{Label: ab.Label, Mana: true}); err != nil {
		return err
	}
	// #1183: "Activate each exhaust ability only once", the mana
	// half of #1181. The key is taken HERE, before anything is
	// validated or paid, for the reason activationTallyKeyLocked
	// gives: a cost that moves the source (a Lotus Petal's sacrifice,
	// a self-bounce) ends the object and carries Card.ObjectEpoch
	// with it, so a key built after the payment would be written
	// against an object that never had the ability. The same key is
	// read now and written at the bottom, so the gate and the record
	// cannot address different things.
	//
	// Before the Condition, which is the order ActivateCatalogAbility
	// checks in: a card that prints both is refused as exhausted
	// rather than as ungated, because that is the answer that will
	// still be true tomorrow.
	activationKey := g.activationTallyKeyLocked(cardID, ab.Label)
	if g.ManaAbilityExhausted(playerID, cardID, ab) {
		return ErrAbilityExhausted
	}
	// "Activate only if you control five or more lands" (Temple of
	// the False God), "…three or more artifacts" (Mox Opal). CR
	// 602.5: an activation restriction is checked before anything
	// is paid, so a failed gate costs the player nothing.
	if ab.Condition != nil && !ab.Condition(g, playerID, cardID) {
		return ErrConditionNotMet
	}
	// #1443: a colour named up front has to be one the slot offers,
	// asked of the same list the view published and the mana_pick
	// would have carried. Before any cost is validated, so a refused
	// colour costs nothing and taps nothing.
	if err := g.validateUpfrontManaColors(playerID, cardID, ab, params.Colors); err != nil {
		return err
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
	// #1213: a mana ability announces no X (CR 605.3b — there is no
	// stack item to carry one), so the announced value is 0 and
	// effects.Register refuses a CountFromX sacrifice clause here at
	// boot. An OPEN count ("sacrifice one or more") still works, and
	// reads its floor and ceiling off the clause alone.
	sacrifices, err := g.validateSacrificeCostLocked(playerID, cardID, AbilityCost{
		SacrificeSelf:  ab.SacrificeCost,
		SacrificeOther: ab.SacrificeOther,
	}, params.SacrificeIDs, 0)
	if err != nil {
		return err
	}
	if !g.CanPayLifeLocked(p, ab.LifeCost) {
		// CR 119.4: you can't pay more life than you have. Paying
		// down to exactly 0 is legal and the SBA loop ends the game
		// after. CR 119.8: a player whose life total can't change
		// cannot pay any of it (#1200). Same gate
		// ActivateCatalogAbility applies to AbilityCost.Life, through
		// the same predicate.
		return ErrInvalidParam
	}
	// ADR 0129 §5, CR 118.3: "Pay {E}" (Aether Hub). Checked with the
	// life, so a player short of energy taps nothing.
	if err := EnergyShortfall(p, ab.EnergyCost); err != nil {
		return err
	}
	// #789: the counter components, validated by the SAME functions
	// the CR 602 activated path uses, because it is the same
	// component — a Vivid land's charge counter and Heart of Kiran's
	// loyalty counter are one cost shape with two owners.
	counters, err := g.validateCounterRemovalLocked(playerID, cardID, ab.RemoveCounters, CounterCostPayment{
		SourceIDs: params.CounterSourceIDs,
		Counts:    params.CounterCounts,
		Kind:      params.CounterKind,
		Kinds:     params.CounterKinds,
	})
	if err != nil {
		return err
	}
	if !g.canPlaceCounterLocked(playerID, cardID, ab.AddCounter) {
		return ErrCantPayCounterCost
	}
	// #758: the tap-another component, validated by the SAME
	// function the CR 602 activated path uses, for the same reason
	// the counter components are — Springleaf Drum's "tap an
	// untapped creature you control" and Earthcraft's are one cost
	// shape with two owners.
	if err := g.validateTapOthersCostLocked(playerID, cardID, ab.TapOthers, params.TapIDs); err != nil {
		return err
	}
	// #1213: the discard component, validated by the SAME function
	// the CR 602 activated path uses — Skirge Familiar's "Discard a
	// card" and Cryptbreaker's are one cost shape with two owners.
	// The source is a permanent, so the zone passed is the
	// battlefield and the DiscardSelf branch can never fire here
	// (effects.ManaAbilityCost has no such field to set).
	discards, err := g.validateDiscardCostLocked(playerID, cardID, srcZone, AbilityCost{
		DiscardCards: ab.DiscardCards,
	}, params.DiscardIDs, 0)
	if err != nil {
		return err
	}
	// #1283: the exile-a-card component, validated beside the discard
	// it is the sibling of and against it — a card named to both would
	// pay two components (CR 118.3).
	exiles, err := g.validateExileCardsCostLocked(playerID, cardID, ab.ExileCards, params.ExileIDs, discards)
	if err != nil {
		return err
	}
	// #1228: the exile-this component, validated by the SAME rule one
	// ability kind over — the source has to be in the zone the clause
	// names, which for the Spirit Guides is the hand. Free: srcZone
	// is the zone the activation was already validated against.
	if err := g.validateManaExileSelfCostLocked(srcZone, ab); err != nil {
		return err
	}
	// #1600: the exile-a-permanent component (Food Chain), validated by
	// the SAME function the CR 602 path uses, against the sacrifices
	// (the source among them when the cost sacrifices it) — one
	// permanent pays one component (CR 118.3).
	if err := g.validateExilePermanentsCostLocked(playerID, cardID, ab.ExilePermanents, params.ExilePermanentIDs,
		movedSourceAlso(cardID, ab.ExileSelf, sacrifices)); err != nil {
		return err
	}
	// CR 118.3, as on the activated path: a cost that prints both
	// {T} and "tap another untapped creature you control" (Jaspera
	// Sentinel) has already spent the source, so naming it here
	// would pay one of the N with a permanent that is tapping
	// anyway.
	if ab.TapCost && !ab.TapOthers.Empty() {
		for _, id := range params.TapIDs {
			if id == cardID {
				return ErrInvalidParam
			}
		}
	}
	// A mana component in the cost — the Signet cycle's "{1}, {T}",
	// Cabal Coffers' "{2}, {T}". Priced and checked here, spent
	// below with everything else, so an unaffordable Signet fails
	// with the source still untapped.
	//
	// #1191: priced through the CR 601.2f pass rather than parsed
	// straight off the shape — CR 605.1a makes a mana ability an
	// activated ability, so "Exhaust abilities of other permanents you
	// control cost {2} less to activate" (Boom Scholar) reaches Loot,
	// the Pathfinder's mana half exactly as it reaches a CR 602
	// ability. The same pricer the view and the legal-move enumerator
	// read, so a board with no activation-scoped modifier on it costs
	// nothing extra: ManaAbilityManaCostForEffect returns the printed
	// cost unchanged and walks nothing.
	//
	// #2215: with AutoTap, the part the pool cannot pay is planned from
	// the activator's other sources HERE, read-only, so a cost no plan
	// can fund still fails with nothing tapped. The plan is carried out
	// below, once every other component has been validated and just
	// before the first payment (CR 601.2g via CR 602.2b: mana abilities
	// are activated before costs are paid, and CR 605.3a allows it in
	// the middle of activating one). Without AutoTap the mana has to be
	// floating already, which is how this path always worked.
	var (
		manaCost  ParsedCost
		manaPlan  tapPlan
		manaShort ParsedCost
		// manaLife is the life the announced PhyrexianLife claim costs
		// (ADR 0131 §2); paid below, after the fallible mana half is
		// known payable.
		manaLife int
	)
	// A claim against a mana ability with no mana component is a client
	// firing the wrong ability, so it is refused rather than dropped —
	// the check activated.go makes for a CR 602 ability.
	if params.PhyrexianLife != 0 && ab.ManaCost == "" {
		return ErrInvalidParam
	}
	if ab.ManaCost != "" {
		priced, perr := g.ManaAbilityManaCostForEffect(playerID, *card, ab)
		if perr != nil {
			return ErrInvalidParam
		}
		// The mana pays an ACTIVATION (CR 602.2b), and the source
		// permanent's own characteristics are what a restricted
		// token is tested against — Eldrazi Temple mana can fund a
		// colorless Eldrazi's ability, not a Signet's.
		spendCtx := ManaSpendForAbility(*card)
		// #1600: an Orrery's controller may pay a filter land's {W/U}
		// with colourless — the same widening the pay step below
		// spends under, because it spends this manaCost.
		manaCost = g.costAsPaidByLocked(playerID, spendCtx, priced, 0)
		// CR 107.4f / CR 602.2b (ADR 0131 §2): the symbols the activator
		// announced they pay with life leave the mana cost here, through
		// the helper the cast and CR 602 paths run, BEFORE the auto-tap
		// plan below is made — tapping a land for a pip the player said
		// they would pay 2 life for strands it. A mana ability's own
		// "Pay N life" component is checked on top (CR 119.4 is about
		// the total).
		var serr error
		manaCost, manaLife, serr = g.strikePhyrexianLifeLocked(p, card.Name, manaCost, spendCtx, params.PhyrexianLife, nil)
		if serr != nil {
			return serr
		}
		if manaLife > 0 && !g.CanPayLifeLocked(p, manaLife+ab.LifeCost) {
			return fmt.Errorf("%w: %s cannot pay the %d life for this mana ability's cost (CR 119.4)", ErrInvalidParam, p.Name, manaLife+ab.LifeCost)
		}
		if !p.ManaPool.CanPayFor(manaCost, 0, spendCtx) {
			if !params.AutoTap {
				return &InsufficientManaError{Missing: p.ManaPool.MissingFor(manaCost, 0, spendCtx)}
			}
			plan, short, ok := g.autoTapTopUpLocked(playerID, manaCost, 0, spendCtx,
				ManaActivationAutoTapExclusions(cardID, params), 0)
			// #2461: refused here, before the activation begins, when the
			// plan carried out would not fund the cost.
			if !ok || !g.planFundsLocked(p, plan, short, manaCost, 0, spendCtx) {
				return &InsufficientManaError{Missing: p.ManaPool.MissingFor(manaCost, 0, spendCtx)}
			}
			manaPlan, manaShort = plan, short
		}
	}

	// #1397: every card this cost is about to move, asked about
	// BEFORE the activation begins. CR 605.3a leaves a mana ability no
	// window to pause in once it has started — which is why every move
	// below settles — but nothing has started yet: a commander whose
	// owner has not answered CR 903.9 parks the activation on that
	// question, and the answer activates it again. See
	// cost_commander_choice.go.
	moving := append(append(append([]uuid.UUID(nil), sacrifices...), discards...), exiles...)
	if ab.ExileSelf {
		moving = append(moving, cardID)
	}
	// #1600: a commander exiled to Food Chain is offered the command
	// zone here, before anything is paid.
	moving = append(moving, params.ExilePermanentIDs...)
	// #1427: every permanent the cost TAPS — the source's {T} and
	// the tap-another picks.
	tapping := append([]uuid.UUID(nil), params.TapIDs...)
	if ab.TapCost {
		tapping = append(tapping, cardID)
	}
	// #1474: the source whatever the cost, and every permanent the
	// counter components take counters off — the same list the CR 602
	// path builds.
	spending := append([]uuid.UUID{cardID}, counters.cardIDs()...)
	// #1445 / #1427 / #1474: a card an EFFECT has already paused on
	// its way out cannot pay, or activate. See
	// refusePausedCostCardsLocked.
	if err := g.refusePausedCostCardsLocked(moving, tapping, spending); err != nil {
		return err
	}
	// ADR 0115: nothing a mana ability's cost moves goes to a hand or a
	// library, so no card is asked CR 903.9b before it begins. A
	// sacrificed, discarded or exiled commander is paid like any other
	// card and offered the command zone afterwards by the CR 903.9a
	// state-based action.

	// needStateChecks is set by any component of this activation
	// that can kill a player or a permanent — a sacrifice, a life
	// payment, a damage rider. A single deferred pass covers all of
	// them, so a painland activated at 1 life loses the game on the
	// way out of this call rather than at some later boundary.
	needStateChecks := false

	// #789: the one paid-cost record, built as the components are
	// charged. A mana ability has no stack item to hang it on (CR
	// 605.3b — it never uses the stack), so it lives for the length
	// of this call and is handed to ProducedForPaid, which is how
	// "Add {C} for each storage counter removed this way" knows how
	// many came off.
	paid := PaidCost{}

	// #1212: what this permanent IS, read before a single component
	// of the cost runs. It has to be here and not at the mint below,
	// because the commonest source a card asks about pays for its own
	// ability by SACRIFICING itself — a Treasure is in a graveyard
	// (and a Treasure token has ceased to exist, CR 111.7) three
	// statements before its mana reaches the pool, and `card` is
	// deliberately nil by then.
	//
	// #1228: and the same sentence one zone over — a Spirit Guide is
	// in EXILE by the time its {R} is minted, for exactly the same
	// reason. Off the battlefield the read is of printed
	// characteristics, which is the honest answer: nothing outside
	// the battlefield has layers (CR 613 runs on permanents), so a
	// Simian Spirit Guide in hand is the creature card it prints.
	srcKinds := manaSourceKindsOf(*card)

	// #1183: the activation record, in both scopes, written HERE —
	// the moment every gate and every cost has been checked and the
	// first payment is about to be made. CR 605.3b makes activating a
	// mana ability one indivisible step with no stack and no priority
	// window inside it, so there is no later "announcement finished"
	// to hang this on, and an activation that begins paying has
	// happened. That is the mana-ability spelling of #1181's "an
	// exhaust ability countered on the stack is still spent".
	//
	// Unconditional, exactly as ActivateCatalogAbility's write is:
	// `Ever` is what exhaust reads and `Turn` is the per-turn count
	// the "Activate only once each turn" cards will read, and both are
	// one write at one call site.
	// #2215: the plan made above, carried out now that nothing left can
	// refuse the activation. Other mana abilities, activated to pay this
	// one's mana component (CR 605.3a); whatever they trigger or
	// sacrifice is drained by the state checks on the way out.
	if len(manaPlan) > 0 {
		g.materializePlanLocked(p, manaPlan, manaShort)
		needStateChecks = true
	}

	g.noteAbilityActivationLocked(activationKey)

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
		spent, ok := p.ManaPool.SpendManaFor(manaCost, 0, ManaSpendForAbility(*card))
		if !ok {
			return &InsufficientManaError{Missing: []string{manaCost.String()}}
		}
		paid.Mana = spent
		g.EmitEvent(manaSpentEvent(playerID, cardID, spent))
	}
	// ADR 0131 §2: the announced Phyrexian life, paid with the mana
	// component it replaces (CR 602.2b pays every cost together). The
	// total was checked above; PayLifeForEffect can still refuse a
	// CR 119.8 lock that appeared in between, so the error is returned
	// rather than swallowed.
	if manaLife > 0 {
		if err := g.payPhyrexianLifeLocked(cardID, playerID, manaLife); err != nil {
			return err
		}
		paid.LifePaid += manaLife
		needStateChecks = true
	}
	// #763 / ADR 0074: the permanent as it was at the moment it was
	// tapped for mana (CR 106.12a), copied because the same cost may
	// still sacrifice it (Lotus Petal) and a triggered mana ability's
	// predicate has to be able to ask what it was. Zero value when the
	// ability has no {T}, which is also when nothing fires.
	var tappedForMana Card
	if ab.TapCost {
		card.Tapped = true
		tappedForMana = *card
		g.EmitEvent(Event{Kind: EventTapCard, Actor: playerID, CardID: cardID})
	}
	// ADR 0130 §4: "Exert this land" (Arena of Glory), beside the {T}
	// as the CR 602 path pays it, keyed to the activator's untap step
	// (CR 701.43a). Always payable on the battlefield (CR 701.43b).
	if ab.ExertCost {
		g.exertLocked(cardID, playerID, uuid.Nil)
	}
	// #758: the tap-another half, after the source's own {T} and
	// before the life and the sacrifices — the same component order
	// the CR 602 path pays in, and before the sacrifices for the
	// same reason: a sacrifice moves cards and would take a named
	// permanent off the battlefield before it could be tapped.
	//
	// There is no "after the ability is on the stack" here to
	// respect: a mana ability never uses the stack (CR 605.3b), so
	// whatever the taps trigger is drained by the state-check pass
	// on the way out with everything else this activation queued.
	g.payTapOthersCostLocked(playerID, params.TapIDs)
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
		paid.LifePaid += ab.LifeCost
		needStateChecks = true
	}
	// ADR 0129 §5: the energy, after the life, through the one path
	// that pays energy (CR 107.14). Validated above.
	if err := g.payEnergyLocked(playerID, ab.EnergyCost, cardID); err != nil {
		return err
	}
	// #789: counters after the life and before the sacrifice, the
	// same component order ActivateCatalogAbility pays in — a self
	// removal has to find the source still on the battlefield, and a
	// sacrifice takes it off.
	if err := g.payCounterRemovalLocked(counters); err != nil {
		return err
	}
	paid.CountersRemoved = counters.total
	if ab.AddCounter != nil {
		added, err := g.payCounterAddLocked(playerID, cardID, ab.AddCounter)
		if err != nil {
			return err
		}
		paid.CountersAdded = added
	}
	// S21 sub-PR 1 made sacrifice-self costs real (Treasure, Eldrazi
	// Spawn, Lotus Petal); the mana-cost pass adds sacrifice-another
	// (Ashnod's Altar, Phyrexian Altar). The mana still lands in the
	// pool below — CR 605.3b: a mana ability resolves immediately,
	// without the stack, so the sacrifices and the mana are one
	// atomic step.
	// ADR 0113 §1 (#2072): which objects the sacrifice takes, named
	// before they move.
	paid.SacrificedObjects = g.sacrificeRefsLocked(sacrifices)
	if len(sacrifices) > 0 {
		// One payment, one simultaneous exit (#747): a Blood Artist
		// paid in alongside another creature sees both deaths.
		if err := g.payCostSacrificesLocked(sacrifices, params.commanderAnswers); err != nil {
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
	paid.Sacrificed = len(sacrifices)
	// #1600: the exile-a-permanent component, beside the sacrifices and
	// for the same reason: it moves permanents, so it goes after the
	// tap and the counters, and `card` may no longer point at the
	// source afterwards (a removal shifts the battlefield slice).
	// Recorded on the paid-cost record, which is how Food Chain's
	// ProducedForPaid finds "the exiled creature". The leaves-triggers
	// it queues wait for the state-check pass on the way out, so they
	// reach the stack with the mana already in the pool (CR 605.3a).
	if len(params.ExilePermanentIDs) > 0 {
		exiledPermanents, err := g.payExilePermanentsCostLocked(playerID, cardID, params.ExilePermanentIDs, params.commanderAnswers)
		if err != nil {
			return err
		}
		paid.Exiled = append(paid.Exiled, exiledPermanents...)
		card = nil
		needStateChecks = true
	}
	// #1213: the discard component, LAST — it moves cards out of the
	// hand, and it goes through the ONE discard helper with cause
	// cost, so EventDiscardCard still fires per card, the CR 614
	// window still runs over the exit and madness still sees it. A
	// mana ability resolves immediately (CR 605.3b), so the discard
	// may not pause any more than a spell's cost discard may:
	// MustSettleNow is the route's, not this call site's.
	if len(discards) > 0 {
		if err := g.discardCardsLocked(playerID, discards, discardOptions{
			cause:            DiscardCauseCost,
			source:           cardID,
			commanderAnswers: params.commanderAnswers,
		}); err != nil {
			return err
		}
		// A discard can queue "whenever you discard a card" triggers
		// (Marauding Mako) and a madness exile, for the same reason a
		// sacrifice queues dies-triggers: drain them on the way out,
		// with the mana already in the pool.
		needStateChecks = true
	}
	// #1283: the exile-a-card component, after the discards and before
	// the exile-this — it moves cards out of the hand, not the source,
	// so it keeps the discard's slot in the order. Not a discard: the
	// one exit primitive, no EventDiscardCard, nothing for madness to
	// see (game.ExileCost). The exit can still queue leave-the-hand
	// triggers, so drain on the way out as the discard does.
	if len(exiles) > 0 {
		if err := g.payExileCardsCostLocked(playerID, cardID, exiles, params.commanderAnswers); err != nil {
			return err
		}
		needStateChecks = true
	}
	// #1228: the exile-this component, LAST — after the discards, for
	// the reason they are last: it moves the SOURCE and invalidates
	// every pointer this block was holding. Through the one exit
	// primitive with MustSettleNow, so a commander exiled to a Spirit
	// Guide's cost still gets its CR 903.9 answer and the CR 601.2h
	// indivisible step cannot pause on a prompt.
	//
	// Everything below works off `cardID`, `ab` and the `srcKinds`
	// snapshot taken before the first payment; `card` is deliberately
	// dangling from here, exactly as it is after a sacrifice.
	if ab.ExileSelf {
		if err := g.payExileSelfCostLocked(playerID, cardID, true, params.commanderAnswers); err != nil {
			return err
		}
		card = nil
		// The exit opens a CR 614 window and can queue leave-the-zone
		// triggers, so drain them on the way out with the mana
		// already in the pool — the posture the sacrifice and the
		// discard above each take.
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
		// #1210: the same CardID stamp EventActivateAbility carries —
		// see the auto-tapper's emit above.
		CardID: cardID,
		// #1184: the ability's identity and its exhaust bit, the same
		// two stamps ActivateCatalogAbility's EventActivateAbility
		// carries. A mana ability is an activated ability (CR 605.1a),
		// so "whenever you activate an exhaust ability" watches this
		// kind too — it just never used the stack to get here.
		Label:   ab.Label,
		Exhaust: ab.Exhaust,
	})
	// --- pre-rider ------------------------------------------------
	//
	// #1370: PreRider is everything the oracle text says BEFORE the
	// "Add …" clause — Empowered Autogenerator's "Put a charge counter
	// on this artifact." It runs here, after the cost is paid and
	// before the produced string is computed, so a ProducedFunc that
	// reads the board (Empowered Autogenerator's charge-counter count)
	// sees the placement already landed, counter doublers and all.
	// This is Rider's mirror image: Rider is the tail of the printed
	// sentence, PreRider is its head, and a card declares whichever
	// one its own printed order actually needs — never both.
	//
	// Runs under g.mu held for write, exactly like Rider, and any
	// mutation it makes (a counter placement) goes through a
	// mustSettleNow entry point: CR 605.3b leaves no priority window
	// inside a mana ability's resolution for a CR 616 ordering prompt
	// to occupy, so a Doubling Season beside a Hardened Scales settles
	// on the gathered order instead of asking (see
	// AddCounterMustSettleNowForEffect in counter_tail.go).
	if ab.PreRider != nil {
		needStateChecks = true
		if err := ab.PreRider(g, playerID, cardID); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				Actor:    playerID,
				Source:   cardID,
				ErrorMsg: err.Error(),
			})
		}
	}
	// Materialise the produced mana. An ability with a ProducedFunc
	// computes its output now, from the board as it stands AFTER the
	// cost was paid AND after any PreRider has landed — which is what
	// CR 605.3b's "resolves immediately" means, and what makes Cabal
	// Coffers count the Swamps that are still there and Empowered
	// Autogenerator count the charge counter it just placed.
	produced := ab.Produced
	if ab.DerivedMatch != nil {
		// #1323: same seeded visited set as manaAbilityProducedLocked
		// — `card` is already the exact Card this ability belongs to,
		// so no lookup is needed.
		produced = pipeString(g.derivedManaLocked(*card, &ab, map[uuid.UUID]bool{cardID: true}))
	} else if ab.ProducedFunc != nil {
		produced = ab.ProducedFunc(g, playerID, cardID)
	}
	// #789: an output that depends on what the cost PAID wins over
	// both — "Add {C} for each storage counter removed this way" can
	// only be answered from the record, because the counters it
	// counts came off a moment ago and the board no longer shows
	// how many there were.
	if ab.ProducedForPaid != nil {
		produced = ab.ProducedForPaid(g, playerID, cardID, paid)
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
	// The commander's colour identity, read once for every slot — and
	// only when a slot could use it, because the read walks every zone
	// and a Forest's "{G}" has nothing to narrow or order.
	var identity commanderIdentity
	if ab.NarrowToCommanderIdentity || hasMultiOptionSlot(slots) {
		identity = commanderIdentityFor(g, p)
	}
	// #763: the mana types this activation put in the pool DIRECTLY,
	// for the triggered mana abilities fired at the bottom of this
	// function. A slot that queued a pick contributes nothing here —
	// its colour is not known yet, and ResolveManaChoice fires for it.
	var addedColors []string
	// #1443: the colours named up front, consumed one per picking slot
	// in output order — the order ManaAbilityColorOptions lists them.
	upfront := params.Colors
	for _, slot := range slots {
		// The printed option set, commander identity first (Birds of
		// Paradise offers all five colours), or narrowed to the
		// identity when the printed text says so (Command Tower,
		// Arcane Signet: NarrowToCommanderIdentity).
		options := manaPickOptions(slot.Options, identity, ab.NarrowToCommanderIdentity)
		if len(options) == 0 {
			// CR 903.4f (#844): "in your commander's color identity"
			// with no identity — no commander, or a colourless one —
			// adds no mana. No token, and no prompt: an empty picker
			// is not a choice anybody can answer.
			continue
		}
		// #2558: "N mana of different colors". Named up front, the N
		// colours (validated distinct before the cost was paid) are
		// produced together; otherwise one pick asks for them one at a
		// time and adds nothing until the last (ResolveManaChoice).
		if slot.DifferentColors() {
			n := slot.DistinctCount(options)
			if len(upfront) >= n {
				named := upfront[:n]
				upfront = upfront[n:]
				if differentColorsAllowed(options, n, named) {
					addedColors = append(addedColors, g.produceManaLocked(
						p, cardID,
						append([]string(nil), named...),
						restrictionsFor(g, &ab, playerID, cardID),
						ab.SpendRiders,
						srcKinds,
						ab.TapCost,
						nil,
					)...)
					continue
				}
			}
			g.queueDifferentColorsPickLocked(PendingChoice{
				Chooser:          playerID,
				FromPlayer:       playerID,
				Source:           cardID,
				ManaRestrictions: restrictionsFor(g, &ab, playerID, cardID),
				ManaSourceKinds:  srcKinds,
				ManaRiders:       copyManaRiders(ab.SpendRiders),
				ManaTapped:       ab.TapCost,
			}, options, n, ab.Label)
			continue
		}
		// The PRINTED width decides whether this is a pick, not the
		// narrowed one: Command Tower under a mono-green commander
		// still prompts, with one button, because the player is
		// choosing a colour the card told them to choose.
		if len(slot.Options) == 1 {
			// Single-color slot — through the one production body,
			// which opens the CR 106.12b window on the amount (#1222)
			// and then mints one token per mana it settles on,
			// carrying the ability's spend restrictions (Eldrazi
			// Temple's "colorless Eldrazi only") and #1212's snapshot
			// of what the source was. A Mana Reflection board makes
			// this Forest two {G}, both of them still Forest mana.
			addedColors = append(addedColors, g.produceManaLocked(
				p, cardID,
				[]string{options[0]},
				restrictionsFor(g, &ab, playerID, cardID),
				ab.SpendRiders,
				srcKinds,
				ab.TapCost,
				nil,
			)...)
			continue
		}
		// #1443: the activator named this slot's colour before the
		// cost was paid. Produce it here, through the same production
		// body the answered pick uses (the CR 106.12b window, the
		// restrictions, the source snapshot, the slot's amount), and
		// queue nothing. Re-checked against the list read NOW, after
		// the cost: a derived output the payment itself changed falls
		// back to the prompt rather than minting a colour it no longer
		// offers.
		if len(upfront) > 0 {
			color := upfront[0]
			upfront = upfront[1:]
			if containsColor(options, color) {
				addedColors = append(addedColors, g.produceManaLocked(
					p, cardID,
					repeatColor(color, slot.AmountFor(color)),
					restrictionsFor(g, &ab, playerID, cardID),
					ab.SpendRiders,
					srcKinds,
					ab.TapCost,
					nil,
				)...)
				continue
			}
		}
		// Multi-option slot — see manaPickOptions above.
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
			ColorOptions:     options,
			ManaRestrictions: restrictionsFor(g, &ab, playerID, cardID),
			// #1212: and the source snapshot rides the choice for the
			// same reason the restrictions do — the token is minted
			// later, in ResolveManaChoice, and a Treasure's pick is
			// answered after the Treasure has gone. Without this the
			// one source every "mana from a Treasure" card is printed
			// about would be the one source that records nothing.
			ManaSourceKinds: srcKinds,
			// #1547: and the spend riders, for the reason the
			// restrictions ride here — a Cavern of Souls pick answered
			// without them would mint mana whose spell CAN be countered.
			ManaRiders: copyManaRiders(ab.SpendRiders),
			// #742: "N mana of any one color" — one pick, N tokens.
			ManaAmounts: copyManaAmounts(slot.Amounts),
			// #763: this pick is part of TAPPING A PERMANENT FOR MANA
			// (CR 106.12a), so answering it is what fires the triggered
			// mana abilities — this is the only moment the colour a
			// Birds-style source produced is known. A pick queued by a
			// resolving spell leaves this false and fires nothing.
			ManaTapped: ab.TapCost,
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
	// --- triggered mana abilities (#763, CR 605.1b) ---------------
	//
	// Wild Growth's extra {G}. It resolves HERE — immediately, with no
	// stack and no priority window (CR 605.4a), after the mana ability
	// that triggered it has finished resolving, rider and all.
	//
	// Only for an activation that put mana in a pool DIRECTLY. One
	// that queued a pick instead fires from ResolveManaChoice, where
	// the colour is finally known; no printed mana ability does both,
	// so each card takes exactly one branch and fires exactly once
	// (ADR 0074 §3). An activation that added nothing fires nothing: a
	// permanent tapped for no mana was not tapped for mana.
	if ab.TapCost && len(addedColors) > 0 {
		g.fireManaTriggersLocked(ManaProduced{
			Source:     tappedForMana,
			Controller: playerID,
			Colors:     addedColors,
		}, nil)
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
//
// And since ADR 0093 a THIRD half, always last: the mana abilities a
// layer-6 effect granted the object (Cryptolith Rite, Chromatic
// Lantern). The body — and each row's stable ref — is manaAbilityRows
// in granted_abilities.go.
func ManaAbilitiesForCard(c Card) []ManaAbilityShape { return manaAbilitiesOf(&c) }

// manaAbilitiesOf is ManaAbilitiesForCard without the copy (#1498).
func manaAbilitiesOf(c *Card) []ManaAbilityShape {
	abs, _ := manaAbilityRows(c, false)
	return abs
}

// landTypeMana lists the five basic land types (CR 305.6) with the
// mana their intrinsic ability produces. Order is only a tie-break
// for cards whose effective subtypes are unordered; the real
// ordering comes from the subtype list itself.
//
// Produced and Label are the ability's two strings, spelled out rather
// than concatenated from Color per call: intrinsicLandManaAbilities
// runs for every land in every view, and the concatenation was most of
// its cost (#1498). TestLandTypeManaStringsMatchTheColor pins them.
var landTypeMana = [...]struct{ Subtype, Color, Produced, Label string }{
	{"Plains", "W", "{W}", "Add {W}"},
	{"Island", "U", "{U}", "Add {U}"},
	{"Swamp", "B", "{B}", "Add {B}"},
	{"Mountain", "R", "{R}", "Add {R}"},
	{"Forest", "G", "{G}", "Add {G}"},
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
func intrinsicLandManaAbilities(c *Card) []ManaAbilityShape {
	if !hasCardType(c, "land") {
		return nil
	}
	var subtypes []string
	switch {
	case c.effective != nil:
		subtypes = c.effective.Subtypes
	case c.faceDownPermanent():
		// CR 708.2: the body's subtypes, not the card's — a
		// manifested Forest has none, Yedora's face-down Forest has
		// "Forest" (#1270). HasCardType's cold path makes the same
		// guard for the IsLand above.
		subtypes = faceDownCharacteristic(*c).Subtypes
	default:
		_, _, subtypes = printedTypeParts(c)
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
				Produced: lt.Produced,
				Label:    lt.Label,
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

// manaPickOptionsFor is the colour list a multi-option mana slot
// offers `p`: the one list ActivateManaAbility, AddManaForEffect and
// the auto-tapper (planner and executor alike) all read, so the three
// can never disagree about what a source produces.
//
// Owner decision (2026-09-17): "any color" offers ALL FIVE colours,
// with the commander's colour identity listed first. So by default
// the slot keeps its printed width and is only REORDERED — identity
// colours first, the rest after, each group in printed order. Birds
// of Paradise in a mono-green deck offers G, W, U, B, R; a Scrubland
// in a mono-white deck offers W, B. The order is what makes the
// client's first button, the legal enumerator's first answer (the
// bot's tie-break) and the auto-tapper's default pick
// (pickColorForSlot's options[0]) an identity colour, without taking
// the others away.
//
// `narrow` is for the four cards whose printed text says "any color
// in your commander's color identity" — Command Tower, Arcane Signet,
// Commander's Sphere, Path of Ancestry (NarrowToCommanderIdentity).
// Those INTERSECT with the identity, and the intersection is allowed
// to come back EMPTY. CR 903.4f: a player with no commander has no
// such quality — "that part of the ability won't do anything" — and a
// commander whose identity is colourless (Kozilek, Karn) leaves the
// same nothing to add. An empty list is this seam's "adds no mana":
// every caller skips the slot rather than prompting for a colour that
// does not exist, and nothing OFFERS a source that would add nothing
// (ManaAbilityAddsNoMana). The one exception is a commander whose
// colour data is MISSING rather than colourless — a placeholder card
// with no Scryfall record behind it — where the printed set stands,
// because a data gap must not switch a real card off. See
// commanderIdentityFor for that tri-state.
//
// The identity is the player's commander's wherever it is
// (commanderIdentityFor, CR 903.4a). Always returns a fresh slice
// when it changes anything, so a caller may keep it on a
// PendingChoice.
func manaPickOptionsFor(g *Game, options []string, p *Player, narrow bool) []string {
	return manaPickOptions(options, commanderIdentityFor(g, p), narrow)
}

// manaPickOptions is manaPickOptionsFor with the identity already in
// hand — the auto-tapper computes it once per plan. This is the one
// narrowing function: order by the identity, or intersect with it.
// Nothing else in the engine reads the identity to decide what a mana
// source offers.
func manaPickOptions(options []string, identity commanderIdentity, narrow bool) []string {
	if narrow {
		// A gap in the card data is not a rules state (#844): with
		// nothing trustworthy to narrow against, the printed card
		// stands. commanderIdentityFor has already logged it.
		if identity.State == identityUnknown {
			return options
		}
		// CR 903.4f. No commander, or a colourless identity, means
		// there are no colours to intersect with, and a printed colour
		// outside the identity falls away for the same reason: the
		// ability can only add a colour the identity names. Nil, not
		// an empty slice, so "adds nothing" has one spelling.
		narrowed := intersectColors(options, identity.Colors)
		if len(narrowed) == 0 {
			return nil
		}
		return narrowed
	}
	if len(options) <= 1 || len(identity.Colors) == 0 {
		return options
	}
	return identityFirst(options, identity.Colors)
}

// identityFirst returns `options` reordered so the colours in
// `identity` come first. Stable within each group, so WUBRG printed
// order survives inside both halves.
func identityFirst(options, identity []string) []string {
	in := intersectColors(options, identity)
	if len(in) == 0 || len(in) == len(options) {
		return options
	}
	out := make([]string, 0, len(options))
	out = append(out, in...)
	for _, o := range options {
		if !containsColor(in, o) {
			out = append(out, o)
		}
	}
	return out
}

func containsColor(set []string, c string) bool {
	for _, x := range set {
		if x == c {
			return true
		}
	}
	return false
}

// manaAbilityProducedLocked is the produced-mana string one mana
// ability would add, given what its cost paid. THE one place the
// three-slot precedence lives — ProducedForPaid beats ProducedFunc
// (or DerivedMatch) beats Produced — so the activation, the
// auto-tapper's planner and executor, CR 106.7's "could produce"
// reader and CR 903.4f's adds-no-mana check cannot disagree about
// what a source makes.
//
// Read-only. Caller must hold g.mu.
func manaAbilityProducedLocked(g *Game, playerID, cardID uuid.UUID, ab *ManaAbilityShape, paid PaidCost) string {
	if ab == nil {
		return ""
	}
	out := ab.Produced
	if ab.DerivedMatch != nil {
		// #1323: seed the visited set with the source's OWN instance
		// ID before recursing, exactly as producibleManaVisitingLocked
		// does — otherwise a cycle that loops all the way back to
		// THIS activation (not just to an intermediate card) would
		// not be caught by a fresh, empty set.
		if c, ok := g.LookupCardForEffect(cardID); ok {
			out = pipeString(g.derivedManaLocked(c, ab, map[uuid.UUID]bool{cardID: true}))
		}
	} else if ab.ProducedFunc != nil {
		out = ab.ProducedFunc(g, playerID, cardID)
	}
	if ab.ProducedForPaid != nil {
		out = ab.ProducedForPaid(g, playerID, cardID, paid)
	}
	return out
}

// maxCounterPaymentLocked is the LARGEST counter removal `rc` could
// be paid with right now, as a PaidCost — what a reader that asks
// "what could this source produce" has to assume, because "could
// produce" (CR 106.7) is about what is possible, not about what some
// hypothetical activation chose.
//
// A Mage-Ring Network with three storage counters could produce
// {C}{C}{C}; one with none could produce nothing, and answering
// otherwise would let a cast be planned against mana that cannot
// exist. For a fixed cost the answer is simply N.
//
// Caller must hold g.mu (read or write).
func (g *Game) maxCounterPaymentLocked(playerID, sourceID uuid.UUID, rc *CounterRemovalCost) PaidCost {
	if rc == nil {
		return PaidCost{}
	}
	if !rc.Variable {
		return PaidCost{CountersRemoved: rc.N}
	}
	total := 0
	for _, o := range g.CounterCostOptionsForEffect(playerID, sourceID, rc) {
		best := 0
		for _, k := range o.Kinds {
			if k.Count > best {
				best = k.Count
			}
		}
		total += best
		if !rc.Among {
			// A variable cost that is not an among cost is paid off
			// one permanent — the source — so the first option is
			// the whole answer.
			break
		}
	}
	return PaidCost{CountersRemoved: total}
}

// MaxCounterPaymentForEffect is the largest number of counters a
// variable removal could be paid with right now — the ceiling the
// client's picker bounds its stepper with, from the same walk the
// engine validates against, so a player can never name a number the
// server then refuses. Zero for a fixed cost's nil component.
//
// Caller must hold g.mu (read or write).
func (g *Game) MaxCounterPaymentForEffect(playerID, sourceID uuid.UUID, rc *CounterRemovalCost) int {
	return g.maxCounterPaymentLocked(playerID, sourceID, rc).CountersRemoved
}

// ManaAbilityAddsNoMana reports that one mana ability would add no
// mana at all right now, so nothing should offer it. Two cases:
//
//   - CR 903.4f: an ability whose printed text says "any color in your
//     commander's color identity" (NarrowToCommanderIdentity) activated
//     by a player who has no commander, or whose commander's colour
//     identity is colourless — the identity is undefined or empty, and
//     "that part of the ability won't do anything".
//   - ADR 0117 §5: an ability whose output is COMPUTED (ProducedFunc,
//     ProducedForPaid or DerivedMatch) and computes to no mana right
//     now — a power-0 Vivi Ornitier, a Selvala whose greatest power is
//     0, an Exotic Orchard with nothing to copy, a Mage-Ring Network
//     with no counters. It is read the way every "what would this
//     make" reader reads it: manaAbilityProducedLocked with the largest
//     counter payment. A static Produced of "" stays out of it.
//
// The engine still ACCEPTS such an activation (CR 605.1a: the ability
// exists and may be activated; it just does nothing, which is what the
// rulings on Command Tower and its kin say). This is what keeps it from
// being OFFERED: the legal-move enumerator drops the move, the view
// greys the row, and the auto-tapper skips the source (it never plans
// a source with no output).
//
// Read-only. Callers hold whatever lock their read path already holds
// — the legal enumerator reads the battlefield the same way.
func ManaAbilityAddsNoMana(g *Game, playerID, cardID uuid.UUID, ab ManaAbilityShape) bool {
	if g == nil {
		return false
	}
	computed := ab.ProducedFunc != nil || ab.ProducedForPaid != nil || ab.DerivedMatch != nil
	if !ab.NarrowToCommanderIdentity && !computed {
		return false
	}
	produced := manaAbilityProducedLocked(g, playerID, cardID, &ab, g.maxCounterPaymentLocked(playerID, cardID, ab.RemoveCounters))
	slots, err := ParseProducedMana(produced)
	if err != nil {
		// A broken declaration is not this rule's business.
		return false
	}
	if len(slots) == 0 {
		// Empty output: only a computed one is "adds nothing right
		// now"; a static empty declaration is left alone.
		return computed
	}
	if !ab.NarrowToCommanderIdentity {
		return false
	}
	identity := commanderIdentityFor(g, g.playerByIDLocked(playerID))
	for _, slot := range slots {
		if len(manaPickOptions(slot.Options, identity, true)) > 0 {
			return false
		}
	}
	return true
}

// identityState is what "your commander's color identity" IS for a
// player right now, as CR 903.4f divides it. Three states, because
// before #844 an empty colour list answered two different questions
// and the engine could not tell them apart: a real colourless
// commander, and a commander nobody has colour data for.
type identityState uint8

const (
	// identityNoCommander — the player has no commander at all.
	// CR 903.4f: the quality is undefined, and an ability that
	// refers to it does nothing.
	identityNoCommander identityState = iota
	// identityKnown — the commander's colour data is authoritative.
	// Colors MAY be empty, and that is a real answer rather than a
	// gap: Kozilek, Butcher of Truth is a legal commander whose
	// colour identity is colourless, so "one mana of any color in
	// your commander's color identity" is no colour at all.
	identityKnown
	// identityUnknown — the player has a commander, but nothing on
	// it says what its colour identity is: a placeholder card with
	// no Scryfall record behind it (the demo seed, a fixture). Not a
	// rules state but a data gap, and the narrowing falls back to
	// the printed colours so a missing import never turns Command
	// Tower off in a real game.
	identityUnknown
)

// commanderIdentity is the tri-state result of the one identity
// computation. Callers read State, never `len(Colors) == 0`, which is
// the whole point of the type: the empty slice is ambiguous and this
// is not.
type commanderIdentity struct {
	State  identityState
	Colors []string
}

// missingIdentityOnce keeps the data-gap warning to one line per
// process. The condition is a broken import rather than a game event
// — it cannot change during a game, and one line is enough to find
// it.
var missingIdentityOnce sync.Once

// commanderIdentityFor returns the colour identity of the player's
// commander as uppercase single-character strings, WITH the tri-state
// that says what an empty list means (CR 903.4f, #844).
//
// The commander is found in EVERY zone, not just the command zone.
// CR 903.4a fixes colour identity before the game begins, so it does
// not change when the commander is cast, dies or is exiled, and a
// commander on the battlefield is the ordinary mid-game state. The
// search matches on IsCommander and Owner (the flag survives zone
// changes, and commander damage already relies on that); ownership,
// not control, because a stolen commander is still its owner's.
// Zones are read in the order a commander usually sits: command zone,
// stack, battlefield, graveyard, exile, hand, library.
//
// Partners and backgrounds: every commander the player owns
// contributes, and the result is their union in discovery order
// (CR 903.4: the deck's identity is the combined identity). A pair
// where one half's colours are known and the other's are missing is
// KNOWN at the colours actually found — the weaker direction, and the
// one that keeps a real deck's Command Tower working. Only when no
// colour was found anywhere AND some commander's data is missing does
// the answer come back unknown.
//
// The per-card read, and the "is this authoritative" question, are
// printedIdentityOf's.
//
// The result feeds manaPickOptions, which ORDERS an ordinary pipe by
// it (a wrong identity only changes which colour is listed first) and
// NARROWS the four "in your commander's color identity" cards, where
// it can now also mean "this ability adds nothing".
func commanderIdentityFor(g *Game, p *Player) commanderIdentity {
	if p == nil {
		return commanderIdentity{State: identityNoCommander}
	}
	var out commanderIdentity
	var found, missing bool
	add := func(c *Card) {
		if !c.IsCommander || c.Owner != p.ID {
			return
		}
		found = true
		colors, known := printedIdentityOf(c)
		if !known {
			missing = true
		}
		for _, col := range colors {
			if !containsColor(out.Colors, col) {
				out.Colors = append(out.Colors, col)
			}
		}
	}
	zones := []*Zone{p.Command}
	if g != nil {
		zones = append(zones, g.Stack, g.Battlefield)
	}
	zones = append(zones, p.Graveyard)
	if g != nil {
		zones = append(zones, g.Exile)
	}
	zones = append(zones, p.Hand, p.Library)
	for _, z := range zones {
		if z == nil {
			continue
		}
		for i := range z.Cards {
			add(&z.Cards[i])
		}
	}
	switch {
	case !found:
		out.State = identityNoCommander
	case missing && len(out.Colors) == 0:
		out.State = identityUnknown
		missingIdentityOnce.Do(func() {
			slog.Warn("commander has no colour identity data: identity mana falls back to the printed colours",
				"player", p.ID, "rule", "CR 903.4f", "issue", 844)
		})
	default:
		out.State = identityKnown
	}
	return out
}

// printedIdentityOf reads one commander's colour identity and says
// whether that answer is authoritative — the two halves #844 asks to
// be separated, decided in one place so no caller has to infer them
// from an empty slice.
//
// Four sources per commander, in order:
//
//  1. **Card.ColorIdentity** — Scryfall's own `color_identity`,
//     copied at deck import. This is the only one of the four that is
//     actually CR 903.4: it folds in mana symbols in rules text and
//     BOTH faces of a double-faced card. Issue #276: a transform /
//     modal-DFC commander has a null top-level mana_cost and colors
//     (the real ones live on card_faces[0]), so sources 3 and 4 both
//     returned empty and Command Tower offered all five colours to an
//     Azorius deck.
//  2. **An imported card's EMPTY ColorIdentity** — authoritative too,
//     and the #844 fix: a card with a Scryfall printing behind it
//     (ScryfallID / OracleID, stamped by deck.ToGameCard, the one
//     path a real deck takes) that lists no colours really has none.
//     Kozilek, Butcher of Truth is colourless, so "any color in your
//     commander's color identity" adds nothing for its deck.
//  3. **Effective().Colors** (S16) — the commander's post-layer
//     colours. Not identity, but a good proxy for any commander whose
//     identity is fully captured by their mana cost.
//  4. **distinctColorsInManaCost** — the original S15 proxy. Covers
//     placeholder commanders with no stamped colours (the demo seed)
//     and the lobby-time pre-effects-init path.
//
// A copy keeps its original printed values in PrintedSelf: a copied
// commander's colour identity comes from those, never from the
// creature it copied (CR 707.2 changes copiable values, not the
// identity CR 903.4a fixed before the game began).
//
// The colour fallbacks can only be narrower than the truth. `known`
// is false only when nothing at all is on the card AND no Scryfall
// printing stands behind it, which no imported deck produces.
func printedIdentityOf(c *Card) (colors []string, known bool) {
	cost, imported := c.ManaCost, c.fromScryfallPrinting()
	colors = c.ColorIdentity
	if original := c.PrintedSelf; original != nil {
		colors, cost = original.ColorIdentity, original.ManaCost
		imported = original.ScryfallID != "" || original.OracleID != ""
		if len(colors) == 0 {
			colors = original.Colors
		}
	} else if len(colors) == 0 {
		colors = c.Effective().Colors
	}
	if len(colors) == 0 {
		colors = distinctColorsInManaCost(cost)
	}
	if len(colors) > 0 {
		return colors, true
	}
	return nil, imported
}

// fromScryfallPrinting reports whether this instance was stamped from
// a real Scryfall record (deck.ToGameCard) rather than conjured as a
// placeholder — the demo seed, a token template, a test fixture. Both
// IDs come off the same record, so either one standing is the whole
// answer.
//
// ADR 0078: `ScryfallID != ""` no longer means that on its own for a
// token whose id is only a resolved ART printing (TokenArtOnly) — a
// vanilla Treasure or Soldier does not become "imported" just because
// it got a picture. `&& !c.TokenArtOnly` on that half keeps this
// reading "was this instance stamped from deck import" exactly as
// documented; a token COPY's ScryfallID is a real copied identity
// (TokenArtOnly false there) and still answers true, as does the
// OracleID half for anything that carries one.
func (c Card) fromScryfallPrinting() bool {
	return (c.ScryfallID != "" && !c.TokenArtOnly) || c.OracleID != ""
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

// PassPriority is the priority holder passing (CR 117.3d). The pass
// joins Turn.PassedInSuccession; until every seat still in the game
// has passed in succession, priority moves to the next non-eliminated
// seat in turn order — the active seat included. Once every seat has
// (CR 117.4, #2275), the top of the stack resolves and the active
// player receives priority again (CR 117.3b), or, with an empty stack,
// the step ends and PriorityHolder is reset to the new step's
// ActiveSeat. Arriving back at the active seat is not the test: that
// is the same thing only when the round began there. See
// priority_succession.go for what restarts a succession.
//
// S13: returns ErrNoPriority during the Untap and Cleanup steps,
// which don't grant priority (CR 502.4 / 514.3). Eliminated seats are
// skipped during the rotation and are not waited for, so a 4-player
// game with one dead seat still ends the round on the survivors'
// passes.
//
// #661 / CR 514.3a is cleanup's exception, on both halves: the step
// DOES grant priority when a state-based action fired or a trigger
// was waiting there, and the round of passes that ends that window
// begins another cleanup step instead of the next turn.
func (g *Game) PassPriority() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	// #1665: passing is the decline of a resolved miracle's cast
	// (CR 702.94a) — see miracle.go. Before the pass, because this
	// pass can resolve the next miracle trigger.
	//
	// ADR 0099 §4: passing is also the decline of a discover's or a
	// cascade's free cast. The card goes to hand (discover) or to the
	// bottom (cascade) before the pass runs. When that moved something
	// that asks or triggers — a commander's CR 903.9 offer, a "whenever
	// you discover" payoff — the holder keeps priority with it on the
	// stack or on the table, exactly as if the discover had put the
	// card into their hand during its resolution; they pass again to
	// move on.
	if h := g.Turn.PriorityHolder; h >= 0 && h < len(g.Seats) && g.Seats[h] != nil {
		if g.closePassWindowsLocked(g.Seats[h].ID) {
			if g.blockingChoiceLocked() != nil || len(g.PendingTriggers) > 0 {
				g.runStateChecksLocked()
				return nil
			}
		}
	}
	return g.passPriorityLocked()
}

// passPriorityLocked is PassPriority's whole body, under a lock the
// caller already holds. Extracted in #914 so AdvanceStep can drive
// the table's passes through the ONE priority engine rather than
// growing a second one beside it — see driveStepToEndLocked (game.go).
// Caller must hold g.mu.
func (g *Game) passPriorityLocked() error {
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
	// #1571 / CR 508.1d: the ACTIVE player's pass in declare_attackers
	// is where they say their attack declaration is done, so it is the
	// declaration's requirement checkpoint — refused, with the unmet
	// requirement named, while a free addition would still obey one
	// more ("Zurgo Helmsmasher attacks each combat if able"). Only
	// once per combat: after it is accepted the declaration is judged,
	// and priority coming back round in the step asks nothing again.
	// A table with no requirement on the board never gets past the
	// fast path. See attack_requirements.go.
	if g.Turn.Step == StepDeclareAttackers && g.Turn.PriorityHolder == g.Turn.ActiveSeat {
		if err := g.attackCheckpointLocked(); err != nil {
			return err
		}
	}
	// #1279 / CR 509.1: a defending player who passes in the
	// declare-blockers step has finished declaring blockers — the pass
	// is how a table that blocks by hand has always said "done". When
	// that completes the LAST pending declaration, the turn-based
	// action is over: its triggers go on the stack and the ACTIVE
	// player receives priority (CR 509.2, 117.3a), rather than the
	// rotation carrying on as if the attacker had already had its
	// post-declaration window. #1501: priority is parked while anyone
	// is declaring, so this is the rare case of a pending defender who
	// holds priority anyway (block_completion.go's header).
	if g.Turn.Step == StepDeclareBlockers {
		if h := g.Turn.PriorityHolder; h >= 0 && h < numSeats && g.Seats[h] != nil {
			// #1597 / CR 509.1c: and so it is the declaration's
			// requirement checkpoint — refused, with the requirement
			// named, while a legal addition would still obey one more
			// ("Grizzly Bears must block Prized Unicorn if able").
			// Only while this defender's declaration is pending; the
			// fast path keeps every table without a requirement out.
			if err := g.blockCheckpointLocked(g.Seats[h].ID); err != nil {
				return err
			}
			if g.completeBlockDeclarationLocked(g.Seats[h].ID) && g.closeBlockDeclarationIfCompleteLocked() {
				return nil
			}
		}
	}
	// #2275 / CR 117.4: this pass joins the succession. Until every
	// seat still in the game has passed without anyone acting in
	// between, priority goes to the next player in turn order (CR
	// 117.3d) — the ACTIVE seat included. Arriving back at the active
	// seat used to be read as "everyone passed", which is only true
	// when the round began there: a non-active caster keeps priority
	// (CR 117.3c), so their spell resolved on their own pass without
	// the active player ever holding priority over it. See
	// priority_succession.go.
	//
	// The walk is bounded by numSeats iterations so a fully-eliminated
	// table can't infinite-loop (the surrounding game-end check in
	// Concede flips State to StateEnded in that case; the early
	// ErrGameNotActive guard above catches subsequent calls).
	g.notePassLocked(g.Turn.PriorityHolder)
	if !g.allPassedInSuccessionLocked() {
		next := g.Turn.PriorityHolder
		for i := 0; i < numSeats; i++ {
			next = (next + 1) % numSeats
			if s := g.Seats[next]; s != nil && !s.Eliminated {
				g.Turn.PriorityHolder = next
				return nil
			}
		}
	}
	// Every seat still in the game has passed in succession.
	//
	// #830 and #859 first: priority has passed all the way around, so
	// the combat declaration staged in this step is complete
	// (CR 508.1 / CR 509.1). Lock it in BEFORE the advance-or-resolve
	// choice below, so its triggers reach the stack inside the
	// declaring step rather than a step late, after blockers or after
	// combat damage. If it put anything there, priority returns to
	// the active player (CR 508.2 / 509.2a) and the table gets a
	// window to respond before those triggers resolve — the step does
	// not advance on this pass.
	//
	// #1279: a defender still pending as the step would end has
	// declared whatever is staged — the step cannot end on an
	// unfinished turn-based action. Completing announces it, so the
	// boundary below runs for it exactly as for a staged lock-in.
	blocksCompleted := g.completeAllBlockDeclarationsLocked()
	if blocksCompleted || g.blockDeclarationPendingLocked() || g.attackDeclarationPendingLocked() {
		g.runStateChecksLocked()
		// A blocking choice counts as much as a stack item here: an
		// optional declaration trigger (Grazilaxx, Legion Loyalty's
		// myriad) queues its yes/no rather than an item, and walking
		// the cursor past an unanswered prompt is the thing #730
		// forbids.
		if g.stackHasItemsLocked() || g.blockingChoiceLocked() != nil {
			g.grantPriorityLocked(g.Turn.ActiveSeat)
			return nil
		}
	}
	// Then the two ordinary cases:
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
		before := g.Turn
		// #489: a resolution that FAILED is still a resolution that
		// happened. The old `return err` here skipped both of the
		// lines below, so one bad resolution left the game half
		// resolved — the item off the stack, its effect run, and the
		// table still waiting on a priority holder that had already
		// passed, with no state-based actions and no triggers. The
		// error is reported the way every other resolution-time error
		// is (fireEffectResolverLocked, CR 608.2's "the game
		// continues"), and the two post-resolution jobs run either
		// way.
		if err := g.resolveTopOfStackLocked(); err != nil {
			// No Actor: the failure belongs to the resolution, not to
			// the player who happened to pass last — the same shape
			// fireEffectResolverLocked emits.
			g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
		}
		// Drain any pending APNAP triggers onto the stack now that
		// we've crossed a priority-grant boundary (CR 603.3b).
		g.runStateChecksLocked()
		// #2165: unless the resolution moved the cursor itself — it
		// ended the combat phase (CR 724.2) or the turn (CR 724.1) —
		// in which case the step it moved to has already decided who
		// holds priority, and a cleanup step waiting on a discard has
		// decided nobody does. See end_turn.go.
		if g.cursorMovedSince(before) {
			return nil
		}
		// Priority returns to the active player after a resolution
		// (CR 117.3b), and a new succession of passes begins with
		// them (#2275). The step doesn't change.
		g.grantPriorityLocked(g.Turn.ActiveSeat)
		return nil
	}
	// CR 514.3a (#661): the pass that closes a priority window in the
	// CLEANUP step does not end the turn. "Once the stack is empty and
	// all players pass in succession, another cleanup step begins" —
	// hand size is checked again and the CR 514.2 sweep runs again,
	// which is how an effect created during cleanup still ends this
	// turn. See cleanup.go.
	if g.Turn.Step == StepCleanup {
		g.repeatCleanupStepLocked()
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
// without haste (CR 302.6, 702.10), ErrDefender for a defender
// creature (CR 702.3), and an *AttackLimitError (wrapping
// ErrAttackLimit) when one more attacker would break a CR 508.1c
// count limit such as Silent Arbiter's (#1507), and an
// *AttackRequirementError (wrapping ErrAttackRequirement) when the
// declaration would make a CR 508.1d requirement unobeyable (#1571).
// Re-declaring the same attacker against a
// different target overwrites the previous target.
//
// Caller authorization (was-it-the-controller) is intentionally
// NOT enforced — sandbox flexibility for casual play. Summoning
// sickness and defender are enforced as of S18.
//
// S22 / #859: the verb STAGES the attack and announces nothing.
// CR 508.1 declares attackers as ONE turn-based action, so the
// EventAttack behind "whenever ~ attacks" — still exactly one per
// declared attacker (see events.go) — is emitted by
// commitAttackDeclarationLocked when the declaration is locked in, at
// the first priority boundary of the step. That is still inside
// declare_attackers and still ahead of blockers; what changed is that
// a creature re-pointed at a different defender before the lock-in
// announces once, naming the defender it ends on, instead of the one
// it was first declared against. See attackers.go.
func (g *Game) DeclareAttacker(attackerID, targetPlayerID uuid.UUID) error {
	return g.DeclareAttackerWith(attackerID, targetPlayerID, DeclareAttackersParams{})
}

// DeclareAttackerWith is DeclareAttacker with the CR 508.1a payment
// posture named (ADR 0080, #1063) — whether the attack tax may tap
// lands, which lands are locked, and how many Phyrexian symbols are
// being paid with life.
//
// A twin rather than a changed signature because DeclareAttacker has
// 190 call sites and every one of them is at a table with no attack
// tax on it, where the zero value costs nothing and does nothing. The
// zero value is also the STRICT posture — pay from the pool, refuse if
// short — so a caller that forgets the params cannot waive a tax.
func (g *Game) DeclareAttackerWith(attackerID, targetPlayerID uuid.UUID, params DeclareAttackersParams) error {
	return g.DeclareAttackerDeclWith(AttackDeclaration{Attacker: attackerID, Target: targetPlayerID}, params)
}

// DeclareAttackerDeclWith is DeclareAttackerWith taking the whole
// declaration, so it can carry the choice to exert the attacker as it
// attacks (ADR 0130 §2): `exert: true` on declare_attacker. The
// choice is validated before anything is paid or staged
// (ErrCantExert), staged on Card.ExertOnAttack and paid at the
// lock-in. A creature re-pointed before the lock-in keeps a choice it
// already staged.
func (g *Game) DeclareAttackerDeclWith(one AttackDeclaration, params DeclareAttackersParams) error {
	attackerID, targetPlayerID := one.Attacker, one.Target
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
			// ADR 0106 §2 (#1794): the creature's own CR 508.1c target
			// restrictions ("can't attack its owner") ride the same
			// check, refused with a sentence naming them.
			if err := g.canAttackTargetWithLocked(card, targetPlayerID); err != nil {
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
			// CR 508.1a, ADR 0080: the attack tax. Priced against the
			// one attack this verb declares, and paid BEFORE anything
			// is staged, so a refusal leaves the creature untapped,
			// undeclared and the board exactly as it was.
			//
			// This verb is deliberately lax about ELIGIBILITY for the
			// sandbox's hand-forcing, and the tax is not eligibility.
			// A free declaration under Propaganda is the card played
			// as a blank, which is the #259 direction the roadmap
			// refuses — and it is also what the enumerator would then
			// be offering, since it emits this verb (#544).
			decl := []AttackDeclaration{{Attacker: attackerID, Target: targetPlayerID}}
			// CR 508.1c, #1507: a count limit — Silent Arbiter's "no
			// more than one creature can attack each combat",
			// Crawlspace's two "attacking you". Judged against every
			// creature already attacking, and before the tax: a
			// declaration the limit refuses owes nothing. Not relaxed
			// for the sandbox's hand-forcing, for the reason
			// "can't attack" above is not — see attack_limits.go.
			if err := g.attackLimitRefusalLocked(decl); err != nil {
				return err
			}
			// CR 508.1d, #1571: a declaration that makes a requirement
			// that could still be obeyed unobeyable — a goaded creature
			// at its goader while another opponent is open, a creature
			// with no requirement into the only slot a creature that
			// must attack needs. Not relaxed for hand-forcing either:
			// a requirement ignored is the card played as a blank,
			// Zurgo's drawback and goad's whole point. Before the tax,
			// for the limit's reason. See attack_requirements.go.
			if err := g.attackRequirementRefusalLocked(decl); err != nil {
				return err
			}
			// ADR 0130 §2: the choice to exert it as it attacks
			// (CR 508.1g), checked before the tax so a refusal leaves
			// the board as it was. The requirement check above ignores
			// the choice: a requirement never forces an optional cost
			// (CR 508.1d).
			if err := g.validateExertChoicesLocked([]AttackDeclaration{one}); err != nil {
				return err
			}
			price := g.priceAttackDeclarationLocked(decl)
			if err := g.payAttackTaxLocked(card.Controller, price, params); err != nil {
				return err
			}
			// Staged, paid at the lock-in. A re-point keeps a choice
			// already staged; clear_combat or undo is how it is taken
			// back.
			if one.Exert {
				card.ExertOnAttack = true
			}
			// #859: staging, not announcing. A creature is declared
			// as an attacker once (CR 508.1), and the sandbox lets a
			// player re-point an already-attacking creature at a
			// different defender before the declaration is complete;
			// that is one decision being revised, so the single
			// EventAttack the creature is owed has to name the
			// defender it ENDS on. Both halves are the lock-in's job
			// — see commitAttackDeclarationLocked.
			// #1364: setAttackTargetLocked also records the defending
			// player, for CR 506.4c's "it may be blocked" after the
			// planeswalker or battle it attacks has left.
			g.setAttackTargetLocked(card, targetPlayerID)
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
			card.clearBlocking()
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
	// Exert is the player's choice to exert the attacker as it attacks
	// (CR 701.43d, 508.1g; ADR 0130 §2). Staged on Card.ExertOnAttack
	// and paid as the declaration locks in. A creature that can't be
	// exerted as it attacks refuses the whole declaration with
	// ErrCantExert. False is every declaration made before exert.
	Exert bool
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
// trigger goes on the stack together in APNAP order. Here every card
// is staged first, then a single runStateChecksLocked locks the
// declaration in — announcing every EventAttack in one batch and
// draining the triggers together.
//
// #859: that lock-in is the same commitAttackDeclarationLocked the
// per-creature verb goes through, so there is exactly ONE place an
// attack declaration is announced, whichever verb staged it. The
// contract here is unchanged — one event per declared creature, one
// batch, one drain — and the per-creature verb now matches it
// instead of announcing a click at a time. The only difference is
// that the commit walks the battlefield, so the events come out in
// battlefield order rather than submission order; that order is the
// order same-controller triggers are offered in, and nothing else.
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
//
// ADR 0080 adds the one exception to "skip what doesn't fit": the
// CR 508.1a attack tax is ALL OR NOTHING. See DeclareAttackersWith.
// #1507 adds the second, for the same reason: a CR 508.1c count limit
// (Silent Arbiter, Crawlspace) refuses the whole eligible set with an
// *AttackLimitError rather than choosing which creatures to drop.
func (g *Game) DeclareAttackers(decls []AttackDeclaration) ([]uuid.UUID, error) {
	return g.DeclareAttackersWith(decls, DeclareAttackersParams{})
}

// DeclareAttackersWith is DeclareAttackers with the CR 508.1a payment
// posture named (ADR 0080, #1063).
//
// The tax is priced against the ELIGIBLE set — what this verb would
// actually declare after skipping tapped, sick and restricted entries
// — and paid as ONE payment before any creature is staged. If it
// cannot be paid in full, nothing is declared and an
// *AttackTaxUnpaidError comes back: a partially paid declaration is
// not a legal declaration, and which attacks to drop when the seat can
// afford only some of them is the player's choice, made by submitting
// a smaller set. That is the one place this verb is not "skip what
// doesn't fit", and it has to be.
func (g *Game) DeclareAttackersWith(decls []AttackDeclaration, params DeclareAttackersParams) ([]uuid.UUID, error) {
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

	// ADR 0080 split what used to be one loop into two passes. The
	// first decides WHICH entries are declarations; the second stages
	// them. The CR 508.1a payment goes between, because it is priced
	// against the eligible set (charging for an attack that gets
	// skipped would be wrong) and because an unpayable tax must leave
	// nothing staged (a half-declared swing is not a declaration).
	eligible := make([]AttackDeclaration, 0, len(decls))
	// The eligible entries grouped by whose creatures they are, in
	// first-seen order. Almost always ONE group: authorization one
	// layer up refuses a batch mixing seats for every caller but an
	// admin session, which the sandbox lets drive another seat's
	// board. Grouping rather than assuming one payer means an admin's
	// mixed batch charges each seat its own share instead of billing
	// the first one for everybody's attacks.
	payers := make([]uuid.UUID, 0, 1)
	byPayer := make(map[uuid.UUID][]AttackDeclaration, 1)
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
		// for the player-only case. ADR 0106 §2 (#1794): and the
		// creature's own target restrictions, so a Xantcha pointed at
		// its owner is skipped like any other illegal entry.
		if g.canAttackTargetWithLocked(card, d.Target) != nil {
			continue
		}
		eligible = append(eligible, d)
		if _, seen := byPayer[card.Controller]; !seen {
			payers = append(payers, card.Controller)
		}
		byPayer[card.Controller] = append(byPayer[card.Controller], d)
	}
	if len(eligible) == 0 {
		return nil, ErrNoLegalAttackers
	}
	// CR 508.1c, #1507: the count limits, ALL OR NOTHING. Every entry
	// here is individually eligible, so there is no "ineligible" one to
	// skip — which creatures to leave home under Silent Arbiter is the
	// attacking player's choice, made by submitting a smaller set, as
	// ADR 0080 has it for a tax the seat can only partly afford.
	// Judged before the tax so a refused declaration is not priced.
	if err := g.attackLimitRefusalLocked(eligible); err != nil {
		return nil, err
	}
	// CR 508.1d, #1571: ALL OR NOTHING like the limit — a swing that
	// leaves a requirement unobeyable (every creature at the goader
	// while another opponent is open) is refused whole, and which
	// creatures to point elsewhere is the player's choice.
	if err := g.attackRequirementRefusalLocked(eligible); err != nil {
		return nil, err
	}
	// ADR 0130 §2: an exert choice on an eligible attacker that can't
	// be exerted refuses the whole batch, before any tax is paid. An
	// INELIGIBLE attacker was skipped silently above (#318) and its
	// choice with it; an exert flag the creature can't honour is a
	// client bug, and dropping it would attack without the cost the
	// player chose.
	if err := g.validateExertChoicesLocked(eligible); err != nil {
		return nil, err
	}
	// CR 508.1a, before anything is staged and before the CR 508.1f
	// taps: the declaration's attack tax, all or nothing.
	//
	// Priced for every payer BEFORE any of them is charged, so a
	// mixed batch whose second seat cannot pay does not leave the
	// first seat's mana spent on a declaration that never happened.
	// That pre-check is exact rather than optimistic: the seats' mana
	// pools and untapped permanents are disjoint, so nothing one
	// payment does can change what another can afford.
	prices := make([]AttackTaxPrice, len(payers))
	for i, p := range payers {
		prices[i] = g.priceAttackDeclarationLocked(byPayer[p])
		if err := g.attackTaxAffordableLocked(p, prices[i], params); err != nil {
			return nil, err
		}
	}
	for i, p := range payers {
		if err := g.payAttackTaxLocked(p, prices[i], params); err != nil {
			return nil, err
		}
	}

	declared := make([]uuid.UUID, 0, len(eligible))
	for _, d := range eligible {
		card := findBattlefieldCard(g, d.Attacker)
		if card == nil {
			// Unreachable: nothing between the passes moves a
			// permanent. Skipped rather than dereferenced, because a
			// nil here would take the whole declaration down with it.
			continue
		}
		g.setAttackTargetLocked(card, d.Target)
		// ADR 0130 §2: staged, paid by the lock-in below.
		if d.Exert {
			card.ExertOnAttack = true
		}
		// CR 508.1f: declaring an attacker taps it unless it has
		// vigilance (CR 702.20).
		if !HasKeyword(card, "vigilance") {
			card.Tapped = true
		}
		// Attacking and blocking are mutually exclusive per card.
		card.clearBlocking()
		declared = append(declared, d.Attacker)
	}
	if len(declared) == 0 {
		return nil, ErrNoLegalAttackers
	}
	// One lock-in and one drain for the whole declaration — see the
	// CR 508.2 note in the doc comment. runStateChecksLocked's first
	// act is commitAttackDeclarationLocked, which announces every
	// creature staged above in one event batch.
	g.runStateChecksLocked()
	return declared, nil
}

// DeclareBlocker marks a battlefield card as blocking a specific
// declared attacker: exactly a one-entry DeclareBlockers
// (block_declaration.go), which is where the rules live. Gated by the
// declare_blockers step. Both IDs must exist on the battlefield, and
// the blocker must be a creature.
//
// Returns ErrWrongStep outside the declare_blockers step,
// ErrNotACreature for a non-creature blocker, ErrCardNotFound when
// either card is missing from the battlefield, and a
// *BlockRefusedError (which wraps ErrIllegalBlock) for a pair
// BlockPairRefusalLocked refuses. Idempotent on the same pair.
//
// #750: it can now also refuse for a block COUNT. One creature is not
// a legal block on a menace attacker (CR 702.111b), and this verb has
// no second entry to make it one, so the refusal is
// too_few_blockers and the caller must send both blockers in one
// DeclareBlockers. That is the whole point: the engine used to accept
// the lone block here and undo it later, after the defender had been
// told it was good.
//
// #1339: the attacker must be attacking the blocker's controller, a
// planeswalker they control or a battle they protect (CR 802.4a);
// anything else — including a creature that is not attacking at all —
// is refused with not_defending. #1364: a creature whose planeswalker
// or battle has left combat is still blockable, by the player who was
// defending it (CR 506.4c).
func (g *Game) DeclareBlocker(blockerID, attackerID uuid.UUID) error {
	return g.DeclareBlockers([]BlockDeclaration{{Blocker: blockerID, Attacker: attackerID}})
}

// ResolveCombatDamage applies the regular combat damage step's damage.
// For each card with AttackingTarget set:
//
//   - If it was never blocked, the attacker's CurrentPower is
//     subtracted from the target player's life — recorded via the
//     same path as ChangeLife so the change shows up in LifeHistory.
//   - If it is blocked, its damage goes to the creatures blocking it,
//     and if none are left it deals none at all (CR 510.1c) unless it
//     tramples (CR 702.19d/e). Blocked is the lock-in's record, not
//     the live blocker count (#715, CR 509.1h).
//
// AttackingTarget / BlockingTarget are intentionally NOT cleared
// here — a creature stays in combat until the end of combat step ends
// (CR 511.3), which is where AdvanceStep calls clearCombatLocked, and
// what actually retracts the client's combat arrows.
//
// This is the REGULAR combat damage step only. First-strike damage is
// its own step's turn-based action (CR 510.4, #717) and the cursor
// runs it on entry to StepFirstStrikeDamage; a caller that reaches
// past the cursor and invokes this directly gets the single-step
// combat a board with no first or double strike would have had.
//
// Auto-invoked by AdvanceStep when entering the combat_damage step,
// so under normal play this never needs to be called explicitly.
func (g *Game) ResolveCombatDamage() {
	// internal helper — caller holds g.mu via AdvanceStep, or we
	// take it ourselves below.
	g.resolveCombatDamageLocked()
}

// resolveCombatDamageLocked is the REGULAR combat damage step's
// turn-based action (CR 510.4) — the second of the two combat damage
// steps when StepFirstStrikeDamage happened, and the only one when it
// did not. runStepEntryHooksLocked calls it on entry to
// StepCombatDamage.
//
// Who deals damage here is read off g.firstStrikeStepParticipants, the
// record taken as the FIRST step began, never off the creatures'
// keywords now (#716, CR 702.7c): the second step includes every
// combatant that had neither first strike nor double strike then, plus
// the ones that have double strike right now. Between the two steps
// sits a real priority window, so those two questions genuinely have
// different answers — see participatesInStepLocked.
//
// Combat step tag (#187, ADR 0053 Decision 1): when the first-strike
// step ran, every combat damage event of both steps is stamped with
// Event.CombatStep — CombatStepFirstStrike there, CombatStepRegular
// here. When it did not, the step is "" and nothing is tagged, so the
// tag's presence alone still says there were two beats. The value is
// passed down as an argument, never kept on Game.
//
// The pass delegates to assignAndDealCombatDamageLocked, which handles
// unblocked straight-to-player damage, single-blocker damage with
// trample overflow, multi-blocker via the CR 510.1c damage-assignment
// prompt, menace close-out validation, and the lifelink / deathtouch
// hooks.
func (g *Game) resolveCombatDamageLocked() {
	if g.State != StateActive {
		return
	}
	// Ensure post-layer effective P/T + Abilities are current before
	// reading CurrentPower / HasKeyword. The fast-path no-ops when
	// no relevant event has fired since the last recompute — but
	// after a first-strike step it rarely does, because something
	// died or was cast in the window.
	g.RecomputeLayersIfStaleLocked()

	// Untagged unless the first-strike step ran, which is exactly
	// "the participation record is non-empty": the step exists only
	// when at least one combatant is in it (stepExistsLocked).
	step := ""
	if len(g.firstStrikeStepParticipants) > 0 {
		step = CombatStepRegular
	}
	g.assignAndDealCombatDamageLocked(step)
	g.runStateChecksLocked()

	// Combat state is intentionally left in place. The cursor clears
	// it as it LEAVES end_combat (CR 511.3, #785), which clears
	// AttackingTarget / BlockingTarget — and with them the blocked
	// state and the participation record. Attackers and blockers are
	// therefore still in combat for the whole end of combat step.
}

// resolveFirstStrikeCombatDamageLocked is the FIRST combat damage
// step's turn-based action (CR 510.4, #717). runStepEntryHooksLocked
// calls it on entry to StepFirstStrikeDamage — a step that exists only
// when this pass has someone to deal damage for.
//
// It does two things, in this order:
//
//  1. RECORDS the participation set, which is what "as the first
//     combat damage step began" means (CR 702.7c, #716). Both steps
//     read it; nothing re-reads the keywords afterwards.
//  2. Deals this step's damage and runs the SBA + trigger drain, so
//     creatures that died are gone and the triggers the damage caused
//     are on the stack before the active player receives priority.
//
// The priority window itself is not here: it is what the step BEING a
// step gives, the same way every other step grants priority on entry.
//
// Caller must hold g.mu.
func (g *Game) resolveFirstStrikeCombatDamageLocked() {
	if g.State != StateActive {
		return
	}
	// Same recompute the regular step does, and for the same reason:
	// the record below and the damage pass both read effective
	// keywords and power.
	g.RecomputeLayersIfStaleLocked()
	g.firstStrikeStepParticipants = g.firstStrikeStepParticipantSetLocked()
	g.assignAndDealCombatDamageLocked(CombatStepFirstStrike)
	// SBA + trigger drain so creatures that died to first-strike
	// damage exit, and the triggers that damage caused go on the
	// stack, BEFORE priority (CR 510.3a puts them there first).
	g.runStateChecksLocked()
}

// firstStrikeStepParticipantSetLocked is the CR 510.4 / 702.7c scan:
// the attacking and blocking creatures that have first strike or
// double strike RIGHT NOW. Read at one moment only — the instant the
// first combat damage step would begin — where it answers two
// questions at once:
//
//   - whether that step exists at all (CR 506.1: it does when the set
//     is non-empty), asked by stepExistsLocked; and
//   - who deals damage in it, and by elimination who is left for the
//     second step, recorded on Game.firstStrikeStepParticipants.
//
// One scan, one record, no per-card special cases.
//
// Caller must hold g.mu, and should have recomputed layers — granted
// first strike is a Layer 6 ability like any other.
func (g *Game) firstStrikeStepParticipantSetLocked() map[uuid.UUID]bool {
	var out map[uuid.UUID]bool
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.AttackingTarget == uuid.Nil && c.BlockingTarget == uuid.Nil {
			continue
		}
		if !HasKeyword(c, "first strike") && !HasKeyword(c, "double strike") {
			continue
		}
		if out == nil {
			out = map[uuid.UUID]bool{}
		}
		out[c.InstanceID] = true
	}
	return out
}

// participatesInStepLocked reports whether a combatant deals damage in
// the given combat damage step. Per CR 510.4 and CR 702.7c, read off
// the ONE participation record taken as the first step began:
//
//   - First combat damage step: the creatures in the record — the
//     ones that had first strike or double strike then.
//   - Second combat damage step: the creatures NOT in the record —
//     the ones that had neither — plus any that have double strike
//     now.
//
// Reading live keywords for the second step is the bug in #716: a
// creature whose granted first strike died with its lord in the first
// step would look like a non-first-striker and deal damage twice, and
// one that gained first strike in the priority window between the
// steps would be skipped by a step it never dealt damage in.
//
// An empty record means the first step did not happen, so every
// combatant participates in the single combat damage step — which is
// the same answer the old keyword read gave for a combat with no first
// strike anywhere.
//
// Caller must hold g.mu.
func (g *Game) participatesInStepLocked(c *Card, firstStrike bool) bool {
	if firstStrike {
		return g.firstStrikeStepParticipants[c.InstanceID]
	}
	return !g.firstStrikeStepParticipants[c.InstanceID] || HasKeyword(c, "double strike")
}

// assignAndDealCombatDamageLocked assigns and applies damage for one
// combat damage step. step is that step's Event.CombatStep value:
// CombatStepFirstStrike selects the first-strike step, and anything
// else ("" or CombatStepRegular) is the regular step, tagged only
// when a first-strike step ran before it. Builds per-attacker blocker
// lists (filtered to the step's participants per
// participatesInStepLocked), snapshots power, then iterates
// attackers:
//
//   - Unblocked → straight to AttackingTarget via
//     markCombatDamageToPlayerLocked.
//   - Blocked, but nothing is blocking it any more → no damage at
//     all (CR 510.1c), unless it has trample, which assigns all of
//     it to the player or planeswalker it is attacking
//     (CR 702.19d/e). See the blocked state below.
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
// "Blocked" is READ here, never decided here (#715, CR 509.1h): it is
// Game.blockedAttackers, written once by the block declaration's
// lock-in (commitBlockDeclarationLocked). This pass used to derive it
// from the live battlefield, so an attacker whose only blocker died
// in the first-strike step looked unblocked and hit the player, and a
// menace attacker that lost one of its two blockers had its block
// reverted. Block legality — menace included — is judged at the
// declaration and nowhere else.
//
// Caller must hold g.mu.
func (g *Game) assignAndDealCombatDamageLocked(step string) {
	firstStrike := step == CombatStepFirstStrike
	// ADR 0107 §1: combat damage is dealt simultaneously (CR 510.2),
	// and this loop deals it attacker by attacker. A life total that has
	// taken half of it is not a game state, so the per-event CR 603.8
	// check waits until every assignment here has been dealt.
	defer g.holdStateTriggersLocked()()
	// ADR 0108 §7 decision 6, CR 615.7: while a charged shield is live,
	// the step's damage is collected and dealt all at once as the loop
	// ends, once each shield it meets more of than it can cover has
	// been divided by the player it protects (divide_shield.go). A
	// deferred call runs before the state-trigger hold is released.
	if g.anyChargedShieldLocked() {
		if prev, staged := g.openDamageStageLocked(g.combatDamageInstanceLocked(), true, false); staged {
			defer g.closeDamageStageLocked(prev)
		}
	}

	blockersByAttacker := make(map[uuid.UUID][]int, len(g.Battlefield.Cards))
	// #1706: a blocker that still blocks two or more live attackers
	// divides its damage among them (CR 510.1d) through a prompt,
	// queued after the attacker loop, instead of dealing it to each.
	var dividing []uuid.UUID
	divides := map[uuid.UUID]bool{}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		for _, atk := range c.BlockedAttackers() {
			blockersByAttacker[atk] = append(blockersByAttacker[atk], i)
		}
		if len(c.AlsoBlocking) > 0 && len(g.liveBlockedAttackersLocked(c)) >= 2 {
			divides[c.InstanceID] = true
			dividing = append(dividing, c.InstanceID)
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
			continue // died during this step's iteration
		}
		atkParticipates := g.participatesInStepLocked(atk, firstStrike)
		atkPower := power[atkID]
		// Blocker list for this attacker — taken as the live (not
		// reverted) set. Per-step participation is checked later
		// per-blocker so a first-strike blocker can still hit a
		// vanilla attacker in the first-strike step even when the
		// attacker itself doesn't participate (CR 510.4; blocker
		// damage is independent of the attacker's keywords).
		blkIdxs := blockersByAttacker[atkID]
		liveBlockers := make([]uuid.UUID, 0, len(blkIdxs))
		for _, bi := range blkIdxs {
			blk := &g.Battlefield.Cards[bi]
			if !blk.IsBlockingAttacker(atkID) {
				continue // reverted earlier in this step
			}
			liveBlockers = append(liveBlockers, blk.InstanceID)
		}

		// Attacker's damage — gated on the attacker participating
		// in this step. When it doesn't (e.g. a vanilla attacker in
		// the first-strike step), skip the whole attacker-side
		// branch but fall through to blocker damage below.
		if atkParticipates {
			if len(liveBlockers) == 0 {
				// Nothing is blocking it now — but that is not the
				// same question as whether it is BLOCKED (CR 509.1h).
				// A blocked attacker with no blockers left assigns no
				// combat damage at all (CR 510.1c); a TRAMPLING one
				// assigns all of it to the player or planeswalker it
				// is attacking (CR 702.19d/e). An attacker that was
				// never blocked deals its damage to what it attacked,
				// as it always did.
				blocked := g.attackerBlockedLocked(atkID)
				if atkPower > 0 && (!blocked || HasKeyword(atk, "trample")) {
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
				// marks. The prompt blocks the table
				// (choice_gate.go), so the cursor cannot leave this
				// damage step until it is answered — which is what
				// stops a first-strike assignment being overtaken by
				// the regular step (#702). Until resume, the
				// attacker's damage is NOT applied. Blocker damage
				// still flows (simultaneous)
				// because blockers deal power back regardless of
				// attacker's assignment (CR 510.1d).
				g.queueDamageAssignmentPromptLocked(atk, liveBlockers, atkPower, step)
			}
		}

		// Blockers assign damage to the attacker simultaneously
		// (CR 510.1d). Each blocker's damage flows independently of
		// the attacker's split — and independently of whether the
		// attacker itself participates in this step, so a first-
		// strike blocker can still hit a vanilla attacker in the
		// first-strike step (CR 510.4).
		for _, blkID := range liveBlockers {
			if divides[blkID] {
				continue // divides its damage below (#1706)
			}
			blk := findBattlefieldCard(g, blkID)
			if blk == nil || !g.participatesInStepLocked(blk, firstStrike) {
				continue
			}
			blkPower := power[blkID]
			if blkPower <= 0 {
				continue
			}
			g.markCombatDamageOnCardLocked(atkID, blkPower, blkID, step)
		}
	}

	// #1706, CR 510.1d: a blocker blocking two or more attackers
	// assigns its combat damage "divided as its controller chooses
	// among them". The CR 510.1c prompt carries it, with the roles
	// turned round (queueBlockerDamageDivisionLocked). Its set is
	// re-read now, after the attackers' damage, because the
	// simultaneity is the prompt's: every mark it makes is on the same
	// board the other blockers' marks landed on, and SBAs run only once
	// the step's marks are all down.
	for _, blkID := range dividing {
		blk := findBattlefieldCard(g, blkID)
		if blk == nil || !g.participatesInStepLocked(blk, firstStrike) || power[blkID] <= 0 {
			continue
		}
		if live := g.liveBlockedAttackersLocked(blk); len(live) >= 2 {
			g.queueBlockerDamageDivisionLocked(blk, live, power[blkID], step)
		} else if len(live) == 1 {
			g.markCombatDamageOnCardLocked(live[0], power[blkID], blkID, step)
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
// step is the queuing step's Event.CombatStep value. It is stored on
// the frame as-is ("" when no first-strike step ran) so the resume
// tags the assigned damage the way the same step's direct paths are
// tagged; FirstStrike is kept alongside it as the frame's own record
// of which of the two steps queued it.
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
		// #662: CR 702.16e is read off the attacker as it was when it
		// assigned, for the same reason lifelink and deathtouch are.
		SourceLKI: SourceCharacteristics(atk),
	}
	// ADR 0056 Decision 2: the damage-result keywords are snapshotted
	// by the same reader the direct combat paths use.
	res := SourceDamageResultTraits(atk)
	frame.SourceInfect, frame.SourceWither, frame.SourceToxic = res.Infect, res.Wither, res.ToxicTotal
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
// step is the damage step's Event.CombatStep value (#187). It rides the
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
		// ADR 0108 PR 0, CR 510.2: one instance for the whole step.
		DamageInstance: g.combatDamageInstanceLocked(),
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
// that resolved in the same damage step before the prompt fires.
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
		// ADR 0108 PR 0, CR 510.2: the answered prompt's damage is the
		// step's, so it joins the step's instance.
		DamageInstance: g.combatDamageInstanceLocked(),
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
		// ADR 0108 PR 0, CR 510.2: the step's instance.
		DamageInstance: g.combatDamageInstanceLocked(),
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
// step is the damage step's Event.CombatStep value (#187). It rides the
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
		// ADR 0108 PR 0, CR 510.2: the step's instance.
		DamageInstance: g.combatDamageInstanceLocked(),
	}
	// A CR 616 ordering prompt lands the damage through the same tail
	// when it is answered (#694). Before that fix the pause dropped the
	// life loss entirely.
	_, _ = g.damageThroughReplacementsLocked(ev)
}

// ClearCombat resets every card on the battlefield to "not attacking
// and not blocking". The sandbox verb; the cursor does this by itself
// as the end of combat step ends (CR 511.3, advanceCursorLocked).
// Callable by anyone — the sandbox doesn't gate it. Cheap O(n) pass
// over the battlefield.
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
	// #1218: one bump for the whole sweep, not one per creature, and
	// only when something actually left combat attacking — a combat
	// with no attackers (or none at all this turn) invalidates
	// nothing. See invalidateLayersForAttackChangeLocked.
	attackerLeft := false
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].AttackingTarget != uuid.Nil {
			attackerLeft = true
		}
		g.Battlefield.Cards[i].AttackingTarget = uuid.Nil
		g.Battlefield.Cards[i].ExertOnAttack = false
		g.Battlefield.Cards[i].clearBlocking()
	}
	if attackerLeft {
		g.invalidateLayersForAttackChangeLocked()
	}
	// #830 / #859: the combat declarations' announcements describe
	// the declarations being wiped here, so they go with them.
	g.clearBlockStateLocked()
	g.clearAttackAnnouncementsLocked()
	// #716: and so does the combat damage steps' participation
	// record — it describes these same attackers and blockers.
	g.firstStrikeStepParticipants = nil
}

// Concede marks the given player as eliminated (CR 104.3a). If one
// non-eliminated player remains after the mutation, the game ends and
// Game.Outcome names that player the winner (CR 104.2a,
// OutcomeCauseLastStanding); if none remains, it is a draw.
//
// Never gated: a player who "can't lose the game" can still concede
// (CR 104.3a; the Platinum Angel and Everybody Lives! rulings).
// EventConcede is still emitted for its listeners, but the log
// projects only the EventPlayerEliminated that follows it, whose
// Label is "concede" — one concession, one log line (ADR 0057
// Decision 1).
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
	if g.OpeningRoll != nil {
		// ADR 0121 §1: nobody is the active player yet, nothing is on
		// the stack and nobody has a hand, so the seat leaves through
		// the opening roll's own path.
		g.leaveOpeningRollLocked(p)
		return nil
	}
	if g.MulligansOpen {
		// CR 103.5: a seat that leaves mid-decision is skipped; the
		// next seat decides, and a round or the window may now be done.
		defer g.settleMulliganLocked()
	}
	// #1529: whether the trigger queue is being held for a trigger
	// that is still announcing (see drainPendingTriggersAPNAPLocked).
	heldForAnnouncement := g.triggerAnnouncementOpenLocked()
	// S13.1: delegate to the unified elimination path so concede
	// fires the same stack cleanup + cursor advance + game-end
	// check as an SBA-driven loss. eliminatePlayerLocked leaves with
	// LossConcede and reads no gate.
	g.eliminatePlayerLocked(p)
	// #1289: the departure may have dropped the last prompt a paused
	// resolution was waiting on (ADR 0018 §6's departure table), which
	// finishes that resolution. Its CR 704.3 boundary is owed now.
	g.settleResolutionLocked()
	// #1529: the same for a trigger batch. If the departure dropped
	// the last announcement prompt the drain was holding for, the rest
	// of the batch goes on the stack now rather than waiting for an
	// unrelated action.
	if heldForAnnouncement && g.State == StateActive && !g.triggerAnnouncementOpenLocked() {
		g.runStateChecksLocked()
	}
	// #1501: and for a block declaration. A defender who leaves while
	// priority is parked for their declaration was the one the table
	// was waiting on; if nobody else is, the declaration is over.
	g.settleBlockDeclarationLocked()
	return nil
}

// PassTurn skips to the next player's untap step, regardless of
// whatever step the current turn is in. Useful for forfeiting a turn
// or when all steps are uneventful. Lands on Untap with NoPriority
// (S13); the entry hook auto-untaps and walks the cursor on to
// Upkeep, matching the normal-flow behaviour of priority wraps and
// AdvanceStep so callers always end at a priority-granting step.
//
// The rest of the turn ends through the rotation seam (rotation.go,
// #766): attackers and blockers leave combat, and the cleanup sweep
// removes marked damage and ends "until end of turn" effects. The
// steps in between do not happen — no end step, so no "at the
// beginning of the end step" triggers — and the cleanup discard to
// hand size is skipped. This is a sandbox verb, not a rules action;
// a player who wants the discard and the end step passes priority
// through them instead.
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
	// End the turn through the shared seam so eliminated seats are
	// skipped, the cleanup sweep runs and per-turn caches clear.
	g.clearCombatLocked()
	g.sweepTurnEndLocked()
	g.beginNextTurnLocked()
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
	if g.OpeningRoll != nil {
		// ADR 0121 §1: no hand has been dealt yet.
		return ErrOpeningRollOpen
	}
	if newHandSize < 0 {
		return ErrInvalidParam
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if g.MulligansOpen && g.MulliganDeciderLocked() != p.Seat {
		// CR 103.5: decisions go in turn order.
		return ErrNotYourMulligan
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
	// Answered for this round; the next decision of this seat comes in
	// the next round, after everyone ahead of it has answered again.
	if g.MulligansOpen {
		p.MulliganDecided = true
	}
	g.settleMulliganLocked()
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
	if g.OpeningRoll != nil {
		// ADR 0121 §1: no hand has been dealt yet.
		return ErrOpeningRollOpen
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
	if g.MulligansOpen && g.MulliganDeciderLocked() != p.Seat {
		// CR 103.5: decisions go in turn order.
		return ErrNotYourMulligan
	}
	p.HandKept = true
	p.MulliganDecided = true
	// Close the window once every live seat has kept; otherwise move
	// the turn to the next decider. Eliminated seats don't gate it.
	g.settleMulliganLocked()
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
	// #1335: a plain shuffle is always the player's own library.
	g.EmitEvent(Event{Kind: EventSearchLibrary, Actor: playerID, Target: playerID, Label: "shuffle"})
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
	return g.applyResolvedCounterLocked(out)
}

// applyCounterLocked is the actual counter-map mutation, extracted
// from AddCounter so the replacement pipeline and the AddCounterForEffect
// helper can share the body.
//
// The placement names NO PLACER and no source. That is the right answer
// for every caller that reaches this spelling — a paid cost, a
// rules-driven removal, the CR 704.5q cancel, a sandbox edit — none of
// which is a player putting counters on something (ADR 0056
// Decision 5). A caller that DOES know calls applyCounterByLocked.
//
// Caller must hold g.mu.
func (g *Game) applyCounterLocked(cardID uuid.UUID, name string, delta int) error {
	return g.applyCounterByLocked(cardID, name, delta, uuid.Nil, uuid.Nil)
}

// applyCounterByLocked is applyCounterLocked with the CR 120.3d placer
// and the source card named: `placer` is who PUTS the counters and
// `source` the card whose effect or damage placed them. Both ride onto
// the emitted EventCounterPlaced as Actor and Source, which is how
// "whenever YOU put one or more counters" (Nest of Scarabs, Exemplar of
// Light) stops meaning "whoever resolved anything most recently".
//
// Caller must hold g.mu.
func (g *Game) applyCounterByLocked(cardID uuid.UUID, name string, delta int, placer, source uuid.UUID) error {
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
			// #1824: a counter that can't be removed stays put.
			if delta < 0 && g.counterRemovalLockedLocked(&z.Cards[i], name) {
				return nil
			}
			hadCounters := len(z.Cards[i].Counters) > 0
			if z.Cards[i].Counters == nil {
				z.Cards[i].Counters = make(map[string]int)
			}
			z.Cards[i].Counters[name] += delta
			newAmount := z.Cards[i].Counters[name]
			// ADR 0101 / CR 613.7c: a keyword counter's timestamp moves
			// on every placement and stays put on a removal.
			z.Cards[i].stampKeywordCounter(name, delta, newAmount, timeNowUnixNano)
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
				Actor:  placer,
				Source: source,
			})
			// CR 714.2b: lore counters put on a Saga fire the chapters
			// they crossed, whatever put them there (#2123). After the
			// counter event, so a chapter's trigger sees the count.
			if name == CounterLore && delta > 0 && z == g.Battlefield {
				g.loreCountersPutLocked(cardID, newAmount-delta, newAmount)
			}
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

// SetMonarch designates the given player as the monarch (CR 725 —
// "an effect instructs a player to become the monarch"). Pass uuid.Nil
// to clear (no current monarch — the rare case where a card explicitly
// removes monarchy).
//
// #375: this is the ENTRY to the designation, not the whole mechanic
// any more. Once a player is the monarch, monarch.go runs CR 725.2's
// two inherent triggered abilities — the end-step draw and the
// combat-damage transfer — off the listener registry, so the crown
// moves and draws without anybody clicking it. The action stays
// because a card has to be able to hand the crown out in the first
// place, and because the sandbox posture is that a table can always
// correct the board by hand. A card's effect uses SetMonarchForEffect,
// under the lock it already holds (#1722).
//
// Returns ErrPlayerNotFound if playerID isn't seated or has left the
// game, or ErrGameNotActive in lobby/ended state.
func (g *Game) SetMonarch(playerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if playerID != uuid.Nil {
		if p := g.playerByIDLocked(playerID); p == nil || p.Eliminated {
			return ErrPlayerNotFound
		}
	}
	// #1722: the same write, and the same EventMonarchChanged, as a
	// card's "you become the monarch" (SetMonarchForEffect) — so a
	// table correcting the crown by hand fires "whenever you become
	// the monarch" and re-reads every "as long as you're the monarch"
	// static exactly as the card would have.
	g.writeMonarchLocked(playerID)
	// #1729: and it ends a Palace Jailer's "until an opponent becomes
	// the monarch" now (CR 610.3), not at the next priority pass.
	g.resolveUntilReturnsLocked()
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

// SetGoaded is the sandbox goad: `by` goads the battlefield creature,
// adding to any other player's goad on it (CR 701.15c) or refreshing
// `by`'s own (goadLocked), and the goad ends as `by`'s next turn begins
// like any other (CR 701.15a). Pass uuid.Nil for `by` to clear EVERY
// goad on the creature. The card must be on the battlefield; goading a
// card in any other zone is meaningless.
//
// Since #1571 the marker is enforced: the goaded creature's CR 701.15b
// requirements are judged with every other CR 508.1d requirement
// (attack_requirements.go).
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
			if by == uuid.Nil {
				g.Battlefield.Cards[i].Goads = nil
			} else {
				g.goadLocked(&g.Battlefield.Cards[i], by)
			}
			return nil
		}
	}
	return ErrCardNotFound
}

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
	if p.Agent {
		return ErrSeatKindTaken
	}
	p.IsBot = true
	p.BotTier = tier
	p.BotDeck = deckID
	return nil
}

// SetAgent marks a seat as played by an AI agent through an MCP client
// (ADR 0122 §7), with client the name the client declared. The lobby
// calls it inside the join's own apply, so the first capture that
// shows the seat already carries the badge.
//
// It is one way. There is no ClearAgent and no other writer of
// Player.Agent: once declared, the badge stays for the life of the
// game, across every action, restore and undo
// (TestAgentBadgeHasNoClearingWriter). A second call on a seat that
// is already an agent keeps the first client name, so a later
// declaration cannot relabel the seat either. Lobby state only, like
// SetBot, and refused on a bot seat.
func (g *Game) SetAgent(playerID uuid.UUID, client string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateLobby {
		return ErrGameNotInLobby
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if p.IsBot {
		return ErrSeatKindTaken
	}
	if p.Agent {
		return nil
	}
	p.Agent = true
	p.AgentClient = client
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
		// Lobby seats compact after a departure. Keep deck source ordinals
		// aligned so a later join cannot reuse a surviving card's ordinal.
		for _, zone := range []*Zone{p.Library, p.Command} {
			for _, card := range zone.Cards {
				if ordinal, ok := g.sourceOrdinals[card.InstanceID]; ok && ordinal>>32 == 0 {
					g.setDeckSourceOrdinalLocked(card.InstanceID, i, int(ordinal&0xffff))
				}
			}
		}
	}
	return nil
}

// SetPoison sets a player's poison counter total. 10 is loss in MTG
// (state-based action, deferred to S13+). amount is clamped at 0 from
// below; there's no upper clamp because some cards / formats deal
// arbitrary poison. Replaces (not increments) — clients send the new
// total so two stale tabs don't double-count.
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
	before := p.Counters[CounterPoison]
	p.Poison = amount
	// S13.2: keep the unified Counters map in sync so the SBA loop
	// reads the same value the legacy SetPoison action wrote.
	setPlayerCounterLocked(p, CounterPoison, amount)
	// ADR 0056 Decision 5: the sandbox verbs SKIP the replacement
	// window — this is SET semantics and there is nothing for a
	// replacement to say about "make the total 7" — but they still owe
	// the table the event and the layer bump that rides it, so a
	// "corrupted" static switches on in the same action that sets the
	// poison. No placer: a hand-edit is nobody putting counters.
	g.emitPlayerCounterDeltaLocked(playerID, CounterPoison, before, amount, uuid.Nil, uuid.Nil)
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
// Emits EventPlayerCounterPlaced with the delta that landed, which
// bumps the layer version (ADR 0056 Decision 5). Like set_poison it
// does NOT open the CR 614 counter window: a manual board fix is not a
// player putting counters, and the two verbs are the "make the board
// say this" controls. The effect-time path
// (AddPlayerCounterByForEffect) is the one that goes through the
// window.
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
	mirrorLegacyPlayerCounterLocked(p, name, next)
	// ADR 0056 Decision 5, the same reason as SetPoison above: the
	// sandbox verb skips the window and still emits, so the event and
	// the layer bump do not depend on WHICH way a counter arrived.
	g.emitPlayerCounterDeltaLocked(playerID, name, cur, next, uuid.Nil, uuid.Nil)
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

// SetUndoLimit is the legacy set_undo_limit action's mutation, now a
// thin wrapper over UpdateSettings (ADR 0075 §2.3): it sets
// Settings.UndoLimit, refreshes every seat's UndosRemaining to the new
// value immediately, and emits EventSettingsChanged. actor is the
// player who asked (uuid.Nil for the admin), recorded on the event.
//
// Two behaviours are kept from before the settings struct existed, on
// purpose, because this action's callers were written against them:
//
//   - Only valid while the game is active (ErrGameNotActive). Use
//     UpdateSettings for a lobby-phase change.
//   - A negative limit is clamped to 0, "no undos allowed". That clamp
//     now matters more than it did: UndoUnlimited is -1, and the legacy
//     action must not be a way to switch the budget off by passing a
//     negative number. Unlimited is reachable only through
//     UpdateSettings (the set_table_settings action, ADR 0075 sub-PR 3).
//
// WHO may send the action is decided in the actions / room layer, not
// here.
func (g *Game) SetUndoLimit(actor uuid.UUID, limit int) error {
	if limit < 0 {
		limit = 0
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	return g.updateSettingsLocked(actor, SettingsPatch{UndoLimit: &limit})
}

// SpendUndo decrements the named player's UndosRemaining if they
// have any to spend; returns ErrNoUndosRemaining when the budget is
// exhausted. Called from the room layer immediately before a successful
// RestoreFrom so the budget is only debited on a successful undo.
//
// Returns ErrPlayerNotFound for an unseated playerID. Admin / spectator
// undos (callerID uuid.Nil) bypass this entirely — the room layer
// short-circuits to never call SpendUndo for those.
//
// Under Settings.UndoLimit == UndoUnlimited the budget is never
// debited and never refuses (ADR 0075 §2.2).
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
	if g.Settings.UndoLimit == UndoUnlimited {
		return nil
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
// seated at the BOTTOM of the destination library instead of the top.
// It backs the move_card action's `to_bottom` flag, which the admin
// context menu (#170) needs for "put this on the bottom of your
// library" — the one destination MoveCard (which always PushTops)
// can't express.
//
// #707: the bottom now rides the exit primitive's route
// (zoneRoute.ToBottom) rather than being a post-move reorder. The
// reorder could not survive a pause — a commander tucked to the bottom
// paused on the CR 903.9b prompt was still in its old zone when the
// reorder ran, so it found nothing and the card later landed on TOP —
// and the route already knows how to do this for every effect-side
// tuck (Condemn, Hinder). Like those, it is honoured only against the
// SETTLED destination: a commander whose owner takes the command zone
// is not on the bottom of anything, and a destination that is not a
// library has no bottom worth naming.
func (g *Game) MoveCardByIDToBottom(src, dst ZoneRef, cardID uuid.UUID, asCommander bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.moveCardByRefLocked(src, dst, cardID, asCommander, true)
}

// cancelPlusMinusCountersLocked is CR 704.5q: N +1/+1 and N -1/-1
// counters are removed from a creature that has both, N the smaller
// count. Reports whether it removed any.
//
// Caller must hold g.mu.
func (g *Game) cancelPlusMinusCountersLocked() (fired bool) {
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.IsCreature() || c.Counters == nil {
			continue
		}
		if c.Counters[CounterPlusOne] > 0 && c.Counters[CounterMinusOne] > 0 {
			*c = afterPlusMinusCancel(*c)
			fired = true
		}
	}
	return fired
}

// afterPlusMinusCancel is the card as CR 704.5q leaves it: N of each of
// +1/+1 and -1/-1 removed, N the smaller count, on a fresh counter map.
// A card with not both kinds comes back unchanged.
func afterPlusMinusCancel(c Card) Card {
	plus, minus := c.Counters[CounterPlusOne], c.Counters[CounterMinusOne]
	if plus <= 0 || minus <= 0 {
		return c
	}
	cancel := min(plus, minus)
	counters := copyStringIntMap(c.Counters)
	counters[CounterPlusOne] -= cancel
	counters[CounterMinusOne] -= cancel
	if counters[CounterPlusOne] <= 0 {
		delete(counters, CounterPlusOne)
	}
	if counters[CounterMinusOne] <= 0 {
		delete(counters, CounterMinusOne)
	}
	if len(counters) == 0 {
		counters = nil
		c.LostLastCounter = true
	}
	c.Counters = counters
	return c
}
