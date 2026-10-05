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
// DECLARED SIMPLIFICATION, weaker than printed: dredge is offered on
// each individual draw as it begins, not on the extra draws a doubler
// such as Thought Reflection adds to it.

// Dredge is "Dredge N" for a card in a graveyard.
func Dredge(n int) game.ReplacementEffect {
	label := fmt.Sprintf("Dredge %d", n)
	return game.ReplacementEffect{
		Watches:       []game.EventKind{game.EventDrawCard},
		FromGraveyard: true,
		Optional:      true,
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventDraw || ev.DrawCount <= 0 || src == nil {
				return false
			}
			p := g.PlayerByIDForEffect(ev.DrawPlayer)
			// CR 702.52b: at least N cards in the library.
			return p != nil && p.Library != nil && p.Library.Size() >= n
		},
		DrawInstead: game.RegisterDrawInstead(fmt.Sprintf("dredge-%d", n),
			func(g *game.Game, drawer, card uuid.UUID, done func(*game.Game) error) error {
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
func dredgeBody(g *game.Game, drawer, card uuid.UUID, n int, done func(*game.Game) error) error {
	return g.MillToZoneThenForEffect(drawer, n, game.ZoneGraveyard, nil,
		func(g *game.Game, _ []uuid.UUID) error {
			if z := g.FindCardZoneForEffect(card); z != nil && z.Kind == game.ZoneGraveyard {
				if err := g.ReturnFromGraveyardForEffect(card, game.ZoneHand); err != nil {
					return err
				}
			}
			return done(g)
		})
}

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
func ChooseFromGraveyardToHandInstead(question string, none func(g *game.Game, drawer, source uuid.UUID) error) game.DrawInsteadFunc {
	return func(g *game.Game, drawer, source uuid.UUID, done func(*game.Game) error) error {
		cards := gyCardIDs(g, drawer, nil)
		if len(cards) == 0 {
			if none == nil {
				return done(g)
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
			Zone:     game.ZoneGraveyard,
			Then: func(g *game.Game, picked []uuid.UUID) error {
				for _, id := range picked {
					if err := g.ReturnFromGraveyardForEffect(id, game.ZoneHand); err != nil {
						return err
					}
				}
				return done(g)
			},
		})
		return nil
	}
}
