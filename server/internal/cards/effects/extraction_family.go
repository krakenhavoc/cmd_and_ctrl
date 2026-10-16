package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// extraction_family.go — the shared body of the extraction family
// (#4033): "Search <that player's> graveyard, hand, and library for
// <any number of | all> cards with the same name as that <card | creature
// | spell …> and exile them. Then that player shuffles." Surgical
// Extraction, Extirpate, Eradicate, Scour, Splinter, Sowing Salt and
// Quash each print it after their own first sentence, which picks the
// object whose name is searched for and the player whose zones are
// searched. Kotose, the Silent Spider prints it too and keeps its own
// body (it grants a play over what it exiled).
//
// The spell's controller searches (CR 701.23a), in one prompt over the
// three zones (game.SearchLibrarySpec.AlsoGraveyardAndHand). "Any number"
// is Unbounded: every match is the searcher's to take or leave. "All" is
// TakeAll: the graveyard's matches are found (a public zone has no fail to
// find, CR 701.23b), and the hand's and library's may still be left,
// because those zones are hidden. The library is shuffled whatever was
// taken.

// extractionNames is every name `c` has (CR 201.2a): one for an ordinary
// card, two for a split card, read off the object as it is now — a
// permanent's copied name, a spell's cast half. A face-down permanent has
// no name and shares a name with nothing (CR 708.2a, 201.2a).
func extractionNames(c game.Card) []string {
	if c.FaceDown {
		return nil
	}
	c.Name = c.Effective().Name
	return game.NamesOf(c)
}

// sharesANameWith is "with the same name as that card": at least one name
// in common (CR 201.2a). No names match nothing.
func sharesANameWith(names []string) func(game.Card) bool {
	return func(c game.Card) bool {
		for _, n := range names {
			if game.HasName(c, n) {
				return true
			}
		}
		return false
	}
}

// notABasicLandCard is "card other than a basic land card".
func notABasicLandCard() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return !IsBasicLand(c) }
}

// extractionSearch is the family's second and third sentences: the
// controller of the resolving spell searches `victim`'s graveyard, hand
// and library for cards sharing a name with `names`, exiles what is
// found, and `victim` shuffles. `all` is "all cards", false is "any
// number of cards". Victim uuid.Nil (nobody to search) does nothing.
func extractionSearch(ctx *Context, card string, victim uuid.UUID, names []string, all bool) error {
	if victim == uuid.Nil {
		return nil
	}
	named := strings.Join(names, " or ")
	if named == "" {
		named = "a name (it has none)"
	}
	reason := card + " — search their graveyard, hand and library for cards named " + named
	if all {
		reason += "; every one in their graveyard is exiled"
	}
	return SearchLibrary{
		Player:               ctx.Controller(),
		LibraryOwner:         victim,
		Predicate:            sharesANameWith(names),
		Dest:                 game.ZoneExile,
		Unbounded:            !all,
		TakeAll:              all,
		AlsoGraveyardAndHand: true,
		Shuffle:              true,
		Reason:               reason,
	}.Apply(ctx)
}

// extractFromGraveyardTarget is Surgical Extraction's and Extirpate's
// body: "Choose target card in a graveyard other than a basic land card.
// Search its owner's graveyard, hand, and library …". The target is only
// chosen, not exiled; it is found by the graveyard half of the search like
// any other card with its name. A target that has left the graveyard
// makes the spell do nothing (CR 608.2b).
func extractFromGraveyardTarget(card string, all bool) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		target, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		c, ok := ctx.Game.LookupCardForEffect(target)
		if !ok {
			return nil
		}
		return extractionSearch(ctx, card, c.Owner, extractionNames(c), all)
	}
}

// exilePermanentThenExtract is Eradicate's, Scour's, Splinter's and Sowing
// Salt's body: "Exile target <permanent>. Search its controller's
// graveyard, hand, and library for all cards with the same name as that
// <permanent> …". The name and the controller are the permanent's as it
// last existed on the battlefield, read before it is exiled. The search
// happens whether or not the permanent reached exile (a replacement may
// send it elsewhere); a target that is no longer legal makes the spell do
// nothing (CR 608.2b).
func exilePermanentThenExtract(card string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		target, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		c, ok := ctx.Game.LookupCardForEffect(target)
		if !ok {
			return nil
		}
		victim, names := c.Controller, extractionNames(c)
		return ExileTarget{
			Target: target,
			Then: func(ctx *Context, _ bool) error {
				return extractionSearch(ctx, card, victim, names, true)
			},
		}.Apply(ctx)
	}
}
