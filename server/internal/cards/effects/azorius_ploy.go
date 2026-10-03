package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azorius Ploy — Instant {1}{W}{W}{U}:
//
//	"Prevent all combat damage target creature would deal this turn.
//	 Prevent all combat damage that would be dealt to target creature this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): two target clauses, two shields.
// The first names its creature as the source, pinned as the spell
// resolves (CR 400.7); the second protects its creature from combat
// damage from any source. The two targets may be the same creature (its
// ruling; CR 601.2c). A target that has become illegal is skipped and
// the other sentence still happens (CR 608.2b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "b9b58c3a-5a76-4b6a-884b-74f54f0ca53c",
		Name:         "Azorius Ploy",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature (its combat damage is prevented)"),
			TargetCreature("target creature (combat damage to it is prevented)"),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (PreventDamageFromSource{From: t.ID, CombatOnly: true, Protect: ShieldAnything}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
				return PreventDamageFromSource{CombatOnly: true, Protect: ShieldObject(t.ID)}.Apply(ctx)
			}
			return nil
		},
	})
}
