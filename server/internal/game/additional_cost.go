package game

import "github.com/google/uuid"

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
	// names one in CastSpellParams.SacrificeIDs.
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
}

// Empty reports whether the cost demands nothing. Nil-safe.
func (c *AdditionalCost) Empty() bool {
	return c == nil || (c.DiscardCards == 0 && c.Sacrifice == nil && !c.PayLifeX)
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

// validateAdditionalCostLocked checks that the caster named exactly
// the right cards to pay `cost`, without paying anything — the same
// validate-all-then-pay discipline ActivateCatalogAbility uses, so a
// rejected cast never leaves a half-paid cost behind. `castID` is
// the spell being cast, which is never a legal discard.
//
// A card with no additional cost that arrives WITH discard IDs is a
// client bug, not a no-op: rejecting it keeps the wire honest.
//
// Caller must hold g.mu.
func (g *Game) validateAdditionalCostLocked(playerID, castID uuid.UUID, cost *AdditionalCost, discardIDs, sacrificeIDs []uuid.UUID, xValue int) error {
	if cost.Empty() {
		if len(discardIDs) > 0 || len(sacrificeIDs) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	// CR 119.4: a player may pay N life only with a life total of at
	// least N. Checked at announce with the rest of the choices, so
	// an unpayable X is a rejected cast rather than a player at -3.
	if cost.PayLifeX && xValue > p.Life {
		return ErrInvalidParam
	}
	if len(discardIDs) != cost.DiscardCards {
		return ErrInvalidParam
	}
	seen := make(map[uuid.UUID]bool, len(discardIDs))
	for _, id := range discardIDs {
		if id == castID || seen[id] {
			return ErrInvalidParam
		}
		seen[id] = true
		if !p.Hand.Contains(id) {
			return ErrCardNotFound
		}
	}
	// The sacrifice clause reuses the activated-ability validator, so
	// "you may only sacrifice what you control" (CR 701.21a) and the
	// spec's own predicate are enforced in one place rather than two.
	if cost.Sacrifice == nil {
		if len(sacrificeIDs) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	if _, err := g.validateSacrificeCostLocked(playerID, castID, AbilityCost{
		SacrificeOther: cost.Sacrifice,
	}, sacrificeIDs); err != nil {
		return err
	}
	return nil
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
func (g *Game) payAdditionalCostLocked(playerID uuid.UUID, discardIDs, sacrificeIDs []uuid.UUID, payLife int) error {
	// Life first: it is the component with no choice attached, and
	// paying it before the sacrifices keeps the event order matching
	// the way the clauses are read aloud.
	if payLife > 0 {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, playerID, -payLife); err != nil {
			return err
		}
	}
	for _, id := range sacrificeIDs {
		if err := g.sacrificePermanentLocked(id); err != nil {
			return err
		}
	}
	if len(discardIDs) == 0 {
		return nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	for _, id := range discardIDs {
		if _, err := MoveCard(p.Hand, p.Graveyard, id); err != nil {
			return err
		}
		g.markCardKnownInZoneLocked(p.Graveyard, id)
		g.EmitEvent(Event{
			Kind:    EventDiscardCard,
			Actor:   playerID,
			CardID:  id,
			OldZone: ZoneHand,
			NewZone: ZoneGraveyard,
		})
	}
	return nil
}
