package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "2d0aa801-1227-4226-8c27-3bdd441857aa",
		Name:            "Photon, Lady of Light",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Photon — exile up to one other target creature you control, then return it", flickerFirstLegalTargetUnderOwner),
				Another(UpToOneTargetCreature("up to one other target creature you control", YouControl())),
			),
		},
	})
}
