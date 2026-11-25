package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urza's Tome — Artifact — Book {2} (Dominaria, S82 OLD, #5530):
//
//	"{3}, {T}: Draw a card. Then discard a card unless you exile a historic
//	 card from your graveyard. (Artifacts, legendaries, and Sagas are historic.)"
//
// The draw resolves first. Then, with a historic card in the graveyard, the
// player is asked which one to exile (choosing none is a choice, and then they
// discard); with none there is nothing to exile and the discard is made at once.
// The discard is the player's own pick (CR 701.9a), and an empty hand discards
// nothing. The graveyard is read after the draw, as the sentence orders it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4734b73e-104c-4b25-ab52-0145f361c9d1",
		Name:         "Urza's Tome",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Draw a card. Then discard a card unless you exile a historic card from your graveyard.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect: effects.Do(
				effects.DrawCards{N: 1},
				effects.Instruction(effects.S82OldB212UrzasTomeDiscardUnlessExile),
			),
		}},
	})
}
