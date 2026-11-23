package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Start the TARDIS — Sorcery {1}{U} (Doctor Who):
//
//	"Surveil 2, then draw a card. You may planeswalk. (To surveil 2, look at
//	 the top two cards of your library, then put any number of them into your
//	 graveyard and the rest on top of your library in any order.)
//	 Jump-start (You may cast this card from your graveyard by discarding a card
//	 in addition to paying its other costs. Then exile this card.)"
//
// Surveil, then the draw (each waits for the question the one before asks),
// then, in a Planechase game, a yes/no for the planeswalk (CR 701.31). Outside
// Planechase there is no planar deck to planeswalk with, so nothing is offered.
// Jump-start is Flashback's offer with a discard (cmdr2S39JumpStart).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "31b72bcc-2f45-4ff8-8e4b-27c020ff0e32",
		Name:             "Start the TARDIS",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Draws: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Cmdr2S39JumpStart("{1}{U}")},
		OnResolve: effects.ResolveSteps(
			effects.Surveil{N: 2},
			effects.DrawCards{N: 1},
			effects.Instruction(func(ctx *effects.Context) error {
				if ctx.Game.Planechase == nil {
					return nil
				}
				return effects.MayChoice{
					Question: "Start the TARDIS — planeswalk?",
					Then:     startTheTardisPlaneswalk,
				}.Apply(ctx)
			}),
		),
	})
}

// startTheTardisPlaneswalk is the "you may planeswalk" answer: a yes planeswalks
// its controller. An on-disk key.
var startTheTardisPlaneswalk = effects.MayYes("confirm/start-the-tardis-planeswalk", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return ctx.Game.PlaneswalkForEffect(ctx.Controller())
})
