package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daze — Instant {1}{U}:
//
//	"You may return an Island you control to its owner's hand rather
//	 than pay this spell's mana cost.
//	 Counter target spell unless its controller pays {1}."
//
// The third shape of the non-mana alternative cost, after Force of
// Will's pitch and Snuff Out's life: a permanent goes back to hand.
//
// The bounce is a COST, so it happens at announce with the Daze
// already on the stack — the Island is in hand before the spell Daze
// is answering has resolved, and before its controller decides
// whether to pay the {1}. A version that bounced on resolution would
// let the victim see a tapped-out board and decline to pay against an
// opponent who had, in fact, just untapped a land's worth of mana.
//
// The "unless its controller pays {1}" half reuses the S19 pay-unless
// prompt (Rhystic Study's machinery): the prompt goes to the SPELL'S
// controller, not to Daze's, and a "pay" they cannot fund degrades to
// a decline server-side.
func init() {
	Register(Spec{
		OracleID:     "70486bee-6ee7-41ea-b834-8caf4699302b",
		Name:         "Daze",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		AlternativeCosts: []game.AlternativeCost{
			ReturnInstead(
				"Return an Island you control to its owner's hand",
				PermanentYouControl("an Island you control", HasSubtype("Island")),
				"an Island you control",
			),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			// Read the victim's controller BEFORE anything touches
			// the stack — CounterTarget deletes the StackMeta entry,
			// so the pay-unless prompt has to capture the payer now.
			target := ctx.Game.StackItemForEffect(stackID)
			if target == nil {
				return nil
			}
			victim := target.Controller
			return PayUnless{
				Chooser:  victim,
				Cost:     "{1}",
				Question: "Daze — pay {1} or your spell is countered",
				OnDecline: func(ctx *Context) error {
					return CounterTarget{StackID: stackID}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
