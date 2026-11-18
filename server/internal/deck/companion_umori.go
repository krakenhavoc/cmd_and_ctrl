package deck

import (
	"fmt"
	"sort"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// companion_umori.go — Umori, the Collector's companion condition (CR
// 702.139, ADR 0172 §9, S82 reprint pass #5908):
//
//	"Companion — Each nonland card in your starting deck shares a card
//	 type."
//
// Read as the printed rulings do: one card type that every nonland card in
// the starting deck has. A deck with no nonland card meets it trivially.

// umoriOracleID is Umori's oracle ID.
const umoriOracleID = "913383a1-9c0d-40a1-8c5a-b0a8a95c09a6"

func init() {
	companionConditions[umoriOracleID] = companionCondition{
		Text:  "each nonland card in your starting deck must share a card type",
		Unmet: umoriUnmet,
	}
}

// cardTypeWords are the card types a type line can carry before its dash
// (CR 205.2a). Supertypes and subtypes are not card types.
var cardTypeWords = map[string]string{
	"artifact": "Artifact", "battle": "Battle", "creature": "Creature", "enchantment": "Enchantment",
	"instant": "Instant", "kindred": "Kindred", "tribal": "Kindred", "land": "Land",
	"planeswalker": "Planeswalker", "sorcery": "Sorcery",
}

// companionCardTypes is the set of card types a card has outside the game,
// as a companion condition sees it:
//
//   - a split card has each card type on either of its halves (CR 709.4c);
//   - a double-faced card has only its front face's (CR 712.8a);
//   - an adventurer has only its normal characteristics (CR 715.4).
//
// Scryfall joins a multi-face card's type lines with " // ".
func companionCardTypes(c cards.Card) map[string]bool {
	lines := []string{c.TypeLine}
	if c.Layout == "split" && len(c.CardFaces) > 1 {
		lines = lines[:0]
		for _, f := range c.CardFaces {
			lines = append(lines, f.TypeLine)
		}
	} else if c.TypeLine == "" && len(c.CardFaces) > 0 {
		lines = []string{c.CardFaces[0].TypeLine}
	}
	out := map[string]bool{}
	for _, line := range lines {
		if c.Layout != "split" {
			line, _, _ = strings.Cut(line, " // ")
		}
		line, _, _ = strings.Cut(line, "—")
		for _, w := range strings.Fields(line) {
			if t, ok := cardTypeWords[strings.ToLower(w)]; ok {
				out[t] = true
			}
		}
	}
	return out
}

// umoriUnmet intersects the card types of every nonland card. It names the
// cards that share nothing with the type most of them have, which is the
// shortest list of cards to cut.
func umoriUnmet(starting []cards.Card, _ game.FormatSpec) string {
	var nonland []cards.Card
	for _, c := range starting {
		if !companionCardTypes(c)["Land"] {
			nonland = append(nonland, c)
		}
	}
	if len(nonland) == 0 {
		return ""
	}
	common := companionCardTypes(nonland[0])
	for _, c := range nonland[1:] {
		have := companionCardTypes(c)
		for t := range common {
			if !have[t] {
				delete(common, t)
			}
		}
	}
	if len(common) > 0 {
		return ""
	}
	// Count each type to point at the nearest fix.
	counts := map[string]int{}
	for _, c := range nonland {
		for t := range companionCardTypes(c) {
			counts[t]++
		}
	}
	best, bestN := "", 0
	var types []string
	for t := range counts {
		types = append(types, t)
	}
	sort.Strings(types)
	for _, t := range types {
		if counts[t] > bestN {
			best, bestN = t, counts[t]
		}
	}
	if best == "" {
		return "no card type is shared by every nonland card"
	}
	var off []string
	seen := map[string]bool{}
	for _, c := range nonland {
		if !companionCardTypes(c)[best] && !seen[c.Name] {
			seen[c.Name] = true
			off = append(off, c.Name)
		}
	}
	if len(off) > maxCompanionOffenders {
		off = off[:maxCompanionOffenders]
	}
	return fmt.Sprintf("no card type is shared by every nonland card; %d of them are %s cards, and these are not: %s",
		bestN, best, strings.Join(off, ", "))
}
