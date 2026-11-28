package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Call of the Wild — Enchantment {2}{G}{G} (Eighth Edition, S82 OLD batch 612,
// #5607):
//
//	"{2}{G}{G}: Reveal the top card of your library. If it's a creature card, put
//	 it onto the battlefield. Otherwise, put it into your graveyard."
//
// Lurking Predators' reveal (RevealTopThenPutIfMatch): a creature card goes
// onto the battlefield under your control, anything else into your graveyard
// (an ordinary move, not a mill, so graveyard hosers see it); an empty library
// does nothing and a revealed token is neither put nor moved (CR 111.8). Any
// number of activations: nothing limits them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "882ee72e-2149-4778-a81c-8bf09beb5d66",
		Name:         "Call of the Wild",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}{G}: Reveal the top card of your library. If it's a creature card, put it onto the battlefield. Otherwise, put it into your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{G}{G}"),
			Effect:  effects.OldB612RevealTopCreatureElseGraveyard,
		}},
	})
}
