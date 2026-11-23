package testsu

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestUrborgMakesEveryLandTapForBlack is the headline: a Mountain
// with Urborg out taps for {B}, and the mana lands in the pool.
func TestUrborgMakesEveryLandTapForBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mountain := effectstest.SeedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")

	// Before Urborg: one ability, {R}.
	if got := effectstest.ManaAbilityProduced(t, g, mountain); len(got) != 1 || got[0] != "{R}" {
		t.Fatalf("bare Mountain produces %v, want [{R}]", got)
	}

	effectstest.SeedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", effectstest.UrborgOracle)

	got := effectstest.ManaAbilityProduced(t, g, mountain)
	if len(got) != 2 || got[0] != "{R}" || got[1] != "{B}" {
		t.Fatalf("Mountain under Urborg produces %v, want [{R} {B}]", got)
	}

	if err := g.ActivateManaAbility(me.ID, mountain, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility for the Urborg-granted {B}: %v", err)
	}
	if colors := effectstest.PoolColors(me); len(colors) != 1 || colors[0] != "B" {
		t.Fatalf("mana pool = %v, want [B]", colors)
	}
}

// TestUrborgIsItselfASwamp — Urborg is a land, so its own static
// applies to it. A lone Urborg taps for {B} and nothing else.
func TestUrborgIsItselfASwamp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	urborg := effectstest.SeedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", effectstest.UrborgOracle)

	if got := effectstest.ManaAbilityProduced(t, g, urborg); len(got) != 1 || got[0] != "{B}" {
		t.Fatalf("lone Urborg produces %v, want [{B}]", got)
	}
	if err := g.ActivateManaAbility(me.ID, urborg, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if colors := effectstest.PoolColors(me); len(colors) != 1 || colors[0] != "B" {
		t.Fatalf("mana pool = %v, want [B]", colors)
	}
}

// TestUrborgAppliesToEveryPlayersLands — "each land", not "each land
// you control". The opponent's Island taps for {B} too.
func TestUrborgAppliesToEveryPlayersLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	island := effectstest.SeedLand(g, them.ID, "Island", "Basic Land — Island", "")
	effectstest.SeedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", effectstest.UrborgOracle)

	got := effectstest.ManaAbilityProduced(t, g, island)
	if len(got) != 2 || got[0] != "{U}" || got[1] != "{B}" {
		t.Fatalf("opponent's Island under Urborg produces %v, want [{U} {B}]", got)
	}
}

// TestUrborgLeavingRevokesTheBlackMana — the static lives exactly as
// long as the permanent (CR 113.6). Bounce Urborg and the Mountain
// is a Mountain again.
func TestUrborgLeavingRevokesTheBlackMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mountain := effectstest.SeedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	urborg := effectstest.SeedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", effectstest.UrborgOracle)

	if got := effectstest.ManaAbilityProduced(t, g, mountain); len(got) != 2 {
		t.Fatalf("precondition: Mountain produces %v, want two abilities", got)
	}
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID},
		urborg,
	); err != nil {
		t.Fatalf("bounce Urborg: %v", err)
	}
	if got := effectstest.ManaAbilityProduced(t, g, mountain); len(got) != 1 || got[0] != "{R}" {
		t.Fatalf("Mountain after Urborg left produces %v, want [{R}]", got)
	}
}

// TestUrborgDoesNotMakeNonlandsSwamps — the static is gated on
// IsLand, and IsLand is now an effective read, so this is also the
// regression guard that the predicate didn't quietly become "every
// permanent."
func TestUrborgDoesNotMakeNonlandsSwamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedCreature(g, "Grizzly Bears", me.ID)
	effectstest.SeedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", effectstest.UrborgOracle)

	if got := effectstest.ManaAbilityProduced(t, g, bear); len(got) != 0 {
		t.Errorf("a creature under Urborg produces %v, want none", got)
	}
	if effectstest.ContainsString(effectstest.EffectiveSubtypes(t, g, bear), "Swamp") {
		t.Error("a creature under Urborg gained the Swamp subtype")
	}
}

// TestUrborgReachesTheWire — the CardView the client renders lists
// the granted ability, so a player can actually click it. Without
// this, the engine would be right and the game unplayable.
func TestUrborgReachesTheWire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	effectstest.SeedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", effectstest.UrborgOracle)

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

// TestUrborgTurnsOnACheckland — the second thing Urborg does at a
// real table, and the one that proves a Layer-4 type reaches code
// nobody wrote for Urborg. Drowned Catacomb checks for an Island or
// a Swamp; a Forest is neither
// (TestChecklandRejectsTheWrongLandType pins that), but under
// Urborg a Forest is a Swamp, so the checkland enters untapped.
func TestUrborgTurnsOnACheckland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.SeedPermanentFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.SeedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", effectstest.UrborgOracle)

	id := effectstest.PlayLandFromHand(t, g, "Drowned Catacomb", effectstest.DrownedCatacombOracle)
	effectstest.Top100AssertEnteredUntapped(t, g, id, "Drowned Catacomb")
}

// TestUrborgAddsBlackToACatalogDualWithoutDoublingIt pins the merge
// rule in ManaAbilitiesForCard. Sunpetal Grove declares a "{G} or
// {W}" pipe; Urborg gives it the {B} the pipe cannot make, and does
// not hand it a second way to make {G}. Drowned Catacomb's pipe
// already covers {B}, so it gains nothing at all and its client row
// stays one click.
func TestUrborgAddsBlackToACatalogDualWithoutDoublingIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	grove := effectstest.SeedLand(g, me.ID, "Sunpetal Grove", "Land — Forest Plains", effectstest.SunpetalGroveOracle)
	catacomb := effectstest.SeedLand(g, me.ID, "Drowned Catacomb", "Land — Island Swamp", effectstest.DrownedCatacombOracle)
	effectstest.SeedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", effectstest.UrborgOracle)

	got := effectstest.ManaAbilityProduced(t, g, grove)
	if len(got) != 2 || got[0] != "{G|W}" || got[1] != "{B}" {
		t.Errorf("Sunpetal Grove under Urborg produces %v, want [{G|W} {B}]", got)
	}
	if got := effectstest.ManaAbilityProduced(t, g, catacomb); len(got) != 1 || got[0] != "{U|B}" {
		t.Errorf("Drowned Catacomb under Urborg produces %v, want just its [{U|B}] pipe", got)
	}
}
