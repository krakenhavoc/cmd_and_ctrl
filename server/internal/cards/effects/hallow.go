package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hallow — Instant {W}:
//
//	"Prevent all damage target spell would deal this turn. You gain life equal to the damage prevented this way."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): a preventFromSource record whose
// source is the targeted spell, pinned on the stack as Hallow resolves.
// A permanent spell's shield covers the permanent it becomes, as the
// Shieldmage Elder ruling reads "target spell" (CR 609.7a's reading). The
// life gain is the shield's CR 615.5 follow-up, run once per damage
// instance with what it prevented (CR 615.13), so a spell that deals
// damage twice gains life twice.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "60b6a31f-043e-4de3-ad37-67eec02114f8",
		Name:         "Hallow",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return PreventDamageFromSource{From: t.ID, Protect: ShieldAnything, Then: preventedGainLifeBody}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
