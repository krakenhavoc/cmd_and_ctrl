package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boon Satyr — Enchantment Creature — Satyr {1}{G}{G}, 4/2 (EDHREC
// rank 16799):
//
//	"Flash
//	 Bestow {3}{G}{G} (If you cast this card for its bestow cost, it's
//	 an Aura spell with enchant creature. It becomes a creature again
//	 if it's not attached.)
//	 Enchanted creature gets +4/+2."
//
// A combat trick that leaves a body behind. Flash is the card's own
// keyword, so it opens both casts at instant speed: the bestowed one
// mid-combat on an attacker, and the creature one at end of turn. The
// bestow half is ADR 0141 (#2862).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "aedcd2fb-813a-4915-b7b9-ac7479863462",
		Name:            "Boon Satyr",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		AlternativeCosts: []game.AlternativeCost{
			Bestow("{3}{G}{G}"),
		},
		Static: []game.StaticAbility{
			PumpAttached(4, 2),
		},
	})
}
