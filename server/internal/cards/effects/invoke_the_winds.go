package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "721c3255-953c-49cd-a5dd-852c3d09eae4",
		Name:         "Invoke the Winds",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			id := item.Targets[0].ID
			if err := (GainControl{Target: id, Duration: game.IndefiniteDuration(),
				Label: "Invoke the Winds — gain control"}).Apply(ctx); err != nil {
				return err
			}
			return UntapTarget{Target: id}.Apply(ctx)
		},
	})
}
