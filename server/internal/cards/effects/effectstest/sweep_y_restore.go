package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- a "may sacrifice" whose run carries more than one number -----------

// testSacrificedNums is what testSacrificedTwoNums was handed.
var TestSacrificedNums []int

// testSacrificedTwoNums records the numbers its run carried.
var TestSacrificedTwoNums = effects.RegisterSacrificedOneRun("test/4253-y-sacrificed-two-nums", func(_ *effects.Context, _ uuid.UUID, r game.RunDone) error {
	TestSacrificedNums = append([]int(nil), r.Nums...)
	return nil
})
