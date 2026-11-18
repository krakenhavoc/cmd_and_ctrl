package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Swarmlord — Creature — Phyrexian Insect Horror {4}{G}{G}, 4/4:
//
//	"Infect
//	 At the beginning of your upkeep, create a 1/1 green Phyrexian Insect
//	 creature token with infect for each poison counter your opponents have."
//
// The number is the total of poison counters over every opponent still in the
// game, read as the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4fc2755c-ae7b-4147-a573-b33679a461e8",
		Name:            "Phyrexian Swarmlord",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Phyrexian Swarmlord — create a 1/1 green Phyrexian Insect creature token with infect for each poison counter your opponents have",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := 0
					for _, opp := range ctx.Opponents() {
						if p := g.PlayerByIDForEffect(opp); p != nil {
							n += p.Counters[game.CounterPoison]
						}
					}
					if n <= 0 {
						return nil
					}
					return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 green Phyrexian Insect with infect"), N: n}.Apply(ctx)
				}),
		},
	})
}
