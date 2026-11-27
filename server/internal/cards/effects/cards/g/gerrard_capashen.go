package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gerrard Capashen — Legendary Creature — Human Soldier {3}{W}{W}, 3/4
// (Apocalypse, #5614):
//
//	"At the beginning of your upkeep, you gain 1 life for each card in target
//	 opponent's hand.
//	 {3}{W}: Tap target creature. Activate only if Gerrard Capashen is
//	 attacking."
//
// The opponent is chosen as the upkeep trigger goes on the stack and their
// hand is counted as it resolves (CR 608.2h). The tap ability is gated on the
// source being an attacking creature, as Forgestoker Dragon's is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27521e70-3a1f-4819-a111-c38216ffe7c3",
		Name:         "Gerrard Capashen",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtYourUpkeep("Gerrard Capashen — you gain 1 life for each card in target opponent's hand", effects.S82OldB634GerrardUpkeep),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "{3}{W}: Tap target creature. Activate only if Gerrard Capashen is attacking.",
			Cost:      effects.ManaCost("{3}{W}"),
			Targets:   effects.TargetCreature("target creature"),
			Condition: effects.SourceIsAttacking(),
			Effect:    effects.ThousandMoonsCrackshotTap,
		}},
	})
}
