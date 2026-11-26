package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Treasure Mage — Creature — Human Wizard {2}{U}, 2/2 (Mirrodin Besieged, S82
// OLD b466, #5576):
//
//	"When this creature enters, you may search your library for an artifact
//	 card with mana value 6 or greater, reveal it, put it into your hand, then
//	 shuffle."
//
// A tutor that names a quality, so the controller may find nothing (CR
// 701.23b); "you may search" lets them decline the search and the shuffle
// both. The mana value is read off the card in the library (an {X} in the
// cost is 0 there, CR 202.3e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0848bb49-5758-41d6-a3e8-0c62e1768f52",
		Name:         "Treasure Mage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1, Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Treasure Mage — you may search for an artifact card with mana value 6 or greater",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsArtifact() && c.ManaValue() >= 6 },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Optional:  true,
						Shuffle:   true,
						Reason:    "Treasure Mage — an artifact card with mana value 6 or greater",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
