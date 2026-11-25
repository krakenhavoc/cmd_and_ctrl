package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weapons Trainer — Creature — Human Soldier Ally {R}{W}, 3/2 (Oath of the
// Gatewatch, #5545):
//
//	"Other creatures you control get +1/+0 as long as you control an
//	 Equipment."
//
// A layer 7c anthem over the other creatures its controller controls, applying
// only while that player controls an Equipment, read live on every recompute (so
// it switches off the moment the last Equipment leaves). The Trainer itself is
// not pumped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d193f3b7-d205-4c49-ae52-6a095282e0b3",
		Name:         "Weapons Trainer",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.B16Anthem(func(target *game.Card, g *game.Game, source *game.Card) bool {
			return effects.OtherCreaturesYouControl(target, g, source) && effects.OldB300ControlsAnEquipment(g, source)
		}, 1, 0)},
	})
}
