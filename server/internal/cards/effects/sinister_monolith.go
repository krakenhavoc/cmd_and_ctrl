package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ae8003eb-9f19-4f23-be3b-0ac3ddc03a72",
		Name:         "Sinister Monolith",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan, StepBegan(game.StepBeginCombat, true),
				"Sinister Monolith — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{T}, Pay 2 life, Sacrifice this artifact: Draw two cards. Activate only as a sorcery.",
			Cost:         Plus(TapCost(), PayLife(2), SacrificeThis()),
			SorcerySpeed: true,
			Effect:       Do(DrawCards{N: 2}),
		}},
	})
}
