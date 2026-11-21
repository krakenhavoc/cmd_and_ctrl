package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rukarumel, Biologist — Legendary Creature — Human Wizard
// {W}{U}{B}{R}{G}, 3/3 (Commander Masters):
//
//	"As Rukarumel enters, choose a creature type.
//	 Slivers you control and nontoken creatures you control are the chosen
//	 type in addition to their other creature types. The same is true for
//	 creature spells you control and creature cards you own that aren't on
//	 the battlefield.
//	 {3}, {T}: Create a 1/1 colorless Sliver creature token."
//
// Leyline of Transformation's three layer-4 statics (CR 613.1d): the
// battlefield one is narrowed to your Slivers and your nontoken creatures,
// so a token that is not a Sliver does not get the type; the spell and
// card ones are the same sentence as the Leyline's. "In addition": the
// type is appended and the others are kept. Before a type is chosen
// nothing is added. The type is the stored "as this enters" choice.
//
// No simplification.
func init() {
	grant := func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
		if source.NamedTribe != "" && !effects.HasFold(c.Subtypes, source.NamedTribe) {
			c.Subtypes = append(c.Subtypes, source.NamedTribe)
		}
	}
	effects.Register(effects.Spec{
		OracleID:     "84391c32-5bb7-4c36-be50-bbb5f8732156",
		Name:         "Rukarumel, Biologist",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCreatureTypeAsEnters("Rukarumel, Biologist — choose a creature type"),
		Static: []game.StaticAbility{
			{
				Label: "Slivers you control and nontoken creatures you control are the chosen type in addition to their other creature types.",
				Layer: game.Layer4Type,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && target.IsCreature() &&
						(!target.IsToken() || target.HasSubtype(source.TextSubtype("Sliver")))
				},
				Apply: grant,
			},
			{
				Label:        "Creature spells you control are the chosen type in addition to their other creature types.",
				Layer:        game.Layer4Type,
				AffectsZones: effects.SpellsOnTheStack(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && target.IsCreature()
				},
				Apply: grant,
			},
			{
				Label:        "Creature cards you own that aren't on the battlefield are the chosen type in addition to their other creature types.",
				Layer:        game.Layer4Type,
				AffectsZones: effects.CardsOffTheBattlefield(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Owner == source.Controller && target.IsCreature()
				},
				Apply: grant,
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Create a 1/1 colorless Sliver creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Sliver"), N: 1}),
		}},
	})
}
