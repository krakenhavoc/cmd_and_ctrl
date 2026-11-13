package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f7c237e6-17a4-4e72-8cc3-456aec738394",
		Name:         "Barrin's Codex",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(AtYourUpkeep("Barrin's Codex — put a page counter on this artifact", putACounterOnThis(page)),
				"Barrin's Codex — put a page counter on it?"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}, Sacrifice this artifact: Draw X cards, where X is the number of page counters on this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), TapCost(), SacrificeThis()),
			Effect: Do(Instruction(func(ctx *Context) error {
				return DrawCards{N: CountersOnThisAsItResolves(page)(ctx)}.Apply(ctx)
			})),
		}},
	})
}
