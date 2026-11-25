package game

import (
	"testing"

	"github.com/google/uuid"
)

// full_text_test.go — #6233, CR 612.6: "One card (Volrath's
// Shapeshifter) states that an object has the 'full text' of another
// object. This changes not just the text that appears in the object's
// text box and type line, but also changes the text that represents its
// name, mana cost, color indicator, power, and toughness."

const (
	fullTextOracle = "oracle-Volrath's Shapeshifter"
	fullTextKeep   = "probe/discard-a-card"
)

// stubFullTextCatalog wires a catalog in which the Shapeshifter prints
// the static and the kept "{2}: Discard a card." is a grant.
func stubFullTextCatalog(t *testing.T) {
	t.Helper()
	prev := CatalogLookup
	t.Cleanup(func() { CatalogLookup = prev })
	discard := ActivatedAbilityShape{Label: "{2}: Discard a card."}
	defs := map[string]*CardDef{
		fullTextOracle: {
			FullTextOfTopGraveyardCreature: &FullTextStatic{Keep: fullTextKeep},
			Activated:                      []ActivatedAbilityShape{discard},
		},
		"oracle-Grizzly Bears": {},
		GrantKey(fullTextKeep): {Activated: []ActivatedAbilityShape{discard}},
	}
	CatalogLookup = func(key string) *CardDef { return defs[key] }
}

// toGraveyardTop puts a fresh card on top of the player's graveyard the
// way a discard does, event and all.
func toGraveyardTop(g *Game, owner uuid.UUID, name, typeLine string, p, tough int) uuid.UUID {
	id := uuid.New()
	seat := g.PlayerByIDForEffect(owner)
	seat.Graveyard.PushTop(Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		OracleID: "oracle-" + name, ScryfallID: "scry-" + name,
		ManaCost: "{1}{G}", Colors: []string{"G"},
		Power: p, Toughness: tough, PrintedPTKnown: true,
		Owner: owner, Controller: owner,
	})
	g.mu.Lock()
	g.EmitEvent(Event{Kind: EventZoneMove, CardID: id, OldZone: ZoneHand, NewZone: ZoneGraveyard})
	g.mu.Unlock()
	return id
}

func shapeshifterOnBattlefield(g *Game, owner uuid.UUID) uuid.UUID {
	return vanillaOnBattlefield(g, owner, "Volrath's Shapeshifter", "Creature — Shapeshifter", 0, 1)
}

func keepsDiscard(t *testing.T, c Card) bool {
	t.Helper()
	d := catalogDef(CatalogKey(c))
	if d == nil {
		return false
	}
	for _, a := range d.Activated {
		if a.Label == "{2}: Discard a card." {
			return true
		}
	}
	return false
}

// TestFullTextOfTheTopCreatureCard is the issue's acceptance: with a
// Grizzly Bears on top of the graveyard the Shapeshifter is a 2/2 named
// Grizzly Bears that keeps "{2}: Discard a card." (CR 612.6), and when
// a non-creature card covers the Bears it is itself again (CR 404.2:
// only the top of the pile counts).
func TestFullTextOfTheTopCreatureCard(t *testing.T) {
	stubFullTextCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	ss := shapeshifterOnBattlefield(g, me.ID)

	if c := copyProbe(t, g, ss); c.Name != "Volrath's Shapeshifter" || c.FullTextBase != nil {
		t.Fatalf("an empty graveyard changed the Shapeshifter: %q", c.Name)
	}

	toGraveyardTop(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	c := copyProbe(t, g, ss)
	if c.Name != "Grizzly Bears" || c.OracleID != "oracle-Grizzly Bears" || c.ManaCost != "{1}{G}" {
		t.Fatalf("full text did not land: name %q oracle %q cost %q", c.Name, c.OracleID, c.ManaCost)
	}
	if e := c.Effective(); e.Power != 2 || e.Toughness != 2 || !c.IsCreature() {
		t.Errorf("P/T = %d/%d, want the Bears' 2/2", e.Power, e.Toughness)
	}
	if len(c.Colors) != 1 || c.Colors[0] != "G" {
		t.Errorf("colours = %v, want the Bears' green", c.Colors)
	}
	if !keepsDiscard(t, c) {
		t.Error("the Shapeshifter lost \"{2}: Discard a card.\"")
	}
	if c.IsCopy() {
		t.Error("a full-text effect is not a copy effect (CR 612.6, 707.2)")
	}

	toGraveyardTop(g, me.ID, "Shock", "Instant", 0, 0)
	c = copyProbe(t, g, ss)
	if c.Name != "Volrath's Shapeshifter" || c.OracleID != fullTextOracle || c.FullTextBase != nil {
		t.Fatalf("with a non-creature on top: name %q oracle %q base %v", c.Name, c.OracleID, c.FullTextBase)
	}
	if e := c.Effective(); e.Power != 0 || e.Toughness != 1 {
		t.Errorf("P/T = %d/%d, want its own 0/1", e.Power, e.Toughness)
	}
	if !keepsDiscard(t, c) {
		t.Error("the Shapeshifter's own \"{2}: Discard a card.\" is missing")
	}
}

// TestFullTextReadsOnlyItsControllersGraveyard — "the top card of YOUR
// graveyard".
func TestFullTextReadsOnlyItsControllersGraveyard(t *testing.T) {
	stubFullTextCatalog(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ss := shapeshifterOnBattlefield(g, me.ID)
	toGraveyardTop(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	if c := copyProbe(t, g, ss); c.Name != "Volrath's Shapeshifter" {
		t.Fatalf("an opponent's graveyard gave it full text: %q", c.Name)
	}
}

// TestFullTextIsNotCopiable — CR 707.2: a text-changing effect is not a
// copiable value, so the Shapeshifter's copiable values are its own.
func TestFullTextIsNotCopiable(t *testing.T) {
	stubFullTextCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	ss := shapeshifterOnBattlefield(g, me.ID)
	toGraveyardTop(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	c := copyProbe(t, g, ss)
	if v := CopiableValuesOf(c); v.Name != "Volrath's Shapeshifter" || v.OracleID != fullTextOracle || len(v.GrantedAbilities) != 0 {
		t.Errorf("copiable values = %q / %q / %v, want the Shapeshifter's own", v.Name, v.OracleID, v.GrantedAbilities)
	}
	if v := OwnPrintedValues(c); v.Name != "Volrath's Shapeshifter" {
		t.Errorf("own printed values = %q", v.Name)
	}
}

// TestFullTextEndsWhenThePermanentLeaves — CR 400.7: the card that
// leaves the battlefield is a Volrath's Shapeshifter, not a Bears.
func TestFullTextEndsWhenThePermanentLeaves(t *testing.T) {
	stubFullTextCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	ss := shapeshifterOnBattlefield(g, me.ID)
	toGraveyardTop(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	_ = copyProbe(t, g, ss)
	c := Card{}
	g.WithWriteLock(func() {
		r, _ := g.battlefieldCardLocked(ss)
		c = *r
	})
	c.restorePrintedSelf()
	if c.Name != "Volrath's Shapeshifter" || c.OracleID != fullTextOracle || c.FullTextBase != nil || len(c.GrantedAbilities) != 0 {
		t.Errorf("after leaving: %q / %q / base %v / grants %v", c.Name, c.OracleID, c.FullTextBase, c.GrantedAbilities)
	}
}

// TestFullTextIsStableAcrossRecomputes — the baseline is
// deep-copied, and a recompute that changes nothing rewrites nothing.
func TestFullTextIsStableAcrossRecomputes(t *testing.T) {
	stubFullTextCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	ss := shapeshifterOnBattlefield(g, me.ID)
	toGraveyardTop(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	_ = copyProbe(t, g, ss)
	before := g.layerVersion.Load()
	g.WithWriteLock(func() { g.materialiseDurationCopiesLocked() })
	if after := g.layerVersion.Load(); after != before {
		t.Errorf("an idempotent pass bumped the layer version %d -> %d", before, after)
	}
	var raw Card
	g.WithWriteLock(func() {
		r, _ := g.battlefieldCardLocked(ss)
		raw = *r
	})
	cl := cloneCard(raw)
	if cl.FullTextBase == raw.FullTextBase {
		t.Error("clone shares the baseline pointer with the live card")
	}
	if cl.FullTextBase == nil || cl.FullTextBase.Name != "Volrath's Shapeshifter" {
		t.Fatalf("clone lost the baseline: %+v", cl.FullTextBase)
	}
}
