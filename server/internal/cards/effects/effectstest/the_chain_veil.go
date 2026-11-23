package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// chandraPlusOne activates Chandra's "+1: Add {R}{R}." and lets it
// resolve.
func ChandraPlusOne(t *testing.T, g *game.Game, p *game.Player, c uuid.UUID) {
	t.Helper()
	B16Activate(t, g, p.ID, c, GuffbChandraAddRR, game.ActivateAbilityParams{})
	PassPriorityAroundTable(t, g)
}

// loyaltyRowsOffered counts the enumerator's loyalty-row moves for
// `walker`.
func LoyaltyRowsOffered(g *game.Game, p *game.Player, walker uuid.UUID) int {
	n := 0
	for _, m := range legal.EnumerateFor(g, p.ID) {
		if m.Type == "activate_ability" && m.Source == walker {
			n++
		}
	}
	return n
}

// walkerGreyed is the wire's loyalty_activated flag for `walker`, as
// p's client sees it.
func WalkerGreyed(t *testing.T, g *game.Game, p *game.Player, walker uuid.UUID) bool {
	t.Helper()
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.KnownBy == nil {
				c.KnownBy = map[uuid.UUID]bool{}
			}
			for _, s := range g.Seats {
				c.KnownBy[s.ID] = true
			}
		}
	})
	v := protocol.ViewOfGameFor(g, p.ID.String())
	for _, cv := range v.Battlefield.Cards {
		if cv.InstanceID == walker.String() {
			return cv.LoyaltyActivated
		}
	}
	t.Fatalf("%s is not in the view", walker)
	return false
}
