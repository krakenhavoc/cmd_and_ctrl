package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b692_test.go — Mercadian Masques pump and combat tricks, S82 OLD
// batch 692 (#5619), played through the real catalog entries.

const (
	o692BriarPatch    = "247793af-be9e-47c7-8887-e99e49478b0b"
	o692Downfall      = "2c1201b2-5036-4e68-b36d-cdaf54b4993d"
	o692Gahr          = "3e7f67d6-db12-4034-b8e3-9c9d401c8f57"
	o692Indignation   = "97409c60-537e-4eeb-9d63-be0fde7e8284"
	o692Spidersilk    = "ad73e5eb-197c-45a3-9d20-a406ba126c79"
	o692DrakeHatch    = "bd55bcc4-f0c5-4c09-8b3b-2e25e9fbd9ee"
	o692CavernCrawler = "c7cfb120-1344-4b4e-9c6f-f969ba9c4a18"
	o692Wolverine     = "ca8060a7-30e7-49a4-a720-da5b995dd31f"
	o692CrenWall      = "ca9e41e8-7830-4721-b131-07d0b60297ad"
	o692CommonCause   = "d1dac139-de12-46ba-94b1-42e107c2c95f"
	o692Sustenance    = "d32a32d2-203d-4be1-8a33-e037747053c7"
	o692Boa           = "f0ab6ca6-098a-416a-89b2-9a93c9d0e60b"
)

func TestS82OldB692CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o692BriarPatch: "Briar Patch", o692Downfall: "Mercadia's Downfall", o692Gahr: "Snorting Gahr",
		o692Indignation: "Righteous Indignation", o692Spidersilk: "Spidersilk Armor", o692DrakeHatch: "Drake Hatchling",
		o692CavernCrawler: "Cavern Crawler", o692Wolverine: "Deepwood Wolverine", o692CrenWall: "Crenellated Wall",
		o692CommonCause: "Common Cause", o692Sustenance: "Sustenance", o692Boa: "Boa Constrictor",
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

func o692Colored(g *game.Game, owner uuid.UUID, name, typeLine, cost string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func o692PT(t *testing.T, g *game.Game, id uuid.UUID, p, tough int, what string) {
	t.Helper()
	if gp, gt := effectstest.EffectivePTOf(t, g, id); gp != p || gt != tough {
		t.Errorf("%s = %d/%d, want %d/%d", what, gp, gt, p, tough)
	}
}

func TestS82OldB692BriarPatchShrinksCreaturesThatAttackYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Briar Patch", "Enchantment", o692BriarPatch, false)
	toOpp := effectstest.PushVanillaCreature(g, me.ID, "At Briar Patch's controller", 3, 3)
	toThird := effectstest.PushVanillaCreature(g, me.ID, "At someone else", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(toOpp, opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(toThird, g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.InitSettle(t, g)
	o692PT(t, g, toOpp, 2, 3, "the creature attacking Briar Patch's controller")
	o692PT(t, g, toThird, 3, 3, "the creature attacking another player")
}

func TestS82OldB692MercadiasDownfallCountsTheDefendersNonbasicLandsPerAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B12Permanent(g, opp.ID, "Karoo", "Land")
	effectstest.B12Permanent(g, opp.ID, "Snow-Covered Forest", "Basic Snow Land — Forest")
	effectstest.B12Permanent(g, third.ID, "Cave", "Land")
	effectstest.B12Permanent(g, third.ID, "Cavern", "Land")
	effectstest.B12Permanent(g, third.ID, "Grotto", "Land")
	one := effectstest.PushVanillaCreature(g, me.ID, "One", 2, 2)
	three := effectstest.PushVanillaCreature(g, me.ID, "Three", 2, 2)
	idle := effectstest.PushVanillaCreature(g, me.ID, "Idle", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(one, opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(three, third.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.FloatForTest(g, me, "CCR")
	id := effectstest.HandCardOf(g, me, "Mercadia's Downfall", "Instant", "{2}{R}", o692Downfall)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	o692PT(t, g, one, 3, 2, "the creature attacking the player with one nonbasic land")
	o692PT(t, g, three, 5, 2, "the creature attacking the player with three nonbasic lands")
	o692PT(t, g, idle, 2, 2, "the creature that is not attacking")
}

func TestS82OldB692BlockedPumpCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		base, bonusP, bonusT   int
	}{
		{"Snorting Gahr", o692Gahr, "Creature — Rhino Beast", 3, 2, 2},
		{"Deepwood Wolverine", o692Wolverine, "Creature — Wolverine", 1, 2, 0},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		c := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.base, tc.base)
		blocker := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 20)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(c, opp.ID); err != nil {
			t.Fatal(err)
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if err := g.DeclareBlocker(blocker, c); err != nil {
			t.Fatal(err)
		}
		effectstest.LockInBlocks(t, g)
		effectstest.InitSettle(t, g)
		o692PT(t, g, c, tc.base+tc.bonusP, tc.base+tc.bonusT, tc.name+" after being blocked")
	}
}

func TestS82OldB692BlockedPumpCreatureStaysSmallWhenUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := effectstest.B12Push(g, me.ID, "Snorting Gahr", "Creature — Rhino Beast", o692Gahr, 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(c, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.LockInBlocks(t, g)
	effectstest.InitSettle(t, g)
	o692PT(t, g, c, 3, 3, "an unblocked Snorting Gahr")
}

func TestS82OldB692RighteousIndignationPumpsBlockersOfBlackOrRedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Righteous Indignation", "Enchantment", o692Indignation, false)
	black := o692Colored(g, me.ID, "Black Attacker", "Creature — Zombie", "{1}{B}", 1, 9)
	green := o692Colored(g, me.ID, "Green Attacker", "Creature — Bear", "{1}{G}", 1, 9)
	blackBlocker := effectstest.PushVanillaCreature(g, opp.ID, "Blocks Black", 1, 9)
	greenBlocker := effectstest.PushVanillaCreature(g, opp.ID, "Blocks Green", 1, 9)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range []uuid.UUID{black, green} {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blackBlocker, black); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareBlocker(greenBlocker, green); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.InitSettle(t, g)
	o692PT(t, g, blackBlocker, 2, 10, "the creature that blocked a black creature")
	o692PT(t, g, greenBlocker, 1, 9, "the creature that blocked a green creature")
}

func TestS82OldB692SpidersilkArmorGivesYourCreaturesReachAndToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Spidersilk Armor", "Enchantment", o692Spidersilk, false)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	o692PT(t, g, mine, 2, 3, "my creature")
	o692PT(t, g, theirs, 2, 2, "their creature")
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "reach") {
		t.Error("my creature lacks reach")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "reach") {
		t.Error("their creature has reach")
	}
	late := o692Colored(g, me.ID, "Late", "Creature — Bear", "", 1, 1)
	o692PT(t, g, late, 1, 2, "a creature that arrived later")
}

func TestS82OldB692DrakeHatchlingPumpsOnlyOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drake := effectstest.B12Push(g, me.ID, "Drake Hatchling", "Creature — Drake", o692DrakeHatch, 1, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "UU")
	if err := g.ActivateCatalogAbility(me.ID, drake, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	o692PT(t, g, drake, 2, 3, "the Hatchling after one activation")
	if err := g.ActivateCatalogAbility(me.ID, drake, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the second activation in one turn was allowed")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, drake, "flying") {
		t.Error("no flying")
	}
}

func TestS82OldB692CavernCrawlerTradesToughnessForPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crawler := effectstest.B12Push(g, me.ID, "Cavern Crawler", "Creature — Insect", o692CavernCrawler, 0, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RR")
	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, crawler, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	o692PT(t, g, crawler, 2, 1, "the Crawler after two activations")
	if !effectstest.HasKeywordOnBattlefield(t, g, crawler, "mountainwalk") {
		t.Error("no mountainwalk")
	}
	effectstest.FloatForTest(g, me, "R")
	if err := g.ActivateCatalogAbility(me.ID, crawler, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("third activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, crawler) != nil {
		t.Error("the Crawler at toughness 0 stayed on the battlefield")
	}
}

func TestS82OldB692BoaConstrictorTapsForPlusThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	boa := effectstest.B12Push(g, me.ID, "Boa Constrictor", "Creature — Snake", o692Boa, 3, 3)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, boa, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	o692PT(t, g, boa, 6, 6, "the Boa")
	if !effectstest.TappedOnBattlefield(t, g, boa) {
		t.Error("the {T} cost was not paid")
	}
}

func TestS82OldB692CrenellatedWallGivesAnyCreaturePlusFourToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.B12Push(g, me.ID, "Crenellated Wall", "Artifact Creature — Wall", o692CrenWall, 0, 4)
	other := effectstest.PushVanillaCreature(g, opp.ID, "Other", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, wall, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	o692PT(t, g, other, 2, 6, "the target")
	if !effectstest.HasKeywordOnBattlefield(t, g, wall, "defender") {
		t.Error("no defender")
	}
}

func TestS82OldB692CommonCauseBoostsOnlyWhileAllNonartifactCreaturesShareAColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Common Cause", "Enchantment", o692CommonCause, false)
	w1 := o692Colored(g, me.ID, "White One", "Creature — Soldier", "{W}", 1, 1)
	w2 := o692Colored(g, opp.ID, "White Two", "Creature — Soldier", "{1}{W}", 2, 2)
	golem := o692Colored(g, opp.ID, "Golem", "Artifact Creature — Golem", "{3}", 3, 3)
	o692PT(t, g, w1, 3, 3, "a white creature when all share white")
	o692PT(t, g, w2, 4, 4, "an opponent's white creature")
	o692PT(t, g, golem, 3, 3, "an artifact creature")
	green := o692Colored(g, opp.ID, "Green", "Creature — Elf", "{G}", 1, 1)
	o692PT(t, g, w1, 1, 1, "a white creature once a green one enters")
	o692PT(t, g, green, 1, 1, "the green creature")
	effectstest.KillCreature(t, g, opp.ID, green)
	o692PT(t, g, w1, 3, 3, "a white creature after the green one leaves")
	colorless := o692Colored(g, opp.ID, "Colorless", "Creature — Bear", "", 1, 1)
	o692PT(t, g, w1, 1, 1, "a white creature beside a colorless one")
	o692PT(t, g, colorless, 1, 1, "the colorless creature")
}

func TestS82OldB692CommonCauseSharedColorNeedsOneColorInAll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Common Cause", "Enchantment", o692CommonCause, false)
	a := o692Colored(g, me.ID, "Gold A", "Creature — Soldier", "{W}{U}", 1, 1)
	b := o692Colored(g, me.ID, "Gold B", "Creature — Soldier", "{W}{B}", 1, 1)
	o692PT(t, g, a, 3, 3, "a creature when W is common to both")
	o692PT(t, g, b, 3, 3, "the other creature")
	c := o692Colored(g, me.ID, "Blue Black", "Creature — Soldier", "{U}{B}", 1, 1)
	o692PT(t, g, a, 1, 1, "a creature when no single color is in all three")
	o692PT(t, g, c, 1, 1, "the third creature")
}

func TestS82OldB692SustenanceSacrificesALandForPlusOnePlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ench := effectstest.PushCatalogPermanent(g, me.ID, "Sustenance", "Enchantment", o692Sustenance, false)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, ench, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("sacrificed a creature to a land cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, ench, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	o692PT(t, g, bear, 3, 3, "the target")
	if effectstest.FindBattlefieldCardForTest(g, land) != nil {
		t.Error("the land was not sacrificed")
	}
}
