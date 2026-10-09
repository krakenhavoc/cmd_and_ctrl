package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Memory Trap — Enchantment {2}{W} (Reality Fracture):
//
//	"When this enchantment enters, exile target nonland permanent an
//	 opponent controls until this enchantment leaves the battlefield."
//
// Oblivion Ring's shape (CR 610.3): the exile records its own return, so
// the permanent comes back as a new object under its owner's control the
// moment this leaves. If the enchantment leaves before the trigger
// resolves, nothing is exiled (CR 610.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70616a8c-9b6c-408e-8f3e-2c348ec136d8",
		Name:         "Memory Trap",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Targets:   TargetPermanent("target nonland permanent an opponent controls", And(Nonland(), OpponentControls())),
				Key:       memoryTrapExileLabel,
				Effect:    exileChosenTargetUntilThisLeaves("Memory Trap — the exiled card returns when Memory Trap leaves the battlefield"),
			},
			UntilThisLeavesLegacyReturn("Memory Trap — return the exiled card", memoryTrapExileLabel),
		},
	})
}

const memoryTrapExileLabel = "Memory Trap — exile target nonland permanent an opponent controls"
