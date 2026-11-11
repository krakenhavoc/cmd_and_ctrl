package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Migration Path — Sorcery {3}{G} (Tarkir: Dragonstorm Commander):
//
//	"Search your library for up to two basic land cards, put them onto
//	 the battlefield tapped, then shuffle.
//	 Cycling {2}"
//
// Explosive Vegetation's body (cmdrB21UpToTwoBasicsTapped) plus Cycling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "db4907f3-4071-47c4-9329-96b4813435be",
		Name:         "Migration Path",
		Purpose:      game.Purpose{Lands: 2},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return cmdrB21UpToTwoBasicsTapped(ctx, "Migration Path")
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
