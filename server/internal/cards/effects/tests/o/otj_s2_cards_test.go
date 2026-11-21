package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	for oracle, cost := range map[string]string{
		otjS2SheriffOracle:    "{1}{W}",
		otjS2LockpickerOracle: "{2}{U}",
	} {
		if sa := effectstest.PlotOf(oracle); sa == nil || sa.Cost != cost {
			t.Errorf("%s: plot %+v, want %s", oracle, sa, cost)
		}
	}
}

// --- Selvala, Eager Trailblazer ----------------------------------------------

func TestOtjS2SelvalaMakesAMercenaryForACreatureSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	selvala := effectstest.B12Push(g, me.ID, "Selvala, Eager Trailblazer", "Legendary Creature — Elf Scout", otjS2SelvalaOracle, 4, 5)
	if !effectstest.HasEffectiveKeyword(t, g, selvala, "vigilance") {
		t.Error("vigilance is printed")
	}
	effectstest.AdvanceToMain(t, g)
	otjB1BoltOpponent(t, g, opp.ID)
	if got := len(effectstest.OtjMPMercenaries(g, me.ID)); got != 0 {
		t.Fatalf("an instant made %d Mercenaries", got)
	}
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Small Bear", TypeLine: "Creature — Bear", ManaCost: "{0}", Power: 2, Toughness: 2})
	effectstest.OtjMPOneMercenary(t, g, me.ID)
}

func TestOtjS2SelvalaAddsOneManaPerDifferentPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	selvala := effectstest.B12Push(g, me.ID, "Selvala, Eager Trailblazer", "Legendary Creature — Elf Scout", otjS2SelvalaOracle, 4, 5)
	effectstest.PushVanillaCreature(g, me.ID, "Two", 2, 2)
	effectstest.PushVanillaCreature(g, me.ID, "Two again", 2, 3)
	effectstest.PushVanillaCreature(g, me.ID, "One", 1, 1)
	effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Theirs", 7, 7)
	if err := g.ActivateManaAbility(me.ID, selvala, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no colour pick")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "W"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	// Powers 4, 2 and 1 among your creatures: three different powers.
	if got := effectstest.PoolColors(me); len(got) != 3 || got[0] != "W" || got[2] != "W" {
		t.Errorf("pool = %v, want {W}{W}{W}", got)
	}
}

// --- Servant of the Stinger --------------------------------------------------

func TestOtjS2ServantOfTheStingerTradesItselfForATutorAfterACrime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "Alpha", "Beta", "Gamma")
	servant := effectstest.B12Push(g, me.ID, "Servant of the Stinger", "Creature — Human Warlock", otjS2ServantOracle, 1, 3)
	if !effectstest.HasEffectiveKeyword(t, g, servant, "deathtouch") {
		t.Error("deathtouch is printed")
	}
	effectstest.AdvanceToMain(t, g)
	otjB1BoltOpponent(t, g, opp.ID)
	effectstest.LciB1Attack(t, g, servant, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	ask := effectstest.AwaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Beta")
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "Alpha", "Beta", "Gamma")
	servant := effectstest.B12Push(g, me.ID, "Servant of the Stinger", "Creature — Human Warlock", otjS2ServantOracle, 1, 3)
	effectstest.LciB1Attack(t, g, servant, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil || effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Fatal("no crime this turn: the ability does not trigger")
	}
	if !g.Battlefield.Contains(servant) {
		t.Error("the Servant stays")
	}
}

// --- Shackle Slinger ---------------------------------------------------------

func TestOtjS2ShackleSlingerTapsAnUntappedCreatureAndStunsATappedOne(t *testing.T) {
	for _, tapped := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Shackle Slinger", "Creature — Human Soldier", otjS2ShackleOracle, 3, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		if tapped {
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == theirs {
						g.Battlefield.Cards[i].Tapped = true
					}
				}
			})
		}
		effectstest.B24CastNoncreature(t, g, "First", "Instant", "{U}")
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			t.Fatal("the first spell triggered it")
		}
		effectstest.B24CastNoncreature(t, g, "Second", "Instant", "{U}")
		p := effectstest.LatestPickTarget(g, me.ID)
		if p == nil {
			effectstest.PassPriorityAroundTable(t, g)
			p = effectstest.LatestPickTarget(g, me.ID)
		}
		if p == nil {
			t.Fatal("the second spell asked for no target")
		}
		if effectstest.HasID(p.PickTargetCards, mine) || !effectstest.HasID(p.PickTargetCards, theirs) {
			t.Fatalf("target creature an opponent controls: %v", p.PickTargetCards)
		}
		effectstest.PickCard(t, g, me.ID, theirs)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.B16Tapped(t, g, theirs) {
			t.Errorf("tapped=%v: the creature ends tapped", tapped)
		}
		want := 0
		if tapped {
			want = 1
		}
		if got := effectstest.CountersOn(g, theirs, game.CounterStun); got != want {
			t.Errorf("tapped=%v: %d stun counters, want %d", tapped, got, want)
		}
	}
}

// --- Shepherd of the Clouds --------------------------------------------------

func TestOtjS2ShepherdOfTheCloudsReturnsToHandOrToTheBattlefieldWithAMount(t *testing.T) {
	for _, mount := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		small := effectstest.GyCard(me, "Small Rock", "Artifact", "{3}", "")
		big := effectstest.GyCard(me, "Big Rock", "Artifact", "{4}", "")
		spell := effectstest.GyCard(me, "Old Bolt", "Instant", "{R}", "")
		if mount {
			effectstest.B12Creature(g, me.ID, "My Horse", "Creature — Horse Mount", 2, 2)
		}
		shepherd := effectstest.CastHoldCreature(t, g, "Shepherd of the Clouds", "Creature — Pegasus", otjS2ShepherdOracle, 4, 3)
		if !effectstest.HasEffectiveKeyword(t, g, shepherd, "flying") || !effectstest.HasEffectiveKeyword(t, g, shepherd, "vigilance") {
			t.Error("flying and vigilance are printed")
		}
		p := effectstest.LatestPickTarget(g, me.ID)
		if p == nil {
			t.Fatal("the enters trigger asked for no target")
		}
		if effectstest.HasID(p.PickTargetCards, big) || effectstest.HasID(p.PickTargetCards, spell) || !effectstest.HasID(p.PickTargetCards, small) {
			t.Fatalf("a permanent card with mana value 3 or less: %v", p.PickTargetCards)
		}
		effectstest.PickCard(t, g, me.ID, small)
		effectstest.PassPriorityAroundTable(t, g)
		if mount {
			if !g.Battlefield.Contains(small) {
				t.Error("with a Mount the card returns to the battlefield")
			}
		} else if !effectstest.HandHasID(me, small) {
			t.Error("without a Mount the card returns to your hand")
		}
	}
}

// --- Sheriff of Safe Passage -------------------------------------------------

func TestOtjS2SheriffEntersWithACounterPlusOneForEachOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "One", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	sheriff := effectstest.CastHoldCreature(t, g, "Sheriff of Safe Passage", "Creature — Human Knight", otjS2SheriffOracle, 0, 0)
	if got := effectstest.PlusOneCounters(g, sheriff); got != 3 {
		t.Fatalf("%d +1/+1 counters, want 3 (one plus two other creatures you control)", got)
	}
	if p, tt := effectstest.PtOf(t, g, sheriff); p != 3 || tt != 3 {
		t.Errorf("the Sheriff is %d/%d, want 3/3", p, tt)
	}
}

func TestOtjS2SheriffAloneEntersWithOneCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	sheriff := effectstest.CastHoldCreature(t, g, "Sheriff of Safe Passage", "Creature — Human Knight", otjS2SheriffOracle, 0, 0)
	if got := effectstest.PlusOneCounters(g, sheriff); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

// --- Shifting Grift ----------------------------------------------------------

func TestOtjS2ShiftingGriftExchangesCreaturesAndEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myBear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirDragon := effectstest.B12Creature(g, opp.ID, "Their Dragon", "Creature — Dragon", 5, 5)
	myAura := effectstest.B12Permanent(g, me.ID, "My Enchantment", "Enchantment")
	theirAura := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	myRock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.CastModal(t, g, "Shifting Grift", "Sorcery", otjS2GriftOracle, []int{0, 2}, []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, myBear, 0, 0),
		effectstest.ModeRef(game.TargetCard, theirDragon, 0, 0),
		effectstest.ModeRef(game.TargetCard, myAura, 1, 0),
		effectstest.ModeRef(game.TargetCard, theirAura, 1, 0),
	})
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]uuid.UUID{
		myBear: opp.ID, theirDragon: me.ID, myAura: opp.ID, theirAura: me.ID, myRock: me.ID,
	} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Fatalf("%s left the battlefield", id)
		}
		if c.Controller != want {
			t.Errorf("%s is controlled by %s, want %s", c.Name, c.Controller, want)
		}
	}
}

func TestOtjS2ShiftingGriftExchangesNothingWhenATargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myRock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirRock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.CastModal(t, g, "Shifting Grift", "Sorcery", otjS2GriftOracle, []int{1}, []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, myRock, 0, 0),
		effectstest.ModeRef(game.TargetCard, theirRock, 0, 0),
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirRock) })
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, myRock); !ok || c.Controller != me.ID {
		t.Error("with one target gone, no control is exchanged (CR 701.12a)")
	}
}

// --- Silver Deputy -----------------------------------------------------------

func TestOtjS2SilverDeputyPutsADesertOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Filler One", "Filler Two")
	desert := effectstest.StapleLibraryCard(me, "Sandy Desert", "Land — Desert")
	effectstest.StapleLibraryCard(me, "Not A Land", "Sorcery")
	effectstest.AdvanceToMain(t, g)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Silver Deputy", TypeLine: "Artifact Creature — Mercenary",
		ManaCost: "{2}", OracleID: otjS2SilverDeputyOracle, Power: 1, Toughness: 2})
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Not A Land") != uuid.Nil {
		t.Error("a sorcery is neither a basic land card nor a Desert card")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, desert)
	effectstest.PassPriorityAroundTable(t, g)
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != desert {
		t.Errorf("the top card is %s, want the Desert", top.Name)
	}
}

func TestOtjS2SilverDeputyPumpsAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	deputy := effectstest.B12Push(g, me.ID, "Silver Deputy", "Artifact Creature — Mercenary", otjS2SilverDeputyOracle, 1, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp.ID))
	if err := g.ActivateCatalogAbility(me.ID, deputy, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(mine)}); err == nil {
		t.Fatal("activate only as a sorcery: refused with a spell on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, deputy, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(theirs)}); err == nil {
		t.Fatal("an opponent's creature is not a creature you control")
	}
	effectstest.B16Activate(t, g, me.ID, deputy, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(mine)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, mine); p != 3 || tt != 2 {
		t.Errorf("your Bear is %d/%d, want 3/2", p, tt)
	}
	if !effectstest.B16Tapped(t, g, deputy) {
		t.Error("{T} is the cost")
	}
}

// --- Skulduggery -------------------------------------------------------------

func TestOtjS2SkulduggeryGrowsYoursAndShrinksTheirs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Skulduggery", "Instant", otjS2SkulduggeryOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}, {Kind: game.TargetCard, ID: mine}}); err == nil {
		t.Fatal("the clauses are bound to a side: yours first, then theirs")
	}
	effectstest.CastCatalogSpell(t, g, "Skulduggery", "Instant", otjS2SkulduggeryOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, mine); p != 3 || tt != 3 {
		t.Errorf("your Bear is %d/%d, want 3/3", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, theirs); p != 1 || tt != 1 {
		t.Errorf("their Bear is %d/%d, want 1/1", p, tt)
	}
	if !effects.YouCommittedACrimeThisTurn(g, me.ID) {
		t.Error("targeting an opponent's creature is a crime")
	}
}

// --- Slickshot Lockpicker ----------------------------------------------------

func TestOtjS2SlickshotLockpickerGrantsFlashbackAtTheCardsManaCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	rock := effectstest.GyCard(me, "Old Rock", "Artifact", "{1}", "")
	effectstest.CastHoldCreature(t, g, "Slickshot Lockpicker", "Creature — Human Rogue", otjS2LockpickerOracle, 2, 3)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if effectstest.HasID(p.PickTargetCards, rock) || !effectstest.HasID(p.PickTargetCards, bolt) {
		t.Fatalf("target instant or sorcery card in your graveyard: %v", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	effectstest.B06AddMana(me, "R")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Targets: effectstest.B16TargetPlayer(opp.ID),
	}); err != nil {
		t.Fatalf("flashback from the graveyard for {R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-3)
	}
	if !g.Exile.Contains(bolt) {
		t.Error("a flashed-back card is exiled")
	}
}

// --- Spinewoods Armadillo ----------------------------------------------------

func TestOtjS2SpinewoodsArmadilloDiscardsToFindALandAndGainThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	desert := effectstest.StapleLibraryCard(me, "Sandy Desert", "Land — Desert")
	effectstest.StapleLibraryCard(me, "Not A Land", "Sorcery")
	effectstest.AdvanceToMain(t, g)
	armadillo := effectstest.PushCatalogHandCard(me, "Spinewoods Armadillo", "Creature — Armadillo", otjS2ArmadilloOracle)
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
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Not A Land") != uuid.Nil {
		t.Error("a sorcery is neither a basic land card nor a Desert card")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, desert)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, desert) {
		t.Error("the Desert goes to your hand")
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	spec, _ := effects.Lookup(otjS2ArmadilloOracle)
	if len(spec.Triggered) != 1 || len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "reach" {
		t.Errorf("want reach and one ward trigger: %v, %d triggers", spec.PrintedKeywords, len(spec.Triggered))
	}
}

// --- Stop Cold ---------------------------------------------------------------

func TestOtjS2StopColdTapsAndSilencesAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"})
	ring := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: effectstest.SolRingOracle, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Stop Cold", "Enchantment — Aura", otjS2StopColdOracle, effectstest.B16TargetCard(land)); err == nil {
		t.Fatal("a land is not an artifact or a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Stop Cold", "Enchantment — Aura", otjS2StopColdOracle, effectstest.B16TargetCard(ring))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.DftCG2Tapped(g, ring) {
		t.Error("the enchanted artifact is tapped")
	}
	if got := game.ManaAbilitiesForCard(effectstest.LayeredCard(t, g, ring)); len(got) != 0 {
		t.Errorf("mana abilities = %v, want none", got)
	}
}

func TestOtjS2StopColdKeepsACreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Stop Cold", "Enchantment — Aura", otjS2StopColdOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.DftCG2Tapped(g, bear) {
		t.Fatal("the enchanted creature is not tapped")
	}
	effectstest.DftMPPassTurn(t, g)
	if g.Seats[g.Turn.ActiveSeat].ID != opp.ID {
		t.Fatalf("the next turn is not the creature's controller's: seat %d", g.Turn.ActiveSeat)
	}
	effectstest.AdvanceToMain(t, g)
	if !effectstest.DftCG2Tapped(g, bear) {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
	spec, _ := effects.Lookup(otjS2StopColdOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("flash is printed: %v", spec.PrintedKeywords)
	}
}
