package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "143714b8-30ae-42c2-84d9-e58aed7dff5b",
		Name:         "Geistflame Reservoir",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()),
				"Geistflame Reservoir — put a charge counter on it",
				Do(CounterOnThis{Kind: game.CounterCharge, N: 1})),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}{R}, {T}, Remove any number of charge counters from this artifact: It deals that much damage to any target.",
				Purpose: ForTargets(DamageXToTarget(0)),
				Cost:    Plus(ManaCost("{1}{R}"), TapCost(), RemoveCountersXFromThis(game.CounterCharge, 0)),
				Targets: TargetAny(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					x := ctx.CountersRemoved()
					for _, t := range ctx.LegalTargets() {
						return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: x}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "{1}{R}, {T}: Exile the top card of your library. You may play that card this turn.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}{R}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
