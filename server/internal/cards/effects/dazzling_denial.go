package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dazzling Denial — Instant {1}{U} (EDHREC rank 4197):
//
//	"Counter target spell unless its controller pays {2}. If you
//	 control a Bird, counter that spell unless its controller pays {4}
//	 instead."
//
// A Mana Leak that becomes a Force Spike-proof Mana Leak in a Bird
// deck. Two mana to tax four is a real counterspell in the late game;
// two mana to tax two is a turn-three answer that stops mattering by
// turn eight.
//
// Daze's machinery, with one branch:
//
//   - The tax is read at RESOLUTION, not at cast. "If you control a
//     Bird" is not an intervening-if and not a targeting restriction
//     — it is part of the effect, so a Bird that arrives while Dazzling
//     Denial is on the stack raises the tax to {4}, and a Bird killed
//     in response lowers it to {2}. Both are printed behaviour and
//     both are why the check is inside OnResolve rather than beside
//     the target clause.
//   - The prompt goes to the SPELL's controller, not to this card's,
//     and it is captured BEFORE anything touches the stack:
//     CounterTarget deletes the stack entry, so the payer has to be
//     read first (the note Daze carries, for the same reason).
//   - A "pay" the victim cannot fund degrades to a decline
//     server-side, so the counter still happens.
//
// "A Bird" is effective subtypes, so a changeling counts and a Bird
// token counts, which is exactly what the Bird deck this is printed
// for is doing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6d56bd32-f47a-4e54-a88e-b69d16283ea4",
		Name:         "Dazzling Denial",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			target := ctx.Game.StackItemForEffect(stackID)
			if target == nil {
				return nil
			}
			cost := "{2}"
			if b40ControlsA(ctx.Game, ctx.Controller(), And(Creature(), HasSubtype("Bird"))) {
				cost = "{4}"
			}
			return CounterUnlessPaid{
				StackID:  stackID,
				Cost:     cost,
				Question: "Dazzling Denial — pay " + cost + " or your spell is countered",
			}.Apply(ctx)
		},
	})
}
