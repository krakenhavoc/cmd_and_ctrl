package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4940783f-2534-41c1-b614-5dabb0dff405",
		Name:         "Animist's Might",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetPermanent("target creature or planeswalker you don't control",
				effects.Or(effects.Creature(), effects.Planeswalker()), effects.Not(effects.YouControl())),
		),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a legendary creature you control.",
				effects.And(effects.Creature(), effects.Legendary(), effects.YouControl())),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error { return effects.OneSidedBiteTimes(ctx, 2) },
	})
}
