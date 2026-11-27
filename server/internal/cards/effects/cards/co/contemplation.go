package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Contemplation — Enchantment {1}{W}{W} (Stronghold, S82 OLD b750, #5627):
//
//	"Whenever you cast a spell, you gain 1 life."
//
// A cast trigger (YouCast(nil), as Aetherflux Reservoir's) for a fixed 1 life.
// A copy is not cast and does not trigger it (CR 707.10).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa7efcef-a688-4e25-a823-4d53b2e96508",
		Name:         "Contemplation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCast(nil), "Contemplation — you gain 1 life",
				effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
