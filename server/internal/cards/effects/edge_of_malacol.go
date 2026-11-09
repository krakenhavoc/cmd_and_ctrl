package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Edge of Malacol — Plane — Belenon (#4870):
//
//	"If a creature you control would untap during your untap step, put two
//	 +1/+1 counters on it instead.
//	 Whenever chaos ensues, untap each creature you control."
//
// The static is a game.UntapReplacement read where the untap step untaps its
// set (CR 502.3, CR 614.1a): "you" is the planar controller, who is the
// active player, so only their tapped creatures are replaced; the creature
// stays tapped and gets the counters instead. A creature that is already
// untapped would not untap and gets nothing, and the chaos ability's untap is
// an ordinary untap, not "during your untap step", so it is not replaced.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb467901-a68d-4922-8ceb-867b582c8fe7",
		Name:         "Edge of Malacol",
		Completeness: CompletenessFull,
		UntapReplacements: []game.UntapReplacement{{
			Label: "If a creature you control would untap during your untap step, put two +1/+1 counters on it instead.",
			AppliesTo: func(_ *game.Game, source *game.Card, c game.Card, active uuid.UUID) bool {
				return c.IsCreature() && c.Controller == source.Controller && active == source.Controller
			},
			Instead: func(g *game.Game, source *game.Card, c game.Card) {
				_ = g.AddCounterByForEffect(source.Controller, c.InstanceID, game.CounterPlusOne, 2)
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Edge of Malacol — untap each creature you control", Do(UntapAllCreaturesYouControl{})),
		},
	})
}
