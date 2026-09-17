package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghalta, Primal Hunger — Legendary Creature — Elder Dinosaur
// {10}{G}{G}, 12/12:
//
//	"This spell costs {X} less to cast, where X is the total power of
//	 creatures you control.
//	 Trample"
//
// #746: the reduction is a self cost modifier. It spends generic mana
// only, so with enough power the spell costs exactly {G}{G} and never
// less (the ruling, and ADR 0048 §3). A board under Sphere of
// Resistance adds its {1} first and the reduction eats it too
// (CR 601.2f: increases, then reductions). A total power of zero or
// less reduces nothing.
func init() {
	Register(Spec{
		OracleID:        "b0b6be0c-41cf-4757-9f0e-87227b6ba6b3",
		Name:            "Ghalta, Primal Hunger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(TotalPowerOfCreaturesYouControl(),
				"This spell costs {X} less to cast, where X is the total power of creatures you control."),
		},
	})
}
