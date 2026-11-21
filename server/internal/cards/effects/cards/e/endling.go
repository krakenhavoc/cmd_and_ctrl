package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Endling — Creature — Zombie Shapeshifter {2}{B}{B}, 3/3:
//
//	"{B}: This creature gains menace until end of turn.
//	 {B}: This creature gains deathtouch until end of turn.
//	 {B}: This creature gains undying until end of turn.
//	 {1}: This creature gets +1/-1 or -1/+1 until end of turn."
//
// Each undying activation is another instance (CR 113.2c): they all
// trigger and the first to resolve returns it (#2075). The +1/-1 or
// -1/+1 choice is made as the last ability resolves (the ruling), a
// two-way question at resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4da3e510-e495-4077-997f-8efc04f01892",
		Name:         "Endling",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{B}: This creature gains menace until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{B}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("menace", "Endling — menace until end of turn"),
			},
			{
				Label:   "{B}: This creature gains deathtouch until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{B}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("deathtouch", "Endling — deathtouch until end of turn"),
			},
			{
				Label:   "{B}: This creature gains undying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{B}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn(game.KeywordUndying, "Endling — undying until end of turn"),
			},
			{
				Label:   "{1}: This creature gets +1/-1 or -1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{1}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Endling — +1/-1 or -1/+1 until end of turn?",
						YesLabel: "+1/-1",
						NoLabel:  "-1/+1",
						Then:     endlingPump,
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

// endlingPump is Endling's choice: +1/-1 on a yes, -1/+1 on a no. Keyed
// (#4291).
var endlingPump = effects.MayYesNo("confirm/endling-pump",
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		return effects.ThisGetsUntilEndOfTurn(1, -1, "Endling — +1/-1 until end of turn")(ctx.Game, ctx.Item)
	},
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		return effects.ThisGetsUntilEndOfTurn(-1, 1, "Endling — -1/+1 until end of turn")(ctx.Game, ctx.Item)
	})
