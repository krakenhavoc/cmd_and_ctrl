package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Call of the Ring — Enchantment {1}{B}:
//
//	"At the beginning of your upkeep, the Ring tempts you.
//	 Whenever you choose a creature as your Ring-bearer, you may pay
//	 2 life. If you do, draw a card."
//
// The upkeep trigger is the keyword action (ADR 0114, CR 701.54). The
// second ability is WheneverYouChooseARingBearer: it triggers only when
// a creature was chosen — "If the Ring tempts you but you can't choose
// a creature as your Ring-bearer …, Call of the Ring's last ability
// won't trigger" — and it does trigger when the creature chosen already
// was the Ring-bearer (2023-06-16 rulings). Any temptation counts, not
// only the upkeep one.
//
// The "you may pay 2 life" is asked as the trigger resolves (CR
// 608.2d, MayChoice), and the payment is re-checked there: a player
// with less than 2 life cannot pay (CR 119.4) and draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9fcb920c-b8d7-4a79-a335-f63050182cca",
		Name:         "Call of the Ring",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Call of the Ring — the Ring tempts you", Do(TheRingTemptsYou{})),
			WheneverYouChooseARingBearer("Call of the Ring — you may pay 2 life to draw a card", callOfTheRingPayAndDraw),
		},
	})
}

// callOfTheRingPayAndDraw is the second ability's body: the "you may",
// then the payment and the draw it buys.
func callOfTheRingPayAndDraw(g *game.Game, item *game.StackItem) error {
	return MayChoice{
		Question: "Call of the Ring — pay 2 life to draw a card?",
		LifeCost: 2,
		OnYes: func(ctx *Context) error {
			return payLifeThenDraw(ctx, 2, 1)
		},
	}.Apply(NewContext(g, item))
}
