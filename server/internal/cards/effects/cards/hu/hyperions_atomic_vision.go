package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hyperion's Atomic Vision — Instant {1}{W} (Marvel Super Heroes
// Commander):
//
//	"As an additional cost to cast this spell, you may behold a Hero.
//	 (You may choose a Hero you control or reveal a Hero card from your
//	 hand.)
//	 Destroy target tapped creature. If a Hero was beheld, scry 2."
//
// The behold is an optional additional cost (OptionalBehold, #4642). Its
// answer is a fact about the announcement (CR 701.4b), so the scry reads
// ctx.WasBeheld() and not the card that was chosen, which may be gone or
// no longer a Hero by then. The scry is not conditional on the creature
// dying: a target that went illegal in response fizzles the whole spell
// (CR 608.2b), as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "e1e82283-01ec-447f-b0d4-5bd4619ab3d8",
		Name:          "Hyperion's Atomic Vision",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.OptionalBehold("a", "Hero")},
		Targets:       effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.WasBeheld() {
				return nil
			}
			return effects.Scry{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
