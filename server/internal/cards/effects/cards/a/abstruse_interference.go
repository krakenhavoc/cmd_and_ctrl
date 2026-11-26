package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abstruse Interference — {2}{U} Instant (Oath of the Gatewatch, S82 OLD
// b296):
//
//	"Devoid (This card has no color.)
//	 Counter target spell unless its controller pays {1}. You create a 1/1
//	 colorless Eldrazi Scion creature token. It has "Sacrifice this token: Add
//	 {C}.""
//
// Miscalculation's counter-unless-pays and Brood Monitor's Scion. The token is
// made whether or not the spell's controller pays, but only if the target was
// still legal as this resolved (CR 608.2b): a spell that fizzles does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e8ebf5cb-3a26-4bbd-812b-414c129e4bb7",
		Name:            "Abstruse Interference",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         game.Purpose{Tokens: 1},
		Targets:         effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			if err := (effects.CounterUnlessPaid{
				StackID:  t.ID,
				Cost:     "{1}",
				Question: "Abstruse Interference — pay {1} or your spell is countered",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.B28EldraziScionToken(), N: 1}.Apply(ctx)
		},
	})
}
