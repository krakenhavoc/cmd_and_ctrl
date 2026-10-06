package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cosmotronic Wave — Sorcery {3}{R}:
//
//	"Cosmotronic Wave deals 1 damage to each creature your opponents
//	 control. Creatures your opponents control can't block this turn."
//
// Two sentences, two different kinds of effect, and they treat a
// creature that arrives later differently:
//
//   - The damage is one-shot. It hits the creatures there as the spell
//     resolves (Blazing Volley's damageEachMatching), and the ones that
//     took lethal die at the next state-based-action check.
//   - "Can't block" is a rules effect, so CR 611.2c does not lock its
//     set (#1650). A creature an opponent flashes in, or gains control
//     of, after the spell resolved can't block either. RestrictUntilEOT
//     reads game.ScopeOpponentsCreatures live until cleanup.
//
// Hazardous Blast prints the same text and shares
// pingThenOpponentsCantBlock().
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "079ac521-9414-46c1-acd9-2ff2ff047d22",
		Name:         "Cosmotronic Wave",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}},
		Completeness: CompletenessFull,
		OnResolve:    pingThenOpponentsCantBlock("Cosmotronic Wave"),
	})
}

// pingThenOpponentsCantBlock is "<this> deals 1 damage to each creature
// your opponents control. Creatures your opponents control can't block
// this turn." — Cosmotronic Wave and Hazardous Blast.
func pingThenOpponentsCantBlock(name string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		if err := damageEachMatching(ctx, And(Creature(), OpponentControls()), 1); err != nil {
			return err
		}
		return RestrictUntilEOT{
			Scope:        game.ScopeOpponentsCreatures,
			Restrictions: game.CantBlock,
			Label:        name + " — creatures your opponents control can't block",
		}.Apply(ctx)
	}
}
