package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Force of Will's pitch costs 1 life. At 1 life it is on the table
// and the cast goes through; at 0 it is neither offered nor accepted.
func TestForceOfWillPitchOfferedOnlyWhenTheLifeIsThere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	fow := effectstest.VisibleHandCard(me, "Force of Will", "Instant", "{3}{U}{U}", effectstest.ForceOfWillOracle, []string{"U"})
	pitch := effectstest.VisibleHandCard(me, "Brainstorm", "Instant", "{U}", "test-blue-pitch", []string{"U"})
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: victim}}

	// Paying down to exactly zero is legal (CR 119.4), so 1 life is
	// the last life total that still carries the offer.
	me.Life = 1
	if keys := effectstest.OfferedKeys(t, g, me, fow); !effectstest.HasKey(keys, "pitch") {
		t.Fatalf("at 1 life the pitch offer is missing: %v", keys)
	}

	me.Life = 0
	if keys := effectstest.OfferedKeys(t, g, me, fow); effectstest.HasKey(keys, "pitch") {
		t.Errorf("at 0 life the pitch offer is still shown: %v", keys)
	}
	if err := g.CastSpell(me.ID, fow, game.CastSpellParams{
		Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{pitch}, Targets: target,
	}); err == nil {
		t.Errorf("Force of Will pitched at 0 life was allowed")
	}
	if !me.Hand.Contains(fow) || !me.Hand.Contains(pitch) {
		t.Errorf("a refused pitch moved cards out of hand")
	}

	// Back at 1 life the view offers it and the engine takes it —
	// the same board, the same offer, the same answer.
	me.Life = 1
	if keys := effectstest.OfferedKeys(t, g, me, fow); !effectstest.HasKey(keys, "pitch") {
		t.Fatalf("at 1 life the pitch offer is missing: %v", keys)
	}
	if err := g.CastSpell(me.ID, fow, game.CastSpellParams{
		Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{pitch}, Targets: target,
	}); err != nil {
		t.Fatalf("Force of Will pitched at exactly 1 life: %v", err)
	}
	if me.Life != 0 {
		t.Errorf("life after the pitch = %d, want 0", me.Life)
	}
}

// Snuff Out's 4 life, with the Swamp already there so the Condition
// is not what is being measured.
func TestSnuffOutOfferHiddenBelowFourLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	snuff := effectstest.VisibleHandCard(me, "Snuff Out", "Instant", "{3}{B}", effectstest.SnuffOutOracle, []string{"B"})
	effectstest.FreeSpellPermanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	bear := effectstest.FreeSpellPermanent(g, them.ID, "Grizzly Bears", "Creature — Bear")
	target := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}

	me.Life = 3
	if keys := effectstest.OfferedKeys(t, g, me, snuff); effectstest.HasKey(keys, "pay_life") {
		t.Errorf("at 3 life the pay_life offer is still shown: %v", keys)
	}
	if err := g.CastSpell(me.ID, snuff, game.CastSpellParams{
		Strict: true, AlternativeCost: "pay_life", Targets: target,
	}); err == nil {
		t.Errorf("Snuff Out for 4 life at 3 life was allowed")
	}

	me.Life = 4
	if keys := effectstest.OfferedKeys(t, g, me, snuff); !effectstest.HasKey(keys, "pay_life") {
		t.Fatalf("at exactly 4 life the pay_life offer is missing: %v", keys)
	}
	if err := g.CastSpell(me.ID, snuff, game.CastSpellParams{
		Strict: true, AlternativeCost: "pay_life", Targets: target,
	}); err != nil {
		t.Fatalf("Snuff Out at exactly 4 life: %v", err)
	}
	if me.Life != 0 {
		t.Errorf("life after Snuff Out = %d, want 0 — paying down to zero is legal", me.Life)
	}
}
