package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Collective Voyage — Sorcery {G} (Secret Lair Drop):
//
//	"Join forces — Starting with you, each player may pay any amount of mana.
//	 Each player searches their library for up to X basic land cards, where X is
//	 the total amount of mana paid this way, puts them onto the battlefield
//	 tapped, then shuffles."
//
// Minds Aglow's join forces (MayPayX over {1}, asked of each living seat in turn
// order starting with the caster, each paying from their own mana, CR 608.2d), with
// the running total riding the prompt as data: the seats still to ask, one nil
// marker, then one placeholder per mana paid. When the last has answered, each
// player in turn order from the caster searches their own library for up to the
// total in basic land cards, puts them onto the battlefield tapped and shuffles
// (CR 701.24). A total of zero is still a search that finds nothing and a shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00f82f84-81a5-45e5-9c35-9d627a180950",
		Name:         "Collective Voyage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1, Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return collectiveVoyageAsk(ctx, effects.SeatsFromController(ctx.Game, item.Controller), 0)
		},
	})
}

// collectiveVoyageAsk asks seats[0] how much to pay, then carries on with the rest
// and the new total. With nobody left to ask, everyone searches.
func collectiveVoyageAsk(ctx *effects.Context, seats []uuid.UUID, total int) error {
	if len(seats) == 0 {
		return collectiveVoyageSearch(ctx, total)
	}
	carry := append([]uuid.UUID(nil), seats[1:]...)
	carry = append(carry, uuid.Nil) // the marker between the seats and the total
	for i := 0; i < total; i++ {
		carry = append(carry, uuid.Nil)
	}
	return effects.MayPayX{
		Chooser: seats[0],
		Symbol:  "{1}",
		Label:   "Collective Voyage",
		Buys:    "each player searches for up to X basic lands",
		Unit:    game.PayAmountOther,
		Then:    collectiveVoyagePaid,
		Carry:   carry,
	}.Apply(ctx)
}

// collectiveVoyagePaid is the join-forces step once a seat has settled its amount:
// the pay branch adds X, the decline branch adds nothing, and either asks the next
// seat. An on-disk key. Registered in init because its branches ask the next seat.
var collectiveVoyagePaid game.PayUnlessElse

func init() {
	collectiveVoyagePaid = effects.PayUnlessBranches("pay-unless/collective-voyage-paid",
		func(ctx *effects.Context, r game.PayUnlessAnswer) error {
			return collectiveVoyageNext(ctx, r, r.Number)
		},
		func(ctx *effects.Context, r game.PayUnlessAnswer) error { return collectiveVoyageNext(ctx, r, 0) })
}

// collectiveVoyageNext reads the seats still to ask and the total so far off the
// carry and asks the next seat with `paid` added.
func collectiveVoyageNext(ctx *effects.Context, r game.PayUnlessAnswer, paid int) error {
	var seats []uuid.UUID
	total := 0
	pastMarker := false
	for _, id := range r.Carry {
		switch {
		case id != uuid.Nil && !pastMarker:
			seats = append(seats, id)
		case id == uuid.Nil && !pastMarker:
			pastMarker = true
		default:
			total++
		}
	}
	return collectiveVoyageAsk(ctx, seats, total+max(paid, 0))
}

// collectiveVoyageSearch has each player search for up to `total` basic lands, in
// turn order from the caster.
func collectiveVoyageSearch(ctx *effects.Context, total int) error {
	for _, id := range effects.SeatsFromController(ctx.Game, ctx.Controller()) {
		if total <= 0 {
			if err := (effects.ShuffleLibrary{Player: id}).Apply(ctx); err != nil {
				return err
			}
			continue
		}
		if err := (effects.SearchLibrary{
			Player:        id,
			Predicate:     effects.IsBasicLand,
			Dest:          game.ZoneBattlefield,
			Limit:         total,
			TappedOnEntry: true,
			Shuffle:       true,
			Reason:        "Collective Voyage — up to X basic land cards",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
