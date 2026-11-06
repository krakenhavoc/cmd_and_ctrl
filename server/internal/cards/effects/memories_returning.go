package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Memories Returning — Sorcery {2}{U}{U} (Final Fantasy, #4462):
//
//	"Reveal the top five cards of your library. Put one of them into your
//	 hand. Then choose an opponent. They put one on the bottom of your
//	 library. Then you put one into your hand. Then they put one on the
//	 bottom of your library. Put the other into your hand.
//	 Flashback {7}{U}{U}"
//
// Fact or Fiction's reveal, then five alternating steps, each its own
// prompt and each waiting on the one before:
//
//  1. you pick one of the revealed cards still in your library and it
//     goes to your hand (a choose_cards pick, then the keyed
//     library-to-hand move, whose continuation is the next step);
//  2. you choose an opponent (asked only when there is more than one);
//  3. they pick one of the rest (a reveal_pick: the reveal is what lets
//     them see cards in your library) and it goes to the bottom;
//  4. you pick one of the rest for your hand;
//  5. they pick one of the rest for the bottom, and the last card goes
//     to your hand.
//
// The chosen opponent rides each later prompt and tail as data (the
// first ID of its Carry), and the step number rides as a number, so a
// table waiting on any of the prompts is a restore point. A library of
// fewer than five cards runs the steps until the revealed cards run
// out. A revealed card that has left the library in the meantime is
// skipped (CR 400.7). An opponent who has left the game picks nothing
// and the steps go on without them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "6cdc6548-cb32-437f-850a-9002542ad3d5",
		Name:             memoriesReturningName,
		Purpose:          game.Purpose{Answers: game.AnswerValue},
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{7}{U}{U}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			revealed := ctx.Game.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), 5,
				"Memories Returning — reveal the top five cards of your library")
			memoriesReturningAskYou(ctx.Game, ctx.Controller(), ctx.Source(), uuid.Nil, revealed, 1)
			return nil
		},
	})
}

const memoriesReturningName = "Memories Returning"

// The registered continuations. They are assigned in init, not at
// package level, because the steps form a loop (each one names the next,
// and step 4 is step 1 again), which Go would see as an initialisation
// cycle. Every key is an on-disk identity: never renamed, never reused.
var (
	// memoriesReturningTake is your pick: it goes to your hand, and the
	// next step waits on the move.
	memoriesReturningTake game.ChooseCardsThen
	// memoriesReturningTaken runs once your pick has reached your hand:
	// after step 1 you choose an opponent, after step 4 they pick again.
	memoriesReturningTaken game.TailRef
	// memoriesReturningOpponent is step 2's answer: the chosen opponent
	// picks next. With nobody to choose the rest stay where they are.
	memoriesReturningOpponent game.OptionPickThen
	// memoriesReturningBottom is the opponent's pick: it goes to the
	// bottom of your library, and the next step waits on the move.
	memoriesReturningBottom game.ChooseCardsThen
	// memoriesReturningBottomed runs once the opponent's pick is on the
	// bottom.
	memoriesReturningBottomed game.TailRef
)

func init() {
	memoriesReturningTake = game.RegisterChooseCardsThen("choose-cards/memories-returning-take", func(g *game.Game, r game.CardsChosen) error {
		stage := memoriesReturningStage(r.Nums)
		carry := r.Carry
		opp := uuid.Nil
		if stage > 1 && len(carry) > 0 {
			opp, carry = carry[0], carry[1:]
		}
		rest := memoriesReturningWithout(carry, r.Picked)
		return g.TakeFromLibraryToHandKeyedForEffect(r.Chooser, r.Picked, game.KeyedTail{
			Ref:    memoriesReturningTaken,
			Player: r.Chooser,
			Source: r.Source,
			N:      stage,
			Carry:  append([]uuid.UUID{opp}, rest...),
		})
	})

	memoriesReturningTaken = CardTail("then/memories-returning-taken", func(ctx *Context, o game.TailOutcome) error {
		if len(o.Tail.Carry) == 0 {
			return nil
		}
		opp, rest := o.Tail.Carry[0], cardsStillInALibrary(ctx.Game, o.Tail.Carry[1:])
		if len(rest) == 0 {
			return nil
		}
		if o.Tail.N > 1 {
			return memoriesReturningAskThem(ctx.Game, o.Tail.Player, o.Tail.Source, opp, rest, o.Tail.N)
		}
		return ChoosePlayer{
			Among:    Opponents,
			Question: memoriesReturningName + " — choose an opponent; they put revealed cards on the bottom of your library",
			ThenKey:  memoriesReturningOpponent,
			Carry:    rest,
		}.Apply(ctx)
	})

	memoriesReturningOpponent = ChoosePlayerThen("option-pick/memories-returning-opponent", func(ctx *Context, chosen uuid.UUID, rest []uuid.UUID) error {
		if chosen == uuid.Nil {
			return nil
		}
		return memoriesReturningAskThem(ctx.Game, ctx.Controller(), ctx.Source(), chosen, rest, 1)
	})

	memoriesReturningBottom = game.RegisterChooseCardsThen("choose-cards/memories-returning-bottom", func(g *game.Game, r game.CardsChosen) error {
		stage := memoriesReturningStage(r.Nums)
		picked := cardsStillInALibrary(g, r.Picked)
		rest := memoriesReturningWithout(r.Candidates, picked)
		if len(picked) == 0 {
			return memoriesReturningAfterBottom(g, r.From, r.Source, r.Chooser, rest, stage)
		}
		return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
			Chooser:   r.From,
			Source:    r.Source,
			Cards:     picked,
			From:      game.ZoneLibrary,
			Placement: game.LibraryPlaceBottom,
			Reason:    memoriesReturningName + " — put it on the bottom of your library",
			ThenTail: game.KeyedTail{
				Ref:    memoriesReturningBottomed,
				Player: r.From,
				Source: r.Source,
				N:      stage,
				Carry:  append([]uuid.UUID{r.Chooser}, rest...),
			},
		})
	})

	memoriesReturningBottomed = game.RegisterTail("then/memories-returning-bottomed", func(g *game.Game, o game.TailOutcome) error {
		if len(o.Tail.Carry) == 0 {
			return nil
		}
		return memoriesReturningAfterBottom(g, o.Tail.Player, o.Tail.Source, o.Tail.Carry[0], o.Tail.Carry[1:], o.Tail.N)
	})
}

// memoriesReturningAskYou is steps 1 and 4: you put one of `rest` into
// your hand. `opp` is the chosen opponent (uuid.Nil before step 2).
func memoriesReturningAskYou(g *game.Game, you, source, opp uuid.UUID, rest []uuid.UUID, stage int) {
	rest = cardsStillInALibrary(g, rest)
	if len(rest) == 0 {
		return
	}
	carry := rest
	if stage > 1 {
		carry = append([]uuid.UUID{opp}, rest...)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:   you,
		Source:    source,
		Question:  memoriesReturningName + " — put one of the revealed cards into your hand",
		Cards:     rest,
		Min:       1,
		Max:       1,
		LiveZone:  game.ZoneLibrary,
		ThenKey:   memoriesReturningTake,
		Carry:     carry,
		CarryNums: []int{stage},
	})
}

// memoriesReturningAskThem is steps 3 and 5: the opponent puts one of
// `rest` on the bottom of your library. An opponent who can't be asked
// (they have left the game) picks nothing.
func memoriesReturningAskThem(g *game.Game, you, source, opp uuid.UUID, rest []uuid.UUID, stage int) error {
	rest = cardsStillInALibrary(g, rest)
	if len(rest) == 0 {
		return nil
	}
	n, err := g.RevealPickThenForEffect(game.RevealPickPrompt{
		Chooser:  opp,
		Owner:    you,
		Source:   source,
		Question: memoriesReturningName + " — put one of these cards on the bottom of their library",
		Cards:    rest,
		Min:      1,
		Max:      1,
		ThenKey:  memoriesReturningBottom,
		Nums:     []int{stage},
	})
	if err != nil || n > 0 {
		return err
	}
	return memoriesReturningAfterBottom(g, you, source, opp, rest, stage)
}

// memoriesReturningAfterBottom is what follows an opponent's step: after
// step 3 you pick again (step 4); after step 5 "the other" goes to your
// hand.
func memoriesReturningAfterBottom(g *game.Game, you, source, opp uuid.UUID, rest []uuid.UUID, stage int) error {
	rest = cardsStillInALibrary(g, rest)
	if len(rest) == 0 {
		return nil
	}
	if stage > 1 {
		return g.TakeFromLibraryToHandKeyedForEffect(you, rest, game.KeyedTail{})
	}
	memoriesReturningAskYou(g, you, source, opp, rest, stage+1)
	return nil
}

// memoriesReturningStage reads the step number a prompt carried.
func memoriesReturningStage(nums []int) int {
	if len(nums) == 0 {
		return 1
	}
	return nums[0]
}

// memoriesReturningWithout is `ids` less `picked`, in order.
func memoriesReturningWithout(ids, picked []uuid.UUID) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if !slices.Contains(picked, id) {
			out = append(out, id)
		}
	}
	return out
}
