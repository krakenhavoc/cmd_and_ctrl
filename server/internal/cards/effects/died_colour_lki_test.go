package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// died_colour_lki_test.go — #1689, CR 603.10a: a dies /
// leaves-the-battlefield condition is judged on the permanent's
// COLOUR as it last existed on the battlefield, not the card's colour
// in its new zone. Reader: leftAsColor. Teysa, Orzhov Scion's
// "whenever another black creature you control dies" is the only
// catalog condition that tested a departed permanent's colour.

// colorByEffectOracle is a test oracle whose static paints the
// permanent black — a layer-5 colour change (CR 105.3), the shape
// Darkest Hour gives every creature its controller controls.
const colorByEffectOracle = "test-1689-cards-color-by-effect"

// paintedNonBlackOracle is a test oracle whose static paints the
// permanent white — a printed black creature an effect has made
// another colour, the reverse direction #1689 also has to get right.
const paintedNonBlackOracle = "test-1689-cards-painted-non-black"

func init() {
	paint := func(color string) func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
		return func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Colors = []string{color}
		}
	}
	self := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID
	}
	Register(Spec{
		OracleID: colorByEffectOracle,
		Name:     "Test Green Creature Painted Black",
		Static: []game.StaticAbility{{
			Layer: game.Layer5Color, AppliesTo: self, Apply: paint("B"),
		}},
	})
	Register(Spec{
		OracleID: paintedNonBlackOracle,
		Name:     "Test Black Creature Painted White",
		Static: []game.StaticAbility{{
			Layer: game.Layer5Color, AppliesTo: self, Apply: paint("W"),
		}},
	})
}

// pushCardsColorByEffect seeds a printed-green creature whose own
// static paints it black, and checks the layer took.
func pushCardsColorByEffect(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	c := game.Card{
		InstanceID: uuid.New(), Name: "Test Green Beast", TypeLine: "Creature — Beast",
		OracleID: colorByEffectOracle, Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	}
	id := pushBattlefieldCardWithTimestamp(g, c)
	g.ReadSnapshot(func() {})
	if card, ok := battlefieldCard(g, id); !ok || !card.HasColor("B") {
		t.Fatal("setup: the static did not paint the creature black")
	}
	return id
}

// pushCardsPaintedNonBlack seeds a printed-black creature whose own
// static paints it white, and checks the layer took.
func pushCardsPaintedNonBlack(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	c := game.Card{
		InstanceID: uuid.New(), Name: "Test Black Creature", TypeLine: "Creature — Zombie",
		OracleID: paintedNonBlackOracle, Colors: []string{"B"},
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	}
	id := pushBattlefieldCardWithTimestamp(g, c)
	g.ReadSnapshot(func() {})
	if card, ok := battlefieldCard(g, id); !ok || card.HasColor("B") || !card.HasColor("W") {
		t.Fatal("setup: the static did not paint the creature white")
	}
	return id
}

// A creature black only through an effect (Darkest Hour, a
// black-making static) dying still triggers Teysa, though the card in
// the graveyard is the printed green creature underneath.
func TestTeysaTriggersOnACreatureBlackOnlyThroughAnEffect(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Teysa, Orzhov Scion", "Legendary Creature — Human Advisor", teysaOrzhovScionOracle, 2, 3)
	beast := pushCardsColorByEffect(t, g, me.ID)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(beast) })
	if gy, ok := g.LookupCardForEffect(beast); ok && gy.HasColor("B") {
		t.Fatal("setup: the card in the graveyard should not be black — the point of the test")
	}
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Spirit"); got != 1 {
		t.Errorf("Spirit count = %d, want 1 — a creature black only by effect should trigger Teysa", got)
	}
}

// A printed black creature an effect painted another colour does NOT
// trigger Teysa when it dies — weaker than printed is fine, stronger
// is not, and the card in the graveyard (printed black) must not be
// consulted over the stamp.
func TestTeysaDoesNotTriggerOnABlackCreaturePaintedNonBlackBeforeDying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Teysa, Orzhov Scion", "Legendary Creature — Human Advisor", teysaOrzhovScionOracle, 2, 3)
	zombie := pushCardsPaintedNonBlack(t, g, me.ID)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(zombie) })
	if gy, ok := g.LookupCardForEffect(zombie); !ok || !gy.HasColor("B") {
		t.Fatal("setup: the card in the graveyard should read black (printed) — the point of the test")
	}
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Spirit"); got != 0 {
		t.Errorf("Spirit count = %d, want 0 — a creature painted non-black before it died must not trigger Teysa", got)
	}
}

// The condition itself, read directly: it answers from the stamp, not
// the graveyard card, exactly as the #1682 supertype and controller
// conditions do.
func TestTeysaConditionReadsTheStampedColourNotTheCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	appliesTo := ltbTriggerOf(t, teysaOrzhovScionOracle).AppliesTo

	bear := b16Creature(g, me.ID, "Black Bear", "Creature — Bear", 2, 2, "B")
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
