package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// ability_cost_either_test.go — #4566: ChosenAbilityCost, the fold every
// reader of an either/or activation cost goes through.

func TestChosenAbilityCostFoldsTheNamedBranch(t *testing.T) {
	cost := AbilityCost{Mana: "{3}", Tap: true, Either: []AbilityCostBranch{
		{Kind: AbilityCostBranchSacrifice, Count: 1, Label: "an artifact", Of: PermanentQuery{Types: []string{"artifact"}}},
		{Kind: AbilityCostBranchDiscard, Count: 1, Label: "a nonland card", Nonland: true},
	}}

	zero, one := 0, 1
	got, err := ChosenAbilityCost(cost, &zero)
	if err != nil || got.SacrificeOther == nil || got.DiscardCards != nil || got.Branched() || got.Mana != "{3}" || !got.Tap {
		t.Fatalf("branch 0 = %+v, %v; want the sacrifice folded into {3}, {T}", got, err)
	}
	if s := got.SacrificeOther; s.Min != 1 || s.Max != 1 || s.Label != "an artifact" ||
		!s.CardOK(nil, uuid.Nil, Card{TypeLine: "Artifact"}, ZoneBattlefield) ||
		s.CardOK(nil, uuid.Nil, Card{TypeLine: "Creature — Bear"}, ZoneBattlefield) {
		t.Errorf("the sacrifice clause %+v does not take exactly one artifact", s)
	}
	got, err = ChosenAbilityCost(cost, &one)
	if err != nil || got.DiscardCards == nil || got.SacrificeOther != nil || got.Branched() {
		t.Fatalf("branch 1 = %+v, %v; want the discard folded in", got, err)
	}
	if d := got.DiscardCards; d.N != 1 || d.Match == nil || d.Match(Card{TypeLine: "Basic Land — Forest"}) || !d.Match(Card{TypeLine: "Instant"}) {
		t.Errorf("the discard clause %+v does not take exactly one nonland card", d)
	}
	if len(cost.Either) != 2 {
		t.Error("the fold wrote through to the catalog's cost")
	}
	if cost.Either[0].BranchLabel() != "Sacrifice an artifact" || cost.Either[1].BranchLabel() != "Discard a nonland card" {
		t.Errorf("labels %q / %q", cost.Either[0].BranchLabel(), cost.Either[1].BranchLabel())
	}
}

// No branch is ever guessed: a branched cost needs one in range, and an
// unbranched one refuses any (ErrAbilityCostBranch, an ErrInvalidParam).
func TestChosenAbilityCostRefusesAMissingOrStrayBranch(t *testing.T) {
	branched := AbilityCost{Either: []AbilityCostBranch{{Kind: AbilityCostBranchDiscard, Count: 1, Label: "a card"}, {Kind: AbilityCostBranchSacrifice, Count: 1, Label: "a permanent"}}}
	two, neg, zero := 2, -1, 0
	for name, branch := range map[string]*int{"missing": nil, "past the end": &two, "negative": &neg} {
		if _, err := ChosenAbilityCost(branched, branch); !errors.Is(err, ErrAbilityCostBranch) || !errors.Is(err, ErrInvalidParam) {
			t.Errorf("%s: err %v", name, err)
		}
	}
	plain := AbilityCost{Mana: "{1}"}
	if _, err := ChosenAbilityCost(plain, &zero); !errors.Is(err, ErrAbilityCostBranch) {
		t.Errorf("a branch on an unbranched cost: err %v", err)
	}
	if got, err := ChosenAbilityCost(plain, nil); err != nil || got.Mana != "{1}" {
		t.Errorf("an unbranched cost = %+v, %v", got, err)
	}
}

// A PayActionMana branch names no cards and is payable only when the
// mana can be paid; a nonland discard takes no land (#4566).
func TestEitherPaymentManaAndNonlandBranches(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	mana := &PayAction{Kind: PayActionMana, Mana: "{2}"}
	if !mana.NamesNothing() {
		t.Error("a mana branch names cards")
	}
	if g.payActionPayableLocked(p.ID, mana) {
		t.Error("{2} payable from an empty pool with no sources")
	}
	p.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
	if !g.payActionPayableLocked(p.ID, mana) {
		t.Error("{2} not payable from two in the pool")
	}
	if PayActionPhrase(*mana) != "pay {2}" {
		t.Errorf("phrase %q", PayActionPhrase(*mana))
	}

	land := NewCard("Forest", p.ID)
	land.TypeLine = "Basic Land — Forest"
	spell := NewCard("Shock", p.ID)
	spell.TypeLine = "Instant"
	p.Hand.PushTop(land)
	p.Hand.PushTop(spell)
	nonland := &PayAction{Kind: PayActionDiscard, Count: 1, Nonland: true}
	opts := map[uuid.UUID]bool{}
	for _, id := range g.payActionOptionsLocked(p.ID, nonland) {
		opts[id] = true
	}
	if !opts[spell.InstanceID] || opts[land.InstanceID] {
		t.Errorf("nonland discard options %v: want the spell and not the land", opts)
	}
	if n := len(g.payActionOptionsLocked(p.ID, &PayAction{Kind: PayActionDiscard, Count: 1})); n != p.Hand.Size() {
		t.Errorf("a plain discard offers %d of %d cards", n, p.Hand.Size())
	}
}
