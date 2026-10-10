package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "4da3e510-e495-4077-997f-8efc04f01892",
		Name:         "Endling",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:  "{B}: This creature gains menace until end of turn.",
				Cost:   ManaCost("{B}"),
				Effect: thisGainsKeywordUntilEndOfTurn("menace", "Endling — menace until end of turn"),
			},
			{
				Label:  "{B}: This creature gains deathtouch until end of turn.",
				Cost:   ManaCost("{B}"),
				Effect: thisGainsKeywordUntilEndOfTurn("deathtouch", "Endling — deathtouch until end of turn"),
			},
			{
				Label:   "{B}: This creature gains undying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    ManaCost("{B}"),
				Effect:  thisGainsKeywordUntilEndOfTurn(game.KeywordUndying, "Endling — undying until end of turn"),
			},
			{
				Label:   "{1}: This creature gets +1/-1 or -1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{1}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Endling — +1/-1 or -1/+1 until end of turn?",
						YesLabel: "+1/-1",
						NoLabel:  "-1/+1",
						OnYes: func(ctx *Context) error {
							return thisGetsUntilEndOfTurn(1, -1, "Endling — +1/-1 until end of turn")(ctx.Game, ctx.Item)
						},
						OnNo: func(ctx *Context) error {
							return thisGetsUntilEndOfTurn(-1, 1, "Endling — -1/+1 until end of turn")(ctx.Game, ctx.Item)
						},
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
