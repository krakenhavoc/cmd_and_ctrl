package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oviya Pashiri, Sage Lifecrafter — Legendary Creature — Human Artificer
// {G}, 1/2 (Commander Masters):
//
//	"{2}{G}, {T}: Create a 1/1 colorless Servo artifact creature token.
//	 {4}{G}, {T}: Create an X/X colorless Construct artifact creature token,
//	 where X is the number of creatures you control."
//
// X is counted as the second ability resolves (CR 608.2h reads it from the
// board as it is), the Construct itself not yet among them, and is the size
// the token enters with (the printed Construct token is */*, so the template
// is the 0/0 row every X-sized Construct shares).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "41e03224-bcdd-4127-a19a-fe66196339b9",
		Name:         "Oviya Pashiri, Sage Lifecrafter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}{G}, {T}: Create a 1/1 colorless Servo artifact creature token.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    Plus(ManaCost("{2}{G}"), TapCost()),
				Effect:  Do(CreateToken{Template: TokenCard("1/1 colorless Servo artifact"), N: 1}),
			},
			{
				Label:   "{4}{G}, {T}: Create an X/X colorless Construct artifact creature token, where X is the number of creatures you control.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    Plus(ManaCost("{4}{G}"), TapCost()),
				Effect:  oviyaConstruct,
			},
		},
	})
}

func oviyaConstruct(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() {
			x++
		}
	}
	tok := ctx.TextToken(TokenCard("0/0 colorless Construct artifact"))
	tok.Power, tok.Toughness = x, x
	return ctx.Game.CreateTokenForEffect(item.Controller, tok, 1)
}
