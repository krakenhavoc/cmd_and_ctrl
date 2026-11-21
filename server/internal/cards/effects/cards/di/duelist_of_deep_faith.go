package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Duelist of Deep Faith — Creature — Phyrexian Soldier {1}{W}, 2/2
// (Phyrexia: All Will Be One, #5479):
//
//	"Toxic 1
//	 During your turn, this creature has first strike."
//
// Feisty Spikeling's first strike: a layer 6 grant on itself that holds
// only while its controller is the active player, so it is gone when it
// blocks on an opponent's turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "30d69e88-2bf6-43b8-8678-2cf37f6a5cf5",
		Name:            "Duelist of Deep Faith",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"toxic 1"},
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.IsActivePlayer(g, source.Controller)
			}, "first strike"),
		},
	})
}
