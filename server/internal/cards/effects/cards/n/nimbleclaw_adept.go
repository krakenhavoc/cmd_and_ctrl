package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nimbleclaw Adept — Creature — Dragon Wizard {3}{U}, 2/3 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Bigby's Hand — {T}: Untap two other target permanents. Activate only as a
//	 sorcery and only once each turn."
//
// Thunderhead Gunner's sorcery-speed, once-a-turn activation (a per-object
// tally, so a second Adept has its own turn) with Hope Tender's untap of the
// two targets; "other" is object identity, the Adept never targets itself. The
// ability needs both targets (a 2/2 clause) and untaps each one still legal as
// it resolves.
//
// No simplification.
func init() {
	const label = "Bigby's Hand — {T}: Untap two other target permanents. Activate only as a sorcery and only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "32e5a4ac-69c7-4de0-8b9f-dd86682f2ffd",
		Name:         "Nimbleclaw Adept",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        label,
			Cost:         effects.TapCost(),
			SorcerySpeed: true,
			Condition:    effects.OncePerTurnActivation(label),
			Targets:      effects.Another(effects.TargetPermanent("two other target permanents").WithCount(2, 2)),
			Effect:       effects.B22UntapEachLegalTarget,
		}},
	})
}
