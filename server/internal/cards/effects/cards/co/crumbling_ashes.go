package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crumbling Ashes — {1}{B} Enchantment:
//
//	"At the beginning of your upkeep, destroy target creature with a -1/-1
//	 counter on it."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e4030eb-c74f-4f81-8d35-8f1d080b15cf",
		Name:         "Crumbling Ashes",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtYourUpkeep("Crumbling Ashes — destroy target creature with a -1/-1 counter on it", effects.DestroyFirstLegalTarget),
				effects.TargetCreature("target creature with a -1/-1 counter on it", effects.S82OldB522WithCounter(game.CounterMinusOne))),
		},
	})
}
