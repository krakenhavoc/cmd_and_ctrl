package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdn_ab2_cards_test.go — Foundations, A–B batch 2 (#4465): the A–B cards
// the first A–B batch left unattempted. Each test plays the real catalog
// entry through the engine and asserts what its text says.

const (
	fdnAB2GearOracle      = "17042684-7af4-43a6-88c7-885032cfb27c"
	fdnAB2AjaniOracle     = "1d5ff280-7f4d-4801-aefd-9b6e6b1f2818"
	fdnAB2AleshaOracle    = "17ce2c17-b69b-4ebf-9c3c-c64e771af101"
	fdnAB2ArmasaurOracle  = "1d4e7422-2e26-49ce-beef-08f61583089a"
	fdnAB2AshrootOracle   = "3f008b57-6471-4785-9571-3b90193da7e2"
	fdnAB2AyliOracle      = "e6a35729-edee-4a8c-9aa1-041114d66c69"
	fdnAB2BannerOracle    = "0ad27d0e-b6c3-4c3f-a389-285d889b6a65"
	fdnAB2BalmorOracle    = "7da19fc5-cb85-4ec7-beea-95210378bbfe"
	fdnAB2BiogenicOracle  = "30812870-87c0-469c-95a5-811d5a181e72"
	fdnAB2BushwhackOracle = "c61374e5-a7f6-455e-a40b-a481751b536b"
)

func TestFdnAB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		fdnAB2GearOracle:      "Adventuring Gear",
		fdnAB2AjaniOracle:     "Ajani, Caller of the Pride",
		fdnAB2AleshaOracle:    "Alesha, Who Laughs at Fate",
		fdnAB2ArmasaurOracle:  "Armasaur Guide",
		fdnAB2AshrootOracle:   "Ashroot Animist",
		fdnAB2AyliOracle:      "Ayli, Eternal Pilgrim",
		fdnAB2BannerOracle:    "Ballyrush Banneret",
		fdnAB2BalmorOracle:    "Balmor, Battlemage Captain",
		fdnAB2BiogenicOracle:  "Biogenic Upgrade",
		fdnAB2BushwhackOracle: "Bushwhack",
	} {
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

func TestFdnAB2PrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Alesha, Who Laughs at Fate", "Legendary Creature — Human Warrior", fdnAB2AleshaOracle, 2, 2, []string{"first strike"}},
		{"Armasaur Guide", "Creature — Dinosaur", fdnAB2ArmasaurOracle, 4, 4, []string{"vigilance"}},
		{"Ashroot Animist", "Creature — Lizard Druid", fdnAB2AshrootOracle, 4, 4, []string{"trample"}},
		{"Ayli, Eternal Pilgrim", "Legendary Creature — Kor Cleric", fdnAB2AyliOracle, 2, 3, []string{"deathtouch"}},
		{"Balmor, Battlemage Captain", "Legendary Creature — Bird Wizard", fdnAB2BalmorOracle, 1, 3, []string{"flying"}},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.B12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		effectstest.AssertKeywords(t, g, id, c.keywords...)
	}
}

// --- Adventuring Gear ------------------------------------------------------

func TestFdnAB2AdventuringGearPumpsTheEquippedCreatureOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	other := effectstest.SeedBear(g, me.ID)
	gear := effectstest.SeedEquipment(g, me.ID, "Adventuring Gear", fdnAB2GearOracle)
	effectstest.EquipTo(t, g, me.ID, gear, bear)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("equipped Bear is %d/%d after a land drop, want 4/4", p, tough)
	}
	if p := effectstest.EffectivePower(t, g, other); p != 2 {
		t.Errorf("the unequipped Bear has power %d, want 2", p)
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("the pump outlived the turn: power %d", p)
	}
}

func TestFdnAB2AdventuringGearUnattachedPumpsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.SeedEquipment(g, me.ID, "Adventuring Gear", fdnAB2GearOracle)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("an unattached Gear pumped a creature to power %d", p)
	}
}

// --- Ajani, Caller of the Pride ---------------------------------------------

func fdnAB2Ajani(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	return effectstest.PushCatalogWalkerWithLine(g, owner, "Ajani, Caller of the Pride", "Legendary Planeswalker — Ajani", fdnAB2AjaniOracle, loyalty)
}

func TestFdnAB2AjaniPlusOnePutsACounterOnUpToOneTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	ajani := fdnAB2Ajani(g, me.ID, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B16Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.LoyaltyCount(g, ajani); n != 5 {
		t.Errorf("loyalty %d, want 5", n)
	}
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
}

func TestFdnAB2AjaniMinusThreeGrantsFlyingAndDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	ajani := fdnAB2Ajani(g, me.ID, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B16Activate(t, g, me.ID, ajani, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.LoyaltyCount(g, ajani); n != 1 {
		t.Errorf("loyalty %d, want 1", n)
	}
	for _, kw := range []string{"flying", "double strike"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the target lacks %s", kw)
		}
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("double strike outlived the turn")
	}
}

func TestFdnAB2AjaniMinusEightMakesACatPerLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	ajani := fdnAB2Ajani(g, me.ID, 8)
	me.Life = 5
	effectstest.B16Activate(t, g, me.ID, ajani, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	cats := effectstest.BattlefieldIDsNamed(g, "Cat")
	if len(cats) != 5 {
		t.Fatalf("%d Cats at 5 life, want 5", len(cats))
	}
	if p, tough := effectstest.EffectivePower(t, g, cats[0]), effectstest.EffectiveToughness(t, g, cats[0]); p != 2 || tough != 2 {
		t.Errorf("a Cat is %d/%d, want 2/2", p, tough)
	}
}

// --- Alesha, Who Laughs at Fate ---------------------------------------------

func TestFdnAB2AleshaGrowsOnAttackAndReturnsByHerPowerAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	three := effectstest.PushGraveyardPermanent(me, "Three Drop", "Creature — Bear", "{2}{B}")
	four := effectstest.PushGraveyardPermanent(me, "Four Drop", "Creature — Bear", "{3}{B}")
	rock := effectstest.PushGraveyardPermanent(me, "Cheap Rock", "Artifact", "{1}")
	alesha := effectstest.B12Push(g, me.ID, "Alesha, Who Laughs at Fate", "Legendary Creature — Human Warrior", fdnAB2AleshaOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, alesha)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, alesha, game.CounterPlusOne); n != 1 {
		t.Fatalf("Alesha has %d +1/+1 counters after attacking, want 1", n)
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	for _, bad := range []uuid.UUID{four, rock} {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Errorf("an illegal card %v was accepted", bad)
		}
	}
	effectstest.PickTriggerTarget(t, g, me.ID, three)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(three) {
		t.Error("the mana value 3 creature card did not return")
	}
}

func TestFdnAB2AleshaWithoutAnAttackDoesNotTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushGraveyardPermanent(me, "Two Drop", "Creature — Bear", "{1}{B}")
	alesha := effectstest.B12Push(g, me.ID, "Alesha, Who Laughs at Fate", "Legendary Creature — Human Warrior", fdnAB2AleshaOracle, 2, 2)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.TriggersOnStackFrom(g, alesha) != 0 {
		t.Fatal("raid triggered without an attack")
	}
}

// --- Armasaur Guide ---------------------------------------------------------

func TestFdnAB2ArmasaurGuideCountersATargetWhenThreeAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guide := effectstest.B12Push(g, me.ID, "Armasaur Guide", "Creature — Dinosaur", fdnAB2ArmasaurOracle, 4, 4)
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 2, 2)
	c := effectstest.PushVanillaCreature(g, me.ID, "C", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, a, b, c)
	effectstest.PickTriggerTarget(t, g, me.ID, guide)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, guide, game.CounterPlusOne); n != 1 {
		t.Errorf("the target has %d +1/+1 counters, want 1", n)
	}
}

func TestFdnAB2ArmasaurGuideIgnoresTwoAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guide := effectstest.B12Push(g, me.ID, "Armasaur Guide", "Creature — Dinosaur", fdnAB2ArmasaurOracle, 4, 4)
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, guide, a)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.TriggersOnStackFrom(g, guide) != 0 {
		t.Error("two attackers triggered the Guide")
	}
	if n := effectstest.CountersOn(g, guide, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters after attacking with two, want 0", n)
	}
}

// --- Ashroot Animist --------------------------------------------------------

func TestFdnAB2AshrootAnimistPumpsAnotherCreatureByItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	animist := effectstest.B12Push(g, me.ID, "Ashroot Animist", "Creature — Lizard Druid", fdnAB2AshrootOracle, 4, 4)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, animist)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the attack trigger asked for no target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: animist}); err == nil {
		t.Fatal("the Animist targeted itself")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("the Bear is %d/%d, want 6/6", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear did not gain trample")
	}
}

// --- Ayli, Eternal Pilgrim --------------------------------------------------

func TestFdnAB2AyliGainsTheSacrificedCreaturesToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ayli := effectstest.B12Push(g, me.ID, "Ayli, Eternal Pilgrim", "Legendary Creature — Kor Cleric", fdnAB2AyliOracle, 2, 3)
	wall := effectstest.PushVanillaCreature(g, me.ID, "Wall", 0, 5)
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, ayli, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{wall}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d → %d, want +5", life, me.Life)
	}
	if g.Battlefield.Contains(wall) {
		t.Error("the Wall was not sacrificed")
	}
	if err := g.ActivateCatalogAbility(me.ID, ayli, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{ayli}}); err == nil {
		t.Error("Ayli sacrificed herself to her own ability")
	}
}

func TestFdnAB2AyliExilesOnlyAtTenAboveStartingLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ayli := effectstest.B12Push(g, me.ID, "Ayli, Eternal Pilgrim", "Legendary Creature — Kor Cleric", fdnAB2AyliOracle, 2, 3)
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID,
	})
	params := game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
		SacrificeIDs: []uuid.UUID{fodder},
	}
	me.Life = game.StartingLife + 9
	effectstest.AcRefused(t, g, me.ID, ayli, 1, params)
	me.Life = game.StartingLife + 10
	effectstest.B16Activate(t, g, me.ID, ayli, 1, params)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the target nonland permanent was not exiled")
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice was not paid")
	}
}

// --- Ballyrush Banneret -----------------------------------------------------

func TestFdnAB2BallyrushBanneretDiscountsKithkinAndSoldierSpellsOnce(t *testing.T) {
	cases := []struct {
		typeLine, cost string
		lands          int
		ok             bool
	}{
		{"Creature — Kithkin Rogue", "{2}{R}", 2, true},
		{"Creature — Human Soldier", "{2}{R}", 2, true},
		{"Creature — Kithkin Soldier", "{3}{R}", 2, false}, // {1} less once, not twice
		{"Creature — Kithkin Soldier", "{3}{R}", 3, true},
		{"Creature — Human Warrior", "{2}{R}", 2, false},
	}
	for _, tc := range cases {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.B27Push(g, me.ID, "Ballyrush Banneret", "Creature — Kithkin Soldier", fdnAB2BannerOracle, "{1}{W}", 2, 1, "W")
		_, err := effectstest.TrekB37CastFromHand(t, g, game.Card{Name: "Spell", TypeLine: tc.typeLine, ManaCost: tc.cost, Power: 1, Toughness: 1}, tc.lands)
		if (err == nil) != tc.ok {
			t.Errorf("%s %s with %d Mountains: cast error %v, want ok=%v", tc.typeLine, tc.cost, tc.lands, err, tc.ok)
		}
	}
}

// --- Balmor, Battlemage Captain ---------------------------------------------

func TestFdnAB2BalmorPumpsTheTeamOnAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	balmor := effectstest.B12Push(g, me.ID, "Balmor, Battlemage Captain", "Legendary Creature — Bird Wizard", fdnAB2BalmorOracle, 1, 3)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []struct {
		id          uuid.UUID
		power, toug int
	}{{balmor, 2, 3}, {bear, 3, 2}} {
		if p, tough := effectstest.EffectivePower(t, g, c.id), effectstest.EffectiveToughness(t, g, c.id); p != c.power || tough != c.toug {
			t.Errorf("%v is %d/%d, want %d/%d", c.id, p, tough, c.power, c.toug)
		}
		if !effectstest.HasEffectiveKeyword(t, g, c.id, "trample") {
			t.Errorf("%v did not gain trample", c.id)
		}
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature was pumped")
	}
}

// --- Biogenic Upgrade -------------------------------------------------------

func TestFdnAB2BiogenicUpgradeDistributesThenDoubles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == b {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 1}
			}
		}
	})
	effectstest.D1658Cast(t, g, "Biogenic Upgrade", "Sorcery", fdnAB2BiogenicOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, a, game.CounterPlusOne); n != 4 {
		t.Errorf("a: %d +1/+1 counters, want 4 (2 doubled)", n)
	}
	if n := effectstest.CountersOn(g, b, game.CounterPlusOne); n != 4 {
		t.Errorf("b: %d +1/+1 counters, want 4 (1 already + 1, doubled)", n)
	}
}

func TestFdnAB2BiogenicUpgradeRefusesMoreThanThreeTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := make([]uuid.UUID, 4)
	for i := range ids {
		ids[i] = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	}
	err := effectstest.D1658CastErr(t, g, "Biogenic Upgrade", "Sorcery", fdnAB2BiogenicOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(ids...), Distribution: map[uuid.UUID]int{ids[0]: 1, ids[1]: 1, ids[2]: 1, ids[3]: 0},
	})
	if err == nil {
		t.Error("four targets were accepted")
	}
}

// --- Bushwhack --------------------------------------------------------------

func TestFdnAB2BushwhackFetchesABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.CastCatalogSpellWithModes(t, g, "Bushwhack", "Sorcery", fdnAB2BushwhackOracle, []int{0})
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
}

func TestFdnAB2BushwhackFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.Pr7Creature(g, me.ID, "Mine", 4, "G")
	theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 1, "B")
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandCardFull(me, "Bushwhack", "Sorcery", "", fdnAB2BushwhackOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: theirs, Slot: 1},
	}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, theirs); got != 4 && g.Battlefield.Contains(theirs) {
		t.Errorf("their creature took %d, want 4", got)
	}
	if got := effectstest.Pr6Marked(g, mine); got != 1 {
		t.Errorf("your creature took %d, want 1", got)
	}
}
