package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	dreadReturnOracle    = "352b64d2-2ae5-44ee-a64f-94932ef545d3"
	cruelCelebrantOracle = "3ee78cfc-0e9e-4737-a7e2-b42f94228040"
)

// The hand-cast half: an ordinary reanimation.
func TestDreadReturnReanimatesFromHand(t *testing.T) {
	g := newCatalogGame(t)
	victim := seedGraveyardCard(t, g, "Fodder Bear", "Creature — Bear", "")

	castCatalogSpell(t, g, "Dread Return", "Sorcery", dreadReturnOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: victim},
	})
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(victim) {
		t.Fatalf("the targeted creature should be on the battlefield")
	}
}

// dreadReturnFlashbackBoard seats Dread Return in the active seat's
// graveyard beside the creature card it will return, and gives the
// seat `n` creatures to sacrifice.
func dreadReturnFlashbackBoard(t *testing.T, g *game.Game, n int) (dread, wurm uuid.UUID, fodder []uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	wurm = seedGraveyardCard(t, g, "Buried Wurm", "Creature — Wurm", "")
	dread = seedGraveyardCard(t, g, "Dread Return", "Sorcery", dreadReturnOracle)
	for i := 0; i < n; i++ {
		fodder = append(fodder, seedCreature(g, "Fodder", me.ID))
	}
	return dread, wurm, fodder
}

func flashBackDreadReturn(g *game.Game, dread, wurm uuid.UUID, pay []uuid.UUID) error {
	return g.CastSpell(g.Seats[g.Turn.ActiveSeat].ID, dread, game.CastSpellParams{
		FromZone:        "graveyard",
		AlternativeCost: "flashback",
		AltCostIDs:      pay,
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: wurm}},
	})
}

// The flashback half (#1727), end to end against the real catalog:
// three creatures sacrificed at announce, the creature card returned
// on resolution, and Dread Return exiled rather than left in the
// graveyard to be flashed back again.
func TestDreadReturnFlashbackSacrificesThreeAndReanimates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dread, wurm, fodder := dreadReturnFlashbackBoard(t, g, 3)

	if !game.CardCastableFromZone(dreadReturnOracle, game.ZoneGraveyard) {
		t.Fatalf("Dread Return does not declare the graveyard as a cast zone")
	}
	if err := flashBackDreadReturn(g, dread, wurm, fodder); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	// Paid at announce: dead before the spell resolves.
	for _, id := range fodder {
		if g.Battlefield.Contains(id) {
			t.Error("a sacrificed creature is still on the battlefield with the spell on the stack")
		}
		if !me.Graveyard.Contains(id) {
			t.Error("a sacrificed creature is not in its owner's graveyard")
		}
	}
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(wurm) {
		t.Errorf("the targeted creature card did not return to the battlefield")
	}
	if me.Graveyard.Contains(dread) {
		t.Errorf("a flashed-back Dread Return returned to the graveyard")
	}
	if !g.Exile.Contains(dread) {
		t.Errorf("a flashed-back Dread Return was not exiled (CR 702.34a)")
	}
}

// The three dies triggers are ABOVE the spell, so they resolve first:
// with Cruel Celebrant out, the first item off the stack is a drain
// and the Wurm is still in the graveyard.
func TestDreadReturnFlashbackDeathsResolveBeforeTheReanimation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	dread, wurm, fodder := dreadReturnFlashbackBoard(t, g, 3)
	pushCatalogPermanent(g, me.ID, "Cruel Celebrant", "Creature — Vampire", cruelCelebrantOracle, false)
	oppBefore := opp.Life

	if err := flashBackDreadReturn(g, dread, wurm, fodder); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	spell := g.StackMeta[dread]
	if spell == nil {
		t.Fatalf("Dread Return is not on the stack")
	}
	above := 0
	for _, item := range g.StackMeta {
		if item.Kind == game.StackItemTriggered && item.Seq > spell.Seq {
			above++
		}
	}
	if above != 3 {
		t.Fatalf("%d Cruel Celebrant triggers above Dread Return, want 3 — one per creature sacrificed to the cost", above)
	}

	// One full round of passes resolves the top object: a drain.
	for range g.Seats {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if opp.Life != oppBefore-1 {
		t.Errorf("after one resolution the opponent lost %d, want the 1 a drain takes", oppBefore-opp.Life)
	}
	if g.Battlefield.Contains(wurm) {
		t.Errorf("the reanimation resolved before the dies triggers above it")
	}

	passPriorityAroundTable(t, g)
	if opp.Life != oppBefore-3 {
		t.Errorf("the opponent lost %d in all, want 3 — one drain per sacrifice", oppBefore-opp.Life)
	}
	if !g.Battlefield.Contains(wurm) {
		t.Errorf("the Wurm never returned")
	}
}

// Countering a flashed-back Dread Return gives nothing back: the three
// creatures stay in the graveyard, the Wurm stays with them, and the
// card is exiled.
func TestCounteredDreadReturnFlashbackKeepsTheCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dread, wurm, fodder := dreadReturnFlashbackBoard(t, g, 3)

	if err := flashBackDreadReturn(g, dread, wurm, fodder); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	if err := g.CounterSpell(dread, nil); err != nil {
		t.Fatalf("CounterSpell: %v", err)
	}
	for _, id := range fodder {
		if !me.Graveyard.Contains(id) {
			t.Error("a sacrificed creature left the graveyard when the spell was countered")
		}
	}
	if !me.Graveyard.Contains(wurm) {
		t.Error("the Wurm moved although the spell never resolved")
	}
	if !g.Exile.Contains(dread) {
		t.Error("a countered flashed-back Dread Return was not exiled")
	}
}

// Two creatures cannot pay "sacrifice three": the cast is refused with
// nothing paid, and the printed {2}{B}{B} is not a way out of the
// graveyard either — flashback is the only price there.
func TestDreadReturnFlashbackRefusesTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dread, wurm, fodder := dreadReturnFlashbackBoard(t, g, 2)

	if err := flashBackDreadReturn(g, dread, wurm, fodder); err == nil {
		t.Fatalf("a two-creature payment bought a three-creature flashback")
	}
	if err := g.CastSpell(me.ID, dread, game.CastSpellParams{
		FromZone: "graveyard",
		Targets:  []game.TargetRef{{Kind: game.TargetCard, ID: wurm}},
	}); err == nil {
		t.Fatalf("Dread Return was cast from the graveyard for its printed cost")
	}
	if !me.Graveyard.Contains(dread) {
		t.Error("a refused cast moved Dread Return")
	}
	for _, id := range fodder {
		if !g.Battlefield.Contains(id) {
			t.Error("a refused cast sacrificed a creature anyway")
		}
	}
}
