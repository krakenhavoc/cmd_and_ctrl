package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grave Birthing — Instant {2}{B} (Battle for Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Target opponent exiles a card from their graveyard. You create a 1/1
//	 colorless Eldrazi Scion creature token. It has "Sacrifice this token: Add
//	 {C}."
//	 Draw a card."
//
// Scrabbling Claws' "target player exiles a card from their graveyard" (they
// choose it, and may have none), then the Scion and the draw once that choice
// is made (CR 608.2c). The Scion and the draw happen whether or not a card was
// exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b25962ab-21b5-47f9-bdef-35b911b45671",
		Name:            "Grave Birthing",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1, Draws: 1},
		Targets:         effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.OldB178TargetExilesOwnGraveyardCard(ctx.Game, ctx.Item)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.CreateToken{Template: effects.B28EldraziScionToken(), N: 1}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{N: 1}.Apply(ctx)
			}),
		),
	})
}
