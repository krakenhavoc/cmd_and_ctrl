package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Locust God — Legendary Creature — God {4}{U}{R}, 4/4 (EDHREC
// rank 1723):
//
//	"Flying
//	 Whenever you draw a card, create a 1/1 blue and red Insect
//	 creature token with flying and haste.
//	 {2}{U}{R}: Draw a card, then discard a card.
//	 When The Locust God dies, return it to its owner's hand at the
//	 beginning of the next end step."
//
// The wheel deck's commander: every card drawn is a hasty flyer.
// Three abilities. The draw trigger is Sheoldred's own-draw
// condition, one Insect per card (EventDrawCard fires once per
// card, so a Windfall is a swarm). The loot is a CR 602 activation
// with a mana cost — draw first, then the discard prompt, so the
// drawn card is a legal discard (lootOne). The dies trigger
// schedules a CR 603.7 delayed trigger for the next end step
// carrying the God's own ID; when it fires, the card comes back to
// hand only if it is still in a graveyard — a God tucked into the
// command zone by CR 903.9, or reanimated in the meantime, is a
// different object and is left alone, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e025a714-02da-4b0c-8021-cf3e8dc9b19e",
		Name:            "The Locust God",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label: "{2}{U}{R}: Draw a card, then discard a card.",
			Cost:  ManaCost("{2}{U}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return lootOne(g, item, 1)
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverYouDraw("The Locust God — create a 1/1 Insect with flying and haste", Do(CreateToken{Template: TokenCard("1/1 blue and red Insect with flying and haste"), N: 1})),
			WhenThisDies("The Locust God — return it to hand at the next end step", func(g *game.Game, item *game.StackItem) error {
				return ScheduleDelayedTrigger{
					Label: "The Locust God — return to its owner's hand",
					Cards: []uuid.UUID{item.SourceCardID},
					Body:  returnListedGraveyardToHandBody,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
