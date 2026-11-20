package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charmed Clothier — Creature — Faerie Advisor {4}{W}, 3/3:
//
//	"Flying
//	 When this creature enters, create a Royal Role token attached to
//	 another target creature you control. (If you control another Role
//	 on it, put that one into the graveyard. Enchanted creature gets
//	 +1/+1 and has ward {1}.)"
//
// "Another" excludes the Clothier by object; with no other creature the
// trigger is removed (CR 603.3d). No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ad0e6b12-a6bd-4860-9eac-68b2492d9567",
		Name:            "Charmed Clothier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Charmed Clothier — create a Royal Role token attached to another target creature you control",
					effects.CreateRoleOnFirstTarget(effects.RoleRoyal)),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
