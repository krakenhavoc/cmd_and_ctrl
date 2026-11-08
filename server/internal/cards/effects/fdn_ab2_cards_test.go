package effects

import (
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
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
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		assertKeywords(t, g, id, c.keywords...)
	}
}

// --- Adventuring Gear ------------------------------------------------------

func TestFdnAB2AdventuringGearPumpsTheEquippedCreatureOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	other := seedBear(g, me.ID)
	gear := seedEquipment(g, me.ID, "Adventuring Gear", fdnAB2GearOracle)
	equipTo(t, g, me.ID, gear, bear)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("equipped Bear is %d/%d after a land drop, want 4/4", p, tough)
	}
	if p := effectivePower(t, g, other); p != 2 {
		t.Errorf("the unequipped Bear has power %d, want 2", p)
	}
	advanceToEndStepOf(t, g, 0)
	advanceToUpkeepOf(t, g, 1)
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("the pump outlived the turn: power %d", p)
	}
}

func TestFdnAB2AdventuringGearUnattachedPumpsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	seedEquipment(g, me.ID, "Adventuring Gear", fdnAB2GearOracle)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("an unattached Gear pumped a creature to power %d", p)
	}
}

// --- Ajani, Caller of the Pride ---------------------------------------------

func fdnAB2Ajani(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	return pushCatalogWalkerWithLine(g, owner, "Ajani, Caller of the Pride", "Legendary Planeswalker — Ajani", fdnAB2AjaniOracle, loyalty)
}

func TestFdnAB2AjaniPlusOnePutsACounterOnUpToOneTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	ajani := fdnAB2Ajani(g, me.ID, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b16Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	passPriorityAroundTable(t, g)
	if n := loyaltyCount(g, ajani); n != 5 {
		t.Errorf("loyalty %d, want 5", n)
	}
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
}

func TestFdnAB2AjaniMinusThreeGrantsFlyingAndDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	ajani := fdnAB2Ajani(g, me.ID, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b16Activate(t, g, me.ID, ajani, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	passPriorityAroundTable(t, g)
	if n := loyaltyCount(g, ajani); n != 1 {
		t.Errorf("loyalty %d, want 1", n)
	}
	for _, kw := range []string{"flying", "double strike"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the target lacks %s", kw)
		}
	}
	advanceToEndStepOf(t, g, 0)
	advanceToUpkeepOf(t, g, 1)
	if hasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("double strike outlived the turn")
	}
}

func TestFdnAB2AjaniMinusEightMakesACatPerLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	ajani := fdnAB2Ajani(g, me.ID, 8)
	me.Life = 5
	b16Activate(t, g, me.ID, ajani, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	cats := battlefieldIDsNamed(g, "Cat")
	if len(cats) != 5 {
		t.Fatalf("%d Cats at 5 life, want 5", len(cats))
	}
	if p, tough := effectivePower(t, g, cats[0]), effectiveToughness(t, g, cats[0]); p != 2 || tough != 2 {
		t.Errorf("a Cat is %d/%d, want 2/2", p, tough)
	}
}

// --- Alesha, Who Laughs at Fate ---------------------------------------------

func TestFdnAB2AleshaGrowsOnAttackAndReturnsByHerPowerAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	three := pushGraveyardPermanent(me, "Three Drop", "Creature — Bear", "{2}{B}")
	four := pushGraveyardPermanent(me, "Four Drop", "Creature — Bear", "{3}{B}")
	rock := pushGraveyardPermanent(me, "Cheap Rock", "Artifact", "{1}")
	alesha := b12Push(g, me.ID, "Alesha, Who Laughs at Fate", "Legendary Creature — Human Warrior", fdnAB2AleshaOracle, 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, alesha)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, alesha, game.CounterPlusOne); n != 1 {
		t.Fatalf("Alesha has %d +1/+1 counters after attacking, want 1", n)
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	for _, bad := range []uuid.UUID{four, rock} {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Errorf("an illegal card %v was accepted", bad)
		}
	}
	pickTriggerTarget(t, g, me.ID, three)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(three) {
		t.Error("the mana value 3 creature card did not return")
	}
}

func TestFdnAB2AleshaWithoutAnAttackDoesNotTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushGraveyardPermanent(me, "Two Drop", "Creature — Bear", "{1}{B}")
	alesha := b12Push(g, me.ID, "Alesha, Who Laughs at Fate", "Legendary Creature — Human Warrior", fdnAB2AleshaOracle, 2, 2)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || triggersOnStackFrom(g, alesha) != 0 {
		t.Fatal("raid triggered without an attack")
	}
}

// --- Armasaur Guide ---------------------------------------------------------

func TestFdnAB2ArmasaurGuideCountersATargetWhenThreeAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guide := b12Push(g, me.ID, "Armasaur Guide", "Creature — Dinosaur", fdnAB2ArmasaurOracle, 4, 4)
	a := pushVanillaCreature(g, me.ID, "A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "B", 2, 2)
	c := pushVanillaCreature(g, me.ID, "C", 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, a, b, c)
	pickTriggerTarget(t, g, me.ID, guide)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, guide, game.CounterPlusOne); n != 1 {
		t.Errorf("the target has %d +1/+1 counters, want 1", n)
	}
}

func TestFdnAB2ArmasaurGuideIgnoresTwoAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guide := b12Push(g, me.ID, "Armasaur Guide", "Creature — Dinosaur", fdnAB2ArmasaurOracle, 4, 4)
	a := pushVanillaCreature(g, me.ID, "A", 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, guide, a)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || triggersOnStackFrom(g, guide) != 0 {
		t.Error("two attackers triggered the Guide")
	}
	if n := countersOn(g, guide, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters after attacking with two, want 0", n)
	}
}

// --- Ashroot Animist --------------------------------------------------------

func TestFdnAB2AshrootAnimistPumpsAnotherCreatureByItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	animist := b12Push(g, me.ID, "Ashroot Animist", "Creature — Lizard Druid", fdnAB2AshrootOracle, 4, 4)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, animist)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the attack trigger asked for no target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: animist}); err == nil {
		t.Fatal("the Animist targeted itself")
	}
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("the Bear is %d/%d, want 6/6", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear did not gain trample")
	}
}

// --- Ayli, Eternal Pilgrim --------------------------------------------------

func TestFdnAB2AyliGainsTheSacrificedCreaturesToughness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ayli := b12Push(g, me.ID, "Ayli, Eternal Pilgrim", "Legendary Creature — Kor Cleric", fdnAB2AyliOracle, 2, 3)
	wall := pushVanillaCreature(g, me.ID, "Wall", 0, 5)
	life := me.Life
	b16Activate(t, g, me.ID, ayli, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{wall}})
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	ayli := b12Push(g, me.ID, "Ayli, Eternal Pilgrim", "Legendary Creature — Kor Cleric", fdnAB2AyliOracle, 2, 3)
	fodder := pushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID,
	})
	params := game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
		SacrificeIDs: []uuid.UUID{fodder},
	}
	me.Life = game.StartingLife + 9
	acRefused(t, g, me.ID, ayli, 1, params)
	me.Life = game.StartingLife + 10
	b16Activate(t, g, me.ID, ayli, 1, params)
	passPriorityAroundTable(t, g)
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
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		b27Push(g, me.ID, "Ballyrush Banneret", "Creature — Kithkin Soldier", fdnAB2BannerOracle, "{1}{W}", 2, 1, "W")
		_, err := trekB37CastFromHand(t, g, game.Card{Name: "Spell", TypeLine: tc.typeLine, ManaCost: tc.cost, Power: 1, Toughness: 1}, tc.lands)
		if (err == nil) != tc.ok {
			t.Errorf("%s %s with %d Mountains: cast error %v, want ok=%v", tc.typeLine, tc.cost, tc.lands, err, tc.ok)
		}
	}
}

// --- Balmor, Battlemage Captain ---------------------------------------------

func TestFdnAB2BalmorPumpsTheTeamOnAnInstant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	balmor := b12Push(g, me.ID, "Balmor, Battlemage Captain", "Legendary Creature — Bird Wizard", fdnAB2BalmorOracle, 1, 3)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	for _, c := range []struct {
		id          uuid.UUID
		power, toug int
	}{{balmor, 2, 3}, {bear, 3, 2}} {
		if p, tough := effectivePower(t, g, c.id), effectiveToughness(t, g, c.id); p != c.power || tough != c.toug {
			t.Errorf("%v is %d/%d, want %d/%d", c.id, p, tough, c.power, c.toug)
		}
		if !hasEffectiveKeyword(t, g, c.id, "trample") {
			t.Errorf("%v did not gain trample", c.id)
		}
	}
	if p := effectivePower(t, g, theirs); p != 2 || hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature was pumped")
	}
}

// --- Biogenic Upgrade -------------------------------------------------------

func TestFdnAB2BiogenicUpgradeDistributesThenDoubles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == b {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 1}
			}
		}
	})
	d1658Cast(t, g, "Biogenic Upgrade", "Sorcery", fdnAB2BiogenicOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 1},
	})
	passPriorityAroundTable(t, g)
	if n := countersOn(g, a, game.CounterPlusOne); n != 4 {
		t.Errorf("a: %d +1/+1 counters, want 4 (2 doubled)", n)
	}
	if n := countersOn(g, b, game.CounterPlusOne); n != 4 {
		t.Errorf("b: %d +1/+1 counters, want 4 (1 already + 1, doubled)", n)
	}
}

func TestFdnAB2BiogenicUpgradeRefusesMoreThanThreeTargets(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := make([]uuid.UUID, 4)
	for i := range ids {
		ids[i] = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	}
	err := d1658CastErr(t, g, "Biogenic Upgrade", "Sorcery", fdnAB2BiogenicOracle, game.CastSpellParams{
		Targets: cardRefs(ids...), Distribution: map[uuid.UUID]int{ids[0]: 1, ids[1]: 1, ids[2]: 1, ids[3]: 0},
	})
	if err == nil {
		t.Error("four targets were accepted")
	}
}

// --- Bushwhack --------------------------------------------------------------

func TestFdnAB2BushwhackFetchesABasicLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	castCatalogSpellWithModes(t, g, "Bushwhack", "Sorcery", fdnAB2BushwhackOracle, []int{0})
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerSearchNamed(t, g, me.ID, "Forest")
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
}

func TestFdnAB2BushwhackFights(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pr7Creature(g, me.ID, "Mine", 4, "G")
	theirs := pr7Creature(g, opp.ID, "Theirs", 1, "B")
	toMainForCost(t, g)
	id := handCardFull(me, "Bushwhack", "Sorcery", "", fdnAB2BushwhackOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: theirs, Slot: 1},
	}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, theirs); got != 4 && g.Battlefield.Contains(theirs) {
		t.Errorf("their creature took %d, want 4", got)
	}
	if got := pr6Marked(g, mine); got != 1 {
		t.Errorf("your creature took %d, want 1", got)
	}
}
