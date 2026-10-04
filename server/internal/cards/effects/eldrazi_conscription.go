package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eldrazi Conscription — Kindred Enchantment — Eldrazi Aura {8}:
//
//	"Enchant creature
//	 Enchanted creature gets +10/+10 and has trample and annihilator 2."
//
// The grant goes through the cumulative keyword append, so a creature
// that already has annihilator keeps it and gains a second instance:
// each triggers separately (CR 702.86b, the 2018-12-07 ruling). The
// engine turns annihilator into an attack trigger (ADR 0113 §2). The
// Aura is an Eldrazi; the creature it enchants does not become one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3635e4e5-e655-45f4-a444-39756f55129a",
		Name:         "Eldrazi Conscription",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(10, 10),
			GrantToAttached("trample", "annihilator 2"),
		},
	})
}
