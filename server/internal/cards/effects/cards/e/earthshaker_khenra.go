package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Earthshaker Khenra — {1}{R} Creature — Jackal Warrior 2/1 (Hour of Devastation):
//
//	"Haste
//	 When this creature enters, target creature with power less than or equal
//	 to this creature's power can't block this turn.
//	 Eternalize {4}{R}{R}"
//
// The comparison is with the source's power, judged as the trigger is put on
// the stack and again as it resolves (RelativeToSource, #2146).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "df660a8e-39c2-455d-8860-522e2998a95e",
		Name:            "Earthshaker Khenra",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Earthshaker Khenra — target creature with power less than or equal to this creature's power can't block this turn",
					effects.TargetCreatureCantBlockThisTurn("Earthshaker Khenra — target creature can't block this turn")),
				effects.RelativeToSource(effects.TargetCreature("target creature with power less than or equal to this creature's power"), effects.PowerNoGreater()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.Eternalize("{4}{R}{R}")},
	})
}
