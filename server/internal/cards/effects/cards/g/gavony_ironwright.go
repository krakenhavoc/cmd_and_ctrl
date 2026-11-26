package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gavony Ironwright — Creature — Human Soldier {2}{W} 1/4:
//
//	"Fateful hour — As long as you have 5 or less life, other creatures
//	 you control get +1/+4."
//
// The life read is a layer-7c static that declares DependsOnLifeTotal
// (LifeGatedPump), so it turns on and off as the life total moves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "99a4f104-c358-442a-8273-6fe94286edb5",
		Name:         "Gavony Ironwright",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.OldB441OtherCreaturesPumpWhileFatefulHour(1, 4)},
	})
}
