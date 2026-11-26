package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nausea — Sorcery {1}{B} (Eternal Masters, S82 OLD, #5542):
//
//	"All creatures get -1/-1 until end of turn."
//
// Glistening Deluge's first sentence: the creatures on the battlefield as it
// resolves get -1/-1 until end of turn (CR 611.2c), a creature that arrives
// later does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2b8abae2-dd15-41ee-81ae-6a463723b43e",
		Name:         "Nausea",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{Match: effects.Creature(), Power: -1, Toughness: -1, Label: "Nausea — -1/-1"}.Apply(ctx)
		},
	})
}
