package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Seventh Doctor — Legendary Creature — Time Lord Doctor {3}{W}{U}, 3/6
// (Doctor Who, #5948):
//
//	"Whenever The Seventh Doctor attacks, choose a card in your hand.
//	 Defending player guesses whether that card's mana value is greater than
//	 the number of artifacts you control. If they guessed wrong, you may cast
//	 it without paying its mana cost. If you don't cast a spell this way,
//	 investigate."
//
// The attack trigger asks the controller for a card in their hand, then asks
// the defending player to guess "greater" or "not greater" (the question
// names the artifact count as it is when asked; the card stays hidden). The
// answer is judged against the card's mana value (X is 0 in a hand, CR
// 107.3b) and the number of artifacts the controller controls as the guess
// is revealed. A wrong guess offers the free cast from the hand as part of
// the resolution (CR 608.2g, the shared resolution-cast offer); a right
// guess, a declined or impossible cast (a land, no legal target) and an empty
// hand all investigate. Only the defending player guesses; in a multiplayer
// game it is the player the Doctor attacked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "43e1b053-2cb9-4b1d-9e38-100cfa7267b6",
		Name:         "The Seventh Doctor",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("The Seventh Doctor — choose a card in your hand; defending player guesses whether its mana value is greater than the number of artifacts you control; if wrong you may cast it free; if you don't cast a spell this way, investigate",
				seventhDoctorAttack),
		},
	})
}

func seventhDoctorAttack(g *game.Game, item *game.StackItem) error {
	var defender uuid.UUID
	if item.Trigger != nil {
		defender = b17DefendingPlayer(g, item.Trigger.Event)
	}
	hand := allHandCardIDs(g, item.Controller)
	if len(hand) == 0 || defender == uuid.Nil {
		return seventhDoctorInvestigate(g, item.Controller, item.SourceCardID)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "The Seventh Doctor — choose a card in your hand",
		Cards:      hand,
		Min:        1,
		Max:        1,
		ThenKey:    seventhDoctorChosen,
		Carry:      []uuid.UUID{defender},
		LiveZone:   game.ZoneHand,
	})
	return nil
}

func seventhDoctorInvestigate(g *game.Game, controller, source uuid.UUID) error {
	return g.InvestigateForEffect(controller, source, 1)
}

// seventhDoctorArtifacts is the number of artifacts controller controls.
func seventhDoctorArtifacts(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsArtifact() {
			n++
		}
	}
	return n
}

// seventhDoctorChosen is the controller's pick; the defending player is asked
// to guess. An on-disk key: never renamed, never reused.
var seventhDoctorChosen = game.RegisterChooseCardsThen("choose-cards/the-seventh-doctor-chosen", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 {
		return seventhDoctorInvestigate(g, r.Chooser, r.Source)
	}
	ctx := keyedContext(g, r.Chooser, r.Source)
	return MayChoice{
		Player: r.Carry[0],
		Question: "The Seventh Doctor — is the mana value of the card they chose greater than the number of artifacts they control (" +
			strconv.Itoa(seventhDoctorArtifacts(g, r.Chooser)) + " now)?",
		YesLabel: "Greater",
		NoLabel:  "Not greater",
		Then:     seventhDoctorGuessed,
		Carry:    []uuid.UUID{r.Picked[0]},
	}.Apply(ctx)
})

// seventhDoctorGuessed judges the guess (Accepted is "greater"). An on-disk
// key: never renamed, never reused.
var seventhDoctorGuessed = MayAnswer("confirm/the-seventh-doctor-guess", func(ctx *Context, r game.ConfirmAnswer) error {
	g, controller, source := ctx.Game, ctx.Controller(), ctx.Source()
	if len(r.Carry) == 0 {
		return seventhDoctorInvestigate(g, controller, source)
	}
	card, ok := g.LookupCardForEffect(r.Carry[0])
	if !ok {
		return seventhDoctorInvestigate(g, controller, source)
	}
	mv, _ := g.ManaValueForEffect(card)
	greater := mv > seventhDoctorArtifacts(g, controller)
	if greater == r.Accepted {
		return seventhDoctorInvestigate(g, controller, source)
	}
	perm := g.FreeCastFromHandPermission(controller, source, "The Seventh Doctor")
	candidates := freeCastCandidates(g, &perm, []uuid.UUID{r.Carry[0]})
	if len(candidates) == 0 || !g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       "The Seventh Doctor",
		Cards:        candidates,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordHandCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Investigate instead",
		Then:         seventhDoctorCastAnswered,
		Carry:        candidates,
	}) {
		return seventhDoctorInvestigate(g, controller, source)
	}
	return nil
})

// seventhDoctorCastAnswered is "if you don't cast a spell this way,
// investigate". An on-disk key: never renamed, never reused.
var seventhDoctorCastAnswered = game.RegisterChooseCardsThen("resolution-cast/the-seventh-doctor-investigate", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) > 0 {
		return nil
	}
	return seventhDoctorInvestigate(g, r.Chooser, r.Source)
})
