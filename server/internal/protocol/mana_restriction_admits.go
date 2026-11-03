package protocol

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ManaRestrictionsAdmitCast reports whether mana carrying `restrictions`
// (a ManaAbilityView's Restrictions) may pay for casting `card` from the
// zone `fromZone` names ("" is the hand). It is the engine's matcher
// (game.CastAdmitsRestrictedMana) asked with the card's projected type
// line and colours, and its changeling keyword for "every creature type"
// (CR 702.73a).
//
// The bot's turn planner reads it (#3976): a policy may not import
// internal/game (ADR 0033 §3), so the engine hands it the answer rather
// than letting it re-implement the tag grammar. An estimate for planning,
// not a legality answer: the enumerator decides the cast.
//
// It judges the face the card shows with {X} at 0; a reader that knows
// the face being cast and the X announced asks ManaRestrictionsAdmitCastAs.
func ManaRestrictionsAdmitCast(restrictions []string, card *CardView, fromZone string) bool {
	face := 0
	if card != nil {
		face = card.ActiveFace
	}
	return ManaRestrictionsAdmitCastAs(restrictions, card, fromZone, face, 0)
}

// ManaRestrictionsAdmitCastAs is ManaRestrictionsAdmitCast for a cast of
// printed face `face` (ADR 0034) with X announced as x, so a restriction
// on the spell's mana value (game.ManaRestrictManaValueAtLeast, #4694)
// reads the mana value the spell will have on the stack (CR 202.3e).
func ManaRestrictionsAdmitCastAs(restrictions []string, card *CardView, fromZone string, face, x int) bool {
	if len(restrictions) == 0 {
		return true
	}
	if card == nil {
		return false
	}
	changeling := false
	for _, a := range card.Abilities {
		if strings.EqualFold(strings.TrimSpace(a), game.KeywordChangeling) {
			changeling = true
			break
		}
	}
	return game.CastAdmitsRestrictedMana(restrictions, card.TypeLine, card.Colors, changeling, fromZone, castManaCostOf(card, face), x)
}

// castManaCostOf is the cost a cast of `face` of card computes its mana
// value from: that face's cost for a card with several castable faces (a
// modal DFC, a split card's half, an adventure), and the front face's for
// a transform or flip card, whose back face has none of its own (CR
// 712.8c, CR 710.1c). A card with no faces is its own cost. A face-down
// cast is not told apart here: an estimate, as above.
func castManaCostOf(card *CardView, face int) string {
	if len(card.Faces) == 0 {
		return card.ManaCost
	}
	if card.Layout == game.LayoutTransform || card.Layout == game.LayoutFlip || face < 0 || face >= len(card.Faces) {
		face = 0
	}
	return card.Faces[face].ManaCost
}
