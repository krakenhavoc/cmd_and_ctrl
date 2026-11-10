package game

import "github.com/google/uuid"

// counter_keyed.go — "counter <a spell or ability>" with a keyed
// continuation told whether it WAS countered (CR 701.6a), for a card that
// counts what it countered: Glen Elendra's Answer's "Create a 1/1 blue and
// black Faerie creature token with flying for each spell and ability
// countered this way" (Lorwyn Eclipsed, #4458).
//
// "Countered this way" is not "left the stack". A spell that can't be
// countered stays; an ability under "abilities can't be countered" stays;
// a counter a replacement turned into an exile (Guile, ExileCounteredSpell-
// InsteadForEffect) moves the spell but does not counter it; and a spell
// that was countered may land somewhere other than its graveyard (a
// flashback spell's exile, a commander's command zone, Rest in Peace). The
// one mark every real counter leaves is its EventCounterSpell (the route's
// Countered move for a spell, counterAbilityLocked for an ability), so the
// continuation is told the object exactly when one naming it was emitted
// after the counter began. A counter paused on a CR 616.1 ordering prompt
// tells its continuation once the prompt is answered: the tail rides the
// counter's route as data, so the pause is a restore point.

// CounterKeyedForEffect counters the spell or ability `stackID` as
// CounterTargetForEffect does, with its gates ("can't be countered" leaves
// it on the stack), and then runs `then`, told the object as
// TailOutcome.IDs when it was countered and no IDs otherwise (it can't be
// countered, it already left the stack, or a replacement made the counter
// something else). A zero tail is "no continuation".
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) CounterKeyedForEffect(stackID uuid.UUID, then KeyedTail) error {
	item, ok := g.StackMeta[stackID]
	if !ok || item == nil {
		return g.runKeyedTailLocked(then, nil, 0)
	}
	mark := g.eventSeq
	switch item.Kind {
	case StackItemActivated, StackItemTriggered:
		if g.abilityCantBeCounteredLocked(item) {
			return g.runKeyedTailLocked(then, nil, 0)
		}
		if err := g.counterAbilityLocked(stackID); err != nil {
			return err
		}
		return g.runKeyedTailLocked(then, g.counteredSinceLocked(stackID, mark), 0)
	case StackItemSpell:
		if !g.counterableSpellOnStackLocked(stackID) {
			return g.runKeyedTailLocked(then, nil, 0)
		}
		var told KeyedTail
		if !then.IsZero() {
			next := copyKeyedTail(then)
			told = KeyedTail{Ref: counteredTold, Carry: []uuid.UUID{stackID}, Nums: []int{int(mark)}, Next: &next}
		}
		return g.counterSpellByEffectLocked(stackID, nil, TuckOptions{}, told)
	default:
		return g.runKeyedTailLocked(then, nil, 0)
	}
}

// counteredSinceLocked is [id] when an EventCounterSpell naming id was
// emitted after the event sequence number `mark`, and nil otherwise.
//
// Caller must hold g.mu.
func (g *Game) counteredSinceLocked(id uuid.UUID, mark uint64) []uuid.UUID {
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Seq <= mark {
			break
		}
		if ev.Kind == EventCounterSpell && ev.Target == id {
			return []uuid.UUID{id}
		}
	}
	return nil
}

// counteredTold is CounterKeyedForEffect's filter on a spell counter's
// route: Carry[0] is the spell, Nums[0] the event mark the counter began
// at, and Next the caller's tail, told the spell only when it was
// countered. An on-disk key: never renamed, never reused.
var counteredTold = RegisterTail("engine/countered-told", func(g *Game, o TailOutcome) error {
	t := o.Tail
	if len(t.Carry) == 0 || len(t.Nums) == 0 {
		return g.runKeyedTailLocked(nextTailValue(t.Next), nil, 0)
	}
	return g.runKeyedTailLocked(nextTailValue(t.Next), g.counteredSinceLocked(t.Carry[0], uint64(t.Nums[0])), 0)
})
