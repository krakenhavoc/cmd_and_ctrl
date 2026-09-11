package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Locust God — Legendary Creature — God {4}{U}{R}, 4/4:
//
//	"Flying"
//	"Whenever you draw a card, create a 1/1 blue and red Insect
//	 creature token with flying and haste."
//	"{2}{U}{R}: Draw a card, then discard a card."
//	"When The Locust God dies, return it to its owner's hand at the
//	 beginning of the next end step."
//
// The other side of the draw deck's win condition: where Psychosis
// Crawler turns draws into damage, this turns them into a board. The
// tokens have HASTE, which is what makes a big draw spell a lethal
// attack the same turn rather than a nice board next turn.
//
// The dies-trigger is a CR 603.7 delayed trigger, the mechanism S22
// already had: the God goes to the graveyard, and at the beginning of
// the next end step it comes back to hand. Scheduling on death rather
// than returning immediately is what makes it a recursion engine that
// the opponent gets one window to answer.
//
// Sandbox note on the activated ability: "draw a card, THEN discard a
// card" is sequenced through the existing discard prompt, so the
// draw resolves first and the discard is chosen from the hand that
// includes it — which is the whole point of looting.
func init() {
	Register(Spec{
		OracleID:        "e025a714-02da-4b0c-8021-cf3e8dc9b19e",
		Name:            "The Locust God",
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventDrawCard},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor == source.Controller
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "The Locust God — create a 1/1 Insect",
						func(g *game.Game, item *game.StackItem) error {
							return CreateToken{
								Controller: item.Controller,
								Template:   InsectToken(),
								N:          1,
							}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					// CR 700.4: "dies" is battlefield → graveyard
					// specifically, not any leave-the-battlefield.
					return ev.CardID == source.InstanceID && ev.NewZone == game.ZoneGraveyard
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					dead := source.InstanceID
					return game.NewTriggeredItem(source, "The Locust God — return it at the next end step",
						func(g *game.Game, item *game.StackItem) error {
							return ScheduleDelayedTrigger{
								At:    game.StepEnd,
								Label: "The Locust God — return to its owner's hand",
								Cards: []uuid.UUID{dead},
								Effect: func(g *game.Game, _ *game.StackItem) error {
									return g.BounceToHandForEffect(dead)
								},
							}.Apply(NewContext(g, item))
						})
				},
			},
		},
		Activated: []ActivatedAbility{{
			Label: "{2}{U}{R}: Draw a card, then discard a card.",
			Cost:  game.AbilityCost{Mana: "{2}{U}{R}"},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				// "Then" is ordinary sequencing here, not the
				// prompt-continuation kind: the draw is synchronous,
				// so the discard prompt already sees the drawn card
				// in hand.
				if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return DiscardCards{Player: item.Controller, N: 1}.Apply(ctx)
			},
		}},
	})
}
