package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ledger Shredder — Creature — Bird Advisor {1}{U}, 1/3:
//
//	"Flying
//	 Whenever a player casts their second spell each turn, this
//	 creature connives. (Draw a card, then discard a card. If you
//	 discarded a nonland card, put a +1/+1 counter on this creature.)"
//
// "A player casts their second spell" is `g.CastTallyFor(ev.Actor).Total
// == 2` with no controller filter — Lotho, Corrupt Shirriff's exact
// condition, any seat's second spell, Ledger Shredder's own
// controller included.
//
// Connive (CR 702.146) has no shared primitive yet — the catalog's
// one other user (Lethal Scheme) declared it uncomposable because it
// has to connive an unbounded, dynamically-discovered SET of
// convoking creatures. This card connives exactly one fixed
// creature — itself — which composes cleanly out of what already
// exists: draw a card, queue a real player-chosen discard prompt
// (`QueueDiscardChoiceForEffect`, the same one every other chosen
// discard uses), and in its `Then` continuation — which runs only
// once the card has actually left the hand — look up what was
// discarded and put a +1/+1 counter on this creature if it wasn't a
// land. `AddCounter` is a no-op if the creature is no longer there to
// receive it (e.g. removed in response to its own trigger), which is
// the correct outcome for a counter that arrives after the fact.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e9117015-1050-44dd-a46b-e7ffe2085fae",
		Name:            "Ledger Shredder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return g.CastTallyFor(ev.Actor).Total == 2
			}, "Ledger Shredder — connive", ledgerShredderConnive),
		},
	})
}

// ledgerShredderConnive is CR 702.146's "draw a card, then discard a
// card. If you discarded a nonland card, put a +1/+1 counter on this
// creature" for exactly one fixed creature.
func ledgerShredderConnive(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	source := item.SourceCardID
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player:   item.Controller,
		Source:   source,
		N:        1,
		Question: "Ledger Shredder — connive: discard a card",
		Then: func(g *game.Game, _ uuid.UUID, discarded []uuid.UUID) error {
			if len(discarded) == 0 {
				return nil
			}
			card, ok := g.LookupCardForEffect(discarded[0])
			if !ok || card.IsLand() {
				return nil
			}
			return AddCounter{Target: source, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
		},
	})
	return nil
}
