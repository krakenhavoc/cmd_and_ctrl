package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b684_test.go — Mercadian Masques, "Creatures with keywords", batch
// 684 (#5619). Every test plays the real catalog entry through the engine.

const (
	o684Glider    = "6ae3a901-9f9a-4a99-94db-22c38fda7aef"
	o684Rider     = "84750c9a-ddf7-4ee1-b93b-3dc9ab73e4d5"
	o684Dragonfly = "8e3e5264-d472-41ce-ad65-a53f8b24fc40"
	o684Nightwind = "9a642db9-c337-481c-b260-11e01f64e68e"
	o684Hounds    = "b50c2f63-750c-4cfc-8a74-5a7a3d9b46c4"
	o684Harpy     = "c8dcb8d8-ae27-43ac-88e7-e59f2656e0ce"
	o684Thermal   = "cbe5c74b-1006-4c63-8b46-e06b61e6fe02"
	o684Queen     = "e1a48505-acbe-4eb4-8e83-3c76cf7a8829"
)

func TestS82OldB684EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{o684Glider, o684Rider, o684Dragonfly, o684Nightwind, o684Hounds, o684Harpy, o684Thermal, o684Queen} {
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

func TestS82OldB684PrintedKeywordsAreOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
		keywords               []string
	}{
		{"Kyren Glider", "Creature — Goblin", o684Glider, 1, 1, []string{"flying"}},
		{"Jhovall Rider", "Creature — Human Rebel", o684Rider, 3, 3, []string{"trample"}},
		{"Venomous Dragonfly", "Creature — Insect", o684Dragonfly, 1, 1, []string{"flying"}},
		{"Nightwind Glider", "Creature — Human Rebel", o684Nightwind, 2, 1, []string{"flying", "protection from black"}},
		{"Lightning Hounds", "Creature — Dog", o684Hounds, 3, 2, []string{"first strike"}},
		{"Molting Harpy", "Creature — Harpy Mercenary", o684Harpy, 2, 1, []string{"flying"}},
		{"Thermal Glider", "Creature — Human Rebel", o684Thermal, 2, 1, []string{"flying", "protection from red"}},
		{"Jhovall Queen", "Creature — Cat Rebel", o684Queen, 4, 7, []string{"vigilance"}},
	} {
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
		for _, kw := range tc.keywords {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("%s does not have %s", tc.name, kw)
			}
		}
	}
}

func TestS82OldB684KyrenGliderCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	glider := effectstest.B12Push(g, opp.ID, "Kyren Glider", "Creature — Goblin", o684Glider, 1, 1)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird", Power: 2, Toughness: 2,
		Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID})
	effectstest.AttackIntoBlocks(t, g, opp.ID, flier)
	if reason := effectstest.BlockRefusal(t, g, glider, flier); reason == "" {
		t.Error("Kyren Glider blocked")
	}
}

func TestS82OldB684JhovallQueenAttacksWithoutTapping(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	queen := effectstest.B12Push(g, me.ID, "Jhovall Queen", "Creature — Cat Rebel", o684Queen, 4, 7)
	effectstest.AttackWith(t, g, opp.ID, queen)
	if c, _ := effectstest.BattlefieldCard(g, queen); c.Tapped {
		t.Error("the attacking Queen tapped")
	}
}

func TestS82OldB684LightningHoundsFirstStrikeKillsTheBlockerFirst(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hounds := effectstest.B12Push(g, me.ID, "Lightning Hounds", "Creature — Dog", o684Hounds, 3, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Ogre", 3, 3)
	effectstest.UnblockedAttackIntoBlockers(t, g, hounds)
	if err := g.DeclareBlocker(blocker, hounds); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if effectstest.FindBattlefieldCardForTest(g, blocker) != nil {
		t.Error("the blocker survived first-strike damage")
	}
	if effectstest.FindBattlefieldCardForTest(g, hounds) == nil {
		t.Error("the Hounds died, so they did not deal damage first")
	}
}

func TestS82OldB684VenomousDragonflyDestroysWhoeverBlocksOrIsBlockedByIt(t *testing.T) {
	// Blocked: the Dragonfly attacks, a flier blocks it.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fly := effectstest.B12Push(g, me.ID, "Venomous Dragonfly", "Creature — Insect", o684Dragonfly, 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == fly {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
			}
		}
	})
	blocker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Flier", TypeLine: "Creature — Drake", Power: 1, Toughness: 9,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.UnblockedAttackIntoBlockers(t, g, fly)
	if err := g.DeclareBlocker(blocker, fly); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, blocker) != nil {
		t.Error("the creature that blocked the Dragonfly survived end of combat")
	}

	// Blocking: the Dragonfly blocks a ground attacker.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	attacker := effectstest.B12Creature(g2, me2.ID, "Wall of Bricks", "Creature — Wall", 1, 9)
	fly2 := effectstest.B12Push(g2, opp2.ID, "Venomous Dragonfly", "Creature — Insect", o684Dragonfly, 1, 1)
	effectstest.UnblockedAttackIntoBlockers(t, g2, attacker)
	if err := g2.DeclareBlocker(fly2, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g2)
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AdvanceTo(t, g2, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.FindBattlefieldCardForTest(g2, attacker) != nil {
		t.Error("the creature the Dragonfly blocked survived end of combat")
	}
}

func TestS82OldB684ProtectionFromAColorPreventsThatColorsDamage(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, color string
	}{
		{"Nightwind Glider", o684Nightwind, "B"},
		{"Thermal Glider", o684Thermal, "R"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			glider := effectstest.B12Push(g, me.ID, tc.name, "Creature — Human Rebel", tc.oracle, 2, 5)
			same := effectstest.SeedColoredCreature(g, opp.ID, "Same", []string{tc.color}, 0)
			other := effectstest.SeedColoredCreature(g, opp.ID, "Other", []string{"G"}, 0)
			effectstest.Cmdr2B08Damage(g, same, glider, 2)
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.DamageMarkedOn(g, glider); n != 0 {
				t.Errorf("%d damage marked from a %s source, want 0", n, tc.color)
			}
			effectstest.Cmdr2B08Damage(g, other, glider, 2)
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.DamageMarkedOn(g, glider); n != 2 {
				t.Errorf("%d damage marked from a green source, want 2", n)
			}
		})
	}
}

func o684HarpyUpkeep(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	harpy := effectstest.B12Push(g, me.ID, "Molting Harpy", "Creature — Harpy Mercenary", o684Harpy, 2, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	return g, me, harpy
}

func TestS82OldB684MoltingHarpyStaysWhenYouPayTwo(t *testing.T) {
	g, me, harpy := o684HarpyUpkeep(t)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no pay prompt for the upkeep trigger")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(harpy) {
		t.Error("the Harpy should stay after paying {2}")
	}
}

func TestS82OldB684MoltingHarpyIsSacrificedWhenYouDontPay(t *testing.T) {
	g, me, harpy := o684HarpyUpkeep(t)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		effectstest.AnswerPayUnless(t, g, me.ID, false)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(harpy) {
		t.Error("the Harpy stayed without payment")
	}
}
