package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mirelurk Queen — Creature — Crab Mutant {4}{U}, 4/4:
//
//	"Vigilance
//	 When this creature enters, target player gets two rad counters.
//	 Whenever one or more nonland cards are milled, draw a card, then put
//	 a +1/+1 counter on this creature. This ability triggers only once
//	 each turn."
//
// #2809. The mill trigger is WheneverOneOrMoreNonlandCardsAreMilled: one
// trigger per mill, whoever milled (the rad counters' own trigger
// included), and none for a mill of only lands or for a surveil, which
// is not a mill (CR 701.25a). "Only once each turn" is checked as it
// triggers (TriggersOnlyOnceEachTurn), so the second mill of a turn
// triggers nothing. The draw comes first and the counter second, in the
// printed order; a Queen that left the battlefield in response still
// draws, and the counter has nowhere to go (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "116ef9ea-d199-49b8-89c7-d61dd92711f5",
		Name:            "Mirelurk Queen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Mirelurk Queen — target player gets two rad counters",
				targetPlayerGetsTwoRadCounters), TargetPlayer("target player")),
			TriggersOnlyOnceEachTurn(WheneverOneOrMoreNonlandCardsAreMilled(
				"Mirelurk Queen — draw a card, then put a +1/+1 counter on this creature",
				Do(DrawCards{N: 1}, CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}

// targetPlayerGetsTwoRadCounters is "target player gets two rad
// counters" (Mirelurk Queen, The Master, Transcendent): the targeted
// player, if still a legal target as it resolves (CR 608.2b), gets
// them from the item's controller.
func targetPlayerGetsTwoRadCounters(g *game.Game, item *game.StackItem) error {
	return playerGetsRadCounters(g, item.Controller, targetedPlayerOf(NewContext(g, item)), 2)
}
