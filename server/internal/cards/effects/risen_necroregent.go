package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Risen Necroregent — Creature — Zombie Cat Knight {4}{B}, 5/4:
//
//	"Start your engines!
//	 Max speed — At the beginning of your end step, create a 2/2 black
//	 Zombie creature token."
//
// ADR 0138 (#2711). The trigger is gated on its controller's max speed
// as the end step begins.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "527cc22f-b471-4a84-93a8-be56193a9723",
		Name:            "Risen Necroregent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		Triggered: []game.TriggeredAbility{
			MaxSpeedTrigger(TriggerWithPurpose(
				AtYourEndStep("Risen Necroregent — create a 2/2 black Zombie",
					Do(CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1})),
				game.Purpose{Tokens: 1})),
		},
	})
}
