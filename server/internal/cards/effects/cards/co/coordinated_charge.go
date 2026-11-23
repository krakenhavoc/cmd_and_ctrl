package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coordinated Charge — Instant {4}{W} (Ikoria: Lair of Behemoths, #5512):
//
//	"Creatures you control get +2/+1 until end of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Heroic Charge's team pump over the creatures you control as it
// resolves, and cycling from hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "688a0b7d-bf00-45ab-ab3f-9d19f52f019a",
		Name:         "Coordinated Charge",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(effects.BoostUntilEOT{
			Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Toughness: 1,
			Label: "Coordinated Charge — creatures you control get +2/+1",
		}),
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
