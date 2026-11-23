package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treacherous Link — Enchantment — Aura {1}{B}:
//
//	"Enchant creature
//	 All damage that would be dealt to enchanted creature is dealt to its
//	 controller instead."
//
// ADR 0108 §9 decision 4 (#1905): a static redirection (CR 614.9) from
// the enchanted creature to its controller as the damage would be dealt.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b7902113-9ddf-4d81-a76a-b54b8a36c154",
		Name:         "Treacherous Link",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Replacements: []game.ReplacementEffect{
			effects.StaticRedirection("Treacherous Link — damage to enchanted creature is dealt to its controller instead",
				effects.RedirectWhere{Applies: effects.DamageToTheAttachedHost, To: effects.ToTheAttachedHostsController}),
		},
	})
}
