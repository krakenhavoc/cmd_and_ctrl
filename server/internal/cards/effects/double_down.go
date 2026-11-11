package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Double Down — Enchantment {3}{U} (Outlaws of Thunder Junction):
//
//	"Whenever you cast an outlaw spell, copy that spell. (Assassins,
//	 Mercenaries, Pirates, Rogues, and Warlocks are outlaws. Copies of
//	 permanent spells become tokens.)"
//
// The cast trigger over the outlaw batch (b40Outlaw reads the spell's
// effective creature types, so a changeling spell is one). "That spell"
// names the spell without targeting it, so one countered before the trigger
// resolves is still copied from last-known information (CR 608.2h). The
// printed text does not offer new targets (CR 707.10c), so the copy keeps
// the original's. A copy of a permanent spell becomes a token as it
// resolves (CR 707.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60dd4233-ee50-42b8-88f5-ac5d105d83cd",
		Name:         "Double Down",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(b40Outlaw(), "Double Down — copy that outlaw spell",
				cmdr2S41CopyTheSpellJustCast),
		},
	})
}
