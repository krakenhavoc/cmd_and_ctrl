package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Call the Scions — Sorcery {2}{G} (Battle for Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Create two 1/1 colorless Eldrazi Scion creature tokens. They have
//	 "Sacrifice this token: Add {C}.""
//
// Two of the catalog's Eldrazi Scions, as one creation.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "731a70e8-40d6-47b9-b2b3-b16c818e1935",
		Name:            "Call the Scions",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.B28EldraziScionToken(), N: 2}.Apply(ctx)
		},
	})
}
