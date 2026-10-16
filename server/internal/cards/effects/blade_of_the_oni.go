package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blade of the Oni — Artifact Creature — Equipment Demon {1}{B}, 3/1:
//
//	"Menace
//	 Equipped creature has base power and toughness 5/5, has menace, and
//	 is a black Demon in addition to its other colors and types.
//	 Reconfigure {2}{B}{B}"
//
// One continuous effect in four layers: the Demon subtype in layer 4,
// black in layer 5, menace in layer 6 and base 5/5 in layer 7b. "In
// addition to" keeps the host's own colours and types (CR 205.1b). Every
// part declares ContinuesAfterRemoval, because CR 613.6 keeps the later
// layers applying once the layer-4 part has started, even if the Blade
// loses its abilities part-way through the pass.
//
// Reconfigure is reconfigure.go (#2639). While attached the Blade is not
// a creature and loses its own Demon type (CR 702.151b, 205.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6d9ac636-3c6c-40ca-90f3-1a3387836bb4",
		Name:            "Blade of the Oni",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{
			bladeOfTheOniMakes(game.Layer4Type, "Equipped creature is a Demon in addition to its other types.", func(c *game.Characteristic, source *game.Card) {
				demon := source.TextSubtype("Demon")
				if !hasFold(c.Subtypes, demon) {
					c.Subtypes = append(c.Subtypes, demon)
				}
			}),
			bladeOfTheOniMakes(game.Layer5Color, "Equipped creature is black in addition to its other colors.", func(c *game.Characteristic, source *game.Card) {
				black := source.TextColor("B")
				if !hasFold(c.Colors, black) {
					c.Colors = append(c.Colors, black)
				}
			}),
			equippedCreatureIs(game.Layer6Ability, "Equipped creature has menace.", func(c *game.Characteristic) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "menace")
			}),
			SetAttachedBasePT(5, 5),
		},
		Activated: Reconfigure("{2}{B}{B}"),
	})
}

// bladeOfTheOniMakes is equippedCreatureIs with the Equipment in hand:
// "a Demon" and "black" are words of its text (CR 612.2, #3984), read
// through it as the pass applies them.
func bladeOfTheOniMakes(layer game.Layer, label string, apply func(c *game.Characteristic, source *game.Card)) game.StaticAbility {
	return game.StaticAbility{
		Label:                 label,
		Layer:                 layer,
		ContinuesAfterRemoval: true,
		AppliesTo:             AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			apply(c, source)
		},
	}
}
