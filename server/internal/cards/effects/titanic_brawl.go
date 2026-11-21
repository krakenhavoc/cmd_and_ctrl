package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e8434c77-10df-4456-af27-acb95d624189",
		Name:         "Titanic Brawl",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature you don't control", OpponentControls()),
		),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(1, "This spell costs {1} less to cast if it targets a creature you control with a +1/+1 counter on it.",
				And(Creature(), YouControl(), gapB78HasPlusOneCounter())),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return slotZeroFightsSlotOne(ctx.Game, item)
		},
	})
}
