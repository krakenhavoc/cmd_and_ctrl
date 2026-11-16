package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Almost Perfect — Enchantment — Aura {4}{G}{W} (Fallout, #5469):
//
//	"Enchant creature
//	 Enchanted creature has base power and toughness 9/10 and has
//	 indestructible."
//
// Coerced to Kill's base-P/T Aura at 9/10 (SetAttachedBasePT, layer
// 7b), so a +1/+1 counter or a pump still lands on top of the 9/10,
// and indestructible granted to the host (GrantToAttached, layer 6).
// Both end when the Aura leaves or moves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "617eae63-b5f8-4014-8269-7c66f4388ac7",
		Name:         "Almost Perfect",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			SetAttachedBasePT(9, 10),
			GrantToAttached("indestructible"),
		},
	})
}
