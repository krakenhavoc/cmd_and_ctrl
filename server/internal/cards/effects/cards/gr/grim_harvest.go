package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grim Harvest — Instant {1}{B} (Coldsnap):
//
//	"Return target creature card from your graveyard to your hand.
//	 Recover {2}{B}"
//
// Recover (CR 702.59a) is S82OldB564Recover: from the graveyard, when a
// creature is put into its owner's graveyard from the battlefield, the owner
// may pay {2}{B} to return the card to hand, otherwise it is exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "231ca538-b277-4c12-979f-57640a0e87f0",
		Name:         "Grim Harvest",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreatureInYourGraveyard(),
		OnResolve:    effects.ReturnTargetFromGraveyardToHand,
		Triggered:    []game.TriggeredAbility{effects.S82OldB564Recover("Grim Harvest", "{2}{B}")},
	})
}
