package testsa

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const acesBaseballBatOracle = "3f47ec32-4ee9-4bc0-96f2-310a939cb973"

func TestAcesBaseballBatIsFull(t *testing.T) {
	spec, ok := effects.Lookup(acesBaseballBatOracle)
	if !ok || spec.Name != "Ace's Baseball Bat" {
		t.Fatalf("registered=%v name=%q", ok, spec.Name)
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("%v %q, want Full with no caveat", spec.Completeness, spec.Caveats)
	}
}

// batEquip equips the Bat to host with ability `idx` (0: equip legendary
// creature {1}, 1: equip {3}), resolving it.
func batEquip(t *testing.T, g *game.Game, bat, host uuid.UUID, idx int) error {
	t.Helper()
	me := g.Seats[0]
	effectstest.FillPool(me, 3)
	err := g.ActivateCatalogAbility(me.ID, bat, idx, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}},
		Strict:  true,
	})
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

// The cheap equip takes only a legendary creature; the plain one any.
// Equipped creature gets +3/+0.
func TestAcesBaseballBatEquipsAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	legend := effectstest.SeedLegendaryCreature(g, me.ID, "Ace")
	bat := effectstest.SeedEquipment(g, me.ID, "Ace's Baseball Bat", acesBaseballBatOracle)

	if err := batEquip(t, g, bat, bear, 0); err == nil {
		t.Fatal("equip legendary creature {1} accepted a non-legendary creature")
	}
	if err := batEquip(t, g, bat, legend, 0); err != nil {
		t.Fatalf("equip legendary creature {1}: %v", err)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, legend); p != 5 || tough != 2 {
		t.Errorf("equipped legend = %d/%d, want 5/2", p, tough)
	}
	if err := batEquip(t, g, bat, bear, 1); err != nil {
		t.Fatalf("equip {3}: %v", err)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 5 {
		t.Errorf("equipped bear power %d, want 5", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, legend); p != 2 {
		t.Errorf("the legend kept the pump after the Bat moved: power %d", p)
	}
}

// batAttack puts the Bat on a 2/2 of seat 0's, and the given blockers
// under seat 1, and declares the attack into the declare-blockers step.
func batAttack(t *testing.T, g *game.Game) (host uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	host = effectstest.SeedBear(g, me.ID)
	bat := effectstest.SeedEquipment(g, me.ID, "Ace's Baseball Bat", acesBaseballBatOracle)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(bat, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
	if effectstest.EffectiveAbilitiesContain(t, g, host, "first strike") {
		t.Fatal("first strike before the creature attacks")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(host, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, host, "first strike") {
		t.Error("the attacking equipped creature has no first strike")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	return host
}

// A Dalek that can block the attacking equipped creature must; a non-Dalek
// may block it too, and on its own satisfies nothing.
func TestAcesBaseballBatMustBeBlockedByADalek(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	dalek := effectstest.B12Creature(g, opp.ID, "Dalek Drone", "Artifact Creature — Dalek", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	host := batAttack(t, g)

	if err := g.DeclareBlocker(bear, host); err != nil {
		t.Fatalf("a non-Dalek blocking it: %v", err)
	}
	err := g.FinishBlocks(opp.ID)
	if !errors.Is(err, game.ErrIllegalBlock) {
		t.Fatalf("finishing without the Dalek: err = %v, want a block requirement refusal", err)
	}
	var br *game.BlockRefusedError
	if !errors.As(err, &br) {
		t.Fatalf("refusal is %T", err)
	}
	if got := br.Sentence(opp.ID); got != "Bear must be blocked by a Dalek if able (Ace's Baseball Bat)." {
		t.Errorf("sentence = %q", got)
	}
	if err := g.DeclareBlocker(dalek, host); err != nil {
		t.Fatalf("the Dalek joins: %v", err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("FinishBlocks with the Dalek blocking: %v", err)
	}
}

// No Dalek: nothing is owed, and not blocking at all is legal.
func TestAcesBaseballBatWithoutADalekAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.B12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	batAttack(t, g)
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("no Dalek, no block: %v", err)
	}
}

// The requirement is only while attacking: an unattacking equipped creature
// carries none.
func TestAcesBaseballBatAsksNothingOfANonAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.SeedBear(g, me.ID)
	bat := effectstest.SeedEquipment(g, me.ID, "Ace's Baseball Bat", acesBaseballBatOracle)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(bat, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
	var reqs []game.BlockRequirement
	g.ReadSnapshot(func() {
		reqs = effectstest.FindBattlefieldCardForTest(g, host).Effective().BlockRequirements
	})
	if len(reqs) != 0 {
		t.Errorf("requirements %+v on a creature that is not attacking", reqs)
	}
}
