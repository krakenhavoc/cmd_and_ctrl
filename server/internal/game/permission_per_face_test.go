package game

import (
	"testing"

	"github.com/google/uuid"
)

// permission_per_face_test.go — #4696: two live cast permissions that
// open different faces of one multi-face card. A cast is legal when any
// rule or effect allows it (CR 601.3), and each face is judged as the spell
// it would be (CR 712.11c, 715.3a), so each face is offered under, cast
// through and priced by the permission that opens it, whichever of the two
// the lookup finds first.

// giantStompFixture is the Bonecrusher Giant shape: a {2}{R} creature
// whose Adventure is a {1}{R} instant.
func giantStompFixture(owner uuid.UUID) Card {
	c := Card{
		InstanceID: uuid.New(),
		OracleID:   "cccccccc-dddd-eeee-ffff-000000004696",
		Owner:      owner,
		Controller: owner,
		Layout:     LayoutAdventure,
		Faces: []Face{
			{Name: "Fixture Giant", TypeLine: "Creature — Giant", ManaCost: "{2}{R}", Colors: []string{"R"},
				Power: 4, Toughness: 3},
			{Name: "Fixture Stomp", TypeLine: "Instant — Adventure", ManaCost: "{1}{R}", Colors: []string{"R"}},
		},
	}
	c.SetFace(0)
	return c
}

// grantStandingFilterCostLess is grantStandingFilter with "spells you cast
// this way cost {n} less", so a price shows which permission it came from.
func grantStandingFilterCostLess(g *Game, p *Player, zone ZoneKind, f PermissionFilter, label string, n int) {
	g.WithWriteLock(func() {
		g.GrantCastPermissionForEffect(CastPermission{
			Player: p.ID, Zone: zone, Scope: ScopeStanding, Filter: f, Label: label, CostLess: n,
		})
	})
}

// faceGrantsOf is CastableFaceGrantsLocked for the card, as the
// enumerator and the view ask it: face -> the permission's label.
func faceGrantsOf(g *Game, p *Player, id uuid.UUID, zone ZoneKind) map[int]string {
	out := map[int]string{}
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return
		}
		first := g.CastPermissionForLocked(p.ID, c, zone)
		for _, fg := range g.CastableFaceGrantsLocked(p.ID, c, zone, first) {
			label := ""
			if fg.Permission != nil {
				label = fg.Permission.Label
			}
			out[fg.Face] = label
		}
	})
	return out
}

// priceOfFace is the total mana value a cast of `face` is charged.
func priceOfFace(t *testing.T, g *Game, p *Player, id uuid.UUID, face int) int {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatal("card not found")
	}
	price, err := g.PriceCast(p.ID, c, CastSpellParams{FromZone: "graveyard", Face: face})
	if err != nil {
		t.Fatalf("PriceCast face %d: %v", face, err)
	}
	return price.Total.ManaValue()
}

func TestTwoPermissionsOpenDifferentFacesOfOneCard(t *testing.T) {
	for _, creatureFirst := range []bool{true, false} {
		name := "creature permission first"
		if !creatureFirst {
			name = "instant-and-sorcery permission first"
		}
		t.Run(name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			toMainPhase(t, g)
			creature := func() {
				grantStandingFilterCostLess(g, me, ZoneGraveyard, PermissionFilter{CreatureOnly: true}, "creature spells", 0)
			}
			spells := func() {
				grantStandingFilterCostLess(g, me, ZoneGraveyard, PermissionFilter{InstantOrSorceryOnly: true}, "instant and sorcery spells", 1)
			}
			if creatureFirst {
				creature()
				spells()
			} else {
				spells()
				creature()
			}
			id := seedInGraveyard(g, me, giantStompFixture(me.ID))

			got := faceGrantsOf(g, me, id, ZoneGraveyard)
			if got[0] != "creature spells" || got[1] != "instant and sorcery spells" || len(got) != 2 {
				t.Fatalf("faces offered = %v, want the Giant under the creature permission and Stomp under the other", got)
			}
			// Priced by the permission that opens each face: only the
			// instant-and-sorcery one takes {1} off.
			if mv := priceOfFace(t, g, me, id, 0); mv != 3 {
				t.Errorf("the Giant is priced at mana value %d, want 3 (the creature permission takes nothing off)", mv)
			}
			if mv := priceOfFace(t, g, me, id, 1); mv != 1 {
				t.Errorf("Stomp is priced at mana value %d, want 1 ({1}{R} less {1})", mv)
			}

			// Stomp, through the instant-and-sorcery permission, for {R}.
			g.WithWriteLock(func() { floatMana(me, "R") })
			if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 1, Strict: true}); err != nil {
				t.Fatalf("cast Stomp from the graveyard: %v", err)
			}
			spell := findCardForTest(g.Stack, id)
			if spell == nil || spell.ActiveFace != 1 || !spell.IsInstant() || spell.IsCreature() {
				t.Fatalf("on the stack: %+v, want the Adventure", spell)
			}
			AssertFaceInvariant(t, g)
		})
	}
}

// The other half through the other permission: the Giant itself, cast
// through the creature permission when the instant-and-sorcery one is the
// first found.
func TestTwoPermissionsCastTheCreatureThroughItsOwn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantStandingFilterCostLess(g, me, ZoneGraveyard, PermissionFilter{InstantOrSorceryOnly: true}, "instant and sorcery spells", 1)
	grantStandingFilterCostLess(g, me, ZoneGraveyard, PermissionFilter{CreatureOnly: true}, "creature spells", 0)
	id := seedInGraveyard(g, me, giantStompFixture(me.ID))

	g.WithWriteLock(func() { floatMana(me, "RR") })
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 0, Strict: true}); err == nil {
		t.Fatal("the Giant was cast for {R}{R}: priced by the instant-and-sorcery permission's reduction")
	}
	g.WithWriteLock(func() { floatMana(me, "R") })
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 0, Strict: true}); err != nil {
		t.Fatalf("cast the Giant from the graveyard: %v", err)
	}
	if spell := findCardForTest(g.Stack, id); spell == nil || spell.ActiveFace != 0 || !spell.IsCreature() {
		t.Fatalf("on the stack: %+v, want the creature", spell)
	}
}

// A permission that NAMES a face (CR 715.4's "cast the creature") beside
// one that opens the card's other face: both faces, each through its own.
func TestNamedFacePermissionBesideAnotherOpensBothFaces(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	card := giantStompFixture(me.ID)
	id := seedInGraveyard(g, me, card)
	g.WithWriteLock(func() {
		g.GrantCastPermissionToCardsForEffect(CastPermission{
			Player: me.ID, Zone: ZoneGraveyard, Faces: []int{0}, Label: "the creature",
			Duration: g.UntilEndOfTurnDuration(),
		}, []Card{card})
	})
	grantStandingFilterCostLess(g, me, ZoneGraveyard, PermissionFilter{InstantOrSorceryOnly: true}, "instant and sorcery spells", 0)

	got := faceGrantsOf(g, me, id, ZoneGraveyard)
	if got[0] != "the creature" || got[1] != "instant and sorcery spells" {
		t.Fatalf("faces offered = %v, want the creature under its named grant and Stomp under the other", got)
	}
	g.WithWriteLock(func() { floatMana(me, "RR") })
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 1, Strict: true}); err != nil {
		t.Fatalf("cast Stomp beside a grant naming the creature: %v", err)
	}
	if spell := findCardForTest(g.Stack, id); spell == nil || spell.ActiveFace != 1 {
		t.Fatalf("on the stack: %+v, want the Adventure", spell)
	}
}

// One permission alone is unchanged: the face it does not open is still
// refused, and a grant naming one face still settles a cast that names
// none.
func TestOnePermissionStillOpensOnlyItsFaces(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantStandingFilterCostLess(g, me, ZoneGraveyard, PermissionFilter{CreatureOnly: true}, "creature spells", 0)
	id := seedInGraveyard(g, me, giantStompFixture(me.ID))
	if got := faceGrantsOf(g, me, id, ZoneGraveyard); len(got) != 1 || got[0] != "creature spells" {
		t.Fatalf("faces offered = %v, want the creature alone", got)
	}
	g.WithWriteLock(func() { floatMana(me, "RR") })
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 1, Strict: true}); err == nil {
		t.Fatal("Stomp was cast through a creature-spells permission")
	}
}
