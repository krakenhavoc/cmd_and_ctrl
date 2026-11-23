package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warg Tactics — Instant {1}{G} (The Hobbit):
//
//	"Choose one —
//	 • Destroy target creature with flying.
//	 • Put a +1/+1 counter on target creature you control. It gains trample
//	   and hexproof until end of turn."
//
// #4454, S76. Crushing Canopy's flying-destroy mode; the second mode is the
// counter and then the keyword grant as two steps, so the grant waits for any
// prompt the counter asks (CR 608.2c, Pawpatch Formation's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1663b51-37a1-45a0-b477-2d3e7dd0828e",
		Name:         "Warg Tactics",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Destroy target creature with flying.",
				effects.TargetCreature("target creature with flying", effects.HasKeyword("flying"))),
			effects.Mode("Put a +1/+1 counter on target creature you control. It gains trample and hexproof until end of turn.",
				effects.TargetCreature("target creature you control", effects.YouControl())),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(0) {
					return nil
				}
				id, legal := effects.B16FirstLegalTargetCard(ctx)
				if !legal {
					return nil
				}
				return effects.DestroyTarget{Target: id}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(1) {
					return nil
				}
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(1) {
					return nil
				}
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"trample", "hexproof"},
					Label:    "Warg Tactics — trample and hexproof until end of turn",
				}.Apply(ctx)
			}),
		),
	})
}
