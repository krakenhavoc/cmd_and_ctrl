package effectstest

import (
	"testing"

	"github.com/google/uuid"

	cardsj "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/j"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// great_synthesis_hand_size_test.go — The Great Synthesis chapter I's "you
// have no maximum hand size for as long as you control this Saga" as an
// effect the resolving ability CREATES (ADR 0113's amendment of 2026-10-13,
// #3247), not a static of the Saga.

// greatSynthesisToChapterOne flips Jin-Gitaxias with seven cards in hand and
// stops with the Saga on the battlefield and its chapter I trigger on the
// stack. Returns the Saga's instance ID.
func GreatSynthesisToChapterOne(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	jin := PushBattlefieldCardWithTimestamp(g, JinCard(me.ID, false))
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	PraetorFillHand(me, 7)
	ToMainForCost(t, g)
	FloatForTest(g, me, "UUUU")
	if err := g.ActivateCatalogAbility(me.ID, jin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("transform: %v", err)
	}
	for i := 0; i < 12; i++ {
		if saga := PraetorOnBattlefield(g, cardsj.JinGitaxiasOracleID); saga != nil && saga.Name == "The Great Synthesis" &&
			TriggerOnStack(g, saga.InstanceID) != nil {
			return saga.InstanceID
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("chapter I never reached the stack")
	return uuid.Nil
}
