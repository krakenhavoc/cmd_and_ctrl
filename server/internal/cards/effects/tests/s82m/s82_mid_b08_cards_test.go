package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 MID batch 8: Cathar's Call, Stuffed Bear, Bramble Armor, Howl of the
// Hunt, Hedgewitch's Mask, Silver Bolt, Necrosynthesis, Grafted Identity,
// Borrowed Time, Geistflame Reservoir and Avacyn's Memorial.

const (
	s82m8CatharsCall   = "433e92d2-899a-4eed-98b9-80de5d25494b"
	s82m8StuffedBear   = "1402c116-194e-49ca-a8c1-09ca659c220c"
	s82m8BrambleArmor  = "25328cab-c7c0-4c55-99bb-a4149f6c511a"
	s82m8HowlOfHunt    = "22fe038e-033e-4632-a0d7-116492cded83"
	s82m8Mask          = "c74f6a83-a280-450e-969d-6b3fa28c8744"
	s82m8SilverBolt    = "b29a0686-2246-47a9-974d-4425e8a5546c"
	s82m8Necrosynth    = "661618a2-54c7-4136-beff-fff5ff7d84cb"
	s82m8Grafted       = "e4fb97b5-0d4a-4c9a-bd51-42ff723b5854"
	s82m8BorrowedTime  = "f1ac4e0e-5633-46e4-9a21-d3924e481d13"
	s82m8Reservoir     = "143714b8-30ae-42c2-84d9-e58aed7dff5b"
	s82m8Memorial      = "a8abc5db-4e0b-488d-898a-8109690d5099"
	s82m8AuraTypeLine  = "Enchantment — Aura"
	s82m8EquipTypeLine = "Artifact — Equipment"
)

func TestS82MidB08CardsAreFullWithNoCaveats(t *testing.T) {
	for _, oracle := range []string{s82m8CatharsCall, s82m8StuffedBear, s82m8BrambleArmor, s82m8HowlOfHunt, s82m8Mask,
		s82m8SilverBolt, s82m8Necrosynth, s82m8Grafted, s82m8BorrowedTime, s82m8Reservoir, s82m8Memorial} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v with caveats %v, want full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func s82m8Aura(t *testing.T, g *game.Game, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, name, s82m8AuraTypeLine, oracle, []game.TargetRef{{Kind: game.TargetCard, ID: host}})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.E2Card(t, g, id); c.AttachedTo.ID != host {
		t.Fatalf("%s is attached to %v, want %v", name, c.AttachedTo.ID, host)
	}
	return id
}

func TestS82MidB08CatharsCallGivesVigilanceAndTheControllersEndStepHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	s82m8Aura(t, g, "Cathar's Call", s82m8CatharsCall, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the enchanted creature has no vigilance")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Human", me.ID); n != 1 {
		t.Errorf("%d Human tokens for the creature's controller, want 1", n)
	}
	if n := effectstest.NamedOnBattlefield(g, "Human", opp.ID); n != 0 {
		t.Errorf("%d Human tokens for the opponent, want 0", n)
	}
}

func TestS82MidB08CatharsCallOnAnOpponentsCreatureMakesTheirTokenOnTheirTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	s82m8Aura(t, g, "Cathar's Call", s82m8CatharsCall, theirs)
	// My end step: the creature's controller is the opponent, so "your end
	// step" is theirs and nothing happens now.
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Human", me.ID) + effectstest.NamedOnBattlefield(g, "Human", opp.ID); n != 0 {
		t.Errorf("%d Human tokens at the Aura controller's end step, want 0", n)
	}
}

func TestS82MidB08StuffedBearBecomesAGreenFourFourBearUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Stuffed Bear", "Artifact", s82m8StuffedBear, false)
	if effectstest.EffectiveCardForTest(t, g, bear).IsCreature() {
		t.Fatal("it is a creature before it is animated")
	}
	effectstest.FillPool(me, 2)
	if err := effectstest.S82b11Activate(t, g, me, bear, 0); err != nil {
		t.Fatal(err)
	}
	c := effectstest.EffectiveCardForTest(t, g, bear)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Bear") {
		t.Errorf("animated: creature %v artifact %v Bear %v", c.IsCreature(), c.IsArtifact(), c.HasSubtype("Bear"))
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("it is %d/%d, want 4/4", p, tough)
	}
	if cols := c.Effective().Colors; len(cols) != 1 || cols[0] != "G" {
		t.Errorf("colours %v, want [G]", cols)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	for g.Turn.Step == game.StepEnd || g.Turn.Step == game.StepCleanup {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if effectstest.EffectiveCardForTest(t, g, bear).IsCreature() {
		t.Error("it is still a creature after the turn ended")
	}
}

func TestS82MidB08StuffedBearRefusesAnEmptyPool(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Stuffed Bear", "Artifact", s82m8StuffedBear, false)
	if err := effectstest.S82b11Activate(t, g, me, bear, 0); err == nil {
		t.Error("activated with no mana")
	}
}

func TestS82MidB08BrambleArmorAttachesOnEntryAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	armor := effectstest.CastCatalogSpell(t, g, "Bramble Armor", s82m8EquipTypeLine, s82m8BrambleArmor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.E2Card(t, g, armor); c.AttachedTo.ID != bear {
		t.Fatalf("attached to %v, want the Bear", c.AttachedTo.ID)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("the equipped Bear is %d/%d, want 4/3", p, tough)
	}
}

func TestS82MidB08BrambleArmorEquipCostsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	armor := effectstest.PushCatalogPermanent(g, me.ID, "Bramble Armor", s82m8EquipTypeLine, s82m8BrambleArmor, false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FillPool(me, 3)
	if err := effectstest.S82b11Activate(t, g, me, armor, 0, bear); err == nil {
		t.Fatal("equipped for three mana")
	}
	effectstest.FillPool(me, 1)
	if err := effectstest.S82b11Activate(t, g, me, armor, 0, bear); err != nil {
		t.Fatalf("equip for four: %v", err)
	}
	if c := effectstest.E2Card(t, g, armor); c.AttachedTo.ID != bear {
		t.Errorf("attached to %v, want the Bear", c.AttachedTo.ID)
	}
}

func TestS82MidB08HowlOfTheHuntPumpsAndUntapsAWolfOrWerewolfOnly(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		untaps   bool
	}{
		{"Creature — Wolf", true},
		{"Creature — Human Werewolf", true},
		{"Creature — Bear", false},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		host := effectstest.B12Creature(g, me.ID, "Host", tc.typeLine, 2, 2)
		g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, host).Tapped = true })
		s82m8Aura(t, g, "Howl of the Hunt", s82m8HowlOfHunt, host)
		if got := !effectstest.E2Card(t, g, host).Tapped; got != tc.untaps {
			t.Errorf("%s: untapped = %v, want %v", tc.typeLine, got, tc.untaps)
		}
		if p, tough := effectstest.PtOf(t, g, host); p != 4 || tough != 4 {
			t.Errorf("%s: %d/%d, want 4/4", tc.typeLine, p, tough)
		}
		if !effectstest.HasEffectiveKeyword(t, g, host, "vigilance") {
			t.Errorf("%s: no vigilance", tc.typeLine)
		}
	}
}

func TestS82MidB08HowlOfTheHuntHasFlash(t *testing.T) {
	spec, _ := effects.Lookup(s82m8HowlOfHunt)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("keywords %v, want flash", spec.PrintedKeywords)
	}
}

func TestS82MidB08HedgewitchsMaskPumpsAndBarsBigBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mask := effectstest.PushCatalogPermanent(g, me.ID, "Hedgewitch's Mask", s82m8EquipTypeLine, s82m8Mask, false)
	effectstest.AttachForTest(t, g, mask, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("equipped Bear is %d/%d, want 3/3", p, tough)
	}
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 4)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Giant", 3, 3)
	var vsBig, vsSmall bool
	g.ReadSnapshot(func() {
		a, _ := effectstest.BattlefieldCard(g, bear)
		b, _ := effectstest.BattlefieldCard(g, big)
		s, _ := effectstest.BattlefieldCard(g, small)
		vsBig = g.CanBlockLocked(&a, &b)
		vsSmall = g.CanBlockLocked(&a, &s)
	})
	if vsBig {
		t.Error("a power 4 creature blocked the equipped creature")
	}
	if !vsSmall {
		t.Error("a power 3 creature could not block the equipped creature")
	}
}

func TestS82MidB08SilverBoltDealsThreeAndDestroysAWerewolf(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		dies     bool
	}{
		{"Creature — Werewolf", true},
		{"Creature — Bear", false},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bolt := effectstest.PushCatalogPermanent(g, me.ID, "Silver Bolt", "Artifact", s82m8SilverBolt, false)
		victim := effectstest.B12Creature(g, opp.ID, "Victim", tc.typeLine, 2, 6)
		effectstest.FillPool(me, 3)
		if err := effectstest.S82b11Activate(t, g, me, bolt, 0, victim); err != nil {
			t.Fatal(err)
		}
		if g.Battlefield.Contains(bolt) {
			t.Errorf("%s: the Bolt was not sacrificed", tc.typeLine)
		}
		if got := g.Battlefield.Contains(victim); got == tc.dies {
			t.Errorf("%s: on battlefield = %v, want %v", tc.typeLine, got, !tc.dies)
		}
		if !tc.dies {
			if d := effectstest.E2Card(t, g, victim).DamageMarked; d != 3 {
				t.Errorf("%s: %d damage, want 3", tc.typeLine, d)
			}
		}
	}
}

func TestS82MidB08SilverBoltRefusesANonCreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.PushCatalogPermanent(g, me.ID, "Silver Bolt", "Artifact", s82m8SilverBolt, false)
	relic := effectstest.PushCatalogPermanent(g, opp.ID, "Relic", "Artifact", "", false)
	effectstest.FillPool(me, 3)
	if err := effectstest.S82b11Activate(t, g, me, bolt, 0, relic); err == nil {
		t.Error("targeted a noncreature artifact")
	}
}

func TestS82MidB08NecrosynthesisGrowsTheCreatureAndDrawsOnItsDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 3, 3)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	s82m8Aura(t, g, "Necrosynthesis", s82m8Necrosynth, host)

	effectstest.KillCreature(t, g, me.ID, other)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, host); got != 1 {
		t.Fatalf("%d counters after another creature died, want 1", got)
	}

	ids := effectstest.SamiLibraryTopFirst(me, game.Card{Name: "A"}, game.Card{Name: "B"}, game.Card{Name: "C"}, game.Card{Name: "D"}, game.Card{Name: "E"})
	effectstest.KillCreature(t, g, me.ID, host)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick prompt")
	}
	if len(pick.ChooseCards) != 4 {
		t.Errorf("%d cards offered, want 4 (its power 3 plus the counter)", len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[2]) {
		t.Error("the chosen card did not reach the hand")
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[4] {
		t.Error("the card below the looked-at ones must still be on top")
	}
	if bottom := me.Library.Cards[0].InstanceID; bottom != ids[0] && bottom != ids[1] && bottom != ids[3] {
		t.Error("the rest did not go to the bottom")
	}
}

func TestS82MidB08GraftedIdentityNeedsASacrificeAndStealsTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// Strict: no creature to sacrifice is refused.
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Grafted Identity", TypeLine: s82m8AuraTypeLine, OracleID: s82m8Grafted, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FillPool(me, 4)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err == nil {
		t.Fatal("cast without naming a creature to sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
		SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	c := effectstest.E2Card(t, g, theirs)
	if c.Controller != me.ID {
		t.Errorf("controller %v, want me", c.Controller)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 3 || tough != 3 {
		t.Errorf("it is %d/%d, want 3/3", p, tough)
	}
}

func TestS82MidB08BorrowedTimeExilesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	spec, _ := effects.Lookup(s82m8BorrowedTime)
	effectstest.RfPermAExpectLegal(t, g, me.ID, spec.Triggered[0].Targets, map[uuid.UUID]bool{mine: false, land: false, victim: true})
	time := effectstest.CastCatalogSpell(t, g, "Borrowed Time", "Enchantment", s82m8BorrowedTime, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature is not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(time) })
	g.RunStateChecksForTest()
	if n := effectstest.NamedOnBattlefield(g, "Their Threat", opp.ID); n != 1 {
		t.Errorf("after Borrowed Time left, %d copies under its owner's control, want 1", n)
	}
}

func TestS82MidB08GeistflameReservoirCountsInstantsAndSorceriesAndShootsThatMuch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	res := effectstest.PushCatalogPermanent(g, me.ID, "Geistflame Reservoir", "Artifact", s82m8Reservoir, false)
	effectstest.CastCatalogSpell(t, g, "Bolt", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, res, game.CounterCharge); got != 2 {
		t.Fatalf("%d charge counters, want 2 (the creature spell does not count)", got)
	}
	effectstest.FillPoolColored(me, "R", 2)
	before := opp.Life
	err := g.ActivateCatalogAbility(me.ID, res, 0, game.ActivateAbilityParams{
		Strict: true, CounterSourceIDs: []uuid.UUID{res}, CounterCounts: []int{2},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("life %d, want %d", opp.Life, before-2)
	}
	if got := effectstest.CountersOn(g, res, game.CounterCharge); got != 0 {
		t.Errorf("%d counters left, want 0", got)
	}
	if !g.Battlefield.Contains(res) {
		t.Error("the Reservoir left the battlefield")
	}
}

func TestS82MidB08GeistflameReservoirExilesTheTopCardToPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	res := effectstest.PushCatalogPermanent(g, me.ID, "Geistflame Reservoir", "Artifact", s82m8Reservoir, false)
	top := effectstest.SeedLibraryTop(me, "Top", "Sorcery", "{1}")
	effectstest.FillPoolColored(me, "R", 2)
	if err := g.ActivateCatalogAbility(me.ID, res, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Contains(top) {
		t.Error("the top card is still in the library")
	}
	if !g.Exile.Contains(top) {
		t.Error("the top card is not in exile")
	}
}

func TestS82MidB08AvacynsMemorialProtectsOtherLegendaryPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mem := effectstest.PushCatalogPermanent(g, me.ID, "Avacyn's Memorial", "Legendary Artifact", s82m8Memorial, false)
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Human", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Legend", "Legendary Creature — Human", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, legend, "indestructible") {
		t.Error("my other legendary permanent is not indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, plain, "indestructible") {
		t.Error("a nonlegendary permanent is indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("an opponent's legendary permanent is indestructible")
	}
	spec, _ := effects.Lookup(s82m8Memorial)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "indestructible" {
		t.Errorf("printed keywords %v, want indestructible", spec.PrintedKeywords)
	}
	_ = mem
}
