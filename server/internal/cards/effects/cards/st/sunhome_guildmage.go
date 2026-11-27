package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunhome Guildmage — Creature — Human Wizard {R}{W}, 2/2 (Gatecrash, S82 OLD
// #5565):
//
//	"{1}{R}{W}: Creatures you control get +1/+0 until end of turn.
//	 {2}{R}{W}: Create a 1/1 red and white Soldier creature token with haste."
//
// Ruby Collector's team pump over the creatures its controller controls as the
// ability resolves, and a plain token ability on the existing "1/1 red and
// white Soldier with haste" template. Neither needs {T}, so each can be
// activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f211d21f-7d78-4435-a592-a6896023af3b",
		Name:         "Sunhome Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{R}{W}: Creatures you control get +1/+0 until end of turn.",
				Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1}, Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{1}{R}{W}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1,
						Label: "Sunhome Guildmage — creatures you control get +1/+0",
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{2}{R}{W}: Create a 1/1 red and white Soldier creature token with haste.",
				Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{2}{R}{W}"),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red and white Soldier with haste"), N: 1}),
			},
		},
	})
}
