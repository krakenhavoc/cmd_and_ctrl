package game

import "github.com/google/uuid"

// exile_card_additional_cost.go — #3536: "As an additional cost to cast
// this spell, exile six cards from your graveyard" (Abhorrent Oculus), per
// ADR 0189 §2.
//
// The clause is the activated-ability component, ExileCost, on the
// MANDATORY slot of a cast (AdditionalCost.ExileCards), so the candidate
// walk (ExileCostOptionsForEffect), the validator
// (validateExileCardsCostLocked) and the payer (payExileCardsCostLocked)
// are the activation's own and the two surfaces cannot disagree about who
// may pay (#544). What a cast adds is only where the picks ride and when
// they are paid:
//
//   - The caster names them on CastSpellParams.ExileIDs, the name an
//     activation already uses (`exile_ids`).
//   - They are paid with the spell already on the stack (CR 601.2a before
//     601.2h), beside the sacrifices, so a "whenever a card leaves your
//     graveyard" watcher lands above the spell, and a countered spell does
//     not hand them back.
//   - They are recorded on PaidCost.Exiled, which Context.Exiled reads at
//     resolution, and which a CR 707.10 copy keeps.
//   - The spell itself is never a candidate: the walk and the validator
//     exclude the cast card, which is what keeps Skaab Ruinator, cast from
//     its own graveyard, from paying with itself.
//
// The cost is charged under every alternative cost and every "without
// paying its mana cost" (the Oculus ruling, 2024-09-20): it lives on the
// additional-cost slot, which neither replaces.

// planExileCards is the exile-N-cards component the announced plan
// demands, or nil. effects.Register allows it only on the mandatory
// slot, so there is at most one.
func planExileCards(plan []costPayment) *ExileCost {
	for _, pay := range plan {
		if pay.cost.ExileCards != nil {
			return pay.cost.ExileCards
		}
	}
	return nil
}

// validateCastExileCardsLocked checks the cards named to pay the plan's
// exile-N-cards cost without moving anything: none when the plan has no
// such component, else the activation validator's rules (exactly N, each
// once, in the caster's own graveyard, matching the clause, none of them
// the spell). `alsoSpent` is every card another component of the same cast
// spends — one card pays one component (CR 118.3).
//
// Caller must hold g.mu.
func (g *Game) validateCastExileCardsLocked(playerID, castID uuid.UUID, plan []costPayment, ids, alsoSpent []uuid.UUID) error {
	ec := planExileCards(plan)
	_, err := g.validateExileCardsCostLocked(playerID, castID, ec, ids, alsoSpent)
	return err
}

// paidWithExiledCards folds the cards the additional cost exiles into the
// record that lands on the stack item.
func paidWithExiledCards(paid PaidCost, ids []uuid.UUID) PaidCost {
	if len(ids) == 0 {
		return paid
	}
	paid.Exiled = append(paid.Exiled, ids...)
	return paid
}

// castExileAlsoSpent is every card another component of this cast spends,
// which the exile-cards cost may not also name (CR 118.3): the hand
// discards, the alternative cost's pitched or escaped cards, and delve's
// graveyard exiles.
func castExileAlsoSpent(params CastSpellParams) []uuid.UUID {
	out := append([]uuid.UUID(nil), params.DiscardIDs...)
	out = append(out, params.AltCostIDs...)
	return append(out, params.DelveIDs...)
}
