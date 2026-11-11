package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaya, Geist Hunter — Legendary Planeswalker — Kaya {1}{W}{B}, loyalty 3:
//
//	"+1: Creatures you control gain deathtouch until end of turn. Put a
//	 +1/+1 counter on up to one target creature token you control.
//	 −2: Until end of turn, if one or more tokens would be created under
//	 your control, twice that many of those tokens are created instead.
//	 −6: Exile all cards from all graveyards, then create a 1/1 white
//	 Spirit creature token with flying for each card exiled this way."
//
// The +1 locks in the creatures you control as it resolves (CR 611.2c) and
// puts the counter on the optional token target if it is still legal. The
// −2 is a turn-scoped token doubler (game.ModMultiplyTokens, CR 614.1a,
// 701.7b), so the −6's Spirits made on a later activation the same turn
// would be doubled, and a Doubling Season beside it multiplies as well. The
// −6 exiles every card in every graveyard as one batch and the Spirits
// follow the exile, one for each card that actually landed in exile (a
// commander that went to the command zone instead is not counted).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e0edb2fc-2533-407c-85e5-4337f4233f05",
		Name:            "Kaya, Geist Hunter",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Creatures you control gain deathtouch until end of turn. Put a +1/+1 counter on up to one target creature token you control.",
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("up to one target creature token you control", YouControl(), IsTokenPredicate()).WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (GrantKeywordUntilEOT{Match: And(Creature(), YouControl()), Keywords: []string{"deathtouch"}, Label: "Kaya, Geist Hunter — deathtouch"}).Apply(ctx); err != nil {
						return err
					}
					return PlusOneCounterOnChosenTarget(g, item)
				},
			},
			{
				Label: "−2: Until end of turn, if one or more tokens would be created under your control, twice that many of those tokens are created instead.",
				Cost:  LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					g.MultiplyTokensThisTurnForEffect(item.SourceCardID, item.Controller, 2, "Kaya, Geist Hunter — twice that many tokens")
					return nil
				},
			},
			{
				Label: "−6: Exile all cards from all graveyards, then create a 1/1 white Spirit creature token with flying for each card exiled this way.",
				Cost:  LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ids := allGraveyardsCardIDs(ctx, func(game.Card) bool { return true })
					return g.ExileCardsKeyedForEffect(ids, ctx.Tail(kayaGeistHunterSpirits))
				},
			},
		},
	})
}

// kayaGeistHunterSpirits is the −6's "then create a Spirit for each card
// exiled this way". An on-disk key: never renamed, never reused.
var kayaGeistHunterSpirits = CardTail("kaya-geist-hunter/spirits", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: len(o.IDs)}.Apply(ctx)
})
