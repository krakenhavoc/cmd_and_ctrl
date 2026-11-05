package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b20_cards_test.go — Star Trek, batch 20 (#4973). Each test plays the
// real catalog entry through the engine and asserts what its text says.

const (
	trekB20DathonOracle     = "76b7e1ee-43f0-4b56-822a-71b064c9201b"
	trekB20MarcusOracle     = "15885e58-76a9-49fa-ae67-a0482aacd81c"
	trekB20AssaultOracle    = "b093916c-cc0a-4b3b-9de6-8a2db53cdfe0"
	trekB20SupervisorOracle = "d151c8ed-58bf-4b82-a993-1d81f789aa0e"
	trekB20MBengaOracle     = "046c4052-4c3c-4207-b44e-e9a98a8c017f"
	trekB20McCoyOracle      = "0157faee-f307-43c0-b308-bbf36ce02e32"
	trekB20FrontierOracle   = "26f7e903-2a00-4878-b8b7-d5abe35c9526"
	trekB20ContactOracle    = "81229e0a-8413-40cf-8087-9d580cbb21c3"
	trekB20GenesisOracle    = "bb86849c-fdec-4ca5-a8c0-ed682d921759"
	trekB20HoloOracle       = "0d830d0c-3339-49cc-8527-4f4edf787925"
)

func TestTrekB20EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB20DathonOracle:     "Dathon and Picard at El-Adrel",
		trekB20MarcusOracle:     "David and Carol Marcus, Creators",
		trekB20AssaultOracle:    "Dominion Assault Force",
		trekB20SupervisorOracle: "Dominion Supervisor",
		trekB20MBengaOracle:     "Dr. Joseph M'Benga",
		trekB20McCoyOracle:      "Dr. Leonard \"Bones\" McCoy",
		trekB20FrontierOracle:   "Final Frontier",
		trekB20ContactOracle:    "First Contact",
		trekB20GenesisOracle:    "Genesis Device",
		trekB20HoloOracle:       "Holo-Engineer",
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

// trekB20PickMode answers the enters trigger's mode pick.
func trekB20PickMode(t *testing.T, g *game.Game, chooser uuid.UUID, mode int) {
	t.Helper()
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, chooser)
	if c == nil {
		t.Fatalf("no mode pick: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, chooser, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick(%d): %v", mode, err)
	}
}

// trekB20Basic puts a basic Forest onto the battlefield under owner.
func trekB20Basic(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return b12Permanent(g, owner, name, "Basic Land — Forest")
}

// --- Dathon and Picard at El-Adrel -------------------------------------------

func TestTrekB20DathonTembaPumpsAndGivesTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	castModal(t, g, "Dathon and Picard at El-Adrel", "Instant", trekB20DathonOracle, []int{0}, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the target is %d/%d, want 5/5", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the target did not gain trample")
	}
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	if p, _ := effectivePTOf(t, g, bear); p != 2 || hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the pump outlasted the turn")
	}
	_ = me
}

func TestTrekB20DathonShakaEdictsNontokenArtifactsAndEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	token := b12Permanent(g, opp.ID, "Treasure", "Token Artifact — Treasure")
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	aura := b12Permanent(g, opp.ID, "Omen", "Enchantment")
	other := b12Permanent(g, third.ID, "Other Omen", "Enchantment")
	creature := b12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)
	mine := b12Permanent(g, me.ID, "My Rock", "Artifact")
	castModal(t, g, "Dathon and Picard at El-Adrel", "Instant", trekB20DathonOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	c := sacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the opponent with an artifact and an enchantment was not asked to choose")
	}
	if edictOffers(sacrificeOptions(c), token) {
		t.Error("a token was offered")
	}
	answerSacrifice(t, g, opp.ID, aura)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || !g.Battlefield.Contains(rock) || !g.Battlefield.Contains(token) {
		t.Error("the opponent should lose exactly the enchantment they chose")
	}
	if g.Battlefield.Contains(other) || !g.Battlefield.Contains(creature) {
		t.Error("the other opponent should lose their enchantment and keep the creature")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the caster sacrificed")
	}
}

// --- David and Carol Marcus, Creators ----------------------------------------

func TestTrekB20MarcusCreatesATappedNewPlanet(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB2Cast(t, g, "David and Carol Marcus, Creators", "Legendary Creature — Human Scientist", trekB20MarcusOracle, 1, 3)
	passPriorityAroundTable(t, g)
	id := findBattlefieldByName(g, "New Planet")
	if id == uuid.Nil {
		t.Fatal("no New Planet token")
	}
	planet := tmtB1Card(t, g, id)
	if !planet.Tapped || planet.Controller != me.ID || !planet.IsLand() || !planet.IsToken() ||
		!planet.HasSubtype("Planet") || len(planet.Colors) != 0 || planet.IsCreature() {
		t.Errorf("New Planet: tapped %v mine %v land %v token %v planet %v colors %v",
			planet.Tapped, planet.Controller == me.ID, planet.IsLand(), planet.IsToken(), planet.HasSubtype("Planet"), planet.Colors)
	}
	if mana := game.ManaAbilitiesOf(&planet); len(mana) != 1 || mana[0].Produced != "{W|U|B|R|G}" || !mana[0].TapCost {
		t.Errorf("New Planet mana abilities = %+v, want one {T}: any color", mana)
	}
}

// --- Dominion Assault Force ----------------------------------------------------

func TestTrekB20AssaultForceExilesAGraveyardCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Their Corpse", TypeLine: "Creature — Zombie", Owner: opp.ID, Controller: opp.ID})
	})
	force := trekB2Cast(t, g, "Dominion Assault Force", "Creature — Jem'Hadar Warrior", trekB20AssaultOracle, 4, 4)
	assertKeywords(t, g, force, "vigilance")
	trekB20PickMode(t, g, me.ID, 0)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(dead)...)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(dead) || opp.Graveyard.Contains(dead) {
		t.Error("the graveyard card was not exiled")
	}
}

func TestTrekB20AssaultForceSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB2Cast(t, g, "Dominion Assault Force", "Creature — Jem'Hadar Warrior", trekB20AssaultOracle, 4, 4)
	trekB20PickMode(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no surveil prompt: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("surveil looked at %d cards, want 1", len(c.ScryCards))
	}
}

// --- Dominion Supervisor --------------------------------------------------------

func TestTrekB20SupervisorCountersAndTramples(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	trekB2Counter(t, g, theirs, 1)
	sup := trekB2Cast(t, g, "Dominion Supervisor", "Creature — Vorta Doctor", trekB20SupervisorOracle, 3, 1)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(bear)...)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 1 {
		t.Errorf("the target has %d counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("a creature you control with a counter lacks trample")
	}
	if hasEffectiveKeyword(t, g, plain, "trample") || hasEffectiveKeyword(t, g, sup, "trample") {
		t.Error("a creature with no counters has trample")
	}
	if hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature with a counter has trample")
	}
}

// --- Dr. Joseph M'Benga ------------------------------------------------------

func TestTrekB20MBengaDoublesYourCountersAndGivesHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mbenga := b12Push(g, me.ID, "Dr. Joseph M'Benga", "Legendary Creature — Human Doctor", trekB20MBengaOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	trekB2Cast(t, g, "Dominion Supervisor", "Creature — Vorta Doctor", trekB20SupervisorOracle, 3, 1)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(bear)...)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 2 {
		t.Errorf("one +1/+1 counter you put became %d, want 2", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("another creature you control with counters lacks hexproof")
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(me.ID, mbenga, "shield", 1); err != nil {
			t.Fatalf("AddCounterByForEffect: %v", err)
		}
	})
	if got := countersOn(g, mbenga, "shield"); got != 2 {
		t.Errorf("M'Benga himself got %d shield counters, want 2 (he is a creature you control)", got)
	}
	if hasEffectiveKeyword(t, g, mbenga, "hexproof") {
		t.Error("M'Benga gave himself hexproof (other creatures)")
	}
}

func TestTrekB20MBengaLeavesOthersCountersAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Dr. Joseph M'Benga", "Legendary Creature — Human Doctor", trekB20MBengaOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	g.WithWriteLock(func() {
		_ = g.AddCounterByForEffect(opp.ID, bear, game.CounterPlusOne, 1)
		_ = g.AddCounterByForEffect(me.ID, theirs, game.CounterPlusOne, 1)
		_ = g.AddCounterByForEffect(me.ID, rock, "charge", 1)
	})
	if got := plusOnes(g, bear); got != 1 {
		t.Errorf("an opponent's counter on your creature became %d, want 1", got)
	}
	if got := plusOnes(g, theirs); got != 1 {
		t.Errorf("your counter on an opponent's creature became %d, want 1", got)
	}
	if got := countersOn(g, rock, "charge"); got != 1 {
		t.Errorf("your counter on a noncreature artifact became %d, want 1", got)
	}
	if hasEffectiveKeyword(t, g, theirs, "hexproof") {
		t.Error("an opponent's creature with a counter got hexproof")
	}
}

// --- Dr. Leonard "Bones" McCoy ------------------------------------------------

func trekB20CastMcCoy(t *testing.T, g *game.Game, mode int) uuid.UUID {
	t.Helper()
	mccoy := trekB2Cast(t, g, "Dr. Leonard \"Bones\" McCoy", "Legendary Creature — Human Doctor", trekB20McCoyOracle, 3, 2)
	trekB20PickMode(t, g, g.Seats[0].ID, mode)
	return mccoy
}

func TestTrekB20McCoyReturnsANonDoctorPermanentCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := uuid.New()
	doctor := uuid.New()
	sorcery := uuid.New()
	theirs := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: rock, Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: doctor, Name: "Medic", TypeLine: "Creature — Human Doctor", Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: sorcery, Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		g.Seats[1].Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Rock", TypeLine: "Artifact", Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	})
	assertKeywords(t, g, trekB20CastMcCoy(t, g, 0), "flash")
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for name, id := range map[string]uuid.UUID{"a Doctor": doctor, "a sorcery": sorcery, "an opponent's card": theirs} {
		if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(id)); err == nil {
			t.Errorf("McCoy targeted %s", name)
		}
	}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(rock)...)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Error("the permanent card did not return to hand")
	}
}

func TestTrekB20McCoyBringsADeadCreatureBackAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mccoy := trekB20CastMcCoy(t, g, 1)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(mccoy)); err == nil {
		t.Error("McCoy targeted himself (other)")
	}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(theirs)...)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(theirs); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(theirs) {
		t.Fatal("the creature did not die")
	}
	advanceToStepOf(t, g, 0, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Fatal("the creature did not return at the beginning of the next end step")
	}
	back := tmtB1Card(t, g, theirs)
	if back.Controller != opp.ID {
		t.Error("it returned under McCoy's controller, not its owner")
	}
	if got := plusOnes(g, theirs); got != 1 {
		t.Errorf("it returned with %d +1/+1 counters, want 1", got)
	}
}

func TestTrekB20McCoyDoesNothingIfTheCreatureSurvivesTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	trekB20CastMcCoy(t, g, 1)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(bear)...)
	passPriorityAroundTable(t, g)
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	advanceToStepOf(t, g, 1, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a creature that died on a later turn came back")
	}
}

func TestTrekB20McCoyLeavesACardThatLeftTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	trekB20CastMcCoy(t, g, 1)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(bear)...)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(bear); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	advanceToStepOf(t, g, 0, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a card that left the graveyard came back anyway")
	}
}

// --- Final Frontier ---------------------------------------------------------

func TestTrekB20FinalFrontierFetchesAPlanetTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	planet := stapleLibraryCard(me, "Vulcan", "Land — Planet")
	castCatalogSpell(t, g, "Final Frontier", "Sorcery", trekB20FrontierOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, planet)
	if !g.Battlefield.Contains(planet) || !tmtB1Card(t, g, planet).Tapped {
		t.Error("the Planet did not enter tapped")
	}
}

func TestTrekB20FinalFrontierRefusesANonbasicNonPlanetLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	dual := stapleLibraryCard(me, "Temple Garden", "Land — Forest Plains")
	castCatalogSpell(t, g, "Final Frontier", "Sorcery", trekB20FrontierOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	beginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{dual}); err == nil {
		t.Error("the search took a nonbasic land that is not a Planet")
	}
}

func TestTrekB20FinalFrontierBeamsUpFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	spell := seedGraveyardCard(t, g, "Final Frontier", "Sorcery", trekB20FrontierOracle)
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == spell {
				me.Graveyard.Cards[i].ManaCost = "{2}{G}"
			}
		}
	})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	beam := func(ids ...uuid.UUID) error {
		return g.CastSpell(me.ID, spell, game.CastSpellParams{
			Strict: true, FromZone: "graveyard", AlternativeCost: "beam_me_up", AltCostIDs: ids,
		})
	}
	floatMana(t, g, me, "{C}")
	if err := beam(bear); err == nil {
		t.Fatal("beamed up for {1}")
	}
	floatMana(t, g, me, "{G}")
	if err := beam(); err == nil {
		t.Fatal("beamed up without returning a creature")
	}
	if err := beam(bear); err != nil {
		t.Fatalf("beam me up: %v", err)
	}
	if !me.Hand.Contains(bear) {
		t.Error("the creature was not returned as the cost")
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, forest)
	if !g.Battlefield.Contains(forest) {
		t.Error("the basic did not reach the battlefield")
	}
	if !g.Exile.Contains(spell) {
		t.Error("the beamed-up spell was not exiled")
	}
}

// --- First Contact --------------------------------------------------------------

func trekB20CastFirstContact(t *testing.T, g *game.Game) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	toMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := mm2Library(me, mm2Spell("Spell A"), game.Card{Name: "Prize", TypeLine: "Artifact"}, mm2Spell("Spell B"), mm2Spell("Spell C"), mm2Spell("Deep"))
	id := handCardFull(me, "First Contact", "Instant", "", trekB20ContactOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	return id, ids
}

func TestTrekB20FirstContactAsTheFirstSpellGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	_, ids := trekB20CastFirstContact(t, g)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, ids[1])
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the permanent card should be in hand")
	}
	if me.Library.Size() != 1 || !me.Graveyard.Contains(ids[0]) {
		t.Errorf("exactly the top four should be milled (library %d)", me.Library.Size())
	}
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d (first spell this game)", me.Life, life+2)
	}
}

func TestTrekB20FirstContactLaterInTheGameGainsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB2Cast(t, g, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	// A later turn: "this game", not "this turn".
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	advanceToStepOf(t, g, 0, game.StepPrecombatMain)
	life := me.Life
	_, ids := trekB20CastFirstContact(t, g)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, ids[1])
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the permanent card should be in hand")
	}
	if me.Life != life {
		t.Errorf("life = %d, want %d (not the first spell)", me.Life, life)
	}
}

func TestTrekB20FirstContactStaysFirstWhenASpellIsCastInResponse(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	trekB20CastFirstContact(t, g)
	// A second spell cast on top of it is not the first.
	_, second := trekB20CastFirstContact(t, g)
	for i := 0; i < 6; i++ {
		passPriorityAroundTable(t, g)
		if c := chooseCardsChoiceFor(g, me.ID); c != nil {
			answerChooseCards(t, g, me.ID)
		}
	}
	_ = second
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d: only the first spell gains 2", me.Life, life+2)
	}
	if me.FirstSpellCast.ID == uuid.Nil {
		t.Error("the first cast was not recorded")
	}
}

// --- Genesis Device ---------------------------------------------------------

func TestTrekB20GenesisDeviceFetchesOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	planet := stapleLibraryCard(me, "Risa", "Land — Planet")
	trekB2Cast(t, g, "Genesis Device", "Legendary Artifact", trekB20GenesisOracle, 0, 0)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, planet)
	if !g.Battlefield.Contains(planet) || !tmtB1Card(t, g, planet).Tapped {
		t.Error("the Planet did not enter tapped")
	}
}

func TestTrekB20GenesisDeviceSacrificesLandsForThatManyLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	device := b12Push(g, me.ID, "Genesis Device", "Legendary Artifact", trekB20GenesisOracle, 0, 0)
	lands := []uuid.UUID{trekB20Basic(g, me.ID, "Forest A"), trekB20Basic(g, me.ID, "Forest B"), trekB20Basic(g, me.ID, "Forest C")}
	fetch := seedTypedLibrary(me, "Land — Planet", "Planet One", "Planet Two", "Planet Three")
	activate := func() error {
		return g.ActivateCatalogAbility(me.ID, device, 0, game.ActivateAbilityParams{Strict: true})
	}
	// The harness starts in the draw step: not a sorcery-speed window.
	floatMana(t, g, me, "{C}{C}{C}{G}{G}")
	if err := activate(); err == nil {
		t.Fatal("activated outside a main phase")
	}
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{C}{G}")
	if err := activate(); err == nil {
		t.Fatal("activated for {3}{G}")
	}
	floatMana(t, g, me, "{G}")
	if err := activate(); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(device) {
		t.Error("Genesis Device was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	sac := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if sac == nil {
		t.Fatalf("no sacrifice prompt: %+v", g.PendingChoices)
	}
	if sac.ChooseMin != 0 || sac.ChooseMax != 3 {
		t.Errorf("any number: floor %d ceiling %d, want 0 and 3", sac.ChooseMin, sac.ChooseMax)
	}
	if err := g.ResolveOwnPermanents(sac.ID, me.ID, lands[:2]); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	search := searchPromptFor(g, me.ID)
	if search == nil {
		t.Fatalf("no search after the sacrifice: %+v", g.PendingChoices)
	}
	if search.SearchMax != 2 {
		t.Errorf("up to that many: ceiling %d, want 2", search.SearchMax)
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, fetch[:2]); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	for _, id := range fetch[:2] {
		if !g.Battlefield.Contains(id) || !tmtB1Card(t, g, id).Tapped {
			t.Error("a fetched land did not enter tapped")
		}
	}
	if g.Battlefield.Contains(lands[0]) || g.Battlefield.Contains(lands[1]) || !g.Battlefield.Contains(lands[2]) {
		t.Error("the wrong lands were sacrificed")
	}
}

// --- Holo-Engineer ------------------------------------------------------------

func TestTrekB20HoloEngineerMakesAnIllusion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB2Cast(t, g, "Holo-Engineer", "Creature — Human Artificer", trekB20HoloOracle, 3, 3)
	passPriorityAroundTable(t, g)
	id := findBattlefieldByName(g, "Illusion")
	if id == uuid.Nil {
		t.Fatal("no Illusion token")
	}
	c := tmtB1Card(t, g, id)
	if c.Controller != me.ID || !c.IsToken() || !c.IsArtifact() || !c.IsCreature() || len(c.Colors) != 0 || !c.HasSubtype("Illusion") {
		t.Errorf("Illusion = %+v", c)
	}
	if p, tough := effectivePTOf(t, g, id); p != 3 || tough != 3 {
		t.Errorf("Illusion is %d/%d, want 3/3", p, tough)
	}
}

func TestTrekB20HoloEngineerTurnsEveryTokenYouMakeIntoIllusions(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Holo-Engineer", "Creature — Human Artificer", trekB20HoloOracle, 3, 3)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 white Soldier"), 2)
		_ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1)
		_ = g.CreateTokenForEffect(opp.ID, TreasureToken(), 1)
	})
	if n := countBattlefieldNamed(g, me.ID, "Illusion"); n != 3 {
		t.Errorf("two Soldiers and a Treasure became %d Illusions, want 3", n)
	}
	if countBattlefieldNamed(g, me.ID, "Soldier") != 0 || countBattlefieldNamed(g, me.ID, "Treasure") != 0 {
		t.Error("an original token was made anyway")
	}
	if countBattlefieldNamed(g, opp.ID, "Treasure") != 1 || countBattlefieldNamed(g, opp.ID, "Illusion") != 0 {
		t.Error("an opponent's token was replaced")
	}
}
