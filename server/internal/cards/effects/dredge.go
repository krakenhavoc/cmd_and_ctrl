package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dredge.go — dredge (CR 702.52, #2127) and the draw replacements that
// stop to ask what you take (#2168, draw_instead.go in the engine).
//
//	CR 702.52a  Dredge N means "As long as you have at least N cards in
//	            your library, if you would draw a card, you may instead
//	            mill N cards and return this card from your graveyard
//	            to your hand."
//	CR 702.52b  A player with fewer than N cards in their library can't
//	            dredge N.
//
// It is a replacement effect that works from the GRAVEYARD, so it rides
// Spec.Replacements with FromGraveyard set: the engine finds it in the
// drawing player's graveyard instead of on the battlefield, asks the
// "may" through the ordinary yes/no prompt, and on "yes" cancels the
// draw and runs the body once the window settles. Several dredge cards
// in one graveyard each offer the choice in turn, so the player takes
// one of them or none; the first "yes" ends that draw. A multi-card
// draw ("draw three") asks once per card, because every individual draw
// opens its own window (CR 121.2).
//
// Declined, a dredge changes nothing: the card is drawn as usual.
//
// Draws added by a doubler such as Thought Reflection each open their
// own window too, so each can independently be replaced by dredge.

// Dredge is "Dredge N" for a card in a graveyard.
func Dredge(n int) game.ReplacementEffect {
	label := fmt.Sprintf("Dredge %d", n)
	return game.ReplacementEffect{
		Watches:       []game.EventKind{game.EventDrawCard},
		FromGraveyard: true,
		Optional:      true,
		Dredge:        n,
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventDraw || ev.DrawCount <= 0 || src == nil {
				return false
			}
			p := g.PlayerByIDForEffect(ev.DrawPlayer)
			// CR 702.52b: at least N cards in the library.
			return p != nil && p.Library != nil && p.Library.Size() >= n
		},
		DrawInstead: game.RegisterDrawInstead(fmt.Sprintf("dredge-%d", n),
			func(g *game.Game, drawer, card uuid.UUID, done game.DrawDone) error {
				return dredgeBody(g, drawer, card, n, done)
			}),
		Controller: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) uuid.UUID {
			return ev.DrawPlayer
		},
		PromptQuestion: fmt.Sprintf("Dredge %d — mill %d cards and return this card to your hand instead of drawing?", n, n),
		Label:          label,
	}
}

// dredgeBody is "mill N cards and return <the card> to your hand": the
// mill first (it can pause on a replacement's prompt), then the return.
// The card is still read from the graveyard when the mill is done, so a
// card something else removed in between is simply not returned.
//
// The return rides the mill as data (#4253): the card and the parked draw
// (DrawDone.Park) are the keyed tail's Carry, so a CR 616 prompt the mill
// pauses on is a restore point.
func dredgeBody(g *game.Game, drawer, card uuid.UUID, n int, done game.DrawDone) error {
	then := game.KeyedTail{Ref: dredgeReturn, Player: drawer, Carry: []uuid.UUID{card, done.Park(g)}}
	return g.MillToZoneKeyedForEffect(drawer, n, game.ZoneGraveyard, game.MillUntil{}, then)
}

// dredgeReturn is the end of a dredge's mill: the card (Carry[0]) returns
// from the graveyard to its owner's hand if it is still there, then the
// rest of the draw (the parked DrawDone, Carry[1]) is paid. An on-disk
// key.
var dredgeReturn = game.RegisterTail("dredge/return-then-draw", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) < 2 {
		return nil
	}
	card := o.Tail.Carry[0]
	if z := g.FindCardZoneForEffect(card); z != nil && z.Kind == game.ZoneGraveyard {
		if err := g.ReturnFromGraveyardForEffect(card, game.ZoneHand); err != nil {
			return err
		}
	}
	return game.DrawDoneByID(o.Tail.Carry[1]).Run(g)
})

// gyCardIDs lists the cards in a player's graveyard, oldest
// first, optionally filtered.
func gyCardIDs(g *game.Game, player uuid.UUID, match func(game.Card) bool) []uuid.UUID {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var out []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if match == nil || match(c) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// ChooseFromGraveyardToHandInstead is the body of "if you would draw a
// card, return a card from your graveyard to your hand instead": a
// mandatory pick, then the return, then the rest of the draw
// instruction. With no candidate it calls `none` — Forbidden Crypt's
// "if you can't, you lose the game" — and does NOT call done.
//
// The rest of the draw rides the prompt as data (#4253 slice 10c): `done`
// is parked on the game and its ID is the prompt's Carry, so the table is a
// restore point while the player picks.
func ChooseFromGraveyardToHandInstead(question string, none func(g *game.Game, drawer, source uuid.UUID) error) game.DrawInsteadFunc {
	return func(g *game.Game, drawer, source uuid.UUID, done game.DrawDone) error {
		cards := gyCardIDs(g, drawer, nil)
		if len(cards) == 0 {
			if none == nil {
				return done.Run(g)
			}
			return none(g, drawer, source)
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  drawer,
			Source:   source,
			Question: question,
			Cards:    cards,
			Min:      1,
			Max:      1,
			LiveZone: game.ZoneGraveyard,
			ThenKey:  drawInsteadGraveyardReturn,
			Carry:    []uuid.UUID{done.Park(g)},
		})
		return nil
	}
}

// drawInsteadGraveyardReturn is the pick of ChooseFromGraveyardToHandInstead:
// the chosen card returns to hand, then the rest of the draw is paid. The
// carried ID is the parked draw. An on-disk key.
var drawInsteadGraveyardReturn = game.RegisterChooseCardsThen("choose-cards/draw-instead-graveyard-return", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	done := game.DrawDoneByID(r.Carry[0])
	for _, id := range r.Picked {
		if err := g.ReturnFromGraveyardForEffect(id, game.ZoneHand); err != nil {
			return err
		}
	}
	return done.Run(g)
})
