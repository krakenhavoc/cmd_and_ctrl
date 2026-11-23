package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// U.S.S. Farragut, Bold Explorer — {6} Legendary Artifact — Spacecraft
// 4/4 (Star Trek Commander, #4974):
//
//	"Landfall — Whenever a land you control enters, you may put a
//	 creature card from your hand onto the battlefield. That creature
//	 gains haste. At the beginning of the next end step, put it on the
//	 bottom of your library. When you do, draw a card.
//	 Station (… It's an artifact creature at 7+.)
//	 7+ | Flying"
//
// Landfall is above the station line, so it triggers at any charge
// count (CR 721.4). The put is Quicksilver Amulet's optional
// PutFromHandOntoBattlefield; its tail grants haste with no duration
// (CR 611.2a), pinned to the creature as that object, and schedules the
// delayed trigger (CR 603.7) with the object's epoch. The delayed
// trigger does nothing to a creature that has left or come back as a new
// object (CR 603.7c), and the draw is a reflexive trigger (CR 603.12)
// that exists only when the creature really reached the library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b2e3deb-1f21-4311-b8a9-4153089ac5ff",
		Name:         "U.S.S. Farragut, Bold Explorer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("U.S.S. Farragut — you may put a creature card from your hand onto the battlefield; it gains haste, and at the next end step goes to the bottom of your library",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PutFromHandOntoBattlefield{
						Match:    effects.Creature(),
						Optional: true,
						Label:    "U.S.S. Farragut — you may put a creature card from your hand onto the battlefield",
						ThenTail: effects.TrekB43FarragutHasteAndBottom,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(7, 4, 4),
			effects.ThresholdKeywords(7, "flying"),
		),
	})
}
