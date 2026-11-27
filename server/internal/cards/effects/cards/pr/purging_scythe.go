package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Purging Scythe — Artifact {5} (Urza's Saga, S82 OLD b723):
//
//	At the beginning of your upkeep, this artifact deals 2 damage to the creature with the least toughness. If two or more creatures are tied for least toughness, you choose one of them.
//
// Porphyry Nodes' least-power shape with toughness: read as the ability resolves from every creature on the battlefield, one tied creature takes the damage and the controller picks among a tie.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4410e3c4-b8c5-46d7-9418-7490bb69d8c4",
		Name:         "Purging Scythe",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Purging Scythe — 2 damage to the creature with the least toughness", effects.S82OldB723ScytheDamage),
		},
	})
}
