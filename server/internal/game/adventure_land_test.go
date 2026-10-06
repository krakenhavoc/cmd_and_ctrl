package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// adventure_land_test.go — #2176: an Adventure whose main half is a
// LAND (the Final Fantasy Town lands). CR 715.3d lets the controller
// PLAY it from exile, and a land play spends the turn's land drop
// (CR 305.2).

func landAdventureFixture(owner uuid.UUID) Card {
	c := Card{
		InstanceID: uuid.New(),
		OracleID:   "cccccccc-dddd-eeee-ffff-000000000002",
		Owner:      owner,
		Controller: owner,
		Layout:     LayoutAdventure,
		Faces: []Face{
			{Name: "Fixture Town", TypeLine: "Land — Town"},
			{Name: "Fixture Overture", TypeLine: "Sorcery — Adventure", ManaCost: "{0}", Colors: []string{"U"}},
		},
	}
	c.SetFace(0)
	return c
}

func handWithLandAdventure(t *testing.T, g *Game, p *Player) uuid.UUID {
	t.Helper()
	card := landAdventureFixture(p.ID)
	g.WithWriteLock(func() {
		p.Hand.PushTop(card)
		g.markCardKnownInZoneLocked(p.Hand, card.InstanceID)
	})
	return card.InstanceID
}

// exiledLandAdventure casts the Adventure half and lets it resolve.
func exiledLandAdventure(t *testing.T, g *Game, me *Player) uuid.UUID {
	t.Helper()
	id := handWithLandAdventure(t, g, me)
	if err := g.CastSpell(me.ID, id, CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Adventure half of a land card: %v", err)
	}
	passPriorityUntilResolvedForTest(t, g, id)
	if !g.Exile.Contains(id) {
		t.Fatal("the resolved Adventure is not in exile")
	}
	return id
}

func TestLandAdventureOffersBothHalvesFromHand(t *testing.T) {
	card := landAdventureFixture(uuid.New())
	if faces := card.CastableFaces(); len(faces) != 2 {
		t.Fatalf("CastableFaces = %v, want both halves", faces)
	}
	if !adventureMainHalfIsLand(card) {
		t.Fatal("main half not recognised as a land")
	}
	if adventureMainHalfIsLand(adventureFixture(uuid.New())) {
		t.Fatal("a creature main half read as a land")
	}
}

func TestLandAdventureGrantIsAPlayPermission(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := exiledLandAdventure(t, g, me)
	perm := adventureGrantFor(g, id)
	if !perm.Granted() {
		t.Fatal("no permission on the exiled land card")
	}
	if perm.CastOnly {
		t.Error("CR 715.3d says play: a land main half must not be cast-only")
	}
	if face, ok := perm.NamedFace(); !ok || face != 0 {
		t.Errorf("grant faces = %v, want [0]", perm.Faces)
	}
}

func TestLandAdventurePlayLandFromExileUsesTheLandDrop(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := exiledLandAdventure(t, g, me)
	before := g.LandDropsRemainingFor(me.ID)
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("play the land from exile: %v", err)
	}
	landed := findCardForTest(g.Battlefield, id)
	if landed == nil || landed.Name != "Fixture Town" || !landed.IsLand() {
		t.Fatalf("land did not reach the battlefield as the land face: %+v", landed)
	}
	if got := g.LandDropsRemainingFor(me.ID); got != before-1 {
		t.Errorf("land drops remaining = %d, want %d", got, before-1)
	}
	if g.Exile.Contains(id) {
		t.Error("the land is still in exile")
	}
}

func TestLandAdventurePlayFromExileRefusedWithoutALandDrop(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := exiledLandAdventure(t, g, me)
	// Spend the turn's drop on another land from hand.
	other := handWithLandAdventure(t, g, me)
	if err := g.CastSpell(me.ID, other, CastSpellParams{}); err != nil {
		t.Fatalf("first land drop: %v", err)
	}
	err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile"})
	if !errors.Is(err, ErrLandDropUnavailable) {
		t.Fatalf("err = %v, want ErrLandDropUnavailable", err)
	}
	if !g.Exile.Contains(id) {
		t.Error("a refused land play moved the card")
	}
}

func TestLandAdventureCannotBePlayedOnAnOpponentsTurn(t *testing.T) {
	g := newActiveGame(t)
	me, you := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	id := exiledLandAdventure(t, g, me)
	passTurn(t, g)
	toMainPhase(t, g)
	if g.Turn.ActiveSeat == 0 {
		t.Fatal("expected the opponent's turn")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile"}); err == nil {
		t.Fatal("the land was played on an opponent's turn")
	}
	if err := g.CastSpell(you.ID, id, CastSpellParams{FromZone: "exile"}); err == nil {
		t.Fatal("an opponent played the exiled land")
	}
	if !g.Exile.Contains(id) {
		t.Error("the card left exile")
	}
}

func TestCounteredLandAdventureGoesToTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := handWithLandAdventure(t, g, me)
	if err := g.CastSpell(me.ID, id, CastSpellParams{Face: 1}); err != nil {
		t.Fatal(err)
	}
	if err := g.CounterSpell(id, nil); err != nil {
		t.Fatal(err)
	}
	if g.Exile.Contains(id) || !me.Graveyard.Contains(id) {
		t.Fatal("a countered land Adventure must be in the graveyard, not exile")
	}
	if adventureGrantFor(g, id).Granted() {
		t.Error("a countered Adventure left a permission")
	}
}

func TestLandAdventureCastFromHandIsAPlainLandPlay(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := handWithLandAdventure(t, g, me)
	if err := g.CastSpell(me.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("play the land from hand: %v", err)
	}
	if findCardForTest(g.Battlefield, id) == nil {
		t.Fatal("land not on the battlefield")
	}
}
