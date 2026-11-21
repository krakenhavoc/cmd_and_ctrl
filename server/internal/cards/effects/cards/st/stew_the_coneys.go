package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stew the Coneys — Instant {2}{G} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Target creature you control deals damage equal to its power to target
//	 creature you don't control. Create a Food token. (It's an artifact with
//	 "{2}, {T}, Sacrifice this token: You gain 3 life.")"
//
// Bite Down's one-sided bite (the creature you control is the source, so
// deathtouch and lifelink see it, and nothing is dealt unless both targets are
// still legal, CR 608.2b), then the Food, after any prompt the damage asks.
// With either target illegal the bite does nothing; the spell still resolves
// as long as one target is legal, so the Food is still made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6f1b8da-5e7c-4121-9b0f-fd6678b04991",
		Name:         "Stew the Coneys",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.OneSidedBite(ctx.Item, ctx) }),
			effects.CreateToken{Template: effects.FoodToken(), N: 1},
		),
	})
}
