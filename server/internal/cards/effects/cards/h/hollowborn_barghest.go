package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hollowborn Barghest — Creature — Demon Dog {5}{B}{B}, 7/6 (Shadowmoor, S82
// OLD b529, #5591):
//
//	"At the beginning of your upkeep, if you have no cards in hand, each
//	 opponent loses 2 life.
//	 At the beginning of each opponent's upkeep, if that player has no cards in
//	 hand, they lose 2 life."
//
// Two printed triggers, each behind an intervening "if" (CR 603.4) checked as
// it triggers and again as it resolves. The second reads the player whose
// upkeep it is off the event, not the Barghest's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b08f344-1ff5-4d04-a435-d3a1c5e4211b",
		Name:         "Hollowborn Barghest",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(
				effects.CondBoard{Name: "you have no cards in hand", Test: func(ctx *effects.Context) bool {
					return effects.OldB529HandIsEmpty(ctx.Game, ctx.Controller())
				}},
				effects.AtYourUpkeep("Hollowborn Barghest — if you have no cards in hand, each opponent loses 2 life",
					effects.Do(effects.EachOpponentLosesLife{N: 2}))),
			hollowbornBarghestOpponentUpkeep(),
		},
	})
}

// hollowbornBarghestOpponentUpkeep is the second trigger: the opponent whose
// upkeep it is, with no cards in hand, loses 2 life.
func hollowbornBarghestOpponentUpkeep() game.TriggeredAbility {
	t := effects.AtEachOpponentsUpkeep("Hollowborn Barghest — if that player has no cards in hand, they lose 2 life",
		func(g *game.Game, item *game.StackItem) error {
			victim := item.Trigger.Event.Actor
			if !effects.OldB529HandIsEmpty(g, victim) {
				return nil
			}
			return g.ChangePlayerLifeForEffect(effects.NewContext(g, item).Source(), victim, -2)
		})
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return effects.OldB529HandIsEmpty(g, ev.Actor) && (when == nil || when(ev, source, lki, g))
	}
	return effects.InterveningIf(effects.OldB529PlayerHasNoCardsInHand{}, t)
}
