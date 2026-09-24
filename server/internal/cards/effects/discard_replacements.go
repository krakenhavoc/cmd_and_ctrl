package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_replacements.go — the CR 614 replacement on "if you would
// discard a card". Its own file rather than helpers.go, per the
// convention enters_tapped.go set.
//
// The event is game.RepEventDiscard (#650, ADR 0061), opened by the
// shared exit primitive for every discard in the engine. It carries
// the discarding player, the cause (effect, cost or the cleanup step's
// turn-based action) and the card that asked, alongside the move
// payload a replacement rewrites. That is everything the
// cause-sensitive family needs: Library of Leng reads the cause,
// madness (#657) reads none of it, and the Obstinate Baloth shape
// reads the cause plus the controller of the source.

// DiscardBecomes describes "if you would discard a card, put it
// <Dst> instead" — the one shape the whole family shares, with the
// clauses that differ between its members as fields.
type DiscardBecomes struct {
	// Dst is where the card goes instead of the graveyard. Library of
	// Leng's is game.ZoneLibrary, which means the TOP of the library:
	// the exit primitive puts a redirected card on top unless the
	// route itself asked for the bottom, and a replacement's
	// redirection never inherits the route's to-the-bottom clause.
	Dst game.ZoneKind

	// Causes narrows the replacement to particular causes. Library of
	// Leng is Effect only ("if an EFFECT causes you to discard a
	// card"; costs are not effects, and the cleanup discard is a
	// turn-based action). Empty means every discard, which is what
	// madness replaces.
	Causes []game.DiscardCause

	// Optional makes it a "may" — the controller is asked
	// each time, per card. Library of Leng's "you MAY put it on top of
	// your library instead" is one.
	Optional bool

	// AnyPlayer widens it from "you" to every player's discards. False
	// — the printed default — restricts it to discards by the source's
	// controller.
	AnyPlayer bool

	// Label is the CR 616 prompt header; Question is the yes/no
	// prompt's text when Optional.
	Label    string
	Question string
}

// Build turns the description into the ReplacementEffect a Spec
// declares.
func (d DiscardBecomes) Build() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:        []game.EventKind{game.EventDiscardCard},
		Optional:       d.Optional,
		PromptQuestion: d.Question,
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventDiscard || ev.NewZone == d.Dst {
				return false
			}
			if !d.AnyPlayer && ev.DiscardPlayer != src.Controller {
				return false
			}
			return d.causeMatches(ev.DiscardCause)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.NewZone = d.Dst
			// The discarding player's own zone, named rather than
			// inferred: "put it on top of YOUR library" is the card,
			// and every hand in this engine holds only its owner's
			// cards. A shared zone (exile) ignores the owner.
			ev.NewZoneOwner = ev.DiscardPlayer
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: d.Label,
	}
}

// causeMatches reports whether this discard's cause is one the card
// replaces. No Causes at all means every cause.
func (d DiscardBecomes) causeMatches(cause game.DiscardCause) bool {
	if len(d.Causes) == 0 {
		return true
	}
	for _, c := range d.Causes {
		if c == cause {
			return true
		}
	}
	return false
}
