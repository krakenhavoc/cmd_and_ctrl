package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "87a82441-9dfb-49bd-8b9e-eaab2c552843",
		Name:         "Court of Embereth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouBecomeTheMonarch("Court of Embereth"),
			effects.AtYourUpkeep("Court of Embereth — create a 3/1 red Knight; if you're the monarch, it deals X damage to each opponent",
				effects.Cmdr2S48CourtOfEmberethUpkeep),
		},
	})
}
