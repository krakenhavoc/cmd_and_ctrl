package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nissa, Nature's Artisan — Legendary Planeswalker — Nissa {4}{G}{G}, starting
// loyalty 5 (Kaladesh, #5539):
//
//	"+3: You gain 3 life.
//	 −4: Reveal the top two cards of your library. Put all land cards from among
//	 them onto the battlefield and the rest into your hand.
//	 −12: Creatures you control get +5/+5 and gain trample until end of turn."
//
// The −4 is Wrenn and Seven's reveal (no choice: the cards are revealed, not
// looked at) with a keyed put of every revealed land onto the battlefield
// untapped (Zimone's Experiment's put) and, behind it as its tail, the rest of
// the two cards, those still in the library, into the hand: so a land that a
// replacement effect kept off the battlefield goes to the hand with the rest.
// The −12 locks the set of creatures as it resolves (CR 611.2c) and gives each
// +5/+5 and trample until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5e2aff26-c9f0-4e4b-bc5e-b2cb46e04293",
		Name:            "Nissa, Nature's Artisan",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+3: You gain 3 life.",
				Cost:   effects.LoyaltyCost(3),
				Effect: effects.Do(effects.GainLife{Amount: 3}),
			},
			{
				Label:  "−4: Reveal the top two cards of your library. Put all land cards from among them onto the battlefield and the rest into your hand.",
				Cost:   effects.LoyaltyCost(-4),
				Effect: nissaNaturesArtisanMinusFour,
			},
			{
				Label:  "−12: Creatures you control get +5/+5 and gain trample until end of turn.",
				Cost:   effects.LoyaltyCost(-12),
				Effect: nissaNaturesArtisanMinusTwelve,
			},
		},
	})
}

// nissaNaturesArtisanMinusFour reveals the top two cards, puts every land among
// them onto the battlefield and sends the rest to the hand.
func nissaNaturesArtisanMinusFour(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	revealed := g.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), 2,
		"Nissa, Nature's Artisan — reveal the top two cards of your library")
	return effects.PutFromLibraryOntoBattlefield{
		Player:   ctx.Controller(),
		Cards:    revealed,
		Match:    effects.Land(),
		All:      true,
		ThenTail: nissaNaturesArtisanRestToHand,
	}.Apply(ctx)
}

// nissaNaturesArtisanRestToHand is "the rest into your hand": every revealed
// card still in the library. An on-disk key.
var nissaNaturesArtisanRestToHand = effects.PutFromLibraryTail("nissa-natures-artisan/rest-to-hand",
	func(ctx *effects.Context, res effects.PutFromLibraryResult) error {
		ctx.Game.TakeFromLibraryToHandForEffect(res.Player, res.Rest)
		return nil
	})

// nissaNaturesArtisanMinusTwelve is +5/+5 and trample for each creature its
// controller controls as the ability resolves.
func nissaNaturesArtisanMinusTwelve(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	mine := effects.And(effects.Creature(), effects.YouControl())
	const label = "Nissa, Nature's Artisan — creatures you control get +5/+5 and gain trample"
	if err := (effects.BoostUntilEOT{Match: mine, Power: 5, Toughness: 5, Label: label}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Match: mine, Keywords: []string{"trample"}, Label: label}.Apply(ctx)
}
