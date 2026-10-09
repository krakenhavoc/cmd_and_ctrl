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
			equippedCreatureIs(game.Layer4Type, func(c *game.Characteristic) {
				if !hasFold(c.Subtypes, "Demon") {
					c.Subtypes = append(c.Subtypes, "Demon")
				}
			}),
			equippedCreatureIs(game.Layer5Color, func(c *game.Characteristic) {
				if !hasFold(c.Colors, "B") {
					c.Colors = append(c.Colors, "B")
				}
			}),
			equippedCreatureIs(game.Layer6Ability, func(c *game.Characteristic) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "menace")
			}),
			SetAttachedBasePT(5, 5),
		},
		Activated: Reconfigure("{2}{B}{B}"),
	})
}
