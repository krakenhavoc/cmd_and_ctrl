package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Conclave Guildmage — Creature — Elf Cleric {G}{W}, 2/2 (Guilds of Ravnica,
// #5525):
//
//	"{G}, {T}: Creatures you control gain trample until end of turn.
//	 {5}{W}, {T}: Create a 2/2 green and white Elf Knight creature token with
//	 vigilance."
//
// Drover Grizzly's team trample, over the creatures its controller controls as
// the ability resolves (the set is fixed then, CR 611.2c), and Sprouting
// Renewal's Elf Knight token. Both need the Guildmage untapped and, being {T}
// costs, no summoning sickness.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d0eb1f4-08ef-4ac9-902b-47335ab46984",
		Name:         "Conclave Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{G}, {T}: Creatures you control gain trample until end of turn.",
				Purpose: game.Purpose{TeamPump: &game.Pump{Keywords: []string{"trample"}}, Answers: game.AnswerPump},
				Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.Creature(), effects.YouControl()),
						Keywords: []string{"trample"},
						Label:    "Conclave Guildmage — trample until end of turn",
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{5}{W}, {T}: Create a 2/2 green and white Elf Knight creature token with vigilance.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    effects.Plus(effects.ManaCost("{5}{W}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("2/2 green and white Elf Knight with vigilance"), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
