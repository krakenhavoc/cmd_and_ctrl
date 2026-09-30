package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const charmingPrinceOracle = "c48d844c-3976-4fa5-8e0d-3f0e535e7619"

// TestCharmingPrinceScryMode picks the first bullet.
func TestCharmingPrinceScryMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Bottom Me")

	castAndResolveCreature(t, g, "Charming Prince", "Creature — Human Noble", charmingPrinceOracle)

	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick prompt from the ETB: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick(scry): %v", err)
	}
	passPriorityAroundTable(t, g)

	if sc := scryChoiceFor(g, me.ID); sc == nil {
		t.Error("choosing the scry bullet should queue a scry-2 prompt")
	}
}

// TestCharmingPrinceLifeGainMode picks the second bullet.
func TestCharmingPrinceLifeGainMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life

	castAndResolveCreature(t, g, "Charming Prince", "Creature — Human Noble", charmingPrinceOracle)

	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick(life): %v", err)
	}
	passPriorityAroundTable(t, g)

	if me.Life != life+3 {
		t.Errorf("life after the gain-3 bullet: %d, want %d", me.Life, life+3)
	}
}

// TestCharmingPrinceExileMode picks the third bullet: another target
// creature the caster owns is exiled and returns at the next end
// step, as a new object.
func TestCharmingPrinceExileMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pal := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pal", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})

	prince := castAndResolveCreature(t, g, "Charming Prince", "Creature — Human Noble", charmingPrinceOracle)

	modePick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if modePick == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(modePick.ID, me.ID, []int{2}); err != nil {
		t.Fatalf("ResolveModePick(exile): %v", err)
	}
	pickCard(t, g, me.ID, pal)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(pal) {
		t.Fatal("the chosen creature should be exiled")
	}
	if !g.Exile.Contains(pal) {
		t.Fatal("it should be sitting in exile, waiting on the delayed trigger")
	}

	// Charming Prince itself is not a legal target for its own
	// exile bullet ("another").
	_ = prince

	advanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward the end step: %v", err)
		}
	}
	passPriorityAroundTable(t, g)

	if g.Exile.Contains(pal) {
		t.Error("the creature should have returned to the battlefield at the next end step")
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Pal" && c.Controller == me.ID {
			found = true
		}
	}
	if !found {
		t.Error("the returned creature should be under its owner's control")
	}
}
