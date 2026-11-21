package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Protege's Awakening — Sorcery {3}{U} (Reality Fracture):
//
//	"Empower Jace 6.
//	 Draw a card."
//
// ADR 0139: the draw is printed after the keyword action, so it is the
// action's Then and waits for the Jace choice and any counter
// replacement ordering.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "07617648-6490-4719-acec-7705671f7e01",
		Name:         "Protege's Awakening",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.EmpowerJace{N: 6, Then: protegesAwakeningDraw}.Apply(ctx)
		},
	})
}

// protegesAwakeningDraw is "draw a card" once the counters are on, keyed
// (#4290).
var protegesAwakeningDraw = effects.RegisterEmpowerThen("prompt-run/proteges-awakening-empower", func(ctx *effects.Context) error {
	return effects.DrawCards{N: 1}.Apply(ctx)
})
