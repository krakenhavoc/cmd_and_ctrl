package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loran, Disciple of History — Legendary Creature — Human Artificer {3}{W},
// 3/3 (The Brothers' War, #5482):
//
//	"Whenever Loran or another legendary creature you control enters,
//	 return target artifact card from your graveyard to your hand."
//
// One trigger over two conditions: Loran's own entry, and any other
// legendary creature entering under your control (Loran is legendary
// itself, so the second covers every legendary creature you control).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9d1f232c-e904-46f6-a4b7-24ca2ede711c",
		Name:         "Loran, Disciple of History",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, thisOrAnotherLegendaryCreatureYouControlEntered,
					"Loran, Disciple of History — return target artifact card from your graveyard to your hand",
					returnTargetGraveyardCardToHand),
				TargetCardInGraveyard("target artifact card from your graveyard", YouOwn(), Artifact()),
			),
		},
	})
}
