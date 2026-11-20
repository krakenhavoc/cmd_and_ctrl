package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Garrison Griffin — Creature — Griffin {2}{W}, 2/2 (Throne of Eldraine, #5516):
//
//	"Flying
//	 Whenever this creature attacks, target Knight you control gains flying until end of turn."
//
// An attack trigger with a target clause.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2ae828d1-0c8b-4558-bd0b-512760020913",
		Name:            "Garrison Griffin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Garrison Griffin — target Knight you control gains flying until end of turn",
					jmpB7FirstLegalTargetGains("Garrison Griffin — flying", "flying")),
				TargetCreature("target Knight you control", HasSubtype("Knight"), YouControl())),
		},
	})
}
