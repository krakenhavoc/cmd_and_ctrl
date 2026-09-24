package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Caverns of Despair — World Enchantment {2}{R}{R} (EDHREC rank
// 21037):
//
//	"No more than two creatures can attack each combat.
//	 No more than two creatures can block each combat."
//
// Silent Arbiter's lines with a bound of two (#1507). Both are exact.
//
// Sandbox simplification, declared: the WORLD supertype's world rule
// (CR 704.5k — when two world permanents are on the battlefield, all
// but the newest go to the graveyard) is not modelled, the gap
// Concordant Crossroads already declares. The Caverns would survive
// beside another world enchantment where the printed card would not.
func init() {
	Register(Spec{
		OracleID:     "a1034a02-36cf-4586-a001-9dc3fb76e904",
		Name:         "Caverns of Despair",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The world rule isn't applied — it isn't put into the graveyard when another world enchantment enters."},
		AttackLimits: []game.AttackLimit{NoMoreThanNCanAttackEachCombat(2)},
		BlockRules:   []game.BlockRule{NoMoreThanNCanBlockEachCombat(2)},
	})
}
