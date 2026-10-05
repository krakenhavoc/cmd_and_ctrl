package game

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// turn_tally_damage_dealers_test.go — #2149. TurnTally.DamageDealers
// records which creature OBJECTS dealt damage to which player this
// turn, for "a creature that dealt combat damage to you this turn".

func TestDamageDealersRecordsTheAttackerThatConnected(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp, third := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp, Controller: opp})
	idle := pushPermanent(g, Card{Name: "Idle", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp, Controller: opp})
	emit(g, Event{Kind: EventDealDamage, Actor: opp, Source: bear, Target: me, Amount: 2, Combat: true})

	g.WithWriteLock(func() {
		if got := g.CreaturesThatDealtCombatDamageToThisTurn(me); len(got) != 1 || got[0].ID != bear {
			t.Errorf("combat dealers to me = %v, want [bear]", got)
		}
		if !g.ObjectDealtCombatDamageToPlayerThisTurn(bear, me) {
			t.Error("the bear that connected is not recorded")
		}
		if g.ObjectDealtCombatDamageToPlayerThisTurn(idle, me) {
			t.Error("a creature that dealt nothing is recorded")
		}
		if g.ObjectDealtCombatDamageToPlayerThisTurn(bear, third) {
			t.Error("the record is per victim")
		}
		if got := g.CreaturesThatDealtCombatDamageToThisTurn(third); len(got) != 0 {
			t.Errorf("nobody hit the third seat: %v", got)
		}
	})
}

func TestDamageDealersSeparatesCombatFromOtherDamageAndIgnoresNonCreatures(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	pinger := pushPermanent(g, Card{Name: "Pinger", TypeLine: "Creature — Construct", Power: 1, Toughness: 1, Owner: opp, Controller: opp})
	rock := pushPermanent(g, Card{Name: "Rock", TypeLine: "Artifact", Owner: opp, Controller: opp})
	emit(g,
		Event{Kind: EventDealDamage, Actor: opp, Source: pinger, Target: me, Amount: 1},
		Event{Kind: EventDealDamage, Actor: opp, Source: rock, Target: me, Amount: 1},
	)
	g.WithWriteLock(func() {
		if g.ObjectDealtCombatDamageToPlayerThisTurn(pinger, me) {
			t.Error("a ping is not combat damage")
		}
		if !g.ObjectDealtDamageToPlayerThisTurn(pinger, me) {
			t.Error("a creature's ping is damage it dealt")
		}
		if g.ObjectDealtDamageToPlayerThisTurn(rock, me) {
			t.Error("a noncreature source is not a creature that dealt damage")
		}
	})
}

// CR 400.7: a creature that leaves and returns is a new object and did
// not deal the damage.
func TestDamageDealersDoesNotCountAFlickeredCreature(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp, Controller: opp})
	emit(g, Event{Kind: EventDealDamage, Actor: opp, Source: bear, Target: me, Amount: 2, Combat: true})
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(bear); err != nil {
			t.Fatal(err)
		}
		back, err := g.ReturnFromExileToBattlefieldForEffect(bear, uuid.Nil, false)
		if err != nil {
			t.Fatal(err)
		}
		if g.ObjectDealtCombatDamageToPlayerThisTurn(bear, me) || g.ObjectDealtCombatDamageToPlayerThisTurn(back, me) {
			t.Error("a flickered creature is a new object and did not deal the damage")
		}
		if got := g.CreaturesThatDealtCombatDamageToThisTurn(me); len(got) != 1 {
			t.Errorf("the record of the old object stays: %v", got)
		}
	})
}

func TestDamageDealersRecordOutlivesTheCreatureAndResetsNextTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp, Controller: opp})
	emit(g, Event{Kind: EventDealDamage, Actor: opp, Source: bear, Target: me, Amount: 2, Combat: true})
	removeFromBattlefield(g, bear)
	g.WithWriteLock(func() {
		if got := g.CreaturesThatDealtCombatDamageToThisTurn(me); len(got) != 1 {
			t.Errorf("a dealer that died is still on the record: %v", got)
		}
		if g.ObjectDealtCombatDamageToPlayerThisTurn(bear, me) {
			t.Error("a card that is nowhere is not an object that can be picked")
		}
	})
	seat := g.Turn.ActiveSeat
	for i := 0; i < 40 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	g.WithWriteLock(func() {
		if got := g.CreaturesThatDealtDamageToThisTurn(me); len(got) != 0 {
			t.Errorf("last turn's damage is not this turn's: %v", got)
		}
	})
}

func TestDamageDealersSurviveCloneRestoreAndTheSnapshotWire(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp, Controller: opp})
	emit(g, Event{Kind: EventDealDamage, Actor: opp, Source: bear, Target: me, Amount: 2, Combat: true})

	snap := g.Clone()
	other := pushPermanent(g, Card{Name: "Ox", TypeLine: "Creature — Ox", Power: 2, Toughness: 2, Owner: opp, Controller: opp})
	emit(g, Event{Kind: EventDealDamage, Actor: opp, Source: other, Target: me, Amount: 2, Combat: true})
	snap.WithWriteLock(func() {
		if got := snap.CreaturesThatDealtCombatDamageToThisTurn(me); len(got) != 1 {
			t.Errorf("the clone shares its record with the live game: %v", got)
		}
	})
	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	g.WithWriteLock(func() {
		if got := g.CreaturesThatDealtCombatDamageToThisTurn(me); len(got) != 1 || got[0].ID != bear {
			t.Errorf("restore did not bring the record back: %v", got)
		}
	})

	raw, err := json.Marshal(g.TurnTally)
	if err != nil {
		t.Fatal(err)
	}
	var back TurnTally
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	if len(back.DamageDealers) != 1 || back.DamageDealers[0].Source != bear {
		t.Errorf("the wire lost the record: %s", raw)
	}
	var empty TurnTally
	raw, _ = json.Marshal(empty)
	if strings.Contains(string(raw), "damageDealers") {
		t.Errorf("an empty tally must stay additive (omitempty): %s", raw)
	}
}
