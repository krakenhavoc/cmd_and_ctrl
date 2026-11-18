package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const satsukiOracle = "e41be9ad-3717-4f25-a8e0-c9947a2a0227"

func neoB19Rows() map[string]cards.Card {
	return map[string]cards.Card{
		"eiganjo": transformRow(theRestorationOfEiganjoOracleID, "The Restoration of Eiganjo", "Enchantment — Saga", "{2}{W}",
			"Architect of Restoration", "Enchantment Creature — Fox Monk", "3", "4", []string{"W"}),
		"dragonkami": transformRow(theDragonKamiRebornOracleID, "The Dragon-Kami Reborn", "Enchantment — Saga", "{2}{G}",
			"Dragon-Kami's Egg", "Enchantment Creature — Egg", "0", "1", []string{"G"}),
		"kumano": transformRow(kumanoFacesKakkazanOracleID, "Kumano Faces Kakkazan", "Enchantment — Saga", "{R}",
			"Etching of Kumano", "Enchantment Creature — Human Shaman", "2", "2", []string{"R"}),
		"jugan": transformRow(juganDefendsTheTempleOracleID, "Jugan Defends the Temple", "Enchantment — Saga", "{2}{G}",
			"Remnant of the Rising Star", "Enchantment Creature — Dragon Spirit", "2", "2", []string{"G"}),
		"iteration": transformRow(inventiveIterationOracleID, "Inventive Iteration", "Enchantment — Saga", "{3}{U}",
			"Living Breakthrough", "Enchantment Creature — Moonfolk", "3", "3", []string{"U"}),
		"kamiwar": transformRow(theKamiWarOracleID, "The Kami War", "Enchantment — Saga", "{1}{W}{U}{B}{R}{G}",
			"O-Kagachi Made Manifest", "Enchantment Creature — Dragon Spirit", "6", "6", []string{"W", "U", "B", "R", "G"}),
	}
}

func TestS82NeoB19EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		theRestorationOfEiganjoOracleID:        "The Restoration of Eiganjo",
		theRestorationOfEiganjoOracleID + "#1": "Architect of Restoration",
		theDragonKamiRebornOracleID:            "The Dragon-Kami Reborn",
		theDragonKamiRebornOracleID + "#1":     "Dragon-Kami's Egg",
		kumanoFacesKakkazanOracleID:            "Kumano Faces Kakkazan",
		kumanoFacesKakkazanOracleID + "#1":     "Etching of Kumano",
		juganDefendsTheTempleOracleID:          "Jugan Defends the Temple",
		juganDefendsTheTempleOracleID + "#1":   "Remnant of the Rising Star",
		inventiveIterationOracleID:             "Inventive Iteration",
		inventiveIterationOracleID + "#1":      "Living Breakthrough",
		theKamiWarOracleID:                     "The Kami War",
		theKamiWarOracleID + "#1":              "O-Kagachi Made Manifest",
		satsukiOracle:                          "Satsuki, the Living Lore",
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

// neoB19BackFace puts the back face of a row straight onto the battlefield
// under p, not summoning sick: what is under test is the back face's
// abilities, not how it got there (chapter III is proved per card below).
func neoB19BackFace(t *testing.T, g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	t.Helper()
	id := importToHand(row, p)
	g.WithWriteLock(func() {
		for i := range p.Hand.Cards {
			if p.Hand.Cards[i].InstanceID == id {
				p.Hand.Cards[i].SetFace(1)
			}
		}
	})
	var entered uuid.UUID
	var err error
	g.WithWriteLock(func() {
		entered, err = g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{})
	})
	if err != nil || entered == uuid.Nil {
		t.Fatalf("put the back face onto the battlefield: %v", err)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == entered {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	return entered
}

// neoB19ReachChapterThree casts the Saga and walks it to the end of chapter
// III, answering nothing: a test that has chapter I or II prompts handles
// them first and calls neoB19FinishChapters for the rest.
func neoB19AssertTransformed(t *testing.T, g *game.Game, saga uuid.UUID, back string, p *game.Player) *game.Card {
	t.Helper()
	if battlefieldCardFor(g, saga) != nil {
		t.Fatal("the original Saga is still on the battlefield")
	}
	if inGraveyardOf(g, p.ID, saga) {
		t.Fatal("the Saga was sacrificed instead of coming back transformed")
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == back {
			c := &g.Battlefield.Cards[i]
			if c.ActiveFace != 1 || c.Controller != p.ID || c.InstanceID == saga {
				t.Fatalf("%s came back as %+v", back, c)
			}
			assertFaceInvariant(t, g)
			return c
		}
	}
	t.Fatalf("%s never reached the battlefield", back)
	return nil
}

// --- The Restoration of Eiganjo // Architect of Restoration ----------

func TestRestorationOfEiganjoWalksItsChaptersAndTransforms(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	plains := pushLibraryTyped(me, "Plains", "Basic Land — Plains")
	pushLibraryTyped(me, "Forest", "Basic Land — Forest")
	// Tapped Cheap is a permanent card with mana value 2 in the graveyard;
	// Big Dead is mana value 3, and a Sorcery is not a permanent card.
	cheap := pushGraveyardCreature(g, me.ID, "Cheap Dead", "{1}{G}")
	big := pushGraveyardCreature(g, me.ID, "Big Dead", "{2}{G}")
	spell := pushGraveyardCardTyped(me, "Dead Spell", "Sorcery")

	saga := importAndCast(t, g, neoB19Rows()["eiganjo"], me)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(plains) {
		t.Fatal("chapter I did not put the basic Plains into hand")
	}

	// II — discard a card, then a reflexive trigger returns a permanent
	// card with mana value 2 or less, tapped.
	discard := pushHandCard(g, me)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("chapter II offered no discard")
	}
	answerDiscard(t, g, me.ID, discard)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the reflexive trigger asked for no target")
	}
	if hasID(p.PickTargetCards, big) || hasID(p.PickTargetCards, spell) || !hasID(p.PickTargetCards, cheap) {
		t.Errorf("targets %v: only the mana value 2 permanent card is legal", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if c, ok := g.LookupCardForEffect(cheap); !ok || !g.Battlefield.Contains(cheap) || !c.Tapped {
		t.Fatalf("the permanent card did not return tapped: %+v", c)
	}

	// III — exile and return transformed.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	back := neoB19AssertTransformed(t, g, saga, "Architect of Restoration", me)
	if back.Power != 3 || back.Toughness != 4 {
		t.Errorf("Architect is %d/%d, want 3/4", back.Power, back.Toughness)
	}
}

func TestRestorationOfEiganjoChapterTwoDeclinedReturnsNothing(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	cheap := pushGraveyardCreature(g, me.ID, "Cheap Dead", "{1}{G}")
	pushHandCard(g, me)
	importAndCast(t, g, neoB19Rows()["eiganjo"], me)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || g.Battlefield.Contains(cheap) {
		t.Fatal("declining the discard still returned a card")
	}
}

func TestArchitectOfRestorationMakesASpiritWhenItAttacksOrBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arch := neoB19BackFace(t, g, neoB19Rows()["eiganjo"], me)
	if !hasEffectiveKeyword(t, g, arch, "vigilance") {
		t.Fatal("no vigilance")
	}
	declareAttack(t, g, opp.ID, arch)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldByName(g, "Spirit"); got != 1 {
		t.Fatalf("Spirits after attacking = %d, want 1", got)
	}
}

func TestArchitectOfRestorationMakesASpiritWhenItBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arch := neoB19BackFace(t, g, neoB19Rows()["eiganjo"], opp)
	bear := pushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(arch, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldByName(g, "Spirit"); got != 1 {
		t.Fatalf("Spirits after blocking = %d, want 1", got)
	}
}

// --- Kumano Faces Kakkazan // Etching of Kumano ---------------------

func TestKumanoFacesKakkazanWalksItsChaptersAndTransforms(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp, opp2 := g.Seats[seat], g.Seats[(seat+1)%4], g.Seats[(seat+2)%4]
	walker := pushCatalogPermanent(g, opp.ID, "Their Walker", "Planeswalker — Test", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == walker {
				g.Battlefield.Cards[i].Counters = map[string]int{"loyalty": 3}
			}
		}
	})
	myWalker := pushCatalogPermanent(g, me.ID, "My Walker", "Planeswalker — Test", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == myWalker {
				g.Battlefield.Cards[i].Counters = map[string]int{"loyalty": 3}
			}
		}
	})
	lifeBefore := [3]int{me.Life, opp.Life, opp2.Life}

	saga := importAndCast(t, g, neoB19Rows()["kumano"], me)
	passPriorityAroundTable(t, g)
	if opp.Life != lifeBefore[1]-1 || opp2.Life != lifeBefore[2]-1 || me.Life != lifeBefore[0] {
		t.Errorf("life after chapter I: me %d, opp %d, opp2 %d", me.Life, opp.Life, opp2.Life)
	}
	if c, _ := g.LookupCardForEffect(walker); c.Counters["loyalty"] != 2 {
		t.Errorf("their planeswalker has loyalty %d, want 2", c.Counters["loyalty"])
	}
	if c, _ := g.LookupCardForEffect(myWalker); c.Counters["loyalty"] != 3 {
		t.Errorf("my planeswalker was damaged: loyalty %d", c.Counters["loyalty"])
	}

	// II — the next creature spell this turn enters with an extra counter.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	first := importToHand(cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: "First Bear", TypeLine: "Creature — Bear", ManaCost: "{0}", Power: "2", Toughness: "2"}, me)
	second := importToHand(cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: "Second Bear", TypeLine: "Creature — Bear", ManaCost: "{0}", Power: "2", Toughness: "2"}, me)
	for _, id := range []uuid.UUID{first, second} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if c, _ := g.LookupCardForEffect(first); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("the first creature has %d +1/+1 counters, want 1", c.Counters[game.CounterPlusOne])
	}
	if c, _ := g.LookupCardForEffect(second); c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("the second creature has %d +1/+1 counters, want none", c.Counters[game.CounterPlusOne])
	}

	// III
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	back := neoB19AssertTransformed(t, g, saga, "Etching of Kumano", me)
	if back.Power != 2 || back.Toughness != 2 {
		t.Errorf("Etching of Kumano is %d/%d, want 2/2", back.Power, back.Toughness)
	}
}

// The promise lapses at end of turn when no creature spell is cast.
func TestKumanoFacesKakkazanChapterTwoLapsesAtEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	importAndCast(t, g, neoB19Rows()["kumano"], me)
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, (seat+1)%4)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	bear := importToHand(cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: "Late Bear", TypeLine: "Creature — Bear", ManaCost: "{0}", Power: "2", Toughness: "2"}, me)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(bear); c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("a creature cast on a later turn got %d counters", c.Counters[game.CounterPlusOne])
	}
}

func TestEtchingOfKumanoExilesCreaturesDamagedByYourSources(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	etching := neoB19BackFace(t, g, neoB19Rows()["kumano"], me)
	if !hasEffectiveKeyword(t, g, etching, "haste") {
		t.Fatal("no haste")
	}
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	other := pushVanillaCreature(g, opp.ID, "Untouched Bear", 2, 2)
	untouched := pushVanillaCreature(g, opp.ID, "Damaged By Their Own", 2, 2)

	g.WithWriteLock(func() {
		// My source damages their bear; their source damages "Damaged By Their Own".
		if err := g.DealDamageToCreatureForEffect(mine, theirs, 2); err != nil {
			t.Fatalf("damage: %v", err)
		}
		if err := g.DealDamageToCreatureForEffect(other, untouched, 2); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	if inGraveyardOf(g, opp.ID, theirs) || !inExile(g, theirs) {
		t.Errorf("a creature dealt damage by my source did not go to exile instead of the graveyard")
	}
	if !inGraveyardOf(g, opp.ID, untouched) {
		t.Errorf("a creature dealt damage only by the opponent's source was exiled")
	}
	_ = other
}

// --- Satsuki, the Living Lore ----------------------------------------

func TestSatsukiPutsALoreCounterOnEachSagaYouControl(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga := importAndCast(t, g, neoB19Rows()["iteration"], me)
	passPriorityAroundTable(t, g)
	satsuki := pushCatalogPermanent(g, me.ID, "Satsuki, the Living Lore", "Legendary Creature — Human Druid", satsukiOracle, false)
	theirs := g.Seats[(seat+1)%4]
	oppSaga := pushCatalogPermanent(g, theirs.ID, "Their Saga", "Enchantment — Saga", "", false)
	if got := loreCountersOn(g, saga); got != 1 {
		t.Fatalf("lore counters before Satsuki = %d, want 1", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, satsuki, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := loreCountersOn(g, saga); got != 2 {
		t.Errorf("lore counters on my Saga = %d, want 2", got)
	}
	if got := loreCountersOn(g, oppSaga); got != 0 {
		t.Errorf("lore counters on their Saga = %d, want none", got)
	}
}

func TestSatsukiRefusesToActivateAtInstantSpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	satsuki := pushCatalogPermanent(g, me.ID, "Satsuki, the Living Lore", "Legendary Creature — Human Druid", satsukiOracle, false)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, satsuki, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("Satsuki activated outside a main phase")
	}
}

func TestSatsukiDiesReturnsASagaFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	satsuki := pushCatalogPermanent(g, me.ID, "Satsuki, the Living Lore", "Legendary Creature — Human Druid", satsukiOracle, false)
	saga := pushGraveyardCardTyped(me, "Dead Saga", "Enchantment — Saga")
	aura := pushGraveyardCardTyped(me, "Dead Aura", "Enchantment — Aura")
	killCreature(t, g, me.ID, satsuki)
	g.RunStateChecksForTest()
	c := modePickChoiceFor(g, me.ID)
	if c == nil || c.ModeMin != 0 {
		t.Fatalf("a real 'choose up to one' prompt, declinable: %+v", c)
	}
	answerMode(t, g, c, me.ID, 1)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if p == nil || hasID(p.PickTargetCards, aura) || !hasID(p.PickTargetCards, saga) {
		t.Fatalf("graveyard targets %+v: only the Saga card is legal", p)
	}
	pickCard(t, g, me.ID, saga)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(saga) {
		t.Error("the Saga card did not return to hand")
	}
}

func TestSatsukiDiesReturnsASagaOrEnchantmentCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	satsuki := pushCatalogPermanent(g, me.ID, "Satsuki, the Living Lore", "Legendary Creature — Human Druid", satsukiOracle, false)
	ec := pushCatalogPermanent(g, me.ID, "Kami", "Enchantment Creature — Spirit", "", false)
	plain := pushCatalogPermanent(g, me.ID, "Plain Bear", "Creature — Bear", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Their Kami", "Enchantment Creature — Spirit", "", false)
	killCreature(t, g, me.ID, satsuki)
	g.RunStateChecksForTest()
	answerMode(t, g, modePickChoiceFor(g, me.ID), me.ID, 0)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if p == nil || hasID(p.PickTargetCards, plain) || hasID(p.PickTargetCards, theirs) || !hasID(p.PickTargetCards, ec) {
		t.Fatalf("targets %+v: only my enchantment creature is legal", p)
	}
	pickCard(t, g, me.ID, ec)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ec) {
		t.Error("the enchantment creature did not return to hand")
	}
}

func TestSatsukiDiesMayChooseNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	satsuki := pushCatalogPermanent(g, me.ID, "Satsuki, the Living Lore", "Legendary Creature — Human Druid", satsukiOracle, false)
	saga := pushGraveyardCardTyped(me, "Dead Saga", "Enchantment — Saga")
	killCreature(t, g, me.ID, satsuki)
	g.RunStateChecksForTest()
	if c := modePickChoiceFor(g, me.ID); c != nil {
		if err := g.ResolveModePick(c.ID, me.ID, nil); err != nil {
			t.Fatalf("choose nothing: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(saga) {
		t.Error("choosing no mode still returned the Saga")
	}
}

// --- The Dragon-Kami Reborn // Dragon-Kami's Egg ---------------------

// neoB19HatchTable casts The Dragon-Kami Reborn with three creature cards on
// top of the library and runs chapter I, exiling `pick` (an index into
// want, top card first). It returns the Saga and the three cards.
func neoB19HatchTable(t *testing.T, g *game.Game, me *game.Player, pick int) (saga uuid.UUID, top []uuid.UUID) {
	t.Helper()
	c := pushLibraryTyped(me, "Third", "Creature — Bear")
	b := pushLibraryTyped(me, "Second", "Creature — Bear")
	a := pushLibraryTyped(me, "First", "Creature — Bear")
	top = []uuid.UUID{a, b, c}
	saga = importAndCast(t, g, neoB19Rows()["dragonkami"], me)
	passPriorityAroundTable(t, g)
	prompt := chooseCardsChoiceFor(g, me.ID)
	if prompt == nil || prompt.ChooseMin != 1 || prompt.ChooseMax != 1 || len(prompt.ChooseCards) != 3 {
		t.Fatalf("want a pick of exactly one of the top three, got %+v", prompt)
	}
	answerChooseCards(t, g, me.ID, top[pick])
	if order := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID); order != nil {
		var rest []uuid.UUID
		for i, id := range top {
			if i != pick {
				rest = append(rest, id)
			}
		}
		if err := g.ResolvePutInLibrary(order.ID, me.ID, rest, nil); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	return saga, top
}

func TestDragonKamiRebornChapterOneGainsLifeAndExilesOneFaceDownWithAHatchingCounter(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	life := me.Life
	_, top := neoB19HatchTable(t, g, me, 1)
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d", me.Life, life+2)
	}
	exiled, ok := g.LookupCardForEffect(top[1])
	if !ok || !inExile(g, top[1]) {
		t.Fatal("the chosen card is not in exile")
	}
	if exiled.Counters[counterHatching] != 1 {
		t.Errorf("hatching counters = %d, want 1", exiled.Counters[counterHatching])
	}
	if exiled.FaceDownKind == "" {
		t.Error("the card was exiled face up")
	}
	for _, id := range []uuid.UUID{top[0], top[2]} {
		if !me.Library.Contains(id) {
			t.Errorf("card %s is not back in the library", id)
		}
	}
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID == top[0] || me.Library.Cards[len(me.Library.Cards)-1].InstanceID == top[2] {
		t.Error("the rest were not put on the bottom")
	}
}

func TestDragonKamiRebornChapterTwoDoesItAgainAndChapterThreeTransforms(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga, _ := neoB19HatchTable(t, g, me, 0)
	more := pushLibraryTyped(me, "Fourth", "Creature — Bear")
	pushLibraryTyped(me, "Fifth", "Creature — Bear")
	pushLibraryTyped(me, "Sixth", "Creature — Bear")
	life := me.Life
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	prompt := chooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("chapter II asked for no pick")
	}
	answerChooseCards(t, g, me.ID, prompt.ChooseCards[0])
	if order := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID); order != nil {
		if err := g.ResolvePutInLibrary(order.ID, me.ID, order.ScryCards, nil); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("chapter II life = %d, want %d", me.Life, life+2)
	}
	hatching := 0
	for _, c := range g.Exile.Cards {
		if c.Counters[counterHatching] > 0 {
			hatching++
		}
	}
	if hatching != 2 {
		t.Errorf("%d exiled cards carry hatching counters, want 2", hatching)
	}
	_ = more
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	back := neoB19AssertTransformed(t, g, saga, "Dragon-Kami's Egg", me)
	if back.Power != 0 || back.Toughness != 1 {
		t.Errorf("Dragon-Kami's Egg is %d/%d, want 0/1", back.Power, back.Toughness)
	}
}

func TestDragonKamisEggCastsAHatchingCreatureFreeWhenADragonDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	egg := neoB19BackFace(t, g, neoB19Rows()["dragonkami"], me)
	dragon := pushCatalogPermanent(g, me.ID, "Test Dragon", "Creature — Dragon", "", false)
	bear := importToHand(cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: "Hatchling", TypeLine: "Creature — Bear", ManaCost: "{4}{G}{G}", Power: "2", Toughness: "2"}, me)
	spell := importToHand(cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: "Hatching Spell", TypeLine: "Sorcery", ManaCost: "{4}", Power: "", Toughness: ""}, me)
	for _, id := range []uuid.UUID{bear, spell} {
		g.WithWriteLock(func() {
			if err := g.ExileCardForEffect(id); err != nil {
				t.Fatalf("exile: %v", err)
			}
		})
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, counterHatching, 1) })
	}
	killCreature(t, g, me.ID, dragon)
	g.RunStateChecksForTest()
	ask := awaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, bear)
	if slicesContainsID(offer.DiscardOptions, spell) {
		t.Error("a noncreature card with a hatching counter is offered")
	}
	if err := castThroughResolutionCast(t, g, me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the hatchling for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the hatchling did not resolve")
	}
	_ = egg
}

func TestDragonKamisEggIgnoresANonDragonDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	neoB19BackFace(t, g, neoB19Rows()["dragonkami"], me)
	bear := pushCatalogPermanent(g, me.ID, "Plain Bear", "Creature — Bear", "", false)
	killCreature(t, g, me.ID, bear)
	g.RunStateChecksForTest()
	if latestTriggerPrompt(g, me.ID) != nil || !stackFullyEmpty(g) {
		t.Error("a non-Dragon dying triggered the Egg")
	}
}

func TestDragonKamisEggTriggersOnItsOwnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	egg := neoB19BackFace(t, g, neoB19Rows()["dragonkami"], me)
	killCreature(t, g, me.ID, egg)
	g.RunStateChecksForTest()
	// With nothing exiled the trigger has nothing to offer, but it did
	// trigger: it asked the optional question.
	if latestTriggerPrompt(g, me.ID) == nil && stackFullyEmpty(g) {
		t.Error("the Egg's own death did not trigger it")
	}
}

// --- Jugan Defends the Temple // Remnant of the Rising Star ----------

func TestJuganDefendsTheTempleWalksItsChaptersAndTransforms(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	other := pushVanillaCreature(g, me.ID, "Other Bear", 2, 2)
	third := pushVanillaCreature(g, g.Seats[1].ID, "Third Bear", 2, 2)

	saga := importAndCast(t, g, neoB19Rows()["jugan"], me)
	passPriorityAroundTable(t, g)
	var monk *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Human Monk" {
			monk = &g.Battlefield.Cards[i]
		}
	}
	if monk == nil || !monk.IsToken() || monk.Power != 1 || monk.Toughness != 1 || len(monk.Colors) != 1 || monk.Colors[0] != "G" || !monk.HasSubtype("Monk") || !monk.HasSubtype("Human") {
		t.Fatalf("Human Monk = %+v", monk)
	}
	monkID := monk.InstanceID
	// Its mana ability: {T}: Add {G}. (It is summoning sick, which the {T}
	// cost respects, so clear that for the check.)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == monkID {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	if err := g.ActivateManaAbility(me.ID, monkID, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the Monk's mana ability: %v", err)
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("pool has %d mana, want {G}", got)
	}

	// II — a +1/+1 counter on each of up to two target creatures.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil || p.PickTargetMax != 2 || p.PickTargetMin != 0 {
		t.Fatalf("chapter II target pick %+v, want up to two", p)
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: third}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, third} {
		if counterCount(g, id, game.CounterPlusOne) != 1 {
			t.Errorf("creature %s lacks its counter", id)
		}
	}
	if counterCount(g, other, game.CounterPlusOne) != 0 {
		t.Error("an untargeted creature got a counter")
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	back := neoB19AssertTransformed(t, g, saga, "Remnant of the Rising Star", me)
	if back.Power != 2 || back.Toughness != 2 {
		t.Errorf("Remnant is %d/%d, want 2/2", back.Power, back.Toughness)
	}
}

func TestRemnantOfTheRisingStarPaysXForXCountersOnTheCreatureThatEntered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	remnant := neoB19BackFace(t, g, neoB19Rows()["jugan"], me)
	if !hasEffectiveKeyword(t, g, remnant, "flying") {
		t.Fatal("no flying")
	}
	fillPool(me, 3)
	bear := hobEnter(t, g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	pa := answerNumber(t, g, me.ID, 2)
	if pa.Min != 0 || pa.Max != 3 {
		t.Errorf("X prompt %d..%d, want 0..3", pa.Min, pa.Max)
	}
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the Bear has %d +1/+1 counters, want 2", got)
	}
	if got := counterCount(g, remnant, game.CounterPlusOne); got != 0 {
		t.Errorf("the Remnant got %d counters", got)
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("%d mana left, want 1", got)
	}
}

func TestRemnantOfTheRisingStarPayingNothingAndOpponentCreaturesDoNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	neoB19BackFace(t, g, neoB19Rows()["jugan"], me)
	fillPool(me, 3)
	hobEnter(t, g, opp.ID, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	if payAmountPrompt(g, me.ID) != nil {
		t.Fatal("an opponent's creature entering asked for a payment")
	}
	bear := hobEnter(t, g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	answerNumber(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 0 || me.ManaPool.Total() != 3 {
		t.Errorf("X=0 left %d counters and %d mana", got, me.ManaPool.Total())
	}
}

func TestRemnantOfTheRisingStarGetsFivePowerAndTrampleWithFiveModifiedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	remnant := neoB19BackFace(t, g, neoB19Rows()["jugan"], me)
	base := apaLive(g, remnant).CurrentPower()
	var modified []uuid.UUID
	for i := 0; i < 4; i++ {
		id := pushVanillaCreature(g, me.ID, "Modified Bear", 2, 2)
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterPlusOne, 1) })
		modified = append(modified, id)
	}
	// An opponent's modified creature does not count, and neither does an
	// unmodified one of mine.
	theirs := pushVanillaCreature(g, opp.ID, "Their Modified Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 1) })
	pushVanillaCreature(g, me.ID, "Plain Bear", 2, 2)
	if p := apaLive(g, remnant).CurrentPower(); p != base {
		t.Fatalf("power %d with four modified creatures, want %d", p, base)
	}
	if hasEffectiveKeyword(t, g, remnant, "trample") {
		t.Fatal("trample with four modified creatures")
	}
	// The fifth: the Remnant itself, once it carries a counter.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(remnant, game.CounterPlusOne, 1) })
	live := apaLive(g, remnant)
	p, tough := live.CurrentPower(), live.CurrentToughness()
	if p != base+1+5 || tough < 5 {
		t.Errorf("Remnant is %d/%d, want +1/+1 counter and +5/+5", p, tough)
	}
	if !hasEffectiveKeyword(t, g, remnant, "trample") {
		t.Error("no trample with five modified creatures")
	}
	// Losing one drops both.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(modified[0], game.CounterPlusOne, -1) })
	if p := apaLive(g, remnant).CurrentPower(); p != base+1 {
		t.Errorf("power %d after dropping to four modified creatures, want %d", p, base+1)
	}
	if hasEffectiveKeyword(t, g, remnant, "trample") {
		t.Error("trample stayed with four modified creatures")
	}
}

// --- Inventive Iteration // Living Breakthrough ----------------------

func TestInventiveIterationWalksItsChaptersAndTransforms(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	rock := pushGraveyardCardTyped(me, "Dead Rock", "Artifact")
	lib := len(me.Library.Cards)
	_ = lib

	saga := importAndCast(t, g, neoB19Rows()["iteration"], me)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !opp.Hand.Contains(victim) {
		t.Fatal("chapter I did not return the creature to its owner's hand")
	}

	// II — return an artifact card from your graveyard; no draw then.
	advanceToPrecombatMainOf(t, g, seat)
	hand := me.Hand.Size() // after the draw step
	passPriorityAroundTable(t, g)
	prompt := chooseCardsChoiceFor(g, me.ID)
	if prompt == nil || !hasID(prompt.ChooseCards, rock) {
		t.Fatalf("chapter II prompt %+v does not offer the artifact card", prompt)
	}
	answerChooseCards(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) || me.Hand.Size() != hand+1 {
		t.Errorf("hand grew by %d, want just the artifact card", me.Hand.Size()-hand)
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	back := neoB19AssertTransformed(t, g, saga, "Living Breakthrough", me)
	if back.Power != 3 || back.Toughness != 3 {
		t.Errorf("Living Breakthrough is %d/%d, want 3/3", back.Power, back.Toughness)
	}
}

func TestInventiveIterationChapterOneTakesUpToOnePlaneswalkerOrNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := pushCatalogPermanent(g, opp.ID, "Their Walker", "Planeswalker — Test", "", false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(walker, "loyalty", 3) })
	land := pushCatalogPermanent(g, opp.ID, "Their Land", "Land", "", false)
	importAndCast(t, g, neoB19Rows()["iteration"], me)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil || p.PickTargetMin != 0 || p.PickTargetMax != 1 || !hasID(p.PickTargetCards, walker) || hasID(p.PickTargetCards, land) {
		t.Fatalf("chapter I pick %+v: up to one creature or planeswalker", p)
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
		t.Fatalf("choosing no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(walker) {
		t.Error("choosing no target still bounced something")
	}
}

func TestInventiveIterationChapterTwoDrawsWhenThereIsNoArtifactCard(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	pushLibraryTyped(me, "Top", "Creature — Bear")
	importAndCast(t, g, neoB19Rows()["iteration"], me)
	passPriorityAroundTable(t, g)
	for latestPickTarget(g, me.ID) != nil {
		if err := g.ResolvePickTargets(latestPickTarget(g, me.ID).ID, me.ID, nil); err != nil {
			t.Fatalf("no target: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	advanceToPrecombatMainOf(t, g, seat)
	hand := me.Hand.Size() // after the draw step
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Fatalf("hand grew by %d: chapter II should draw exactly one card", got)
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a pick was offered with no artifact card in the graveyard")
	}
}

func neoB19Spell(p *game.Player, name, cost string) uuid.UUID {
	return importToHand(cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: name, TypeLine: "Instant", ManaCost: cost}, p)
}

func TestLivingBreakthroughBansTheCastSpellsManaValueUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	breakthrough := neoB19BackFace(t, g, neoB19Rows()["iteration"], me)
	if !hasEffectiveKeyword(t, g, breakthrough, "flying") {
		t.Fatal("no flying")
	}
	fillPool(me, 2)
	if err := g.CastSpell(me.ID, neoB19Spell(me, "My Spell", "{2}"), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	for _, o := range []*game.Player{opp, opp2} {
		fillPool(o, 5)
		if err := g.CastSpell(o.ID, neoB19Spell(o, "Same MV", "{2}"), game.CastSpellParams{Strict: true}); !errors.Is(err, game.ErrCantCast) {
			t.Fatalf("%s cast a mana value 2 spell: %v, want ErrCantCast", o.ID, err)
		}
	}
	if err := g.CastSpell(opp.ID, neoB19Spell(opp, "Other MV", "{3}"), game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a mana value 3 spell is not banned: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The ban does not bind its own controller.
	fillPool(me, 2)
	if err := g.CastSpell(me.ID, neoB19Spell(me, "Same MV Again", "{2}"), game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the controller is not banned: %v", err)
	}
	passPriorityAroundTable(t, g)

	// Until your next turn: it ends as that turn begins.
	advanceToUpkeepOf(t, g, 0)
	fillPool(opp, 5)
	if err := g.CastSpell(opp.ID, neoB19Spell(opp, "Same MV Later", "{2}"), game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the ban should be over once the next turn has begun: %v", err)
	}
}

func TestLivingBreakthroughBansEachCastSpellsValueInTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	neoB19BackFace(t, g, neoB19Rows()["iteration"], me)
	fillPool(me, 1)
	if err := g.CastSpell(me.ID, neoB19Spell(me, "One", "{1}"), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	fillPool(me, 4)
	if err := g.CastSpell(me.ID, neoB19Spell(me, "Four", "{4}"), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	fillPool(opp, 9)
	for _, cost := range []string{"{1}", "{4}"} {
		if err := g.CastSpell(opp.ID, neoB19Spell(opp, "Banned "+cost, cost), game.CastSpellParams{Strict: true}); !errors.Is(err, game.ErrCantCast) {
			t.Errorf("a %s spell: %v, want ErrCantCast", cost, err)
		}
	}
	if err := g.CastSpell(opp.ID, neoB19Spell(opp, "Free", "{0}"), game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a mana value 0 spell is not banned: %v", err)
	}
}

// --- The Kami War // O-Kagachi Made Manifest -------------------------

func TestTheKamiWarWalksItsChaptersAndTransforms(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	theirs := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	land := pushCatalogPermanent(g, opp.ID, "Their Land", "Land", "", false)
	mine := pushCatalogPermanent(g, me.ID, "My Rock", "Artifact", "", false)
	pushHandCard(g, opp)
	pushHandCard(g, g.Seats[(seat+2)%4])

	saga := importAndCast(t, g, neoB19Rows()["kamiwar"], me)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil || hasID(p.PickTargetCards, land) || hasID(p.PickTargetCards, mine) || !hasID(p.PickTargetCards, theirs) {
		t.Fatalf("chapter I targets %+v: only the opponent's nonland permanent", p)
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !inExile(g, theirs) {
		t.Fatal("chapter I did not exile the target")
	}

	// II — bounce up to one other target nonland permanent, then each
	// opponent discards a card.
	advanceToPrecombatMainOf(t, g, seat)
	// Counted after the round of draw steps the advance walks through.
	hands := []int{g.Seats[(seat+1)%4].Hand.Size(), g.Seats[(seat+2)%4].Hand.Size(), g.Seats[(seat+3)%4].Hand.Size()}
	passPriorityAroundTable(t, g)
	p = latestPickTarget(g, me.ID)
	if p == nil || p.PickTargetMin != 0 || p.PickTargetMax != 1 || hasID(p.PickTargetCards, saga) || hasID(p.PickTargetCards, land) || !hasID(p.PickTargetCards, mine) {
		t.Fatalf("chapter II targets %+v: up to one other nonland permanent", p)
	}
	pickCard(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	for i := 1; i <= 3; i++ {
		o := g.Seats[(seat+i)%4]
		for discardChoiceFor(g, o.ID) != nil {
			answerDiscard(t, g, o.ID, o.Hand.Cards[0].InstanceID)
			passPriorityAroundTable(t, g)
		}
		if got := o.Hand.Size(); got != hands[i-1]-1 {
			t.Errorf("opponent %d hand %d -> %d, want one fewer", i, hands[i-1], got)
		}
	}
	if !me.Hand.Contains(mine) {
		t.Error("chapter II did not return my permanent to my hand")
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	back := neoB19AssertTransformed(t, g, saga, "O-Kagachi Made Manifest", me)
	if back.Power != 6 || back.Toughness != 6 || len(back.Colors) != 5 {
		t.Errorf("O-Kagachi is %d/%d with colors %v, want 6/6 in all five", back.Power, back.Toughness, back.Colors)
	}
}

func TestOKagachiChoosesTheDefendingPlayerAndPumpsByManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := neoB19BackFace(t, g, neoB19Rows()["kamiwar"], me)
	if !hasEffectiveKeyword(t, g, dragon, "flying") || !hasEffectiveKeyword(t, g, dragon, "trample") {
		t.Fatal("missing flying or trample")
	}
	land := pushGraveyardCardTyped(me, "Dead Land", "Land")
	four := pushGraveyardCreature(g, me.ID, "Four Drop", "{2}{G}{G}")
	one := pushGraveyardCreature(g, me.ID, "One Drop", "{G}")
	before := apaLive(g, dragon).CurrentPower()

	declareAttack(t, g, opp.ID, dragon)
	passPriorityAroundTable(t, g)
	prompt := chooseCardsChoiceFor(g, opp.ID)
	if prompt == nil {
		t.Fatalf("the defending player was not asked: %+v", g.PendingChoices)
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the attacker was asked; the defending player chooses")
	}
	if hasID(prompt.ChooseCards, land) || !hasID(prompt.ChooseCards, four) || !hasID(prompt.ChooseCards, one) {
		t.Fatalf("candidates %v: nonland cards in the attacker's graveyard", prompt.ChooseCards)
	}
	answerChooseCards(t, g, opp.ID, four)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(four) {
		t.Error("the chosen card did not return to the attacker's hand")
	}
	after := apaLive(g, dragon).CurrentPower()
	if after != before+4 {
		t.Errorf("power %d -> %d, want +4 (mana value of the chosen card)", before, after)
	}
}

func TestOKagachiWithNothingToChooseDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := neoB19BackFace(t, g, neoB19Rows()["kamiwar"], me)
	pushGraveyardCardTyped(me, "Dead Land", "Land")
	before := apaLive(g, dragon).CurrentPower()
	declareAttack(t, g, opp.ID, dragon)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, opp.ID) != nil {
		t.Fatal("asked to choose with only a land card in the graveyard")
	}
	if after := apaLive(g, dragon).CurrentPower(); after != before {
		t.Errorf("power changed %d -> %d", before, after)
	}
}

func TestOKagachiIsAllColors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dragon := neoB19BackFace(t, g, neoB19Rows()["kamiwar"], me)
	c := apaLive(g, dragon)
	if len(c.Colors) != 5 {
		t.Errorf("colors %v, want all five", c.Colors)
	}
}
