package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loathsome Catoblepas — Creature — Beast {5}{B}, 3/3 (Theros):
//
//	"{2}{G}: This creature must be blocked this turn if able.
//	 When this creature dies, target creature an opponent controls gets -3/-3
//	 until end of turn."
//
// The requirement is a data record pinned to the Catoblepas until end of turn
// (BlockRequirementUntilEOT), so it does not follow a new object (CR 400.7).
// The death trigger is Black Dragon's -3/-3 on a creature an opponent
// controls, chosen as it goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6f68e4b-af2b-43d5-8052-104defe7f3ec",
		Name:         "Loathsome Catoblepas",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Loathsome Catoblepas — target creature an opponent controls gets -3/-3 until end of turn",
					effects.BoostTargetUntilEOT("Loathsome Catoblepas: -3/-3 until end of turn", -3, -3)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: This creature must be blocked this turn if able.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{G}"),
			Effect:  effects.OldB395MustBeBlockedThis,
		}},
	})
}
