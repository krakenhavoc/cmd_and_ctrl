package gamestats

// deckfacts.go is ADR 0217 §3's facts read from a deck list rather
// than from games (slice 5 of #4880): composition by card type, the
// mana curve, the average mana value and the colour share, plus §5's
// version hash. Everything here is pure: the caller resolves the list
// against the card index and passes the cards in.

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// CardTypes is the order composition lists card types in. A card
// counts once under each type its front face has, so an artifact
// creature counts as both.
var CardTypes = []string{"land", "creature", "artifact", "enchantment", "instant", "sorcery", "planeswalker", "battle"}

// CurveTop is the last bucket of the mana curve: mana value 7 and up.
const CurveTop = 7

// DeckCard is one line of a resolved deck list: the card and how many
// copies, with commanders marked. The sideboard, a companion,
// Attractions and sticker sheets are not deck cards and are left out
// by the caller.
type DeckCard struct {
	Card      cards.Card
	Count     int
	Commander bool
}

// TypeCount is how many cards (copies) of a deck have one card type.
type TypeCount struct {
	Type  string `json:"type"`
	Cards int    `json:"cards"`
}

// ColourShare is one colour's share of the coloured mana symbols in a
// deck's non-land costs. Symbols is fractional: a hybrid symbol counts
// half to each of its colours. Share is nil when the deck has no
// coloured symbols at all.
type ColourShare struct {
	Colour  string   `json:"colour"`
	Symbols float64  `json:"symbols"`
	Share   *float64 `json:"share"`
}

// DeckFacts is what ADR 0217 reads from a deck list.
type DeckFacts struct {
	// Hash is §5's version hash of the main deck and commanders, ""
	// when no card resolved.
	Hash string `json:"hash"`
	// Cards is the copies counted (main deck and commanders) and
	// Unresolved the copies whose name the card index does not know,
	// which no other number here includes.
	Cards      int `json:"cards"`
	Unresolved int `json:"unresolved"`
	// Types is every card type in CardTypes order, zeros included.
	Types []TypeCount `json:"types"`
	// RolesAvailable is false while the server has no card roles
	// (ADR 0217 §7, slice 4): composition is then card types only.
	RolesAvailable bool `json:"roles_available"`
	// ManaCurve counts non-land cards at mana value 0, 1, … 6 and 7+
	// (CurveTop + 1 buckets). X counts as 0.
	ManaCurve []int `json:"mana_curve"`
	// NonLand is the non-land copies, commanders included.
	NonLand int `json:"nonland"`
	// AvgManaValue is the mean mana value of the non-land copies, nil
	// for a deck with none.
	AvgManaValue *float64 `json:"avg_mana_value"`
	// ColourShare is W, U, B, R and G in that order.
	ColourShare []ColourShare `json:"colour_share"`
	// ColourIdentity is the union of the commanders' colour identities,
	// in WUBRG order.
	ColourIdentity []string `json:"colour_identity"`
}

// wubrg is the colour order every list here uses.
var wubrg = []string{"W", "U", "B", "R", "G"}

// FactsOf computes a deck's facts from its resolved cards. unresolved
// is the copies the caller could not resolve.
func FactsOf(list []DeckCard, unresolved int) DeckFacts {
	out := DeckFacts{Unresolved: unresolved, ManaCurve: make([]int, CurveTop+1),
		Types: make([]TypeCount, len(CardTypes)), ColourIdentity: []string{}}
	for i, t := range CardTypes {
		out.Types[i].Type = t
	}
	symbols := map[string]float64{}
	identity := map[string]bool{}
	oracle := map[string]int{}
	mvSum := 0
	for _, dc := range list {
		if dc.Count < 1 {
			continue
		}
		c := dc.Card
		out.Cards += dc.Count
		if c.OracleID != uuid.Nil {
			oracle[c.OracleID.String()] += dc.Count
		}
		if dc.Commander {
			for _, l := range c.ColorIdentity {
				identity[strings.ToUpper(l)] = true
			}
		}
		front := strings.ToLower(FrontTypeLine(c))
		for i, t := range CardTypes {
			if strings.Contains(front, t) {
				out.Types[i].Cards += dc.Count
			}
		}
		if IsLandCard(c) {
			continue
		}
		mv := ManaValue(c)
		out.NonLand += dc.Count
		mvSum += mv * dc.Count
		out.ManaCurve[min(mv, CurveTop)] += dc.Count
		colourSymbols(castCost(c), func(letter string, weight float64) {
			symbols[letter] += weight * float64(dc.Count)
		})
	}
	if out.NonLand > 0 {
		v := float64(mvSum) / float64(out.NonLand)
		out.AvgManaValue = &v
	}
	total := 0.0
	for _, l := range wubrg {
		total += symbols[l]
	}
	for _, l := range wubrg {
		cs := ColourShare{Colour: l, Symbols: symbols[l]}
		if total > 0 {
			v := symbols[l] / total
			cs.Share = &v
		}
		out.ColourShare = append(out.ColourShare, cs)
		if identity[l] {
			out.ColourIdentity = append(out.ColourIdentity, l)
		}
	}
	out.Hash = DeckHash(oracle)
	return out
}

// FrontTypeLine is the type line of a card's front face: the part
// before " // " of a multi-faced card's line, or its first face's.
func FrontTypeLine(c cards.Card) string {
	if c.TypeLine != "" {
		front, _, _ := strings.Cut(c.TypeLine, " // ")
		return front
	}
	if len(c.CardFaces) > 0 {
		return c.CardFaces[0].TypeLine
	}
	return ""
}

// IsLandCard reports a card whose front face is a land. A modal
// double-faced card with a land back is a spell here, as it is on the
// mana curve: its front can be cast.
func IsLandCard(c cards.Card) bool { return isLandFace(FrontTypeLine(c)) }

// IsBasicLand reports a basic land (any "Basic" supertype), which
// "cards in the most decks" leaves out.
func IsBasicLand(c cards.Card) bool {
	front := FrontTypeLine(c)
	return strings.Contains(front, "Basic") && isLandFace(front)
}

// castCost is the mana cost a card is cast for: the printed cost, or
// for a card whose cost sits on its faces (a transforming or modal
// double-faced card), its front face's. A split card's printed cost
// carries both halves ("{1}{R} // {2}{U}").
func castCost(c cards.Card) string {
	if c.ManaCost != "" {
		return c.ManaCost
	}
	if len(c.CardFaces) > 0 {
		return c.CardFaces[0].ManaCost
	}
	return ""
}

// ManaValue is a card's mana value from its cost, X as 0. A split
// card's is the total of its halves; any other multi-faced card's is
// its front face's.
func ManaValue(c cards.Card) int {
	cost := castCost(c)
	if c.Layout == "split" {
		mv := 0
		for _, half := range strings.Split(cost, " // ") {
			mv += parseCost(half).mv
		}
		return mv
	}
	return parseCost(cost).mv
}

// colourSymbols calls add for each coloured mana symbol in a cost
// string, every face of it: a one-colour symbol (a Phyrexian or {2/W}
// one included) adds 1 to its colour, and a hybrid adds an equal part
// to each of its colours. Generic, colourless, X and snow add nothing.
func colourSymbols(cost string, add func(letter string, weight float64)) {
	for {
		open := strings.IndexByte(cost, '{')
		if open < 0 {
			return
		}
		end := strings.IndexByte(cost[open:], '}')
		if end < 0 {
			return
		}
		sym := strings.ToUpper(cost[open+1 : open+end])
		cost = cost[open+end+1:]
		var letters []string
		for _, p := range strings.Split(sym, "/") {
			switch p {
			case "W", "U", "B", "R", "G":
				letters = append(letters, p)
			}
		}
		for _, l := range letters {
			add(l, 1/float64(len(letters)))
		}
	}
}
