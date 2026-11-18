package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Denry Klin, Editor in Chief — Legendary Creature — Cat Advisor {2}{W}{U},
// 2/2 (Streets of New Capenna Commander, #5491):
//
//	"Denry Klin enters with your choice of a +1/+1, first strike, or
//	 vigilance counter on it.
//	 Whenever a nontoken creature you control enters, if Denry Klin has
//	 counters on it, put the same number of each kind of counter on that
//	 creature."
//
// The first line is an as-enters choice (CR 614.12) asked as Denry arrives,
// the way riot is. The trigger watches every nontoken creature entering under
// the controller's control, Denry included, so Denry's own counter is copied
// onto Denry. "If Denry has counters" is an intervening "if" (CR 603.4),
// read from Denry's last-known counters if it has left (CR 608.2h). The same
// number of each kind: a creature entering with counters of its own gets more
// on top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba477a45-84cc-4a85-98e4-50dca3e58788",
		Name:         "Denry Klin, Editor in Chief",
		Completeness: CompletenessFull,
		AsEnters: func(card *game.Card, ctx *Context) error {
			kinds := []string{game.CounterPlusOne, game.CounterFirstStrike, game.CounterVigilance}
			options := make([]game.ChoiceOption, len(kinds))
			for i, kind := range kinds {
				options[i] = game.ChoiceOption{Label: counterOptionLabel(kind)}
			}
			// Queued straight off the entry (no stack item), like every as-enters
			// choice; energyChosenCounter reads the controller, then the permanent
			// and the placer from the carry.
			ctx.Game.QueueOptionPickForEffect(game.OptionPickPrompt{
				Chooser:  card.Controller,
				Source:   card.InstanceID,
				Question: "Denry Klin — choose a counter to enter with",
				Options:  options,
				ThenKey:  energyChosenCounter,
				Carry:    []uuid.UUID{card.Controller, card.InstanceID, card.Controller},
			})
			return nil
		},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{Name: "Denry Klin has counters on it", Test: func(ctx *Context) bool {
				return len(denryKlinCounters(ctx)) > 0
			}}, On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && !IsToken(c) && sncB07HasCounters(source.Counters)
			}, "Denry Klin — put the same number of each kind of counter on that creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := NewContext(g, item)
					counters := denryKlinCounters(ctx)
					creature := item.Trigger.Event.CardID
					if !onBattlefield(g, creature) {
						return nil
					}
					kinds := make([]string, 0, len(counters))
					for kind := range counters {
						kinds = append(kinds, kind)
					}
					sort.Strings(kinds)
					for _, kind := range kinds {
						if err := g.AddCounterByForEffect(item.Controller, creature, kind, counters[kind]); err != nil {
							return err
						}
					}
					return nil
				})),
		},
	})
}

// denryKlinCounters is Denry's counters, with a kind at zero left out: now,
// or as it last existed if it has left the battlefield.
func denryKlinCounters(ctx *Context) map[string]int {
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	out := map[string]int{}
	for kind, n := range info.Counters {
		if n > 0 {
			out[kind] = n
		}
	}
	return out
}
