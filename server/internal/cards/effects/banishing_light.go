package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Banishing Light — Enchantment {2}{W} (The Hobbit Eternal, HOC):
//
//	"When this enchantment enters, exile target nonland permanent an
//	 opponent controls until this enchantment leaves the battlefield."
//
// Oblivion Ring's shape (CR 610.3), word for word Memory Trap (#4454):
// the exile records its own return, so the permanent comes back as a new
// object under its owner's control the moment this leaves. If the
// enchantment leaves before the trigger resolves, nothing is exiled
// (CR 610.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f28b21a6-f7ce-437a-8c5b-0423cb55cefb",
		Name:         "Banishing Light",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Targets:   TargetPermanent("target nonland permanent an opponent controls", And(Nonland(), OpponentControls())),
				Key:       banishingLightExileLabel,
				Effect:    exileChosenTargetUntilThisLeaves("Banishing Light — the exiled card returns when Banishing Light leaves the battlefield"),
			},
			UntilThisLeavesLegacyReturn("Banishing Light — return the exiled card", banishingLightExileLabel),
		},
	})
}

const banishingLightExileLabel = "Banishing Light — exile target nonland permanent an opponent controls"
