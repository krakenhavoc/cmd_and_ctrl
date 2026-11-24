package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stormtide Leviathan — Creature — Leviathan {5}{U}{U}{U}, 8/8 (Commander
// Legends):
//
//	"Islandwalk (This creature can't be blocked as long as defending player
//	 controls an Island.)
//	 All lands are Islands in addition to their other types.
//	 Creatures without flying or islandwalk can't attack."
//
// Islandwalk is the engine's keyword. The second line is Urborg's layer-4
// static over every land, whoever controls it (each keeps its other types
// and abilities, and gains the intrinsic {T}: Add {U}, CR 305.6). The third
// is CreaturesWithoutAnyCantAttack: the restriction is folded in after the
// layers, so flying or islandwalk granted at any timestamp lets a creature
// attack, and the Leviathan itself, which has islandwalk, may.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "131a7475-0b9e-46eb-9ade-fef1b9820184",
		Name:            "Stormtide Leviathan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		Static: []game.StaticAbility{
			{
				Label: "All lands are Islands in addition to their other types.",
				Layer: game.Layer4Type,
				AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
					return target.IsLand()
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					for _, st := range c.Subtypes {
						if st == "Island" {
							return
						}
					}
					c.Subtypes = append(c.Subtypes, "Island")
				},
			},
			effects.CreaturesWithoutAnyCantAttack("Creatures without flying or islandwalk can't attack.", "flying", "islandwalk"),
		},
	})
}
