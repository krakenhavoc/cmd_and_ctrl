package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mastermind's Acquisition — Sorcery {2}{B}{B}:
//
//	"Choose one —
//	 • Search your library for a card, put it into your hand, then
//	   shuffle.
//	 • Put a card you own from outside the game into your hand."
//
// The wish with no reveal and no "may" (ADR 0191): the second mode is a
// mandatory pick (TakeFromOutside.Required), so a player with a sideboard
// takes a card and cannot decline it, and a player with none takes
// nothing. The first mode is Demonic Tutor's search.
func init() {
	Register(Spec{
		OracleID:     "6aec5b1f-cea6-4ec3-9759-e4544da8c52b",
		Name:         "Mastermind's Acquisition",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Search your library for a card, put it into your hand, then shuffle."),
			Mode("Put a card you own from outside the game into your hand."),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				return SearchLibrary{
					Player:  ctx.Controller(),
					Dest:    game.ZoneHand,
					Limit:   1,
					Reveal:  false,
					Shuffle: true,
					Reason:  "Mastermind's Acquisition — search your library for a card",
				}.Apply(ctx)
			case ctx.HasMode(1):
				return TakeFromOutside{
					Pick:     mastermindsAcquisitionTake,
					Question: "Mastermind's Acquisition — put a card you own from outside the game into your hand",
					Required: true,
				}.Apply(ctx)
			}
			return nil
		},
	})
}

// mastermindsAcquisitionTake is the second mode's keyed pick (#4256),
// so a table waiting on it is a restore point. It does not reveal.
var mastermindsAcquisitionTake = registerOutsideTake("choose-cards/masterminds-acquisition-take", nil, false)
