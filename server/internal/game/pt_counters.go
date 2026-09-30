package game

// pt_counters.go — every power/toughness counter kind, not just two
// (#1664, CR 122.1a).
//
// CR 122.1a: "A counter on a permanent that reads +X/+Y ... adds X to
// its power and Y to its toughness", and the same for -X/-Y. The
// engine used to read exactly two such kinds — +1/+1 and -1/-1 — so a
// -2/-1 counter from Contagion, a +1/+0 from Dwarven Armorer or a
// -0/-1 paid into Wall of Roots was stored on the card, drawn as a pip,
// and changed nothing. ParsePTCounter is the one place that decides
// whether a counter name is a P/T counter and what it is worth;
// PTCounterDelta sums a whole counter map through it, and
// Card.PowerForComparison / Card.CurrentToughness (the CR 613.4c
// counter step, which this engine applies after the layer system
// rather than inside it) read that sum.
//
// What does NOT change, deliberately:
//
//   - CR 704.5q's annihilation is about +1/+1 and -1/-1 counters and
//     nothing else. A +1/+0 and a -1/-0 on one creature stay; the SBA in
//     stateBasedActionsLocked still names the two kinds it cancels.
//   - "+1/+1 counters" printed on a card means that one kind. Hardened
//     Scales, Branching Evolution, "the number of +1/+1 counters on it"
//     and every other kind-specific reader keep naming CounterPlusOne.

import "strconv"

// ParsePTCounter reports whether name is a power/toughness counter —
// "[+-]N/[+-]M", both signs present, N and M plain decimal digits —
// and the power and toughness it adds. "+1/+0" is (1, 0, true),
// "-2/-1" is (-2, -1, true), "-0/-1" is (0, -1, true). Anything else
// — "loyalty", "charge", "1/1", "+1/+1 " with trailing space, "+X/+X"
// — is (0, 0, false), so a homebrew or unknown counter name is never
// mistaken for a stat change.
func ParsePTCounter(name string) (power, toughness int, ok bool) {
	// Fast path for the two kinds every board carries.
	switch name {
	case CounterPlusOne:
		return 1, 1, true
	case CounterMinusOne:
		return -1, -1, true
	}
	slash := -1
	for i := 0; i < len(name); i++ {
		if name[i] == '/' {
			if slash >= 0 {
				return 0, 0, false
			}
			slash = i
		}
	}
	if slash < 0 {
		return 0, 0, false
	}
	p, okP := parseSignedPTHalf(name[:slash])
	t, okT := parseSignedPTHalf(name[slash+1:])
	if !okP || !okT {
		return 0, 0, false
	}
	return p, t, true
}

// parseSignedPTHalf parses one side of a P/T counter name: a mandatory
// '+' or '-' followed by one to three decimal digits. The digit bound
// keeps a pathological name from overflowing anything downstream; no
// printed counter is wider than one digit.
func parseSignedPTHalf(s string) (int, bool) {
	if len(s) < 2 || len(s) > 4 {
		return 0, false
	}
	sign := 1
	switch s[0] {
	case '+':
	case '-':
		sign = -1
	default:
		return 0, false
	}
	digits := s[1:]
	for i := 0; i < len(digits); i++ {
		if digits[i] < '0' || digits[i] > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(digits)
	if err != nil {
		return 0, false
	}
	return sign * n, true
}

// PTCounterDelta is the power and toughness a counter map adds, summed
// over every P/T counter kind on it (CR 122.1a). Non-P/T kinds and
// non-positive counts contribute nothing.
func PTCounterDelta(counters map[string]int) (power, toughness int) {
	for name, n := range counters {
		if n <= 0 {
			continue
		}
		p, t, ok := ParsePTCounter(name)
		if !ok {
			continue
		}
		power += p * n
		toughness += t * n
	}
	return power, toughness
}
