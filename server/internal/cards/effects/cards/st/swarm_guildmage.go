package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Swarm Guildmage — Creature — Elf Shaman {B}{G}, 2/2 (Guilds of Ravnica,
// #5525):
//
//	"{4}{B}, {T}: Creatures you control get +1/+0 and gain menace until end of
//	 turn.
//	 {1}{G}, {T}: You gain 2 life."
//
// Make a Stand's pump and grant over the creatures its controller controls as
// the first ability resolves (the set is fixed then, CR 611.2c), and Campfire's
// life gain for the second. Both need the Guildmage untapped and, being {T}
// costs, no summoning sickness.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14b8b35e-57b6-4577-97a8-9d9c513db0b8",
		Name:         "Swarm Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{4}{B}, {T}: Creatures you control get +1/+0 and gain menace until end of turn.",
				Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Keywords: []string{"menace"}}, Answers: game.AnswerPump},
				Cost:    effects.Plus(effects.ManaCost("{4}{B}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					yours := effects.And(effects.Creature(), effects.YouControl())
					if err := (effects.BoostUntilEOT{Match: yours, Power: 1, Label: "Swarm Guildmage — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"menace"}, Label: "Swarm Guildmage — menace"}.Apply(ctx)
				},
			},
			{
				Label:   "{1}{G}, {T}: You gain 2 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.TapCost()),
				Effect:  effects.Do(effects.GainLife{Amount: 2}),
			},
		},
	})
}
