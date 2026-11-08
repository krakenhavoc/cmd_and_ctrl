package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// otj_s2_cards_test.go — Outlaws of Thunder Junction (OTJ), the S cards,
// second batch (#4468). Each test plays the real catalog entry through the
// engine and asserts what the printed text says.

const (
	otjS2SelvalaOracle      = "95b20721-1af9-4f3f-a39b-7cd1d4ce0b20"
	otjS2ServantOracle      = "0c9ec9a1-738a-4f69-b367-5dbf3d046bf1"
	otjS2ShackleOracle      = "59e8747a-3e3f-48d2-aaee-af94ecee378b"
	otjS2ShepherdOracle     = "c0535854-e4e7-43e4-bc5c-7cc95fb17929"
	otjS2SheriffOracle      = "ec16008f-18d6-4461-b99c-0a504f211618"
	otjS2GriftOracle        = "763d60ea-a0ad-4f13-b8a7-aa1515283266"
	otjS2SilverDeputyOracle = "7bbede9a-8981-4435-9854-af61a18c6166"
	otjS2SkulduggeryOracle  = "9100530e-2bd0-487e-a4eb-13dabd55b678"
	otjS2LockpickerOracle   = "e6e4ad3c-079d-429b-9f29-b9e42a1a236f"
	otjS2ArmadilloOracle    = "8c81f7b6-38e7-482b-afff-727074a23888"
	otjS2StopColdOracle     = "e0206af2-8034-4351-a797-a1d996f080b5"
)

func TestOtjS2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		otjS2SelvalaOracle:      "Selvala, Eager Trailblazer",
		otjS2ServantOracle:      "Servant of the Stinger",
		otjS2ShackleOracle:      "Shackle Slinger",
		otjS2ShepherdOracle:     "Shepherd of the Clouds",
		otjS2SheriffOracle:      "Sheriff of Safe Passage",
		otjS2GriftOracle:        "Shifting Grift",
		otjS2SilverDeputyOracle: "Silver Deputy",
		otjS2SkulduggeryOracle:  "Skulduggery",
		otjS2LockpickerOracle:   "Slickshot Lockpicker",
		otjS2ArmadilloOracle:    "Spinewoods Armadillo",
		otjS2StopColdOracle:     "Stop Cold",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	for oracle, cost := range map[string]string{
		otjS2SheriffOracle:    "{1}{W}",
		otjS2LockpickerOracle: "{2}{U}",
	} {
		if sa := plotOf(oracle); sa == nil || sa.Cost != cost {
			t.Errorf("%s: plot %+v, want %s", oracle, sa, cost)
		}
	}
}

// --- Selvala, Eager Trailblazer ----------------------------------------------

func TestOtjS2SelvalaMakesAMercenaryForACreatureSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	selvala := b12Push(g, me.ID, "Selvala, Eager Trailblazer", "Legendary Creature — Elf Scout", otjS2SelvalaOracle, 4, 5)
	if !hasEffectiveKeyword(t, g, selvala, "vigilance") {
		t.Error("vigilance is printed")
	}
	advanceToMain(t, g)
	otjB1BoltOpponent(t, g, opp.ID)
	if got := len(otjMPMercenaries(g, me.ID)); got != 0 {
		t.Fatalf("an instant made %d Mercenaries", got)
	}
	msc17Play(t, g, me, game.Card{Name: "Small Bear", TypeLine: "Creature — Bear", ManaCost: "{0}", Power: 2, Toughness: 2})
	otjMPOneMercenary(t, g, me.ID)
}

func TestOtjS2SelvalaAddsOneManaPerDifferentPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	selvala := b12Push(g, me.ID, "Selvala, Eager Trailblazer", "Legendary Creature — Elf Scout", otjS2SelvalaOracle, 4, 5)
	pushVanillaCreature(g, me.ID, "Two", 2, 2)
	pushVanillaCreature(g, me.ID, "Two again", 2, 3)
	pushVanillaCreature(g, me.ID, "One", 1, 1)
	pushVanillaCreature(g, g.Seats[1].ID, "Theirs", 7, 7)
	if err := g.ActivateManaAbility(me.ID, selvala, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no colour pick")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "W"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	// Powers 4, 2 and 1 among your creatures: three different powers.
	if got := poolColors(me); len(got) != 3 || got[0] != "W" || got[2] != "W" {
		t.Errorf("pool = %v, want {W}{W}{W}", got)
	}
}

// --- Servant of the Stinger --------------------------------------------------

func TestOtjS2ServantOfTheStingerTradesItselfForATutorAfterACrime(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Alpha", "Beta", "Gamma")
	servant := b12Push(g, me.ID, "Servant of the Stinger", "Creature — Human Warlock", otjS2ServantOracle, 1, 3)
	if !hasEffectiveKeyword(t, g, servant, "deathtouch") {
		t.Error("deathtouch is printed")
	}
	advanceToMain(t, g)
	otjB1BoltOpponent(t, g, opp.ID)
	lciB1Attack(t, g, servant, opp.ID)
	advanceTo(t, g, game.StepCombatDamage)
	ask := awaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	answerSearchNamed(t, g, me.ID, "Beta")
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(servant) || !me.Graveyard.Contains(servant) {
		t.Error("the Servant is sacrificed")
	}
	found := false
	for _, c := range me.Hand.Cards {
		if c.Name == "Beta" {
			found = true
		}
	}
	if !found {
		t.Error("the searched card goes to your hand")
	}
}

func TestOtjS2ServantOfTheStingerAsksNothingWithoutACrime(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Alpha", "Beta", "Gamma")
	servant := b12Push(g, me.ID, "Servant of the Stinger", "Creature — Human Warlock", otjS2ServantOracle, 1, 3)
	lciB1Attack(t, g, servant, opp.ID)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil || searchChoiceFor(g, me.ID) != nil {
		t.Fatal("no crime this turn: the ability does not trigger")
	}
	if !g.Battlefield.Contains(servant) {
		t.Error("the Servant stays")
	}
}

// --- Shackle Slinger ---------------------------------------------------------

func TestOtjS2ShackleSlingerTapsAnUntappedCreatureAndStunsATappedOne(t *testing.T) {
	for _, tapped := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		b12Push(g, me.ID, "Shackle Slinger", "Creature — Human Soldier", otjS2ShackleOracle, 3, 2)
		theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		if tapped {
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == theirs {
						g.Battlefield.Cards[i].Tapped = true
					}
				}
			})
		}
		b24CastNoncreature(t, g, "First", "Instant", "{U}")
		passPriorityAroundTable(t, g)
		if latestPickTarget(g, me.ID) != nil {
			t.Fatal("the first spell triggered it")
		}
		b24CastNoncreature(t, g, "Second", "Instant", "{U}")
		p := latestPickTarget(g, me.ID)
		if p == nil {
			passPriorityAroundTable(t, g)
			p = latestPickTarget(g, me.ID)
		}
		if p == nil {
			t.Fatal("the second spell asked for no target")
		}
		if hasID(p.PickTargetCards, mine) || !hasID(p.PickTargetCards, theirs) {
			t.Fatalf("target creature an opponent controls: %v", p.PickTargetCards)
		}
		pickCard(t, g, me.ID, theirs)
		passPriorityAroundTable(t, g)
		if !b16Tapped(t, g, theirs) {
			t.Errorf("tapped=%v: the creature ends tapped", tapped)
		}
		want := 0
		if tapped {
			want = 1
		}
		if got := countersOn(g, theirs, game.CounterStun); got != want {
			t.Errorf("tapped=%v: %d stun counters, want %d", tapped, got, want)
		}
	}
}

// --- Shepherd of the Clouds --------------------------------------------------

func TestOtjS2ShepherdOfTheCloudsReturnsToHandOrToTheBattlefieldWithAMount(t *testing.T) {
	for _, mount := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		small := gyCard(me, "Small Rock", "Artifact", "{3}", "")
		big := gyCard(me, "Big Rock", "Artifact", "{4}", "")
		spell := gyCard(me, "Old Bolt", "Instant", "{R}", "")
		if mount {
			b12Creature(g, me.ID, "My Horse", "Creature — Horse Mount", 2, 2)
		}
		shepherd := castHoldCreature(t, g, "Shepherd of the Clouds", "Creature — Pegasus", otjS2ShepherdOracle, 4, 3)
		if !hasEffectiveKeyword(t, g, shepherd, "flying") || !hasEffectiveKeyword(t, g, shepherd, "vigilance") {
			t.Error("flying and vigilance are printed")
		}
		p := latestPickTarget(g, me.ID)
		if p == nil {
			t.Fatal("the enters trigger asked for no target")
		}
		if hasID(p.PickTargetCards, big) || hasID(p.PickTargetCards, spell) || !hasID(p.PickTargetCards, small) {
			t.Fatalf("a permanent card with mana value 3 or less: %v", p.PickTargetCards)
		}
		pickCard(t, g, me.ID, small)
		passPriorityAroundTable(t, g)
		if mount {
			if !g.Battlefield.Contains(small) {
				t.Error("with a Mount the card returns to the battlefield")
			}
		} else if !handHasID(me, small) {
			t.Error("without a Mount the card returns to your hand")
		}
	}
}

// --- Sheriff of Safe Passage -------------------------------------------------

func TestOtjS2SheriffEntersWithACounterPlusOneForEachOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, me.ID, "One", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	b12Permanent(g, me.ID, "My Rock", "Artifact")
	b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	sheriff := castHoldCreature(t, g, "Sheriff of Safe Passage", "Creature — Human Knight", otjS2SheriffOracle, 0, 0)
	if got := plusOneCounters(g, sheriff); got != 3 {
		t.Fatalf("%d +1/+1 counters, want 3 (one plus two other creatures you control)", got)
	}
	if p, tt := ptOf(t, g, sheriff); p != 3 || tt != 3 {
		t.Errorf("the Sheriff is %d/%d, want 3/3", p, tt)
	}
}

func TestOtjS2SheriffAloneEntersWithOneCounter(t *testing.T) {
	g := newCatalogGame(t)
	sheriff := castHoldCreature(t, g, "Sheriff of Safe Passage", "Creature — Human Knight", otjS2SheriffOracle, 0, 0)
	if got := plusOneCounters(g, sheriff); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

// --- Shifting Grift ----------------------------------------------------------

func TestOtjS2ShiftingGriftExchangesCreaturesAndEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myBear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirDragon := b12Creature(g, opp.ID, "Their Dragon", "Creature — Dragon", 5, 5)
	myAura := b12Permanent(g, me.ID, "My Enchantment", "Enchantment")
	theirAura := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	myRock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	castModal(t, g, "Shifting Grift", "Sorcery", otjS2GriftOracle, []int{0, 2}, []game.TargetRef{
		modeRef(game.TargetCard, myBear, 0, 0),
		modeRef(game.TargetCard, theirDragon, 0, 0),
		modeRef(game.TargetCard, myAura, 1, 0),
		modeRef(game.TargetCard, theirAura, 1, 0),
	})
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]uuid.UUID{
		myBear: opp.ID, theirDragon: me.ID, myAura: opp.ID, theirAura: me.ID, myRock: me.ID,
	} {
		c, ok := battlefieldCard(g, id)
		if !ok {
			t.Fatalf("%s left the battlefield", id)
		}
		if c.Controller != want {
			t.Errorf("%s is controlled by %s, want %s", c.Name, c.Controller, want)
		}
	}
}

func TestOtjS2ShiftingGriftExchangesNothingWhenATargetIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myRock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	theirRock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	castModal(t, g, "Shifting Grift", "Sorcery", otjS2GriftOracle, []int{1}, []game.TargetRef{
		modeRef(game.TargetCard, myRock, 0, 0),
		modeRef(game.TargetCard, theirRock, 0, 0),
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirRock) })
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, myRock); !ok || c.Controller != me.ID {
		t.Error("with one target gone, no control is exchanged (CR 701.12b)")
	}
}

// --- Silver Deputy -----------------------------------------------------------

func TestOtjS2SilverDeputyPutsADesertOnTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Filler One", "Filler Two")
	desert := stapleLibraryCard(me, "Sandy Desert", "Land — Desert")
	stapleLibraryCard(me, "Not A Land", "Sorcery")
	advanceToMain(t, g)
	msc17Play(t, g, me, game.Card{Name: "Silver Deputy", TypeLine: "Artifact Creature — Mercenary",
		ManaCost: "{2}", OracleID: otjS2SilverDeputyOracle, Power: 1, Toughness: 2})
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Not A Land") != uuid.Nil {
		t.Error("a sorcery is neither a basic land card nor a Desert card")
	}
	answerSearchByID(t, g, me.ID, desert)
	passPriorityAroundTable(t, g)
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != desert {
		t.Errorf("the top card is %s, want the Desert", top.Name)
	}
}

func TestOtjS2SilverDeputyPumpsAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	deputy := b12Push(g, me.ID, "Silver Deputy", "Artifact Creature — Mercenary", otjS2SilverDeputyOracle, 1, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	if err := g.ActivateCatalogAbility(me.ID, deputy, 0, game.ActivateAbilityParams{Targets: b16TargetCard(mine)}); err == nil {
		t.Fatal("activate only as a sorcery: refused with a spell on the stack")
	}
	passPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, deputy, 0, game.ActivateAbilityParams{Targets: b16TargetCard(theirs)}); err == nil {
		t.Fatal("an opponent's creature is not a creature you control")
	}
	b16Activate(t, g, me.ID, deputy, 0, game.ActivateAbilityParams{Targets: b16TargetCard(mine)})
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, mine); p != 3 || tt != 2 {
		t.Errorf("your Bear is %d/%d, want 3/2", p, tt)
	}
	if !b16Tapped(t, g, deputy) {
		t.Error("{T} is the cost")
	}
}

// --- Skulduggery -------------------------------------------------------------

func TestOtjS2SkulduggeryGrowsYoursAndShrinksTheirs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := castCatalogSpellErr(t, g, "Skulduggery", "Instant", otjS2SkulduggeryOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}, {Kind: game.TargetCard, ID: mine}}); err == nil {
		t.Fatal("the clauses are bound to a side: yours first, then theirs")
	}
	castCatalogSpell(t, g, "Skulduggery", "Instant", otjS2SkulduggeryOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, mine); p != 3 || tt != 3 {
		t.Errorf("your Bear is %d/%d, want 3/3", p, tt)
	}
	if p, tt := ptOf(t, g, theirs); p != 1 || tt != 1 {
		t.Errorf("their Bear is %d/%d, want 1/1", p, tt)
	}
	if !YouCommittedACrimeThisTurn(g, me.ID) {
		t.Error("targeting an opponent's creature is a crime")
	}
}

// --- Slickshot Lockpicker ----------------------------------------------------

func TestOtjS2SlickshotLockpickerGrantsFlashbackAtTheCardsManaCost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	rock := gyCard(me, "Old Rock", "Artifact", "{1}", "")
	castHoldCreature(t, g, "Slickshot Lockpicker", "Creature — Human Rogue", otjS2LockpickerOracle, 2, 3)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if hasID(p.PickTargetCards, rock) || !hasID(p.PickTargetCards, bolt) {
		t.Fatalf("target instant or sorcery card in your graveyard: %v", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	life := opp.Life
	b06AddMana(me, "R")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Targets: b16TargetPlayer(opp.ID),
	}); err != nil {
		t.Fatalf("flashback from the graveyard for {R}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-3)
	}
	if !g.Exile.Contains(bolt) {
		t.Error("a flashed-back card is exiled")
	}
}

// --- Spinewoods Armadillo ----------------------------------------------------

func TestOtjS2SpinewoodsArmadilloDiscardsToFindALandAndGainThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	desert := stapleLibraryCard(me, "Sandy Desert", "Land — Desert")
	stapleLibraryCard(me, "Not A Land", "Sorcery")
	advanceToMain(t, g)
	armadillo := pushCatalogHandCard(me, "Spinewoods Armadillo", "Creature — Armadillo", otjS2ArmadilloOracle)
	life := me.Life
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, armadillo, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	if !me.Graveyard.Contains(armadillo) {
		t.Fatal("discarding the Armadillo is the cost")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Not A Land") != uuid.Nil {
		t.Error("a sorcery is neither a basic land card nor a Desert card")
	}
	answerSearchByID(t, g, me.ID, desert)
	passPriorityAroundTable(t, g)
	if !handHasID(me, desert) {
		t.Error("the Desert goes to your hand")
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	spec, _ := Lookup(otjS2ArmadilloOracle)
	if len(spec.Triggered) != 1 || len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "reach" {
		t.Errorf("want reach and one ward trigger: %v, %d triggers", spec.PrintedKeywords, len(spec.Triggered))
	}
}

// --- Stop Cold ---------------------------------------------------------------

func TestOtjS2StopColdTapsAndSilencesAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	land := pushPermanent(g, opp.ID, game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"})
	ring := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: solRingOracle, Owner: opp.ID, Controller: opp.ID,
	})
	advanceToMain(t, g)
	if err := castCatalogSpellErr(t, g, "Stop Cold", "Enchantment — Aura", otjS2StopColdOracle, b16TargetCard(land)); err == nil {
		t.Fatal("a land is not an artifact or a creature")
	}
	castCatalogSpell(t, g, "Stop Cold", "Enchantment — Aura", otjS2StopColdOracle, b16TargetCard(ring))
	passPriorityAroundTable(t, g)
	if !dftCG2Tapped(g, ring) {
		t.Error("the enchanted artifact is tapped")
	}
	if got := game.ManaAbilitiesForCard(layeredCard(t, g, ring)); len(got) != 0 {
		t.Errorf("mana abilities = %v, want none", got)
	}
}

func TestOtjS2StopColdKeepsACreatureTapped(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Stop Cold", "Enchantment — Aura", otjS2StopColdOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if !dftCG2Tapped(g, bear) {
		t.Fatal("the enchanted creature is not tapped")
	}
	dftMPPassTurn(t, g)
	if g.Seats[g.Turn.ActiveSeat].ID != opp.ID {
		t.Fatalf("the next turn is not the creature's controller's: seat %d", g.Turn.ActiveSeat)
	}
	advanceToMain(t, g)
	if !dftCG2Tapped(g, bear) {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
	spec, _ := Lookup(otjS2StopColdOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("flash is printed: %v", spec.PrintedKeywords)
	}
}
