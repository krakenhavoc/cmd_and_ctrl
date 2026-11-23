package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunbathing Rootwalla — Creature — Lizard {1}{G}, 2/2:
//
//	"Domain — {3}{G}: Until end of turn, this creature gets +1/+1 for
//	 each basic land type among lands you control. Activate only once
//	 each turn."
//
// The count is read as the ability resolves (CR 611.2a: the amount is
// fixed then, not tracked afterwards), from the shared domain count
// (DomainCount, #5666). The limit is the announcement tally
// (OncePerTurnActivation, CR 602.1b).
//
// No simplification.
func init() {
	const label = "Domain — {3}{G}: Until end of turn, this creature gets +1/+1 for each basic land type among lands you control. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "c47ae79c-4e07-47f2-8920-2ab62b471e4f",
		Name:         "Sunbathing Rootwalla",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{3}{G}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := effects.DomainCount(g, item.Controller)
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n, Label: label}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
