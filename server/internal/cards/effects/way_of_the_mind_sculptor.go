package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Way of the Mind Sculptor — Legendary Enchantment {4}{U} (Reality
// Fracture, tracker #2795):
//
//	"When Way of the Mind Sculptor enters, empower Jace 5.
//	 Whenever you activate a loyalty ability, if you removed two or more
//	 loyalty counters to activate it, draw a card."
//
// Empower Jace is the keyword action (ADR 0139). The trigger watches the
// activation announcement (the event's loyalty bit, printed or granted)
// and reads the ability's loyalty cost off the planeswalker by its label
// (frLoyaltyAbilityRemovedAtLeast): a −2 or lower counts, a plus ability
// does not. The intervening "if" (CR 603.4) is checked when it would
// trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "94be2e86-ba0e-4b8d-9b9c-114adacb26db",
		Name:         "Way of the Mind Sculptor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Way of the Mind Sculptor — empower Jace 5", Do(EmpowerJace{N: 5})),
			On(game.EventActivateAbility,
				func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
					return frYouActivatedALoyaltyAbility(ev, source, ch, g) && frLoyaltyAbilityRemovedAtLeast(ev, g, 2)
				},
				"Way of the Mind Sculptor — draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
