package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Merciless Eviction — Sorcery {4}{W}{B}:
//
//	"Choose one —
//	 • Exile all artifacts.
//	 • Exile all creatures.
//	 • Exile all enchantments.
//	 • Exile all planeswalkers."
//
// EXILE rather than destroy is the reason to play it over Wrath of
// God: no dies-triggers, no regeneration, no recursion, and
// indestructible doesn't help. Against the aristocrats payoffs in the
// catalog that difference decides games — a Blood Artist board wiped
// by Damnation drains the table, wiped by Eviction it doesn't.
//
// Four untargeted modes, so ChooseOne carries no target clauses: the
// engine validates the pick at announce and OnResolve reads it back
// with ctx.HasMode in printed order (CR 700.2c).
func init() {
	Register(Spec{
		OracleID: "3c8d4999-e18b-48d8-8ed9-f2feaa38300d",
		Name:     "Merciless Eviction",
		Modes: ChooseOne(
			Mode("Exile all artifacts."),
			Mode("Exile all creatures."),
			Mode("Exile all enchantments."),
			Mode("Exile all planeswalkers."),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			// Mode index → that sweep's type test, in printed order.
			sweeps := []func(game.Card) bool{
				func(c game.Card) bool { return c.IsArtifact() },
				func(c game.Card) bool { return c.IsCreature() },
				func(c game.Card) bool { return c.IsEnchantment() },
				func(c game.Card) bool { return c.IsPlaneswalker() },
			}
			for i, matches := range sweeps {
				if !ctx.HasMode(i) {
					continue
				}
				// Snapshot the IDs before exiling anything: ExileTarget
				// removes cards from the battlefield zone we would
				// otherwise be ranging over, and an exile-triggered
				// ability could add one mid-sweep.
				var doomed []uuid.UUID
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if matches(c) {
						doomed = append(doomed, c.InstanceID)
					}
				}
				for _, id := range doomed {
					if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
