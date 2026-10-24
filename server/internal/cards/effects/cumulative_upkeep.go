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
// A mana cost. "Cumulative upkeep—Sacrifice a land", "—Discard a card"
// and "—Pay 1 life" are CumulativeUpkeepPaying (ADR 0108 §5; the life
// payment is LifePayment, #3869).
func CumulativeUpkeep(label, cost string) game.TriggeredAbility {
	return CumulativeUpkeepPaying(label, UpkeepPayment{Mana: cost}, "", "")
}

// CumulativeUpkeepPaying is CumulativeUpkeep for any payment the upkeep
// prompt takes, including the two that are not mana (ADR 0108 §5, owner
// decision 3): Polar Kraken's "Cumulative upkeep—Sacrifice a land" is
//
//	CumulativeUpkeepPaying(label, SacrificePayment(1, "land", "lands", game.PermanentQuery{Types: []string{"land"}}), "land", "lands")
//
// With three age counters that is "Sacrifice three lands", chosen and
// paid all at once or not at all (CR 702.24a: "either the entire set of
// costs is paid, or none of them is paid"). `noun` and `plural` name a
// sacrifice's permanents in the prompt; a mana, discard or life payment
// ignores them.
//
// The trigger carries game.KeywordCumulativeUpkeep (#3869), which is
// how game.HasCumulativeUpkeep finds an instance, printed or granted.
func CumulativeUpkeepPaying(label string, each UpkeepPayment, noun, plural string) game.TriggeredAbility {
	t := AtYourUpkeep(label, func(g *game.Game, item *game.StackItem) error {
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
			owed := each.times(age, noun, plural)
			cost, action := owed.prompt()
			prompt := UpkeepPayUnless{
				Chooser:  ctx.Controller(),
				Cost:     cost,
				Action:   action,
				Question: label,
			}
			if owed.Flips > 0 {
				// The flips are the payment, queued once it is made
				// (cumulativeUpkeepFlipCoins).
				prompt.Else = cumulativeUpkeepFlipCoins
			} else {
				// #4264: keyed, so a table waiting on the payer is a
				// restore point.
				prompt.Else = upkeepSacrificeSource
			}
			return prompt.Apply(ctx)
		})
	})
	t.Keyword = game.KeywordCumulativeUpkeep
	return t
}

// cumulativeUpkeepFlipCoins is the two answers of "Cumulative upkeep—Flip
// a coin" (#4063, Karplusan Minotaur), keyed (#4264): saying yes costs
// nothing to say, and the flips are the payment, queued once it is made,
// under one call for the age counters' worth (CR 705.2); a decline
// sacrifices the permanent. A coin per age counter is the Minotaur's
// payment; the coins and the words are worked out from the counters when
// the answer arrives, which is what the prompt was queued with, because the
// upkeep holds the table until it is answered. The question is the asking
// trigger's label.
var cumulativeUpkeepFlipCoins = PayUnlessBranches("pay-unless/cumulative-upkeep-flip-coins",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		owed := FlipCoinsPayment(1).times(counterCountOn(ctx.Game, r.Source, game.CounterAge), "", "")
		ctx.Game.FlipCoinForEffect(game.CoinFlipSpec{
			Flipper:  ctx.Controller(),
			Source:   r.Source,
			Question: ctx.Item.Label + ": " + strings.ToLower(owed.Words[:1]) + owed.Words[1:],
			Coins:    owed.Flips,
		})
		return nil
	},
	func(ctx *Context, r game.PayUnlessAnswer) error {
		return SacrificePermanent{Target: r.Source}.Apply(ctx)
	})

// CumulativeUpkeepGrant is the ADR 0093 bundle for a QUOTED cumulative
// upkeep that another effect gives a permanent (#3869): Mana Chains'
// "Enchanted creature has 'Cumulative upkeep {1}.'", Balduvian Shaman's
// "That enchantment gains 'Cumulative upkeep {1}.'". `key` is the
// catalog-wide bundle key and `cost` the quoted mana cost, charged once
// per age counter.
//
// Nothing about the trigger changes when it is granted. A granted
// trigger's source is the HOST (ADR 0093 Decision 4), so the age
// counter goes on the host, the host's controller is asked, and a
// decline sacrifices the host. A permanent with several instances (its
// own and a granted one, or two grants) triggers once per instance, and
// each instance counts every age counter on the permanent as it
// resolves (CR 702.24b), which is what reading the counters at
// resolution already does.
func CumulativeUpkeepGrant(key, cost string) AbilityGrant {
	return CumulativeUpkeepGrantPaying(key, UpkeepPayment{Mana: cost}, "Cumulative upkeep "+cost)
}

// CumulativeUpkeepGrantPaying is CumulativeUpkeepGrant for a payment
// that is not mana. Decomposition's "Cumulative upkeep—Pay 1 life." is
//
//	CumulativeUpkeepGrantPaying(key, LifePayment(1), "Cumulative upkeep—Pay 1 life")
//
// `text` is the quoted ability as printed, without its closing period;
// it is also the trigger's stack label.
func CumulativeUpkeepGrantPaying(key string, each UpkeepPayment, text string) AbilityGrant {
	return AbilityGrant{
		Key:       key,
		Triggered: []game.TriggeredAbility{CumulativeUpkeepPaying(text, each, "", "")},
		Text:      text + ".",
	}
}

// DoesNotHaveCumulativeUpkeep is the target clause "that doesn't have
// cumulative upkeep" (Balduvian Shaman, #3869): no instance of CR
// 702.24 on the object, printed or granted, as the layers leave it.
func DoesNotHaveCumulativeUpkeep() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return !game.HasCumulativeUpkeep(c) }
}

// counterCountOn reads one counter kind off a card in any zone.
func counterCountOn(g *game.Game, id uuid.UUID, kind string) int {
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return 0
	}
	return c.Counters[kind]
}
