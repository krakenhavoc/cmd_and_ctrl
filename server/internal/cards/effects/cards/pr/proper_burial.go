package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Proper Burial — Enchantment {3}{W} (Dissension, #5598):
//
//	"Whenever a creature you control dies, you gain life equal to that
//	 creature's toughness."
//
// The toughness is read as the ability resolves, from the creature as it
// last existed on the battlefield (CR 608.2h, GainLifeEqualToSubject).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5bf77c2-a9b2-4366-8412-e263160403c1",
		Name:         "Proper Burial",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverACreatureYouControlDies("Proper Burial — gain life equal to its toughness", effects.Do(effects.GainLifeEqualToSubject{Stat: effects.SubjectToughness})),
		},
	})
}
