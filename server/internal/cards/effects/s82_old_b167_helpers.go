package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b167_helpers.go — shared vocabulary of War of the Spark batch 167
// (S82 older sets, #5520).

// OldB167SourceDealsItsLastKnownPower is "it deals damage equal to its power to
// <the target>" for a dies trigger whose item carries the creature's last-known
// power (CR 603.10, 608.2h) in Params.Amount (EerieDiesWithPower): the damage is
// dealt by the creature as it last existed, to the first target still legal.
func OldB167SourceDealsItsLastKnownPower(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	t, ok := FirstLegalTarget(ctx)
	if !ok {
		return nil
	}
	return DealDamage{SourceObject: &ref, Target: t.ID, Amount: item.Params.Amount}.Apply(ctx)
}

// OldB167LoyaltyCounterOnTheTarget is "put a loyalty counter on target
// <planeswalker>": one loyalty counter on the first target still legal, through
// the counter pipeline.
func OldB167LoyaltyCounterOnTheTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterLoyalty, N: 1}.Apply(ctx)
}

// OldB167ReturnExiledCreatureWithCounterBody is Teferi's Time Twist's delayed
// return: the listed cards come back together under their owners' control, and
// each that enters as a creature enters with an additional +1/+1 counter on it,
// judged card by card as each would exist on the battlefield (CR 614.12). A
// noncreature permanent comes back with nothing extra. An on-disk key: never
// renamed, never reused.
var OldB167ReturnExiledCreatureWithCounterBody = game.SimpleDelayedBody(
	"flicker/s82-old-b167-return-exiled-creature-with-counter", func(g *game.Game, item *game.StackItem) error {
		var ids []uuid.UUID
		for _, t := range item.Targets {
			if t.Kind == game.TargetCard && t.ID != uuid.Nil {
				ids = append(ids, t.ID)
			}
		}
		return ReturnFromExileTogether{
			Targets:        ids,
			CountersIfType: []game.EntryCountersIfType{{CardType: "creature", Counter: game.CounterPlusOne, N: 1}},
		}.Apply(NewContext(g, item))
	})
