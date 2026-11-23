package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sinister Monolith — Artifact {3}{B} (Bloomburrow, #4467):
//
//	"At the beginning of combat on your turn, each opponent loses 1 life
//	 and you gain 1 life.
//	 {T}, Pay 2 life, Sacrifice this artifact: Draw two cards. Activate
//	 only as a sorcery."
//
// The trigger is Additive Evolution's beginning-of-combat step on your
// turn only, with Acolyte of Aclazotz's drain: every opponent still in
// the game loses 1 life, then you gain 1, whatever was lost. The
// activated ability pays all three costs at once (CR 601.2h); paying 2
// life needs a life total of at least 2 (CR 119.4). It is sorcery speed,
// so it declares no Answers (ADR 0142).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae8003eb-9f19-4f23-be3b-0ac3ddc03a72",
		Name:         "Sinister Monolith",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true),
				"Sinister Monolith — each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}, Pay 2 life, Sacrifice this artifact: Draw two cards. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.TapCost(), effects.PayLife(2), effects.SacrificeThis()),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.DrawCards{N: 2}),
		}},
	})
}
