package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// land_adventure_cards_test.go — the four Town lands with an Adventure
// half (#2176). The engine lifecycle is pinned in
// game/adventure_land_test.go; these prove each card's two halves.

// landAdventureCard builds a printed land // adventure card the way the
// deck importer does.
func landAdventureCard(owner uuid.UUID, oracle, landName, adventureName, cost string, colors []string) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracle,
		Layout:     game.LayoutAdventure,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: landName, TypeLine: "Land — Town"},
			{Name: adventureName, TypeLine: "Sorcery — Adventure", ManaCost: cost, Colors: colors},
		},
	}
	c.SetFace(0)
	return c
}

// seatLandAdventure puts the card in the active seat's hand at main
// phase 1 and returns the game, seat and card ID.
func seatLandAdventure(t *testing.T, oracle, landName, adventureName, cost string, colors []string) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := landAdventureCard(me.ID, oracle, landName, adventureName, cost, colors)
	me.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return g, me, c.InstanceID
}

// assertLandPlayableFromExile checks CR 715.3d's grant is a play
// permission, then plays the land and checks it entered tapped.
func assertLandPlayableFromExile(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) {
	t.Helper()
	if !g.Exile.Contains(id) {
		t.Fatal("CR 715.3d: the resolved Adventure is not in exile")
	}
	perm := g.CastPermissionOnCardByIDForEffect(id)
	if !perm.Granted() || perm.CastOnly {
		t.Fatalf("grant = %+v, want a play (not cast-only) permission", perm)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("play the land from exile: %v", err)
	}
	var landed *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			landed = &g.Battlefield.Cards[i]
		}
	}
	if landed == nil {
		t.Fatal("the land did not reach the battlefield")
	}
	if !landed.IsLand() || !landed.Tapped {
		t.Errorf("land = %+v, want a land that entered tapped", landed)
	}
}

func TestJidoorOvertureMillsHalfThenTheLandIsPlayableFromExile(t *testing.T) {
	g, me, id := seatLandAdventure(t, jidoorOracleID, "Jidoor, Aristocratic Capital", "Overture", "{4}{U}{U}", []string{"U"})
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	before := opp.Library.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Face:    1,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("cast Overture: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != before-before/2 {
		t.Errorf("library %d -> %d, want half milled", before, got)
	}
	assertLandPlayableFromExile(t, g, me, id)
}

func TestJidoorIsAPlainLandPlayFromHand(t *testing.T) {
	g, me, id := seatLandAdventure(t, jidoorOracleID, "Jidoor, Aristocratic Capital", "Overture", "{4}{U}{U}", []string{"U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play Jidoor: %v", err)
	}
	if !g.Battlefield.Contains(id) || g.Stack.Contains(id) {
		t.Fatal("Jidoor should be on the battlefield as a land play, not on the stack")
	}
}

func TestLastingFaythMakesAHeroWithACounterPerLand(t *testing.T) {
	g, me, id := seatLandAdventure(t, zanarkandOracleID, "Zanarkand, Ancient Metropolis", "Lasting Fayth", "{4}{G}{G}", []string{"G"})
	for i := 0; i < 3; i++ {
		l := game.NewCard("Test Forest", me.ID)
		l.TypeLine = "Basic Land — Forest"
		g.Battlefield.PushTop(l)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Lasting Fayth: %v", err)
	}
	passPriorityAroundTable(t, g)
	var hero *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Hero" {
			hero = &g.Battlefield.Cards[i]
		}
	}
	if hero == nil {
		t.Fatal("no Hero token")
	}
	if got := hero.CurrentPower(); got != 4 {
		t.Errorf("Hero power = %d, want 1 + 3 counters (three lands)", got)
	}
	assertLandPlayableFromExile(t, g, me, id)
}

func TestTakeATripToDrawsTwoAndHitsEachOpponent(t *testing.T) {
	g, me, id := seatLandAdventure(t, valueTownOracleID, "Value Town", "Take a Trip to...", "{4}{U}{R}", []string{"U", "R"})
	handBefore := me.Hand.Size()
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Take a Trip to...: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The spell left the hand (-1) and two cards were drawn (+2).
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d -> %d, want +1 net (drew two, cast one)", handBefore, got)
	}
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want -= 2
		}
		if p.Life != want {
			t.Errorf("%s life = %d, want %d", p.Name, p.Life, want)
		}
	}
	assertLandPlayableFromExile(t, g, me, id)
}

func TestFaithAndGriefReturnsUpToTwoArtifactsOrEnchantments(t *testing.T) {
	g, me, id := seatLandAdventure(t, ishgardOracleID, "Ishgard, the Holy See", "Faith & Grief", "{3}{W}{W}", []string{"W"})
	mk := func(name, typeLine string) uuid.UUID {
		c := game.NewCard(name, me.ID)
		c.TypeLine = typeLine
		me.Graveyard.PushTop(c)
		return c.InstanceID
	}
	rock := mk("Test Rock", "Artifact")
	aura := mk("Test Ward", "Enchantment")
	bear := mk("Test Bear", "Creature — Bear")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Face: 1,
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: rock},
			{Kind: game.TargetCard, ID: aura},
		},
	}); err != nil {
		t.Fatalf("cast Faith & Grief: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) || !me.Hand.Contains(aura) {
		t.Error("both targets should return to hand")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the creature was never targetable and should stay put")
	}
	assertLandPlayableFromExile(t, g, me, id)

	// A creature card is not a legal target.
	g2, me2, id2 := seatLandAdventure(t, ishgardOracleID, "Ishgard, the Holy See", "Faith & Grief", "{3}{W}{W}", []string{"W"})
	c := game.NewCard("Test Bear", me2.ID)
	c.TypeLine = "Creature — Bear"
	me2.Graveyard.PushTop(c)
	if err := g2.CastSpell(me2.ID, id2, game.CastSpellParams{
		Face:    1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: c.InstanceID}},
	}); err == nil {
		t.Error("a creature card was accepted as a target")
	}
}
