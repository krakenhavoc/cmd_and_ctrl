package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sanctified Charge — Instant {4}{W} (Magic 2015, #5556):
//
//	"Creatures you control get +2/+1 until end of turn. White creatures you
//	 control also gain first strike until end of turn."
//
// Both sets are fixed as the spell resolves (CR 611.2c): the creatures you
// control then get the pump, and those of them that are white then also get
// first strike. A creature that enters later, or turns white later, gets
// neither.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e89b3133-4cd6-44fb-ae86-a01b86bb970c",
		Name:         "Sanctified Charge",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mine := effects.And(effects.Creature(), effects.YouControl())
			if err := (effects.BoostUntilEOT{Match: mine, Power: 2, Toughness: 1, Label: "Sanctified Charge — creatures you control get +2/+1"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Match:    effects.And(mine, effects.OfColor("W")),
				Keywords: []string{"first strike"},
				Label:    "Sanctified Charge — white creatures you control gain first strike",
			}.Apply(ctx)
		},
	})
}
