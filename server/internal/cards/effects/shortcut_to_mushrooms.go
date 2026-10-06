package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shortcut to Mushrooms — Enchantment {1}{G}:
//
//	"When this enchantment enters, the Ring tempts you.
//	 At the beginning of your end step, if a permanent you controlled
//	 left the battlefield this turn, put a +1/+1 counter on target
//	 creature you control."
//
// The end-step "if" is revolt (revolt.go, #2148): an intervening if,
// checked as the step begins and again as the ability resolves
// (CR 603.4). The target is chosen when it triggers, and the trigger is
// removed when you control no creature (CR 603.3d).
//
// No simplification.
func init() {
	counter := AtYourEndStepIfRevolt("Shortcut to Mushrooms — put a +1/+1 counter on target creature you control",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}
			}
			return nil
		})
	counter.Targets = TargetCreature("target creature you control", YouControl())
	Register(Spec{
		OracleID:     "2dc6637a-0b51-462b-9f18-571146bb4558",
		Name:         "Shortcut to Mushrooms",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Shortcut to Mushrooms — the Ring tempts you", Do(TheRingTemptsYou{})),
			counter,
		},
	})
}
