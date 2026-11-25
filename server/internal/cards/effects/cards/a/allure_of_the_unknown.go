package cardsa

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Allure of the Unknown — Sorcery {3}{B}{R} (Theros Beyond Death, #5514):
//
//	"Reveal the top six cards of your library. An opponent exiles a nonland card
//	 from among them, then you put the rest into your hand. That opponent may cast
//	 the exiled card without paying its mana cost."
//
// In printed order, each step waiting for the one before it: the reveal (CR
// 701.20); the controller chooses an opponent (ChoosePlayer, as Fact or Fiction
// names its splitter); that opponent picks a nonland card from among the revealed
// ones (one choose_cards prompt, mandatory when there is one) and exiles it; the
// rest go to the controller's hand, lands included; then that opponent may cast the
// exiled card for free during the resolution (CR 608.2g, GrantFreeCasts with the
// cast left in exile if declined). With no nonland card among them nothing is
// exiled and everything goes to the hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "252d9b25-45a1-40ca-b563-4cdefc791a44",
		Name:         "Allure of the Unknown",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			revealed := ctx.Game.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), 6,
				"Allure of the Unknown — reveal the top six cards of your library")
			if len(revealed) == 0 {
				return nil
			}
			return effects.ChoosePlayer{
				Among:    effects.Opponents,
				Question: "Allure of the Unknown — choose an opponent to exile a nonland card from among the revealed cards",
				ThenKey:  allureChosenOpponent,
				Carry:    revealed,
			}.Apply(ctx)
		},
	})
}

// allureChosenOpponent asks the chosen opponent to exile a nonland card. Keyed:
// the revealed cards ride the prompt as its Carry. An on-disk key.
var allureChosenOpponent = effects.ChoosePlayerThen("option-pick/allure-of-the-unknown-opponent", func(ctx *effects.Context, opp uuid.UUID, revealed []uuid.UUID) error {
	var nonland []uuid.UUID
	for _, id := range revealed {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && !c.IsLand() {
			nonland = append(nonland, id)
		}
	}
	if opp == uuid.Nil || len(nonland) == 0 {
		return effects.FactOrFictionSettle(ctx.Game, revealed, nil)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    opp,
		FromPlayer: ctx.Controller(),
		Source:     ctx.Source(),
		Question:   "Allure of the Unknown — exile a nonland card from among the revealed cards",
		Cards:      nonland,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneLibrary,
		ThenKey:    allureExile,
		Carry:      append([]uuid.UUID{ctx.Controller()}, revealed...),
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
})

// allureExile exiles the opponent's pick, then (in its tail) puts the rest into the
// controller's hand and offers the free cast.
var allureExile = effects.AskedPick("choose-cards/allure-of-the-unknown-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Carry) < 1 {
		return nil
	}
	controller, revealed := r.Carry[0], r.Carry[1:]
	if len(r.Picked) != 1 {
		return effects.FactOrFictionSettle(ctx.Game, revealed, nil)
	}
	tail := ctx.ItemTail(allureAfterExile)
	tail.Player = r.Chooser
	tail.Carry = append([]uuid.UUID{controller}, revealed...)
	return ctx.Game.ExileCardsKeyedForEffect(r.Picked, tail)
})

// allureAfterExile puts the cards that were not exiled into the controller's hand
// and lets the opponent who exiled the card cast it free.
var allureAfterExile = effects.CardTail("exile-then/allure-of-the-unknown-rest", func(ctx *effects.Context, o game.TailOutcome) error {
	carry := o.Tail.Carry
	if len(carry) < 1 {
		return nil
	}
	opp, revealed := o.Tail.Player, carry[1:]
	var rest []uuid.UUID
	for _, id := range revealed {
		if !slices.Contains(o.IDs, id) {
			rest = append(rest, id)
		}
	}
	if err := effects.FactOrFictionSettle(ctx.Game, rest, nil); err != nil {
		return err
	}
	if len(o.IDs) == 0 || opp == uuid.Nil {
		return nil
	}
	effects.GrantFreeCasts(ctx.Game, opp, ctx.Source(), "Allure of the Unknown", game.TimingFlash, game.LapseStaysInExile, o.IDs)
	return nil
})
