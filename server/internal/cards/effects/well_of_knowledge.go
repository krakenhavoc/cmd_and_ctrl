package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Well of Knowledge — Artifact {3}:
//
//	"{2}: Draw a card. Any player may activate this ability but only
//	 during their draw step."
//
// An any-player row (CR 602.2, 602.1b) with a timing instruction:
// "only during their draw step" names the ACTIVATOR's draw step, so the
// Condition asks two things of the activating player — that the turn is
// theirs and that the step is the draw step (CR 504). Each player may
// use it once their draw step's turn-based draw is done and they have
// priority (CR 504.1, 504.2), as many times as they can pay, and never
// in another player's draw step. Whoever activates it pays the {2} out
// of their own pool (CR 602.1a) and draws (CR 109.5, 113.8).
//
// Purpose {Draws: 1}: the effect plainly helps whoever activates it, so
// the bot may use another player's Well (ADR 0106 owner decision 2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "99e2b36b-1fab-4b09-b371-d97a6854dfbc",
		Name:         "Well of Knowledge",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{2}: Draw a card. Any player may activate this ability but only during their draw step.",
			Cost:      ManaCost("{2}"),
			Condition: AllConditions(DuringStep(game.StepDraw), DuringYourTurn()),
			AnyPlayer: true,
			Purpose:   game.ActivationPurpose{Draws: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
