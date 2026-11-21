package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conclave Phalanx — Creature — Human Soldier {4}{W}, 2/4 (Ravnica: City
// of Guilds, #5600):
//
//	"Convoke
//	 When this creature enters, you gain 1 life for each creature you
//	 control."
//
// Convoke is the spell's tap cost (Spec.TapCost, CR 702.51). The enters
// trigger counts as it resolves (Mastery of the Unseen's body,
// cmdr2S37GainLifePerCreature), so the Phalanx counts itself while it is
// still on the battlefield, and the creatures tapped to convoke it count
// too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a1bebe9-c940-4fac-8572-0833808eaa55",
		Name:         "Conclave Phalanx",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Conclave Phalanx — you gain 1 life for each creature you control", effects.Cmdr2S37GainLifePerCreature),
		},
	})
}
