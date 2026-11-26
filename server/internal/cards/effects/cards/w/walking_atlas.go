package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Walking Atlas — Artifact Creature — Construct {2}, 1/1 (Worldwake, S82 OLD
// b487):
//
//	"{T}: You may put a land card from your hand onto the battlefield."
//
// A put, not a play (CR 305.4), so the turn's land drop is untouched, and the land
// arrives untapped through the entry pipeline (MayPutALandFromHand, Eureka Moment's
// clause); the "you may" is the optional pick as the ability resolves, so a hand
// with no land asks nothing. It had been a declared skip waiting for exactly this
// hand-to-battlefield seam.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "855a4837-7fc8-4b97-afbd-daa88e322c89",
		Name:         "Walking Atlas",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: You may put a land card from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1, LandsUntapped: 1},
			Cost:    effects.TapCost(),
			Effect:  effects.Do(effects.MayPutALandFromHand("Walking Atlas")),
		}},
	})
}
