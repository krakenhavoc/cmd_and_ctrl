package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// urborg_test.go is the end-to-end proof that Layer 4 is load-bearing
// rather than wire-only. Urborg, Tomb of Yawgmoth is the card #258
// wrote up and then refused to ship: its one static "would apply
// cleanly and do nothing," because everything downstream of the
// layer engine read the printed type line.
//
// The assertions deliberately go all the way to real mana in a real
// pool via the real activation path. An assertion on
// Effective().Subtypes alone would have passed before this branch
// and is exactly the check that let the gap survive four sprints.

const urborgOracle = "db6174d7-211d-4817-b8e4-8384594c83f9"

// seedLand puts a land on the battlefield through the layer-aware
// push so the listener stamps EnteredBattlefieldAt and invalidates
// the cached resolution.
func seedLand(g *game.Game, owner uuid.UUID, name, typeLine, oracleID string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      owner,
		Controller: owner,
	})
}

// poolColors lists the colour letters currently in a player's mana
// pool, in pool order.
func poolColors(p *game.Player) []string {
	out := make([]string, 0, len(p.ManaPool))
	for _, tok := range p.ManaPool {
		out = append(out, tok.Color)
	}
	return out
}

// TestUrborgMakesEveryLandTapForBlack is the headline: a Mountain
// with Urborg out taps for {B}, and the mana lands in the pool.
func TestUrborgMakesEveryLandTapForBlack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mountain := seedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")

	// Before Urborg: one ability, {R}.
	if got := manaAbilityProduced(t, g, mountain); len(got) != 1 || got[0] != "{R}" {
		t.Fatalf("bare Mountain produces %v, want [{R}]", got)
	}

	seedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", urborgOracle)

	got := manaAbilityProduced(t, g, mountain)
	if len(got) != 2 || got[0] != "{R}" || got[1] != "{B}" {
		t.Fatalf("Mountain under Urborg produces %v, want [{R} {B}]", got)
	}

	if err := g.ActivateManaAbility(me.ID, mountain, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility for the Urborg-granted {B}: %v", err)
	}
	if colors := poolColors(me); len(colors) != 1 || colors[0] != "B" {
		t.Fatalf("mana pool = %v, want [B]", colors)
	}
}

// TestUrborgIsItselfASwamp — Urborg is a land, so its own static
// applies to it. A lone Urborg taps for {B} and nothing else.
func TestUrborgIsItselfASwamp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	urborg := seedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", urborgOracle)

	if got := manaAbilityProduced(t, g, urborg); len(got) != 1 || got[0] != "{B}" {
		t.Fatalf("lone Urborg produces %v, want [{B}]", got)
	}
	if err := g.ActivateManaAbility(me.ID, urborg, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if colors := poolColors(me); len(colors) != 1 || colors[0] != "B" {
		t.Fatalf("mana pool = %v, want [B]", colors)
	}
}

// TestUrborgAppliesToEveryPlayersLands — "each land", not "each land
// you control". The opponent's Island taps for {B} too.
func TestUrborgAppliesToEveryPlayersLands(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	island := seedLand(g, them.ID, "Island", "Basic Land — Island", "")
	seedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", urborgOracle)

	got := manaAbilityProduced(t, g, island)
	if len(got) != 2 || got[0] != "{U}" || got[1] != "{B}" {
		t.Fatalf("opponent's Island under Urborg produces %v, want [{U} {B}]", got)
	}
}

// TestUrborgLeavingRevokesTheBlackMana — the static lives exactly as
// long as the permanent (CR 113.6). Bounce Urborg and the Mountain
// is a Mountain again.
func TestUrborgLeavingRevokesTheBlackMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mountain := seedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	urborg := seedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", urborgOracle)

	if got := manaAbilityProduced(t, g, mountain); len(got) != 2 {
		t.Fatalf("precondition: Mountain produces %v, want two abilities", got)
	}
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID},
		urborg,
	); err != nil {
		t.Fatalf("bounce Urborg: %v", err)
	}
	if got := manaAbilityProduced(t, g, mountain); len(got) != 1 || got[0] != "{R}" {
		t.Fatalf("Mountain after Urborg left produces %v, want [{R}]", got)
	}
}

// TestUrborgDoesNotMakeNonlandsSwamps — the static is gated on
// IsLand, and IsLand is now an effective read, so this is also the
// regression guard that the predicate didn't quietly become "every
// permanent."
func TestUrborgDoesNotMakeNonlandsSwamps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedCreature(g, "Grizzly Bears", me.ID)
	seedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", urborgOracle)

	if got := manaAbilityProduced(t, g, bear); len(got) != 0 {
		t.Errorf("a creature under Urborg produces %v, want none", got)
	}
	if containsString(effectiveSubtypes(t, g, bear), "Swamp") {
		t.Error("a creature under Urborg gained the Swamp subtype")
	}
}

// TestUrborgReachesTheWire — the CardView the client renders lists
// the granted ability, so a player can actually click it. Without
// this, the engine would be right and the game unplayable.
func TestUrborgReachesTheWire(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	seedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", urborgOracle)

	// ViewOfGame takes the read lock (and the layer refresh) itself,
	// so it must not be nested inside another ReadSnapshot.
	gv := protocol.ViewOfGame(g)
	var view *protocol.CardView
	for i := range gv.Battlefield.Cards {
		if gv.Battlefield.Cards[i].InstanceID == forest.String() {
			view = &gv.Battlefield.Cards[i]
			break
		}
	}
	if view == nil {
		t.Fatal("Forest missing from the battlefield view")
	}
	if len(view.ManaAbilities) != 2 {
		t.Fatalf("CardView.ManaAbilities = %+v, want two entries", view.ManaAbilities)
	}
	if view.TypeLine != "Basic Land — Forest Swamp" {
		t.Errorf("CardView.TypeLine = %q, want the rebuilt line with Swamp", view.TypeLine)
	}
}

// manaAbilityProduced returns the Produced string of each mana
// ability the engine currently offers on a battlefield card,
// through a snapshot so the layer engine has resolved.
func manaAbilityProduced(t *testing.T, g *game.Game, id uuid.UUID) []string {
	t.Helper()
	var out []string
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != id {
				continue
			}
			found = true
			for _, ab := range game.ManaAbilitiesForCard(c) {
				out = append(out, ab.Produced)
			}
			return
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}

// effectiveSubtypes returns Effective().Subtypes for a battlefield
// card, forcing a snapshot.
func effectiveSubtypes(t *testing.T, g *game.Game, id uuid.UUID) []string {
	t.Helper()
	var out []string
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c.Effective().Subtypes, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}
