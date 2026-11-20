package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skarrgan Hellkite — Creature — Dragon {3}{R}{R}, 4/4:
//
//	"Riot (This creature enters with your choice of a +1/+1 counter or
//	 haste.)
//	 Flying
//	 {3}{R}: This creature deals 2 damage divided as you choose among
//	 one or two targets. Activate only if this creature has a +1/+1
//	 counter on it."
//
// Riot is the engine's keyword (ADR 0109 §10), so the counter its
// activation needs is the one riot offers — or any other +1/+1 counter.
// The division is announced with the targets as the ability is
// activated (CR 601.2d via CR 602.2b), Mogg Mob's shape.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bfbe28a2-b656-4575-8b7e-9dbc22f8d4bc",
		Name:            "Skarrgan Hellkite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRiot, "flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     "{3}{R}: This creature deals 2 damage divided as you choose among one or two targets. Activate only if this creature has a +1/+1 counter on it.",
			Cost:      effects.ManaCost("{3}{R}"),
			Condition: effects.SourceHasCountersAtLeast(game.CounterPlusOne, 1),
			Targets:   effects.TargetAny().WithCount(1, 2).Dividing(effects.Divide(2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DealDividedDamage(effects.NewContext(g, item))
			},
		}},
	})
}
