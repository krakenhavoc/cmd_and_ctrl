package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// enters_from_graveyard_test.go — EventETB carries the zone (and its
// owner) the permanent came from (#2135, ADR 0113 amendment 2026-10-08).

const (
	efgPitDweller = "899387c5-6781-485b-b8bc-aa89d3b97413"
	efgFlayer     = "ba9f3c5d-556f-48d1-8d87-723f7893cfcd"
	efgKelpie     = "ef5c93ae-91e0-4edb-b8a3-1171a3694bb8"
)

func efgController(g *game.Game, id uuid.UUID) uuid.UUID {
	var out uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out = c.Controller
			}
		}
	})
	return out
}

func TestEfgRegisteredFull(t *testing.T) {
	for _, o := range []string{efgPitDweller, efgFlayer, efgKelpie} {
		spec, ok := effects.Lookup(o)
		if !ok || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: registered=%v completeness=%v", o, ok, spec.Completeness)
		}
	}
}

// The event names the graveyard and its owner when undying returns a
// creature, and nothing when a creature is put onto the battlefield from
// nowhere.
func TestEnterEventCarriesOriginZone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ghoul := effectstest.PoolPush(g, me.ID, "Sightless Ghoul", effectstest.PoolSightlessGhoul, 2, 2)
	effectstest.B25Destroy(g, ghoul)
	effectstest.SettleProwess(t, g)
	if !effects.OnBattlefield(g, ghoul) {
		t.Fatal("undying did not return the Ghoul")
	}
	ev := effectstest.EfgLastETB(t, g, ghoul)
	if ev.EnteredFrom != game.ZoneGraveyard || ev.EnteredFromOwner != me.ID {
		t.Errorf("ETB origin = %q/%s, want graveyard/%s", ev.EnteredFrom, ev.EnteredFromOwner, me.ID)
	}
}

// Treacherous Pit-Dweller: undying hands it to the chosen opponent; a
// cast does not.
func TestPitDwellerChangesHandsOnlyWhenItReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dw := effectstest.PoolPush(g, me.ID, "Treacherous Pit-Dweller", efgPitDweller, 4, 3)
	effectstest.B25Destroy(g, dw)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.SettleProwess(t, g)
	if !effects.OnBattlefield(g, dw) || effectstest.CountersOn(g, dw, game.CounterPlusOne) != 1 {
		t.Fatal("undying did not return it with a +1/+1 counter")
	}
	if got := efgController(g, dw); got != opp.ID {
		t.Errorf("controller %s, want the opponent %s", got, opp.ID)
	}
}

func TestPitDwellerCastKeepsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.CastCatalogSpell(t, g, "Treacherous Pit-Dweller", "Creature — Demon", efgPitDweller, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, id) {
		t.Fatal("the creature did not resolve")
	}
	if got := efgController(g, id); got != me.ID {
		t.Errorf("controller %s, want the caster %s", got, me.ID)
	}
	for _, o := range g.Seats {
		if effectstest.LatestPickTarget(g, o.ID) != nil {
			t.Error("a cast Pit-Dweller asked for a target")
		}
	}
}

// River Kelpie draws when any permanent enters from a graveyard, and
// not when one is cast.
func TestRiverKelpieDrawsOnGraveyardEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PoolPush(g, me.ID, "River Kelpie", efgKelpie, 3, 3)
	ghoul := effectstest.PoolPush(g, opp.ID, "Sightless Ghoul", effectstest.PoolSightlessGhoul, 2, 2)
	hand := me.Hand.Size()
	effectstest.B25Destroy(g, ghoul)
	effectstest.SettleProwess(t, g)
	if !effects.OnBattlefield(g, ghoul) {
		t.Fatal("undying did not return the opponent's Ghoul")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("Kelpie's controller drew %d, want 1", got)
	}
}

func TestRiverKelpieDoesNotDrawForACast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PoolPush(g, me.ID, "River Kelpie", efgKelpie, 3, 3)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want no draw for a cast creature", hand, me.Hand.Size())
	}
}

// The sandbox move (the context menu) stamps the origin too.
func TestSandboxMoveFromGraveyardStampsOrigin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushGraveyardCreature(g, me.ID, "Grizzly Bears", "{1}{G}")
	err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, id)
	if err != nil {
		t.Fatal(err)
	}
	ev := effectstest.EfgLastETB(t, g, id)
	if ev.EnteredFrom != game.ZoneGraveyard || ev.EnteredFromOwner != me.ID {
		t.Errorf("ETB origin = %q/%s", ev.EnteredFrom, ev.EnteredFromOwner)
	}
}

// Flayer of the Hatebound triggers off YOUR graveyard only.
func TestFlayerDamagesOffYourGraveyardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PoolPush(g, me.ID, "Flayer of the Hatebound", efgFlayer, 4, 2)

	theirs := effectstest.PoolPush(g, opp.ID, "Sightless Ghoul", effectstest.PoolSightlessGhoul, 2, 2)
	effectstest.B25Destroy(g, theirs)
	effectstest.SettleProwess(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("an opponent's creature returning from their graveyard triggered the Flayer")
	}

	mine := effectstest.PoolPush(g, me.ID, "Sightless Ghoul", effectstest.PoolSightlessGhoul, 2, 2)
	life := opp.Life
	effectstest.B25Destroy(g, mine)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.SettleProwess(t, g)
	// Returned with a +1/+1 counter: 3 power.
	if opp.Life != life-3 {
		t.Errorf("opponent life %d → %d, want -3 (the returned Ghoul's power)", life, opp.Life)
	}
}
