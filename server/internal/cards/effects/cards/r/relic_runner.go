package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Relic Runner — Creature — Human Rogue {1}{U}, 2/1 (Dominaria, S82 OLD,
// #5530):
//
//	"This creature can't be blocked if you've cast a historic spell this turn.
//	 (Artifacts, legendaries, and Sagas are historic.)"
//
// A block rule on the Runner itself (Thieves' Tools' CantBeBlockedWhile) whose
// condition is read when blockers are declared (CR 509.1b): the Runner's
// controller has cast an artifact, legendary or Saga spell this turn. The turn's
// cast log is read, so a historic spell that was countered still counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69e4e5bf-f0e2-4ed1-8b17-c8c7a9381551",
		Name:         "Relic Runner",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{effects.CantBeBlockedWhile(effects.OnSelf(), func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			return effects.S82OldB213CastHistoricSpellThisTurn(g, controller)
		})},
	})
}
