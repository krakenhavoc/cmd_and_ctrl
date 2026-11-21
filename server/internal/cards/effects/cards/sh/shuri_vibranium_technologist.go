package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	shuri := effects.WhenThisEnters("Shuri, Vibranium Technologist — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	shuri.Modes = effects.ChooseOne(
		effects.ModeDoing("Create a 1/1 colorless Robot Hero artifact creature token with flying.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Template: effects.TokenCard("1/1 colorless Robot Hero artifact with flying"), N: 1}.Apply(ctx)
			}),
		effects.ModeDoing("Draw a card.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.DrawCards{N: 1}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:        "94b2a7a6-9e9a-437f-9cde-111b1b02a68e",
		Name:            "Shuri, Vibranium Technologist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{shuri},
	})
}
