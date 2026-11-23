package testsm

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mkm_dl3_cards_test.go — Murders at Karlov Manor cards D–L, third batch
// (#4469). Each test plays the real catalog entry and asserts what its
// text says. The face-down helpers are the M–P batch's (mkmMPCastFaceDown,
// mkmMPTurnFaceUp, mkmMPPT).

const (
	mkmDL3AccusationOracle = "677b9c19-4886-4e06-88f8-9afa94e58aa8"
	mkmDL3DiligenceOracle  = "d412150d-3c70-44cf-97cd-965c438e7303"
	mkmDL3RageOracle       = "e1905321-180b-41d2-b7ef-0974ab90f188"
	mkmDL3BloomKinOracle   = "66e74e8f-036e-4b26-a4a8-2365891765bc"
	mkmDL3FussOracle       = "88e9006a-cfdf-40f3-80d5-23465f3dceca"
	mkmDL3OrangutanOracle  = "1f3e7420-86aa-4847-bc2d-21f6bae470c8"
	mkmDL3StriderOracle    = "9ea58a77-9b2e-4ae2-8a31-a51804900a3b"
	mkmDL3DronesmithOracle = "a6e2d8bd-8fe7-4de0-af25-7ad3e9dcb81b"
	mkmDL3HomicideOracle   = "dc4d5602-47cb-47c8-8a43-3b840e12b79c"
	mkmDL3HotshotOracle    = "288fcf5c-f5aa-4f62-9dd8-ba295df48f7a"
	mkmDL3HustleOracle     = "e294d85f-caf2-4939-99d3-7ca5bba2b8d1"
	mkmDL3BystanderOracle  = "c293f1c0-93de-4a73-aaf6-4e9f01e557ec"
	mkmDL3KellanOracle     = "2f2d75a0-d751-4c74-b357-8f92a992447f"
	mkmDL3KrenkoOracle     = "91b2520e-85b6-4e1f-88cf-a585feeb8e65"
	mkmDL3HaunchOracle     = "85fdfa3c-4912-496a-bbcb-1ae13a34e917"
)

func TestMkmDL3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mkmDL3AccusationOracle:    "Dramatic Accusation",
		mkmDL3DiligenceOracle:     "Due Diligence",
		mkmDL3RageOracle:          "Felonious Rage",
		mkmDL3BloomKinOracle:      "Flourishing Bloom-Kin",
		mkmDL3FussOracle:          "Fuss",
		mkmDL3FussOracle + "#1":   "Bother",
		mkmDL3OrangutanOracle:     "Gearbane Orangutan",
		mkmDL3StriderOracle:       "Gravestone Strider",
		mkmDL3DronesmithOracle:    "Harried Dronesmith",
		mkmDL3HomicideOracle:      "Homicide Investigator",
		mkmDL3HotshotOracle:       "Hotshot Investigators",
		mkmDL3HustleOracle:        "Hustle",
		mkmDL3HustleOracle + "#1": "Bustle",
		mkmDL3BystanderOracle:     "Innocent Bystander",
		mkmDL3KellanOracle:        "Kellan, Inquisitive Prodigy",
		mkmDL3KellanOracle + "#1": "Tail the Suspect",
		mkmDL3KrenkoOracle:        "Krenko, Baron of Tin Street",
		mkmDL3HaunchOracle:        "Krovod Haunch",
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
	ac := game.AlternativeCostByKey(mkmDL3BloomKinOracle, "disguise")
	if ac == nil || ac.FaceDown == nil || ac.ManaCost != effects.MorphCastCost ||
		ac.FaceDown.Kind != game.FaceDownDisguised || ac.FaceDown.FaceUpCost != "{4}{G}" {
		t.Errorf("Flourishing Bloom-Kin's disguise is not wired: %+v", ac)
	}
}

// mkmDL3Split puts a split or Adventure card in p's hand, both faces
// costing nothing so the test needs no mana.
func mkmDL3Split(p *game.Player, oracle, layout string, left, right game.Face) game.Card {
	left.ManaCost, right.ManaCost = "{0}", "{0}"
	card := effectstest.MdfcCard(p.ID, oracle, layout, left, right)
	p.Hand.PushTop(card)
	return card
}

// --- Dramatic Accusation -------------------------------------------------------

func TestMkmDL3DramaticAccusationTapsLocksAndShufflesTheCreatureAway(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aura := effectstest.CastCatalogSpell(t, g, "Dramatic Accusation", "Enchantment — Aura", mkmDL3AccusationOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedTo(t, g, aura, theirs) {
		t.Fatal("the Aura is not on the creature")
	}
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Fatal("the enters trigger did not tap the enchanted creature")
	}
	effectstest.EndTurn(t, g)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("the enchanted creature untapped in its controller's untap step")
	}

	effectstest.FloatMana(t, g, me, "{U}{U}")
	effectstest.P7Activate(t, g, me, aura, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the creature is still on the battlefield")
	}
	if !opp.Library.Contains(theirs) {
		t.Error("the creature is not in its owner's library")
	}
}

// --- Due Diligence -------------------------------------------------------------

func TestMkmDL3DueDiligencePumpsTheHostAndAnotherCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "My Elk", "Creature — Elk", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Due Diligence", "Enchantment — Aura", mkmDL3DiligenceOracle, effectstest.B16TargetCard(host))
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, effectstest.CardRef(host)); err == nil {
		t.Fatal("the enchanted creature was accepted as the trigger's target")
	}
	effectstest.AnswerPickTarget(t, g, other)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmMPPT(t, g, host); p != 4 || tough != 4 || !effectstest.HasEffectiveKeyword(t, g, host, "vigilance") {
		t.Errorf("the enchanted creature is %d/%d (vigilance %v), want 4/4 with vigilance",
			p, tough, effectstest.HasEffectiveKeyword(t, g, host, "vigilance"))
	}
	if p, tough := effectstest.MkmMPPT(t, g, other); p != 5 || tough != 5 || !effectstest.HasEffectiveKeyword(t, g, other, "vigilance") {
		t.Errorf("the target is %d/%d (vigilance %v), want 5/5 with vigilance",
			p, tough, effectstest.HasEffectiveKeyword(t, g, other, "vigilance"))
	}
	effectstest.EndTurn(t, g)
	if p, _ := effectstest.MkmMPPT(t, g, other); p != 3 || effectstest.HasEffectiveKeyword(t, g, other, "vigilance") {
		t.Error("the trigger's pump outlasted the turn")
	}
	if p, _ := effectstest.MkmMPPT(t, g, host); p != 4 {
		t.Error("the Aura's static ended with the turn")
	}
}

// --- Felonious Rage ------------------------------------------------------------

func TestMkmDL3FeloniousRagePumpsHastesAndLeavesADetectiveWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Felonious Rage", "Instant", mkmDL3RageOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmMPPT(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the target is %d/%d, want 4/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the target has no haste")
	}
	if effectstest.TmtB1Count(g, me.ID, "Detective") != 0 {
		t.Fatal("a Detective came before the creature died")
	}
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Detective"); got != 1 {
		t.Errorf("%d Detectives after the creature died, want 1", got)
	}
}

func TestMkmDL3FeloniousRageTargetsOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Felonious Rage", "Instant", mkmDL3RageOracle, effectstest.B16TargetCard(theirs)); err == nil {
		t.Error("an opponent's creature was accepted as the target")
	}
}

// --- Flourishing Bloom-Kin -----------------------------------------------------

func TestMkmDL3FlourishingBloomKinCountsForestsAndFetchesTwoWhenTurnedFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushTypedCard(g, me.ID, "Forest", "Basic Land — Forest", "")
	effectstest.PushTypedCard(g, opp.ID, "Forest", "Basic Land — Forest", "")
	id := effectstest.MkmMPCastFaceDown(t, g, "Flourishing Bloom-Kin", "Creature — Plant Elemental", mkmDL3BloomKinOracle, 0, 0)
	if p, tough := effectstest.MkmMPPT(t, g, id); p != 2 || tough != 2 {
		t.Fatalf("face down the Bloom-Kin is %d/%d, want 2/2", p, tough)
	}

	first := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	second := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Overgrown Tomb", TypeLine: "Land — Swamp Forest"})
	effectstest.MkmMPTurnFaceUp(t, g, me, id, "G", "C", "C", "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, second)
	effectstest.AnswerSearchByID(t, g, me.ID, first)
	if !g.Battlefield.Contains(second) || !me.Hand.Contains(first) {
		t.Fatalf("the nonbasic Forest is on the battlefield %v, the other in hand %v; want both",
			g.Battlefield.Contains(second), me.Hand.Contains(first))
	}
	if !effectstest.B16Tapped(t, g, second) {
		t.Error("the Forest entered untapped")
	}
	// Two Forests I control now; the opponent's does not count.
	if p, tough := effectstest.MkmMPPT(t, g, id); p != 2 || tough != 2 {
		t.Errorf("face up with two Forests the Bloom-Kin is %d/%d, want 2/2", p, tough)
	}
}

// --- Fuss // Bother ------------------------------------------------------------

func TestMkmDL3FussPutsACounterOnEachAttackingCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.B12Creature(g, me.ID, "Attacker A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Attacker B", "Creature — Bear", 2, 2)
	home := effectstest.B12Creature(g, me.ID, "Stays Home", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)

	card := mkmDL3Split(me, mkmDL3FussOracle, game.LayoutSplit,
		game.Face{Name: "Fuss", TypeLine: "Instant"}, game.Face{Name: "Bother", TypeLine: "Sorcery"})
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Fuss: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"attacker A", a, 1}, {"attacker B", b, 1}, {"the creature at home", home, 0}, {"the opponent's creature", theirs, 0}} {
		if got := effectstest.CounterCount(g, c.id, game.CounterPlusOne); got != c.want {
			t.Errorf("%s has %d +1/+1 counters, want %d", c.name, got, c.want)
		}
	}
}

func TestMkmDL3BotherMakesThreeThoptersThenSurveilsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	effectstest.AdvanceToMain(t, g)
	card := mkmDL3Split(me, mkmDL3FussOracle, game.LayoutSplit,
		game.Face{Name: "Fuss", TypeLine: "Instant"}, game.Face{Name: "Bother", TypeLine: "Sorcery"})
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Bother: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	thopters := effectstest.TmtB1Named(g, "Thopter")
	if len(thopters) != 3 {
		t.Fatalf("%d Thopters, want 3", len(thopters))
	}
	for _, id := range thopters {
		c := effectstest.TmtB1Card(t, g, id)
		if c.Controller != me.ID || c.Power != 1 || c.Toughness != 1 || !c.IsArtifact() || !game.HasKeyword(&c, "flying") {
			t.Errorf("a Thopter is a %d/%d (artifact %v) under %s, want my 1/1 flying artifact", c.Power, c.Toughness, c.IsArtifact(), c.Controller)
		}
	}
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("no surveil 2 after the Thopters: %+v", c)
	}
}

// --- Gearbane Orangutan --------------------------------------------------------

func TestMkmDL3GearbaneOrangutanDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.PushTypedCard(g, opp.ID, "Their Rock", "Artifact", "{2}")
	ape := effectstest.CastAndResolveCreature(t, g, "Gearbane Orangutan", "Creature — Ape", mkmDL3OrangutanOracle)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	effectstest.AssertKeywords(t, g, ape, "reach")
	if got := effectstest.CounterCount(g, ape, game.CounterPlusOne); got != 0 {
		t.Errorf("the destroy mode put %d counters on the Ape", got)
	}
}

func TestMkmDL3GearbaneOrangutanSacrificesAnArtifactForTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	clue := effectstest.PushToken(g, me.ID, effects.ClueToken())
	ape := effectstest.CastAndResolveCreature(t, g, "Gearbane Orangutan", "Creature — Ape", mkmDL3OrangutanOracle)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, clue)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(clue) {
		t.Error("the Clue was not sacrificed")
	}
	if got := effectstest.CounterCount(g, ape, game.CounterPlusOne); got != 2 {
		t.Errorf("the Ape has %d +1/+1 counters, want 2", got)
	}
}

func TestMkmDL3GearbaneOrangutanWithNoArtifactGetsNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ape := effectstest.CastAndResolveCreature(t, g, "Gearbane Orangutan", "Creature — Ape", mkmDL3OrangutanOracle)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, ape, game.CounterPlusOne); got != 0 {
		t.Errorf("with nothing sacrificed the Ape has %d counters, want 0", got)
	}
}

// --- Gravestone Strider --------------------------------------------------------

func TestMkmDL3GravestoneStriderFiltersOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	strider := effectstest.B12Push(g, me.ID, "Gravestone Strider", "Artifact Creature — Golem", mkmDL3StriderOracle, 1, 3)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FillPoolColored(me, "C", 2)
	if err := g.ActivateManaAbility(me.ID, strider, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("{1}: add {B}: %v", err)
	}
	effectstest.MkWantPool(t, me, map[string]int{"C": 1, "B": 1})
	if err := g.ActivateManaAbility(me.ID, strider, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Error("a second activation this turn was allowed")
	}
}

func TestMkmDL3GravestoneStriderExilesACardFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	onBattlefield := effectstest.B12Push(g, me.ID, "Gravestone Strider", "Artifact Creature — Golem", mkmDL3StriderOracle, 1, 3)
	victim := effectstest.Batch01GraveyardCard(opp, "Their Card", "Sorcery")
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, onBattlefield, 0, game.ActivateAbilityParams{
		Strict: true, Targets: effectstest.B16TargetCard(victim)}); err == nil {
		t.Fatal("the graveyard ability was activated from the battlefield")
	}

	id := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Gravestone Strider", TypeLine: "Artifact Creature — Golem",
		OracleID: mkmDL3StriderOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Strict: true, Targets: effectstest.B16TargetCard(victim)}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	if !g.Exile.Contains(id) {
		t.Error("the Strider was not exiled as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(victim) || opp.Graveyard.Contains(victim) {
		t.Error("the target card was not exiled from the graveyard")
	}
}

// --- Harried Dronesmith --------------------------------------------------------

func TestMkmDL3HarriedDronesmithMakesAHastyThopterAndSacrificesItAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Harried Dronesmith", "Creature — Human Artificer", mkmDL3DronesmithOracle, 2, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	thopters := effectstest.TmtB1Named(g, "Thopter")
	if len(thopters) != 1 {
		t.Fatalf("%d Thopters at the beginning of combat, want 1", len(thopters))
	}
	thopter := thopters[0]
	if c := effectstest.TmtB1Card(t, g, thopter); c.Controller != me.ID || !game.HasKeyword(&c, "flying") {
		t.Error("the Thopter is not my flyer")
	}
	if !effectstest.HasEffectiveKeyword(t, g, thopter, "haste") {
		t.Error("the Thopter has no haste")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(thopter) {
		t.Error("the Thopter survived my end step")
	}
}

// --- Homicide Investigator -----------------------------------------------------

func TestMkmDL3HomicideInvestigatorInvestigatesOnceEachTurnForNontokenDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Homicide Investigator", "Creature — Human Detective", mkmDL3HomicideOracle, 2, 2)
	token := effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 red Goblin"))
	effectstest.KillCreature(t, g, me.ID, token)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 0 {
		t.Fatalf("a token dying made %d Clues, want 0", got)
	}

	first := effectstest.B12Creature(g, me.ID, "First Bear", "Creature — Bear", 2, 2)
	second := effectstest.B12Creature(g, me.ID, "Second Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, first)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Fatalf("%d Clues after a nontoken creature died, want 1", got)
	}
	effectstest.KillCreature(t, g, me.ID, second)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues after a second death this turn, want it to stay 1", got)
	}

	effectstest.EndTurn(t, g)
	third := effectstest.B12Creature(g, me.ID, "Third Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, third)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 2 {
		t.Errorf("%d Clues after a death on the next turn, want 2", got)
	}
}

func TestMkmDL3HomicideInvestigatorIgnoresAnOpponentsCreatureAndSeesItsOwnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	investigator := effectstest.B12Push(g, me.ID, "Homicide Investigator", "Creature — Human Detective", mkmDL3HomicideOracle, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.TmtB1Named(g, "Clue")); got != 0 {
		t.Fatalf("an opponent's creature dying made %d Clues", got)
	}
	effectstest.KillCreature(t, g, me.ID, investigator)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("the Investigator's own death made %d Clues, want 1", got)
	}
}

// --- Hotshot Investigators -----------------------------------------------------

func TestMkmDL3HotshotInvestigatorsBouncesYourCreatureAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	hotshot := effectstest.CastHoldCreature(t, g, "Hotshot Investigators", "Creature — Vedalken Detective", mkmDL3HotshotOracle, 4, 4)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, effectstest.CardRef(hotshot)); err == nil {
		t.Fatal("the Investigators may target themselves, but the trigger says \"other\"")
	}
	effectstest.AnswerPickTarget(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) {
		t.Error("my creature is not back in my hand")
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

func TestMkmDL3HotshotInvestigatorsBouncesAnOpponentsCreatureWithoutAClue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Hotshot Investigators", "Creature — Vedalken Detective", mkmDL3HotshotOracle, 4, 4)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) {
		t.Error("the opponent's creature is not in its owner's hand")
	}
	if got := len(effectstest.TmtB1Named(g, "Clue")); got != 0 {
		t.Errorf("%d Clues for a creature I did not control, want 0", got)
	}
	_ = me
}

// --- Hustle // Bustle ----------------------------------------------------------

func TestMkmDL3HustleMakesTheCreatureAttackOrBlockThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	card := mkmDL3Split(me, mkmDL3HustleOracle, game.LayoutSplit,
		game.Face{Name: "Hustle", TypeLine: "Instant"}, game.Face{Name: "Bustle", TypeLine: "Sorcery"})
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("cast Hustle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	blocks := false
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModAddBlockRequirement && m.Text == string(game.BlockRequirementBlocks) {
				blocks = true
			}
		}
	}
	if !blocks {
		t.Error("no blocks-if-able requirement on the target")
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Errorf("passing the declare attackers step without the creature: %v, want an attack requirement", err)
	}
}

func TestMkmDL3BustlePumpsTheTeamAndTurnsACreatureFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	phantom := effectstest.MkmMPCastFaceDown(t, g, "Defenestrated Phantom", "Creature — Spirit", mkmDLPhantomOracle, 4, 3)
	card := mkmDL3Split(me, mkmDL3HustleOracle, game.LayoutSplit,
		game.Face{Name: "Hustle", TypeLine: "Instant"}, game.Face{Name: "Bustle", TypeLine: "Sorcery"})
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Bustle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, phantom)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindCardAnywhere(t, g, phantom).FaceDown {
		t.Fatal("the face-down creature is still face down")
	}
	if p, tough := effectstest.MkmMPPT(t, g, phantom); p != 6 || tough != 5 {
		t.Errorf("the turned-up Phantom is %d/%d, want 6/5 (4/3 plus the pump)", p, tough)
	}
	if p, tough := effectstest.MkmMPPT(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("my Bear is %d/%d, want 4/4", p, tough)
	}
	for _, id := range []uuid.UUID{bear, phantom} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
			t.Error("a creature I control has no trample")
		}
	}
	effectstest.AssertKeywords(t, g, phantom, "flying", "trample")
}

// --- Innocent Bystander --------------------------------------------------------

func TestMkmDL3InnocentBystanderInvestigatesOnThreeDamageAtOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Push(g, me.ID, "Innocent Bystander", "Creature — Goblin Citizen", mkmDL3BystanderOracle, 2, 1)
	effectstest.B27Damage(g, foe, bystander, 3)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bystander) {
		t.Error("the Bystander survived 3 damage")
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues after 3 damage, want 1 even though it died", got)
	}
}

func TestMkmDL3InnocentBystanderDoesNotAddUpSeparateDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	// A Bystander made big enough to take both hits.
	bystander := effectstest.B12Push(g, me.ID, "Innocent Bystander", "Creature — Goblin Citizen", mkmDL3BystanderOracle, 2, 6)
	effectstest.B27Damage(g, foe, bystander, 2)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B27Damage(g, foe, bystander, 2)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 0 {
		t.Errorf("2 damage then 2 more made %d Clues, want 0", got)
	}
}

// --- Kellan, Inquisitive Prodigy // Tail the Suspect ---------------------------

func TestMkmDL3KellanDestroysAnArtifactAndDrawsOnlyIfItWasYours(t *testing.T) {
	for _, mine := range []bool{true, false} {
		name := "an opponent's artifact"
		if mine {
			name = "my artifact"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			owner := opp
			if mine {
				owner = me
			}
			rock := effectstest.PushTypedCard(g, owner.ID, "A Rock", "Artifact", "{2}")
			kellan := effectstest.B12Push(g, me.ID, "Kellan, Inquisitive Prodigy", "Legendary Creature — Human Faerie Detective", mkmDL3KellanOracle, 3, 4)
			effectstest.AssertKeywords(t, g, kellan, "flying", "vigilance")
			effectstest.SeedLibrary(me, "One", "Two")
			effectstest.DeclareAttack(t, g, opp.ID, kellan)
			hand := me.Hand.Size()
			effectstest.PickCard(t, g, me.ID, rock)
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(rock) {
				t.Error("the artifact was not destroyed")
			}
			want := 0
			if mine {
				want = 1
			}
			if got := me.Hand.Size() - hand; got != want {
				t.Errorf("drew %d cards, want %d", got, want)
			}
		})
	}
}

func TestMkmDL3TailTheSuspectInvestigatesAndGrantsALandPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	card := mkmDL3Split(me, mkmDL3KellanOracle, game.LayoutAdventure,
		game.Face{Name: "Kellan, Inquisitive Prodigy", TypeLine: "Legendary Creature — Human Faerie Detective", Power: 3, Toughness: 4},
		game.Face{Name: "Tail the Suspect", TypeLine: "Sorcery — Adventure"})
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Tail the Suspect: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
	if got := g.ExtraLandDropsThisTurn[me.ID]; got != 1 {
		t.Errorf("extra land plays = %d, want 1", got)
	}
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("the Adventure card is not in exile after it resolved")
	}
}

// --- Krenko, Baron of Tin Street -----------------------------------------------

func TestMkmDL3KrenkoSacrificesAnArtifactToGrowGoblinsAndPaysForAGoblin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	krenko := effectstest.B12Push(g, me.ID, "Krenko, Baron of Tin Street", "Legendary Creature — Goblin", mkmDL3KrenkoOracle, 3, 3)
	effectstest.AssertKeywords(t, g, krenko, "haste")
	myGoblin := effectstest.B12Creature(g, me.ID, "My Goblin", "Creature — Goblin Warrior", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirGoblin := effectstest.B12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	clue := effectstest.PushToken(g, me.ID, effects.ClueToken())

	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a creature that is not an artifact was accepted as the sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{clue}}); err != nil {
		t.Fatalf("activate with a Clue: %v", err)
	}
	// The sacrificed Clue went to the graveyard from the battlefield.
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	for _, c := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"Krenko", krenko, 1}, {"my Goblin", myGoblin, 1}, {"my Bear", bear, 0}, {"the opponent's Goblin", theirGoblin, 0}} {
		if got := effectstest.CounterCount(g, c.id, game.CounterPlusOne); got != c.want {
			t.Errorf("%s has %d +1/+1 counters, want %d", c.name, got, c.want)
		}
	}
	var made []uuid.UUID
	for _, id := range effectstest.TmtB1Named(g, "Goblin") {
		if c := effectstest.TmtB1Card(t, g, id); c.IsToken() && c.Controller == me.ID {
			made = append(made, id)
		}
	}
	if len(made) != 1 {
		t.Fatalf("%d Goblin tokens, want 1", len(made))
	}
	if !effectstest.HasEffectiveKeyword(t, g, made[0], "haste") {
		t.Error("the Goblin token has no haste")
	}
}

func TestMkmDL3KrenkoSeesAnOpponentsArtifactAndMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Krenko, Baron of Tin Street", "Legendary Creature — Goblin", mkmDL3KrenkoOracle, 3, 3)
	rock := effectstest.PushTypedCard(g, opp.ID, "Their Rock", "Artifact", "{2}")
	effectstest.WoeB1Destroy(t, g, rock)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("an opponent's artifact dying asked nothing")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Goblin"); got != 0 {
		t.Errorf("declining made %d Goblins", got)
	}
}

// --- Krovod Haunch -------------------------------------------------------------

func TestMkmDL3KrovodHaunchPumpsAndPaysForDogsWhenEaten(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	haunch := effectstest.B12Push(g, me.ID, "Krovod Haunch", "Artifact — Food Equipment", mkmDL3HaunchOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.P7Activate(t, g, me, haunch, 1, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)})
	if !effectstest.AttachedTo(t, g, haunch, bear) {
		t.Fatal("the Haunch is not attached")
	}
	if p, tough := effectstest.MkmMPPT(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the equipped creature is %d/%d, want 4/2", p, tough)
	}

	life := me.Life
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, haunch, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("eat the Haunch: %v", err)
	}
	if g.Battlefield.Contains(haunch) {
		t.Fatal("the Haunch was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.FillPoolColored(me, "C", 1)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Dog"); got != 2 {
		t.Errorf("%d Dogs after paying, want 2", got)
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestMkmDL3KrovodHaunchDestroyedAndDeclinedMakesNoDogs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	haunch := effectstest.B12Push(g, me.ID, "Krovod Haunch", "Artifact — Food Equipment", mkmDL3HaunchOracle, 0, 0)
	effectstest.WoeB1Destroy(t, g, haunch)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the Haunch going to the graveyard asked nothing")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Dog"); got != 0 {
		t.Errorf("declining made %d Dogs", got)
	}
}
