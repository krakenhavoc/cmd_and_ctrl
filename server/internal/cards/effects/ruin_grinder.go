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
							Draws:    7,
							OnYes: func(c *Context) error {
								return ruinGrinderWheel.RunAt(c, seatIndexOf(c.Game, player))
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

// ruinGrinderWheel is one player's "discard your hand and draw seven cards".
// Every player's question is open at once, so the steps find their player
// from the seat RunAt was given (ctx.Occurrence()). The draw waits for a
// prompt a discard asks (a replacement offer such as Library of Leng's; CR 608.2c, #4155).
var ruinGrinderWheel = RegisterThenSteps("then-steps/ruin-grinder-wheel",
	Instruction(func(ctx *Context) error {
		_, err := discardWholeHand(ctx.Game, seatAtOccurrence(ctx))
		return err
	}),
	Instruction(func(ctx *Context) error { return ctx.Game.DrawNForEffect(seatAtOccurrence(ctx), 7) }),
)
