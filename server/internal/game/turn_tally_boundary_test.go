package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// turn_tally_boundary_test.go — #1009. Two dimensions the tally grew
// so that the last "this turn" questions in the catalog could stop
// walking the event log:
//
//   - TurnTally.Entered, the per-OBJECT "did this permanent enter this
//     turn" cell, whose whole point is the boundary: the walk it
//     replaces stopped at EventBeginUpkeep, and the turn does not
//     begin at the upkeep.
//   - TurnTally.CombatDamagedPlayers, "which players did a creature of
//     mine with this name / this subtype hit this turn", recorded at
//     the damage because the creature that dealt it is the one thing
//     combat is most likely to remove before anybody asks.

// pushPermanent puts a permanent on the battlefield without any
// entry machinery — the tests below drive the listener with the
// events they mean to test and nothing else.
func pushPermanent(g *Game, c Card) uuid.UUID {
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	g.WithWriteLock(func() { g.Battlefield.PushTop(c) })
	return c.InstanceID
}

// removeFromBattlefield is CR 704.5d for a token, spelled out: the
// object is gone from every zone and no later lookup can find it.
func removeFromBattlefield(g *Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards = append(g.Battlefield.Cards[:i], g.Battlefield.Cards[i+1:]...)
				return
			}
		}
	})
}

// THE BUG (#1009). The upkeep is not the start of the turn. A
// permanent that entered during the untap step — an untap-step
// trigger or an untap-step choice putting one onto the battlefield
// (#70, ADR 0070) — is before EventBeginUpkeep, and the walk this
// replaced stopped there and answered "it did not enter this turn".
func TestEnteredThisTurnCountsAnEntryBeforeTheUpkeep(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	id := pushPermanent(g, Card{Name: "Untap Step Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Owner: me, Controller: me})

	// The turn begins. resetTurnTallyLocked is what onTurnBeganLocked
	// calls, and it runs BEFORE the untap step's turn-based actions.
	g.WithWriteLock(func() { g.resetTurnTallyLocked() })
	// The untap step puts the permanent onto the battlefield…
	emit(g, Event{Kind: EventETB, CardID: id})
	if !g.EnteredThisTurn(id) {
		t.Fatal("setup: the entry was not counted at all")
	}
	// …and then the upkeep begins, which used to erase the answer.
	emit(g, Event{Kind: EventBeginUpkeep, Actor: me})
	if !g.EnteredThisTurn(id) {
		t.Error("a permanent that entered in the untap step entered THIS TURN; the upkeep is not the turn boundary")
	}
	if g.EnteredThisTurn(uuid.New()) {
		t.Error("a permanent that never entered did not enter this turn")
	}
	if g.EnteredThisTurn(uuid.Nil) {
		t.Error("the nil ID is nobody")
	}
}

// The other half of the boundary: it IS a boundary. The tally resets
// when the turn begins, so last turn's entries are gone.
func TestEnteredThisTurnResetsAtTheTurnBoundary(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	id := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me, Controller: me})
	emit(g, Event{Kind: EventETB, CardID: id})
	if !g.EnteredThisTurn(id) {
		t.Fatal("setup: the entry was not counted")
	}

	seat := g.Turn.ActiveSeat
	for i := 0; i < 40 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat == seat {
		t.Fatal("setup: the turn never passed")
	}
	if g.EnteredThisTurn(id) {
		t.Error("\"entered this turn\" does not reach back into the previous turn")
	}
}

// Counted per ENTRY: a permanent blinked twice in one turn entered
// twice, and the cell says so even though nothing reads the count yet.
func TestEnteredThisTurnCountsEveryEntry(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	id := pushPermanent(g, Card{Name: "Blinker", TypeLine: "Creature — Spirit", Power: 1, Toughness: 1, Owner: me, Controller: me})
	emit(g, Event{Kind: EventETB, CardID: id}, Event{Kind: EventETB, CardID: id})
	if got := g.TurnTally.Entered[id]; got != 2 {
		t.Errorf("two entries counted %d times", got)
	}
}

// #596, the combat half. A Faerie token connects and then trades.
// CR 704.5d takes it out of the graveyard at the next state-based
// check, so by the time Alela's target predicate asks "which players
// did my Faeries hit this turn" there is no object anywhere to look
// up — which is exactly when the question is asked, because a target
// predicate is not handed the trigger's event.
func TestCombatDamageTallyRemembersATokenThatTraded(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp, third := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	tok := pushPermanent(g, Card{
		Name: "Faerie Rogue", TypeLine: "Token Creature — Faerie Rogue",
		Power: 1, Toughness: 1, Owner: me, Controller: me,
	})
	emit(g, Event{Kind: EventDealDamage, Actor: me, Source: tok, Target: opp, Amount: 1, Combat: true})
	removeFromBattlefield(g, tok)
	if _, ok := g.LookupCardForEffect(tok); ok {
		t.Fatal("setup: the token should be nowhere")
	}

	hit := g.PlayersDealtCombatDamageThisTurnBySubtype(me, "Faerie")
	if !hit[opp] {
		t.Error("a Faerie token that traded still dealt the damage it dealt")
	}
	if hit[third] || hit[me] {
		t.Errorf("only the damaged player is in the set: %v", hit)
	}
	if named := g.PlayersDealtCombatDamageThisTurnByName(me, "Faerie Rogue"); !named[opp] {
		t.Error("the same record answers by name, for Trygon Predator's clause")
	}
	if named := g.PlayersDealtCombatDamageThisTurnByName(me, "Faerie"); named[opp] {
		t.Error("a name and a subtype that spell the same must not collide")
	}
	if wrong := g.PlayersDealtCombatDamageThisTurnBySubtype(opp, "Faerie"); len(wrong) != 0 {
		t.Errorf("the set is per controller: %v", wrong)
	}
	if none := g.PlayersDealtCombatDamageThisTurnBySubtype(me, "Goblin"); len(none) != 0 {
		t.Errorf("no Goblin of mine hit anyone: %v", none)
	}
	if none := g.PlayersDealtCombatDamageThisTurnBySubtype(uuid.Nil, "Faerie"); len(none) != 0 {
		t.Errorf("nobody controls nothing: %v", none)
	}
}

// Noncombat damage is not combat damage, and a changeling is every
// creature type — the enteredAllCreatureTypes fold EnteredSubtypes
// already has.
func TestCombatDamageTallyIsCombatOnlyAndFoldsChangelings(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	bolt := pushPermanent(g, Card{Name: "Pinger", TypeLine: "Artifact Creature — Construct", Power: 1, Toughness: 1, Owner: me, Controller: me})
	shifter := pushPermanent(g, Card{
		Name: "Shifter", TypeLine: "Creature — Shapeshifter", Keywords: []string{KeywordChangeling},
		Power: 1, Toughness: 1, Owner: me, Controller: me,
	})
	emit(g,
		Event{Kind: EventDealDamage, Actor: me, Source: bolt, Target: opp, Amount: 1},
		Event{Kind: EventDealDamage, Actor: me, Source: shifter, Target: opp, Amount: 1, Combat: true},
	)
	// By NAME, because the changeling below is a Construct too.
	if hit := g.PlayersDealtCombatDamageThisTurnByName(me, "Pinger"); len(hit) != 0 {
		t.Errorf("a ping is not combat damage: %v", hit)
	}
	if hit := g.PlayersDealtCombatDamageThisTurnBySubtype(me, "Faerie"); !hit[opp] {
		t.Error("a changeling is a Faerie for this too (CR 702.73a)")
	}
	if hit := g.PlayersDealtCombatDamageThisTurnBySubtype(me, "Swamp"); len(hit) != 0 {
		t.Errorf("a changeling has creature types, not land types: %v", hit)
	}
}

// The tally is turn-scoped like every other cell.
func TestCombatDamageTallyResetsAtTheTurnBoundary(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	faerie := pushPermanent(g, Card{Name: "Faerie", TypeLine: "Creature — Faerie", Power: 1, Toughness: 1, Owner: me, Controller: me})
	emit(g, Event{Kind: EventDealDamage, Actor: me, Source: faerie, Target: opp, Amount: 1, Combat: true})
	if hit := g.PlayersDealtCombatDamageThisTurnBySubtype(me, "Faerie"); !hit[opp] {
		t.Fatal("setup: the damage was not recorded")
	}
	seat := g.Turn.ActiveSeat
	for i := 0; i < 40 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if hit := g.PlayersDealtCombatDamageThisTurnBySubtype(me, "Faerie"); len(hit) != 0 {
		t.Errorf("last turn's combat damage is not this turn's: %v", hit)
	}
}

// Both new cells are CARRIED: they ride Clone / RestoreFrom and the
// JSON snapshot, like every other field of TurnTally (the drift test
// classifies Game.TurnTally; this is the deep copy and the wire).
func TestNewTallyCellsSurviveCloneAndTheSnapshotWire(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	faerie := pushPermanent(g, Card{Name: "Faerie", TypeLine: "Creature — Faerie", Power: 1, Toughness: 1, Owner: me, Controller: me})
	emit(g,
		Event{Kind: EventETB, CardID: faerie},
		Event{Kind: EventDealDamage, Actor: me, Source: faerie, Target: opp, Amount: 1, Combat: true},
	)

	snap := g.Clone()
	// The clone must not share the maps with the live game.
	other := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me, Controller: me})
	emit(g, Event{Kind: EventETB, CardID: other})
	if snap.EnteredThisTurn(other) {
		t.Error("the clone shares TurnTally.Entered with the live game")
	}
	if !snap.EnteredThisTurn(faerie) || !snap.PlayersDealtCombatDamageThisTurnBySubtype(me, "Faerie")[opp] {
		t.Error("the clone did not take the cells with it")
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	if !g.EnteredThisTurn(faerie) {
		t.Error("restore did not bring Entered back")
	}
	if g.EnteredThisTurn(other) {
		t.Error("restore did not roll Entered back")
	}
	if !g.PlayersDealtCombatDamageThisTurnBySubtype(me, "Faerie")[opp] {
		t.Error("restore did not bring CombatDamagedPlayers back")
	}

	// The wire: a UUID-keyed map has to marshal and unmarshal, which
	// is the one thing a new map type can get wrong silently.
	blob, err := json.Marshal(cloneTurnTally(g.TurnTally))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back TurnTally
	if err := json.Unmarshal(blob, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Entered[faerie] != g.TurnTally.Entered[faerie] || back.Entered[faerie] == 0 {
		t.Errorf("Entered did not survive JSON: %v", back.Entered)
	}
	if len(back.CombatDamagedPlayers) != len(g.TurnTally.CombatDamagedPlayers) || len(back.CombatDamagedPlayers) == 0 {
		t.Errorf("CombatDamagedPlayers did not survive JSON: %v", back.CombatDamagedPlayers)
	}
}
