package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Price of Freedom — Sorcery — Lesson {1}{R}:
//
//	"Destroy target artifact or land an opponent controls. Its
//	 controller may search their library for a basic land card, put it
//	 onto the battlefield tapped, then shuffle.
//	 Draw a card."
//
// Avengers Disassembled's land bullet over an artifact or land, then a
// draw. The search is optional, and a player who declines does not
// shuffle (the ruling); SearchLibrary's Optional does both. The draw
// waits for the search's answer (CR 608.2c), so it is a second step.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e6faa98-cf55-430d-87ce-8085f68aa20a",
		Name:         "Price of Freedom",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or land an opponent controls", effects.Or(effects.Artifact(), effects.Land()), effects.OpponentControls()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.DestroyThenItsControllerMaySearchForATappedBasic(ctx, id,
					"Price of Freedom — you may search for a basic land")
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
		),
	})
}
