package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b1_cards_test.go — Star Trek, batch 1 (#4973). Each test plays
// the real catalog entry and asserts what its text says.

const (
	trekB1LightYearsOracle   = "9356aee1-3d2b-4d17-8863-663e1ddf56f3"
	trekB1AutomatonsOracle   = "2da322e6-a8e2-4d7a-bd34-54a3dfa85940"
	trekB1IncorporatorOracle = "6f0b814b-94e6-48f2-9ccf-c376ca03bba3"
	trekB1AssemblyOracle     = "deb26a1b-2b83-4da2-a2e5-31c17a57eb1e"
	trekB1BoimlerOracle      = "10af9cd9-1700-48f9-97e1-61e239536fef"
	trekB1PicardOracle       = "04624f75-5981-4607-aa8f-706554f826e1"
	trekB1DecisionOracle     = "1c4afb0d-6d6e-4d21-9fe0-5621cd2fd27c"
	trekB1SubadjunctOracle   = "85b67121-340f-4cff-a26d-2160aa923ec3"
	trekB1ConquestOracle     = "35906166-71d1-4089-bbe4-4055edb81fb0"
	trekB1SeriemaOracle      = "a6bcec1f-f515-4e63-9e84-8eb04cc582ff"
)

func TestTrekB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB1LightYearsOracle:   "70,000 Light-Years from Home",
		trekB1AutomatonsOracle:   "Andromedan Automatons",
		trekB1IncorporatorOracle: "Antagonist Incorporator",
		trekB1AssemblyOracle:     "Assembly Specialist",
		trekB1BoimlerOracle:      "Brad Boimler, Eager Ensign",
		trekB1PicardOracle:       "Captain Jean-Luc Picard",
		trekB1DecisionOracle:     "Command Decision",
		trekB1SubadjunctOracle:   "Command Subadjunct",
		trekB1ConquestOracle:     "Conquest Mission",
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

// --- shared helpers ------------------------------------------------------------

// trekB1Library puts a card on top of p's library and returns its id.
func trekB1Library(g *game.Game, p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	})
	return id
}

// trekB1Settle passes priority until the stack is empty or a prompt waits.
func trekB1Settle(t *testing.T, g *game.Game) {
	t.Helper()
	passPriorityAroundTable(t, g)
}

// trekB1Borgs is the Borg tokens p controls.
func trekB1Borgs(t *testing.T, g *game.Game, p uuid.UUID) []game.Card {
	t.Helper()
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Borg" && c.Controller == p && IsToken(c) {
				out = append(out, c)
			}
		}
	})
	return out
}

func trekB1AssertBorg(t *testing.T, c game.Card) {
	t.Helper()
	if !c.IsArtifact() || !c.IsCreature() || !c.HasSubtype("Borg") || len(c.Colors) != 0 || c.Power != 1 || c.Toughness != 1 {
		t.Errorf("Borg token = %s %d/%d %v, want a 1/1 colorless Borg artifact creature", c.TypeLine, c.Power, c.Toughness, c.Colors)
	}
}

// --- 70,000 Light-Years from Home ---------------------------------------------

func TestTrekB1LightYearsExilesAndItsControllerMayFetchABasicToHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	forest := trekB1Library(g, opp, "Forest", "Basic Land — Forest", "")
	trekB1Library(g, opp, "Island", "Basic Land — Island", "")
	trekB1Library(g, opp, "Grizzly", "Creature — Bear", "")
	addPool(t, g, me.ID, "{C}{W}")
	castFromHandForTest(t, g, me, "70,000 Light-Years from Home", "Sorcery", "{1}{W}", trekB1LightYearsOracle,
		game.CastSpellParams{Targets: cardRefs(bear), Strict: true})
	trekB1Settle(t, g)
	if b12ZoneOf(g, bear) != game.ZoneExile {
		t.Fatalf("the creature is in %v, want exile", b12ZoneOf(g, bear))
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Fatal("the caster was asked to search; the search is the exiled permanent's controller's")
	}
	c := searchChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the creature's controller was not offered the search")
	}
	if searchOptionNamed(g, c, "Grizzly") != uuid.Nil {
		t.Error("the search offered a nonland card")
	}
	answerSearchByID(t, g, opp.ID, forest)
	trekB1Settle(t, g)
	if b12ZoneOf(g, forest) != game.ZoneHand {
		t.Errorf("the Forest is in %v, want its owner's hand", b12ZoneOf(g, forest))
	}
}

func TestTrekB1LightYearsTargetsASpacecraftButNotAnotherArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	ship := b27Push(g, opp.ID, "Their Ship", "Artifact — Spacecraft", "", "{3}", 0, 0)
	rock := b27Push(g, opp.ID, "Their Rock", "Artifact", "", "{2}", 0, 0)
	addPool(t, g, me.ID, "{C}{W}")
	if _, err := castFromHandErr(g, me, "70,000 Light-Years from Home", "Sorcery", "{1}{W}", trekB1LightYearsOracle,
		game.CastSpellParams{Targets: cardRefs(rock), Strict: true}); err == nil {
		t.Fatal("targeted a noncreature artifact that is not a Spacecraft")
	}
	castFromHandForTest(t, g, me, "70,000 Light-Years from Home", "Sorcery", "{1}{W}", trekB1LightYearsOracle,
		game.CastSpellParams{Targets: cardRefs(ship), Strict: true})
	trekB1Settle(t, g)
	if b12ZoneOf(g, ship) != game.ZoneExile {
		t.Errorf("the Spacecraft is in %v, want exile", b12ZoneOf(g, ship))
	}
	// The controller may decline the search.
	if c := searchChoiceFor(g, opp.ID); c != nil {
		answerSearchFailToFind(t, g, opp.ID)
	}
	trekB1Settle(t, g)
	if b12ZoneOf(g, rock) != game.ZoneBattlefield {
		t.Error("the other artifact left the battlefield")
	}
}

// --- Andromedan Automatons ------------------------------------------------------

func TestTrekB1AndromedanAutomatonsFetchesAnyNumberOfItsNamesakes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	a := trekB1Library(g, me, "Andromedan Automatons", "Artifact Creature — Android", trekB1AutomatonsOracle)
	b := trekB1Library(g, me, "Andromedan Automatons", "Artifact Creature — Android", trekB1AutomatonsOracle)
	trekB1Library(g, me, "Other Android", "Artifact Creature — Android", "")
	castFromHandForTest(t, g, me, "Andromedan Automatons", "Artifact Creature — Android", "{2}{W}", trekB1AutomatonsOracle, game.CastSpellParams{})
	trekB1Settle(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search was offered when the Automatons entered")
	}
	if searchOptionNamed(g, c, "Other Android") != uuid.Nil {
		t.Error("the search offered a card with another name")
	}
	answerSearchByID(t, g, me.ID, a, b)
	trekB1Settle(t, g)
	if b12ZoneOf(g, a) != game.ZoneHand || b12ZoneOf(g, b) != game.ZoneHand {
		t.Errorf("the namesakes are in %v and %v, want both in hand", b12ZoneOf(g, a), b12ZoneOf(g, b))
	}
}

// --- Assembly Specialist ---------------------------------------------------------

func TestTrekB1AssemblySpecialistMakesABorgOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	castFromHandForTest(t, g, me, "Assembly Specialist", "Artifact Creature — Borg", "{1}{W}", trekB1AssemblyOracle, game.CastSpellParams{})
	trekB1Settle(t, g)
	borgs := trekB1Borgs(t, g, me.ID)
	if len(borgs) != 1 {
		t.Fatalf("%d Borg tokens, want 1", len(borgs))
	}
	trekB1AssertBorg(t, borgs[0])
}

// --- Antagonist Incorporator ------------------------------------------------------

func TestTrekB1AntagonistIncorporatorGrantsArtifactCreaturesAPaidBorg(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Antagonist Incorporator", "Artifact Creature — Borg", trekB1IncorporatorOracle, "{1}{W}", 2, 3, "W")
	drone := b27Push(g, me.ID, "Drone", "Artifact Creature — Construct", "", "{1}", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, drone, bear)
	advanceTo(t, g, game.StepCombatDamage)
	trekB1Settle(t, g)
	// Only the artifact creature connected with the ability: one prompt.
	asked := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == me.ID {
			asked++
		}
	}
	if asked != 1 {
		t.Fatalf("%d prompts to pay {1} for a Borg, want 1 (the Drone's, not the Bear's)", asked)
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	answerPayUnless(t, g, me.ID, true)
	trekB1Settle(t, g)
	borgs := trekB1Borgs(t, g, me.ID)
	if len(borgs) != 1 {
		t.Fatalf("%d Borg tokens after paying, want 1", len(borgs))
	}
	trekB1AssertBorg(t, borgs[0])
}

func TestTrekB1AntagonistIncorporatorDeclinedOrUnpaidMakesNoBorg(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inc := b27Push(g, me.ID, "Antagonist Incorporator", "Artifact Creature — Borg", trekB1IncorporatorOracle, "{1}{W}", 2, 3, "W")
	declareAttack(t, g, opp.ID, inc)
	advanceTo(t, g, game.StepCombatDamage)
	trekB1Settle(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("the Incorporator, an artifact creature itself, has no granted trigger")
	}
	answerPayUnless(t, g, me.ID, false)
	trekB1Settle(t, g)
	if n := len(trekB1Borgs(t, g, me.ID)); n != 0 {
		t.Errorf("declined: %d Borgs, want 0", n)
	}
}

func TestTrekB1AntagonistIncorporatorGrantSkipsAnOpponentsArtifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, opp.ID, "Antagonist Incorporator", "Artifact Creature — Borg", trekB1IncorporatorOracle, "{1}{W}", 2, 3, "W")
	drone := b27Push(g, me.ID, "Drone", "Artifact Creature — Construct", "", "{1}", 1, 1)
	declareAttack(t, g, opp.ID, drone)
	advanceTo(t, g, game.StepCombatDamage)
	trekB1Settle(t, g)
	if hasPayUnlessFor(g, me.ID) || hasPayUnlessFor(g, opp.ID) {
		t.Error("an opponent's Incorporator granted my artifact creature its ability")
	}
}

// --- Brad Boimler, Eager Ensign ----------------------------------------------------

func TestTrekB1BradBoimlerTappedAddsOneMoreCounterThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brad := b27Push(g, me.ID, "Brad Boimler, Eager Ensign", "Legendary Creature — Human Officer", trekB1BoimlerOracle, "{1}{W}", 2, 2, "W")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	assertKeywords(t, g, brad, "lifelink")

	// Before Brad is tapped, nothing extra.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	if got := plusOnes(g, bear); got != 1 {
		t.Fatalf("untapped Brad: %d counters, want 1", got)
	}

	declareAttack(t, g, opp.ID, brad)
	trekB1Settle(t, g)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, game.CounterPlusOne, 2)
		_ = g.AddCounterForEffect(bear, "charge", 1)
		_ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 1)
	})
	if got := plusOnes(g, bear); got != 1+3 {
		t.Errorf("+1/+1 counters on my Bear = %d, want 4 (1 + two plus one)", got)
	}
	if got := countersOn(g, bear, "charge"); got != 2 {
		t.Errorf("charge counters = %d, want 2 (each kind gets one more)", got)
	}
	if got := plusOnes(g, theirs); got != 1 {
		t.Errorf("an opponent's creature got %d counters, want 1", got)
	}

	// Until end of turn only.
	advanceToUpkeepOf(t, g, 1)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	if got := plusOnes(g, bear); got != 5 {
		t.Errorf("next turn: +1/+1 counters = %d, want 5", got)
	}
}

func TestTrekB1BradBoimlerVigilanceAttackDoesNotTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brad := b27Push(g, me.ID, "Brad Boimler, Eager Ensign", "Legendary Creature — Human Officer", trekB1BoimlerOracle, "{1}{W}", 2, 2, "W")
	b27Push(g, me.ID, "Command Subadjunct", "Artifact Creature — Borg", trekB1SubadjunctOracle, "{W}", 2, 1, "W")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	// Brad is no artifact, so the Subadjunct does not stop him tapping:
	// this pins only that an untapped attacker would not count.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	declareAttack(t, g, opp.ID, brad)
	trekB1Settle(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	if got := plusOnes(g, bear); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", got)
	}
}

// --- Captain Jean-Luc Picard ----------------------------------------------------------

func TestTrekB1PicardAnthemsOthersAndMakesOfficersOnCreatureOrSpacecraftCasts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	picard := b27Push(g, me.ID, "Captain Jean-Luc Picard", "Legendary Creature — Human Officer", trekB1PicardOracle, "{2}{W}", 2, 3, "W")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p, tg := b38PT(t, g, picard); p != 2 || tg != 3 {
		t.Errorf("Picard = %d/%d, want 2/3 (\"other\")", p, tg)
	}
	if p, tg := b38PT(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("my Bear = %d/%d, want 3/3", p, tg)
	}
	if p, tg := b38PT(t, g, theirs); p != 2 || tg != 2 {
		t.Errorf("an opponent's Bear = %d/%d, want 2/2", p, tg)
	}

	officers := func() int { return len(tmtB1Named(g, "Officer")) }
	castFromHandForTest(t, g, me, "Grizzly", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{})
	trekB1Settle(t, g)
	if officers() != 1 {
		t.Fatalf("a creature spell made %d Officers, want 1", officers())
	}
	o, _ := g.LookupCardForEffect(tmtB1Named(g, "Officer")[0])
	if !o.IsCreature() || !o.HasSubtype("Officer") || len(o.Colors) != 1 || o.Colors[0] != "R" || o.Power != 1 || o.Toughness != 1 {
		t.Errorf("Officer token = %s %d/%d %v, want a 1/1 red Officer", o.TypeLine, o.Power, o.Toughness, o.Colors)
	}
	castFromHandForTest(t, g, me, "Test Ship", "Artifact — Spacecraft", "{3}", "", game.CastSpellParams{})
	trekB1Settle(t, g)
	if officers() != 2 {
		t.Fatalf("a Spacecraft spell: %d Officers, want 2", officers())
	}
	castFromHandForTest(t, g, me, "Divination", "Sorcery", "{2}{U}", "", game.CastSpellParams{})
	trekB1Settle(t, g)
	if officers() != 2 {
		t.Errorf("a sorcery made an Officer: %d", officers())
	}
}

func TestTrekB1PicardTapsSevenOfficersForAChargedSpacecraft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	picard := b27Push(g, me.ID, "Captain Jean-Luc Picard", "Legendary Creature — Human Officer", trekB1PicardOracle, "{2}{W}", 2, 3, "W")
	var crew []uuid.UUID
	for i := 0; i < 6; i++ {
		crew = append(crew, b12Creature(g, me.ID, "Ensign", "Creature — Human Officer", 1, 1))
	}
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ship := trekB1Library(g, me, "The Seriema", "Legendary Artifact — Spacecraft", trekB1SeriemaOracle)
	trekB1Library(g, me, "Grizzly", "Creature — Bear", "")

	// Six Officers and a Bear are not seven Officers.
	if err := g.ActivateCatalogAbility(me.ID, picard, 0, game.ActivateAbilityParams{TapIDs: append(append([]uuid.UUID(nil), crew...), bear), Strict: true}); err == nil {
		t.Fatal("paid with a non-Officer")
	}
	if err := g.ActivateCatalogAbility(me.ID, picard, 0, game.ActivateAbilityParams{TapIDs: crew, Strict: true}); err == nil {
		t.Fatal("paid with six Officers")
	}
	// Picard is an Officer and may tap for his own cost.
	if err := g.ActivateCatalogAbility(me.ID, picard, 0, game.ActivateAbilityParams{TapIDs: append(append([]uuid.UUID(nil), crew...), picard), Strict: true}); err != nil {
		t.Fatalf("activate with seven Officers: %v", err)
	}
	trekB1Settle(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Grizzly") != uuid.Nil {
		t.Error("the search offered a non-Spacecraft")
	}
	answerSearchByID(t, g, me.ID, ship)
	trekB1Settle(t, g)
	if b12ZoneOf(g, ship) != game.ZoneBattlefield {
		t.Fatalf("the Spacecraft is in %v, want the battlefield", b12ZoneOf(g, ship))
	}
	if got := countersOn(g, ship, "charge"); got != 7 {
		t.Errorf("charge counters = %d, want 7", got)
	}
	if !cardByID(g, ship).IsCreature() {
		t.Error("The Seriema at seven charge counters is not a creature")
	}
}

// --- Command Decision ---------------------------------------------------------------

func TestTrekB1CommandDecisionDestroysPowerFourOrDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	big := b12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 4)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	addPool(t, g, me.ID, "{C}{W}")
	if _, err := castFromHandErr(g, me, "Command Decision", "Sorcery", "{1}{W}", trekB1DecisionOracle,
		game.CastSpellParams{Modes: []int{0}, Targets: cardRefs(small), Strict: true}); err == nil {
		t.Fatal("targeted a creature with power 3")
	}
	castFromHandForTest(t, g, me, "Command Decision", "Sorcery", "{1}{W}", trekB1DecisionOracle,
		game.CastSpellParams{Modes: []int{0}, Targets: cardRefs(big), Strict: true})
	trekB1Settle(t, g)
	if b12ZoneOf(g, big) != game.ZoneGraveyard {
		t.Errorf("the 4-power creature is in %v, want the graveyard", b12ZoneOf(g, big))
	}

	hand := me.Hand.Size()
	addPool(t, g, me.ID, "{C}{W}")
	castFromHandForTest(t, g, me, "Command Decision", "Sorcery", "{1}{W}", trekB1DecisionOracle,
		game.CastSpellParams{Modes: []int{1}, Strict: true})
	trekB1Settle(t, g)
	// The cast card left the hand; the draw put one back.
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d", got, hand+1)
	}
	if b12ZoneOf(g, small) != game.ZoneBattlefield {
		t.Error("the draw mode destroyed something")
	}
}

// --- Command Subadjunct ---------------------------------------------------------------

func TestTrekB1CommandSubadjunctGivesYourArtifactCreaturesVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sub := b27Push(g, me.ID, "Command Subadjunct", "Artifact Creature — Borg", trekB1SubadjunctOracle, "{W}", 2, 1, "W")
	drone := b27Push(g, me.ID, "Drone", "Artifact Creature — Construct", "", "{1}", 1, 1)
	rock := b27Push(g, me.ID, "Rock", "Artifact", "", "{1}", 0, 0)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b27Push(g, opp.ID, "Their Drone", "Artifact Creature — Construct", "", "{1}", 1, 1)
	assertKeywords(t, g, sub, "vigilance")
	assertKeywords(t, g, drone, "vigilance")
	for name, id := range map[string]uuid.UUID{"a noncreature artifact": rock, "a non-artifact creature": bear, "an opponent's artifact creature": theirs} {
		if effectiveAbilitiesContain(t, g, id, "vigilance") {
			t.Errorf("%s has vigilance", name)
		}
	}
	declareAttack(t, g, opp.ID, drone, bear)
	if cardByID(g, drone).Tapped {
		t.Error("the artifact creature tapped to attack")
	}
	if !cardByID(g, bear).Tapped {
		t.Error("the Bear did not tap to attack")
	}
}

// --- Conquest Mission ------------------------------------------------------------------

func TestTrekB1ConquestMissionCountsQuestsAndDoublesAtSix(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mission := b27Push(g, me.ID, "Conquest Mission", "Enchantment", trekB1ConquestOracle, "{2}{W}", 0, 0, "W")
	b24CastNoncreature(t, g, "Sol Ring", "Artifact", "{1}")
	trekB1Settle(t, g)
	if got := countersOn(g, mission, "quest"); got != 1 {
		t.Fatalf("quest counters = %d, want 1", got)
	}
	borgs := trekB1Borgs(t, g, me.ID)
	if len(borgs) != 1 {
		t.Fatalf("%d Borgs after one artifact, want 1 (the Borg token does not chain)", len(borgs))
	}
	trekB1AssertBorg(t, borgs[0])

	// A creature that is not an artifact, and an opponent's artifact, do nothing.
	castFromHandForTest(t, g, me, "Grizzly", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{})
	trekB1Settle(t, g)
	if got := countersOn(g, mission, "quest"); got != 1 {
		t.Errorf("a nonartifact spell added a quest counter: %d", got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TokenCard("1/1 colorless Borg artifact"), 1) })
	trekB1Settle(t, g)

	// Counters 2..5 make one Borg each; the sixth artifact's own counter
	// is already on when the count is read, so it makes two.
	for i := 2; i <= 6; i++ {
		before := len(trekB1Borgs(t, g, me.ID))
		b24CastNoncreature(t, g, "Mind Stone", "Artifact", "{2}")
		trekB1Settle(t, g)
		want := 1
		if i >= 6 {
			want = 2
		}
		if got := len(trekB1Borgs(t, g, me.ID)) - before; got != want {
			t.Errorf("artifact %d: %d Borgs, want %d", i, got, want)
		}
	}
	if got := countersOn(g, mission, "quest"); got != 6 {
		t.Errorf("quest counters = %d, want 6", got)
	}
}
