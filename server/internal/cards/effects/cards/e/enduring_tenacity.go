package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enduring Tenacity — Enchantment Creature — Snake Glimmer {2}{B}{B},
// 4/3:
//
//	"Whenever you gain life, target opponent loses that much life.
//	 When Enduring Tenacity dies, if it was a creature, return it to
//	 the battlefield under its owner's control. It's an enchantment.
//	 (It's not a creature.)"
//
// The drain half is Sanguine Bond on a body: every positive
// EventChangeLife on the controller — lifelink included — becomes a
// targeted loss, and the amount is captured by value when the trigger
// is built, so a life total that moves again before it resolves does
// not change the drain.
//
// The Glimmer half is the Enduring cycle's shared dies trigger
// (WhenThisDiesReturnItAsAnEnchantment, glimmer_return.go), written
// for Enduring Curiosity: the card comes back once as an enchantment
// that is not a creature, pinned to that object, and a second death
// finds "if it was a creature" false and leaves it in the graveyard.
// It was held out until the per-object type change existed, because
// a return as an ordinary creature would have been an unkillable
// drain engine — stronger than printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98e698ae-1a69-469c-9cfb-0e3fedeb71d4",
		Name:         "Enduring Tenacity",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventChangeLife},
			AppliesTo: effects.YouGainedLife,
			Targets:   effects.TargetPlayer("target opponent", effects.Opponent()),
			Key:       "Enduring Tenacity — target opponent loses that much life",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrainTargetedOpponent(item.Trigger.Event.Amount)(g, item)
			},
		},
			effects.WhenThisDiesReturnItAsAnEnchantment("Enduring Tenacity"),
		},
	})
}
