package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shocking Sharpshooter — Creature — Human Archer {1}{R}, 1/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Reach
//	 Whenever another creature you control enters, this creature deals 1
//	 damage to target opponent."
//
// Raphael, Tough Turtle's trigger: the opponent is chosen as the trigger
// goes on the stack (CR 603.3d), and the Sharpshooter is the damage's
// source, as it last existed if it has left (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fbb7735f-ec91-4ed4-b735-d94a76e68a80",
		Name:            "Shocking Sharpshooter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverAnotherCreatureEntersUnderYourControl("Shocking Sharpshooter — 1 damage to target opponent",
					b36DesertDualDamage),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
