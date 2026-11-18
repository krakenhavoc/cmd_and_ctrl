package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreadhound — Creature — Demon Dog {4}{B}{B}, 6/6 (Innistrad: Midnight
// Hunt, S82):
//
//	"When this creature enters, mill three cards.
//	 Whenever a creature dies or a creature card is put into a graveyard
//	 from a library, each opponent loses 1 life."
//
// The second line is one trigger with two conditions (CR 603.2c), Syr
// Konrad's shape: a creature dying, the Dreadhound itself included (it
// says "a creature", not "another"), or a creature card put into a
// graveyard from a library, whether milled, searched for or put there by
// a look. A creature card discarded from a hand or countered on the stack
// is not from a library. The enters trigger mills the controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5ed533a7-82b0-4a71-a9e9-4dffe9d18851",
		Name:         "Dreadhound",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dreadhound — mill three cards", Do(MillCards{N: 3})),
			TriggerWithPurpose(OnAny([]game.EventKind{game.EventLTB, game.EventMill, game.EventZoneMove},
				func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Kind == game.EventLTB {
						_, died := diedCreature(ev, g)
						return died
					}
					if ev.NewZone != game.ZoneGraveyard || ev.OldZone != game.ZoneLibrary {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.IsCreature() && !c.IsToken()
				}, "Dreadhound — each opponent loses 1 life", Do(EachOpponentLosesLife{N: 1})),
				game.Purpose{DeathPayoff: true}),
		},
	})
}
