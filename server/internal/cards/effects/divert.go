package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Divert — Instant {U}:
//
//	"Change the target of target spell with a single target unless
//	 that spell's controller pays {2}."
//
// The reason this card waited for #2365: the "unless" is a CR 118.12
// tax on a spell that is STILL ON THE STACK. Written with a plain
// PayUnless the targeted spell resolves while its controller is still
// being asked, so the redirect arrives after the spell has already
// done its work. GuardedPayUnless is the shape that stops the table
// until the payer answers (the engine derives the halt from the
// guarded object, as it does for Spell Pierce and ward).
//
// Pay {2} and nothing changes. Decline — or answer "pay" without the
// mana — and Divert's controller is offered the CR 115.7b change
// (mandatory, exactly one slot) through the same ChangeTargets
// primitive Misdirection and Bolt Bend use; a target with nowhere
// else legal to go is unchanged, per CR 115.7a. The new target is
// chosen by Divert's controller, not by the payer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "561aed01-322b-4699-b7b0-f4416075aa9b",
		Name:         "Divert",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell with a single target", HasASingleTarget()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			return GuardedPayUnless{
				StackID:  target,
				Cost:     "{2}",
				Question: "Divert — pay {2} or the target of your spell is changed",
				OnDecline: func(ctx *Context) error {
					return ChangeTargets{
						StackID: target,
						Policy:  game.RetargetChangeOne,
						Reason:  "Divert — change the target",
					}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
