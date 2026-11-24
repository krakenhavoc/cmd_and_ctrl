package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bookwurm — Creature — Wurm {7}{G}, 7/7 (Strixhaven: School of Mages, #5502):
//
//	"Trample
//	 When this creature enters, you gain 3 life and draw a card.
//	 {2}{G}: Put this card from your graveyard into your library third from the
//	 top."
//
// The enters trigger gains the life and then draws (one printed sentence, two
// instructions, in that order). The graveyard ability is Screaming Swarm's with
// a depth of three: it does nothing if the card has left the graveyard by
// resolution (CR 602.5, 608.2b), and the card goes third from the top of its
// owner's library (a library of fewer than two cards holds it at the bottom,
// CR 401.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ce77fff1-32f8-4a55-9124-b1dafe95840d",
		Name:            "Bookwurm",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Draws: 1, LifeGain: 3},
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bookwurm — you gain 3 life and draw a card",
				effects.Do(effects.GainLife{Amount: 3}, effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: Put this card from your graveyard into your library third from the top.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{G}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				id := item.SourceCardID
				if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
					return nil
				}
				return effects.PutInLibraryInAnyOrder{
					Chooser:   item.Controller,
					Cards:     []uuid.UUID{id},
					From:      game.ZoneGraveyard,
					Placement: game.LibraryPlaceTop,
					TopDepth:  3,
					Label:     "Bookwurm — put it into your library third from the top",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
