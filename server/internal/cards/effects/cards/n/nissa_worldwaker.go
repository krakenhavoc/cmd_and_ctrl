package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nissa, Worldwaker — Legendary Planeswalker — Nissa {3}{G}{G}, loyalty 3
// (Magic 2015, #5556):
//
//	"+1: Target land you control becomes a 4/4 Elemental creature with
//	     trample. It's still a land.
//	 +1: Untap up to four target Forests.
//	 −7: Search your library for any number of basic land cards, put them
//	     onto the battlefield, then shuffle. Those lands become 4/4 Elemental
//	     creatures with trample. They're still lands."
//
// The first +1 is Elemental Uprising's animation with no duration (CR 611.2a):
// the land keeps its types and abilities and stays animated for good. The
// second untaps each Forest still legal as it resolves. The −7 is a search
// for any number (SearchLibrary.Unbounded, a real choice from none to all),
// then the shuffle, and its keyed tail animates exactly the lands that were
// found and entered (S82OldB358BecomeElementalLand).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "468f9b8e-4423-4131-8611-d34b0aa3d96f",
		Name:            "Nissa, Worldwaker",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Target land you control becomes a 4/4 Elemental creature with trample. It's still a land.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					land := effects.FirstLegalBattlefieldTarget(ctx)
					if land == uuid.Nil {
						return nil
					}
					return effects.S82OldB358BecomeElementalLand(ctx, land, "Nissa, Worldwaker — the land becomes a 4/4 Elemental creature with trample")
				},
			},
			{
				Label:   "+1: Untap up to four target Forests.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetPermanent("up to four target Forests", effects.Land(), effects.HasSubtype("Forest")).WithCount(0, 4),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.UntapEach(ctx, effects.LegalTargetCardIDs(ctx))
				},
			},
			{
				Label: "−7: Search your library for any number of basic land cards, put them onto the battlefield, then shuffle. Those lands become 4/4 Elemental creatures with trample. They're still lands.",
				Cost:  effects.LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.SearchLibrary{
						Player:    ctx.Controller(),
						Predicate: func(c game.Card) bool { return c.IsLand() && c.HasSupertype("basic") },
						Dest:      game.ZoneBattlefield,
						Unbounded: true,
						Shuffle:   true,
						Source:    ctx.Source(),
						Reason:    "Nissa, Worldwaker — any number of basic land cards",
						ThenTail:  ctx.ItemTail(nissaWorldwakerAnimate),
					}.Apply(ctx)
				},
			},
		},
	})
}

// nissaWorldwakerAnimate is the −7's "those lands become 4/4 Elemental
// creatures with trample": each land found that is on the battlefield. An
// on-disk key.
var nissaWorldwakerAnimate = effects.CardTail("then/nissa-worldwaker-animate", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := effects.S82OldB358BecomeElementalLand(ctx, id, "Nissa, Worldwaker — the land becomes a 4/4 Elemental creature with trample"); err != nil {
			return err
		}
	}
	return nil
})
