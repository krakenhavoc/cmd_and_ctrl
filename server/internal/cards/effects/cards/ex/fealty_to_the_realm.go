package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fealty to the Realm — Enchantment — Aura {4}{U} (Tales of Middle-earth
// Commander):
//
//	"Enchant creature
//	 When this Aura enters, you become the monarch.
//	 The monarch controls enchanted creature.
//	 Enchanted creature attacks each combat if able and can't attack you."
//
// Coerced to Kill's control-changing static (layer 2, CR 613.2) with the
// controller read off the crown as the layers run, instead of the Aura's own
// controller; the crown moving invalidates the layers (EventMonarchChanged),
// so the creature changes hands with it. With no monarch nobody controls it
// by this effect and it stays with whoever has it otherwise (CR 725.4: the
// designation can be gone). The requirement is the Aura's, Ghoulish
// Impetus-style (AttacksEachCombatWhere over the attached creature); "can't
// attack you" is the Aura controller (not the monarch) and the player only,
// not their planeswalkers, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0128941-240a-4ca1-9816-b32f47523fdb",
		Name:         "Fealty to the Realm",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			{
				Label:     "The monarch controls enchanted creature.",
				Layer:     game.Layer2Control,
				AppliesTo: effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					if g.Monarch == uuid.Nil {
						return
					}
					c.Controller = g.Monarch
					c.ControlSource = source.InstanceID
				},
			},
			func() game.StaticAbility {
				s := effects.AttacksEachCombatWhere(effects.AttachedToSource)
				s.Label = "Enchanted creature attacks each combat if able."
				return s
			}(),
			{
				Label:     "Enchanted creature can't attack you.",
				Layer:     game.Layer6Ability,
				AppliesTo: effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					if source == nil {
						return
					}
					c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, game.AttackTargetRestriction{
						Source:        source.InstanceID,
						SourceName:    source.Name,
						NotPlayer:     source.Controller,
						NotPlayerOnly: true,
					})
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Fealty to the Realm — you become the monarch", effects.Do(effects.BecomeTheMonarch{})),
		},
	})
}
