package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const lobeliaDefenderOracle = "0f2e3e2e-9857-4be5-9bf0-62da1cc147fc"

func TestLobeliaDefenderOfBagEndIsRegisteredFull(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, lobeliaDefenderOracle, "Lobelia, Defender of Bag End")
}

// lobeliaTable casts Lobelia and resolves its enters trigger, after
// seeding each opponent's top card. It answers Lobelia and, per seat, the
// card seeded on top (uuid.Nil for Lobelia's controller).
func lobeliaTable(t *testing.T, tops map[int][3]string) (*game.Game, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	seeded := make([]uuid.UUID, len(g.Seats))
	for i, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		x := tops[i]
		_ = effectstest.TopOfLibrary(p, "Under", "Sorcery", "{B}")
		seeded[i] = effectstest.TopOfLibrary(p, x[0], x[1], x[2])
	}
	lobelia := effectstest.CastCatalogSpell(t, g, "Lobelia, Defender of Bag End", "Legendary Creature — Halfling Citizen", lobeliaDefenderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g) // Lobelia resolves
	effectstest.PassPriorityAroundTable(t, g) // its enters trigger resolves
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == lobelia {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	return g, lobelia, seeded
}

func lobeliaArtifact(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Trinket", TypeLine: "Artifact", Owner: owner, Controller: owner})
}

func lobeliaActivate(g *game.Game, me uuid.UUID, lobelia, artifact uuid.UUID, mode int) error {
	return g.ActivateCatalogAbility(me, lobelia, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{artifact}, Modes: []int{mode}})
}

// The enters trigger exiles each opponent's top card face down, linked to
// Lobelia, which only Lobelia's controller may look at. The first mode
// lets them play ONE of those cards free this turn: a land with their
// land drop; after that, nothing else.
func TestLobeliaExilesEachOpponentsTopAndPlaysOneFree(t *testing.T) {
	seat := 0
	g := effectstest.NewCatalogGame(t)
	seat = g.Turn.ActiveSeat
	tops := map[int][3]string{}
	for i := range g.Seats {
		tops[i] = [3]string{"Their Sorcery", "Sorcery", "{5}{U}{U}"}
	}
	tops[(seat+1)%4] = [3]string{"Their Forest", "Basic Land — Forest", ""}
	g, lobelia, seeded := lobeliaTable(t, tops)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	for i, id := range seeded {
		if id == uuid.Nil {
			continue
		}
		c := effectstest.CardInExile(t, g, id)
		if c.ExiledWith.ID != lobelia {
			t.Fatalf("seat %d's top card linked to %s, want Lobelia", i, c.ExiledWith.ID)
		}
		if !c.FaceDown {
			t.Fatalf("seat %d's top card is exiled face up", i)
		}
	}
	land := seeded[(g.Turn.ActiveSeat+1)%4]
	spell := seeded[(g.Turn.ActiveSeat+2)%4]
	if effectstest.ViewCastable(t, g, opp.ID, land) {
		t.Error("an opponent may play a card exiled with Lobelia")
	}
	// Without the activation, nothing is playable.
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("played a card exiled with Lobelia before the activation")
	}
	if err := lobeliaActivate(g, me.ID, lobelia, lobeliaArtifact(g, me.ID), 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("play the exiled land: %v", err)
	}
	if z := effectstest.E2Zone(g, land); z != game.ZoneBattlefield {
		t.Fatalf("the land is in %q, want the battlefield", z)
	}
	// One play only: the seven-mana sorcery is no longer free.
	if effectstest.EnumeratedFor(g, me.ID, spell) {
		t.Error("a second card is still offered after the one play")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("cast a second card exiled with Lobelia")
	}
}

// The spell is cast without paying its mana cost, and the permission ends
// with the turn.
func TestLobeliaCastsASpellFreeOnlyThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	tops := map[int][3]string{}
	for i := range g.Seats {
		tops[i] = [3]string{"Their Sorcery", "Sorcery", "{5}{U}{U}"}
	}
	g, lobelia, seeded := lobeliaTable(t, tops)
	me := g.Seats[g.Turn.ActiveSeat]
	a, b := seeded[(g.Turn.ActiveSeat+1)%4], seeded[(g.Turn.ActiveSeat+2)%4]
	if err := lobeliaActivate(g, me.ID, lobelia, lobeliaArtifact(g, me.ID), 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, a, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the exiled sorcery free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.E2Zone(g, a); z != game.ZoneGraveyard {
		t.Fatalf("the cast sorcery is in %q, want its owner's graveyard", z)
	}
	// A second activation next turn: the earlier grant has ended, the new
	// one opens one more play.
	mine := g.Turn.ActiveSeat
	effectstest.AdvanceToMainOf(t, g, (mine+1)%4)
	effectstest.AdvanceToMainOf(t, g, mine)
	if err := g.CastSpell(me.ID, b, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("cast a card exiled with Lobelia on a later turn without activating")
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// The second mode drains each opponent for 2.
func TestLobeliaDrainMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	lobelia := effectstest.PushCatalogPermanent(g, me.ID, "Lobelia, Defender of Bag End", "Legendary Creature — Halfling Citizen", lobeliaDefenderOracle, false)
	if err := lobeliaActivate(g, me.ID, lobelia, uuid.New(), 1); err == nil {
		t.Fatal("activated without sacrificing an artifact")
	}
	if err := lobeliaActivate(g, me.ID, lobelia, lobeliaArtifact(g, me.ID), 1); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := 38
		if p.ID == me.ID {
			want = 42
		}
		if p.Life != want {
			t.Errorf("%s at %d life, want %d", p.Name, p.Life, want)
		}
	}
}
