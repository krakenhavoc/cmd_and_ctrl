package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invoke the Winds — Sorcery {1}{U}{U}{U}{U} (Kamigawa: Neon Dynasty, #5493):
//
//	"Gain control of target artifact or creature. Untap it."
//
// No duration is stated, so the control change lasts indefinitely
// (CR 611.2a), a layer 2 effect pinned to the object it took (CR 400.7).
// The untap follows it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "721c3255-953c-49cd-a5dd-852c3d09eae4",
		Name:         "Invoke the Winds",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			id := item.Targets[0].ID
			if err := (effects.GainControl{Target: id, Duration: game.IndefiniteDuration(),
				Label: "Invoke the Winds — gain control"}).Apply(ctx); err != nil {
				return err
			}
			return effects.UntapTarget{Target: id}.Apply(ctx)
		},
	})
}
