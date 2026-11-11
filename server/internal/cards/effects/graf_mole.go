package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Graf Mole — Creature — Mole Beast {2}{G}, 2/4 (Murders at Karlov
// Manor Commander, #5368):
//
//	"Whenever you sacrifice a Clue, you gain 3 life."
//
// Tireless Tracker's condition (lonisYouSacrificedAClue): the sacrifice
// event fires before the Clue leaves, so its subtype is still readable,
// and any Clue counts, token or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b2c6b44-99b3-496d-93ca-d5211222038f",
		Name:         "Graf Mole",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, lonisYouSacrificedAClue,
				"Graf Mole — you gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
