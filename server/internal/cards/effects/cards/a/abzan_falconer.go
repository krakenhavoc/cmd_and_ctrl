package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abzan Falconer — Creature — Human Soldier {2}{W}, 2/3:
//
//	"Outlast {W} ({W}, {T}: Put a +1/+1 counter on this creature. Outlast
//	 only as a sorcery.)
//	 Each creature you control with a +1/+1 counter on it has flying."
//
// Abzan Battle Priest's shape with flying for lifelink.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b972819-507a-40a9-ab1f-1a674ea56083",
		Name:         "Abzan Falconer",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.RprB04Outlast("{W}")},
		Static: []game.StaticAbility{
			effects.CreaturesYouControlWithCountersHave(game.CounterPlusOne, false, "flying"),
		},
	})
}
