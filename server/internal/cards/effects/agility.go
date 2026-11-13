package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agility — Enchantment — Aura {1}{R} (Mirage, #5633, #5784):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has flanking. (Whenever a creature
//	 without flanking blocks this creature, the blocking creature gets
//	 -1/-1 until end of turn.)"
//
// The shared Aura shapes (EnchantCreature, PumpAttached,
// GrantToAttached). The granted flanking is the keyword token, so the
// engine derives its trigger (game/flanking.go) as it does for a printed
// one. Flanking is cumulative (CR 702.25b): on a creature that already
// has flanking it is a second instance, and a second trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0f5f8b47-d104-4e29-b052-02d730d76ec0",
		Name:         "Agility",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached(game.KeywordFlanking),
		},
	})
}
