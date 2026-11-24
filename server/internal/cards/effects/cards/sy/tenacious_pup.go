package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tenacious Pup — Creature — Wolf {G}, 1/2 (Mystery Booster 2, #5948):
//
//	"When Tenacious Pup enters the battlefield, you gain 1 life. You get a
//	 one-time boon with 'When you cast a creature spell, that creature
//	 enters the battlefield with an additional +1/+1 counter, trample
//	 counter, and vigilance counter on it.'"
//
// The boon is EmblemSpec{Boon: true} (ADR 0099 amendment 2026-11-24): it
// sits beside its owner's emblems, its one trigger triggers only once,
// and it is gone the moment it has. The trigger marks the spell it
// watched, as Wildgrowth Archaic's does, so the three counters ride the
// creature's entry and every entry replacement sees them; a spell that
// is countered or leaves the stack gets nothing, and the boon is spent
// all the same.
//
// No simplification.
func init() {
	const boonText = "When you cast a creature spell, that creature enters the battlefield with an additional +1/+1 counter, trample counter, and vigilance counter on it."
	effects.Register(effects.Spec{
		OracleID:     "70c9cfcf-6f23-41de-af63-dbe212018035",
		Name:         "Tenacious Pup",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tenacious Pup — you gain 1 life and get a one-time boon",
				effects.Do(effects.GainLife{Amount: 1}, effects.CreateEmblem{})),
		},
		Emblem: &effects.EmblemSpec{
			Label:     "Tenacious Pup boon",
			Text:      boonText,
			Boon:      true,
			Triggered: []game.TriggeredAbility{tenaciousPupBoonTrigger()},
		},
	})
}

// tenaciousPupBoonTrigger is the boon's ability: on resolution it marks
// the creature spell it watched with the three counters.
func tenaciousPupBoonTrigger() game.TriggeredAbility {
	const label = "Tenacious Pup boon — that creature enters with an additional +1/+1 counter, trample counter, and vigilance counter"
	t := effects.On(game.EventCast, effects.YouCast(effects.Creature()), label, func(g *game.Game, item *game.StackItem) error {
		for _, kind := range []string{game.CounterPlusOne, game.CounterTrample, game.CounterVigilance} {
			g.MarkSpellEntersWithCountersForEffect(item.Params.Object, game.PromisedCounter{
				Kind:       kind,
				Count:      1,
				Source:     item.SourceCardID,
				SourceName: "Tenacious Pup",
			})
		}
		return nil
	})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Object = game.ObjectRef{ID: ev.CardID, Epoch: g.ObjectEpochForEffect(ev.CardID)}
		return item
	}
	return t
}
