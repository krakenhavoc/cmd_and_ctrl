package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eyeblight Assassin — Creature — Elf Assassin {2}{B}, 2/2 (Magic Origins, #5549):
//
//	"When this creature enters, target creature an opponent controls gets -1/-1 until end of turn."
//
// Splatter Goblin's -1/-1 on an entry trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09cacd93-5c20-4993-9ee2-a756204796f7",
		Name:         "Eyeblight Assassin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEnters("Eyeblight Assassin — target creature an opponent controls gets -1/-1 until end of turn",
					effects.PumpAndGrantFirstTarget(-1, -1, "Eyeblight Assassin — -1/-1 until end of turn")),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.HarmUntilEOT(0, -1, -1))),
		},
	})
}
