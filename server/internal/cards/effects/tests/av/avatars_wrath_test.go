package testsav

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatars_wrath_test.go — #1316's first proof card, through the
// engine rather than by reading its own Spec back: the spared
// creature stays, every other creature is airbent, and every opponent
// is locked out of casting from anywhere but their hand until the
// caster's next turn.

const avatarsWrathOracle = "e3431dae-969c-4896-9f9e-a80e7bec4bdf"

// TestAvatarsWrathSparesTheChosenCreatureAndAirbendsTheRest is the
// board half: "choose up to one target creature, then airbend all
// other creatures" — every creature but the target, regardless of
// controller, and the target itself untouched.
func TestAvatarsWrathSparesTheChosenCreatureAndAirbendsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spared := effectstest.PushCatalogPermanent(g, me.ID, "Spared Bear", "Creature — Bear", "", false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "My Other Bear", "Creature — Bear", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	land := effectstest.PushPermanentForTest(g, opp.ID, "Their Forest", "test-avatars-wrath-forest", "Basic Land — Forest")

	effectstest.CastCatalogSpell(t, g, "Avatar's Wrath", "Sorcery", avatarsWrathOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spared}})
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(spared) {
		t.Error("the chosen creature was airbent; it should have been spared")
	}
	if g.Battlefield.Contains(mine) {
		t.Error("another creature the caster controls was not airbent")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's creature was not airbent")
	}
	if !g.Exile.Contains(mine) || !g.Exile.Contains(theirs) {
		t.Error("the airbent creatures did not reach exile")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a LAND was airbent; the card says creatures")
	}

	// Airbend's own grant: each owner may cast their exiled creature
	// for {2} while it remains exiled.
	var perm *game.CastPermission
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(theirs)
		perm = g.CastPermissionForLocked(opp.ID, c, game.ZoneExile)
	})
	if !perm.Granted() {
		t.Error("the opponent's airbent creature carries no cast-from-exile permission")
	}
}

// TestAvatarsWrathLocksEveryOpponentToHandUntilTheCastersNextTurn is
// #1316's own proof: the rider, granted to EACH opponent, reaching
// every zone but the hand, and gone the instant the caster's next
// turn begins.
func TestAvatarsWrathLocksEveryOpponentToHandUntilTheCastersNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]

	effectstest.CastCatalogSpell(t, g, "Avatar's Wrath", "Sorcery", avatarsWrathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	spell := effectstest.InstantCardFor(t, g)
	for _, opp := range []*game.Player{g.Seats[1], g.Seats[2], g.Seats[3]} {
		if err := effectstest.CastGateFor(g, opp.ID, spell, game.ZoneHand); err != nil {
			t.Errorf("seat %s cast from hand should be allowed: %v", opp.ID, err)
		}
		var banErr *game.CantCastError
		if err := effectstest.CastGateFor(g, opp.ID, spell, game.ZoneGraveyard); !errors.As(err, &banErr) {
			t.Errorf("seat %s cast from graveyard: got %v, want a *CantCastError", opp.ID, err)
		}
	}
	// The caster is not their own opponent.
	if err := effectstest.CastGateFor(g, me.ID, spell, game.ZoneGraveyard); err != nil {
		t.Errorf("the caster's own cast from graveyard was refused: %v", err)
	}

	for i := 0; i < 3; i++ {
		effectstest.AdvanceOneTurnForTest(t, g)
		var banErr *game.CantCastError
		if err := effectstest.CastGateFor(g, g.Seats[1].ID, spell, game.ZoneGraveyard); !errors.As(err, &banErr) {
			t.Fatalf("the ban ended after %d opponent turns; it lasts until the CASTER's next turn", i+1)
		}
	}
	effectstest.AdvanceOneTurnForTest(t, g)
	if err := effectstest.CastGateFor(g, g.Seats[1].ID, spell, game.ZoneGraveyard); err != nil {
		t.Errorf("the ban survived the caster's next turn beginning: %v", err)
	}
}

// TestAvatarsWrathExilesItself — "Exile Avatar's Wrath."
func TestAvatarsWrathExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	spell := effectstest.CastCatalogSpell(t, g, "Avatar's Wrath", "Sorcery", avatarsWrathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	var inExile bool
	g.ReadSnapshot(func() { inExile = g.Exile.Contains(spell) })
	if !inExile {
		t.Error("Avatar's Wrath did not exile itself")
	}
}

// TestAvatarsWrathHasNoCaveats is the ADR 0037 coverage signal.
func TestAvatarsWrathHasNoCaveats(t *testing.T) {
	spec, ok := effects.Lookup(avatarsWrathOracle)
	if !ok {
		t.Fatal("Avatar's Wrath is not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Avatar's Wrath is %v with %d caveats, want full with none", spec.Completeness, len(spec.Caveats))
	}
}
