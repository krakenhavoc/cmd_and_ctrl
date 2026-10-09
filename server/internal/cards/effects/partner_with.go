package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// partner_with.go — the entry half of "Partner with [name]" (#2142,
// CR 702.124j). Append-only.
//
// CR 702.124j makes the keyword two abilities. The deck-construction
// one ("You may designate two legendary cards as your commander rather
// than one if each has a 'partner with [name]' ability with the other's
// name") is read off the oracle text by internal/deck, for every card,
// catalog or not. The other is the trigger here:
//
//	"When this permanent enters, target player may search their library
//	 for a card named [name], reveal it, put it into their hand, then
//	 shuffle."
//
// Three pieces of existing machinery, nothing new in the engine:
//
//   - The target is chosen as the trigger goes on the stack (CR 603.3d)
//     by its controller, and can be any player, an opponent included.
//     It is re-checked on resolution (CR 608.2b): a target who has left
//     the game means the ability does nothing.
//   - "May" is a choice made on resolution (CR 603.5), and it is the
//     TARGET PLAYER's: the MayChoice is addressed to them. Declining is
//     not searching at all, so nothing is revealed and the library is
//     not shuffled.
//   - The search is an ordinary search of their own library for a card
//     with a stated quality, a name, so they may fail to find it even
//     when it is there (CR 701.23b): the search prompt is Optional and
//     taking nothing still shuffles. The card found is revealed to the
//     table (CR 701.23e: the instruction says to) and goes to the hand.
//
// The row declares its Effect (ADR 0041 P9), so a trigger waiting on
// the stack is a restore point and rebuilds from the row on restore.

// partnerWithKeyMark is the text every partner-with row's label carries
// between the card's name and its partner's. The roadmap's probe finds
// the rows by it.
const partnerWithKeyMark = " — partner with "

// PartnerWith is the entry trigger of "Partner with [partner]" on the
// card named `card`: "When this permanent enters, target player may
// search their library for a card named [partner], reveal it, put it
// into their hand, then shuffle" (CR 702.124j).
func PartnerWith(card, partner string) game.TriggeredAbility {
	return Targeting(
		WhenThisEnters(card+partnerWithKeyMark+partner, partnerWithSearch(partner)),
		TargetPlayer("target player"))
}

// IsPartnerWithRow reports whether a triggered row is PartnerWith's.
func IsPartnerWithRow(t game.TriggeredAbility) bool {
	return strings.Contains(t.Key, partnerWithKeyMark)
}

// partnerWithSearch is the trigger's resolution: the target player is
// asked, and searches on a yes.
func partnerWithSearch(partner string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetPlayer {
				continue
			}
			searcher := t.ID
			return MayChoice{
				Player:   searcher,
				Question: "Partner with " + partner + " — search your library for a card named " + partner + "?",
				YesLabel: "Search",
				NoLabel:  "Don't search",
				OnYes: func(ctx *Context) error {
					return SearchLibrary{
						Player:    searcher,
						Predicate: func(c game.Card) bool { return game.HasName(c, partner) },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Partner with — choose a card named " + partner + " to reveal and put into your hand",
					}.Apply(ctx)
				},
			}.Apply(ctx)
		}
		return nil
	}
}
