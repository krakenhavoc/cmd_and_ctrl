package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// grist_the_hunger_tide_test.go — Grist, the Hunger Tide (#5427): the
// "isn't on the battlefield" sentence in every zone (Spec.OffBattlefield,
// game/off_battlefield_self.go), and the three loyalty abilities.

const (
	gristOracle   = "0efb0d7e-dea0-4817-a243-15066e9ef333"
	gristTypeLine = "Legendary Planeswalker — Grist"
)

// gristCard is a Grist card for `owner`, as the importer stamps it: no
// printed power or toughness.
func gristCard(owner uuid.UUID) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: "Grist, the Hunger Tide", TypeLine: gristTypeLine,
		ManaCost: "{1}{B}{G}", OracleID: gristOracle, Colors: []string{"B", "G"},
		Owner: owner, Controller: owner,
	}
}

// gristTable seeds Grist on seat 0's battlefield with `loyalty`, in a
// main phase, over an EMPTY library, so a test stacks the cards the +1
// mills.
func gristTable(t *testing.T, loyalty int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	me.Library.Cards = nil
	grist := pushBattlefieldCardWithTimestamp(g, gristCard(me.ID))
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(grist, game.CounterLoyalty, loyalty) })
	return g, me, grist
}

func gristLookup(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var c game.Card
	var ok bool
	g.WithWriteLock(func() { c, ok = g.LookupCardForEffect(id) })
	if !ok {
		t.Fatalf("card %s is nowhere", id)
	}
	return c
}

// requireGristCreatureCard asserts the off-battlefield reading: a 1/1
// Insect creature in addition to a legendary Grist planeswalker.
func requireGristCreatureCard(t *testing.T, where string, c game.Card) {
	t.Helper()
	if !c.IsCreature() || !c.IsPlaneswalker() {
		t.Errorf("%s: creature %v, planeswalker %v; want both", where, c.IsCreature(), c.IsPlaneswalker())
	}
	if !c.HasSubtype("Insect") || !c.HasSubtype("Grist") {
		t.Errorf("%s: Insect %v, Grist %v; want both", where, c.HasSubtype("Insect"), c.HasSubtype("Grist"))
	}
	if !c.HasSupertype("legendary") {
		t.Errorf("%s: not legendary", where)
	}
	if p, tough := c.CurrentPower(), c.CurrentToughness(); p != 1 || tough != 1 {
		t.Errorf("%s: %d/%d, want 1/1", where, p, tough)
	}
	if c.HasSubtype("Elf") {
		t.Errorf("%s: an Elf with no effect making it one", where)
	}
	eff := c.Effective()
	if len(eff.Types) != 2 || len(eff.Subtypes) != 2 {
		t.Errorf("%s: types %v subtypes %v, want Planeswalker Creature — Grist Insect", where, eff.Types, eff.Subtypes)
	}
}

func TestGristIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(gristOracle)
	if !ok {
		t.Fatal("Grist, the Hunger Tide is not registered")
	}
	if spec.Name != "Grist, the Hunger Tide" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// CR 113.6c: in every zone but the battlefield Grist is a 1/1 Insect
// creature card as well as a planeswalker card.
func TestGristIsACreatureCardInEveryZoneOffTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	zones := map[string]*game.Zone{
		"hand": me.Hand, "library": me.Library, "graveyard": me.Graveyard,
		"command zone": me.Command, "exile": g.Exile,
	}
	for where, z := range zones {
		c := gristCard(me.ID)
		z.PushTop(c)
		requireGristCreatureCard(t, where, gristLookup(t, g, c.InstanceID))
	}
}

// On the battlefield the sentence is off: a planeswalker with loyalty and
// no power, toughness or creature type, before and after the layer pass.
func TestGristIsOnlyAPlaneswalkerOnTheBattlefield(t *testing.T) {
	g, _, grist := gristTable(t, 3)
	for _, c := range []game.Card{gristLookup(t, g, grist), layeredCard(t, g, grist)} {
		if c.IsCreature() || c.HasSubtype("Insect") {
			t.Errorf("on the battlefield: creature %v, Insect %v; want neither", c.IsCreature(), c.HasSubtype("Insect"))
		}
		if !c.IsPlaneswalker() || !c.HasSubtype("Grist") {
			t.Error("on the battlefield Grist is a Grist planeswalker")
		}
	}
}

// Cast from hand it is a creature spell: "whenever you cast a creature
// spell" sees it. Then it resolves as a planeswalker with 3 loyalty, and "whenever
// another creature enters" does not see a creature enter.
func TestGristIsACreatureSpellAndEntersAsAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Beast Whisperer", "Creature — Elf Druid", beastWhispererOracle, false)
	pushCatalogPermanent(g, me.ID, "Soul Warden", "Creature — Human Cleric", b04SoulWardenOracle, false)
	advanceToMain(t, g)
	c := gristCard(me.ID)
	me.Hand.PushTop(c)
	hand, life := me.Hand.Size(), me.Life

	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Grist: %v", err)
	}
	requireGristCreatureCard(t, "stack", gristLookup(t, g, c.InstanceID))
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand-1+1 {
		t.Errorf("hand %d → %d: Beast Whisperer draws for a creature spell", hand, got)
	}
	perm, ok := battlefieldCard(g, c.InstanceID)
	if !ok {
		t.Fatal("Grist did not resolve onto the battlefield")
	}
	if perm.IsCreature() || !perm.IsPlaneswalker() {
		t.Errorf("resolved: creature %v, planeswalker %v", perm.IsCreature(), perm.IsPlaneswalker())
	}
	if got := countersOn(g, c.InstanceID, game.CounterLoyalty); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
	if me.Life != life {
		t.Errorf("life %d → %d: Soul Warden saw a creature enter, and Grist is not one on the battlefield", life, me.Life)
	}
}

// "Counter target creature spell" may target Grist on the stack, and a
// noncreature planeswalker spell is still refused.
func TestGristOnTheStackIsCounteredAsACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	other := game.Card{
		InstanceID: uuid.New(), Name: "Plain Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: me.ID, Controller: me.ID,
	}
	me.Hand.PushTop(other)
	if err := g.CastSpell(me.ID, other.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the plain planeswalker: %v", err)
	}
	scatter := handCardFull(opp, "Essence Scatter", "Instant", "{1}{U}", b25EssenceScatterOracle, nil)
	if err := g.CastSpell(opp.ID, scatter, game.CastSpellParams{Targets: b16TargetCard(other.InstanceID)}); err == nil {
		t.Fatal("a planeswalker spell that is not a creature spell is refused as a target")
	}
	passPriorityAroundTable(t, g)

	c := gristCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Grist: %v", err)
	}
	if err := g.CastSpell(opp.ID, scatter, game.CastSpellParams{Targets: b16TargetCard(c.InstanceID)}); err != nil {
		t.Fatalf("Essence Scatter on Grist: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(c.InstanceID) || !me.Graveyard.Contains(c.InstanceID) {
		t.Error("Grist was not countered")
	}
	requireGristCreatureCard(t, "graveyard after the counter", gristLookup(t, g, c.InstanceID))
}
