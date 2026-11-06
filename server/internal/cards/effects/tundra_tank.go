package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tundra Tank — Artifact — Vehicle {2}{B}, 4/4:
//
//	"Firebending 1 (Whenever this creature attacks, add {R}. This mana
//	 lasts until end of combat.)
//	 When this Vehicle enters, target creature you control gains
//	 indestructible until end of turn.
//	 Crew 1"
//
// Firebending is Avatar Roku's attack trigger with N = 1: a triggered
// ability that uses the stack (the rulings), its {R} marked to last
// until end of combat. It only matters while the Vehicle is crewed,
// since only a creature attacks. The entry trigger targets a creature
// its controller controls (the Vehicle itself only if it is somehow
// already a creature); a target gone by resolution gets nothing (CR
// 608.2b). Crew 1 is the engine's crew cost.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "2ded4b02-7aa4-46a1-b10f-8480eb942cd4",
		Name:         "Tundra Tank",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Tundra Tank — firebending 1", Do(firebending(1))),
			Targeting(WhenThisEnters("Tundra Tank — target creature you control gains indestructible until end of turn",
				jmpB7FirstLegalTargetGains("Tundra Tank — indestructible until end of turn", "indestructible")),
				TargetCreature("target creature you control", YouControl())),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Tundra Tank"),
		}},
	})
}
