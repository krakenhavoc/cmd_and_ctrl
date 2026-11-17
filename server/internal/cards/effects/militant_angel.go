package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Militant Angel — Creature — Angel {3}{W}{W}, 3/4 (Game Night, #5523):
//
//	"Flying, lifelink
//	 When this creature enters, create a number of 2/2 white Knight
//	 creature tokens with vigilance equal to the number of opponents you
//	 attacked this turn."
//
// The count is OpponentsYouAttackedThisTurn (Kurn, Son of Mogh's), taken
// as the trigger resolves (CR 608.2h) from this turn's attack records:
// each opponent once however many creatures attacked them, an attack on a
// planeswalker or a battle counts no player, and a creature that attacked
// and has since died still counts, because the records are history. An
// opponent who has left the game since is no longer an opponent and is
// not counted. Only the active player declares attackers, so the Angel
// entering on another player's turn makes no Knights.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ad645324-204a-49d2-b128-f98eee7c9bb9",
		Name:            "Militant Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Militant Angel — create a 2/2 white Knight with vigilance for each opponent you attacked this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := OpponentsYouAttackedThisTurn(g, ctx.Controller())
					if n <= 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("2/2 white Knight with vigilance"), N: n}.Apply(ctx)
				}),
		},
	})
}
