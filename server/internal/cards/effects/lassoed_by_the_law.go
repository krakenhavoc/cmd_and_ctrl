package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lassoed by the Law — Enchantment {3}{W} (Outlaws of Thunder Junction,
// #4468):
//
//	"When this enchantment enters, exile target nonland permanent an
//	 opponent controls until this enchantment leaves the battlefield.
//	 When this enchantment enters, create a 1/1 red Mercenary creature
//	 token with '{T}: Target creature you control gets +1/+0 until end of
//	 turn. Activate only as a sorcery.'"
//
// The first ability is Banishing Light's (exileUntilThisLeavesTriggers,
// CR 610.3): the card returns when the enchantment leaves, and if the
// enchantment leaves before the trigger resolves, nothing is exiled. The
// second is a separate enter trigger making the set's Mercenary token
// (MercenaryToken), so the two can be ordered and the token is made
// even with no opponent's permanent to target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9457f3c1-a087-46cd-a8be-cf665466b77c",
		Name:         "Lassoed by the Law",
		Completeness: CompletenessFull,
		Triggered: append(exileUntilThisLeavesTriggers("Lassoed by the Law"),
			WhenThisEnters("Lassoed by the Law — create a 1/1 red Mercenary creature token",
				Do(CreateToken{Template: MercenaryToken(), N: 1}))),
	})
}
