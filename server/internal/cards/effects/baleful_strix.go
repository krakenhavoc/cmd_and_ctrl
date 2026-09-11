package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baleful Strix — Artifact Creature — Bird, {U}{B}, 1/1:
//
//	"Flying, deathtouch"
//	"When this creature enters, draw a card."
//
// Two mana that replaces itself and then blanks the biggest creature
// on the table for the rest of the game. Every word of it is live:
// flying and deathtouch are both among the twelve keywords the
// combat code honours, so the Strix really does block a 12/12 and
// kill it.
//
// The ETB draw is a TRIGGERED ability, not an OnETB hook. The
// difference is observable and the printed card is unambiguous —
// "When this creature enters" goes on the stack, so opponents get a
// response window and the draw can be countered or responded to.
// OnETB would resolve it invisibly as part of the creature entering.
//
// "Draw a card" is the controller's draw, so it reads
// item.Controller off the trigger rather than the card's owner: a
// Strix stolen with a Control Magic draws for whoever controls it
// when it enters.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "37688720-03de-4eca-a82d-a0afe8d58adc",
		Name:            "Baleful Strix",
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Baleful Strix — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
