package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raven's Run — Plane — Shadowmoor (Planechase Anthology, #4870):
//
//	"All creatures have wither. (They deal damage to creatures in the form
//	 of -1/-1 counters.)
//	 Whenever chaos ensues, put a -1/-1 counter on target creature, two
//	 -1/-1 counters on another target creature, and three -1/-1 counters on
//	 a third target creature."
//
// A plane (ADR 0216 §6). Three target clauses, each different from the
// earlier ones (the ruling: three different creatures, even ones the
// planar controller controls); the trigger is removed if three can't be
// chosen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b53e06af-be18-4f75-9dd0-7494ec867b13",
		Name:         "Raven's Run",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{b9AllCreaturesHave("wither")},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("Raven's Run — a -1/-1 counter on one target creature, two on another, three on a third",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for i, t := range item.Targets {
						if !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (AddCounter{Target: t.ID, Kind: game.CounterMinusOne, N: i + 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
				TargetCreature("target creature").Then(
					Distinct(TargetCreature("another target creature")),
					Distinct(TargetCreature("a third target creature")))),
		},
	})
}
