package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leyline of Transformation — Enchantment {2}{U}{U} (Duskmourn):
//
//	"If this card is in your opening hand, you may begin the game with it on
//	 the battlefield.
//	 As this enchantment enters, choose a creature type.
//	 Creatures you control are the chosen type in addition to their other
//	 types. The same is true for creature spells you control and creature
//	 cards you own that aren't on the battlefield."
//
// Maskwood Nexus with ONE chosen type instead of every type. The opening
// hand is Leyline of Hope's (BeginTheGameOnTheBattlefield), and the type is
// the stored "as this enters" creature-type prompt (ChooseCreatureTypeAsEnters,
// Card.NamedTribe), answered when it begins the game on the battlefield too.
// The grant is three layer-4 statics, exactly the Nexus's: creatures you
// control, creature spells you control (on the stack), and creature cards you
// own off the battlefield (#3862). "In addition": the type is appended and
// the others are kept. Before a type is chosen nothing is added.
func init() {
	grant := func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
		if source.NamedTribe != "" && !hasFold(c.Subtypes, source.NamedTribe) {
			c.Subtypes = append(c.Subtypes, source.NamedTribe)
		}
	}
	Register(Spec{
		OracleID:     "cfaa13d6-9992-4223-8bd7-c44046505555",
		Name:         "Leyline of Transformation",
		Completeness: CompletenessFull,
		OpeningHand:  BeginTheGameOnTheBattlefield(),
		AsEnters:     ChooseCreatureTypeAsEnters("Leyline of Transformation — choose a creature type"),
		Static: []game.StaticAbility{
			{
				Label:     "Creatures you control are the chosen type in addition to their other types.",
				Layer:     game.Layer4Type,
				AppliesTo: TribeFilter{YoursOnly: true}.Matches,
				Apply:     grant,
			},
			{
				Label:        "Creature spells you control are the chosen type in addition to their other types.",
				Layer:        game.Layer4Type,
				AffectsZones: SpellsOnTheStack(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && target.IsCreature()
				},
				Apply: grant,
			},
			{
				Label:        "Creature cards you own that aren't on the battlefield are the chosen type in addition to their other types.",
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
