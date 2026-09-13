package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drown in the Loch — Instant {U}{B} (EDHREC rank 1441):
//
//	"Choose one —
//	 • Counter target spell with mana value less than or equal to the
//	   number of cards in its controller's graveyard.
//	 • Destroy target creature with mana value less than or equal to
//	   the number of cards in its controller's graveyard."
//
// A counterspell or a removal spell, both scaled by the opponent's
// own graveyard. The clause is one predicate shared by the two modes
// (b13ManaValueAtMostControllersGraveyard): on the stack the mana
// value includes X (CR 202.3e) and the controller is the spell's; on
// the battlefield X is zero and the controller is the permanent's.
// It is a target predicate, so it is checked at announce and again
// at resolution — a graveyard exiled in response takes the target
// out and the spell fizzles, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0f264e5b-264e-4e97-9a8d-8ae1d6a286ce",
		Name:         "Drown in the Loch",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Counter target spell with mana value less than or equal to the number of cards in its controller's graveyard.",
				TargetSpell("target spell with mana value at most its controller's graveyard size", b13ManaValueAtMostControllersGraveyard())),
			Mode("Destroy target creature with mana value less than or equal to the number of cards in its controller's graveyard.",
				TargetCreature("target creature with mana value at most its controller's graveyard size", b13ManaValueAtMostControllersGraveyard())),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			switch {
			case ctx.HasMode(0):
				if ctx.Game.StackItemForEffect(target) == nil {
					return nil
				}
				return CounterTarget{StackID: target}.Apply(ctx)
			case ctx.HasMode(1):
				return DestroyTarget{Target: target}.Apply(ctx)
			}
			return nil
		},
	})
}
