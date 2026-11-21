package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Photon, Lady of Light — Legendary Creature — Human Hero {3}{W}, 3/3
// (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever Photon attacks, exile up to one other target creature you
//	 control, then return that card to the battlefield under its owner's
//	 control."
//
// The target is chosen as the trigger goes on the stack ("up to one",
// so none is legal), and the blink is the shared Flicker, which returns
// the card as a new object under its owner (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2d0aa801-1227-4226-8c27-3bdd441857aa",
		Name:            "Photon, Lady of Light",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Photon — exile up to one other target creature you control, then return it", effects.FlickerFirstLegalTargetUnderOwner),
				effects.Another(effects.UpToOneTargetCreature("up to one other target creature you control", effects.YouControl())),
			),
		},
	})
}
