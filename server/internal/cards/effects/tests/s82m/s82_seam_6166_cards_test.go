package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_seam_6166_cards_test.go — #6166: Soulfire Eruption exiles the top
// card for each target in turn and deals that card's mana value to it, and
// the exiled cards are playable until the end of the caster's next turn.
// CR 608.2c: the instructions are followed in the order written.

const seam6166SoulfireOracle = "089dd3b7-c3ed-4734-8789-a3c1ed9a4e40"

func TestSeam6166SoulfireIsFull(t *testing.T) {
	spec, ok := effects.Lookup(seam6166SoulfireOracle)
	if !ok || spec.Name != "Soulfire Eruption" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Soulfire Eruption: %+v %v", spec, ok)
	}
}

// Three targets take damage equal to three successive top cards' mana
// values; the cards stay playable through the caster's next turn and not
// after.
func TestSoulfireEruptionDamagesEachTargetByItsOwnCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp2 := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	beast := effectstest.B12Creature(g, opp.ID, "Big Beast", "Creature — Beast", 6, 6)
	// Pushed bottom-up: the top card is the last pushed.
	fourth := effectstest.TopCard(me, "Unused", "Sorcery", "{7}")
	third := effectstest.TopCard(me, "Third", "Sorcery", "{3}{R}")
	second := effectstest.TopCard(me, "Second", "Land", "")
	first := effectstest.TopCard(me, "First", "Instant", "{1}{R}{R}")
	lifeOpp, lifeOpp2 := opp.Life, opp2.Life
	effectstest.FillPoolColored(me, "R", 3)
	effectstest.FillPool(me, 6)
	effectstest.CastFromHandForTest(t, g, me, "Soulfire Eruption", "Sorcery", "{6}{R}{R}{R}", seam6166SoulfireOracle,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
			{Kind: game.TargetPlayer, ID: opp.ID},
			{Kind: game.TargetPlayer, ID: opp2.ID},
			{Kind: game.TargetCard, ID: beast},
		}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != lifeOpp-3 {
		t.Errorf("first target: life %d → %d, want 3 damage (the first card's mana value)", lifeOpp, opp.Life)
	}
	if opp2.Life != lifeOpp2 {
		t.Errorf("second target: life %d → %d, want no damage (a land's mana value is 0)", lifeOpp2, opp2.Life)
	}
	if got := effectstest.DamageOn(t, g, beast); got != 4 {
		t.Errorf("third target has %d damage, want 4 (the third card's mana value)", got)
	}
	for _, id := range []uuid.UUID{first, second, third} {
		if !g.Exile.Contains(id) {
			t.Fatalf("card %s was not exiled", id)
		}
		if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, id), me.ID) {
			t.Errorf("card %s is not playable", id)
		}
		if effectstest.PermissionLive(g, effectstest.ExiledPermission(g, id), opp.ID) {
			t.Errorf("an opponent may play card %s", id)
		}
	}
	if !me.Library.Contains(fourth) {
		t.Error("a fourth card was exiled for three targets")
	}
	for range g.Seats {
		effectstest.EndTurn(t, g)
	}
	if g.Seats[g.Turn.ActiveSeat].ID != me.ID {
		t.Fatalf("not the caster's next turn")
	}
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, first), me.ID) {
		t.Error("the window closed before the end of the caster's next turn")
	}
	effectstest.EndTurn(t, g)
	if effectstest.PermissionLive(g, effectstest.ExiledPermission(g, first), me.ID) {
		t.Error("the window is still open after the caster's next turn")
	}
}

// With no targets chosen nothing is exiled.
func TestSoulfireEruptionWithNoTargetsExilesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	top := effectstest.TopCard(me, "Top", "Sorcery", "{3}")
	effectstest.FillPoolColored(me, "R", 3)
	effectstest.FillPool(me, 6)
	effectstest.CastFromHandForTest(t, g, me, "Soulfire Eruption", "Sorcery", "{6}{R}{R}{R}", seam6166SoulfireOracle,
		game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Library.Contains(top) {
		t.Error("a card was exiled with no target")
	}
}
