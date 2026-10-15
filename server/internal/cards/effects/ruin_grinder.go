package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ruin Grinder — Artifact Creature — Construct {5}{R}, 7/4:
//
//	"Menace
//	 When this creature dies, each player may discard their hand and
//	 draw seven cards.
//	 Mountaincycling {2}"
//
// The dies trigger is Variable Chaser's wheel: each player, starting
// with the controller, chooses for themselves. Mountaincycling is the
// shared Typecycling constructor (any land with the Mountain type).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8d0445ef-370b-4f4f-8975-a40cb02c031e",
		Name:            "Ruin Grinder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Ruin Grinder — each player may discard their hand and draw seven cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, who := range tablePlayers(ctx) {
						player := who
						if err := (MayChoice{
							Player:   player,
							Question: "Ruin Grinder — discard your hand and draw seven cards?",
							OnYes: func(c *Context) error {
								if _, err := discardWholeHand(c.Game, player); err != nil {
									return err
								}
								return c.Game.DrawNForEffect(player, 7)
							},
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
		Activated: []ActivatedAbility{
			Typecycling("Mountaincycling", "{2}", "a Mountain card", IsLandWithSubtype("Mountain")),
		},
	})
}
