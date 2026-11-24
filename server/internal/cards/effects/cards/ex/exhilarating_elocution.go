package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Exhilarating Elocution — Sorcery {2}{W}{B} (Strixhaven: School of Mages,
// #5502):
//
//	"Put two +1/+1 counters on target creature you control. Other creatures
//	 you control get +1/+1 until end of turn."
//
// The counters go on the target (if still legal, CR 608.2b); the pump
// is for the creatures you control as it resolves, the target left out
// (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58acd10d-d81e-49ae-9e08-cff4c6f0c6a1",
		Name:         "Exhilarating Elocution",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if ctx.IsTargetLegal(item.Targets[0]) {
				if err := (effects.AddCounter{Target: target, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return c.InstanceID != target
				}),
				Power: 1, Toughness: 1,
				Label: "Exhilarating Elocution — other creatures you control get +1/+1",
			}.Apply(ctx)
		},
	})
}
