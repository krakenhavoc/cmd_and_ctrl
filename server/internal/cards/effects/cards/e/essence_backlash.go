package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Essence Backlash — Instant {2}{U}{R} (Return to Ravnica, S82 OLD):
//
//	"Counter target creature spell. Essence Backlash deals damage equal to
//	 that spell's power to its controller."
//
// Psychic Barrier's counter-then-hurt with damage in place of life loss. The
// spell's controller and its power (as it sat on the stack, CR 608.2h) are
// read before it is countered, so the damage is dealt even though the spell
// has gone. A spell that can't be countered is still not countered but its
// controller is still dealt the damage. A spell that has left the stack
// makes the target illegal and Essence Backlash does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f51ed0b4-c144-458e-b97c-618b9e1cf25d",
		Name:         "Essence Backlash",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target creature spell", effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				spell := ctx.Game.StackItemForEffect(t.ID)
				if spell == nil {
					return nil
				}
				controller := spell.Controller
				power := 0
				if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok {
					power = ctx.Game.PowerAnywhereForEffect(c)
				}
				if err := (effects.CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if power <= 0 {
					return nil
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: controller, Amount: power}.Apply(ctx)
			}
			return nil
		},
	})
}
