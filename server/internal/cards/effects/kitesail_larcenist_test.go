package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const kitesailLarcenistOracle = "2452be47-cc23-47f7-a3a1-fec900bb0119"

// Flying is a printed keyword and reaches the effective
// characteristic like any other.
func TestKitesailLarcenistFlies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Kitesail Larcenist",
		"Creature — Human Pirate", kitesailLarcenistOracle, false)
	if !hasAbility(effectiveAbilities(t, g, id), "flying") {
		t.Errorf("Kitesail Larcenist has flying, got %v", effectiveAbilities(t, g, id))
	}
}

// Ward {1} is a triggered ability (CR 702.21a), so an opponent's spell
// is a LEGAL target and the tax lands on the caster as a payment
// prompt once the ward trigger resolves.
func TestKitesailLarcenistTaxesAnOpponentsSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	larcenist := pushCatalogPermanent(g, me.ID, "Kitesail Larcenist",
		"Creature — Human Pirate", kitesailLarcenistOracle, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Doom Blade", TypeLine: "Instant",
		OracleID: doomBladeOracle, Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: larcenist}},
	}); err != nil {
		t.Fatalf("a warded permanent is a legal target: %v", err)
	}
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatalf("ward {1} asks the CASTER to pay: %+v", g.PendingChoices)
	}
}

// The enters trigger is deferred whole. Shipping the type change and
// the ability loss without the mana ability they come with would turn
// an opponent's Sol Ring into an inert rock — strictly better removal
// than the card prints (#259) — so nothing of it ships, and the
// declaration says so.
func TestKitesailLarcenistDefersItsEntersTrigger(t *testing.T) {
	spec, ok := Lookup(kitesailLarcenistOracle)
	if !ok {
		t.Fatal("Kitesail Larcenist is registered")
	}
	if spec.Completeness != CompletenessCaveats {
		t.Fatalf("completeness = %v, want caveats", spec.Completeness)
	}
	if len(spec.Caveats) != 1 || !strings.Contains(spec.Caveats[0], "Treasure") {
		t.Errorf("the caveat names the Treasures it does not make: %v", spec.Caveats)
	}
	if spec.AsEnters != nil {
		t.Error("nothing happens as it enters")
	}
	for _, ab := range spec.Triggered {
		for _, w := range ab.Watches {
			if w == game.EventETB {
				t.Error("no enters-the-battlefield trigger ships")
			}
		}
	}
	if !hasAbility(spec.PrintedKeywords, "flying") {
		t.Errorf("flying is printed, got %v", spec.PrintedKeywords)
	}
}
