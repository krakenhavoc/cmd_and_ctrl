package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sanctum Gargoyle — Artifact Creature — Gargoyle {3}{W}, 2/3 (Shards of
// Alara, #5588):
//
//	"Flying
//	 When this creature enters, you may return target artifact card from
//	 your graveyard to your hand."
//
// Treasure Hunter's enters trigger on a flier (CR 702.9b): the target is
// an artifact card in its controller's own graveyard, chosen as the
// trigger goes on the stack (CR 603.3d), and the return is a "you may"
// (CR 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6ee9802a-7a44-49c7-93d9-f75c4d42e71e",
		Name:            "Sanctum Gargoyle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					WhenThisEnters("Sanctum Gargoyle — return an artifact card to your hand",
						returnFirstLegalGraveyardTargetToHand),
					"Sanctum Gargoyle — return an artifact card from your graveyard to your hand?"),
				TargetCardInGraveyard("target artifact card from your graveyard", YouOwn(), Artifact())),
		},
	})
}
