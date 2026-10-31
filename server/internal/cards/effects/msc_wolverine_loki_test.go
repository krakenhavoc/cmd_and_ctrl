package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// msc_wolverine_loki_test.go — Wolverine, Claws Out and Lady Loki, Agent of
// Chaos (#4455). Each test plays the real catalog entry through the engine.

const (
	mscWlWolverineOracle = "7100fe0c-1c2a-461a-83b3-684d44279081"
	mscWlLokiOracle      = "6efae65b-5ae5-4f62-97e7-cd4d1249fa9e"
	mscWlWolverineLine   = "Legendary Creature — Mutant Berserker Hero"
)

func TestMscWlEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mscWlWolverineOracle: "Wolverine, Claws Out",
		mscWlLokiOracle:      "Lady Loki, Agent of Chaos",
	} {
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

// --- Wolverine, Claws Out --------------------------------------------------

func mscWlWolverine(g *game.Game) uuid.UUID {
	return apaPush(g, g.Seats[0].ID, g.Seats[0].ID, game.Card{
		Name: "Wolverine, Claws Out", TypeLine: mscWlWolverineLine, OracleID: mscWlWolverineOracle,
		Power: 2, Toughness: 5,
	})
}

func mscWlBlocker(g *game.Game, name string, p, tough int) uuid.UUID {
	return apaPush(g, g.Seats[1].ID, g.Seats[1].ID, game.Card{
		Name: name, TypeLine: "Creature — Test", Power: p, Toughness: tough,
	})
}

// mscWlBlockedCombat attacks seat 1 with the attackers, has each blocker
// block its attacker, and walks to the combat damage step.
func mscWlBlockedCombat(t *testing.T, g *game.Game, attackers []uuid.UUID, blocks map[uuid.UUID]uuid.UUID) {
	t.Helper()
	pr7bAttack(t, g, g.Seats[1].ID, attackers...)
	pr7bBlock(t, g, blocks)
	pr7bDamageStep(t, g)
}

func mscWlDamagePrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	return pr7bPrompt(t, g, game.PendingChoiceDamageAssignment)
}

func mscWlAnswer(t *testing.T, g *game.Game, c *game.PendingChoice, toPlayer int, shares ...game.DamageAssignmentEntry) error {
	t.Helper()
	return g.ResolveDamageAssignment(c.ID, g.Seats[0].ID, shares, toPlayer)
}

func TestMscWlWolverineDoublesAMutantThatAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wolverine := mscWlWolverine(g)
	other := apaPush(g, me.ID, me.ID, game.Card{Name: "Other Mutant", TypeLine: "Creature — Mutant", Power: 3, Toughness: 3})
	human := apaPush(g, me.ID, me.ID, game.Card{Name: "Plain Human", TypeLine: "Creature — Human", Power: 3, Toughness: 3})

	pr7bAttack(t, g, g.Seats[2].ID, wolverine, other, human)
	pr7bDamageStep(t, g)

	if lost := 40 - g.Seats[2].Life; lost != 13 {
		t.Errorf("the defender lost %d, want 13 (Wolverine 4 + the other Mutant 6 + the Human 3, undoubled)", lost)
	}
}

func TestMscWlBlockedWolverineMayAssignAsThoughUnblocked(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	wolverine := mscWlWolverine(g)
	wall := mscWlBlocker(g, "Wall", 1, 3)
	mscWlBlockedCombat(t, g, []uuid.UUID{wolverine}, map[uuid.UUID]uuid.UUID{wall: wolverine})

	c := mscWlDamagePrompt(t, g)
	f := c.DamageAssignment
	if !f.AsThoughUnblocked || len(f.BlockerIDs) != 1 || f.AttackerPower != 4 {
		t.Fatalf("the prompt is %+v, want one blocker, power 4 (doubled), as-though-unblocked", f)
	}
	if opp.Life != 40 {
		t.Fatalf("damage was dealt before the answer: %d", opp.Life)
	}
	// Mixed, or short of all of it, is not "as though it weren't blocked".
	if err := mscWlAnswer(t, g, c, 3, game.DamageAssignmentEntry{BlockerID: wall, Amount: 1}); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("part to the player and part to the blocker: %v, want ErrInvalidParam", err)
	}
	if err := mscWlAnswer(t, g, c, 2, game.DamageAssignmentEntry{BlockerID: wall, Amount: 0}); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("only some of it to the player: %v, want ErrInvalidParam", err)
	}
	if err := mscWlAnswer(t, g, c, 4, game.DamageAssignmentEntry{BlockerID: wall, Amount: 0}); err != nil {
		t.Fatalf("all of it to the player: %v", err)
	}
	if opp.Life != 36 {
		t.Errorf("the defending player has %d life, want 36", opp.Life)
	}
	if got := pr6Marked(g, wall); got != 0 {
		t.Errorf("the blocker was dealt %d damage, want none", got)
	}
	if got := pr6Marked(g, wolverine); got != 1 {
		t.Errorf("the blocker's damage back to Wolverine is %d, want 1 (blocked all the same)", got)
	}
}

func TestMscWlBlockedWolverineMayStillAssignNormally(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	wolverine := mscWlWolverine(g)
	wall := mscWlBlocker(g, "Wall", 1, 3)
	mscWlBlockedCombat(t, g, []uuid.UUID{wolverine}, map[uuid.UUID]uuid.UUID{wall: wolverine})

	c := mscWlDamagePrompt(t, g)
	if err := mscWlAnswer(t, g, c, 0, game.DamageAssignmentEntry{BlockerID: wall, Amount: 4}); err != nil {
		t.Fatalf("all of it to the blocker: %v", err)
	}
	if opp.Life != 40 {
		t.Errorf("the defending player lost life (%d) though Wolverine assigned to the blocker", opp.Life)
	}
	if findBattlefieldCardForTest(g, wall) != nil {
		t.Error("the blocker survived 4 damage on toughness 3")
	}
}

func TestMscWlWolverineBlockedByTwoMayAssignAsThoughUnblocked(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	wolverine := mscWlWolverine(g)
	a := mscWlBlocker(g, "Wall A", 0, 3)
	b := mscWlBlocker(g, "Wall B", 0, 3)
	mscWlBlockedCombat(t, g, []uuid.UUID{wolverine}, map[uuid.UUID]uuid.UUID{a: wolverine, b: wolverine})

	c := mscWlDamagePrompt(t, g)
	if err := mscWlAnswer(t, g, c, 4,
		game.DamageAssignmentEntry{BlockerID: a}, game.DamageAssignmentEntry{BlockerID: b}); err != nil {
		t.Fatalf("all of it to the player past two blockers: %v", err)
	}
	if opp.Life != 36 || pr6Marked(g, a) != 0 || pr6Marked(g, b) != 0 {
		t.Errorf("life %d, blockers marked %d and %d: want 36, 0, 0", opp.Life, pr6Marked(g, a), pr6Marked(g, b))
	}
}

func TestMscWlBlockedWolverineWithNoBlockerLeftMayStillHitThePlayer(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		wolverine := mscWlWolverine(g)
		wall := mscWlBlocker(g, "Wall", 0, 3)
		pr7bAttack(t, g, opp.ID, wolverine)
		pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{wall: wolverine})
		g.WithWriteLock(func() {
			if err := g.ExileCardForEffect(wall); err != nil {
				t.Fatalf("exile the blocker: %v", err)
			}
		})
		pr7bDamageStep(t, g)

		c := mscWlDamagePrompt(t, g)
		if len(c.DamageAssignment.BlockerIDs) != 0 {
			t.Fatalf("blockers left: %v", c.DamageAssignment.BlockerIDs)
		}
		toPlayer := 0
		if yes {
			toPlayer = 4
		}
		if err := mscWlAnswer(t, g, c, toPlayer); err != nil {
			t.Fatalf("answer %v: %v", yes, err)
		}
		want := 40
		if yes {
			want = 36
		}
		if opp.Life != want {
			t.Errorf("assign-to-player %v: the defender has %d life, want %d", yes, opp.Life, want)
		}
	}
}

func TestMscWlWolverineThatLostItsAbilitiesAssignsToTheBlocker(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	wolverine := mscWlWolverine(g)
	wall := mscWlBlocker(g, "Wall", 1, 3)
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(wolverine),
			[]game.Mod{game.LoseAllAbilitiesMod()}, g.UntilEndOfTurnDuration(), "test — loses all abilities")
	})
	mscWlBlockedCombat(t, g, []uuid.UUID{wolverine}, map[uuid.UUID]uuid.UUID{wall: wolverine})

	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceDamageAssignment {
			t.Fatal("a Wolverine without his abilities is asked where to assign his damage")
		}
	}
	if got := pr6Marked(g, wall); got != 2 {
		t.Errorf("the blocker was dealt %d, want Wolverine's 2 (no doubling, no option)", got)
	}
	if opp.Life != 40 {
		t.Errorf("the defending player lost life: %d", opp.Life)
	}
}

func TestMscWlOtherCreaturesAreNeverOfferedTheOption(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mscWlWolverine(g)
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 4, Toughness: 4})
	a := mscWlBlocker(g, "Wall A", 0, 3)
	b := mscWlBlocker(g, "Wall B", 0, 3)
	mscWlBlockedCombat(t, g, []uuid.UUID{bear}, map[uuid.UUID]uuid.UUID{a: bear, b: bear})
	c := mscWlDamagePrompt(t, g)
	if c.DamageAssignment.AsThoughUnblocked {
		t.Error("an ordinary multi-blocked attacker is offered the option")
	}
	if err := mscWlAnswer(t, g, c, 4,
		game.DamageAssignmentEntry{BlockerID: a}, game.DamageAssignmentEntry{BlockerID: b}); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("an ordinary attacker put damage on the player past its blockers: %v", err)
	}
}

func TestMscWlWolverineSingleBlockerAnswersTheEnumeratorAndTheBotOffer(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	wolverine := mscWlWolverine(g)
	wall := mscWlBlocker(g, "Wall", 1, 3)
	mscWlBlockedCombat(t, g, []uuid.UUID{wolverine}, map[uuid.UUID]uuid.UUID{wall: wolverine})
	c := mscWlDamagePrompt(t, g)

	var offers []legal.OfferedDamageSplit
	g.WithWriteLock(func() { offers = legal.DamageSplits(g, c) })
	if len(offers) != 2 {
		t.Fatalf("%d offers, want the canonical split and the as-though-unblocked one: %+v", len(offers), offers)
	}
	if offers[0].Split.TrampleTo != 0 || offers[1].Split.TrampleTo != 4 {
		t.Errorf("offers %+v: want the canonical (kill the blocker) first, then all 4 to the player", offers)
	}
	// Both are answers the engine takes.
	for i, o := range offers {
		g2 := newCatalogGame(t)
		w2 := mscWlWolverine(g2)
		wall2 := mscWlBlocker(g2, "Wall", 1, 3)
		mscWlBlockedCombat(t, g2, []uuid.UUID{w2}, map[uuid.UUID]uuid.UUID{wall2: w2})
		c2 := mscWlDamagePrompt(t, g2)
		var shares []game.DamageAssignmentEntry
		for _, a := range o.Split.Assignments {
			// the new table has new ids: the single blocker is wall2.
			shares = append(shares, game.DamageAssignmentEntry{BlockerID: wall2, Amount: a.Amount})
		}
		if err := g2.ResolveDamageAssignment(c2.ID, g2.Seats[0].ID, shares, o.Split.TrampleTo); err != nil {
			t.Errorf("offer %d (%s) was refused: %v", i, o.Label, err)
		}
	}
	_, _ = wall, opp
}

func TestMscWlWolverineCommanderDamageGoesToThePlayerAsThoughUnblocked(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	wolverine := mscWlWolverine(g)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, wolverine).IsCommander = true })
	wall := mscWlBlocker(g, "Wall", 0, 3)
	mscWlBlockedCombat(t, g, []uuid.UUID{wolverine}, map[uuid.UUID]uuid.UUID{wall: wolverine})
	c := mscWlDamagePrompt(t, g)
	if err := mscWlAnswer(t, g, c, 4, game.DamageAssignmentEntry{BlockerID: wall}); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if opp.Life != 36 {
		t.Errorf("life %d, want 36", opp.Life)
	}
	if got := opp.CommanderDamage[wolverine]; got != 4 {
		t.Errorf("commander damage from Wolverine is %d, want 4", got)
	}
}
