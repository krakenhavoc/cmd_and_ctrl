package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Out Cold — Instant {3}{U} (Murders at Karlov Manor, #4469):
//
//	"This spell can't be countered. (This includes by the ward ability.)
//	 Tap up to two target creatures and put a stun counter on each of
//	 them. Investigate. (If a permanent with a stun counter would become
//	 untapped, remove one from it instead.)"
//
// Spec.CantBeCountered (Abrupt Decay's), so ward's counter does nothing
// to it either. Each still-legal target is tapped and gets a stun counter
// (CR 122.1d), an already-tapped one included; the Clue comes whether or
// not any target was chosen or is still legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2d16d51b-47e0-4cdc-83dd-a805ba50bd83",
		Name:            "Out Cold",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		Targets:         effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, id := range effects.LegalTargetCardIDs(ctx) {
				if err := (effects.TapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.AddCounter{Target: id, Kind: game.CounterStun, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.Investigate{N: 1}.Apply(ctx)
		},
	})
}
