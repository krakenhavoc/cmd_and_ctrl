package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- fixtures ------------------------------------------------------------------

// avB1Hand puts a card in p's hand and returns its id.
func AvB1Hand(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Hand.PushTop(c)
	return c.InstanceID
}

// avB1Cast casts a card from p's hand from a main phase, failing the test on
// a refusal.
func AvB1Cast(t *testing.T, g *game.Game, p *game.Player, c game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	AdvanceToMain(t, g)
	id := AvB1Hand(p, c)
	if err := g.CastSpell(p.ID, id, params); err != nil {
		t.Fatalf("casting %s: %v", c.Name, err)
	}
	return id
}

// avB1Plains seeds n Plains for owner.
func AvB1Plains(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		PushLandFor(g, owner, "Plains", "Basic Land — Plains")
	}
}

// avB1Card reads a card through the layers.
func AvB1Card(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var c game.Card
	var ok bool
	g.ReadSnapshot(func() { c, ok = g.LookupCardForEffect(id) })
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c
}

// avB1Named returns the battlefield permanent controlled by owner with the
// given name, or uuid.Nil.
func AvB1Named(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == owner {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

// avB1ResolveCreature casts a creature from p's hand and resolves it, its
// entry triggers left for the caller.
func AvB1ResolveCreature(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := AvB1Cast(t, g, p, game.Card{Name: name, TypeLine: typeLine, OracleID: oracle, Power: 2, Toughness: 2}, game.CastSpellParams{})
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if !g.Battlefield.Contains(id) {
		t.Fatalf("%s never reached the battlefield", name)
	}
	return id
}

// avB1OpponentCreatureEnters has a creature token enter under owner's control
// outside their turn, so its entry is an event the source can see.
func AvB1OpponentCreatureEnters(t *testing.T, g *game.Game, owner uuid.UUID, typeLine string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(owner, game.Card{Name: "Their Token", TypeLine: "Token " + typeLine, Power: 1, Toughness: 1}, 1); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	PassPriorityAroundTable(t, g)
}
