package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "bfbe28a2-b656-4575-8b7e-9dbc22f8d4bc",
		Name:            "Skarrgan Hellkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRiot, "flying"},
		Activated: []ActivatedAbility{{
			Label:     "{3}{R}: This creature deals 2 damage divided as you choose among one or two targets. Activate only if this creature has a +1/+1 counter on it.",
			Cost:      ManaCost("{3}{R}"),
			Condition: SourceHasCountersAtLeast(game.CounterPlusOne, 1),
			Targets:   TargetAny().WithCount(1, 2).Dividing(Divide(2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DealDividedDamage(NewContext(g, item))
			},
		}},
	})
}
