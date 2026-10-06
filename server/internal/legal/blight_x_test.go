package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// blight_x_test.go — the enumerator half of #2174: "blight X" is an X
// announced by a non-mana cost, capped by the greatest toughness among
// the seat's creatures, paid onto one creature. Every move offered is
// one CastSpell accepts (#544), and none is offered at X=0.

const oracleSoulImmolation = "338747e8-bed8-4e60-8b19-5c2b80799477"

func immolationInHand(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	clearHand(p)
	cid := handCard(p, game.Card{
		Name: "Soul Immolation", TypeLine: "Sorcery", ManaCost: "{3}{R}{R}", OracleID: oracleSoulImmolation,
	})
	lands(g, p, "Mountain", "Mountain", 5)
	return cid
}

func TestBlightXIsAnnouncedAtTheGreatestToughnessAndPaidOntoACreature(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	card := immolationInHand(t, g, active)
	goblin := battlefieldCard(g, active, creature("Goblin", "{R}", 1, 1))
	battlefieldCard(g, active, creature("Ogre", "{2}{R}", 3, 3))
	battlefieldCard(g, opp, creature("Wall", "{1}", 0, 10))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, card)
	if len(casts) != 1 {
		t.Fatalf("%d casts, want 1: %v", len(casts), labels(casts))
	}
	p := teamworkBlightParams(t, casts[0])
	if x := xValueOf(t, casts[0]); x < 1 || x > 3 {
		t.Errorf("offered X=%d, want 1..3 (greatest toughness 3; 0 does nothing)", x)
	}
	if len(p.BlightIDs) != 1 || p.BlightIDs[0] != goblin.String() {
		t.Errorf("blights %v, want the Goblin (least power)", p.BlightIDs)
	}
	dispatchAll(t, g, active.ID, moves)
}

func TestBlightXIsNotOfferedWithNoCreature(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	card := immolationInHand(t, g, active)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	if casts := castMovesFor(moves, card); len(casts) != 0 {
		t.Errorf("offered %d casts with no creature to blight: %v", len(casts), labels(casts))
	}
	dispatchAll(t, g, active.ID, moves)
}

func TestBlightXCeilingAgreesWithTheEngine(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	card := immolationInHand(t, g, active)
	battlefieldCard(g, active, creature("Wall", "{1}", 0, 2))
	advanceTo(t, g, game.StepPrecombatMain)

	var ceiling int
	g.WithWriteLock(func() { ceiling = g.BlightXCeilingForEffect(active.ID) })
	casts := castMovesFor(legal.EnumerateFor(g, active.ID), card)
	if len(casts) != 1 {
		t.Fatalf("%d casts, want 1", len(casts))
	}
	if x := xValueOf(t, casts[0]); x != ceiling {
		t.Errorf("offered X=%d, engine ceiling %d", x, ceiling)
	}
}
