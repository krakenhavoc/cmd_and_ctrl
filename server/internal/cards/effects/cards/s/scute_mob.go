package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scute Mob — Creature — Insect {G}, 1/1:
//
//	"At the beginning of your upkeep, if you control five or more lands, put
//	 four +1/+1 counters on this creature."
//
// An intervening "if" (CR 603.4): the land count is checked when the upkeep
// begins and again as the ability resolves. The counters go on the source
// only while it is still the same permanent (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1e0eda0-4a5e-4f1f-8131-ba58a58887cf",
		Name:         "Scute Mob",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control five or more lands",
				func(g *game.Game, you uuid.UUID) bool { return effects.B10LandsControlled(g, you) >= 5 },
				effects.AtYourUpkeep("Scute Mob — put four +1/+1 counters on this creature",
					effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 4}))),
		},
	})
}
