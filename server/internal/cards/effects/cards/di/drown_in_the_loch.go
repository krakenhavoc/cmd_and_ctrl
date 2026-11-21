package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0f264e5b-264e-4e97-9a8d-8ae1d6a286ce",
		Name:         "Drown in the Loch",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Counter target spell with mana value less than or equal to the number of cards in its controller's graveyard.",
				effects.TargetSpell("target spell with mana value at most its controller's graveyard size", effects.B13ManaValueAtMostControllersGraveyard())),
			effects.Mode("Destroy target creature with mana value less than or equal to the number of cards in its controller's graveyard.",
				effects.TargetCreature("target creature with mana value at most its controller's graveyard size", effects.B13ManaValueAtMostControllersGraveyard())),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			switch {
			case ctx.HasMode(0):
				if ctx.Game.StackItemForEffect(target) == nil {
					return nil
				}
				return effects.CounterTarget{StackID: target}.Apply(ctx)
			case ctx.HasMode(1):
				return effects.DestroyTarget{Target: target}.Apply(ctx)
			}
			return nil
		},
	})
}
