package game

import "github.com/google/uuid"

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
	return g.drawCardLocked(playerID)
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
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	c, err := p.Library.PopTop()
	if err != nil {
		if err == ErrZoneEmpty {
			p.LosesAtNextSBA = true
		}
		return err
	}
	p.Hand.PushTop(c)
	// S13.5: drawn cards become known to the owner. Cards drawn from
	// scry-positioned tops keep any pre-existing scry knowledge via
	// the sticky map; the AddKnower call is idempotent.
	g.markCardKnownInZoneLocked(p.Hand, c.InstanceID)
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
// and special actions (CR 702.79). The flag is mirrored on
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
	// Sorcery-speed gate. Lands are special-action-fast (CR 305 is
	// "you may play a land during your main phase if the stack is
	// empty"); they're handled implicitly by the same gate below.
	// Instants bypass the gate entirely. Anything else (sorceries,
	// permanents that aren't instants) needs sorcery speed.
	requiresSorcerySpeed := !card.IsInstant() && !card.IsLand()
	if card.IsLand() || requiresSorcerySpeed {
		if !g.sorcerySpeedOpenLocked(playerID) {
			return ErrSorcerySpeedRequired
		}
	}
	// Lands skip the stack entirely (CR 305). Move the card to the
	// battlefield and stamp the controller — same shape as PlayCard.
	if card.IsLand() {
		moved, err := MoveCard(src, g.Battlefield, cardID)
		if err != nil {
			return err
		}
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == moved.InstanceID {
				g.Battlefield.Cards[i].Controller = playerID
			}
		}
		g.markCardKnownInZoneLocked(g.Battlefield, moved.InstanceID)
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
		g.fireETBHookLocked(moved.InstanceID, moved.OracleID)
		return nil
	}
	// S15 strict-mode cost gate. Only engaged for non-land casts —
	// lands have no mana cost, and the land-cast branch above has
	// already returned. Strict + payable deducts the cost from the
	// pool; strict + not payable + not forced rejects with a
	// structured InsufficientManaError; permissive or forced emits
	// an EventCostWarning and proceeds without touching the pool
	// (sandbox posture — paper tracking remains valid).
	if err := g.applyCastCostLocked(p, card, params, cardID); err != nil {
		return err
	}
	// Non-land: route through the stack. The card lives in
	// Game.Stack; the announce-time choices live in StackMeta.
	if _, err := MoveCard(src, g.Stack, cardID); err != nil {
		return err
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
	}
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
	g.EmitEvent(Event{
		Kind:   EventCast,
		Actor:  playerID,
		Source: cardID,
		CardID: cardID,
	})
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
// Empty / unparseable ManaCost short-circuits the gate (treats
// the card as costless — matches the sandbox posture for cards
// the importer couldn't parse). Caller must hold g.mu.
func (g *Game) applyCastCostLocked(p *Player, card Card, params CastSpellParams, cardID uuid.UUID) error {
	cost, err := g.effectiveCostLocked(p, card, params)
	if err != nil {
		// Unparseable ManaCost — treat as costless so a Scryfall
		// gap doesn't wedge sandbox casts. Emit a warning for
		// debug visibility.
		g.EmitEvent(Event{
			Kind:     EventCostWarning,
			Actor:    p.ID,
			Source:   cardID,
			ErrorMsg: err.Error(),
		})
		return nil
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
	if !p.ManaPool.CanPay(cost, params.XValue) {
		return &InsufficientManaError{Missing: p.ManaPool.Missing(cost, params.XValue)}
	}
	// Payable under strict mode — commit the spend.
	p.ManaPool.SpendMana(cost, params.XValue)
	g.EmitEvent(Event{
		Kind:   EventManaSpent,
		Actor:  p.ID,
		Source: cardID,
	})
	return nil
}

// effectiveCostLocked parses the card's printed ManaCost and adds
// the commander tax surcharge for casts from the command zone
// (CR 903.8 — each previous cast of THIS commander adds {2} to the
// cost). The CommanderCasts counter is incremented AFTER CastSpell
// reaches the stack, so reading it here returns the prior-cast
// count: first cast pays cost+0, second pays cost+2, third pays
// cost+4. Non-command casts return the raw parsed cost. Errors on
// an unparseable ManaCost.
func (g *Game) effectiveCostLocked(p *Player, card Card, params CastSpellParams) (ParsedCost, error) {
	cost, err := ParseCost(card.ManaCost)
	if err != nil {
		return ParsedCost{}, err
	}
	if params.FromZone == "command" {
		tax := p.CommanderCasts[card.InstanceID]
		cost.Generic += tax * 2
	}
	return cost, nil
}

// castSourceZoneLocked resolves the FromZone string to the zone
// the card should leave from. Unknown values fall back to hand —
// keeps the action forward-compatible with sub-PR 8's "command"
// addition without breaking older clients that omit the field.
func (g *Game) castSourceZoneLocked(p *Player, fromZone string) (*Zone, error) {
	switch fromZone {
	case "", "hand":
		return p.Hand, nil
	case "command":
		return p.Command, nil
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
// graveyard (CR 608.2f). For activated / triggered abilities, the
// item simply ceases to exist (CR 608.2m) — there is no source-card
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
		return g.resolveTopAbilityLocked()
	}
	// Top of the stack is the last card in the slice (LIFO).
	top := g.Stack.Cards[len(g.Stack.Cards)-1]
	item, ok := g.StackMeta[top.InstanceID]
	if !ok || item == nil {
		// Defensive: a card on the stack without a meta entry should
		// not happen — but if it does, route to graveyard so the
		// stack doesn't wedge.
		return g.routeStackCardToGraveyardLocked(top)
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
		return g.routeStackCardToGraveyardLocked(top)
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
	g.fireEffectResolverLocked(item, top.OracleID, top.InstanceID)
	if top.IsPermanent() {
		// Permanents resolve to the battlefield with the announce-time
		// controller (which may differ from owner — e.g. cast via a
		// "play this from exile" effect that change controller).
		moved, err := MoveCard(g.Stack, g.Battlefield, top.InstanceID)
		if err != nil {
			return err
		}
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == moved.InstanceID {
				g.Battlefield.Cards[i].Controller = item.Controller
				break
			}
		}
		g.markCardKnownInZoneLocked(g.Battlefield, moved.InstanceID)
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			Actor:   item.Controller,
			CardID:  moved.InstanceID,
			OldZone: ZoneStack,
			NewZone: ZoneBattlefield,
		})
		g.EmitEvent(Event{
			Kind:   EventETB,
			Actor:  item.Controller,
			CardID: moved.InstanceID,
		})
		g.fireETBHookLocked(moved.InstanceID, moved.OracleID)
		return nil
	}
	// Instants / sorceries: resolve to the owner's graveyard.
	return g.routeStackCardToGraveyardLocked(top)
}

// spellAllTargetsIllegalLocked reports whether a resolved stack item
// has at least one targeted slot (player or card) and every one of
// those targets is now illegal per the CR 608.2b existence check.
// A slot with Kind Self or None is always legal. Items with no
// targets at all return false (nothing to re-check).
//
// Caller must hold g.mu.
func spellAllTargetsIllegalLocked(g *Game, item *StackItem) bool {
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
			if targetStillExistsLocked(g, t) {
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

// resolveTopAbilityLocked resolves the most-recently-added ability
// item in StackMeta (no underlying card on Game.Stack). Returns nil
// if there are no abilities to resolve. Caller must hold g.mu.
func (g *Game) resolveTopAbilityLocked() error {
	if len(g.StackMeta) == 0 {
		return nil
	}
	// Without an ordering hint, pick any ability item — sub-PR 6
	// will replace this with proper LIFO ordering once
	// activate_ability / announce_trigger are wired.
	for id, item := range g.StackMeta {
		if item != nil && (item.Kind == StackItemActivated || item.Kind == StackItemTriggered) {
			delete(g.StackMeta, id)
			g.recomputeSplitSecondLocked()
			g.EmitEvent(Event{
				Kind:   EventResolve,
				Actor:  item.Controller,
				Source: item.SourceCardID,
			})
			return nil
		}
	}
	return nil
}

// routeStackCardToGraveyardLocked moves a card off Game.Stack and
// into its owner's graveyard. Used by resolveTopOfStackLocked for
// instants / sorceries and by the "countered by game rules" path
// (sub-PR 3) when every target is illegal on resolve. Caller must
// hold g.mu.
func (g *Game) routeStackCardToGraveyardLocked(c Card) error {
	owner := g.playerByIDLocked(c.Owner)
	if owner == nil {
		// Owner is no longer seated — drop the card to exile so the
		// stack doesn't carry a reference to a dead player. (S13.1
		// sub-PR 10 will fold this into the leaving-game cleanup.)
		if _, err := MoveCard(g.Stack, g.Exile, c.InstanceID); err != nil {
			return err
		}
		g.markCardKnownInZoneLocked(g.Exile, c.InstanceID)
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			CardID:  c.InstanceID,
			OldZone: ZoneStack,
			NewZone: ZoneExile,
		})
		return nil
	}
	if _, err := MoveCard(g.Stack, owner.Graveyard, c.InstanceID); err != nil {
		return err
	}
	g.markCardKnownInZoneLocked(owner.Graveyard, c.InstanceID)
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		Actor:   owner.ID,
		CardID:  c.InstanceID,
		OldZone: ZoneStack,
		NewZone: ZoneGraveyard,
	})
	return nil
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
// item (CR 608.2m).
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
	}
	return nil
}

// ActivateLoyalty applies a planeswalker's loyalty ability. Sandbox
// shape: the engine doesn't model the activation as a proper stack
// item (full loyalty-on-stack lands in S14+ alongside the effect
// catalog). Instead, the loyalty delta is applied immediately and
// the once-per-turn flag is set. Sorcery-speed gate enforced per
// CR 606.5.
//
// `delta` is the loyalty change announced by the ability:
// +1 / +2 / -3 / etc. Applied via AddCounter to the planeswalker's
// "loyalty" counter; negative deltas that would drive loyalty below
// zero are clamped (the planeswalker leaves via SBA in sub-PR 7).
//
// Returns:
//   - ErrCardNotFound if planeswalkerID is not on the battlefield.
//   - ErrSorcerySpeedRequired if the gate is closed.
//   - ErrLoyaltyAlreadyActivated if the planeswalker has already
//     activated a loyalty ability this turn (CR 606.5).
//
// Caller must NOT hold g.mu — this method takes the write lock.
//
// S13.1.
func (g *Game) ActivateLoyalty(playerID, planeswalkerID uuid.UUID, label string, delta int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
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
	if pw.Counters == nil {
		pw.Counters = make(map[string]int)
	}
	pw.Counters["loyalty"] += delta
	if pw.Counters["loyalty"] <= 0 {
		// Loyalty reaching zero is the SBA's responsibility (sub-PR 7);
		// here we just keep the bookkeeping clean by removing the key
		// when it drops to zero. SBA will pick up "loyalty == 0" by
		// counting positive entries.
		if pw.Counters["loyalty"] == 0 {
			delete(pw.Counters, "loyalty")
		}
		// Negative is allowed for the moment — clamping would discard
		// "minus more than you have" intent that the SBA needs to
		// see. Sub-PR 7 will read this as "loyalty <= 0".
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
// the loop (CR 704.4 — SBAs repeat until none fire).
//
// Player-loss SBAs (S13.1):
//   - 704.5a: a player at 0 or less life loses
//   - 704.5b: a player who tried to draw from an empty library loses
//   - 704.5v / 903.14a: 21 commander damage from a single source
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
//   - 704.5p: a battle with 0 defense counters is moved to its
//     owner's graveyard
//   - 704.5q: +1/+1 and -1/-1 counters on the same creature
//     cancel out — remove min(N, M) of each
//   - 704.5u: a saga whose final-chapter lore counter is set is
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
	fired := false

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
			}
			fired = true
		}
	}

	// Player-loss SBAs.
	for _, p := range g.Seats {
		if p.Eliminated {
			continue
		}
		if p.Life <= 0 || p.LosesAtNextSBA || p.IsDeadByCommanderDamage() {
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
	// destroy every demo seed card.
	var doomed []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsCreature() {
			if c.Toughness == 0 && len(c.Counters) == 0 {
				continue
			}
			curT := c.CurrentToughness()
			if curT <= 0 {
				doomed = append(doomed, c.InstanceID)
				continue
			}
			if c.DamageMarked >= curT {
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
		// 704.5p — battle with 0 defense counters.
		if c.IsBattle() {
			if c.Counters == nil || c.Counters[CounterDefense] <= 0 {
				doomed = append(doomed, c.InstanceID)
			}
			continue
		}
	}
	for _, id := range doomed {
		if err := g.routeBattlefieldCardToOwnerGraveyardLocked(id); err == nil {
			fired = true
		}
	}

	return fired
}

// runStateChecksLocked runs the SBA + APNAP-trigger-drain loop until
// the game is quiet (no SBAs fire AND no triggers are pending).
// CR 704.4 + 603.3b — both checks are paired at every priority-grant
// boundary.
//
// Bounded at 32 iterations as a safety belt against an unintended
// SBA / trigger ping-pong; in practice the loop terminates after at
// most a handful of passes (one creature destroyed → one Concede-
// adjacent trigger → one resolution).
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
		if hasPending {
			g.runStateChecksLocked()
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
	p.LosesAtNextSBA = false
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
	g.recomputeSplitSecondLocked()
}

// routeBattlefieldCardToOwnerGraveyardLocked moves a battlefield
// card to its owner's graveyard, clearing battlefield-only state
// (combat declarations and damage marked are zeroed by MoveCard's
// CR 400.7 cleanup; we additionally clear DamageMarked here since
// MoveCard predates the field). Used by SBAs that destroy creatures.
//
// If the owner is no longer seated, the card lands in exile so the
// engine doesn't carry a stale reference. Caller must hold g.mu.
func (g *Game) routeBattlefieldCardToOwnerGraveyardLocked(cardID uuid.UUID) error {
	var owner *Player
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			owner = g.playerByIDLocked(g.Battlefield.Cards[i].Owner)
			// Zero battlefield-only state in place before the move.
			// MoveCard handles tapped + counters + position via
			// existing CR 400.7 cleanup, but DamageMarked is new.
			g.Battlefield.Cards[i].DamageMarked = 0
			break
		}
	}
	if owner == nil {
		if _, err := MoveCard(g.Battlefield, g.Exile, cardID); err != nil {
			return err
		}
		g.markCardKnownInZoneLocked(g.Exile, cardID)
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			CardID:  cardID,
			OldZone: ZoneBattlefield,
			NewZone: ZoneExile,
		})
		g.EmitEvent(Event{Kind: EventLTB, CardID: cardID})
		return nil
	}
	if _, err := MoveCard(g.Battlefield, owner.Graveyard, cardID); err != nil {
		return err
	}
	g.markCardKnownInZoneLocked(owner.Graveyard, cardID)
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		Actor:   owner.ID,
		CardID:  cardID,
		OldZone: ZoneBattlefield,
		NewZone: ZoneGraveyard,
	})
	g.EmitEvent(Event{Kind: EventLTB, CardID: cardID, Actor: owner.ID})
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
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			g.Battlefield.Cards[i].DamageMarked += delta
			if g.Battlefield.Cards[i].DamageMarked < 0 {
				g.Battlefield.Cards[i].DamageMarked = 0
			}
			if delta > 0 {
				g.EmitEvent(Event{
					Kind:   EventDealDamage,
					Target: cardID,
					Amount: delta,
				})
			}
			g.runStateChecksLocked()
			return nil
		}
	}
	return ErrCardNotFound
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
// S13.1.
func (g *Game) drainPendingTriggersAPNAPLocked() {
	if len(g.PendingTriggers) == 0 {
		return
	}
	numSeats := len(g.Seats)
	if numSeats == 0 {
		return
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
	g.PendingTriggers = nil
	if g.StackMeta == nil {
		g.StackMeta = make(map[uuid.UUID]*StackItem)
	}
	// APNAP: active player first, then clockwise.
	for offset := 0; offset < numSeats; offset++ {
		seat := (g.Turn.ActiveSeat + offset) % numSeats
		for _, t := range bySeat[seat] {
			g.StackMeta[t.ID] = t
		}
	}
	g.recomputeSplitSecondLocked()
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
	item, ok := g.StackMeta[spellID]
	if !ok || item == nil || item.Kind != StackItemSpell {
		return ErrCardNotOnStack
	}
	if g.Stack == nil || !g.Stack.Contains(spellID) {
		return ErrCardNotOnStack
	}
	// Resolve the destination. nil → owner's graveyard. Battlefield /
	// stack are illegal — see method docstring.
	var destZone *Zone
	if dst == nil {
		owner := g.playerByIDLocked(item.Owner)
		if owner == nil {
			// Owner has left the game — exile rather than wedging.
			destZone = g.Exile
		} else {
			destZone = owner.Graveyard
		}
	} else {
		if dst.Kind == ZoneBattlefield || dst.Kind == ZoneStack {
			return ErrInvalidStackDestination
		}
		destZone = g.zoneFromRefLocked(*dst)
		if destZone == nil {
			return ErrZoneNotFound
		}
	}
	if _, err := MoveCard(g.Stack, destZone, spellID); err != nil {
		return err
	}
	g.markCardKnownInZoneLocked(destZone, spellID)
	delete(g.StackMeta, spellID)
	g.recomputeSplitSecondLocked()
	g.EmitEvent(Event{
		Kind:   EventCounterSpell,
		Target: spellID,
		CardID: spellID,
	})
	return nil
}

// CounterAbility removes an activated / triggered ability from the
// stack. Abilities cease to exist on resolution (CR 608.2m); a
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
	// Commander zone replacement (CR 903.9): if the caller flagged
	// the move and the card is a commander headed to a destination
	// the rule covers, rewrite dst to the owner's command zone
	// before resolving the destination zone.
	if asCommander {
		if rewritten, ok := g.applyCommanderZoneReplacementLocked(dst, cardID); ok {
			dst = rewritten
		}
	}
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
	if _, err := MoveCard(srcZone, dstZone, cardID); err != nil {
		return err
	}
	g.markCardKnownInZoneLocked(dstZone, cardID)
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		CardID:  cardID,
		OldZone: srcZone.Kind,
		NewZone: dstZone.Kind,
	})
	if srcZone.Kind == ZoneBattlefield {
		g.EmitEvent(Event{Kind: EventLTB, CardID: cardID})
	}
	if dstZone.Kind == ZoneBattlefield {
		g.EmitEvent(Event{Kind: EventETB, CardID: cardID})
		// ETB hook for admin direct-drop onto battlefield (and the
		// commander zone-replacement destination). Find the card in
		// the destination zone to pull its Scryfall ID.
		for _, c := range dstZone.Cards {
			if c.InstanceID == cardID {
				g.fireETBHookLocked(cardID, c.OracleID)
				break
			}
		}
	}
	return nil
}

// applyCommanderZoneReplacementLocked returns a rewritten ZoneRef
// pointing at the owner's command zone if the move is eligible for
// the CR 903.9 replacement (the card is a commander and the
// destination is graveyard / exile / hand / library). Returns
// (input, false) when not eligible. Caller must hold g.mu.
func (g *Game) applyCommanderZoneReplacementLocked(dst ZoneRef, cardID uuid.UUID) (ZoneRef, bool) {
	switch dst.Kind {
	case ZoneGraveyard, ZoneExile, ZoneHand, ZoneLibrary:
		// Eligible destination.
	default:
		return dst, false
	}
	// Find the card and confirm it's a commander.
	z := g.findCardZoneLocked(cardID)
	if z == nil {
		return dst, false
	}
	for _, c := range z.Cards {
		if c.InstanceID != cardID {
			continue
		}
		if !c.IsCommander {
			return dst, false
		}
		return ZoneRef{Kind: ZoneCommand, Owner: c.Owner}, true
	}
	return dst, false
}

// TapCard sets the tapped state of a card on the battlefield. Returns
// ErrCardNotFound if the card is not currently on the battlefield —
// tapping a card in any other zone is meaningless.
func (g *Game) TapCard(cardID uuid.UUID, tapped bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			g.Battlefield.Cards[i].Tapped = tapped
			if tapped {
				g.EmitEvent(Event{Kind: EventTapCard, CardID: cardID})
			} else {
				g.EmitEvent(Event{Kind: EventUntapCard, CardID: cardID})
			}
			return nil
		}
	}
	return ErrCardNotFound
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
// controller's commander's color identity before queuing the pick.
//
// Added in S15 sub-PR 2.
func (g *Game) ActivateManaAbility(playerID, cardID uuid.UUID, abilityIdx int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
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
	abilities := ManaAbilitiesForCard(*card)
	if abilityIdx < 0 || abilityIdx >= len(abilities) {
		return ErrInvalidParam
	}
	ab := abilities[abilityIdx]
	// Pay costs. Tap: the card must be untapped; we flip it, emit the
	// tap event. Sacrifice: not exercised by S15 catalog; rejected
	// here as unsupported so a future Lotus-Petal-style spec fails
	// loudly rather than silently producing mana without a cost.
	if ab.SacrificeCost {
		return ErrInvalidParam
	}
	if ab.TapCost {
		if card.Tapped {
			return ErrAlreadyTapped
		}
		card.Tapped = true
		g.EmitEvent(Event{Kind: EventTapCard, Actor: playerID, CardID: cardID})
	}
	g.EmitEvent(Event{
		Kind:   EventManaAbilityActivated,
		Actor:  playerID,
		Source: cardID,
	})
	// Materialise the produced mana.
	slots, err := ParseProducedMana(ab.Produced)
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
			// Single-color slot — straight into the pool.
			p.ManaPool.AddMana(ManaToken{Color: options[0], Source: cardID})
			g.EmitEvent(Event{Kind: EventManaAdded, Actor: playerID, Source: cardID})
			continue
		}
		// Multi-option slot — intersect with the controller's commander
		// identity (Arcane Signet). When no commander exists or the
		// identity is empty (placeholder commanders from the demo seed),
		// fall through to the raw option set so Birds of Paradise still
		// offers the full five colors.
		filtered := filterPipeByCommanderIdentity(options, p)
		// Queue the pick.
		g.QueueChoiceForEffect(PendingChoice{
			Kind:         PendingChoiceMana,
			Chooser:      playerID,
			FromPlayer:   playerID,
			Count:        1,
			Source:       cardID,
			Reason:       ab.Label,
			ColorOptions: filtered,
		})
	}
	return nil
}

// ManaAbilitiesForCard returns the abilities available on a card,
// preferring the catalog declaration and falling back to the
// synthetic basic-land shape (Forest → "{G}", etc.) when the catalog
// has nothing registered. The protocol layer calls this to stamp
// CardView.ManaAbilities; the dispatcher calls it to resolve an
// incoming ability index. Caller must hold g.mu (the fallback path
// reads the card's TypeLine, which is stable under lock).
func ManaAbilitiesForCard(c Card) []ManaAbilityShape {
	if CatalogManaAbilities != nil {
		if list := CatalogManaAbilities(c.OracleID); len(list) > 0 {
			return list
		}
	}
	if color := basicLandColor(c.TypeLine); color != "" {
		return []ManaAbilityShape{{
			TapCost:  true,
			Produced: "{" + color + "}",
			Label:    "Add {" + color + "}",
		}}
	}
	return nil
}

// basicLandColor returns the single-letter mana color produced by a
// basic land subtype on a card's TypeLine. Lowercase substring check
// so "Basic Land — Forest" and "Land — Forest" both match.
// Multi-type basic lands (Snow-Covered basics, Wastes) fall out of
// this path and would need catalog entries; for S15 we ship only
// the five plain basics.
func basicLandColor(typeLine string) string {
	if !typeLineHas(typeLine, "basic") || !typeLineHas(typeLine, "land") {
		return ""
	}
	switch {
	case typeLineHas(typeLine, "plains"):
		return "W"
	case typeLineHas(typeLine, "island"):
		return "U"
	case typeLineHas(typeLine, "swamp"):
		return "B"
	case typeLineHas(typeLine, "mountain"):
		return "R"
	case typeLineHas(typeLine, "forest"):
		return "G"
	}
	return ""
}

// filterPipeByCommanderIdentity narrows `options` to just the colors
// in the active player's commander's color identity. Used by the
// `"{W|U|B|R|G}"` → Arcane Signet path. If the player has no
// commander (or the commander has no color identity), returns the
// raw options unchanged — Birds of Paradise falls through this path
// with the full 5-color set intact.
func filterPipeByCommanderIdentity(options []string, p *Player) []string {
	// Find the player's commander on the battlefield or in the
	// command zone. The color identity is on game.Card directly
	// (copied at deck-import time via cards.Card.ColorIdentity).
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
func commanderIdentityFor(p *Player) []string {
	if p == nil || p.Command == nil {
		return nil
	}
	for _, c := range p.Command.Cards {
		if !c.IsCommander {
			continue
		}
		// Card.ColorIdentity isn't carried on game.Card today — the
		// deck importer keeps it on cards.Card only. S15 doesn't
		// thread it yet; use the TypeLine / ManaCost as a proxy by
		// scanning the card's ManaCost for WUBRG letters. Good
		// enough for every sandbox Commander whose commander's mana
		// cost hints at the identity (true for >99% of EDH
		// commanders — the escape hatch ("has color abilities in
		// rules text") lands alongside S17 layer work).
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

// untapAllForLocked untaps every battlefield card controlled by the
// given seat. Used both by the public UntapAll mutation and by the
// StepUntap auto-action in runStepEntryHooksLocked. A negative or
// out-of-range seat is a no-op (the caller has already validated the
// seat or is the auto-fire path which only calls with the active
// seat). Caller must hold g.mu.
func (g *Game) untapAllForLocked(seat int) {
	if seat < 0 || seat >= len(g.Seats) {
		return
	}
	playerID := g.Seats[seat].ID
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Controller == playerID {
			g.Battlefield.Cards[i].Tapped = false
		}
	}
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
	prev := g.Turn
	g.Turn = g.Turn.advance(numSeats)
	g.onTurnAdvanceLocked(prev, g.Turn)
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
// unknown target player, and ErrCardNotFound for an unknown
// attacker card. Re-declaring the same attacker against a different
// target overwrites the previous target.
//
// Caller authorization (was-it-the-controller) is intentionally
// NOT enforced — sandbox flexibility for casual play. Untap state
// and summoning sickness are also not enforced; those land with
// rules enforcement in S13+.
func (g *Game) DeclareAttacker(attackerID, targetPlayerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.Turn.Step != StepDeclareAttackers {
		return ErrWrongStep
	}
	if g.playerByIDLocked(targetPlayerID) == nil {
		return ErrPlayerNotFound
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == attackerID {
			if !g.Battlefield.Cards[i].IsCreature() {
				return ErrNotACreature
			}
			g.Battlefield.Cards[i].AttackingTarget = targetPlayerID
			// A card declared as attacker can't simultaneously be a
			// blocker — clearing the other field keeps the per-card
			// combat state coherent.
			g.Battlefield.Cards[i].BlockingTarget = uuid.Nil
			return nil
		}
	}
	return ErrCardNotFound
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
	// Verify the attacker exists on the battlefield. Without this the
	// blocker would silently point at a non-existent attacker ID.
	attackerExists := false
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == attackerID {
			attackerExists = true
			break
		}
	}
	if !attackerExists {
		return ErrCardNotFound
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == blockerID {
			if !g.Battlefield.Cards[i].IsCreature() {
				return ErrNotACreature
			}
			g.Battlefield.Cards[i].BlockingTarget = attackerID
			g.Battlefield.Cards[i].AttackingTarget = uuid.Nil
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

func (g *Game) resolveCombatDamageLocked() {
	if g.State != StateActive {
		return
	}
	// Pre-compute the set of attacker IDs that have at least one
	// declared blocker. O(n) two-pass keeps the per-attacker check
	// O(1) even with many blockers.
	blocked := make(map[uuid.UUID]bool, len(g.Battlefield.Cards))
	for _, c := range g.Battlefield.Cards {
		if c.BlockingTarget != uuid.Nil {
			blocked[c.BlockingTarget] = true
		}
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.AttackingTarget == uuid.Nil {
			continue
		}
		if blocked[c.InstanceID] {
			continue
		}
		target := g.playerByIDLocked(c.AttackingTarget)
		if target == nil {
			continue
		}
		damage := c.CurrentPower()
		if damage <= 0 {
			continue
		}
		// Player.ChangeLife appends a LifeHistory entry. We're inside
		// the game's write lock; ChangeLife operates on the player
		// struct directly without any internal locking, so this is
		// safe.
		target.ChangeLife(-damage)
	}
	// Combat state is intentionally left in place here. AdvanceStep's
	// transition into end_combat invokes clearCombatLocked, which is
	// what actually clears AttackingTarget / BlockingTarget. Deferring
	// the clear lets the client keep combat arrows drawn for the full
	// duration of the combat_damage step.
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
	nextSeat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	nextNumber := g.Turn.Number
	if nextSeat == 0 {
		nextNumber++
	}
	prev := g.Turn
	g.Turn = Turn{
		Number:         nextNumber,
		ActiveSeat:     nextSeat,
		PriorityHolder: initialPriorityHolder(StepUntap, nextSeat),
		Phase:          PhaseOf(StepUntap),
		Step:           StepUntap,
	}
	g.onTurnAdvanceLocked(prev, g.Turn)
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
	p.Library.Shuffle(g.rng)
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
// using the RNG captured by Start so deterministic test runs stay
// deterministic. S13.5: clears KnownBy on every library card —
// any prior scry / top-of-library knowledge dissolves with the
// shuffle (CR 701.20 + the per-instance KnownBy invariant).
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
	p.Library.Shuffle(g.rng)
	clearKnownInZoneLocked(p.Library)
	g.EmitEvent(Event{Kind: EventSearchLibrary, Actor: playerID, Label: "shuffle"})
	return nil
}

// ChangePlayerLife adjusts a player's life total by delta (positive
// for gain, negative for loss) and returns the new total.
func (g *Game) ChangePlayerLife(playerID uuid.UUID, delta int) (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return 0, ErrGameNotActive
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return 0, ErrPlayerNotFound
	}
	newLife := p.ChangeLife(delta)
	g.EmitEvent(Event{
		Kind:   EventChangeLife,
		Target: playerID,
		Amount: delta,
	})
	return newLife, nil
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
	z := g.findCardZoneLocked(cardID)
	if z == nil {
		return ErrCardNotFound
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID == cardID {
			if delta == 0 {
				return nil
			}
			if z.Cards[i].Counters == nil {
				z.Cards[i].Counters = make(map[string]int)
			}
			z.Cards[i].Counters[name] += delta
			newAmount := z.Cards[i].Counters[name]
			if z.Cards[i].Counters[name] <= 0 {
				delete(z.Cards[i].Counters, name)
				if len(z.Cards[i].Counters) == 0 {
					z.Cards[i].Counters = nil
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

// SetCommanderDamage sets the total commander damage dealt to the
// target player by the given opponent's commander(s). This is a set
// operation, not an increment — the client computes and sends the new
// total. Intended to back a 4×4 commander-damage grid UI.
//
// Both `from` and `to` must be IDs of seated players; passing an ID
// that doesn't correspond to a seated player returns ErrPlayerNotFound
// rather than silently polluting the target's damage map.
//
// NOTE: CommanderDamage is keyed by opponent player ID, not by
// commander instance ID, so partner commanders are not correctly
// distinguished. See the package-level note on Player.CommanderDamage
// in player.go — S10 fix.
func (g *Game) SetCommanderDamage(from, to uuid.UUID, amount int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.playerByIDLocked(from) == nil {
		return ErrPlayerNotFound
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
