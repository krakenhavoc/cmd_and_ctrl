package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Relic Amulet — Artifact {2} (Zendikar Rising, S82 OLD #5508):
//
//	"Whenever you cast an instant, sorcery, or Wizard spell, put a charge
//	 counter on this artifact.
//	 {2}, {T}, Remove all charge counters from this artifact: It deals that
//	 much damage to target creature."
//
// The counters are removed as a cost (CR 602.2b) and their number carries to
// the effect, so the damage is what was taken off even if the Amulet is gone
// by resolution; a target that is illegal then makes the ability do nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3998d9c-56b1-486c-818c-a6882e999d54",
		Name:         "Relic Amulet",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery(), effects.HasSubtype("Wizard")),
				"Relic Amulet — put a charge counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterCharge, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Remove all charge counters from this artifact: It deals that much damage to target creature.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.RemoveAllCountersFromThis(game.CounterCharge)),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: ctx.CountersRemoved()}.Apply(ctx)
			},
		}},
	})
}
