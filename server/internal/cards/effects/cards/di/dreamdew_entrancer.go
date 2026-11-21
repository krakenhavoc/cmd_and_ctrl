package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreamdew Entrancer —
//
// "Reach When this creature enters, tap up to one target creature and put
// three stun counters on it. If you control that creature, draw two cards."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b6b2d63f-5b8c-47ee-8712-58596e0e9e94",
		Name:            "Dreamdew Entrancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.WhenThisEnters("Dreamdew Entrancer — tap and stun target creature", effects.Do(
				// The draw waits for a prompt the counters ask (CR 608.2c, #4155).
				effects.Instruction(func(ctx *effects.Context) error {
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					if err := (effects.TapTarget{Target: ts[0].ID}).Apply(ctx); err != nil {
						return err
					}
					return effects.AddCounter{Target: ts[0].ID, Kind: game.CounterStun, N: 3}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					c, ok := ctx.Game.LookupCardForEffect(ts[0].ID)
					if ok && c.Controller == ctx.Item.Controller {
						return effects.DrawCards{N: 2}.Apply(ctx)
					}
					return nil
				}),
			)),
			effects.TargetCreature("up to one target creature").WithCount(0, 1),
		)},
	})
}
