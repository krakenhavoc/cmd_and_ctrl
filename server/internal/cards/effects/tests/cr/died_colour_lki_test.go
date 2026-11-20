package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pushCardsPaintedNonBlack seeds a printed-black creature whose own
// static paints it white, and checks the layer took.
func pushCardsPaintedNonBlack(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	c := game.Card{
		InstanceID: uuid.New(), Name: "Test Black Creature", TypeLine: "Creature — Zombie",
		OracleID: effectstest.PaintedNonBlackOracle, Colors: []string{"B"},
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	}
	id := effectstest.PushBattlefieldCardWithTimestamp(g, c)
	g.ReadSnapshot(func() {})
	if card, ok := effectstest.BattlefieldCard(g, id); !ok || card.HasColor("B") || !card.HasColor("W") {
		t.Fatal("setup: the static did not paint the creature white")
	}
	return id
}

// A creature black only through an effect (Darkest Hour, a
// black-making static) dying still triggers Teysa, though the card in
// the graveyard is the printed green creature underneath.
func TestTeysaTriggersOnACreatureBlackOnlyThroughAnEffect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Teysa, Orzhov Scion", "Legendary Creature — Human Advisor", effectstest.TeysaOrzhovScionOracle, 2, 3)
	beast := effectstest.PushCardsColorByEffect(t, g, me.ID)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(beast) })
	if gy, ok := g.LookupCardForEffect(beast); ok && gy.HasColor("B") {
		t.Fatal("setup: the card in the graveyard should not be black — the point of the test")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Spirit"); got != 1 {
		t.Errorf("Spirit count = %d, want 1 — a creature black only by effect should trigger Teysa", got)
	}
}

// A printed black creature an effect painted another colour does NOT
// trigger Teysa when it dies — weaker than printed is fine, stronger
// is not, and the card in the graveyard (printed black) must not be
// consulted over the stamp.
func TestTeysaDoesNotTriggerOnABlackCreaturePaintedNonBlackBeforeDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Teysa, Orzhov Scion", "Legendary Creature — Human Advisor", effectstest.TeysaOrzhovScionOracle, 2, 3)
	zombie := pushCardsPaintedNonBlack(t, g, me.ID)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(zombie) })
	if gy, ok := g.LookupCardForEffect(zombie); !ok || !gy.HasColor("B") {
		t.Fatal("setup: the card in the graveyard should read black (printed) — the point of the test")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Spirit"); got != 0 {
		t.Errorf("Spirit count = %d, want 0 — a creature painted non-black before it died must not trigger Teysa", got)
	}
}

// The condition itself, read directly: it answers from the stamp, not
// the graveyard card, exactly as the #1682 supertype and controller
// conditions do.
func TestTeysaConditionReadsTheStampedColourNotTheCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	appliesTo := ltbTriggerOf(t, effectstest.TeysaOrzhovScionOracle).AppliesTo

	bear := effectstest.B16Creature(g, me.ID, "Black Bear", "Creature — Bear", 2, 2, "B")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	ev := lastLTBOf(t, g, bear)
	if was, known := ev.WasColor("B"); !was || !known {
		t.Fatalf("setup: the stamp says %v, %v for a printed black creature", was, known)
	}
	source := &game.Card{InstanceID: uuid.New(), Controller: me.ID, Owner: me.ID}
	var none game.Characteristic
	if !appliesTo(ev, source, none, g) {
		t.Error("a printed black creature that died black did not count")
	}
	ev.LastKnownColors = nil
	if appliesTo(ev, source, none, g) {
		t.Error("the condition read the graveyard card's colour over a stamp that says it was not black")
	}
}
