package protocol

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_x_view_test.go — #2527. "Discard X cards" ships no count (the
// count is the X the client is about to announce), the count_from_x
// flag, demands_x so the activation flow knows an X is owed, and the
// hand cards that could pay — to the controller alone (#1369).

func seatDiscardXSource(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Discard X Source", TypeLine: "Creature — Praetor",
		OracleID: "00000000-0000-0000-0000-000000002527",
		Owner:    owner, Controller: owner, Power: 3, Toughness: 3,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "{4}{B}{B}{B}, Discard X cards: do a thing",
			Cost: game.AbilityCost{
				Mana:         "{4}{B}{B}{B}",
				DiscardCards: &game.DiscardCost{Label: "X cards", CountFromX: true},
			},
		}},
	})
	return id
}

func TestDiscardXViewStampsTheFlagAndTheHandToTheControllerAlone(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var src uuid.UUID
	var hand []string
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		src = seatDiscardXSource(g, me.ID)
		hand = sortedIDs(
			handCardFor(me, "Hidden Bear", "Creature — Bear"),
			handCardFor(me, "Hidden Bolt", "Instant"),
		)
	})
	g.BumpLayerVersionForTest()

	mine := controllerFrameCard(t, g, src).ActivatedAbilities[0]
	if !mine.DiscardCostCountFromX || mine.DiscardCostN != 0 || mine.DiscardCostLabel != "X cards" {
		t.Errorf("controller row: count_from_x=%v n=%d label=%q, want true / 0 / \"X cards\"",
			mine.DiscardCostCountFromX, mine.DiscardCostN, mine.DiscardCostLabel)
	}
	if !mine.DemandsX {
		t.Error("demands_x is not set, so the client would never learn an X is owed")
	}
	if got := sortedStrings(mine.DiscardCostOptions); !reflect.DeepEqual(got, hand) {
		t.Errorf("controller options = %v, want the whole hand %v", got, hand)
	}

	theirs := frameCard(t, g, opp.ID.String(), src).ActivatedAbilities[0]
	if !theirs.DiscardCostCountFromX || !theirs.DemandsX {
		t.Error("the opponent's row lost the flag; it names nothing hidden")
	}
	if len(theirs.DiscardCostOptions) != 0 {
		t.Errorf("the opponent's row carries the controller's hand: %v", theirs.DiscardCostOptions)
	}
}

func sortedStrings(in []string) []string {
	out := append([]string(nil), in...)
	return sortedIDsFromStrings(out)
}

func sortedIDsFromStrings(s []string) []string {
	ids := make([]uuid.UUID, 0, len(s))
	for _, v := range s {
		ids = append(ids, uuid.MustParse(v))
	}
	return sortedIDs(ids...)
}
