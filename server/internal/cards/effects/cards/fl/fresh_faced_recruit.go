package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "14319020-faf2-49c4-a808-c39650203383",
		Name:         "Fresh-Faced Recruit",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.IsActivePlayer(g, source.Controller)
			}, "first strike"),
		},
	})
}
