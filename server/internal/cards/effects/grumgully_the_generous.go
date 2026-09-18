package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grumgully, the Generous — Legendary Creature — Goblin Shaman
// {1}{R}{G}, 3/3 (EDHREC rank 4353):
//
//	"Each other non-Human creature you control enters with an
//	 additional +1/+1 counter on it."
//
// A three-mana Gruul commander whose text is one replacement effect,
// and the reason it is played is a combo rather than the beatdown: a
// persist creature that comes back with a −1/−1 counter and gets a
// +1/+1 counter at the same time has the two cancel out (CR 704.5q),
// so it persists forever. Grumgully himself is a Goblin, which is
// non-Human, which is exactly why the card says "other".
//
// It is a REAL CR 614 entry replacement and not a counter added a beat
// later, which is load-bearing three ways: the creature is never on the
// battlefield without the counter, so nothing sees it at its printed
// size; Hardened Scales and Doubling Season apply to it, because they
// replace the same event; and the persist interaction works, because
// the two counters arrive together and the state-based action annihilates
// them.
//
// "Non-Human" reads EFFECTIVE subtypes, so a changeling IS a Human and
// gets nothing, and a creature something turned into a Human stops
// qualifying. Tokens get the counter: a created token runs the same
// entry pipeline every other permanent runs.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa4afcf7-9cf7-4f67-95b9-1f7fe66bf332",
		Name:         "Grumgully, the Generous",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b41OtherNonHumanCreaturesYouControlEnterWithACounter(
				"Grumgully, the Generous: enters with an additional +1/+1 counter"),
		},
	})
}
