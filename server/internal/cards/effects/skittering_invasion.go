package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skittering Invasion — Kindred Sorcery — Eldrazi {7} (Modern Horizons 3
// Commander, #5948):
//
//	"Create five 0/1 colorless Eldrazi Spawn creature tokens. They have
//	 "Sacrifice this token: Add {C}.""
//
// The Spawn are the catalog's own (EldraziSpawnToken), whose sacrifice
// mana ability is real, not text on a vanilla token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf890abd-b282-4dbb-ac86-9e7caec1c1b8",
		Name:         "Skittering Invasion",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 5, CreatureTokens: 5},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: EldraziSpawnToken(), N: 5}.Apply(ctx)
		},
	})
}
