package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aegis of Honor — Enchantment {W}:
//
//	"{1}: The next time an instant or sorcery spell would deal damage to
//	 you this turn, that spell deals that damage to its controller instead."
//
// ADR 0108 §9 (#1905): a "next time" redirection with no chosen source,
// keyed on the property (CR 609.7b: rechecked as the spell would deal its
// damage), dealt to the spell's controller as it is then.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea93acc8-0c1f-42a2-bed3-f385d210d58f",
		Name:         "Aegis of Honor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.RedirectRow("{1}: The next time an instant or sorcery spell would deal damage to you this turn, that spell deals that damage to its controller instead.",
				effects.ManaCost("{1}"), nil,
				effects.RedirectDamage{Queries: []game.PermanentQuery{effects.QueryTypes("Instant", "Sorcery")}, Protect: effects.ShieldYou, Next: true, To: effects.RedirectToSourceController}),
		},
	})
}
