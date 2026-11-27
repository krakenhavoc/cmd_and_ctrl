package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Looming Hoverguard — Creature — Drone {4}{U}{U}, 3/3 (Mirrodin, S82 OLD,
// #5606):
//
//	"Flying
//	 When this creature enters, put target artifact on top of its owner's
//	 library."
//
// Run Aground's tuck on an enters trigger: the target is any artifact,
// whoever controls it, chosen as the trigger goes on the stack, and it goes
// on top of its OWNER's library as the trigger resolves (CR 608.2b if it is
// no longer there). The Hoverguard itself is not an artifact; an artifact
// creature is a legal target. A token ceases to exist.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e1c49bdb-cfd1-4ebf-90f7-73bd5009f75e",
		Name:            "Looming Hoverguard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Looming Hoverguard — put target artifact on top of its owner's library",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PutTargetOnTopOfOwnersLibrary(item, effects.NewContext(g, item))
				}), effects.TargetPermanent("target artifact", effects.Artifact())),
		},
	})
}
