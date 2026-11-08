package effects

import (
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		want := CompletenessFull
		if oracle == fdnHLHarbingerOracle {
			want = CompletenessCaveats
		}
		if spec.Name != name || spec.Completeness != want {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
		if want == CompletenessFull && len(spec.Caveats) != 0 {
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
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		assertKeywords(t, g, id, c.keyword)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped := pushVanillaCreature(g, opp.ID, "Tapped", 2, 2)
	assistCardTapped(t, g, tapped)
	castCatalogSpell(t, g, "Harbinger of the Tides", "Creature — Merfolk Wizard", fdnHLHarbingerOracle, nil)
	pick := game.TargetRef{Kind: game.TargetCard, ID: tapped}
	assistCardSettle(t, g, me.ID, &pick, true)
	if assistCardOnBattlefield(g, tapped) || !opp.Hand.Contains(tapped) {
		t.Error("the tapped creature did not go to its owner's hand")
	}
}

func TestFdnHLHarbingerRefusesAnUntappedOrYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	untapped := pushVanillaCreature(g, opp.ID, "Untapped", 2, 2)
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	assistCardTapped(t, g, mine)
	legal := pushVanillaCreature(g, opp.ID, "Legal", 2, 2)
	assistCardTapped(t, g, legal)
	castCatalogSpell(t, g, "Harbinger of the Tides", "Creature — Merfolk Wizard", fdnHLHarbingerOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if hasID(p.PickTargetCards, untapped) || hasID(p.PickTargetCards, mine) {
		t.Errorf("an untapped creature or your own was offered: %v", p.PickTargetCards)
	}
	if !hasID(p.PickTargetCards, legal) {
		t.Errorf("a tapped creature an opponent controls was not offered: %v", p.PickTargetCards)
	}
}

func TestFdnHLHarbingerHasFlashForTwoMore(t *testing.T) {
	spec, _ := Lookup(fdnHLHarbingerOracle)
	if len(spec.AlternativeCosts) != 1 {
		t.Fatalf("%d alternative costs, want the flash surcharge only", len(spec.AlternativeCosts))
	}
	if ac := spec.AlternativeCosts[0]; !ac.AsThoughFlash || ac.ManaCost != "{2}{U}{U}" {
		t.Errorf("flash for {2} more: %+v", ac)
	}
}

// --- Heartfire Immolator --------------------------------------------------------

func TestFdnHLHeartfireImmolatorSacrificesForItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	immolator := b12Push(g, me.ID, "Heartfire Immolator", "Creature — Human Wizard", fdnHLHeartfireOracle, 2, 2)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	fillPoolColored(me, "R", 2)
	if err := g.ActivateCatalogAbility(me.ID, immolator, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Fatal("a player was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, immolator, 0, game.ActivateAbilityParams{Targets: cardRefs(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(immolator) {
		t.Error("the Immolator was not sacrificed")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the 2/2 survived 2 damage")
	}
}

// --- Herald of Faith ------------------------------------------------------------

func TestFdnHLHeraldOfFaithGainsTwoLifeWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	herald := b12Push(g, me.ID, "Herald of Faith", "Creature — Angel", fdnHLHeraldOfFaithOracle, 4, 3)
	life := me.Life
	declareAttack(t, g, opp.ID, herald)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

// --- Heroes' Bane ---------------------------------------------------------------

func TestFdnHLHeroesBaneEntersWithFourAndDoublesItsCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bane := avB16CastCreature(t, g, "Heroes' Bane", "Creature — Hydra", fdnHLHeroesBaneOracle, 0, 0)
	if n := countersOn(g, bane, game.CounterPlusOne); n != 4 {
		t.Fatalf("%d +1/+1 counters on entry, want 4", n)
	}
	fillPoolColored(me, "G", 4)
	if err := g.ActivateCatalogAbility(me.ID, bane, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bane, game.CounterPlusOne); n != 8 {
		t.Errorf("%d +1/+1 counters after the ability, want 8 (four more, its power)", n)
	}
}

// --- Hungry Ghoul ---------------------------------------------------------------

func TestFdnHLHungryGhoulSacrificesAnotherForACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ghoul := b12Push(g, me.ID, "Hungry Ghoul", "Creature — Zombie", fdnHLHungryGhoulOracle, 2, 2)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	fillPoolColored(me, "B", 2)
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{ghoul}}); err == nil {
		t.Fatal("the Ghoul sacrificed itself to its \"another creature\" cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the other creature was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, ghoul, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

// --- Icewind Elemental ----------------------------------------------------------

func TestFdnHLIcewindElementalLootsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	yard := me.Graveyard.Size()
	elemental := castHoldCreature(t, g, "Icewind Elemental", "Creature — Elemental", fdnHLIcewindOracle, 3, 4)
	discardFromHand(t, g, me.ID)
	g.SettleResolution()
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	sage := b12Push(g, me.ID, "Infestation Sage", "Creature — Elf Warlock", fdnHLInfestationOracle, 1, 1)
	b25Destroy(g, sage)
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Insect", me.ID); n != 1 {
		t.Fatalf("%d Insect tokens, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Insect" {
			continue
		}
		if c.Power != 1 || c.Toughness != 1 || !hasKeywordOnBattlefield(t, g, c.InstanceID, "flying") {
			t.Errorf("the Insect is %d/%d, flying %v; want a 1/1 flier", c.Power, c.Toughness, hasKeywordOnBattlefield(t, g, c.InstanceID, "flying"))
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
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			seedLibrary(me, "A", "B", "C")
			target := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			hand := me.Hand.Size()
			if _, err := castWithOptionalCosts(t, g, "Into the Roil", "Instant", fdnHLIntoTheRoilOracle, b16TargetCard(target), tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	assistCardTapped(t, g, theirs)
	castCatalogSpell(t, g, "Involuntary Employment", "Sorcery", fdnHLInvoluntaryOracle, cardRefs(theirs))
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, theirs); got != me.ID {
		t.Fatal("you did not gain control of the creature")
	}
	if isTapped(g, theirs) {
		t.Error("the creature was not untapped")
	}
	if !hasKeywordOnBattlefield(t, g, theirs, "haste") {
		t.Error("the creature did not gain haste")
	}
	if n := namedOnBattlefield(g, "Treasure", me.ID); n != 1 {
		t.Errorf("%d Treasure tokens, want 1", n)
	}
}

// --- Joust Through --------------------------------------------------------------

func TestFdnHLJoustThroughHitsOnlyACreatureInCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	if err := castCatalogSpellErr(t, g, "Joust Through", "Instant", fdnHLJoustThroughOracle, cardRefs(bear)); err == nil {
		t.Fatal("a creature that is neither attacking nor blocking was accepted")
	}
	fdnHLSetAttacking(g, bear, me.ID)
	life := me.Life
	castCatalogSpell(t, g, "Joust Through", "Instant", fdnHLJoustThroughOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the 3/3 attacker survived 3 damage")
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

// --- Juggernaut -----------------------------------------------------------------

func TestFdnHLJuggernautMustAttackAndWallsCantBlockIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jugg := b12Push(g, me.ID, "Juggernaut", "Artifact Creature — Juggernaut", fdnHLJuggernautOracle, 5, 3)
	var reqs []game.AttackRequirement
	g.ReadSnapshot(func() {
		if c, ok := battlefieldCard(g, jugg); ok {
			reqs = c.Effective().AttackRequirements
		}
	})
	if len(reqs) == 0 {
		t.Error("Juggernaut carries no attack requirement")
	}
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(jugg, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, jugg); err == nil {
		t.Error("a Wall blocked Juggernaut")
	}
	if err := g.DeclareBlocker(bear, jugg); err != nil {
		t.Errorf("a non-Wall could not block: %v", err)
	}
}

// --- Kargan Dragonrider ---------------------------------------------------------

func TestFdnHLKarganDragonriderFliesWhileYouControlADragon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rider := b12Push(g, me.ID, "Kargan Dragonrider", "Creature — Human Warrior", fdnHLKarganOracle, 2, 2)
	if hasKeywordOnBattlefield(t, g, rider, "flying") {
		t.Fatal("flying without a Dragon")
	}
	b12Creature(g, opp.ID, "Their Dragon", "Creature — Dragon", 4, 4)
	if hasKeywordOnBattlefield(t, g, rider, "flying") {
		t.Fatal("flying from an opponent's Dragon")
	}
	dragon := b12Creature(g, me.ID, "My Dragon", "Creature — Dragon", 4, 4)
	if !hasKeywordOnBattlefield(t, g, rider, "flying") {
		t.Fatal("no flying while you control a Dragon")
	}
	b25Destroy(g, dragon)
	if hasKeywordOnBattlefield(t, g, rider, "flying") {
		t.Error("still flying after your Dragon left")
	}
}

// --- Kindled Fury ---------------------------------------------------------------

func TestFdnHLKindledFuryPumpsAndGrantsFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Kindled Fury", "Instant", fdnHLKindledFuryOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 3/2", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "first strike") {
		t.Error("the Bear did not gain first strike")
	}
}

// --- Kitesail Corsair -----------------------------------------------------------

func TestFdnHLKitesailCorsairFliesOnlyWhileAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	corsair := b12Push(g, me.ID, "Kitesail Corsair", "Creature — Human Pirate", fdnHLKitesailOracle, 2, 1)
	if hasKeywordOnBattlefield(t, g, corsair, "flying") {
		t.Fatal("flying while not attacking")
	}
	ground := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, corsair)
	if !hasKeywordOnBattlefield(t, g, corsair, "flying") {
		t.Fatal("no flying while attacking")
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(ground, corsair); err == nil {
		t.Error("a creature without flying or reach blocked the attacking Corsair")
	}
}

// --- Luminous Rebuke ------------------------------------------------------------

func TestFdnHLLuminousRebukeCostsThreeLessAgainstATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	untapped := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	tapped := b12Creature(g, opp.ID, "Elk", "Creature — Elk", 2, 2)
	g.WithWriteLock(func() { b16Tap(g, tapped) })
	advanceToMain(t, g)
	card := game.Card{Name: "Luminous Rebuke", TypeLine: "Instant", ManaCost: "{4}{W}", OracleID: fdnHLLuminousOracle}
	if err := socB19CastWithPool(g, me, card, "W", 2, game.CastSpellParams{Targets: cardRefs(untapped)}); err == nil {
		t.Fatal("cast at the full price with two mana against an untapped creature")
	}
	if err := socB19CastWithPool(g, me, card, "W", 2, game.CastSpellParams{Targets: cardRefs(tapped)}); err != nil {
		t.Fatalf("cast for {1}{W} against a tapped creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature survived")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature died")
	}
}
