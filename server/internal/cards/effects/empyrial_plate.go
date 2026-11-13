package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Empyrial Plate — Artifact — Equipment {2} (Mirrodin, #5606):
//
//	"Equipped creature gets +1/+1 for each card in your hand.
//	 Equip {2}"
//
// Banshee's Blade's per-count pump (PumpAttachedPer), counting the
// cards in the hand of the Plate's controller, who is "you" even when
// the equipped creature is someone else's. It is a layer 7c
// modification (CR 613.4c), and a hand is not on the battlefield, so
// the static declares DependsOnHandSize as Stingerback Terror's does:
// every draw, discard or cast resizes the creature at once.
//
// No simplification.
func init() {
	pump := PumpAttachedPer(1, 1, func(g *game.Game, source *game.Card) int { return cardsInHand(g, source.Controller) })
	pump.DependsOnHandSize = true
	Register(Spec{
		OracleID:     "e95e100e-4886-4bde-825c-017e878ac825",
		Name:         "Empyrial Plate",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{pump},
		Activated:    []ActivatedAbility{EquipAbility("{2}")},
	})
}
