package game

import "github.com/google/uuid"

// CanEffectCauseSacrificeForEffect applies the effect's protection gate.
// Costs do not use this query. Caller must hold g.mu.
func (g *Game) CanEffectCauseSacrificeForEffect(playerID uuid.UUID) bool {
	p := g.playerByIDLocked(playerID)
	return p != nil && !p.Eliminated && !g.effectSacrificeBlockedLocked(playerID)
}

// EffectSacrificeCandidatesForEffect returns controlled permanents eligible
// for this effect's instruction. It does not target. Caller must hold g.mu.
func (g *Game) EffectSacrificeCandidatesForEffect(playerID uuid.UUID, spec *TargetSpec) []uuid.UUID {
	if !g.CanEffectCauseSacrificeForEffect(playerID) {
		return nil
	}
	return g.sacrificeCandidatesLocked(playerID, spec)
}

type permanentChoiceLeg struct {
	prompt ChooseCardsPrompt
	epochs map[uuid.UUID]int // immutable after construction
}

type permanentChoiceReceipt struct {
	seat   uuid.UUID
	cards  []uuid.UUID
	epochs map[uuid.UUID]int // the initial candidates' immutable epochs
}

// EarlierPermanentChoicesForEffect returns earlier public choices, in the
// supplied seat order. Copies prevent a consumer from mutating the frame.
func (c *PendingChoice) EarlierPermanentChoicesForEffect() []SeatCards {
	if c == nil || c.chooseCardsResume == nil {
		return nil
	}
	return copyPermanentChoices(c.chooseCardsResume.earlierPermanentChoices)
}

func copyPermanentChoices(choices []SeatCards) []SeatCards {
	if len(choices) == 0 {
		return nil
	}
	out := make([]SeatCards, len(choices))
	for i, choice := range choices {
		out[i] = SeatCards{Seat: choice.Seat, Cards: append([]uuid.UUID(nil), choice.Cards...)}
	}
	return out
}

// ChoosePermanentsInOrderThenForEffect collects public choices one seat at a
// time (CR 101.4b), then runs one action. The caller supplies APNAP order.
// No permanent moves while choosing. A leg with a forced valid set needs no
// prompt; empty or departed legs are skipped. Leg-level Then is rejected.
// Receipts bind to the original objects, so leaving and returning never
// makes a replacement object part of the final action. Caller must hold g.mu.
func (g *Game) ChoosePermanentsInOrderThenForEffect(legs []ChooseCardsPrompt, then func(*Game, PromptedPicks) error) error {
	prepared := make([]permanentChoiceLeg, len(legs))
	for i, leg := range legs {
		if leg.Then != nil || (leg.Zone != "" && leg.Zone != ZoneBattlefield) {
			return ErrInvalidParam
		}
		leg.Zone = ZoneBattlefield
		leg.promptRun = uuid.Nil
		leg.Cards = append([]uuid.UUID(nil), leg.Cards...)
		epochs := make(map[uuid.UUID]int, len(leg.Cards))
		seen := make(map[uuid.UUID]bool, len(leg.Cards))
		for _, id := range leg.Cards {
			if seen[id] {
				return ErrInvalidParam
			}
			seen[id] = true
			if c := findBattlefieldCard(g, id); c != nil {
				epochs[id] = c.ObjectEpoch
			}
		}
		prepared[i] = permanentChoiceLeg{prompt: leg, epochs: epochs}
	}
	return g.collectPermanentChoicesLocked(prepared, nil, then)
}

func (g *Game) permanentChoicePicksLocked(receipts []permanentChoiceReceipt) PromptedPicks {
	var out PromptedPicks
	for _, receipt := range receipts {
		entry := SeatCards{Seat: receipt.seat}
		for _, id := range receipt.cards {
			if c := findBattlefieldCard(g, id); c != nil && c.ObjectEpoch == receipt.epochs[id] {
				entry.Cards = append(entry.Cards, id)
			}
		}
		out = append(out, entry)
	}
	return out
}

func (g *Game) collectPermanentChoicesLocked(legs []permanentChoiceLeg, receipts []permanentChoiceReceipt, then func(*Game, PromptedPicks) error) error {
	if len(legs) == 0 {
		if then != nil {
			return then(g, g.permanentChoicePicksLocked(receipts))
		}
		return nil
	}
	leg := legs[0]
	p := leg.prompt
	player := g.playerByIDLocked(p.Chooser)
	if player == nil || player.Eliminated {
		return g.collectPermanentChoicesLocked(legs[1:], receipts, then)
	}
	var cards []uuid.UUID
	for _, id := range p.Cards {
		epoch, ok := leg.epochs[id]
		if c := findBattlefieldCard(g, id); ok && c != nil && c.ObjectEpoch == epoch {
			cards = append(cards, id)
		}
	}
	if len(cards) == 0 {
		return g.collectPermanentChoicesLocked(legs[1:], receipts, then)
	}
	p.Cards = cards
	p.permanentEpochs = leg.epochs
	p.earlierPermanentChoices = copyPermanentChoices([]SeatCards(g.permanentChoicePicksLocked(receipts)))
	// Each continuation owns a new receipt slice. The remaining legs and
	// epoch maps are immutable, so clones/undo may share the closure safely.
	seat := p.Chooser
	p.Then = func(g *Game, picked []uuid.UUID) error {
		next := make([]permanentChoiceReceipt, len(receipts), len(receipts)+1)
		copy(next, receipts)
		next = append(next, permanentChoiceReceipt{seat: seat, cards: append([]uuid.UUID(nil), picked...), epochs: leg.epochs})
		return g.collectPermanentChoicesLocked(legs[1:], next, then)
	}
	hi := p.Max
	if hi <= 0 || hi > len(cards) {
		hi = len(cards)
	}
	lo := max(0, min(p.Min, hi))
	if lo == len(cards) && hi == len(cards) {
		values := make([]Card, 0, len(cards))
		for _, id := range cards {
			c, _ := g.LookupCardForEffect(id)
			values = append(values, c)
		}
		if p.Validate == nil || p.Validate(values) {
			return p.Then(g, cards)
		}
	}
	g.QueueChooseCardsForEffect(p)
	return nil
}
