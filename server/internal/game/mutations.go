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
// runStepEntryHooksLocked. Caller must hold g.mu.
func (g *Game) drawCardLocked(playerID uuid.UUID) error {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	c, err := p.Library.PopTop()
	if err != nil {
		return err
	}
	p.Hand.PushTop(c)
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
}

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
				break
			}
		}
		return nil
	}
	// Non-land: route through the stack. The card lives in
	// Game.Stack; the announce-time choices live in StackMeta.
	if _, err := MoveCard(src, g.Stack, cardID); err != nil {
		return err
	}
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
	return nil
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
		return nil
	}
	// Instants / sorceries: resolve to the owner's graveyard.
	return g.routeStackCardToGraveyardLocked(top)
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
		_, err := MoveCard(g.Stack, g.Exile, c.InstanceID)
		return err
	}
	_, err := MoveCard(g.Stack, owner.Graveyard, c.InstanceID)
	return err
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
func (g *Game) MoveCardByID(src, dst ZoneRef, cardID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	srcZone := g.zoneFromRefLocked(src)
	if srcZone == nil {
		return ErrZoneNotFound
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
	_, err := MoveCard(srcZone, dstZone, cardID)
	return err
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
			return nil
		}
	}
	return ErrCardNotFound
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
		// Priority returns to the active player after a resolution
		// (CR 117.3b). The step doesn't change.
		g.Turn.PriorityHolder = g.Turn.ActiveSeat
		return nil
	}
	g.Turn = g.Turn.advance(numSeats)
	g.runStepEntryHooksLocked()
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
	p.Eliminated = true

	// If the conceding seat held priority or was the active turn, the
	// turn cursor must advance past them — otherwise the table sits
	// waiting on a player who can never act again. Pass to the next
	// non-eliminated seat in turn order.
	g.advancePastEliminatedLocked()

	// Game-end check: exactly one survivor → StateEnded.
	survivors := 0
	for _, s := range g.Seats {
		if !s.Eliminated {
			survivors++
		}
	}
	if survivors <= 1 {
		g.State = StateEnded
	}
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
	g.Turn = Turn{
		Number:         nextNumber,
		ActiveSeat:     nextSeat,
		PriorityHolder: initialPriorityHolder(StepUntap, nextSeat),
		Phase:          PhaseOf(StepUntap),
		Step:           StepUntap,
	}
	// Refresh per-turn budgets (undo, future per-turn counters) on
	// the new active seat — same hook AdvanceStep / PassPriority's
	// wrap branch run when stepping into untap. The hook also auto-
	// untaps and advances past Untap (no priority) so the cursor
	// lands at Upkeep.
	g.runStepEntryHooksLocked()
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
	for range newHandSize {
		if p.Library.Size() == 0 {
			break
		}
		c, _ := p.Library.PopTop()
		p.Hand.PushTop(c)
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
// deterministic.
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
	return p.ChangeLife(delta), nil
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
			if z.Cards[i].Counters[name] <= 0 {
				delete(z.Cards[i].Counters, name)
				if len(z.Cards[i].Counters) == 0 {
					z.Cards[i].Counters = nil
				}
			}
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
	return nil
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
	return nil
}
