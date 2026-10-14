package deck

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// companion.go — the companion deck check (CR 702.139, ADR 0172 §2).
//
// A deck may reveal one companion: a card from outside the game whose
// companion condition its starting deck fulfils (CR 103.2b). The
// condition is printed on the companion, and each one is a different
// rule about the deck's printed characteristics, so each is code. The
// registry below is keyed by oracle ID and is ALSO the list of
// companions the server supports: a companion with no entry is refused
// as `unsupported_mechanic` rather than seated on trust.
//
// "Starting deck" is the commanders plus the mainboard (CR 702.139b:
// in Commander, before the commander is set aside). The companion is
// never part of it.
//
// The conditions read Scryfall data only, the same record the colour
// identity check reads. This package can't import the effects catalog
// (see deckcoverage's package comment), and a deck condition is about
// printed characteristics, not about what the engine automates.

// companionCondition is one companion's deck-building condition.
type companionCondition struct {
	// Text is the condition as a player reads it, for the violation
	// message.
	Text string
	// Offenders returns the names of the starting-deck cards that
	// break the condition, in deck order. Empty means fulfilled.
	Offenders func(starting []cards.Card) []string
}

// companionConditions is the registry. One entry per supported
// companion, added with its card (ADR 0172 §9 lists the next nine).
var companionConditions = map[string]companionCondition{
	// Jegantha, the Wellspring: "No card in your starting deck has more
	// than one of the same mana symbol in its mana cost."
	"f0587c55-06c1-4930-a62e-668f37464ce9": {
		Text:      "no card in your starting deck may have more than one of the same mana symbol in its mana cost",
		Offenders: offendersEach(repeatsAManaSymbol),
	},
}

// maxCompanionOffenders caps how many offending cards one failed
// condition lists, one violation row each (ADR 0172 §2).
const maxCompanionOffenders = 10

// offendersEach turns a per-card test into an Offenders function.
func offendersEach(breaks func(cards.Card) bool) func([]cards.Card) []string {
	return func(starting []cards.Card) []string {
		var out []string
		seen := map[string]bool{}
		for _, c := range starting {
			if seen[c.Name] || !breaks(c) {
				continue
			}
			seen[c.Name] = true
			out = append(out, c.Name)
		}
		return out
	}
}

// manaSymbolRE matches one mana symbol in Scryfall brace notation.
var manaSymbolRE = regexp.MustCompile(`\{[^}]+\}`)

// repeatsAManaSymbol is Jegantha's test: does any one symbol appear
// twice in the card's mana cost, compared exactly as printed? {R/G}{R/G}
// and {X}{X}{R} repeat; {R}{R/G} does not (Scryfall rulings,
// 2020-04-17). The cost is the card's characteristics outside the game
// (companionManaCosts).
func repeatsAManaSymbol(c cards.Card) bool {
	seen := map[string]bool{}
	for _, cost := range companionManaCosts(c) {
		for _, sym := range manaSymbolRE.FindAllString(cost, -1) {
			if seen[sym] {
				return true
			}
			seen[sym] = true
		}
	}
	return false
}

// companionManaCosts is the mana cost a companion condition sees for a
// card in a deck, which is outside the game, as one or more printed
// costs whose symbols are counted together:
//
//   - a split card is both halves combined, but "an effect that refers
//     specifically to the symbols in a split card's mana cost sees the
//     separate symbols" (CR 709.4b), so Fire // Ice's {1}{R} and {1}{U}
//     hold {1} twice;
//   - a double-faced card outside the game has only its front face's
//     characteristics (CR 712.8a);
//   - an adventurer card has only its normal characteristics (CR
//     715.4), and a flip card's mana cost is its top half's (CR 710.1c);
//   - a card with no mana cost (a land) has no symbols.
//
// Scryfall writes a split or adventurer card's top-level cost as the
// halves joined by " // " and leaves a double-faced card's empty, with
// the cost on each face. So: a split card takes every face, anything
// else takes the top-level cost before any " // ", or the front face's.
func companionManaCosts(c cards.Card) []string {
	if c.Layout == "split" && len(c.CardFaces) > 1 {
		out := make([]string, 0, len(c.CardFaces))
		for _, f := range c.CardFaces {
			out = append(out, f.ManaCost)
		}
		return out
	}
	if c.ManaCost != "" {
		front, _, _ := strings.Cut(c.ManaCost, " // ")
		return []string{front}
	}
	if len(c.CardFaces) > 0 {
		return []string{c.CardFaces[0].ManaCost}
	}
	return nil
}

// companionViolations judges the list's companion (ADR 0172 §2). Nil
// when the list has none.
func companionViolations(list *List) []Violation {
	var vs []Violation
	for _, name := range list.extraCompanions {
		vs = append(vs, Violation{
			Code:    CodeTooManyCompanions,
			Card:    name,
			Message: fmt.Sprintf("a deck can reveal only one companion; %q is a second", name),
		})
	}
	c := list.Companion
	if c == nil {
		return vs
	}
	if list.CompanionFromSideboard {
		vs = append(vs, Violation{
			Code:    CodeCompanionFromSideboard,
			Card:    c.Name,
			Message: fmt.Sprintf("%q was the only card in the sideboard, so it is read as your companion; put it under a Companion heading to say so", c.Name),
		})
	}
	if !printsCompanion(*c) {
		return append(vs, Violation{
			Code:    CodeNotACompanion,
			Card:    c.Name,
			Message: fmt.Sprintf("%q has no companion ability, so it can't be your companion", c.Name),
		})
	}
	for _, cmd := range list.Commanders {
		if cmd.OracleID == c.OracleID {
			vs = append(vs, Violation{
				Code:    CodeCompanionIsCommander,
				Card:    c.Name,
				Message: fmt.Sprintf("%q is your commander, so it can't also be your companion", c.Name),
			})
			break
		}
	}
	if legal := c.Legalities["commander"]; legal != "legal" {
		vs = append(vs, Violation{
			Code:    CodeCompanionNotLegal,
			Card:    c.Name,
			Message: fmt.Sprintf("%q is not legal in the commander format (status: %q)", c.Name, legal),
		})
	}
	if len(list.Commanders) == 1 || len(list.Commanders) == 2 {
		allowed := map[string]struct{}{}
		for _, cmd := range list.Commanders {
			for sym := range identitySet(cmd.ColorIdentity) {
				allowed[sym] = struct{}{}
			}
		}
		for _, sym := range c.ColorIdentity {
			if _, ok := allowed[sym]; !ok {
				vs = append(vs, Violation{
					Code:    CodeCompanionColorIdentity,
					Card:    c.Name,
					Message: fmt.Sprintf("companion %q has color identity %v but commander allows %v", c.Name, c.ColorIdentity, commanderIdentityList(allowed)),
				})
				break
			}
		}
	}
	cond, ok := companionConditions[c.OracleID.String()]
	if !ok {
		return append(vs, Violation{
			Code:    CodeUnsupportedMechanic,
			Card:    c.Name,
			Message: fmt.Sprintf("%q can't be a companion here yet: its companion condition isn't checked by the server", c.Name),
		})
	}
	offenders := cond.Offenders(allCards(list))
	if len(offenders) > maxCompanionOffenders {
		offenders = offenders[:maxCompanionOffenders]
	}
	for _, name := range offenders {
		vs = append(vs, Violation{
			Code:    CodeCompanionConditionNotMet,
			Card:    name,
			Message: fmt.Sprintf("%q breaks %s's companion condition: %s", name, c.Name, cond.Text),
		})
	}
	return vs
}
