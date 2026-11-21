package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beyeen Veil // Beyeen Coast — Modal DFC. Front: Instant {1}{U}:
// "Creatures your opponents control get -2/-0 until end of turn." Back:
// Land (mdfc_lands.go).
//
// The set is pinned at resolution (CR 611.2c): a creature an opponent
// plays later this turn is not shrunk. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b03de49d-246f-44e2-9487-9e4e43ec7be4",
		Name:         "Beyeen Veil",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.OpponentControls()),
				Power: -2,
				Label: "Beyeen Veil — creatures your opponents control get -2/-0",
			}.Apply(ctx)
		},
	})
}
