package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Suit — Artifact — Equipment {1} (Marvel's Spider-Man, #4460):
//
//	"Equipped creature gets +2/+2 and is a Spider Hero in addition to its
//	 other types.
//	 Equip {3} ({3}: Attach to target creature you control. Equip only as a
//	 sorcery.)"
//
// One continuous effect in two layers, as Blade of the Oni's Demon: the
// Spider and Hero subtypes in layer 4 (read through the Equipment's text,
// CR 612.2) and +2/+2 in layer 7c. "In addition to" keeps the host's own
// types (CR 205.1b). Both parts declare ContinuesAfterRemoval, because
// CR 613.6 keeps the later layer applying once the layer-4 part has started.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba08b4cf-9a13-40fd-8522-9504d2ff1a19",
		Name:         "Spider-Suit",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:                 "Equipped creature is a Spider Hero in addition to its other types.",
				Layer:                 game.Layer4Type,
				ContinuesAfterRemoval: true,
				AppliesTo:             AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					for _, word := range []string{"Spider", "Hero"} {
						sub := source.TextSubtype(word)
						if !hasFold(c.Subtypes, sub) {
							c.Subtypes = append(c.Subtypes, sub)
						}
					}
				},
			},
			{
				Label:                 "Equipped creature gets +2/+2.",
				Layer:                 game.Layer7PT,
				SubLayer:              game.SubLayer7C_Modify,
				ContinuesAfterRemoval: true,
				AppliesTo:             AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
					c.Toughness += 2
				},
			},
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
