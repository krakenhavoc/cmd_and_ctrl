package gamestats

// mana.go is ADR 0217 §4's colour check: a spell's cost as coloured
// symbols, and whether the seat's mana sources can pay them, one source
// a symbol, by a small bipartite matching.

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// CardIndex is what Compute reads from the Scryfall card index: a
// printing by its Scryfall id. *cards.Index satisfies it. A nil index
// leaves the deck hash empty and reads mana sources off the view.
type CardIndex interface {
	Get(id uuid.UUID) (cards.Card, bool)
}

// colours is a set of the mana letters a source can make or a symbol
// accepts: W, U, B, R, G and C (colourless).
type colours uint8

const (
	colW colours = 1 << iota
	colU
	colB
	colR
	colG
	colC
)

func colourOf(letter string) colours {
	switch strings.ToUpper(strings.TrimSpace(letter)) {
	case "W":
		return colW
	case "U":
		return colU
	case "B":
		return colB
	case "R":
		return colR
	case "G":
		return colG
	case "C":
		return colC
	}
	return 0
}

// cost is a parsed mana cost.
type cost struct {
	// mv is the mana value, X as 0.
	mv int
	// pips are the symbols that need a source of a colour: one entry
	// per symbol, the colours it accepts (two for a hybrid). Generic,
	// Phyrexian and {2/W} symbols are not here: generic is paid from the
	// sources left, a Phyrexian symbol can be paid with life, and a
	// {2/W} symbol with two generic.
	pips []colours
}

// parseCost reads a Scryfall mana cost ("{2}{G}{G/U}{W/P}{X}"). A split
// or modal card's cost is its first face's (the part before " // ").
func parseCost(s string) cost {
	if i := strings.Index(s, " // "); i >= 0 {
		s = s[:i]
	}
	var c cost
	for {
		open := strings.IndexByte(s, '{')
		if open < 0 {
			return c
		}
		end := strings.IndexByte(s[open:], '}')
		if end < 0 {
			return c
		}
		sym := strings.ToUpper(s[open+1 : open+end])
		s = s[open+end+1:]
		if n, err := strconv.Atoi(sym); err == nil {
			c.mv += n
			continue
		}
		switch sym {
		case "X", "Y", "Z":
			continue
		case "S", "H", "HW", "HR":
			c.mv++ // snow and half symbols: paid like generic here
			continue
		}
		parts := strings.Split(sym, "/")
		phyrexian := false
		var accept colours
		twobrid := false
		for _, p := range parts {
			switch p {
			case "P":
				phyrexian = true
			case "2":
				twobrid = true
			default:
				accept |= colourOf(p)
			}
		}
		switch {
		case twobrid:
			c.mv += 2
		case phyrexian:
			c.mv++
		case accept != 0:
			c.mv++
			c.pips = append(c.pips, accept)
		default:
			c.mv++
		}
	}
}

// payable reports whether every pip can be matched to a different
// source that makes one of its colours (Kuhn's augmenting paths; a few
// pips against at most a few dozen sources).
func payable(pips []colours, sources []colours) bool {
	if len(pips) > len(sources) {
		return false
	}
	owner := make([]int, len(sources)) // source → pip index + 1; 0 is free
	var try func(p int, seen []bool) bool
	try = func(p int, seen []bool) bool {
		for s, made := range sources {
			if made&pips[p] == 0 || seen[s] {
				continue
			}
			seen[s] = true
			if owner[s] == 0 || try(owner[s]-1, seen) {
				owner[s] = p + 1
				return true
			}
		}
		return false
	}
	for p := range pips {
		if !try(p, make([]bool, len(sources))) {
			return false
		}
	}
	return true
}

// castable is ADR 0217 §4's check of one card against the sources:
// by amount (its mana value at most the number of sources) and by colour
// (its coloured symbols matched to different sources).
func castable(manaCost string, sources []colours) (byAmount, byColour bool) {
	c := parseCost(manaCost)
	byAmount = c.mv <= len(sources)
	byColour = byAmount && payable(c.pips, sources)
	return byAmount, byColour
}

// producedBy is the colours a permanent can make: the card index's
// ProducedMana for its printing, or else what its mana abilities on the
// view say they add. Zero when it makes no mana.
func producedBy(c *protocol.CardView, idx CardIndex) colours {
	var out colours
	if idx != nil && c.ScryfallID != "" {
		if id, err := uuid.Parse(c.ScryfallID); err == nil {
			if card, ok := idx.Get(id); ok && len(card.ProducedMana) > 0 {
				for _, l := range card.ProducedMana {
					out |= colourOf(l)
				}
				return out
			}
		}
	}
	for _, a := range c.ManaAbilities {
		if a.AddsNoMana {
			continue
		}
		out |= lettersIn(a.Produced)
		for _, opt := range a.ColorOptions {
			for _, l := range opt {
				out |= colourOf(l)
			}
		}
	}
	return out
}

// lettersIn reads the colours of mana symbols in a string like "{G}{G}"
// or "{W/U}".
func lettersIn(s string) colours {
	var out colours
	for _, r := range s {
		switch r {
		case 'W', 'U', 'B', 'R', 'G', 'C':
			out |= colourOf(string(r))
		}
	}
	return out
}

// isLandFace reports a type line with the Land type.
func isLandFace(typeLine string) bool { return strings.Contains(typeLine, "Land") }

// landKind is how ADR 0217 §4 counts a card in hand: a land (its face up
// is a land), a land only on another face (a modal double-faced card
// with a land back: a land for screw, neither for flood), or a spell.
type landKind int

const (
	spellCard landKind = iota
	landCard
	landBackCard
)

func landKindOf(c *protocol.CardView) landKind {
	front, _, _ := strings.Cut(c.TypeLine, " // ")
	if isLandFace(front) {
		return landCard
	}
	if isLandFace(c.TypeLine) {
		return landBackCard
	}
	for _, f := range c.Faces {
		if isLandFace(f.TypeLine) {
			return landBackCard
		}
	}
	return spellCard
}
