package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Feint — Instant {R}:
//
//	"Tap all creatures blocking target attacking creature. Prevent all combat damage that would be dealt this turn by that creature and each creature blocking it."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the blockers are read as the
// spell resolves, tapped (which does not remove them from combat or stop
// their damage, CR 506.4b), and then the attacker and each of them gets
// a combat-damage shield with itself as the source, pinned now (CR
// 400.7). One record per source: a damage event has one source, so no
// event meets two of them and nothing differs from one effect naming
// them all.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "1bb8fe05-abb3-40a8-9e80-5d99ed0e4284",
		Name:         "Feint",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking creature", AttackingCreature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				blockers := blockersOf(ctx.Game, t.ID)
				for _, b := range blockers {
					if err := (TapTarget{Target: b}).Apply(ctx.asGroupMember()); err != nil {
						return err
					}
				}
				for _, src := range append([]uuid.UUID{t.ID}, blockers...) {
					if err := (PreventDamageFromSource{From: src, CombatOnly: true, Protect: ShieldAnything}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}
			return nil
		},
	})
}
