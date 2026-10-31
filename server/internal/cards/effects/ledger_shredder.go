package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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

// ledgerShredderConnive makes the Shredder connive (CR 701.50a): the shared
// ConniveEach on the one creature, so the connive is an event a replacement
// can apply to (#4589, Leader, Super-Genius) and says it has connived
// (CR 701.50f, #4641) like every other.
func ledgerShredderConnive(g *game.Game, item *game.StackItem) error {
	return sourceConnives(NewContext(g, item))
}

// ledgerShredderDiscarded is the discard continuation the Shredder's own
// connive used before it went through ConniveEach (#4589). Nothing queues it
// now; it stays registered because a saved table can hold a discard prompt
// that names it (an on-disk key: never renamed, never reused).
var _ = game.RegisterDiscardThen("choose-cards/ledger-shredder-connive", func(g *game.Game, r game.DiscardsDone) error {
	if err := ledgerShredderCounter(g, r); err != nil {
		return err
	}
	// CR 701.50f, #4641: the Shredder has connived whatever the discard was.
	g.ConnivedForEffect(r.Source, r.Player)
	return nil
})

// ledgerShredderCounter is the +1/+1 counter for a nonland card discarded.
func ledgerShredderCounter(g *game.Game, r game.DiscardsDone) error {
	discarded := r.Discarded.Cards()
	if len(discarded) == 0 {
		return nil
	}
	card, ok := g.LookupCardForEffect(discarded[0])
	if !ok || card.IsLand() {
		return nil
	}
	return AddCounter{Target: r.Source, Kind: game.CounterPlusOne, N: 1}.Apply(keyedContext(g, r.Player, r.Source))
}
