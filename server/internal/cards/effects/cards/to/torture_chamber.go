package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Torture Chamber — Artifact {3} (Tempest, S82 OLD batch 759, #5628):
//
//	"At the beginning of your upkeep, put a pain counter on this artifact.
//	 At the beginning of your end step, this artifact deals damage to you
//	 equal to the number of pain counters on it.
//	 {1}, {T}, Remove all pain counters from this artifact: It deals damage
//	 to target creature equal to the number of pain counters removed this
//	 way."
//
// The end-step damage counts the counters as the trigger resolves
// (CR 608.2h). The activation removes every pain counter as a cost (Jar of
// Eyeballs' "remove all"), and the damage is the number removed, a fact of
// the announcement read back as the ability resolves; with the Chamber gone
// by then it still deals the damage (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12436f2f-eaed-4b46-8570-81316ce15962",
		Name:         "Torture Chamber",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Torture Chamber — put a pain counter on this artifact", effects.PutACounterOnThis("pain")),
			effects.AtYourEndStep("Torture Chamber — damage to you equal to the number of pain counters on it",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					n := effects.CountersOnThisAsItResolves("pain")(ctx)
					if n <= 0 {
						return nil
					}
					return effects.DealDamage{Source: ctx.Source(), Target: ctx.Controller(), Amount: n}.Apply(ctx)
				}))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Remove all pain counters from this artifact: It deals damage to target creature equal to the number of pain counters removed this way.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.RemoveAllCountersFromThis("pain")),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				n := ctx.CountersRemoved()
				if n <= 0 {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
