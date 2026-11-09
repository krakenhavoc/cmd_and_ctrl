package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const dragonhawkOracle = "f7f435bc-1716-4412-9505-a1f2f2d83d08"

func TestDragonhawkIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(dragonhawkOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Dragonhawk: %+v %v", spec, ok)
	}
}

func dragonhawkBoard(t *testing.T, g *game.Game) (me, opp *game.Player, dragon uuid.UUID, ids []uuid.UUID) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	dragon = dskB1Creature(g, me.ID, "Dragonhawk, Fate's Tempest", "Legendary Creature — Bird Dragon", dragonhawkOracle, 5, 5)
	b12Creature(g, me.ID, "Big Bear", "Creature — Bear", 4, 4)
	b12Creature(g, me.ID, "Small Bear", "Creature — Bear", 3, 3)
	ids = seedLibrary(me, "A", "B", "C", "D", "E")
	return
}

func TestDragonhawkAttackExilesXAndDamagesForEachStillExiled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, dragon, ids := dragonhawkBoard(t, g)
	declareAttack(t, g, opp.ID, dragon)
	passPriorityAroundTable(t, g)
	// X = 2: Dragonhawk (5) and the 4/4. The 3/3 does not count.
	for _, id := range ids[:2] {
		if z := zoneOfCard(g, id); z != game.ZoneExile {
			t.Fatalf("card %v is in %v, want exile", id, z)
		}
	}
	if z := zoneOfCard(g, ids[2]); z != game.ZoneLibrary {
		t.Errorf("a third card left the library (%v)", z)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	life := opp.Life
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 4 {
		t.Errorf("end-step damage %d, want 4 (2 for each of 2 cards still exiled)", got)
	}
	_ = me
}

func TestDragonhawkDamagesOnlyForCardsStillExiled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, dragon, ids := dragonhawkBoard(t, g)
	declareAttack(t, g, opp.ID, dragon)
	passPriorityAroundTable(t, g)
	// One of the exiled cards leaves exile (played or otherwise).
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, ids[0]); err != nil {
		t.Fatalf("move: %v", err)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	life := opp.Life
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 2 {
		t.Errorf("end-step damage %d, want 2 (one card still exiled)", got)
	}
}

func zoneOfCard(g *game.Game, id uuid.UUID) game.ZoneKind {
	var k game.ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			k = z.Kind
		}
	})
	return k
}
