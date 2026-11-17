package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b15_helpers.go — shared helpers for The Brothers' War batch 15.

// s82BroTotalValueRuleMax is the largest total-mana-value ceiling that has
// its own registered set rule.
const s82BroTotalValueRuleMax = 24

// s82BroTotalValueRules are the registered set rules "with total mana value
// N or less" for N = 0..s82BroTotalValueRuleMax (index = N), plus one
// unbounded rule at the end. A set rule is a function of the picked cards
// alone (RegisterPutRule), so a ceiling that is only known at resolution
// (Smelting Vat's "the sacrificed artifact's mana value") picks the rule
// registered for that number. Each key is on disk and never renamed.
var s82BroTotalValueRules = func() []PutRule {
	rules := make([]PutRule, 0, s82BroTotalValueRuleMax+2)
	for n := 0; n <= s82BroTotalValueRuleMax; n++ {
		rules = append(rules, RegisterPutRule(
			fmt.Sprintf("choose-cards/put-from-library-s82-bro-total-mana-value-%d", n),
			b23TotalManaValueAtMost(n)))
	}
	return append(rules, RegisterPutRule("choose-cards/put-from-library-s82-bro-total-mana-value-unbounded",
		func([]game.Card) bool { return true }))
}()

// s82BroTotalValueRule is the registered rule for "total mana value
// `limit` or less" over the looked-at cards. A ceiling no set of the
// looked-at cards can reach (at most `picks` of them are put) is the same
// as no ceiling, so it is clamped to the largest sum there is before the
// rule is picked; a clamp above the registered range is unbounded.
func s82BroTotalValueRule(g *game.Game, looked []uuid.UUID, limit, picks int) PutRule {
	var values []int
	for _, id := range looked {
		if c, ok := g.LookupCardForEffect(id); ok {
			values = append(values, c.ManaValue())
		}
	}
	// The `picks` largest values, summed.
	sum := 0
	for i := 0; i < picks; i++ {
		best := -1
		for j, v := range values {
			if v >= 0 && (best < 0 || v > values[best]) {
				best = j
			}
		}
		if best < 0 {
			break
		}
		sum += values[best]
		values[best] = -1
	}
	limit = min(max(limit, 0), sum)
	if limit > s82BroTotalValueRuleMax {
		return s82BroTotalValueRules[len(s82BroTotalValueRules)-1]
	}
	return s82BroTotalValueRules[limit]
}
