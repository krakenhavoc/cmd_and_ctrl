package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_creature_d_test.go — the Reality Fracture creatures of
// slice fra-creature-d.

const (
	rfdMarwynClearcutterOracle = "aff4e43d-cafa-47de-8e39-011f9254a26a"
	rfdMarwynPreserverOracle   = "eb03f6c1-c94f-4f09-b2e9-a55ce5533653"
	rfdMasterOfBarbsOracle     = "e8c36337-135c-460d-a1da-cc01743dda71"
	rfdMemnarchOracle          = "f2630b02-fb48-4167-8a47-8284f7d4c484"
	rfdNivMizzetOracle         = "a5350327-e84c-44b6-8359-87882e521048"
	rfdObNixilisOracle         = "3a49b518-61c8-4dea-8008-c5edf8384581"
	rfdOmnathOracle            = "579225b5-50e2-4891-8871-0bf6dbc07e33"
	rfdPiaAetherOracle         = "6519dc16-4960-4220-aaa8-7e4279743714"
	rfdPiaRebuilderOracle      = "a49c0c64-ea87-44f4-9acf-e7181962aeb2"
	rfdWitchstalkerOracle      = "c21fafaf-4510-4752-a501-606ae68e6699"
	rfdProftConsultingOracle   = "45473bdb-b96f-4213-a46d-b2dbdf6b5e56"
	rfdProftSinisterOracle     = "a350856c-87f2-4605-999f-4e26aa362a55"
	rfdRampartHunterOracle     = "0867755a-9fb9-42ad-897e-6206fbfbc1f8"
	rfdRankRatOracle           = "64e7ef06-ba30-4b19-9d44-ca77944cc930"
	rfdRescueGirlOracle        = "df6fc0b4-333b-4f6c-804c-ba6c8e29c88e"
	rfdRuricBiomagusOracle     = "873fa095-5662-407a-bd24-9a1c4712a428"
	rfdRuricMagecrusherOracle  = "453b2b3c-7df5-4515-9565-82f603cf451b"
	rfdSaheeliConsulOracle     = "dfa95c0e-393d-4b22-9de3-45a7efc0ba14"
	rfdSaheeliJewelOracle      = "83705258-1f42-41db-a32e-de99ffd759eb"
)

// rfdPush puts a catalog creature (or other permanent) with a real
// power and toughness onto the battlefield, ready to act.
func rfdPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// rfdHandCard seeds a card with real power/toughness into a hand.
func rfdHandCard(p *game.Player, name, typeLine, manaCost, oracle string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// rfdCast casts a creature from the active seat's hand and lets it
// (and anything it triggers) settle.
func rfdCast(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfdHandCard(me, name, typeLine, "", oracle, power, toughness)
	toMainForCost(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func rfdCountNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

func rfdOnBattlefield(g *game.Game, id uuid.UUID) bool {
	_, ok := battlefieldCard(g, id)
	return ok
}

func rfdBolt(t *testing.T, g *game.Game, target game.TargetRef) {
	t.Helper()
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{target})
	passPriorityAroundTable(t, g)
}

func TestRFDCardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		rfdMarwynClearcutterOracle: "Marwyn, the Clearcutter",
		rfdMarwynPreserverOracle:   "Marwyn, the Preserver",
		rfdMasterOfBarbsOracle:     "Master of Barbs",
		rfdMemnarchOracle:          "Memnarch, the Warden",
		rfdNivMizzetOracle:         "Niv-Mizzet, Ghost Counsel",
		rfdObNixilisOracle:         "Ob Nixilis, the Ascended",
		rfdOmnathOracle:            "Omnath, Locus of the Void",
		rfdPiaAetherOracle:         "Pia, Aether Ascetic",
		rfdPiaRebuilderOracle:      "Pia, Determined Rebuilder",
		rfdWitchstalkerOracle:      "Primal Witchstalker",
		rfdProftConsultingOracle:   "Proft, Consulting Detective",
		rfdProftSinisterOracle:     "Proft, Sinister Mastermind",
		rfdRampartHunterOracle:     "Rampart Hunter",
		rfdRankRatOracle:           "Rank Rat",
		rfdRescueGirlOracle:        "Rescue Girl, First Responder",
		rfdRuricBiomagusOracle:     "Ruric Thar, Biomagus",
		rfdRuricMagecrusherOracle:  "Ruric Thar, Magecrusher",
		rfdSaheeliConsulOracle:     "Saheeli, Consul of Oversight",
		rfdSaheeliJewelOracle:      "Saheeli, Jewel of Avishkar",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// --- Rank Rat ----------------------------------------------------------

func TestRFDRankRatMakesEachOpponentDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Hand.Size()
	}
	rfdCast(t, g, "Rank Rat", "Creature — Zombie Rat", rfdRankRatOracle, 1, 1)
	for _, p := range g.Seats[1:] {
		if discardChoiceFor(g, p.ID) == nil {
			t.Fatalf("opponent %s was not asked to discard", p.ID)
		}
	}
	if discardChoiceFor(g, me.ID) != nil {
		t.Error("the caster must not be asked to discard")
	}
	for _, p := range g.Seats[1:] {
		discardFromHand(t, g, p.ID)
		if got := p.Hand.Size(); got != before[p.ID]-1 {
			t.Errorf("opponent hand = %d, want %d", got, before[p.ID]-1)
		}
	}
}

// --- Rampart Hunter ----------------------------------------------------

func TestRFDRampartHunterPumpsAndGrantsDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := rfdPush(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	rfdHandCard(me, "x", "Land", "", "", 0, 0) // keep the hand non-empty
	hunter := rfdHandCard(me, "Rampart Hunter", "Creature — Horror", "", rfdRampartHunterOracle, 3, 3)
	toMainForCost(t, g)
	if err := g.CastSpell(me.ID, hunter, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("the enters trigger should ask for a target creature")
	}
	pickTriggerTarget(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, theirs); got != 4 {
		t.Errorf("target power = %d, want 4", got)
	}
	if got := effectiveToughness(t, g, theirs); got != 4 {
		t.Errorf("target toughness = %d, want 4", got)
	}
	if !hasAbility(effectiveAbilities(t, g, theirs), "deathtouch") {
		t.Error("the target should gain deathtouch")
	}
	if !hasAbility(effectiveAbilities(t, g, hunter), "deathtouch") {
		t.Error("Rampart Hunter has deathtouch itself")
	}
}

// --- Pia, Determined Rebuilder ------------------------------------------

func TestRFDPiaDeterminedRebuilderMakesAThopterAndPumpsPerArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pia := rfdCast(t, g, "Pia, Determined Rebuilder", "Legendary Creature — Human Artificer", rfdPiaRebuilderOracle, 2, 2)
	if got := rfdCountNamed(g, me.ID, "Thopter"); got != 1 {
		t.Fatalf("Thopters = %d, want 1", got)
	}
	pushPermanentForTest(g, me.ID, "Trinket", "", "Artifact")
	fillPoolColored(me, "R", 1)
	fillPool(me, 5)
	if err := g.ActivateCatalogAbility(me.ID, pia, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: pia}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Two artifacts: the Thopter token and the Trinket.
	if got := effectivePower(t, g, pia); got != 4 {
		t.Errorf("Pia power = %d, want 4 (2 + two artifacts)", got)
	}
	if got := effectiveToughness(t, g, pia); got != 2 {
		t.Errorf("Pia toughness = %d, want 2 (+X/+0)", got)
	}
}

// --- Memnarch, the Warden ----------------------------------------------

func TestRFDMemnarchMakesTwoMyrAndDrawsPerArtifactOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mem := rfdCast(t, g, "Memnarch, the Warden", "Legendary Artifact Creature — Wizard", rfdMemnarchOracle, 8, 9)
	if got := rfdCountNamed(g, me.ID, "Myr"); got != 2 {
		t.Fatalf("Myr = %d, want 2", got)
	}
	// Clear the sickness the cast gave him so he can attack.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == mem {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, mem)
	passPriorityAroundTable(t, g)
	// Memnarch and two Myr: three artifacts, three cards.
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand = %d, want %d (a card per artifact)", got, hand+3)
	}
}

// --- Marwyn, the Clearcutter --------------------------------------------

func TestRFDMarwynClearcutterSacrificesAnArtifactOrLandToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	marwyn := rfdPush(g, me.ID, "Marwyn, the Clearcutter", "Legendary Creature — Elf Warrior", rfdMarwynClearcutterOracle, 2, 1)
	bear := rfdPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	rock := pushPermanentForTest(g, me.ID, "Rock", "", "Artifact")
	fillPool(me, 2)
	hand := me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, marwyn, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a plain creature is not an artifact or land")
	}
	if !rfdOnBattlefield(g, bear) || me.Hand.Size() != hand {
		t.Fatal("a refused activation must not pay any part of the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, marwyn, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if rfdOnBattlefield(g, rock) {
		t.Error("the artifact should have been sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d", got, hand+1)
	}
}

// --- Marwyn, the Preserver ---------------------------------------------

func TestRFDMarwynPreserverGivesYourLandsHexproofAndRecurs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marwyn := rfdPush(g, me.ID, "Marwyn, the Preserver", "Legendary Creature — Elf Druid", rfdMarwynPreserverOracle, 3, 2)
	mine := rfdPush(g, me.ID, "My Forest", "Basic Land — Forest", "", 0, 0)
	theirs := rfdPush(g, opp.ID, "Their Forest", "Basic Land — Forest", "", 0, 0)
	if !hasAbility(effectiveAbilities(t, g, mine), "hexproof") {
		t.Error("my land should have hexproof")
	}
	if hasAbility(effectiveAbilities(t, g, theirs), "hexproof") {
		t.Error("an opponent's land must not gain hexproof")
	}

	land := pushGraveyardCardTyped(me, "Dead Forest", "Basic Land — Forest")
	spell := pushGraveyardCardTyped(me, "Dead Spell", "Sorcery")
	fillPool(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, marwyn, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	}); err == nil {
		t.Fatal("a nonland card is not a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, marwyn, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the land card should be back in hand")
	}
}

// --- Master of Barbs ---------------------------------------------------

func TestRFDMasterOfBarbsPumpsOnNoncombatDamageToAnOpponentOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := rfdPush(g, me.ID, "Master of Barbs", "Creature — Lizard Bard", rfdMasterOfBarbsOracle, 2, 1)
	other := rfdPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := rfdPush(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)

	// Damage to a creature is not damage to an opponent.
	rfdBolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: theirs})
	if got := effectivePower(t, g, master); got != 2 {
		t.Fatalf("power after bolting a creature = %d, want 2", got)
	}
	rfdBolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if got := effectivePower(t, g, master); got != 3 {
		t.Errorf("Master power = %d, want 3", got)
	}
	if got := effectivePower(t, g, other); got != 3 {
		t.Errorf("the other creature's power = %d, want 3", got)
	}
}

func TestRFDMasterOfBarbsIgnoresCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := rfdPush(g, me.ID, "Master of Barbs", "Creature — Lizard Bard", rfdMasterOfBarbsOracle, 2, 1)
	attackWith(t, g, opp.ID, master)
	passPriorityAroundTable(t, g)
	if opp.Life == 40 {
		t.Fatal("the attack should have connected")
	}
	if got := effectivePower(t, g, master); got != 2 {
		t.Errorf("combat damage must not trigger it: power = %d, want 2", got)
	}
}

// --- Niv-Mizzet, Ghost Counsel -----------------------------------------

func TestRFDNivMizzetPaysLifeToDrawThatMany(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rfdPush(g, me.ID, "Niv-Mizzet, Ghost Counsel", "Legendary Creature — Spirit Dragon", rfdNivMizzetOracle, 4, 4)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	passPriorityAroundTable(t, g)
	life, hand := me.Life, me.Hand.Size()
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life = %d, want %d (paid 3)", me.Life, life-3)
	}
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand = %d, want %d", got, hand+3)
	}
}

func TestRFDNivMizzetDeclineDoesNothingAndDrainFeedsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	niv := rfdPush(g, me.ID, "Niv-Mizzet, Ghost Counsel", "Legendary Creature — Spirit Dragon", rfdNivMizzetOracle, 4, 4)
	life, oppLife, hand := me.Life, opp.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, niv, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != life+1 {
		t.Fatalf("drain: opp %d, me %d; want %d and %d", opp.Life, me.Life, oppLife-1, life+1)
	}
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 || me.Hand.Size() != hand {
		t.Errorf("declined: life %d hand %d, want %d and %d", me.Life, me.Hand.Size(), life+1, hand)
	}
}

// --- Pia, Aether Ascetic -----------------------------------------------

func TestRFDPiaAetherAsceticDiscardsToTutorAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ench := pushLibraryCardForTest(me, game.Card{
		InstanceID: uuid.New(), Name: "Test Enchantment", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID,
	})
	rfdCast(t, g, "Pia, Aether Ascetic", "Legendary Creature — Human Druid", rfdPiaAetherOracle, 2, 2)
	c := discardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the controller should be offered a discard")
	}
	answerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ench) {
		t.Error("the enchantment should have been searched into hand")
	}
}

func TestRFDPiaAetherAsceticDecliningTheDiscardSearchesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ench := pushLibraryCardForTest(me, game.Card{
		InstanceID: uuid.New(), Name: "Test Enchantment", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID,
	})
	rfdCast(t, g, "Pia, Aether Ascetic", "Legendary Creature — Human Druid", rfdPiaAetherOracle, 2, 2)
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(ench) {
		t.Error("no discard, no search")
	}
}

// --- Primal Witchstalker -----------------------------------------------

func TestRFDPrimalWitchstalkerMillsFourAndReturnsALandTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushGraveyardCardTyped(me, "Dead Forest", "Basic Land — Forest")
	lib := me.Library.Size()
	rfdCast(t, g, "Primal Witchstalker", "Creature — Wolf", rfdWitchstalkerOracle, 2, 1)
	if got := me.Library.Size(); got != lib-4 {
		t.Fatalf("library = %d, want %d (milled four)", got, lib-4)
	}
	if latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, land)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, land)
	if !ok {
		t.Fatal("the land card should have returned to the battlefield")
	}
	if !c.Tapped {
		t.Error("it returns tapped")
	}
}

func TestRFDPrimalWitchstalkerWithNoLandStillMills(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Graveyard.Cards = nil
	lib := me.Library.Size()
	rfdCast(t, g, "Primal Witchstalker", "Creature — Wolf", rfdWitchstalkerOracle, 2, 1)
	// The library is basic fillers (no land type line), so nothing is a
	// legal target and the reflexive trigger is removed.
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("no land card in the graveyard: no prompt")
	}
	if got := me.Library.Size(); got != lib-4 {
		t.Errorf("library = %d, want %d", got, lib-4)
	}
}

// --- Proft, Consulting Detective ---------------------------------------

func TestRFDProftConsultingPaysTwoForACounterAndACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	proft := rfdPush(g, me.ID, "Proft, Consulting Detective", "Legendary Creature — Human Detective", rfdProftConsultingOracle, 2, 2)
	g.WithWriteLock(func() { g.ScryForEffect(me.ID, uuid.Nil, 1) })
	scry := scryChoiceFor(g, me.ID)
	if scry == nil {
		t.Fatal("no scry prompt")
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("scrying should offer to pay {2}")
	}
	b06AddMana(me, "C", "C")
	hand := me.Hand.Size()
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := allCountersOn(t, g, proft)[game.CounterPlusOne]; got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d", got, hand+1)
	}
}

func TestRFDProftConsultingDeclinedPaysNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	proft := rfdPush(g, me.ID, "Proft, Consulting Detective", "Legendary Creature — Human Detective", rfdProftConsultingOracle, 2, 2)
	g.WithWriteLock(func() { g.SurveilThenForEffect(me.ID, uuid.Nil, 1, nil) })
	sv := surveilChoiceFor(g, me.ID)
	if sv == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(sv.ID, me.ID, sv.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := allCountersOn(t, g, proft)[game.CounterPlusOne]; got != 0 {
		t.Errorf("counters = %d, want 0", got)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand changed to %d, want %d", me.Hand.Size(), hand)
	}
}

// --- Proft, Sinister Mastermind ----------------------------------------

func TestRFDProftSinisterMastermindNeedsThresholdToCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Graveyard.Cards = nil
	for i := 0; i < 6; i++ {
		pushGraveyardCardTyped(me, "Filler", "Sorcery")
	}
	proft := rfdHandCard(me, "Proft, Sinister Mastermind", "Legendary Creature — Human Rogue", "", rfdProftSinisterOracle, 5, 5)
	toMainForCost(t, g)
	if err := g.CastSpell(me.ID, proft, game.CastSpellParams{}); err == nil {
		t.Fatal("six cards in the graveyard: the cast must be refused")
	}
	pushGraveyardCardTyped(me, "Seventh", "Sorcery")
	if err := g.CastSpell(me.ID, proft, game.CastSpellParams{}); err != nil {
		t.Fatalf("seven cards in the graveyard: CastSpell: %v", err)
	}
}

func TestRFDProftSinisterMastermindDiscardShrinksATarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := rfdPush(g, opp.ID, "Their Bear", "Creature — Bear", "", 4, 4)
	proft := rfdHandCard(me, "Proft, Sinister Mastermind", "Legendary Creature — Human Rogue", "", rfdProftSinisterOracle, 5, 5)
	b06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, proft, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility from hand: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(proft) {
		t.Error("Proft is discarded as the cost")
	}
	if p, tg := effectivePower(t, g, theirs), effectiveToughness(t, g, theirs); p != 1 || tg != 3 {
		t.Errorf("target = %d/%d, want 1/3", p, tg)
	}
}

// --- Rescue Girl, First Responder --------------------------------------

func TestRFDRescueGirlBouncesAnotherPermanentOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	girl := rfdPush(g, me.ID, "Rescue Girl, First Responder", "Legendary Creature — Human Cleric", rfdRescueGirlOracle, 1, 3)
	bear := rfdPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)

	if err := g.ActivateCatalogAbility(me.ID, girl, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: girl}},
	}); err == nil {
		t.Fatal("she can't target herself: it says another")
	}
	if err := g.ActivateCatalogAbility(me.ID, girl, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if rfdOnBattlefield(g, bear) || !me.Hand.Contains(bear) {
		t.Error("the Bear should be back in its owner's hand")
	}
}

func TestRFDRescueGirlRefusesOnAnotherPlayersTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	girl := rfdPush(g, me.ID, "Rescue Girl, First Responder", "Legendary Creature — Human Cleric", rfdRescueGirlOracle, 1, 3)
	bear := rfdPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	advanceTo(t, g, game.StepEnd)
	advanceTo(t, g, game.StepDraw)
	if g.Seats[g.Turn.ActiveSeat].ID == me.ID {
		t.Fatal("expected another seat's turn")
	}
	if err := g.ActivateCatalogAbility(me.ID, girl, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("activation must be refused off your own turn")
	}
}

// --- Ruric Thar, Biomagus ----------------------------------------------

func TestRFDRuricTharBiomagusHasProwessTwiceAndDrawsWhenAnOpponentTargetsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ruric := rfdPush(g, opp.ID, "Ruric Thar, Biomagus", "Legendary Creature — Ogre Crab Wizard", rfdRuricBiomagusOracle, 4, 6)
	if n := prowessCountOf(t, g, ruric); n != 2 {
		t.Fatalf("prowess instances = %d, want 2", n)
	}
	if !hasAbility(effectiveAbilities(t, g, ruric), "flying") {
		t.Error("Ruric Thar has flying")
	}
	hand := opp.Hand.Size()
	rfdBolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: ruric})
	if got := opp.Hand.Size(); got != hand+1 {
		t.Errorf("its controller's hand = %d, want %d (drew for the opponent's targeting)", got, hand+1)
	}
	_ = me
}

func TestRFDRuricTharBiomagusIgnoresItsOwnControllersTargeting(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ruric := rfdPush(g, me.ID, "Ruric Thar, Biomagus", "Legendary Creature — Ogre Crab Wizard", rfdRuricBiomagusOracle, 4, 6)
	hand := me.Hand.Size()
	rfdBolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: ruric})
	// The Bolt itself is the cast card (+1 pushed, -1 cast) so the hand is level.
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand = %d, want %d: targeting your own creature draws nothing", got, hand)
	}
}

// --- Ruric Thar, Magecrusher -------------------------------------------

func TestRFDRuricTharMagecrusherCantBeCounteredAndHasItsKeywords(t *testing.T) {
	spec, ok := Lookup(rfdRuricMagecrusherOracle)
	if !ok {
		t.Fatal("Ruric Thar, Magecrusher is not registered")
	}
	if !spec.CantBeCountered {
		t.Error("this spell can't be countered")
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	ruric := rfdPush(g, me.ID, "Ruric Thar, Magecrusher", "Legendary Creature — Ogre Warrior", rfdRuricMagecrusherOracle, 7, 7)
	for _, kw := range []string{"reach", "vigilance", "trample"} {
		if !hasAbility(effectiveAbilities(t, g, ruric), kw) {
			t.Errorf("missing %s", kw)
		}
	}
	if hasAbility(effectiveAbilities(t, g, ruric), "hexproof") {
		t.Error("the declared caveat is that it never has hexproof; granting it would be a stronger card")
	}
}

// --- Saheeli, Consul of Oversight --------------------------------------

func TestRFDSaheeliConsulMakesOneThopterPerTurnOnScryOrSurveil(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rfdPush(g, me.ID, "Saheeli, Consul of Oversight", "Legendary Creature — Human Advisor", rfdSaheeliConsulOracle, 4, 4)

	g.WithWriteLock(func() { g.SurveilThenForEffect(me.ID, uuid.Nil, 1, nil) })
	sv := surveilChoiceFor(g, me.ID)
	if sv == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(sv.ID, me.ID, sv.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := rfdCountNamed(g, me.ID, "Thopter"); got != 1 {
		t.Fatalf("Thopters after a surveil = %d, want 1", got)
	}

	g.WithWriteLock(func() { g.ScryForEffect(me.ID, uuid.Nil, 1) })
	scry := scryChoiceFor(g, me.ID)
	if scry == nil {
		t.Fatal("no scry prompt")
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := rfdCountNamed(g, me.ID, "Thopter"); got != 1 {
		t.Errorf("a second scry the same turn made another Thopter: %d", got)
	}
}

// --- Saheeli, Jewel of Avishkar ----------------------------------------

func TestRFDSaheeliJewelMakesThoptersOnNoncreatureSpellsAndGivesThemHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rfdPush(g, me.ID, "Saheeli, Jewel of Avishkar", "Legendary Creature — Human Artificer", rfdSaheeliJewelOracle, 2, 4)
	rfdBolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if got := rfdCountNamed(g, me.ID, "Thopter"); got != 1 {
		t.Fatalf("Thopters = %d, want 1", got)
	}
	var thopter uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Thopter" {
			thopter = c.InstanceID
		}
	}
	if !hasAbility(effectiveAbilities(t, g, thopter), "haste") {
		t.Error("my Thopters have haste")
	}
	rfdCast(t, g, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	if got := rfdCountNamed(g, me.ID, "Thopter"); got != 1 {
		t.Errorf("a creature spell made a Thopter: %d", got)
	}
}

// --- Ob Nixilis, the Ascended ------------------------------------------

func TestRFDObNixilisDestroysTappedOpposingCreaturesAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped1 := rfdPush(g, opp.ID, "Tapped One", "Creature — Bear", "", 2, 2)
	tapped2 := rfdPush(g, g.Seats[2].ID, "Tapped Two", "Creature — Bear", "", 2, 2)
	untapped := rfdPush(g, opp.ID, "Untapped", "Creature — Bear", "", 2, 2)
	mineTapped := rfdPush(g, me.ID, "Mine Tapped", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case tapped1, tapped2, mineTapped:
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	life := me.Life
	rfdCast(t, g, "Ob Nixilis, the Ascended", "Legendary Creature — Angel", rfdObNixilisOracle, 4, 4)
	if rfdOnBattlefield(g, tapped1) || rfdOnBattlefield(g, tapped2) {
		t.Error("tapped opposing creatures should be destroyed")
	}
	if !rfdOnBattlefield(g, untapped) {
		t.Error("an untapped creature survives")
	}
	if !rfdOnBattlefield(g, mineTapped) {
		t.Error("my own tapped creature survives")
	}
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d (one per creature destroyed)", me.Life, life+2)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := rfdCountNamed(g, me.ID, "Angel"); got != 1 {
		t.Errorf("Angel tokens at the end step = %d, want 1 (life was gained this turn)", got)
	}
}

func TestRFDObNixilisMakesNoAngelWithoutLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rfdPush(g, me.ID, "Ob Nixilis, the Ascended", "Legendary Creature — Angel", rfdObNixilisOracle, 4, 4)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := rfdCountNamed(g, me.ID, "Angel"); got != 0 {
		t.Errorf("Angel tokens = %d, want 0", got)
	}
}

// --- Omnath, Locus of the Void -----------------------------------------

func TestRFDOmnathGrowsWithUnspentManaKeepsItColorlessAndLandfallAddsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	omnath := rfdPush(g, me.ID, "Omnath, Locus of the Void", "Legendary Creature — Elemental", rfdOmnathOracle, 6, 6)
	if got := effectivePower(t, g, omnath); got != 6 {
		t.Fatalf("power with an empty pool = %d, want 6", got)
	}
	floatMana(t, g, me, "{R}{G}{C}")
	if got := effectivePower(t, g, omnath); got != 9 {
		t.Errorf("power with three unspent mana = %d, want 9", got)
	}

	// Mana that would empty becomes colorless instead.
	advanceTo(t, g, game.StepBeginCombat)
	if len(me.ManaPool) != 3 {
		t.Fatalf("pool = %d mana after a step ended, want 3 kept", len(me.ManaPool))
	}
	for _, tok := range me.ManaPool {
		if tok.Color != "C" {
			t.Errorf("kept mana colour = %q, want colorless", tok.Color)
		}
	}

	before := len(me.ManaPool)
	b31PlayLand(t, g, "Plains", "Basic Land — Plains", "")
	passPriorityAroundTable(t, g)
	if got := len(me.ManaPool); got != before+2 {
		t.Errorf("pool after landfall = %d, want %d", got, before+2)
	}
}
