package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// MoveAPlusOneCounterFromFirstToSecond is "Move a +1/+1 counter from target
// creature onto another target creature" for the Guildmage-style two-clause
// statement (Combine Guildmage): both clauses must still be legal and the
// first creature must still hold a counter (CR 122.5). The counter lands on
// the second creature first, through the ordinary counter window, and comes
// off the first only if it did.
func MoveAPlusOneCounterFromFirstToSecond(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	from, okFrom := ctx.ClauseTarget(0)
	to, okTo := ctx.ClauseTarget(1)
	if !okFrom || !okTo || from.Kind != game.TargetCard || to.Kind != game.TargetCard {
		return nil
	}
	src, ok := g.LookupCardForEffect(from.ID)
	if !ok || src.Counters[game.CounterPlusOne] <= 0 {
		return nil
	}
	then := ctx.Tail(s82OldB177TakeOne)
	then.Carry = []uuid.UUID{from.ID}
	return g.AddCounterByKeyedForEffect(item.Controller, to.ID, game.CounterPlusOne, 1, then)
}

// s82OldB177TakeOne is "remove a +1/+1 counter from the first creature", if
// the one onto the second landed. An on-disk key: never renamed, never reused.
var s82OldB177TakeOne = game.RegisterTail("combine-guildmage/take-one", func(g *game.Game, o game.TailOutcome) error {
	if o.Amount <= 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	if c, ok := g.LookupCardForEffect(o.Tail.Carry[0]); !ok || c.Counters[game.CounterPlusOne] <= 0 {
		return nil
	}
	return g.AddCounterForEffect(o.Tail.Carry[0], game.CounterPlusOne, -1)
})

// PlusOneCountersOnCreaturesYouControl is the number of +1/+1 counters on the
// creatures `player` controls right now (Rumbling Ruin: "count the number of
// +1/+1 counters on creatures you control").
func PlusOneCountersOnCreaturesYouControl(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			n += c.Counters[game.CounterPlusOne]
		}
	}
	return n
}

// AttackDefenderDiscardsACard is "Whenever <this> attacks, defending player
// discards a card" as a trigger's effect (Shrieking Specter, The Haunt of
// Hightower): the defending player (the one attacked, or the controller of the
// planeswalker or battle attacked) discards by their own choice.
func AttackDefenderDiscardsACard(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	defender := DefendingPlayerOfAttack(g, item.Trigger.Event)
	if defender == uuid.Nil {
		return nil
	}
	return DefendingPlayerDiscardsACard(g, item, defender)
}
