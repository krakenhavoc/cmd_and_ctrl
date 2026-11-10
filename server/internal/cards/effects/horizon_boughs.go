package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Horizon Boughs — Plane — Pyrulea (#4870):
//
//	"All permanents untap during each player's untap step.
//	 Whenever chaos ensues, you may search your library for up to three basic
//	 land cards, put them onto the battlefield tapped, then shuffle."
//
// An untap-step permission read from the face-up plane (every player's untap
// step, every permanent: CR 502.3 still lets restrictions such as "doesn't
// untap" win). The chaos search is optional by its "up to" and puts the lands
// onto the battlefield tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "065346cd-cd81-478e-8a62-2c7ea421bdcd",
		Name:         "Horizon Boughs",
		Completeness: CompletenessFull,
		UntapStep: []game.UntapStepPermission{{
			Label:     "All permanents untap during each player's untap step.",
			AppliesTo: func(*game.Game, *game.Card, uuid.UUID) bool { return true },
			Untaps:    func(*game.Game, *game.Card, *game.Card) bool { return true },
		}},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Horizon Boughs — search your library for up to three basic land cards, put them onto the battlefield tapped, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         3,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        "Horizon Boughs — up to three basic lands, onto the battlefield tapped",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
