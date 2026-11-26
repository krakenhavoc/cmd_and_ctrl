package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aetherling — Creature — Shapeshifter {4}{U}{U}, 4/5 (Dragon's Maze, S82 OLD
// batch 404, #5564):
//
//	"{U}: Exile this creature. Return it to the battlefield under its owner's
//	 control at the beginning of the next end step.
//	 {U}: This creature can't be blocked this turn.
//	 {1}: This creature gets +1/-1 until end of turn.
//	 {1}: This creature gets -1/+1 until end of turn."
//
// Four activated abilities, each a body that exists already. The flicker is
// Djinn of the Fountain's (it leaves and comes back as a new object, CR 400.7, at
// the next end step under its owner's control), so the other three do nothing for a
// copy of the old object: Agent of Horizons's unblockable is guarded on the source
// still being this permanent, and the two pumps are layer-7 modifications on it
// until cleanup (Rootwalla's body).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b980bf11-9e12-4ee9-83c1-45706d9ae344",
		Name:         "Aetherling",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}: Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{U}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.ExileThisThenScheduleReturn(effects.NewContext(g, item), item, "Aetherling — return the exiled creature")
				},
			},
			{
				Label:   "{U}: This creature can't be blocked this turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{U}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
						Label: "Aetherling — can't be blocked this turn"}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{1}: This creature gets +1/-1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{1}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, -1, "Aetherling — +1/-1 until end of turn"),
			},
			{
				Label:   "{1}: This creature gets -1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{1}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(-1, 1, "Aetherling — -1/+1 until end of turn"),
			},
		},
	})
}
