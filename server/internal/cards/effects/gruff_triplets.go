package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gruff Triplets — Creature — Satyr Warrior {3}{G}{G}{G}, 3/3 (Wilds of
// Eldraine, #4471):
//
//	"Trample
//	 When this creature enters, if it isn't a token, create two tokens that
//	 are copies of it.
//	 When this creature dies, put a number of +1/+1 counters equal to its
//	 power on each creature you control named Gruff Triplets."
//
// The entry is an intervening "if" (CR 603.4) checked as it enters and again
// as it resolves (condThisIsNotAToken): a token copy never makes more, so
// the three stop there. The copies are CreateTokenCopy over the Triplets as
// they last existed on the battlefield. The death trigger reads the power it
// last had (the departure record, counters and pumps included, CR 608.2h)
// and puts that many counters on each creature you control then named Gruff
// Triplets (the name read through the layers, so a copy named so counts, a
// Triplets that is not named so does not), as one placement per creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "32de1b70-f993-4423-b0ef-0762c2b4d1b1",
		Name:            "Gruff Triplets",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(condThisIsNotAToken{}, WhenThisEnters("Gruff Triplets — if it isn't a token, create two tokens that are copies of it",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 2}.Apply(NewContext(g, item))
				})),
			WhenThisDies("Gruff Triplets — put a number of +1/+1 counters equal to its power on each creature you control named Gruff Triplets",
				gruffTripletsDies),
		},
	})
}

// condThisIsNotAToken is "if it isn't a token", read of the ability's own
// source as the trigger resolves.
type condThisIsNotAToken struct{}

func (condThisIsNotAToken) Holds(ctx *Context) bool {
	c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	return ok && !c.IsToken()
}

func (condThisIsNotAToken) condition() (game.FootprintStep, bool) { return game.FootprintStep{}, false }

func gruffTripletsDies(g *game.Game, item *game.StackItem) error {
	info, ok := g.LastKnownPermanentForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	n := max(0, info.Power)
	if n == 0 {
		return nil
	}
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && game.HasName(c, "Gruff Triplets") {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
