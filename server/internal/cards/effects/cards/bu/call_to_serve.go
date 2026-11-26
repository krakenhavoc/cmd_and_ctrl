package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Call to Serve — {1}{W} Aura: "Enchant nonblack creature. Enchanted
// creature gets +1/+2, has flying, and is an Angel in addition to its
// other types."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb4be6b2-507c-4d7c-bea4-010670dce3da",
		Name:         "Call to Serve",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("enchant nonblack creature", effects.NonBlack()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 2),
			effects.GrantToAttached("flying"),
			effects.AddAttachedSubtypes("Angel"),
		},
	})
}
