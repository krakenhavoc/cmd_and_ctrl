package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b08_cards_test.go — Streets of New Capenna counters-matter
// batch (S82, #5491), played through the real catalog entries.

const (
	sncCelestialRegulatorOracle  = "3ee68758-9431-4446-8539-f42f545f3bd0"
	sncBrokersAscendancyOracle   = "809edfed-89a9-4327-b7db-1bcec63f3e66"
	sncMetropolisAngelOracle     = "c29b592c-7115-4464-beba-e79f62691593"
	sncRevelationOfPowerOracle   = "287f60a6-e83b-487b-add9-2048a06dcca1"
	sncBackstreetBruiserOracle   = "354d9db1-18a6-4701-be3c-c4ded900e2ff"
	sncBrokersInitiateOracle     = "9bdf033e-87d4-4bc4-ba08-4b497cc1b559"
	sncKrosOracle                = "debeb004-3157-43cf-95fd-220777417ff7"
	sncFirstResponderOracle      = "8f5aed57-d0f3-4fa8-826c-0d94fa143f08"
	sncSkyboonEvangelistOracle   = "def2087a-5020-41e8-aa7b-d1f0c1f0fdb6"
	sncParkHeightsMaverickOracle = "5dd54b2a-7d6a-44f5-8827-ff3660a4b12b"
)

func TestSncB08AllRegisteredFull(t *testing.T) {
	for _, id := range []string{
		sncCelestialRegulatorOracle, sncBrokersAscendancyOracle, sncMetropolisAngelOracle,
		sncRevelationOfPowerOracle, sncBackstreetBruiserOracle, sncBrokersInitiateOracle,
		sncKrosOracle, sncFirstResponderOracle, sncSkyboonEvangelistOracle, sncParkHeightsMaverickOracle,
	} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func sncCounter(g *game.Game, id uuid.UUID, kind string, n int) {
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, kind, n) })
}

func sncMana(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

func sncPushCard(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int) uuid.UUID {
	return effectstest.B12Push(g, owner, name, typeLine, oracle, p, tough)
}

// --- Backstreet Bruiser ----------------------------------------------------

func TestSncB08BackstreetBruiserAttacksOnlyWithTwoCountersAmongYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bruiser := sncPushCard(g, me.ID, "Backstreet Bruiser", "Creature — Octopus Rogue", sncBackstreetBruiserOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	sncCounter(g, theirs, game.CounterPlusOne, 5)
	sncCounter(g, bear, game.CounterPlusOne, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bruiser, opp.ID); err == nil {
		t.Fatal("it attacked with one counter among my creatures (their counters do not count)")
	}
	sncCounter(g, bruiser, game.CounterShield, 1)
	if err := g.DeclareAttacker(bruiser, opp.ID); err != nil {
		t.Fatalf("two counters among my creatures, of different kinds, and it still cannot attack: %v", err)
	}
}

func TestSncB08BackstreetBruiserSumsCountersOnOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bruiser := sncPushCard(g, me.ID, "Backstreet Bruiser", "Creature — Octopus Rogue", sncBackstreetBruiserOracle, 3, 3)
	sncCounter(g, bruiser, game.CounterPlusOne, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bruiser, opp.ID); err != nil {
		t.Fatalf("two counters on itself: %v", err)
	}
}

func TestSncB08BackstreetBruiserStrictCastRefused(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sncMana(me, "U")
	card := game.Card{Name: "Backstreet Bruiser", TypeLine: "Creature — Octopus Rogue", OracleID: sncBackstreetBruiserOracle, ManaCost: "{1}{U}", Power: 3, Toughness: 3}
	if _, err := effectstest.Msc17Cast(t, g, me, card, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a strict cast with one mana for {1}{U} was accepted")
	}
}

// --- Brokers Initiate ------------------------------------------------------

func TestSncB08BrokersInitiateBecomesFiveFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	adv := sncPushCard(g, me.ID, "Brokers Initiate", "Creature — Cat Citizen", sncBrokersInitiateOracle, 0, 4)
	effectstest.AdvanceToMain(t, g)
	sncMana(me, "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, adv, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {4}{G/U} with four mana")
	}
	sncMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, adv, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with a hybrid G: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, adv), effectstest.EffectiveToughness(t, g, adv); p != 5 || tg != 5 {
		t.Errorf("it is %d/%d, want 5/5", p, tg)
	}
}

func TestSncB08BrokersInitiateAcceptsBlueForTheHybrid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	adv := sncPushCard(g, me.ID, "Brokers Initiate", "Creature — Cat Citizen", sncBrokersInitiateOracle, 0, 4)
	effectstest.AdvanceToMain(t, g)
	sncMana(me, "C", "C", "C", "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, adv, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with a hybrid U: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, adv), effectstest.EffectiveToughness(t, g, adv); p != 5 || tg != 5 {
		t.Errorf("it is %d/%d, want 5/5", p, tg)
	}
}

// --- Metropolis Angel ------------------------------------------------------

func TestSncB08MetropolisAngelDrawsOncePerAttackWithACounteredCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sncPushCard(g, me.ID, "Metropolis Angel", "Creature — Angel Soldier", sncMetropolisAngelOracle, 3, 1)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	sncCounter(g, a, game.CounterShield, 1)
	sncCounter(g, b, game.CounterPlusOne, 1)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Fatalf("drew %d cards for two countered attackers, want exactly 1", got)
	}
}

func TestSncB08MetropolisAngelIgnoresAnAttackWithoutCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sncPushCard(g, me.ID, "Metropolis Angel", "Creature — Angel Soldier", sncMetropolisAngelOracle, 3, 1)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("drew %d with no countered attacker", me.Hand.Size()-hand)
	}
}

// --- Revelation of Power ---------------------------------------------------

func sncCastRevelation(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID, strict bool) error {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	sncMana(me, "C", "W")
	card := game.Card{Name: "Revelation of Power", TypeLine: "Instant", OracleID: sncRevelationOfPowerOracle, ManaCost: "{1}{W}", Colors: []string{"W"}}
	_, err := effectstest.Msc17Cast(t, g, me, card, game.CastSpellParams{Strict: strict, Targets: effectstest.TdCard(target)})
	return err
}

func TestSncB08RevelationOfPowerPumpsAndGrantsEvasionOnlyWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	marked := effectstest.B12Creature(g, me.ID, "Marked", "Creature — Bear", 2, 2)
	sncCounter(g, marked, game.CounterShield, 1)
	if err := sncCastRevelation(t, g, me, plain, true); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, plain), effectstest.EffectiveToughness(t, g, plain); p != 4 || tg != 4 {
		t.Errorf("plain is %d/%d, want 4/4", p, tg)
	}
	if effectstest.EffectiveAbilitiesContain(t, g, plain, "flying") || effectstest.EffectiveAbilitiesContain(t, g, plain, "lifelink") {
		t.Error("a creature with no counter gained flying or lifelink")
	}
	if err := sncCastRevelation(t, g, me, marked, true); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, marked), effectstest.EffectiveToughness(t, g, marked); p != 4 || tg != 4 {
		t.Errorf("marked is %d/%d, want 4/4", p, tg)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, marked, "flying") || !effectstest.EffectiveAbilitiesContain(t, g, marked, "lifelink") {
		t.Error("a creature with a counter did not gain flying and lifelink")
	}
}

func TestSncB08RevelationOfPowerRefusesANonCreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	if err := sncCastRevelation(t, g, me, rock, true); err == nil {
		t.Fatal("a strict cast at an artifact was accepted")
	}
}

// --- Celestial Regulator ---------------------------------------------------

func sncCastRegulator(t *testing.T, g *game.Game, me *game.Player, theirs uuid.UUID) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	sncMana(me, "C", "W", "U")
	card := game.Card{Name: "Celestial Regulator", TypeLine: "Creature — Angel Advisor", OracleID: sncCelestialRegulatorOracle, ManaCost: "{1}{W}{U}", Power: 2, Toughness: 3, Colors: []string{"W", "U"}}
	if _, err := effectstest.Msc17Cast(t, g, me, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSncB08CelestialRegulatorTapsAndFreezesWhenYouControlACounteredCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	sncCounter(g, mine, game.CounterPlusOne, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	sncCastRegulator(t, g, me, theirs)
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("the chosen creature is not tapped")
	}
	if len(effectstest.TmtB1Card(t, g, theirs).NextUntapSkips) != 1 {
		t.Errorf("the chosen creature has %d untap skips, want 1", len(effectstest.TmtB1Card(t, g, theirs).NextUntapSkips))
	}
}

func TestSncB08CelestialRegulatorOnlyTapsWithoutACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	sncCastRegulator(t, g, me, theirs)
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("the chosen creature is not tapped")
	}
	if len(effectstest.TmtB1Card(t, g, theirs).NextUntapSkips) != 0 {
		t.Error("it was frozen with no creature carrying a counter")
	}
}

// --- Brokers Ascendancy ----------------------------------------------------

func TestSncB08BrokersAscendancyCountersEveryCreatureAndPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sncPushCard(g, me.ID, "Brokers Ascendancy", "Enchantment", sncBrokersAscendancyOracle, 0, 0)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	pw := effectstest.B12Permanent(g, me.ID, "Walker", "Legendary Planeswalker — Test")
	sncCounter(g, pw, game.CounterLoyalty, 3)
	theirPW := effectstest.B12Permanent(g, opp.ID, "Their Walker", "Legendary Planeswalker — Test")
	sncCounter(g, theirPW, game.CounterLoyalty, 3)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if effectstest.PlusOnes(g, id) != 1 {
			t.Errorf("my creature has %d +1/+1 counters, want 1", effectstest.PlusOnes(g, id))
		}
	}
	if effectstest.PlusOnes(g, theirs) != 0 {
		t.Error("an opponent's creature got a counter")
	}
	if effectstest.CountersOn(g, pw, game.CounterLoyalty) != 4 {
		t.Errorf("my planeswalker has %d loyalty, want 4", effectstest.CountersOn(g, pw, game.CounterLoyalty))
	}
	if effectstest.CountersOn(g, theirPW, game.CounterLoyalty) != 3 {
		t.Error("an opponent's planeswalker got a loyalty counter")
	}
}

// --- First Responder -------------------------------------------------------

func sncFirstResponderBoard(g *game.Game, me *game.Player) (resp, big uuid.UUID) {
	resp = sncPushCard(g, me.ID, "First Responder", "Creature — Ogre Citizen", sncFirstResponderOracle, 3, 3)
	big = effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 4, 4)
	return resp, big
}

func TestSncB08FirstResponderReturnsACreatureAndGrowsByItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	resp, big := sncFirstResponderBoard(g, me)
	sncCounter(g, big, game.CounterPlusOne, 1) // power 5 as it leaves
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no creature choice")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1", pick.ChooseMin, pick.ChooseMax)
	}
	if effectstest.HasID(pick.ChooseCards, resp) {
		t.Error("First Responder offered itself")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{big}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(big) {
		t.Error("the chosen creature did not return to its owner's hand")
	}
	if effectstest.PlusOnes(g, resp) != 5 {
		t.Errorf("First Responder has %d counters, want 5 (the creature's power)", effectstest.PlusOnes(g, resp))
	}
}

func TestSncB08FirstResponderMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	resp, big := sncFirstResponderBoard(g, me)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no creature choice")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, big) || effectstest.PlusOnes(g, resp) != 0 {
		t.Error("declining still bounced or grew")
	}
}

// --- Skyboon Evangelist ----------------------------------------------------

func TestSncB08SkyboonEvangelistSupportsSixOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	sncMana(me, "C", "C", "C", "C", "W")
	card := game.Card{Name: "Skyboon Evangelist", TypeLine: "Creature — Bird Advisor", OracleID: sncSkyboonEvangelistOracle, ManaCost: "{4}{W}", Power: 3, Toughness: 3, Colors: []string{"W"}}
	evang, err := effectstest.Msc17Cast(t, g, me, card, game.CastSpellParams{Strict: true})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no support target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, append(effectstest.TdCard(a), effectstest.TdCard(b)...)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, a) != 1 || effectstest.PlusOnes(g, b) != 1 {
		t.Error("support did not put a counter on each target")
	}
	if effectstest.PlusOnes(g, evang) != 0 {
		t.Error("support targeted the Evangelist itself")
	}
}

func TestSncB08SkyboonEvangelistGivesFlyingToACounteredAttackerOfAnyone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sncPushCard(g, opp.ID, "Skyboon Evangelist", "Creature — Bird Advisor", sncSkyboonEvangelistOracle, 3, 3)
	marked := effectstest.B12Creature(g, me.ID, "Marked", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	sncCounter(g, marked, game.CounterPlusOne, 1)
	// The Evangelist is the opponent's: my creatures attack their controller's
	// opponent? No: they attack the Evangelist's controller, not an opponent
	// of it, so nothing triggers.
	effectstest.DeclareAttack(t, g, opp.ID, marked, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, marked, "flying") {
		t.Fatal("a creature attacking the Evangelist's own controller gained flying")
	}
}

func TestSncB08SkyboonEvangelistFlyingOnlyForCounteredAttackersAtOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sncPushCard(g, me.ID, "Skyboon Evangelist", "Creature — Bird Advisor", sncSkyboonEvangelistOracle, 3, 3)
	marked := effectstest.B12Creature(g, me.ID, "Marked", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	sncCounter(g, marked, game.CounterPlusOne, 1)
	effectstest.DeclareAttack(t, g, opp.ID, marked, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, marked, "flying") {
		t.Error("the countered attacker did not gain flying")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, plain, "flying") {
		t.Error("an attacker with no counter gained flying")
	}
}

// --- Park Heights Maverick -------------------------------------------------

func TestSncB08ParkHeightsMaverickDethronesAndCannotBeBlockedBySmallCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mav := sncPushCard(g, me.ID, "Park Heights Maverick", "Creature — Human Soldier", sncParkHeightsMaverickOracle, 2, 2)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 3, 3)
	// Everyone is tied for the most life, so the attack is at a leader.
	effectstest.BrAttack(t, g, mav)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, mav) != 1 {
		t.Errorf("dethrone put %d counters, want 1", effectstest.PlusOnes(g, mav))
	}
	effectstest.BrRefusal(t, g.DeclareBlocker(small, mav), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(big, mav); err != nil {
		t.Fatalf("a power-3 creature could not block: %v", err)
	}
}

func TestSncB08ParkHeightsMaverickNoDethroneAgainstTheTrailingPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mav := sncPushCard(g, me.ID, "Park Heights Maverick", "Creature — Human Soldier", sncParkHeightsMaverickOracle, 2, 2)
	opp.Life = 5
	effectstest.DeclareAttack(t, g, opp.ID, mav)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, mav) != 0 {
		t.Fatal("dethrone triggered against a player who is not the life leader")
	}
}

func TestSncB08ParkHeightsMaverickProliferatesOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mav := sncPushCard(g, me.ID, "Park Heights Maverick", "Creature — Human Soldier", sncParkHeightsMaverickOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sncCounter(g, bear, game.CounterPlusOne, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, me.ID, mav)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerProliferate(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, bear) != 2 {
		t.Errorf("the Bear has %d counters, want 2", effectstest.PlusOnes(g, bear))
	}
}

func TestSncB08ParkHeightsMaverickProliferatesOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mav := sncPushCard(g, me.ID, "Park Heights Maverick", "Creature — Human Soldier", sncParkHeightsMaverickOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sncCounter(g, bear, game.CounterPlusOne, 1)
	opp.Life = 5 // no dethrone counter on the Maverick
	effectstest.DeclareAttack(t, g, opp.ID, mav)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerProliferate(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, bear) != 2 {
		t.Errorf("the Bear has %d counters, want 2", effectstest.PlusOnes(g, bear))
	}
}

// --- Kros, Defense Contractor ----------------------------------------------

func TestSncB08KrosShieldsTapsGoadsAndGivesTrampleToTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sncPushCard(g, me.ID, "Kros, Defense Contractor", "Legendary Creature — Cat Advisor", sncKrosOracle, 2, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, theirs, game.CounterShield) != 1 {
		t.Fatalf("no shield counter on the target (%d)", effectstest.CountersOn(g, theirs, game.CounterShield))
	}
	c := effectstest.TmtB1Card(t, g, theirs)
	if !c.Tapped {
		t.Error("the creature was not tapped")
	}
	if !c.IsGoadedBy(me.ID) {
		t.Error("the creature was not goaded")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, theirs, "trample") {
		t.Error("the creature did not gain trample")
	}
	if effectstest.TmtB1Card(t, g, other).Tapped || effectstest.EffectiveAbilitiesContain(t, g, other, "trample") {
		t.Error("an unrelated creature was affected")
	}
}

func TestSncB08KrosTriggersOnAnyCountersYouPutOnAnOpponentsCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sncPushCard(g, me.ID, "Kros, Defense Contractor", "Legendary Creature — Cat Advisor", sncKrosOracle, 2, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	// A counter on my own creature does nothing.
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, mine, game.CounterPlusOne, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TmtB1Card(t, g, mine).Tapped {
		t.Fatal("my own creature was tapped")
	}
	// A counter the opponent put on their own creature is not me putting it.
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(opp.ID, theirs, game.CounterPlusOne, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TmtB1Card(t, g, theirs).Tapped {
		t.Fatal("a counter the opponent put on their own creature triggered Kros")
	}
	// One I put there does, whatever the kind.
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, theirs, game.CounterPlusOne, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TmtB1Card(t, g, theirs).Tapped {
		t.Fatal("a counter I put on their creature did not trigger Kros")
	}
}
