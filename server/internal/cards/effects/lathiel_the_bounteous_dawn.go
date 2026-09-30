package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lathiel, the Bounteous Dawn — Legendary Creature — Unicorn
// {2}{G}{W}, 2/2 (EDHREC rank 3472):
//
//	"Lifelink
//	 At the beginning of each end step, if you gained life this turn,
//	 distribute up to that many +1/+1 counters among any number of
//	 other target creatures."
//
// The lifegain deck's counters engine: every point of life gained
// becomes a counter at the end of the turn — any player's end step,
// as printed. The intervening if is b15LifeGainedThisTurn, checked
// at announce and again at resolution (CR 603.4); the targets are
// "any number of other target creatures", Appa, Steadfast Guardian's
// Min-0 unbounded clause with Lathiel kept out by name.
//
// "Distribute up to that many" is a divided clause (#1657, ADR 0065's
// 2026-09-28 (b) amendment): UpTo(DivideBy(DivideLifeYouGainedThisTurn)).
// The amount is the life gained this turn, read as the trigger is put
// on the stack — the target walk — and never again, which is the
// ruling: "Gaining more life in response to the triggered ability
// won't change how many counters will be distributed, nor will it
// change the distribution." The controller announces the split with
// the targets; UpTo lets it add up to LESS than the life gained, and
// each chosen creature still gets at least one (the ruling again), so
// three life can be 1 + 1 on two creatures. Choosing no creature at
// all is the clause's Min 0 — "you may choose no targets". At
// resolution each creature still a legal target gets exactly its
// announced share (PutDividedCounters); a departed creature's share
// is lost, not moved.
//
// With no other creature on the battlefield the trigger is dropped
// before the prompt (CR 603.3d); the printed card would put it on
// the stack to do nothing. No difference anyone can observe.
func init() {
	Register(Spec{
		OracleID:        "d3c56fc4-3611-41b1-952e-4c5311b1510b",
		Name:            "Lathiel, the Bounteous Dawn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginEndStep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b33EndStepAndYouGainedLifeThisTurn(ev, source, g)
			},
			Targets: Another(TargetCreature("any number of other target creatures")).WithCount(0, 0).
				Dividing(UpTo(DivideBy(DivideLifeYouGainedThisTurn))),
			Key:    "Lathiel, the Bounteous Dawn — distribute +1/+1 counters among the chosen creatures",
			Effect: lathielDistributeCounters,
		}},
	})
}

// lathielDistributeCounters is Lathiel's body: the intervening if
// re-checked (CR 603.4), then each surviving target's announced share
// of +1/+1 counters.
func lathielDistributeCounters(g *game.Game, item *game.StackItem) error {
	if b15LifeGainedThisTurn(g, item.Controller) <= 0 {
		return nil
	}
	return PutDividedCounters(NewContext(g, item), game.CounterPlusOne)
}
