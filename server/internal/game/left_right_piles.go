package game

import (
	"github.com/google/uuid"
)

// left_right_piles.go — Raging River (#6200): "Whenever one or more
// creatures you control attack, each defending player divides all
// creatures without flying they control into a "left" pile and a "right"
// pile. Then, for each attacking creature you control, choose "left" or
// "right." That creature can't be blocked this combat except by creatures
// with flying and creatures in a pile with the chosen label."
//
//	700.3a Each of the affected objects must be put into exactly one of
//	       those piles, unless the effect specifies otherwise.
//	700.3d A pile can contain zero or more objects.
//	101.4. If multiple players would make choices and/or take actions at
//	       the same time, the active player … makes any choices required,
//	       then the next player in turn order …
//	509.1b The defending player checks each creature they control to see
//	       whether it's affected by any restrictions …
//
// Two keyed legs, both ordinary prompts the client and the bot already
// answer:
//
//  1. each defending player, in APNAP order, picks their LEFT pile from
//     the creatures without flying they control (choose_cards, any
//     number); the rest are their right pile (CR 700.3a);
//  2. the trigger's controller labels each attacking creature they
//     control, one option pick ("Left" / "Right") per attacker.
//
// Each label is a cantBeBlockedExceptBy record pinned to that attacker
// for the rest of combat: blockers with flying, or the creatures in a
// pile with that label (every defending player's), as the objects they
// were when the piles were made (CR 400.7: a creature that leaves and
// comes back, or arrives later, is in no pile). Everything rides the
// prompts as data, so a table waiting on either leg is a restore point.

var (
	leftRightPilesSplit ChooseCardsThen
	leftRightPilesLabel OptionPickThen
)

func init() {
	// On-disk keys: never renamed, never reused.
	leftRightPilesSplit = RegisterChooseCardsThen("choose-cards/left-right-piles", leftRightPilesSplitAnswered)
	leftRightPilesLabel = RegisterOptionPickThen("option-pick/left-right-piles-label", leftRightPilesLabelAnswered)
}

// QueueLeftRightPilesForEffect runs Raging River's effect for the
// attacking player `controller`, whose card is `source`.
//
// Caller must hold g.mu (write).
func (g *Game) QueueLeftRightPilesForEffect(source, controller uuid.UUID) {
	g.askNextLeftPileLocked(source, controller, uuid.Nil, nil, nil)
}

// leftPileCandidatesLocked is the creatures without flying `seat`
// controls.
//
// Caller must hold g.mu (read or write), with fresh layers.
func (g *Game) leftPileCandidatesLocked(seat uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == seat && c.IsCreature() && !HasKeyword(c, "flying") {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// askNextLeftPileLocked asks the next defending player after `after`
// (the first, when uuid.Nil) for their left pile, and, with nobody left
// to ask, moves on to the labels. `left` and `right` are the piles made
// so far, every defending player's together.
//
// Caller must hold g.mu (write).
func (g *Game) askNextLeftPileLocked(source, controller, after uuid.UUID, left, right []uuid.UUID) {
	g.RecomputeLayersIfStaleLocked()
	seats := g.defendingSeatsAPNAPLocked()
	start := 0
	if after != uuid.Nil {
		start = len(seats)
		for i, s := range seats {
			if s == after {
				start = i + 1
				break
			}
		}
	}
	for _, seat := range seats[start:] {
		candidates := g.leftPileCandidatesLocked(seat)
		if len(candidates) == 0 {
			continue
		}
		carry := append(append([]uuid.UUID{controller}, left...), right...)
		if g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser:    seat,
			FromPlayer: seat,
			Source:     source,
			Question:   "Divide your creatures without flying into two piles: choose the \"left\" pile. The rest are the \"right\" pile.",
			Cards:      candidates,
			Min:        0,
			Max:        len(candidates),
			Zone:       ZoneBattlefield,
			ThenKey:    leftRightPilesSplit,
			Carry:      carry,
			CarryNums:  []int{len(left)},
		}) != uuid.Nil {
			return
		}
	}
	g.askNextLabelLocked(source, controller, left, right, g.labelAttackersLocked(controller))
}

// leftRightPilesSplitAnswered records one defending player's piles and
// asks the next. Carry is the attacking player, then the left piles so
// far, then the right; Nums[0] is how many are left. A player who left
// the game made no piles (CR 800.4a).
func leftRightPilesSplitAnswered(g *Game, r CardsChosen) error {
	if len(r.Carry) == 0 || len(r.Nums) == 0 {
		return nil
	}
	controller, rest := r.Carry[0], r.Carry[1:]
	n := r.Nums[0]
	if n > len(rest) {
		n = len(rest)
	}
	left := copyUUIDs(rest[:n])
	right := copyUUIDs(rest[n:])
	if r.Answered {
		one, two := partitionPiles(r.Candidates, r.Picked)
		left = append(left, one...)
		right = append(right, two...)
	}
	g.askNextLeftPileLocked(r.Source, controller, r.Chooser, left, right)
	return nil
}

// labelAttackersLocked is the attacking creatures `controller` controls,
// in battlefield order.
//
// Caller must hold g.mu (read or write).
func (g *Game) labelAttackersLocked(controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && c.AttackingTarget != uuid.Nil {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// askNextLabelLocked asks `controller` for the label of the first of
// `attackers` still attacking.
//
// Caller must hold g.mu (write).
func (g *Game) askNextLabelLocked(source, controller uuid.UUID, left, right, attackers []uuid.UUID) {
	for len(attackers) > 0 {
		atk := findBattlefieldCard(g, attackers[0])
		if atk == nil || atk.AttackingTarget == uuid.Nil || atk.Controller != controller {
			attackers = attackers[1:]
			continue
		}
		carry := append(append(append([]uuid.UUID(nil), left...), right...), attackers...)
		if g.QueueOptionPickForEffect(OptionPickPrompt{
			Chooser:  controller,
			Source:   source,
			Question: "Choose \"left\" or \"right\" for " + atk.Name + ": it can't be blocked this combat except by creatures with flying and creatures in a pile with that label.",
			Options: []ChoiceOption{
				{Label: "Left", Cards: []uuid.UUID{atk.InstanceID}},
				{Label: "Right", Cards: []uuid.UUID{atk.InstanceID}},
			},
			ThenKey: leftRightPilesLabel,
			Carry:   carry,
			Nums:    []int{len(left), len(right)},
		}) != uuid.Nil {
			return
		}
		attackers = attackers[1:]
	}
}

// leftRightPilesLabelAnswered pins the chosen label's restriction to the
// attacker being labelled and asks for the next one. Carry is the left
// piles, the right piles, then the attackers still to label (the first
// is this one's); Nums the two piles' sizes. Unanswered (the chooser
// left), nothing is pinned.
func leftRightPilesLabelAnswered(g *Game, r OptionPicked) error {
	if len(r.Nums) < 2 || r.Nums[0]+r.Nums[1] > len(r.Carry) {
		return nil
	}
	nl, nr := r.Nums[0], r.Nums[1]
	left := copyUUIDs(r.Carry[:nl])
	right := copyUUIDs(r.Carry[nl : nl+nr])
	attackers := copyUUIDs(r.Carry[nl+nr:])
	if len(attackers) == 0 || r.Option == nil {
		return nil
	}
	pile, text := left, "creatures with flying and creatures in a \"left\" pile"
	if r.Option.Label == "Right" {
		pile, text = right, "creatures with flying and creatures in a \"right\" pile"
	}
	g.pinPileLabelLocked(r.Source, attackers[0], pile, text)
	g.askNextLabelLocked(r.Source, r.Chooser, left, right, attackers[1:])
	return nil
}

// pinPileLabelLocked registers the attacker's restriction for the rest
// of combat. Nothing outside a combat phase or for an attacker that has
// gone.
//
// Caller must hold g.mu (write).
func (g *Game) pinPileLabelLocked(source, attacker uuid.UUID, pile []uuid.UUID, text string) {
	d, ok := g.UntilEndOfCombatDuration()
	if !ok {
		return
	}
	affected := g.PinnedObjectsLocked(attacker)
	if len(affected) == 0 {
		return
	}
	var objects []ObjectRef
	for _, id := range pile {
		if c := findBattlefieldCard(g, id); c != nil {
			objects = append(objects, ObjectRef{ID: id, Epoch: c.ObjectEpoch})
		}
	}
	g.appendScopedEffectLocked(source, affected, ScopeNone, uuid.Nil,
		[]Mod{CantBeBlockedExceptByObjectsMod([]string{"flying"}, objects, text)},
		d, "can't be blocked this combat except by "+text, timeNowUnixNano())
}
