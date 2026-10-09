package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Variable Chaser // Arc of Fortune — Creature — Human Wizard {2}{U},
// 2/3 // Sorcery {2}{U} (preparation card, CR 722):
//
//	"Flying, prowess
//	 This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)"
//
//	Arc of Fortune — "Each player may discard their hand and draw seven
//	 cards."
//
// Each player is asked in turn order starting with the caster and
// finishes their own discard-and-draw before the next is asked.
//
// No simplification.
func init() {
	const id = "67c603cc-ad66-4a1c-8386-5901c9c01bfb"
	Register(Spec{
		OracleID:        id,
		Name:            "Variable Chaser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "prowess"},
		Replacements:    []game.ReplacementEffect{SelfEntersPrepared()},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Arc of Fortune",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, who := range tablePlayers(ctx) {
				player := who
				if err := (MayChoice{
					Player:   player,
					Question: "Arc of Fortune — discard your hand and draw seven cards?",
					OnYes: func(ctx *Context) error {
						if _, err := discardWholeHand(ctx.Game, player); err != nil {
							return err
						}
						return ctx.Game.DrawNForEffect(player, 7)
					},
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
