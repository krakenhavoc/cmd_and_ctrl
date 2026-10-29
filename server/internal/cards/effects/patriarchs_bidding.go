package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Patriarch's Bidding — Sorcery {3}{B}{B}:
//
//	"Each player chooses a creature type. Each player returns all
//	 creature cards of a type chosen this way from their graveyard to
//	 the battlefield."
//
// "Each player chooses" is a chain of resolution-time prompts (#2382),
// one per living seat in turn order from the caster, each addressed to
// that seat. The chosen types ride the continuation BY VALUE. When the
// last seat has answered, every player returns every creature card of
// any chosen type from their own graveyard, as one simultaneous entry
// under its owner's control. A seat that leaves with its prompt open
// chooses nothing and the chain carries on.
//
// The prompts are asked one at a time, caster first, so a later seat
// sees the earlier answers — which is the rules' own order for an
// "each player chooses" (CR 101.4: APNAP, with earlier choices known).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "25fc3bc2-d852-4f52-9adb-c7e35c06f3af",
		Name:         "Patriarch's Bidding",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return patriarchsBiddingAsk(ctx, seatsFromController(ctx.Game, item.Controller), nil)
		},
	})
}

// patriarchsBiddingAsk asks seats[0], then carries on with the rest and
// the answer appended. A seat that cannot be asked (queue refused) is
// skipped without an answer. The seats still to ask ride the prompt's tail
// as its Carry and the answers so far as its Words (#4253), so a table
// waiting on any seat is a restore point.
func patriarchsBiddingAsk(ctx *Context, seats []uuid.UUID, chosen []string) error {
	if len(seats) == 0 {
		return patriarchsBiddingReturn(ctx, chosen)
	}
	then := ctx.ItemTail(patriarchsBiddingTypeChosen)
	then.Carry = append([]uuid.UUID(nil), seats[1:]...)
	then.Words = append([]string(nil), chosen...)
	if !ChooseCreatureTypeTail(ctx.Game, seats[0], ctx.Source(), "Patriarch's Bidding — choose a creature type", then) {
		return patriarchsBiddingAsk(ctx, seats[1:], chosen)
	}
	return nil
}

// patriarchsBiddingTypeChosen records one seat's type (the tail's Word, ""
// when that seat left) and asks the next. An on-disk key, registered in
// init because its body asks the next seat with it.
var patriarchsBiddingTypeChosen game.TailRef

func init() {
	patriarchsBiddingTypeChosen = CardTail("then/patriarchs-bidding-type-chosen", func(ctx *Context, o game.TailOutcome) error {
		chosen := append([]string(nil), o.Tail.Words...)
		if o.Tail.Word != "" {
			chosen = append(chosen, o.Tail.Word)
		}
		return patriarchsBiddingAsk(ctx, o.Tail.Carry, chosen)
	})
}

func patriarchsBiddingReturn(ctx *Context, chosen []string) error {
	if len(chosen) == 0 {
		return nil
	}
	ids := allGraveyardsCardIDs(ctx, func(c game.Card) bool {
		if !c.IsCreature() {
			return false
		}
		for _, t := range chosen {
			if c.HasSubtype(t) {
				return true
			}
		}
		return false
	})
	// #1997: "each player returns" — each card is put by the player it
	// enters under (CR 110.2a), which Arboria reads.
	return ctx.Game.PutByEnteringPlayerForEffect(func() error {
		return ReturnFromGraveyardTogether{Targets: ids}.Apply(ctx)
	})
}
