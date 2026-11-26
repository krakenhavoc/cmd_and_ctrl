package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 Dragons of Tarkir batch 330 (#5551): bolster.

const (
	dtk330Gleam    = "68a91b87-1845-419b-aae7-ee1237554946"
	dtk330Captain  = "74333718-2a05-4643-9a37-013df39b32e0"
	dtk330Victory  = "955ff3bf-4323-452e-a8b1-b7662034473d"
	dtk330Feast    = "9812bdf9-98b6-4b4d-8381-a7bbb25069d8"
	dtk330Echoes   = "9863a7da-6f51-4d3e-9ace-e8ff201faf13"
	dtk330Sunbring = "bbfced8f-0f83-4cc9-9143-aa2473828b48"
)

func TestDtk330CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dtk330Gleam: "Gleam of Authority", dtk330Captain: "Dromoka Captain", dtk330Victory: "Enduring Victory",
		dtk330Feast: "Pinion Feast", dtk330Echoes: "Echoes of the Kin Tree", dtk330Sunbring: "Sunbringer's Touch",
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

func TestDtk330DromokaCaptainBolstersTheLeastToughnessWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	captain := effectstest.PushCatalogPermanent(g, me.ID, "Dromoka Captain", "Creature — Human Soldier", dtk330Captain, false)
	effectstest.Unsick(g, captain)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(captain, game.CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
	})
	small := effectstest.PushVanillaCreature(g, me.ID, "Small", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, captain)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, small, game.CounterPlusOne); got != 1 {
		t.Errorf("the 1/1 has %d counters, want 1", got)
	}
	if got := effectstest.CountersOn(g, captain, game.CounterPlusOne); got != 1 {
		t.Errorf("the Captain has %d counters, want its original 1", got)
	}
}

func TestDtk330DromokaCaptainHasFirstStrikeAndDoesNotBolsterWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	captain := effectstest.PushCatalogPermanent(g, me.ID, "Dromoka Captain", "Creature — Human Soldier", dtk330Captain, false)
	if !effectstest.HasKeywordOf(g, captain, "first strike") {
		t.Error("the Captain lacks first strike")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, captain, game.CounterPlusOne); got != 0 {
		t.Errorf("the Captain has %d counters without attacking", got)
	}
}

func dtk330Cast(t *testing.T, g *game.Game, me *game.Player, name, oracle string, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	id := effectstest.B13HandCard(me, name, "Instant", oracle)
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets, Strict: true})
}

func TestDtk330EnduringVictoryDestroysAnAttackerThenBolsters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	effectstest.Unsick(g, attacker)
	small := effectstest.PushVanillaCreature(g, me.ID, "Small", 1, 1)
	idle := effectstest.PushVanillaCreature(g, opp.ID, "Idle", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	if _, err := dtk330Cast(t, g, me, "Enduring Victory", dtk330Victory, effectstest.CardRefs(idle)); err == nil {
		t.Fatal("Enduring Victory targeted a creature that is not attacking or blocking")
	}
	if _, err := dtk330Cast(t, g, me, "Enduring Victory", dtk330Victory, effectstest.CardRefs(attacker)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacker was not destroyed")
	}
	if got := effectstest.CountersOn(g, small, game.CounterPlusOne); got != 1 {
		t.Errorf("the 1/1 has %d counters, want 1", got)
	}
}

func TestDtk330PinionFeastDestroysAFlyerThenBolstersTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	small := effectstest.PushVanillaCreature(g, me.ID, "Small", 1, 1)
	ground := effectstest.PushVanillaCreature(g, opp.ID, "Walker", 2, 2)
	flyer := effectstest.PushCatalogPermanent(g, opp.ID, "Drake", "Creature — Drake", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flyer {
				g.Battlefield.Cards[i].Keywords = []string{"flying"}
			}
		}
	})
	if _, err := dtk330Cast(t, g, me, "Pinion Feast", dtk330Feast, effectstest.CardRefs(ground)); err == nil {
		t.Fatal("Pinion Feast targeted a creature without flying")
	}
	if _, err := dtk330Cast(t, g, me, "Pinion Feast", dtk330Feast, effectstest.CardRefs(flyer)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flyer) {
		t.Error("the flyer was not destroyed")
	}
	if got := effectstest.CountersOn(g, small, game.CounterPlusOne); got != 2 {
		t.Errorf("the 1/1 has %d counters, want 2", got)
	}
}

func TestDtk330EchoesOfTheKinTreeBolstersForItsCostOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	echoes := effectstest.PushCatalogPermanent(g, me.ID, "Echoes of the Kin Tree", "Enchantment", dtk330Echoes, false)
	big := effectstest.PushVanillaCreature(g, me.ID, "Big", 3, 3)
	small := effectstest.PushVanillaCreature(g, me.ID, "Small", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, echoes, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("Echoes activated without paying {2}{W}")
	}
	effectstest.FillPoolColored(me, "W", 3)
	if err := g.ActivateCatalogAbility(me.ID, echoes, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, small, game.CounterPlusOne); got != 1 {
		t.Errorf("the 1/1 has %d counters, want 1", got)
	}
	if got := effectstest.CountersOn(g, big, game.CounterPlusOne); got != 0 {
		t.Errorf("the 3/3 has %d counters, want 0", got)
	}
}

func TestDtk330SunbringersTouchBolstersByHandSizeAndGrantsTrampleToCounteredCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	countered := effectstest.PushVanillaCreature(g, me.ID, "Countered", 2, 2)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(countered, game.CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
	})
	small := effectstest.PushVanillaCreature(g, me.ID, "Small", 1, 1)
	plain := effectstest.PushVanillaCreature(g, me.ID, "Plain", 4, 4)
	id := effectstest.B13HandCard(me, "Sunbringer's Touch", "Sorcery", dtk330Sunbring)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	x := me.Hand.Size()
	if x == 0 {
		t.Fatal("the opening hand is empty; the test needs cards in hand")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, small, game.CounterPlusOne); got != x {
		t.Errorf("the 1/1 has %d counters, want %d (the cards in hand)", got, x)
	}
	if !effectstest.HasKeywordOf(g, small, "trample") || !effectstest.HasKeywordOf(g, countered, "trample") {
		t.Error("a creature with a +1/+1 counter did not gain trample")
	}
	if effectstest.HasKeywordOf(g, plain, "trample") {
		t.Error("a creature with no counter gained trample")
	}
}

func TestDtk330GleamOfAuthorityCountsCountersOnOtherCreaturesGrantsVigilanceAndBolster(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.UmbraCreature(g, me.ID, 2, 2)
	effectstest.Unsick(g, host)
	other := effectstest.UmbraCreature(g, me.ID, 1, 1)
	theirs := effectstest.UmbraCreature(g, opp.ID, 1, 1)
	g.WithWriteLock(func() {
		for _, add := range []struct {
			id uuid.UUID
			n  int
		}{{host, 5}, {other, 2}, {theirs, 4}} {
			if err := g.AddCounterForEffect(add.id, game.CounterPlusOne, add.n); err != nil {
				t.Fatal(err)
			}
		}
	})
	hp, ht := effectstest.EffectivePTOf(t, g, host)
	effectstest.AuraCast(t, g, "Gleam of Authority", dtk330Gleam, host)
	effectstest.AdvanceToMain(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, host); p != hp+2 || tough != ht+2 {
		t.Errorf("enchanted creature %d/%d, want %d/%d (+1/+1 per counter on my other creatures only)", p, tough, hp+2, ht+2)
	}
	if !effectstest.HasKeywordOf(g, host, "vigilance") {
		t.Error("the enchanted creature lacks vigilance")
	}
	idx, ref := effectstest.GrantedActivatedIndex(t, g, host)
	if idx < 0 {
		t.Fatal("the enchanted creature has no granted activated ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, host, idx, game.ActivateAbilityParams{Ref: ref, Strict: true}); err == nil {
		t.Fatal("the granted ability activated without {W}")
	}
	effectstest.FillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, host, idx, game.ActivateAbilityParams{Ref: ref, Strict: true}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The least toughness among my creatures is the other creature (3/3 against the host's 9/9).
	if got := effectstest.CountersOn(g, other, game.CounterPlusOne); got != 3 {
		t.Errorf("the other creature has %d counters, want 3", got)
	}
	if !effectstest.B13Tapped(t, g, host) {
		t.Error("the enchanted creature was not tapped by its own {T} cost")
	}
	// The new counter raises the pump by one: more counters on other creatures, a bigger host.
	if p, _ := effectstest.EffectivePTOf(t, g, host); p != hp+3 {
		t.Errorf("enchanted creature power %d, want %d after the bolster", p, hp+3)
	}
}
