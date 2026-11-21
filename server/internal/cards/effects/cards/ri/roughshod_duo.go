package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roughshod Duo — Creature — Mouse Raccoon {2}{R}, 3/2 (Bloomburrow,
// #4467):
//
//	"Trample
//	 Whenever you expend 4, target creature you control gets +1/+1 and
//	 gains trample until end of turn. (You expend 4 as you spend your
//	 fourth total mana to cast spells during a turn.)"
//
// Trample is a printed keyword. Expend is WheneverYouExpend (expend.go),
// the turn's tally of mana spent on spells, with the target declared on
// the trigger: chosen as it goes on the stack, re-checked as it resolves
// (pumpAndGrantFirstTarget).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0cc89c15-5a48-4998-b6df-21791b6ed9a3",
		Name:            "Roughshod Duo",
		ReadsManaSpent:  game.ReadsYourCasts,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WheneverYouExpend(4, "Roughshod Duo — target creature you control gets +1/+1 and gains trample until end of turn",
					effects.PumpAndGrantFirstTarget(1, 1, "Roughshod Duo — +1/+1 and trample", "trample")),
				effects.TargetCreature("target creature you control", effects.YouControl())),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "trample"))),
		},
	})
}
