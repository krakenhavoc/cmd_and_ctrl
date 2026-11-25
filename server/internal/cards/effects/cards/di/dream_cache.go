package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dream Cache — Sorcery {2}{U} (Commander 2018, S82 OLD batch 193, #5526):
//
//	"Draw three cards, then put two cards from your hand both on top of your
//	 library or both on the bottom of your library."
//
// Brainstorm's shape: the put-back is the draw's continuation (CR 608.2c), so
// it is asked after the three draws, even when a draw paused to ask a
// question. You pick the two cards (a hand of one or two cards puts what it
// has, CR 608.2), then say whether both go on top or both go on the bottom,
// then order them. All three prompts are restore points: the pick and the
// choice carry plain data under registered keys.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "81338d64-8fc3-41ac-afc8-e2a79df1159b",
		Name:         "Dream Cache",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.DrawNThenForEffect(item.Controller, 3, game.DrawThen{
				Ref: dreamCachePutBack, Player: item.Controller, Source: item.SourceCardID, N: 2,
			})
		},
	})
}

const (
	dreamCacheTop    = "Put both cards on top of your library"
	dreamCacheBottom = "Put both cards on the bottom of your library"
)

// dreamCachePutBack is the draw's continuation: pick the two cards, or skip
// the pick when every card in hand goes back. An on-disk key.
var dreamCachePutBack = game.RegisterDrawThen("old-b193/dream-cache-put-back", func(g *game.Game, d game.DrawThen) error {
	hand := effects.AllHandCardIDs(g, d.Player)
	if len(hand) == 0 || d.N <= 0 {
		return nil
	}
	floor := min(d.N, len(hand))
	if floor == len(hand) {
		return dreamCacheAskSide(g, d.Player, d.Source, hand)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  d.Player,
		Source:   d.Source,
		Question: "Dream Cache — put two cards from your hand on top or on the bottom of your library",
		Cards:    hand,
		Min:      floor,
		Max:      floor,
		LiveZone: game.ZoneHand,
		ThenKey:  dreamCachePicked,
	})
	return nil
})

// dreamCachePicked is the two cards picked. An on-disk key.
var dreamCachePicked = game.RegisterChooseCardsThen("choose-cards/old-b193-dream-cache-put-back", func(g *game.Game, r game.CardsChosen) error {
	return dreamCacheAskSide(g, r.Chooser, r.Source, r.Picked)
})

// dreamCacheAskSide asks "both on top or both on the bottom".
func dreamCacheAskSide(g *game.Game, player, source uuid.UUID, cards []uuid.UUID) error {
	g.QueueOptionPickForEffect(game.OptionPickPrompt{
		Chooser:  player,
		Source:   source,
		Question: "Dream Cache — put both cards on top of your library or both on the bottom?",
		Options:  []game.ChoiceOption{{Label: dreamCacheTop}, {Label: dreamCacheBottom}},
		ThenKey:  dreamCacheSide,
		Carry:    cards,
	})
	return nil
}

// dreamCacheSide puts the carried cards on the chosen side, in the order the
// player picks. An on-disk key.
var dreamCacheSide = game.RegisterOptionPickThen("option-pick/old-b193-dream-cache-side", func(g *game.Game, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) == 0 {
		return nil
	}
	placement := game.LibraryPlaceTop
	reason := "Dream Cache — in what order? (the first is on top)"
	if r.Option.Label == dreamCacheBottom {
		placement = game.LibraryPlaceBottom
		reason = "Dream Cache — in what order? (the first is closest to the top of the bottom pile)"
	}
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   r.Chooser,
		Source:    r.Source,
		Cards:     r.Carry,
		From:      game.ZoneHand,
		Placement: placement,
		Reason:    reason,
	})
})
