package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tuskguard Captain — Creature — Human Warrior {2}{G}, 2/3 (Khans of
// Tarkir, #5554):
//
//	"Outlast {G} ({G}, {T}: Put a +1/+1 counter on this creature.
//	 Outlast only as a sorcery.)
//	 Each creature you control with a +1/+1 counter on it has trample."
//
// Outlast is the shared sorcery-speed tap-and-pay counter row
// (rprB04Outlast, CR 702.107a). The static is Abzan Battle Priest's
// layer-6 grant over every creature you control with a +1/+1 counter,
// this one included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa149597-087a-4e1f-953c-2cd04e90a93b",
		Name:         "Tuskguard Captain",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{rprB04Outlast("{G}")},
		Static: []game.StaticAbility{
			CreaturesYouControlWithCountersHave(game.CounterPlusOne, false, "trample"),
		},
	})
}
