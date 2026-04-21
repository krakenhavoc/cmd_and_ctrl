package effects

import "fmt"

// cost.go parses Scryfall-style mana-cost strings into a ParsedCost
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
// hybrid alternative (Reaper King). Both are scaffolded for S17 and
// unused by S15's validator / auto-tapper.
type ColorRequirement struct {
	Options       []string
	Phyrexian     bool
	NumericAlt    int
	HasNumericAlt bool
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
