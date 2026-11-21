package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// snc_b10_cards_test.go — Streets of New Capenna batch 10 (S82, #5491).

const (
	sncB10WarmWelcome  = "53807ded-5bbb-44a1-9c22-ffdd4b8523e0"
	sncB10Libation     = "aa4f9618-8106-451e-889a-79a5e8159256"
	sncB10Magician     = "758e68fd-76ba-467e-a756-c5daa7db16d5"
	sncB10Briefcase    = "00cd3bf9-0dd6-4ec0-a841-1839a71604ab"
	sncB10Stimulus     = "ce76d17e-d536-46f0-98e2-3e1b1b4f61a4"
	sncB10Partnership  = "a40de3b1-b24b-4da4-b550-9ac9838aecba"
	sncB10Scepter      = "c930b1f3-9bc3-4708-8d05-02295483c9d8"
	sncB10Upstart      = "e0a9c4be-74fe-4984-9028-60380db1575b"
	sncB10Pegasus      = "fd92b588-f273-4206-9887-34e6877fc14f"
	sncB10Bess         = "df72b5d6-6f9e-4d6b-acf6-7dec4ff35468"
	sncB10Gardener     = "9eb7f301-b9b6-48c3-9dad-e7df9cc15355"
	sncB10CreatureType = "Creature — Test"
	sncB10CitizenName  = "Citizen"
)

func TestSncB10CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		sncB10WarmWelcome: "Warm Welcome",
		sncB10Libation:    "Luxurious Libation",
		sncB10Magician:    "Exhibition Magician",
		sncB10Briefcase:   "Courier's Briefcase",
		sncB10Stimulus:    "Stimulus Package",
		sncB10Partnership: "Prosperous Partnership",
		sncB10Scepter:     "Scepter of Celebration",
		sncB10Upstart:     "Brazen Upstart",
		sncB10Pegasus:     "Park Heights Pegasus",
		sncB10Bess:        "Bess, Soul Nourisher",
		sncB10Gardener:    "Civic Gardener",
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
}

func sncB10Citizens(g *game.Game, owner uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == sncB10CitizenName && c.IsToken() {
			out = append(out, c)
		}
	}
	return out
}

func sncB10CheckCitizen(t *testing.T, c game.Card) {
	t.Helper()
	if c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 2 || !c.HasSubtype("Citizen") {
		t.Errorf("token %s %d/%d %v, want a 1/1 green and white Citizen", c.Name, c.Power, c.Toughness, c.Colors)
	}
}

func sncB10Library(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		cards[i].InstanceID = uuid.New()
		cards[i].Owner, cards[i].Controller = p.ID, p.ID
		ids[i] = cards[i].InstanceID
		p.Library.PushTop(cards[i])
	}
	return ids
}

func sncB10Dig() []game.Card {
	return []game.Card{
		{Name: "Forest", TypeLine: "Basic Land — Forest"},
		{Name: "Bear", TypeLine: "Creature — Bear"},
		{Name: "Shock", TypeLine: "Instant"},
		{Name: "Golem", TypeLine: "Artifact Creature — Golem"},
		{Name: "Rancor", TypeLine: "Enchantment — Aura"},
	}
}

func TestSncB10WarmWelcomeDigsForACreatureAndMakesACitizen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := sncB10Library(me, sncB10Dig()...)
	effectstest.CastCatalogSpell(t, g, "Warm Welcome", "Instant", sncB10WarmWelcome, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick opened")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if len(offered) != 2 || !offered[top[1]] || !offered[top[3]] || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (min %d max %d), want the two creature cards, up to one", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[1]) {
		t.Error("the creature card is not in hand")
	}
	effectstest.PassPriorityAroundTable(t, g)
	cs := sncB10Citizens(g, me.ID)
	if len(cs) != 1 {
		t.Fatalf("%d Citizens, want 1", len(cs))
	}
	sncB10CheckCitizen(t, cs[0])
}

func TestSncB10WarmWelcomeMakesACitizenEvenIfYouTakeNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sncB10Library(me, sncB10Dig()...)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Warm Welcome", "Instant", sncB10WarmWelcome, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, me.ID) != nil {
		effectstest.AnswerChooseCards(t, g, me.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d after taking nothing", hand, me.Hand.Size())
	}
	if n := len(sncB10Citizens(g, me.ID)); n != 1 {
		t.Errorf("%d Citizens, want 1", n)
	}
}

func TestSncB10LuxuriousLibationPumpsByXAndMakesACitizen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", sncB10CreatureType, 2, 2)
	effectstest.FloatForTest(g, me, "GGGG")
	if _, err := effectstest.CastWithTapParams(t, g, "Luxurious Libation", "Instant", "{X}{G}", sncB10Libation,
		game.CastSpellParams{Strict: true, XValue: 3, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.PtOf(t, g, bear); p != 5 || tg != 5 {
		t.Errorf("Bear is %d/%d, want 5/5", p, tg)
	}
	cs := sncB10Citizens(g, me.ID)
	if len(cs) != 1 {
		t.Fatalf("%d Citizens, want 1", len(cs))
	}
	sncB10CheckCitizen(t, cs[0])
}

func TestSncB10LuxuriousLibationNeedsACreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	effectstest.FloatForTest(g, me, "GG")
	_, err := effectstest.CastWithTapParams(t, g, "Luxurious Libation", "Instant", "{X}{G}", sncB10Libation,
		game.CastSpellParams{Strict: true, XValue: 1, Targets: effectstest.CardRefs(relic)})
	if err == nil {
		t.Fatal("cast with a noncreature target")
	}
}

func TestSncB10ExhibitionMagicianChoosesACitizenOrATreasure(t *testing.T) {
	for _, tc := range []struct {
		name      string
		mode      int
		citizens  int
		treasures int
	}{{"citizen", 0, 1, 0}, {"treasure", 1, 0, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.CastRingCreature(t, g, "Exhibition Magician", "Creature — Human Wizard", sncB10Magician, 2, 1)
			effectstest.PassPriorityAroundTable(t, g)
			pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
			if pick == nil {
				t.Fatalf("no mode prompt: %+v", g.PendingChoices)
			}
			if err := g.ResolveModePick(pick.ID, me.ID, []int{tc.mode}); err != nil {
				t.Fatalf("ResolveModePick: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if n := len(sncB10Citizens(g, me.ID)); n != tc.citizens {
				t.Errorf("%d Citizens, want %d", n, tc.citizens)
			}
			if n := effectstest.CountTreasures(g, me.ID); n != tc.treasures {
				t.Errorf("%d Treasures, want %d", n, tc.treasures)
			}
		})
	}
}

func TestSncB10CouriersBriefcaseMakesACitizenAndIsATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.CastCatalogSpell(t, g, "Courier's Briefcase", "Artifact — Treasure", sncB10Briefcase, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(sncB10Citizens(g, me.ID)); n != 1 {
		t.Fatalf("%d Citizens, want 1", n)
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok || !c.HasSubtype("Treasure") || c.IsToken() {
		t.Fatalf("the Briefcase is not a nontoken Treasure: %+v", c)
	}
}

func TestSncB10CouriersBriefcaseSacrificesForManaAndDrawsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	briefcase := effectstest.B12Push(g, me.ID, "Courier's Briefcase", "Artifact — Treasure", sncB10Briefcase, 0, 0)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	effectstest.B06AddMana(me, "W", "U", "B", "R", "G")
	effectstest.B16Activate(t, g, me.ID, briefcase, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d -> %d, want +3", hand, me.Hand.Size())
	}
	if g.Battlefield.Contains(briefcase) {
		t.Error("the Briefcase was not sacrificed")
	}
}

func TestSncB10CouriersBriefcaseDrawNeedsAllFiveColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	briefcase := effectstest.B12Push(g, me.ID, "Courier's Briefcase", "Artifact — Treasure", sncB10Briefcase, 0, 0)
	effectstest.B06AddMana(me, "W", "U", "B", "R")
	if err := g.ActivateCatalogAbility(me.ID, briefcase, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("drew three without {G}")
	}
	if !g.Battlefield.Contains(briefcase) {
		t.Error("the Briefcase was sacrificed by a refused activation")
	}
}

func TestSncB10StimulusPackageMakesTwoTreasuresAndSacrificesOneForACitizen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pkg := effectstest.CastCatalogSpell(t, g, "Stimulus Package", "Enchantment", sncB10Stimulus, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountTreasures(g, me.ID); n != 2 {
		t.Fatalf("%d Treasures, want 2", n)
	}
	var treasure uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" {
			treasure = c.InstanceID
		}
	}
	effectstest.B16Activate(t, g, me.ID, pkg, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{treasure}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(treasure) {
		t.Error("the Treasure was not sacrificed")
	}
	cs := sncB10Citizens(g, me.ID)
	if len(cs) != 1 {
		t.Fatalf("%d Citizens, want 1", len(cs))
	}
	sncB10CheckCitizen(t, cs[0])
}

func TestSncB10StimulusPackageRefusesANonTreasureSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	pkg := effectstest.B12Push(g, me.ID, "Stimulus Package", "Enchantment", sncB10Stimulus, 0, 0)
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, pkg, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{relic}}); err == nil {
		t.Fatal("sacrificed a non-Treasure")
	}
}

func TestSncB10ProsperousPartnershipMakesTwoCitizensAndTapsThreeCreaturesForATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pp := effectstest.CastCatalogSpell(t, g, "Prosperous Partnership", "Enchantment", sncB10Partnership, nil)
	effectstest.PassPriorityAroundTable(t, g)
	cs := sncB10Citizens(g, me.ID)
	if len(cs) != 2 {
		t.Fatalf("%d Citizens, want 2", len(cs))
	}
	sncB10CheckCitizen(t, cs[0])
	third := effectstest.B12Creature(g, me.ID, "Third", sncB10CreatureType, 2, 2)
	ids := []uuid.UUID{cs[0].InstanceID, cs[1].InstanceID, third}
	if err := g.ActivateCatalogAbility(me.ID, pp, 0, game.ActivateAbilityParams{Strict: true, TapIDs: ids[:2]}); err == nil {
		t.Fatal("two creatures paid a three-creature cost")
	}
	// Summoning-sick tokens may be tapped: it is not the {T} symbol.
	effectstest.B16Activate(t, g, me.ID, pp, 0, game.ActivateAbilityParams{Strict: true, TapIDs: ids})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountTreasures(g, me.ID); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
	for _, id := range ids {
		if !effectstest.IsTapped(g, id) {
			t.Errorf("creature %v was not tapped", id)
		}
	}
}

func sncB10Equip(g *game.Game, owner, host uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Scepter of Celebration",
		TypeLine: "Artifact — Equipment", OracleID: sncB10Scepter, Owner: owner, Controller: owner,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host}})
}

func TestSncB10ScepterOfCelebrationPumpsAndMakesThatManyCitizens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "Host", sncB10CreatureType, 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Other", sncB10CreatureType, 2, 2)
	sncB10Equip(g, me.ID, host)
	if p, _ := effectstest.PtOf(t, g, host); p != 3 {
		t.Errorf("Host power %d, want 3", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, host, "trample") {
		t.Error("the equipped creature lacks trample")
	}
	effectstest.DealCombatDamageToPlayer(g, other, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(sncB10Citizens(g, me.ID)); n != 0 {
		t.Fatalf("%d Citizens after an unequipped creature connected", n)
	}
	effectstest.DealCombatDamageToPlayer(g, host, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	cs := sncB10Citizens(g, me.ID)
	if len(cs) != 3 {
		t.Fatalf("%d Citizens, want 3", len(cs))
	}
	sncB10CheckCitizen(t, cs[0])
}

func TestSncB10BrazenUpstartDigsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	up := effectstest.B12Push(g, me.ID, "Brazen Upstart", "Creature — Elf Shaman", sncB10Upstart, 4, 2)
	top := sncB10Library(me, sncB10Dig()...)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(up) })
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick opened")
	}
	if len(pick.ChooseCards) != 2 || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (min %d max %d), want the two creature cards, up to one", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[3]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[3]) {
		t.Error("the creature card is not in hand")
	}
}

func TestSncB10ParkHeightsPegasusDrawsOnlyAfterTwoCreaturesEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	pegasus := effectstest.B12Push(g, me.ID, "Park Heights Pegasus", "Creature — Pegasus", sncB10Pegasus, 2, 1)
	effectstest.SeedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 green and white Citizen"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DealCombatDamageToPlayer(g, pegasus, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("drew with one creature entered: hand %d -> %d", hand, me.Hand.Size())
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 green and white Citizen"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DealCombatDamageToPlayer(g, pegasus, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d after two creatures entered, want +1", hand, me.Hand.Size())
	}
}

func TestSncB10BessGrowsOnceForBaseOneByOnesAndPumpsThemOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bess := effectstest.B12Push(g, me.ID, "Bess, Soul Nourisher", "Legendary Creature — Human Citizen", sncB10Bess, 1, 1)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 green and white Citizen"), 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bess, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d counters after two 1/1s entered together, want 1", n)
	}
	effectstest.B12Creature(g, me.ID, "Bear", sncB10CreatureType, 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 black Zombie"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bess, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d counters after a 2/2 entered, want 1", n)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bess {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.DeclareAttack(t, g, opp.ID, bess)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	var ones, bears int
	for _, c := range g.Battlefield.Cards {
		if c.Controller != me.ID || c.InstanceID == bess {
			continue
		}
		p, tg := effectstest.PtOf(t, g, c.InstanceID)
		switch {
		case c.Name == sncB10CitizenName && p == 2 && tg == 2:
			ones++
		case c.Name == "Bear" && p == 2 && tg == 2:
			bears++
		}
	}
	if ones != 2 || bears != 1 {
		t.Errorf("pumped Citizens %d (want 2), unchanged Bear %d (want 1)", ones, bears)
	}
}

func TestSncB10CivicGardenerUntapsACreatureOrLandOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gardener := effectstest.B12Push(g, me.ID, "Civic Gardener", "Creature — Human Citizen", sncB10Gardener, 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.TapForTest(t, g, land)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == gardener {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.DeclareAttack(t, g, opp.ID, gardener)
	effectstest.LockInAttacks(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, land)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, land) {
		t.Error("the land is still tapped")
	}
}
