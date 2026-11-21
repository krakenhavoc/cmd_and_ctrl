package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geistflame Reservoir — Artifact {2}{R} (Innistrad: Midnight Hunt):
//
//	"Whenever you cast an instant or sorcery spell, put a charge counter
//	 on this artifact.
//	 {1}{R}, {T}, Remove any number of charge counters from this artifact:
//	 It deals that much damage to any target.
//	 {1}{R}, {T}: Exile the top card of your library. You may play that
//	 card this turn."
//
// The counter trigger is Cursed Recording's cast watcher. The damage
// ability announces how many counters to remove as a cost (Infused Arrows'
// RemoveCountersXFromThis, zero allowed) and reads the count back from the
// paid-cost record, so "that much" is the number that really came off. The
// artifact stays on the battlefield, so it is the damage source. The
// exile is Ghost-Spider's ExileTopWithPermission.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "143714b8-30ae-42c2-84d9-e58aed7dff5b",
		Name:         "Geistflame Reservoir",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Geistflame Reservoir — put a charge counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterCharge, N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{R}, {T}, Remove any number of charge counters from this artifact: It deals that much damage to any target.",
				Purpose: effects.ForTargets(effects.DamageXToTarget(0)),
				Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost(), effects.RemoveCountersXFromThis(game.CounterCharge, 0)),
				Targets: effects.TargetAny(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := ctx.CountersRemoved()
					for _, t := range ctx.LegalTargets() {
						return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: x}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "{1}{R}, {T}: Exile the top card of your library. You may play that card this turn.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
