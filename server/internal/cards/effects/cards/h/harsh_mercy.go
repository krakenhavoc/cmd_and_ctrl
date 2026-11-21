package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Harsh Mercy — Sorcery {2}{W} (Commander Masters, #5948):
//
//	"Each player chooses a creature type. Destroy all creatures that aren't
//	 of a type chosen this way. They can't be regenerated."
//
// Patriarch's Bidding's chain of resolution-time prompts: one per living
// seat in turn order from the caster, each seeing the earlier answers (CR
// 101.4), the chosen types riding the continuation by value. When the last
// seat has answered, every creature with none of the chosen types is
// destroyed as one sweep that ignores regeneration (CR 701.19c); a
// changeling has every type, so it always survives a choice. A seat that
// leaves with its prompt open chooses nothing and the chain carries on.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dcd65c9b-4aaa-42af-869f-de179ae57c9f",
		Name:         "Harsh Mercy",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true},
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return harshMercyAsk(ctx, effects.SeatsFromController(ctx.Game, item.Controller), nil)
		},
	})
}

// harshMercyAsk asks seats[0], then carries on with the rest and the answer
// appended. The seats still to ask ride the prompt's tail as its Carry and
// the answers so far as its Words, so a table waiting on any seat is a
// restore point.
func harshMercyAsk(ctx *effects.Context, seats []uuid.UUID, chosen []string) error {
	if len(seats) == 0 {
		return harshMercyDestroy(ctx, chosen)
	}
	then := ctx.ItemTail(harshMercyTypeChosen)
	then.Carry = append([]uuid.UUID(nil), seats[1:]...)
	then.Words = append([]string(nil), chosen...)
	if !effects.ChooseCreatureTypeTail(ctx.Game, seats[0], ctx.Source(), "Harsh Mercy — choose a creature type", then) {
		return harshMercyAsk(ctx, seats[1:], chosen)
	}
	return nil
}

// harshMercyTypeChosen records one seat's type (the tail's Word, "" when
// that seat left) and asks the next. An on-disk key, registered in init
// because its body asks the next seat with it.
var harshMercyTypeChosen game.TailRef

func init() {
	harshMercyTypeChosen = effects.CardTail("then/harsh-mercy-type-chosen", func(ctx *effects.Context, o game.TailOutcome) error {
		chosen := append([]string(nil), o.Tail.Words...)
		if o.Tail.Word != "" {
			chosen = append(chosen, o.Tail.Word)
		}
		return harshMercyAsk(ctx, o.Tail.Carry, chosen)
	})
}

// harshMercyDestroy destroys every creature that has none of the chosen
// types, regeneration shields ignored. No type chosen (every seat left)
// leaves every creature "not of a chosen type", as printed.
func harshMercyDestroy(ctx *effects.Context, chosen []string) error {
	return effects.DestroyAllMatching{
		Match: func(g *game.Game, you uuid.UUID, c game.Card) bool {
			if !c.IsCreature() {
				return false
			}
			for _, t := range chosen {
				if c.HasSubtype(t) {
					return false
				}
			}
			return true
		},
		CantBeRegenerated: true,
	}.Apply(ctx)
}
