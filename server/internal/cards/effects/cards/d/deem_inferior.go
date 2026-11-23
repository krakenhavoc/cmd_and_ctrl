package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deem Inferior — Sorcery {3}{U} (Modern Horizons 3):
//
//	"This spell costs {1} less to cast for each card you've drawn this turn.
//	 The owner of target nonland permanent puts it into their library second
//	 from the top or on the bottom."
//
// Temporal Cleansing's resolution (the permanent's OWNER chooses the lane,
// PutIntoLibraryAtDepthOrBottom) with a self cost reduction read off the
// caster's per-turn tally of cards drawn, as Fractal Anomaly reads it: every
// card drawn this turn counts, whatever drew it, and the reduction comes off
// the generic part only (CR 601.2f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0bd4a29-ac13-4d4f-8479-e0e80f8d846f",
		Name:         "Deem Inferior",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Not(effects.Land())),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessPerCardDrawnThisTurn(),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.PutIntoLibraryAtDepthOrBottom{
				Card:  item.Targets[0].ID,
				Depth: 2,
				Label: "Deem Inferior — put it into your library second from the top or on the bottom",
			}.Apply(ctx)
		},
	})
}
