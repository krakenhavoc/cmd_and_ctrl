package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dune Chanter:
//
//	"Reach
//	 Lands you control and land cards you own that aren't on the
//	 battlefield are Deserts in addition to their other types.
//	 Lands you control have "{T}: Add one mana of any color."
//	 {T}: Mill two cards. You gain 1 life for each land card milled this
//	 way."
//
// Urborg's layer-4 static for the lands you control, plus Maskwood
// Nexus' zone-layer static for the land cards you own in a hand,
// library, graveyard, exile or the command zone (so Hazezon's Desert
// permission, a Desert search and Cataclysmic Prospecting's source
// kinds all see them). The mana grant is Chromatic Lantern's layer-6
// grant. The life is counted from the cards the mill put into the
// graveyard, by their card types.
//
// No simplification.
const duneChanterGrant = "dune-chanter/any-color"

func init() {
	addDesert := func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
		desert := source.TextSubtype("Desert")
		for _, st := range c.Subtypes {
			if st == desert {
				return
			}
		}
		c.Subtypes = append(c.Subtypes, desert)
	}
	Register(Spec{
		OracleID:        "7811b50d-af76-43cd-8753-43cf65fbefa3",
		Name:            "Dune Chanter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Grants:          []AbilityGrant{AnyColorManaGrant(duneChanterGrant)},
		Static: []game.StaticAbility{
			{
				Layer: game.Layer4Type,
				Label: "Lands you control are Deserts in addition to their other types.",
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsLand() && target.Controller == source.Controller
				},
				Apply: addDesert,
			},
			{
				Layer:        game.Layer4Type,
				Label:        "Land cards you own that aren't on the battlefield are Deserts in addition to their other types.",
				AffectsZones: CardsOffTheBattlefield(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Owner == source.Controller && target.IsLand()
				},
				Apply: addDesert,
			},
			Labelled(GrantAbilities(landsYouControl, duneChanterGrant), "Lands you control have \"{T}: Add one mana of any color.\""),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Mill two cards. You gain 1 life for each land card milled this way.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return MillToZone{N: 2, ThenTail: ctx.ItemTail(duneChanterMilled)}.Apply(ctx)
			},
		}},
	})
}

// duneChanterMilled is "You gain 1 life for each land card milled this
// way": the cards the mill put into the graveyard, counted by type.
var duneChanterMilled = CardTail("dune-chanter/after-mill", func(ctx *Context, o game.TailOutcome) error {
	n := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsLand() {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return GainLife{Player: ctx.Controller(), Amount: n}.Apply(ctx)
})
