package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Apes of Rath — Creature — Ape {2}{G}{G}, 5/4 (Tempest, S82 OLD, #5628):
//
//	"Whenever this creature attacks, it doesn't untap during its controller's
//	 next untap step."
//
// An attack trigger (OldB813DoesntUntapWhenThisAttacks, shared with Lead Golem)
// that records the skip on the Apes as the object they are (CR 502.3, CR
// 400.7): the marker belongs to whoever controls them when that
// untap step comes. A vigilant or untapped Apes lose nothing, as the skip only
// matters to a tapped permanent.
//
// No simplification.
func init() {
	const name = "Apes of Rath"
	effects.Register(effects.Spec{
		OracleID:     "b52f5850-f1aa-412a-bae9-2ce76c287188",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB813DoesntUntapWhenThisAttacks(name),
		},
	})
}
