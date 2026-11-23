package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flanking_grant_test.go — #5784. Flanking's trigger is the engine's,
// derived from the token on the layered ability list (game/flanking.go):
// a granted flanking brings its trigger, a lost one takes it away, and a
// printed one triggers once per instance (CR 702.25b), never twice.

const (
	flankingTelimTorOracle = "b13a723d-80a9-4116-9857-baf3b373f87a" // Telim'Tor
	flankingAgilityOracle  = "0f5f8b47-d104-4e29-b052-02d730d76ec0" // Agility
	flankingFoliageOracle  = "f9275c99-0616-44b3-a19e-98c683e87ad3" // Barbed Foliage
)

func TestFlankingMirageCardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		flankingTelimTorOracle: "Telim'Tor",
		flankingAgilityOracle:  "Agility",
		flankingFoliageOracle:  "Barbed Foliage",
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

// flankingWithAgility casts a real Agility on `creature` and resolves it.
func flankingWithAgility(t *testing.T, g *game.Game, creature uuid.UUID) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Agility", "Enchantment — Aura", flankingAgilityOracle, effectstest.CardRefs(creature))
	effectstest.PassPriorityAroundTable(t, g)
	if !game.HasKeyword(effectstest.ApaLive(g, creature), game.KeywordFlanking) {
		t.Fatal("the enchanted creature does not have flanking")
	}
}

// The issue's first case: a Grizzly Bears enchanted with Agility is 3/3
// with flanking, and its trigger shrinks a blocker without flanking once.
func TestAgilityGivesFlankingThatShrinksABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bears := effectstest.ApaPush(g, me.ID, me.ID, effectstest.FlankingBears(t))
	flankingWithAgility(t, g, bears)
	if c := effectstest.ApaLive(g, bears); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Fatalf("the enchanted Bears is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	big := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Big Blocker", "", 2, 5))
	effectstest.FlankingBlocks(t, g, opp.ID, bears, big)
	if got := effectstest.ApaLive(g, big).CurrentToughness(); got != 4 {
		t.Errorf("a 2/5 blocking the Bears with Agility has toughness %d, want 4 (-1/-1 once)", got)
	}
}

// Agility on a creature that prints flanking is a second instance (CR
// 702.25b), and so a second trigger.
func TestAgilityOnAPrintedFlankerIsASecondInstance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crusader := effectstest.ApaPush(g, me.ID, me.ID, flankingCrusader(t))
	flankingWithAgility(t, g, crusader)
	if got := game.FlankingCount(effectstest.ApaLive(g, crusader)); got != 2 {
		t.Fatalf("the Crusader with Agility has %d instances of flanking, want 2", got)
	}
	big := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Big Blocker", "", 2, 5))
	effectstest.FlankingBlocks(t, g, opp.ID, crusader, big)
	if got := effectstest.ApaLive(g, big).CurrentToughness(); got != 3 {
		t.Errorf("a 2/5 blocker has toughness %d, want 3 (-1/-1 twice)", got)
	}
}

// A real Zhalfirin Crusader, which carries the retired row beside its
// token, triggers exactly once: the row never triggers, so the token is
// the only trigger.
func TestPrintedFlankingTriggersOncePerInstance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crusader := effectstest.ApaPush(g, me.ID, me.ID, flankingCrusader(t))
	if got := game.FlankingCount(effectstest.ApaLive(g, crusader)); got != 1 {
		t.Fatalf("the imported Crusader has %d instances of flanking, want 1", got)
	}
	big := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Big Blocker", "", 2, 5))
	effectstest.FlankingBlocks(t, g, opp.ID, crusader, big)
	if got := effectstest.ApaLive(g, big).CurrentToughness(); got != 4 {
		t.Errorf("a 2/5 blocker has toughness %d, want 4 (-1/-1 once)", got)
	}
}

// A card the catalog does not know gets the keyword from the importer,
// and now its trigger with it.
func TestUncataloguedFlankerTriggers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := effectstest.ApaPush(g, me.ID, me.ID, effectstest.FlankingImport(t, "", "Uncatalogued Knight", "Creature — Human Knight", "{1}{W}", "2", "2", "Flanking"))
	bears := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.FlankingBears(t))
	effectstest.FlankingBlocks(t, g, opp.ID, knight, bears)
	if got := effectstest.ApaLive(g, bears).CurrentToughness(); got != 1 {
		t.Errorf("the Bears blocking an uncatalogued flanker has toughness %d, want 1", got)
	}
}

// A flanking creature that loses flanking has no trigger: here a direct
// layer-6 removal (CR 613.1f).
func TestFlankingLostToARemovalHasNoTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crusader := effectstest.ApaPush(g, me.ID, me.ID, flankingCrusader(t))
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(crusader),
			[]game.Mod{game.RemoveKeywordsMod(game.KeywordFlanking)}, g.UntilEndOfTurnDuration(),
			"test — the Crusader loses flanking")
	})
	if game.HasKeyword(effectstest.ApaLive(g, crusader), game.KeywordFlanking) {
		t.Fatal("the Crusader lost flanking and still has it")
	}
	bears := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.FlankingBears(t))
	effectstest.FlankingBlocks(t, g, opp.ID, crusader, bears)
	if got := effectstest.ApaLive(g, bears).CurrentToughness(); got != 2 {
		t.Errorf("the Bears blocking a Crusader without flanking has toughness %d, want 2", got)
	}
}

// The issue's second case: Barbed Foliage takes flanking from a creature
// that attacks its controller, so the blocker is not shrunk, and deals
// it 1 damage because it has no flying. That holds for a granted
// flanking too: the removal's timestamp is later than Agility's.
func TestBarbedFoliageTakesFlankingAndItsTrigger(t *testing.T) {
	for _, tc := range []struct {
		name    string
		agility bool
	}{{"a printed flanker", false}, {"a printed flanker with Agility", true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Barbed Foliage", TypeLine: "Enchantment", OracleID: flankingFoliageOracle})
			crusader := effectstest.ApaPush(g, me.ID, me.ID, flankingCrusader(t))
			if tc.agility {
				flankingWithAgility(t, g, crusader)
			}
			big := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Big Blocker", "", 2, 5))
			// Both of the Foliage's triggers fire on the one attack, so
			// its controller orders them (CR 603.3b) before the blocks.
			effectstest.DeclareAttack(t, g, opp.ID, crusader)
			effectstest.B35SettleTriggerOrders(t, g, opp.ID)
			effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
			if err := g.DeclareBlocker(big, crusader); err != nil {
				t.Fatal(err)
			}
			if !effectstest.FinishBlockDeclarations(t, g) {
				t.Fatal("no block declaration was left to finish")
			}
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.ApaLive(g, crusader)
			if game.HasKeyword(c, game.KeywordFlanking) {
				t.Error("the Crusader attacked Barbed Foliage's controller and still has flanking")
			}
			if c.DamageMarked != 1 {
				t.Errorf("the Crusader has %d damage marked, want 1", c.DamageMarked)
			}
			if got := effectstest.ApaLive(g, big).CurrentToughness(); got != 5 {
				t.Errorf("the 2/5 blocker has toughness %d, want 5: the attacker lost flanking", got)
			}
		})
	}
}

// Barbed Foliage watches attacks on its controller only, and its damage
// skips a creature with flying.
func TestBarbedFoliageReadsWhoIsAttackedAndFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Barbed Foliage", TypeLine: "Enchantment", OracleID: flankingFoliageOracle})
	crusader := effectstest.ApaPush(g, me.ID, me.ID, flankingCrusader(t))
	flier := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Flier", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(crusader, third.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(flier, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ApaLive(g, crusader)
	if !game.HasKeyword(c, game.KeywordFlanking) || c.DamageMarked != 0 {
		t.Errorf("the Crusader attacked another player: flanking %v, damage %d; want true, 0",
			game.HasKeyword(c, game.KeywordFlanking), c.DamageMarked)
	}
	if f := effectstest.ApaLive(g, flier); f == nil || f.DamageMarked != 0 {
		t.Error("the flier attacking Barbed Foliage's controller was dealt damage")
	}
}

// Telim'Tor: "all attacking creatures with flanking get +1/+1 until end of
// turn", read as it resolves. Telim'Tor and a Bears given flanking by
// Agility get it; an attacker without flanking, and a flanker that stayed
// home, do not.
func TestTelimTorPumpsAttackingCreaturesWithFlanking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	telim := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Telim'Tor", flankingTelimTorOracle, 2, 2))
	if !game.HasKeyword(effectstest.ApaLive(g, telim), game.KeywordFlanking) {
		t.Fatal("Telim'Tor does not have flanking")
	}
	enchanted := effectstest.ApaPush(g, me.ID, me.ID, effectstest.FlankingBears(t))
	flankingWithAgility(t, g, enchanted)
	plain := effectstest.ApaPush(g, me.ID, me.ID, effectstest.FlankingBears(t))
	home := effectstest.ApaPush(g, me.ID, me.ID, flankingCrusader(t))
	effectstest.DeclareAttack(t, g, opp.ID, telim, enchanted, plain)
	effectstest.PassPriorityAroundTable(t, g)
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"Telim'Tor", telim, 3}, {"the Bears with Agility", enchanted, 4}, {"the plain Bears", plain, 2}, {"the Crusader at home", home, 2}} {
		if got := effectstest.ApaLive(g, tc.id).CurrentPower(); got != tc.want {
			t.Errorf("%s has power %d, want %d", tc.name, got, tc.want)
		}
	}
}
