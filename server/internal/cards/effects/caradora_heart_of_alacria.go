package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Caradora, Heart of Alacria — Legendary Creature — Human Knight
// {2}{G}{W}, 4/2 (Aetherdrift, #4464):
//
//	"When Caradora enters, you may search your library for a Mount or
//	 Vehicle card, reveal it, put it into your hand, then shuffle.
//	 If one or more +1/+1 counters would be put on a creature or Vehicle
//	 you control, that many plus one +1/+1 counters are put on it instead."
//
// The search is optional and revealed; the replacement is Hardened Scales'
// with a Vehicle named beside a creature, so an uncrewed Vehicle gets the
// extra counter too, and a creature or Vehicle entering with counters
// gets one more (CR 614.1, 122.6). Two Caradoras each add one (CR 616.1).
// "Mount" is a word of the card's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a9d60d80-bcef-45e0-8ded-d70bd3c2780f",
		Name:         "Caradora, Heart of Alacria",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Caradora — search for a Mount or Vehicle card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: b13MountOrVehicleCard,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Caradora — a Mount or Vehicle card",
				}.Apply(NewContext(g, item))
			}),
		},
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventCounterPlaced},
			AppliesTo: plusOneCounterPlacementOnYourCreatureOrVehicle,
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				ev.CounterDelta += 1
				return nil
			},
			Controller: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Caradora: +1 +1/+1 counter",
		}},
	})
}
