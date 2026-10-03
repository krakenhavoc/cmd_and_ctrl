package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hazduhr the Abbot — Legendary Creature — Human Cleric {3}{W}{W}, 2/5:
//
//	"{X}, {T}: The next X damage that would be dealt this turn to target
//	 white creature you control is dealt to Hazduhr instead."
//
// ADR 0108 §9 (#1905): a charged redirection (CR 615.7) of X, fixed as
// the ability was activated, from the target to Hazduhr as it is now.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "35844d8b-8f68-4248-b921-e766757a9e26",
		Name:         "Hazduhr the Abbot",
		XMatters:     true,
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{X}, {T}: The next X damage that would be dealt this turn to target white creature you control is dealt to Hazduhr instead.",
			Cost:    Plus(ManaCost("{X}"), TapCost()),
			Targets: TargetCreature("target white creature you control", OfColor("W"), YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if ctx.X() < 1 {
					return nil
				}
				return RedirectDamage{Protect: ShieldTheTarget, Amount: ctx.X(), To: RedirectToThis}.Apply(ctx)
			},
		}},
	})
}
