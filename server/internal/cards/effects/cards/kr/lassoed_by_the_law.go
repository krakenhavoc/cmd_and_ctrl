package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lassoed by the Law — Enchantment {3}{W} (Outlaws of Thunder Junction):
//
//	"When this enchantment enters, exile target nonland permanent an opponent
//	 controls until this enchantment leaves the battlefield.
//	 When this enchantment enters, create a 1/1 red Mercenary creature token
//	 with "{T}: Target creature you control gets +1/+0 until end of turn.
//	 Activate only as a sorcery.""
//
// Two enters triggers, as printed (the controller puts them on the stack in
// the order they like): Banishing Light's exile-until-leaves pair (the exile
// and its return, CR 610.3) and the Mercenary token of the set
// (MercenaryToken). If the enchantment leaves before the first resolves, nothing
// is exiled (CR 610.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9457f3c1-a087-46cd-a8be-cf665466b77c",
		Name:         "Lassoed by the Law",
		Completeness: effects.CompletenessFull,
		Triggered: append(effects.ExileUntilThisLeavesTriggers("Lassoed by the Law"),
			effects.WhenThisEnters("Lassoed by the Law — create a 1/1 red Mercenary creature token",
				effects.Do(effects.CreateToken{Template: effects.MercenaryToken(), N: 1}))),
	})
}
