package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func ultraMagnusRow() cards.Card {
	return transformerRow(ultraMagnusOracleID,
		disturbFace{"Ultra Magnus, Tactician", "Legendary Artifact Creature — Robot", "{4}{R}{G}{W}", "7", "7"},
		disturbFace{"Ultra Magnus, Armored Carrier", "Legendary Artifact — Vehicle", "", "4", "7"},
		"More Than Meets the Eye {2}{R}{G}{W}\nWard {2}\nWhenever Ultra Magnus attacks, you may put an artifact creature card from your hand onto the battlefield tapped and attacking. If you do, convert Ultra Magnus at end of combat.",
		"Living metal\nHaste\nFormidable — Whenever Ultra Magnus attacks, attacking creatures you control gain indestructible until end of turn. If those creatures have total power 8 or greater, convert Ultra Magnus.",
		"Living metal", "Convert", "Ward", "Haste")
}

// The front face puts an artifact creature card tapped and attacking,
// and converts at end of combat.
func TestUltraMagnusPutsAnAttackerInAndConvertsAtEndOfCombat(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerIn(t, g, ultraMagnusRow(), me)
	robot := handCardForTest(me, "Hired Robot", "Artifact Creature — Robot", "")
	handCardForTest(me, "Plain Bear", "Creature — Bear", "")

	transformerAttacks(t, g, id, opp.ID)
	answerChooseCards(t, g, me.ID, robot)
	answerOptionPick(t, g, me.ID, 0) // Magnus's own defender comes first
	tfSettle(t, g)

	c := flipState(t, g, robot)
	if !c.Tapped || c.AttackingTarget != opp.ID {
		t.Fatalf("the put creature is tapped=%v attacking %v, want tapped and attacking the opponent", c.Tapped, c.AttackingTarget)
	}
	wantFace(t, g, id, 0, "Ultra Magnus, Tactician")

	advanceTo(t, g, game.StepEndCombat)
	tfSettle(t, g)
	c = wantFace(t, g, id, 1, "Ultra Magnus, Armored Carrier")
	if c.TurnOvers != 1 {
		t.Errorf("turn-overs = %d, want 1", c.TurnOvers)
	}
}

// Only artifact creature cards are offered: with none in hand there is
// no prompt at all.
func TestUltraMagnusOffersOnlyArtifactCreatures(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerIn(t, g, ultraMagnusRow(), me)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	handCardForTest(me, "Plain Bear", "Creature — Bear", "")
	handCardForTest(me, "Plain Rock", "Artifact", "")

	transformerAttacks(t, g, id, opp.ID)
	tfSettle(t, g)
	if latestChooseCards(g, me.ID) != nil {
		t.Fatal("offered a card although none in hand is an artifact creature")
	}
	advanceTo(t, g, game.StepEndCombat)
	tfSettle(t, g)
	wantFace(t, g, id, 0, "Ultra Magnus, Tactician")
}

// Declining the "may" puts nothing in, so there is no "if you do" and
// Magnus does not convert.
func TestUltraMagnusDecliningConvertsNothing(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerIn(t, g, ultraMagnusRow(), me)
	robot := handCardForTest(me, "Hired Robot", "Artifact Creature — Robot", "")
	transformerAttacks(t, g, id, opp.ID)
	answerChooseCards(t, g, me.ID)
	tfSettle(t, g)
	if zone, _ := cardWhere(g, me, robot); zone != "hand" {
		t.Errorf("declined, yet the creature is in %s", zone)
	}
	advanceTo(t, g, game.StepEndCombat)
	tfSettle(t, g)
	wantFace(t, g, id, 0, "Ultra Magnus, Tactician")
}

// CR 701.28e: a Magnus that turned over before end of combat ignores the
// delayed convert.
func TestUltraMagnusDelayedConvertIsIgnoredAfterAnotherTurnOver(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerIn(t, g, ultraMagnusRow(), me)
	robot := handCardForTest(me, "Hired Robot", "Artifact Creature — Robot", "")
	transformerAttacks(t, g, id, opp.ID)
	answerChooseCards(t, g, me.ID, robot)
	answerOptionPick(t, g, me.ID, 0)
	tfSettle(t, g)

	g.WithWriteLock(func() {
		if err := g.ConvertPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	advanceTo(t, g, game.StepEndCombat)
	tfSettle(t, g)
	c := flipState(t, g, id)
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
			g, me, opp := flipMain(t)
			id := transformerConverted(t, g, ultraMagnusRow(), me)
			other := pushVanillaCreature(g, me.ID, "Ally", tc.otherPower, 4)

			transformerAttacks(t, g, id, opp.ID, other)
			tfSettle(t, g)

			for _, who := range []uuid.UUID{id, other} {
				c := flipState(t, g, who)
				if !game.HasKeyword(&c, "indestructible") {
					t.Errorf("%s did not gain indestructible", c.Name)
				}
			}
			if tc.wantConvert {
				wantFace(t, g, id, 0, "Ultra Magnus, Tactician")
			} else {
				wantFace(t, g, id, 1, "Ultra Magnus, Armored Carrier")
			}
		})
	}
}

// The Vehicle has haste, so one converted this turn can attack.
func TestUltraMagnusArmoredCarrierHasHaste(t *testing.T) {
	g, me, opp := flipMain(t)
	id := importToBattlefield(t, g, ultraMagnusRow(), me) // summoning sick
	g.WithWriteLock(func() {
		if err := g.ConvertPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
		g.RecomputeLayersIfStaleLocked()
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Errorf("the hasty Vehicle can't attack: %v", err)
	}
}
