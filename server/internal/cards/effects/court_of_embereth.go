package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Court of Embereth — Enchantment {2}{R}{R} (Wilds of Eldraine Commander,
// #5368):
//
//	"When this enchantment enters, you become the monarch.
//	 At the beginning of your upkeep, create a 3/1 red Knight creature
//	 token. Then if you're the monarch, this enchantment deals X damage to
//	 each opponent, where X is the number of creatures you control."
//
// The Knight is made first, so it counts toward X. "Then if you're the
// monarch" is a clause of the effect, read as the trigger resolves (not an
// intervening "if"), and X is counted then too: the Court is the source of
// the damage, by last-known information if it has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "87a82441-9dfb-49bd-8b9e-eaab2c552843",
		Name:         "Court of Embereth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Court of Embereth"),
			AtYourUpkeep("Court of Embereth — create a 3/1 red Knight; if you're the monarch, it deals X damage to each opponent",
				cmdr2S48CourtOfEmberethUpkeep),
		},
	})
}
