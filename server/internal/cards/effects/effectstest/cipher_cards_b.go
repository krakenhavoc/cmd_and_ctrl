package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cipher_cards_b_test.go — the cipher cards' spell text (#3867): Shadow
// Slice, Stolen Identity, Undercity Plague, Voidwalk, Whispering Madness
// and Writ of Return, and the whole cipher loop for two of them (encode,
// combat damage to a player, the free copy with new targets).

const (
	CibShadowSliceOracle       = "f4d2eca4-4645-4044-b75f-442cd726cedf"
	CibStolenIdentityOracle    = "30b533a3-afc3-4500-a8f0-54e84301988f"
	CibUndercityPlagueOracle   = "aa997e63-be19-4b1a-8972-eebc8a720ded"
	CibVoidwalkOracle          = "6706ebe4-4f96-4788-a934-f9b9aac5746f"
	CibWhisperingMadnessOracle = "7bbc5f3a-f8fb-47e8-837e-ce319bc489de"
	CibWritOfReturnOracle      = "47735ef2-fc4c-4634-83ab-715b1e86da0c"
)

func CibPlayer(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}
