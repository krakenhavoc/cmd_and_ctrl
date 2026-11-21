package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disrupting Shoal — Instant — Arcane {X}{U}{U} (Betrayers of Kamigawa,
// #5602):
//
//	"You may exile a blue card with mana value X from your hand rather
//	 than pay this spell's mana cost.
//	 Counter target spell if its mana value is X."
//
// The pitch is Shining Shoal's (bokB1ShoalPitch): X is announced and the
// exiled blue card's mana value must equal it. The target is any spell;
// the mana value test is made as Disrupting Shoal resolves, against the
// target's mana value on the stack with its own announced X counted (CR
// 202.3e), so a spell of another mana value stays where it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "2912d93d-ed51-41ef-b64f-3cb6581c09e5",
		Name:             "Disrupting Shoal",
		XMatters:         true,
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetSpell("target spell"),
		AlternativeCosts: []game.AlternativeCost{effects.BokB1ShoalPitch("blue", "U")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			t := targets[0]
			c, ok := ctx.Game.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			if mv, ok := ctx.Game.ManaValueForEffect(c); !ok || mv != ctx.X() {
				return nil
			}
			return effects.CounterTarget{StackID: t.ID}.Apply(ctx)
		},
	})
}
