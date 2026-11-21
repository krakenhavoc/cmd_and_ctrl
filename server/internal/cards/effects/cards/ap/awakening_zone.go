package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Awakening Zone — Enchantment for {2}{G}:
//
//	"At the beginning of your upkeep, create a 0/1 colorless Eldrazi
//	Spawn creature token. It has 'Sacrifice this creature: Add {C}.'"
//
// S19 sub-PR 5: a pure token-generating "your upkeep" trigger.
// Mandatory; the token arrives when the trigger resolves.
//
// No simplification remains. The Spawn's "Sacrifice this creature:
// Add {C}" went live in S21 sub-PR 1 — EldraziSpawnToken declares it
// on Card.ManaAbilities with no tap in the cost, so a Spawn can be
// cracked the turn it arrives. The note here calling it deferred
// outlived the fix; corrected in the #338 sweep.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f955bc96-d602-4142-a9a2-87009cc7028c",
		Name:         "Awakening Zone",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Awakening Zone — create an Eldrazi Spawn", effects.Do(effects.CreateToken{
				Template: effects.EldraziSpawnToken(),
				N:        1,
			})),
		},
	})
}
