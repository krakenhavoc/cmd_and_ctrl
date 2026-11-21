package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ignis Scientia — Legendary Creature — Human Advisor {1}{G}{U}, 2/2 (Final
// Fantasy, #4462):
//
//	"When Ignis Scientia enters, look at the top six cards of your library.
//	 You may put a land card from among them onto the battlefield tapped.
//	 Put the rest on the bottom of your library in a random order.
//	 I've Come Up with a New Recipe! — {1}{G}{U}, {T}: Exile target card
//	 from a graveyard. If a creature card was exiled this way, create a Food
//	 token."
//
// The activated ability exiles the target card and, when it was a creature
// card (read as it was in the graveyard), makes a Food. If the target is
// illegal as it resolves the whole ability does nothing, so no Food.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "23320d55-80cb-4da3-a53c-8c75abdc8c8f",
		Name:         "Ignis Scientia",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ignis Scientia — look at the top six cards of your library; you may put a land card from among them onto the battlefield tapped; put the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					return ignisLookAtTopSix(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "I've Come Up with a New Recipe! — {1}{G}{U}, {T}: Exile target card from a graveyard. If a creature card was exiled this way, create a Food token.",
			Cost:    effects.Plus(effects.ManaCost("{1}{G}{U}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
			Effect:  ignisNewRecipe,
		}},
	})
}

// ignisLookAtTopSix is the enters trigger: look at the top six, a land card
// from among them may go onto the battlefield tapped, and the rest go on the
// bottom of the library in a random order.
func ignisLookAtTopSix(ctx *effects.Context) error {
	player := ctx.Controller()
	looked := ctx.Game.LookAtTopOfLibraryForEffect(player, 6)
	return effects.PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    looked,
		Match:    effects.Land(),
		Max:      1,
		Optional: true,
		Tapped:   true,
		Label:    "Ignis Scientia — you may put a land card from among them onto the battlefield tapped",
		ThenTail: effects.PutRestOnBottomInRandomOrderTail,
	}.Apply(ctx)
}

func ignisNewRecipe(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return effects.ExileThenIfItWas{
			Target: t.ID,
			Was:    effects.WasCreatureCard,
			Keyed:  ignisFoodIfCreature,
		}.Apply(ctx)
	}
	return nil
}

// ignisFoodIfCreature is "if a creature card was exiled this way, create a
// Food token", told the card reached exile. An on-disk key: never renamed.
var ignisFoodIfCreature = effects.IfItWasTail("if-it-was/ignis-scientia-food", func(ctx *effects.Context) error {
	return effects.CreateToken{Template: effects.FoodToken(), N: 1}.Apply(ctx)
}, nil)
