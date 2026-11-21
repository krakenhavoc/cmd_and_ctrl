package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Highway Robbery — Sorcery {1}{R} (Outlaws of Thunder Junction):
//
//	"You may discard a card or sacrifice a land. If you do, draw two cards.
//	 Plot {1}{R}"
//
// One choice at resolution among three branches: do neither, discard, or
// sacrifice a land (an option you can't take is not offered, CR 608.2d).
// The chosen cost is paid, and only then are the two cards drawn: the
// discard rides the keyed discard-then-draw continuation, the sacrifice a
// keyed tail, so a table waiting on either prompt is a restore point.
//
// No simplification.
var highwayRobberyChoice = effects.OptionPickThen("option-pick/highway-robbery-cost", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil {
		return nil
	}
	player := ctx.Controller()
	switch r.Option.Label {
	case highwayRobberyDiscard:
		return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player: player, Source: ctx.Source(), N: 1,
			Question: "Highway Robbery — discard a card",
		}, effects.DrawTwoIfDiscarded.Then, nil)
	case highwayRobberySacrifice:
		return effects.SacrificeChoice{
			Player:     player,
			Candidates: effects.LandsControlledByPlayer(ctx.Game, player),
			Question:   "Highway Robbery — sacrifice a land",
			ThenTail:   ctx.Tail(highwayRobberyDraw),
		}.Apply(ctx)
	}
	return nil
})

const (
	highwayRobberyDiscard   = "Discard a card"
	highwayRobberySacrifice = "Sacrifice a land"
)

// highwayRobberyDraw is "draw two cards" once the land has been sacrificed.
var highwayRobberyDraw = effects.CardTail("then/highway-robbery-draw-two", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.DrawCards{N: 2}.Apply(ctx)
})

func init() {
	effects.Register(effects.Spec{
		OracleID:       "df020a47-1784-479b-9882-be77b7f30950",
		Name:           "Highway Robbery",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Draws: 2},
		SpecialActions: []game.SpecialAction{effects.Plot("{1}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			opts := []game.ChoiceOption{{Label: "Do neither"}}
			if p := ctx.PlayerByID(player); p != nil && p.Hand.Size() > 0 {
				opts = append(opts, game.ChoiceOption{Label: highwayRobberyDiscard})
			}
			if len(effects.LandsControlledByPlayer(ctx.Game, player)) > 0 {
				opts = append(opts, game.ChoiceOption{Label: highwayRobberySacrifice})
			}
			if len(opts) == 1 {
				return nil
			}
			return effects.PickOption{
				Player:   player,
				Question: "Highway Robbery — discard a card or sacrifice a land to draw two cards?",
				Options:  opts,
				ThenKey:  highwayRobberyChoice,
			}.Apply(ctx)
		},
	})
}
