package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Celestial Dawn — Enchantment {1}{W}{W} (Mirage, #5633):
//
//	"Lands you control are Plains.
//	 Nonland permanents you control are white. The same is true for spells
//	 you control and nonland cards you own that aren't on the battlefield.
//	 You may spend white mana as though it were mana of any color. You may
//	 spend other mana only as though it were colorless mana."
//
// Four layered facts and one spend rule:
//
//   - layer 4: each land its controller controls becomes a Plains, the static
//     form of CR 305.7 (SetsBasicLandType, Blood Moon's shape): its land types
//     are replaced, it loses the abilities its rules text gives it and taps
//     for {W} (CR 305.6);
//   - layer 5: nonland permanents you control are white, and, as Maskwood
//     Nexus does for creature types, two statics over objects off the
//     battlefield: spells you control (nonland ones) and nonland cards you own
//     in any other zone (#3862);
//   - the spend rule is Celestial Dawn's own seam (#6202): white mana pays
//     as any colour and every other mana pays only generic costs and {C}.
//
// No simplification.
func init() {
	isWhite := func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) { c.Colors = []string{"W"} }
	effects.Register(effects.Spec{
		OracleID:      "b4d27fad-9a94-4760-80a6-6a2948109864",
		Name:          "Celestial Dawn",
		Completeness:  effects.CompletenessFull,
		AnyColorSpend: effects.YouMaySpendColorAsAnyOthersOnlyColorless("W"),
		Static: []game.StaticAbility{
			effects.SetsBasicLandType(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsLand() && target.Controller == source.Controller
			}, nil, []string{"Plains"}),
			{
				Label: "Nonland permanents you control are white.",
				Layer: game.Layer5Color,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return !target.IsLand() && target.Controller == source.Controller
				},
				Apply: isWhite,
			},
			{
				Label:        "Spells you control are white.",
				Layer:        game.Layer5Color,
				AffectsZones: effects.SpellsOnTheStack(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return !target.IsLand() && target.Controller == source.Controller
				},
				Apply: isWhite,
			},
			{
				Label:        "Nonland cards you own that aren't on the battlefield are white.",
				Layer:        game.Layer5Color,
				AffectsZones: effects.CardsOffTheBattlefield(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return !target.IsLand() && target.Owner == source.Controller
				},
				Apply: isWhite,
			},
		},
	})
}
