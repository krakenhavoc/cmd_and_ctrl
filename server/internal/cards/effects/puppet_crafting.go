package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d13c111a-ebfc-48ac-b734-f203c7c29d21",
		Name:         "Puppet Crafting",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("enchant artifact or non-Aura enchantment",
			Or(Artifact(), And(Enchantment(), Not(HasSubtype("Aura"))))),
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer4Type,
				AppliesTo: AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !zoeticHas(c.Types, "Creature") {
						c.Types = append(c.Types, "Creature")
					}
					if !zoeticHas(c.Subtypes, "Construct") {
						c.Subtypes = append(c.Subtypes, "Construct")
					}
				},
			},
			SetAttachedBasePT(5, 5),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{G}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{4}{G}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  returnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
