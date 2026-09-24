package game

import (
	"fmt"

	"github.com/google/uuid"
)

// additional_cost.go — S21: "As an additional cost to cast this
// spell, discard a card" (sub-PR 5) or "sacrifice a creature"
// (sub-PR 6) — CR 601.2f–h. The third kind of cost
// the engine knows about, after a spell's mana cost (S15) and an
// activated ability's cost (S21 sub-PR 2).
//
// The distinction worth modelling: an additional cost is paid to
// CAST the spell, not during its resolution. So it is paid even if
// the spell is countered, and — the part that changes gameplay —
// the discard happens while the spell is still on the stack, which
// means Mary Read and Anne Bonny, Marauding Mako and Glint-Horn
// Buccaneer all see it and trigger BEFORE the spell resolves.
// Folding the discard into OnResolve would get the card draw right
// and the triggers wrong.
//
// Modelled as a struct of components rather than a parsed cost
// string, for the same reason AbilityCost is: the shapes are few,
// and a second mini-language would have to be maintained against a
// handful of cards.

// AdditionalCost is what a spell demands on top of its mana cost.
// The zero value demands nothing.
type AdditionalCost struct {
	// DiscardCards is "discard a card" (Thrill of Possibility) or
	// "discard two cards". The caster names them in
	// CastSpellParams.DiscardIDs. The spell being cast is never a
	// legal choice: CR 601.2a moves it to the stack before costs
	// are paid, so it is no longer in hand.
	DiscardCards int

	// Sacrifice is "sacrifice a creature" (Village Rites, Altar's
	// Reap) or "sacrifice an artifact or creature" (Deadly Dispute),
	// as a spec matched against the caster's permanents. The caster
	// names them in CastSpellParams.SacrificeIDs — exactly the
	// clause's count, which is 1 unless the spec says otherwise
	// ("sacrifice two creatures", effects.SacrificeNCost, #747).
	//
	// Same reasoning as DiscardCards, one zone over: the creature
	// dies while the spell is on the stack, so a Blood Artist or
	// Zulaport Cutthroat trigger goes ABOVE the spell and drains
	// before it resolves. Sacrificing in OnResolve inverts that, and
	// also hands the creature back when the spell is countered.
	//
	// Unlike a discard, the spell being cast is never a candidate for
	// a different reason: it is on the stack, not the battlefield.
	Sacrifice *TargetSpec

	// PayLifeX is "As an additional cost to cast this spell, pay X
	// life" (Toxic Deluge), where X is the value announced on
	// CastSpellParams.XValue. Added in S23.
	//
	// The X here is NOT a mana-cost X — Toxic Deluge prints {2}{B}
	// with no {X} anywhere in it. It is a free variable the caster
	// names at announce whose only consumer is this clause and the
	// spell's own text ("all creatures get -X/-X"), which is why it
	// rides the existing XValue slot rather than growing a second
	// one: the two are the same number by definition, and a card
	// that announced them separately could set them differently.
	//
	// CR 119.4 — paying life is legal only when the life total is at
	// least the amount, so the cast is rejected at announce when it
	// isn't. Paying 0 is always legal and always a no-op.
	PayLifeX bool

	// Label is the cost clause as printed ("Discard a card"), shown
	// in the client's cost picker so the prompt reads like the card
	// rather than like a schema.
	Label string

	// Optional marks a cost the caster CHOOSES whether to pay while
	// announcing the spell (CR 601.2b) — kicker (CR 702.33), buyback
	// (CR 702.27). ADR 0073 §1.
	//
	// A flag on this struct rather than a type of its own, for the
	// reason ADR 0021 gave the sacrifice component: Constant Mists'
	// "Buyback—Sacrifice a land" IS the Sacrifice clause above with
	// this bool set, and a separate type would have had to grow its
	// own copy of every component and its own validator. What changes
	// is WHEN the cost is settled, not what it is made of.
	//
	// An optional cost is declared in Spec.OptionalCosts and is
	// INDEXED there, because the announcement names which ones were
	// paid; the mandatory one stays in Spec.AdditionalCost. Register
	// cross-checks the flag against the slot, so the two cannot
	// disagree.
	Optional bool

	// Key is the stable identity of an optional cost — "kicker",
	// "multikicker", "buyback". It never crosses the wire (the
	// announcement names indices, not keys) and it is not what makes
	// the cost payable; it is what a card's own OnResolve asks for
	// through ctx.WasKicked / ctx.OptionalCostTimes, so a resolution
	// never has to know its own declaration order.
	//
	// Required and unique per card on an optional cost, meaningless
	// on a mandatory one. Register enforces both.
	Key string

	// ManaCost is the mana half of the clause — kicker {4}, buyback
	// {3} — in the same Scryfall brace notation Card.ManaCost uses.
	// Empty for a purely non-mana cost (Constant Mists' buyback,
	// Gatekeeper of Malakir's kicker).
	//
	// The one component ADR 0073 genuinely adds rather than reuses:
	// every additional cost the engine had before was non-mana,
	// because ADR 0021 shipped discard and sacrifice. It joins the
	// total at CR 601.2f, where an additional cost has always gone —
	// after the alternative-cost swap and the commander tax, before
	// the cost modifiers.
	ManaCost string

	// Repeat is CR 702.33d's multikicker: the maximum number of times
	// this cost may be paid for one cast. 0 and 1 both mean "once".
	//
	// Only a MANA-ONLY cost may repeat. Register panics on a Repeat
	// above 1 that also carries a card- or permanent-shaped
	// component, because every printed multikicker is mana and N
	// independent card payments per cast is a wire shape nothing asks
	// for (ADR 0073 §4).
	Repeat int

	// ChoosesOpponent is gift's cost (CR 702.174a): "As an additional
	// cost to cast this spell, you may choose an opponent." Paying it
	// is NAMING a player rather than handing anything over — the
	// caster sends the opponent on CastSpellParams.GiftOpponent, the
	// engine checks it is an opponent still in the game, and the
	// choice lands on PaidCost.GiftOpponent. ADR 0089 §1.
	//
	// A component of this struct for the reason Optional is a flag
	// rather than a type: CR 702.174a says paying it "follows the
	// rules for paying additional costs in rules 601.2b and
	// 601.2f–h", so it is announced, recorded, copied and carried
	// onto a permanent exactly as a kicker is — only what it is made
	// of differs. Register allows it only on an optional cost keyed
	// GiftKey, and only on its own.
	ChoosesOpponent bool

	// Targets, when non-nil, REPLACES the spell's target clause when
	// this optional cost is paid — Long River's Pull's "counter target
	// creature spell. If the gift was promised, instead counter target
	// spell", Wear Down's "instead destroy two target artifacts and/or
	// enchantments". The same shape AlternativeCost.Targets gives
	// cleave, for the same reason: the cost is announced at CR 601.2b
	// and the targets at 601.2c, so the clause the targets are judged
	// against is the one the announcement produced (CR 702.174m, ADR
	// 0089 §3).
	//
	// A clause the unpaid spell does not have at all (Mind Spiral's
	// "if the gift was promised, tap target creature an opponent
	// controls") is the printed clause list with the extra clause
	// appended — CR 702.174m's "chooses those targets only if the gift
	// was promised" is exactly a clause list that differs by the
	// cost.
	Targets *TargetSpec
}

// MaxPayments is how many times this cost may be paid for one cast:
// 1 for a mandatory or ordinary optional cost, Repeat for a
// multikicker. Nil-safe.
func (c *AdditionalCost) MaxPayments() int {
	if c == nil || c.Repeat < 1 {
		return 1
	}
	return c.Repeat
}

// Empty reports whether the cost demands nothing. Nil-safe.
func (c *AdditionalCost) Empty() bool {
	return c == nil || (c.DiscardCards == 0 && c.Sacrifice == nil && !c.PayLifeX && c.ManaCost == "" && !c.ChoosesOpponent)
}

// CardsDemanded reports whether paying this cost needs the caster to
// NAME something — cards to discard, permanents to sacrifice. A
// mana-only cost needs no payment list, which is what lets a
// multikicker be paid N times off one announcement (ADR 0073 §4).
// Nil-safe.
func (c *AdditionalCost) CardsDemanded() bool {
	return c != nil && (c.DiscardCards > 0 || c.Sacrifice != nil || c.PayLifeX)
}

// CatalogAdditionalCost is the catalog hook the effects package
// wires at init, mirroring CatalogTargetSpec and CatalogModeSpec.
// Nil, or a nil return, means the card has no additional cost.
var CatalogAdditionalCost func(oracleID string) *AdditionalCost

// AdditionalCostFor returns a card's additional cost, or nil.
func AdditionalCostFor(oracleID string) *AdditionalCost {
	if CatalogAdditionalCost == nil || oracleID == "" {
		return nil
	}
	return CatalogAdditionalCost(oracleID)
}

// CatalogOptionalCosts is the catalog hook for the costs a card lets
// the caster CHOOSE to pay (ADR 0073) — kicker, multikicker, buyback.
// The slice order is the card's declaration order, and it is the
// INDEX SPACE the announcement names: CastSpellParams.OptionalCosts
// and PaidCost.OptionalCosts both hold positions in this slice.
//
// A separate hook from CatalogAdditionalCost, not a widening of it,
// because the mandatory cost has no index and ~20 card files already
// read it as "the one cost this spell demands".
var CatalogOptionalCosts func(oracleID string) []AdditionalCost

// OptionalCostsFor returns the optional additional costs a card
// offers, in declaration order, or nil.
func OptionalCostsFor(oracleID string) []AdditionalCost {
	if CatalogOptionalCosts == nil || oracleID == "" {
		return nil
	}
	return CatalogOptionalCosts(oracleID)
}

// The three optional-cost keys the engine itself knows about. A card
// may use any key it likes — the engine only ever compares strings —
// but these three are the ones the ENGINE reads rather than the card:
// kicker and multikicker because CR 702.33 counts them together, and
// buyback because CR 702.27b changes where the spell goes.
const (
	// KickerKey is CR 702.33's kicker, paid at most once.
	KickerKey = "kicker"
	// MultikickerKey is CR 702.33d's multikicker, paid any number of
	// times. Counted with KickerKey by KickedTimesPaid, because "the
	// number of times it was kicked" does not distinguish them.
	MultikickerKey = "multikicker"
	// BuybackKey is CR 702.27's buyback. Read by the stack-exit route
	// and by nothing else.
	BuybackKey = "buyback"
	// GiftKey is CR 702.174's gift. The engine reads it in one place:
	// the announce check that a gift cost names an opponent
	// (validateGiftChoiceLocked). Everything the gift DOES is the
	// card's, grown by effects.Gift (ADR 0089).
	GiftKey = "gift"
)

// OptionalCostTimesPaid counts how many times the optional cost whose
// Key is `key` was paid, given the card it belongs to and a list of
// paid indices — PaidCost.OptionalCosts for a spell on the stack,
// Card.Provenance.OptionalCosts for a permanent that has entered.
//
// Keyed rather than indexed so a card's own resolution never has to
// know its declaration order, and so the two readers above share one
// lookup.
func OptionalCostTimesPaid(card Card, paid []int, key string) int {
	if len(paid) == 0 || key == "" {
		return 0
	}
	costs := OptionalCostsFor(CatalogKey(card))
	if len(costs) == 0 {
		return 0
	}
	n := 0
	for _, i := range paid {
		if i >= 0 && i < len(costs) && costs[i].Key == key {
			n++
		}
	}
	return n
}

// KickedTimesPaid is CR 702.33's "the number of times it was kicked":
// kicker and multikicker together, because no rules text tells them
// apart and no card prints both.
func KickedTimesPaid(card Card, paid []int) int {
	return OptionalCostTimesPaid(card, paid, KickerKey) +
		OptionalCostTimesPaid(card, paid, MultikickerKey)
}

// CardKickedTimes is KickedTimesPaid for a permanent that has already
// entered, reading the record the resolution path carried onto it
// (CR 400.7d, ADR 0073 §5). Zero for a permanent that did not arrive
// by resolving a kicked spell — including one reanimated out of a
// graveyard, whose record was cleared on the way out.
func CardKickedTimes(c Card) int {
	return KickedTimesPaid(c, c.Provenance.OptionalCosts)
}

// costPayment is one component of a cast's CR 601.2f cost: the card's
// mandatory additional cost, or ONE payment of one optional cost. A
// multikicker paid three times contributes three entries.
//
// The plan exists so there is still exactly one validator and one
// payer (ADR 0073 §4): the caster's discard_ids and sacrifice_ids are
// flat wire lists, and the plan is the ORDER they are walked in —
// mandatory first, then each chosen optional cost in index order.
type costPayment struct {
	cost AdditionalCost
	// index is -1 for the mandatory cost, else the cost's position in
	// the card's OptionalCosts slice.
	index int
}

// castCostPayments builds the CR 601.2f payment plan for one cast:
// the mandatory cost (if any), then one entry per announced payment
// of an optional cost, in index order so the flat payment lists are
// walked deterministically whatever order the client sent the
// indices in.
//
// `chosen` is CastSpellParams.OptionalCosts, already validated by
// validateOptionalCostChoice. A nil mandatory cost and an empty
// choice produce a nil plan, which is every cast in the catalog that
// predates ADR 0073.
func castCostPayments(mandatory *AdditionalCost, optional []AdditionalCost, chosen []int) []costPayment {
	var plan []costPayment
	if !mandatory.Empty() {
		plan = append(plan, costPayment{cost: *mandatory, index: -1})
	}
	if len(chosen) == 0 || len(optional) == 0 {
		return plan
	}
	counts := make([]int, len(optional))
	for _, i := range chosen {
		if i >= 0 && i < len(counts) {
			counts[i]++
		}
	}
	for i, n := range counts {
		for range n {
			plan = append(plan, costPayment{cost: optional[i], index: i})
		}
	}
	return plan
}

// paidWithOptionalCosts folds the plan's optional-cost payments into
// the payment record that lands on the stack item. The indices come
// out of the plan rather than straight off the wire, so they arrive
// in ascending order however the client sent them — two clients
// announcing the same multikicker produce the same record, which is
// what makes the snapshot round-trip and the undo comparison stable.
func paidWithOptionalCosts(paid PaidCost, plan []costPayment) PaidCost {
	for _, pay := range plan {
		if pay.index >= 0 {
			paid.OptionalCosts = append(paid.OptionalCosts, pay.index)
		}
	}
	return paid
}

// paidWithGift folds the gift recipient into the record (ADR 0089
// §2). Already validated: non-nil exactly when a gift cost was
// announced.
func paidWithGift(paid PaidCost, to uuid.UUID) PaidCost {
	paid.GiftOpponent = to
	return paid
}

// paidWithSacrifices folds the number of permanents the additional
// cost sacrificed into the record that lands on the stack item
// (#1213) — the same shape and the same reason as the fold above:
// by resolution the permanents are in graveyards, so the number has to
// be a fact about the announcement rather than a count of the board.
func paidWithSacrifices(paid PaidCost, n int) PaidCost {
	paid.Sacrificed = n
	return paid
}

// validateOptionalCostChoice checks the ANNOUNCEMENT itself (CR
// 601.2b) before anything is priced or paid: every index names a cost
// the card offers, and no cost is named more times than it may be
// paid (CR 702.33d's multikicker cap, 1 for everything else).
//
// Pure, so the bot enumerator and the view can ask the same question
// the cast path asks.
func validateOptionalCostChoice(optional []AdditionalCost, chosen []int) error {
	if len(chosen) == 0 {
		return nil
	}
	if len(optional) == 0 {
		return ErrInvalidParam
	}
	counts := make([]int, len(optional))
	for _, i := range chosen {
		if i < 0 || i >= len(optional) {
			return ErrInvalidParam
		}
		counts[i]++
		if counts[i] > optional[i].MaxPayments() {
			return ErrInvalidParam
		}
	}
	return nil
}

// validateAdditionalCostLocked checks that the caster named exactly
// the right cards to pay `plan`, without paying anything — the same
// validate-all-then-pay discipline ActivateCatalogAbility uses, so a
// rejected cast never leaves a half-paid cost behind. `castID` is
// the spell being cast, which is never a legal discard.
//
// The flat discard and sacrifice lists are walked in plan order and
// must be consumed EXACTLY: a card with no additional cost that
// arrives with discard IDs is a client bug, not a no-op, and so is a
// kicked Gatekeeper of Malakir that sends two creatures for a
// one-creature kicker.
//
// Caller must hold g.mu.
func (g *Game) validateAdditionalCostLocked(playerID, castID uuid.UUID, plan []costPayment, discardIDs, sacrificeIDs []uuid.UUID, xValue int) error {
	if len(plan) == 0 {
		if len(discardIDs) > 0 || len(sacrificeIDs) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	// One `seen` set across the whole plan: a card in hand pays one
	// discard and a permanent pays one sacrifice, however many
	// clauses are demanding them.
	discarded := make(map[uuid.UUID]bool, len(discardIDs))
	sacrificed := make(map[uuid.UUID]bool, len(sacrificeIDs))
	di, si := 0, 0
	for _, pay := range plan {
		cost := pay.cost
		// CR 119.4: a player may pay N life only with a life total of
		// at least N. Checked at announce with the rest of the
		// choices, so an unpayable X is a rejected cast rather than a
		// player at -3.
		if cost.PayLifeX && xValue > p.Life {
			return ErrInvalidParam
		}
		if di+cost.DiscardCards > len(discardIDs) {
			return ErrInvalidParam
		}
		for _, id := range discardIDs[di : di+cost.DiscardCards] {
			if id == castID || discarded[id] {
				return ErrInvalidParam
			}
			discarded[id] = true
			if !p.Hand.Contains(id) {
				return ErrCardNotFound
			}
		}
		di += cost.DiscardCards
		if cost.Sacrifice == nil {
			continue
		}
		n := SacrificeCostCount(cost.Sacrifice)
		if si+n > len(sacrificeIDs) {
			return ErrInvalidParam
		}
		slice := sacrificeIDs[si : si+n]
		for _, id := range slice {
			if sacrificed[id] {
				return ErrInvalidParam
			}
			sacrificed[id] = true
		}
		// The sacrifice clause reuses the activated-ability validator,
		// so "you may only sacrifice what you control" (CR 701.21a)
		// and the spec's own predicate are enforced in one place
		// rather than two.
		//
		// #1213: a cast's sacrifice clause is always a FIXED count —
		// effects.Register refuses a variable one here, because the
		// flat wire lists are walked in plan order and a clause with
		// no fixed width could not be split from the next one's
		// payment. So there is no announced X to pass, and 0 reads
		// the clause exactly as it did before.
		if _, err := g.validateSacrificeCostLocked(playerID, castID, AbilityCost{
			SacrificeOther: cost.Sacrifice,
		}, slice, 0); err != nil {
			return err
		}
		si += n
	}
	if di != len(discardIDs) || si != len(sacrificeIDs) {
		return ErrInvalidParam
	}
	return nil
}

// AddOptionalCostMana is the mana half of the announced optional
// costs, added into `cost` at CR 601.2f. ONE helper, so the cast
// path, the auto-tapper and the bot enumerator price a kicked spell
// identically and a bot is never offered a kicked move at the
// unkicked price (ADR 0073 §3, #544).
//
// Exported because the enumerator lives in another package and there
// is no second copy of this arithmetic to be had.
//
// An unparseable optional cost string refuses the cast for the same
// reason an unparseable printed cost does (#289): there is no price
// for the player to have paid. Register refuses the same string at
// boot, so reaching this branch means a card was built around the
// catalog.
func AddOptionalCostMana(cost ParsedCost, optional []AdditionalCost, chosen []int) (ParsedCost, error) {
	for _, i := range chosen {
		if i < 0 || i >= len(optional) || optional[i].ManaCost == "" {
			continue
		}
		add, err := ParseCost(optional[i].ManaCost)
		if err != nil {
			return cost, fmt.Errorf("%w for %s: %w", ErrUnparseableCost, optional[i].Label, err)
		}
		cost.Generic += add.Generic
		cost.Required = append(cost.Required, add.Required...)
		cost.XSlots += add.XSlots
		cost.HasPhyrexian = cost.HasPhyrexian || add.HasPhyrexian
		cost.HasSnow = cost.HasSnow || add.HasSnow
	}
	return cost, nil
}

// payAdditionalCostLocked pays the cost's components: sacrifices the
// named permanents (emitting EventSacrifice and routing them to their
// owners' graveyards, so aristocrats payoffs trigger) and discards the
// named cards to their owner's graveyard, emitting EventDiscardCard
// for each so discard payoffs trigger. Call only after validateAdditionalCostLocked has passed
// and after the spell itself has moved to the stack (CR 601.2a
// before 601.2h), so a discard trigger sees the spell above it.
//
// Caller must hold g.mu.
func (g *Game) payAdditionalCostLocked(playerID uuid.UUID, discardIDs, sacrificeIDs []uuid.UUID, payLife int, answers map[uuid.UUID]bool) error {
	// Life first: it is the component with no choice attached, and
	// paying it before the sacrifices keeps the event order matching
	// the way the clauses are read aloud.
	//
	// #793: through the COST path. Paying life is losing life
	// (CR 119.4), so the CR 614 window still runs and a life-loss
	// replacement still sees it — but it settles without a prompt,
	// because CR 601.2h pays a spell's costs as one indivisible step
	// and a paused ordering prompt here would leave the spell on the
	// stack half paid for.
	if payLife > 0 {
		if err := g.PayLifeForEffect(uuid.Nil, playerID, payLife); err != nil {
			return err
		}
	}
	// The N permanents of a "sacrifice two creatures" clause leave as
	// one simultaneous exit (#747, ADR 0021 addendum), so the order
	// the IDs arrived in cannot change what the watchers see.
	if err := g.payCostSacrificesLocked(sacrificeIDs, answers); err != nil {
		return err
	}
	if len(discardIDs) == 0 {
		return nil
	}
	if g.playerByIDLocked(playerID) == nil {
		return ErrPlayerNotFound
	}
	// Through the COST path of the one discard helper (discard.go):
	// a discard is a discard, but CR 601.2h pays a spell's costs as
	// one indivisible step, so this one may not pause. A commander
	// among the discards carries the CR 903.9 answer its owner gave
	// before the cast was paid for (#1397).
	return g.discardCardsLocked(playerID, discardIDs, discardOptions{
		cause:            DiscardCauseCost,
		commanderAnswers: answers,
	})
}
