package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Super Strength — {4}{G} Enchantment — Aura (Marvel Super Heroes, #4455):
//
//	"Enchant creature
//	 Enchanted creature gets +4/+4 and has trample and ward {1}."
//
// The +4/+4 and trample are the Aura's statics; the ward is its own
// trigger watching the host become a target (WardAttached), as Sheltered
// by Ghosts grants its own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e1b577ad-1488-43b5-83d9-e40bd60c74be",
		Name:         "Super Strength",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(4, 4),
			GrantToAttached("trample"),
		},
		Triggered: []game.TriggeredAbility{
			WardAttached(WardMana("{1}"), "Super Strength — enchanted creature's ward {1}"),
		},
	})
}
