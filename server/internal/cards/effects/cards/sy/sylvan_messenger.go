package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sylvan Messenger — Creature — Elf {3}{G}, 2/2 (Kaldheim Commander, #5506):
//
//	"Trample
//	 When this creature enters, reveal the top four cards of your library.
//	 Put all Elf cards revealed this way into your hand and the rest on the
//	 bottom of your library in any order."
//
// The reveal is public, no choice is offered for the Elves (all of them
// are taken), and the controller orders the rest on the bottom. "Elf" is a
// word of this card's text. A library shorter than four reveals what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ba310102-8272-4d64-a00b-9e7346cb5f46",
		Name:            "Sylvan Messenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sylvan Messenger — reveal the top four cards; Elf cards to your hand, the rest on the bottom in any order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					elf := ctx.TextSubtype("Elf")
					return effects.S82OldB074RevealTopTakeAllToHandRestBottom(g, item.Controller, ctx.Source(), 4,
						"Sylvan Messenger — reveal the top four cards, Elf cards to your hand and the rest on the bottom of your library",
						effects.HasSubtype(elf))
				}),
		},
	})
}
