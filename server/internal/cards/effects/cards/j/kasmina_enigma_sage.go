package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kasmina, Enigma Sage — Legendary Planeswalker — Kasmina {1}{G}{U}, loyalty 2
// (Strixhaven: School of Mages, #5502):
//
//	"Each other planeswalker you control has the loyalty abilities of
//	 Kasmina.
//	 +2: Scry 1.
//	 −X: Create a 0/0 green and blue Fractal creature token. Put X +1/+1
//	 counters on it.
//	 −8: Search your library for an instant or sorcery card that shares a
//	 color with this planeswalker, exile that card, then shuffle. You may cast
//	 that card without paying its mana cost."
//
// # The static
//
// Nicol Bolas, Dragon-God's loyalty-only borrow (BorrowExceptNonLoyalty), run
// the other way: the recipients are the other planeswalkers its controller
// controls (GrantActivatedAbilitiesOf) and the one lender is Kasmina. A lent
// row is activated by the recipient, so it pays that planeswalker's loyalty,
// counts against its one loyalty activation each turn (CR 606.3), and every
// "this planeswalker" in it is the recipient.
//
// # The abilities
//
// −X is Ugin's X loyalty cost with Snarl Song's Fractal; the counters arrive on
// the token as it is created. The −8 is Bring to Light's exile-and-free-cast
// tail: the search predicate compares a candidate's colors with the colors of
// the planeswalker whose ability this is as it resolves (Kasmina, or the
// planeswalker she lent it to); a colorless planeswalker finds nothing. The
// card stays in exile when the cast is declined.
//
// No simplification.
func init() {
	const clause = "Each other planeswalker you control has the loyalty abilities of Kasmina."
	effects.Register(effects.Spec{
		OracleID:        "06425cb9-d157-4155-a727-78656df12458",
		Name:            "Kasmina, Enigma Sage",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 2,
		Static: []game.StaticAbility{effects.GrantActivatedAbilitiesOf(
			func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return source != nil && target.InstanceID != source.InstanceID &&
					target.IsPlaneswalker() && target.Controller == source.Controller
			},
			game.AbilityBorrow{
				Label: clause,
				Lenders: func(_ *game.Game, source *game.Card) []*game.Card {
					if source == nil {
						return nil
					}
					return []*game.Card{source}
				},
				Except: game.BorrowExceptNonLoyalty,
			})},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Scry 1.",
				Cost:    effects.LoyaltyCost(2),
				Purpose: game.Purpose{Scry: 1},
				Effect:  effects.Do(effects.Scry{N: 1}),
			},
			{
				Label:   "−X: Create a 0/0 green and blue Fractal creature token. Put X +1/+1 counters on it.",
				Cost:    effects.LoyaltyMinusX(),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.CreateTokenAdvanced{
						Controller: item.Controller,
						N:          1,
						Spec: effects.TokenSpec{
							Template: effects.TokenCard("0/0 green and blue Fractal"),
							Counters: map[string]int{game.CounterPlusOne: ctx.X()},
						},
					}.Apply(ctx)
				},
			},
			{
				Label: "−8: Search your library for an instant or sorcery card that shares a color with this planeswalker, exile that card, then shuffle. You may cast that card without paying its mana cost.",
				Cost:  effects.LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					walker, _ := g.LookupCardForEffect(ctx.Source())
					return effects.SearchLibrary{
						Player: ctx.Controller(),
						Limit:  1,
						Dest:   game.ZoneExile,
						Predicate: func(c game.Card) bool {
							return (c.HasCardType("instant") || c.HasCardType("sorcery")) && effects.SharesAColorNow(c, walker)
						},
						Shuffle:  true,
						Reason:   "Kasmina, Enigma Sage — search for an instant or sorcery card that shares a color with this planeswalker",
						ThenTail: ctx.ItemTail(kasminaCast),
					}.Apply(ctx)
				},
			},
		},
	})
}

var kasminaCast = effects.CardTail("old-b051/kasmina-cast", func(ctx *effects.Context, o game.TailOutcome) error {
	effects.GrantFreeCasts(ctx.Game, ctx.Controller(), ctx.Source(), "Kasmina, Enigma Sage",
		game.TimingFlash, game.LapseStaysInExile, o.IDs)
	return nil
})
