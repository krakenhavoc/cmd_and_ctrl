package game

import "strconv"

// ManaRestrictColorCountExactly builds "spend this mana only to cast spells
// with exactly N colors" (Meeting of the Five, #5491, CR 105.2): a clause
// about a SPELL, judged on the colours the spell being cast has (a hybrid
// spell is every colour of its symbols, CR 202.2d). It refuses an
// activation, an unlock and a payment with no stated purpose, which have no
// spell to count colours of (weaker than printed, never stronger).
// ManaSpendContext.Colors is what it reads. The tag is additive: an older
// snapshot never carries it.
func ManaRestrictColorCountExactly(n int) string {
	return colorCountKey + ":" + strconv.Itoa(n)
}

// colorCountKey is the tag key of ManaRestrictColorCountExactly.
const colorCountKey = "colorcount"

// matchesColorCount decides one ManaRestrictColorCountExactly tag.
func (ctx ManaSpendContext) matchesColorCount(value string) bool {
	if ctx.Purpose != SpendPurposeCast {
		return false
	}
	n, err := strconv.Atoi(value)
	return err == nil && distinctColorCount(ctx.Colors) == n
}
