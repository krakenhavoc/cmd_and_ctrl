package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pariah — Enchantment — Aura {2}{W}:
//
//	"Enchant creature
//	 All damage that would be dealt to you is dealt to enchanted creature
//	 instead."
//
// ADR 0108 §9 decision 4 (#1905): a static redirection (CR 614.9) through
// the one primitive. The rulings: redirected combat damage is still combat
// damage; two Pariahs on two creatures are two effects, and you choose
// which one applies to each event (CR 616.1) — the damage is never split.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c5c8250-1860-42a1-a335-071f54830d37",
		Name:         "Pariah",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Replacements: []game.ReplacementEffect{
			effects.RedirectYourDamageToAttached("Pariah — damage to you is dealt to enchanted creature instead"),
		},
	})
}
