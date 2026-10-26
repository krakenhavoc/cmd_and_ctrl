package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// asked_pick.go — #4253 sweep: a choose_cards pick whose continuation reads
// the item that asked it, written as a registered body plus the item as data.
//
// A pick queued with ThenKey holds no closure, so a table waiting on it is a
// restore point. What the old closure captured is plain data on the prompt:
// the asking item (ChooseCardsPrompt.Asked, game.NewAskedItem), the IDs
// (Carry), the numbers (CarryNums) and the one word (CarryWord). The body is
// handed a Context rebuilt from that item, so it reads the controller, the
// source card and object, the targets, the X and the trigger's event as it did.

// askedPick registers a keyed choose_cards continuation that runs against a
// Context rebuilt from the prompt's asking item (ChooseCardsPrompt.Asked).
// A prompt that carried no item runs against a stand-in with the chooser as
// the controller and the asking card as the source. Call it once, from a
// package-level var. The key is an on-disk identity ("choose-cards/<card>-<what>"),
// ledgered: never renamed, never reused.
func askedPick(key string, body func(ctx *Context, r game.CardsChosen) error) game.ChooseCardsThen {
	return askedPickRule(key, nil, body)
}

// askedPickRule is askedPick for a pick with a set rule over what is chosen
// (game.RegisterChooseCardsThenRule): the rule is a function of the cards
// alone and captures nothing.
func askedPickRule(key string, rule game.SetRuleFunc, body func(ctx *Context, r game.CardsChosen) error) game.ChooseCardsThen {
	return game.RegisterChooseCardsThenRule(key, rule, func(g *game.Game, r game.CardsChosen) error {
		ctx, _ := askedContext(g, r.Asked, r.Chooser, r.Source, nil)
		return body(ctx, r)
	})
}
