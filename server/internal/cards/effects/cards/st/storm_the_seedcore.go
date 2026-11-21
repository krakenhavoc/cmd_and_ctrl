package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm the Seedcore — Sorcery {2}{G}{G} (March of the Machine):
//
//	"Distribute four +1/+1 counters among up to four target creatures you
//	 control. Creatures you control gain vigilance and trample until end of
//	 turn."
//
// Blessings of Nature's division (PutDividedCounters) over up to four
// creatures its controller controls, each target at least one counter
// (CR 601.2d). The keywords go to the creatures that controller controls as
// the spell resolves, a set fixed then (CR 611.2c): the targets and every
// other creature they control, including one that was an illegal target by
// then; a creature that arrives later gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae7643ea-ef1c-43a3-9605-4e8585239ea3",
		Name:         "Storm the Seedcore",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCreature("up to four target creatures you control", effects.YouControl()).WithCount(0, 4).Dividing(effects.Divide(4)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.PutDividedCounters(ctx, game.CounterPlusOne); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"vigilance", "trample"},
				Label:    "Storm the Seedcore — creatures you control gain vigilance and trample until end of turn",
			}.Apply(ctx)
		},
	})
}
