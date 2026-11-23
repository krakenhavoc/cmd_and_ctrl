package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// You're Confronted by Robbers — Instant {3}{W} (Commander Legends: Battle
// for Baldur's Gate, S82 GAP #5948):
//
//	"Choose one —
//	 • Stall for Time — Tap up to three target creatures.
//	 • Call for Aid — Create three 1/1 white Soldier creature tokens."
//
// Two bullets, each with its own body (ModeDoing). Stall for Time taps every
// pick that is still legal as the spell resolves (Frost Breath's tap, at three
// targets); Call for Aid makes three Soldiers for the caster.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d10ad59-e90d-4726-bd55-5680d7c21638",
		Name:         "You're Confronted by Robbers",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Stall for Time — Tap up to three target creatures.",
				effects.TargetCreature("up to three target creatures").WithCount(0, 3),
				func(_ *game.StackItem, ctx *effects.Context, occurrence int) error {
					for _, t := range ctx.ModeTargets(occurrence) {
						if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (effects.TapTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Call for Aid — Create three 1/1 white Soldier creature tokens.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 3}.Apply(ctx)
				}), game.Purpose{Tokens: 3, CreatureTokens: 3}),
		),
	})
}
