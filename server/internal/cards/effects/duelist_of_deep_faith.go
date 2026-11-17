package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "30d69e88-2bf6-43b8-8678-2cf37f6a5cf5",
		Name:            "Duelist of Deep Faith",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"toxic 1"},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return thisPermanent(target, g, source) && isActivePlayer(g, source.Controller)
			}, "first strike"),
		},
	})
}
