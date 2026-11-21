package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disposal Mummy — Creature — Zombie Jackal {2}{W}, 2/3 (Hour of
// Devastation, #5534):
//
//	"When this creature enters, exile target card from an opponent's
//	 graveyard."
//
// Ruin Rat's clause on an enters trigger: the target is a card in a
// graveyard its controller doesn't own, and the trigger is removed when
// there is none (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "002a9cea-8cf7-48ba-83eb-e1c87a7024e5",
		Name:         "Disposal Mummy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Disposal Mummy — exile target card from an opponent's graveyard", effects.ExileTheFirstLegalTarget),
				effects.TargetCardInGraveyard("target card from an opponent's graveyard", effects.Not(effects.YouOwn())),
			),
		},
	})
}
