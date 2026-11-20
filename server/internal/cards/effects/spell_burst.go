package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spell Burst — Instant {X}{U} (Time Spiral, #5596):
//
//	"Buyback {3} (You may pay an additional {3} as you cast this spell. If
//	 you do, put this card into your hand as it resolves.)
//	 Counter target spell with mana value X."
//
// The target's mana value is checked against the announced X when it is
// chosen and again as Spell Burst resolves (CR 608.2b). Buyback is the
// shared constructor (CR 702.27a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "141d9b88-5326-4e3d-8f16-5293246fab88",
		Name:          "Spell Burst",
		Completeness:  CompletenessFull,
		Targets:       TargetSpell("target spell with mana value X").WithManaValueEqualsX(),
		OptionalCosts: []game.AdditionalCost{Buyback("{3}")},
		OnResolve:     counterTheTargetSpell,
	})
}
