package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eyeblight Massacre — Sorcery {2}{B}{B} (Kaldheim Commander, #5506):
//
//	"Non-Elf creatures get -2/-2 until end of turn."
//
// Every creature that is not an Elf as the spell resolves, yours and
// the opponents', gets -2/-2 (layer 7c); the set locks as the effect begins
// (CR 611.2c), so a creature that enters later this turn is not shrunk. A
// changeling is an Elf and is spared (Crippling Fear's reading of the
// effective subtypes).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b90dd70-798e-4dc4-99e7-b36d6823af2b",
		Name:         "Eyeblight Massacre",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{Match: effects.NotOfCreatureType(ctx.TextSubtype("Elf")), Power: -2, Toughness: -2, Label: "Eyeblight Massacre — -2/-2"}.Apply(ctx)
		},
	})
}
