package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Selective Snare — Sorcery {X}{U} (Guilds of Ravnica, #5525):
//
//	"Return X target creatures of the creature type of your choice to their
//	 owner's hand."
//
// The #6245 seam: the caster names a creature type as the spell is cast
// (CR 601.2c), and every one of the X targets must be a creature of that type,
// at announce and again as the spell resolves (CR 608.2b). Each target still
// legal goes home; a target that stopped being of the type is left alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                  "85442273-9596-47bf-9b40-6d48fd7183a6",
		Name:                      "Selective Snare",
		Completeness:              effects.CompletenessFull,
		ChoosesCreatureTypeAsCast: true,
		Targets:                   effects.TargetsCountedByX(effects.OfTheChosenType(effects.TargetCreature("X target creatures of the creature type of your choice"))),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.NeoB20BounceTargetedPermanents(ctx)
		},
	})
}
