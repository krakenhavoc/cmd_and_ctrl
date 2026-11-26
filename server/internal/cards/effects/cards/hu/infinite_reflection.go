package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Infinite Reflection — Enchantment — Aura {5}{U} (Commander 2014):
//
//	"Enchant creature
//	 When this Aura enters attached to a creature, each other nontoken
//	 creature you control becomes a copy of that creature.
//	 Nontoken creatures you control enter as a copy of enchanted creature."
//
// The enters trigger reads the creature the Aura is attached to as it
// resolves (CR 608.2h) and makes every other nontoken creature its
// controller has a copy of it with no duration (CR 707.2, 611.2a). The
// static half is #6262's replacement: a nontoken creature entering under
// the Aura controller's control enters as a copy of the enchanted creature,
// and nothing happens while the Aura is attached to no creature.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5bbcfb73-cccc-47b0-81ae-17eb7c33ff5e",
		Name:         "Infinite Reflection",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Replacements: []game.ReplacementEffect{
			effects.NontokenCreaturesYouControlEnterAsCopyOfEnchanted("Infinite Reflection: nontoken creatures you control enter as a copy of enchanted creature"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Infinite Reflection — each other nontoken creature you control becomes a copy of enchanted creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					aura, ok := ctx.TriggeringPermanent()
					if !ok || aura.AttachedTo.Kind != game.TargetCard {
						return nil
					}
					host := aura.AttachedTo.ID
					if !effects.OnBattlefield(g, host) {
						return nil
					}
					others := effects.PermanentsControlledByMatching(g, item.Controller, effects.And(effects.Creature(), effects.NotSelf(host),
						func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return !c.IsToken() }))
					if len(others) == 0 {
						return nil
					}
					return effects.BecomeCopy{
						Targets:  others,
						Of:       host,
						Duration: effects.CopyIndefinite,
						Label:    "Infinite Reflection — becomes a copy of the enchanted creature",
					}.Apply(ctx)
				}),
		},
	})
}
