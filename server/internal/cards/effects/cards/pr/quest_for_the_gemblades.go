package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quest for the Gemblades — Enchantment {1}{G} (Zendikar, S82 OLD b493):
//
//	"Whenever a creature you control deals combat damage to a creature, you may
//	 put a quest counter on this enchantment.
//	 Remove a quest counter from this enchantment and sacrifice it: Put four
//	 +1/+1 counters on target creature."
//
// Quest for Pure Flame's damage read (B34DamageSourceController) narrowed to
// combat damage dealt to a creature; each creature that deals damage is its own
// event and its own "may". The activation is Khalni Heart Expedition's cost
// shape with Agent of Kotis' counters on a target creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c762361a-0b72-4db5-8b6f-715ca42d5067",
		Name:         "Quest for the Gemblades",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !ev.Combat || ev.Amount <= 0 {
					return false
				}
				target, ok := g.LookupCardForEffect(ev.Target)
				if !ok || !target.IsCreature() {
					return false
				}
				controller, ok := effects.B34DamageSourceController(ev, g)
				return ok && controller == source.Controller
			}, "Quest for the Gemblades — put a quest counter",
				effects.Do(effects.CounterOnThis{Kind: "quest", N: 1})), "Quest for the Gemblades — put a quest counter on it?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove a quest counter from this enchantment and sacrifice it: Put four +1/+1 counters on target creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.RemoveCountersFromThis("quest", 1), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.RenewCountersOnTarget(game.CounterPlusOne, 4),
		}},
	})
}
