package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Conspiracy — Enchantment {3}{B}{B} (Mercadian Masques,
// #5948):
//
//	"As this enchantment enters, choose a creature type.
//	 Creatures you control are the chosen type. The same is true for creature
//	 spells you control and creature cards you own that aren't on the
//	 battlefield."
//
// Leyline of Transformation's three layer-4 statics (creatures you control,
// creature spells you control, creature cards you own off the battlefield,
// #3862) with the one difference the text makes: "are the chosen type" and not
// "in addition", so the creature types are replaced (CR 205.1a:
// Characteristic.SetCreatureTypes: every creature type goes, the card types and
// other subtypes stay). The type is the stored "as this enters" choice
// (Card.NamedTribe); before one is chosen nothing is changed.
//
// No simplification.
func init() {
	grant := func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
		if source.NamedTribe != "" {
			c.SetCreatureTypes([]string{source.NamedTribe})
		}
	}
	Register(Spec{
		OracleID:     "775d284d-0884-4ae3-a1b5-2b2a50a895b4",
		Name:         "Conspiracy",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Conspiracy — choose a creature type"),
		Static: []game.StaticAbility{
			{
				Label:     "Creatures you control are the chosen type.",
				Layer:     game.Layer4Type,
				AppliesTo: TribeFilter{YoursOnly: true}.Matches,
				Apply:     grant,
			},
			{
				Label:        "Creature spells you control are the chosen type.",
				Layer:        game.Layer4Type,
				AffectsZones: SpellsOnTheStack(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && target.IsCreature()
				},
				Apply: grant,
			},
			{
				Label:        "Creature cards you own that aren't on the battlefield are the chosen type.",
				Layer:        game.Layer4Type,
				AffectsZones: CardsOffTheBattlefield(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Owner == source.Controller && target.IsCreature()
				},
				Apply: grant,
			},
		},
	})
}
