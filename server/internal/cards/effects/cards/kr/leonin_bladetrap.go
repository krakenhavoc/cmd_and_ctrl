package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leonin Bladetrap — Artifact {3} (Commander 2013, S82 OLD b386):
//
//	"Flash
//	 {2}, Sacrifice this artifact: It deals 2 damage to each attacking
//	 creature without flying."
//
// Rolling Temblor's damage over attacking creatures lacking flying, from the
// sacrificed Bladetrap on its last-known information (CR 113.7a). One damage
// event; outside combat nothing is attacking and nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ca21d597-40a5-4724-9d53-1ab6e6c6e767",
		Name:            "Leonin Bladetrap",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice this artifact: It deals 2 damage to each attacking creature without flying.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageEachMatching(effects.NewContext(g, item),
					effects.And(effects.AttackingCreature(), effects.WithoutKeyword("flying")), 2)
			},
		}},
	})
}
