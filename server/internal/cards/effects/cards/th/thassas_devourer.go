package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thassa's Devourer — Enchantment Creature — Elemental {4}{U}, 2/6 (Journey into Nyx, #5558):
//
//	"Constellation — Whenever this creature or another enchantment you
//	 control enters, target player mills two cards."
//
// Constellation is Eidolon of Blossoms' condition: its own entry (it is an
// enchantment) or another enchantment entering under your control.
// Scrabbling Skullcrab's mill.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9794b16-3ef4-424a-b1ee-e2077f53873d",
		Name:         "Thassa's Devourer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.ConstellationTriggers, "Thassa's Devourer — target player mills two cards (constellation)",
				effects.TargetPlayerMillsEffect(2)), effects.TargetPlayer("target player")),
		},
	})
}
