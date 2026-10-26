package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep02_helpers.go — #4253's closure sweep, batch 02: the small shared
// shapes the batch's cards key their prompts with.
//
// #4253 slice 10b merged the batch's two duplicates into the shared forms:
// itemChoice is askedPick (asked_pick.go), and ChoosePlayerKeyed and its
// ChoosePlayerThen became ChoosePlayer.ThenKey and ChoosePlayerThen
// (choose_player.go), which also carry the asking item and record the answer
// on it. The keys the batch registered are unchanged.

// itemChoice is askedPick under the name batch 02's cards use: a keyed
// choose_cards continuation that runs against a Context rebuilt from the
// asking item. The key is an on-disk identity ("choose-cards/<card>-<what>"),
// ledgered: never renamed, never reused.
func itemChoice(key string, body func(ctx *Context, r game.CardsChosen) error) game.ChooseCardsThen {
	return askedPick(key, body)
}
