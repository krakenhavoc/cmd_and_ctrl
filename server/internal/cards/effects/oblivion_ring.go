package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oblivion Ring — Enchantment {2}{W}:
//
//	"When this enchantment enters, exile another target nonland
//	 permanent.
//	 When this enchantment leaves the battlefield, return the exiled
//	 card to the battlefield under its owner's control."
//
// The return is CR 610.3's one-shot, not a second trigger (#1729):
// ExileUntil records it and the engine performs it the moment the
// enchantment leaves, with no stack in between. The permanent comes
// back as a new object under its OWNER's control (CR 610.3c). If the
// enchantment leaves before the trigger resolves, nothing is exiled
// (CR 610.3b).
//
// "Another" excludes the enchantment itself, by instance. The target
// is chosen when the trigger goes on the stack, so a removal in
// response leaves the target where it was.
//
// The legacy leave trigger stays for one job: a restore point holding a
// card an older binary exiled with no "until" record.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bd9b9772-f5f9-4c6b-913e-7193bea5d0a7",
		Name:         "Oblivion Ring",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Targets:   Another(TargetPermanent("another target nonland permanent", Nonland())),
				Key:       oblivionRingExileLabel,
				Effect:    exileChosenTargetUntilThisLeaves("Oblivion Ring — the exiled card returns when Oblivion Ring leaves the battlefield"),
			},
			UntilThisLeavesLegacyReturn("Oblivion Ring — return the exiled card", oblivionRingExileLabel),
		},
	})
}

const oblivionRingExileLabel = "Oblivion Ring — exile another target nonland permanent"
