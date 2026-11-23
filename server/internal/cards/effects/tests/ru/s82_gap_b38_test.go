package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b38_test.go — Assassin's Creed cards of S82 GAP batch 38 (#5948).
// Each test plays the real catalog entry.

const (
	gapb38BasimOracle     = "8a70bde5-c514-4dcf-919a-48bdd1c17768"
	gapb38LydiaOracle     = "0315f5e7-2867-428e-8e43-66dda0db57a0"
	gapb38RatonOracle     = "a37f8381-83a4-4b6b-8176-c90f6a61da74"
	gapb38ShayOracle      = "95892392-2ed3-4c1a-bbaa-4083f526769c"
	gapb38ArchersOracle   = "4887b9e1-ea84-4543-b272-c106a13a33dc"
	gapb38EivorOracle     = "cb8289e9-1e59-41f1-a994-6e34d2da2da2"
	gapb38HaviOracle      = "3f58e0f2-7cb5-4a82-b794-071f433fe583"
	gapb38SigurdOracle    = "eedd74c4-42d1-41ba-9051-dc518e45ebbb"
	gapb38ShaunOracle     = "795a2b09-2ed1-4bc8-b6c9-fb26a81e6790"
	gapb38LabyrinthOracle = "00a9bdd0-8a39-4489-bd42-2ac92e550781"
	gapb38SurtrOracle     = "424084cf-418f-47b4-915f-b254c8adb3bd"
)

func TestGapB38CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapb38BasimOracle:     "Basim Ibn Ishaq",
		gapb38LydiaOracle:     "Lydia Frye",
		gapb38RatonOracle:     "Ratonhnhaké꞉ton",
		gapb38ShayOracle:      "Shay Cormac",
		gapb38ArchersOracle:   "Palazzo Archers",
		gapb38EivorOracle:     "Eivor, Wolf-Kissed",
		gapb38HaviOracle:      "Havi, the All-Father",
		gapb38SigurdOracle:    "Sigurd, Jarl of Ravensthorpe",
		gapb38ShaunOracle:     "Shaun & Rebecca, Agents",
		gapb38LabyrinthOracle: "Labyrinth Adversary",
		gapb38SurtrOracle:     "Surtr, Fiery Jötun",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

// gapb38Cast casts a hand card with mana floated for it.
func gapb38Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, mana string) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := gapb38Hand(p, name, typeLine, cost)
	effectstest.FloatForTest(g, p, mana)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
}

// --- Labyrinth Adversary ------------------------------------------------------

func gapb38LabyrinthAttack(t *testing.T, pay bool) (*game.Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adv := effectstest.B12Push(g, me.ID, "Labyrinth Adversary", "Creature — Minotaur", gapb38LabyrinthOracle, 4, 3)
	blocker := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(adv, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatForTest(g, me, "RR")
	effectstest.AnswerPayUnless(t, g, me.ID, pay)
	if pay {
		effectstest.AnswerPickTargetIfOpen(t, g, blocker)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return g, adv, blocker
}

func TestGapB38LabyrinthAdversaryPaidMakesTheTargetUnableToBlock(t *testing.T) {
	g, adv, blocker := gapb38LabyrinthAttack(t, true)
	if !effectstest.EffectiveAbilitiesContain(t, g, adv, "trample") {
		t.Error("Labyrinth Adversary has no trample")
	}
	effectstest.AssertRestrictions(t, g, blocker, game.CantBlock)
}

func TestGapB38LabyrinthAdversaryDeclinedLeavesBlockersAlone(t *testing.T) {
	g, _, blocker := gapb38LabyrinthAttack(t, false)
	effectstest.AssertRestrictions(t, g, blocker, 0)
}

// --- Surtr, Fiery Jötun -------------------------------------------------------

func TestGapB38SurtrDealsThreeOnlyForAHistoricSpell(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		want           bool
	}{
		{"artifact", "Artifact", true},
		{"legendary", "Legendary Creature — Human", true},
		{"saga", "Enchantment — Saga", true},
		{"plain creature", "Creature — Bear", false},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		surtr := effectstest.B12Push(g, me.ID, "Surtr, Fiery Jötun", "Legendary Creature — Giant God Warrior", gapb38SurtrOracle, 5, 5)
		if !effectstest.EffectiveAbilitiesContain(t, g, surtr, "trample") {
			t.Error("Surtr has no trample")
		}
		life := opp.Life
		gapb38Cast(t, g, me, "Test "+tc.name, tc.typeLine, "{1}", "C")
		if tc.want {
			effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := life
		if tc.want {
			want = life - 3
		}
		if opp.Life != want {
			t.Errorf("%s: opponent life %d -> %d, want %d", tc.name, life, opp.Life, want)
		}
	}
}

// --- Basim Ibn Ishaq ----------------------------------------------------------

func TestGapB38BasimDrawsOncePerTurnAndBecomesUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	basim := effectstest.B12Push(g, me.ID, "Basim Ibn Ishaq", "Legendary Creature — Human Assassin", gapb38BasimOracle, 2, 2)
	hand := me.Hand.Size()
	gapb38Cast(t, g, me, "Relic One", "Artifact", "{1}", "C")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("first historic spell: hand %d -> %d, want one card drawn (the spell left)", hand, got)
	}
	effectstest.AssertRestrictions(t, g, basim, game.CantBeBlocked)
	hand = me.Hand.Size()
	gapb38Cast(t, g, me, "Relic Two", "Artifact", "{1}", "C")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("second historic spell the same turn drew: hand %d -> %d", hand, got)
	}
	effectstest.DftMPPassTurn(t, g)
	effectstest.PassTurnsTo(t, g, 0)
	effectstest.AssertRestrictions(t, g, basim, 0)
}

func TestGapB38BasimGrowsWhenItConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	basim := effectstest.B12Push(g, me.ID, "Basim Ibn Ishaq", "Legendary Creature — Human Assassin", gapb38BasimOracle, 2, 2)
	effectstest.AttackWith(t, g, opp.ID, basim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, basim, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

// --- Lydia Frye ---------------------------------------------------------------

func TestGapB38LydiaCantBeBlockedByPowerThreeOrGreater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lydia := effectstest.B12Push(g, me.ID, "Lydia Frye", "Legendary Creature — Human Assassin", gapb38LydiaOracle, 3, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 3, 3)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(lydia, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(big, lydia); err == nil {
		t.Error("a creature with power 3 blocked Lydia")
	}
	if err := g.DeclareBlocker(small, lydia); err != nil {
		t.Errorf("a creature with power 2 couldn't block Lydia: %v", err)
	}
}

func TestGapB38LydiaSurveilsForEachTappedAssassin(t *testing.T) {
	for _, tc := range []struct {
		tapped int
		want   bool
	}{{0, false}, {2, true}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		lydia := effectstest.B12Push(g, me.ID, "Lydia Frye", "Legendary Creature — Human Assassin", gapb38LydiaOracle, 3, 2)
		other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Human Assassin", 1, 1)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
		gapb38Tap(g, bear)
		if tc.tapped == 2 {
			gapb38Tap(g, lydia)
			gapb38Tap(g, other)
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		ch := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID)
		if (ch != nil) != tc.want {
			t.Errorf("%d tapped Assassins: surveil prompt open = %v, want %v", tc.tapped, ch != nil, tc.want)
			continue
		}
		if ch != nil && len(ch.ScryCards) != 2 {
			t.Errorf("surveil looked at %d cards, want 2", len(ch.ScryCards))
		}
	}
}

// --- Palazzo Archers ----------------------------------------------------------

func TestGapB38PalazzoArchersShootsAFlierThatAttacksYou(t *testing.T) {
	for _, flying := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		archers := effectstest.B12Push(g, me.ID, "Palazzo Archers", "Creature — Human Archer", gapb38ArchersOracle, 2, 3)
		if !effectstest.EffectiveAbilitiesContain(t, g, archers, "reach") {
			t.Error("Palazzo Archers has no reach")
		}
		var kw []string
		if flying {
			kw = []string{"flying"}
		}
		attacker := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Raider", TypeLine: "Creature — Bird", Power: 1, Toughness: 2, Keywords: kw})
		effectstest.PassTurnsTo(t, g, 1)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(attacker, me.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		dead := !g.Battlefield.Contains(attacker)
		if dead != flying {
			t.Errorf("flying=%v: attacker died=%v, want %v", flying, dead, flying)
		}
	}
}

// --- Havi, the All-Father -----------------------------------------------------

func gapb38HaviPush(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Havi, the All-Father", OracleID: gapb38HaviOracle,
		TypeLine: "Legendary Creature — God Warrior", ManaCost: "{3}{R}{G}{W}", Power: 6, Toughness: 6,
		Owner: owner, Controller: owner,
	})
}

func gapb38Grave(g *game.Game, p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
		g.BumpLayerVersionForTest()
	})
	return id
}

func TestGapB38HaviIsIndestructibleWithFourHistoricCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	havi := gapb38HaviPush(g, me.ID)
	for i := 0; i < 3; i++ {
		gapb38Grave(g, me, "Relic", "Artifact", "{1}")
	}
	gapb38Grave(g, me, "Plain", "Creature — Bear", "{1}{G}")
	if effectstest.EffectiveAbilitiesContain(t, g, havi, "indestructible") {
		t.Error("Havi is indestructible with three historic cards")
	}
	gapb38Grave(g, me, "Old Saga", "Enchantment — Saga", "{2}{W}")
	if !effectstest.EffectiveAbilitiesContain(t, g, havi, "indestructible") {
		t.Error("Havi is not indestructible with four historic cards")
	}
}

func TestGapB38HaviReturnsALesserLegendaryCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	havi := gapb38HaviPush(g, me.ID)
	cheap := gapb38Grave(g, me, "Cheap Legend", "Legendary Creature — Human", "{1}{R}")
	pricey := gapb38Grave(g, me, "Pricey Legend", "Legendary Creature — Dragon", "{5}{R}{R}")
	plain := gapb38Grave(g, me, "Plain Creature", "Creature — Bear", "{1}")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(havi) })
	g.RunStateChecksForTest()
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if !effectstest.HasID(p.PickTargetCards, cheap) || effectstest.HasID(p.PickTargetCards, pricey) || effectstest.HasID(p.PickTargetCards, plain) {
		t.Errorf("targets %v, want only the cheaper legendary creature", p.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, cheap)
	if !ok || !c.Tapped {
		t.Errorf("the cheaper legend: on battlefield=%v tapped=%v, want both", ok, c.Tapped)
	}
}

func TestGapB38HaviTriggersForAnotherLegendaryCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gapb38HaviPush(g, me.ID)
	legend := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mid Legend", TypeLine: "Legendary Creature — Human", ManaCost: "{3}", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	cheap := gapb38Grave(g, me, "Cheap Legend", "Legendary Creature — Human", "{1}")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(legend) })
	g.RunStateChecksForTest()
	effectstest.AnswerPickTarget(t, g, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the cheaper legend did not return when another legendary creature died")
	}
}

// --- Shaun & Rebecca, Agents --------------------------------------------------

func TestGapB38ShaunAndRebeccaFetchTheAnimusAndTapToMill(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	animus := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: animus, Name: "The Animus", TypeLine: "Legendary Artifact", Owner: me.ID, Controller: me.ID})
	effectstest.CastHoldCreature(t, g, "Shaun & Rebecca, Agents", "Legendary Creature — Human Assassin Scientist", gapb38ShaunOracle, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "The Animus")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(animus) {
		t.Fatal("The Animus was not put onto the battlefield")
	}
	var shaun uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Shaun & Rebecca, Agents" {
			shaun = c.InstanceID
		}
	}
	// Summoning sick: untap-and-ready her by clearing the sickness.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == shaun {
				g.Battlefield.Cards[i].SummonedThisTurn = false
				g.Battlefield.Cards[i].EnteredBattlefieldAt = 0
			}
		}
	})
	grave := me.Graveyard.Size()
	if err := g.ActivateManaAbility(me.ID, shaun, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != grave+2 {
		t.Errorf("graveyard %d -> %d, want two cards milled", grave, got)
	}
}

// --- Eivor, Wolf-Kissed -------------------------------------------------------

func TestGapB38EivorMillsThatManyAndPutsASagaAndALandOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eivor := effectstest.B12Push(g, me.ID, "Eivor, Wolf-Kissed", "Legendary Creature — Human Assassin Warrior", gapb38EivorOracle, 7, 6)
	for _, kw := range []string{"trample", "haste"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, eivor, kw) {
			t.Errorf("Eivor has no %s", kw)
		}
	}
	mk := func(name, typeLine string) uuid.UUID {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		return id
	}
	// Top of library last: 7 cards milled are the last seven pushed.
	bear := mk("Bear", "Creature — Bear")
	saga1 := mk("Saga One", "Enchantment — Saga")
	saga2 := mk("Saga Two", "Enchantment — Saga")
	land1 := mk("Land One", "Land")
	land2 := mk("Land Two", "Land")
	mk("Filler A", "Instant")
	mk("Filler B", "Instant")
	effectstest.AttackWith(t, g, opp.ID, eivor)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick was offered")
	}
	if effectstest.HasID(pick.ChooseCards, bear) {
		t.Error("a creature card was offered")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{saga1, saga2}); err == nil {
		t.Error("two Sagas were accepted")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{land1, land2}); err == nil {
		t.Error("two lands were accepted")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{saga1, land1}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{saga1, land1} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%v was not put onto the battlefield", id)
		}
	}
	if g.Battlefield.Contains(saga2) || g.Battlefield.Contains(land2) {
		t.Error("an unchosen card entered")
	}
}

// --- Ratonhnhaké꞉ton ----------------------------------------------------------

func TestGapB38RatonhnhaketonIsHexproofAndUnblockableUntilItDealsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raton := effectstest.B12Push(g, me.ID, "Ratonhnhaké꞉ton", "Legendary Creature — Human Assassin", gapb38RatonOracle, 3, 3)
	if !effectstest.EffectiveAbilitiesContain(t, g, raton, "hexproof") {
		t.Error("no hexproof before it has dealt damage")
	}
	effectstest.AssertRestrictions(t, g, raton, game.CantBeBlocked)
	effectstest.AttackWith(t, g, opp.ID, raton)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, raton, "hexproof") {
		t.Error("hexproof after it dealt damage")
	}
	effectstest.AssertRestrictions(t, g, raton, 0)
	tokens := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Assassin" {
			tokens++
			if !effectstest.EffectiveAbilitiesContain(t, g, c.InstanceID, "menace") {
				t.Error("the Assassin token has no menace")
			}
		}
	}
	if tokens != 1 {
		t.Errorf("%d Assassin tokens, want 1", tokens)
	}
}

func TestGapB38RatonhnhaketonReturnsAnEquipmentAttachedToTheToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raton := effectstest.B12Push(g, me.ID, "Ratonhnhaké꞉ton", "Legendary Creature — Human Assassin", gapb38RatonOracle, 3, 3)
	sword := gapb38Grave(g, me, "Test Sword", "Artifact — Equipment", "{1}")
	effectstest.AttackWith(t, g, opp.ID, raton)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, sword)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, sword)
	if !ok {
		t.Fatal("the Equipment did not return")
	}
	host, hostOK := effectstest.BattlefieldCard(g, c.AttachedTo.ID)
	if !hostOK || !host.IsToken() || host.Name != "Assassin" {
		t.Errorf("the Equipment is attached to %q (token=%v), want the Assassin token", host.Name, host.IsToken())
	}
}

// --- Sigurd, Jarl of Ravensthorpe ---------------------------------------------

func gapb38Saga(g *game.Game, owner uuid.UUID, lore int) uuid.UUID {
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Saga", TypeLine: "Enchantment — Saga", Owner: owner, Controller: owner,
	})
	if lore > 0 {
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterLore, lore) })
	}
	return id
}

func TestGapB38SigurdBoastRemovesALoreCounterOnlyAfterAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sigurd := effectstest.B12Push(g, me.ID, "Sigurd, Jarl of Ravensthorpe", "Legendary Creature — Human Warrior", gapb38SigurdOracle, 3, 3)
	for _, kw := range []string{"vigilance", "trample", "lifelink"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, sigurd, kw) {
			t.Errorf("Sigurd has no %s", kw)
		}
	}
	saga := gapb38Saga(g, me.ID, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, sigurd, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(saga), Strict: true}); err == nil {
		t.Fatal("boast was activated before Sigurd attacked")
	}
	effectstest.AttackWith(t, g, opp.ID, sigurd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, sigurd, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(saga), Strict: true}); err != nil {
		t.Fatalf("boast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
	if ch == nil {
		t.Fatal("no put-or-remove choice")
	}
	before := effectstest.CounterOn(g, saga, game.CounterLore) // the Saga's own turn-based counter has landed by now
	if err := g.ResolveOptionPick(ch.ID, me.ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, saga, game.CounterLore); got != before-1 {
		t.Errorf("%d lore counters, want %d after removing one", got, before-1)
	}
}

func TestGapB38SigurdPutsAPlusOneCounterWhenYouPutALoreCounterOnASaga(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Sigurd, Jarl of Ravensthorpe", "Legendary Creature — Human Warrior", gapb38SigurdOracle, 3, 3)
	saga := gapb38Saga(g, me.ID, 0)
	theirs := gapb38Saga(g, opp.ID, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(opp.ID, theirs, game.CounterLore, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if pickTargetPromptOpen(g) != nil {
		t.Fatal("an opponent's Saga triggered Sigurd")
	}
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, saga, game.CounterLore, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the other creature has %d +1/+1 counters, want 1", got)
	}
}

// --- Shay Cormac --------------------------------------------------------------

func TestGapB38ShayStripsOpponentsPermanentsOfTheirProtections(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shay := effectstest.B12Push(g, me.ID, "Shay Cormac", "Legendary Creature — Human Knight Rogue", gapb38ShayOracle, 1, 1)
	theirs := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Fortress", TypeLine: "Creature — Ogre", Power: 2, Toughness: 2,
		Keywords: []string{"hexproof", "indestructible", "shroud", "protection from red", "hexproof from white", "flying"}})
	mine := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Mine", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Keywords: []string{"hexproof", "protection from red"}})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, shay, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.EffectiveAbilities(t, g, theirs)
	if len(got) != 1 || got[0] != "flying" {
		t.Errorf("their creature has %v, want only flying", got)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, mine, "hexproof") || !effectstest.EffectiveAbilitiesContain(t, g, mine, "protection from red") {
		t.Error("my own creature lost its keywords")
	}
	late := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Late", TypeLine: "Creature — Ogre", Power: 2, Toughness: 2, Keywords: []string{"hexproof"}})
	if !effectstest.EffectiveAbilitiesContain(t, g, late, "hexproof") {
		t.Error("a permanent that arrived later lost hexproof (the set is locked as the ability resolves)")
	}
	if !gapb38WardSuppressed(g, theirs) {
		t.Error("ward is not suppressed for the stripped creature")
	}
	if gapb38WardSuppressed(g, late) {
		t.Error("ward is suppressed for a permanent the ability did not find")
	}
}

func TestGapB38ShayPutsABountyCounterAndGrowsWhenThatCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shay := effectstest.B12Push(g, me.ID, "Shay Cormac", "Legendary Creature — Human Knight Rogue", gapb38ShayOracle, 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 2, 2)
	bolt := effectstest.HandCardFull(me, "Test Bolt", "Instant", "{R}", "", []string{"R"})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, victim, "bounty"); got != 1 {
		t.Fatalf("%d bounty counters on the targeted creature, want 1", got)
	}
	// Targeting your own creature, or an opponent's creature with a spell
	// they control, puts no bounty counter (checked through the first).
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, shay, game.CounterPlusOne); got != 2 {
		t.Errorf("Shay has %d +1/+1 counters, want 2", got)
	}
}

// gapb38WardSuppressed asks the ward gate about a battlefield permanent.
func gapb38WardSuppressed(g *game.Game, id uuid.UUID) bool {
	out := false
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				out = g.WardSuppressedForEffect(&g.Battlefield.Cards[i])
			}
		}
	})
	return out
}

func gapb38Hand(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

func gapb38Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
}
