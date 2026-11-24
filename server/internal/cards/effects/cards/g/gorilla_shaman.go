package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gorilla Shaman — Creature — Ape Shaman {R}, 1/1 (Modern Horizons 2):
//
//	"{X}{X}{1}: Destroy target noncreature artifact with mana value X."
//
// {X}{X}{1} is two X slots and a generic: X=2 costs five. The target's mana
// value is checked against the announced X when it is chosen and again as the
// ability resolves (CR 608.2b); X=0 is legal and reaches tokens and other
// zero-cost artifacts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9f53b691-da24-4264-a2a0-d0b95a385963",
		Name:         "Gorilla Shaman",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{X}{1}: Destroy target noncreature artifact with mana value X.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.ManaCost("{X}{X}{1}"),
			Targets: effects.TargetPermanent("target noncreature artifact with mana value X", effects.Artifact(), effects.Not(effects.Creature())).WithManaValueEqualsX(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyTheTargetPermanent(item, effects.NewContext(g, item))
			},
		}},
	})
}
