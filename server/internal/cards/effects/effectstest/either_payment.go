package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// testAnyPlayerProbe is an any-player offer's two consequences as a
// registered key (#4253 retired the OnPaid / OnNobody closures): it counts
// what ran, and who paid, in the package-level probe below.
var (
	TestAnyPlayerPaid, TestAnyPlayerNobody int
	TestAnyPlayerPayer                     uuid.UUID
	TestAnyPlayerProbe                     = effects.AnyPlayerConsequences("test/4253-any-player-probe",
		func(_ *effects.Context, r game.AnyPlayerAnswer) error {
			TestAnyPlayerPaid++
			TestAnyPlayerPayer = r.Payer
			return nil
		},
		func(*effects.Context, game.AnyPlayerAnswer) error { TestAnyPlayerNobody++; return nil },
		nil)
)

// resetAnyPlayerProbe zeroes testAnyPlayerProbe's counts.
func ResetAnyPlayerProbe() {
	TestAnyPlayerPaid, TestAnyPlayerNobody, TestAnyPlayerPayer = 0, 0, uuid.Nil
}
