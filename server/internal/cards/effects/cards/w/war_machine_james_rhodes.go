package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "0a512e55-793e-4936-bd27-80d749f4657e",
		Name:            "War Machine, James Rhodes",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("War Machine — tap up to one target creature", effects.TapChosenPermanent),
				effects.UpToOneTargetCreature("up to one target creature")),
		},
	})
}
