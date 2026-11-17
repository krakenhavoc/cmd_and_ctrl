package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reveka, Wizard Savant — Legendary Creature — Dwarf Wizard {2}{U}{U}, 0/1
// (Homelands, #5753):
//
//	"{T}: Reveka deals 2 damage to any target and doesn't untap during
//	 your next untap step."
//
// The second half is ADR 0058's one-shot next-untap-step mark on Reveka,
// keyed to the ACTIVATOR's untap step ("your", UntapSkip.Player), not to
// whoever controls Reveka then. The ruling: "Reveka's effect causes it not
// to untap during your next untap step. If another player gains control
// of it, it will untap during that player's next untap step." The mark is
// used up by that step even if Reveka is untapped by then (CR 502.3: the
// step only declines to untap it), and it goes with Reveka if she leaves
// the battlefield (CR 400.7).
//
// A target gone by resolution makes the whole ability do nothing (CR
// 608.2b), so Reveka is not marked either. A Reveka that left and came
// back in response is a new object and is not marked (#1432).
//
// No simplification.
func init() {
	const label = "{T}: Reveka deals 2 damage to any target and doesn't untap during your next untap step."
	Register(Spec{
		OracleID:     "7085d5c6-da96-465e-90cf-8469fcacb474",
		Name:         "Reveka, Wizard Savant",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   label,
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Cost:    TapCost(),
			Targets: TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := DealDamageToTheTarget(2)(g, item); err != nil {
					return err
				}
				ctx := NewContext(g, item)
				return DoesntUntapNextUntapStep{
					Targets: []uuid.UUID{ctx.Source()},
					Player:  ctx.Controller(),
					Label:   "Reveka, Wizard Savant — doesn't untap during your next untap step",
				}.Apply(ctx)
			},
		}},
	})
}
