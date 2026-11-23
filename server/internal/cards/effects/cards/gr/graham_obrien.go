package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Graham O'Brien — Legendary Creature — Human Pilot {1}{G}, 2/2 (Doctor Who):
//
//	"Paradox — Whenever you cast a spell from anywhere other than your hand,
//	 create a Food token.
//	 Doctor's companion"
//
// The zone is the one the spell left (Event.OldZone), so a commander cast
// from the command zone, a flashback or an exile cast all count and a cast
// from the hand does not; only the controller's casts count. Doctor's
// companion is a deck-construction rule read off the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7986908d-8af2-465a-8597-26e92ca2f1c9",
		Name:         "Graham O'Brien",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.GapB85YouCastFromOutsideYourHand,
				"Paradox — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
	})
}
