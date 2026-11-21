package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angel of Renewal — Creature — Angel Ally {5}{W}, 4/4 (Battle for
// Zendikar, #5547):
//
//	"Flying
//	 When this creature enters, you gain 1 life for each creature you
//	 control."
//
// Conclave Phalanx's entry trigger on a flyer: the creatures are
// counted as it resolves (CR 608.2h), the Angel itself included while
// it is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c604fda4-164c-4966-94f8-1e3a96d5ef9d",
		Name:            "Angel of Renewal",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Angel of Renewal — you gain 1 life for each creature you control", effects.Cmdr2S37GainLifePerCreature),
		},
	})
}
