package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phytohydra — Creature — Plant Hydra {2}{G}{W}{W}, 1/1 (Ravnica Remastered,
// #5948):
//
//	"If damage would be dealt to this creature, put that many +1/+1 counters
//	 on it instead."
//
// A replacement effect, not a prevention (CR 614.1a): it applies to damage
// that can't be prevented too, and it is not a "prevent" for anything that
// reads one. The damage is not dealt (so no lifelink and no deathtouch from
// it, CR 614.6) and the counters go on through the counter window, where a
// doubling effect applies to them. The amount is the damage as it would be
// dealt, after any earlier replacement has changed it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f5a7b65-b7f4-4bad-acfd-35edfb8f86a7",
		Name:         "Phytohydra",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: false, // CR 614.1a: a replacement, not a prevention
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return src != nil && ev.Kind == game.RepEventDamage && ev.DamageTarget == src.InstanceID && ev.DamageAmount > 0
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				n := ev.DamageAmount
				ev.Cancel()
				_, err := g.AddCounterMustSettleNowForEffect(src.InstanceID, game.CounterPlusOne, n)
				return err
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Phytohydra — put that many +1/+1 counters on it instead of dealing damage to it",
		}},
	})
}
