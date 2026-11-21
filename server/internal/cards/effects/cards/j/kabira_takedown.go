package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kabira Takedown // Kabira Plateau — Modal DFC. Front: Instant {1}{W}:
// "Kabira Takedown deals damage equal to the number of creatures you
// control to target creature or planeswalker." Back: Land (mdfc_lands.go).
//
// The count is taken as the spell resolves (CR 608.2h), from the
// creatures the caster controls then. A target that is no longer legal
// means the spell does nothing. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0bb73c07-0220-4ba9-8d85-3c357c223833",
		Name:         "Kabira Takedown",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := len(effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.YouControl())))
			if n <= 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: n}.Apply(ctx)
		},
	})
}
