package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b146_test.go — GAP batch 146 (#5948): Golem Artisan, Phyrexian
// Ironfoot and Shambling Suit.

const (
	gap146Artisan  = "8537fc2c-c027-4440-bff4-d046fc198428"
	gap146Ironfoot = "83b4acd0-751d-4882-a32d-6c089bc652ee"
	gap146Suit     = "862b777a-4816-4a3e-be8d-ddfcb72268e8"
)

func TestGap146EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gap146Artisan: "Golem Artisan", gap146Ironfoot: "Phyrexian Ironfoot", gap146Suit: "Shambling Suit",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gap146Keyword(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	return effectstest.HasKeywordOnBattlefield(t, g, id, kw)
}

func TestGap146ArtisanPumpsOnlyArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	artisan := effectstest.PushCatalogPermanent(g, me.ID, "Golem Artisan", "Artifact Creature — Golem", gap146Artisan, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	robot := effectstest.B12Creature(g, opp.ID, "Robot", "Artifact Creature — Robot", 1, 1)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.FillPool(me, 2)
	activate := func(target uuid.UUID) error {
		err := g.ActivateCatalogAbility(me.ID, artisan, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(target)})
		if err == nil {
			effectstest.PassPriorityAroundTable(t, g)
		}
		return err
	}
	if err := activate(bear); err == nil {
		t.Fatal("a nonartifact creature was a legal target")
	}
	if err := activate(rock); err == nil {
		t.Fatal("a noncreature artifact was a legal target")
	}
	if err := activate(robot); err != nil {
		t.Fatalf("{2}: %v", err)
	}
	if got := effectstest.EffectivePower(t, g, robot); got != 2 {
		t.Errorf("robot power %d, want 2", got)
	}
	effectstest.FillPool(me, 2)
	before := effectstest.EffectivePower(t, g, artisan)
	if err := activate(artisan); err != nil {
		t.Fatalf("{2} on itself: %v", err)
	}
	if got := effectstest.EffectivePower(t, g, artisan); got != before+1 {
		t.Errorf("artisan power %d, want %d", got, before+1)
	}
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, artisan, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(robot)}); err == nil {
		t.Fatal("activated with one mana")
	}
}

func TestGap146ArtisanGrantsTheChosenKeyword(t *testing.T) {
	for idx, kw := range []string{"flying", "trample", "haste"} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		artisan := effectstest.PushCatalogPermanent(g, me.ID, "Golem Artisan", "Artifact Creature — Golem", gap146Artisan, false)
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		robot := effectstest.B12Creature(g, opp.ID, "Robot", "Artifact Creature — Robot", 1, 1)
		effectstest.FillPool(me, 2)
		if err := g.ActivateCatalogAbility(me.ID, artisan, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
			t.Fatal("a nonartifact creature was a legal target")
		}
		if err := g.ActivateCatalogAbility(me.ID, artisan, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(robot)}); err != nil {
			t.Fatalf("{2}: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionPick)
		if c == nil {
			t.Fatal("no keyword prompt")
		}
		if err := g.ResolveOptionPick(c.ID, c.Chooser, idx); err != nil {
			t.Fatalf("ResolveOptionPick: %v", err)
		}
		for i, other := range []string{"flying", "trample", "haste"} {
			if got := gap146Keyword(t, g, robot, other); got != (i == idx) {
				t.Errorf("chose %s: robot has %s = %v", kw, other, got)
			}
		}
		if gap146Keyword(t, g, artisan, kw) {
			t.Errorf("the Artisan gained %s too", kw)
		}
	}
}

func TestGap146IronfootUntapsForSnowManaOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	foot := effectstest.PushCatalogPermanent(g, me.ID, "Phyrexian Ironfoot", "Snow Artifact Creature — Phyrexian Construct", gap146Ironfoot, false)
	effectstest.BlbB1Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", 0, 0)
	effectstest.BlbB1Push(g, me.ID, "Island", "Basic Land — Island", "", 0, 0)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, foot).Tapped = true })
	activate := func() error {
		err := g.ActivateCatalogAbility(me.ID, foot, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true})
		if err == nil {
			effectstest.PassPriorityAroundTable(t, g)
		}
		return err
	}
	if err := activate(); err == nil {
		t.Fatal("paid {S} with nonsnow lands")
	}
	if !effectstest.IsTapped(g, foot) {
		t.Fatal("untapped without paying")
	}
	effectstest.BlbB1Push(g, me.ID, "Snow-Covered Island", "Basic Snow Land — Island", "", 0, 0)
	if err := activate(); err != nil {
		t.Fatalf("{1}{S}: %v", err)
	}
	if effectstest.IsTapped(g, foot) {
		t.Error("still tapped after the ability resolved")
	}
}

func TestGap146IronfootDoesntUntapDuringItsControllersUntapStep(t *testing.T) {
	spec, _ := effects.Lookup(gap146Ironfoot)
	if len(spec.UntapStepRestrictions) != 1 {
		t.Fatalf("untap-step restrictions %d, want 1", len(spec.UntapStepRestrictions))
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	foot := effectstest.PushCatalogPermanent(g, me.ID, "Phyrexian Ironfoot", "Snow Artifact Creature — Phyrexian Construct", gap146Ironfoot, false)
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, foot).Tapped = true
		effectstest.FindBattlefieldCardForTest(g, other).Tapped = true
	})
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceToMainOf(t, g, 0)
	if !effectstest.IsTapped(g, foot) {
		t.Error("the Ironfoot untapped in its controller's untap step")
	}
	if effectstest.IsTapped(g, other) {
		t.Error("an ordinary creature stayed tapped through the untap step")
	}
}

func TestGap146SuitCountsArtifactsAndEnchantmentsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	suit := effectstest.PushCatalogPermanent(g, me.ID, "Shambling Suit", "Artifact Creature — Construct", gap146Suit, false)
	if got := effectstest.EffectivePower(t, g, suit); got != 1 {
		t.Fatalf("power alone %d, want 1 (itself)", got)
	}
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Aura", "Enchantment")
	effectstest.B12Permanent(g, me.ID, "Both", "Enchantment Artifact")
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	if got := effectstest.EffectivePower(t, g, suit); got != 4 {
		t.Errorf("power %d, want 4 (suit, rock, aura, artifact enchantment once)", got)
	}
}
