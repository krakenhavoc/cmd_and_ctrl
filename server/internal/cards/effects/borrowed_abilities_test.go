package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// borrowed_abilities_test.go — ADR 0155 Decision 1's registration guard:
// a borrow is a layer-6 static that stands alone, names its lenders and
// its clause, and never sits inside a granted bundle.

func fixtureBorrow() game.AbilityBorrow {
	return game.AbilityBorrow{
		Lenders: func(*game.Game, *game.Card) []*game.Card { return nil },
		Label:   "Has all activated abilities of the fixture.",
	}
}

func TestCheckStaticBorrowsRefusesWhatABorrowCannotSay(t *testing.T) {
	wrongLayer := HasActivatedAbilitiesOf(fixtureBorrow())
	wrongLayer.Layer = game.Layer7PT
	expectRegisterPanic(t, "layer-6 effect", func() {
		checkStaticGrants("Fixture", "static", []game.StaticAbility{wrongLayer})
	})
	withRemoval := HasActivatedAbilitiesOf(fixtureBorrow())
	withRemoval.RemovesAbilities = true
	expectRegisterPanic(t, "beside a removal", func() {
		checkStaticGrants("Fixture", "static", []game.StaticAbility{withRemoval})
	})
	withGrant := HasActivatedAbilitiesOf(fixtureBorrow())
	withGrant.GrantAbilities = []string{"fixture/x"}
	expectRegisterPanic(t, "beside a removal", func() {
		checkStaticGrants("Fixture", "static", []game.StaticAbility{withGrant})
	})
	noLenders := fixtureBorrow()
	noLenders.Lenders = nil
	expectRegisterPanic(t, "no Lenders", func() {
		checkStaticGrants("Fixture", "static", []game.StaticAbility{HasActivatedAbilitiesOf(noLenders)})
	})
	noLabel := fixtureBorrow()
	noLabel.Label = ""
	expectRegisterPanic(t, "no printed Label", func() {
		checkStaticGrants("Fixture", "static", []game.StaticAbility{HasActivatedAbilitiesOf(noLabel)})
	})
	// The right shape is fine, for itself and for others.
	checkStaticGrants("Fixture", "static", []game.StaticAbility{
		HasActivatedAbilitiesOf(fixtureBorrow()),
		GrantActivatedAbilitiesOf(func(*game.Card, *game.Game, *game.Card) bool { return true }, fixtureBorrow()),
	})
}

// A bundle that lends abilities would put a borrow inside a borrow.
func TestCheckGrantsRefusesABorrowInsideABundle(t *testing.T) {
	expectRegisterPanic(t, "borrows abilities", func() {
		checkGrants("Fixture", []AbilityGrant{{Key: "fixture/lends", Text: "x",
			Static: []game.StaticAbility{HasActivatedAbilitiesOf(fixtureBorrow())}}})
	})
}

// HasActivatedAbilitiesOf's one recipient is its own source.
func TestHasActivatedAbilitiesOfAppliesToItsSourceAlone(t *testing.T) {
	s := HasActivatedAbilitiesOf(fixtureBorrow())
	src := &game.Card{Name: "Ooze"}
	src.InstanceID = [16]byte{1}
	other := &game.Card{Name: "Bear"}
	other.InstanceID = [16]byte{2}
	if !s.AppliesTo(src, nil, src) || s.AppliesTo(other, nil, src) {
		t.Error("HasActivatedAbilitiesOf should apply to its source and nothing else")
	}
	if s.Borrows == nil || s.Borrows.Label == "" || s.Layer != game.Layer6Ability {
		t.Errorf("static = %+v", s)
	}
}
