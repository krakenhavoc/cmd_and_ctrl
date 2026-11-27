package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b569_creatures_test.go — the four Guildpact haunt creatures of OLD
// batch 569 (#5599) that were not yet on develop. It reuses the settle and pick
// helpers of s82_old_b569_haunt_test.go.

const (
	oldB569DominatorOracle = "5462d711-037b-4c62-9238-c557a66534f8"
	oldB569AbsolverOracle  = "a9002245-0f33-48cb-96d3-fde884236eb5"
	oldB569BelfryOracle    = "c171065f-7947-4161-a481-1ea2d3ce8422"
	oldB569ExhumerOracle   = "cd2ae326-1a74-4b67-9718-92b6a209303b"
)

func TestOldB569CreaturesAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB569DominatorOracle: "Graven Dominator",
		oldB569AbsolverOracle:  "Absolver Thrull",
		oldB569BelfryOracle:    "Belfry Spirit",
		oldB569ExhumerOracle:   "Exhumer Thrull",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

// oldB569HauntIt kills the haunt creature and has it haunt victim.
func oldB569HauntIt(t *testing.T, g *game.Game, owner, haunter, victim uuid.UUID) {
	t.Helper()
	oldB569Kill(t, g, owner, haunter)
	oldB569Settle(t, g, oldB569Card(victim))
	if h, ok := oldB569Haunting(g, haunter); !ok || h != victim {
		t.Fatalf("haunting: in exile %v, haunts %v; want %v", ok, h, victim)
	}
}

func TestOldB569GravenDominatorMakesEveryOtherCreatureOneOneOnEntryAndOnTheHauntedDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.P1g2Seats(g)
	mine := effectstest.DftABPush(g, me.ID, "My Ogre", "Creature — Ogre", "", 4, 4)
	theirs := effectstest.DftABPush(g, opp.ID, "Opp Wurm", "Creature — Wurm", "", 5, 5)
	effectstest.AdvanceToMain(t, g)
	dom := effectstest.DftABCast(t, g, me, game.Card{Name: "Graven Dominator", TypeLine: "Creature — Gargoyle", OracleID: oldB569DominatorOracle,
		ManaCost: "{4}{W}{W}", Power: 4, Toughness: 4})
	oldB569Settle(t, g, oldB569NoPick)
	for _, id := range []uuid.UUID{mine, theirs} {
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != 1 || tough != 1 {
			t.Errorf("on entry: creature is %d/%d, want 1/1", p, tough)
		}
	}
	if p, tough := effectstest.EffectivePTOf(t, g, dom); p != 4 || tough != 4 {
		t.Errorf("the Dominator itself is %d/%d, want 4/4", p, tough)
	}

	// Dies, haunts a creature, and the haunted creature's death shrinks the rest.
	oldB569HauntIt(t, g, me.ID, dom, theirs)
	other := effectstest.DftABPush(g, opp.ID, "Opp Elk", "Creature — Elk", "", 3, 3)
	oldB569Kill(t, g, opp.ID, theirs)
	oldB569Settle(t, g, oldB569NoPick)
	if p, tough := effectstest.EffectivePTOf(t, g, other); p != 1 || tough != 1 {
		t.Errorf("after the haunted creature died: elk is %d/%d, want 1/1", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 1 || tough != 1 {
		t.Errorf("after the haunted creature died: my ogre is %d/%d, want 1/1", p, tough)
	}
}

func TestOldB569AbsolverThrullDestroysAnEnchantmentOnEntryAndOnTheHauntedDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.P1g2Seats(g)
	e1 := effectstest.DftABPush(g, opp.ID, "Opp Aura One", "Enchantment", "", 0, 0)
	e2 := effectstest.DftABPush(g, opp.ID, "Opp Aura Two", "Enchantment", "", 0, 0)
	bear := effectstest.DftABPush(g, opp.ID, "Opp Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	thrull := effectstest.DftABCast(t, g, me, game.Card{Name: "Absolver Thrull", TypeLine: "Creature — Thrull Cleric", OracleID: oldB569AbsolverOracle,
		ManaCost: "{3}{W}", Power: 2, Toughness: 3})
	oldB569Settle(t, g, oldB569Card(e1))
	if onBattlefield(g, e1) {
		t.Error("the entry trigger left its enchantment")
	}
	// The bear is not an enchantment: not a legal choice.
	oldB569HauntIt(t, g, me.ID, thrull, bear)
	oldB569Kill(t, g, opp.ID, bear)
	oldB569Settle(t, g, oldB569Card(e2))
	if onBattlefield(g, e2) {
		t.Error("the haunt trigger left its enchantment")
	}
}

func onBattlefield(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				found = true
			}
		}
	})
	return found
}

func TestOldB569BelfryTwoBatsOnEntryAndTwoOnTheHauntedDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.P1g2Seats(g)
	bear := effectstest.DftABPush(g, opp.ID, "Opp Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	spirit := effectstest.DftABCast(t, g, me, game.Card{Name: "Belfry Spirit", TypeLine: "Creature — Spirit", OracleID: oldB569BelfryOracle,
		ManaCost: "{3}{W}{W}", Power: 1, Toughness: 1})
	oldB569Settle(t, g, oldB569NoPick)
	if n := oldB569Bats(g, me.ID); n != 2 {
		t.Fatalf("on entry: %d Bats, want 2", n)
	}
	oldB569HauntIt(t, g, me.ID, spirit, bear)
	oldB569Kill(t, g, opp.ID, bear)
	oldB569Settle(t, g, oldB569NoPick)
	if n := oldB569Bats(g, me.ID); n != 4 {
		t.Errorf("after the haunted creature died: %d Bats, want 4", n)
	}
}

func oldB569Bats(g *game.Game, controller uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && c.Name == "Bat" && c.IsToken() {
				n++
			}
		}
	})
	return n
}

func TestOldB569ExhumerThrullReturnsACreatureCardOnEntryAndOnTheHauntedDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.P1g2Seats(g)
	c1 := effectstest.Batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	c2 := effectstest.Batch01GraveyardCard(me, "Dead Elk", "Creature — Elk")
	land := effectstest.Batch01GraveyardCard(me, "Dead Land", "Land")
	bear := effectstest.DftABPush(g, opp.ID, "Opp Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	thrull := effectstest.DftABCast(t, g, me, game.Card{Name: "Exhumer Thrull", TypeLine: "Creature — Thrull", OracleID: oldB569ExhumerOracle,
		ManaCost: "{5}{B}", Power: 3, Toughness: 3})
	// A land card is not a legal choice.
	if c := effectstest.FirstPendingTarget(g); c != nil && effectstest.HasID(c.PickTargetCards, land) {
		t.Fatal("a land card was offered")
	}
	oldB569Settle(t, g, oldB569Card(c1))
	if !me.Hand.Contains(c1) || me.Graveyard.Contains(c1) {
		t.Fatal("on entry: the creature card did not return to hand")
	}
	oldB569HauntIt(t, g, me.ID, thrull, bear)
	oldB569Kill(t, g, opp.ID, bear)
	oldB569Settle(t, g, oldB569Card(c2))
	if !me.Hand.Contains(c2) {
		t.Error("after the haunted creature died: the second creature card did not return")
	}
	if me.Hand.Contains(land) {
		t.Error("the land card returned")
	}
}
