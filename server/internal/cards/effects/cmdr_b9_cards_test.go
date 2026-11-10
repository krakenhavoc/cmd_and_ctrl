package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b9_cards_test.go — Commander deck cards, batches 9 to 12 (#5368).

const (
	cmdrB9SecludedSteppeOracle   = "8dec6fcf-1254-4b1b-ba23-7a3e492a7241"
	cmdrB9TranquilThicketOracle  = "9f8fe514-77ed-41b4-a6f3-c6f095bb97be"
	cmdrB9LonelySandbarOracle    = "765863c8-1be0-4bb1-9e9c-db7701cffde3"
	cmdrB9DakmorSalvageOracle    = "cdc4048a-73ec-4ec1-a179-2b36c397bf1a"
	cmdrB9RazortideBridgeOracle  = "6cb37ac1-dd11-4a8c-bca5-ef44d828059f"
	cmdrB9RustvaleBridgeOracle   = "a3faf70d-c034-4692-9e92-1922029e3852"
	cmdrB9SilverbluffBridgeOracl = "081bfd50-a436-463b-9d2c-5bc8a32b387c"
)

// cmdrB9Registered pins that every oracle id is a Full card with no caveats.
func cmdrB9Registered(t *testing.T, want map[string]string) {
	t.Helper()
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestCmdrB9LandsAreRegisteredFull(t *testing.T) {
	cmdrB9Registered(t, map[string]string{
		cmdrB9SecludedSteppeOracle: "Secluded Steppe", cmdrB9TranquilThicketOracle: "Tranquil Thicket",
		cmdrB9LonelySandbarOracle: "Lonely Sandbar", cmdrB9DakmorSalvageOracle: "Dakmor Salvage",
		cmdrB9RazortideBridgeOracle: "Razortide Bridge", cmdrB9RustvaleBridgeOracle: "Rustvale Bridge",
		cmdrB9SilverbluffBridgeOracl: "Silverbluff Bridge",
	})
}

func TestCmdrB9CyclingLandsEnterTappedAndCycleForTheirColour(t *testing.T) {
	for _, tc := range []struct{ name, oracle, mana string }{
		{"Secluded Steppe", cmdrB9SecludedSteppeOracle, "{W}"},
		{"Tranquil Thicket", cmdrB9TranquilThicketOracle, "{G}"},
		{"Lonely Sandbar", cmdrB9LonelySandbarOracle, "{U}"},
	} {
		g := newCatalogGame(t)
		id := b12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})
		if !tappedForTest(t, g, id) {
			t.Errorf("%s did not enter tapped", tc.name)
		}
		g2 := newCatalogGame(t)
		me := g2.Seats[g2.Turn.ActiveSeat]
		before := me.Hand.Size()
		cid, _ := cycleFromHand(t, g2, tc.name, "Land", tc.oracle, tc.mana)
		if me.Hand.Contains(cid) {
			t.Errorf("%s is still in hand after cycling", tc.name)
		}
		if me.Hand.Size() != before {
			t.Errorf("%s: hand %d after cycling, want %d (discard one, draw one)", tc.name, me.Hand.Size(), before)
		}
	}
}

func TestCmdrB9CyclingNeedsItsColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := pushCatalogHandCard(me, "Secluded Steppe", "Land", cmdrB9SecludedSteppeOracle)
	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("cycling {W} was paid with colourless mana")
	}
}

func TestCmdrB9BridgesEnterTappedIndestructibleAndTapForEitherColour(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		colours      []string
	}{
		{"Razortide Bridge", cmdrB9RazortideBridgeOracle, []string{"{W}", "{U}"}},
		{"Rustvale Bridge", cmdrB9RustvaleBridgeOracle, []string{"{R}", "{W}"}},
		{"Silverbluff Bridge", cmdrB9SilverbluffBridgeOracl, []string{"{U}", "{R}"}},
	} {
		g := newCatalogGame(t)
		id := b12PlayFromHand(t, g, tc.name, "Artifact Land", tc.oracle, game.CastSpellParams{})
		if !tappedForTest(t, g, id) {
			t.Errorf("%s did not enter tapped", tc.name)
		}
		if !hasKeywordOnBattlefield(t, g, id, "indestructible") {
			t.Errorf("%s is not indestructible", tc.name)
		}
		spec, _ := Lookup(tc.oracle)
		if len(spec.ManaAbilities) != 1 {
			t.Fatalf("%s: %d mana abilities", tc.name, len(spec.ManaAbilities))
		}
		produced := spec.ManaAbilities[0].Produced
		for _, col := range tc.colours {
			if !strings.Contains(produced, col[1:2]) {
				t.Errorf("%s produces %q, missing %s", tc.name, produced, col)
			}
		}
	}
}

func TestCmdrB9DakmorSalvageDredgesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Dakmor Salvage", TypeLine: "Land",
		OracleID: cmdrB9DakmorSalvageOracle, Owner: me.ID, Controller: me.ID})
	lib := me.Library.Size()
	_ = g.DrawCard(me.ID)
	offer := latestChoiceOfKindFor(g, game.PendingChoiceOptionalReplacement, me.ID)
	if offer == nil || offer.DredgeOffer() != 2 {
		t.Fatalf("no dredge 2 offer: %+v", offer)
	}
	if err := g.ResolveOptionalReplacement(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if !me.Hand.Contains(land) {
		t.Error("Dakmor Salvage did not return to hand")
	}
	if got := me.Library.Size(); got != lib-2 {
		t.Errorf("library %d, want %d (milled two, drew none)", got, lib-2)
	}
}
