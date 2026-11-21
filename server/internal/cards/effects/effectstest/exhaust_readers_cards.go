package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exhaust_readers_cards_test.go — #1184, the catalog half of the four
// cards that read the exhaust record from OUTSIDE the ability that
// owns it. The engine half (the event, the permission, the priced
// activation) is in game/exhaust_readers_test.go.
//
//	Rangers' Refueler    a trigger over YOUR activation
//	Afterburner Expert   the same trigger, from the GRAVEYARD
//	Elvish Refueler      the permission that suspends the gate
//	Boom Scholar         the cost modifier that can see the keyword

const (
	RangersRefuelerOracle   = "8e7ac64d-42c3-4e3b-9a20-9f09ee9d8a7a"
	AfterburnerExpertOracle = "b4d19e3a-c2f8-4503-bcfa-1b46bde972e5"
	ElvishRefuelerOracle    = "4b62f9fe-e3f4-4a21-aa6b-4dbd411d0c44"
	BoomScholarOracle       = "48296cc0-0141-47c6-9ec8-ada6171fee6f"
)

// handSizeOf is the seat's hand count, for the draw payoff.
func HandSizeOf(g *game.Game, p *game.Player) int { return len(p.Hand.Cards) }

// hasEffectiveType reports whether the permanent's POST-LAYER type
// line carries `want` — the only reading that can see an animation.
func HasEffectiveType(t *testing.T, g *game.Game, cardID uuid.UUID, want string) bool {
	t.Helper()
	for _, ty := range EffectiveTypes(t, g, cardID) {
		if ty == want {
			return true
		}
	}
	return false
}
