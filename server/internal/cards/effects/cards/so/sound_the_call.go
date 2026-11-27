package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sound the Call — Sorcery {2}{G} (Coldsnap, S82 OLD b560):
//
//	"Create a 1/1 green Wolf creature token. It has "This token gets +1/+1 for
//	 each card named Sound the Call in each graveyard.""
//
// The token carries its own static ability (a layer 7c modify that counts the
// cards named Sound the Call in every graveyard as it is recomputed), so it
// grows when the spell itself reaches the graveyard and for as long as any
// copy lies in any graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f0ab2fb-297a-4824-8d2c-0120eec09e5e",
		Name:         "Sound the Call",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenFromCatalog(effects.PrintedOldB560SoundTheCallWolf), N: 1}.Apply(ctx)
		},
	})
}
