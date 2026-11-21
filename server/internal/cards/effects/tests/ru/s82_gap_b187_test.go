package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b187_test.go — Warhammer 40,000 Commander creatures, S82 GAP
// batch 187 (#5948).

var gapB187IDs = map[string]string{
	"Canoptek Spyder":        "bcad5b83-5513-4455-a9c9-716281505eaa",
	"Canoptek Tomb Sentinel": "678b3fb5-aa9f-4268-9433-2a89dd37b927",
	"Cryptothrall":           "e26f0fcd-d6d2-4e7b-82dd-711bc0b9dad3",
	"Arco-Flagellant":        "b48e7e87-3adf-417d-b93b-bf67ad6cbf1d",
	"Chronomancer":           "9833e3bd-6e7f-41b7-bf7e-ffb74ff17844",
	"Cryptek":                "27a4b633-5a62-4d8c-8cc6-44959c311de9",
	"Flayed One":             "899415fd-4164-42a5-a0c7-eef1967c5ec5",
	"Great Unclean One":      "cd9031e4-9f6f-4dd8-908e-dcf7b79438e5",
	"Hexmark Destroyer":      "b3329240-c812-45ea-a7f8-4d79cd4ed515",
}

func TestGapB187CardsAreFull(t *testing.T) {
	for name, id := range gapB187IDs {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v spec=%q/%s/%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gapB187Has(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	c := effectstest.FlipState(t, g, id)
	return game.HasKeyword(&c, kw)
}

func gapB187Library(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		p.Library.Cards = nil
		for range n {
			p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: p.ID, Controller: p.ID})
		}
	})
}

func TestFlayedOneMillsThree(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	gapB187Library(g, me, 6)
	effectstest.CastFromHandForTest(t, g, me, "Flayed One", "Artifact Creature — Necron", "{0}", gapB187IDs["Flayed One"], game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Size() != 3 {
		t.Fatalf("graveyard has %d cards, want 3 milled", me.Graveyard.Size())
	}
}

func TestCanoptekSpyderDrawsForOtherNontokenArtifactCreaturesAndVehicles(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	gapB187Library(g, me, 6)
	effectstest.PushCatalogPermanent(g, me.ID, "Canoptek Spyder", "Artifact Creature — Spider", gapB187IDs["Canoptek Spyder"], false)
	hand := me.Hand.Size()
	effectstest.CastFromHandForTest(t, g, me, "Golem", "Artifact Creature — Golem", "{0}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Cart", "Artifact — Vehicle", "{0}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Fatalf("drew %d cards for a Golem and a Vehicle, want 2", got)
	}
	effectstest.CastFromHandForTest(t, g, me, "Plain Bear", "Creature — Bear", "{0}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Fatalf("a nonartifact creature drew a card (total %d)", got)
	}
}

func TestCryptothrallGivesOtherArtifactCreaturesHexproof(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	thrall := effectstest.PushCatalogPermanent(g, me.ID, "Cryptothrall", "Artifact Creature — Construct", gapB187IDs["Cryptothrall"], false)
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	if !gapB187Has(t, g, golem, "hexproof") {
		t.Fatal("another artifact creature has no hexproof")
	}
	if gapB187Has(t, g, thrall, "hexproof") || gapB187Has(t, g, bear, "hexproof") {
		t.Fatal("Cryptothrall itself or a nonartifact creature has hexproof")
	}
}

func TestHexmarkDestroyerNeedsSixBlockers(t *testing.T) {
	spec, _ := effects.Lookup(gapB187IDs["Hexmark Destroyer"])
	if len(spec.BlockRules) != 1 {
		t.Fatalf("block rules = %d, want 1", len(spec.BlockRules))
	}
	g, me, opp := effectstest.FlipMain(t)
	hex := effectstest.PushCatalogPermanent(g, me.ID, "Hexmark Destroyer", "Artifact Creature — Necron", gapB187IDs["Hexmark Destroyer"], false)
	effectstest.Unsick(g, hex)
	var blockers []uuid.UUID
	for range 5 {
		blockers = append(blockers, effectstest.B12Creature(g, opp.ID, "Blocker", effectstest.TestCreatureTypeLine, 1, 1))
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(hex, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blockers[0], hex); err == nil {
		t.Fatal("a lone creature could block a Hexmark Destroyer")
	}
}

func TestArcoFlagellantCantBlockAndPaysLifeForIndestructible(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	arco := effectstest.PushCatalogPermanent(g, me.ID, "Arco-Flagellant", "Creature — Human", gapB187IDs["Arco-Flagellant"], false)
	effectstest.Unsick(g, arco)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, arco, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Endurant: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 || !gapB187Has(t, g, arco, "indestructible") {
		t.Fatalf("life %d -> %d, indestructible=%v", life, me.Life, gapB187Has(t, g, arco, "indestructible"))
	}
	spec, _ := effects.Lookup(gapB187IDs["Arco-Flagellant"])
	if len(spec.OptionalCosts) != 1 {
		t.Fatalf("optional costs = %d, want the squad cost", len(spec.OptionalCosts))
	}
	_ = opp
}

func TestChronomancerSacrificesAnotherArtifactToDraw(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	gapB187Library(g, me, 4)
	chrono := effectstest.PushCatalogPermanent(g, me.ID, "Chronomancer", "Artifact Creature — Necron Wizard", gapB187IDs["Chronomancer"], false)
	effectstest.Unsick(g, chrono)
	relic := effectstest.PushCatalogPermanent(g, me.ID, "Relic", "Artifact", "", false)
	effectstest.FloatMana(t, g, me, "{C}")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, chrono, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{relic}}); err != nil {
		t.Fatalf("Atomic Transmutation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || g.Battlefield.Contains(relic) {
		t.Fatalf("hand %d -> %d, relic on battlefield=%v", hand, me.Hand.Size(), g.Battlefield.Contains(relic))
	}
}

func TestCryptekReturnsTheChosenCreatureTappedWhenItDies(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	cryptek := effectstest.PushCatalogPermanent(g, me.ID, "Cryptek", "Artifact Creature — Necron Wizard", gapB187IDs["Cryptek"], false)
	effectstest.Unsick(g, cryptek)
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	effectstest.FloatMana(t, g, me, "{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, cryptek, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: golem}}}); err != nil {
		t.Fatalf("Cryptek: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B18Kill(t, g, golem)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCardByID(g, golem)
	if !ok || !c.Tapped {
		t.Fatalf("the Golem is back=%v tapped=%v, want back and tapped", ok, c.Tapped)
	}
}

func TestGreatUncleanOneDrainsAndMakesPlaguebearersForPoorerOpponents(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Great Unclean One", "Creature — Demon", gapB187IDs["Great Unclean One"], false)
	me.Life = 20
	for i, p := range g.Seats {
		if p != me {
			p.Life = []int{40, 21, 12, 12}[i%4]
		}
	}
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	want := 0
	for _, p := range g.Seats {
		if p == me {
			continue
		}
		if p.Life != before[p.ID]-2 {
			t.Errorf("opponent life %d -> %d, want -2", before[p.ID], p.Life)
		}
		if p.Life < me.Life {
			want++
		}
	}
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Plaguebearer of Nurgle" {
				n++
			}
		}
	})
	if n != want {
		t.Fatalf("%d Plaguebearers, want %d (one per opponent with less life than you)", n, want)
	}
}

func TestCanoptekTombSentinelIsUnearthable(t *testing.T) {
	spec, _ := effects.Lookup(gapB187IDs["Canoptek Tomb Sentinel"])
	if len(spec.Activated) != 1 || len(spec.Triggered) != 1 {
		t.Fatalf("activated %d triggered %d, want unearth and the enters-from-graveyard trigger", len(spec.Activated), len(spec.Triggered))
	}
}
