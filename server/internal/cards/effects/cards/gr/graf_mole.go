package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "5b2c6b44-99b3-496d-93ca-d5211222038f",
		Name:         "Graf Mole",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, effects.LonisYouSacrificedAClue,
				"Graf Mole — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
