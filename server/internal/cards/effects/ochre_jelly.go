package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ochre Jelly — Creature — Ooze {X}{G}, 0/0 (Adventures in the Forgotten
// Realms):
//
//	"Trample
//	 This creature enters with X +1/+1 counters on it.
//	 Split — When this creature dies, if it had two or more +1/+1 counters on
//	 it, create a token that's a copy of it at the beginning of the next end
//	 step. The token enters with half that many +1/+1 counters on it, rounded
//	 down."
//
// The entry counters are the X-counter self-replacement (Benevolent Hydra's
// XCounters). The death trigger carries an intervening "if" (CR 603.4) read
// from the +1/+1 counters the Jelly had as it last existed (CR 603.10a,
// Iron Apprentice's LastKnownCountersForEffect); that count cannot change
// before the trigger resolves. Resolving it schedules a delayed trigger
// (CR 603.7) for the next end step, carrying the half-count as plain data.
//
// At that end step the token copies the dead Jelly (CR 707.2: its
// last-known copiable values while the engine still holds them, else the card
// it now is). A token is never cast, so its own "enters with X counters" has
// X = 0 (CR 107.3f) and adds none; the token enters with only the half, so
// "that many" is the number the Jelly had, halved and rounded down.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "2b757d32-09a3-49cd-ada4-84b5335178ce",
		Name:                       "Ochre Jelly",
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"trample"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return cardDied(ev, source) && g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] >= 2
			}, "Ochre Jelly — create a token that's a copy of it at the beginning of the next end step, with half that many +1/+1 counters",
				ochreJellySplit),
		},
	})
}

// ochreJellySplit is the dies trigger's body: it fixes the number of counters
// the token will enter with and schedules the token for the next end step.
func ochreJellySplit(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	had := g.LastKnownCountersForEffect(ctx.Trigger().Event.CardID)[game.CounterPlusOne]
	if had < 2 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label:  "Ochre Jelly — create a token that's a copy of it with half that many +1/+1 counters",
		Cards:  []uuid.UUID{item.SourceCardID},
		Body:   ochreJellySplitBody,
		Params: game.EffectParams{Amount: had / 2},
	}.Apply(ctx)
}

// ochreJellySplitBody is the delayed trigger: the token copy with the counters.
// An on-disk key: never renamed, never reused.
var ochreJellySplitBody = game.DelayedBody("ochre-jelly/split-token",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		if p.Amount <= 0 || len(item.Targets) == 0 {
			return nil
		}
		tmpl, ok := TokenCopyTemplate(g, item.Targets[0].ID)
		if !ok {
			return nil
		}
		_, err := g.CreateTokensForEffect(item.Controller, copiedToken(tmpl), 1,
			game.TokenEntryOptions{Counters: map[string]int{game.CounterPlusOne: p.Amount}})
		return err
	})
