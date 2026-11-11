package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Panoptic Projektor — Artifact {4} (Murders at Karlov Manor Commander):
//
//	"{T}: The next face-down creature spell you cast this turn costs {3}
//	 less to cast.
//	 If turning a face-down permanent face up causes a triggered ability of a
//	 permanent you control to trigger, that ability triggers an additional
//	 time."
//
// The tap ability leaves a one-shot promise on its controller until end of
// turn, narrowed to a spell cast face down (CR 708.4), Goblin Maskmaker's
// FaceDownOnly shape: a reduction of generic mana only, so a morph's {3}
// costs {0}. The doubler is CR 603.2d's, on the turned-face-up event only:
// every triggered ability of a permanent its controller controls that the
// turn-up triggers does so twice, the turned-up permanent's own included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1e1a787a-2294-472d-908d-d4d72821a7cc",
		Name:            "Panoptic Projektor",
		Completeness:    CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{cmdr2S37PanopticDoubler()},
		Activated: []ActivatedAbility{{
			Label:   "{T}: The next face-down creature spell you cast this turn costs {3} less to cast.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantNextSpellPromise{From: "Panoptic Projektor", Promise: game.NextSpellPromise{
					FaceDownOnly: true,
					Reduce:       3,
					Text:         "The next face-down creature spell you cast this turn costs {3} less to cast.",
				}}.Apply(NewContext(g, item))
			},
		}},
	})
}
