package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b29_cards_test.go — Star Trek, batch 29 (#4973). Each test plays the
// real catalog entry through the engine.

const (
	trekB29NiVarOracle    = "45bf09bb-4141-4166-bdae-6dac39e7f316"
	trekB29OrionOracle    = "47e3aa22-8806-408f-b93e-0d0b800866f6"
	trekB29QContOracle    = "457cf958-8129-4bf6-b4d5-211eb8c58763"
	trekB29QonosOracle    = "cab718af-8e6f-4766-856e-3a61ba98f94d"
	trekB29QuarksOracle   = "800d0f33-31c9-43da-a649-8860646eca1f"
	trekB29RisaOracle     = "365a4644-74f4-4af9-81ab-0d328ca5edce"
	trekB29RomulusOracle  = "14c4ad24-7f38-409b-8ef0-341f233982da"
	trekB29SpatialOracle  = "a4f31512-a91c-42e0-9929-5069386c3c71"
	trekB29AcademyOracle  = "de8ebef9-fa0e-42df-a1bc-5c4b86d1bdd1"
	trekB29SubspaceOracle = "1a28a83a-de27-4aaa-b35f-1e7075fc1e43"
)

func TestTrekB29EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB29NiVarOracle:    "Ni'Var Science Academy",
		trekB29OrionOracle:    "Orion",
		trekB29QContOracle:    "Q Continuum",
		trekB29QonosOracle:    "Qo'noS",
		trekB29QuarksOracle:   "Quark's Bar",
		trekB29RisaOracle:     "Risa",
		trekB29RomulusOracle:  "Romulus",
		trekB29SpatialOracle:  "Spatial Anomaly",
		trekB29AcademyOracle:  "Starfleet Academy",
		trekB29SubspaceOracle: "Subspace Anomaly",
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

// trekB29PoolOf counts the mana of one colour in a player's pool.
func trekB29PoolOf(p *game.Player, colour string) int {
	n := 0
	for _, tok := range p.ManaPool {
		if tok.Color == colour {
			n += tok.N()
		}
	}
	return n
}

// --- the Planets (gain lands) --------------------------------------------

func TestTrekB29PlanetsEnterTappedGainOneAndTapForTheirPair(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, a, b string
	}{
		{"Orion", trekB29OrionOracle, "R", "G"},
		{"Qo'noS", trekB29QonosOracle, "B", "R"},
		{"Risa", trekB29RisaOracle, "G", "W"},
		{"Romulus", trekB29RomulusOracle, "U", "B"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			before := lifeOf(g, me.ID)
			id := b12PlayFromHand(t, g, tc.name, "Land — Planet", tc.oracle, game.CastSpellParams{})
			if got := lifeOf(g, me.ID); got != before {
				t.Errorf("life %d before the trigger resolved, want %d (a real trigger)", got, before)
			}
			passPriorityAroundTable(t, g)
			if got := lifeOf(g, me.ID); got != before+1 {
				t.Errorf("life = %d, want %d", got, before+1)
			}
			if !b16Tapped(t, g, id) {
				t.Error("the Planet entered untapped")
			}
			untap(g, id)
			if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{tc.b}}); err != nil {
				t.Fatalf("tap for {%s}: %v", tc.b, err)
			}
			if trekB29PoolOf(me, tc.b) != 1 {
				t.Errorf("pool = %v, want one {%s}", me.ManaPool, tc.b)
			}
		})
	}
}

// --- the surveil lands -------------------------------------------------------

// trekB29PlaySurveilLand plays the land with `others` other lands already
// under the active player's control.
func trekB29PlaySurveilLand(t *testing.T, g *game.Game, name, oracle string, others int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < others; i++ {
		seedLandOnBattlefield(g, me.ID, "Wastes", "Basic Land — Wastes")
	}
	return b12PlayFromHand(t, g, name, "Land", oracle, game.CastSpellParams{})
}

func TestTrekB29SurveilLandsEnterTappedWithOneOtherLandAndDoNotTrigger(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Ni'Var Science Academy", trekB29NiVarOracle},
		{"Q Continuum", trekB29QContOracle},
		{"Quark's Bar", trekB29QuarksOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			seedLibrary(me, "Next Draw")
			id := trekB29PlaySurveilLand(t, g, tc.name, tc.oracle, 1)
			passPriorityAroundTable(t, g)
			if !b16Tapped(t, g, id) {
				t.Error("one other land: the land enters tapped")
			}
			if modePickChoiceFor(g, me.ID) != nil || surveilChoiceFor(g, me.ID) != nil || g.Stack.Size() != 0 {
				t.Error("a tapped entry does not trigger")
			}
		})
	}
}

func TestTrekB29SurveilLandUntappedSurveilsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Bin Me")
	id := trekB29PlaySurveilLand(t, g, "Quark's Bar", trekB29QuarksOracle, 2)
	if b16Tapped(t, g, id) {
		t.Fatal("two other lands: Quark's Bar enters untapped")
	}
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("an untapped entry offers the modes")
	}
	answerMode(t, g, c, me.ID, 1)
	passPriorityAroundTable(t, g)
	s := surveilChoiceFor(g, me.ID)
	if s == nil || len(s.ScryCards) != 1 {
		t.Fatalf("surveil 1 prompt = %+v", s)
	}
	if err := g.ResolveSurveil(s.ID, me.ID, s.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if names := graveyardNames(me); len(names) != 1 || names[0] != "Bin Me" {
		t.Errorf("graveyard = %v, want [Bin Me]", names)
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for {R}: %v", err)
	}
	if trekB29PoolOf(me, "R") != 1 {
		t.Errorf("pool = %v, want {R}", me.ManaPool)
	}
}

func TestTrekB29QuarksBarDealsOneToEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := lifeOfOpponents(g)
	mine := lifeOf(g, me.ID)
	trekB29PlaySurveilLand(t, g, "Quark's Bar", trekB29QuarksOracle, 2)
	passPriorityAroundTable(t, g)
	answerMode(t, g, modePickChoiceFor(g, me.ID), me.ID, 0)
	passPriorityAroundTable(t, g)
	after := lifeOfOpponents(g)
	for i := range before {
		if after[i] != before[i]-1 {
			t.Errorf("opponent %d life %d -> %d, want 1 damage", i, before[i], after[i])
		}
	}
	if lifeOf(g, me.ID) != mine {
		t.Error("Quark's Bar damaged its controller")
	}
}

func TestTrekB29NiVarGrantsFlyingUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := trekB29PlaySurveilLand(t, g, "Ni'Var Science Academy", trekB29NiVarOracle, 2)
	passPriorityAroundTable(t, g)
	answerMode(t, g, modePickChoiceFor(g, me.ID), me.ID, 0)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the Bear did not gain flying")
	}
	if c, _ := battlefieldCard(g, bear); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("the Bear is %d/%d; Ni'Var grants no pump", c.CurrentPower(), c.CurrentToughness())
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for {U}: %v", err)
	}
	if trekB29PoolOf(me, "U") != 1 {
		t.Errorf("pool = %v, want {U}", me.ManaPool)
	}
	advanceToNextSeatsTurn(t, g)
	if hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("flying outlasted the turn")
	}
}

func TestTrekB29QContinuumPumpsAndGrantsVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	trekB29PlaySurveilLand(t, g, "Q Continuum", trekB29QContOracle, 3)
	passPriorityAroundTable(t, g)
	answerMode(t, g, modePickChoiceFor(g, me.ID), me.ID, 0)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, bear)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if !hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the Bear did not gain vigilance")
	}
	advanceToNextSeatsTurn(t, g)
	c, _ = battlefieldCard(g, bear)
	if c.CurrentPower() != 2 || hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the pump outlasted the turn")
	}
}

// --- the anomalies -------------------------------------------------------------

func trekB29NewPlanetFor(t *testing.T, g *game.Game, controller uuid.UUID) *game.Card {
	t.Helper()
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "New Planet" && c.Controller == controller {
			return c
		}
	}
	return nil
}

func TestTrekB29SpatialAnomalyShootsACreatureAndMakesANewPlanet(t *testing.T) {
	g, me, opp := trekB16Main(t)
	anomaly := pushCatalogPermanent(g, me.ID, "Spatial Anomaly", "Land", trekB29SpatialOracle, false)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)

	if err := g.ActivateManaAbility(me.ID, anomaly, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for {C}: %v", err)
	}
	if trekB29PoolOf(me, "C") != 1 {
		t.Fatalf("pool = %v, want {C}", me.ManaPool)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	untap(g, anomaly)

	// A pool that can't pay the {R} is refused.
	floatMana(t, g, me, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, anomaly, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{trekB16Ref(bear, 0)}}); err == nil {
		t.Fatal("activated without {R}")
	}
	if !onBattlefield(g, anomaly) {
		t.Fatal("a refused activation sacrificed the land")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })

	floatMana(t, g, me, "{R}{C}{C}{C}{C}")
	b16Activate(t, g, me.ID, anomaly, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{trekB16Ref(bear, 0)}})
	if onBattlefield(g, anomaly) {
		t.Error("the land was not sacrificed")
	}
	if onBattlefield(g, bear) {
		t.Error("the 2/2 survived 2 damage")
	}
	planet := trekB29NewPlanetFor(t, g, me.ID)
	if planet == nil || !planet.Tapped || !planet.HasSubtype("Planet") {
		t.Fatalf("New Planet = %+v, want a tapped Planet", planet)
	}
}

func TestTrekB29SpatialAnomalyNeedsACreatureTarget(t *testing.T) {
	g, me, opp := trekB16Main(t)
	anomaly := pushCatalogPermanent(g, me.ID, "Spatial Anomaly", "Land", trekB29SpatialOracle, false)
	land := pushLandFor(g, opp.ID, "Mountain", "Basic Land — Mountain")
	floatMana(t, g, me, "{R}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, anomaly, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{trekB16Ref(land, 0)}}); err == nil {
		t.Fatal("Spatial Anomaly targeted a land")
	}
	if err := g.ActivateCatalogAbility(me.ID, anomaly, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("Spatial Anomaly targeted a player")
	}
}

// The creature left in response: the ability doesn't resolve (CR 608.2b),
// so no New Planet.
func TestTrekB29SpatialAnomalyFizzlesWithItsTargetGone(t *testing.T) {
	g, me, opp := trekB16Main(t)
	anomaly := pushCatalogPermanent(g, me.ID, "Spatial Anomaly", "Land", trekB29SpatialOracle, false)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{R}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, anomaly, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{trekB16Ref(bear, 0)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	b27Kill(g, bear)
	passPriorityAroundTable(t, g)
	if trekB29NewPlanetFor(t, g, me.ID) != nil {
		t.Error("a fizzled ability still made a New Planet")
	}
}

func TestTrekB29SubspaceAnomalyMakesTwoOfficersAndANewPlanet(t *testing.T) {
	g, me, _ := trekB16Main(t)
	anomaly := pushCatalogPermanent(g, me.ID, "Subspace Anomaly", "Land", trekB29SubspaceOracle, false)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, anomaly, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without {W}")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatMana(t, g, me, "{W}{C}{C}{C}{C}")
	b16Activate(t, g, me.ID, anomaly, 0, game.ActivateAbilityParams{})
	if onBattlefield(g, anomaly) {
		t.Error("the land was not sacrificed")
	}
	officers := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Officer" && c.Controller == me.ID {
			officers++
			if !c.IsCreature() || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "R" || !c.HasSubtype("Officer") {
				t.Errorf("Officer = %+v, want a 1/1 red Officer creature", c)
			}
		}
	}
	if officers != 2 {
		t.Errorf("%d Officers, want 2", officers)
	}
	planet := trekB29NewPlanetFor(t, g, me.ID)
	if planet == nil || !planet.Tapped {
		t.Fatalf("New Planet = %+v, want a tapped one", planet)
	}
	mana := game.ManaAbilitiesForCard(*planet)
	if len(mana) != 1 || mana[0].Produced != "{W|U|B|R|G}" {
		t.Errorf("New Planet mana = %+v", mana)
	}
}

// --- Starfleet Academy ----------------------------------------------------------

func TestTrekB29StarfleetAcademyCountsCreatureTypesAmongNonBorg(t *testing.T) {
	g, me, opp := trekB16Main(t)
	academy := pushCatalogPermanent(g, me.ID, "Starfleet Academy", "Legendary Land", trekB29AcademyOracle, false)
	b12Creature(g, me.ID, "Vulcan Scientist", "Creature — Vulcan Scientist", 1, 1)
	b12Creature(g, me.ID, "Human Scientist", "Creature — Human Scientist", 1, 1)
	b12Creature(g, me.ID, "Drone", "Creature — Human Borg", 1, 1)       // a Borg: nothing counts
	b12Creature(g, opp.ID, "Their Klingon", "Creature — Klingon", 1, 1) // not mine
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Shapeshifter", TypeLine: "Creature — Shapeshifter",
		Keywords: []string{"changeling"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID}) // every type, so a Borg

	if err := g.ActivateManaAbility(me.ID, academy, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for {C}: %v", err)
	}
	if trekB29PoolOf(me, "C") != 1 {
		t.Fatalf("pool = %v, want {C}", me.ManaPool)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	untap(g, academy)

	// Nothing to pay the {5} with: refused.
	if err := g.ActivateManaAbility(me.ID, academy, 1, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Fatal("the Federation ability activated without {5}")
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{C}")
	if err := g.ActivateManaAbility(me.ID, academy, 1, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("Federation: %v", err)
	}
	// Vulcan, Scientist, Human: three types.
	if got := trekB29PoolOf(me, "G"); got != 3 {
		t.Errorf("{G} added = %d, want 3 (Vulcan, Human, Scientist)", got)
	}
	if trekB29PoolOf(me, "C") != 0 || !b16Tapped(t, g, academy) {
		t.Error("the {5} was not paid or the land did not tap")
	}
}
