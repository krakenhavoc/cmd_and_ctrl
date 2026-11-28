package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wave of Terror — {2}{B} Enchantment (Weatherlight, S82 OLD batch 772, #5629):
//
//	"Cumulative upkeep {1}
//	 At the beginning of your draw step, destroy each creature with mana value
//	 equal to the number of age counters on this enchantment. They can't be
//	 regenerated."
//
// The draw-step trigger fires on its controller's draw step only, and destroys
// every creature, any player's, whose mana value equals the age counters the
// enchantment has as the ability resolves, as one event that ignores regeneration
// (CR 701.19c). A token has mana value 0, so it falls only at zero counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2346be19-7344-4f11-be2c-a22750cb7e1d",
		Name:         "Wave of Terror",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true},
		},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Wave of Terror — cumulative upkeep {1}", "{1}"),
			effects.On(game.EventBeginDrawStep, effects.ByYou, "Wave of Terror — destroy each creature with mana value equal to the number of age counters on it; they can't be regenerated",
				effects.S82OldB772WaveOfTerrorDestroy),
		},
	})
}
