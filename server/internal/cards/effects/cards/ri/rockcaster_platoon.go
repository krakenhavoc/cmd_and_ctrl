package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rockcaster Platoon — {5}{W}{W} Creature — Rhino Soldier 5/7 (Shards of Alara,
// S82 OLD b515):
//
//	"{4}{G}: This creature deals 2 damage to each creature with flying and
//	 each player."
//
// Hammerfist Giant's body (one damage instance to the matching creatures, then
// to every player) over flying creatures. A partial sweep (ADR 0126).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3bc82914-1524-4d57-8b38-9788867a4046",
		Name:         "Rockcaster Platoon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}: This creature deals 2 damage to each creature with flying and each player.",
			Cost:    effects.ManaCost("{4}{G}"),
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
			Effect:  effects.ThisDealsDamageToEachCreatureMatchingAndEachPlayer(effects.HasKeyword("flying"), 2),
		}},
	})
}
