package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The front face puts an artifact creature card tapped and attacking,
// and converts at end of combat.
func TestUltraMagnusPutsAnAttackerInAndConvertsAtEndOfCombat(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, effectstest.UltraMagnusRow(), me)
	robot := effectstest.HandCardForTest(me, "Hired Robot", "Artifact Creature — Robot", "")
	effectstest.HandCardForTest(me, "Plain Bear", "Creature — Bear", "")

	effectstest.TransformerAttacks(t, g, id, opp.ID)
	effectstest.AnswerChooseCards(t, g, me.ID, robot)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // Magnus's own defender comes first
	effectstest.TfSettle(t, g)

	c := effectstest.FlipState(t, g, robot)
	if !c.Tapped || c.AttackingTarget != opp.ID {
		t.Fatalf("the put creature is tapped=%v attacking %v, want tapped and attacking the opponent", c.Tapped, c.AttackingTarget)
	}
	effectstest.WantFace(t, g, id, 0, "Ultra Magnus, Tactician")

	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.TfSettle(t, g)
	c = effectstest.WantFace(t, g, id, 1, "Ultra Magnus, Armored Carrier")
	if c.TurnOvers != 1 {
		t.Errorf("turn-overs = %d, want 1", c.TurnOvers)
	}
}

// Only artifact creature cards are offered: with none in hand there is
// no prompt at all.
func TestUltraMagnusOffersOnlyArtifactCreatures(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, effectstest.UltraMagnusRow(), me)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	effectstest.HandCardForTest(me, "Plain Bear", "Creature — Bear", "")
	effectstest.HandCardForTest(me, "Plain Rock", "Artifact", "")

	effectstest.TransformerAttacks(t, g, id, opp.ID)
	effectstest.TfSettle(t, g)
	if effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Fatal("offered a card although none in hand is an artifact creature")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.TfSettle(t, g)
	effectstest.WantFace(t, g, id, 0, "Ultra Magnus, Tactician")
}

// Declining the "may" puts nothing in, so there is no "if you do" and
// Magnus does not convert.
func TestUltraMagnusDecliningConvertsNothing(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, effectstest.UltraMagnusRow(), me)
	robot := effectstest.HandCardForTest(me, "Hired Robot", "Artifact Creature — Robot", "")
	effectstest.TransformerAttacks(t, g, id, opp.ID)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.TfSettle(t, g)
	if zone, _ := effectstest.CardWhere(g, me, robot); zone != "hand" {
		t.Errorf("declined, yet the creature is in %s", zone)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.TfSettle(t, g)
	effectstest.WantFace(t, g, id, 0, "Ultra Magnus, Tactician")
}

// CR 701.28e: a Magnus that turned over before end of combat ignores the
// delayed convert.
func TestUltraMagnusDelayedConvertIsIgnoredAfterAnotherTurnOver(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, effectstest.UltraMagnusRow(), me)
	robot := effectstest.HandCardForTest(me, "Hired Robot", "Artifact Creature — Robot", "")
	effectstest.TransformerAttacks(t, g, id, opp.ID)
	effectstest.AnswerChooseCards(t, g, me.ID, robot)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.TfSettle(t, g)

	g.WithWriteLock(func() {
		if err := g.ConvertPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.TfSettle(t, g)
	c := effectstest.FlipState(t, g, id)
	if c.TurnOvers != 1 || c.ActiveFace != 1 {
		t.Errorf("turn-overs %d face %d: the delayed convert should have been ignored", c.TurnOvers, c.ActiveFace)
	}
}

// Formidable: attacking creatures gain indestructible, and with total
// power 8 or more Magnus converts.
func TestUltraMagnusArmoredCarrierFormidable(t *testing.T) {
	for _, tc := range []struct {
		name        string
		otherPower  int
		wantConvert bool
	}{{"total 8", 4, true}, {"total 7", 3, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := effectstest.FlipMain(t)
			id := effectstest.TransformerConverted(t, g, effectstest.UltraMagnusRow(), me)
			other := effectstest.PushVanillaCreature(g, me.ID, "Ally", tc.otherPower, 4)

			effectstest.TransformerAttacks(t, g, id, opp.ID, other)
			effectstest.TfSettle(t, g)

			for _, who := range []uuid.UUID{id, other} {
				c := effectstest.FlipState(t, g, who)
				if !game.HasKeyword(&c, "indestructible") {
					t.Errorf("%s did not gain indestructible", c.Name)
				}
			}
			if tc.wantConvert {
				effectstest.WantFace(t, g, id, 0, "Ultra Magnus, Tactician")
			} else {
				effectstest.WantFace(t, g, id, 1, "Ultra Magnus, Armored Carrier")
			}
		})
	}
}

// The Vehicle has haste, so one converted this turn can attack.
func TestUltraMagnusArmoredCarrierHasHaste(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.ImportToBattlefield(t, g, effectstest.UltraMagnusRow(), me) // summoning sick
	g.WithWriteLock(func() {
		if err := g.ConvertPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
		g.RecomputeLayersIfStaleLocked()
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Errorf("the hasty Vehicle can't attack: %v", err)
	}
}
