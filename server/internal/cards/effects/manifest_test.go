package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// manifest_test.go — #1194 / ADR 0082, the manifest / disguise half of
// the CARD side. The engine-side object model is pinned in
// server/internal/game/face_down_test.go and the mechanics in
// face_down_cast_test.go; what belongs here is that two cards reach
// them and that the ward hook is really wired.

const (
	soulSummonsOracle   = "0f5b79ca-9f80-420b-a6c5-bb2a9a95c7e7"
	forumFamiliarOracle = "7ea0012b-b90b-43d6-bb5b-e4e92452bea7"
)

// TestSoulSummonsManifestsTheTopCard is CR 701.34a end to end: the top
// card of the library arrives face down as a 2/2 its controller alone
// may look at, with no ETB trigger and no text (CR 708.2a).
func TestSoulSummonsManifestsTheTopCard(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	// A loud card on top, so anything that leaked would be obvious.
	loud := game.NewCard("Griselbrand", active.ID)
	loud.TypeLine = "Legendary Creature — Demon"
	loud.OracleID = "oracle-manifested-demon"
	loud.ManaCost = "{4}{B}{B}"
	loud.Power, loud.Toughness = 7, 7
	loud.Keywords = []string{"flying", "lifelink"}
	top := loud.InstanceID
	g.WithWriteLock(func() { active.Library.PushTop(loud) })

	castCatalogSpell(t, g, "Soul Summons", "Sorcery", soulSummonsOracle, nil)
	passPriorityAroundTable(t, g)

	c := findCardAnywhere(t, g, top)
	if !c.FaceDownIsPermanent() || c.FaceDownKind != game.FaceDownManifested {
		t.Fatalf("face-down state = (%v, %q), want a manifested permanent", c.FaceDown, c.FaceDownKind)
	}
	if c.Controller != active.ID {
		t.Errorf("controller = %s, want the caster", c.Controller)
	}
	if !c.IsKnownTo(active.ID) {
		t.Error("the manifesting player cannot look at their own manifest (CR 708.5)")
	}
	if c.IsKnownTo(opp.ID) {
		t.Error("an opponent can read a manifested card — it was never revealed")
	}
	if game.HasKeyword(&c, "flying") {
		t.Error("a manifested permanent kept the card's printed keyword (CR 708.2a)")
	}
	eff := c.Effective()
	if eff.Power != 2 || eff.Toughness != 2 || eff.Name != "" {
		t.Errorf("body = %q %d/%d, want a nameless 2/2", eff.Name, eff.Power, eff.Toughness)
	}

	// CR 701.40b: a manifested CREATURE card may be turned face up for
	// its MANA COST — not for any morph cost, and not at all if it is
	// not a creature card. The offer is the engine's; this asserts the
	// card reaches it.
	offer := game.TurnFaceUpOffer(c)
	if offer == nil {
		t.Fatal("a manifested creature card offers no way up (CR 701.40b)")
	}
	if offer.Cost != "{4}{B}{B}" {
		t.Errorf("turn-face-up cost = %q, want the card's mana cost", offer.Cost)
	}
}

// TestAManifestedNoncreatureStaysFaceDown is the other half of
// CR 701.40b, and the reason the offer cannot be a catalog
// declaration: the same spell, a different top card, no way up.
func TestAManifestedNoncreatureStaysFaceDown(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]

	land := game.NewCard("Mountain", active.ID)
	land.TypeLine = "Basic Land — Mountain"
	top := land.InstanceID
	g.WithWriteLock(func() { active.Library.PushTop(land) })

	castCatalogSpell(t, g, "Soul Summons", "Sorcery", soulSummonsOracle, nil)
	passPriorityAroundTable(t, g)

	c := findCardAnywhere(t, g, top)
	if !c.FaceDownIsPermanent() {
		t.Fatal("the land did not manifest — CR 701.34a manifests ANY card")
	}
	if !c.IsCreature() {
		t.Error("a manifested land is not a creature; CR 708.2 makes the object a 2/2 creature")
	}
	if offer := game.TurnFaceUpOffer(c); offer != nil {
		t.Errorf("a manifested land may be turned face up for %q — CR 701.40b says only a creature card may", offer.Cost)
	}
}

// TestForumFamiliarIsDisguisedWithWard is the disguise round trip, and
// the only thing that separates disguise from morph: the face-down
// object has ward {2} (CR 702.168a), read off the STATE and gone the
// moment the permanent is face up.
func TestForumFamiliarIsDisguisedWithWard(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]

	id := castWithAltCost(t, g, "Forum Familiar", "Creature — Cat", forumFamiliarOracle, "disguise")
	passPriorityAroundTable(t, g)

	down := findCardAnywhere(t, g, id)
	if down.FaceDownKind != game.FaceDownDisguised {
		t.Fatalf("face-down kind = %q, want disguised", down.FaceDownKind)
	}
	wards := game.TriggersForCard(down)
	if len(wards) != 1 || wards[0].Key != "Ward {2}" {
		t.Fatalf("a disguised permanent's abilities = %+v, want exactly ward {2}", wards)
	}
	if offer := game.TurnFaceUpOffer(down); offer == nil || offer.Cost != "{1}{W}" {
		t.Fatalf("turn-face-up offer = %+v, want the disguise cost {1}{W}", offer)
	}

	// A permanent to bounce, so the CR 708.8 trigger has a legal
	// target when it fires.
	bear := game.NewCard("Grizzly Bears", active.ID)
	bear.TypeLine = "Creature — Bear"
	bear.Controller = active.ID
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(bear)
		active.ManaPool.AddMana(manaTokens("W", "C")...)
	})

	if err := g.PerformSpecialAction(active.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	up := findCardAnywhere(t, g, id)
	if up.FaceDown {
		t.Fatal("still face down")
	}
	if got := game.TriggersForCard(up); len(got) == 0 || got[0].Key == "Ward {2}" {
		t.Errorf("abilities after turning up = %+v, want the card's own rather than the state's ward", got)
	}
}
