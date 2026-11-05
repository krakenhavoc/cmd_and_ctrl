package protocol

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_different_names_view_test.go — #4896. "Discard three cards
// with different names" ships the public different_names flag beside the
// count, and the hand cards that could be one of the set (a nameless card
// never is, CR 201.2b) to the controller alone (#1369).

func TestDifferentNamesDiscardViewStampsTheRuleAndTheNamedCards(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := uuid.New()
	var want []string
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		g.Battlefield.PushTop(game.Card{
			InstanceID: src, Name: "Different Names Source", TypeLine: "Creature — Sphinx",
			OracleID: "00000000-0000-0000-0000-000000004896",
			Owner:    me.ID, Controller: me.ID, Power: 5, Toughness: 5,
			ActivatedAbilities: []game.ActivatedAbilityShape{{
				Label: "Discard three cards with different names: Draw five cards.",
				Cost: game.AbilityCost{DiscardCards: &game.DiscardCost{
					N: 3, Label: "three cards with different names", DifferentNames: true,
				}},
			}},
		})
		a := handCardFor(me, "Opt", "Instant")
		b := handCardFor(me, "Opt", "Instant")
		c := handCardFor(me, "Shock", "Instant")
		handCardFor(me, "", "Instant")
		want = sortedIDs(a, b, c)
	})
	g.BumpLayerVersionForTest()

	mine := controllerFrameCard(t, g, src).ActivatedAbilities[0]
	if !mine.DiscardCostDifferentNames || mine.DiscardCostN != 3 {
		t.Errorf("controller row: different_names=%v n=%d", mine.DiscardCostDifferentNames, mine.DiscardCostN)
	}
	if got := sortedStrings(mine.DiscardCostOptions); !reflect.DeepEqual(got, want) {
		t.Errorf("controller options = %v, want the named cards %v", got, want)
	}

	theirs := frameCard(t, g, opp.ID.String(), src).ActivatedAbilities[0]
	if !theirs.DiscardCostDifferentNames {
		t.Error("the opponent's row lost the flag; it is the printed rule")
	}
	if len(theirs.DiscardCostOptions) != 0 {
		t.Errorf("the opponent's row carries the controller's hand: %v", theirs.DiscardCostOptions)
	}
}
