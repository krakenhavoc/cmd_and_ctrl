package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reroute — Instant {1}{R}:
//
//	"Change the target of target activated ability with a single
//	 target. (Mana abilities can't be targeted.)
//	 Draw a card."
//
// The clause admits an activated ability only — not a triggered
// ability and not a spell — with exactly one target slot. A mana
// ability never reaches the stack, so the reminder text needs no
// predicate. If the ability is countered or gone in response, Reroute
// has no legal target and does nothing at all, draw included (CR
// 608.2b); if it is still there but nowhere else is a legal
// destination, the target stays and the card is still drawn.
func init() {
	Register(Spec{
		OracleID:     "997f5a5a-9ee1-4b66-9ffc-d1a2e996e615",
		Name:         "Reroute",
		Completeness: CompletenessFull,
		Targets: TargetSpellOrAbility("target activated ability with a single target",
			AnActivatedAbilityItem(), ItemHasASingleTarget()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := retargetTheTargetedItem(ctx, "Reroute — change the target", game.RetargetChangeOne, false); err != nil {
				return err
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
