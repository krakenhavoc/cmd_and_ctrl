package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lotus Bloom — Artifact with NO mana cost:
//
//	"Suspend 3—{0}
//	 {T}, Sacrifice this artifact: Add three mana of any one color."
//
// A Black Lotus you have to ask for three turns in advance. Like
// Ancestral Vision it can never be cast from hand — an empty mana
// cost is unpayable, not free (CR 118.6) — and unlike it, the suspend
// cost really is {0}, which is a printed cost and not an omission.
//
// It is the PERMANENT half of suspend's proof: the free cast puts an
// artifact onto the battlefield, so the time counters must be gone
// from the permanent that arrives (CR 400.7 — MoveCard clears
// counters on the way out of exile) and CR 702.62a's haste has
// nothing to do, because an artifact's {T} ability is not gated on
// summoning sickness.
//
// The mana ability is Gilded Lotus's, written in the #742 grammar:
// ONE colour pick that adds three tokens of that colour, not three
// independent picks.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04cf02dc-f053-414e-87d8-1537f25bcbf4",
		Name:         "Lotus Bloom",
		Completeness: effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{
			effects.Suspend(3, "{0}"),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Sacrifice: true},
			Produced: effects.OneColorOfAmount(3),
			Label:    "{T}, Sacrifice this artifact: Add three mana of any one color",
		}},
	})
}
