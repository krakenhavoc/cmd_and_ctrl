package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rootpath Purifier — Creature — Elf Druid {3}{G}, 3/4 (The Brothers' War):
//
//	"Lands you control and land cards in your library are basic."
//
// Two layer 4 statics (CR 613.1d), as Maskwood Nexus splits its two
// halves: lands its controller controls on the battlefield gain the Basic
// supertype, and land cards in its controller's library gain it through
// the zone layer step (AffectsZones, a library card's characteristic is
// re-derived each pass), so a "basic land card" search of that library
// (IsBasicLand reads the effective supertypes) finds a nonbasic land.
// Basic is a supertype, not a land type, so the lands gain no mana
// ability and keep every other ability and type. Both end the moment
// Rootpath Purifier leaves the battlefield or loses its abilities.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "442d06e6-8cf6-4980-b2bf-73cfb4d3a468",
		Name:         "Rootpath Purifier",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label: "Lands you control are basic.",
				Layer: game.Layer4Type,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && target.IsLand()
				},
				Apply: makeBasic,
			},
			{
				Label:        "Land cards in your library are basic.",
				Layer:        game.Layer4Type,
				AffectsZones: []game.ZoneKind{game.ZoneLibrary},
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Owner == source.Controller && target.IsLand()
				},
				Apply: makeBasic,
			},
		},
	})
}

// makeBasic adds the Basic supertype to the characteristic in progress.
func makeBasic(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
	for _, s := range c.Supertypes {
		if s == "Basic" {
			return
		}
	}
	c.Supertypes = append(c.Supertypes, "Basic")
}
