package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pyxis of Pandemonium — Artifact {1} (Theros):
//
//	"{T}: Each player exiles the top card of their library face down.
//	 {7}, {T}, Sacrifice this artifact: Each player turns face up all cards
//	 they own exiled with this artifact, then puts all permanent cards among
//	 them onto the battlefield."
//
// The two abilities are linked (CR 607.2a). Each player's top card goes to
// exile face down, nobody looking, linked to this Pyxis object in turn order
// from the active player; the second ability sacrifices the Pyxis as a cost,
// so the link is read from its last-known object (a Pyxis that came back is
// a new object with no claim, CR 400.7). Every linked card is turned face up
// and the permanent cards among them enter together under their owners'
// control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b746fe1-4c11-4dfd-bdc9-edd394478861",
		Name:         "Pyxis of Pandemonium",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Each player exiles the top card of their library face down.",
				Cost:    effects.TapCost(),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  effects.PyxisOfPandemoniumExile,
			},
			{
				Label:   "{7}, {T}, Sacrifice this artifact: Each player turns face up all cards they own exiled with this artifact, then puts all permanent cards among them onto the battlefield.",
				Cost:    effects.Plus(effects.ManaCost("{7}"), effects.TapCost(), effects.SacrificeThis()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  effects.PyxisOfPandemoniumRelease,
			},
		},
	})
}
