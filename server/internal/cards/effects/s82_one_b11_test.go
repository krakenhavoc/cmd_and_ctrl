package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b11_test.go — Phyrexia: All Will Be One, oil counters and
// graveyard triggers (S82, batch 11).

const (
	b11Charforger      = "d6ded5b0-1309-4f1e-be10-d84ca0a5dd20"
	b11Fuseling        = "754e5238-00a2-45fa-9b20-40980b73f1cc"
	b11Centurion       = "fa519a46-11ab-4ab8-9286-904ce2746c03"
	b11Necrosquito     = "41dd7ec7-a238-4723-a244-9a112d4fbfd2"
	b11Adaptive        = "e350c2d5-0ca2-4b9c-b01f-a3c31cbf8d10"
	b11Scrapgorger     = "0d3a4db2-df79-4946-8f87-b174cf9c293c"
	b11Golem           = "2feca0cd-6bd3-4e19-8cfd-81d7bd1118ee"
	b11Anointer        = "a09c134b-7802-4526-9250-ec5cc637d9c0"
	b11Cackler         = "e1a41091-6ff1-42bb-b248-90040107efe1"
	b11Troll           = "238f2c86-4376-48f2-ae71-c10d3d785163"
	b11Ravager         = "a1c6bd47-19ef-44f9-921a-e4ffa80b25e1"
	b11Skitterfang     = "8255b408-cc8c-453f-982c-f361b5559cec"
	b11CreatureTypeBug = "Creature — Phyrexian Insect"
)

func TestS82OneB11CardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		b11Charforger: "Charforger", b11Fuseling: "Exuberant Fuseling", b11Centurion: "Forgehammer Centurion",
		b11Necrosquito: "Necrosquito", b11Adaptive: "Evolving Adaptive", b11Scrapgorger: "Armored Scrapgorger",
		b11Golem: "Ichorplate Golem", b11Anointer: "Urabrask's Anointer", b11Cackler: "Kuldotha Cackler",
		b11Troll: "Oil-Gorger Troll", b11Ravager: "Cinderslash Ravager", b11Skitterfang: "Atraxa's Skitterfang",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: registered=%v spec=%+v", name, ok, spec)
		}
	}
}

// b11Cast casts a creature with the given catalog identity from the
// active seat's hand in a main phase and resolves it.
func b11Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID}
	castFromHandAndResolve(t, g, me, c)
	drcSettle(t, g, me.ID)
	return c.InstanceID
}

func b11Oil(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return battlefieldCardCopy(t, g, id).Counters["oil"]
}

func b11PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	var p, tough int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				e := c.Effective()
				p, tough = e.Power, e.Toughness
			}
		}
	})
	return p, tough
}

func b11SetOil(g *game.Game, id uuid.UUID, n int) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				if g.Battlefield.Cards[i].Counters == nil {
					g.Battlefield.Cards[i].Counters = map[string]int{}
				}
				g.Battlefield.Cards[i].Counters["oil"] = n
			}
		}
	})
}

func b11OilRock(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Oily Rock", TypeLine: "Artifact", Owner: owner, Controller: owner,
		Counters: map[string]int{"oil": 1},
	})
}

func b11PickPlayer(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: who}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

func b11Die(t *testing.T, g *game.Game, owner, id uuid.UUID) {
	t.Helper()
	killCreature(t, g, owner, id)
	drcSettle(t, g, owner)
}

// --- Necrosquito ---------------------------------------------------

func TestS82NecrosquitoEntersWithTwoOilAndGrowsOnYourDepartures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	n := b11Cast(t, g, me, "Necrosquito", b11CreatureTypeBug, b11Necrosquito, 0, 0)
	if got := b11Oil(t, g, n); got != 2 {
		t.Fatalf("%d oil counters on entry, want 2", got)
	}
	if p, tough := b11PT(t, g, n); p != 2 || tough != 2 {
		t.Fatalf("%d/%d, want 2/2", p, tough)
	}
	b11Die(t, g, me.ID, b12Creature(g, me.ID, "Bear", drcCreatureTypeLine, 2, 2))
	b11Die(t, g, me.ID, b12Permanent(g, me.ID, "Rock", "Artifact"))
	if got := b11Oil(t, g, n); got != 4 {
		t.Fatalf("%d oil counters after a creature and an artifact, want 4", got)
	}
	// An opponent's creature and a noncreature nonartifact do not count.
	b11Die(t, g, opp.ID, b12Creature(g, opp.ID, "Bear", drcCreatureTypeLine, 2, 2))
	b11Die(t, g, me.ID, b12Permanent(g, me.ID, "Aura", "Enchantment"))
	if got := b11Oil(t, g, n); got != 4 {
		t.Fatalf("%d oil counters after excluded departures, want 4", got)
	}
	if p, tough := b11PT(t, g, n); p != 4 || tough != 4 {
		t.Fatalf("%d/%d, want 4/4", p, tough)
	}
}

// --- Exuberant Fuseling --------------------------------------------

func TestS82ExuberantFuselingCountsItsEntryAndDepartures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	f := b11Cast(t, g, me, "Exuberant Fuseling", "Creature — Phyrexian Goblin Warrior", b11Fuseling, 0, 1)
	if got := b11Oil(t, g, f); got != 1 {
		t.Fatalf("%d oil counters after entering, want 1", got)
	}
	if p, tough := b11PT(t, g, f); p != 1 || tough != 1 {
		t.Fatalf("%d/%d, want 1/1", p, tough)
	}
	b11Die(t, g, me.ID, b12Creature(g, me.ID, "Bear", drcCreatureTypeLine, 2, 2))
	if got := b11Oil(t, g, f); got != 2 {
		t.Fatalf("%d oil counters after a death, want 2", got)
	}
	if p, _ := b11PT(t, g, f); p != 2 {
		t.Fatalf("power %d, want 2", p)
	}
}

// --- Oil-Gorger Troll ----------------------------------------------

func TestS82OilGorgerTrollDrawsOnlyWithAnOilPermanent(t *testing.T) {
	for _, withOil := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		for i := 0; i < 3; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		}
		if withOil {
			b11OilRock(g, me.ID)
		}
		life, hand := me.Life, me.Hand.Size()
		b11Cast(t, g, me, "Oil-Gorger Troll", "Creature — Phyrexian Troll Warrior", b11Troll, 3, 4)
		if me.Life != life+3 {
			t.Errorf("withOil=%v: life %d, want %d", withOil, me.Life, life+3)
		}
		want := hand
		if withOil {
			want = hand + 1
		}
		if me.Hand.Size() != want {
			t.Errorf("withOil=%v: hand %d, want %d", withOil, me.Hand.Size(), want)
		}
	}
}

// --- Urabrask's Anointer -------------------------------------------

func TestS82UrabrasksAnointerDealsDamageEqualToYourOilPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b11OilRock(g, me.ID)
	b11OilRock(g, me.ID)
	b11OilRock(g, opp.ID) // an opponent's does not count
	life := opp.Life
	c := game.Card{InstanceID: uuid.New(), Name: "Urabrask's Anointer", TypeLine: "Artifact Creature — Phyrexian Wizard",
		OracleID: b11Anointer, Power: 4, Toughness: 2, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	b11PickPlayer(t, g, opp.ID)
	drcSettle(t, g, me.ID)
	if opp.Life != life-2 {
		t.Fatalf("life %d, want %d", opp.Life, life-2)
	}
}

// --- Kuldotha Cackler ----------------------------------------------

func TestS82KuldothaCacklerGetsPlusXWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := b12Push(g, me.ID, "Kuldotha Cackler", "Creature — Phyrexian Hyena", b11Cackler, 2, 3)
	b11OilRock(g, me.ID)
	b11OilRock(g, me.ID)
	b11OilRock(g, opp.ID)
	declareAttack(t, g, opp.ID, c)
	drcSettle(t, g, me.ID)
	if p, _ := b11PT(t, g, c); p != 4 {
		t.Fatalf("power %d, want 4", p)
	}
}

// --- Cinderslash Ravager -------------------------------------------

func TestS82CinderslashRavagerCostsLessPerOilPermanentAndPingsOpposingCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	card := func() game.Card {
		return game.Card{InstanceID: uuid.New(), Name: "Cinderslash Ravager", TypeLine: "Creature — Phyrexian Warrior",
			ManaCost: "{4}{R}{G}", Colors: []string{"R", "G"}, OracleID: b11Ravager, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID}
	}
	// Without oil it costs all six.
	full := card()
	me.Hand.PushTop(full)
	b06AddMana(me, "R", "G", "C", "C")
	if err := g.CastSpell(me.ID, full.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for four with no oil permanent")
	}
	b11OilRock(g, me.ID)
	b11OilRock(g, me.ID)
	b11OilRock(g, opp.ID) // an opponent's does not reduce it
	small := b12Creature(g, opp.ID, "Mite", drcCreatureTypeLine, 1, 1)
	big := b12Creature(g, opp.ID, "Ox", drcCreatureTypeLine, 2, 3)
	mine := b12Creature(g, me.ID, "Bear", drcCreatureTypeLine, 1, 1)
	if err := g.CastSpell(me.ID, full.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for four with two oil permanents: %v", err)
	}
	passPriorityAroundTable(t, g)
	drcSettle(t, g, me.ID)
	if battlefieldCardOK(g, small) {
		t.Error("the 1/1 survived the damage")
	}
	if !battlefieldCardOK(g, big) || !battlefieldCardOK(g, mine) {
		t.Error("the Ox or my own Bear died")
	}
	if c := battlefieldCardCopy(t, g, big); c.DamageMarked != 1 {
		t.Errorf("Ox has %d damage, want 1", c.DamageMarked)
	}
}

func battlefieldCardOK(g *game.Game, id uuid.UUID) bool {
	_, ok := battlefieldCard(g, id)
	return ok
}

// --- Armored Scrapgorger -------------------------------------------

func TestS82ArmoredScrapgorgerExilesAGraveyardCardWhenItBecomesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s := b12Push(g, me.ID, "Armored Scrapgorger", "Creature — Phyrexian Beast", b11Scrapgorger, 0, 3)
	dead := drcGraveyardCard(opp, "Dead Bear", drcCreatureTypeLine)
	advanceToMain(t, g)
	if p, _ := b11PT(t, g, s); p != 0 {
		t.Fatalf("power %d before any oil", p)
	}
	if err := g.ActivateManaAbility(me.ID, s, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	answerPickTarget(t, g, dead)
	if pick := riderLatestManaPick(g, me.ID); pick != nil {
		if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	drcSettle(t, g, me.ID)
	if got := b11Oil(t, g, s); got != 1 {
		t.Fatalf("%d oil counters, want 1", got)
	}
	for _, c := range opp.Graveyard.Cards {
		if c.InstanceID == dead {
			t.Fatal("the graveyard card was not exiled")
		}
	}
	if err := g.AddCounterForEffect(s, "oil", 2); err != nil {
		t.Fatal(err)
	}
	if p, _ := b11PT(t, g, s); p != 3 {
		t.Fatalf("power %d with three oil, want 3", p)
	}
	if err := g.AddCounterForEffect(s, "oil", -1); err != nil {
		t.Fatal(err)
	}
	if p, _ := b11PT(t, g, s); p != 0 {
		t.Fatalf("power %d with two oil, want 0", p)
	}
}

func TestS82ArmoredScrapgorgerGetsNoCounterWithNoGraveyardCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	s := b12Push(g, me.ID, "Armored Scrapgorger", "Creature — Phyrexian Beast", b11Scrapgorger, 0, 3)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, s, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if pick := riderLatestManaPick(g, me.ID); pick != nil {
		if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	drcSettle(t, g, me.ID)
	if got := b11Oil(t, g, s); got != 0 {
		t.Fatalf("%d oil counters with nothing to exile, want 0", got)
	}
}

// --- Charforger ----------------------------------------------------

func TestS82CharforgerMakesAGoblinCountsDeparturesAndPaysThreeOil(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	c := b11Cast(t, g, me, "Charforger", "Creature — Phyrexian Beast", b11Charforger, 2, 3)
	if n := cardsNamed(g, "Phyrexian Goblin", me.ID); n != 1 {
		t.Fatalf("%d Goblin tokens, want 1", n)
	}
	b11Die(t, g, me.ID, findBattlefieldByName(g, "Phyrexian Goblin"))
	if got := b11Oil(t, g, c); got != 1 {
		t.Fatalf("%d oil counters after the Goblin died, want 1", got)
	}

	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	b11SetOil(g, c, 2)
	if err := g.ActivateCatalogAbility(me.ID, c, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with two oil counters")
	}
	b11SetOil(g, c, 4)
	b16Activate(t, g, me.ID, c, 0, game.ActivateAbilityParams{})
	drcSettle(t, g, me.ID)
	if got := b11Oil(t, g, c); got != 1 {
		t.Fatalf("%d oil counters after paying three of four, want 1", got)
	}
	found := false
	for _, x := range g.Exile.Cards {
		if x.Name == "Top" {
			found = true
		}
	}
	if !found {
		t.Fatal("the top card was not exiled")
	}
}

// --- Evolving Adaptive ---------------------------------------------

func TestS82EvolvingAdaptiveEvolvesOnGreaterPowerOrToughnessOfYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b11Cast(t, g, me, "Evolving Adaptive", "Creature — Phyrexian Warrior", b11Adaptive, 0, 0)
	if got := b11Oil(t, g, a); got != 1 {
		t.Fatalf("%d oil on entry, want 1", got)
	}
	// 1/1 now. A 1/1 is not greater; a 2/1 has greater power.
	b11Cast(t, g, me, "Same", drcCreatureTypeLine, "", 1, 1)
	if got := b11Oil(t, g, a); got != 1 {
		t.Fatalf("%d oil after an equal creature, want 1", got)
	}
	b11Cast(t, g, me, "Power", drcCreatureTypeLine, "", 2, 1)
	if got := b11Oil(t, g, a); got != 2 {
		t.Fatalf("%d oil after a greater-power creature, want 2", got)
	}
	// 2/2 now. A 1/3 has greater toughness only.
	b11Cast(t, g, me, "Tough", drcCreatureTypeLine, "", 1, 3)
	if got := b11Oil(t, g, a); got != 3 {
		t.Fatalf("%d oil after a greater-toughness creature, want 3", got)
	}
	// An opponent's creature never evolves it.
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Theirs", TypeLine: drcCreatureTypeLine,
		Power: 9, Toughness: 9, Owner: opp.ID, Controller: opp.ID})
	drcSettle(t, g, me.ID)
	if got := b11Oil(t, g, a); got != 3 {
		t.Fatalf("%d oil after an opponent's creature, want 3", got)
	}
}

// --- Ichorplate Golem ----------------------------------------------

func TestS82IchorplateGolemGrowsEnteringOilCreaturesAndPumpsThem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	golem := b12Push(g, me.ID, "Ichorplate Golem", "Artifact Creature — Phyrexian Golem", b11Golem, 2, 3)
	bear := b12Creature(g, me.ID, "Bear", drcCreatureTypeLine, 2, 2)
	if p, _ := b11PT(t, g, bear); p != 2 {
		t.Fatalf("a Bear with no oil got the bonus: %d", p)
	}
	n := b11Cast(t, g, me, "Necrosquito", b11CreatureTypeBug, b11Necrosquito, 0, 0)
	if got := b11Oil(t, g, n); got != 3 {
		t.Fatalf("%d oil on the entering Necrosquito, want 3 (2 + the Golem's)", got)
	}
	// 0/0 + 3 oil, +1/+1 from the Golem.
	if p, tough := b11PT(t, g, n); p != 4 || tough != 4 {
		t.Fatalf("%d/%d, want 4/4", p, tough)
	}
	// A creature with no oil entering gets none.
	other := b11Cast(t, g, me, "Plain", drcCreatureTypeLine, "", 1, 1)
	if got := b11Oil(t, g, other); got != 0 {
		t.Fatalf("%d oil on a creature that entered with none", got)
	}
	_ = golem
}

// --- Forgehammer Centurion -----------------------------------------

func b11Centurion2(t *testing.T, g *game.Game, me *game.Player, oil int) uuid.UUID {
	t.Helper()
	c := b12Push(g, me.ID, "Forgehammer Centurion", "Creature — Phyrexian Warrior", b11Centurion, 3, 2)
	b11SetOil(g, c, oil)
	return c
}

func TestS82ForgehammerCenturionRemovesTwoOilToStopABlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := b11Centurion2(t, g, me, 3)
	blocker := b12Creature(g, opp.ID, "Wall", drcCreatureTypeLine, 0, 4)
	declareAttack(t, g, opp.ID, c)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	answerPickTarget(t, g, blocker)
	drcSettle(t, g, me.ID)
	if got := b11Oil(t, g, c); got != 1 {
		t.Fatalf("%d oil left, want 1", got)
	}
	if r := battlefieldCardCopy(t, g, blocker).Effective().Restrictions; r&game.CantBlock == 0 {
		t.Fatal("the target can still block")
	}
}

func TestS82ForgehammerCenturionDeclinedOrShortDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := b11Centurion2(t, g, me, 3)
	blocker := b12Creature(g, opp.ID, "Wall", drcCreatureTypeLine, 0, 4)
	declareAttack(t, g, opp.ID, c)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	drcSettle(t, g, me.ID)
	if got := b11Oil(t, g, c); got != 3 {
		t.Fatalf("declined, but %d oil left", got)
	}
	if r := battlefieldCardCopy(t, g, blocker).Effective().Restrictions; r&game.CantBlock != 0 {
		t.Fatal("declined, but the target can't block")
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	c2 := b11Centurion2(t, g2, me2, 1)
	declareAttack(t, g2, opp2.ID, c2)
	drcSettle(t, g2, me2.ID)
	if got := b11Oil(t, g2, c2); got != 1 {
		t.Fatalf("one oil can't pay for two, but %d left", got)
	}
}

// --- Atraxa's Skitterfang ------------------------------------------

func TestS82AtraxasSkitterfangGrantsYourChoiceOfKeywordForAnOilCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	f := b11Cast(t, g, me, "Atraxa's Skitterfang", "Artifact Creature — Phyrexian Insect", b11Skitterfang, 2, 2)
	if got := b11Oil(t, g, f); got != 3 {
		t.Fatalf("%d oil on entry, want 3", got)
	}
	bear := b12Creature(g, me.ID, "Bear", drcCreatureTypeLine, 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	for i := 0; i < 20 && latestOptionPickFor(g, me.ID) == nil; i++ {
		if len(g.PendingChoices) > 0 {
			answerPickTarget(t, g, bear)
		} else {
			passPriorityAroundTable(t, g)
		}
	}
	answerOptionPick(t, g, me.ID, 3) // lifelink
	drcSettle(t, g, me.ID)
	if got := b11Oil(t, g, f); got != 2 {
		t.Fatalf("%d oil left, want 2", got)
	}
	if !p1g2HasKeyword(g, bear, "lifelink") {
		t.Fatal("the Bear did not gain lifelink")
	}
	if p1g2HasKeyword(g, bear, "flying") {
		t.Fatal("the Bear gained a keyword that was not chosen")
	}
}
