package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hand of Emrakul — Creature — Eldrazi {9}, 7/7:
//
//	"You may sacrifice four Eldrazi Spawn rather than pay this spell's
//	 mana cost.
//	 Annihilator 1"
//
// The alternative cost is the shared sacrifice-instead offer (Fireblast's
// shape): four Eldrazi Spawn you control, sacrificed together as the
// spell is cast. It does not change when the Hand may be cast, nor its
// mana value, and cost increases and reductions still apply (the
// 2010-06-15 rulings). Annihilator 1 is the canonical keyword token
// (ADR 0113 §2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b206d3bc-1203-4f00-997e-5da706346a24",
		Name:            "Hand of Emrakul",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"annihilator 1"},
		AlternativeCosts: []game.AlternativeCost{
			SacrificeInstead(4, "four Eldrazi Spawn", 0, HasSubtype("Eldrazi"), HasSubtype("Spawn")),
		},
	})
}
