package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prosperous Bandit — Creature — Raccoon Rogue {2}{R}, 2/2 (Bloomburrow
// Commander, #5368):
//
//	"Offspring {1} ... First strike
//	 Whenever this creature deals combat damage to a player, create that
//	 many tapped Treasure tokens."
//
// One tapped Treasure per point of the triggering damage.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2a14670b-2ab6-4e1d-9b33-610a87bdb7ce",
		Name:            "Prosperous Bandit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		OptionalCosts:   []game.AdditionalCost{effects.Offspring("{1}")},
		Triggered: []game.TriggeredAbility{
			effects.OffspringToken("Prosperous Bandit"),
			effects.WheneverThisDealsCombatDamageToAPlayer("Prosperous Bandit — create that many tapped Treasure tokens",
				effects.Do(effects.Cmdr2S18TappedTreasuresForDamage{Count: func(ctx *effects.Context) int { return ctx.Trigger().Event.Amount }})),
		},
	})
}
