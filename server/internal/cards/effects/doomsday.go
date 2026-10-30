package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Doomsday — {B}{B}{B} Sorcery:
//
//	"Search your library and graveyard for five cards and exile the rest.
//	 Put the chosen cards on top of your library in any order. You lose
//	 half your life, rounded up."
//
// One search over two zones, in the printed order:
//
//  1. The controller picks five cards from the whole library and the whole
//     graveyard together (fewer only when there are fewer: CR 701.23d, a
//     search for a quantity must find that many or as many as possible). The
//     library is looked at, so it is the chooser's to see and nobody
//     else's (LookAtTopOfLibraryForEffect marks the knower), and the search
//     is announced (EventSearchLibrary, CR 701.23), which is what "whenever
//     an opponent searches their library" watches. It does not shuffle.
//  2. Every other card in the library and graveyard is exiled, simultaneously
//     (a CR 903.9 commander prompt is a restore point).
//  3. The chosen cards go on top of the library in the order the controller
//     chooses (ADR 0088), from whichever zone each is in.
//  4. The controller loses half their life, rounded up (CR 119.3): a player at
//     a life total of zero or less loses nothing.
//
// The chain is data: the pick, the exile, the ordering prompt and the life
// loss are each a registered key (never renamed, never reused).
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	Register(Spec{
		OracleID:     "721eb5a2-d7cf-4db0-8013-ef3f596c52a5",
		Name:         "Doomsday",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			you := ctx.Controller()
			p := ctx.PlayerByID(you)
			if p == nil || p.Eliminated {
				return nil
			}
			var pool []uuid.UUID
			if p.Library != nil {
				pool = append(pool, ctx.Game.LookAtTopOfLibraryForEffect(you, len(p.Library.Cards))...)
			}
			ctx.Game.EmitEvent(game.Event{Kind: game.EventSearchLibrary, Actor: you, Target: you})
			if p.Graveyard != nil {
				for _, c := range p.Graveyard.Cards {
					pool = append(pool, c.InstanceID)
				}
			}
			if len(pool) == 0 {
				return msc18DoomsdayLoseHalfOfYourLife(ctx, you)
			}
			n := min(5, len(pool))
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  you,
				Source:   ctx.Source(),
				Question: "Doomsday — search your library and graveyard for five cards (the rest are exiled)",
				Cards:    pool,
				Min:      n,
				Max:      n,
				ThenKey:  msc18DoomsdayChosen,
			})
			return nil
		},
	})
}

// msc18DoomsdayChosen is "exile the rest" once the five are picked: every
// other candidate goes to exile in one batch, then the chosen cards are
// placed. An on-disk key.
var msc18DoomsdayChosen = game.RegisterChooseCardsThen("choose-cards/doomsday-search", func(g *game.Game, r game.CardsChosen) error {
	if !r.Answered {
		return nil
	}
	keep := make(map[uuid.UUID]bool, len(r.Picked))
	for _, id := range r.Picked {
		keep[id] = true
	}
	var rest []uuid.UUID
	for _, id := range r.Candidates {
		if !keep[id] {
			rest = append(rest, id)
		}
	}
	place := game.KeyedTail{Ref: msc18DoomsdayPlace, Player: r.Chooser, Source: r.Source, Carry: r.Picked}
	if len(rest) == 0 {
		return place.Run(g, nil)
	}
	return g.ExileCardsKeyedForEffect(rest, place)
})

// msc18DoomsdayPlace puts the chosen cards on top of the library in the
// controller's order, then loses the life. Tail.Carry is the chosen cards. An
// on-disk key.
var msc18DoomsdayPlace = game.RegisterTail("then/doomsday-put-on-top", func(g *game.Game, o game.TailOutcome) error {
	then := game.KeyedTail{Ref: msc18DoomsdayLose, Player: o.Tail.Player, Source: o.Tail.Source}
	if len(o.Tail.Carry) == 0 {
		return then.Run(g, nil)
	}
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   o.Tail.Player,
		Source:    o.Tail.Source,
		Cards:     o.Tail.Carry,
		Placement: game.LibraryPlaceTop,
		Reason:    "Doomsday — put the chosen cards on top of your library in any order",
		ThenTail:  then,
	})
})

// msc18DoomsdayLose is "you lose half your life, rounded up". An on-disk key.
var msc18DoomsdayLose = CardTail("then/doomsday-lose-half-your-life", func(ctx *Context, o game.TailOutcome) error {
	return msc18DoomsdayLoseHalfOfYourLife(ctx, o.Tail.Player)
})

func msc18DoomsdayLoseHalfOfYourLife(ctx *Context, who uuid.UUID) error {
	p := ctx.PlayerByID(who)
	if p == nil || p.Eliminated || p.Life <= 0 {
		return nil
	}
	return GainLife{Player: who, Amount: -((p.Life + 1) / 2)}.Apply(ctx)
}
