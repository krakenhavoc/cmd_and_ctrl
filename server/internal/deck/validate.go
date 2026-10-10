package deck

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
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
	// CodeUnsupportedLayout names cards whose printing carries more
	// faces than the engine can play (ADR 0034). Non-fatal: the card
	// imports as its front face and works as that half. The message
	// says which half, and what is lost.
	CodeUnsupportedLayout = "unsupported_layout"
	// CodeInvalidPartnerPair flags two commanders that no partner
	// ability pairs, when at least one of them has one or is a
	// Background: a "Partner with" card naming a different card, a
	// Choose a Background commander beside a card that is not a
	// Background, and so on (CR 702.124, #2142, #2874). Carries the
	// card at fault.
	CodeInvalidPartnerPair = "invalid_partner_pair"
	// CodeUnknownCard flags a decklist row whose name did not resolve
	// against the Scryfall index. Surfaced via UnknownCardError and
	// translated into the 422 violations[] list so the client can
	// highlight every offending row at once.
	CodeUnknownCard = "unknown_card"
	// CodeUnsupportedMechanic flags a resolved card that leans on a
	// mechanic (companion) the server hasn't modeled yet.
	CodeUnsupportedMechanic = "unsupported_mechanic"
	// CodeUnknownSource flags a URL-based import whose host isn't
	// one of the supported deck-builders (S06.5). Carries the offending
	// URL in the `card` field so the client can echo it back.
	CodeUnknownSource = "unknown_source"
	// CodeDeckNotFound is a URL-import failure where the upstream
	// returned 404 — the deck was deleted or the ID was wrong.
	CodeDeckNotFound = "deck_not_found"
	// CodeDeckPrivate is a URL-import failure where the upstream
	// returned a JSON 401/403 — the deck exists but is not publicly
	// readable. The user's fix is to flip the deck to Public or paste
	// the JSON/text export directly.
	CodeDeckPrivate = "deck_private"
	// CodeUpstreamBlocked is a URL-import failure where the upstream's
	// CDN (Cloudflare, typically) returned a 403 bot-wall HTML page
	// instead of a JSON response. The deck itself may be public; the
	// server's egress IP is being rate-limited or reputation-scored
	// by the CDN. Distinguished from CodeDeckPrivate so the client
	// can surface the right fix (try Archidekt, paste the text
	// export, or route outbound traffic through a different IP)
	// rather than nudging the user to change Moxfield's privacy
	// setting on a deck that's already public.
	CodeUpstreamBlocked = "upstream_blocked"
	// CodeExternalAPIUnavailable is a URL-import failure for upstream
	// 5xx / connection / timeout errors. Retry-after-a-bit semantics;
	// not the user's fault.
	CodeExternalAPIUnavailable = "external_api_unavailable"
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
//  1. Exactly one commander, or two that a partner ability pairs
//     (CR 702.124, #2142, #2874). A Background is never the only
//     commander (CR 702.124k).
//  2. Commander itself must be Scryfall-legal in the commander format
//     (type line is "Legendary Creature" or oracle text contains the
//     "can be your commander" clause via the "legalities.commander"
//     entry being "legal").
//  3. Mainboard + Commanders must total exactly 100 cards.
//  4. All non-basic-land cards are singletons (at most one copy).
//  5. Every card is legal in the commander format — specifically,
//     legalities.commander != "banned" and != "not_legal".
//  6. Every card's color_identity is a subset of the commander's
//     color identity (the union of a pair's, CR 702.124c).
//
// Sideboard cards are surfaced as a warning, not an error — Commander
// doesn't use a sideboard but "I accidentally pasted a Standard deck"
// is a recoverable mistake that shouldn't reject the whole upload.
func Validate(list *List) error {
	if list == nil {
		return fmt.Errorf("deck: nil list")
	}
	var vs []Violation

	// Commander presence. Two commanders are allowed when a partner
	// ability pairs them (partner.go, CR 702.124, #2142, #2874); no
	// partner ability allows more than two (CR 702.124g).
	switch len(list.Commanders) {
	case 0:
		vs = append(vs, Violation{Code: CodeMissingCommander, Message: "deck has no commander"})
	case 1:
		if v := loneCommanderViolation(list.Commanders[0]); v != nil {
			vs = append(vs, *v)
		}
	case 2:
		if v := commanderPairViolation(list.Commanders[0], list.Commanders[1]); v != nil {
			vs = append(vs, *v)
		}
	default:
		vs = append(vs, Violation{
			Code:    CodeTooManyCommanders,
			Message: fmt.Sprintf("deck has %d commanders; at most two are allowed, and only when a partner ability pairs them", len(list.Commanders)),
		})
	}

	// Commander legality: type line must include "Legendary" and the
	// card must be legal in the commander format. The "can be your
	// commander" oracle clause (Planeswalker commanders like Oloro or
	// the "creature type commander" mechanics) is captured by the
	// legalities.commander = "legal" flag. Each commander of a pair is
	// judged on its own (CR 702.124a: "two legendary cards"); a
	// Background is a legendary enchantment, so it passes here and
	// the pair rule above decides whether it may be a commander.
	for _, cmd := range list.Commanders {
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

	// Singleton: at most one of each non-basic card, unless the card's
	// own text says otherwise (CR 113.6n, CR 903.5b's exception).
	seen := make(map[string]int, len(list.Mainboard))
	limits := make(map[string]copyLimit, len(list.Mainboard))
	for _, c := range list.Mainboard {
		if isBasicLand(c) {
			continue
		}
		seen[c.Name]++
		limits[c.Name] = deckCopyLimit(c)
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		n, lim := seen[name], limits[name]
		if lim.any || n <= lim.max {
			continue
		}
		msg := fmt.Sprintf("%q appears %d times; Commander is singleton", name, n)
		if lim.max > 1 {
			msg = fmt.Sprintf("%q appears %d times; a deck can have up to %d cards named %s", name, n, lim.max, name)
		}
		vs = append(vs, Violation{Code: CodeSingleton, Card: name, Message: msg})
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
	//
	// Two commanders' identities combine (CR 702.124c, 903.4).
	if len(list.Commanders) == 1 || len(list.Commanders) == 2 {
		allowed := map[string]struct{}{}
		for _, cmd := range list.Commanders {
			for sym := range identitySet(cmd.ColorIdentity) {
				allowed[sym] = struct{}{}
			}
		}
		allowedList := commanderIdentityList(allowed)
		for _, c := range list.Mainboard {
			for _, sym := range c.ColorIdentity {
				if _, ok := allowed[sym]; !ok {
					vs = append(vs, Violation{
						Code:    CodeColorIdentity,
						Card:    c.Name,
						Message: fmt.Sprintf("%q has color identity %v but commander allows %v", c.Name, c.ColorIdentity, allowedList),
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

	// ADR 0034 — declared simplifications on multi-face layouts, in
	// the same non-fatal shape as the sideboard warning above.
	//
	// Refusing the import would reject a whole deck over a card that
	// is merely cosmetically wrong; SILENCE is what produced #265,
	// where a Sea Gate Restoration arrived on the battlefield as a
	// land with no prompt and nothing said why. A yellow banner is
	// the correct middle.
	vs = append(vs, unsupportedLayoutViolations(list)...)

	if len(vs) == 0 {
		return nil
	}
	return &ValidationError{Violations: vs}
}

// isLegalCommander captures the "can this card be in the command
// zone" predicate. Covers legendary creatures (the default case) and
// the handful of non-creature commanders whose oracle text says
// "can be your commander". Walks every face so DFC/MDFC legendary
// commanders (e.g. Esika, God of the Tree // The Prismatic Bridge)
// aren't rejected when their top-level TypeLine concatenates oddly.
func isLegalCommander(c cards.Card) bool {
	if c.Legalities["commander"] != "legal" {
		return false
	}
	if commanderOnFace(c.TypeLine, c.OracleText) {
		return true
	}
	for _, face := range c.CardFaces {
		if commanderOnFace(face.TypeLine, face.OracleText) {
			return true
		}
	}
	return false
}

// commanderOnFace is the per-face predicate used by isLegalCommander:
// a legendary type line qualifies, OR an oracle text that explicitly
// states "can be your commander" (the rider clause on non-creature
// commanders like Faceless Haven's legendary transformations).
func commanderOnFace(typeLine, oracleText string) bool {
	if strings.Contains(typeLine, "Legendary") {
		return true
	}
	return strings.Contains(oracleText, "can be your commander")
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

// layoutSimplifications names, per Scryfall layout, what the engine
// does with a card the multi-face model does not fully play yet
// (ADR 0034).
//
// modal_dfc is deliberately ABSENT: both its faces are playable and
// the face picker chooses between them, so there is nothing to warn
// about. So is `normal` and every other single-faced layout.
//
// `split` left it with ADR 0103: either half is cast (CR 709.3), fuse
// from hand and aftermath from a graveyard, and a Room's doors unlock.
//
// `adventure` LEFT this map in #719, for the same reason modal_dfc was
// never in it: CR 715 is played now. Both halves are offered from
// hand, the Adventure spell exiles as it resolves (CR 715.3d) and its
// owner may cast the creature from exile (CR 715.4). `prepare` left it
// in #1328 for the same reason (ADR 0090): the card is cast as its
// creature half, which CR 722.3 says is the only way it is ever cast,
// and a prepared permanent's prepare spell is cast as a copy out of
// exile. What both still share with every other card is the ordinary
// catalog gap — an Adventure half with no Spec resolves manually, and a
// preparation card with no Spec never becomes prepared, which is the
// `unimplemented` badge's job (ADR 0037) and not a layout warning.
//
// `transform` left it too: the engine turns a permanent over in place
// (game.TransformPermanentForEffect, server/internal/game/transform.go),
// and the roadmap marks transform implemented. A transform card with a
// back face the engine cannot yet produce says so through its Caveats
// or its `unimplemented` badge (ADR 0037), not through a layout banner.
var layoutSimplifications = map[string]string{
	"flip": "imports as its front face only",
	"meld": "imports as its front face only; melding isn't implemented yet",
}

// unsupportedLayoutViolations reports one non-fatal violation per
// distinct simplification, naming the cards it applies to.
//
// Grouped by layout rather than emitted per card because a deck can
// legitimately hold a dozen transform cards and twelve identical
// banners is noise, not information. The cards are named inside the
// message so the player can still see exactly which of their deck is
// affected.
func unsupportedLayoutViolations(list *List) []Violation {
	byLayout := map[string][]string{}
	seen := map[string]bool{}
	for _, c := range allCards(list) {
		note, ok := layoutSimplifications[c.Layout]
		if !ok || note == "" {
			continue
		}
		// One mention per distinct card — a deck running four copies
		// of a transform card lists it once.
		key := c.Layout + "\x00" + c.Name
		if seen[key] {
			continue
		}
		seen[key] = true
		byLayout[c.Layout] = append(byLayout[c.Layout], frontFaceName(c))
	}
	if len(byLayout) == 0 {
		return nil
	}
	layouts := make([]string, 0, len(byLayout))
	for layout := range byLayout {
		layouts = append(layouts, layout)
	}
	sort.Strings(layouts)

	out := make([]Violation, 0, len(layouts))
	for _, layout := range layouts {
		names := byLayout[layout]
		sort.Strings(names)
		out = append(out, Violation{
			Code: CodeUnsupportedLayout,
			Card: names[0],
			Message: fmt.Sprintf("%s: %s (%s)",
				strings.Join(names, ", "),
				layoutSimplifications[layout],
				layout),
		})
	}
	return out
}

// frontFaceName is the name a player will actually see once the card
// is imported — the front face's, not Scryfall's "A // B"
// composite, which is what game.Card carries since ADR 0034.
func frontFaceName(c cards.Card) string {
	if len(c.CardFaces) > 1 && c.CardFaces[0].Name != "" {
		return c.CardFaces[0].Name
	}
	return c.Name
}

// copyLimit is how many copies of one card a deck may hold. The zero
// value is the Commander default of one.
type copyLimit struct {
	any bool // "any number of cards named ~"
	max int  // "up to N cards named ~"; 1 when the card says nothing
}

// deckCopyRule matches the printed deck-construction exception
// (CR 113.6n): "A deck can have any number of cards named ~." and
// "A deck can have up to nine cards named ~."
var deckCopyRule = regexp.MustCompile(`(?i)a deck can have (any number of|up to ([a-z0-9-]+)) cards? named`)

var numberWords = map[string]int{
	"one": 1, "two": 2, "three": 3, "four": 4, "five": 5, "six": 6, "seven": 7,
	"eight": 8, "nine": 9, "ten": 10, "eleven": 11, "twelve": 12,
	"thirteen": 13, "fourteen": 14, "fifteen": 15, "twenty": 20,
}

// deckCopyLimit reads a card's allowance from its own oracle text, on
// any face, so no card is named in the validator. A phrase it cannot
// read (an unknown number word) leaves the singleton default, the
// strict direction.
func deckCopyLimit(c cards.Card) copyLimit {
	lim := copyLimit{max: 1}
	texts := []string{c.OracleText}
	for _, f := range c.CardFaces {
		texts = append(texts, f.OracleText)
	}
	for _, t := range texts {
		m := deckCopyRule.FindStringSubmatch(t)
		if m == nil {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(m[1]), "up to") {
			return copyLimit{any: true}
		}
		w := strings.ToLower(m[2])
		n, ok := numberWords[w]
		if !ok {
			n, _ = strconv.Atoi(w)
		}
		if n > lim.max {
			lim.max = n
		}
	}
	return lim
}

// commanderIdentityList is a combined identity set as WUBRG-ordered
// letters, for a violation message.
func commanderIdentityList(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for _, sym := range []string{"W", "U", "B", "R", "G"} {
		if _, ok := set[sym]; ok {
			out = append(out, sym)
		}
	}
	return out
}
