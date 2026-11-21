package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b39_test.go — Assassin's Creed cards of S82 GAP batch 39 (#5948).
// Each test plays the real catalog entry.

const (
	gapb39AyaOracle       = "27b690f2-d7ce-43b6-8e8b-98a8ffe487bb"
	gapb39BayekOracle     = "b0322e2e-e9f9-4fe2-a7a5-3e49665ad177"
	gapb39EivorOracle     = "1b4a07bd-5345-4520-bd05-b5173e08135c"
	gapb39EzioOracle      = "d4e2b541-31b6-4bcb-a28a-d897deb91484"
	gapb39KassandraOracle = "745b9a2c-bcd7-4db9-8cd0-d38be2ef6270"
	gapb39SpyOracle       = "26d800b2-b312-4400-8357-615301971086"
	gapb39EvieOracle      = "248cafe8-cc37-4cf3-8239-ac10e55ced13"
	gapb39LeonardoOracle  = "92a287a0-b3cc-4040-a530-fb71510cfc67"
	gapb39InventorOracle  = "eff22b28-88b8-4251-937f-4c06b6ca7bd2"
	gapb39HaythamOracle   = "ee6a9bac-cf5b-4edd-8cb2-874d8dd0e62a"
	gapb39SokratesOracle  = "f4a9d567-513b-487e-afed-c6e6513a7862"
)

func TestGapB39CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapb39AyaOracle:       "Aya of Alexandria",
		gapb39BayekOracle:     "Bayek of Siwa",
		gapb39EivorOracle:     "Eivor, Battle-Ready",
		gapb39EzioOracle:      "Ezio, Brash Novice",
		gapb39KassandraOracle: "Kassandra, Eagle Bearer",
		gapb39SpyOracle:       "Brotherhood Spy",
		gapb39EvieOracle:      "Evie Frye",
		gapb39LeonardoOracle:  "Leonardo da Vinci",
		gapb39InventorOracle:  "Loyal Inventor",
		gapb39HaythamOracle:   "Haytham Kenway",
		gapb39SokratesOracle:  "Sokrates, Athenian Teacher",
	} {
		def, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

func gapb39Hand(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

func gapb39Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
}

func gapb39Assassins(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Assassin" && c.IsToken() {
			n++
		}
	}
	return n
}

// --- Aya of Alexandria --------------------------------------------------------

func TestGapB39AyaMakesAnAssassinForEachHistoricCreatureThatConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aya := b12Push(g, me.ID, "Aya of Alexandria", "Legendary Creature — Human Assassin", gapb39AyaOracle, 4, 3)
	for _, kw := range []string{"menace", "lifelink"} {
		if !effectiveAbilitiesContain(t, g, aya, kw) {
			t.Errorf("Aya has no %s", kw)
		}
	}
	relic := b12Creature(g, me.ID, "Relic Golem", "Artifact Creature — Golem", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	attackWith(t, g, opp.ID, aya, relic, bear)
	passPriorityAroundTable(t, g)
	if got := gapb39Assassins(g, me.ID); got != 2 {
		t.Errorf("%d Assassin tokens, want 2 (Aya and the artifact creature; not the Bear)", got)
	}
}

// --- Eivor, Battle-Ready ------------------------------------------------------

func TestGapB39EivorDealsDamageForEachEquipmentToEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eivor := b12Push(g, me.ID, "Eivor, Battle-Ready", "Legendary Creature — Human Assassin Warrior", gapb39EivorOracle, 5, 5)
	for _, kw := range []string{"vigilance", "haste"} {
		if !effectiveAbilitiesContain(t, g, eivor, kw) {
			t.Errorf("Eivor has no %s", kw)
		}
	}
	b12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	b12Permanent(g, me.ID, "Axe", "Artifact — Equipment")
	b12Permanent(g, opp.ID, "Their Sword", "Artifact — Equipment")
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	attackWith(t, g, opp.ID, eivor)
	passPriorityAroundTable(t, g)
	for i := 1; i < len(g.Seats); i++ {
		want := lives[i] - 2
		if g.Seats[i].ID == opp.ID {
			want -= 5
		}
		if g.Seats[i].Life != want {
			t.Errorf("opponent %d life %d -> %d, want %d", i, lives[i], g.Seats[i].Life, want)
		}
	}
}

// --- Brotherhood Spy ----------------------------------------------------------

func TestGapB39BrotherhoodSpyNeedsALegendaryAssassin(t *testing.T) {
	for _, legend := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		spy := b12Push(g, me.ID, "Brotherhood Spy", "Creature — Human Assassin", gapb39SpyOracle, 1, 3)
		typeLine := "Creature — Human Assassin"
		if legend {
			typeLine = "Legendary Creature — Human Assassin"
		}
		b12Creature(g, me.ID, "Friend", typeLine, 2, 2)
		advanceTo(t, g, game.StepBeginCombat)
		passPriorityAroundTable(t, g)
		power := effectivePower(t, g, spy)
		if legend {
			assertRestrictions(t, g, spy, game.CantBeBlocked)
			if power != 2 {
				t.Errorf("with a legendary Assassin the Spy's power is %d, want 2", power)
			}
			continue
		}
		assertRestrictions(t, g, spy, 0)
		if power != 1 {
			t.Errorf("without one the Spy's power is %d, want 1", power)
		}
	}
}

// --- Ezio, Brash Novice -------------------------------------------------------

func TestGapB39EzioBecomesAnAssassinWithFirstStrikeAtTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ezio := b12Push(g, me.ID, "Ezio, Brash Novice", "Legendary Creature — Human", gapb39EzioOracle, 1, 1)
	if effectiveAbilitiesContain(t, g, ezio, "first strike") || hasFold(effectiveSubtypes(t, g, ezio), "Assassin") {
		t.Fatal("Ezio starts with first strike or as an Assassin")
	}
	attackWith(t, g, opp.ID, ezio)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, ezio, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d counters after one attack, want 1", got)
	}
	if effectiveAbilitiesContain(t, g, ezio, "first strike") {
		t.Error("first strike with one counter")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ezio, "charge", 1) })
	if !effectiveAbilitiesContain(t, g, ezio, "first strike") {
		t.Error("no first strike with two counters (of any kind)")
	}
	subs := effectiveSubtypes(t, g, ezio)
	if !hasFold(subs, "Assassin") || !hasFold(subs, "Human") {
		t.Errorf("subtypes %v, want Human and Assassin", subs)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ezio, "charge", -1) })
	if effectiveAbilitiesContain(t, g, ezio, "first strike") {
		t.Error("first strike stayed after a counter was removed")
	}
}

// --- Bayek of Siwa ------------------------------------------------------------

func TestGapB39BayekGrantsDoubleStrikeToOtherHistoricCreaturesOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bayek := b12Push(g, me.ID, "Bayek of Siwa", "Legendary Creature — Human Assassin", gapb39BayekOracle, 3, 4)
	if !effectiveAbilitiesContain(t, g, bayek, "double strike") {
		t.Error("Bayek has no double strike")
	}
	golem := b12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 2, 2)
	if !effectiveAbilitiesContain(t, g, golem, "double strike") {
		t.Error("my historic creature has no double strike on my turn")
	}
	if effectiveAbilitiesContain(t, g, bear, "double strike") || effectiveAbilitiesContain(t, g, theirs, "double strike") {
		t.Error("a nonhistoric or opposing creature got double strike")
	}
	dftMPPassTurn(t, g)
	if effectiveAbilitiesContain(t, g, golem, "double strike") {
		t.Error("double strike on an opponent's turn")
	}
	def, _ := Lookup(gapb39BayekOracle)
	if len(def.AlternativeCosts) != 1 || def.AlternativeCosts[0].Key != "disguise" {
		t.Errorf("Bayek's alternative costs %+v, want disguise", def.AlternativeCosts)
	}
}

// --- Kassandra, Eagle Bearer --------------------------------------------------

func TestGapB39KassandraFetchesTheSpearAndDrawsForALegendaryEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spear := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: spear, Name: "The Spear of Leonidas", TypeLine: "Legendary Artifact — Equipment", Owner: me.ID, Controller: me.ID})
	castHoldCreature(t, g, "Kassandra, Eagle Bearer", "Legendary Creature — Human Assassin Warrior", gapb39KassandraOracle, 2, 2)
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "The Spear of Leonidas")
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(spear) {
		t.Fatal("The Spear of Leonidas was not put onto the battlefield")
	}
	for _, legend := range []bool{true, false} {
		g2 := newCatalogGame(t)
		me2, opp2 := g2.Seats[0], g2.Seats[1]
		b12Push(g2, me2.ID, "Kassandra, Eagle Bearer", "Legendary Creature — Human Assassin Warrior", gapb39KassandraOracle, 2, 2)
		bearer := b12Creature(g2, me2.ID, "Bearer", "Creature — Human", 2, 2)
		typeLine := "Artifact — Equipment"
		if legend {
			typeLine = "Legendary Artifact — Equipment"
		}
		pushBattlefieldCardWithTimestamp(g2, game.Card{
			InstanceID: uuid.New(), Name: "Blade", TypeLine: typeLine, Owner: me2.ID, Controller: me2.ID,
			AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bearer},
		})
		hand := me2.Hand.Size()
		attackWith(t, g2, opp2.ID, bearer)
		passPriorityAroundTable(t, g2)
		want := hand
		if legend {
			want = hand + 1
		}
		if got := me2.Hand.Size(); got != want {
			t.Errorf("legendary=%v: hand %d -> %d, want %d", legend, hand, got, want)
		}
	}
	_ = opp
}

// --- Loyal Inventor -----------------------------------------------------------

func TestGapB39LoyalInventorPutsTheArtifactInHandOrOnTop(t *testing.T) {
	for _, assassin := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		if assassin {
			b12Creature(g, me.ID, "Friend", "Creature — Human Assassin", 1, 1)
		}
		relic := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: relic, Name: "Gadget", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
		for i := 0; i < 3; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		}
		castHoldCreature(t, g, "Loyal Inventor", "Creature — Human Artificer", gapb39InventorOracle, 2, 3)
		passPriorityAroundTable(t, g)
		answerSearchNamed(t, g, me.ID, "Gadget")
		passPriorityAroundTable(t, g)
		z := g.FindCardZoneForEffect(relic)
		if z == nil {
			t.Fatal("the artifact card vanished")
		}
		if assassin && z.Kind != game.ZoneHand {
			t.Errorf("with an Assassin the card is in %v, want the hand", z.Kind)
		}
		if !assassin {
			top, _ := me.Library.Top()
			if z.Kind != game.ZoneLibrary || top.InstanceID != relic {
				t.Errorf("without an Assassin the card is in %v and the top card is %q, want it on top of the library", z.Kind, top.Name)
			}
		}
	}
}

// --- Evie Frye ----------------------------------------------------------------

func gapb39EvieLoot(t *testing.T, discard string) (*game.Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	evie := b12Push(g, me.ID, "Evie Frye", "Legendary Creature — Human Assassin", gapb39EvieOracle, 2, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	discardCard := gapb39Hand(me, "Discard Me", discard, "{1}")
	advanceToMain(t, g)
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, evie, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, discardCard)
	passPriorityAroundTable(t, g)
	return g, bear, discardCard
}

func TestGapB39EvieFryeMakesACreatureUnblockableWhenYouDiscardACreatureCard(t *testing.T) {
	g, bear, _ := gapb39EvieLoot(t, "Creature — Bear")
	answerPickTargetIfOpen(t, g, bear)
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, bear, game.CantBeBlocked)
}

func TestGapB39EvieFryeDoesNothingExtraForANoncreatureDiscard(t *testing.T) {
	g, bear, _ := gapb39EvieLoot(t, "Sorcery")
	if pickTargetPromptOpen(g) != nil {
		t.Fatal("a target prompt opened for a noncreature discard")
	}
	assertRestrictions(t, g, bear, 0)
}

// --- Leonardo da Vinci --------------------------------------------------------

func TestGapB39LeonardoSetsThoptersToTheCardsInHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	leo := b12Push(g, me.ID, "Leonardo da Vinci", "Legendary Creature — Human Artificer", gapb39LeonardoOracle, 3, 3)
	thopter := pushPermanent(g, me.ID, game.Card{Name: "Thopter", TypeLine: "Token Artifact Creature — Thopter", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	size := me.Hand.Size()
	floatForTest(g, me, "CCCUU")
	if err := g.ActivateCatalogAbility(me.ID, leo, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := effectivePTOf(t, g, thopter); p != size || tt != size {
		t.Errorf("the Thopter is %d/%d, want %d/%d", p, tt, size, size)
	}
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Errorf("the Bear's power is %d, want 2", p)
	}
	late := pushPermanent(g, me.ID, game.Card{Name: "Thopter", TypeLine: "Token Artifact Creature — Thopter", Power: 1, Toughness: 1})
	if p, _ := effectivePTOf(t, g, late); p != 1 {
		t.Errorf("a Thopter that arrived later is affected: power %d", p)
	}
	dftMPPassTurn(t, g)
	if p, _ := effectivePTOf(t, g, thopter); p != 1 {
		t.Errorf("the base P/T outlasted the turn: power %d", p)
	}
}

func TestGapB39LeonardoCopiesADiscardedArtifactAsAThopter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	leo := b12Push(g, me.ID, "Leonardo da Vinci", "Legendary Creature — Human Artificer", gapb39LeonardoOracle, 3, 3)
	relic := gapb39Hand(me, "Great Relic", "Legendary Artifact", "{3}")
	advanceToMain(t, g)
	floatForTest(g, me, "CCU")
	if err := g.ActivateCatalogAbility(me.ID, leo, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(relic); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the discarded artifact is in %v, want exile", z != nil)
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Great Relic" && c.IsToken() {
			found = true
			if p, tt := effectivePTOf(t, g, c.InstanceID); p != 0 || tt != 2 {
				t.Errorf("the copy is %d/%d, want 0/2", p, tt)
			}
			types, subs := effectiveTypes(t, g, c.InstanceID), effectiveSubtypes(t, g, c.InstanceID)
			if !hasFold(types, "Artifact") || !hasFold(types, "Creature") || !hasFold(subs, "Thopter") {
				t.Errorf("types %v subtypes %v, want an artifact creature Thopter", types, subs)
			}
			if !effectiveAbilitiesContain(t, g, c.InstanceID, "flying") {
				t.Error("the copy has no flying")
			}
		}
	}
	if !found {
		t.Error("no token copy of the discarded artifact")
	}
}

func TestGapB39LeonardoLeavesANonartifactDiscardInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	leo := b12Push(g, me.ID, "Leonardo da Vinci", "Legendary Creature — Human Artificer", gapb39LeonardoOracle, 3, 3)
	spell := gapb39Hand(me, "Plain Spell", "Sorcery", "{1}")
	advanceToMain(t, g)
	floatForTest(g, me, "CCU")
	if err := g.ActivateCatalogAbility(me.ID, leo, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, spell)
	passPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(spell); z == nil || z.Kind != game.ZoneGraveyard {
		t.Error("a nonartifact discard left the graveyard")
	}
}

// --- Haytham Kenway -----------------------------------------------------------

func TestGapB39HaythamExilesOneCreatureForEachOpponentUntilHeLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	a1 := b12Creature(g, a.ID, "A One", "Creature — Ogre", 2, 2)
	a2 := b12Creature(g, a.ID, "A Two", "Creature — Ogre", 2, 2)
	b1 := b12Creature(g, b.ID, "B One", "Creature — Ogre", 2, 2)
	castHoldCreature(t, g, "Haytham Kenway", "Legendary Creature — Human Knight", gapb39HaythamOracle, 3, 3)
	var haytham uuid.UUID
	for _, c := range g.Stack.Cards {
		if c.Name == "Haytham Kenway" {
			haytham = c.InstanceID
		}
	}
	passPriorityAroundTable(t, g)
	// One "up to one" clause per opponent, answered in turn: each prompt offers
	// only that opponent's creatures.
	prompts := 0
	for p := latestPickTarget(g, me.ID); p != nil && prompts < 4; p = latestPickTarget(g, me.ID) {
		prompts++
		var pick []game.TargetRef
		for _, want := range []uuid.UUID{a1, b1} {
			if hasID(p.PickTargetCards, want) {
				pick = cardRefs(want)
			}
		}
		if hasID(p.PickTargetCards, a1) && hasID(p.PickTargetCards, b1) {
			t.Fatal("one clause offered creatures of two opponents")
		}
		if err := g.ResolvePickTargets(p.ID, me.ID, pick); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	}
	if prompts == 0 {
		t.Fatal("no target prompt")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a1) || g.Battlefield.Contains(b1) {
		t.Fatal("the chosen creatures were not exiled")
	}
	if !g.Battlefield.Contains(a2) {
		t.Error("an unchosen creature was exiled")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Haytham Kenway" {
			haytham = c.InstanceID
		}
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(haytham) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	// A creature that returns from exile is a new object, so find it by name.
	named := func(name string) bool {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				return true
			}
		}
		return false
	}
	if !named("A One") {
		t.Error("the first exiled creature did not return when Haytham left")
	}
	if !named("B One") {
		t.Error("the second exiled creature did not return when Haytham left")
	}
}

func TestGapB39HaythamLordsTheOtherKnights(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	haytham := b12Push(g, me.ID, "Haytham Kenway", "Legendary Creature — Human Knight", gapb39HaythamOracle, 3, 3)
	knight := b12Creature(g, me.ID, "Knight", "Creature — Human Knight", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Knight", "Creature — Human Knight", 2, 2)
	const prot = "protection from creatures of type Assassin"
	if !effectiveAbilitiesContain(t, g, haytham, prot) {
		t.Error("Haytham has no protection from Assassins")
	}
	if p, tt := effectivePTOf(t, g, knight); p != 4 || tt != 4 {
		t.Errorf("my Knight is %d/%d, want 4/4", p, tt)
	}
	if !effectiveAbilitiesContain(t, g, knight, prot) {
		t.Error("my Knight has no protection from Assassins")
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 2 || effectiveAbilitiesContain(t, g, theirs, prot) {
		t.Error("an opponent's Knight got the bonus")
	}
	if p, _ := effectivePTOf(t, g, haytham); p != 3 {
		t.Errorf("Haytham's power is %d, want 3 (other Knights only)", p)
	}
}

// --- Sokrates, Athenian Teacher -----------------------------------------------

func TestGapB39SokratesIsHexproofOnlyWhileUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sok := b12Push(g, me.ID, "Sokrates, Athenian Teacher", "Legendary Creature — Human Advisor", gapb39SokratesOracle, 0, 4)
	if !effectiveAbilitiesContain(t, g, sok, "defender") || !effectiveAbilitiesContain(t, g, sok, "hexproof") {
		t.Error("untapped Sokrates lacks defender or hexproof")
	}
	gapb39Tap(g, sok)
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if effectiveAbilitiesContain(t, g, sok, "hexproof") {
		t.Error("tapped Sokrates has hexproof")
	}
}

func TestGapB39SokratesPreventsCombatDamageToAPlayerAndBothDrawHalf(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sok := b12Push(g, me.ID, "Sokrates, Athenian Teacher", "Legendary Creature — Human Advisor", gapb39SokratesOracle, 0, 4)
	brute := b12Creature(g, me.ID, "Brute", "Creature — Ogre", 5, 5)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sok, 0, game.ActivateAbilityParams{Targets: cardRefs(brute), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	life, mine, theirs := opp.Life, me.Hand.Size(), opp.Hand.Size()
	attackWith(t, g, opp.ID, brute)
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("the player lost %d life, want the damage prevented", life-opp.Life)
	}
	if got := me.Hand.Size() - mine; got != 2 {
		t.Errorf("the creature's controller drew %d, want 2 (half of 5, rounded down)", got)
	}
	if got := opp.Hand.Size() - theirs; got != 2 {
		t.Errorf("the damaged player drew %d, want 2", got)
	}
}
