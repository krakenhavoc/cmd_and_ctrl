package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sqBoltInHand puts a Lightning Bolt in the active seat's hand at its
// main phase and returns it.
func SqBoltInHand(t *testing.T, g *game.Game) (me *game.Player, bolt uuid.UUID) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	AdvanceToMain(t, g)
	bolt = uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}",
			OracleID: LightningBoltOracle, Owner: me.ID, Controller: me.ID})
	})
	return me, bolt
}

// sqResolveAll passes until the stack is empty, keeping each copy's
// target when the copy asks.
func SqResolveAll(t *testing.T, g *game.Game, me *game.Player) (prompts int) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if StackFullyEmpty(g) && LatestPickTarget(g, me.ID) == nil {
			return prompts
		}
		if p := LatestPickTarget(g, me.ID); p != nil {
			prompts++
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: g.Seats[1].ID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
	return prompts
}
