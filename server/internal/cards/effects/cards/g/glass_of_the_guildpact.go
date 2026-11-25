package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glass of the Guildpact — Artifact {2} (Ravnica Allegiance):
//
//	"Multicolored creatures you control get +1/+1."
//
// A layer 7c anthem over the creatures you control that are two or more
// colors, judged on each layer pass, so a creature that gains or loses a
// color is counted from then on.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a967921c-cfd4-422c-a08b-42cdd80f89c0",
		Name:         "Glass of the Guildpact",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.B16Anthem(func(target *game.Card, g *game.Game, source *game.Card) bool {
			return effects.CreaturesYouControl(target, g, source) && effects.Multicolored()(g, source.Controller, *target)
		}, 1, 1)},
	})
}
