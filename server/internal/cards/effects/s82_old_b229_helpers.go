package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b229_helpers.go — the shape Arcane Adaptation shares with the
// Leyline of Transformation family (S82 older sets, batch 229).

// ChosenTypeInAdditionStatics is "Creatures you control are the chosen type in
// addition to their other types. The same is true for creature spells you
// control and creature cards you own that aren't on the battlefield": three
// layer-4 statics that append the source's chosen creature type
// (Card.NamedTribe). Before a type is chosen nothing is added.
func ChosenTypeInAdditionStatics() []game.StaticAbility {
	grant := func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
		if source.NamedTribe != "" && !HasFold(c.Subtypes, source.NamedTribe) {
			c.Subtypes = append(c.Subtypes, source.NamedTribe)
		}
	}
	return []game.StaticAbility{
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
	}
}
