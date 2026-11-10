package game

import "github.com/google/uuid"

// chosen_cost.go — "As an additional cost to cast this spell, choose a
// creature you control or reveal a creature card from your hand. This
// spell deals damage equal to the power of the creature you chose or the
// card you revealed" (Monstrous Emergence, Duskmourn, #4452): the payment
// record of a behold whose pick the spell's text reads back.
//
// The cost is the behold component the Champions and the either/or
// branches already use (RevealCost{Behold}) with CardType "creature" (no
// creature type) and Chosen set, on the mandatory slot. The candidate
// walk, the validator and the payer are the reveal cost's own
// (RevealCostOptionsForEffect, validateRevealLocked, payRevealLocked); what
// Chosen adds is one entry on PaidCost.Chosen, the PaidTap record a station
// ability's tapped creature already uses (#759):
//
//   - a permanent on the battlefield is recorded as the object it is
//     ({instance, epoch}) with its power as it was chosen. The spell reads
//     its power as it resolves, so a pump in response counts; once it has
//     left the battlefield the exit choke point rewrites the entry to the
//     power it last had there and marks it Left (freezePaidTapsOnExitLocked,
//     CR 608.2h), and a new object under the same instance ID is not the
//     creature that was chosen (CR 400.7). That is the printed ruling: "use
//     that creature's power as it last existed on the battlefield".
//   - a card in the hand is revealed, not chosen from the battlefield, and
//     its power is the printed one: it is recorded Left with that power, so
//     it is final from the moment of the reveal.
//
// A CR 707.10 copy of the spell carries the record (copiedPaidCost), since
// its text refers to the objects chosen as the original was cast.

// recordChosenLocked writes the pick onto the cast's stack item. Called
// from payRevealLocked, with the spell already on the stack.
//
// Caller must hold g.mu in write mode.
func (g *Game) recordChosenLocked(playerID, castID, id uuid.UUID) {
	item := g.StackMeta[castID]
	if item == nil {
		return
	}
	if p := g.playerByIDLocked(playerID); p != nil && p.Hand != nil {
		for i := range p.Hand.Cards {
			if c := &p.Hand.Cards[i]; c.InstanceID == id {
				item.Paid.Chosen = append(item.Paid.Chosen, PaidTap{ID: id, Epoch: c.ObjectEpoch, Power: c.PowerForComparison(), Left: true})
				return
			}
		}
	}
	g.RecomputeLayersIfStaleLocked()
	if c := findBattlefieldCard(g, id); c != nil {
		item.Paid.Chosen = append(item.Paid.Chosen, PaidTap{ID: id, Epoch: c.ObjectEpoch, Power: c.PowerForComparison()})
	}
}

// ChosenPowerForEffect is the power of the first object a Chosen behold
// recorded on `item` (the spell as it resolves), and whether one was. A
// permanent still on the battlefield as the same object is read live
// (CR 608.2h); one that has left, and a card revealed from the hand, answer
// with the recorded power. Not clamped, like PaidTapPowerForEffect.
//
// Caller must hold g.mu in write mode.
func (g *Game) ChosenPowerForEffect(item *StackItem) (int, bool) {
	if item == nil || len(item.Paid.Chosen) == 0 {
		return 0, false
	}
	return g.PaidTapPowerForEffect(item.Paid.Chosen[0]), true
}
