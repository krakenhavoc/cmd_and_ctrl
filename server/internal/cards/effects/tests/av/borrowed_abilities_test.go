package testsav

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	wrongLayer := effects.HasActivatedAbilitiesOf(fixtureBorrow())
	wrongLayer.Layer = game.Layer7PT
	effectstest.ExpectRegisterPanic(t, "layer-6 effect", func() {
		effects.CheckStaticGrants("Fixture", "static", []game.StaticAbility{wrongLayer})
	})
	withRemoval := effects.HasActivatedAbilitiesOf(fixtureBorrow())
	withRemoval.RemovesAbilities = true
	effectstest.ExpectRegisterPanic(t, "beside a removal", func() {
		effects.CheckStaticGrants("Fixture", "static", []game.StaticAbility{withRemoval})
	})
	withGrant := effects.HasActivatedAbilitiesOf(fixtureBorrow())
	withGrant.GrantAbilities = []string{"fixture/x"}
	effectstest.ExpectRegisterPanic(t, "beside a removal", func() {
		effects.CheckStaticGrants("Fixture", "static", []game.StaticAbility{withGrant})
	})
	noLenders := fixtureBorrow()
	noLenders.Lenders = nil
	effectstest.ExpectRegisterPanic(t, "no Lenders", func() {
		effects.CheckStaticGrants("Fixture", "static", []game.StaticAbility{effects.HasActivatedAbilitiesOf(noLenders)})
	})
	noLabel := fixtureBorrow()
	noLabel.Label = ""
	effectstest.ExpectRegisterPanic(t, "no printed Label", func() {
		effects.CheckStaticGrants("Fixture", "static", []game.StaticAbility{effects.HasActivatedAbilitiesOf(noLabel)})
	})
	// The right shape is fine, for itself and for others.
	effects.CheckStaticGrants("Fixture", "static", []game.StaticAbility{
		effects.HasActivatedAbilitiesOf(fixtureBorrow()),
		effects.GrantActivatedAbilitiesOf(func(*game.Card, *game.Game, *game.Card) bool { return true }, fixtureBorrow()),
	})
}

// A bundle that lends abilities would put a borrow inside a borrow.
func TestCheckGrantsRefusesABorrowInsideABundle(t *testing.T) {
	effectstest.ExpectRegisterPanic(t, "borrows abilities", func() {
		effects.CheckGrants("Fixture", []effects.AbilityGrant{{Key: "fixture/lends", Text: "x",
			Static: []game.StaticAbility{effects.HasActivatedAbilitiesOf(fixtureBorrow())}}})
	})
}

// HasActivatedAbilitiesOf's one recipient is its own source.
func TestHasActivatedAbilitiesOfAppliesToItsSourceAlone(t *testing.T) {
	s := effects.HasActivatedAbilitiesOf(fixtureBorrow())
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
