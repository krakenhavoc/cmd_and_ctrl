package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haste Magic — Instant {1}{R}:
//
//	"Target creature gets +3/+1 and gains haste until end of turn.
//	 Exile the top card of your library. You may play it until your
//	 next end step."
//
// The first two clauses are until-end-of-turn effects locked on the
// target now (CR 611.2c); the impulse window is
// game.UntilYourNextEndStep (#2373), so cast in your own main phase
// the card must be played before this turn's end step, and cast on an
// opponent's turn it stays playable to the end step of your next turn.
// If the target is illegal on resolution the spell does not resolve
// at all (CR 608.2b), so nothing is exiled, which the engine enforces.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5249dd0d-3bda-452e-ac6e-774c0ebc2e41",
		Name:         "Haste Magic",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 1, "haste")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				target := item.Targets[0].ID
				if err := (effects.BoostUntilEOT{Target: target, Power: 3, Toughness: 1, Label: "Haste Magic — +3/+1"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"haste"}, Label: "Haste Magic — haste"}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.ExileTopNUntilYourNextEndStep(ctx, 1)
		},
	})
}
