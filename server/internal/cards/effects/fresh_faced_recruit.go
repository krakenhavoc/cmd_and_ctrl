package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fresh-Faced Recruit — Creature — Human Soldier {1}{R/W}, 2/1 (Ravnica:
// Clue Edition, #5929):
//
//	"During your turn, this creature has first strike."
//
// Feisty Spikeling's first strike: a layer 6 grant on itself that holds
// only while its controller is the active player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "14319020-faf2-49c4-a808-c39650203383",
		Name:         "Fresh-Faced Recruit",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return thisPermanent(target, g, source) && isActivePlayer(g, source.Controller)
			}, "first strike"),
		},
	})
}
