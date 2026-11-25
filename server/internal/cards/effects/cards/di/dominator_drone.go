package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dominator Drone — Creature — Eldrazi Drone {2}{B}, 3/2 (Battle for
// Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Ingest (Whenever this creature deals combat damage to a player, that
//	 player exiles the top card of their library.)
//	 When this creature enters, if you control another colorless creature, each
//	 opponent loses 2 life."
//
// The enters trigger has an intervening "if" (CR 603.4): another colorless
// creature is checked as it enters and again as the trigger resolves.
//
// No simplification.
func init() {
	const name = "Dominator Drone"
	effects.Register(effects.Spec{
		OracleID:        "3c98c925-f380-4520-970d-4c1abbebefe3",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			effects.OldB305Ingest(name),
			effects.OldB305IfYouControlAnotherColorless(
				effects.WhenThisEnters(name+" — each opponent loses 2 life",
					effects.Do(effects.EachOpponentLosesLife{N: 2}))),
		},
	})
}
