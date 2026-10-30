package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "e1e82283-01ec-447f-b0d4-5bd4619ab3d8",
		Name:          "Hyperion's Atomic Vision",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{OptionalBehold("a", "Hero")},
		Targets:       TargetCreature("target tapped creature", tappedPermanent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.WasBeheld() {
				return nil
			}
			return Scry{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
