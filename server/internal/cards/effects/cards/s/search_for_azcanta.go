package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Search for Azcanta // Azcanta, the Sunken Ruin — Legendary Enchantment
// {1}{U} // Legendary Land (Secret Lair Drop, #5462):
//
//	Front: "At the beginning of your upkeep, surveil 1. Then if you have
//	        seven or more cards in your graveyard, you may transform Search
//	        for Azcanta."
//	Back:  "{T}: Add {U}.
//	        {2}{U}, {T}: Look at the top four cards of your library. You may
//	        reveal a noncreature, nonland card from among them and put it into
//	        your hand. Put the rest on the bottom of your library in any order."
//
// The surveil runs first and the rest waits for the cards to be filed
// (Surveil.ThenTail), so the graveyard counted is the one the surveil left.
// This is not an intervening "if": the count is read once, after the surveil,
// and then the player is asked. The transform is in place (CR 701.27, 712.18).
// The back's look is Avengers Tower's TakeFromLibraryToHand.
//
// No simplification.
func init() {
	const oracle = "f74c4d96-bc4a-4d32-9519-a753d192144e"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Search for Azcanta",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Search for Azcanta — surveil 1, then if you have seven or more cards in your graveyard, you may transform Search for Azcanta",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.Surveil{N: 1, ThenTail: ctx.ItemTail(searchForAzcantaAfterSurveil)}.Apply(ctx)
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:      oracle + "#1",
		Name:          "Azcanta, the Sunken Ruin",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{U}", Label: "{T}: Add {U}."}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, {T}: Look at the top four cards of your library. You may reveal a noncreature, nonland card from among them and put it into your hand. Put the rest on the bottom of your library in any order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				player := ctx.Controller()
				return effects.TakeFromLibraryToHand{
					Player:   player,
					Cards:    g.LookAtTopOfLibraryForEffect(player, 4),
					Match:    effects.And(effects.Noncreature(), effects.Nonland()),
					Max:      1,
					Optional: true,
					Label:    "Azcanta, the Sunken Ruin — you may reveal a noncreature, nonland card and put it into your hand",
					Key:      effects.TakeRevealRestBottomAny,
				}.Apply(ctx)
			},
		}},
	})
}

// searchForAzcantaAfterSurveil asks "you may transform" when seven cards are
// in the graveyard. An on-disk key: never renamed, never reused.
var searchForAzcantaAfterSurveil = effects.CardTail("search-for-azcanta/after-surveil", func(ctx *effects.Context, _ game.TailOutcome) error {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.Graveyard.Size() < 7 {
		return nil
	}
	return effects.MayChoice{
		Question: "Search for Azcanta — transform it?",
		YesLabel: "Transform",
		NoLabel:  "Don't",
		Then:     searchForAzcantaTransform,
	}.Apply(ctx)
})

// searchForAzcantaTransform is the "yes". An on-disk key: never renamed,
// never reused.
var searchForAzcantaTransform = effects.MayYes("confirm/search-for-azcanta-transform", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.TransformThis{}.Apply(ctx)
})
