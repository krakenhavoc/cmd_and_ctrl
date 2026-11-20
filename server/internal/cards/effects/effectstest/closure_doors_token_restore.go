package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// closure_doors_token_restore_test.go — #4253 slice 6, the token half: the
// "then" of a token creation is a registered tail, so the CR 616 order
// prompt a Doubling Season beside a Parallel Lives asks is a restore point.
// Each test takes a STRICT restore while the prompt is open, answers it on
// the restored table and checks the card still finishes off the disk.
// (The helpers libTop and tokenOrderBoard are in
// closure_doors_restore_test.go.)

// orderPromptOpen reports whether the one open prompt is a CR 616 order.
func OrderPromptOpen(g *game.Game) bool {
	return len(g.PendingChoices) == 1 && g.PendingChoices[0].Kind == game.PendingChoiceReplacementOrder
}
