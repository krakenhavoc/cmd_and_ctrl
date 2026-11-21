package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vinebred Brawler — Creature — Elf Berserker {2}{G} (Lorwyn Eclipsed, #4458):
//
//	"This creature must be blocked if able.
//	 Whenever this creature attacks, another target Elf you control gets
//	 +2/+1 until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5fffd840-1e6e-4f5d-b847-802f9a25eb5c",
		Name:         "Vinebred Brawler",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.MustBeBlocked()},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Vinebred Brawler — another target Elf you control gets +2/+1 until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						return effects.VsPumpTheTarget(effects.NewContext(g, item), 2, 1, "Vinebred Brawler — +2/+1 until end of turn")
					}),
				effects.Another(effects.TargetCreature("another target Elf you control", effects.YouControl(), effects.HasSubtype("Elf")))),
		},
	})
}
