package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// devoid_test.go pins #2152 (CR 702.114a): devoid is a characteristic-
// defining ability that makes the card colourless in every zone, a
// layer-5 colour effect still applies on top of it, a copy keeps it,
// and colour identity (CR 903.4) does not move.

// devoidDrone is a devoid creature as the deck importer stamps one:
// Scryfall's empty `colors`, the coloured pips in the cost, and the
// keyword in Card.Keywords.
func devoidDrone(owner uuid.UUID) Card {
	c := NewCard("Fixture Drone", owner)
	c.TypeLine = "Creature — Eldrazi Drone"
	c.ManaCost = "{2}{U}"
	c.Colors = []string{}
	c.ColorIdentity = []string{"U"}
	c.Keywords = []string{KeywordDevoid, "flying"}
	c.Power, c.Toughness = 2, 1
	return c
}

// assertColourless fails unless every colour read the engine has
// answers "colourless" for c.
func assertColourless(t *testing.T, where string, c Card) {
	t.Helper()
	if got := c.EffectiveColors(); len(got) != 0 {
		t.Errorf("%s: EffectiveColors = %v, want none", where, got)
	}
	if got := c.Effective().Colors; len(got) != 0 {
		t.Errorf("%s: Effective().Colors = %v, want none", where, got)
	}
	if !c.IsColorless() {
		t.Errorf("%s: IsColorless = false", where)
	}
	if c.HasColor("U") {
		t.Errorf("%s: HasColor(U) = true for a devoid card", where)
	}
	if got := SourceCharacteristics(&c).Colors; len(got) != 0 {
		t.Errorf("%s: SourceCharacteristics colours = %v, want none (protection from blue would stop it)", where, got)
	}
}

// TestDevoidIsColourlessInEveryZone: CR 702.114a's "This ability
// functions everywhere". The same card, put in each zone in turn.
func TestDevoidIsColourlessInEveryZone(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	zones := map[string]*Zone{
		"hand": p.Hand, "library": p.Library, "graveyard": p.Graveyard,
		"command": p.Command, "exile": g.Exile, "stack": g.Stack,
	}
	for name, z := range zones {
		c := devoidDrone(p.ID)
		z.PushTop(c)
		got, ok := g.LookupCardForEffect(c.InstanceID)
		if !ok {
			t.Fatalf("%s: card not found", name)
		}
		assertColourless(t, name, got)
	}

	id := pushTypedTestCard(g, devoidDrone(p.ID))
	assertColourless(t, "battlefield", layeredBattlefieldCard(t, g, id))
}

// TestUnstampedCardWithoutDevoidKeepsTheCostFallback is the other side
// of the fix: a token or fixture with no stamped colours is still the
// colour of its cost. Devoid suppresses the fallback; nothing else
// does.
func TestUnstampedCardWithoutDevoidKeepsTheCostFallback(t *testing.T) {
	c := devoidDrone(uuid.New())
	c.Keywords = []string{"flying"}
	if got := c.EffectiveColors(); !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("EffectiveColors = %v, want [U] from the cost", got)
	}
	if got := c.Effective().Colors; !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("Effective().Colors = %v, want [U] from the cost", got)
	}
	if got := printedColors(c); !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("printedColors = %v, want [U]", got)
	}
}

// TestDevoidFromTheCatalog: a card that never went through the deck
// importer (a fixture, a restore point written before devoid was a
// canonical token) is devoid when its catalog entry says so. The two
// roads are the two inputs printedInputs merges.
func TestDevoidFromTheCatalog(t *testing.T) {
	stubPrintedKeywords(t, map[string][]string{"devoid-catalog-probe": {KeywordDevoid}})
	c := NewCard("Catalog Drone", uuid.New())
	c.OracleID = "devoid-catalog-probe"
	c.TypeLine = "Instant"
	c.ManaCost = "{1}{R}"
	assertColourless(t, "catalog-declared devoid", c)
	if got := printedColors(c); len(got) != 0 {
		t.Errorf("printedColors = %v, want none", got)
	}
	// The same card under a key whose entry declares nothing is red:
	// the answer comes from the catalog, not from the cost's shape.
	c.OracleID = "some-other-card"
	if got := c.EffectiveColors(); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("an uncatalogued card with no devoid keyword: EffectiveColors = %v, want [R]", got)
	}
}

// TestStampedColoursBeatDevoid: a non-empty Colors is an explicit value
// that already accounts for colour-defining abilities. The case that
// makes the order matter is CR 707.9d: a copy "except it's black"
// does not copy the colour-defining ability, so an eternalized devoid
// card is black, not colourless.
func TestStampedColoursBeatDevoid(t *testing.T) {
	c := devoidDrone(uuid.New())
	c.Colors = []string{"B"}
	if got := c.EffectiveColors(); !reflect.DeepEqual(got, []string{"B"}) {
		t.Errorf("EffectiveColors = %v, want the stamped [B]", got)
	}
	if got := c.Effective().Colors; !reflect.DeepEqual(got, []string{"B"}) {
		t.Errorf("Effective().Colors = %v, want the stamped [B]", got)
	}
}

const (
	devoidPainterOracle  = "devoid-test-painter"
	devoidSilencerOracle = "devoid-test-silencer"
)

// TestDevoidUnderLayerFiveAndSix: devoid is applied first in layer 5
// (CR 613.3), so a colour-setting effect still paints the permanent;
// and an effect that removes all abilities in layer 6 comes too late to
// make it coloured again (the Battle for Zendikar release notes: "If a
// card loses devoid, it will still be colorless").
func TestDevoidUnderLayerFiveAndSix(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	painted := map[uuid.UUID]bool{}
	silenced := map[uuid.UUID]bool{}
	withStaticAbilities(t, func(key string) []StaticAbility {
		switch key {
		case devoidPainterOracle:
			return []StaticAbility{{
				Layer:     Layer5Color,
				AppliesTo: func(target *Card, _ *Game, _ *Card) bool { return painted[target.InstanceID] },
				Apply:     func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.Colors = []string{"R"} },
			}}
		case devoidSilencerOracle:
			return []StaticAbility{{
				Layer:     Layer6Ability,
				AppliesTo: func(target *Card, _ *Game, _ *Card) bool { return silenced[target.InstanceID] },
				Apply:     func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.Abilities = nil },
			}}
		}
		return nil
	})
	paintedID := pushTypedTestCard(g, devoidDrone(p.ID))
	silencedID := pushTypedTestCard(g, devoidDrone(p.ID))
	painted[paintedID] = true
	silenced[silencedID] = true
	pushTypedTestCard(g, Card{Name: "Painter", TypeLine: "Enchantment", OracleID: devoidPainterOracle, Owner: p.ID, Controller: p.ID})
	pushTypedTestCard(g, Card{Name: "Silencer", TypeLine: "Enchantment", OracleID: devoidSilencerOracle, Owner: p.ID, Controller: p.ID})

	if got := layeredBattlefieldCard(t, g, paintedID).EffectiveColors(); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("a layer-5 colour effect on a devoid permanent: colours = %v, want [R]", got)
	}
	quiet := layeredBattlefieldCard(t, g, silencedID)
	if HasKeyword(&quiet, KeywordDevoid) {
		t.Error("the layer-6 effect should have removed devoid from the ability list")
	}
	assertColourless(t, "devoid permanent that lost all abilities", quiet)
}

// TestCopyOfADevoidCardStaysColourless: devoid rides the copiable
// values (CR 707.2) — Keywords, and the oracle ID for a catalogued card
// — so a Clone of an Eldrazi is colourless, and so is a copy of the
// spell. A copy "except it's black" is black (CR 707.9d).
func TestCopyOfADevoidCardStaysColourless(t *testing.T) {
	owner := uuid.New()
	src := devoidDrone(owner)

	clone := NewCard("Fixture Clone", owner)
	clone.TypeLine = "Creature — Shapeshifter"
	clone.ManaCost = "{3}{U}"
	clone.Colors = []string{"U"}
	clone.applyCopy(CopiableValuesOf(src), src)
	assertColourless(t, "a Clone copying a devoid creature", clone)

	spellCopy := NewCard("", owner)
	spellCopy.setPrintedValues(CopiableValuesOf(src))
	assertColourless(t, "a copy of a devoid spell", spellCopy)

	black := CopiableValuesOf(src)
	black.Colors = []string{"B"}
	zombie := NewCard("", owner)
	zombie.setPrintedValues(black)
	if got := zombie.EffectiveColors(); !reflect.DeepEqual(got, []string{"B"}) {
		t.Errorf("a copy \"except it's black\": colours = %v, want [B] (CR 707.9d)", got)
	}
}

// TestDevoidDoesNotChangeColourIdentity: CR 903.4 reads mana symbols,
// and devoid removes colours, not symbols. Both the imported shape
// (Scryfall's color_identity) and the fixture shape (identity derived
// from the cost) keep the pips' identity.
func TestDevoidDoesNotChangeColourIdentity(t *testing.T) {
	for _, tc := range []struct {
		name     string
		identity []string
		cost     string
		want     []string
	}{
		{"imported identity", []string{"U", "B"}, "{2}{U}{B}", []string{"U", "B"}},
		{"identity from the cost", nil, "{2}{U}{B}", []string{"U", "B"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			p := g.Seats[0]
			p.Command.Cards = nil
			cmdr := devoidDrone(p.ID)
			cmdr.Name = "Devoid Commander"
			cmdr.TypeLine = "Legendary Creature — Eldrazi"
			cmdr.IsCommander = true
			cmdr.ManaCost = tc.cost
			cmdr.ColorIdentity = tc.identity
			p.Command.PushTop(cmdr)
			if got := commanderIdentityFor(g, p).Colors; !reflect.DeepEqual(got, tc.want) {
				t.Errorf("identity = %v, want %v", got, tc.want)
			}
			in, _ := g.LookupCardForEffect(cmdr.InstanceID)
			assertColourless(t, "the commander itself", in)
		})
	}
}

// TestDevoidSurvivesSnapshotAndClone: the keyword is ordinary printed
// data in Card.Keywords, so a restore point and an undo clone carry it
// with no new field — and Scryfall's empty colour list, which
// `omitempty` drops on the wire, comes back as "not stamped", which is
// now the same answer for a devoid card.
func TestDevoidSurvivesSnapshotAndClone(t *testing.T) {
	g := newRestorableGame(t)
	p := g.Seats[0]
	inHand := devoidDrone(p.ID)
	p.Hand.PushTop(inHand)
	inYard := devoidDrone(p.ID)
	p.Graveyard.PushTop(inYard)
	onField := pushTypedTestCard(g, devoidDrone(p.ID))

	_, restored := roundTrip(t, g)
	clone := g.Clone()
	for name, gg := range map[string]*Game{"restored": restored, "clone": clone} {
		for _, id := range []uuid.UUID{inHand.InstanceID, inYard.InstanceID} {
			c, ok := gg.LookupCardForEffect(id)
			if !ok {
				t.Fatalf("%s: card %s missing", name, id)
			}
			if !containsKeyword(c.Keywords, KeywordDevoid) {
				t.Errorf("%s: devoid keyword lost: %v", name, c.Keywords)
			}
			assertColourless(t, name, c)
		}
		assertColourless(t, name+" battlefield", layeredBattlefieldCard(t, gg, onField))
	}
}

// TestDevoidIsACanonicalKeyword: the importer filters Scryfall's
// keyword array through CanonicalKeyword, so this is what stamps devoid
// onto the ~130 devoid cards that have no catalog entry.
func TestDevoidIsACanonicalKeyword(t *testing.T) {
	if kw, ok := CanonicalKeyword("Devoid"); !ok || kw != KeywordDevoid {
		t.Errorf("CanonicalKeyword(\"Devoid\") = (%q, %v), want (%q, true)", kw, ok, KeywordDevoid)
	}
	if NeedsCatalogEffect("Creature — Eldrazi Drone", "Devoid (This card has no color.)\nFlying") {
		t.Error("a devoid flier prints nothing the engine cannot run")
	}
}
