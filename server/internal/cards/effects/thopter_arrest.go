package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thopter Arrest — Enchantment {2}{W} (Aether Revolt,
// #5536):
//
//	"When this enchantment enters, exile target artifact or creature an
//	 opponent controls until this enchantment leaves the battlefield."
//
// Mystical Tether without the flash cost: the exile is linked to the
// enchantment and the card returns when it leaves (exile_until.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9b9cc771-3988-41af-a49e-f3bb7a8e397c",
		Name:         "Thopter Arrest",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Thopter Arrest — exile target artifact or creature an opponent controls until this enchantment leaves the battlefield",
				exileChosenTargetUntilThisLeaves("Thopter Arrest — the exiled card returns when Thopter Arrest leaves the battlefield")),
				TargetPermanent("target artifact or creature an opponent controls", Or(Artifact(), Creature()), OpponentControls())),
		},
	})
}
