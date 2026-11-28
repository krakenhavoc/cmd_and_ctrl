package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shivan Emissary (Invasion, S82 OLD batch 648, #5616):
//
//	Creature — Human Wizard {2}{R}, 1/1:
//	"Kicker {1}{B}
//	When this creature enters, if it was kicked, destroy target nonblack creature. It can't be regenerated."
func init() {
	effects.Register(effects.Spec{
		OracleID:      "1b885600-def1-4b2c-a6cb-02df34d58a1b",
		Name:          "Shivan Emissary",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Shivan Emissary — kicked, destroy target nonblack creature; it can't be regenerated",
					effects.Do(effects.Instruction(effects.S82OldB648DestroyTheTargetNoRegen))),
				effects.TargetCreature("target nonblack creature", effects.NonBlack())),
		},
	})
}
