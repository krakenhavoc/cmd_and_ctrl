package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ephara's Enlightenment — {1}{W}{U} Enchantment — Aura (Born of the Gods):
//
//	"Enchant creature
//	 When this Aura enters, put a +1/+1 counter on enchanted creature.
//	 Enchanted creature has flying.
//	 Whenever a creature you control enters, you may return this Aura to its owner's
//	 hand."
//
// The counter goes on the creature the Aura is attached to as the trigger resolves. The
// return is a "may" asked as the second trigger resolves; an Aura that has left (a new
// object, CR 400.7) returns nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da848568-1c66-4ce1-87f0-a35c317f2c61",
		Name:         "Ephara's Enlightenment",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("flying")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ephara's Enlightenment — put a +1/+1 counter on enchanted creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					host, ok := effects.EnchantedCreatureAsItResolves(ctx)
					if !ok {
						return nil
					}
					return effects.AddCounter{Target: host, Kind: "+1/+1", N: 1}.Apply(ctx)
				}),
			effects.Optional(effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl,
				"Ephara's Enlightenment — you may return this Aura to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					if effects.SourceIsNewObject(g, item) {
						return nil
					}
					return effects.BounceToHand{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
				"Ephara's Enlightenment — return this Aura to its owner's hand?"),
		},
	})
}
