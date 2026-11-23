package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jared Carthalion — Legendary Planeswalker — Jared {W}{U}{B}{R}{G}, loyalty
// 5 (Dominaria United, S82 DMU batch 15):
//
//	"+1: Create a 3/3 Kavu creature token with trample that's all colors.
//	 −3: Choose up to two target creatures. For each of them, put a number of
//	 +1/+1 counters on it equal to the number of colors it is.
//	 −6: Return target multicolored card from your graveyard to your hand. If
//	 that card was all colors, draw a card and create two Treasure tokens.
//	 Jared Carthalion can be your commander."
//
// The +1 token is the printed all-colors Kavu. The −3 counts each target's
// layered colors as it resolves, each target judged on its own (CR 608.2b); a
// colorless one gets none. The −6 reads the card's colors before it moves,
// and the draw and Treasures follow only when it was all five and really
// went to the hand. "Can be your commander" is deck-building, read from the
// oracle text by the deck validator.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f9485c76-122d-4faf-9194-b84ed9c41378",
		Name:            "Jared Carthalion",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Create a 3/3 Kavu creature token with trample that's all colors.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Template: effects.TokenCard(jaredCarthalionKavu), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "−3: Choose up to two target creatures. For each of them, put a number of +1/+1 counters on it equal to the number of colors it is.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("up to two target creatures").WithCount(0, 2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						c, ok := g.LookupCardForEffect(t.ID)
						if !ok {
							continue
						}
						n := effects.ColorsOf(c)
						if n <= 0 {
							continue
						}
						if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: n}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label:   "−6: Return target multicolored card from your graveyard to your hand. If that card was all colors, draw a card and create two Treasure tokens.",
				Cost:    effects.LoyaltyCost(-6),
				Targets: effects.TargetCardInGraveyard("target multicolored card in your graveyard", effects.YouOwn(), effects.Multicolored()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					c, ok := g.LookupCardForEffect(id)
					if !ok {
						return nil
					}
					all := effects.IsAllColors(c)
					if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
						return err
					}
					if !all {
						return nil
					}
					if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
						return err
					}
					return effects.CreateToken{Template: effects.TreasureToken(), N: 2}.Apply(ctx)
				},
			},
		},
	})
}

// jaredCarthalionKavu is the printed 3/3 all-colors Kavu with trample.
const jaredCarthalionKavu = "3/3 white and blue and black and red and green Kavu with trample"
