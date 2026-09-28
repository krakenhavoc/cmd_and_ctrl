package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hunt Down — Sorcery, {G}:
//
//	"Target creature blocks target creature this turn if able."
//
// #1684 leftover: two "target creature" slots, positional like Arc
// Trail (the ORDER matters, and the engine already refuses a
// duplicate pick within one clause) — the first target is the
// would-be blocker, the second is the object it must block if able.
// The second target's current object (instance and epoch, CR 400.7)
// is read at resolution via PermanentRefForEffect and pinned into a
// BlocksAttackerUntilEOT record, the same shape Provoke and
// Grappling Hook use. If the second target never attacks this turn,
// the requirement asks nothing (CR 509.1c: it is only judged when
// that object is a declared attacker).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "29b0f3bd-d9bd-4c54-b4c3-9c2e01720e34",
		Name:         "Hunt Down",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature").WithCount(2, 2),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) < 2 {
				return nil
			}
			blocker, named := item.Targets[0], item.Targets[1]
			if !ctx.IsTargetLegal(blocker) || !ctx.IsTargetLegal(named) {
				return nil
			}
			ref, ok := ctx.Game.PermanentRefForEffect(named.ID)
			if !ok {
				return nil
			}
			return BlocksAttackerUntilEOT{
				Blocker:  blocker.ID,
				Attacker: ref,
				Label:    "Hunt Down — blocks the other target creature if able",
			}.Apply(ctx)
		},
	})
}
