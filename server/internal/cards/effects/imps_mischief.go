package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Imp's Mischief — Instant {1}{B}:
//
//	"Change the target of target spell with a single target. You lose
//	 life equal to that spell's mana value."
//
// The black Misdirection, at a third of the cost and with the price
// paid in life instead of a card. Named on the "Stack-item retarget"
// seam row (#298) and unblocked by the same primitive
// (docs/decisions/0019-structured-targeting.md, the 2026-09-22
// amendment).
//
// Two notes on how it behaves here:
//
//   - THE MANA VALUE IS READ BEFORE THE PROMPT OPENS, off the spell
//     on the stack, so {X} counts as the X that was announced
//     (CR 202.3e, ManaValueForEffect). It cannot change afterwards —
//     the redirected spell's identity is not what the prompt
//     alters — so reading it first costs nothing.
//   - THE LIFE IS LOST EVEN IF THE TARGET CANNOT MOVE. "Change the
//     target" and "you lose life" are two sentences of one
//     resolution, and CR 115.7a's "the original target is unchanged"
//     is not a failure to resolve. An Imp's Mischief pointed at a
//     Lightning Bolt aimed at the only creature on the board costs
//     its controller one life and does nothing else, which is the
//     printed outcome.
//
// The life loss runs while the redirect prompt is still open rather
// than after it is answered. Nothing depends on the order — neither
// sentence reads the other's result — and the alternative would be
// parking the rest of the card in a continuation for no gain.
//
// No caveat: the card prints "target SPELL", so the half the engine
// cannot reach (targeting an ability on the stack, ADR 0065's open
// item) is a half this card never had.
func init() {
	Register(Spec{
		OracleID:     "30ec73ad-c7a1-4527-8dc6-44bdd65b9ba6",
		Name:         "Imp's Mischief",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell with a single target", HasASingleTarget()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			stackID := item.Targets[0].ID
			life := 0
			if c, ok := ctx.Game.LookupCardForEffect(stackID); ok {
				if mv, ok := ctx.Game.ManaValueForEffect(c); ok {
					life = mv
				}
			}
			if err := (ChangeTargets{
				StackID: stackID,
				Policy:  game.RetargetChangeOne,
				Reason:  "Imp's Mischief — change the target",
			}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: ctx.Controller(), Amount: -life}.Apply(ctx)
		},
	})
}
