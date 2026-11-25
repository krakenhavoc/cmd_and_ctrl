package game

import (
	"testing"

	"github.com/google/uuid"
)

// could_attack_test.go — #6211, Season of the Witch's "except for
// creatures that couldn't attack": the turn remembers every creature that
// could have been declared as an attacker (CR 508.1a, 508.1c), whether or
// not it was.

func couldAttack(g *Game, id uuid.UUID) bool {
	var out bool
	g.WithWriteLock(func() { out = g.CouldHaveAttackedThisTurnForEffect(id) })
	return out
}

func setCardField(g *Game, id uuid.UUID, f func(c *Card)) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				f(&g.Battlefield.Cards[i])
			}
		}
		g.layerVersion.Add(1)
	})
}

// A summoning-sick creature, a tapped one and one under a "can't attack"
// restriction couldn't attack; an untapped creature that stayed home and
// one that attacked could; the defending player's creatures couldn't.
func TestCouldAttackRecordsTheLegalAttackersOfTheDeclaration(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushCombatant(t, g, me, "Attacker", 2, 2)
	home := pushCombatant(t, g, me, "Stayed Home", 2, 2)
	sick := pushCombatant(t, g, me, "Sick", 2, 2)
	hasty := pushCombatant(t, g, me, "Hasty", 2, 2, "haste")
	tapped := pushCombatant(t, g, me, "Tapped", 2, 2)
	theirs := pushCombatant(t, g, opp, "Theirs", 2, 2)
	setCardField(g, sick, func(c *Card) { c.SummonedThisTurn = true })
	setCardField(g, hasty, func(c *Card) { c.SummonedThisTurn = true })
	setCardField(g, tapped, func(c *Card) { c.Tapped = true })

	if couldAttack(g, home) {
		t.Fatal("the record says a creature could attack before any declaration")
	}
	declareAttacks(t, g, attacker)

	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want bool
	}{
		{"the attacker", attacker, true},
		{"an untapped creature that stayed home", home, true},
		{"a summoning-sick creature", sick, false},
		{"a summoning-sick creature with haste", hasty, true},
		{"a tapped creature", tapped, false},
		{"the defending player's creature", theirs, false},
	} {
		if got := couldAttack(g, tc.id); got != tc.want {
			t.Errorf("%s: could attack = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// A creature that leaves and comes back is a new object that could not
// attack (CR 400.7), and the record ends with the turn.
func TestCouldAttackIsPerObjectAndPerTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	home := pushCombatant(t, g, me, "Stayed Home", 2, 2)
	blinked := pushCombatant(t, g, me, "Blinked", 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	advanceIntoStep(t, g, StepPostcombatMain)
	if !couldAttack(g, home) || !couldAttack(g, blinked) {
		t.Fatal("setup: the untapped creatures are not recorded")
	}
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, blinked); err != nil {
			t.Fatal(err)
		}
		if _, err := MoveCard(me.Graveyard, g.Battlefield, blinked); err != nil {
			t.Fatal(err)
		}
	})
	if couldAttack(g, blinked) {
		t.Error("a creature that left and came back still reads as able to attack")
	}
	advanceOneTurn(t, g)
	if couldAttack(g, home) {
		t.Error("the record outlived the turn")
	}
}
