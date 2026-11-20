package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Greater Sandwurm — Creature — Wurm {5}{G}{G}, 7/7
// (Amonkhet, #5535):
//
//	"This creature can't be blocked by creatures with power 2 or less.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Enraged Ceratok's evasion, plus cycling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d1dfed64-6e88-44c2-b479-9a2d072e4be7",
		Name:         "Greater Sandwurm",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"),
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
