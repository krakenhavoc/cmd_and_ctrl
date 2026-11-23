package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cipher_cards_a_test.go — the cipher cards of #3867's batch A (CR
// 702.99): each card's own text, and the whole cipher loop (encode,
// combat damage to a player, the free copy with new targets) for Call
// of the Nightwing and Hands of Binding.

const (
	CiaArcaneHeistOracle       = "839748e7-ccb4-421e-9626-b6d8be9390ab"
	CiaCallOfTheNightwingOracl = "919d21ad-3d1e-4f80-8005-64e3f8bf64d1"
	CiaHandsOfBindingOracle    = "9bc82078-2f3e-44ee-906d-707f3430d414"
	CiaHiddenStringsOracle     = "a50737d7-dd2d-40da-817c-f913f76b5011"
	CiaMentalVaporsOracle      = "e0151f37-790b-47f4-9928-cd8f32f68e4c"
	CiaMidnightRecoveryOracle  = "7904025c-7120-4362-abe9-a9f5fb64447f"
	CiaParanoidDelusionsOracle = "129cec2c-fc1b-4ce6-98f2-2123e4dbd1ae"
)

// ciaSeats is the active player and the next one, at the active
// player's main phase.
func CiaSeats(t *testing.T, g *game.Game) (*game.Player, *game.Player) {
	t.Helper()
	AdvanceToMain(t, g)
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
}

func CiaTapped(g *game.Game, id uuid.UUID) bool {
	tapped := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				tapped = c.Tapped
			}
		}
	})
	return tapped
}
