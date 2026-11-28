package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mana Severance — Sorcery {1}{U} (Tempest, S82 OLD batch 766, #5628):
//
//	"Search your library for any number of land cards, exile them, then
//	 shuffle."
//
// An unbounded search (Endless Horizons': a real choice from none to every land
// card, CR 701.23b), the found cards exiled face up, and the library shuffled
// whatever was found.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34c66f52-dcbf-4d0a-bf7f-cb294bd93e9c",
		Name:         "Mana Severance",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: effects.IsLandCard,
				Dest:      game.ZoneExile,
				Unbounded: true,
				Shuffle:   true,
				Source:    item.SourceCardID,
				Reason:    "Mana Severance — exile any number of land cards",
			}.Apply(ctx)
		},
	})
}
