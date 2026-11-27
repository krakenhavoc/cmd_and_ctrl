package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Druids' Repository — Enchantment {1}{G}{G}:
//
//	"Whenever a creature you control attacks, put a charge counter on this
//	 enchantment.
//	 Remove a charge counter from this enchantment: Add one mana of any color."
//
// One counter per attacking creature: the engine raises one attack event for
// each. The mana ability is a cost-only one, so it does not use the stack.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cbbc530f-e51b-423e-b219-23bb580fd834",
		Name:         "Druids' Repository",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Druids' Repository — put a charge counter on this enchantment", effects.PutChargeCounterOnThis),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{RemoveCounters: effects.RemoveCountersFromThis(game.CounterCharge, 1).RemoveCounters},
			Produced: "{W|U|B|R|G}",
			Label:    "Remove a charge counter from this enchantment: Add one mana of any color",
		}},
	})
}
