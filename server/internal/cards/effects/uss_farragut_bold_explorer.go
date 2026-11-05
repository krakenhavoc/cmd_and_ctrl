package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3b2e3deb-1f21-4311-b8a9-4153089ac5ff",
		Name:         "U.S.S. Farragut, Bold Explorer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("U.S.S. Farragut — you may put a creature card from your hand onto the battlefield; it gains haste, and at the next end step goes to the bottom of your library",
				func(g *game.Game, item *game.StackItem) error {
					return PutFromHandOntoBattlefield{
						Match:    Creature(),
						Optional: true,
						Label:    "U.S.S. Farragut — you may put a creature card from your hand onto the battlefield",
						ThenTail: trekB43FarragutHasteAndBottom,
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(7, 4, 4),
			ThresholdKeywords(7, "flying"),
		),
	})
}
