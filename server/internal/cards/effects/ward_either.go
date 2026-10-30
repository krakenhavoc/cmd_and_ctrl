package effects

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ward_either.go — #4566: a ward with two ways to pay. Titania, Rugged
// Rumbler prints "Ward—Discard a card or pay {2}": CR 702.21a's "counter
// that spell or ability unless that player pays [cost]", where the cost
// is a choice. CR 118.12a reads it as "that player may pay [cost]; if they
// don't, counter it", so it is the guarded pay-unless every ward uses
// (QueueCounterUnlessPaidForEffect), carrying an either payment (ADR 0187
// §5): one question, one answer naming the branch, and a decline that
// counters. The mana branch is game.PayActionMana, paid by the prompt's
// own mana path; the discard branch is the payer's own pick from hand.

// validateEither panics on a malformed either/or ward at catalog init:
// fewer than two branches, a component beside the branches, or a branch
// that is not a plain mana, life or chosen-discard ward (the shapes an
// either payment can carry).
func (c WardCost) validateEither() {
	if len(c.Either) < 2 {
		panic(fmt.Sprintf("effects: an either/or ward needs two or more branches: %+v", c))
	}
	rest := c
	rest.Either = nil
	if rest.Mana != "" || rest.Waterbend != "" || rest.Life != 0 || rest.Sacrifice != nil || rest.Discard != 0 || rest.DiscardAtRandom || rest.DiscardKind != nil || rest.Poison != 0 {
		panic(fmt.Sprintf("effects: an either/or ward has a component beside its branches: %+v", c))
	}
	for _, b := range c.Either {
		if wardBranchAction(b) == nil {
			panic(fmt.Sprintf("effects: an either/or ward branch must be WardMana, WardLife or WardDiscard: %+v", b))
		}
	}
}

// wardBranchAction is one branch of an either/or ward as a payment, or
// nil for a shape the either payment can't carry.
func wardBranchAction(b WardCost) *game.PayAction {
	switch {
	case len(b.Either) > 0 || b.Waterbend != "" || b.Sacrifice != nil || b.DiscardAtRandom || b.DiscardKind != nil:
		return nil
	case b.Mana != "" && b.Life == 0 && b.Discard == 0:
		if _, err := game.ParseCost(b.Mana); err != nil {
			return nil
		}
		return &game.PayAction{Kind: game.PayActionMana, Mana: b.Mana, Words: "pay " + b.Mana}
	case b.Life > 0 && b.Mana == "" && b.Discard == 0:
		return &game.PayAction{Kind: game.PayActionLife, Count: b.Life, Words: "pay " + strconv.Itoa(b.Life) + " life"}
	case b.Discard > 0 && b.Mana == "" && b.Life == 0:
		return &game.PayAction{Kind: game.PayActionDiscard, Count: b.Discard, Words: "discard " + wardDiscardLabel(b.Discard, false, nil)}
	}
	return nil
}

// wardEitherAction is the whole either/or ward as one either payment.
func wardEitherAction(cost WardCost) *game.PayAction {
	branches := make([]game.PayAction, 0, len(cost.Either))
	for _, b := range cost.Either {
		if a := wardBranchAction(b); a != nil {
			branches = append(branches, *a)
		}
	}
	return &game.PayAction{Kind: game.PayActionEither, Either: branches}
}

// wardEitherOrCounter is "counter it unless that player <pays one of the
// branches>". A payer who can pay no branch at all can't pay (CR 118.3),
// so the spell is countered with no prompt, as a discard ward with an
// empty hand is.
func wardEitherOrCounter(g *game.Game, source, payer uuid.UUID, cost WardCost, targetingItem uuid.UUID) error {
	action := wardEitherAction(cost)
	if !g.PayActionCanPayForEffect(payer, action) {
		return wardCounter(g, targetingItem)
	}
	phrases := make([]string, 0, len(action.Either))
	for _, b := range action.Either {
		phrases = append(phrases, b.Words)
	}
	words := strings.Join(phrases, " or ")
	return g.QueueCounterUnlessPaidForEffect(game.CounterUnlessPaidPrompt{
		StackItem: targetingItem,
		Chooser:   payer,
		Source:    source,
		Cost:      strings.ToUpper(words[:1]) + words[1:],
		Action:    action,
		Question:  "Ward — " + words + ", or the spell is countered",
	})
}
