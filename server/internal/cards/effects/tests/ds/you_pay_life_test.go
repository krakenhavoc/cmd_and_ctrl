package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// you_pay_life_test.go — #6234: effects.YouPayLife is "Whenever you pay
// life" (Font of Agonies). A payment by the source's controller matches;
// another player's payment, damage and an effect's life loss do not.
func TestYouPayLifeMatchesOnlyYourPayments(t *testing.T) {
	me, opp := uuid.New(), uuid.New()
	src := &game.Card{Controller: me}
	paid := game.Event{Kind: game.EventChangeLife, Target: me, Amount: -2, LifePayment: true}
	if !effects.YouPayLife(paid, src, game.Characteristic{}, nil) {
		t.Error("your life payment did not match")
	}
	for name, ev := range map[string]game.Event{
		"an opponent's payment": {Kind: game.EventChangeLife, Target: opp, Amount: -2, LifePayment: true},
		"damage or a loss":      {Kind: game.EventChangeLife, Target: me, Amount: -2},
		"a gain":                {Kind: game.EventChangeLife, Target: me, Amount: 2},
	} {
		if effects.YouPayLife(ev, src, game.Characteristic{}, nil) {
			t.Errorf("%s matched", name)
		}
	}
}
