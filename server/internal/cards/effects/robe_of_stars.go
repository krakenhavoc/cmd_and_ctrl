package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Robe of Stars — Artifact — Equipment {1}{W}:
//
//	"Equipped creature gets +0/+3.
//	 Astral Projection — {1}{W}: Equipped creature phases out.
//	 Equip {1}"
//
// The phase-out is the engine's (ADR 0084): the Robe is attached, so it
// phases out with the creature and both come back attached during their
// controller's next untap step. The ability can be activated at instant
// speed, which is the point: a creature can dodge removal or a wipe.
//
// No simplification. Mass Casualty deck request (#3245).
func init() {
	Register(Spec{
		OracleID:     "416f49e2-feb1-43e7-befc-b0e02181b202",
		Name:         "Robe of Stars",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(0, 3)},
		Activated: []ActivatedAbility{
			{
				Label:   "Astral Projection — {1}{W}: Equipped creature phases out.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    ManaCost("{1}{W}"),
				Effect:  phaseOutAttachedHost,
			},
			EquipAbility("{1}"),
		},
	})
}
