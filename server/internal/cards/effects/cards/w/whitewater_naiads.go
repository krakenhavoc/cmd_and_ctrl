package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whitewater Naiads — Enchantment Creature — Nymph {3}{U}{U}, 4/4 (Journey into Nyx, #5558):
//
//	"Constellation — Whenever this creature or another enchantment you
//	 control enters, target creature can't be blocked this turn."
//
// Constellation is Eidolon of Blossoms' condition: its own entry (it is an
// enchantment) or another enchantment entering under your control. Neurok
// Invisimancer's unblockability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c73e533-9ed1-4df0-af12-c0174c0b52b9",
		Name:         "Whitewater Naiads",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.ConstellationTriggers, "Whitewater Naiads — target creature can't be blocked this turn (constellation)",
				effects.TargetCantBeBlockedThisTurn("Whitewater Naiads")), effects.TargetCreature("target creature")),
		},
	})
}
