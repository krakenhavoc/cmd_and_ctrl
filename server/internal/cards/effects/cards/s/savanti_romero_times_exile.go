package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savanti Romero, Time's Exile — {3}{B}{B} Legendary Creature — Demon Wizard
// 4/4 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Trample
//	 At the beginning of combat on your turn, put a +1/+1 counter on
//	 Savanti Romero. Then you draw X cards and lose X life, where X is
//	 the number of counters on Savanti Romero."
//
// Two steps of one resolution, in printed order (CR 608.2c): the counter
// goes on first, then X is counted. X is every counter on him, of any kind,
// read as the second step runs (CR 608.2h); if he has left the battlefield in
// response, no counter goes on and X is the counters he had as he last
// existed there. Drawing and losing life share the one X, and X of zero does
// neither.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6074632b-4520-4ae8-bcdc-26b1eff71af9",
		Name:            "Savanti Romero, Time's Exile",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Savanti Romero — put a +1/+1 counter on him, then draw X cards and lose X life",
				effects.Do(
					effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1},
					effects.Instruction(func(ctx *effects.Context) error {
						info, ok := ctx.SourcePermanent()
						if !ok {
							return nil
						}
						x := 0
						for _, n := range info.Counters {
							if n > 0 {
								x += n
							}
						}
						if x <= 0 {
							return nil
						}
						if err := (effects.DrawCards{Player: ctx.Controller(), N: x}).Apply(ctx); err != nil {
							return err
						}
						return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -x)
					}),
				)),
		},
	})
}
