package cardsco

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Countervailing Winds — Instant {2}{U} (Hour of Devastation, #5534):
//
//	"Counter target spell unless its controller pays {1} for each card in
//	 your graveyard.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Circular Logic's counter: the graveyard is counted as the spell resolves,
// when this spell is still on the stack, and its controller is asked to
// pay that much. The shared Cycling row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c51e20b-07b2-4c96-8d73-87379aafbc7c",
		Name:         "Countervailing Winds",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			cost := "{" + strconv.Itoa(effects.B31GraveyardSize(ctx.Game, ctx.Controller())) + "}"
			return effects.CounterUnlessPaid{
				StackID:  t.ID,
				Cost:     cost,
				Question: "Countervailing Winds — pay " + cost + " or your spell is countered",
			}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
