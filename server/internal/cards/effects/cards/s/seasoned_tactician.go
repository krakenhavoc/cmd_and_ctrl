package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seasoned Tactician — Creature — Human Advisor {2}{W}, 1/3:
//
//	"{3}, Exile the top four cards of your library: The next time a
//	 source of your choice would deal damage to you this turn, prevent
//	 that damage."
//
// ADR 0109 §7 (#1902): "Exile the top N cards of your library" is a
// cost with nothing to choose. A library of fewer than N cards can't pay
// it (CR 118.3), and it is paid after every other cost (CR 601.2h).
// The shield is ADR 0107 §6's: the next instance of damage from a source
// chosen as the ability resolves (CR 615.8, 609.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7c2f27c-8c82-4283-a17d-f979ee636542",
		Name:         "Seasoned Tactician",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{effects.NextDamageShieldRow(
			"{3}, Exile the top four cards of your library: The next time a source of your choice would deal damage to you this turn, prevent that damage.",
			effects.Plus(effects.ManaCost("{3}"), effects.ExileTopOfLibrary(4)), nil, effects.PreventNextDamageFromChosenSource(effects.ShieldYou))},
	})
}
