package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nahiris_lithoforming_test.go — Nahiri's Lithoforming (#6147): "Sacrifice
// X lands. For each land sacrificed this way, draw a card. You may play X
// additional lands this turn. Lands you control enter tapped this turn."

const nahirisLithoformingOracle = "888e711a-1251-4704-8ef1-b4b98ec54b5a"

func TestNahirisLithoformingIsFullWithNoCaveats(t *testing.T) {
	spec, ok := effects.Lookup(nahirisLithoformingOracle)
	if !ok || spec.Name != "Nahiri's Lithoforming" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec %+v (registered %v)", spec, ok)
	}
}

func lithoformingLand(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Basic Land — Mountain",
		Owner: owner, Controller: owner,
	})
}

// playLandFromHandTapped plays a land from the active player's hand
// through the real land-play path and reports whether it entered tapped.
func playLandFromHandTapped(t *testing.T, g *game.Game, p *game.Player, name string) bool {
	t.Helper()
	id := effectstest.HandCard(p, name, "Basic Land — Mountain")
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok || !g.Battlefield.Contains(id) {
		t.Fatalf("%s is not on the battlefield", name)
	}
	return c.Tapped
}

func TestNahirisLithoformingSacrificesXDrawsGrantsPlaysAndTapsYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	l1 := lithoformingLand(g, me.ID, "Mountain 1")
	l2 := lithoformingLand(g, me.ID, "Mountain 2")
	l3 := lithoformingLand(g, me.ID, "Mountain 3")
	oppLand := lithoformingLand(g, opp.ID, "Their Mountain")
	for i := 0; i < 3; i++ {
		effectstest.SeedLibraryTop(me, "Library Card", "Sorcery", "{R}")
	}
	effectstest.PaCast(t, g, "Nahiri's Lithoforming", "Sorcery", "{X}{R}{R}", nahirisLithoformingOracle, game.CastSpellParams{XValue: 2})
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatalf("no sacrifice prompt: %+v", g.PendingChoices)
	}
	if c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Errorf("sacrifice bounds %d..%d, want exactly 2", c.ChooseMin, c.ChooseMax)
	}
	for _, id := range c.ChooseCards {
		if id == oppLand {
			t.Error("an opponent's land was offered")
		}
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{l1, l3}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(l1) || g.Battlefield.Contains(l3) || !g.Battlefield.Contains(l2) {
		t.Fatal("exactly the two chosen lands are sacrificed")
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d, want two cards drawn", hand, got)
	}
	if got := g.LandDropsRemainingFor(me.ID); got != 3 {
		t.Errorf("land plays left = %d, want 3 (the turn's one plus X = 2)", got)
	}
	for i := 1; i <= 3; i++ {
		if !playLandFromHandTapped(t, g, me, "New Mountain") {
			t.Errorf("land play %d entered untapped", i)
		}
	}
	if g.LandDropsRemainingFor(me.ID) != 0 {
		t.Error("more land plays than the turn's one plus X")
	}
	// An opponent's land is not "lands you control".
	theirs := effectstest.HandCard(opp, "Their New Land", "Basic Land — Mountain")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, theirs); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if c, _ := g.LookupCardForEffect(theirs); c.Tapped {
		t.Error("an opponent's land entered tapped")
	}
	// The next turn, the caster's lands enter untapped again.
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	effectstest.AdvanceToMainOf(t, g, 0)
	if playLandFromHandTapped(t, g, me, "Next Turn Mountain") {
		t.Error("the enters-tapped effect outlived its turn")
	}
}

// With X = 0 nothing is sacrificed or drawn, and lands still enter tapped.
func TestNahirisLithoformingWithXZeroStillTapsYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lithoformingLand(g, me.ID, "Mountain")
	effectstest.PaCast(t, g, "Nahiri's Lithoforming", "Sorcery", "{X}{R}{R}", nahirisLithoformingOracle, game.CastSpellParams{XValue: 0})
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); c != nil {
		t.Fatal("X = 0 asked for a sacrifice")
	}
	if me.Hand.Size() != hand {
		t.Error("X = 0 drew a card")
	}
	if g.LandDropsRemainingFor(me.ID) != 1 {
		t.Errorf("land plays left = %d, want the turn's one", g.LandDropsRemainingFor(me.ID))
	}
	if !playLandFromHandTapped(t, g, me, "New Mountain") {
		t.Error("the land entered untapped")
	}
}
