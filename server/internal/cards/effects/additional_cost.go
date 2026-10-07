package effects

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// additional_cost.go — S21 sub-PR 5: constructors for Spec.
// AdditionalCost, the "As an additional cost to cast this spell, …"
// clause (CR 601.2f).
//
// Two shapes so far. Exile-from-graveyard still has no card in the
// current decklists, and a constructor with no caller is a guess about
// an API rather than an API.

// DiscardCost is "As an additional cost to cast this spell, discard
// N cards." The engine validates the caster's picks at announce and
// pays them with the spell already on the stack, so a discard
// payoff triggers above the spell and resolves first.
func DiscardCost(n int) *game.AdditionalCost {
	return &game.AdditionalCost{DiscardCards: n, Label: discardLabel(n)}
}

// discardLabel spells the clause the way the card prints it, since
// the client shows it verbatim above the picker.
func discardLabel(n int) string {
	switch n {
	case 1:
		return "Discard a card"
	case 2:
		return "Discard two cards"
	case 3:
		return "Discard three cards"
	}
	return "Discard " + strconv.Itoa(n) + " cards"
}

// SacrificeCost is "As an additional cost to cast this spell,
// sacrifice a creature" (Village Rites, Altar's Reap) — or any wider
// clause, via the predicates. Build the spec with the same
// sacrificeSpec helper the activated and mana abilities use, so
// "you may only sacrifice what you control" is enforced in one place.
//
// Paid with the spell already on the stack, so an aristocrats payoff
// watching the death triggers above the spell and drains first.
func SacrificeCost(label string, preds ...CardPredicate) *game.AdditionalCost {
	return &game.AdditionalCost{
		Sacrifice: sacrificeSpec(label, preds...),
		Label:     "Sacrifice " + label,
	}
}

// SacrificeNCost is "As an additional cost to cast this spell,
// sacrifice N <permanents>": SacrificeNCost(2, "two creatures",
// Creature()). SacrificeCost is the n = 1 case and stays the way to
// write it. The count rides on the clause exactly as SacrificeN's
// does (#747, ADR 0021 addendum), so the caster names exactly n
// permanents and they leave as one simultaneous exit while the spell
// is on the stack.
func SacrificeNCost(n int, label string, preds ...CardPredicate) *game.AdditionalCost {
	return &game.AdditionalCost{
		Sacrifice: sacrificeSpec(label, preds...).WithCount(n, n),
		Label:     "Sacrifice " + label,
	}
}

// SacrificeXCost is "As an additional cost to cast this spell,
// sacrifice X <permanents>" — Devastating Summons' "sacrifice X
// lands", Eliminate the Competition's "sacrifice X creatures" (ADR 0100
// §3). The count is the X announced with the cast (CR 107.3a,
// CastSpellParams.XValue), the same X the spell's text reads back with
// ctx.X(): CR 107.3i makes every X on the object one number, so "destroy
// X target creatures" is the existing CountFromX target clause.
//
// Register refuses it beside a "pay X life" clause: no card prints the
// pair, and one announced X paying two printed costs is a price nobody
// chose.
func SacrificeXCost(label string, preds ...CardPredicate) *game.AdditionalCost {
	spec := sacrificeSpec(label, preds...)
	spec.CountFromX = true
	return &game.AdditionalCost{
		Sacrifice: spec,
		Label:     "Sacrifice " + label,
	}
}

// SacrificeAnyNumberCost is "As an additional cost to cast this spell,
// sacrifice any number of <permanents>" (Vicious Betrayal) and its
// "you may" spellings — "you may sacrifice any number of creatures"
// (Torgaar, Famine Incarnate), "you may sacrifice one or more
// creatures" (Plumb the Forbidden). ADR 0100 §3: for these cards the
// printed "you may" and "any number" mean the same thing, because
// sacrificing none is not paying, so all three are one mandatory clause
// whose count runs from zero with no printed ceiling.
//
// The caster names the permanents on cast_spell's sacrifice_ids, and
// the list's length IS the count — there is no separate announcement
// to disagree with it. The count is recorded on PaidCost.Sacrificed and
// read back with ctx.Sacrificed(); a per-sacrifice discount reads it at
// CR 601.2f through CostsLessPerSacrificed.
//
// Only the mandatory slot may carry it: Register refuses it in an
// optional cost, in an either/or branch and on any ability.
func SacrificeAnyNumberCost(label string, preds ...CardPredicate) *game.AdditionalCost {
	return &game.AdditionalCost{
		Sacrifice: sacrificeSpec(label, preds...).WithCount(0, 0),
		Label:     "Sacrifice " + label,
	}
}

// PayXLifeCost is "As an additional cost to cast this spell, pay X
// life" — Toxic Deluge, and the third shape the clause takes.
//
// The X is announced with the cast and is the SAME number the spell's
// text reads back with ctx.X(): Toxic Deluge pays X life and gives
// every creature -X/-X, and a card that could announce those
// separately would be a different card. Paid with the spell already
// on the stack, like the other two components, so anything watching
// the life loss triggers above it.
func PayXLifeCost() *game.AdditionalCost {
	return &game.AdditionalCost{PayLifeX: true, Label: "Pay X life"}
}

// BlightXCost is "As an additional cost to cast this spell, blight X.
// X can't be greater than the greatest toughness among creatures you
// control" — Soul Immolation (#2174), and the fourth shape the clause
// takes after discard, sacrifice and pay-X-life.
//
// X is announced with the cast (CR 107.3a) and is the SAME number the
// spell's text reads back with ctx.X(); the caster names the one
// creature that takes the X -1/-1 counters on cast_spell's blight_ids.
// The counters are placed at CR 601.2h with the spell already on the
// stack, so a creature that dies of them dies before the spell
// resolves, and the cost stays paid. See game.AdditionalCost.BlightX.
func BlightXCost() *game.AdditionalCost {
	return &game.AdditionalCost{BlightX: true, Label: "Blight X"}
}

// --- either/or additional costs (ADR 0100 §2, #1732) ---------------
//
// "As an additional cost to cast this spell, sacrifice an artifact or
// discard a card" is ONE mandatory cost with branches: the caster
// announces which branch (cast_spell's cost_branch, CR 601.2b) and the
// engine pays that branch as the mandatory entry of the plan. Each
// branch is an ordinary cost with a Key, which is what the card's
// resolution asks for (ctx.PaidCostBranch), built with the usual
// constructors and .Keyed:
//
//	AdditionalCost: EitherCost(
//	    SacrificeCost("an artifact", Artifact()).Keyed("sacrifice"),
//	    DiscardCost(1).Keyed("discard"),
//	), // Demand Answers
//	AdditionalCost: EitherCost(DiscardCost(1).Keyed("discard"), ManaAdditionalCost("{5}").Keyed("mana")), // Lightning Axe
//
// A branch may carry mana (ManaAdditionalCost), a discard, a
// fixed-count sacrifice, a fixed life payment (PayLifeCost) or a blight
// (BlightCost). Register refuses the rest: see checkEitherCost.

// EitherCost is an either/or additional cost over two or more branches,
// in printed order. Its Label spells the whole clause ("Discard a card
// or pay {5}") for the prompt heading.
func EitherCost(branches ...*game.AdditionalCost) *game.AdditionalCost {
	out := &game.AdditionalCost{}
	labels := make([]string, 0, len(branches))
	for _, b := range branches {
		if b == nil {
			panic("effects.EitherCost: a nil branch")
		}
		out.Either = append(out.Either, *b)
		labels = append(labels, b.Label)
	}
	out.Label = joinBranchLabels(labels)
	return out
}

// joinBranchLabels reads "Sacrifice a creature", "Discard a card" as
// "Sacrifice a creature or discard a card", and three branches as
// "A, b, or c" — the way the card prints the clause.
func joinBranchLabels(labels []string) string {
	for i := 1; i < len(labels); i++ {
		labels[i] = lowerFirst(labels[i])
	}
	switch len(labels) {
	case 0:
		return ""
	case 1:
		return labels[0]
	case 2:
		return labels[0] + " or " + labels[1]
	}
	out := ""
	for i, l := range labels {
		switch {
		case i == 0:
			out = l
		case i == len(labels)-1:
			out += ", or " + l
		default:
			out += ", " + l
		}
	}
	return out
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// ManaAdditionalCost is a branch that pays mana — Lightning Axe's "pay
// {5}". It joins the total at CR 601.2f like a kicker's mana does, so a
// cost modifier sees it (ADR 0100 §2). Only meaningful as a branch of
// EitherCost; Register refuses it anywhere else.
func ManaAdditionalCost(mana string) *game.AdditionalCost {
	return &game.AdditionalCost{ManaCost: mana, Label: "Pay " + mana}
}

// PayLifeCost is a branch that pays a fixed amount of life — Bitter
// Triumph's "pay 3 life" (CR 119.4). Refused at announce when the
// caster's life total is below it.
func PayLifeCost(n int) *game.AdditionalCost {
	return &game.AdditionalCost{PayLife: n, Label: fmt.Sprintf("Pay %d life", n)}
}

// BlightCost is a branch that blights (CR 701.68a) — Wild Unraveling's
// "blight 2 or pay {1}": put N -1/-1 counters on a creature you
// control. The caster names the creature on cast_spell as blight_ids,
// exactly as for an optional blight (OptionalBlight).
func BlightCost(n int) *game.AdditionalCost {
	return &game.AdditionalCost{Blight: n, Label: fmt.Sprintf("Blight %d", n)}
}

// RevealCardCost is a branch that reveals a card — Wren's Run
// Vanquisher's "reveal an Elf card from your hand" (CR 701.20). The
// caster names the card on cast_spell as reveal_ids; paying shows it to
// the table and moves nothing. `article` is the printed "a" or "an".
func RevealCardCost(article, subtype string) *game.AdditionalCost {
	return &game.AdditionalCost{
		Reveal: &game.RevealCost{Subtype: subtype},
		Label:  "Reveal " + article + " " + subtype + " card from your hand",
	}
}

// BeholdCost is a branch that beholds — Silvergill Mentor's "behold a
// Merfolk": choose a Merfolk you control or reveal a Merfolk card from
// your hand. The same reveal_ids pick as RevealCardCost, widened to the
// caster's permanents; a permanent chosen is not revealed.
func BeholdCost(article, subtype string) *game.AdditionalCost {
	return &game.AdditionalCost{
		Reveal: &game.RevealCost{Subtype: subtype, Behold: true},
		Label:  "Behold " + article + " " + subtype,
	}
}

// checkEitherCost is Register's guard for an either/or cost (ADR 0100
// §2). Each refused shape compiles and then pays something the card
// does not print:
//
//   - a branched cost that also has components of its own;
//   - fewer than two branches;
//   - a branch that is empty, Optional, repeating or itself branched;
//   - a missing or duplicate branch Key;
//   - a variable sacrifice clause in a branch (no printed card has
//     one, and it would reopen the flat sacrifice_ids width question);
//   - a branch component the cast path has no shape for — teamwork, a
//     gift's opponent, "pay X life";
//   - an unparseable branch mana cost.
//
// And, outside a branch, the two components that exist only for one:
// a fixed PayLife in an optional cost and a mandatory mana cost.
func checkEitherCost(spec Spec) {
	for _, oc := range spec.OptionalCosts {
		if oc.PayLife != 0 || len(oc.Either) > 0 || oc.Reveal != nil {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q pays fixed life or has branches — both exist only on a branch of the mandatory EitherCost (ADR 0100)", spec.Name, oc.Key))
		}
	}
	ac := spec.AdditionalCost
	if ac == nil {
		return
	}
	if !ac.Branched() {
		if ac.Reveal != nil {
			panic(fmt.Sprintf("effects.Register: %q declares a reveal / behold cost outside an either/or branch — every printed one is a branch beside mana (ADR 0100 amendment 2026-10-07)", spec.Name))
		}
		if ac.ManaCost != "" {
			panic(fmt.Sprintf("effects.Register: %q declares a mandatory additional MANA cost — mana on the mandatory slot is a branch of EitherCost (ADR 0100); a single mana cost is part of the mana cost", spec.Name))
		}
		return
	}
	own := *ac
	own.Either = nil
	own.Label = ""
	if !own.Empty() || own.Key != "" || own.Targets != nil {
		panic(fmt.Sprintf("effects.Register: %q declares an either/or cost with components of its own — every component belongs to a branch (ADR 0100)", spec.Name))
	}
	if len(ac.Either) < 2 {
		panic(fmt.Sprintf("effects.Register: %q declares an either/or cost with %d branch(es) — build it with EitherCost over two or more", spec.Name, len(ac.Either)))
	}
	seen := make(map[string]bool, len(ac.Either))
	for i, b := range ac.Either {
		where := fmt.Sprintf("either/or branch %d (%q)", i, b.Key)
		switch {
		case b.Empty():
			panic(fmt.Sprintf("effects.Register: %q %s demands nothing", spec.Name, where))
		case b.Optional:
			panic(fmt.Sprintf("effects.Register: %q %s is Optional — a branch is paid when it is announced", spec.Name, where))
		case b.MaxPayments() > 1:
			panic(fmt.Sprintf("effects.Register: %q %s repeats — a branch is paid once", spec.Name, where))
		case len(b.Either) > 0:
			panic(fmt.Sprintf("effects.Register: %q %s is itself branched", spec.Name, where))
		case b.Key == "":
			panic(fmt.Sprintf("effects.Register: %q either/or branch %d has no Key — build it with .Keyed(\"…\")", spec.Name, i))
		case seen[b.Key]:
			panic(fmt.Sprintf("effects.Register: %q declares two either/or branches keyed %q", spec.Name, b.Key))
		case b.Teamwork != 0 || b.ChoosesOpponent || b.PayLifeX || b.BlightX || b.Targets != nil:
			panic(fmt.Sprintf("effects.Register: %q %s carries a component a branch has no shape for (teamwork, gift, pay X life, a target rewrite)", spec.Name, where))
		case b.Reveal != nil && b.Reveal.Subtype == "":
			panic(fmt.Sprintf("effects.Register: %q %s reveals a card of no type", spec.Name, where))
		case b.Blight < 0 || b.PayLife < 0 || b.DiscardCards < 0:
			panic(fmt.Sprintf("effects.Register: %q %s has a negative component", spec.Name, where))
		}
		seen[b.Key] = true
		if b.ManaCost != "" {
			if _, err := game.ParseCost(b.ManaCost); err != nil {
				panic(fmt.Sprintf("effects.Register: %q %s declares an unparseable mana cost %q: %v", spec.Name, where, b.ManaCost, err))
			}
		}
		checkSacrificeClause(spec.Name, where, b.Sacrifice, false, false, false)
	}
}

// checkVariableSacrificePlan is Register's cross-slot rule for a
// variable sacrifice on a cast (ADR 0100 §3):
//
//	A cast's payment plan may hold at most one variable sacrifice
//	clause, and if it holds one, no other entry in the plan may carry
//	a sacrifice.
//
// The plan is walked against ONE flat sacrifice_ids list, and a
// variable clause has no printed width, so the validator hands it
// whatever is left of the list (game.validateAdditionalCostLocked).
// That is only unambiguous when nothing else in the plan sacrifices.
// checkSacrificeClause already keeps a variable clause out of every
// slot but the mandatory one (an either/or branch and an optional cost
// both refuse it); this refuses the rest — the mandatory variable
// clause beside a sacrificing kicker or buyback — plus "sacrifice X"
// beside "pay X life", which no card prints (the ADR 0073 Decision 9
// posture). Each would compile and then be paid as something the card
// does not print.
//
// "Sacrifice X" beside an {X} in the MANA cost is not refused here,
// because the mana cost is the printing's, not the Spec's; no card in
// the dump prints it (ADR 0100 §3, the variable-sacrifice count), and
// if one did CR 107.3i would make the two the same number, which is
// what CastSpellParams.XValue already is.
func checkVariableSacrificePlan(spec Spec) {
	ac := spec.AdditionalCost
	if ac == nil || !game.SacrificeCostVariable(ac.Sacrifice) {
		return
	}
	for _, oc := range spec.OptionalCosts {
		if oc.Sacrifice != nil {
			panic(fmt.Sprintf("effects.Register: %q pairs a variable sacrifice with optional cost %q's sacrifice — a cast's plan may hold one variable sacrifice clause and no other sacrifice (ADR 0100 §3)", spec.Name, oc.Key))
		}
	}
	if !game.SacrificeCountFromX(ac.Sacrifice) {
		return
	}
	if ac.PayLifeX {
		panic(fmt.Sprintf("effects.Register: %q sacrifices X and pays X life — one announced X cannot pay both, and no card prints the pair (ADR 0100 §3)", spec.Name))
	}
}

// --- optional additional costs (ADR 0073, #664) -------------------
//
// An optional cost is the SAME struct with Optional set: what changes
// is that the caster chooses at CR 601.2b whether to pay it, and the
// choice is recorded so the resolution — or an entering permanent's
// own trigger — can read it back. One constructor per keyword, for
// the reason Overload and Evoke have one: the keyword carries the Key
// the engine reads, and a hand-rolled game.AdditionalCost{Optional:
// true} compiles and then never returns a bought-back card to hand.

// OptionalAdditionalCostKey is the key of a plain optional additional
// cost that is no keyword — "As an additional cost to cast this spell,
// you may pay {2}{R}" (Undergrowth). It is not a kicker, so nothing that
// asks "was it kicked" sees it.
const OptionalAdditionalCostKey = "additional"

// OptionalAdditionalMana is that cost: "As an additional cost to cast
// this spell, you may pay <mana>" (CR 118.8b, 601.2b). The caster
// chooses as the spell is cast; read it back with
// ctx.OptionalCostTimes(OptionalAdditionalCostKey) > 0, which is "if
// this spell's additional cost was paid".
//
//	OptionalCosts: []game.AdditionalCost{OptionalAdditionalMana("{2}{R}")},   // Undergrowth
func OptionalAdditionalMana(mana string) game.AdditionalCost {
	return game.AdditionalCost{
		Optional: true,
		Key:      OptionalAdditionalCostKey,
		ManaCost: mana,
		Label:    "Pay " + mana,
	}
}

// Kicker is CR 702.33's "Kicker [cost]" — "you may pay an additional
// [cost] as you cast this spell". Read back at resolution with
// ctx.WasKicked(), and from an entering permanent's own trigger with
// game.CardKickedTimes(*source).
//
//	OptionalCosts: []game.AdditionalCost{Kicker("{4}")},   // Burst Lightning
func Kicker(mana string) game.AdditionalCost {
	return game.AdditionalCost{
		Optional: true,
		Key:      game.KickerKey,
		ManaCost: mana,
		Label:    "Kicker " + mana,
	}
}

// Kickers is CR 702.33b's "Kicker [cost 1] and/or [cost 2]", which
// "means the same thing as 'Kicker [cost 1]' and 'Kicker [cost 2]'" —
// so it is exactly two Kicker entries, in printed order, both keyed
// game.KickerKey. Each is a separate once-only toggle: either, both or
// neither may be paid (#2153).
//
//	OptionalCosts: Kickers("{R}", "{W}"),   // Thornscape Battlemage
//
// Every "was it kicked" read counts both (ctx.WasKicked,
// ctx.KickedTimes, game.CardKickedTimes — "kicked twice" is 2), and
// CR 702.33f's "if it was kicked with its {R} kicker" is
// ctx.KickedWith("{R}") on a spell and ThisKickedWith("{R}") on a
// permanent's own trigger. Register refuses a third kicker and two
// with the same mana, which would leave that question unanswerable.
func Kickers(first, second string) []game.AdditionalCost {
	return []game.AdditionalCost{Kicker(first), Kicker(second)}
}

// ThisKickedAtLeast is the intervening if (CR 603.4) "if it was
// kicked" (n = 1) or "if it was kicked twice" (n = 2, Archangel of
// Wrath) on a permanent's own trigger, read off the record the
// resolution carried onto it (CR 400.7d). A permanent that was not
// cast — reanimated, flickered — was not kicked.
func ThisKickedAtLeast(n int) When {
	return func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return game.CardKickedTimes(*source) >= n
	}
}

// ThisKickedWith is CR 702.33f's "if it was kicked with its [cost]
// kicker" on a permanent's own trigger — Thornscape Battlemage's two
// enters abilities, each linked to one of its kicker costs (#2153).
// The cost is spelled as the card's Kickers declaration spells it.
func ThisKickedWith(cost string) When {
	return func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return game.CardKickedWith(*source, cost)
	}
}

// checkKickers holds a card's kicker costs to the shapes CR 702.33
// prints (#2153): one kicker, or "Kicker [A] and/or [B]" (CR 702.33b)
// — never three. Two kickers must each carry mana and differ in it,
// because CR 702.33f's "kicked with its [A] kicker" is answered by
// the mana (game.KickedWithPaid), and two that matched would make
// one of a card's linked abilities unreachable. No printed card has
// a non-mana kicker beside a second kicker.
func checkKickers(spec Spec) {
	kickers := game.KickerIndices(spec.OptionalCosts)
	switch {
	case len(kickers) <= 1:
		return
	case len(kickers) > 2:
		panic(fmt.Sprintf("effects.Register: %q declares %d kicker costs — CR 702.33b prints at most two (\"Kicker [A] and/or [B]\")", spec.Name, len(kickers)))
	}
	a, b := spec.OptionalCosts[kickers[0]], spec.OptionalCosts[kickers[1]]
	if a.ManaCost == "" || b.ManaCost == "" || a.ManaCost == b.ManaCost {
		panic(fmt.Sprintf("effects.Register: %q declares two kicker costs %q and %q — each must be a different mana cost, which is how \"kicked with its [cost] kicker\" names it (CR 702.33f)", spec.Name, a.Label, b.Label))
	}
}

// KickerSacrifice is a NON-MANA kicker — "Kicker—Sacrifice a
// creature" (Gatekeeper of Malakir). The same Sacrifice component
// every mandatory sacrifice cost uses, so it is validated by the same
// validator and paid with the spell already on the stack: a Blood
// Artist drains before the Gatekeeper resolves.
func KickerSacrifice(label string, preds ...CardPredicate) game.AdditionalCost {
	return game.AdditionalCost{
		Optional:  true,
		Key:       game.KickerKey,
		Sacrifice: sacrificeSpec(label, preds...),
		Label:     "Kicker—Sacrifice " + label,
	}
}

// Multikicker is CR 702.33d's "Multikicker [cost]" — "you may pay an
// additional [cost] any number of times as you cast this spell".
// Read back with ctx.KickedTimes() and game.CardKickedTimes.
//
// `max` is the engine's cap on ONE announcement. It is not printed on
// any card — multikicker is unbounded in paper — but an announcement
// has to be finite, the client's stepper has to stop somewhere, and
// the bot's expansion has to terminate. A cap far above what any
// board can pay for is the honest place to put that; say so in the
// card's own comment.
//
//	OptionalCosts: []game.AdditionalCost{Multikicker("{G}", 20)},  // Wolfbriar Elemental
func Multikicker(mana string, max int) game.AdditionalCost {
	if max < 1 {
		max = 1
	}
	return game.AdditionalCost{
		Optional: true,
		Key:      game.MultikickerKey,
		ManaCost: mana,
		Repeat:   max,
		Label:    "Multikicker " + mana,
	}
}

// Buyback is CR 702.27's "Buyback [cost]" — "you may pay an
// additional [cost] as you cast this spell. If the buyback cost was
// paid, put this card into its owner's hand as it resolves."
//
// The return is the ENGINE's, not the card's: the resolution path
// reads the paid record and routes the spell to its owner's hand
// through the same stack-exit primitive flashback uses. A card file
// declares the cost and nothing else — and must NOT also return
// itself in OnResolve, which would move a card that is still on the
// stack.
//
//	OptionalCosts: []game.AdditionalCost{Buyback("{3}")},   // Capsize
func Buyback(mana string) game.AdditionalCost {
	return game.AdditionalCost{
		Optional: true,
		Key:      game.BuybackKey,
		ManaCost: mana,
		Label:    "Buyback " + mana,
	}
}

// BuybackSacrifice is a NON-MANA buyback — "Buyback—Sacrifice a land"
// (Constant Mists). The reason #664 exists at all: without it
// Constant Mists is a {1}{G} Fog.
func BuybackSacrifice(label string, preds ...CardPredicate) game.AdditionalCost {
	return game.AdditionalCost{
		Optional:  true,
		Key:       game.BuybackKey,
		Sacrifice: sacrificeSpec(label, preds...),
		Label:     "Buyback—Sacrifice " + label,
	}
}

// WhenPaid declares the target clause a paid optional cost gives the
// spell, REPLACING the printed one — the kicker, teamwork and blight
// twin of Gift.Instead (#1716, ADR 0089 §3 amendment 2026-09-28):
//
//	Targets:       TargetCreature("target creature with mana value 3 or less", ManaValueLE(3)),
//	OptionalCosts: []game.AdditionalCost{WhenPaid(Teamwork(2), TargetCreature("target creature"))}, // Cruel Alliance
//
//	Targets:       TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
//	OptionalCosts: []game.AdditionalCost{WhenPaid(Kicker("{1}{B}"), TargetPermanent("target nonland permanent", Nonland()))}, // Tear Asunder
//
// "If this spell was kicked, instead destroy target creature" is a
// different clause, not a looser predicate on the same one: the cost
// is announced at CR 601.2b and the targets at 601.2c, so the clause
// the targets are chosen and judged against is the one the
// announcement produced. That is why this rewrites the whole clause
// rather than teaching a TargetSpec predicate to read the stack item:
// the predicate is also asked by the view and the enumerator, which
// have no stack item, and game.TargetSpecUnderOptionalCosts is the one
// function all four readers (announce, the CR 608.2b re-check and
// restore, the view, the enumerator) already share.
//
// A clause the unpaid spell does not have at all is the printed list
// with the extra clause appended, exactly as for gift.
func WhenPaid(cost game.AdditionalCost, clause *game.TargetSpec) game.AdditionalCost {
	if clause == nil {
		panic("effects.WhenPaid: a nil clause rewrites nothing — declare the cost on its own")
	}
	cost.Targets = clause
	return cost
}
