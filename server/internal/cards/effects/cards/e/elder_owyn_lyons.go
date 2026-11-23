package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "19788405-7141-475e-94ee-d0c89488f50f",
		Name:         "Elder Owyn Lyons",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WardGranted(effects.WardMana("{1}"), "Elder Owyn Lyons — artifacts you control have ward {1}",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && target.IsArtifact()
				}),
			effects.Targeting(effects.WhenThisEntersOrDies("Elder Owyn Lyons — return target artifact card from your graveyard to your hand",
				effects.ReturnTargetGraveyardCardToHand),
				effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.YouOwn(), effects.Artifact())),
		},
	})
}
