package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Canoptek Wraith — Artifact Creature — Wraith {3}, 2/1 (Warhammer 40,000
// Commander, S82 GAP2, #5948, #6090):
//
//	"Wraith Form — This creature can't be blocked.
//	 Transdimensional Scout — When this creature deals combat damage to a
//	 player, you may pay {3} and sacrifice it. If you do, choose a land you
//	 control. Then search your library for up to two basic land cards
//	 which have the same name as the chosen land, put them onto the
//	 battlefield tapped, then shuffle."
//
// As the trigger resolves, the {3} is offered only while the Wraith is
// still the permanent that dealt the damage, since it must also be
// sacrificed (CR 400.7). Paying sacrifices it, and only a sacrifice that
// happened goes on ("if you do"). Then the controller chooses a land they
// control (not a target, so any land); the search takes up to two basic
// land cards with that land's name, read as it was chosen, which enter
// tapped, then the library is shuffled. A player with no land, or who chose
// a nonbasic land whose name no basic land card has, still searches and
// finds nothing (CR 701.23b), and the library is still shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61849685-b407-4ed0-aafc-ec3a003200b9",
		Name:         "Canoptek Wraith",
		Completeness: effects.CompletenessFull,
		// Ramp (ADR 0126): up to two lands, tapped.
		Purpose: game.Purpose{Lands: 2},
		Static:  []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(
				"Canoptek Wraith — Transdimensional Scout: you may pay {3} and sacrifice it to search for up to two basic lands named like a land you control",
				canoptekWraithOffer),
		},
	})
}

const canoptekWraithName = "Canoptek Wraith"

// canoptekWraithOffer asks for the {3}, while the Wraith is still here to
// be sacrificed.
func canoptekWraithOffer(g *game.Game, item *game.StackItem) error {
	if !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{3}",
		Question: canoptekWraithName + " — pay {3} and sacrifice it to search for up to two basic lands with the name of a land you control?",
		Then:     canoptekWraithPaid,
	}.Apply(effects.NewContext(g, item))
}

// canoptekWraithPaid sacrifices the Wraith, and goes on only if it went.
// An on-disk key.
var canoptekWraithPaid = effects.PayUnlessBranches("pay-unless/canoptek-wraith-sacrifice",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
			return nil
		}
		return effects.SacrificePermanent{Target: ctx.Item.SourceCardID, ThenTail: ctx.ItemTail(canoptekWraithSacrificed)}.Apply(ctx)
	}, nil)

// canoptekWraithSacrificed is "if you do": choose a land you control. An
// on-disk key.
var canoptekWraithSacrificed = effects.IfItWentTail("canoptek-wraith/sacrificed", func(ctx *effects.Context) error {
	you := ctx.Controller()
	lands := effects.PermanentsControlledByMatching(ctx.Game, you, func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsLand() })
	if len(lands) == 0 {
		return canoptekWraithSearch(ctx, "")
	}
	return effects.ChoosePermanents{
		Question: canoptekWraithName + " — choose a land you control",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			ids := effects.PermanentsControlledByMatching(g, of, func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsLand() })
			if len(ids) == 0 {
				return nil, 0, 0
			}
			return ids, 1, 1
		},
		Keyed: canoptekWraithLandChosen,
	}.Apply(ctx)
})

// canoptekWraithLandChosen searches for basics with the chosen land's
// name. An on-disk key.
var canoptekWraithLandChosen = effects.RegisterPermanentsPick("prompt-run/canoptek-wraith-land", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		name := ""
		if cards := picked.Cards(); len(cards) > 0 {
			if c, ok := ctx.Game.LookupCardForEffect(cards[0]); ok {
				name = c.Name
			}
		}
		return canoptekWraithSearch(ctx, name)
	},
})

// canoptekWraithSearch is "search your library for up to two basic land
// cards which have the same name as the chosen land, put them onto the
// battlefield tapped, then shuffle". An empty name finds nothing.
func canoptekWraithSearch(ctx *effects.Context, name string) error {
	return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player: ctx.Controller(),
		Source: ctx.Source(),
		Pred: func(c game.Card) bool {
			return name != "" && c.IsLand() && c.HasSupertype("Basic") && c.Name == name
		},
		Dest:          game.ZoneBattlefield,
		Limit:         2,
		TappedOnEntry: true,
		Shuffle:       true,
		Reason:        canoptekWraithName + " — search for up to two basic land cards with the chosen land's name",
	})
}
