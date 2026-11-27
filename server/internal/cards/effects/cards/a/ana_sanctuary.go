package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ana Sanctuary — Enchantment {2}{G} (Apocalypse, S82 OLD, #5614):
//
//	"At the beginning of your upkeep, if you control a blue or black
//	 permanent, target creature gets +1/+1 until end of turn. If you control a
//	 blue permanent and a black permanent, that creature gets +5/+5 until end
//	 of turn instead."
//
// One targeted trigger with an intervening if (CR 603.4, read as it goes on the
// stack and as it resolves); the "instead" is read as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51f17fe1-1cf7-4362-b9a2-8ec225d41b03",
		Name:         "Ana Sanctuary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB630UpkeepIfYouControlEitherTargeting("U", "B", "you control a blue or black permanent",
				"Ana Sanctuary — target creature gets +1/+1 until end of turn (+5/+5 if you control a blue permanent and a black permanent)",
				effects.TargetCreature("target creature"),
				func(g *game.Game, item *game.StackItem, both bool) error {
					if both {
						return effects.BoostTargetUntilEOT("Ana Sanctuary — +5/+5", 5, 5)(g, item)
					}
					return effects.BoostTargetUntilEOT("Ana Sanctuary — +1/+1", 1, 1)(g, item)
				}),
		},
	})
}
