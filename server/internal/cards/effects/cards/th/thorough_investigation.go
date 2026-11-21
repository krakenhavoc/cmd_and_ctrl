package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thorough Investigation — Enchantment {2}{W}:
//
//	"Whenever you attack, investigate.
//	 Whenever you sacrifice a Clue, venture into the dungeon."
//
// "Whenever you attack" is one trigger per attack declaration
// (OncePerBatch, Adeline's shape). The Clue sacrifice is Tireless
// Tracker's condition: the event fires before the zone move, so the
// sacrificed Clue is still there to be read.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7504407-bb72-4baf-be4a-8152e4198e2b",
		Name:         "Thorough Investigation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Thorough Investigation — investigate", effects.Do(effects.Investigate{N: 1}))),
			effects.On(game.EventSacrifice, effects.LonisYouSacrificedAClue, effects.VentureLabel("Thorough Investigation"), effects.Venture),
		},
	})
}
