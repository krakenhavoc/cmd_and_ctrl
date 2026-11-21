package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "9b9cc771-3988-41af-a49e-f3bb7a8e397c",
		Name:         "Thopter Arrest",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Thopter Arrest — exile target artifact or creature an opponent controls until this enchantment leaves the battlefield",
				effects.ExileChosenTargetUntilThisLeaves("Thopter Arrest — the exiled card returns when Thopter Arrest leaves the battlefield")),
				effects.TargetPermanent("target artifact or creature an opponent controls", effects.Or(effects.Artifact(), effects.Creature()), effects.OpponentControls())),
		},
	})
}
