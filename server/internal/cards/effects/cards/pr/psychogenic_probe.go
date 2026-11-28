package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Psychogenic Probe — Artifact {2} (Mirrodin, #5606):
//
//	"Whenever a spell or ability causes a player to shuffle their library,
//	 this artifact deals 2 damage to that player."
//
// Widespread Panic's watch on game.EventLibraryShuffled caused by a spell or
// ability (a search's shuffle included; the sandbox's manual shuffle and the
// opening-hand and mulligan shuffles are none of those). "That player" is the
// shuffled library's owner, the event's actor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1beb81a8-b4aa-4964-93d2-2b5414fc6ce3",
		Name:         "Psychogenic Probe",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLibraryShuffled, effects.ASpellOrAbilityShufflesALibrary,
				"Psychogenic Probe — it deals 2 damage to that player",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.DealDamageToPlayerForEffect(ctx.Source(), ctx.Trigger().Event.Actor, 2)
				}),
		},
	})
}
