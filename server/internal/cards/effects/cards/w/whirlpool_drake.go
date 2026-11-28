package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Whirlpool Drake — Creature — Drake {3}{U}, 2/2 (Apocalypse, S82 OLD, #5614):
//
//	"Flying
//	 When this creature enters, shuffle the cards from your hand into your
//	 library, then draw that many cards.
//	 When this creature dies, shuffle the cards from your hand into your
//	 library, then draw that many cards."
//
// Two printed triggers, two registered ones, each Winds of Change's
// shuffle-and-redraw over the controller alone (the dies trigger's controller
// is the creature's last controller, CR 603.10a).
//
// No simplification.
func init() {
	redraw := func(g *game.Game, item *game.StackItem) error {
		return effects.WindsOfChangeShuffleFrom(g, []uuid.UUID{item.Controller}, nil)
	}
	effects.Register(effects.Spec{
		OracleID:        "aef9fcf0-b755-40f6-821d-836ae1b5497a",
		Name:            "Whirlpool Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Whirlpool Drake — shuffle the cards from your hand into your library, then draw that many cards", redraw),
			effects.WhenThisDies("Whirlpool Drake — shuffle the cards from your hand into your library, then draw that many cards", redraw),
		},
	})
}
