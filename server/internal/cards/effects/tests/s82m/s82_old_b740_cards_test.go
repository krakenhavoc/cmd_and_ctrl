package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b740_cards_test.go — Exodus, "Pump and combat tricks", batch 740
// (#5626). Every test plays the real catalog entry through the engine.

const (
	o740Palisade = "0fb94fa4-2aff-4636-ac1b-ed39dc9451a6"
	o740Crawler  = "3850479f-1c1e-4a72-ae47-6da2997202e2"
	o740Licid    = "4f484927-0e8e-493d-93bc-ed190965dad9"
	o740Elite    = "bd9f2c84-0589-4ffe-9bff-ea74be4f77a4"
	o740Berserk  = "c9046281-dedb-4246-8e5d-7786fe0c41c7"
	o740Ogre     = "fe455177-d9e5-4893-ad87-ebbba5e2fa45"
)

func TestS82OldB740EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{o740Palisade, o740Crawler, o740Licid, o740Elite, o740Berserk, o740Ogre} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestS82OldB740ElvenPalisadeShrinksAnAttackerForAForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pal := effectstest.B12Push(g, me.ID, "Elven Palisade", "Enchantment", o740Palisade, 0, 0)
	forest := effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	swamp := effectstest.B12Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", 0, 0)
	atk := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 5)
	idle := effectstest.B12Creature(g, opp.ID, "Idle Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(atk, me.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, pal, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(idle), SacrificeIDs: []uuid.UUID{forest}, Strict: true}); err == nil {
		t.Fatal("a creature that is not attacking was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, pal, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(atk), SacrificeIDs: []uuid.UUID{swamp}, Strict: true}); err == nil {
		t.Fatal("a Swamp paid for a Forest")
	}
	effectstest.ApaActivate(t, g, me, pal, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(atk), SacrificeIDs: []uuid.UUID{forest}})
	if p, tough := effectstest.PtOf(t, g, atk); p != 2 || tough != 5 {
		t.Errorf("the attacker is %d/%d, want 2/5", p, tough)
	}
	if effectstest.FindBattlefieldCardForTest(g, forest) != nil {
		t.Error("the Forest was not sacrificed")
	}
}

func TestS82OldB740CinderCrawlerGrowsOnlyOnceBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crawler := effectstest.B12Push(g, me.ID, "Cinder Crawler", "Creature — Salamander", o740Crawler, 1, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	effectstest.FloatForTest(g, me, "R")
	if err := g.ActivateCatalogAbility(me.ID, crawler, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated while not blocked")
	}
	effectstest.UnblockedAttackIntoBlockers(t, g, crawler)
	if err := g.DeclareBlocker(wall, crawler); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.FloatForTest(g, me, "R")
	effectstest.ApaActivate(t, g, me, crawler, 0, game.ActivateAbilityParams{})
	if p, _ := effectstest.PtOf(t, g, crawler); p != 2 {
		t.Errorf("the Crawler has power %d, want 2", p)
	}
}

func TestS82OldB740SkyshroudEliteNeedsAnOpponentsNonbasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elite := effectstest.B12Push(g, me.ID, "Skyshroud Elite", "Creature — Elf", o740Elite, 1, 1)
	effectstest.B12Push(g, me.ID, "My Nonbasic", "Land", "", 0, 0)
	effectstest.B12Push(g, opp.ID, "Their Forest", "Basic Land — Forest", "", 0, 0)
	if p, tough := effectstest.PtOf(t, g, elite); p != 1 || tough != 1 {
		t.Fatalf("the Elite is %d/%d with no opposing nonbasic land, want 1/1", p, tough)
	}
	effectstest.B12Push(g, opp.ID, "Their Nonbasic", "Land", "", 0, 0)
	if p, tough := effectstest.PtOf(t, g, elite); p != 2 || tough != 3 {
		t.Errorf("the Elite is %d/%d with an opposing nonbasic land, want 2/3", p, tough)
	}
}

func TestS82OldB740ElvishBerserkerGrowsPerBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	berserker := effectstest.B12Push(g, me.ID, "Elvish Berserker", "Creature — Elf Berserker", o740Berserk, 1, 1)
	a := effectstest.B12Creature(g, opp.ID, "Blocker A", "Creature — Wall", 0, 9)
	b := effectstest.B12Creature(g, opp.ID, "Blocker B", "Creature — Wall", 0, 9)
	effectstest.UnblockedAttackIntoBlockers(t, g, berserker)
	for _, blk := range []uuid.UUID{a, b} {
		if err := g.DeclareBlocker(blk, berserker); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, berserker); p != 3 || tough != 3 {
		t.Errorf("the Berserker is %d/%d, want 3/3 (+1/+1 for each of two blockers)", p, tough)
	}
}

func TestS82OldB740RecklessOgreGetsPowerOnlyWhenAttackingAlone(t *testing.T) {
	for _, alone := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		ogre := effectstest.B12Push(g, me.ID, "Reckless Ogre", "Creature — Ogre", o740Ogre, 3, 2)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
		if err := g.DeclareAttacker(ogre, opp.ID); err != nil {
			t.Fatal(err)
		}
		if !alone {
			if err := g.DeclareAttacker(bear, opp.ID); err != nil {
				t.Fatal(err)
			}
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		want := 3
		if alone {
			want = 6
		}
		if p, _ := effectstest.PtOf(t, g, ogre); p != want {
			t.Errorf("alone=%v: the Ogre has power %d, want %d", alone, p, want)
		}
	}
}

func TestS82OldB740TransmogrifyingLicidPumpsAndMakesAnArtifactThenEnds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	licid := effectstest.B12Push(g, me.ID, "Transmogrifying Licid", "Artifact Creature — Licid", o740Licid, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "C")
	effectstest.ApaActivate(t, g, me, licid, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	l := effectstest.B12Card(t, g, licid)
	if !l.IsAura() || l.IsCreature() || l.IsArtifact() || l.AttachedTo.ID != bear {
		t.Fatalf("the licid is not a plain Aura on the bear (aura %v, creature %v, artifact %v)", l.IsAura(), l.IsCreature(), l.IsArtifact())
	}
	b := effectstest.B12Card(t, g, bear)
	if !b.IsArtifact() || !b.IsCreature() {
		t.Errorf("the bear is artifact %v creature %v, want both", b.IsArtifact(), b.IsCreature())
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the bear is %d/%d, want 3/3", p, tough)
	}
	effectstest.FloatForTest(g, me, "C")
	if err := g.PerformSpecialAction(me.ID, licid, game.SpecialActionEndEffect, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("end_effect: %v", err)
	}
	l = effectstest.B12Card(t, g, licid)
	if !l.IsCreature() || !l.IsArtifact() || l.IsAura() {
		t.Error("after ending, the licid is not an artifact creature again")
	}
	if b := effectstest.B12Card(t, g, bear); b.IsArtifact() {
		t.Error("after ending, the bear is still an artifact")
	}
}
