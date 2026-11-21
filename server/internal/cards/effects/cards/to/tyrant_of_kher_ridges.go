package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tyrant of Kher Ridges — Creature — Dragon {4}{R}{R}, 4/5 (The Brothers' War, #5482):
//
//	"Flying
//	 When this creature enters, it deals 4 damage to any target.
//	 {R}: This creature gets +1/+0 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e7c06af6-8d96-4bfa-8baf-9e80f741e8cb",
		Name:            "Tyrant of Kher Ridges",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Tyrant of Kher Ridges — 4 damage to any target",
				effects.SourceDealsDamageToEachLegalTarget(4)), effects.TargetAny()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Tyrant of Kher Ridges — +1/+0 until end of turn"),
		}},
	})
}
