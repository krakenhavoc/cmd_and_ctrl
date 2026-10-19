package game

import (
	"fmt"

	"github.com/google/uuid"
)

// splice_cost.go — #4055, ADR 0202's 2026-10-19 amendment: a splice cost
// that is not mana.
//
// CR 702.47a: "Paying a card's splice cost follows the rules for paying
// additional costs in rules 601.2b and 601.2f-h." The five Kamigawa splice
// cards whose cost is not mana print it as a verb:
//
//	Veil of Secrecy        Return a blue creature you control to its owner's hand.
//	Horobi's Whisper       Exile four cards from your graveyard.
//	Roar of Jukai          An opponent gains 5 life.
//	Hundred-Talon Strike   Tap an untapped white creature you control.
//	Torrent of Stone       Sacrifice two Mountains.
//
// # The shape
//
// SpliceAbility.Cost is the non-mana half, built out of the components
// the activated abilities and the other additional costs already pay
// (a sacrifice clause, an exile-from-graveyard cost, a tap-others cost, a
// return-to-hand cost) plus the one new thing, a chosen opponent gaining
// life. Its walks, validators and payers are theirs, so the view, the bot
// enumerator and the engine cannot disagree about which permanent pays
// (#544).
//
// The payment rides CastSpellParams.SplicePayments, one entry for each
// card in Splices, in that order. It is NOT on the flat discard / sacrifice
// / teamwork lists the cast's own plan walks: a splice card is another
// card's cost, which of its payers belongs to which card is the caster's
// announcement (CR 601.2b), and a flat list would make the client rebuild
// the plan's order to say it. Because the lists are separate, an object
// named by two payments, or by a payment and by the cast's own costs, is
// refused (CR 118.3): fewer choices than the rules allow, never more.
//
// # When it is paid
//
// With the spell already on the stack (CR 601.2a before 601.2h), after the
// cast's own costs, in the order the cards were revealed: a sacrifice
// trigger or a leaves-the-battlefield watcher lands above the spell. A
// spliced card that cannot pay its cost cannot be spliced (CR 118.3), so
// SpliceOffersLocked leaves it out and the announcement refuses it.
//
// Nothing about the payment is recorded on the spell's PaidCost: no
// printed splice text reads what its own cost paid, and the spell's own
// "sacrificed" or "tapped to pay" readers must not see another card's.

// SpliceCost is the non-mana half of a splice cost. Exactly one component
// is set on every printed card; effects.Register refuses none, and refuses
// the components no splice cost uses.
type SpliceCost struct {
	// Label is the cost as printed, without its closing period: "Sacrifice
	// two Mountains". SpliceAbility.Label puts it after the dash.
	Label string

	// Sacrifice is "Sacrifice two Mountains": a fixed-count clause over the
	// caster's permanents (effects.SpliceSacrifice). The picks ride
	// SplicePayment.SacrificeIDs, validated as an additional cost's are.
	Sacrifice *TargetSpec

	// ExileCards is "Exile four cards from your graveyard": a fixed count
	// from the caster's graveyard (effects.SpliceExileFromGraveyard). The
	// picks ride SplicePayment.ExileIDs.
	ExileCards *ExileCost

	// TapOthers is "Tap an untapped white creature you control": the
	// activated abilities' clause, a fixed count, none of them the {T}
	// symbol so summoning sickness is not asked (CR 302.6). The picks ride
	// SplicePayment.TapIDs.
	TapOthers *TapOthersCost

	// ReturnToHand is "Return a blue creature you control to its owner's
	// hand": the activated abilities' clause. The picks ride
	// SplicePayment.ReturnIDs.
	ReturnToHand *ReturnToHandCost

	// OpponentGainsLife is "An opponent gains 5 life": the amount. The
	// caster chooses the opponent the way gift's recipient is chosen, on
	// SplicePayment.Opponent; paying it is that opponent's life change, one
	// the CR 614 window sees and an opponent who can't gain life can't be
	// given (CR 614.17b).
	OpponentGainsLife int
}

// Empty reports whether the cost demands nothing. Nil-safe.
func (c *SpliceCost) Empty() bool {
	return c == nil || (c.Sacrifice == nil && c.ExileCards == nil && c.TapOthers.Empty() &&
		c.ReturnToHand.Empty() && c.OpponentGainsLife <= 0)
}

// SplicePayment is what one revealed splice card's cost is paid with
// (CR 702.47a, 601.2h). The zero value pays a cost that demands nothing.
type SplicePayment struct {
	SacrificeIDs []uuid.UUID
	ExileIDs     []uuid.UUID
	TapIDs       []uuid.UUID
	ReturnIDs    []uuid.UUID
	// Opponent is the player given the life of an "An opponent gains N
	// life" cost.
	Opponent uuid.UUID
}

// objects is every object the payment spends.
func (p SplicePayment) objects() []uuid.UUID {
	out := make([]uuid.UUID, 0, len(p.SacrificeIDs)+len(p.ExileIDs)+len(p.TapIDs)+len(p.ReturnIDs))
	out = append(out, p.SacrificeIDs...)
	out = append(out, p.ExileIDs...)
	out = append(out, p.TapIDs...)
	return append(out, p.ReturnIDs...)
}

// spliceObjectsOf is every object the payments of a cast spend.
func spliceObjectsOf(pays []SplicePayment) []uuid.UUID {
	var out []uuid.UUID
	for _, p := range pays {
		out = append(out, p.objects()...)
	}
	return out
}

// castSpentObjects is every object the cast's own costs spend or tap, which
// no splice payment may name again (CR 118.3).
func castSpentObjects(params CastSpellParams) []uuid.UUID {
	return unionIDList(params.DiscardIDs, params.SacrificeIDs, params.AltCostIDs, params.DelveIDs,
		params.ExilePermanentIDs, params.ExileFromHandIDs, params.ExileIDs, params.TeamworkIDs,
		params.TapIDs, params.BlightIDs)
}

// unionIDList is the ids of several lists in one slice, in order.
func unionIDList(lists ...[]uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, l := range lists {
		out = append(out, l...)
	}
	return out
}

// spliceReturnIDsOf is every permanent the payments return to a hand: the
// ones a commander's owner is asked about first (CR 903.9b).
func spliceReturnIDsOf(pays []SplicePayment) []uuid.UUID {
	var out []uuid.UUID
	for _, p := range pays {
		out = append(out, p.ReturnIDs...)
	}
	return out
}

// spliceCostOf is the non-mana cost the revealed card declares, nil for a
// mana-only splice.
func spliceCostOf(sp SplicedText) *SpliceCost {
	if sa := SpliceAbilityFor(sp.Key); sa != nil {
		return sa.Cost
	}
	return nil
}

// spliceCostPayableLocked reports whether `playerID` could pay `c` right
// now for the spell `castID` (CR 118.3: an unpayable cost can't be paid, so
// a card whose cost the caster lacks is not offered). Mana is not asked,
// for the reason the alternative cost gives: CR 601.2g lets the caster make
// it after the costs are settled. Each component is judged alone; two
// splice cards that want the same permanent are refused at the announcement.
//
// Caller must hold g.mu (read or write).
func (g *Game) spliceCostPayableLocked(playerID, castID uuid.UUID, c *SpliceCost) bool {
	if c.Empty() {
		return true
	}
	if c.Sacrifice != nil &&
		len(g.SacrificeCostCandidatesForEffect(playerID, uuid.Nil, c.Sacrifice)) < SacrificeCostCount(c.Sacrifice) {
		return false
	}
	if ec := c.ExileCards; ec != nil && len(g.ExileCostOptionsForEffect(playerID, castID, ec)) < ec.N {
		return false
	}
	if !g.TapOthersPayable(playerID, castID, c.TapOthers) {
		return false
	}
	if !g.ReturnToHandPayable(playerID, castID, c.ReturnToHand) {
		return false
	}
	if c.OpponentGainsLife > 0 && len(g.SpliceOpponentsLocked(playerID)) == 0 {
		return false
	}
	return true
}

// SpliceOpponentsLocked is who may be given the life of an "An opponent
// gains N life" splice cost: the opponents still in the game who can gain
// life (CR 614.17b: a cost that includes an event that can't happen can't
// be paid). The set the view offers and the announcement checks.
//
// Caller must hold g.mu (read or write).
func (g *Game) SpliceOpponentsLocked(caster uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range g.GiftOpponentsLocked(caster) {
		if p := g.playerByIDLocked(id); p != nil && !g.playerCantGainLifeLocked(p) {
			out = append(out, id)
		}
	}
	return out
}

// validateSplicePaymentsLocked checks, before anything is paid (CR 601.2b,
// 601.2h; validate-all-then-pay), that every revealed splice card's cost is
// paid with exactly what it demands. `pays` is aligned with `spliced`; a
// missing entry pays nothing. `spent` is everything the cast's own costs
// already spend, which no splice payment may name again (CR 118.3).
//
// A payment for a component the card's cost does not have is a client bug
// and is refused rather than ignored, the posture the cast's other flat
// lists take.
//
// Caller must hold g.mu.
func (g *Game) validateSplicePaymentsLocked(playerID, castID uuid.UUID, spliced []SplicedText, pays []SplicePayment, spent []uuid.UUID) error {
	if len(pays) > len(spliced) {
		return fmt.Errorf("%w: a splice payment for a card nothing was spliced for", ErrInvalidParam)
	}
	used := make(map[uuid.UUID]bool, len(spent))
	for _, id := range spent {
		used[id] = true
	}
	for k, sp := range spliced {
		var pay SplicePayment
		if k < len(pays) {
			pay = pays[k]
		}
		cost := spliceCostOf(sp)
		if cost.Empty() {
			if len(pay.objects()) > 0 || pay.Opponent != uuid.Nil {
				return fmt.Errorf("%w: %s's splice cost takes no payment", ErrInvalidParam, sp.Name)
			}
			continue
		}
		for _, id := range pay.objects() {
			if used[id] {
				return fmt.Errorf("%w: an object pays one cost of a cast at most", ErrInvalidParam)
			}
			used[id] = true
		}
		if err := g.validateSplicePaymentLocked(playerID, castID, sp, cost, pay, spent); err != nil {
			return err
		}
	}
	return nil
}

// validateSplicePaymentLocked is one card's payment against its cost.
func (g *Game) validateSplicePaymentLocked(playerID, castID uuid.UUID, sp SplicedText, cost *SpliceCost, pay SplicePayment, spent []uuid.UUID) error {
	bad := func(what string) error {
		return fmt.Errorf("%w: %s's splice cost: %s", ErrInvalidParam, sp.Name, what)
	}
	if cost.Sacrifice == nil && len(pay.SacrificeIDs) > 0 {
		return bad("nothing to sacrifice")
	}
	if cost.ExileCards == nil && len(pay.ExileIDs) > 0 {
		return bad("nothing to exile")
	}
	if cost.TapOthers.Empty() && len(pay.TapIDs) > 0 {
		return bad("nothing to tap")
	}
	if cost.ReturnToHand.Empty() && len(pay.ReturnIDs) > 0 {
		return bad("nothing to return")
	}
	if cost.OpponentGainsLife <= 0 && pay.Opponent != uuid.Nil {
		return bad("no opponent to choose")
	}
	if cost.Sacrifice != nil {
		if _, err := g.validateSacrificeCostLocked(playerID, castID, AbilityCost{SacrificeOther: cost.Sacrifice}, pay.SacrificeIDs, 0); err != nil {
			return err
		}
	}
	if cost.ExileCards != nil {
		// The cast's own spends and the other splice cards' are in `spent`
		// or in the payments checked before this one; the engine's walk
		// then judges the cards themselves.
		if _, err := g.validateExileCardsCostLocked(playerID, castID, cost.ExileCards, pay.ExileIDs, spent); err != nil {
			return err
		}
	}
	if !cost.TapOthers.Empty() {
		if err := g.validateTapOthersCostLocked(playerID, castID, cost.TapOthers, pay.TapIDs); err != nil {
			return err
		}
	}
	if !cost.ReturnToHand.Empty() {
		if err := g.validateReturnToHandCostLocked(playerID, castID, cost.ReturnToHand, pay.ReturnIDs); err != nil {
			return err
		}
	}
	if cost.OpponentGainsLife > 0 {
		ok := false
		for _, id := range g.SpliceOpponentsLocked(playerID) {
			if id == pay.Opponent {
				ok = true
			}
		}
		if !ok {
			return bad("choose an opponent who can gain life")
		}
	}
	return nil
}

// paySplicePaymentsLocked pays every revealed card's cost, in the order
// the cards were revealed (CR 601.2h), with the spell already on the stack.
// Call only after validateSplicePaymentsLocked has passed.
//
// Caller must hold g.mu.
func (g *Game) paySplicePaymentsLocked(playerID, castID uuid.UUID, spliced []SplicedText, pays []SplicePayment, answers map[uuid.UUID]bool) error {
	for k, sp := range spliced {
		cost := spliceCostOf(sp)
		if cost.Empty() || k >= len(pays) {
			continue
		}
		pay := pays[k]
		if cost.Sacrifice != nil {
			if err := g.payCostSacrificesLocked(pay.SacrificeIDs, answers); err != nil {
				return err
			}
		}
		if cost.ExileCards != nil && len(pay.ExileIDs) > 0 {
			if err := g.payExileCardsCostLocked(playerID, castID, pay.ExileIDs, answers); err != nil {
				return err
			}
		}
		if !cost.TapOthers.Empty() {
			g.RecomputeLayersIfStaleLocked()
			g.payTapOthersCostLocked(playerID, pay.TapIDs)
		}
		if !cost.ReturnToHand.Empty() {
			if _, err := g.payReturnToHandCostLocked(playerID, castID, pay.ReturnIDs, answers); err != nil {
				return err
			}
		}
		if cost.OpponentGainsLife > 0 {
			if err := g.gainLifeAsCostLocked(castID, pay.Opponent, cost.OpponentGainsLife); err != nil {
				return err
			}
		}
	}
	return nil
}
