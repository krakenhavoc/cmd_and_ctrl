package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Animist's Might — Sorcery {2}{G} (March of the Machine: The
// Aftermath):
//
//	"This spell costs {2} less to cast if it targets a legendary
//	 creature you control.
//	 Target creature you control deals damage equal to twice its power
//	 to target creature or planeswalker you don't control."
//
// Chocobo Kick's oneSidedBiteTimes (slot 0 deals twice its power, read as
// the spell resolves, to slot 1) over Bite Down's clauses, plus Dark
// Endurance's CostsLessIfItTargets, judged against the announced targets
// when the cost is determined (CR 601.2f). The discount reads a legendary
// creature the caster controls, so a legendary creature as the second
// target (which the caster cannot control) never earns it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4940783f-2534-41c1-b614-5dabb0dff405",
		Name:         "Animist's Might",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetPermanent("target creature or planeswalker you don't control",
				Or(Creature(), Planeswalker()), Not(YouControl())),
		),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a legendary creature you control.",
				And(Creature(), Legendary(), YouControl())),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error { return oneSidedBiteTimes(ctx, 2) },
	})
}
