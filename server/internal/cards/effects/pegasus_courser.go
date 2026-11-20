package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	attack := WheneverThisAttacks(label, firstTargetGainsUntilEndOfTurn(0, label, "flying"))
	attack.Targets = Another(TargetCreature("another target attacking creature", AttackingCreature()))
	Register(Spec{
		OracleID:        "24db5905-d513-42c6-9e9d-114ac5cbd21d",
		Name:            "Pegasus Courser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{attack},
	})
}
