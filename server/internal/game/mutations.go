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
func (g *Game) DrawCard(playerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
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

// PlayCard moves a card from the player's hand to the shared
// battlefield. The card's controller is set to the player (already
// the case for cards entering from your own hand, but stored
// explicitly for clarity).
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
// given player. This is what a player does at the start of their
// untap step.
func (g *Game) UntapAll(playerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.playerByIDLocked(playerID) == nil {
		return ErrPlayerNotFound
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Controller == playerID {
			g.Battlefield.Cards[i].Tapped = false
		}
	}
	return nil
}

// PassPriority rotates priority to the next seat. When priority would
// pass back to the active seat (every other seat has passed in
// succession with nothing on the stack), the step auto-advances and
// PriorityHolder is reset to the new ActiveSeat — mirroring real MTG
// rules where priority passing around in succession ends the step.
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
	next := (g.Turn.PriorityHolder + 1) % numSeats
	if next == g.Turn.ActiveSeat {
		// Wrapped — advance the step. Turn.advance resets PriorityHolder
		// to the new ActiveSeat for us.
		g.Turn = g.Turn.advance(numSeats)
		return nil
	}
	g.Turn.PriorityHolder = next
	return nil
}

// PassTurn skips to the next player's untap step, regardless of
// whatever step the current turn is in. Useful for forfeiting a turn
// or when all steps are uneventful.
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
		PriorityHolder: nextSeat,
		Phase:          PhaseOf(StepUntap),
		Step:           StepUntap,
	}
	return nil
}

// Mulligan shuffles the player's entire hand back into their library
// and draws newHandSize cards. This is the simplified "London
// mulligan" shape without the card-to-bottom penalty — S03 does not
// implement the penalty because it's rules enforcement.
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
