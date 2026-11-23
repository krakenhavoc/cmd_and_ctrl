package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Puppet Crafting — Enchantment — Aura {1}{G} (Reality Fracture):
//
//	"Enchant artifact or non-Aura enchantment
//	 Enchanted permanent is a Construct creature with base power and
//	 toughness 5/5 in addition to its other types.
//	 {4}{G}: Return this card from your graveyard to your hand."
//
// Zoetic Glyph's shape: an additive layer-4 type change (Creature and
// the Construct subtype join what the permanent already is) and a layer
// 7b base P/T set. The return ability functions only from the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d13c111a-ebfc-48ac-b734-f203c7c29d21",
		Name:         "Puppet Crafting",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("enchant artifact or non-Aura enchantment",
			effects.Or(effects.Artifact(), effects.And(effects.Enchantment(), effects.Not(effects.HasSubtype("Aura"))))),
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer4Type,
				AppliesTo: effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					if !effects.ZoeticHas(c.Types, "Creature") {
						c.Types = append(c.Types, "Creature")
					}
					// "Construct" is a word of the Aura's text (CR 612.2, #3984).
					construct := source.TextSubtype("Construct")
					if !effects.ZoeticHas(c.Subtypes, construct) {
						c.Subtypes = append(c.Subtypes, construct)
					}
				},
			},
			effects.SetAttachedBasePT(5, 5),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{4}{G}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
