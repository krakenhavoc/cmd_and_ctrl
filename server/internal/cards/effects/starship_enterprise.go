package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starship Enterprise — Legendary Artifact — Spacecraft {5}, 4/6 (Star
// Trek Commander, #4974):
//
//	"Landfall — Whenever a land you control enters, put four charge
//	 counters on Starship Enterprise.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.
//	 It's an artifact creature at 16+.)
//	 16+ | Flying
//	 Whenever Starship Enterprise deals combat damage to a player, draw
//	 that many cards."
//
// Station is the ordinary keyword. Landfall puts the counters on this
// permanent through the CR 614 counter window (CounterOnThis). The
// combat-damage trigger is on the 16+ line, so it exists only at sixteen
// or more charge counters (CR 721.2a, TriggeredAbility.ActiveWhen);
// "that many" is the damage dealt, read off the triggering event and
// fixed as it triggered (Cold-Eyed Selkie's read), and the draw is not
// optional.
//
// No simplification.
func init() {
	draw := WheneverThisDealsCombatDamageToAPlayer("Starship Enterprise — draw that many cards", func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
			return nil
		}
		return DrawCards{Player: item.Controller, N: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
	})
	draw.ActiveWhen = AtChargeCounters(16)
	Register(Spec{
		OracleID:     "9b627fa4-af4c-482b-b1cf-2ad15938602c",
		Name:         "Starship Enterprise",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Activated:    []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(16, 4, 6),
			ThresholdKeywords(16, "flying"),
		),
		Triggered: []game.TriggeredAbility{
			Landfall("Starship Enterprise — put four charge counters on it", Do(CounterOnThis{Kind: game.CounterCharge, N: 4})),
			draw,
		},
	})
}
