package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leonin Abunas — {3}{W} Creature — Cat Cleric 2/5 (Archenemy):
//
//	"Artifacts you control have hexproof."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10064324-34a0-47eb-a58e-01db10234ed9",
		Name:         "Leonin Abunas",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsArtifact() && target.Controller == source.Controller
		}, "hexproof")},
	})
}
