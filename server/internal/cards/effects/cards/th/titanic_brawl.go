package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Titanic Brawl — Instant {1}{G} (Ravnica Remastered):
//
//	"This spell costs {1} less to cast if it targets a creature you control
//	 with a +1/+1 counter on it.
//	 Target creature you control fights target creature you don't control."
//
// The discount reads the announced targets (CostsLessIfItTargets); the fight
// is Go for Blood's, which deals nothing if either has left (CR 701.14b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e8434c77-10df-4456-af27-acb95d624189",
		Name:         "Titanic Brawl",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(1, "This spell costs {1} less to cast if it targets a creature you control with a +1/+1 counter on it.",
				effects.And(effects.Creature(), effects.YouControl(), effects.GapB78HasPlusOneCounter())),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SlotZeroFightsSlotOne(ctx.Game, item)
		},
	})
}
