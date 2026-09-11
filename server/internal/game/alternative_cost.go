package game

import "github.com/google/uuid"

// alternative_cost.go — S22: a cost paid *instead of* a spell's mana
// cost (CR 118.9). The fourth kind of cost the engine models, after
// a spell's mana cost (S15), an activated ability's cost (S21
// sub-PR 2) and an additional cost (S21 sub-PR 5).
//
// The distinction from AdditionalCost is the whole point of the
// file. An additional cost is paid ALONGSIDE the mana cost; an
// alternative cost is paid INSTEAD of it. "Overload {6}{U}" does not
// mean "{1}{U} and also {6}{U}" — it means the spell costs {6}{U}
// and nothing else. Additional costs survive the swap, because CR
// 601.2f is evaluated independently of the cost chosen at 601.2e, so
// a card that charged both would still charge both.
//
// Three things an alternative cost may carry beyond the price,
// because the keywords that grant one rarely stop there:
//
//   - Overload rewrites the targeting clause out of existence
//     ("change 'target' to 'each'"), so paying it must CLEAR the
//     spell's TargetSpec rather than widen it — otherwise the
//     announce gate would still demand a target the spell no longer
//     has. That is ClearsTargets.
//   - Cleave rewrites the clause into a different one ("remove the
//     words in square brackets"), so paying it SWAPS IN another
//     TargetSpec. That is Targets.
//   - Evoke attaches a triggered ability to the permanent's entry
//     ("it's sacrificed when it enters"). That is SacrificeOnEntry.
//
// Modelled as a slice on the card rather than a single struct: a
// card can offer more than one (spree, and the modal-cost cards),
// and growing a struct into a slice later would churn every
// signature.
//
// Modelled as components rather than a parsed cost string, for the
// same reason AbilityCost and AdditionalCost are: the shapes are
// few, and a cost mini-language has to be maintained against a
// handful of cards.

// AlternativeCost is one "you may cast this spell for X rather than
// its mana cost" option.
type AlternativeCost struct {
	// Key is the stable wire identifier the caster names to claim
	// this cost — "overload", "evoke", "cleave". Rides
	// CastSpellParams.AlternativeCost, lands on StackItem.AltCost,
	// and is read back at resolution so a card whose text changes
	// with the cost can branch on it. Required, and unique among a
	// card's alternative costs.
	Key string

	// Label is the clause as printed ("Overload {4}{R}"), shown in
	// the client's picker so the prompt reads like the card rather
	// than like a schema.
	Label string

	// ManaCost is what the caster pays instead of the printed cost,
	// in the same Scryfall brace notation Card.ManaCost uses. Empty
	// means free ("without paying its mana cost") — ParseCost reads
	// an empty string as the zero cost, which is exactly right.
	//
	// Commander tax still applies on top: CR 903.8 adds {2} per
	// prior cast to whatever the cost is, alternative or not. So
	// effectiveCostLocked layers the tax after the swap, not before.
	ManaCost string

	// Targets replaces the card's target clause when this cost is
	// paid. Cleave's "remove the words in square brackets" turns
	// "counter target spell that wasn't cast from its owner's hand"
	// into "counter target spell" — a different, wider clause, not
	// an absent one. Nil leaves the printed clause alone.
	Targets *TargetSpec

	// ClearsTargets removes the target clause outright — overload's
	// "change 'target' in its text to 'each'". A cast that claims
	// such a cost and still sends targets is rejected rather than
	// ignored: the spell has none, and a client that thinks
	// otherwise is confused about which cost it is paying.
	ClearsTargets bool

	// Condition gates the OFFER — "IF YOU CONTROL A COMMANDER, you
	// may cast this spell without paying its mana cost" (Fierce
	// Guardianship), "IF YOU CONTROL A SWAMP, you may pay 4 life
	// rather than pay this spell's mana cost" (Snuff Out).
	//
	// Checked at announce and again in the view, so an offer the
	// caster cannot take is neither shown nor accepted. Nil means
	// unconditional, which is what overload, evoke and cleave are.
	//
	// Read-only, under g.mu. Added in S28.
	Condition func(g *Game, controller uuid.UUID) bool

	// Life is a "pay N life" component of the alternative cost (CR
	// 118.4) — Force of Will's 1 life, Snuff Out's 4.
	//
	// A COST, not a drawback: it is validated before anything is
	// paid, so a player below N life cannot claim the offer at all.
	// (CR 118.4 lets a player pay life down to exactly zero, and the
	// state-based action kills them afterwards — that is a legal, if
	// unwise, Force of Will.)
	Life int

	// ExileFromHand is "exile a blue card from your hand" (Force of
	// Will) or "exile a white card from your hand" (Solitude's evoke
	// cost), as a spec matched against the caster's hand. The caster
	// names the card in CastSpellParams.AltCostIDs.
	//
	// The spell being cast is never a legal choice: CR 601.2a moves
	// it to the stack before costs are paid, so it is no longer in
	// hand. Force of Will cannot pitch itself.
	ExileFromHand *TargetSpec

	// ReturnToHand is "return an Island you control to its owner's
	// hand" (Daze), matched against the caster's permanents. Also
	// named in CastSpellParams.AltCostIDs.
	ReturnToHand *TargetSpec

	// PayLabel is the picker's prompt copy for the ExileFromHand /
	// ReturnToHand component ("a blue card", "an Island you
	// control"). Empty falls back to Label.
	PayLabel string

	// SacrificeOnEntry is evoke's "it's sacrificed when it enters".
	// Modelled as what CR 702.74b says it is — a triggered ability —
	// rather than as an immediate sacrifice inside the resolution.
	// The difference is observable and is the entire reason to evoke
	// a Slithermuse: the sacrifice uses the stack, so opponents get
	// a window, and the creature's own leaves-the-battlefield
	// trigger goes on the stack above nothing and draws the cards.
	SacrificeOnEntry bool
}

// Clears reports whether paying this cost deletes the spell's target
// clause. Nil-safe, so the cast path can ask without a guard.
func (a *AlternativeCost) Clears() bool {
	return a != nil && a.ClearsTargets
}

// CatalogAlternativeCosts is the catalog hook the effects package
// wires at init, mirroring CatalogAdditionalCost and CatalogModeSpec.
// Nil, or a nil return, means the card offers no alternative cost —
// which is nearly every card.
var CatalogAlternativeCosts func(oracleID string) []AlternativeCost

// AlternativeCostsFor returns the alternative costs a card offers,
// or nil.
func AlternativeCostsFor(oracleID string) []AlternativeCost {
	if CatalogAlternativeCosts == nil || oracleID == "" {
		return nil
	}
	return CatalogAlternativeCosts(oracleID)
}

// AlternativeCostByKey returns the named alternative cost for a card,
// or nil when the card offers none by that name. The empty key is
// how "I am paying the printed mana cost" is spelled on the wire, so
// it always answers nil.
func AlternativeCostByKey(oracleID, key string) *AlternativeCost {
	if key == "" {
		return nil
	}
	for _, ac := range AlternativeCostsFor(oracleID) {
		if ac.Key == key {
			out := ac
			return &out
		}
	}
	return nil
}

// validateAlternativeCost resolves the caster's claim at announce
// (CR 601.2b — the alternative cost is chosen before targets, and
// everything downstream is judged under it).
//
// A key the card doesn't offer is a rejection rather than a
// fall-back to the printed cost: silently charging full price for a
// cast the player thought was an overload is the worst of the
// available failures. A cost that deletes the target clause arriving
// with targets is rejected for the same reason — the client is
// confused about which cost it is paying.
//
// Returns (nil, nil) for the ordinary "paying the mana cost" case.
func validateAlternativeCost(oracleID, key string, targets []TargetRef) (*AlternativeCost, error) {
	if key == "" {
		return nil, nil
	}
	alt := AlternativeCostByKey(oracleID, key)
	if alt == nil {
		return nil, ErrInvalidParam
	}
	if alt.ClearsTargets && len(targets) > 0 {
		return nil, ErrInvalidParam
	}
	return alt, nil
}

// PaysCards reports whether this cost has a component the caster
// must name a card for. Nil-safe.
func (a *AlternativeCost) PaysCards() bool {
	return a != nil && (a.ExileFromHand != nil || a.ReturnToHand != nil)
}

// Available reports whether a player may claim this offer right now
// — its Condition, and nothing else. The payment components are
// checked separately, at announce, because "you control no Swamp" is
// a reason to hide the offer while "you named the wrong card" is a
// reason to reject a cast.
//
// Nil-safe and nil-Condition-safe: an offer with no condition is
// always available. Caller must hold g.mu.
func (a *AlternativeCost) Available(g *Game, controller uuid.UUID) bool {
	if a == nil {
		return false
	}
	return a.Condition == nil || a.Condition(g, controller)
}

// validateAlternativeCostPaymentLocked checks the non-mana half of a
// claimed alternative cost without paying any of it — the same
// validate-all-then-pay discipline the additional cost follows, so a
// rejected cast never leaves a half-paid cost behind.
//
// `castID` is the spell being cast, which is never a legal pitch: CR
// 601.2a has already moved it to the stack.
//
// An offer whose Condition is false is rejected here rather than
// silently downgraded to the printed cost, for the reason
// validateAlternativeCost gives about unknown keys: charging full
// price for a cast the player thought was free is the worst available
// failure.
//
// Caller must hold g.mu.
func (g *Game) validateAlternativeCostPaymentLocked(playerID, castID uuid.UUID, alt *AlternativeCost, ids []uuid.UUID) error {
	if alt == nil {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	if !alt.Available(g, playerID) {
		return ErrInvalidParam
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if alt.Life > 0 && p.Life < alt.Life {
		return ErrInvalidParam
	}
	spec := alt.ExileFromHand
	zone := ZoneHand
	if spec == nil {
		spec, zone = alt.ReturnToHand, ZoneBattlefield
	}
	if spec == nil {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	if len(ids) != 1 {
		return ErrInvalidParam
	}
	id := ids[0]
	if id == castID {
		return ErrInvalidParam
	}
	// Not a target — a cost is not targeted (CR 601.2h), so hexproof
	// and shroud do not apply and the spec is matched directly
	// against the card rather than through the targeting gate.
	c, ok := g.cardInZoneLocked(g.zoneForAltCostLocked(p, zone), id)
	if !ok {
		return ErrCardNotFound
	}
	if zone == ZoneBattlefield && c.Controller != playerID {
		return ErrInvalidParam
	}
	if spec.CardOK != nil && !spec.CardOK(g, playerID, c, zone) {
		return ErrInvalidParam
	}
	return nil
}

// zoneForAltCostLocked picks the zone an alternative cost's card
// component is paid from: the caster's own hand, or the shared
// battlefield.
func (g *Game) zoneForAltCostLocked(p *Player, kind ZoneKind) *Zone {
	if kind == ZoneHand {
		return p.Hand
	}
	return g.Battlefield
}

// payAlternativeCostLocked pays the non-mana components of a claimed
// alternative cost. Call only after
// validateAlternativeCostPaymentLocked has passed and after the spell
// itself has reached the stack (CR 601.2a before 601.2h), so anything
// watching the exile, the life payment or the bounce triggers ABOVE
// the spell and resolves first.
//
// That window is the whole reason this is not folded into the
// resolution: a Daze that returned its Island on resolution would
// hand the opponent a turn of information, and an exiled Force of
// Will pitch that never happened because the spell was countered
// would make Force of Will free.
//
// Caller must hold g.mu.
func (g *Game) payAlternativeCostLocked(playerID uuid.UUID, alt *AlternativeCost, ids []uuid.UUID) error {
	if alt == nil {
		return nil
	}
	if alt.Life > 0 {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, playerID, -alt.Life); err != nil {
			return err
		}
	}
	if len(ids) == 0 {
		return nil
	}
	id := ids[0]
	if alt.ExileFromHand != nil {
		return g.ExileCardForEffect(id)
	}
	if alt.ReturnToHand != nil {
		return g.BounceToHandForEffect(id)
	}
	return nil
}

// alternativeCostString is the "pay" half: the mana cost a cast
// actually owes. The swap is total — nothing adds the printed cost
// back, which is the difference between this file and
// additional_cost.go.
func alternativeCostString(card Card, key string) string {
	if alt := AlternativeCostByKey(CatalogKey(card), key); alt != nil {
		return alt.ManaCost
	}
	return card.ManaCost
}

// TargetSpecUnderAlternativeCost applies an alternative cost's
// rewrite of the target clause: overload deletes it, cleave swaps
// it, anything else leaves the printed clause alone. Used at
// announce (CR 601.2c), again at resolution (CR 608.2b) so both
// checks judge the spell under the text it was actually cast with,
// and by the view layer so the client's picker shows the legal set
// each offer would produce.
func TargetSpecUnderAlternativeCost(base *TargetSpec, alt *AlternativeCost) *TargetSpec {
	if alt == nil {
		return base
	}
	if alt.ClearsTargets {
		return nil
	}
	if alt.Targets != nil {
		return alt.Targets
	}
	return base
}

// queueAltCostEntryTriggerLocked puts evoke's "it's sacrificed when
// it enters" onto the pending-trigger queue as an ordinary triggered
// ability (CR 702.74b). Called from the resolution path right after
// the permanent lands and its ETB hook fires, which is the last
// moment the StackItem — and so the cost that was paid — is still in
// hand.
//
// No-op for a spell cast for its mana cost, and for an alternative
// cost that doesn't carry the clause. Caller must hold g.mu.
func (g *Game) queueAltCostEntryTriggerLocked(card Card, item *StackItem) {
	if item == nil || item.AltCost == "" {
		return
	}
	alt := AlternativeCostByKey(CatalogKey(card), item.AltCost)
	if alt == nil || !alt.SacrificeOnEntry {
		return
	}
	g.queueHarvestedTriggerLocked(&StackItem{
		Kind:         StackItemTriggered,
		Controller:   item.Controller,
		Owner:        item.Controller,
		SourceCardID: card.InstanceID,
		Label:        card.Name + " — " + alt.Label + ", sacrifice it",
		Effect: func(g *Game, it *StackItem) error {
			// The permanent may already have left — the trigger sat
			// on the stack and anyone could answer it. Nothing to
			// sacrifice is not an error; the ability simply does as
			// much as it can (CR 608.2c).
			if g.controllerOfBattlefieldCardLocked(it.SourceCardID) == uuid.Nil {
				return nil
			}
			return g.sacrificePermanentLocked(it.SourceCardID)
		},
	})
}
