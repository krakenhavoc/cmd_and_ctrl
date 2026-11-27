package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b604_cards_test.go — Mirrodin, S82 OLD batch 604 (#5606). Each
// test plays the real catalog entry and asserts what its text says.

const (
	oldB604Stylus    = "75a7fdd4-046a-433b-b16f-a0424a05c904"
	oldB604Joiner    = "84244aa6-11be-4940-b102-e029f955fc9a"
	oldB604Mindserv  = "c0ea83ab-f62e-43ee-b7ad-ddbbc8916632"
	oldB604Tower     = "c383b28c-b319-49cd-acf1-4721a1301bbb"
	oldB604Pearl     = "d08cb402-2138-4e8e-99f4-4a62855dc285"
	oldB604Vermiculo = "88cadf97-4b24-481a-9136-96c017fd96e9"
	oldB604Familiar  = "8e11d1f7-fa98-4cd2-9912-081e9c9e37ee"
	oldB604Forger    = "c94332ac-982a-4c0f-85d4-43e950ac3069"
	oldB604Battlemst = "cf01312d-305a-4587-8f54-03ee471bdf8c"
	oldB604Hoverguar = "e1c49bdb-cfd1-4ebf-90f7-73bd5009f75e"
)

func TestOldB604CardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB604Stylus: "Tel-Jilad Stylus", oldB604Joiner: "Viridian Joiner", oldB604Mindserv: "Myr Mindservant",
		oldB604Tower: "Tower of Murmurs", oldB604Pearl: "Pearl Shard", oldB604Vermiculo: "Vermiculos",
		oldB604Familiar: "Neurok Familiar", oldB604Forger: "Wurmskin Forger", oldB604Battlemst: "Vulshok Battlemaster",
		oldB604Hoverguar: "Looming Hoverguard",
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

func TestOldB604TelJiladStylusPutsOnlyYourOwnPermanentOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	stylus := effectstest.B12Push(g, me.ID, "Tel-Jilad Stylus", "Artifact", oldB604Stylus, 0, 0)
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, stylus, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(theirs)}); err == nil {
		t.Fatal("a permanent I do not own was accepted as the target")
	}
	size := len(me.Library.Cards)
	if err := g.ActivateCatalogAbility(me.ID, stylus, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(mine)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || len(me.Library.Cards) != size+1 || me.Library.Cards[0].InstanceID != mine {
		t.Errorf("the Rock is not at the bottom of my library (library %d → %d)", size, len(me.Library.Cards))
	}
}

func TestOldB604ViridianJoinerMakesManaEqualToItsPower(t *testing.T) {
	for _, counters := range []int{0, 2} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		joiner := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Viridian Joiner", TypeLine: "Creature — Elf Druid",
			OracleID: oldB604Joiner, Power: 1, Toughness: 2, Owner: me.ID, Controller: me.ID, Counters: map[string]int{"+1/+1": counters}})
		if err := g.ActivateManaAbility(me.ID, joiner, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
		if n := len(me.ManaPool.Units()); n != 1+counters {
			t.Errorf("%d counters: %d mana, want %d", counters, n, 1+counters)
		}
		for _, u := range me.ManaPool.Units() {
			if u.Color != "G" {
				t.Errorf("the mana is %s, want G", u.Color)
			}
		}
	}
}

func TestOldB604MyrMindservantShufflesTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	servant := effectstest.B12Push(g, me.ID, "Myr Mindservant", "Artifact Creature — Myr", oldB604Mindserv, 1, 1)
	before := make([]uuid.UUID, 0, len(me.Library.Cards))
	for _, c := range me.Library.Cards {
		before = append(before, c.InstanceID)
	}
	effectstest.ApaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, servant, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	same := len(before) == len(me.Library.Cards)
	for i, c := range me.Library.Cards {
		if i < len(before) && c.InstanceID != before[i] {
			same = false
		}
	}
	if same {
		t.Error("the library is in the same order after the shuffle")
	}
	if !effectstest.TappedForTest(t, g, servant) {
		t.Error("the ability did not tap the Mindservant")
	}
}

func TestOldB604TowerOfMurmursMillsEightFromAnyPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	tower := effectstest.B12Push(g, me.ID, "Tower of Murmurs", "Artifact", oldB604Tower, 0, 0)
	size, grave := len(opp.Library.Cards), opp.Graveyard.Size()
	effectstest.ApaMana(me, "C", "C", "C", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, tower, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(opp.Library.Cards) != size-8 || opp.Graveyard.Size() != grave+8 {
		t.Errorf("library %d → %d, graveyard %d → %d; want eight milled", size, len(opp.Library.Cards), grave, opp.Graveyard.Size())
	}
}

func TestOldB604PearlShardPreventsTwoDamageForEitherCost(t *testing.T) {
	for _, tc := range []struct {
		name string
		idx  int
		mana []string
	}{
		{"{3}, {T}", 0, []string{"C", "C", "C"}},
		{"{W}, {T}", 1, []string{"W"}},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		shard := effectstest.B12Push(g, me.ID, "Pearl Shard", "Artifact", oldB604Pearl, 0, 0)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 4, 4)
		effectstest.ApaMana(me, tc.mana...)
		if err := g.ActivateCatalogAbility(me.ID, shard, tc.idx, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
			t.Fatalf("%s: activate: %v", tc.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		g.WithWriteLock(func() {
			if err := g.DealDamageToCreatureForEffect(shard, bear, 3); err != nil {
				t.Fatal(err)
			}
		})
		if got := effectstest.DamageMarkedOn(g, bear); got != 1 {
			t.Errorf("%s: %d damage marked, want 1 (3 dealt, 2 prevented)", tc.name, got)
		}
	}
}

func TestOldB604VermiculosGrowsForEveryArtifactThatEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	verm := effectstest.B12Push(g, me.ID, "Vermiculos", "Creature — Horror", oldB604Vermiculo, 1, 1)
	effectstest.CastCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, verm); p != 1 {
		t.Fatalf("a nonartifact growing it: power %d", p)
	}
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Golem", "Artifact Creature — Golem", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, verm); p != 9 || tough != 9 {
		t.Errorf("after two artifacts the Vermiculos is %d/%d, want 9/9", p, tough)
	}
	effectstest.DcCleanup(g)
	if p, _ := effectstest.EffectivePTOf(t, g, verm); p != 1 {
		t.Errorf("the bonus outlived the cleanup step (power %d)", p)
	}
}

func TestOldB604NeurokFamiliarSortsTheTopCard(t *testing.T) {
	for _, artifact := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		typeLine := "Instant"
		if artifact {
			typeLine = "Artifact"
		}
		top := effectstest.SeedLibraryTop(me, "Top Card", typeLine, "")
		effectstest.CastAndResolveCreature(t, g, "Neurok Familiar", "Creature — Bird", oldB604Familiar)
		effectstest.PassPriorityAroundTable(t, g)
		if artifact != effectstest.InHandOf(me, top) || artifact == me.Graveyard.Contains(top) {
			t.Errorf("artifact=%v: in hand %v, in graveyard %v", artifact, effectstest.InHandOf(me, top), me.Graveyard.Contains(top))
		}
	}
}

func TestOldB604WurmskinForgerDistributesThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elk := effectstest.B12Creature(g, opp.ID, "Elk", "Creature — Elk", 3, 3)
	effectstest.CastAndResolveCreature(t, g, "Wurmskin Forger", "Creature — Elf Warrior", oldB604Forger)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asked for no targets")
	}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, effectstest.CardRefs(bear, elk), map[uuid.UUID]int{bear: 1, elk: 1}); err == nil {
		t.Fatal("a split of 2 was accepted for three counters")
	}
	effectstest.B17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{bear: 2, elk: 1}, bear, elk)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the Bear has %d counters, want 2", got)
	}
	if got := effectstest.CounterCount(g, elk, game.CounterPlusOne); got != 1 {
		t.Errorf("the Elk has %d counters, want 1", got)
	}
}

func TestOldB604VulshokBattlemasterAttachesEveryEquipmentOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Sword", TypeLine: "Artifact — Equipment",
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear}})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Sword", TypeLine: "Artifact — Equipment",
		Owner: opp.ID, Controller: opp.ID})
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	master := effectstest.CastAndResolveCreature(t, g, "Vulshok Battlemaster", "Creature — Human Warrior", oldB604Battlemst)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		c, _ := effectstest.BattlefieldCard(g, id)
		if c.AttachedTo.ID != master {
			t.Errorf("%s is attached to %v, want the Battlemaster", c.Name, c.AttachedTo.ID)
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, theirs); c.Controller != opp.ID {
		t.Error("attaching changed the Equipment's controller")
	}
	if c, _ := effectstest.BattlefieldCard(g, rock); c.AttachedTo.ID != uuid.Nil {
		t.Error("a non-Equipment artifact was attached")
	}
	if !effectstest.HasEffectiveKeyword(t, g, master, "haste") {
		t.Error("no haste")
	}
}

func TestOldB604LoomingHoverguardTucksAnyArtifactOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, owner := g.Seats[0], g.Seats[1], g.Seats[2]
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Stolen Rock", TypeLine: "Artifact",
		Owner: owner.ID, Controller: opp.ID})
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastAndResolveCreature(t, g, "Looming Hoverguard", "Creature — Drone", oldB604Hoverguar)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil || effectstest.HasID(p.PickTargetCards, bear) || !effectstest.HasID(p.PickTargetCards, stolen) {
		t.Fatalf("pick offered %+v: only artifacts", p)
	}
	effectstest.PickCard(t, g, me.ID, stolen)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(stolen) {
		t.Fatal("the artifact is still on the battlefield")
	}
	if n := len(owner.Library.Cards); n == 0 || owner.Library.Cards[n-1].InstanceID != stolen {
		t.Error("the artifact is not on top of its owner's library")
	}
	if len(opp.Library.Cards) > 0 && opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID == stolen {
		t.Error("the artifact went to its controller's library")
	}
}
