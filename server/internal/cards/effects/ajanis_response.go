package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ajani's Response — Instant {4}{W} (Secrets of Strixhaven, #4456):
//
//	"This spell costs {3} less to cast if it targets a tapped creature.
//	 Destroy target creature."
//
// The reduction reads the announced target (CR 601.2f, after 601.2c):
// CostsLessIfItTargets with the tapped-creature predicate, as Magnificent
// End does. Whether the creature is tapped is judged when the cost is
// determined; the destroy applies to whatever the target is on
// resolution, and does nothing if it is no longer a legal target
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bf7ec320-64bb-46ca-aac0-ca3d4224b69e",
		Name:         "Ajani's Response",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a tapped creature.", hobB3TappedCreature()),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return DestroyTarget{Target: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
