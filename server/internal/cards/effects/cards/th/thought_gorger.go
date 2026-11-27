package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thought Gorger — Creature — Horror {2}{B}{B}, 2/2 (Rise of the Eldrazi,
// S82 OLD batch 482, #5581):
//
//	"Trample
//	 When this creature enters, put a +1/+1 counter on it for each card in
//	 your hand. If you do, discard your hand.
//	 When this creature leaves the battlefield, draw a card for each +1/+1
//	 counter on it."
//
// "If you do" means counters really went on: an empty hand, a Gorger that
// has already gone, or a counter prevention discards nothing. The leaves
// trigger reads the counters as it last existed on the battlefield
// (CR 603.10a, 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ec5dc165-c760-4060-99fb-ec50f87374cb",
		Name:            "Thought Gorger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Thought Gorger — put a +1/+1 counter on it for each card in your hand; if you do, discard your hand",
				thoughtGorgerEnters),
			effects.WhenThisLeaves("Thought Gorger — draw a card for each +1/+1 counter on it",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.DrawCards{N: effects.SocB9PlusOneCountersOnSource(ctx)}.Apply(ctx)
				}))),
		},
	})
}

// thoughtGorgerEnters puts a counter per card in hand, then discards the hand
// if any counter was put.
func thoughtGorgerEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Hand == nil || p.Hand.Size() == 0 || !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	before := thoughtGorgerCounters(g, item)
	if err := (effects.CounterOnThis{Kind: game.CounterPlusOne, N: p.Hand.Size()}).Apply(ctx); err != nil {
		return err
	}
	if thoughtGorgerCounters(g, item) <= before {
		return nil
	}
	_, err := effects.DiscardWholeHand(g, item.Controller)
	return err
}

func thoughtGorgerCounters(g *game.Game, item *game.StackItem) int {
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return 0
	}
	return c.Counters[game.CounterPlusOne]
}
