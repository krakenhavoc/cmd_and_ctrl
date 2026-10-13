package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func blasterRow() cards.Card {
	return transformerRow(blasterOracleID,
		disturbFace{"Blaster, Combat DJ", "Legendary Artifact Creature — Robot", "{3}{R}{G}", "3", "3"},
		disturbFace{"Blaster, Morale Booster", "Legendary Artifact", "", "", ""},
		"More Than Meets the Eye {1}{R}{G}\nOther nontoken artifact creatures and Vehicles you control have modular 1. (They enter with an additional +1/+1 counter on them. When they die, you may put their +1/+1 counters on target artifact creature.)\nWhenever you put one or more +1/+1 counters on Blaster, convert it.",
		"Modular 3\n{X}, {T}: Move X +1/+1 counters from Blaster onto another target artifact. That artifact gains haste until end of turn. If Blaster has no +1/+1 counters on it, convert it. Activate only as a sorcery.",
		"Modular")
}

func hasModularOne(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c := flipState(t, g, id)
	return game.HasKeyword(&c, "modular 1")
}

// The static grants modular 1 to the controller's OTHER nontoken artifact
// creatures and Vehicles, and to nothing else.
func TestBlasterGrantsModularToOtherNontokenArtifactCreaturesAndVehicles(t *testing.T) {
	g, me, opp := flipMain(t)
	blaster := transformerIn(t, g, blasterRow(), me)
	golem := pushCatalogPermanent(g, me.ID, "Golem", "Artifact Creature — Golem", "", false)
	vehicle := pushCatalogPermanent(g, me.ID, "Cart", "Artifact — Vehicle", "", false)
	servo := pushCatalogPermanent(g, me.ID, "Servo", "Token Artifact Creature — Servo", "", false)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "", false)

	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want bool
	}{
		{"nontoken artifact creature", golem, true}, {"Vehicle", vehicle, true},
		{"token", servo, false}, {"nonartifact creature", bear, false},
		{"noncreature, non-Vehicle artifact", rock, false}, {"an opponent's", theirs, false},
		{"Blaster itself", blaster, false},
	} {
		if got := hasModularOne(t, g, tc.id); got != tc.want {
			t.Errorf("%s: modular 1 = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// A granted modular is a real modular: the permanent enters with the
// extra counter.
func TestBlasterGivesEnteringArtifactCreaturesACounter(t *testing.T) {
	g, me, _ := flipMain(t)
	transformerIn(t, g, blasterRow(), me)
	id := handCardForTest(me, "Golem", "Artifact Creature — Golem", "")
	var entered uuid.UUID
	g.WithWriteLock(func() {
		var err error
		entered, err = g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{})
		if err != nil {
			t.Fatal(err)
		}
	})
	if n := countersOn(g, entered, game.CounterPlusOne); n != 1 {
		t.Errorf("the entering Golem has %d +1/+1 counters, want 1 from modular 1", n)
	}
}

// Whenever you put +1/+1 counters on Blaster, convert it.
func TestBlasterConvertsWhenYouPutCountersOnIt(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerIn(t, g, blasterRow(), me)
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(me.ID, id, game.CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
	})
	tfSettle(t, g)
	wantFace(t, g, id, 1, "Blaster, Morale Booster")
}

// Counters another player puts on it, or other counters, do not.
func TestBlasterIgnoresOtherCounters(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerIn(t, g, blasterRow(), me)
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(opp.ID, id, game.CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
		if err := g.AddCounterByForEffect(me.ID, id, "charge", 1); err != nil {
			t.Fatal(err)
		}
	})
	tfSettle(t, g)
	wantFace(t, g, id, 0, "Blaster, Combat DJ")
}

// Cast converted, Blaster enters as the noncreature back face with its
// own Modular 3.
func TestBlasterCastConvertedEntersWithThreeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := importToHand(blasterRow(), me)
	advanceToMain(t, g)
	me.ManaPool = nil
	b06AddMana(me, "C", "R", "G")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: mtmte, Strict: true}); err != nil {
		t.Fatalf("cast converted: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := flipState(t, g, id)
	if c.ActiveFace != 1 || c.Name != "Blaster, Morale Booster" || c.IsCreature() {
		t.Fatalf("entered as face %d %q creature=%v, want the noncreature back face", c.ActiveFace, c.Name, c.IsCreature())
	}
	if c.Counters[game.CounterPlusOne] != 3 {
		t.Errorf("%d +1/+1 counters, want 3 from Modular 3", c.Counters[game.CounterPlusOne])
	}
}

// Morale Booster: move X counters onto another artifact, which gains
// haste; Blaster converts only when none are left.
func TestBlasterMoraleBoosterMovesCountersAndConvertsWhenEmpty(t *testing.T) {
	for _, tc := range []struct {
		name        string
		x           int
		wantMoved   int
		wantConvert bool
	}{{"leaves one", 2, 2, false}, {"moves them all", 3, 3, true}, {"more than it has", 5, 3, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := flipMain(t)
			id := transformerConverted(t, g, blasterRow(), me)
			pfAddCounters(g, id, "+1/+1", 3)
			rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
			for i := 0; i < tc.x; i++ {
				b06AddMana(me, "C")
			}
			p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{
				XValue: tc.x, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
			})
			if n := countersOn(g, rock, game.CounterPlusOne); n != tc.wantMoved {
				t.Errorf("the target holds %d counters, want %d", n, tc.wantMoved)
			}
			r := flipState(t, g, rock)
			if !game.HasKeyword(&r, "haste") {
				t.Error("the target did not gain haste")
			}
			if tc.wantConvert {
				wantFace(t, g, id, 0, "Blaster, Combat DJ")
			} else {
				wantFace(t, g, id, 1, "Blaster, Morale Booster")
			}
		})
	}
}

// It is a sorcery-speed ability and cannot target itself.
func TestBlasterMoraleBoosterCannotTargetItself(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerConverted(t, g, blasterRow(), me)
	pfAddCounters(g, id, "+1/+1", 1)
	b06AddMana(me, "C")
	err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		XValue: 1, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}},
	})
	if err == nil {
		t.Fatal("Blaster moved counters onto itself")
	}
}
