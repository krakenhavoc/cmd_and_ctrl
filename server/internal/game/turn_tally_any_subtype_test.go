package game

import "testing"

// turn_tally_any_subtype_test.go — #3786 (Estinien Varlineau): the
// any-controller subtype read and the ref-keyed object read.

// "A Dragon" is whoever's Dragon it is: the subtype read over every
// controller, where the per-controller read answers only "a Dragon you
// control".
func TestCombatDamageBySubtypeUnionsEveryController(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp, third, fourth := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID, g.Seats[3].ID
	mine := pushPermanent(g, Card{Name: "My Dragon", TypeLine: "Creature — Dragon", Power: 4, Toughness: 4, Owner: me, Controller: me})
	theirs := pushPermanent(g, Card{Name: "Their Dragon", TypeLine: "Creature — Dragon", Power: 4, Toughness: 4, Owner: opp, Controller: opp})
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me, Controller: me})
	emit(g,
		Event{Kind: EventDealDamage, Actor: me, Source: mine, Target: opp, Amount: 4, Combat: true},
		Event{Kind: EventDealDamage, Actor: opp, Source: theirs, Target: third, Amount: 4, Combat: true},
		Event{Kind: EventDealDamage, Actor: me, Source: bear, Target: fourth, Amount: 2, Combat: true},
	)
	g.WithWriteLock(func() {
		anyDragon := g.PlayersDealtCombatDamageThisTurnByAnySubtype("Dragon")
		if len(anyDragon) != 2 || !anyDragon[opp] || !anyDragon[third] {
			t.Errorf("any-controller Dragon set = %v, want the second and third seats", anyDragon)
		}
		if mineOnly := g.PlayersDealtCombatDamageThisTurnBySubtype(me, "Dragon"); len(mineOnly) != 1 || !mineOnly[opp] {
			t.Errorf("the per-controller read must stay narrow: %v", mineOnly)
		}
		if none := g.PlayersDealtCombatDamageThisTurnByAnySubtype("Goblin"); len(none) != 0 {
			t.Errorf("no Goblin hit anyone: %v", none)
		}
		if none := g.PlayersDealtCombatDamageThisTurnByAnySubtype(""); len(none) != 0 {
			t.Errorf("an empty subtype names nothing: %v", none)
		}
	})
}

// The ref read answers for an object that has left the battlefield and
// refuses a new object under the same card ID (CR 400.7, CR 608.2h).
func TestObjectRefDealtCombatDamageSurvivesTheDealerAndRefusesAStranger(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me, Controller: me})
	var ref ObjectRef
	g.WithWriteLock(func() { ref = ObjectRef{ID: bear, Epoch: g.findCardByIDLocked(bear).ObjectEpoch} })
	emit(g, Event{Kind: EventDealDamage, Actor: me, Source: bear, Target: opp, Amount: 2, Combat: true})
	removeFromBattlefield(g, bear)
	g.WithWriteLock(func() {
		if !g.ObjectRefDealtCombatDamageToPlayerThisTurn(ref, opp) {
			t.Error("the object that dealt the damage is gone but still dealt it")
		}
		if g.ObjectRefDealtCombatDamageToPlayerThisTurn(ObjectRef{ID: bear, Epoch: ref.Epoch + 1}, opp) {
			t.Error("a different object under the same card ID did not deal it")
		}
		if g.ObjectRefDealtCombatDamageToPlayerThisTurn(ref, g.Seats[2].ID) {
			t.Error("the record is per victim")
		}
		if g.ObjectRefDealtCombatDamageToPlayerThisTurn(ObjectRef{}, opp) {
			t.Error("the zero ref names nothing")
		}
	})
}
