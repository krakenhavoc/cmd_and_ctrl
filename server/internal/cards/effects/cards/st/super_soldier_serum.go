package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Super-Soldier Serum — Enchantment — Aura {1}{W} (Marvel Super Heroes,
// #4455):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2, has first strike and vigilance, and is
//	 a legendary Soldier in addition to its other types.
//	 Whenever enchanted creature attacks or blocks, attach any number of
//	 target Equipment you control to it."
//
// The pump and keywords are the usual Aura statics. "Legendary Soldier in
// addition" is a layer 4 effect (CR 613.1d) adding the Legendary
// supertype and the Soldier subtype ("Soldier" a word of the Aura's text,
// CR 612.2) without removing anything. The trigger belongs to the Aura and
// fires for the attack or the block of the enchanted creature; it targets
// any number of Equipment its controller controls (zero is legal) and
// attaches each to that creature (AttachForEffect; an Equipment that is
// already attached to it, or that has left, is skipped by the engine).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f2ba142-5e7c-4709-ae03-e028cb2a9872",
		Name:         "Super-Soldier Serum",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("first strike", "vigilance"),
			{
				Label:     "Enchanted creature is a legendary Soldier in addition to its other types.",
				Layer:     game.Layer4Type,
				AppliesTo: effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					if !effects.EotHasType(c.Supertypes, "Legendary") {
						c.Supertypes = append(c.Supertypes, "Legendary")
					}
					soldier := source.TextSubtype("Soldier")
					if !effects.EotHasType(c.Subtypes, soldier) {
						c.Subtypes = append(c.Subtypes, soldier)
					}
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return (ev.Kind == game.EventAttack || ev.Kind == game.EventBlock) && source.IsAttachedTo(ev.CardID)
				}, "Super-Soldier Serum — attach any number of target Equipment you control to the enchanted creature",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						creature := game.TargetRef{Kind: game.TargetCard, ID: item.Trigger.Event.CardID}
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							if err := g.AttachForEffect(t.ID, creature); err != nil {
								return err
							}
						}
						return nil
					}),
				effects.TargetPermanent("any number of target Equipment you control", effects.HasSubtype("Equipment"), effects.YouControl()).WithCount(0, 0)),
		},
	})
}
