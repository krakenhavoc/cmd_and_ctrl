package game

import (
	"fmt"
	"strconv"
)

// mana_cost.go parses Scryfall-style mana-cost strings into a ParsedCost
// (moved from cards/effects/cost.go in S15 sub-PR 2 — `game` owns the
// parser now so Player.ManaPool helpers can consume ParsedCost
// without an import cycle through cards/effects, which itself
// depends on `game`).
// the S15 cost validator + auto-tapper can reason about. Regexp-free;
// walks the string token by token. ~80 LoC, single-pass.
//
// Syntax (what the parser accepts):
//
//	{N}       — generic mana, N is a non-negative integer (0 is legal for
//	            uncast-cost artefact records like Mox Diamond-adjacent printings;
//	            parses but contributes zero to ParsedCost.Generic).
//	{W|U|B|R|G|C}
//	          — colored / colorless mana. Case-insensitive; stored uppercase
//	            so the downstream validator matches ManaToken.Color.
//	{X}       — variable cost. Counted into ParsedCost.XSlots. The caller
//	            supplies the XValue announce-time; the auto-tapper solves
//	            for Generic + XSlots * XValue generic mana.
//	{W/U}     — hybrid. Parsed into a ColorRequirement whose Options set
//	            has both halves. Two-mana hybrids ({2/W}) and Phyrexian
//	            ({W/P}) share the same shape:
//	                - {N/W}:   Options = {"W"}, NumericAlt = N (pay-N-or-W).
//	                - {W/P}:   Options = {"W"}, Phyrexian = true (pay-W-or-2-life).
//	              S15 doesn't implement the life self-pay or the numeric-alt
//	              payment — auto-tapper treats these as the color half.
//	{S}       — snow. Parses but ParsedCost.HasSnow is informational only;
//	            S15 does not enforce "must be paid with snow sources."
//
// Unknown symbols return an error rather than silently succeeding.

// ParsedCost is the result of parsing a Scryfall mana-cost string. Pass
// to the validator / auto-tapper by value; it is not mutated in place.
type ParsedCost struct {
	// Generic is the total generic-mana requirement (the sum of all
	// {N} tokens). {X} is tracked separately in XSlots.
	Generic int

	// Required is one ColorRequirement per colored-mana slot in the
	// cost, in the order they appear in the string. A {1}{R}{R} cost
	// produces two entries in Required, both with Options = {"R"}.
	Required []ColorRequirement

	// XSlots is the count of {X} tokens (always 0 or 1 for real
	// printings, but the parser admits N — deduplication is the
	// caller's job). The cost solver multiplies this by the announced
	// XValue to produce the total generic-mana demand.
	XSlots int

	// HasPhyrexian is true if any {W/P}-style token was parsed. The
	// cost validator records this for the "pay 2 life instead"
	// affordance S17 will wire; S15 always pays the mana half.
	HasPhyrexian bool

	// HasSnow is true if any {S} token was parsed. S15 treats snow
	// sources as identical to generic sources — the flag is
	// informational for S17's snow-routing work.
	HasSnow bool
}

// ColorRequirement is one colored-mana slot. Options is the set of
// colors that may satisfy it; a monocolored {R} has Options = {"R"},
// a hybrid {W/U} has Options = {"W", "U"}. Phyrexian flags the
// "or 2 life" alternative; NumericAlt flags the "{N/COLOR}" two-mana
// hybrid alternative (Reaper King). The validator and auto-tapper do
// not read NumericAlt — they pay the coloured half — but the mana
// value does (CR 202.3f, ColorRequirement.ManaValue).
type ColorRequirement struct {
	Options       []string
	Phyrexian     bool
	NumericAlt    int
	HasNumericAlt bool
}

// ManaValue is what one coloured-mana slot contributes to a mana
// value: the largest component of the symbol (CR 202.3f). That is 1
// for {W}, {C}, snow, a two-colour or colourless hybrid ({W/U},
// {C/W}) and a Phyrexian symbol (CR 202.3g), and N for a monocoloured
// hybrid {N/W} — 2 for every printed one.
func (r ColorRequirement) ManaValue() int {
	if r.HasNumericAlt && r.NumericAlt > 1 {
		return r.NumericAlt
	}
	return 1
}

// ParseCost walks s and produces a ParsedCost. Returns an error on
// any malformed token (unbalanced braces, unknown symbol, non-numeric
// generic). Empty input is not an error — it parses to the zero
// ParsedCost (matches land behaviour).
func ParseCost(s string) (ParsedCost, error) {
	var out ParsedCost
	i := 0
	for i < len(s) {
		if s[i] == ' ' || s[i] == '\t' {
			i++
			continue
		}
		if s[i] != '{' {
			return ParsedCost{}, fmt.Errorf("cost: expected '{' at position %d in %q", i, s)
		}
		// Find the matching '}'.
		end := i + 1
		for end < len(s) && s[end] != '}' {
			end++
		}
		if end >= len(s) {
			return ParsedCost{}, fmt.Errorf("cost: unterminated '{' at position %d in %q", i, s)
		}
		token := s[i+1 : end]
		if err := absorbToken(token, &out); err != nil {
			return ParsedCost{}, fmt.Errorf("cost: %w (token %q in %q)", err, token, s)
		}
		i = end + 1
	}
	return out, nil
}

// absorbToken mutates out with the effect of one "{…}" token body.
func absorbToken(token string, out *ParsedCost) error {
	if token == "" {
		return fmt.Errorf("empty token")
	}
	// Uppercase ASCII in-place (single pass).
	buf := make([]byte, len(token))
	for i := 0; i < len(token); i++ {
		b := token[i]
		if b >= 'a' && b <= 'z' {
			b -= 'a' - 'A'
		}
		buf[i] = b
	}
	u := string(buf)

	// Pure color / colorless / X / S.
	if len(u) == 1 {
		switch u[0] {
		case 'W', 'U', 'B', 'R', 'G', 'C':
			out.Required = append(out.Required, ColorRequirement{Options: []string{u}})
			return nil
		case 'X':
			out.XSlots++
			return nil
		case 'S':
			out.HasSnow = true
			out.Required = append(out.Required, ColorRequirement{Options: []string{"C"}})
			return nil
		}
	}

	// Generic integer: "0", "1", "12", "16", "100"? All digits.
	if allDigits(u) {
		n := 0
		for i := 0; i < len(u); i++ {
			n = n*10 + int(u[i]-'0')
		}
		out.Generic += n
		return nil
	}

	// Hybrid forms — "{W/U}", "{2/W}", "{W/P}".
	if len(u) == 3 && u[1] == '/' {
		left, right := u[0], u[2]
		// Phyrexian: colored / P.
		if right == 'P' && isColor(left) {
			out.Required = append(out.Required, ColorRequirement{
				Options:   []string{string(left)},
				Phyrexian: true,
			})
			out.HasPhyrexian = true
			return nil
		}
		// {N/COLOR}: numeric-alt hybrid (Reaper King's printings).
		if left >= '0' && left <= '9' && isColor(right) {
			out.Required = append(out.Required, ColorRequirement{
				Options:       []string{string(right)},
				NumericAlt:    int(left - '0'),
				HasNumericAlt: true,
			})
			return nil
		}
		// Plain hybrid: two colors or a color + colorless.
		if isColor(left) && isColor(right) {
			out.Required = append(out.Required, ColorRequirement{
				Options: []string{string(left), string(right)},
			})
			return nil
		}
	}

	return fmt.Errorf("unknown token")
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func isColor(b byte) bool {
	switch b {
	case 'W', 'U', 'B', 'R', 'G', 'C':
		return true
	}
	return false
}

// ProducedManaEntry is one "slot" parsed out of a ManaAbility.Produced
// string. Options lists the colors the controller may pick from (one
// entry for `"{W}"`, one per color for `"{W|U|B|R|G}"`). A slot with
// multiple Options is materialised into a PendingChoiceMana at
// activation time; a single-option slot drops straight into the pool.
type ProducedManaEntry struct {
	Options []string

	// Amounts is how many mana the slot adds for each option, for
	// "N mana of any one color" (#742): one pick, N tokens of the
	// picked colour. nil means one of whichever colour is picked, which
	// is every slot the grammar had before. Only a MULTI-option slot
	// carries it — a single-option "{G3}" is expanded by the parser
	// into three ordinary {G} slots, so nothing downstream of the
	// parser has to learn about amounts for the common case.
	Amounts map[string]int
}

// AmountFor is how many mana this slot adds when `color` is picked.
func (e ProducedManaEntry) AmountFor(color string) int {
	if n, ok := e.Amounts[color]; ok {
		return n
	}
	return 1
}

// OneColorAmounts reports whether the slot is a "N mana of any one
// color" pick — several options, at least one adding more than one
// mana. The auto-tapper plans around such a slot (see
// gatherTapSources) because its tokens must all be one colour, which
// the planner's one-slot-one-mana model cannot promise.
func (e ProducedManaEntry) OneColorAmounts() bool {
	return len(e.Options) > 1 && len(e.Amounts) > 0
}

// ParseProducedMana walks a produced-mana declaration (the
// Spec.ManaAbilities.Produced string) and returns one
// ProducedManaEntry per colored brace. Syntax is a superset of
// ParseCost — adds the pipe operator to express "controller picks
// one of these colors":
//
//	"{C}{C}"          → two colorless slots (Sol Ring)
//	"{G}"             → one green slot (basic Forest synthetic)
//	"{W|U|B|R|G}"     → one any-color slot (Birds of Paradise)
//	"{W3|U3|B3|R3|G3}" → one pick adding three mana of the picked
//	                     colour (Gilded Lotus, #742)
//	"{G2|U5}"         → per-option amounts (Nyx Lotus's devotion)
//	"{G3}"            → expanded to three {G} slots
//	"{G0|U2}"         → a zero-amount option is dropped, so this is
//	                     two {U} slots
//
// A count follows the colour letter inside the brace. It is a
// produced-mana extension only — ParseCost has no such form, since a
// cost's "{2}" already means something else.
//
// Generic / X / phyrexian / snow tokens aren't meaningful in produced
// mana and return an error. Empty input parses to a nil slice — a
// "produces nothing" ability, which isn't expected in live catalog
// but is useful for tests.
func ParseProducedMana(s string) ([]ProducedManaEntry, error) {
	var out []ProducedManaEntry
	i := 0
	for i < len(s) {
		if s[i] == ' ' || s[i] == '\t' {
			i++
			continue
		}
		if s[i] != '{' {
			return nil, fmt.Errorf("produced mana: expected '{' at position %d in %q", i, s)
		}
		end := i + 1
		for end < len(s) && s[end] != '}' {
			end++
		}
		if end >= len(s) {
			return nil, fmt.Errorf("produced mana: unterminated '{' at position %d in %q", i, s)
		}
		token := s[i+1 : end]
		if token == "" {
			return nil, fmt.Errorf("produced mana: empty token in %q", s)
		}
		// Uppercase + pipe-split.
		buf := make([]byte, len(token))
		for j := 0; j < len(token); j++ {
			b := token[j]
			if b >= 'a' && b <= 'z' {
				b -= 'a' - 'A'
			}
			buf[j] = b
		}
		u := string(buf)
		var options []string
		amounts := map[string]int{}
		counted := false
		start := 0
		for j := 0; j <= len(u); j++ {
			if j == len(u) || u[j] == '|' {
				seg := u[start:j]
				start = j + 1
				if len(seg) == 0 || !isColor(seg[0]) {
					return nil, fmt.Errorf("produced mana: unknown token %q in %q", seg, s)
				}
				n := 1
				if len(seg) > 1 {
					if !allDigits(seg[1:]) {
						return nil, fmt.Errorf("produced mana: unknown token %q in %q", seg, s)
					}
					v, err := strconv.Atoi(seg[1:])
					if err != nil {
						return nil, fmt.Errorf("produced mana: bad amount in %q: %w", s, err)
					}
					n, counted = v, true
				}
				if n <= 0 {
					continue
				}
				color := seg[:1]
				if _, dup := amounts[color]; dup {
					// A derived pipe that names a colour twice
					// offers it once.
					continue
				}
				options = append(options, color)
				amounts[color] = n
			}
		}
		i = end + 1
		switch {
		case len(options) == 0:
			// Every option counted zero: the slot adds nothing.
		case len(options) == 1:
			// "{G3}" is three ordinary {G} slots.
			for k := 0; k < amounts[options[0]]; k++ {
				out = append(out, ProducedManaEntry{Options: []string{options[0]}})
			}
		default:
			entry := ProducedManaEntry{Options: options}
			if counted {
				for _, c := range options {
					if amounts[c] != 1 {
						entry.Amounts = amounts
						break
					}
				}
			}
			out = append(out, entry)
		}
	}
	return out, nil
}
