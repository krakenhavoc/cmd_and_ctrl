package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1b44fe0a-4a99-4166-b0b3-102b36b54ffa",
		Name:         "Abzan Battle Priest",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{rprB04Outlast("{W}")},
		Static: []game.StaticAbility{
			CreaturesYouControlWithCountersHave(game.CounterPlusOne, false, "lifelink"),
		},
	})
}
