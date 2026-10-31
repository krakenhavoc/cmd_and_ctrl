package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// War Machine, James Rhodes — {3}{W/U} Legendary Artifact Creature —
// Human Hero 3/2 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever War Machine attacks, tap up to one target creature."
//
// Choosing no target is legal; a target that became illegal is skipped on
// resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0a512e55-793e-4936-bd27-80d749f4657e",
		Name:            "War Machine, James Rhodes",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("War Machine — tap up to one target creature", tapChosenPermanent),
				UpToOneTargetCreature("up to one target creature")),
		},
	})
}
