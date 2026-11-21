package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdn_hl_cards_test.go — Foundations, H–L (#4465): the cards from that
// stretch of the alphabet that earlier batches left and that existing
// seams express. Each test plays the real catalog entry through the
// engine and asserts what its text says.

const (
	fdnHLHarbingerOracle     = "9f5277e8-1cbe-478c-a8bc-ac028f13305e"
	fdnHLHeartfireOracle     = "1e414a4f-6e76-4cbe-9c69-d52852cd5bc0"
	fdnHLHeraldOfFaithOracle = "cb97da84-1d13-4795-b68a-2bf111a50067"
	fdnHLHeroesBaneOracle    = "bccd722a-b610-4017-93bd-313f08448f5e"
	fdnHLHungryGhoulOracle   = "4a7a55a3-568c-4c31-9460-5b94dc0121d8"
	fdnHLIcewindOracle       = "5001bf21-9dbb-4229-aeef-25805e71b832"
	fdnHLInfestationOracle   = "d398dca3-f3a5-44c9-998f-b4fce9e5c2bf"
	fdnHLIntoTheRoilOracle   = "c2898bbd-82a4-4d26-b6ee-169b0ebe71b4"
	fdnHLInvoluntaryOracle   = "7e86fb1e-9987-43ba-b59e-0cb03bbdedb6"
	fdnHLJoustThroughOracle  = "33c21ef1-33eb-447a-a318-a352f163a8a0"
	fdnHLJuggernautOracle    = "4ac9116f-36bc-4d71-b696-d6ee064e1d58"
	fdnHLKarganOracle        = "df366871-8ed7-4183-8b0c-414ca361a711"
	fdnHLKindledFuryOracle   = "ebeee06e-3345-4f02-896f-cb0b2cfa5548"
	fdnHLKitesailOracle      = "6103f294-f920-4307-af80-d44cc71167fd"
	fdnHLLuminousOracle      = "d96141af-313f-4ea9-a249-d9eb22cd2431"
)

func TestFdnHLEveryCardIsRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		fdnHLHarbingerOracle:     "Harbinger of the Tides",
		fdnHLHeartfireOracle:     "Heartfire Immolator",
		fdnHLHeraldOfFaithOracle: "Herald of Faith",
		fdnHLHeroesBaneOracle:    "Heroes' Bane",
		fdnHLHungryGhoulOracle:   "Hungry Ghoul",
		fdnHLIcewindOracle:       "Icewind Elemental",
		fdnHLInfestationOracle:   "Infestation Sage",
		fdnHLIntoTheRoilOracle:   "Into the Roil",
		fdnHLInvoluntaryOracle:   "Involuntary Employment",
		fdnHLJoustThroughOracle:  "Joust Through",
		fdnHLJuggernautOracle:    "Juggernaut",
		fdnHLKarganOracle:        "Kargan Dragonrider",
		fdnHLKindledFuryOracle:   "Kindled Fury",
		fdnHLKitesailOracle:      "Kitesail Corsair",
		fdnHLLuminousOracle:      "Luminous Rebuke",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
		if len(spec.Caveats) != 0 {
			t.Errorf("%s: a full card with caveats %v", name, spec.Caveats)
		}
	}
}

func TestFdnHLKeywordBodiesCarryTheirPrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keyword                string
	}{
		{"Herald of Faith", "Creature — Angel", fdnHLHeraldOfFaithOracle, 4, 3, "flying"},
		{"Icewind Elemental", "Creature — Elemental", fdnHLIcewindOracle, 3, 4, "flying"},
		{"Heartfire Immolator", "Creature — Human Wizard", fdnHLHeartfireOracle, 2, 2, "prowess"},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.B12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		effectstest.AssertKeywords(t, g, id, c.keyword)
	}
}

// fdnHLSetAttacking marks a battlefield creature as attacking `defender`
// for a test that only needs the combat flag (Elspeth's Smite's setup).
func fdnHLSetAttacking(g *game.Game, id, defender uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].AttackingTarget = defender
			}
		}
	})
}

// --- Harbinger of the Tides -----------------------------------------------------

func TestFdnHLHarbingerReturnsATappedOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped := effectstest.PushVanillaCreature(g, opp.ID, "Tapped", 2, 2)
	effectstest.AssistCardTapped(t, g, tapped)
	effectstest.CastCatalogSpell(t, g, "Harbinger of the Tides", "Creature — Merfolk Wizard", fdnHLHarbingerOracle, nil)
	pick := game.TargetRef{Kind: game.TargetCard, ID: tapped}
	effectstest.AssistCardSettle(t, g, me.ID, &pick, true)
	if effectstest.AssistCardOnBattlefield(g, tapped) || !opp.Hand.Contains(tapped) {
		t.Error("the tapped creature did not go to its owner's hand")
	}
}

func TestFdnHLHarbingerRefusesAnUntappedOrYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	untapped := effectstest.PushVanillaCreature(g, opp.ID, "Untapped", 2, 2)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	effectstest.AssistCardTapped(t, g, mine)
	legal := effectstest.PushVanillaCreature(g, opp.ID, "Legal", 2, 2)
	effectstest.AssistCardTapped(t, g, legal)
	effectstest.CastCatalogSpell(t, g, "Harbinger of the Tides", "Creature — Merfolk Wizard", fdnHLHarbingerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if effectstest.HasID(p.PickTargetCards, untapped) || effectstest.HasID(p.PickTargetCards, mine) {
		t.Errorf("an untapped creature or your own was offered: %v", p.PickTargetCards)
	}
	if !effectstest.HasID(p.PickTargetCards, legal) {
		t.Errorf("a tapped creature an opponent controls was not offered: %v", p.PickTargetCards)
	}
}

func TestFdnHLHarbingerHasFlashForTwoMore(t *testing.T) {
	spec, _ := effects.Lookup(fdnHLHarbingerOracle)
	// #5207: an optional additional {2} that opens the window, not an
	// alternative cost.
	if len(spec.AlternativeCosts) != 0 || len(spec.OptionalCosts) != 1 {
		t.Fatalf("%d alternative costs, %d optional costs; want the {2} optional cost only", len(spec.AlternativeCosts), len(spec.OptionalCosts))
	}
	if oc := spec.OptionalCosts[0]; !oc.AsThoughFlash || oc.ManaCost != "{2}" {
		t.Errorf("flash for {2} more: %+v", oc)
	}
}

// --- Heartfire Immolator --------------------------------------------------------

func TestFdnHLHeartfireImmolatorSacrificesForItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	immolator := effectstest.B12Push(g, me.ID, "Heartfire Immolator", "Creature — Human Wizard", fdnHLHeartfireOracle, 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.FillPoolColored(me, "R", 2)
	if err := g.ActivateCatalogAbility(me.ID, immolator, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Fatal("a player was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, immolator, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(immolator) {
		t.Error("the Immolator was not sacrificed")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the 2/2 survived 2 damage")
	}
}

// --- Herald of Faith ------------------------------------------------------------

func TestFdnHLHeraldOfFaithGainsTwoLifeWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	herald := effectstest.B12Push(g, me.ID, "Herald of Faith", "Creature — Angel", fdnHLHeraldOfFaithOracle, 4, 3)
	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, herald)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

// --- Heroes' Bane ---------------------------------------------------------------

func TestFdnHLHeroesBaneEntersWithFourAndDoublesItsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bane := effectstest.AvB16CastCreature(t, g, "Heroes' Bane", "Creature — Hydra", fdnHLHeroesBaneOracle, 0, 0)
	if n := effectstest.CountersOn(g, bane, game.CounterPlusOne); n != 4 {
		t.Fatalf("%d +1/+1 counters on entry, want 4", n)
	}
	effectstest.FillPoolColored(me, "G", 4)
	if err := g.ActivateCatalogAbility(me.ID, bane, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bane, game.CounterPlusOne); n != 8 {
		t.Errorf("%d +1/+1 counters after the ability, want 8 (four more, its power)", n)
	}
}

// --- Hungry Ghoul ---------------------------------------------------------------

func TestFdnHLHungryGhoulSacrificesAnotherForACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ghoul := effectstest.B12Push(g, me.ID, "Hungry Ghoul", "Creature — Zombie", fdnHLHungryGhoulOracle, 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "B", 2)
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{ghoul}}); err == nil {
		t.Fatal("the Ghoul sacrificed itself to its \"another creature\" cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the other creature was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, ghoul, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

// --- Icewind Elemental ----------------------------------------------------------

func TestFdnHLIcewindElementalLootsOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	yard := me.Graveyard.Size()
	elemental := effectstest.CastHoldCreature(t, g, "Icewind Elemental", "Creature — Elemental", fdnHLIcewindOracle, 3, 4)
	effectstest.DiscardFromHand(t, g, me.ID)
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(elemental) {
		t.Fatal("the Elemental is not on the battlefield")
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d, want %d (draw one, discard one)", got, before)
	}
	if got := me.Graveyard.Size(); got != yard+1 {
		t.Errorf("graveyard %d, want %d", got, yard+1)
	}
}

// --- Infestation Sage -----------------------------------------------------------

func TestFdnHLInfestationSageLeavesAFlyingInsect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sage := effectstest.B12Push(g, me.ID, "Infestation Sage", "Creature — Elf Warlock", fdnHLInfestationOracle, 1, 1)
	effectstest.B25Destroy(g, sage)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Insect", me.ID); n != 1 {
		t.Fatalf("%d Insect tokens, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Insect" {
			continue
		}
		if c.Power != 1 || c.Toughness != 1 || !effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "flying") {
			t.Errorf("the Insect is %d/%d, flying %v; want a 1/1 flier", c.Power, c.Toughness, effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "flying"))
		}
	}
}

// --- Into the Roil --------------------------------------------------------------

func TestFdnHLIntoTheRoilBouncesAndDrawsOnlyWhenKicked(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		draws    int
	}{
		{"unkicked", nil, 0},
		{"kicked", []int{0}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.SeedLibrary(me, "A", "B", "C")
			target := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			hand := me.Hand.Size()
			if _, err := effectstest.CastWithOptionalCosts(t, g, "Into the Roil", "Instant", fdnHLIntoTheRoilOracle, effectstest.B16TargetCard(target), tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(target) || !opp.Hand.Contains(target) {
				t.Error("the target was not returned to its owner's hand")
			}
			if got := me.Hand.Size() - hand; got != tc.draws {
				t.Errorf("hand changed by %d, want %d draws", got, tc.draws)
			}
		})
	}
}

// --- Involuntary Employment -----------------------------------------------------

func TestFdnHLInvoluntaryEmploymentStealsUntapsHastesAndMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AssistCardTapped(t, g, theirs)
	effectstest.CastCatalogSpell(t, g, "Involuntary Employment", "Sorcery", fdnHLInvoluntaryOracle, effectstest.CardRefs(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, theirs); got != me.ID {
		t.Fatal("you did not gain control of the creature")
	}
	if effectstest.IsTapped(g, theirs) {
		t.Error("the creature was not untapped")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, theirs, "haste") {
		t.Error("the creature did not gain haste")
	}
	if n := effectstest.NamedOnBattlefield(g, "Treasure", me.ID); n != 1 {
		t.Errorf("%d Treasure tokens, want 1", n)
	}
}

// --- Joust Through --------------------------------------------------------------

func TestFdnHLJoustThroughHitsOnlyACreatureInCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	if err := effectstest.CastCatalogSpellErr(t, g, "Joust Through", "Instant", fdnHLJoustThroughOracle, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("a creature that is neither attacking nor blocking was accepted")
	}
	fdnHLSetAttacking(g, bear, me.ID)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Joust Through", "Instant", fdnHLJoustThroughOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the 3/3 attacker survived 3 damage")
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

// --- Juggernaut -----------------------------------------------------------------

func TestFdnHLJuggernautMustAttackAndWallsCantBlockIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jugg := effectstest.B12Push(g, me.ID, "Juggernaut", "Artifact Creature — Juggernaut", fdnHLJuggernautOracle, 5, 3)
	var reqs []game.AttackRequirement
	g.ReadSnapshot(func() {
		if c, ok := effectstest.BattlefieldCard(g, jugg); ok {
			reqs = c.Effective().AttackRequirements
		}
	})
	if len(reqs) == 0 {
		t.Error("Juggernaut carries no attack requirement")
	}
	wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(jugg, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, jugg); err == nil {
		t.Error("a Wall blocked Juggernaut")
	}
	if err := g.DeclareBlocker(bear, jugg); err != nil {
		t.Errorf("a non-Wall could not block: %v", err)
	}
}

// --- Kargan Dragonrider ---------------------------------------------------------

func TestFdnHLKarganDragonriderFliesWhileYouControlADragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rider := effectstest.B12Push(g, me.ID, "Kargan Dragonrider", "Creature — Human Warrior", fdnHLKarganOracle, 2, 2)
	if effectstest.HasKeywordOnBattlefield(t, g, rider, "flying") {
		t.Fatal("flying without a Dragon")
	}
	effectstest.B12Creature(g, opp.ID, "Their Dragon", "Creature — Dragon", 4, 4)
	if effectstest.HasKeywordOnBattlefield(t, g, rider, "flying") {
		t.Fatal("flying from an opponent's Dragon")
	}
	dragon := effectstest.B12Creature(g, me.ID, "My Dragon", "Creature — Dragon", 4, 4)
	if !effectstest.HasKeywordOnBattlefield(t, g, rider, "flying") {
		t.Fatal("no flying while you control a Dragon")
	}
	effectstest.B25Destroy(g, dragon)
	if effectstest.HasKeywordOnBattlefield(t, g, rider, "flying") {
		t.Error("still flying after your Dragon left")
	}
}

// --- Kindled Fury ---------------------------------------------------------------

func TestFdnHLKindledFuryPumpsAndGrantsFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Kindled Fury", "Instant", fdnHLKindledFuryOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 3/2", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "first strike") {
		t.Error("the Bear did not gain first strike")
	}
}

// --- Kitesail Corsair -----------------------------------------------------------

func TestFdnHLKitesailCorsairFliesOnlyWhileAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	corsair := effectstest.B12Push(g, me.ID, "Kitesail Corsair", "Creature — Human Pirate", fdnHLKitesailOracle, 2, 1)
	if effectstest.HasKeywordOnBattlefield(t, g, corsair, "flying") {
		t.Fatal("flying while not attacking")
	}
	ground := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, corsair)
	if !effectstest.HasKeywordOnBattlefield(t, g, corsair, "flying") {
		t.Fatal("no flying while attacking")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(ground, corsair); err == nil {
		t.Error("a creature without flying or reach blocked the attacking Corsair")
	}
}

// --- Luminous Rebuke ------------------------------------------------------------

func TestFdnHLLuminousRebukeCostsThreeLessAgainstATappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	untapped := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	tapped := effectstest.B12Creature(g, opp.ID, "Elk", "Creature — Elk", 2, 2)
	g.WithWriteLock(func() { effectstest.B16Tap(g, tapped) })
	effectstest.AdvanceToMain(t, g)
	card := game.Card{Name: "Luminous Rebuke", TypeLine: "Instant", ManaCost: "{4}{W}", OracleID: fdnHLLuminousOracle}
	if err := effectstest.SocB19CastWithPool(g, me, card, "W", 2, game.CastSpellParams{Targets: effectstest.CardRefs(untapped)}); err == nil {
		t.Fatal("cast at the full price with two mana against an untapped creature")
	}
	if err := effectstest.SocB19CastWithPool(g, me, card, "W", 2, game.CastSpellParams{Targets: effectstest.CardRefs(tapped)}); err != nil {
		t.Fatalf("cast for {1}{W} against a tapped creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature survived")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature died")
	}
}
