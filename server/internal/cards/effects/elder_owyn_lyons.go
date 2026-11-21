package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elder Owyn Lyons — Legendary Creature — Human Knight {2}{W}{U}, 3/3
// (Fallout):
//
//	"Artifacts you control have ward {1}.
//	 When Elder Owyn Lyons enters or dies, return target artifact card from
//	 your graveyard to your hand."
//
// The ward grant is Bronze Guardian's without "other": the Elder is not an
// artifact, so it makes no difference. The trigger is Aerid Konstrari's
// "enters or dies" with Loran's return.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "19788405-7141-475e-94ee-d0c89488f50f",
		Name:         "Elder Owyn Lyons",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WardGranted(WardMana("{1}"), "Elder Owyn Lyons — artifacts you control have ward {1}",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && target.IsArtifact()
				}),
			Targeting(WhenThisEntersOrDies("Elder Owyn Lyons — return target artifact card from your graveyard to your hand",
				returnTargetGraveyardCardToHand),
				TargetCardInGraveyard("target artifact card from your graveyard", YouOwn(), Artifact())),
		},
	})
}
