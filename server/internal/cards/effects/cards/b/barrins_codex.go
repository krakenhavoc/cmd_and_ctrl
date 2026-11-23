package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barrin's Codex — Artifact — Book {4} (Urza's Saga):
//
//	"At the beginning of your upkeep, you may put a page counter on this
//	 artifact.
//	 {4}, {T}, Sacrifice this artifact: Draw X cards, where X is the
//	 number of page counters on this artifact."
//
// The upkeep counter is a "you may" trigger, Darksteel Reactor's shape.
// The Codex is sacrificed as part of the cost, so X is counted from the
// artifact as it last existed on the battlefield (CR 608.2h), the way
// Hoarder's Overflow counts its stash counters.
//
// No simplification.
func init() {
	const page = "page"
	effects.Register(effects.Spec{
		OracleID:     "f7c237e6-17a4-4e72-8cc3-456aec738394",
		Name:         "Barrin's Codex",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Barrin's Codex — put a page counter on this artifact", effects.PutACounterOnThis(page)),
				"Barrin's Codex — put a page counter on it?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}, Sacrifice this artifact: Draw X cards, where X is the number of page counters on this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{N: effects.CountersOnThisAsItResolves(page)(ctx)}.Apply(ctx)
			})),
		}},
	})
}
