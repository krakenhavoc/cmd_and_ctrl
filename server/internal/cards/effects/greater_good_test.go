package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const greaterGoodOracle = "dc0593c2-ccb4-4648-a592-c5bcd121dc72"

// TestGreaterGoodDrawsForPowerThenDiscardsThree activates the
// ability over a 4-power creature and checks both halves land: four
// cards drawn, then exactly three discarded.
func TestGreaterGoodDrawsForPowerThenDiscardsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	greaterGood := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Greater Good", TypeLine: "Enchantment",
		OracleID: greaterGoodOracle, Owner: me.ID, Controller: me.ID,
	})
	beast := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Beast", TypeLine: "Creature — Beast",
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})

	before := me.Hand.Size()
	b16Activate(t, g, me.ID, greaterGood, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{beast}})
	if got := me.Hand.Size(); got != before+4 {
		t.Fatalf("draws four for a 4-power sacrifice: hand %d → %d", before, got)
	}
	if !me.Graveyard.Contains(beast) {
		t.Error("the sacrificed creature is in the graveyard")
	}
	c := discardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("Greater Good then asks the caster to discard three")
	}
	if c.ChooseMin != 3 || c.ChooseMax != 3 {
		t.Errorf("discard bounds = [%d,%d], want [3,3]", c.ChooseMin, c.ChooseMax)
	}
	discardFromHand(t, g, me.ID)
	if got := me.Hand.Size(); got != before+4-3 {
		t.Errorf("hand after the discard = %d, want %d", got, before+4-3)
	}
}

// pushGloriousAnthemFor puts a Glorious Anthem under `owner`'s control:
// creatures they control get +1/+1.
func pushGloriousAnthemFor(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment",
		OracleID: gloriousAnthemOracle, Owner: owner, Controller: owner,
	})
}

// TestGreaterGoodCountsAnAnthemInTheSacrificedPower — the power is the
// creature's last-known information, so an anthem's +1 counts beside a
// +1/+1 counter: a printed 2/2 with one counter under Glorious Anthem
// draws four.
func TestGreaterGoodCountsAnAnthemInTheSacrificedPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	greaterGood := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Greater Good", TypeLine: "Enchantment",
		OracleID: greaterGoodOracle, Owner: me.ID, Controller: me.ID,
	})
	pushGloriousAnthemFor(g, me.ID)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 1},
	})
	if got := effectivePower(t, g, bear); got != 3 {
		t.Fatalf("setup: the anthem makes the bear's layered power %d, want 3", got)
	}

	before := me.Hand.Size()
	b16Activate(t, g, me.ID, greaterGood, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	if got := me.Hand.Size(); got != before+4 {
		t.Errorf("draws four for a 2/2 with a +1/+1 counter under an anthem: hand %d → %d", before, got)
	}
}

// TestGreaterGoodDrawsNothingForAZeroPowerCreature — the sacrifice is
// legal and still pays out the discard, but a 0-power creature draws
// nothing at all (DrawCards{N: 0} is a no-op, not a one-card draw).
func TestGreaterGoodDrawsNothingForAZeroPowerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	greaterGood := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Greater Good", TypeLine: "Enchantment",
		OracleID: greaterGoodOracle, Owner: me.ID, Controller: me.ID,
	})
	nothing := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Zero", TypeLine: "Creature — Illusion",
		Power: 0, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})

	before := me.Hand.Size()
	b16Activate(t, g, me.ID, greaterGood, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{nothing}})
	if got := me.Hand.Size(); got != before {
		t.Fatalf("a 0-power sacrifice draws nothing: hand %d → %d", before, got)
	}
	c := discardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the discard still happens even when the draw is zero")
	}
	discardFromHand(t, g, me.ID)
	if got := me.Hand.Size(); got != before-3 {
		t.Errorf("hand after the discard = %d, want %d", got, before-3)
	}
}
