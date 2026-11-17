package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ayara's Oathsworn — Creature — Human Knight {1}{B}, 2/2 (March of the
// Machine: The Aftermath):
//
//	"Menace
//	 Whenever this creature deals combat damage to a player, if it has
//	 fewer than four +1/+1 counters on it, put a +1/+1 counter on it.
//	 Then if it has exactly four +1/+1 counters on it, search your
//	 library for a card, put it into your hand, then shuffle."
//
// One triggered ability with an intervening "if" (CR 603.4) on the fewer
// than four counters, checked as it triggers and as it resolves (InterveningIf,
// read from the Oathsworn's last-known information if it has left). The
// counter goes on first (CounterOnThis, through the counter pipeline, so a
// doubler or a replacement counts), and the "then if exactly four" is read
// after it, on the permanent as it now stands, so a card that arrives with
// three counters and doubles to five does not search. The search names no
// quality, so it is unrevealed (Servant of the Stinger's).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ff3936bb-3281-4260-84de-1f65dbfbe993",
		Name:            "Ayara's Oathsworn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(ayarasOathswornFewerThanFour, On(game.EventDealDamage,
				AllOf(ThisDealtCombatDamageToAPlayer,
					func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
						return source.Counters[game.CounterPlusOne] < 4
					}),
				"Ayara's Oathsworn — put a +1/+1 counter on it, then search if it has four",
				ayarasOathswornGrow)),
		},
	})
}

// ayarasOathswornPlusOneCounters is the number of +1/+1 counters the
// Oathsworn has, as it is now or as it last was on the battlefield.
func ayarasOathswornPlusOneCounters(ctx *Context) int {
	info, ok := ctx.SourcePermanent()
	if !ok {
		return 0
	}
	return info.Counters[game.CounterPlusOne]
}

var ayarasOathswornFewerThanFour = CondBoard{
	Name: "it has fewer than four +1/+1 counters on it",
	Test: func(ctx *Context) bool { return ayarasOathswornPlusOneCounters(ctx) < 4 },
}

// ayarasOathswornGrow is the trigger's effect: the counter, then the
// search only if the Oathsworn now has exactly four.
func ayarasOathswornGrow(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (CounterOnThis{Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if ayarasOathswornPlusOneCounters(ctx) != 4 {
		return nil
	}
	return SearchLibrary{
		Player:  ctx.Controller(),
		Dest:    game.ZoneHand,
		Limit:   1,
		Shuffle: true,
		Reason:  "Ayara's Oathsworn — search your library for a card",
	}.Apply(ctx)
}
