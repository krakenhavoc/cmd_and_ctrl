package game

import "github.com/google/uuid"

// tap_cost.go — S22: tapping untapped permanents you control as part
// of paying for a spell. Convoke (CR 702.51) and waterbend are the
// same mechanic under two names, so one component serves both.
//
// This is the fifth kind of cost the engine models, after a spell's
// mana cost (S15), an activated ability's cost (S21 sub-PR 2), an
// additional cost (S21 sub-PR 5) and an alternative cost (S22) — and
// it is a different SHAPE from all of them. The first four each
// answer "what does this cast owe?"; this one answers "what may be
// spent against what it owes". It is a cost *component*: it never
// stands alone, it reduces a mana demand that already exists.
//
//	Convoke  — "Your creatures can help cast this spell. Each
//	           creature you tap while casting this spell pays for {1}
//	           or one mana of that creature's color."
//	Waterbend {N} — "While paying a waterbend cost, you can tap your
//	           artifacts and creatures to help. Each one pays for
//	           {1}."
//
// Two differences, both modelled rather than papered over:
//
//   - Convoke's COLOUR CLAUSE is real. A tapped white creature may
//     pay a {W} in the cost, not just a generic {1}. Waterbend has no
//     such clause: each permanent pays for {1} and nothing else. That
//     is the ColorClause field, and it is the whole reason the
//     assignment below is a matching problem rather than a counter.
//   - Convoke helps pay the SPELL's cost; a waterbend cost is a
//     separate cost of its own ("as an additional cost … waterbend
//     {X}") that the tapped permanents help pay. Tapping three
//     creatures for Waterbender's Restoration's waterbend {X} does
//     not make its {U}{U} any cheaper. That is the Extra field: non-
//     empty means the keyword ADDS a mana demand which the tapped
//     permanents — and only the tapped permanents, plus mana — pay.
//
// Tapping this way is not the {T} symbol (CR 702.51b), so summoning
// sickness does not apply: a creature that entered this turn may
// still be convoked. There is deliberately no sickness check below.
//
// Modelled as a struct of components rather than a parsed cost
// string, for the same reason AbilityCost, AdditionalCost and
// AlternativeCost are: the shapes are few, and a mini-language would
// have to be maintained against a handful of cards.

// TapPermanentsCost is a card's "tap permanents you control to help
// pay" clause. The zero value, and nil, demand nothing.
//
// Named for what it does rather than for either keyword, because it
// is one component behind two names — and deliberately NOT `TapCost`,
// which is already an activated-ability cost constructor in the
// catalog package and a field on ManaAbilityShape.
type TapPermanentsCost struct {
	// Key is the stable wire identifier — "convoke" or "waterbend".
	// Unlike an alternative cost's Key this is not claimed by the
	// caster (there is nothing to choose between); it exists so the
	// client's picker and the event log can name the mechanic.
	Key string

	// Label is the clause as printed ("Convoke", "Waterbend {X}"),
	// shown above the client's picker so the prompt reads like the
	// card rather than like a schema.
	Label string

	// Spec is what may be tapped: convoke takes "creatures you
	// control", waterbend takes "artifacts and creatures you
	// control". Built with the same TargetSpec vocabulary a target
	// clause uses, so "you control" and the type predicate are
	// enforced by the one validator — but a cost is not a target
	// (CR 601.2h), so nothing here can be responded to and hexproof
	// never applies.
	//
	// The spec is expected to exclude already-tapped permanents; the
	// validator re-checks Tapped anyway, because a spec that forgot
	// would otherwise let one permanent pay twice.
	Spec *TargetSpec

	// ColorClause is convoke's "or one mana of that creature's
	// color". When set, a tapped permanent may pay a colored symbol
	// in the cost whose options include one of its colors; when
	// clear (waterbend) every tapped permanent pays exactly {1}.
	//
	// A colorless creature has no color, so it can only ever pay
	// {1} — which falls out of the matching below rather than
	// needing a special case.
	ColorClause bool

	// Extra is the mana the keyword itself demands, on top of the
	// card's printed cost, in the same brace notation Card.ManaCost
	// uses. "{X}" for Waterbender's Restoration's "waterbend {X}";
	// empty for convoke, which adds nothing and simply helps pay
	// what the card already costs.
	//
	// Only the GENERIC half of Extra can be paid by tapping — that
	// is what "each one pays for {1}" means. Every printed waterbend
	// cost is generic-only, so in practice the whole of it is
	// payable that way.
	Extra string
}

// Empty reports whether the cost offers nothing. Nil-safe, so the
// cast path can ask without a guard.
func (c *TapPermanentsCost) Empty() bool {
	return c == nil || c.Spec == nil
}

// DemandsX reports whether the keyword's own cost carries an {X} the
// caster must announce (waterbend {X}). Drives the client's X prompt
// for a card whose PRINTED cost has no {X} in it — Waterbender's
// Restoration costs {U}{U} and still needs one.
func (c *TapPermanentsCost) DemandsX() bool {
	if c.Empty() || c.Extra == "" {
		return false
	}
	extra, err := ParseCost(c.Extra)
	if err != nil {
		return false
	}
	return extra.XSlots > 0
}

// CatalogTapPermanentsCost is the catalog hook the effects package
// wires at init, mirroring CatalogAdditionalCost and
// CatalogAlternativeCosts. Nil, or a nil return, means the card has
// no such cost — which is nearly every card.
var CatalogTapPermanentsCost func(oracleID string) *TapPermanentsCost

// TapPermanentsCostFor returns a card's tap-permanents cost, or nil.
func TapPermanentsCostFor(oracleID string) *TapPermanentsCost {
	if CatalogTapPermanentsCost == nil || oracleID == "" {
		return nil
	}
	return CatalogTapPermanentsCost(oracleID)
}

// tapPermanentsBudget is how many permanents the cost will accept —
// the number of mana symbols the tapping is allowed to pay for. You
// may not tap more creatures for convoke than the spell costs
// (CR 702.51a: each tapped creature pays for one mana, and you can't
// pay more than the cost), and you may not tap more than the
// waterbend cost is worth.
//
// `cost` is the mana the cast owes BEFORE any tap payment, with the
// keyword's own Extra already excluded — the caller supplies the
// cost it wants the budget measured against.
func tapPermanentsBudget(c *TapPermanentsCost, cost ParsedCost, xValue int) int {
	if c.Empty() {
		return 0
	}
	if c.Extra != "" {
		extra, err := ParseCost(c.Extra)
		if err != nil {
			return 0
		}
		// Only the generic half is payable by tapping.
		return extra.Generic + extra.XSlots*xValue
	}
	return cost.Generic + cost.XSlots*xValue + len(cost.Required)
}

// TapPermanentsBudgetFor is the view layer's read of the same
// budget, taken against a card's printed mana-cost string rather
// than against a ParsedCost the protocol package has no way to
// build. Returns 0 for a waterbend {X} cost at xValue 0, which is
// the client's cue to size the picker from the X the caster is about
// to announce.
//
// Pure; safe to call under the read lock the snapshot builder holds.
func TapPermanentsBudgetFor(c *TapPermanentsCost, manaCost string, xValue int) int {
	cost, err := ParseCost(manaCost)
	if err != nil {
		return 0
	}
	return tapPermanentsBudget(c, cost, xValue)
}

// tapPermanentsAdjusted layers a tap-permanents cost onto a parsed
// cost: it adds whatever mana the keyword itself demands (waterbend
// {X}) and subtracts what the tapped permanents pay.
//
// Pure — no game state, no lock. The caller resolves `payers` from
// the announce-time IDs; a payer that has since vanished simply
// isn't in the slice and pays nothing.
//
// {X} in the incoming cost is folded into Generic here (using the
// announced xValue) so the subtraction has a single concrete number
// to work against. That is exactly equivalent for the solver, which
// computes Generic + XSlots*xValue anyway.
func tapPermanentsAdjusted(cost ParsedCost, c *TapPermanentsCost, payers []Card, xValue int) ParsedCost {
	if c.Empty() {
		return cost
	}
	out := cost
	out.Required = append([]ColorRequirement(nil), cost.Required...)
	out.Generic += out.XSlots * xValue
	out.XSlots = 0

	if c.Extra != "" {
		// Waterbend: the keyword adds a cost of its own, and the
		// tapped permanents pay that cost and nothing else. The
		// spell's printed {U}{U} is untouched no matter how many
		// permanents are tapped.
		extra, err := ParseCost(c.Extra)
		if err != nil {
			return out
		}
		owed := extra.Generic + extra.XSlots*xValue
		paid := len(payers)
		if paid > owed {
			paid = owed
		}
		out.Generic += owed - paid
		out.Required = append(out.Required, extra.Required...)
		return out
	}

	// Convoke: the tapped creatures help pay the whole cost. Assign
	// each to a colored symbol it can pay when the colour clause
	// allows it, then let the leftovers pay generic.
	spent := 0
	if c.ColorClause && len(out.Required) > 0 {
		assigned := matchPayersToRequirements(out.Required, payers)
		kept := out.Required[:0:0]
		for i, req := range out.Required {
			if assigned[i] >= 0 {
				spent++
				continue
			}
			kept = append(kept, req)
		}
		out.Required = kept
	}
	generic := len(payers) - spent
	if generic > out.Generic {
		generic = out.Generic
	}
	if generic > 0 {
		out.Generic -= generic
	}
	return out
}

// matchPayersToRequirements assigns tapped permanents to colored
// mana symbols in the cost, returning for each requirement the index
// of the payer covering it, or -1.
//
// A maximum bipartite matching (Kuhn's augmenting-path algorithm),
// not a greedy walk, because greedy gets the wrong answer on a real
// board: a {W}{U} cost paid with a white-blue creature and a white
// creature has a two-creature assignment, but a greedy pass that
// hands the white-blue creature to {W} first leaves the mono-white
// creature with nowhere to go and convokes only one. The inputs are
// a handful of symbols against a handful of creatures, so the cost
// of getting it right is a few dozen comparisons.
//
// Colored symbols are matched before generic is considered because a
// colored symbol is strictly harder to pay from a mana pool than a
// generic one — spending a creature on {W} and leaving {1} to the
// lands is never worse for the caster than the reverse.
func matchPayersToRequirements(reqs []ColorRequirement, payers []Card) []int {
	assigned := make([]int, len(reqs))
	for i := range assigned {
		assigned[i] = -1
	}
	var augment func(payer int, seen []bool) bool
	augment = func(payer int, seen []bool) bool {
		for r := range reqs {
			if seen[r] || !payerCoversRequirement(payers[payer], reqs[r]) {
				continue
			}
			seen[r] = true
			if assigned[r] == -1 || augment(assigned[r], seen) {
				assigned[r] = payer
				return true
			}
		}
		return false
	}
	for p := range payers {
		augment(p, make([]bool, len(reqs)))
	}
	return assigned
}

// payerCoversRequirement reports whether a tapped permanent's colors
// let it pay one colored symbol — convoke's "one mana of that
// creature's color".
//
// A colorless permanent has no colors and so covers nothing, which
// is correct: it can still pay {1} as generic, but it can never pay
// a {C} or a colored symbol. Hybrid symbols carry both halves in
// Options, so either colour covers them.
func payerCoversRequirement(c Card, req ColorRequirement) bool {
	for _, color := range c.EffectiveColors() {
		if matchColor(color, req.Options) {
			return true
		}
	}
	return false
}

// tapPermanentsPayersLocked resolves announce-time IDs to the
// battlefield cards they name, skipping any that have since left.
// Read-only. Caller must hold g.mu.
func (g *Game) tapPermanentsPayersLocked(ids []uuid.UUID) []Card {
	if len(ids) == 0 || g.Battlefield == nil {
		return nil
	}
	out := make([]Card, 0, len(ids))
	for _, id := range ids {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				out = append(out, g.Battlefield.Cards[i])
				break
			}
		}
	}
	return out
}

// validateTapPermanentsCostLocked checks that every permanent the
// caster named may actually be tapped for this cost, without tapping
// anything — the same validate-all-then-pay discipline the other
// cost components use, so a rejected cast never leaves a board half
// tapped.
//
// A card with no such cost that arrives WITH tap IDs is a client
// bug, not a no-op: rejecting it keeps the wire honest, exactly as
// the additional-cost validator does.
//
// `budget` caps how many may be tapped; over-tapping is a rejection
// rather than a silent truncation, because the player would have
// tapped a permanent for nothing.
//
// Caller must hold g.mu.
func (g *Game) validateTapPermanentsCostLocked(playerID uuid.UUID, c *TapPermanentsCost, ids []uuid.UUID, budget int) error {
	if c.Empty() {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	if len(ids) > budget {
		return ErrInvalidParam
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return ErrInvalidParam
		}
		seen[id] = true
		card := g.findBattlefieldCardLocked(id)
		if card == nil {
			return ErrCardNotFound
		}
		// "You control" and "untapped" are the two clauses every
		// tap-as-a-cost keyword shares; the spec carries the rest
		// (creature for convoke, artifact-or-creature for
		// waterbend).
		if card.Controller != playerID || card.Tapped {
			return ErrInvalidParam
		}
		if !g.targetLegalLocked(playerID, c.Spec, TargetRef{Kind: TargetCard, ID: id}) {
			return ErrInvalidParam
		}
	}
	return nil
}

// payTapPermanentsCostLocked taps the named permanents, emitting
// EventTapCard for each so anything watching taps sees them. Call
// only after validateTapPermanentsCostLocked has passed and after
// the spell itself has moved to the stack (CR 601.2a before
// 601.2h), so a tap-watching trigger goes on the stack above the
// spell and resolves first — the same ordering payAdditionalCostLocked
// exists to preserve.
//
// A permanent that vanished between announce and payment is skipped
// rather than erroring: the cast is already committed, and the
// caster simply got less help than they asked for.
//
// Caller must hold g.mu.
func (g *Game) payTapPermanentsCostLocked(playerID uuid.UUID, ids []uuid.UUID) {
	for _, id := range ids {
		card := g.findBattlefieldCardLocked(id)
		if card == nil || card.Tapped {
			continue
		}
		card.Tapped = true
		g.EmitEvent(Event{Kind: EventTapCard, Actor: playerID, CardID: id})
	}
}

// findBattlefieldCardLocked returns a pointer into the live
// battlefield slice for the given instance ID, or nil. Unlike
// findCardByIDLocked this looks in one zone only — a cost may only
// tap a permanent, and a "permanent" in exile is not one.
//
// Caller must hold g.mu.
func (g *Game) findBattlefieldCardLocked(cardID uuid.UUID) *Card {
	if g.Battlefield == nil {
		return nil
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}
