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

// CardEvidenceCollected is "if evidence was collected" read from a
// permanent that has entered (Crimestopper Sprite's and Vitu-Ghazi
// Inspector's enters triggers): the spell it resolved from announced its
// optional collect evidence cost (CR 607.2i), the record the resolution
// carried onto it (Card.Provenance.OptionalCosts, CR 400.7d). False for a
// permanent that was not cast, and for one that left and came back.
func CardEvidenceCollected(c Card) bool {
	return OptionalCostTimesPaid(c, c.Provenance.OptionalCosts, CollectEvidenceKey) > 0
}
