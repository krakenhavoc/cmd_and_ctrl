package effects

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ward.go — CR 702.21, the keyword ADR 0038 §7 deliberately left out
// of the targeting choke point, built where it actually belongs.
//
// The ADR's argument is worth keeping in front of the code, because
// the tempting implementation is the wrong one. Ward LOOKS like
// hexproof: both appear on a creature, both make it awkward to
// remove, and hexproof is enforced in ten lines at
// game.CanBeTargetedBy. But CR 702.21a makes ward a TRIGGERED
// ability, and the difference is observable on every play:
//
//   - A warded permanent is a perfectly LEGAL target. The removal
//     spell is announced, it goes on the stack, and the trigger goes
//     on top of it. Enforcing ward at the targeting gate would
//     refuse the announce outright — no trigger, no payment offered,
//     no response window, and no way for the spell's controller to
//     pay and have it resolve.
//   - The payer is not the ward permanent's controller. It is
//     whoever cast the spell, which is exactly the shape the
//     pay-unless and confirm prompts already have: a chooser who is
//     not the controller of the source.
//   - The trigger is an object on the stack in its own right. It can
//     be responded to, and countering IT leaves the removal spell
//     alive and unpaid-for.
//
// PARAMETERISATION. The other half of ADR 0038 §7 is that ward
// carries a cost — "ward {2}", "ward—pay 3 life", "ward—sacrifice a
// creature" — and Characteristic.Abilities is a []string of bare
// tokens with nowhere to put it. That is why ward does NOT join
// canonicalKeywords here: a bare "ward" token in the enforced table
// would tell the coverage signal (ADR 0037) that every ward card is
// implemented while the engine had no idea what to charge. The cost
// lives where the rest of a card's parameters live — on the catalog
// Spec, as an argument to the helper below — and a ward card that is
// not in the catalog still flags as unimplemented, which is the
// honest answer.
//
// SCOPE. These printed cost shapes ship, and each one is charged by
// the prompt that already fits it:
//
//   - "Ward {N}" (Hulking Raptor) — PendingChoicePayUnless, which
//     parses a mana cost and auto-taps for it.
//   - "Ward—Waterbend {N}" (The Unagi of Kyoshi Island, #1311) — the
//     same prompt carrying a waterbend clause, so the payer may tap
//     artifacts and creatures for the generic (CR 701.67a).
//   - "Ward—Pay N life" (Refraction Elemental, Sedgemoor Witch) — a
//     PendingChoiceConfirm with LifeCost declared, Sylvan Library's
//     "pay 4 life or put it back" shape (#552).
//   - "Ward—Sacrifice a creature" (Vein Ripper) — the same Confirm,
//     whose accept branch chains a PendingChoiceChooseCards over the
//     payer's own creatures.
//   - "Ward—Discard a card" (Graveyard Trespasser, Gideon the
//     Oathless, #3172) — the same Confirm, whose accept branch queues
//     the payer's own discard prompt over their hand.
//
//   - "Ward—Get five poison counters" (The Serpent Society, #4613) — the
//     same Confirm with PoisonCost declared, whose accept branch gives the
//     payer the counters. A player can always get counters, so the cost is
//     always offered (CR 118.12a); the tenth one loses the game as a
//     state-based action (CR 704.5c, CR 122.1f), and a payer who takes it
//     at five has chosen to.
//
// The mana, life and sacrifice shapes needed no engine work: the
// Confirm and ChooseCards kinds take any Chooser, have bot enumerator
// cases in legal/, and are rendered by the client. The waterbend shape
// did — the pay-unless prompt learned a tap list (game.
// CounterUnlessPaidPrompt.Waterbend, ResolvePayUnlessWithTaps). Mana
// plus fixed life is one compound payment (Gisa, #2945);
// other mixtures remain unsupported and are rejected.

// WardCost is the payment a ward demands. One field or Mana plus Life is set;
// build it with WardMana, WardManaAndLife, WardWaterbend, WardLife or WardSacrifice
// rather than by hand.
type WardCost struct {
	// Mana is "ward {N}": a mana cost string for the pay-unless prompt.
	Mana string
	// Waterbend is "Ward—Waterbend {N}" (The Unagi of Kyoshi Island,
	// #1311): the same mana payment, where the payer may tap untapped
	// artifacts and creatures they control, each paying {1} of the
	// generic (CR 701.67a). A mana cost string, like Mana.
	Waterbend string
	// Life is "ward—pay N life", alone or alongside Mana.
	Life int
	// Sacrifice is "ward—sacrifice a <thing>".
	Sacrifice *WardSacrificeCost
	// Discard is "ward—discard N cards" (#3172): the payer discards
	// that many cards from their own hand, their choice (CR 701.9a).
	Discard int
	// DiscardAtRandom makes the discard "at random" (Alpharael,
	// Stonechosen: "Ward—Discard a card at random", #2128): the payer
	// accepts the payment, and the engine picks the cards from their hand
	// on the payer's random stream instead of asking. Meaningful only
	// with Discard.
	DiscardAtRandom bool
	// DiscardKind narrows the discard to cards of a kind ("ward—discard an
	// enchantment, instant, or sorcery card", Saruman of Many Colors): the
	// payer must hold at least that many such cards, and the discard is
	// refused unless every card chosen is one. Nil is any card. Build it
	// with WardDiscardKind.Cost.
	DiscardKind *WardDiscardKind
	// Poison is "ward—get N poison counters" (The Serpent Society, #4613):
	// the payer gets that many poison counters. Always payable.
	Poison int

	// Either is a ward with two ways to pay (#4566): Titania, Rugged
	// Rumbler's "Ward—Discard a card or pay {2}". Each branch is a ward
	// cost of its own — mana, life, or a chosen discard — and the payer
	// is asked once, naming the branch (an either payment, ADR 0187 §5).
	// Build it with WardEither; no other field is set beside it.
	Either []WardCost
}

// WardEither builds a ward with two or more ways to pay: WardEither(
// WardDiscard(1), WardMana("{2}")) is "Ward—Discard a card or pay {2}"
// (#4566).
func WardEither(branches ...WardCost) WardCost {
	return WardCost{Either: append([]WardCost(nil), branches...)}
}

// WardSacrificeCost is the "sacrifice a <thing>" half of a ward cost:
// the printed noun phrase for the prompt, and the predicate a permanent
// the payer controls must pass to be sacrificed for it.
type WardSacrificeCost struct {
	Label string
	OK    CardPredicate

	// Count is how many permanents the cost demands — Valgavoth,
	// Terror Eater's "sacrifice THREE nonland permanents". Zero reads
	// as one, so every ward that shipped before the field existed
	// keeps charging exactly what it charged.
	//
	// A count rather than a repeated prompt: CR 118.3 is one payment,
	// so a payer who cannot produce all of them has not paid at all
	// and the spell is countered — never "sacrifice two of the three
	// and keep the spell". The pick is one ChooseCards prompt with
	// Min == Max == n, which the client renders as a multi-select and
	// the bot enumerator already solves for sets rather than singles.
	Count int
}

// count is the printed number, with zero reading as one.
func (s WardSacrificeCost) count() int {
	if s.Count < 1 {
		return 1
	}
	return s.Count
}

// WardMana builds the "ward {N}" cost.
func WardMana(cost string) WardCost { return WardCost{Mana: cost} }

// WardManaAndLife declares one indivisible mana-plus-life payment.
func WardManaAndLife(cost string, life int) WardCost {
	if cost == "" || life <= 0 {
		panic("effects: compound ward needs mana and positive life")
	}
	return WardCost{Mana: cost, Life: life}
}

// WardWaterbend builds the "Ward—Waterbend {N}" cost (#1311). The
// payment is the pay-unless prompt's mana payment with a tap picker
// beside it: charging a plain {N} instead would make the ward harder to
// pay than printed, the card stronger than it is (#259).
func WardWaterbend(cost string) WardCost { return WardCost{Waterbend: cost} }

// WardLife builds the "ward—pay N life" cost.
func WardLife(n int) WardCost { return WardCost{Life: n} }

// WardPoison builds the "ward—get N poison counters" cost (The Serpent
// Society, "Ward—Get five poison counters"). The payer is offered the
// counters and the spell is countered on a decline; there is no "can't
// pay", since a player can always get counters (the tenth loses the game,
// CR 704.5c).
func WardPoison(n int) WardCost { return WardCost{Poison: n} }

// WardSacrifice builds the "ward—sacrifice <label>" cost, where every
// predicate must pass: WardSacrifice("a creature", Creature()).
func WardSacrifice(label string, preds ...CardPredicate) WardCost {
	return WardCost{Sacrifice: &WardSacrificeCost{Label: label, OK: And(preds...)}}
}

// WardSacrificeN builds the "ward—sacrifice N <label>" cost:
// WardSacrificeN(3, "three nonland permanents", Nonland()). The label
// is the printed noun phrase INCLUDING the number, because that is
// what the prompt shows; `n` is what the engine counts.
func WardSacrificeN(n int, label string, preds ...CardPredicate) WardCost {
	return WardCost{Sacrifice: &WardSacrificeCost{Label: label, OK: And(preds...), Count: n}}
}

// WardDiscard builds the "ward—discard N cards" cost (Graveyard
// Trespasser, Gideon the Oathless: "Ward—Discard a card"). The payer
// picks which cards go, and a payer holding fewer than N cannot pay at
// all (CR 118.3), so the spell is countered with no prompt.
func WardDiscard(n int) WardCost { return WardCost{Discard: n} }

// WardDiscardAtRandom builds "ward—discard N cards at random" (CR 701.9b;
// Alpharael, Stonechosen). The payer chooses only whether to pay.
func WardDiscardAtRandom(n int) WardCost { return WardCost{Discard: n, DiscardAtRandom: true} }

// WardDiscardKind is a kind of card a ward asks to be discarded: "an
// enchantment, instant, or sorcery card". It is registered once, from a
// package-level var, so the discard prompt it queues is a restore point (the
// set rule is registered with the key). The key is an on-disk identity, never
// renamed and never reused.
type WardDiscardKind struct {
	key   string
	label string
	match func(game.Card) bool
	then  game.DiscardThen
}

var (
	wardDiscardKindsMu sync.Mutex
	wardDiscardKinds   = map[string]*WardDiscardKind{}
)

// NewWardDiscardKind registers a kind of card a ward may ask for. `match` is
// a function of the card alone: it runs under the discard prompt's set rule,
// on the submit path and in the bot enumerator, and must capture no state.
// `label` is the printed noun phrase, "an enchantment, instant, or sorcery
// card". It panics on a duplicate key.
func NewWardDiscardKind(key, label string, match func(game.Card) bool) *WardDiscardKind {
	k := &WardDiscardKind{key: key, label: label, match: match}
	k.then = game.RegisterDiscardThenRule("choose-cards/ward-discard-"+key,
		func(picked []game.Card) bool {
			for _, c := range picked {
				if !match(c) {
					return false
				}
			}
			return true
		},
		func(*game.Game, game.DiscardsDone) error { return nil })
	wardDiscardKindsMu.Lock()
	defer wardDiscardKindsMu.Unlock()
	if _, dup := wardDiscardKinds[key]; dup {
		panic("effects: duplicate ward discard kind " + key)
	}
	wardDiscardKinds[key] = k
	return k
}

// Cost is "ward—discard a <kind> card": one card of the kind, the payer's
// choice (CR 701.9a).
func (k *WardDiscardKind) Cost() WardCost { return WardCost{Discard: 1, DiscardKind: k} }

// validate panics on an empty cost or an unsupported mixture. It runs
// from Ward() at catalog init, so a malformed card fails at boot rather
// than charging the first component it happens to check.
func (c WardCost) validate() {
	if len(c.Either) > 0 {
		c.validateEither()
		return
	}
	if c.Life < 0 {
		panic("effects: a ward life cost cannot be negative")
	}
	if c.Poison < 0 {
		panic("effects: a ward poison cost cannot be negative")
	}
	n := 0
	if c.Mana != "" {
		n++
	}
	if c.Waterbend != "" {
		n++
	}
	if c.Life > 0 {
		n++
	}
	if c.Sacrifice != nil {
		n++
	}
	if c.Discard > 0 {
		n++
	}
	if c.Poison > 0 {
		n++
	}
	compound := n == 2 && c.Mana != "" && c.Life > 0
	if (n != 1 && !compound) || (c.Sacrifice != nil && c.Sacrifice.OK == nil) {
		panic(fmt.Sprintf("effects: a ward cost needs one component, or Mana plus Life: %+v", c))
	}
	if c.DiscardKind != nil && (c.Discard < 1 || c.DiscardAtRandom) {
		panic("effects: a ward discard kind needs a chosen discard of at least one card")
	}
	if c.Sacrifice != nil {
		// Register the cost's two keyed legs at boot (#4291).
		wardSacrificeKeysFor(*c.Sacrifice)
	}
}

// Ward builds the CR 702.21a triggered ability: "Whenever this
// permanent becomes the target of a spell or ability an opponent
// controls, counter that spell or ability unless its controller pays
// <cost>."
//
// Two details in the AppliesTo that are the rule rather than
// defensive coding:
//
//   - "An OPPONENT controls" is `ev.Actor != <the warded
//     permanent>.Controller`. Your own Giant Growth on your own
//     warded creature does not trigger it, which is the difference
//     between ward and shroud.
//   - The trigger fires per target INSTANCE, because
//     EventBecomesTarget is emitted per target slot (CR 115.3). A
//     spell that targets the same warded creature twice triggers
//     ward twice, and the controller pays twice or the spell is
//     countered. That is correct and it falls out of the event
//     shape rather than needing anything here.
//
// The Effect reads the targeting spell off the event's StackItemID
// rather than its Source, because for an activated or triggered
// ability those differ — Source is the permanent the ability came
// from and would counter the wrong object (or nothing at all).
//
// It is WardGranted with the narrowest possible predicate — the
// warded permanent is the source itself — so the printed ward and
// every granted one share one trigger body and one payment path.
func Ward(cost WardCost, label string) game.TriggeredAbility {
	return WardGranted(cost, label, func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID
	})
}

// WardGranted is "<these permanents> have ward <cost>": a ward one
// object gives to OTHERS. Lavaspur Boots' "equipped creature has
// ward {1}" is one shape of it; Teferi Akosa of Zhalfir's emblem,
// "Knights you control get +1/+0 and have ward {1}", is the other.
//
// It is NOT a layer-6 keyword grant, and the reason is the file
// comment's: CR 702.21a makes ward a TRIGGERED ability, and
// `Characteristic.Abilities` is a []string with nowhere to put the
// cost, which is exactly why "ward" is kept out of
// `canonicalKeywords`. So the grant is the GRANTING object carrying
// the trigger and watching for a permanent it grants ward to becoming
// a target. Nothing is written onto the warded permanent; what reads
// the ward is the harvester, walking the granting object's own zone
// — the battlefield for an Equipment, the command zone for an emblem
// (CR 114.3).
//
// `grants` is the ordinary `StaticAbility.AppliesTo` shape
// (target, game, source), so ONE predicate value serves both halves
// of "Knights you control get +1/+0 and have ward {1}" and the layer
// half and the trigger half can never drift apart. `source` is the
// granting object, so "you control" is the same
// `target.Controller == source.Controller` an anthem uses.
//
// Three details that are the rule rather than defensive coding:
//
//   - The candidate must be a BATTLEFIELD permanent. Ward is a
//     permanent's ability (CR 702.21a), and EventBecomesTarget is
//     emitted for spells on the stack and cards in graveyards too, so
//     without the zone test an emblem's "Knights you control" would
//     answer a Raise Dead aimed at a Knight card in a graveyard.
//   - "An opponent controls" is measured against the WARDED
//     permanent's controller, not the granting object's. The two
//     agree for every printed case and can diverge only behind a
//     control change, and then the rule that matters is the
//     permanent's.
//   - The PAYER is always the spell's controller (ev.Actor), which is
//     right in every case. What is approximated is the trigger's own
//     CONTROLLER: NewTriggeredItem keys on the granting object, so
//     the stack object belongs to the Equipment's or the emblem's
//     controller rather than the warded permanent's. That shows up
//     only in trigger ordering, behind a control change that
//     separated the two, and it is noted rather than modelled.
func WardGranted(cost WardCost, label string, grants func(target *game.Card, g *game.Game, source *game.Card) bool) game.TriggeredAbility {
	cost.validate()
	if label == "" {
		label = "Ward"
	}
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventBecomesTarget},
		// The label is the card's own, unadorned: the cost is already
		// in it ("Hulking Raptor — ward {2}"), and the pay-or-counter
		// wording belongs on the PROMPT, which is where the decision
		// is actually made. It is the row's Key as well, which is the
		// name a restored item checks the row by (ADR 0041 P9).
		Key: label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if grants == nil {
				return false
			}
			// The targeting item and the payer are what the
			// resolution reads off the triggering event; an event
			// naming neither has nothing to charge or counter.
			if ev.StackItemID == uuid.Nil || ev.Actor == uuid.Nil {
				return false
			}
			// One zone lookup, not two: the battlefield test and the
			// warded permanent both come out of the same walk.
			z := g.FindCardZoneForEffect(ev.CardID)
			if z == nil || z.Kind != game.ZoneBattlefield {
				return false
			}
			for i := range z.Cards {
				warded := &z.Cards[i]
				if warded.InstanceID != ev.CardID {
					continue
				}
				if !grants(warded, g, source) || ev.Actor == warded.Controller {
					return false
				}
				// CR 702.21 + Nowhere to Run (#1560): "ward abilities of
				// those creatures don't trigger". Asked of the WARDED
				// permanent, not the source — a granted ward's source
				// is the Equipment or emblem — and only here, at the
				// moment it would trigger, so the static leaving later
				// does not trigger it after the fact.
				return !g.WardSuppressedForEffect(warded)
			}
			return false
		},
		// ADR 0041 P9 (tier 4-2): declared on the row. The targeting
		// item and the payer are read at resolution off the item's
		// triggering event (#1223) rather than captured when the
		// ability triggered, so a restored item charges the same
		// player for the same spell. The cost is the row's own.
		Effect: func(g *game.Game, item *game.StackItem) error {
			if item.Trigger == nil {
				return nil
			}
			ev := item.Trigger.Event
			return wardPayOrCounter(g, item, ev.StackItemID, ev.Actor, cost)
		},
	}
}

// wardPayOrCounter is the trigger's resolution: offer the payment to
// the spell's controller, counter on a decline.
//
// Every cost shape treats the two ways a decline happens — answering
// "no", and being unable to pay (no mana in pool plus untapped
// sources, too little life, nothing to sacrifice) — identically, which
// is CR 118.12.
//
// A spell that has already left the stack when the trigger resolves
// (countered in response, or its own controller responded by
// sacrificing the ward permanent and the spell then fizzled) makes
// the counter a no-op: CounterTargetForEffect returns
// ErrCardNotOnStack and there is nothing to charge for either, so
// the prompt is skipped entirely rather than asking a player to pay
// for a spell that no longer exists.
func wardPayOrCounter(g *game.Game, item *game.StackItem, targetingItem, payer uuid.UUID, cost WardCost) error {
	if g.StackItemForEffect(targetingItem) == nil {
		return nil
	}
	switch {
	case len(cost.Either) > 0:
		return wardEitherOrCounter(g, item.SourceCardID, payer, cost, targetingItem)
	case cost.Life > 0 && cost.Mana == "":
		return wardPayLifeOrCounter(g, item.SourceCardID, payer, cost.Life, targetingItem)
	case cost.Sacrifice != nil:
		return wardSacrificeOrCounter(g, item.SourceCardID, payer, *cost.Sacrifice, targetingItem)
	case cost.Discard > 0:
		return wardDiscardOrCounter(g, item.SourceCardID, payer, cost, targetingItem)
	case cost.Poison > 0:
		return wardPoisonOrCounter(g, item.SourceCardID, payer, cost.Poison, targetingItem)
	}
	if cost.Waterbend != "" {
		// #1311: "Ward—Waterbend {N}" is the mana payment below with
		// CR 701.67a's second way to cover each generic symbol. The
		// clause is effects.Waterbend — the one a spell's waterbend
		// uses — so "you control" means the PAYER's permanents: the
		// validator reads it for the chooser, not for the warded
		// permanent's controller.
		return g.QueueCounterUnlessPaidForEffect(game.CounterUnlessPaidPrompt{
			StackItem: targetingItem,
			Chooser:   payer,
			Source:    item.SourceCardID,
			Cost:      cost.Waterbend,
			Question:  "Ward — waterbend " + cost.Waterbend + " or the spell is countered",
			Waterbend: Waterbend(cost.Waterbend),
		})
	}
	question := "Ward — pay " + cost.Mana
	if cost.Life > 0 {
		question += fmt.Sprintf(" and %d life", cost.Life)
	}
	return g.QueueCounterUnlessPaidForEffect(game.CounterUnlessPaidPrompt{
		StackItem: targetingItem,
		Chooser:   payer,
		Source:    item.SourceCardID,
		Cost:      cost.Mana,
		LifeCost:  cost.Life,
		Question:  question + " or the spell is countered",
	})
}

// wardCounter is the ward's decline: counter the targeting spell or ability
// if it is still on the stack, and do nothing if it is not.
func wardCounter(g *game.Game, targetingItem uuid.UUID) error {
	if g.StackItemForEffect(targetingItem) == nil {
		return nil
	}
	return g.CounterTargetForEffect(targetingItem)
}

// The three confirm legs below are KEYED (#4291): the prompt holds no closure,
// so a table waiting on a ward's life, sacrifice or discard question is a
// restore point. What a leg needs rides the prompt as plain data. The
// targeting spell or ability is its stack ID (the one carried ID), the payer
// is the chooser, the warded permanent is the source, and the life a payment
// costs is the prompt's LifeCost (the discard's count its Word).

// wardPayLifeAnswer is the payer's answer to "pay N life, or it is
// countered". N is the life the prompt declares.
var wardPayLifeAnswer = game.RegisterConfirmThen("confirm/ward-pay-life", func(g *game.Game, r game.ConfirmAnswer) error {
	targetingItem := wardTargetingItem(r.Carry)
	if !r.Accepted {
		return wardCounter(g, targetingItem)
	}
	if g.StackItemForEffect(targetingItem) == nil {
		// Nothing left to pay for.
		return nil
	}
	if p := g.PlayerByIDForEffect(r.Chooser); p == nil || p.Life < r.LifeCost {
		return wardCounter(g, targetingItem)
	}
	// Ward's "unless that player pays N life" is a COST (CR 118.3), so it
	// goes through the cost path: the CR 614 window runs (CR 119.4 — paying
	// life is losing life) but settles in one step rather than leaving the
	// countered-or-not question hanging on a CR 616 prompt.
	return g.PayLifeForEffect(r.Source, r.Chooser, r.LifeCost)
})

// wardGetPoisonAnswer is the payer's answer to "get N poison counters, or it
// is countered". N is the count the prompt declares (PoisonCost, #4613).
var wardGetPoisonAnswer = game.RegisterConfirmThen("confirm/ward-get-poison", func(g *game.Game, r game.ConfirmAnswer) error {
	targetingItem := wardTargetingItem(r.Carry)
	if !r.Accepted {
		return wardCounter(g, targetingItem)
	}
	if g.StackItemForEffect(targetingItem) == nil {
		// Nothing left to pay for.
		return nil
	}
	// Getting counters is a cost paid in full or not at all (CR 118.3). The
	// placement goes through the CR 614 window as every counter does, and
	// the state-based action that loses the game at ten (CR 704.5c) runs at
	// the next check, not here.
	if err := g.AddPlayerCounterByForEffect(r.Chooser, r.Chooser, game.CounterPoison, r.PoisonCost); err != nil {
		if errors.Is(err, game.ErrPlayerEliminated) {
			return wardCounter(g, targetingItem)
		}
		return err
	}
	return nil
})

// wardPoisonOrCounter is "counter it unless that player gets N poison
// counters": a Confirm whose accept branch gives them. Unlike every other
// non-mana ward it has no unpayable case to settle without a prompt, since a
// player can always get counters. PoisonCost is declared on the Confirm so
// the bot's move list prices the accept branch (legal.MoveCost.Poison); a
// bot at nine poison weighs a payment that loses the game as one.
func wardPoisonOrCounter(g *game.Game, source, payer uuid.UUID, n int, targetingItem uuid.UUID) error {
	_, err := g.QueueConfirmKeyedForEffect(game.ConfirmPrompt{
		Chooser:      payer,
		Source:       source,
		Question:     fmt.Sprintf("Ward — get %s, or it is countered", wardPoisonLabel(n)),
		AcceptLabel:  "Get " + wardPoisonLabel(n),
		DeclineLabel: "Let it be countered",
		PoisonCost:   n,
	}, wardGetPoisonAnswer, []uuid.UUID{targetingItem})
	return err
}

// wardPoisonLabel is "5 poison counters" (a single one is "a poison counter").
func wardPoisonLabel(n int) string {
	if n == 1 {
		return "a poison counter"
	}
	return fmt.Sprintf("%d poison counters", n)
}

// wardTargetingItem is the stack ID a ward leg carries.
func wardTargetingItem(carry []uuid.UUID) uuid.UUID {
	if len(carry) == 0 {
		return uuid.Nil
	}
	return carry[0]
}

// wardPayLifeOrCounter is "counter it unless that player pays N life".
//
// CR 119.4: a player may pay life only if their life total is at least
// the payment. Below that the counter is the only legal outcome, so it
// happens without a prompt — an offer the engine would have to refuse
// is how a seat wedges (#544). At exactly N the payment is offered, and
// the payer loses to the state-based action if they take it; that is a
// legal play.
//
// LifeCost is declared on the Confirm so the bot's move list prices the
// accept branch; without it a bot at 3 life answers "pay" and dies
// (#547). The branch re-checks life when the answer arrives rather than
// trusting the prompt-time total.
func wardPayLifeOrCounter(g *game.Game, source, payer uuid.UUID, life int, targetingItem uuid.UUID) error {
	if p := g.PlayerByIDForEffect(payer); p == nil || p.Life < life {
		return wardCounter(g, targetingItem)
	}
	_, err := g.QueueConfirmKeyedForEffect(game.ConfirmPrompt{
		Chooser:      payer,
		Source:       source,
		Question:     fmt.Sprintf("Ward — pay %d life, or it is countered", life),
		AcceptLabel:  fmt.Sprintf("Pay %d life", life),
		DeclineLabel: "Let it be countered",
		LifeCost:     life,
	}, wardPayLifeAnswer, []uuid.UUID{targetingItem})
	return err
}

// wardSacrificeKeys are the two registered legs of one sacrifice ward cost:
// the confirm, and the pick it chains.
type wardSacrificeKeys struct {
	confirm game.ConfirmThen
	pick    game.ChooseCardsThen
	count   int
}

var (
	wardSacrificeMu   sync.Mutex
	wardSacrificeLegs = map[string]wardSacrificeKeys{}
)

// wardSlug is a cost's printed noun phrase as a key suffix: "a creature" is
// "a-creature".
func wardSlug(label string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(label) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			dash = false
		} else if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

// wardSacrificeKeysFor registers, once per cost label, the two legs of a
// "ward—sacrifice <label>" payment as keys ("confirm/ward-sacrifice-<slug>"
// and "choose-cards/ward-sacrifice-<slug>", #4291). The cost's predicate is a
// function and cannot ride a prompt, so the key IS the cost: it names the
// printed noun phrase, and the legs close over this cost, a constant of the
// card. A label therefore names one cost shape. Two costs with one label must
// be the same cost, and one with a different count panics.
func wardSacrificeKeysFor(sac WardSacrificeCost) wardSacrificeKeys {
	wardSacrificeMu.Lock()
	defer wardSacrificeMu.Unlock()
	slug := wardSlug(sac.Label)
	if k, ok := wardSacrificeLegs[slug]; ok {
		if k.count != sac.count() {
			panic(fmt.Sprintf("effects: two ward sacrifice costs share the label %q with different counts", sac.Label))
		}
		return k
	}
	n := sac.count()
	pick := game.RegisterChooseCardsThen("choose-cards/ward-sacrifice-"+slug, wardSacrificePickLeg(sac, n))
	confirm := game.RegisterConfirmThen("confirm/ward-sacrifice-"+slug, wardSacrificeConfirmLeg(sac, n, pick))
	k := wardSacrificeKeys{confirm: confirm, pick: pick, count: n}
	wardSacrificeLegs[slug] = k
	return k
}

// wardSacrificePickLeg is the pick's answer: the payer's permanents, checked
// again against the battlefield, are sacrificed; an incomplete pick (or none,
// when the prompt was withdrawn) leaves the cost unpaid and the spell is
// countered.
func wardSacrificePickLeg(sac WardSacrificeCost, n int) game.ChooseCardsFunc {
	return func(g *game.Game, r game.CardsChosen) error {
		targetingItem := wardTargetingItem(r.Carry)
		if g.StackItemForEffect(targetingItem) == nil {
			return nil
		}
		if !wardSacrificePickComplete(g, r.Chooser, sac, r.Picked, n) {
			return wardCounter(g, targetingItem)
		}
		for _, id := range r.Picked {
			if err := g.SacrificePermanentForEffect(id); err != nil {
				return err
			}
		}
		return nil
	}
}

// wardSacrificeConfirmLeg is the confirm's answer: a decline counters the
// spell, an accept chains the pick over the payer's matching permanents.
func wardSacrificeConfirmLeg(sac WardSacrificeCost, n int, pick game.ChooseCardsThen) game.ConfirmFunc {
	return func(g *game.Game, r game.ConfirmAnswer) error {
		targetingItem := wardTargetingItem(r.Carry)
		if !r.Accepted {
			return wardCounter(g, targetingItem)
		}
		if g.StackItemForEffect(targetingItem) == nil {
			return nil
		}
		candidates := wardSacrificeCandidates(g, r.Chooser, sac)
		if len(candidates) < n {
			return wardCounter(g, targetingItem)
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  r.Chooser,
			Source:   r.Source,
			Question: "Ward — choose " + sac.Label + " to sacrifice",
			Cards:    candidates,
			Min:      n,
			Max:      n,
			ThenKey:  pick,
			Carry:    r.Carry,
			// The pick is withdrawn, and the spell countered, if the payer's
			// matching permanents leave the battlefield under it (#1225).
			LiveZone: game.ZoneBattlefield,
		})
		return nil
	}
}

// wardSacrificeOrCounter is "counter it unless that player sacrifices
// <a creature>": a Confirm whose accept branch chains a one-card pick
// over the payer's own matching permanents, then sacrifices the pick.
//
// Two prompts rather than one ChooseCards with a floor of zero, because
// "choose nothing" is a poor way to say "let my spell be countered",
// and the bot enumerator already treats a Confirm's decline as the
// always-legal answer. A payer with nothing to sacrifice cannot pay, so
// the spell is countered without a prompt.
//
// The candidates are the payer's permanents, not targets: hexproof and
// protection don't stop a player sacrificing their own creature, so the
// predicate is the whole test. Both links re-read the battlefield when
// their answer arrives, and a pick that is no longer a matching
// permanent the payer controls counts as not paying. Both links are keyed
// (wardSacrificeKeysFor), so a table waiting on either is a restore point.
//
// A cost that names more than one permanent (Valgavoth, Terror
// Eater's three) changes only the bounds on the pick and the guard
// in front of it: a payer holding fewer than the count cannot pay at
// all, so the spell is countered with no prompt rather than offering
// a payment that cannot be completed.
func wardSacrificeOrCounter(g *game.Game, source, payer uuid.UUID, sac WardSacrificeCost, targetingItem uuid.UUID) error {
	if len(wardSacrificeCandidates(g, payer, sac)) < sac.count() {
		return wardCounter(g, targetingItem)
	}
	_, err := g.QueueConfirmKeyedForEffect(game.ConfirmPrompt{
		Chooser:      payer,
		Source:       source,
		Question:     "Ward — sacrifice " + sac.Label + ", or it is countered",
		AcceptLabel:  "Sacrifice " + sac.Label,
		DeclineLabel: "Let it be countered",
	}, wardSacrificeKeysFor(sac).confirm, []uuid.UUID{targetingItem})
	return err
}

// wardDiscardWord is a discard ward's Word: the count, then "random" for a
// discard at random, or "kind:<key>" for a discard of a kind of card.
func wardDiscardWord(n int, atRandom bool, kind *WardDiscardKind) string {
	switch {
	case kind != nil:
		return strconv.Itoa(n) + " kind:" + kind.key
	case atRandom:
		return strconv.Itoa(n) + " random"
	}
	return strconv.Itoa(n)
}

// wardDiscardLabel is the printed "a card" / "N cards" / "... at random" /
// "an enchantment, instant, or sorcery card".
func wardDiscardLabel(n int, atRandom bool, kind *WardDiscardKind) string {
	if kind != nil {
		return kind.label
	}
	label := "a card"
	if n > 1 {
		label = fmt.Sprintf("%d cards", n)
	}
	if atRandom {
		label += " at random"
	}
	return label
}

// wardDiscardAnswer is the payer's answer to "discard <N cards>, or it is
// countered". The Word is the count, and "random" after it for a discard at
// random.
var wardDiscardAnswer = game.RegisterConfirmThen("confirm/ward-discard", func(g *game.Game, r game.ConfirmAnswer) error {
	targetingItem := wardTargetingItem(r.Carry)
	if !r.Accepted {
		return wardCounter(g, targetingItem)
	}
	if g.StackItemForEffect(targetingItem) == nil {
		return nil
	}
	count, rest, _ := strings.Cut(r.Word, " ")
	n, _ := strconv.Atoi(count)
	atRandom := rest == "random"
	var kind *WardDiscardKind
	if key, ok := strings.CutPrefix(rest, "kind:"); ok {
		wardDiscardKindsMu.Lock()
		kind = wardDiscardKinds[key]
		wardDiscardKindsMu.Unlock()
	}
	if wardDiscardCardsInHand(g, r.Chooser, kind) < n {
		return wardCounter(g, targetingItem)
	}
	prompt := game.DiscardPrompt{
		Player:   r.Chooser,
		Source:   r.Source,
		N:        n,
		Question: "Ward — discard " + wardDiscardLabel(n, false, kind),
	}
	if kind != nil {
		return g.PlayerDiscardsThenKeyedForEffect(prompt, kind.then, nil)
	}
	if atRandom {
		return g.DiscardRandomForEffect(r.Chooser, n)
	}
	g.QueueDiscardChoiceForEffect(prompt)
	return nil
})

// wardDiscardCardsInHand is how many cards in the payer's hand can pay a
// discard of `kind`: any card for a nil kind.
func wardDiscardCardsInHand(g *game.Game, payer uuid.UUID, kind *WardDiscardKind) int {
	p := g.PlayerByIDForEffect(payer)
	if p == nil || p.Hand == nil {
		return 0
	}
	if kind == nil {
		return p.Hand.Size()
	}
	n := 0
	for _, c := range p.Hand.Cards {
		if kind.match(c) {
			n++
		}
	}
	return n
}

// wardDiscardOrCounter is "counter it unless that player discards <N
// cards>": a Confirm whose accept branch queues the payer's own
// discard prompt, the way the sacrifice ward chains its pick.
//
// A payer with fewer cards in hand than the cost names cannot pay, so
// the spell is countered with no prompt (CR 118.3 — a cost is paid in
// full or not at all; "as many as you can" is CR 701.9a's rule for an
// instruction, not for a cost). Accepting commits the payer to the
// discard: the prompt it queues has no floor below N, and the hand is
// re-read when the answer arrives.
//
// A cost is not an instruction to a player, so nothing here targets:
// hexproof and protection are irrelevant, and the cards are the
// payer's own.
func wardDiscardOrCounter(g *game.Game, source, payer uuid.UUID, cost WardCost, targetingItem uuid.UUID) error {
	n, atRandom, kind := cost.Discard, cost.DiscardAtRandom, cost.DiscardKind
	if wardDiscardCardsInHand(g, payer, kind) < n {
		return wardCounter(g, targetingItem)
	}
	label := wardDiscardLabel(n, atRandom, kind)
	_, err := g.QueueConfirmKeyedForEffect(game.ConfirmPrompt{
		Chooser:      payer,
		Source:       source,
		Question:     "Ward — discard " + label + ", or it is countered",
		AcceptLabel:  "Discard " + label,
		DeclineLabel: "Let it be countered",
		Word:         wardDiscardWord(n, atRandom, kind),
	}, wardDiscardAnswer, []uuid.UUID{targetingItem})
	return err
}

// wardSacrificeCandidates lists the permanents payer controls that the
// sacrifice cost accepts, in battlefield order. A permanent that can't
// be sacrificed can't pay it (#3275, CR 118.3).
func wardSacrificeCandidates(g *game.Game, payer uuid.UUID, sac WardSacrificeCost) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == payer && sac.OK(g, payer, c) && !g.CantBeSacrificedForEffect(c.InstanceID) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// wardSacrificeLegal re-checks one pick against the live battlefield.
func wardSacrificeLegal(g *game.Game, payer uuid.UUID, sac WardSacrificeCost, id uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == id {
			return c.Controller == payer && sac.OK(g, payer, c) && !g.CantBeSacrificedForEffect(id)
		}
	}
	return false
}

// wardSacrificePickComplete re-checks a whole answer against the live
// battlefield: exactly `n` permanents, all distinct, each still one
// the payer controls and each still matching the cost.
//
// Distinctness is checked here rather than assumed, because the
// payment is one cost (CR 118.3) and the same permanent cannot pay
// for itself twice — a repeated ID would otherwise sacrifice one
// permanent and silently discharge a three-permanent cost.
func wardSacrificePickComplete(g *game.Game, payer uuid.UUID, sac WardSacrificeCost, picked []uuid.UUID, n int) bool {
	if len(picked) != n {
		return false
	}
	seen := make(map[uuid.UUID]bool, n)
	for _, id := range picked {
		if seen[id] || !wardSacrificeLegal(g, payer, sac, id) {
			return false
		}
		seen[id] = true
	}
	return true
}
