package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pegasus Courser — Creature — Pegasus {2}{W}, 1/3 (Dominaria, #5530):
//
//	"Flying
//	 Whenever this creature attacks, another target attacking creature gains flying until end of turn."
//
// Aerial Guide's attack trigger: another attacking creature gains
// flying until end of turn.
//
// No simplification.
func init() {
	const label = "Pegasus Courser — another attacking creature gains flying"
	attack := effects.WheneverThisAttacks(label, effects.FirstTargetGainsUntilEndOfTurn(0, label, "flying"))
	attack.Targets = effects.Another(effects.TargetCreature("another target attacking creature", effects.AttackingCreature()))
	effects.Register(effects.Spec{
		OracleID:        "24db5905-d513-42c6-9e9d-114ac5cbd21d",
		Name:            "Pegasus Courser",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{attack},
	})
}
