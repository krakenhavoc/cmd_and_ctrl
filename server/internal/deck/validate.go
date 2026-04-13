package deck

import (
	"fmt"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// Violation is one validation failure. Collected into ValidationError
// so a deck with multiple problems surfaces them all in one round-
// trip — rebuilding a 99-card deck name-by-name because the server
// only reports one issue at a time is miserable UX.
type Violation struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	// Card, if non-empty, names the card the violation is about.
	Card string `json:"card,omitempty"`
}

// Violation codes. Stable strings so the client can key UI treatments
// off them (e.g. highlight the card row on a color-identity fail).
const (
	CodeWrongCardCount       = "wrong_card_count"
	CodeMissingCommander     = "missing_commander"
	CodeTooManyCommanders    = "too_many_commanders"
	CodeNotLegalCommander    = "not_a_legal_commander"
	CodeColorIdentity        = "color_identity_violation"
	CodeSingleton            = "singleton_violation"
	CodeNotLegalInFormat     = "not_legal_in_format"
	CodeSideboardUnsupported = "sideboard_not_supported_in_commander"
)

// ValidationError bundles one or more Violations into a single error
// value. Implements the error interface so validation composes with
// the rest of the package's error surface.
type ValidationError struct {
	Violations []Violation
}

func (e *ValidationError) Error() string {
	if len(e.Violations) == 0 {
		return "deck: validation failed"
	}
	msgs := make([]string, len(e.Violations))
	for i, v := range e.Violations {
		msgs[i] = v.Message
	}
	return "deck: " + strings.Join(msgs, "; ")
}

// Validate enforces Commander-format rules on a resolved List. At S05
// those rules are:
//
//  1. Exactly one commander.
//  2. Commander itself must be Scryfall-legal in the commander format
//     (type line is "Legendary Creature" or oracle text contains the
//     "can be your commander" clause via the "legalities.commander"
//     entry being "legal").
//  3. Mainboard + Commanders must total exactly 100 cards.
//  4. All non-basic-land cards are singletons (at most one copy).
//  5. Every card is legal in the commander format — specifically,
//     legalities.commander != "banned" and != "not_legal".
//  6. Every card's color_identity is a subset of the commander's
//     color identity (union when we add partner support later).
//
// Sideboard cards are surfaced as a warning, not an error — Commander
// doesn't use a sideboard but "I accidentally pasted a Standard deck"
// is a recoverable mistake that shouldn't reject the whole upload.
func Validate(list *List) error {
	if list == nil {
		return fmt.Errorf("deck: nil list")
	}
	var vs []Violation

	// Commander presence
	switch len(list.Commanders) {
	case 0:
		vs = append(vs, Violation{Code: CodeMissingCommander, Message: "deck has no commander"})
	case 1:
		// happy path
	default:
		vs = append(vs, Violation{
			Code:    CodeTooManyCommanders,
			Message: fmt.Sprintf("deck has %d commanders; partner/companion is not supported at S05", len(list.Commanders)),
		})
	}

	// Commander legality: type line must include "Legendary" and the
	// card must be legal in the commander format. The "can be your
	// commander" oracle clause (Planeswalker commanders like Oloro or
	// the "creature type commander" mechanics) is captured by the
	// legalities.commander = "legal" flag.
	if len(list.Commanders) >= 1 {
		cmd := list.Commanders[0]
		if !isLegalCommander(cmd) {
			vs = append(vs, Violation{
				Code:    CodeNotLegalCommander,
				Card:    cmd.Name,
				Message: fmt.Sprintf("%q cannot be a commander", cmd.Name),
			})
		}
	}

	// Total count: commanders + mainboard should be 100.
	total := len(list.Commanders) + len(list.Mainboard)
	if total != 100 {
		vs = append(vs, Violation{
			Code:    CodeWrongCardCount,
			Message: fmt.Sprintf("deck has %d cards; expected 100", total),
		})
	}

	// Singleton: at most one of each non-basic card.
	seen := make(map[string]int, len(list.Mainboard))
	for _, c := range list.Mainboard {
		if isBasicLand(c) {
			continue
		}
		seen[c.Name]++
	}
	for name, n := range seen {
		if n > 1 {
			vs = append(vs, Violation{
				Code:    CodeSingleton,
				Card:    name,
				Message: fmt.Sprintf("%q appears %d times; Commander is singleton", name, n),
			})
		}
	}

	// Format legality: every card must be commander-legal.
	for _, c := range allCards(list) {
		legal, ok := c.Legalities["commander"]
		// Missing entry — treat as unknown / not_legal. Data quality
		// issues in the Scryfall dump shouldn't silently pass.
		if !ok || legal == "not_legal" || legal == "banned" {
			vs = append(vs, Violation{
				Code:    CodeNotLegalInFormat,
				Card:    c.Name,
				Message: fmt.Sprintf("%q is not legal in the commander format (status: %q)", c.Name, legal),
			})
		}
	}

	// Color identity: every mainboard card's identity must be a
	// subset of the commander's identity. Uses string set ops on the
	// single-letter WUBRG values.
	if len(list.Commanders) == 1 {
		allowed := identitySet(list.Commanders[0].ColorIdentity)
		for _, c := range list.Mainboard {
			for _, sym := range c.ColorIdentity {
				if _, ok := allowed[sym]; !ok {
					vs = append(vs, Violation{
						Code:    CodeColorIdentity,
						Card:    c.Name,
						Message: fmt.Sprintf("%q has color identity %v but commander allows %v", c.Name, c.ColorIdentity, list.Commanders[0].ColorIdentity),
					})
					break // one violation per card — don't spam
				}
			}
		}
	}

	// Sideboard warning (surfaced as a violation too, but non-fatal
	// at the caller's discretion). We emit it so the client can show
	// a yellow banner rather than silently discard the sideboard.
	if len(list.Sideboard) > 0 {
		vs = append(vs, Violation{
			Code:    CodeSideboardUnsupported,
			Message: fmt.Sprintf("deck has %d sideboard cards; Commander doesn't use a sideboard and they will be ignored", len(list.Sideboard)),
		})
	}

	if len(vs) == 0 {
		return nil
	}
	return &ValidationError{Violations: vs}
}

// isLegalCommander captures the "can this card be in the command
// zone" predicate. Covers legendary creatures (the default case) and
// the handful of non-creature commanders whose oracle text says
// "can be your commander". The simplest proxy for the latter is
// Scryfall's type-line + legalities combination: if the type line
// includes "Legendary" and the card is commander-legal, it can be a
// commander.
func isLegalCommander(c cards.Card) bool {
	if c.Legalities["commander"] != "legal" {
		return false
	}
	typeLine := c.TypeLine
	if len(c.CardFaces) > 0 && typeLine == "" {
		typeLine = c.CardFaces[0].TypeLine
	}
	if !strings.Contains(typeLine, "Legendary") {
		// Some commanders are tokens/planeswalkers with an explicit
		// "can be your commander" clause in the oracle text.
		return strings.Contains(c.OracleText, "can be your commander")
	}
	return true
}

// isBasicLand returns true for basic-land printings, which are
// exempt from the Commander singleton rule. We match on the type
// line rather than a hard-coded name list so Wastes, Snow-Covered
// Forest, and any future basics land correctly.
func isBasicLand(c cards.Card) bool {
	return strings.Contains(c.TypeLine, "Basic Land")
}

// allCards returns a slice over the commanders + mainboard for
// cards-global checks. Sideboard is excluded — we ignore it for
// gameplay purposes.
func allCards(list *List) []cards.Card {
	out := make([]cards.Card, 0, len(list.Commanders)+len(list.Mainboard))
	out = append(out, list.Commanders...)
	out = append(out, list.Mainboard...)
	return out
}

// identitySet builds an O(1) lookup set from a color_identity slice.
// Uppercase WUBRG letters, empty slice maps to an empty set (colorless
// commanders can only contain colorless cards).
func identitySet(syms []string) map[string]struct{} {
	out := make(map[string]struct{}, len(syms))
	for _, s := range syms {
		out[strings.ToUpper(s)] = struct{}{}
	}
	return out
}
