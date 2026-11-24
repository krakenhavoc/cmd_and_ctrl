package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b12_sea_devils_test.go — S82 GAP2 batch 12: The Sea Devils,
// through the real chapter, delayed-trigger and targeting paths.

const gap2B12SeaDevils = "09d7a2d5-9e6e-4d84-8f97-c36b11ccc6fb"

func TestGap2B12SeaDevilsIsFull(t *testing.T) {
	spec, ok := effects.Lookup(gap2B12SeaDevils)
	if !ok {
		t.Fatal("The Sea Devils is not registered")
	}
	if spec.Name != "The Sea Devils" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

func TestGap2B12SeaDevilsChapterOneMakesASalamander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "The Sea Devils", "Enchantment — Saga", gap2B12SeaDevils, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Alien Salamander" && c.Controller == me.ID {
			n++
			eff := c.Effective()
			if eff.Power != 2 || eff.Toughness != 2 || !c.HasSubtype("Salamander") || !game.HasKeyword(&c, "islandwalk") {
				t.Errorf("token %+v", eff)
			}
		}
	}
	if n != 1 {
		t.Errorf("%d Alien Salamanders, want 1", n)
	}
}

func TestGap2B12SeaDevilsChapterThreeTurnsSalamanderHitsIntoRemoval(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Sea Devils", TypeLine: "Enchantment — Saga", OracleID: gap2B12SeaDevils,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: 2},
	})
	mySal := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Salamander Pal", TypeLine: "Creature — Salamander", Power: 3, Toughness: 3})
	theirSal := effectstest.PushTestPermanent(g, opp1.ID, game.Card{Name: "Their Salamander", TypeLine: "Creature — Salamander Warrior", Power: 2, Toughness: 2})
	bear := effectstest.PushVanillaCreature(g, me.ID, "Not a Salamander", 2, 2)
	victim := effectstest.PushVanillaCreature(g, opp2.ID, "Victim", 3, 3)
	bystander := effectstest.PushVanillaCreature(g, opp1.ID, "Bystander", 5, 5)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 4, 4)

	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()

	// A non-Salamander's hit does nothing.
	effectstest.DealCombatDamageToPlayer(g, bear, opp2.ID, 2)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatalf("a non-Salamander's combat damage triggered: %+v", p)
	}
	// My Salamander hits the second opponent: only their creatures are on offer.
	effectstest.DealCombatDamageToPlayer(g, mySal, opp2.ID, 3)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatalf("no target prompt after a Salamander hit a player: %+v", g.PendingChoices)
	}
	if effectstest.HasID(prompt.PickTargetCards, bystander) || effectstest.HasID(prompt.PickTargetCards, mine) {
		t.Error("a creature the damaged player does not control was offered")
	}
	if !effectstest.HasID(prompt.PickTargetCards, victim) {
		t.Fatal("the damaged player's creature was not offered")
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(victim) {
		t.Error("3 damage from the Salamander did not kill the 3/3")
	}

	// An opponent's Salamander counts too; I choose the target among my own creatures.
	effectstest.DealCombatDamageToPlayer(g, theirSal, me.ID, 2)
	prompt = effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatalf("an opponent's Salamander did not trigger it: %+v", g.PendingChoices)
	}
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := g.LookupCardForEffect(mine); !ok || c.DamageMarked != 2 {
		t.Errorf("Mine has %d damage, want 2", c.DamageMarked)
	}

	// It lasts until end of turn only.
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	effectstest.DealCombatDamageToPlayer(g, mySal, opp2.ID, 3)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Error("the trigger outlived the turn")
	}
}
