package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Street Wraith — Creature — Wraith {3}{B}{B}, 3/4 (Mystery Booster 2):
//
//	"Swampwalk
//	 Cycling—Pay 2 life."
//
// Cycling is an activated ability from the hand (CR 702.29a); this one's
// cost is 2 life instead of mana, with the discard. A player may pay life
// only down to zero or more (CR 119.4), so it can't be activated at 1 life.
//
// No simplification.
func init() {
	cycle := effects.Cycling("{0}")
	cycle.Label = "Cycling—Pay 2 life. (Pay 2 life, Discard this card: Draw a card.)"
	cycle.Cost = effects.Plus(effects.PayLife(2), effects.DiscardThis())
	effects.Register(effects.Spec{
		OracleID:        "bd7867f4-4f58-49c2-9c24-a124dfdb506d",
		Name:            "Street Wraith",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
		Activated:       []effects.ActivatedAbility{cycle},
	})
}
