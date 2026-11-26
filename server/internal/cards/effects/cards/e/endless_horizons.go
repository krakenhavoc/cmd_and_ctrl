package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Endless Horizons — Enchantment {3}{W}:
//
//	"When this enchantment enters, search your library for any number of Plains
//	 cards, exile them, then shuffle.
//	 At the beginning of your upkeep, you may put a card you own exiled with
//	 this enchantment into your hand."
//
// Pit of Offerings' "exiled with" record: the cards the first trigger's
// resolution exiled, read back through B27ExiledWith under that trigger's
// label, so they are remembered while this object stays on the battlefield and
// forgotten if it leaves and returns (CR 400.7). The search is an unbounded
// search (a real choice from none to every Plains card). The upkeep trigger
// offers the cards still in exile as a choose-cards prompt with a floor of zero
// ("you may"), and puts the one picked into its owner's hand (CR 400.3).
const endlessHorizonsSearchLabel = "Endless Horizons — search your library for any number of Plains cards, exile them, then shuffle"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca945c1e-efa3-4bb2-bf5a-379b93409855",
		Name:         "Endless Horizons",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(endlessHorizonsSearchLabel, func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsLandWithSubtype("Plains"),
					Dest:      game.ZoneExile,
					Unbounded: true,
					Shuffle:   true,
					Source:    item.SourceCardID,
					Reason:    "Endless Horizons — exile any number of Plains cards",
				}.Apply(effects.NewContext(g, item))
			}),
			effects.AtYourUpkeep("Endless Horizons — you may put a card you own exiled with this enchantment into your hand", endlessHorizonsReturn),
		},
	})
}

func endlessHorizonsReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var candidates []uuid.UUID
	for _, id := range effects.B27ExiledWith(g, item.SourceCardID, endlessHorizonsSearchLabel) {
		if c, ok := g.LookupCardForEffect(id); ok && c.Owner == item.Controller {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Endless Horizons — put a card you own exiled with this enchantment into your hand",
		Cards:      candidates,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneExile,
		ThenKey:    endlessHorizonsTake,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}

var endlessHorizonsTake = effects.AskedPick("choose-cards/endless-horizons-take", func(ctx *effects.Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := effects.NothingIfGone(ctx.Game.BounceToHandForEffect(id)); err != nil {
			return err
		}
	}
	return nil
})
