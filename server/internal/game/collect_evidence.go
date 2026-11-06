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

// CollectEvidenceFromTargetsCost is the ExileCost clause "collect evidence
// X, where X is the total mana value of the permanents this spell targets"
// is (Urgent Necropsy, #5240): collect evidence's clause with no printed
// floor, the floor read off the cast's targets (ManaValueFromTargets). The
// label is the printed words, with the X left in.
func CollectEvidenceFromTargetsCost(label string) *ExileCost {
	return &ExileCost{
		AnyNumber:            true,
		Label:                label,
		From:                 ZoneGraveyard,
		ManaValueFromTargets: true,
	}
}

// IsCollectEvidenceFromTargets reports whether `e` is the target-sized
// collect evidence clause (CollectEvidenceFromTargetsCost) and nothing
// else. Nil-safe.
func (e *ExileCost) IsCollectEvidenceFromTargets() bool {
	return e != nil && e.ManaValueFromTargets && e.AnyNumber && e.ManaValueAtLeast == 0 &&
		e.SymbolTotal == nil && e.Match == nil && e.N == 0 && !e.CountFromX && e.Zone() == ZoneGraveyard
}

// CollectEvidenceTargetsFloorForEffect is the X of "collect evidence X,
// where X is the total mana value of the permanents this spell targets":
// the sum of the mana values of the distinct permanents among `targets`
// (PermanentManaValueForEffect). A player, a card in another zone and a
// spell are not permanents and add nothing, and a permanent chosen for two
// of the spell's "target" words is one permanent the spell targets (CR
// 115.3), so it is counted once. No targets is X = 0, which any cards pay,
// none included (CR 701.59a).
//
// CastSpell reads it once, after the targets are chosen (CR 601.2c) and as
// the total cost is determined (CR 601.2f), so the number is fixed for the
// payment that follows; the view's per-permanent values and the bot read the
// same function, so the three agree (#544).
//
// CALLER MUST ALREADY HOLD g.mu (read or write).
func (g *Game) CollectEvidenceTargetsFloorForEffect(targets []TargetRef) int {
	seen := make(map[uuid.UUID]bool, len(targets))
	total := 0
	for _, t := range targets {
		if t.Kind != TargetCard || seen[t.ID] {
			continue
		}
		seen[t.ID] = true
		if mv, ok := g.PermanentManaValueForEffect(t.ID); ok {
			total += mv
		}
	}
	return total
}

// PermanentManaValueForEffect is the mana value of the permanent `id` as
// the engine reads it (ManaValueForEffect: a face-down permanent's is 0, CR
// 708.2a and 202.3a), and ok false when `id` is not on the battlefield or
// its cost can't be read.
//
// CALLER MUST ALREADY HOLD g.mu (read or write).
func (g *Game) PermanentManaValueForEffect(id uuid.UUID) (int, bool) {
	c := g.findBattlefieldCardLocked(id)
	if c == nil {
		return 0, false
	}
	return g.ManaValueForEffect(*c)
}

// PermanentManaValuesForEffect is every permanent's mana value
// (PermanentManaValueForEffect), keyed by instance ID: what the view ships
// beside a target-sized collect evidence cost, so the client sums the
// numbers the engine will. Caller must hold g.mu.
func (g *Game) PermanentManaValuesForEffect() map[uuid.UUID]int {
	if g.Battlefield == nil {
		return nil
	}
	out := make(map[uuid.UUID]int, len(g.Battlefield.Cards))
	for i := range g.Battlefield.Cards {
		c := g.Battlefield.Cards[i]
		if mv, ok := g.ManaValueForEffect(c); ok {
			out[c.InstanceID] = mv
		}
	}
	return out
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
// evidence clause (an announced OptionalCollectEvidence, or a mandatory
// target-sized one, #5240), so the cast path emits EventCollectEvidence for
// the cards it exiled. The target-sized one is paid on every cast, collect
// evidence 0 included.
func planCollectsEvidence(plan []costPayment) bool {
	for _, pay := range plan {
		if ec := pay.cost.ExileCards; ec.IsCollectEvidence() || ec.IsCollectEvidenceFromTargets() {
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
