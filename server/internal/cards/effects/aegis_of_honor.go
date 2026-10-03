package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ea93acc8-0c1f-42a2-bed3-f385d210d58f",
		Name:         "Aegis of Honor",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			redirectRow("{1}: The next time an instant or sorcery spell would deal damage to you this turn, that spell deals that damage to its controller instead.",
				ManaCost("{1}"), nil,
				RedirectDamage{Queries: []game.PermanentQuery{QueryTypes("Instant", "Sorcery")}, Protect: ShieldYou, Next: true, To: RedirectToSourceController}),
		},
	})
}
