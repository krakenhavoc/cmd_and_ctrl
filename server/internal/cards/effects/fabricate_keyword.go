package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fabricate_keyword.go — the fabricate keyword (CR 702.123) as a
// constructor a card file spells in one line:
//
//	Triggered: []game.TriggeredAbility{Fabricate("Marionette Apprentice", 1)},
//
// "Fabricate N" is "when this permanent enters, you may put N +1/+1
// counters on it. If you don't, create N 1/1 colorless Servo artifact
// creature tokens." The choice is made as the trigger resolves, and
// CR 702.123a adds the clause this shape depends on: a permanent that
// has left the battlefield by then can't take the counters, so the
// tokens are made without asking.

// Fabricate is the keyword's trigger.
func Fabricate(cardName string, n int) game.TriggeredAbility {
	label := fmt.Sprintf("%s — fabricate %d", cardName, n)
	return WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
		return fabricateChoice(g, item, label, n)
	})
}

// fabricateChoice asks counters-or-Servos, or makes the Servos when
// there is nothing to put counters on.
func fabricateChoice(g *game.Game, item *game.StackItem, label string, n int) error {
	ctx := NewContext(g, item)
	if !onBattlefield(g, item.SourceCardID) {
		return fabricateServos(ctx, n)
	}
	return MayChoice{
		Question: label,
		YesLabel: fmt.Sprintf("Put %d +1/+1 counter(s) on it", n),
		NoLabel:  fmt.Sprintf("Create %d 1/1 colorless Servo artifact creature token(s)", n),
		OnYes: func(ctx *Context) error {
			return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
		},
		OnNo: func(ctx *Context) error { return fabricateServos(ctx, n) },
	}.Apply(ctx)
}

// fabricateServos is the "if you don't" half.
func fabricateServos(ctx *Context, n int) error {
	return CreateToken{Template: TokenCard("1/1 colorless Servo artifact"), N: n}.Apply(ctx)
}
