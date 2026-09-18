package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marketback Walker — Artifact Creature — Construct {X}{X}, 0/0
// (EDHREC rank 4059):
//
//	"This creature enters with X +1/+1 counters on it.
//	 {4}: Put a +1/+1 counter on this creature.
//	 When this creature dies, draw a card for each +1/+1 counter on
//	 it."
//
// A colourless mana sink that turns into cards. {X}{X} is a terrible
// rate on the way in — six mana for a 3/3 — and that is not what the
// card is for: it is for a late game with nothing to do, where the
// Walker grows a counter per four mana and every counter is a card
// the moment somebody kills it. In a deck with a sacrifice outlet it
// is Ambition's Cost you can cash in whenever you like.
//
// {X}{X} IS TWO X SLOTS, so casting it for X=3 costs six mana and
// puts three counters on it. The engine charges both slots and the
// entry clause reads the single announced X, which is the printed
// arithmetic.
//
// THE DEATH TRIGGER READS THE LAST-KNOWN COUNTERS, not the card in
// the graveyard, which has none (CR 400.7 clears them the moment it
// moves). CR 603.10 has a dies-trigger look at the game state just
// before the permanent left, and b13LastKnownCounters is that read,
// reconstructed from the event log. So a Walker that entered with
// three counters, grew two, and then died draws five — and a Walker
// that had its counters removed in response draws only what it still
// held.
//
// The activated ability is a plain {4} with no tap, so it can be used
// at instant speed, any number of times, and the turn the Walker
// lands. That is the printed text and it is what makes the card a
// mana sink rather than a creature.
//
// SANDBOX SIMPLIFICATION, DECLARED — Goldvein Hydra's and Mikaeus's.
// An entry replacement cannot see the X announced for the spell (the
// stack item is gone by the time the entry pipeline runs), so the
// counters go on as the spell RESOLVES, a beat before the card moves
// from the stack to the battlefield, which is the last moment X is
// readable. They are on the card when it lands, so the 0/0 body never
// meets the state-based check without them, and a counter doubler
// still applies. The one observable difference is that a "whenever
// you put counters on a permanent" payoff does not see them — weaker,
// never stronger.
//
// The same engine-side gap those cards declare applies here: cast for
// X=0 the Walker is a printed 0/0 that never had a counter, which the
// toughness state check deliberately skips as a placeholder, so it
// survives and can be grown with {4}. That is the one way this card
// plays stronger than printed, and it is declared to players.
func init() {
	Register(Spec{
		OracleID:     "0405e0a9-6d02-4691-bdb8-59c72b824dab",
		Name:         "Marketback Walker",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The X +1/+1 counters are put on it as the spell resolves, a beat before it enters, so effects that watch you put counters on a permanent don't see them.",
			"Cast for X=0, it stays on the battlefield as a 0/0 instead of dying at once, and can still be grown with {4}.",
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: ctx.X()}.Apply(ctx)
		},
		Activated: []ActivatedAbility{{
			Label:  "{4}: Put a +1/+1 counter on this creature.",
			Cost:   ManaCost("{4}"),
			Effect: putCounterOnSelf,
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Key: "Marketback Walker — draw a card for each +1/+1 counter on it",
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				counters := b13LastKnownCounters(g, source.InstanceID, game.CounterPlusOne)
				return game.NewTriggeredItem(source, "Marketback Walker — draw a card for each +1/+1 counter on it",
					func(g *game.Game, item *game.StackItem) error {
						if counters <= 0 {
							return nil
						}
						return DrawCards{Player: item.Controller, N: counters}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
