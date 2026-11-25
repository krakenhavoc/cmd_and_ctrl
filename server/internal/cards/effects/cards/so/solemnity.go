package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Solemnity — Enchantment {2}{W} (Hour of Devastation, S82 OLD batch 241, #5534):
//
//	"Players can't get counters.
//	 Counters can't be put on artifacts, creatures, enchantments, or lands."
//
// Two world statics. The first is #6241's player prohibition, every kind a
// player can get; the second gives each artifact, creature, enchantment and
// land the permanent prohibition (CR 101.2): the placement does not happen,
// so no replacement or "whenever counters are put" trigger sees it. Counters
// already there stay. Planeswalkers and battles still take counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0b4331b-ddd3-4f11-ba16-67c651c8954c",
		Name:         "Solemnity",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PlayersCantGetCounters(), effects.OldB241NoCountersOnPermanentsOfFourTypes()},
	})
}
