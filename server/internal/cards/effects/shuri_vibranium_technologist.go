package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shuri, Vibranium Technologist — {2}{W/U} Legendary Creature — Human
// Artificer Hero 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Vigilance
//	 When Shuri enters, choose one —
//	 • Create a 1/1 colorless Robot Hero artifact creature token with flying.
//	 • Draw a card."
//
// The token follows the oracle text: no printed token card is 1/1 (the
// only printed Robot Hero is a 2/1), so the row is listed in
// knownUnmatchedTokens with this card's name.
//
// No simplification.
func init() {
	shuri := WhenThisEnters("Shuri, Vibranium Technologist — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	shuri.Modes = ChooseOne(
		ModeDoing("Create a 1/1 colorless Robot Hero artifact creature token with flying.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Template: TokenCard("1/1 colorless Robot Hero artifact with flying"), N: 1}.Apply(ctx)
			}),
		ModeDoing("Draw a card.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return DrawCards{N: 1}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:        "94b2a7a6-9e9a-437f-9cde-111b1b02a68e",
		Name:            "Shuri, Vibranium Technologist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{shuri},
	})
}
