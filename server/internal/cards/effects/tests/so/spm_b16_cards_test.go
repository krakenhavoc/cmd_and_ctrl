package testsso

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm_b16_cards_test.go — Marvel's Spider-Man batch 16 (#4460). Each test
// plays the real catalog entry and asserts what its text says.

const (
	spmB16MassacreGirlOracle = "63e6cb7d-cc27-4200-85fc-ff6472318c1a"
	spmB16NajeelaOracle      = "09619943-6aec-4080-ace5-a0c6ebb23f1c"
	spmB16PositionOracle     = "abd62af0-c17d-4f62-af15-9ea83037b990"
	spmB16RanksOracle        = "0d4ee553-f206-4e70-a7b2-a3670440aee7"
	spmB16FuryOracle         = "f0ff2c1c-702d-406b-a426-4282117afe8e"
	spmB16ShowAndTellOracle  = "b83a3ba0-249e-4c39-bbf0-cb005413f7d2"
	spmB16SilkguardOracle    = "6dbeeb49-55c6-4ceb-94dd-3e4110dff8c0"
	spmB16SkithiryxOracle    = "daf6c421-e7f7-4fc6-967c-65f4ab96fcfd"
	spmB16SteelyOracle       = "48c127f0-2857-4c36-97bc-1291b6fe4a82"
	spmB16StormOracle        = "7c47876d-143f-4e4c-b151-90c099e452a1"
)

func TestSpmB16EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		spmB16MassacreGirlOracle: "Massacre Girl",
		spmB16NajeelaOracle:      "Najeela, the Blade-Blossom",
		spmB16PositionOracle:     "Privileged Position",
		spmB16RanksOracle:        "Return to the Ranks",
		spmB16FuryOracle:         "Righteous Fury",
		spmB16ShowAndTellOracle:  "Show and Tell",
		spmB16SilkguardOracle:    "Silkguard",
		spmB16SkithiryxOracle:    "Skithiryx, the Blight Dragon",
		spmB16SteelyOracle:       "Steely Resolve",
		spmB16StormOracle:        "Storm, Force of Nature",
	}
	for oracle, name := range want {
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

// spmB16Settle passes priority until the stack is empty.
func spmB16Settle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 40; i++ {
		g.RunStateChecksForTest()
		if effectstest.StackFullyEmpty(g) && effectstest.PayAmountPrompt(g, g.Seats[0].ID) == nil && effectstest.PayAmountPrompt(g, g.Seats[1].ID) == nil {
			return
		}
		for _, seat := range g.Seats {
			if c := effectstest.PayAmountPrompt(g, seat.ID); c != nil {
				if err := g.ResolvePayAmount(c.ID, seat.ID, c.PayAmount.Goal); err != nil {
					t.Fatalf("answering the attack split: %v", err)
				}
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	g.RunStateChecksForTest()
	if !effectstest.StackFullyEmpty(g) {
		t.Fatal("the stack never emptied")
	}
}

func spmB16OnBattlefield(g *game.Game, id uuid.UUID) bool { return g.Battlefield.Contains(id) }

func spmB16Tap(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	effectstest.BattlefieldCardPtr(t, g, id).Tapped = true
}

// --- Massacre Girl ---

func TestSpmB16MassacreGirlShrinksOthersAndChainsOnEveryDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.PushDiesCreatureForTest(g, opp.ID, "Big", "", "Creature — Test", 4, 4)
	tokenTmpl := effects.TokenCard("1/1 white Warrior")
	token := effectstest.PushToken(g, me.ID, tokenTmpl)
	mid := effectstest.PushDiesCreatureForTest(g, me.ID, "Mid", "", "Creature — Test", 2, 2)

	girl := effectstest.CastHoldCreature(t, g, "Massacre Girl", "Legendary Creature — Human Assassin", spmB16MassacreGirlOracle, 4, 4)
	spmB16Settle(t, g)

	// -1/-1 to the others: the token dies; its death (a token's counts)
	// is another sweep, which kills Mid (2->1->0), whose death is a third.
	if spmB16OnBattlefield(g, token) {
		t.Error("the 1/1 token survived the first -1/-1")
	}
	if spmB16OnBattlefield(g, mid) {
		t.Error("the 2/2 survived the chain")
	}
	if got := effectstest.EffectiveToughness(t, g, big); got != 1 {
		t.Errorf("the 4/4 has toughness %d, want 1 (first sweep plus a sweep for each of the two deaths)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, girl); got != 4 {
		t.Errorf("Massacre Girl has toughness %d, want 4: she is not shrunk", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, girl, "menace") {
		t.Error("Massacre Girl lacks menace")
	}
}

func TestSpmB16MassacreGirlSweepsAgainWhenAnyCreatureDiesLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.PushDiesCreatureForTest(g, opp.ID, "Big", "", "Creature — Test", 5, 5)
	girl := effectstest.CastHoldCreature(t, g, "Massacre Girl", "Legendary Creature — Human Assassin", spmB16MassacreGirlOracle, 4, 4)
	spmB16Settle(t, g)
	if got := effectstest.EffectiveToughness(t, g, big); got != 4 {
		t.Fatalf("Big has toughness %d after the enter sentence, want 4", got)
	}
	// A creature that dies later this turn, Massacre Girl's own side
	// included, is one more sweep over everything but her.
	victim := effectstest.PushDiesCreatureForTest(g, me.ID, "Victim", "", "Creature — Test", 1, 1)
	effectstest.BattlefieldCardPtr(t, g, victim).DamageMarked = 1
	spmB16Settle(t, g)
	if spmB16OnBattlefield(g, victim) {
		t.Fatal("the damaged creature did not die")
	}
	if got := effectstest.EffectiveToughness(t, g, big); got != 3 {
		t.Errorf("Big has toughness %d after one more death, want 3", got)
	}
	if got := effectstest.EffectiveToughness(t, g, girl); got != 4 {
		t.Errorf("Massacre Girl has toughness %d, want 4", got)
	}
}

// --- Najeela, the Blade-Blossom ---

func TestSpmB16NajeelaMakesTheAttackingWarriorsControllerATokenAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	najeela := effectstest.PushDiesCreatureForTest(g, me.ID, "Najeela, the Blade-Blossom", spmB16NajeelaOracle, "Legendary Creature — Human Warrior", 3, 2)
	bear := effectstest.PushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)

	effectstest.DeclareAttack(t, g, opp.ID, najeela, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	spmB16Settle(t, g)

	tokens := effectstest.BattlefieldIDsNamed(g, "Warrior")
	if len(tokens) != 1 {
		for _, c := range g.PendingChoices {
			t.Logf("pending: %+v", c)
		}
		t.Fatalf("%d Warrior tokens, want 1 (only Najeela is a Warrior)", len(tokens))
	}
	c, _ := effectstest.BattlefieldCard(g, tokens[0])
	if !c.Tapped || c.AttackingTarget != opp.ID || c.Controller != me.ID {
		t.Errorf("token tapped=%v attacking=%s controller=%s; want tapped, attacking the opponent, under Najeela's controller", c.Tapped, c.AttackingTarget, c.Controller)
	}
}

func TestSpmB16NajeelaMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	najeela := effectstest.PushDiesCreatureForTest(g, me.ID, "Najeela, the Blade-Blossom", spmB16NajeelaOracle, "Legendary Creature — Human Warrior", 3, 2)
	effectstest.DeclareAttack(t, g, opp.ID, najeela)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	spmB16Settle(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Warrior")); n != 0 {
		t.Errorf("%d Warrior tokens after declining, want 0", n)
	}
}

func TestSpmB16NajeelaGivesAnOpponentsAttackingWarriorTheToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Najeela, the Blade-Blossom", spmB16NajeelaOracle, "Legendary Creature — Human Warrior", 3, 2)
	raider := effectstest.PushDiesCreatureForTest(g, opp.ID, "Raider", "", "Creature — Human Warrior", 2, 2)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(raider, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	spmB16Settle(t, g)
	tokens := effectstest.BattlefieldIDsNamed(g, "Warrior")
	if len(tokens) != 1 {
		t.Fatalf("%d Warrior tokens, want 1", len(tokens))
	}
	c, _ := effectstest.BattlefieldCard(g, tokens[0])
	if c.Controller != opp.ID {
		t.Errorf("the token is controlled by %s, want the attacking Warrior's controller %s", c.Controller, opp.ID)
	}
	if !c.Tapped || c.AttackingTarget != me.ID {
		t.Errorf("token tapped=%v attacking=%s; want tapped and attacking %s", c.Tapped, c.AttackingTarget, me.ID)
	}
}

func TestSpmB16NajeelaUntapsAttackersGrantsKeywordsAndAddsACombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	najeela := effectstest.PushDiesCreatureForTest(g, me.ID, "Najeela, the Blade-Blossom", spmB16NajeelaOracle, "Legendary Creature — Human Warrior", 3, 2)
	bear := effectstest.PushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	idle := effectstest.PushDiesCreatureForTest(g, me.ID, "Idle", "", "Creature — Bear", 2, 2)
	spmB16Tap(t, g, idle)

	// Not during combat: refused, with the mana to pay for it.
	effectstest.FloatMana(t, g, me, "{W}{U}{B}{R}{G}")
	if err := g.ActivateCatalogAbility(me.ID, najeela, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("Najeela's ability was activated in the main phase")
	}

	effectstest.DeclareAttack(t, g, opp.ID, najeela, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	spmB16Settle(t, g)
	if err := g.ActivateCatalogAbility(me.ID, najeela, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate in combat: %v", err)
	}
	spmB16Settle(t, g)

	for _, id := range []uuid.UUID{najeela, bear} {
		c, _ := effectstest.BattlefieldCard(g, id)
		if c.Tapped {
			t.Errorf("%s is still tapped", c.Name)
		}
		for _, kw := range []string{"trample", "lifelink", "haste"} {
			if !game.HasKeyword(&c, kw) {
				t.Errorf("%s lacks %s", c.Name, kw)
			}
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, idle); !c.Tapped || game.HasKeyword(&c, "trample") {
		t.Errorf("the creature that was not attacking was untapped or granted trample: %+v", c)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepBeginCombat || next.PhaseOrdinal != 2 {
		t.Fatalf("after the first combat: %+v, want an additional combat", next)
	}
}

// --- Privileged Position ---

func TestSpmB16PrivilegedPositionGivesOtherPermanentsYouControlHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pos := effectstest.B12Push(g, me.ID, "Privileged Position", "Enchantment", spmB16PositionOracle, 0, 0)
	land := effectstest.B12Permanent(g, me.ID, "My Land", "Land")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for id, want := range map[uuid.UUID]bool{land: true, bear: true, pos: false, theirs: false} {
		if got := effectstest.HasEffectiveKeyword(t, g, id, "hexproof"); got != want {
			t.Errorf("hexproof on %s = %v, want %v", effectstest.B12Card(t, g, id).Name, got, want)
		}
	}
	// A permanent that arrives later has it at once.
	later := effectstest.B12Permanent(g, me.ID, "Later Rock", "Artifact")
	if !effectstest.HasEffectiveKeyword(t, g, later, "hexproof") {
		t.Error("a permanent that arrived later lacks hexproof")
	}
}

// --- Steely Resolve ---

func TestSpmB16SteelyResolveGivesEveryCreatureOfTheChosenTypeShroud(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushDiesCreatureForTest(g, me.ID, "My Bear", "", "Creature — Bear", 2, 2)
	theirs := effectstest.PushDiesCreatureForTest(g, opp.ID, "Their Bear", "", "Creature — Bear", 2, 2)
	elf := effectstest.PushDiesCreatureForTest(g, opp.ID, "Elf", "", "Creature — Elf", 1, 1)
	effectstest.CastAndPause(t, g, "Steely Resolve", "Enchantment", spmB16SteelyOracle)
	effectstest.AnswerCreatureType(t, g, me.ID, "Bear")
	spmB16Settle(t, g)
	for id, want := range map[uuid.UUID]bool{mine: true, theirs: true, elf: false} {
		if got := effectstest.HasEffectiveKeyword(t, g, id, "shroud"); got != want {
			t.Errorf("shroud on %s = %v, want %v", effectstest.B12Card(t, g, id).Name, got, want)
		}
	}
	later := effectstest.B12Creature(g, opp.ID, "Later Bear", "Creature — Bear", 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, later, "shroud") {
		t.Error("a Bear that arrived later lacks shroud")
	}
}

// --- Skithiryx, the Blight Dragon ---

func TestSpmB16SkithiryxHasFlyingAndInfectAndGainsHasteForB(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sk := effectstest.PushCatalogPermanent(g, me.ID, "Skithiryx, the Blight Dragon", "Legendary Creature — Phyrexian Dragon Skeleton", spmB16SkithiryxOracle, true)
	for _, kw := range []string{"flying", "infect"} {
		if !effectstest.HasEffectiveKeyword(t, g, sk, kw) {
			t.Errorf("Skithiryx lacks %s", kw)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, sk, "haste") {
		t.Fatal("Skithiryx has haste before the ability")
	}
	effectstest.FloatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, sk, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate haste: %v", err)
	}
	spmB16Settle(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, sk, "haste") {
		t.Error("Skithiryx did not gain haste")
	}
}

func TestSpmB16SkithiryxRegeneratesForBB(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sk := effectstest.PushCatalogPermanent(g, me.ID, "Skithiryx, the Blight Dragon", "Legendary Creature — Phyrexian Dragon Skeleton", spmB16SkithiryxOracle, false)
	effectstest.BattlefieldCardPtr(t, g, sk).Power, effectstest.BattlefieldCardPtr(t, g, sk).Toughness = 4, 4
	effectstest.FloatMana(t, g, me, "{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, sk, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate regenerate: %v", err)
	}
	spmB16Settle(t, g)
	effectstest.BattlefieldCardPtr(t, g, sk).DamageMarked = 4
	g.RunStateChecksForTest()
	c, ok := effectstest.BattlefieldCard(g, sk)
	if !ok {
		t.Fatal("Skithiryx died with a regeneration shield up")
	}
	if !c.Tapped || c.DamageMarked != 0 {
		t.Errorf("after regenerating: tapped=%v damage=%d, want tapped with no damage", c.Tapped, c.DamageMarked)
	}
}

// --- Righteous Fury ---

func TestSpmB16RighteousFuryDestroysTappedCreaturesAndGainsTwoEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myTapped := effectstest.PushDiesCreatureForTest(g, me.ID, "My Tapped", "", "Creature — Test", 2, 2)
	myUntapped := effectstest.PushDiesCreatureForTest(g, me.ID, "My Untapped", "", "Creature — Test", 2, 2)
	theirTapped := effectstest.PushDiesCreatureForTest(g, opp.ID, "Their Tapped", "", "Creature — Test", 2, 2)
	theirUntapped := effectstest.PushDiesCreatureForTest(g, opp.ID, "Their Untapped", "", "Creature — Test", 2, 2)
	tappedLand := effectstest.B12Permanent(g, opp.ID, "Tapped Land", "Land")
	spmB16Tap(t, g, myTapped)
	spmB16Tap(t, g, theirTapped)
	spmB16Tap(t, g, tappedLand)
	before := me.Life

	effectstest.CastCatalogSpell(t, g, "Righteous Fury", "Sorcery", spmB16FuryOracle, nil)
	spmB16Settle(t, g)

	if spmB16OnBattlefield(g, myTapped) || spmB16OnBattlefield(g, theirTapped) {
		t.Error("a tapped creature survived")
	}
	if !spmB16OnBattlefield(g, myUntapped) || !spmB16OnBattlefield(g, theirUntapped) || !spmB16OnBattlefield(g, tappedLand) {
		t.Error("an untapped creature, or a tapped land, was destroyed")
	}
	if me.Life != before+4 {
		t.Errorf("life %d -> %d, want +4 (2 for each of two creatures)", before, me.Life)
	}
}

func TestSpmB16RighteousFuryCountsOnlyWhatWasDestroyed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doomed := effectstest.PushDiesCreatureForTest(g, opp.ID, "Doomed", "", "Creature — Test", 2, 2)
	sturdy := effectstest.PushDiesCreatureForTest(g, opp.ID, "Sturdy", "", "Creature — Test", 2, 2)
	effectstest.BattlefieldCardPtr(t, g, sturdy).Keywords = []string{"indestructible"}
	spmB16Tap(t, g, doomed)
	spmB16Tap(t, g, sturdy)
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Righteous Fury", "Sorcery", spmB16FuryOracle, nil)
	spmB16Settle(t, g)
	if spmB16OnBattlefield(g, doomed) || !spmB16OnBattlefield(g, sturdy) {
		t.Fatal("the indestructible creature died, or the other survived")
	}
	if me.Life != before+2 {
		t.Errorf("life %d -> %d, want +2: the indestructible creature was not destroyed this way", before, me.Life)
	}
}

// --- Silkguard ---

func TestSpmB16SilkguardCountersTheTargetsAndProtectsTheModified(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	countered := effectstest.B12Creature(g, me.ID, "Already Countered", "Creature — Bear", 2, 2)
	effectstest.BattlefieldCardPtr(t, g, countered).Counters = map[string]int{game.CounterPlusOne: 1}
	sword := effectstest.B12Permanent(g, me.ID, "My Sword", "Artifact — Equipment")
	aura := effectstest.B12Permanent(g, me.ID, "My Aura", "Enchantment — Aura")
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirSword := effectstest.B12Permanent(g, opp.ID, "Their Sword", "Artifact — Equipment")
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	effectstest.CastXSpell(t, g, "Silkguard", "Instant", spmB16SilkguardOracle, "{X}{G}", 2, effectstest.CardRefs(a, b))
	spmB16Settle(t, g)

	if effectstest.PlusOnes(g, a) != 1 || effectstest.PlusOnes(g, b) != 1 || effectstest.PlusOnes(g, plain) != 0 || effectstest.PlusOnes(g, theirs) != 0 {
		t.Errorf("counters a=%d b=%d plain=%d theirs=%d; want 1, 1, 0, 0", effectstest.PlusOnes(g, a), effectstest.PlusOnes(g, b), effectstest.PlusOnes(g, plain), effectstest.PlusOnes(g, theirs))
	}
	for id, want := range map[uuid.UUID]bool{
		a: true, b: true, countered: true, sword: true, aura: true,
		plain: false, rock: false, theirSword: false, theirs: false,
	} {
		if got := effectstest.HasEffectiveKeyword(t, g, id, "hexproof"); got != want {
			t.Errorf("hexproof on %s = %v, want %v", effectstest.B12Card(t, g, id).Name, got, want)
		}
	}
}

func TestSpmB16SilkguardTargetsAtMostXCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)

	cast := func(x int, mana string, targets []game.TargetRef) error {
		effectstest.FloatMana(t, g, me, mana)
		_, err := effectstest.CastFromHandErr(g, me, "Silkguard", "Instant", "{X}{G}", spmB16SilkguardOracle, game.CastSpellParams{Strict: true, XValue: x, Targets: targets})
		return err
	}
	if err := cast(1, "{G}{C}", effectstest.CardRefs(a, b)); err == nil {
		t.Fatal("two targets with X = 1 were accepted")
	}
	g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
	if err := cast(1, "{G}{C}", effectstest.CardRefs(theirs)); err == nil {
		t.Fatal("a creature the opponent controls was accepted")
	}
	g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
	if err := cast(1, "{G}{C}", effectstest.CardRefs(a)); err != nil {
		t.Fatalf("one target with X = 1 was refused: %v", err)
	}
	spmB16Settle(t, g)
	if err := cast(0, "{G}", nil); err != nil {
		t.Fatalf("no target with X = 0 was refused: %v", err)
	}
	spmB16Settle(t, g)
	if effectstest.PlusOnes(g, a) != 1 || effectstest.PlusOnes(g, b) != 0 {
		t.Errorf("counters a=%d b=%d; want 1, 0", effectstest.PlusOnes(g, a), effectstest.PlusOnes(g, b))
	}
}

// --- Return to the Ranks ---

func TestSpmB16ReturnToTheRanksReturnsXSmallCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	one := effectstest.PushGraveyardPermanent(me, "One Drop", "Creature — Soldier", "{W}")
	two := effectstest.PushGraveyardPermanent(me, "Two Drop", "Creature — Soldier", "{1}{W}")
	three := effectstest.PushGraveyardPermanent(me, "Three Drop", "Creature — Soldier", "{2}{W}")
	rock := effectstest.PushGraveyardPermanent(me, "Rock", "Artifact", "{1}")

	effectstest.CastXSpell(t, g, "Return to the Ranks", "Sorcery", spmB16RanksOracle, "{X}{W}{W}", 2, effectstest.CardRefs(one, two))
	spmB16Settle(t, g)

	if !spmB16OnBattlefield(g, one) || !spmB16OnBattlefield(g, two) {
		t.Error("a targeted creature card did not return")
	}
	if me.Graveyard.Contains(three) == false || me.Graveyard.Contains(rock) == false {
		t.Error("an untargeted card left the graveyard")
	}
}

func TestSpmB16ReturnToTheRanksTargetsExactlyXLegalCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	one := effectstest.PushGraveyardPermanent(me, "One Drop", "Creature — Soldier", "{W}")
	two := effectstest.PushGraveyardPermanent(me, "Two Drop", "Creature — Soldier", "{1}{W}")
	three := effectstest.PushGraveyardPermanent(me, "Three Drop", "Creature — Soldier", "{2}{W}")
	rock := effectstest.PushGraveyardPermanent(me, "Rock", "Artifact", "{1}")
	effectstest.AdvanceToMain(t, g)

	cast := func(x int, targets ...uuid.UUID) error {
		effectstest.FloatMana(t, g, me, "{W}{W}"+string(slices.Repeat([]byte("{C}"), x)))
		_, err := effectstest.CastFromHandErr(g, me, "Return to the Ranks", "Sorcery", "{X}{W}{W}", spmB16RanksOracle, game.CastSpellParams{Strict: true, XValue: x, Targets: effectstest.CardRefs(targets...)})
		g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
		return err
	}
	if err := cast(1, three); err == nil {
		t.Error("a creature card with mana value 3 was a legal target")
	}
	if err := cast(1, rock); err == nil {
		t.Error("a noncreature card was a legal target")
	}
	if err := cast(2, one); err == nil {
		t.Error("one target was accepted for X = 2")
	}
	if err := cast(1, one, two); err == nil {
		t.Error("two targets were accepted for X = 1")
	}
	if err := cast(1, two); err != nil {
		t.Fatalf("X = 1 with a mana value 2 creature card was refused: %v", err)
	}
}

func TestSpmB16ReturnToTheRanksConvokes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	one := effectstest.PushGraveyardPermanent(me, "One Drop", "Creature — Soldier", "{W}")
	helper := effectstest.B12Creature(g, me.ID, "Helper", "Creature — Soldier", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{W}{W}")
	// {1}{W}{W}: the two white from the pool, the generic by tapping.
	if _, err := effectstest.CastFromHandErr(g, me, "Return to the Ranks", "Sorcery", "{X}{W}{W}", spmB16RanksOracle,
		game.CastSpellParams{Strict: true, XValue: 1, Targets: effectstest.CardRefs(one), TapIDs: []uuid.UUID{helper}}); err != nil {
		t.Fatalf("convoke cast: %v", err)
	}
	if c, _ := effectstest.BattlefieldCard(g, helper); !c.Tapped {
		t.Error("the convoking creature was not tapped")
	}
	spmB16Settle(t, g)
	if !spmB16OnBattlefield(g, one) {
		t.Error("the creature card did not return")
	}
}

// --- Show and Tell ---

func TestSpmB16ShowAndTellEveryPlayerMayPutOneCardFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for _, p := range g.Seats {
		g.WithWriteLock(func() { p.Hand.Cards = nil })
	}
	myCreature := effectstest.HandCardForTest(me, "My Creature", "Creature — Bear", "")
	myOther := effectstest.HandCardForTest(me, "My Other Creature", "Creature — Bear", "")
	myInstant := effectstest.HandCardForTest(me, "My Instant", "Instant", "")
	myWalker := effectstest.HandCardForTest(me, "My Walker", "Legendary Planeswalker — Test", "")
	theirLand := effectstest.HandCardForTest(opp, "Their Land", "Land", "")
	theirSorcery := effectstest.HandCardForTest(opp, "Their Sorcery", "Sorcery", "")

	effectstest.CastCatalogSpell(t, g, "Show and Tell", "Sorcery", spmB16ShowAndTellOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	ask := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if ask == nil {
		t.Fatal("the active player was not asked")
	}
	if len(ask.ChooseCards) != 2 || slices.Contains(ask.ChooseCards, myInstant) || slices.Contains(ask.ChooseCards, myWalker) {
		t.Fatalf("candidates %v, want the two creature cards only", ask.ChooseCards)
	}
	if err := g.ResolveChooseCards(ask.ID, me.ID, []uuid.UUID{myCreature, myOther}); err == nil {
		t.Fatal("two cards were accepted: each player puts one")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, myCreature)
	ask = effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if ask == nil {
		t.Fatal("the next player was not asked")
	}
	if len(ask.ChooseCards) != 1 || ask.ChooseCards[0] != theirLand {
		t.Fatalf("their candidates %v, want only the land", ask.ChooseCards)
	}
	if spmB16OnBattlefield(g, myCreature) {
		t.Error("a pick entered before every player had chosen")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, theirLand)
	spmB16Settle(t, g)

	if !spmB16OnBattlefield(g, myCreature) || !spmB16OnBattlefield(g, theirLand) {
		t.Error("a chosen card did not enter")
	}
	if spmB16OnBattlefield(g, myOther) || !me.Hand.Contains(myOther) {
		t.Error("the unchosen creature card left the hand")
	}
	if !me.Hand.Contains(myInstant) || !me.Hand.Contains(myWalker) || !opp.Hand.Contains(theirSorcery) {
		t.Error("an ineligible card left a hand")
	}
	if c, _ := effectstest.BattlefieldCard(g, theirLand); c.Controller != opp.ID {
		t.Errorf("their land entered under %s, want its owner", c.Controller)
	}
}

func TestSpmB16ShowAndTellDecliningPutsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats {
		g.WithWriteLock(func() { p.Hand.Cards = nil })
	}
	card := effectstest.HandCardForTest(me, "My Creature", "Creature — Bear", "")
	effectstest.CastCatalogSpell(t, g, "Show and Tell", "Sorcery", spmB16ShowAndTellOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	spmB16Settle(t, g)
	if spmB16OnBattlefield(g, card) || !me.Hand.Contains(card) {
		t.Error("declining still put the card onto the battlefield")
	}
}

// --- Storm, Force of Nature ---

func TestSpmB16StormGivesTheNextInstantOrSorceryStormOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Storm, Force of Nature", spmB16StormOracle, "Legendary Creature — Mutant Hero", 3, 4)

	effectstest.StormFiller(t, g) // one spell cast before combat damage
	effectstest.AttackWith(t, g, opp.ID, effectstest.BattlefieldIDsNamed(g, "Storm, Force of Nature")[0])
	spmB16Settle(t, g)
	afterCombat := effectstest.LifeOf(g, opp.ID)

	// A creature spell neither storms nor uses the watch up.
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)

	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if effectstest.TriggerOnStack(g, bolt) == nil && len(g.StackMeta) < 2 {
		t.Fatal("no storm trigger went on the stack with the spell")
	}
	// Two spells came before it this turn (the filler and the Bear): two copies.
	effectstest.AnswerCopyPrompts(t, g, me.ID, opp.ID, 2)
	if got := effectstest.LifeOf(g, opp.ID); got != afterCombat-9 {
		t.Errorf("life %d -> %d, want -9 (the spell and two copies, 3 each)", afterCombat, got)
	}

	// The watch was used up by that spell: the next one does not storm.
	again := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	_ = again
	before := effectstest.LifeOf(g, opp.ID)
	spmB16Settle(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-3 {
		t.Errorf("the second spell dealt %d, want 3: it does not have storm", before-got)
	}
}

func TestSpmB16StormDoesNothingWithoutCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Storm, Force of Nature", spmB16StormOracle, "Legendary Creature — Mutant Hero", 3, 4)
	effectstest.StormFiller(t, g)
	before := effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	spmB16Settle(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-3 {
		t.Errorf("Lightning Bolt dealt %d with no storm granted, want 3", before-got)
	}
}
