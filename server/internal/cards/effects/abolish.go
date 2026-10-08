package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abolish — Instant {1}{W}{W}:
//
//	"You may discard a Plains card rather than pay this spell's mana
//	 cost.
//	 Destroy target artifact or enchantment."
//
// ADR 0135 §2 (#2412): Disenchant with Snag's price. The Plains card is
// discarded as a cost with the spell already on the stack (CR 601.2a
// before 601.2h); any card with the land type Plains pays it.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:         "d6d7faa8-ecc9-408b-9c46-a0f62c74f567",
		Name:             "Abolish",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{DiscardInstead("a Plains card", HasSubtype("Plains"))},
		Targets:          TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve:        destroyTheTargetPermanent,
	})
}
