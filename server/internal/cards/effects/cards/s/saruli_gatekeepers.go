package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saruli Gatekeepers — Creature — Elf Warrior {3}{G}, 2/4:
//
//	"When this creature enters, if you control two or more Gates, you
//	 gain 7 life."
//
// An intervening "if" (CR 603.4), Duergar Hedge-Mage's shape: the Gate
// count is checked as the creature enters (no trigger, and no target,
// otherwise) and again as the trigger resolves. A Gate is a land subtype,
// so a land that is a Gate by an effect counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "509ea58a-da0a-4e9a-aafc-15f4f65a7467",
		Name:         "Saruli Gatekeepers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondControlTwoOrMoreOfSubtype("Gate"),
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ControlTwoOrMoreOfSubtype("Gate")),
					"Saruli Gatekeepers — if you control two or more Gates, you gain 7 life",
					effects.Do(effects.GainLife{Amount: 7}))),
		},
	})
}
