package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hallowed Spiritkeeper — Creature — Avatar {1}{W}{W}, 3/2:
//
//	"Vigilance
//	 When this creature dies, create X 1/1 white Spirit creature tokens with
//	 flying, where X is the number of creature cards in your graveyard."
//
// X is counted as the trigger resolves, so the Spiritkeeper itself, now in
// the graveyard, is one of the cards counted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5852ed2f-4332-4c30-b145-373deea471d9",
		Name:            "Hallowed Spiritkeeper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Hallowed Spiritkeeper — create X 1/1 white Spirit creature tokens with flying, where X is the number of creature cards in your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: b11CreatureCardsInGraveyard(g, ctx.Controller())}.Apply(ctx)
				}),
		},
	})
}
