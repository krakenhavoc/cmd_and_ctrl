package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stunning Strike — Enchantment — Aura {2}{U} (Commander Legends: Battle for
// Baldur's Gate, #5948):
//
//	"Flash
//	 Enchant creature
//	 When this Aura enters, tap enchanted creature and remove it from combat.
//	 As long as enchanted creature isn't legendary, it doesn't untap during its
//	 controller's untap step."
//
// Waterknot's entry tap, then the creature leaves combat (CR 506.4) if it was
// in it: a creature not in combat is simply left alone. The untap-step
// restriction is Waterknot's, held only while the enchanted creature is not
// legendary, read as each untap step begins on its effective supertypes
// (CR 205.4), so a creature that becomes or stops being legendary is caught
// either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "139574fd-c61b-4615-9200-59f2c65f5703",
		Name:            "Stunning Strike",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{{
			Label: "enchanted creature doesn't untap while it isn't legendary",
			Restricts: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.AttachedToSource(target, g, source) && !target.IsLegendary()
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Stunning Strike — tap enchanted creature and remove it from combat", func(g *game.Game, item *game.StackItem) error {
				if err := effects.TapEnchantedCreatureOnEntry(g, item); err != nil {
					return err
				}
				aura, ok := effects.NewContext(g, item).TriggeringPermanent()
				if !ok || aura.AttachedTo.Kind != game.TargetCard || aura.AttachedTo.ID == uuid.Nil {
					return nil
				}
				if host, ok := g.LookupCardForEffect(aura.AttachedTo.ID); ok && (host.AttackingTarget != uuid.Nil || host.BlockingTarget != uuid.Nil) {
					return g.RemoveFromCombatForEffect(aura.AttachedTo.ID)
				}
				return nil
			}),
		},
	})
}
