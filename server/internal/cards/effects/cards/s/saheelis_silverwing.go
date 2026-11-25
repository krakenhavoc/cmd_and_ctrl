package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saheeli's Silverwing — Artifact Creature — Drake {4}, 2/3 (War of the Spark, #5520):
//
//	"Flying
//	 When this creature enters, look at the top card of target opponent's
//	 library."
//
// Mishra's Bauble's look, the card going back on top: you look, the owner is
// the opponent. The target is chosen as the trigger goes on the stack and the
// look does nothing if it is illegal on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ab5f6e83-a151-48b6-8cc5-3b2dae94079c",
		Name:            "Saheeli's Silverwing",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.Self,
			Targets:   effects.TargetPlayer("target opponent", effects.Opponent()),
			Key:       "Saheeli's Silverwing — look at the top card of target opponent's library",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				var victim uuid.UUID
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						victim = t.ID
					}
				}
				if victim == uuid.Nil {
					return nil
				}
				return effects.LookAtLibraryThenPlace{
					Owner:     victim,
					N:         1,
					Placement: game.LibraryPlaceTop,
					Label:     "Saheeli's Silverwing — put it back on top",
				}.Apply(ctx)
			},
		}},
	})
}
