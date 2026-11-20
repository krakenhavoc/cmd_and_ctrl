package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Salvage — Sorcery {G} (Portal Second Age, #5625):
//
//	"Put target card from your graveyard on top of your library."
//
// Reclaim at sorcery speed: only the caster's own graveyard, and the
// target is checked again at resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c01a090a-11e4-469e-aded-74203c06fef4",
		Name:         "Salvage",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target card from your graveyard", YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return PutChosenTargetOnTopOfLibrary(ctx.Game, item)
		},
	})
}
