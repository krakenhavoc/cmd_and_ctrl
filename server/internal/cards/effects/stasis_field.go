package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stasis Field — Enchantment — Aura {1}{U} (March of the Machine):
//
//	"Enchant creature
//	 Enchanted creature has base power and toughness 0/2, has defender, and
//	 loses all other abilities."
//
// Spider-Man No More's layers without the type change: layer 6 removes every
// ability but the granted defender in one effect, layer 7b sets base 0/2 so
// counters and anthems still apply on top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83633ea1-a8d5-4351-9047-625d6905386d",
		Name:         "Stasis Field",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			LoseAllAbilities("defender"),
			SetAttachedBasePT(0, 2),
		},
	})
}
