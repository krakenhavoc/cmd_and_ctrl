package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0109 §7 (#1902): the view stamps each new cost component — the
// random discard's flag (and no options: there is nothing to pick), the
// library exile's count, and the hand-to-library put's count, label and
// the activator's hand.
func TestViewStampsLibraryAndRandomCosts(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	var src uuid.UUID
	var hand []string
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		src = uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: src, Name: "Cost Rows", TypeLine: "Enchantment",
			OracleID: "00000000-0000-0000-0000-000000001903",
			Owner:    me.ID, Controller: me.ID,
			ActivatedAbilities: []game.ActivatedAbilityShape{
				{Label: "random", Cost: game.AbilityCost{DiscardCards: &game.DiscardCost{N: 1, Label: "a card at random", Random: true}}},
				{Label: "library", Cost: game.AbilityCost{ExileFromLibraryTop: 4}},
				{Label: "top", Cost: game.AbilityCost{PutFromHandOnLibraryTop: 1}},
			},
		})
		hand = sortedIDs(handCardFor(me, "Hidden Bear", "Creature — Bear"), handCardFor(me, "Hidden Bolt", "Instant"))
	})
	g.BumpLayerVersionForTest()

	rows := frameCard(t, g, me.ID.String(), src).ActivatedAbilities
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(rows))
	}
	if r := rows[0]; !r.DiscardCostRandom || r.DiscardCostN != 1 || r.DiscardCostLabel != "a card at random" || len(r.DiscardCostOptions) != 0 {
		t.Errorf("random row = %+v, want the flag, the count and the label, and no options", r)
	}
	if r := rows[1]; r.LibraryExileCostN != 4 {
		t.Errorf("library row exile count = %d, want 4", r.LibraryExileCostN)
	}
	if r := rows[2]; r.TopCostN != 1 || r.TopCostLabel == "" || !sameStrings(sortedCopy(r.TopCostOptions), hand) {
		t.Errorf("top row = n %d label %q options %v, want 1, the clause and the hand %v", r.TopCostN, r.TopCostLabel, r.TopCostOptions, hand)
	}
}
