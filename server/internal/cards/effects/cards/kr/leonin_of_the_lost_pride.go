package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leonin of the Lost Pride — Creature — Cat Warrior {1}{W}, 3/1:
//
//	"When this creature dies, exile target card from an opponent's
//	 graveyard."
//
// Carrion Locust's clause without the life loss: a target that left in
// response leaves the trigger with no legal target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c692dd9-3986-46de-ac10-cc917b7c41b7",
		Name:         "Leonin of the Lost Pride",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Leonin of the Lost Pride — exile target card from an opponent's graveyard", effects.ExileTheFirstLegalTarget),
				effects.TargetCardInGraveyard("target card from an opponent's graveyard", effects.Not(effects.YouOwn())),
			),
		},
	})
}
