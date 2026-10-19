package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e7504407-bb72-4baf-be4a-8152e4198e2b",
		Name:         "Thorough Investigation",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Thorough Investigation — investigate", Do(CreateToken{Template: ClueToken(), N: 1}))),
			On(game.EventSacrifice, lonisYouSacrificedAClue, ventureLabel("Thorough Investigation"), Venture),
		},
	})
}
