package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glen Elendra Archmage — Creature — Faerie Wizard {3}{U}, 2/2:
//
//	"Flying
//	 {U}, Sacrifice this creature: Counter target noncreature spell.
//	 Persist"
//
// The sacrifice is a cost, so the Archmage dies before the counter
// resolves and persist brings it back, ready to do it once more.
// Flying and persist are PrintedKeywords (#2075).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5f3f68b5-8c6a-4181-bb8a-9c73967d198a",
		Name:            "Glen Elendra Archmage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordPersist},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, Sacrifice this creature: Counter target noncreature spell.",
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.SacrificeThis()),
			Targets: effects.TargetSpell("target noncreature spell", effects.Noncreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CounterTheTargetSpell(item, effects.NewContext(g, item))
			},
		}},
	})
}
