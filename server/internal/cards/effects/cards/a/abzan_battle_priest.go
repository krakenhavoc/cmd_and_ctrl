package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abzan Battle Priest — Creature — Human Cleric {3}{W}, 3/2:
//
//	"Outlast {W} ({W}, {T}: Put a +1/+1 counter on this creature. Outlast
//	 only as a sorcery.)
//	 Each creature you control with a +1/+1 counter on it has lifelink."
//
// Outlast is the shared sorcery-speed tap-and-pay counter ability. The
// static is Drix Fatemaker's layer 6 grant over every creature you control
// with a +1/+1 counter, the Priest included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b44fe0a-4a99-4166-b0b3-102b36b54ffa",
		Name:         "Abzan Battle Priest",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.RprB04Outlast("{W}")},
		Static: []game.StaticAbility{
			effects.CreaturesYouControlWithCountersHave(game.CounterPlusOne, false, "lifelink"),
		},
	})
}
