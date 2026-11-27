package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b392_cards_test.go — Theros removal, batch 392 (#5561, S82 OLD).
// Every test plays the real catalog entry through the engine.

const (
	oldB392Sorrow   = "115fd5ac-460b-49a3-83ed-9d56a7aecb14"
	oldB392Ray      = "509bfcbc-a59a-47c8-be78-29823fc010e3"
	oldB392Vanquish = "6990ac82-49d4-412d-9c3d-1e7fb5479313"
	oldB392Priest   = "706f7d56-f4da-48ba-ac94-fc09580beec7"
	oldB392Revelry  = "7165fade-ea96-4026-bc80-620729b8f70c"
	oldB392Eruption = "f34eb8a9-63af-452b-8957-8587bf4fd956"
)

func TestOldB392CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB392Sorrow: "Artisan's Sorrow", oldB392Ray: "Ray of Dissolution", oldB392Vanquish: "Vanquish the Foul",
		oldB392Priest: "Priest of Iroas", oldB392Revelry: "Destructive Revelry", oldB392Eruption: "Peak Eruption",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB392Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// oldB392Cast casts a free spell strictly at the targets.
func oldB392Cast(g *game.Game, name, typeLine, oracle string, targets ...uuid.UUID) error {
	p := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{0}", OracleID: oracle, Owner: p.ID, Controller: p.ID})
	})
	return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTargets(targets...)})
}

func oldB392Resolve(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
}

func oldB392Scry(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil {
		t.Fatal("the scry asked nothing")
	}
	if err := g.ResolveScry(scry.ID, p.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
}

func TestOldB392ArtisansSorrowDestroysAnArtifactOrEnchantmentThenScries(t *testing.T) {
	g, me, opp := oldB392Table(t)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	ench := effectstest.B21Push(g, opp.ID, "Their Enchantment", "Enchantment", "", 0, 0)
	art := effectstest.B21Push(g, opp.ID, "Their Artifact", "Artifact", "", 0, 0)
	if err := oldB392Cast(g, "Artisan's Sorrow", "Instant", oldB392Sorrow, bear); err == nil {
		t.Fatal("cast at a creature")
	}
	for _, id := range []uuid.UUID{ench, art} {
		if err := oldB392Cast(g, "Artisan's Sorrow", "Instant", oldB392Sorrow, id); err != nil {
			t.Fatalf("cast: %v", err)
		}
		oldB392Resolve(t, g)
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", id)
		}
		oldB392Scry(t, g, me)
		oldB392Resolve(t, g)
	}
}

func TestOldB392RayOfDissolutionDestroysAnEnchantmentAndGainsThree(t *testing.T) {
	g, me, opp := oldB392Table(t)
	art := effectstest.B21Push(g, opp.ID, "Their Artifact", "Artifact", "", 0, 0)
	ench := effectstest.B21Push(g, opp.ID, "Their Enchantment", "Enchantment", "", 0, 0)
	if err := oldB392Cast(g, "Ray of Dissolution", "Instant", oldB392Ray, art); err == nil {
		t.Fatal("cast at an artifact")
	}
	life := me.Life
	if err := oldB392Cast(g, "Ray of Dissolution", "Instant", oldB392Ray, ench); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB392Resolve(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestOldB392VanquishTheFoulNeedsPowerFourAndScries(t *testing.T) {
	g, me, opp := oldB392Table(t)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 9)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 4)
	if err := oldB392Cast(g, "Vanquish the Foul", "Sorcery", oldB392Vanquish, small); err == nil {
		t.Fatal("cast at a creature with power 3")
	}
	if err := oldB392Cast(g, "Vanquish the Foul", "Sorcery", oldB392Vanquish, big); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB392Resolve(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the 4/4 survived")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the other creature was destroyed")
	}
	oldB392Scry(t, g, me)
}

func TestOldB392PriestOfIroasSacrificesItselfToDestroyAnEnchantment(t *testing.T) {
	g, me, opp := oldB392Table(t)
	priest := effectstest.PushCatalogPermanent(g, me.ID, "Priest of Iroas", "Creature — Human Cleric", oldB392Priest, false)
	ench := effectstest.B21Push(g, opp.ID, "Their Enchantment", "Enchantment", "", 0, 0)
	effectstest.B06AddMana(me, "W", "W", "W", "W")
	effectstest.B16Activate(t, g, me.ID, priest, 0, game.ActivateAbilityParams{Targets: effectstest.CardTargets(ench)})
	if g.Battlefield.Contains(priest) {
		t.Error("the Priest was not sacrificed as a cost")
	}
	oldB392Resolve(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
}

func TestOldB392DestructiveRevelryDestroysAndDamagesItsController(t *testing.T) {
	g, _, opp := oldB392Table(t)
	art := effectstest.B21Push(g, opp.ID, "Their Artifact", "Artifact", "", 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := oldB392Cast(g, "Destructive Revelry", "Instant", oldB392Revelry, bear); err == nil {
		t.Fatal("cast at a creature")
	}
	life := opp.Life
	if err := oldB392Cast(g, "Destructive Revelry", "Instant", oldB392Revelry, art); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB392Resolve(t, g)
	if g.Battlefield.Contains(art) {
		t.Error("the artifact survived")
	}
	if opp.Life != life-2 {
		t.Errorf("the controller is at %d, want %d", opp.Life, life-2)
	}
}

func TestOldB392PeakEruptionDestroysAMountainAndDamagesItsController(t *testing.T) {
	g, _, opp := oldB392Table(t)
	mountain := effectstest.PushLandFor(g, opp.ID, "Mountain", "Basic Land — Mountain")
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := oldB392Cast(g, "Peak Eruption", "Sorcery", oldB392Eruption, forest); err == nil {
		t.Fatal("cast at a Forest")
	}
	life := opp.Life
	if err := oldB392Cast(g, "Peak Eruption", "Sorcery", oldB392Eruption, mountain); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB392Resolve(t, g)
	if g.Battlefield.Contains(mountain) {
		t.Error("the Mountain survived")
	}
	if opp.Life != life-3 {
		t.Errorf("the controller is at %d, want %d", opp.Life, life-3)
	}
}
