package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reason // Believe — Sorcery // Sorcery with aftermath (Hour of Devastation):
//
//	Reason {U}: "Scry 3."
//	Believe {4}{G}: "Aftermath (Cast this spell only from your graveyard.
//	 Then exile it.) Look at the top card of your library. You may put it onto
//	 the battlefield if it's a creature card. If you don't, put it into your
//	 hand."
//
// The put is optional and only offered for a creature card; whichever way it
// is not put onto the battlefield, the card goes to your hand (Rulik Mons's
// look-then-put shape).
func init() {
	const oracle = "e2b152ea-82d5-465e-8e13-3d62aacd6692"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Reason",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 3},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.Scry{N: 3}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Believe",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PutFromLibraryOntoBattlefield{
				Player:   ctx.Controller(),
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(ctx.Controller(), 1),
				Match:    effects.Creature(),
				Max:      1,
				Optional: true,
				Label:    "Believe — you may put the creature card onto the battlefield",
				ThenTail: believeRestToHand,
			}.Apply(ctx)
		},
	})
}

var believeRestToHand = effects.PutFromLibraryTail("reason-believe/rest-to-hand",
	func(ctx *effects.Context, res effects.PutFromLibraryResult) error {
		if len(res.Entered) > 0 || len(res.Rest) == 0 {
			return nil
		}
		ctx.Game.TakeFromLibraryToHandForEffect(res.Player, res.Rest)
		return nil
	})
