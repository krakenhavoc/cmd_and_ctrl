package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// transformers_helpers_test.go — what the More Than Meets the Eye card
// tests (#3195, ADR 0152) share. Each card has its own _test.go; the
// engine itself is pinned in goldbug_humanitys_ally_test.go.

// transformerRow is a Transformers card's Scryfall record, down the real
// import road: front face and back face with the text the card prints.
func transformerRow(oracleID string, front, back disturbFace, frontText, backText string, keywords ...string) cards.Card {
	row := disturbRow(oracleID, nil, front, back)
	row.Keywords = append([]string{"More Than Meets the Eye"}, keywords...)
	row.CardFaces[0].OracleText = frontText
	row.CardFaces[1].OracleText = backText
	return row
}

// transformerIn puts the card on the battlefield, front face up and past
// summoning sickness, and returns its instance ID.
func transformerIn(t *testing.T, g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	t.Helper()
	id := importToBattlefield(t, g, row, p)
	unsick(g, id)
	return id
}

// transformerConverted is transformerIn, then converted once (back face
// up, TurnOvers 1).
func transformerConverted(t *testing.T, g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	t.Helper()
	id := transformerIn(t, g, row, p)
	g.WithWriteLock(func() {
		if err := g.ConvertPermanentForEffect(id); err != nil {
			t.Fatalf("convert: %v", err)
		}
		g.RecomputeLayersIfStaleLocked()
	})
	return id
}

// wantFace asserts which face a permanent shows and how many times it
// has turned over.
func wantFace(t *testing.T, g *game.Game, id uuid.UUID, face int, name string) game.Card {
	t.Helper()
	c := flipState(t, g, id)
	if c.ActiveFace != face || c.Name != name {
		t.Fatalf("shows face %d (%q), want face %d (%q)", c.ActiveFace, c.Name, face, name)
	}
	return c
}

// emitCombatDamage fakes one creature's combat damage to a player.
func emitCombatDamage(g *game.Game, source, victim uuid.UUID, amount int) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: source, Target: victim, Amount: amount, Combat: true})
	})
}

// tfSettle passes priority round the table until the stack is empty.
func tfSettle(t *testing.T, g *game.Game) {
	t.Helper()
	passPriorityAroundTable(t, g)
}

// transformerAttacks declares the card (and any others) against the
// defender and lets the attack trigger go on the stack.
func transformerAttacks(t *testing.T, g *game.Game, id, defender uuid.UUID, others ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, defender); err != nil {
		t.Fatalf("the card can't attack: %v", err)
	}
	for _, o := range others {
		if err := g.DeclareAttacker(o, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	tfSettle(t, g)
}
