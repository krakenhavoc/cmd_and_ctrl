package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nick Fury, Agent of S.H.I.E.L.D. — Legendary Creature — Human Spy Hero {W},
// 2/1 (Marvel Super Heroes, #4455):
//
//	"Power-up — {W}{U}{B}{R}{G}: Put two +1/+1 counters on Nick Fury, then
//	 look at the top seven cards of your library. You may put a Hero,
//	 Equipment, or Vehicle card from among them onto the battlefield. If
//	 it's a double-faced card, you may transform it. Put the rest on the
//	 bottom of your library in a random order. (Activate each power-up
//	 ability only once. Reduce the cost by his mana cost if he entered this
//	 turn.)"
//
// Power-up is the engine's (CR 702.193). The counters go on first, then
// Ao, the Dawn Sky's look at seven (PutFromLibraryOntoBattlefield, here one
// optional card of the three kinds, a nonland permanent by card type or
// not: a Hero creature, an Equipment, a Vehicle). The clause after the put
// is the tail's: if the card that arrived is a double-faced card that can
// transform, its controller is asked, and the answer, either way, is
// followed by the rest going to the bottom in a random order, so the rest
// are placed after the transform as printed. Declining the put, or a pick
// that did not arrive, goes straight to the bottom.
//
// No simplification.
func init() {
	ab := PowerUp("{W}{U}{B}{R}{G}: Put two +1/+1 counters on Nick Fury, then look at the top seven cards of your library. You may put a Hero, Equipment, or Vehicle card from among them onto the battlefield. If it's a double-faced card, you may transform it. Put the rest on the bottom of your library in a random order.",
		ManaCost("{W}{U}{B}{R}{G}"), func(g *game.Game, item *game.StackItem) error {
			if err := plusOneCountersOnThis(2)(g, item); err != nil {
				return err
			}
			ctx := NewContext(g, item)
			player := item.Controller
			return PutFromLibraryOntoBattlefield{
				Player:   player,
				Cards:    g.LookAtTopOfLibraryForEffect(player, 7),
				Match:    Or(Subtype("Hero"), Subtype("Equipment"), Subtype("Vehicle")),
				Max:      1,
				Optional: true,
				Label:    "Nick Fury — you may put a Hero, Equipment, or Vehicle card from among them onto the battlefield",
				ThenTail: mshA6NickFuryAfterPut,
			}.Apply(ctx)
		})
	ab.Purpose.Answers = game.AnswerPump
	Register(Spec{
		OracleID:     "62916107-14c0-4516-9fed-a032b1a80e5d",
		Name:         "Nick Fury, Agent of S.H.I.E.L.D.",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{ab},
	})
}

// mshA6NickFuryAfterPut is the clause after the put: ask about transforming
// a double-faced card that arrived, then put the rest on the bottom.
var mshA6NickFuryAfterPut = PutFromLibraryTail("msh-a6/nick-fury-after-put", func(ctx *Context, res PutFromLibraryResult) error {
	if len(res.Entered) == 1 && ctx.Game.CanTransformForEffect(res.Entered[0]) {
		name := "it"
		if c, ok := ctx.Game.LookupCardForEffect(res.Entered[0]); ok {
			name = c.Name
		}
		return MayChoice{
			Player:   res.Player,
			Question: "Nick Fury — transform " + name + "?",
			Then:     mshA6NickFuryTransformAnswer,
			Carry:    append([]uuid.UUID{res.Entered[0]}, res.Rest...),
		}.Apply(ctx)
	}
	return PutRestOnBottomInRandomOrder(ctx.Game, res)
})

// mshA6NickFuryTransformAnswer is the answer to the transform question,
// either way followed by the rest (Carry after the first ID) going to the
// bottom of the library in a random order. An on-disk key.
var mshA6NickFuryTransformAnswer = MayAnswer("confirm/msh-a6-nick-fury-transform", func(ctx *Context, r game.ConfirmAnswer) error {
	if r.Accepted {
		if err := (Transform{Target: carryAt(r, 0)}).Apply(ctx); err != nil {
			return err
		}
	}
	if len(r.Carry) < 2 {
		return nil
	}
	return PutRestOnBottomInRandomOrder(ctx.Game, PutFromLibraryResult{Player: r.Chooser, Rest: cardsStillInALibrary(ctx.Game, r.Carry[1:])})
})
