package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// special_action_charged_cost_view_test.go — #1319: a CR 116.2
// special-action row carries ChargedCost, the printed Cost run
// through the CR 601.2f cost-modifier pass the engine actually
// charges (SpecialActionManaCostForEffect). Mirrors
// charged_mana_cost_view_test.go for the special-action half; Ranar
// the Ever-Watchful's foretell discount is the fixture.

// installSpecialAction hooks game.CatalogSpecialActions for the
// length of the test — installChargedCostModifier's twin for the
// other catalog hook this file needs.
func installSpecialAction(t *testing.T, oracle string, actions ...game.SpecialAction) {
	t.Helper()
	prev := game.CatalogSpecialActions
	game.CatalogSpecialActions = func(id string) []game.SpecialAction {
		if id == oracle {
			return actions
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogSpecialActions = prev })
}

func TestForetellRowShowsTheChargedCostAndThePrintedOneWhenTheyDiffer(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]

	const cardOracle = "protocol-test-foretell-card"
	installSpecialAction(t, cardOracle, game.SpecialAction{
		Kind: game.SpecialActionForetell, Cost: game.ForetellExileCost, CastCost: "{1}{U}", Label: "Foretell {2}",
	})
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Saw It Coming",
		TypeLine:   "Instant",
		OracleID:   cardOracle,
		Owner:      me.ID,
		Controller: me.ID,
		KnownBy:    map[uuid.UUID]bool{me.ID: true},
	})

	// Undiscounted: ChargedCost equals the printed Cost.
	row := handCardView(t, FilterViewFor(ViewOfGame(g), me.ID.String()), 0, id).SpecialActions[0]
	if row.Cost != "{2}" {
		t.Fatalf("printed Cost = %q, want {2}", row.Cost)
	}
	if got := mustChargedCost(t, row.ChargedCost); got != row.Cost {
		t.Errorf("with no discounter on the board, ChargedCost = %q, want it to equal the printed cost %q",
			got, row.Cost)
	}

	const sourceOracle = "protocol-test-foretell-discounter"
	installChargedCostModifier(t, sourceOracle, game.CostModifier{
		Kind:           game.CostReduction,
		SpecialActions: true,
		Label:          "The first card you foretell each turn costs {0} to foretell.",
		AppliesTo: func(q game.CostQuery) bool {
			return q.SpecialAction != nil && q.SpecialAction.Kind == game.SpecialActionForetell &&
				q.Game.ForetoldCountThisTurn(q.Controller) == 0
		},
		Amount: func(game.CostQuery) int { return 2 },
	})
	discounter := game.NewCard("Ranar the Ever-Watchful", me.ID)
	discounter.TypeLine = "Legendary Creature — Spirit Warrior"
	discounter.OracleID = sourceOracle
	g.Battlefield.PushTop(discounter)

	row = handCardView(t, FilterViewFor(ViewOfGame(g), me.ID.String()), 0, id).SpecialActions[0]
	if row.Cost != "{2}" {
		t.Errorf("the discount must not touch the printed field: Cost = %q, want {2}", row.Cost)
	}
	got := mustChargedCost(t, row.ChargedCost)
	if got != "" {
		t.Errorf("ChargedCost = %q, want \"\" — the first foretell this turn is free", got)
	}
	if got == row.Cost {
		t.Errorf("ChargedCost and Cost read the same (%q) — the discount did not reach the row", row.Cost)
	}
}
