package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	rprB04BanisherPriest   = "9f560b83-32d4-4bb4-a956-8f5db18599db"
	rprB04Contender        = "35df179a-c0e6-4ac1-a861-e6e9b4d1614d"
	rprB04AngelSalvation   = "27b945ad-92a9-4615-87ec-790682e8c3e7"
	rprB04Gearhulk         = "2e4865c5-a590-4820-81ec-926014ddb8c4"
	rprB04AidFromTheCowl   = "9597c974-6120-43a7-9ad3-7e4fddc8c9f5"
	rprB04Alharu           = "c8deef08-fcf2-41e1-8aeb-aa4e0af147ca"
	rprB04AbzanFalconer    = "8b972819-507a-40a9-ab1f-1a674ea56083"
	rprB04Stronghold       = "c792229b-4a0f-48d5-93e5-60bd4cae9c42"
	rprB04Arvad            = "d561dcff-0a19-46ac-9dee-f374c840d532"
	rprB04AbzanBattlePrist = "1b44fe0a-4a99-4166-b0b3-102b36b54ffa"
	rprB04Pretender        = "51d6d164-6e5f-4be3-8f09-433e0b8d4c6b"
)

func TestRprB04CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		rprB04BanisherPriest: "Banisher Priest", rprB04Contender: "Acclaimed Contender",
		rprB04AngelSalvation: "Angel of Salvation", rprB04Gearhulk: "Cataclysmic Gearhulk",
		rprB04AidFromTheCowl: "Aid from the Cowl", rprB04Alharu: "Alharu, Solemn Ritualist",
		rprB04AbzanFalconer: "Abzan Falconer", rprB04Stronghold: "Bretagard Stronghold",
		rprB04Arvad: "Arvad the Cursed", rprB04AbzanBattlePrist: "Abzan Battle Priest",
		rprB04Pretender: "Bloodline Pretender",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Banisher Priest -----------------------------------------------------

func TestRprB04BanisherPriestExilesAnOpposingCreatureUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	priest := effectstest.CastCatalogSpell(t, g, "Banisher Priest", "Creature — Human Cleric", rprB04BanisherPriest, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardRefs(mine)); err == nil {
		t.Fatal("Banisher Priest targeted its controller's own creature")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) || g.Battlefield.Contains(theirs) {
		t.Fatal("the opposing creature was not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(priest) })
	g.RunStateChecksForTest()
	if n := effectstest.NamedOnBattlefield(g, "Their Bear", opp.ID); n != 1 {
		t.Errorf("after the Priest left, %d of the exiled creature under its owner's control, want 1", n)
	}
}

// --- Abzan Battle Priest and Abzan Falconer --------------------------------

func rprB04ActivateOutlast(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) error {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{})
}

func TestRprB04OutlastPutsACounterAndGrantsTheKeywordToCreaturesWithCounters(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle, keyword string
		power, toughness                int
	}{
		{"Abzan Battle Priest", "Creature — Human Cleric", rprB04AbzanBattlePrist, "lifelink", 3, 2},
		{"Abzan Falconer", "Creature — Human Soldier", rprB04AbzanFalconer, "flying", 2, 3},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			self := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: c.name, TypeLine: c.typeLine, OracleID: c.oracle,
				Power: c.power, Toughness: c.toughness, Owner: me.ID, Controller: me.ID})
			bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			opp := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
			effectstest.AssertKeywords(t, g, self)
			if err := rprB04ActivateOutlast(t, g, me, self); err != nil {
				t.Fatalf("outlast: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.BattlefieldPlusOnes(g, self); n != 1 {
				t.Fatalf("%d counters after outlast, want 1", n)
			}
			if !effectstest.FindBattlefieldCardForTest(g, self).Tapped {
				t.Error("outlast did not tap the creature")
			}
			effectstest.AssertKeywords(t, g, self, c.keyword)
			effectstest.AssertKeywords(t, g, bear)
			effectstest.AddCountersForTest(g, bear, game.CounterPlusOne, 1)
			effectstest.AssertKeywords(t, g, bear, c.keyword)
			effectstest.AddCountersForTest(g, opp, game.CounterPlusOne, 1)
			effectstest.AssertKeywords(t, g, opp)
		})
	}
}

func TestRprB04OutlastIsSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	self := effectstest.PushCatalogPermanent(g, me.ID, "Abzan Falconer", "Creature — Human Soldier", rprB04AbzanFalconer, false)
	effectstest.B12Creature(g, g.Seats[1].ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, self, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("outlast was activated outside a main phase")
	}
	if effectstest.BattlefieldPlusOnes(g, self) != 0 {
		t.Fatal("a refused outlast put a counter")
	}
}

// --- Arvad the Cursed ------------------------------------------------------

func TestRprB04ArvadPumpsOtherLegendaryCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arvad := effectstest.B12Creature(g, me.ID, "Arvad the Cursed", "Legendary Creature — Vampire Knight", 3, 3)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, arvad).OracleID = rprB04Arvad })
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Human", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Legend", "Legendary Creature — Human", 2, 2)
	for _, c := range []struct {
		id         uuid.UUID
		name       string
		power, tou int
	}{{legend, "your other legend", 4, 4}, {arvad, "Arvad itself", 3, 3}, {plain, "a nonlegendary creature", 2, 2}, {theirs, "an opposing legend", 2, 2}} {
		if p, tt := effectstest.EffectivePTOf(t, g, c.id); p != c.power || tt != c.tou {
			t.Errorf("%s is %d/%d, want %d/%d", c.name, p, tt, c.power, c.tou)
		}
	}
	effectstest.AssertKeywords(t, g, arvad, "deathtouch", "lifelink")
}

// --- Bloodline Pretender ---------------------------------------------------

func TestRprB04BloodlinePretenderGrowsForAnotherCreatureOfTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pretender := effectstest.PushNamedTribePermanent(t, g, me.ID, "Bloodline Pretender", "Artifact Creature — Shapeshifter", rprB04Pretender, "Elf")
	cast := func(name, typeLine string, keywords ...string) {
		t.Helper()
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Keywords: keywords, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	cast("Plain Bear", "Creature — Bear")
	if n := effectstest.BattlefieldPlusOnes(g, pretender); n != 0 {
		t.Fatalf("a Bear grew the Pretender (%d counters)", n)
	}
	cast("Elvish Bear", "Creature — Elf")
	if n := effectstest.BattlefieldPlusOnes(g, pretender); n != 1 {
		t.Fatalf("an Elf gave %d counters, want 1", n)
	}
	cast("Shapeshifter", "Creature — Shapeshifter", "changeling")
	if n := effectstest.BattlefieldPlusOnes(g, pretender); n != 2 {
		t.Fatalf("a changeling gave %d counters in all, want 2", n)
	}
	_ = opp
}

// --- Alharu, Solemn Ritualist ----------------------------------------------

func TestRprB04AlharuCountersTwoOthersAndMakesSpiritsWhenCounteredCreaturesDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	alharu := effectstest.CastCatalogSpell(t, g, "Alharu, Solemn Ritualist", "Legendary Creature — Human Monk", rprB04Alharu, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asked for no targets")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(a, b, c)); err == nil {
		t.Fatal("three targets were accepted")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(alharu)); err == nil {
		t.Fatal("Alharu targeted itself")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(a, b)); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldPlusOnes(g, a) != 1 || effectstest.BattlefieldPlusOnes(g, b) != 1 || effectstest.BattlefieldPlusOnes(g, c) != 0 {
		t.Fatalf("counters %d %d %d, want 1 1 0", effectstest.BattlefieldPlusOnes(g, a), effectstest.BattlefieldPlusOnes(g, b), effectstest.BattlefieldPlusOnes(g, c))
	}
	spirits := func() int { return effectstest.NamedOnBattlefield(g, "Spirit", me.ID) }
	effectstest.B27Kill(g, c)
	effectstest.PassPriorityAroundTable(t, g)
	if spirits() != 0 {
		t.Fatal("a creature with no counter made a Spirit")
	}
	effectstest.B27Kill(g, a)
	effectstest.PassPriorityAroundTable(t, g)
	if spirits() != 1 {
		t.Fatalf("%d Spirits after a countered creature died, want 1", spirits())
	}
	effectstest.AssertKeywords(t, g, findSpirit(g, me.ID), "flying")
	effectstest.B27Kill(g, alharu)
	effectstest.PassPriorityAroundTable(t, g)
	if spirits() != 1 {
		t.Fatalf("%d Spirits after Alharu (no counter) died, want still 1", spirits())
	}
}

func findSpirit(g *game.Game, owner uuid.UUID) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && c.Controller == owner {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

// --- Aid from the Cowl -----------------------------------------------------

func TestRprB04AidFromTheCowlNeedsRevoltAndPutsAPermanentOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Aid from the Cowl", "Enchantment", rprB04AidFromTheCowl, false)
	top := effectstest.PutOnTopOfLibrary(g, me, "Top Bear", "Creature — Bear", "")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil || g.Battlefield.Contains(top) {
		t.Fatal("it triggered with no revolt")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.PushCatalogPermanent(g2, me2.ID, "Aid from the Cowl", "Enchantment", rprB04AidFromTheCowl, false)
	top2 := effectstest.PutOnTopOfLibrary(g2, me2, "Top Bear", "Creature — Bear", "")
	effectstest.RevoltFor(t, g2, me2.ID)
	effectstest.AdvanceTo(t, g2, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerMayChoice(t, g2, me2.ID, true)
	effectstest.PassPriorityAroundTable(t, g2)
	if !g2.Battlefield.Contains(top2) {
		t.Fatal("the revealed permanent card was not put onto the battlefield")
	}
}

func TestRprB04AidFromTheCowlMayDeclineAPermanentAndBottomsALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Aid from the Cowl", "Enchantment", rprB04AidFromTheCowl, false)
	top := effectstest.PutOnTopOfLibrary(g, me, "Top Bear", "Creature — Bear", "")
	effectstest.RevoltFor(t, g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(top) || me.Library.Cards[len(me.Library.Cards)-1].InstanceID != top {
		t.Fatal("declining moved the card")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.PushCatalogPermanent(g2, me2.ID, "Aid from the Cowl", "Enchantment", rprB04AidFromTheCowl, false)
	spell := effectstest.PutOnTopOfLibrary(g2, me2, "Top Bolt", "Instant", "")
	effectstest.RevoltFor(t, g2, me2.ID)
	effectstest.AdvanceTo(t, g2, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerMayChoice(t, g2, me2.ID, true)
	effectstest.PassPriorityAroundTable(t, g2)
	if me2.Library.Cards[0].InstanceID != spell {
		t.Fatal("a nonpermanent card was not put on the bottom of the library")
	}
}

// --- Acclaimed Contender ---------------------------------------------------

func TestRprB04AcclaimedContenderLooksOnlyWithAnotherKnight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.HornTopFive(me, []struct{ Name, TypeLine string }{
		{"Plains", "Basic Land — Plains"}, {"Sir", "Creature — Human Knight"}, {"Sword", "Artifact — Equipment"},
		{"Ring", "Legendary Artifact"}, {"Pacifism", "Enchantment — Aura"},
	})
	effectstest.CastCatalogSpell(t, g, "Acclaimed Contender", "Creature — Human Knight", rprB04Contender, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Fatal("it looked with no other Knight")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.B12Creature(g2, me2.ID, "Other Knight", "Creature — Human Knight", 2, 2)
	top := effectstest.HornTopFive(me2, []struct{ Name, TypeLine string }{
		{"Plains", "Basic Land — Plains"}, {"Sir", "Creature — Human Knight"}, {"Sword", "Artifact — Equipment"},
		{"Ring", "Legendary Artifact"}, {"Pacifism", "Enchantment — Aura"},
	})
	effectstest.CastCatalogSpell(t, g2, "Acclaimed Contender", "Creature — Human Knight", rprB04Contender, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	prompt := effectstest.LatestChooseCards(g2, me2.ID)
	if prompt == nil {
		t.Fatal("no look with another Knight")
	}
	if len(prompt.ChooseCards) != 4 || prompt.ChooseMax != 1 {
		t.Fatalf("offered %d cards (max %d), want the four non-land cards, one at most", len(prompt.ChooseCards), prompt.ChooseMax)
	}
	for _, id := range prompt.ChooseCards {
		if id == top[0] {
			t.Error("a land was offered")
		}
	}
	handBefore := me2.Hand.Size()
	if err := g2.ResolveChooseCards(prompt.ID, me2.ID, []uuid.UUID{top[3]}); err != nil {
		t.Fatalf("take: %v", err)
	}
	if !me2.Hand.Contains(top[3]) || me2.Hand.Size() != handBefore+1 {
		t.Error("the legendary artifact did not reach the hand")
	}
}

// --- Angel of Salvation ----------------------------------------------------

func TestRprB04AngelOfSalvationPreventsFiveDividedAndHasFlashAndConvoke(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	spec, _ := effects.Lookup(rprB04AngelSalvation)
	if spec.TapCost == nil || spec.TapCost.Key != "convoke" {
		t.Error("Angel of Salvation has no convoke")
	}
	angel := effectstest.CastCatalogSpell(t, g, "Angel of Salvation", "Creature — Angel", rprB04AngelSalvation, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, angel, "flash", "flying")
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asked for no targets")
	}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, append(effectstest.CardRefs(bear), game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}),
		map[uuid.UUID]int{bear: 3, opp.ID: 3}); err == nil {
		t.Fatal("a division of six was accepted")
	}
	refs := append(effectstest.CardRefs(bear), game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{bear: 3, me.ID: 2}); err != nil {
		t.Fatalf("a legal division: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 4) })
	if d := effectstest.E2Card(t, g, bear).DamageMarked; d != 1 {
		t.Errorf("4 damage to the Bear with a 3 shield left %d marked, want 1", d)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.New(), me.ID, 5) })
	if me.Life != life-3 {
		t.Errorf("5 damage to me with a 2 shield: life %d → %d, want -3", life, me.Life)
	}
}

// --- Bretagard Stronghold --------------------------------------------------

func TestRprB04BretagardStrongholdCountersTwoCreaturesAndSacrifices(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Bretagard Stronghold", "Land", rprB04Stronghold, false)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	for _, bad := range [][]uuid.UUID{{a, b, c}, {theirs}} {
		if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bad...)}); err == nil {
			t.Fatalf("targets %v were accepted", bad)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(a, b)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldPlusOnes(g, a) != 1 || effectstest.BattlefieldPlusOnes(g, b) != 1 || effectstest.BattlefieldPlusOnes(g, c) != 0 {
		t.Fatal("counters are not on exactly the two chosen creatures")
	}
	effectstest.AssertKeywords(t, g, a, "vigilance", "lifelink")
	effectstest.AssertKeywords(t, g, b, "vigilance", "lifelink")
	effectstest.AssertKeywords(t, g, c)
}

func TestRprB04BretagardStrongholdEntersTappedAndTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.B12PlayFromHand(t, g, "Bretagard Stronghold", "Land", rprB04Stronghold, game.CastSpellParams{})
	if !effectstest.B12Card(t, g, land).Tapped {
		t.Fatal("did not enter tapped")
	}
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.UntapForTest(g, land)
	effectstest.ActivateManaFor(t, g, me.ID, land, 0, game.ManaAbilityParams{})
	if got := effectstest.PoolColours(me); got["G"] != 1 || me.ManaPool.Total() != 1 {
		t.Fatalf("pool %v, want one {G}", got)
	}
}

// --- Cataclysmic Gearhulk --------------------------------------------------

func TestRprB04CataclysmicGearhulkKeepsOneOfEachTypeAndNeverTouchesLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a1 := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	a2 := effectstest.B12Permanent(g, me.ID, "Stone", "Artifact")
	c1 := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	c2 := effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	e1 := effectstest.B12Permanent(g, me.ID, "Ward", "Enchantment")
	pw := effectstest.PushWalkerForTest(g, me.ID, "Walker", "", 3)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	oppLand := effectstest.PushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	oc1 := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	oc2 := effectstest.B12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 2, 2)
	hulk := effectstest.CastCatalogSpell(t, g, "Cataclysmic Gearhulk", "Artifact Creature — Construct", rprB04Gearhulk, nil)
	effectstest.PassPriorityAroundTable(t, g)

	prompt := effectstest.LatestChooseCardsFor(g, me.ID)
	if prompt == nil {
		t.Fatal("the controller was not asked")
	}
	if len(prompt.ChooseCards) != 7 {
		t.Fatalf("offered %d permanents, want the 7 nonland ones", len(prompt.ChooseCards))
	}
	for _, id := range prompt.ChooseCards {
		if id == land {
			t.Fatal("a land was offered")
		}
	}
	// Missing the enchantment and the planeswalker is not a legal keep.
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{a1, c1}); err == nil {
		t.Fatal("a keep that leaves out the enchantment and the planeswalker was accepted")
	}
	// Two permanents for one type is not legal either.
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{a1, a2, c1, e1, pw}); err == nil {
		t.Fatal("two artifacts were accepted")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, a1, c1, e1, pw)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, opp.ID) == nil {
		t.Fatal("the opponent with two creatures was not asked")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, oc2)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"artifact": a1, "creature": c1, "enchantment": e1, "planeswalker": pw, "land": land, "their land": oppLand, "their kept creature": oc2} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("the %s was sacrificed", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"second artifact": a2, "second creature": c2, "Gearhulk": hulk, "their other creature": oc1} {
		if g.Battlefield.Contains(id) {
			t.Errorf("the %s was kept", name)
		}
	}
}

func TestRprB04CataclysmicGearhulkAskedNothingWhenEveryTypeIsAlreadyOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Lone Bear", "Creature — Bear", 2, 2)
	battle := effectstest.B12Permanent(g, opp.ID, "Siege", "Battle — Siege")
	effectstest.CastCatalogSpell(t, g, "Cataclysmic Gearhulk", "Artifact Creature — Construct", rprB04Gearhulk, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil || effectstest.LatestChooseCardsFor(g, opp.ID) != nil {
		t.Fatal("someone was asked with nothing to choose")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("a lone creature was sacrificed")
	}
	if g.Battlefield.Contains(battle) {
		t.Error("a battle (none of the four types) was kept")
	}
}
