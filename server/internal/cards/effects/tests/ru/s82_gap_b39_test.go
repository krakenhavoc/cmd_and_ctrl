package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aya := effectstest.B12Push(g, me.ID, "Aya of Alexandria", "Legendary Creature — Human Assassin", gapb39AyaOracle, 4, 3)
	for _, kw := range []string{"menace", "lifelink"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, aya, kw) {
			t.Errorf("Aya has no %s", kw)
		}
	}
	relic := effectstest.B12Creature(g, me.ID, "Relic Golem", "Artifact Creature — Golem", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, aya, relic, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapb39Assassins(g, me.ID); got != 2 {
		t.Errorf("%d Assassin tokens, want 2 (Aya and the artifact creature; not the Bear)", got)
	}
}

// --- Eivor, Battle-Ready ------------------------------------------------------

func TestGapB39EivorDealsDamageForEachEquipmentToEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eivor := effectstest.B12Push(g, me.ID, "Eivor, Battle-Ready", "Legendary Creature — Human Assassin Warrior", gapb39EivorOracle, 5, 5)
	for _, kw := range []string{"vigilance", "haste"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, eivor, kw) {
			t.Errorf("Eivor has no %s", kw)
		}
	}
	effectstest.B12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	effectstest.B12Permanent(g, me.ID, "Axe", "Artifact — Equipment")
	effectstest.B12Permanent(g, opp.ID, "Their Sword", "Artifact — Equipment")
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	effectstest.AttackWith(t, g, opp.ID, eivor)
	effectstest.PassPriorityAroundTable(t, g)
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		spy := effectstest.B12Push(g, me.ID, "Brotherhood Spy", "Creature — Human Assassin", gapb39SpyOracle, 1, 3)
		typeLine := "Creature — Human Assassin"
		if legend {
			typeLine = "Legendary Creature — Human Assassin"
		}
		effectstest.B12Creature(g, me.ID, "Friend", typeLine, 2, 2)
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.PassPriorityAroundTable(t, g)
		power := effectstest.EffectivePower(t, g, spy)
		if legend {
			effectstest.AssertRestrictions(t, g, spy, game.CantBeBlocked)
			if power != 2 {
				t.Errorf("with a legendary Assassin the Spy's power is %d, want 2", power)
			}
			continue
		}
		effectstest.AssertRestrictions(t, g, spy, 0)
		if power != 1 {
			t.Errorf("without one the Spy's power is %d, want 1", power)
		}
	}
}

// --- Ezio, Brash Novice -------------------------------------------------------

func TestGapB39EzioBecomesAnAssassinWithFirstStrikeAtTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ezio := effectstest.B12Push(g, me.ID, "Ezio, Brash Novice", "Legendary Creature — Human", gapb39EzioOracle, 1, 1)
	if effectstest.EffectiveAbilitiesContain(t, g, ezio, "first strike") || effects.HasFold(effectstest.EffectiveSubtypes(t, g, ezio), "Assassin") {
		t.Fatal("Ezio starts with first strike or as an Assassin")
	}
	effectstest.AttackWith(t, g, opp.ID, ezio)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, ezio, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d counters after one attack, want 1", got)
	}
	if effectstest.EffectiveAbilitiesContain(t, g, ezio, "first strike") {
		t.Error("first strike with one counter")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ezio, "charge", 1) })
	if !effectstest.EffectiveAbilitiesContain(t, g, ezio, "first strike") {
		t.Error("no first strike with two counters (of any kind)")
	}
	subs := effectstest.EffectiveSubtypes(t, g, ezio)
	if !effects.HasFold(subs, "Assassin") || !effects.HasFold(subs, "Human") {
		t.Errorf("subtypes %v, want Human and Assassin", subs)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ezio, "charge", -1) })
	if effectstest.EffectiveAbilitiesContain(t, g, ezio, "first strike") {
		t.Error("first strike stayed after a counter was removed")
	}
}

// --- Bayek of Siwa ------------------------------------------------------------

func TestGapB39BayekGrantsDoubleStrikeToOtherHistoricCreaturesOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bayek := effectstest.B12Push(g, me.ID, "Bayek of Siwa", "Legendary Creature — Human Assassin", gapb39BayekOracle, 3, 4)
	if !effectstest.EffectiveAbilitiesContain(t, g, bayek, "double strike") {
		t.Error("Bayek has no double strike")
	}
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 2, 2)
	if !effectstest.EffectiveAbilitiesContain(t, g, golem, "double strike") {
		t.Error("my historic creature has no double strike on my turn")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "double strike") || effectstest.EffectiveAbilitiesContain(t, g, theirs, "double strike") {
		t.Error("a nonhistoric or opposing creature got double strike")
	}
	effectstest.DftMPPassTurn(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, golem, "double strike") {
		t.Error("double strike on an opponent's turn")
	}
	def, _ := effects.Lookup(gapb39BayekOracle)
	if len(def.AlternativeCosts) != 1 || def.AlternativeCosts[0].Key != "disguise" {
		t.Errorf("Bayek's alternative costs %+v, want disguise", def.AlternativeCosts)
	}
}

// --- Kassandra, Eagle Bearer --------------------------------------------------

func TestGapB39KassandraFetchesTheSpearAndDrawsForALegendaryEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spear := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: spear, Name: "The Spear of Leonidas", TypeLine: "Legendary Artifact — Equipment", Owner: me.ID, Controller: me.ID})
	effectstest.CastHoldCreature(t, g, "Kassandra, Eagle Bearer", "Legendary Creature — Human Assassin Warrior", gapb39KassandraOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "The Spear of Leonidas")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(spear) {
		t.Fatal("The Spear of Leonidas was not put onto the battlefield")
	}
	for _, legend := range []bool{true, false} {
		g2 := effectstest.NewCatalogGame(t)
		me2, opp2 := g2.Seats[0], g2.Seats[1]
		effectstest.B12Push(g2, me2.ID, "Kassandra, Eagle Bearer", "Legendary Creature — Human Assassin Warrior", gapb39KassandraOracle, 2, 2)
		bearer := effectstest.B12Creature(g2, me2.ID, "Bearer", "Creature — Human", 2, 2)
		typeLine := "Artifact — Equipment"
		if legend {
			typeLine = "Legendary Artifact — Equipment"
		}
		effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{
			InstanceID: uuid.New(), Name: "Blade", TypeLine: typeLine, Owner: me2.ID, Controller: me2.ID,
			AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bearer},
		})
		hand := me2.Hand.Size()
		effectstest.AttackWith(t, g2, opp2.ID, bearer)
		effectstest.PassPriorityAroundTable(t, g2)
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		if assassin {
			effectstest.B12Creature(g, me.ID, "Friend", "Creature — Human Assassin", 1, 1)
		}
		relic := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: relic, Name: "Gadget", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
		for i := 0; i < 3; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		}
		effectstest.CastHoldCreature(t, g, "Loyal Inventor", "Creature — Human Artificer", gapb39InventorOracle, 2, 3)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerSearchNamed(t, g, me.ID, "Gadget")
		effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	evie := effectstest.B12Push(g, me.ID, "Evie Frye", "Legendary Creature — Human Assassin", gapb39EvieOracle, 2, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	discardCard := gapb39Hand(me, "Discard Me", discard, "{1}")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, evie, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, discardCard)
	effectstest.PassPriorityAroundTable(t, g)
	return g, bear, discardCard
}

func TestGapB39EvieFryeMakesACreatureUnblockableWhenYouDiscardACreatureCard(t *testing.T) {
	g, bear, _ := gapb39EvieLoot(t, "Creature — Bear")
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, bear, game.CantBeBlocked)
}

func TestGapB39EvieFryeDoesNothingExtraForANoncreatureDiscard(t *testing.T) {
	g, bear, _ := gapb39EvieLoot(t, "Sorcery")
	if pickTargetPromptOpen(g) != nil {
		t.Fatal("a target prompt opened for a noncreature discard")
	}
	effectstest.AssertRestrictions(t, g, bear, 0)
}

// --- Leonardo da Vinci --------------------------------------------------------

func TestGapB39LeonardoSetsThoptersToTheCardsInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	leo := effectstest.B12Push(g, me.ID, "Leonardo da Vinci", "Legendary Creature — Human Artificer", gapb39LeonardoOracle, 3, 3)
	thopter := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Thopter", TypeLine: "Token Artifact Creature — Thopter", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	size := me.Hand.Size()
	effectstest.FloatForTest(g, me, "CCCUU")
	if err := g.ActivateCatalogAbility(me.ID, leo, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePTOf(t, g, thopter); p != size || tt != size {
		t.Errorf("the Thopter is %d/%d, want %d/%d", p, tt, size, size)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 {
		t.Errorf("the Bear's power is %d, want 2", p)
	}
	late := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Thopter", TypeLine: "Token Artifact Creature — Thopter", Power: 1, Toughness: 1})
	if p, _ := effectstest.EffectivePTOf(t, g, late); p != 1 {
		t.Errorf("a Thopter that arrived later is affected: power %d", p)
	}
	effectstest.DftMPPassTurn(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, thopter); p != 1 {
		t.Errorf("the base P/T outlasted the turn: power %d", p)
	}
}

func TestGapB39LeonardoCopiesADiscardedArtifactAsAThopter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	leo := effectstest.B12Push(g, me.ID, "Leonardo da Vinci", "Legendary Creature — Human Artificer", gapb39LeonardoOracle, 3, 3)
	relic := gapb39Hand(me, "Great Relic", "Legendary Artifact", "{3}")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCU")
	if err := g.ActivateCatalogAbility(me.ID, leo, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(relic); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the discarded artifact is in %v, want exile", z != nil)
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Great Relic" && c.IsToken() {
			found = true
			if p, tt := effectstest.EffectivePTOf(t, g, c.InstanceID); p != 0 || tt != 2 {
				t.Errorf("the copy is %d/%d, want 0/2", p, tt)
			}
			types, subs := effectstest.EffectiveTypes(t, g, c.InstanceID), effectstest.EffectiveSubtypes(t, g, c.InstanceID)
			if !effects.HasFold(types, "Artifact") || !effects.HasFold(types, "Creature") || !effects.HasFold(subs, "Thopter") {
				t.Errorf("types %v subtypes %v, want an artifact creature Thopter", types, subs)
			}
			if !effectstest.EffectiveAbilitiesContain(t, g, c.InstanceID, "flying") {
				t.Error("the copy has no flying")
			}
		}
	}
	if !found {
		t.Error("no token copy of the discarded artifact")
	}
}

func TestGapB39LeonardoLeavesANonartifactDiscardInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	leo := effectstest.B12Push(g, me.ID, "Leonardo da Vinci", "Legendary Creature — Human Artificer", gapb39LeonardoOracle, 3, 3)
	spell := gapb39Hand(me, "Plain Spell", "Sorcery", "{1}")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCU")
	if err := g.ActivateCatalogAbility(me.ID, leo, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(spell); z == nil || z.Kind != game.ZoneGraveyard {
		t.Error("a nonartifact discard left the graveyard")
	}
}

// --- Haytham Kenway -----------------------------------------------------------

func TestGapB39HaythamExilesOneCreatureForEachOpponentUntilHeLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	a1 := effectstest.B12Creature(g, a.ID, "A One", "Creature — Ogre", 2, 2)
	a2 := effectstest.B12Creature(g, a.ID, "A Two", "Creature — Ogre", 2, 2)
	b1 := effectstest.B12Creature(g, b.ID, "B One", "Creature — Ogre", 2, 2)
	effectstest.CastHoldCreature(t, g, "Haytham Kenway", "Legendary Creature — Human Knight", gapb39HaythamOracle, 3, 3)
	var haytham uuid.UUID
	for _, c := range g.Stack.Cards {
		if c.Name == "Haytham Kenway" {
			haytham = c.InstanceID
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	// One "up to one" clause per opponent, answered in turn: each prompt offers
	// only that opponent's creatures.
	prompts := 0
	for p := effectstest.LatestPickTarget(g, me.ID); p != nil && prompts < 4; p = effectstest.LatestPickTarget(g, me.ID) {
		prompts++
		var pick []game.TargetRef
		for _, want := range []uuid.UUID{a1, b1} {
			if effectstest.HasID(p.PickTargetCards, want) {
				pick = effectstest.CardRefs(want)
			}
		}
		if effectstest.HasID(p.PickTargetCards, a1) && effectstest.HasID(p.PickTargetCards, b1) {
			t.Fatal("one clause offered creatures of two opponents")
		}
		if err := g.ResolvePickTargets(p.ID, me.ID, pick); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	}
	if prompts == 0 {
		t.Fatal("no target prompt")
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	haytham := effectstest.B12Push(g, me.ID, "Haytham Kenway", "Legendary Creature — Human Knight", gapb39HaythamOracle, 3, 3)
	knight := effectstest.B12Creature(g, me.ID, "Knight", "Creature — Human Knight", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Knight", "Creature — Human Knight", 2, 2)
	const prot = "protection from creatures of type Assassin"
	if !effectstest.EffectiveAbilitiesContain(t, g, haytham, prot) {
		t.Error("Haytham has no protection from Assassins")
	}
	if p, tt := effectstest.EffectivePTOf(t, g, knight); p != 4 || tt != 4 {
		t.Errorf("my Knight is %d/%d, want 4/4", p, tt)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, knight, prot) {
		t.Error("my Knight has no protection from Assassins")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 2 || effectstest.EffectiveAbilitiesContain(t, g, theirs, prot) {
		t.Error("an opponent's Knight got the bonus")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, haytham); p != 3 {
		t.Errorf("Haytham's power is %d, want 3 (other Knights only)", p)
	}
}

// --- Sokrates, Athenian Teacher -----------------------------------------------

func TestGapB39SokratesIsHexproofOnlyWhileUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sok := effectstest.B12Push(g, me.ID, "Sokrates, Athenian Teacher", "Legendary Creature — Human Advisor", gapb39SokratesOracle, 0, 4)
	if !effectstest.EffectiveAbilitiesContain(t, g, sok, "defender") || !effectstest.EffectiveAbilitiesContain(t, g, sok, "hexproof") {
		t.Error("untapped Sokrates lacks defender or hexproof")
	}
	gapb39Tap(g, sok)
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if effectstest.EffectiveAbilitiesContain(t, g, sok, "hexproof") {
		t.Error("tapped Sokrates has hexproof")
	}
}

func TestGapB39SokratesPreventsCombatDamageToAPlayerAndBothDrawHalf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sok := effectstest.B12Push(g, me.ID, "Sokrates, Athenian Teacher", "Legendary Creature — Human Advisor", gapb39SokratesOracle, 0, 4)
	brute := effectstest.B12Creature(g, me.ID, "Brute", "Creature — Ogre", 5, 5)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sok, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(brute), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	life, mine, theirs := opp.Life, me.Hand.Size(), opp.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, brute)
	effectstest.PassPriorityAroundTable(t, g)
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
