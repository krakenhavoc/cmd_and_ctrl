package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Retriever Phoenix — Creature — Phoenix {3}{R}, 2/2 (EDHREC rank
// 26187):
//
//	"Flying, haste
//	 When this creature enters, if you cast it, learn. (You may reveal a
//	 Lesson card you own from outside the game and put it into your
//	 hand, or discard a card to draw a card.)
//	 As long as this card is in your graveyard, if you would learn, you
//	 may instead return this card to the battlefield."
//
// The front half is learn (CR 701.48a, learn.go) behind an intervening
// "if you cast it", read off the event log (b16EnteredFromStack,
// Tiamat's shape): a Phoenix returned, blinked or put onto the
// battlefield another way does not learn, so the Phoenix this card
// brings back is not a source of the next learn.
//
// The back half is a CR 614 replacement effect that works from the
// graveyard (CR 113.6b, "as long as this card is in your graveyard"). It
// rides Spec.Replacements with FromGraveyard set, as a dredge does, and
// replaces the keyword action itself: learn is an event
// (game.LearnForEffect opens a window on it, #3525), so a replaced
// learn never happens (CR 614.6) and the discard, the draw and the
// Lesson are not looked at. Declared Optional, so "you may instead" is
// a yes/no the owner answers each time; declined, the learn goes on as
// printed. Two Phoenixes in a graveyard are two "may"s asked in turn
// (the player takes one of them or none), and a Phoenix beside any other
// replacement of a learn is the ordinary CR 616 ordering prompt.
//
// The return is the Phoenix's owner's: "return this card to the
// battlefield" has no "under your control" clause (CR 400.3). It checks
// the card is still in the graveyard when the body runs.
//
// A learn that is replaced did not happen, so a Lesson sitting
// outside the game is not taken either; that is the card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "279e573d-b7a3-4e62-a708-3ec992d97972",
		Name:            "Retriever Phoenix",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && b16EnteredFromStack(g, source.InstanceID)
			}, "Retriever Phoenix — learn", Do(Learn{})),
		},
		Replacements: []game.ReplacementEffect{ReturnThisInsteadOfLearning("Retriever Phoenix")},
	})
}

// ReturnThisInsteadOfLearning is "as long as this card is in your
// graveyard, if you would learn, you may instead return this card to the
// battlefield" for a card named `name`.
func ReturnThisInsteadOfLearning(name string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:       []game.EventKind{game.EventKeywordAction},
		FromGraveyard: true,
		Optional:      true,
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			// "If YOU would learn": the learner owns the graveyard the
			// gather looked in, and so the card.
			return ev.Kind == game.RepEventKeywordAction &&
				ev.KeywordAction == game.KeywordActionLearn &&
				src != nil && ev.Actor == src.Owner
		},
		KeywordInstead: game.RegisterKeywordInstead("return-this-instead-of-learning", returnThisInsteadOfLearningBody),
		Controller: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) uuid.UUID {
			return ev.Actor
		},
		PromptQuestion: name + " — return it to the battlefield instead of learning?",
		Label:          name + ": return it to the battlefield instead of learning",
		RowLabel:       "While in your graveyard: if you would learn, you may return it to the battlefield instead.",
	}
}

// returnThisInsteadOfLearningBody is "return this card to the
// battlefield" in place of the learn: the card is read from the
// graveyard when the window has settled, so a card that something else
// took out of it in the meantime is simply not returned. The rest of the
// sentence the learn was part of (nothing, in every printing: learn is
// last) runs through done.
func returnThisInsteadOfLearningBody(g *game.Game, _, card uuid.UUID, done func(*game.Game) error) error {
	if z := g.FindCardZoneForEffect(card); z != nil && z.Kind == game.ZoneGraveyard {
		if err := g.ReturnFromGraveyardForEffect(card, game.ZoneBattlefield); err != nil {
			return err
		}
	}
	return done(g)
}
