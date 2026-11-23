package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- shared helpers ------------------------------------------------------------

// mkmB1CastFaceDown casts a catalog card from the active player's hand
// face down for its disguise and resolves it.
func MkmB1CastFaceDown(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{AlternativeCost: "disguise"}); err != nil {
		t.Fatalf("cast %s face down: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
	if c := FindCardAnywhere(t, g, id); !c.FaceDownIsPermanent() {
		t.Fatalf("%s did not enter face down", name)
	}
	return id
}

// mkmB1TurnFaceUp pays the disguise cost from conjured mana and turns
// the permanent face up.
func MkmB1TurnFaceUp(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, mana ...string) {
	t.Helper()
	g.WithWriteLock(func() { p.ManaPool.AddMana(ManaTokens(mana...)...) })
	if err := g.PerformSpecialAction(p.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	if FindCardAnywhere(t, g, id).FaceDown {
		t.Fatal("still face down after the special action")
	}
}

// mkmB1PT is a battlefield permanent's power and toughness after the
// layers and its counters.
func MkmB1PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	c := TmtB1Card(t, g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

// mkmB1MustBlock reports whether a battlefield creature carries a
// "blocks this turn if able" requirement.
func MkmB1MustBlock(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := FindBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	for _, r := range c.Effective().BlockRequirements {
		if r.Kind == game.BlockRequirementBlocks {
			return true
		}
	}
	return false
}

// mkmB1Solved reads a Case's solved designation.
func MkmB1Solved(g *game.Game, id uuid.UUID) bool { return IsSolved(g, id) }
