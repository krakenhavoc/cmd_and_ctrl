package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rfdPush puts a catalog creature (or other permanent) with a real
// power and toughness onto the battlefield, ready to act.
func rfdPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func rfdBolt(t *testing.T, g *game.Game, target game.TargetRef) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{target})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestRFDCardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.RfdMarwynClearcutterOracle: "Marwyn, the Clearcutter",
		effectstest.RfdMarwynPreserverOracle:   "Marwyn, the Preserver",
		effectstest.RfdMasterOfBarbsOracle:     "Master of Barbs",
		effectstest.RfdMemnarchOracle:          "Memnarch, the Warden",
		effectstest.RfdNivMizzetOracle:         "Niv-Mizzet, Ghost Counsel",
		effectstest.RfdObNixilisOracle:         "Ob Nixilis, the Ascended",
		effectstest.RfdOmnathOracle:            "Omnath, Locus of the Void",
		effectstest.RfdPiaAetherOracle:         "Pia, Aether Ascetic",
		effectstest.RfdPiaRebuilderOracle:      "Pia, Determined Rebuilder",
		effectstest.RfdWitchstalkerOracle:      "Primal Witchstalker",
		effectstest.RfdProftConsultingOracle:   "Proft, Consulting Detective",
		effectstest.RfdProftSinisterOracle:     "Proft, Sinister Mastermind",
		effectstest.RfdRampartHunterOracle:     "Rampart Hunter",
		effectstest.RfdRankRatOracle:           "Rank Rat",
		effectstest.RfdRescueGirlOracle:        "Rescue Girl, First Responder",
		effectstest.RfdRuricBiomagusOracle:     "Ruric Thar, Biomagus",
		effectstest.RfdRuricMagecrusherOracle:  "Ruric Thar, Magecrusher",
		effectstest.RfdSaheeliConsulOracle:     "Saheeli, Consul of Oversight",
		effectstest.RfdSaheeliJewelOracle:      "Saheeli, Jewel of Avishkar",
	} {
		spec, ok := effects.Lookup(oracle)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Hand.Size()
	}
	effectstest.RfdCast(t, g, "Rank Rat", "Creature — Zombie Rat", effectstest.RfdRankRatOracle, 1, 1)
	for _, p := range g.Seats[1:] {
		if effectstest.DiscardChoiceFor(g, p.ID) == nil {
			t.Fatalf("opponent %s was not asked to discard", p.ID)
		}
	}
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("the caster must not be asked to discard")
	}
	for _, p := range g.Seats[1:] {
		effectstest.DiscardFromHand(t, g, p.ID)
		if got := p.Hand.Size(); got != before[p.ID]-1 {
			t.Errorf("opponent hand = %d, want %d", got, before[p.ID]-1)
		}
	}
}

// --- Rampart Hunter ----------------------------------------------------

func TestRFDRampartHunterPumpsAndGrantsDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := rfdPush(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.RfdHandCard(me, "x", "Land", "", "", 0, 0) // keep the hand non-empty
	hunter := effectstest.RfdHandCard(me, "Rampart Hunter", "Creature — Horror", "", effectstest.RfdRampartHunterOracle, 3, 3)
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, hunter, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("the enters trigger should ask for a target creature")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, theirs); got != 4 {
		t.Errorf("target power = %d, want 4", got)
	}
	if got := effectstest.EffectiveToughness(t, g, theirs); got != 4 {
		t.Errorf("target toughness = %d, want 4", got)
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, theirs), "deathtouch") {
		t.Error("the target should gain deathtouch")
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, hunter), "deathtouch") {
		t.Error("Rampart Hunter has deathtouch itself")
	}
}

// --- Pia, Determined Rebuilder ------------------------------------------

func TestRFDPiaDeterminedRebuilderMakesAThopterAndPumpsPerArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pia := effectstest.RfdCast(t, g, "Pia, Determined Rebuilder", "Legendary Creature — Human Artificer", effectstest.RfdPiaRebuilderOracle, 2, 2)
	if got := effectstest.RfdCountNamed(g, me.ID, "Thopter"); got != 1 {
		t.Fatalf("Thopters = %d, want 1", got)
	}
	effectstest.PushPermanentForTest(g, me.ID, "Trinket", "", "Artifact")
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.FillPool(me, 5)
	if err := g.ActivateCatalogAbility(me.ID, pia, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: pia}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Two artifacts: the Thopter token and the Trinket.
	if got := effectstest.EffectivePower(t, g, pia); got != 4 {
		t.Errorf("Pia power = %d, want 4 (2 + two artifacts)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, pia); got != 2 {
		t.Errorf("Pia toughness = %d, want 2 (+X/+0)", got)
	}
}

// --- Memnarch, the Warden ----------------------------------------------

func TestRFDMemnarchMakesTwoMyrAndDrawsPerArtifactOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mem := effectstest.RfdCast(t, g, "Memnarch, the Warden", "Legendary Artifact Creature — Wizard", effectstest.RfdMemnarchOracle, 8, 9)
	if got := effectstest.RfdCountNamed(g, me.ID, "Myr"); got != 2 {
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
	effectstest.DeclareAttack(t, g, opp.ID, mem)
	effectstest.PassPriorityAroundTable(t, g)
	// Memnarch and two Myr: three artifacts, three cards.
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand = %d, want %d (a card per artifact)", got, hand+3)
	}
}

// --- Marwyn, the Clearcutter --------------------------------------------

func TestRFDMarwynClearcutterSacrificesAnArtifactOrLandToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	marwyn := rfdPush(g, me.ID, "Marwyn, the Clearcutter", "Legendary Creature — Elf Warrior", effectstest.RfdMarwynClearcutterOracle, 2, 1)
	bear := rfdPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	rock := effectstest.PushPermanentForTest(g, me.ID, "Rock", "", "Artifact")
	effectstest.FillPool(me, 2)
	hand := me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, marwyn, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a plain creature is not an artifact or land")
	}
	if !effectstest.RfdOnBattlefield(g, bear) || me.Hand.Size() != hand {
		t.Fatal("a refused activation must not pay any part of the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, marwyn, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RfdOnBattlefield(g, rock) {
		t.Error("the artifact should have been sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d", got, hand+1)
	}
}

// --- Marwyn, the Preserver ---------------------------------------------

func TestRFDMarwynPreserverGivesYourLandsHexproofAndRecurs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marwyn := rfdPush(g, me.ID, "Marwyn, the Preserver", "Legendary Creature — Elf Druid", effectstest.RfdMarwynPreserverOracle, 3, 2)
	mine := rfdPush(g, me.ID, "My Forest", "Basic Land — Forest", "", 0, 0)
	theirs := rfdPush(g, opp.ID, "Their Forest", "Basic Land — Forest", "", 0, 0)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, mine), "hexproof") {
		t.Error("my land should have hexproof")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, theirs), "hexproof") {
		t.Error("an opponent's land must not gain hexproof")
	}

	land := effectstest.PushGraveyardCardTyped(me, "Dead Forest", "Basic Land — Forest")
	spell := effectstest.PushGraveyardCardTyped(me, "Dead Spell", "Sorcery")
	effectstest.FillPool(me, 4)
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
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the land card should be back in hand")
	}
}

// --- Master of Barbs ---------------------------------------------------

func TestRFDMasterOfBarbsPumpsOnNoncombatDamageToAnOpponentOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := rfdPush(g, me.ID, "Master of Barbs", "Creature — Lizard Bard", effectstest.RfdMasterOfBarbsOracle, 2, 1)
	other := rfdPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := rfdPush(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)

	// Damage to a creature is not damage to an opponent.
	rfdBolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: theirs})
	if got := effectstest.EffectivePower(t, g, master); got != 2 {
		t.Fatalf("power after bolting a creature = %d, want 2", got)
	}
	rfdBolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if got := effectstest.EffectivePower(t, g, master); got != 3 {
		t.Errorf("Master power = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, other); got != 3 {
		t.Errorf("the other creature's power = %d, want 3", got)
	}
}

func TestRFDMasterOfBarbsIgnoresCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := rfdPush(g, me.ID, "Master of Barbs", "Creature — Lizard Bard", effectstest.RfdMasterOfBarbsOracle, 2, 1)
	effectstest.AttackWith(t, g, opp.ID, master)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life == 40 {
		t.Fatal("the attack should have connected")
	}
	if got := effectstest.EffectivePower(t, g, master); got != 2 {
		t.Errorf("combat damage must not trigger it: power = %d, want 2", got)
	}
}

// --- Niv-Mizzet, Ghost Counsel -----------------------------------------

func TestRFDNivMizzetPaysLifeToDrawThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rfdPush(g, me.ID, "Niv-Mizzet, Ghost Counsel", "Legendary Creature — Spirit Dragon", effectstest.RfdNivMizzetOracle, 4, 4)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	life, hand := me.Life, me.Hand.Size()
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life = %d, want %d (paid 3)", me.Life, life-3)
	}
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand = %d, want %d", got, hand+3)
	}
}

func TestRFDNivMizzetDeclineDoesNothingAndDrainFeedsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	niv := rfdPush(g, me.ID, "Niv-Mizzet, Ghost Counsel", "Legendary Creature — Spirit Dragon", effectstest.RfdNivMizzetOracle, 4, 4)
	life, oppLife, hand := me.Life, opp.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, niv, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != life+1 {
		t.Fatalf("drain: opp %d, me %d; want %d and %d", opp.Life, me.Life, oppLife-1, life+1)
	}
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 || me.Hand.Size() != hand {
		t.Errorf("declined: life %d hand %d, want %d and %d", me.Life, me.Hand.Size(), life+1, hand)
	}
}

// --- Pia, Aether Ascetic -----------------------------------------------

func TestRFDPiaAetherAsceticDiscardsToTutorAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ench := effectstest.PushLibraryCardForTest(me, game.Card{
		InstanceID: uuid.New(), Name: "Test Enchantment", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID,
	})
	effectstest.RfdCast(t, g, "Pia, Aether Ascetic", "Legendary Creature — Human Druid", effectstest.RfdPiaAetherOracle, 2, 2)
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the controller should be offered a discard")
	}
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if !me.Hand.Contains(ench) {
		t.Error("the enchantment should have been searched into hand")
	}
}

func TestRFDPiaAetherAsceticDecliningTheDiscardSearchesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ench := effectstest.PushLibraryCardForTest(me, game.Card{
		InstanceID: uuid.New(), Name: "Test Enchantment", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID,
	})
	effectstest.RfdCast(t, g, "Pia, Aether Ascetic", "Legendary Creature — Human Druid", effectstest.RfdPiaAetherOracle, 2, 2)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(ench) {
		t.Error("no discard, no search")
	}
}

// --- Primal Witchstalker -----------------------------------------------

func TestRFDPrimalWitchstalkerMillsFourAndReturnsALandTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushGraveyardCardTyped(me, "Dead Forest", "Basic Land — Forest")
	lib := me.Library.Size()
	effectstest.RfdCast(t, g, "Primal Witchstalker", "Creature — Wolf", effectstest.RfdWitchstalkerOracle, 2, 1)
	if got := me.Library.Size(); got != lib-4 {
		t.Fatalf("library = %d, want %d (milled four)", got, lib-4)
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, land)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, land)
	if !ok {
		t.Fatal("the land card should have returned to the battlefield")
	}
	if !c.Tapped {
		t.Error("it returns tapped")
	}
}

func TestRFDPrimalWitchstalkerWithNoLandStillMills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Graveyard.Cards = nil
	lib := me.Library.Size()
	effectstest.RfdCast(t, g, "Primal Witchstalker", "Creature — Wolf", effectstest.RfdWitchstalkerOracle, 2, 1)
	// The library is basic fillers (no land type line), so nothing is a
	// legal target and the reflexive trigger is removed.
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("no land card in the graveyard: no prompt")
	}
	if got := me.Library.Size(); got != lib-4 {
		t.Errorf("library = %d, want %d", got, lib-4)
	}
}

// --- Proft, Consulting Detective ---------------------------------------

func TestRFDProftConsultingPaysTwoForACounterAndACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	proft := rfdPush(g, me.ID, "Proft, Consulting Detective", "Legendary Creature — Human Detective", effectstest.RfdProftConsultingOracle, 2, 2)
	g.WithWriteLock(func() { g.ScryForEffect(me.ID, uuid.Nil, 1) })
	scry := effectstest.ScryChoiceFor(g, me.ID)
	if scry == nil {
		t.Fatal("no scry prompt")
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("scrying should offer to pay {2}")
	}
	effectstest.B06AddMana(me, "C", "C")
	hand := me.Hand.Size()
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AllCountersOn(t, g, proft)[game.CounterPlusOne]; got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d", got, hand+1)
	}
}

func TestRFDProftConsultingDeclinedPaysNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	proft := rfdPush(g, me.ID, "Proft, Consulting Detective", "Legendary Creature — Human Detective", effectstest.RfdProftConsultingOracle, 2, 2)
	g.WithWriteLock(func() { g.SurveilKeyedForEffect(me.ID, uuid.Nil, 1, game.KeyedTail{}) })
	sv := effectstest.SurveilChoiceFor(g, me.ID)
	if sv == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(sv.ID, me.ID, sv.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AllCountersOn(t, g, proft)[game.CounterPlusOne]; got != 0 {
		t.Errorf("counters = %d, want 0", got)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand changed to %d, want %d", me.Hand.Size(), hand)
	}
}

// --- Proft, Sinister Mastermind ----------------------------------------

func TestRFDProftSinisterMastermindNeedsThresholdToCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Graveyard.Cards = nil
	for i := 0; i < 6; i++ {
		effectstest.PushGraveyardCardTyped(me, "Filler", "Sorcery")
	}
	proft := effectstest.RfdHandCard(me, "Proft, Sinister Mastermind", "Legendary Creature — Human Rogue", "", effectstest.RfdProftSinisterOracle, 5, 5)
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, proft, game.CastSpellParams{}); err == nil {
		t.Fatal("six cards in the graveyard: the cast must be refused")
	}
	effectstest.PushGraveyardCardTyped(me, "Seventh", "Sorcery")
	if err := g.CastSpell(me.ID, proft, game.CastSpellParams{}); err != nil {
		t.Fatalf("seven cards in the graveyard: CastSpell: %v", err)
	}
}

func TestRFDProftSinisterMastermindDiscardShrinksATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := rfdPush(g, opp.ID, "Their Bear", "Creature — Bear", "", 4, 4)
	proft := effectstest.RfdHandCard(me, "Proft, Sinister Mastermind", "Legendary Creature — Human Rogue", "", effectstest.RfdProftSinisterOracle, 5, 5)
	effectstest.B06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, proft, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility from hand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(proft) {
		t.Error("Proft is discarded as the cost")
	}
	if p, tg := effectstest.EffectivePower(t, g, theirs), effectstest.EffectiveToughness(t, g, theirs); p != 1 || tg != 3 {
		t.Errorf("target = %d/%d, want 1/3", p, tg)
	}
}

// --- Rescue Girl, First Responder --------------------------------------

func TestRFDRescueGirlBouncesAnotherPermanentOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	girl := rfdPush(g, me.ID, "Rescue Girl, First Responder", "Legendary Creature — Human Cleric", effectstest.RfdRescueGirlOracle, 1, 3)
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
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RfdOnBattlefield(g, bear) || !me.Hand.Contains(bear) {
		t.Error("the Bear should be back in its owner's hand")
	}
}

func TestRFDRescueGirlRefusesOnAnotherPlayersTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	girl := rfdPush(g, me.ID, "Rescue Girl, First Responder", "Legendary Creature — Human Cleric", effectstest.RfdRescueGirlOracle, 1, 3)
	bear := rfdPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AdvanceTo(t, g, game.StepDraw)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ruric := rfdPush(g, opp.ID, "Ruric Thar, Biomagus", "Legendary Creature — Ogre Crab Wizard", effectstest.RfdRuricBiomagusOracle, 4, 6)
	if n := effectstest.ProwessCountOf(t, g, ruric); n != 2 {
		t.Fatalf("prowess instances = %d, want 2", n)
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, ruric), "flying") {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ruric := rfdPush(g, me.ID, "Ruric Thar, Biomagus", "Legendary Creature — Ogre Crab Wizard", effectstest.RfdRuricBiomagusOracle, 4, 6)
	hand := me.Hand.Size()
	rfdBolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: ruric})
	// The Bolt itself is the cast card (+1 pushed, -1 cast) so the hand is level.
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand = %d, want %d: targeting your own creature draws nothing", got, hand)
	}
}

// --- Ruric Thar, Magecrusher -------------------------------------------

func TestRFDRuricTharMagecrusherCantBeCounteredAndHasItsKeywords(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.RfdRuricMagecrusherOracle)
	if !ok {
		t.Fatal("Ruric Thar, Magecrusher is not registered")
	}
	if !spec.CantBeCountered {
		t.Error("this spell can't be countered")
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ruric := rfdPush(g, me.ID, "Ruric Thar, Magecrusher", "Legendary Creature — Ogre Warrior", effectstest.RfdRuricMagecrusherOracle, 7, 7)
	for _, kw := range []string{"reach", "vigilance", "trample"} {
		if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, ruric), kw) {
			t.Errorf("missing %s", kw)
		}
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, ruric), "hexproof") {
		t.Error("it has hexproof until it deals combat damage (#3172); ruric_thar_hexproof_test.go covers the turn-off")
	}
}

// --- Saheeli, Consul of Oversight --------------------------------------

func TestRFDSaheeliConsulMakesOneThopterPerTurnOnScryOrSurveil(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rfdPush(g, me.ID, "Saheeli, Consul of Oversight", "Legendary Creature — Human Advisor", effectstest.RfdSaheeliConsulOracle, 4, 4)

	g.WithWriteLock(func() { g.SurveilKeyedForEffect(me.ID, uuid.Nil, 1, game.KeyedTail{}) })
	sv := effectstest.SurveilChoiceFor(g, me.ID)
	if sv == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(sv.ID, me.ID, sv.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.RfdCountNamed(g, me.ID, "Thopter"); got != 1 {
		t.Fatalf("Thopters after a surveil = %d, want 1", got)
	}

	g.WithWriteLock(func() { g.ScryForEffect(me.ID, uuid.Nil, 1) })
	scry := effectstest.ScryChoiceFor(g, me.ID)
	if scry == nil {
		t.Fatal("no scry prompt")
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.RfdCountNamed(g, me.ID, "Thopter"); got != 1 {
		t.Errorf("a second scry the same turn made another Thopter: %d", got)
	}
}

// --- Saheeli, Jewel of Avishkar ----------------------------------------

func TestRFDSaheeliJewelMakesThoptersOnNoncreatureSpellsAndGivesThemHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rfdPush(g, me.ID, "Saheeli, Jewel of Avishkar", "Legendary Creature — Human Artificer", effectstest.RfdSaheeliJewelOracle, 2, 4)
	rfdBolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if got := effectstest.RfdCountNamed(g, me.ID, "Thopter"); got != 1 {
		t.Fatalf("Thopters = %d, want 1", got)
	}
	var thopter uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Thopter" {
			thopter = c.InstanceID
		}
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, thopter), "haste") {
		t.Error("my Thopters have haste")
	}
	effectstest.RfdCast(t, g, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	if got := effectstest.RfdCountNamed(g, me.ID, "Thopter"); got != 1 {
		t.Errorf("a creature spell made a Thopter: %d", got)
	}
}

// --- Ob Nixilis, the Ascended ------------------------------------------

func TestRFDObNixilisDestroysTappedOpposingCreaturesAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
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
	effectstest.RfdCast(t, g, "Ob Nixilis, the Ascended", "Legendary Creature — Angel", effectstest.RfdObNixilisOracle, 4, 4)
	if effectstest.RfdOnBattlefield(g, tapped1) || effectstest.RfdOnBattlefield(g, tapped2) {
		t.Error("tapped opposing creatures should be destroyed")
	}
	if !effectstest.RfdOnBattlefield(g, untapped) {
		t.Error("an untapped creature survives")
	}
	if !effectstest.RfdOnBattlefield(g, mineTapped) {
		t.Error("my own tapped creature survives")
	}
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d (one per creature destroyed)", me.Life, life+2)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.RfdCountNamed(g, me.ID, "Angel"); got != 1 {
		t.Errorf("Angel tokens at the end step = %d, want 1 (life was gained this turn)", got)
	}
}

func TestRFDObNixilisMakesNoAngelWithoutLifeGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rfdPush(g, me.ID, "Ob Nixilis, the Ascended", "Legendary Creature — Angel", effectstest.RfdObNixilisOracle, 4, 4)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.RfdCountNamed(g, me.ID, "Angel"); got != 0 {
		t.Errorf("Angel tokens = %d, want 0", got)
	}
}

// --- Omnath, Locus of the Void -----------------------------------------

func TestRFDOmnathGrowsWithUnspentManaKeepsItColorlessAndLandfallAddsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	omnath := rfdPush(g, me.ID, "Omnath, Locus of the Void", "Legendary Creature — Elemental", effectstest.RfdOmnathOracle, 6, 6)
	if got := effectstest.EffectivePower(t, g, omnath); got != 6 {
		t.Fatalf("power with an empty pool = %d, want 6", got)
	}
	effectstest.FloatMana(t, g, me, "{R}{G}{C}")
	if got := effectstest.EffectivePower(t, g, omnath); got != 9 {
		t.Errorf("power with three unspent mana = %d, want 9", got)
	}

	// Mana that would empty becomes colorless instead.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if me.ManaPool.Total() != 3 {
		t.Fatalf("pool = %d mana after a step ended, want 3 kept", me.ManaPool.Total())
	}
	for _, tok := range me.ManaPool.Units() {
		if tok.Color != "C" {
			t.Errorf("kept mana colour = %q, want colorless", tok.Color)
		}
	}

	before := me.ManaPool.Total()
	effectstest.B31PlayLand(t, g, "Plains", "Basic Land — Plains", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != before+2 {
		t.Errorf("pool after landfall = %d, want %d", got, before+2)
	}
}
