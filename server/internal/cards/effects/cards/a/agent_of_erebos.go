package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Agent of Erebos — Enchantment Creature — Zombie {3}{B}, 2/2 (Journey into Nyx, #5558):
//
//	"Constellation — Whenever this creature or another enchantment you
//	 control enters, exile target player's graveyard."
//
// Constellation is Eidolon of Blossoms' condition: its own entry (it is an
// enchantment) or another enchantment entering under your control. Angel
// of Finality's graveyard exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3fa5e109-4a67-4523-8183-90e48a906400",
		Name:         "Agent of Erebos",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.ConstellationTriggers, "Agent of Erebos — exile target player's graveyard (constellation)",
				effects.ExileTargetPlayersGraveyard), effects.TargetPlayer("target player")),
		},
	})
}
