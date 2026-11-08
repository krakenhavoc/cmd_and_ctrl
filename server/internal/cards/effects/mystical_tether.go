package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mystical Tether — Enchantment {2}{W} (Outlaws of Thunder Junction,
// #4468):
//
//	"You may cast this spell as though it had flash if you pay {2} more
//	 to cast it.
//	 When this enchantment enters, exile target artifact or creature an
//	 opponent controls until this enchantment leaves the battlefield."
//
// The surcharge is Asinine Antics's FlashForMore offer ({2}{W} plus
// {2}), so it carries that card's caveat (flash_for_more.go, #5207).
// The exile is Glass Casket's: the target is chosen as the trigger goes
// on the stack (CR 603.3d), and the card returns when the Tether leaves
// the battlefield (CR 610.3). If the Tether has already left by the time
// the trigger resolves, the card is not exiled at all.
func init() {
	Register(Spec{
		OracleID:     "fb1a6986-39f7-4592-9039-7e30f5ccd009",
		Name:         "Mystical Tether",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Paying {2} more to cast this at instant speed can't be combined with another alternative cost.",
		},
		AlternativeCosts: []game.AlternativeCost{FlashForMore("{4}{W}", "{2}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Mystical Tether — exile target artifact or creature an opponent controls until this enchantment leaves the battlefield",
				exileChosenTargetUntilThisLeaves("Mystical Tether — the exiled card returns when Mystical Tether leaves the battlefield")),
				TargetPermanent("target artifact or creature an opponent controls", Or(Artifact(), Creature()), OpponentControls())),
		},
	})
}
