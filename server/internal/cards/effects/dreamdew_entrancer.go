package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreamdew Entrancer —
//
// "Reach When this creature enters, tap up to one target creature and put
// three stun counters on it. If you control that creature, draw two cards."
func init() {
	Register(Spec{
		OracleID:        "b6b2d63f-5b8c-47ee-8712-58596e0e9e94",
		Name:            "Dreamdew Entrancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{Targeting(
			WhenThisEnters("Dreamdew Entrancer — tap and stun target creature", Do(
				// The draw waits for a prompt the counters ask (CR 608.2c, #4155).
				Instruction(func(ctx *Context) error {
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					if err := (TapTarget{Target: ts[0].ID}).Apply(ctx); err != nil {
						return err
					}
					return AddCounter{Target: ts[0].ID, Kind: game.CounterStun, N: 3}.Apply(ctx)
				}),
				Instruction(func(ctx *Context) error {
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					c, ok := ctx.Game.LookupCardForEffect(ts[0].ID)
					if ok && c.Controller == ctx.Item.Controller {
						return DrawCards{N: 2}.Apply(ctx)
					}
					return nil
				}),
			)),
			TargetCreature("up to one target creature").WithCount(0, 1),
		)},
	})
}
