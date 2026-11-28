package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b664_cards_test.go — Invasion creatures, S82 OLD batch 664 (#5616).
// Every test plays the real catalog entry through the engine.

const (
	b664Reaper = "245c3681-fc47-47fe-b6e5-34c96315c135"
	b664Kavu   = "285c9f49-4aeb-468c-bb41-f5ccb5257e2b"
	b664Runner = "2c90b7fa-d5bf-4c46-92d5-3014ea1bec50"
	b664Jade   = "495e08a8-e654-4a0b-a25f-1e344ae2a0fa"
	b664Spy    = "5f1e8aa2-6b85-4e2e-92e4-876470447d6f"
	b664Alab   = "8a8b032e-90b7-4782-b854-b575f316d641"
	b664Puma   = "e28e1921-797b-4e6e-8058-3929af46eeb5"
)

func TestB664EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b664Reaper: "Phyrexian Reaper", b664Kavu: "Yavimaya Kavu", b664Runner: "Kavu Runner", b664Jade: "Jade Leech",
		b664Spy: "Goblin Spy", b664Alab: "Alabaster Leech", b664Puma: "Scarred Puma",
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

func b664Creature(g *game.Game, owner uuid.UUID, name, cost, oracle string, colors []string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", ManaCost: cost, OracleID: oracle,
		Colors: colors, Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

// --- Scarred Puma ----------------------------------------------------------

func TestB664ScarredPumaAttacksOnlyBesideABlackOrGreenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	puma := b664Creature(g, me.ID, "Scarred Puma", "{R}", b664Puma, []string{"R"}, 2, 1)
	red := b664Creature(g, me.ID, "Red Bear", "{1}{R}", "", []string{"R"}, 2, 2)
	green := b664Creature(g, me.ID, "Green Bear", "{1}{G}", "", []string{"G"}, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(red, opp.ID); err != nil {
		t.Fatalf("the red creature was refused: %v", err)
	}
	if err := g.DeclareAttacker(puma, opp.ID); err != nil {
		t.Fatalf("staging the Puma with a green creature available: %v", err)
	}
	// The red creature is no company; the Puma needs the green one.
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackLimit) {
		t.Fatalf("passing with the Puma and only a red companion: %v, want ErrAttackLimit", err)
	}
	if err := g.DeclareAttacker(green, opp.ID); err != nil {
		t.Fatalf("the green companion was refused: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, puma); c == nil || c.AttackingTarget == uuid.Nil {
		t.Error("the Puma is not attacking beside the green creature")
	}
}

func TestB664ScarredPumaCantAttackWithOnlyAWhiteCompanion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	puma := b664Creature(g, g.Seats[0].ID, "Scarred Puma", "{R}", b664Puma, []string{"R"}, 2, 1)
	b664Creature(g, g.Seats[0].ID, "White Knight", "{W}{W}", "", []string{"W"}, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(puma, opp.ID); !errors.Is(err, game.ErrAttackLimit) {
		t.Fatalf("declaring the Puma with only a white companion: %v, want ErrAttackLimit", err)
	}
}

// --- Yavimaya Kavu ---------------------------------------------------------

func TestB664YavimayaKavuCountsRedCreaturesForPowerAndGreenForToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kavu := b664Creature(g, me.ID, "Yavimaya Kavu", "{2}{R}{G}", b664Kavu, []string{"R", "G"}, 0, 0)
	g.BumpLayerVersionForTest()
	if p, tough := effectstest.PtOf(t, g, kavu); p != 1 || tough != 1 {
		t.Fatalf("alone the Kavu is %d/%d, want 1/1 (it is red and green)", p, tough)
	}
	b664Creature(g, opp.ID, "Their Red", "{R}", "", []string{"R"}, 1, 1)
	b664Creature(g, me.ID, "Another Red", "{R}", "", []string{"R"}, 1, 1)
	b664Creature(g, opp.ID, "Their Green", "{G}", "", []string{"G"}, 1, 1)
	b664Creature(g, opp.ID, "Blue", "{U}", "", []string{"U"}, 1, 1)
	// A noncreature red permanent doesn't count.
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Red Rock", TypeLine: "Artifact", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
	})
	g.BumpLayerVersionForTest()
	if p, tough := effectstest.PtOf(t, g, kavu); p != 3 || tough != 2 {
		t.Errorf("the Kavu is %d/%d, want 3/2 (three red creatures, two green)", p, tough)
	}
}

// --- Kavu Runner -----------------------------------------------------------

func TestB664KavuRunnerHasHasteOnlyWhileNoOpponentControlsAWhiteOrBlueCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	runner := b664Creature(g, me.ID, "Kavu Runner", "{3}{R}", b664Runner, []string{"R"}, 3, 3)
	g.BumpLayerVersionForTest()
	if !effectstest.HasKeywordOnBattlefield(t, g, runner, "haste") {
		t.Fatal("no haste with no opposing white or blue creature")
	}
	// Your own white creature, an opposing red one and an opposing white
	// noncreature don't turn it off.
	b664Creature(g, me.ID, "My White", "{W}", "", []string{"W"}, 1, 1)
	b664Creature(g, opp.ID, "Their Red", "{R}", "", []string{"R"}, 1, 1)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "White Rock", TypeLine: "Artifact", Colors: []string{"W"}, Owner: opp.ID, Controller: opp.ID,
	})
	g.BumpLayerVersionForTest()
	if !effectstest.HasKeywordOnBattlefield(t, g, runner, "haste") {
		t.Fatal("lost haste to something that isn't an opposing white or blue creature")
	}
	blue := b664Creature(g, opp.ID, "Their Blue", "{U}", "", []string{"U"}, 1, 1)
	g.BumpLayerVersionForTest()
	if effectstest.HasKeywordOnBattlefield(t, g, runner, "haste") {
		t.Fatal("kept haste beside an opposing blue creature")
	}
	g.WithWriteLock(func() { _, _ = g.Battlefield.Remove(blue) })
	b664Creature(g, opp.ID, "Their White", "{W}", "", []string{"W"}, 1, 1)
	g.BumpLayerVersionForTest()
	if effectstest.HasKeywordOnBattlefield(t, g, runner, "haste") {
		t.Fatal("kept haste beside an opposing white creature")
	}
}

// --- Goblin Spy ------------------------------------------------------------

func TestB664GoblinSpyRevealsTheTopCard(t *testing.T) {
	spec, ok := effects.Lookup(b664Spy)
	if !ok || spec.LibraryTopVisible != game.LibraryTopRevealed {
		t.Errorf("LibraryTopVisible = %v, want revealed", spec.LibraryTopVisible)
	}
	if len(spec.CastPermissions) != 0 {
		t.Error("the Spy gives a permission to play the top card")
	}
}

// --- Jade Leech and Alabaster Leech ----------------------------------------

// b664TryCast puts the Leech on the battlefield of `leechSeat` in a fresh game,
// then tries to cast the spell from the first seat's hand strictly with exactly
// the mana given.
func b664TryCast(t *testing.T, leechOracle, leechName string, leechSeat int, name, cost, colorLetter string, mana ...string) error {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, g.Seats[leechSeat].ID, leechName, "Creature — Leech", leechOracle, false)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Test", ManaCost: cost,
			Colors: []string{colorLetter}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		me.ManaPool = nil
		for _, c := range mana {
			me.ManaPool.AddMana(game.ManaToken{Color: c})
		}
	})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
}

func TestB664JadeLeechTaxesOnlyYourGreenSpells(t *testing.T) {
	cast := func(name, cost, color string, mana ...string) error {
		return b664TryCast(t, b664Jade, "Jade Leech", 0, name, cost, color, mana...)
	}
	if err := cast("Green Bear", "{1}{G}", "G", "G", "C"); err == nil {
		t.Fatal("cast a green spell without paying the extra {G}")
	}
	if err := cast("Green Bear", "{1}{G}", "G", "G", "C", "C"); err == nil {
		t.Fatal("paid the extra {G} with colorless mana")
	}
	if err := cast("Green Bear", "{1}{G}", "G", "G", "G", "C"); err != nil {
		t.Fatalf("cast a green spell paying the extra {G}: %v", err)
	}
	if err := cast("Red Bear", "{1}{R}", "R", "R", "C"); err != nil {
		t.Fatalf("a red spell was taxed: %v", err)
	}
}

func TestB664JadeLeechDoesNotTaxAnOpponentsSpells(t *testing.T) {
	if err := b664TryCast(t, b664Jade, "Jade Leech", 1, "Green Bear", "{1}{G}", "G", "G", "C"); err != nil {
		t.Fatalf("an opponent's Leech taxed my green spell: %v", err)
	}
}

func TestB664AlabasterLeechTaxesOnlyYourWhiteSpells(t *testing.T) {
	cast := func(name, cost, color string, mana ...string) error {
		return b664TryCast(t, b664Alab, "Alabaster Leech", 0, name, cost, color, mana...)
	}
	if err := cast("White Knight", "{W}", "W", "W"); err == nil {
		t.Fatal("cast a white spell without paying the extra {W}")
	}
	if err := cast("White Knight", "{W}", "W", "W", "W"); err != nil {
		t.Fatalf("cast a white spell paying the extra {W}: %v", err)
	}
	if err := cast("Blue Bear", "{U}", "U", "U"); err != nil {
		t.Fatalf("a blue spell was taxed: %v", err)
	}
}

// --- Phyrexian Reaper ------------------------------------------------------

func b664ReaperCombat(t *testing.T, blockerColors []string, regenerating bool) (*game.Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reaper := b664Creature(g, me.ID, "Phyrexian Reaper", "{4}{B}", b664Reaper, []string{"B"}, 3, 3)
	blocker := b664Creature(g, opp.ID, "Blocker", "{1}", "", blockerColors, 1, 5)
	effectstest.AdvanceToMain(t, g)
	if regenerating {
		g.WithWriteLock(func() { _ = g.RegenerateForEffect(blocker) })
	}
	effectstest.Pr7bAttack(t, g, opp.ID, reaper)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blocker: reaper})
	return g, reaper, blocker
}

func TestB664PhyrexianReaperDestroysAGreenBlocker(t *testing.T) {
	g, reaper, blocker := b664ReaperCombat(t, []string{"G"}, false)
	if g.Battlefield.Contains(blocker) {
		t.Error("the green blocker survived")
	}
	if !g.Battlefield.Contains(reaper) {
		t.Error("the Reaper died")
	}
}

func TestB664PhyrexianReaperIgnoresRegeneration(t *testing.T) {
	g, _, blocker := b664ReaperCombat(t, []string{"G"}, true)
	if g.Battlefield.Contains(blocker) {
		t.Error("the regeneration shield saved the green blocker")
	}
}

func TestB664PhyrexianReaperLeavesANongreenBlockerAlone(t *testing.T) {
	g, _, blocker := b664ReaperCombat(t, []string{"W"}, false)
	if !g.Battlefield.Contains(blocker) {
		t.Error("a white blocker was destroyed")
	}
}
