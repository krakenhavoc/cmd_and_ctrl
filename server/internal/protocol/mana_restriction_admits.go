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
func ManaRestrictionsAdmitCast(restrictions []string, card *CardView, fromZone string) bool {
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
	return game.CastAdmitsRestrictedMana(restrictions, card.TypeLine, card.Colors, changeling, fromZone)
}
