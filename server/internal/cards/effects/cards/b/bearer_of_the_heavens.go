package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bearer of the Heavens — Creature — Giant {7}{R}, 10/10 (Journey into Nyx,
// S82 OLD, #5558):
//
//	"When this creature dies, destroy all permanents at the beginning of the
//	 next end step."
//
// The dies trigger schedules a delayed trigger (CR 603.7); when it fires every
// permanent on the battlefield at that moment is destroyed, lands and the
// controller's own included, with regeneration still allowed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "54e86bc8-ff48-49be-902c-4923d59901a7",
		Name:         "Bearer of the Heavens",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.WhenThisDies("Bearer of the Heavens — destroy all permanents at the beginning of the next end step",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ScheduleDelayedTrigger{
							Label: "Bearer of the Heavens — destroy all permanents",
							Body:  bearerOfTheHeavensDestroyBody,
						}.Apply(effects.NewContext(g, item))
					}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepDestroy}}),
		},
	})
}

// bearerOfTheHeavensDestroyBody is the delayed trigger's "destroy all
// permanents". An on-disk key.
var bearerOfTheHeavensDestroyBody = game.SimpleDelayedBody("s82-old-b370/destroy-all-permanents", func(g *game.Game, item *game.StackItem) error {
	return effects.DestroyAllMatching{Match: func(*game.Game, uuid.UUID, game.Card) bool { return true }}.Apply(effects.NewContext(g, item))
})
