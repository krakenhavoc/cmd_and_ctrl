package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mu Yanling, Sky Dancer — Legendary Planeswalker — Yanling {1}{U}{U},
// loyalty 2 (Foundations Jumpstart, #5929):
//
//	"+2: Until your next turn, up to one target creature gets -2/-0 and loses
//	 flying.
//	 −3: Create a 4/4 blue Elemental Bird creature token with flying.
//	 −8: You get an emblem with "Islands you control have '{T}: Draw a
//	 card.'""
//
// The +2 is a scoped effect on the one target, lasting until the activator's
// next turn begins (CR 611.2): -2/-0 in layer 7 and the loss of flying in
// layer 6. The emblem is Liliana of the Dark Realms' shape: a layer-6 grant to
// the Islands its owner controls, the granted ability an ordinary activated
// ability (drawing is not a mana ability, CR 605.1a), so it uses the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1e825298-7c99-40cc-81c9-280cc7ed98d3",
		Name:            "Mu Yanling, Sky Dancer",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 2,
		Grants: []effects.AbilityGrant{{
			Key: muYanlingIslandDraw,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			}},
			Text: "{T}: Draw a card.",
		}},
		Emblem: &effects.EmblemSpec{
			Label: "Mu Yanling, Sky Dancer emblem",
			Text:  "Islands you control have \"{T}: Draw a card.\"",
			Static: []game.StaticAbility{
				effects.GrantAbilities(muYanlingIslandsYouControl, muYanlingIslandDraw),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Until your next turn, up to one target creature gets -2/-0 and loses flying.",
				Cost:    effects.LoyaltyCost(2),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.ScopedEffectFor{
						Target:   t.ID,
						Mods:     []game.Mod{game.ModifyPTMod(-2, 0), game.RemoveKeywordsMod("flying")},
						Duration: effects.DurationUntilYourNextTurn(ctx, item.Controller),
						Label:    "Mu Yanling, Sky Dancer — -2/-0 and loses flying until your next turn",
					}.Apply(ctx)
				},
			},
			{
				Label:   "−3: Create a 4/4 blue Elemental Bird creature token with flying.",
				Cost:    effects.LoyaltyCost(-3),
				Purpose: game.Purpose{Tokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 blue Elemental Bird with flying"), N: 1}),
			},
			{
				Label:  "−8: You get an emblem with \"Islands you control have '{T}: Draw a card.'\"",
				Cost:   effects.LoyaltyCost(-8),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

// muYanlingIslandDraw is the granted bundle's key. An on-disk key.
const muYanlingIslandDraw = "mu-yanling-sky-dancer/island-draw"

// muYanlingIslandsYouControl is "Islands you control", read off the effective
// subtypes so an Island a layer-4 effect made counts (the emblem's controller
// is its owner, CR 114.2).
func muYanlingIslandsYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsLand() && target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Island"))
}
