package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3490: Quantum Misalignment's rebound cast (a targeted spell) from the
// player's next upkeep. The free cast goes through cast_spell with a
// target, like a Staggershock rebound.
func TestQuantumMisalignmentReboundCastsFromExile(t *testing.T) {
	const oracle = "197582b9-4c86-41a3-ad0c-789a0db8e087"
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	bear := pushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	_, id := castFromHandAtMain(t, g, "Quantum Misalignment", "Sorcery", oracle,
		game.TargetRef{Kind: game.TargetCard, ID: bear})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Fatal("Quantum Misalignment was not exiled by rebound")
	}
	bears := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Grizzly Bears" {
				n++
			}
		}
		return n
	}
	if bears() != 2 {
		t.Fatalf("first cast: %d Grizzly Bears, want 2", bears())
	}

	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	offer := latestChoiceOfKind(g, game.PendingChoiceResolutionCast)
	if offer == nil {
		t.Fatal("no rebound offer at the controller's next upkeep")
	}
	// #4250: the cast below answers the resolution_cast prompt.
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict:   true,
		FromZone: "exile",
		Targets:  []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("the free rebound cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if bears() != 3 {
		t.Errorf("rebound cast: %d Grizzly Bears, want 3", bears())
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the rebound recast did not go to the graveyard")
	}
}
