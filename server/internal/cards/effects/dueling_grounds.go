package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dueling Grounds — Enchantment {1}{G}{W} (EDHREC rank 12155):
//
//	"No more than one creature can attack each combat.
//	 No more than one creature can block each combat."
//
// Silent Arbiter's two lines on an enchantment, so the same two
// declarations (#1507): an AttackLimit and a whole-combat
// BlockRule.Limit. The current Oracle text says "each combat" (early
// printings said "each turn"), and that is what the engine counts.
func init() {
	Register(Spec{
		OracleID:     "eb2df4a1-6b63-4f6a-a830-e9487afe59f8",
		Name:         "Dueling Grounds",
		Completeness: CompletenessFull,
		AttackLimits: []game.AttackLimit{NoMoreThanNCanAttackEachCombat(1)},
		BlockRules:   []game.BlockRule{NoMoreThanNCanBlockEachCombat(1)},
	})
}
