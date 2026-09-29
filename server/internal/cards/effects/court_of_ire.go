package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Court of Ire — Enchantment {3}{R}{R}:
//
//	"When this enchantment enters, you become the monarch.
//	 At the beginning of your upkeep, this enchantment deals 2 damage
//	 to any target. If you're the monarch, it deals 7 damage instead."
//
// The red Court (#1722). The target is chosen as the trigger goes on
// the stack (CR 603.3d) and re-checked as it resolves (CR 608.2b); the
// amount is read at resolution, because "if you're the monarch, … 7
// instead" is a clause of the effect. The Court is the source of the
// damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13f58292-9b78-4cd1-a16e-b1779a170d33",
		Name:         "Court of Ire",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Court of Ire"),
			Targeting(AtYourUpkeep("Court of Ire — 2 damage to any target, or 7 if you're the monarch", courtOfIreUpkeep),
				TargetAny()),
		},
	})
}

// courtOfIreUpkeep deals 2, or 7 for the monarch, to the chosen target.
func courtOfIreUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	amount := 2
	if YoureTheMonarch(g, item.Controller) {
		amount = 7
	}
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: targets[0].ID, Amount: amount}.Apply(ctx)
}
