package game

import "github.com/google/uuid"

// library_cost.go — ADR 0109 §7 (#1902): the three activation-cost
// components that move cards to or from a library, or by chance.
//
//	AbilityCost.PutFromHandOnLibraryTop  "Put a card from your hand on
//	                                     top of your library" (Penance,
//	                                     Leashling, Hidden Retreat)
//	AbilityCost.ExileFromLibraryTop      "Exile the top N cards of your
//	                                     library" (Seasoned Tactician,
//	                                     Arc-Slogger, Phyrexian Devourer)
//	DiscardCost.Random                   "Discard a card at random"
//	                                     (owner decision 3: Pyromancy,
//	                                     Mage il-Vec, Meteor Storm)
//
// All three are validated with the other components and paid in
// activateCatalogAbilityLocked, so a refused activation pays nothing.
// CR 601.2h (via CR 602.2b) splits the payment in two: costs with a
// random element, and costs that move objects from the library to a
// public zone, are paid after every other cost. So the hand-to-library
// put is paid with the discards and the exiles, and the random discard
// and the library exile are paid last of all.
//
// Two of them choose nothing, so they ride no params field and no wire
// field: the library exile takes whatever is on top, and the random
// discard is the engine's draw. The hand-to-library put names its card
// at announce in ActivateAbilityParams.TopIDs (`top_ids`), as the
// discard and exile picks do (CR 602.2b).
//
// What each payment moved is recorded for CR 400.7j: the library exile
// on PaidCost.Exiled (Phyrexian Devourer's "the exiled card's mana
// value", Storm Elemental's "if the exiled card is a snow land"), and
// the random discard on PaidCost.Discarded beside any other discard
// (Pyromancy's "the mana value of the discarded card", ADR 0109 §8).
// The card put on top of the library is in a hidden zone and nothing
// reads it, so it is not recorded.

// PutOnTopCostOptionsForEffect lists the cards in `playerID`'s hand that
// could pay a PutFromHandOnLibraryTop component right now, in hand
// order, excluding the source when it is itself in that hand. Every
// printed clause says "a card", so there is no predicate.
//
// The view stamps it so the client can skip its picker when the hand
// holds exactly the cards the clause needs, and the legal enumerator
// pays out of it.
//
// CALLER MUST ALREADY HOLD g.mu (read or write).
func (g *Game) PutOnTopCostOptionsForEffect(playerID, sourceID uuid.UUID, n int) []uuid.UUID {
	if n <= 0 {
		return nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Hand == nil {
		return nil
	}
	var out []uuid.UUID
	for i := range p.Hand.Cards {
		if id := p.Hand.Cards[i].InstanceID; id != sourceID {
			out = append(out, id)
		}
	}
	return out
}

// LibraryExileCostPayableForEffect reports whether `playerID`'s library
// holds the N cards an ExileFromLibraryTop component exiles (CR 118.3:
// "Exile the top four cards of your library" can't be paid from a
// library of three). True for n <= 0, a cost with no such component.
//
// CALLER MUST ALREADY HOLD g.mu (read or write).
func (g *Game) LibraryExileCostPayableForEffect(playerID uuid.UUID, n int) bool {
	if n <= 0 {
		return true
	}
	p := g.playerByIDLocked(playerID)
	return p != nil && p.Library != nil && p.Library.Size() >= n
}

// RandomDiscardPoolForEffect is the hand a "Discard N cards at random"
// component draws from: every card in `playerID`'s hand except the
// source (an ability activated from a hand cannot pay for itself), the
// cards the same payment names to another component (`taken`: a card
// already discarded, exiled or put on top of the library is not in the
// hand when the random discard is paid, CR 601.2h), and a card whose
// exit an effect has already paused (#1445: it is already on its way
// out, and the payment would be refused if it were drawn).
//
// CR 118.3: the component can be paid only when the pool holds N
// cards. The enumerator gates on the pool's size and the activation
// draws from it.
//
// CALLER MUST ALREADY HOLD g.mu (read or write).
func (g *Game) RandomDiscardPoolForEffect(playerID, sourceID uuid.UUID, taken []uuid.UUID) []uuid.UUID {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Hand == nil {
		return nil
	}
	skip := make(map[uuid.UUID]bool, len(taken)+1)
	skip[sourceID] = true
	for _, id := range taken {
		skip[id] = true
	}
	var out []uuid.UUID
	for i := range p.Hand.Cards {
		id := p.Hand.Cards[i].InstanceID
		if skip[id] || g.zoneChangePausedLocked(id) {
			continue
		}
		out = append(out, id)
	}
	return out
}

// RandomDiscardCount is how many cards a cost discards at random: the
// clause's N when its DiscardCards component is random, zero otherwise.
func (c AbilityCost) RandomDiscardCount() int {
	if dc := c.DiscardCards; dc != nil && dc.Random && dc.N > 0 {
		return dc.N
	}
	return 0
}

// validatePutOnTopCostLocked checks a PutFromHandOnLibraryTop component
// without moving anything: exactly n ids, each distinct, each in the
// activator's hand, never the source, and never a card the same payment
// already names to a discard or an exile (`alsoSpent` — one card pays
// one component, CR 118.3). Ids arriving for a cost with no such
// component are refused rather than ignored, as a stray discard_ids is.
//
// Caller must hold g.mu.
func (g *Game) validatePutOnTopCostLocked(playerID, sourceID uuid.UUID, n int, chosen, alsoSpent []uuid.UUID) ([]uuid.UUID, error) {
	if n <= 0 {
		if len(chosen) > 0 {
			return nil, ErrInvalidParam
		}
		return nil, nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return nil, ErrPlayerNotFound
	}
	if len(chosen) != n {
		return nil, ErrInvalidParam
	}
	spent := make(map[uuid.UUID]bool, len(alsoSpent))
	for _, id := range alsoSpent {
		spent[id] = true
	}
	seen := make(map[uuid.UUID]bool, len(chosen))
	for _, id := range chosen {
		if seen[id] || spent[id] || id == sourceID {
			return nil, ErrInvalidParam
		}
		seen[id] = true
		if !p.Hand.Contains(id) {
			return nil, ErrCardNotFound
		}
	}
	return append([]uuid.UUID(nil), chosen...), nil
}

// libraryExileCostCardsLocked is the cards an ExileFromLibraryTop
// component will exile: the top n of the activator's library, top
// first. Read at announce so the commander gate can ask about them;
// nothing between the announce and the payment touches the library
// (the hand-to-library put is refused beside this component at boot).
// ErrCantPayLibraryCost when the library is short (CR 118.3).
//
// Caller must hold g.mu.
func (g *Game) libraryExileCostCardsLocked(playerID uuid.UUID, n int) ([]uuid.UUID, error) {
	if n <= 0 {
		return nil, nil
	}
	if !g.LibraryExileCostPayableForEffect(playerID, n) {
		return nil, ErrCantPayLibraryCost
	}
	lib := g.playerByIDLocked(playerID).Library
	out := make([]uuid.UUID, 0, n)
	for i := len(lib.Cards) - 1; i >= 0 && len(out) < n; i-- {
		out = append(out, lib.Cards[i].InstanceID)
	}
	return out, nil
}

// randomDiscardsLocked draws the cards a "Discard N cards at random"
// component discards (CR 701.9b: the player does not choose).
//
// The draw is made ONCE per announcement, after every other part of it
// has been validated, and the result rides the announcement
// (ActivateAbilityParams.randomDiscards). That is what lets a commander
// among the drawn cards be offered CR 903.9a before anything is paid
// (#1397, cost_commander_choice.go): the parked announcement re-runs
// with the same draw rather than drawing again. The pool excludes every
// card another component of the payment takes, so the draw is the one
// CR 601.2h's "paid after the others" would make.
//
// A re-run's preset draw is validated again: each card still in the
// pool, each once. ErrCantPayRandomDiscard when the pool is short
// (CR 118.3).
//
// Caller must hold g.mu for writing (it draws from the game's RNG).
func (g *Game) randomDiscardsLocked(playerID, sourceID uuid.UUID, n int, preset, taken []uuid.UUID) ([]uuid.UUID, error) {
	if n <= 0 {
		if len(preset) > 0 {
			return nil, ErrInvalidParam
		}
		return nil, nil
	}
	pool := g.RandomDiscardPoolForEffect(playerID, sourceID, taken)
	if len(pool) < n {
		return nil, ErrCantPayRandomDiscard
	}
	if preset != nil {
		in := make(map[uuid.UUID]bool, len(pool))
		for _, id := range pool {
			in[id] = true
		}
		if len(preset) != n {
			return nil, ErrInvalidParam
		}
		seen := make(map[uuid.UUID]bool, n)
		for _, id := range preset {
			if seen[id] || !in[id] {
				return nil, ErrInvalidParam
			}
			seen[id] = true
		}
		return append([]uuid.UUID(nil), preset...), nil
	}
	return g.ChooseAtRandomForEffect(RandomDraw{Player: playerID, Source: sourceID}, pool, n), nil
}

// payPutOnTopCostLocked pays a PutFromHandOnLibraryTop component: each
// named card goes from the activator's hand to the top of their
// library, in the order named (the last one named ends on top), through
// the one exit primitive with MustSettleNow. It is not a discard: no
// EventDiscardCard, nothing for madness to see.
//
// The activator still knows the card they put there (ADR 0088 Decision
// 3's reading: a card placed alone in its lane keeps its knowers), so
// the knower set the library route clears is put back: whoever knew it
// in hand, and the activator.
// A commander whose owner took CR 903.9b's command zone was never put
// in the library and is left alone.
//
// Caller must hold g.mu and have validated the list.
func (g *Game) payPutOnTopCostLocked(playerID, sourceID uuid.UUID, ids []uuid.UUID, answers map[uuid.UUID]bool) error {
	for _, id := range ids {
		card, _ := g.cardInZoneLocked(g.findCardZoneLocked(id), id)
		knowers := libraryOrderKnowers(card, nil, playerID, 1)
		if _, err := g.routeCardToZoneLocked(zoneRoute{
			CardID:          id,
			Dst:             ZoneLibrary,
			Actor:           playerID,
			Source:          sourceID,
			Cause:           MoveCause{Kind: MoveCauseCost, Controller: playerID},
			MustSettleNow:   true,
			commanderAnswer: commanderAnswerFor(answers, id),
		}); err != nil {
			return err
		}
		if z := g.findCardZoneLocked(id); z != nil && z.Kind == ZoneLibrary {
			if c := g.findCardByIDLocked(id); c != nil {
				c.KnownBy = knowers
			}
		}
	}
	return nil
}

// payLibraryExileCostLocked pays an ExileFromLibraryTop component: the
// cards libraryExileCostCardsLocked read at announce, top first, each
// into exile face up through the one exit primitive with
// MustSettleNow. Exile is public, so the table sees them.
//
// Caller must hold g.mu and have validated the list.
func (g *Game) payLibraryExileCostLocked(playerID, sourceID uuid.UUID, ids []uuid.UUID, answers map[uuid.UUID]bool) error {
	return g.payExileCardsCostLocked(playerID, sourceID, ids, answers)
}
