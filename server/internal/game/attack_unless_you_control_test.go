package game

import (
	"slices"
	"testing"
)

// attack_unless_you_control_test.go pins AttackTargetRestriction.ControllerMustControl
// and PermanentQuery.WithExactPT (#5948): Lovestruck Beast's "This creature can't
// attack unless you control a 1/1 creature" (CR 508.1c). It is a fact about the
// attacker's controller, so it refuses every target while unmet.

const lovestruckOracle = "test-lovestruck-beast"

func withControlsOneOneStatic(t *testing.T) {
	t.Helper()
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != lovestruckOracle {
			return nil
		}
		return []StaticAbility{{
			Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.InstanceID == source.InstanceID && target.IsCreature()
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, source *Card) {
				c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, AttackTargetRestriction{
					Source: source.InstanceID, SourceName: source.Name,
					ControllerMustControl: []PermanentQuery{PermanentQuery{Types: []string{"creature"}}.WithExactPT(1, 1)},
				})
			},
		}}
	})
}

func TestCantAttackUnlessYouControlAOneOneCreature(t *testing.T) {
	withControlsOneOneStatic(t)
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := pushOwnedByAnother(g, me.ID, me.ID, "Lovestruck Beast", lovestruckOracle)
	pushTypedTestCard(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	pushTypedTestCard(g, Card{Name: "Their Human", TypeLine: "Creature — Human", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	advanceIntoStep(t, g, StepDeclareAttackers)

	if got := attackTargetsOf(g, beast); len(got) != 0 {
		t.Fatalf("targets = %v with no 1/1 of mine, want none", got)
	}
	re := restrictionErr(t, g.Clone().DeclareAttacker(beast, opp.ID))
	if got, want := re.Sentence(), "Lovestruck Beast can't attack P2: P1 controls no 1/1 creature."; got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}

	one := pushTypedTestCard(g, Card{Name: "Human", TypeLine: "Creature — Human", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if got := attackTargetsOf(g, beast); !slices.Contains(got, opp.ID) {
		t.Fatalf("targets = %v with a 1/1 of mine, want the opponents", got)
	}
	// A counter makes it a 2/2: no longer a 1/1 (CR 122.1a).
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(one, CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
	})
	if got := attackTargetsOf(g, beast); len(got) != 0 {
		t.Errorf("targets = %v once the Human is a 2/2, want none", got)
	}
}
