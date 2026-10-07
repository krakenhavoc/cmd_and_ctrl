package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sewers of Estark — Instant {2}{B}{B}:
//
//	"Choose target creature. If it's attacking, it can't be blocked this
//	 turn. If it's blocking, prevent all combat damage that would be
//	 dealt this combat by it and each creature it's blocking."
//
// ADR 0108 amendment 2026-10-07 (#2027): the prevention lasts "this
// combat" (game.UntilEndOfCombat), the combat phase in progress, rather
// than the turn, so an additional combat phase is not shielded. One shield
// record per source (a damage event has one source), the blocker and each
// attacker it blocks (Card.BlockedAttackers, a creature that can block
// several), each pinned as the object it is now (CR 400.7). The creature
// is attacking or blocking, never both; a target that has gone is skipped
// (CR 608.2b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "2482f98a-eadf-450d-9e73-fc441b572862",
		Name:         "Sewers of Estark",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				switch {
				case c.AttackingTarget != uuid.Nil:
					return RestrictUntilEOT{
						Target:       t.ID,
						Restrictions: game.CantBeBlocked,
						Label:        "Sewers of Estark — can't be blocked",
					}.Apply(ctx)
				case c.BlockingTarget != uuid.Nil:
					for _, src := range append([]uuid.UUID{t.ID}, c.BlockedAttackers()...) {
						if err := (PreventDamageFromSource{
							From: src, CombatOnly: true, Protect: ShieldAnything, Lasts: ShieldThisCombat,
							Label: "Sewers of Estark — prevent all combat damage this combat",
						}).Apply(ctx); err != nil {
							return err
						}
					}
				}
				return nil
			}
			return nil
		},
	})
}
