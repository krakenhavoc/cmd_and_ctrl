package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b532_helpers.go — shared pieces for the S82 OLD batch 532 cards
// (Shadowmoor -1/-1 and other counters, #5591).

// OldB532MinusOneCountersAtEntry is "this creature enters with N -1/-1 counters
// on it" (Chainbreaker, Leech Bonder, Morselhoarder, Grief Tyrant): CR 614.1c
// entry counters, so they are there as it arrives and Hardened Scales sees them.
func OldB532MinusOneCountersAtEntry(name string, n int) game.ReplacementEffect {
	r := B10EntersWithCounters(game.CounterMinusOne, n, name+": enters with "+NumberWord(n)+" -1/-1 counters")
	r.RowLabel = "Enters with " + NumberWord(n) + " -1/-1 counters on it"
	return r
}

// OldB532UntapCost is "<mana>, {Q}" (CR 107.6, AbilityCost.UntapSelf): the source
// must be tapped to pay it.
func OldB532UntapCost(mana string) game.AbilityCost {
	return game.AbilityCost{Mana: mana, UntapSelf: true}
}

// OldB532UntapThisForAMinusOneCounter is the printed untapper of the Shadowmoor
// Kithkin and Elementals, "Put a -1/-1 counter on this creature: Untap this
// creature." The counter is the whole cost (Devoted Druid's untapper): it has no
// {T}, so it can be paid while tapped, and it uses the stack.
func OldB532UntapThisForAMinusOneCounter() ActivatedAbility {
	return ActivatedAbility{
		Label:   "Put a -1/-1 counter on this creature: Untap this creature.",
		Purpose: game.Purpose{Answers: game.AnswerValue},
		Cost:    AddCounterToThis(game.CounterMinusOne, 1),
		Effect: func(g *game.Game, item *game.StackItem) error {
			return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
		},
	}
}

// OldB532RemoveAMinusOneCounterFromTarget is "remove a -1/-1 counter from target
// creature" (Chainbreaker): the first legal target as the ability resolves
// (CR 608.2b), the counter through the effect-side adder at -1 so the layer pass
// sees it go. A creature with none has nothing to lose.
func OldB532RemoveAMinusOneCounterFromTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.Counters[game.CounterMinusOne] <= 0 {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterMinusOne, N: -1}.Apply(ctx)
}

// OldB532MinusOneCountersOnTargetForEachOnThis is Grief Tyrant's dies effect:
// "put a -1/-1 counter on target creature for each -1/-1 counter on this
// creature". The count is the counters it had as it died (last-known information,
// CR 603.10a, 608.2h), read as the ability resolves.
func OldB532MinusOneCountersOnTargetForEachOnThis(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, _ := ctx.TriggeringPermanent()
	n := info.Counters[game.CounterMinusOne]
	target, ok := ctx.ClauseTarget(0)
	if n <= 0 || !ok || target.Kind != game.TargetCard {
		return nil
	}
	return AddCounter{Target: target.ID, Kind: game.CounterMinusOne, N: n}.Apply(ctx)
}
