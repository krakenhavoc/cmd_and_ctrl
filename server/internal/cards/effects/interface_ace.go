package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Interface Ace — Artifact Creature — Robot Pilot {1}{W}, 0/4:
//
//	"This creature saddles Mounts and crews Vehicles using its toughness
//	 rather than its power.
//	 Whenever this creature becomes tapped during your turn, untap it.
//	 This ability triggers only once each turn."
//
// The first line is CrewsUsingToughness (ADR 0186 Decision 4): the Ace
// counts as 4 for crew and saddle, and its real power stays 0.
//
// The trigger is Dionus's, on the Ace itself instead of granted to a
// tribe. "Becomes tapped" is any tap, including the Ace being tapped to
// pay a crew or saddle cost (ThisBecameTapped), and it must be its
// controller's turn. "Only once each turn" is read off the engine's
// per-object trigger tally (TriggeredThisTurn), which counts the trigger
// the moment it goes on the stack, so a second tap in the same turn does
// nothing.
//
// No simplification.
const interfaceAceLabel = "Interface Ace — untap this creature"

func init() {
	Register(Spec{
		OracleID:     "c4272676-a38a-4b23-988e-9cdb2feb46c1",
		Name:         "Interface Ace",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CrewsUsingToughness(game.CostPowerCrew | game.CostPowerSaddle),
		},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack},
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return ThisBecameTapped(ev, source, lki, g) &&
						IsYourTurn(g, source.Controller) &&
						!b11TriggeredThisTurn(g, source.InstanceID, interfaceAceLabel)
				},
				interfaceAceLabel, b31UntapSelf),
		},
	})
}
