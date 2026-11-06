package game

import "github.com/google/uuid"

// collect_evidence.go — #5193: collect evidence N, the Murders at Karlov
// Manor keyword action (CR 701.59).
//
//	701.59a To "collect evidence N" means to exile any number of cards
//	        from your graveyard with total mana value N or greater.
//
// CR 701.59b: a player unable to exile cards with total mana value N or
// greater from their graveyard can't choose to collect evidence. CR
// 701.59c: a spell's "if evidence was collected" is linked to its
// collect evidence additional cost (CR 607.2i).
//
// It is not a component of its own. It is the graveyard ExileCost every
// other "exile cards from your graveyard" cost is (Grim Lavamancer's and
// Abhorrent Oculus's component, delve's and escape's neighbour), in its
// AnyNumber form with a ManaValueAtLeast floor, so the candidate walk
// (ExileCostOptionsForEffect), the payable check (ExileCostPayableForEffect),
// the validator (validateExileCardsCostLocked), the payer
// (payExileCardsCostLocked) and the bot's pick
// (ExileCostCheapestSufficientForEffect) are the ones those costs already
// use, and ExileCost.FloorMet is the one gate all of them read. Three
// owners carry it:
//
//   - an optional additional cost to cast (CollectEvidenceKey): the cards
//     ride cast_spell's exile_ids, are exiled with the spell on the stack,
//     and are recorded on PaidCost.Exiled beside the optional-cost index on
//     PaidCost.OptionalCosts, which "if evidence was collected" reads at
//     resolution and a permanent carries as Card.Provenance.OptionalCosts;
//   - a CR 602 activated ability's cost (AbilityCost.ExileCards): the cards
//     ride activate_ability's exile_ids;
//   - an effect ("you may collect evidence N. If you do, …"): a choose_cards
//     prompt over CollectEvidenceOptionsForEffect whose set rule is
//     CollectEvidenceMet (effects.CollectEvidence).
//
// The mana value of a card in the graveyard is its printed one
// (Card.ManaValue): a modal double-faced card has only its front face
// there (CR 712.8a) and an adventurer card only its normal
// characteristics (CR 715.4).
//
// #5223: each owner announces its collection once the cards are exiled
// (noteCollectEvidenceLocked, EventCollectEvidence), which "whenever you
// collect evidence" watches. Collect evidence X as an effect (X chosen as
// it resolves, CR 107.3f) reads the same clause with its floor at X
// (CollectEvidenceXMetForEffect).

// CollectEvidenceCost is the ExileCost clause "collect evidence N" is: any
// number of the player's own graveyard cards whose total mana value is N
// or greater. The label is the printed words.
func CollectEvidenceCost(n int, label string) *ExileCost {
	return &ExileCost{
		AnyNumber:        true,
		Label:            label,
		From:             ZoneGraveyard,
		ManaValueAtLeast: n,
	}
}

// IsCollectEvidence reports whether `e` is a collect evidence clause: an
// AnyNumber graveyard clause with a total mana value floor and no other.
// Nil-safe.
func (e *ExileCost) IsCollectEvidence() bool {
	return e != nil && e.AnyNumber && e.ManaValueAtLeast > 0 && e.SymbolTotal == nil &&
		e.Match == nil && e.N == 0 && !e.CountFromX && e.Zone() == ZoneGraveyard
}

// CollectEvidenceOptionsForEffect lists the cards in `playerID`'s
// graveyard an effect's "you may collect evidence N" may exile, or nil
// when the whole graveyard's total mana value is below N (CR 701.59b: the
// player can't choose to collect evidence, so nothing is asked). `sourceID`
// is never offered; pass uuid.Nil when the source is not in the graveyard.
//
// CALLER MUST ALREADY HOLD g.mu (read or write).
func (g *Game) CollectEvidenceOptionsForEffect(playerID, sourceID uuid.UUID, n int) []uuid.UUID {
	ec := CollectEvidenceCost(n, "")
	if n <= 0 || !g.ExileCostPayableForEffect(playerID, sourceID, ec) {
		return nil
	}
	return g.ExileCostOptionsForEffect(playerID, sourceID, ec)
}

// CollectEvidenceMet reports whether `cards` collect evidence N: their
// total mana value is N or greater (CR 701.59a). An effect's prompt
// registers it as its set rule, so a restored prompt refuses a pick that
// falls short.
func CollectEvidenceMet(cards []Card, n int) bool {
	return CollectEvidenceCost(n, "").FloorMet(cards)
}

// CollectEvidenceXOptionsForEffect lists the cards in `playerID`'s
// graveyard a "you may collect evidence X" may exile (Incinerator of the
// Guilty, #5223): every one of them. X is the controller's to choose as the
// effect resolves (CR 107.3f, 608.2d), the same collect evidence clause with
// its floor at X (CollectEvidenceCost), so no card is out of reach and X = 0
// is always possible, with no card exiled.
//
// CALLER MUST ALREADY HOLD g.mu (read or write).
func (g *Game) CollectEvidenceXOptionsForEffect(playerID uuid.UUID) []uuid.UUID {
	return g.ExileCostOptionsForEffect(playerID, uuid.Nil, CollectEvidenceCost(0, ""))
}

// CollectEvidenceXCeilingForEffect is the largest X a collect evidence X
// that exiles `ids` may name: their total mana value (CR 701.59a), the
// number FloorMet compares X with (CollectEvidenceXMetForEffect). ok is
// false when one of them is no longer in `playerID`'s graveyard.
//
// CALLER MUST ALREADY HOLD g.mu (read or write).
func (g *Game) CollectEvidenceXCeilingForEffect(playerID uuid.UUID, ids []uuid.UUID) (ceiling int, ok bool) {
	cards, ok := g.collectEvidenceCardsLocked(playerID, ids)
	return TotalManaValue(cards), ok
}

// CollectEvidenceXMetForEffect reports whether exiling `ids` from
// `playerID`'s graveyard collects evidence X: every one is still there,
// and they pass the one gate every collect evidence reads (FloorMet) with
// the floor at X. X = 0 is met by any cards, none included (CR 701.59a);
// an X above their total mana value is refused, and so is a negative X
// (CR 107.1b).
//
// CALLER MUST ALREADY HOLD g.mu (read or write).
func (g *Game) CollectEvidenceXMetForEffect(playerID uuid.UUID, ids []uuid.UUID, x int) bool {
	cards, ok := g.collectEvidenceCardsLocked(playerID, ids)
	return ok && x >= 0 && CollectEvidenceCost(x, "").FloorMet(cards)
}

// collectEvidenceCardsLocked resolves `ids` out of `playerID`'s graveyard,
// with ok false when one of them is not there (or is named twice).
func (g *Game) collectEvidenceCardsLocked(playerID uuid.UUID, ids []uuid.UUID) ([]Card, bool) {
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return nil, false
		}
		seen[id] = true
	}
	cards := g.exileCostCardsLocked(playerID, CollectEvidenceCost(0, ""), ids)
	return cards, len(cards) == len(ids)
}

// noteCollectEvidenceLocked records that `player` collected evidence for
// `source`, exiling `cards`: one EventCollectEvidence, which "whenever you
// collect evidence" (Evidence Examiner, Surveillance Monitor) watches,
// carrying the cards and their total mana value. Called by each owner once
// its collection has been made: the cast path after the spell's exiles, the
// activation after the ability's, the effect form after its own
// (NoteCollectEvidenceForEffect). Caller must hold g.mu in write mode.
func (g *Game) noteCollectEvidenceLocked(player, source uuid.UUID, cards []uuid.UUID) {
	total := 0
	for _, id := range cards {
		if z, pos := g.locateCardLocked(id); z != nil {
			total += z.Cards[pos].ManaValue()
		}
	}
	g.EmitEvent(Event{
		Kind:    EventCollectEvidence,
		Actor:   player,
		Source:  source,
		Amount:  total,
		CardIDs: append([]uuid.UUID(nil), cards...),
	})
}

// NoteCollectEvidenceForEffect is noteCollectEvidenceLocked for the effect
// form (effects.RegisterCollectEvidence, effects.RegisterCollectEvidenceX):
// call it once the chosen cards have been exiled, and before the rest of
// the card runs.
//
// Caller must hold g.mu in write mode.
func (g *Game) NoteCollectEvidenceForEffect(player, source uuid.UUID, cards []uuid.UUID) {
	g.noteCollectEvidenceLocked(player, source, cards)
}

// planCollectsEvidence reports whether a cast's payment plan pays a collect
// evidence clause (an announced OptionalCollectEvidence), so the cast path
// emits EventCollectEvidence for the cards it exiled.
func planCollectsEvidence(plan []costPayment) bool {
	for _, pay := range plan {
		if pay.cost.ExileCards.IsCollectEvidence() {
			return true
		}
	}
	return false
}

// CardEvidenceCollected is "if evidence was collected" read from a
// permanent that has entered (Crimestopper Sprite's and Vitu-Ghazi
// Inspector's enters triggers): the spell it resolved from announced its
// optional collect evidence cost (CR 607.2i), the record the resolution
// carried onto it (Card.Provenance.OptionalCosts, CR 400.7d). False for a
// permanent that was not cast, and for one that left and came back.
func CardEvidenceCollected(c Card) bool {
	return OptionalCostTimesPaid(c, c.Provenance.OptionalCosts, CollectEvidenceKey) > 0
}
