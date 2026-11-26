package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rollick of Abandon — Sorcery {3}{R}{R} (Journey into Nyx, S82 OLD, #5558):
//
//	"All creatures get +2/-2 until end of turn."
//
// Overrun's untargeted until-end-of-turn boost over every creature, yours and
// theirs, chosen once as the spell resolves (CR 611.2c): creatures that enter
// later are not affected. A creature whose toughness reaches 0 dies as a
// state-based action (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ecda721d-ec5d-484b-a834-f4de96083b08",
		Name:         "Rollick of Abandon",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{Match: effects.Creature(), Power: 2, Toughness: -2, Label: "Rollick of Abandon — +2/-2"}.Apply(ctx)
		},
	})
}
