package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hidden Retreat — Enchantment {2}{W}:
//
//	"Put a card from your hand on top of your library: Prevent all damage that would be dealt by target instant or sorcery spell this turn."
//
// ADR 0108 §7 (#1904) and ADR 0109 §7 (#1902): the cost is ADR 0109
// PR 8's "put a card from your hand on top of your library", named at
// announce (CR 602.2b). The shield is a preventFromSource against a
// TARGETED source, pinned as the ability resolves with no prompt: every
// instance of that spell's damage this turn, to anything, is prevented.
// A spell gone from the stack by resolution is an illegal target, and
// the ability does nothing (CR 608.2b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "a116329a-343e-4f10-a122-38bf8b5ac2c8",
		Name:         "Hidden Retreat",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Put a card from your hand on top of your library: Prevent all damage that would be dealt by target instant or sorcery spell this turn.",
			Cost:    PutACardFromHandOnTop(),
			Targets: instantOrSorcerySpell("target instant or sorcery spell"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return PreventDamageFromSource{From: t.ID, Protect: ShieldAnything}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
