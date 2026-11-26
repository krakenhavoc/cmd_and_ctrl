package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chained to the Rocks — Enchantment — Aura {W} (Theros, S82 OLD b390, #5561):
//
//	"Enchant Mountain you control
//	 When this Aura enters, exile target creature an opponent controls
//	 until this Aura leaves the battlefield."
//
// Utopia Sprawl's narrowed enchant clause (the CR 704.5m re-check sends the
// Aura away when the Mountain stops being one or changes hands) over Banisher
// Priest's exile: the card returns under its owner's control when the Aura
// leaves (CR 610.3c); if the Aura has left before the trigger resolves nothing
// is exiled (CR 610.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "afb0c14c-41ac-4536-bfce-f1c36066b278",
		Name:         "Chained to the Rocks",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(effects.Subtype("Mountain"), effects.YouControl()),
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Chained to the Rocks — exile target creature an opponent controls until this Aura leaves the battlefield",
				effects.ExileChosenTargetUntilThisLeaves("Chained to the Rocks — the exiled creature returns when Chained to the Rocks leaves the battlefield")),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
