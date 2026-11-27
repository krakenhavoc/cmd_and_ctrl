package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oak Street Innkeeper — Creature — Elf {2}{G}, 1/2 (Return to Ravnica, #5566):
//
//	"During turns other than yours, tapped creatures you control have
//	 hexproof."
//
// Bedrock Tortoise's layer-6 hexproof grant, read live (CR 611.3a): it applies
// on every turn but its controller's and only to a creature that is tapped at
// that moment, itself included.
//
// No simplification.
func init() {
	hexproof := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsCreature() && target.Tapped &&
			!effects.IsActivePlayer(g, source.Controller)
	}, "hexproof")
	hexproof.Label = "During turns other than yours, tapped creatures you control have hexproof."
	effects.Register(effects.Spec{
		OracleID:     "1fa94925-04e5-4073-91af-1857dbc09991",
		Name:         "Oak Street Innkeeper",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{hexproof},
	})
}
