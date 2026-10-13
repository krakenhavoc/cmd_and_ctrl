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

	// ByOpponent narrows it to "if a spell or ability an OPPONENT
	// controls causes you to discard" (ADR 0153 §1): the replacement
	// applies only when game.DiscardCausedByOpponent(ev). The test is
	// the stack item's controller, never the source card's (CR 113.8).
	// With it, or with a battlefield Dst, AppliesTo also requires the
	// event to still be headed for the graveyard (CR 614.5), the way
	// MadnessReplacement does: once madness or Rest in Peace has moved
	// the card, this no longer applies.
	ByOpponent bool

	// ThisCard narrows it to a discard of the card carrying the
	// replacement ("causes you to discard THIS card"), so it works from
	// the hand and never replaces a discard of a second copy.
	ThisCard bool

	// EntersWithCounters seeds the entry of a battlefield Dst:
	// Dodecapod's two +1/+1 counters. Ignored for any other Dst.
	EntersWithCounters map[string]int

	// OnApply runs when the replacement is applied, before the card is
	// moved: Nephalia Academy's "you may reveal that card".
	OnApply func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error

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
			if d.ThisCard {
				if src == nil || ev.CardID != src.InstanceID {
					return false
				}
			} else if !d.AnyPlayer && ev.DiscardPlayer != src.Controller {
				return false
			}
			if (d.ByOpponent || d.Dst == game.ZoneBattlefield) && ev.NewZone != game.ZoneGraveyard {
				return false
			}
			if d.ByOpponent && !game.DiscardCausedByOpponent(ev) {
				return false
			}
			return d.causeMatches(ev.DiscardCause)
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			if d.OnApply != nil {
				if err := d.OnApply(ev, g, src); err != nil {
					return err
				}
			}
			ev.NewZone = d.Dst
			if d.Dst == game.ZoneBattlefield {
				// The entry pipeline puts it under its owner's control
				// (ADR 0153 §2); the counters ride the entry event.
				ev.NewZoneOwner = uuid.Nil
				for name, n := range d.EntersWithCounters {
					ev.AddCounterAtETB(name, n)
				}
				return nil
			}
			// The discarding player's own zone, named rather than
			// inferred: "put it on top of YOUR library" is the card,
			// and every hand in this engine holds only its owner's
			// cards. A shared zone (exile) ignores the owner.
			ev.NewZoneOwner = ev.DiscardPlayer
			return nil
		},
		Controller: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			if d.ThisCard {
				// A card in a hand has no controller (CR 616.1): its
				// owner, the player discarding it, orders the window.
				return ev.DiscardPlayer
			}
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

// DiscardedByAnOpponent is the triggered half of "if a spell or ability
// an opponent controls causes you to discard" (ADR 0153 §4): the event
// is the source's controller discarding a card because a resolving
// spell or ability an opponent controls told them to.
//
// With self set it is "causes you to discard THIS card" and also
// requires the card to be the source (Sand Golem, Psychic Purge). Wrap
// such a row in InGraveyard, like cycling's "when you cycle this card":
// the card has already reached the graveyard when the ability triggers,
// and "you" is its owner there (CR 108.4). A card a replacement sent to
// exile instead does not trigger from exile, and each card's Caveats
// says so (owner decision 6).
//
// Without self it is "causes you to discard a card" (Spiritual Focus on
// the battlefield).
func DiscardedByAnOpponent(self bool) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		if self && ev.CardID != source.InstanceID {
			return false
		}
		return game.DiscardCausedByOpponentOf(ev, source.Controller)
	}
}

// OntoTheBattlefieldWhenAnOpponentMakesYouDiscardThis is the Obstinate
// Baloth family's replacement (ADR 0153): "If a spell or ability an
// opponent controls causes you to discard this card, put it onto the
// battlefield instead of putting it into your graveyard." `counters` is
// Dodecapod's "with two +1/+1 counters on it"; nil for the rest. The
// card carries it while it is in a hand.
func OntoTheBattlefieldWhenAnOpponentMakesYouDiscardThis(name string, counters map[string]int) game.ReplacementEffect {
	return DiscardBecomes{
		Dst:                game.ZoneBattlefield,
		ByOpponent:         true,
		ThisCard:           true,
		EntersWithCounters: counters,
		Label:              name + ": onto the battlefield instead of the graveyard",
	}.Build()
}
