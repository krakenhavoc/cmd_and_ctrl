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
		Triggered:    exileUntilThisLeavesTriggers("Banishing Light"),
	})
}

// exileUntilThisLeavesTriggers are the two rows of Banishing Light's text
// (the exile and the return), labelled with the card's name. Web Up
// (Marvel Super Heroes, #4455) prints the same text.
func exileUntilThisLeavesTriggers(name string) []game.TriggeredAbility {
	exileLabel := name + " — exile target nonland permanent an opponent controls"
	return []game.TriggeredAbility{
		{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Targets:   TargetPermanent("target nonland permanent an opponent controls", And(Nonland(), OpponentControls())),
			Key:       exileLabel,
			Effect:    exileChosenTargetUntilThisLeaves(name + " — the exiled card returns when " + name + " leaves the battlefield"),
		},
		UntilThisLeavesLegacyReturn(name+" — return the exiled card", exileLabel),
	}
}
