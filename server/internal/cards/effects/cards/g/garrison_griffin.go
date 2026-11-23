package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Garrison Griffin — Creature — Griffin {2}{W}, 2/2 (Throne of Eldraine, #5516):
//
//	"Flying
//	 Whenever this creature attacks, target Knight you control gains flying until end of turn."
//
// An attack trigger with a target clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2ae828d1-0c8b-4558-bd0b-512760020913",
		Name:            "Garrison Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Garrison Griffin — target Knight you control gains flying until end of turn",
					effects.JmpB7FirstLegalTargetGains("Garrison Griffin — flying", "flying")),
				effects.TargetCreature("target Knight you control", effects.HasSubtype("Knight"), effects.YouControl())),
		},
	})
}
