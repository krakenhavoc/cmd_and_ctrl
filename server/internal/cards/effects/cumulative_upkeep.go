package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cumulative_upkeep.go — CR 702.24, the keyword whose cost grows every
// turn (#567).
//
//	702.24a  Cumulative upkeep is a triggered ability that imposes an
//	         increasing cost on a permanent. "Cumulative upkeep [cost]"
//	         means "At the beginning of your upkeep, if this permanent
//	         is on the battlefield, put an age counter on this
//	         permanent, then sacrifice it unless you pay its upkeep
//	         cost for each age counter on it."
//
// One constructor, built out of primitives that already exist: an
// AtYourUpkeep trigger (triggers_common.go), the counter primitive
// (AddCounter), and the "unless that player pays" prompt (PayUnless,
// the same one Rhystic Study raises). Nothing about this keyword is a
// new pipeline — what it needed was a cost that is not fixed when the
// card is written, and that is one strings.Repeat at resolution
// because the age count is known by then.
//
// Three details the rule pins that a looser reading gets wrong:
//
//   - The counter goes on FIRST, then the cost is counted. The first
//     upkeep after the permanent enters already charges once.
//   - The cost is the printed one PER COUNTER, not the printed one
//     times a number: three age counters on a "cumulative upkeep {1}{U}"
//     permanent is {1}{U}{1}{U}{1}{U}, which ParseCost accumulates into
//     {3}{U}{U}{U}. A Phyrexian or hybrid cumulative upkeep would be
//     three separate symbols to pay, which is what repeating the string
//     produces and what multiplying a number could not express.
//   - "Unless you pay" is the CONTROLLER's decision, so declining
//     sacrifices. The prompt therefore BLOCKS the table, unlike every
//     other pay_unless — see UpkeepPayUnless and ADR 0018 §6.
//
// Not a canonical keyword token. `canonicalKeywords`
// (server/internal/game/keywords.go) is a CLOSED set of keywords the
// engine enforces from the bare string the deck importer stamps on
// Card.Keywords, and cumulative upkeep carries a COST, which a bare
// token has nowhere to put — the same reason ward is deliberately
// absent (ward.go). A cumulative-upkeep card with no catalog entry
// still flags as unimplemented, which is the honest answer; the cost
// lives on the Spec, as the argument to this constructor.

// CumulativeUpkeep builds the CR 702.24 triggered ability for one
// card. `label` is the whole stack label, as every constructor in
// triggers_common.go takes it; `cost` is the printed cumulative upkeep
// cost, charged once per age counter.
//
// Mana costs only. "Cumulative upkeep—Pay 2 life" (Glacial Chasm) and
// "Cumulative upkeep—Sacrifice a creature" (Phyrexian Soulgorger) are
// the same trigger with a different payment, and the engine's
// pay-or-else prompt parses a mana cost; those cards wait for the
// life / sacrifice payment shapes rather than being approximated here.
func CumulativeUpkeep(label, cost string) game.TriggeredAbility {
	return AtYourUpkeep(label, func(g *game.Game, item *game.StackItem) error {
		source := item.SourceCardID
		// CR 702.24a's "if this permanent is on the battlefield": a
		// permanent that left in response has nothing to age and
		// nothing to sacrifice, and its controller is not billed.
		// Left and came back is the same answer (#1432, CR 400.7).
		if !onBattlefield(g, source) || sourceIsNewObject(g, item) {
			return nil
		}
		// #1290: what the rule charges for is the counters that are
		// actually there once the placement LANDS, via
		// AddCounterThenForEffect's continuation — not on the next
		// line. The counter may be replaced or doubled on the way in
		// (a Doubling Season / Hardened Scales board), which can pause
		// it on a CR 616 prompt; reading before that resumes would
		// charge for the pre-placement age count.
		return g.AddCounterThenForEffect(source, game.CounterAge, 1, func(g *game.Game, _ int) error {
			ctx := NewContext(g, item)
			age := counterCountOn(g, source, game.CounterAge)
			if age <= 0 {
				return nil
			}
			return UpkeepPayUnless{
				Chooser:   ctx.Controller(),
				Cost:      strings.Repeat(cost, age),
				Question:  label,
				OnDecline: func(ctx *Context) error { return SacrificePermanent{Target: source}.Apply(ctx) },
			}.Apply(ctx)
		})
	})
}

// counterCountOn reads one counter kind off a card in any zone.
func counterCountOn(g *game.Game, id uuid.UUID, kind string) int {
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return 0
	}
	return c.Counters[kind]
}
