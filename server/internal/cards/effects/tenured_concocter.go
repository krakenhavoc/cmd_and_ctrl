package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tenured Concocter — Creature — Troll Druid {4}{G}, 4/5 (Secrets of
// Strixhaven, #4456):
//
//	"Vigilance
//	 Whenever this creature becomes the target of a spell or ability an
//	 opponent controls, you may draw a card.
//	 Infusion — This creature gets +2/+0 as long as you gained life this
//	 turn."
//
// The trigger is Cephalid Illusionist's EventBecomesTarget, fired at announce
// (CR 601.2c) for spells and abilities alike; the event's Actor is the
// controller of the targeting spell or ability, so one of this creature's own
// controller's does not trigger it (Black Bolt's check). "You may draw" is
// the optional-draw shape, so a player a draw limit stops is not asked.
//
// Infusion is an ability word. The pump is a layer 7c modification of itself,
// read live on every layer pass off the controller's turn tally of life gained
// (lifelink, a drain's gain half and a gain-life effect all count; an
// opponent's gain does not), so it turns on when life is gained and off with
// the turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0ecd4493-dfb4-4d56-a90e-08b74b474f39",
		Name:            "Tenured Concocter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			OptionalDraw(On(game.EventBecomesTarget, tenuredConcocterTargetedByAnOpponent,
				"Tenured Concocter — you may draw a card", Do(DrawCards{N: 1})),
				"Draw a card?", 1),
		},
		Static: []game.StaticAbility{{
			Label:    "Infusion — This creature gets +2/+0 as long as you gained life this turn.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			// A gain of life changes the turn's tally this reads, and
			// the layer cache must see it (EventChangeLife arm).
			DependsOnEvents: []game.EventKind{game.EventChangeLife},
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID &&
					b15LifeGainedThisTurn(g, source.Controller) > 0
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
			},
		}},
	})
}

// tenuredConcocterTargetedByAnOpponent is "this creature becomes the target
// of a spell or ability an opponent controls".
func tenuredConcocterTargetedByAnOpponent(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventBecomesTarget && ev.CardID == source.InstanceID &&
		ev.Actor != uuid.Nil && ev.Actor != source.Controller
}
