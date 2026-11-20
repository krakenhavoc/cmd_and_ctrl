package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eyeblight Assassin — Creature — Elf Assassin {2}{B}, 2/2 (Magic Origins, #5549):
//
//	"When this creature enters, target creature an opponent controls gets -1/-1 until end of turn."
//
// Splatter Goblin's -1/-1 on an entry trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "09cacd93-5c20-4993-9ee2-a756204796f7",
		Name:         "Eyeblight Assassin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				WhenThisEnters("Eyeblight Assassin — target creature an opponent controls gets -1/-1 until end of turn",
					pumpAndGrantFirstTarget(-1, -1, "Eyeblight Assassin — -1/-1 until end of turn")),
				TargetCreature("target creature an opponent controls", OpponentControls())),
				ForTargets(HarmUntilEOT(0, -1, -1))),
		},
	})
}
