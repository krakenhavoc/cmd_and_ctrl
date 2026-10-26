package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Círdan the Shipwright — Legendary Creature — Elf Noble {3}{G}{U}, 3/4:
//
//	"Vigilance
//	 Secret council — Whenever Círdan enters or attacks, each player
//	 secretly votes for a player, then those votes are revealed. Each
//	 player draws a card for each vote they received. Each player who
//	 received no votes may put a permanent card from their hand onto
//	 the battlefield."
//
// A secret vote over the players (ADR 0180); a player may vote for
// themselves. Once the votes are revealed, each player draws a card per
// vote they received, in turn order from the active player. Then each
// player who received none chooses, in the same order, whether to put a
// permanent card from their hand (a hidden choice, asked of each in
// turn), and every chosen card enters at once, each under its owner's
// control (CR 101.4: the choices first, then one simultaneous put).
//
// A bot votes for itself (Vote.SelfBest): the votes a player receives
// draw them cards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "063e05b0-554f-4b55-9feb-e11f3927f3a6",
		Name:            "Círdan the Shipwright",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Círdan the Shipwright — each player secretly votes for a player",
				func(g *game.Game, item *game.StackItem) error {
					return Vote{
						Question: "Círdan the Shipwright — secretly vote for a player",
						Options:  VoteForPlayers(g),
						Secret:   true,
						SelfBest: true,
						Then:     cirdanVoted,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

var cirdanVoted = VoteResultThen("vote/cirdan-the-shipwright", func(ctx *Context, r game.VoteResult) error {
	received := VotesReceived(r)
	var voted, unvoted []uuid.UUID
	var counts []int
	for _, p := range apnapLivingSeats(ctx.Game) {
		if n := received[p]; n > 0 {
			voted = append(voted, p)
			counts = append(counts, n)
			continue
		}
		unvoted = append(unvoted, p)
	}
	ctx.RecordStep(cirdanVotedRecord, voted, counts...)
	ctx.RecordStep(cirdanUnvotedRecord, unvoted)
	return cirdanDrawThenAsk.Run(ctx)
})

// What the vote hands the steps behind it: who received votes and how many,
// and who received none (#4163).
const (
	cirdanVotedRecord   = "cirdan-the-shipwright/voted"
	cirdanUnvotedRecord = "cirdan-the-shipwright/unvoted"
)

// cirdanDrawThenAsk is the draws for the players who received votes and then
// the question to the ones who received none: the question waits for a
// prompt a draw asks (a dredge offer; CR 608.2c, #4155).
var cirdanDrawThenAsk = RegisterThenSteps("then-steps/cirdan-the-shipwright-draws",
	Instruction(func(ctx *Context) error {
		r := ctx.StepRecord(cirdanVotedRecord)
		for i, p := range r.IDs {
			if i >= len(r.Ns) {
				break
			}
			if err := ctx.Game.DrawNForEffect(p, r.Ns[i]); err != nil {
				return err
			}
		}
		return nil
	}),
	Instruction(func(ctx *Context) error {
		return cirdanAskToPut(ctx.Game, ctx.Source(), ctx.StepRecord(cirdanUnvotedRecord).IDs, nil)
	}),
)

// cirdanPut is the keyed answer to one player's "put a permanent card from
// your hand onto the battlefield" (#4253): the players still to ask and the
// cards chosen so far ride the prompt as its Carry, and the next player is
// asked from here. An on-disk key: never renamed, never reused.
//
// Registered in init, not at package level: the answer asks the next
// player, so it refers to itself.
var cirdanPut game.ChooseCardsThen

func init() {
	cirdanPut = game.RegisterChooseCardsThen("choose-cards/cirdan-the-shipwright-put", func(g *game.Game, r game.CardsChosen) error {
		rest, so := unpackIDPair(r.Carry)
		return cirdanAskToPut(g, r.Source, rest, append(so, r.Picked...))
	})
}

// cirdanAskToPut asks the first player in `players` which permanent card
// from their hand they will put onto the battlefield, if any, and the
// rest after them; once everyone has answered, every chosen card enters
// at once. Captures only values.
func cirdanAskToPut(g *game.Game, source uuid.UUID, players, chosen []uuid.UUID) error {
	for i, p := range players {
		if !voterStillIn(g, p) {
			continue
		}
		cards := handCardsMatching(g, p, nil)
		if len(cards) == 0 {
			continue
		}
		rest := append([]uuid.UUID(nil), players[i+1:]...)
		so := append([]uuid.UUID(nil), chosen...)
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Destination: game.ChooseOntoBattlefield,
			Chooser:     p,
			Source:      source,
			Question:    "Círdan the Shipwright — you received no votes: you may put a permanent card from your hand onto the battlefield",
			Cards:       cards,
			Min:         0,
			Max:         1,
			ThenKey:     cirdanPut,
			LiveZone:    game.ZoneHand,
			Carry:       packIDLists(rest, so),
		})
		return nil
	}
	// A batch with a card that is no longer in a hand is refused whole,
	// so anything that moved meanwhile is left out.
	var still []uuid.UUID
	for _, id := range chosen {
		if c, ok := g.LookupCardForEffect(id); ok && slices.Contains(handCardsMatching(g, c.Owner, nil), id) {
			still = append(still, id)
		}
	}
	if len(still) == 0 {
		return nil
	}
	return g.PutCardsFromHandOntoBattlefieldKeyedForEffect(still, game.HandEntryOptions{}, game.KeyedTail{})
}
