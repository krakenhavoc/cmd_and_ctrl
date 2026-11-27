package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forked-Branch Garami — Creature — Spirit {3}{G}{G}, 4/4 (Betrayers of Kamigawa, S82 OLD
// b596):
//
//	"Soulshift 4, soulshift 4"
//
// Two instances of soulshift trigger separately (CR 702.46b), each with its own optional
// prompt and its own target, so a dying Garami may return up to two Spirit cards of mana
// value 4 or less (never the same card twice: the second target is chosen after the first
// is on the stack, and a card that is no longer in the graveyard is not returned).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c918baa5-0e7c-4df2-8567-64a7fbef4df8",
		Name:         "Forked-Branch Garami",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Soulshift(4, "Forked-Branch Garami"),
			effects.Soulshift(4, "Forked-Branch Garami"),
		},
	})
}
