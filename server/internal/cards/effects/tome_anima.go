package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tome Anima — Creature — Spirit {3}{U}, 3/3 (Jumpstart):
//
//	"This creature can't be blocked as long as you've drawn two or more
//	 cards this turn."
//
// A CR 509.1b block rule on the creature itself (Slippery Scoundrel's
// CantBeBlockedWhile), read when blocks are declared off its controller's
// per-turn draw tally. Every turn's tally starts at zero, so the evasion
// is on in a turn where you drew two and off again in the next.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4a68bb3-5a79-4764-8223-acddd16fb660",
		Name:         "Tome Anima",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedWhile(OnSelf(), jmpB3YouHaveDrawnTwoOrMoreCards),
		},
	})
}
