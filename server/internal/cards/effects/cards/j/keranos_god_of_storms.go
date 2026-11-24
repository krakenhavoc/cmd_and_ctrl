package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Keranos, God of Storms — Legendary Enchantment Creature — God {3}{U}{R},
// 6/5 (Secret Lair Drop):
//
//	"Indestructible
//	 As long as your devotion to blue and red is less than seven, Keranos isn't
//	 a creature.
//	 Reveal the first card you draw on each of your turns. Whenever you reveal
//	 a land card this way, draw a card. Whenever you reveal a nonland card this
//	 way, Keranos deals 3 damage to any target."
//
// The God clause is the shared Theros static. The reveal and the two "whenever
// you reveal" abilities are one step each, as Inquisitor Eisenhorn's are: the
// first card its controller draws on their own turn (the per-turn draw count is
// one) triggers the ability matching its kind, and its first act as it
// resolves is to reveal the card to the table (if it is still in the hand).
// Every first card is a land or a nonland, so exactly one of the two triggers.
// The nonland ability targets as it goes on the stack, which a card revealed
// "this way" cannot be told from, since the trigger exists only for a nonland
// card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2bd6f13a-ae02-426d-8f99-c20c5c08cf55",
		Name:            "Keranos, God of Storms",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static:          []game.StaticAbility{effects.GodUnlessDevotionToColors(7, "U", "R")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, keranosFirstDrawIs(true), "Keranos, God of Storms — reveal the land card you drew; draw a card",
				func(g *game.Game, item *game.StackItem) error {
					keranosReveal(g, item)
					return effects.DrawCards{N: 1}.Apply(effects.NewContext(g, item))
				}),
			effects.TriggerWithPurpose(effects.Targeting(effects.On(game.EventDrawCard, keranosFirstDrawIs(false),
				"Keranos, God of Storms — reveal the nonland card you drew; 3 damage to any target",
				func(g *game.Game, item *game.StackItem) error {
					keranosReveal(g, item)
					return effects.SourceDealsDamageToEachLegalTarget(3)(g, item)
				}), effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 3))),
		},
	})
}

// keranosFirstDrawIs is the first card the controller drew this turn, on their
// own turn, being a land (or not).
func keranosFirstDrawIs(land bool) effects.When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Actor != source.Controller || g.TurnTallyFor(source.Controller).CardsDrawn != 1 {
			return false
		}
		if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) || g.Seats[g.Turn.ActiveSeat].ID != source.Controller {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.IsLand() == land
	}
}

// keranosReveal shows the drawn card to the table, if it is still in its
// owner's hand.
func keranosReveal(g *game.Game, item *game.StackItem) {
	if item.Trigger == nil {
		return
	}
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneHand {
		return
	}
	g.RevealForEffect(game.RevealSpec{
		Player: item.Controller,
		Source: item.SourceCardID,
		Reason: "Keranos, God of Storms — reveal the first card you draw",
		Cards:  []uuid.UUID{id},
	})
}
