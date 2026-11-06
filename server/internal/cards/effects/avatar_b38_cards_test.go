package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b38_cards_test.go — Avatar: The Last Airbender Eternal, batches
// 38 to 40 (#5075). Each test plays the real catalog entry through the
// engine.

const (
	avB38ZukoOracle         = "d445a584-98e7-4501-bb20-e76c9cafa757"
	avB38BumiOracle         = "301af2b2-e2f7-4220-a3f5-7702fb62664a"
	avB38StudentOracle      = "8f8fd2ba-df0e-49bd-a916-cc4f50b9fa24"
	avB38EelHoundsOracle    = "25432a3a-bcfd-4c43-b662-07676e9b3fdd"
	avB38FrogSquirrelOracle = "b8563788-debc-4f97-adf4-b186bf3960e2"
	avB38HippoCowsOracle    = "f66fca2b-6c44-4079-b29e-4d18ef030a0c"
	avB38MatchOddsOracle    = "797d8776-9dcf-45b1-8c5e-1d025f916773"
	avB38TutelageOracle     = "7ece47b1-0b4b-40ac-a90f-4573a8e3b8bd"
	avB38RampageOracle      = "9e5df175-5f91-47f7-957c-63bac31f0858"
	avB38GliderOracle       = "d53d1805-c4ec-4b42-a226-3cd78cd2be84"
	avB38AangOracle         = "e46b2ac6-97d6-48f2-aba3-cd92f8e092c6"
	avB38AangDefenseOracle  = "92b617f9-6e6b-42a0-9c45-6ff5f6193051"
	avB38SlothOracle        = "1c385880-3eea-45c8-8e89-330ad8ad92b1"
	avB38PathOracle         = "ac94860b-70b8-4624-9758-4bb024a56f60"
	avB38RazorRingsOracle   = "eb5afbd7-4d21-4449-9204-eeb4ebf6a72a"
	avB38DragonMooseOracle  = "4a678c6f-5a93-4a3b-80fd-4c0d0f9040c9"
	avB38ShotOracle         = "a6081d17-599b-4fd2-bd68-55eca939213a"
	avB38FNSoldierOracle    = "6a15fad7-8171-4bce-bb39-fe3a8175f8fa"
	avB38ConquestOracle     = "8893af49-2f90-4aed-bcaf-905bb9e8ed8f"
	avB38KomodoOracle       = "7fe5cad0-8db8-4cb3-9d61-bda84477654d"
	avB38ZhaoOracle         = "92ce3e30-dde3-43bb-a3d6-ccd91ff18bc3"
	avB38OffenseOracle      = "90a2ec9c-2f1c-4ed4-820d-dfdb6b58c0db"
)

func TestAvatarB38EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB38ZukoOracle:         "Zuko, Avatar Hunter",
		avB38BumiOracle:         "Bumi, Eclectic Earthbender",
		avB38StudentOracle:      "Earthbending Student",
		avB38EelHoundsOracle:    "Eel-Hounds",
		avB38FrogSquirrelOracle: "Frog-Squirrels",
		avB38HippoCowsOracle:    "Hippo-Cows",
		avB38MatchOddsOracle:    "Match the Odds",
		avB38TutelageOracle:     "Seismic Tutelage",
		avB38RampageOracle:      "Hog-Monkey Rampage",
		avB38GliderOracle:       "Mechanical Glider",
		avB38AangOracle:         "Aang, Air Nomad",
		avB38AangDefenseOracle:  "Aang's Defense",
		avB38SlothOracle:        "Aardvark Sloth",
		avB38PathOracle:         "Path to Redemption",
		avB38RazorRingsOracle:   "Razor Rings",
		avB38DragonMooseOracle:  "Dragon Moose",
		avB38ShotOracle:         "Explosive Shot",
		avB38FNSoldierOracle:    "Fire Nation Soldier",
		avB38ConquestOracle:     "Fire Nation's Conquest",
		avB38KomodoOracle:       "Komodo Rhino",
		avB38ZhaoOracle:         "Zhao, the Seething Flame",
		avB38OffenseOracle:      "Zuko's Offense",
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

// --- keyword-only creatures ----------------------------------------------------

func TestAvatarB38KeywordCreaturesHaveTheirKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, c := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Frog-Squirrels", "Creature — Frog Squirrel", avB38FrogSquirrelOracle, []string{"reach"}},
		{"Hippo-Cows", "Creature — Hippo Ox", avB38HippoCowsOracle, []string{"trample"}},
		{"Aardvark Sloth", "Creature — Sloth Beast", avB38SlothOracle, []string{"lifelink"}},
		{"Dragon Moose", "Creature — Dragon Elk", avB38DragonMooseOracle, []string{"haste"}},
		{"Fire Nation Soldier", "Creature — Human Soldier", avB38FNSoldierOracle, []string{"haste"}},
		{"Komodo Rhino", "Creature — Lizard Rhino", avB38KomodoOracle, []string{"trample"}},
		{"Zhao, the Seething Flame", "Legendary Creature — Human Soldier", avB38ZhaoOracle, []string{"menace"}},
		{"Zuko, Avatar Hunter", "Legendary Creature — Human Noble", avB38ZukoOracle, []string{"reach"}},
		{"Eel-Hounds", "Creature — Fish Dog", avB38EelHoundsOracle, []string{"trample"}},
		{"Aang, Air Nomad", "Legendary Creature — Human Avatar Ally", avB38AangOracle, []string{"flying", "vigilance"}},
	} {
		id := pushCatalogPermanent(g, me.ID, c.name, c.typeLine, c.oracle, false)
		assertKeywords(t, g, id, c.keywords...)
	}
}

func TestAvatarB38DragonMooseAttacksTheTurnItArrives(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	moose := pushCatalogPermanent(g, me.ID, "Dragon Moose", "Creature — Dragon Elk", avB38DragonMooseOracle, true)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(moose, opp.ID); err != nil {
		t.Fatalf("a summoning-sick creature with haste could not attack: %v", err)
	}
}

// --- Zuko, Avatar Hunter -------------------------------------------------------

func TestAvatarB38ZukoMakesASoldierForEachRedSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Zuko, Avatar Hunter", "Legendary Creature — Human Noble", avB38ZukoOracle, 4, 5)
	floatForTest(g, me, "RG")
	castFromHandForTest(t, g, me, "Red Spell", "Instant", "{R}", "", strict)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Soldier"); got != 1 {
		t.Fatalf("%d Soldiers after a red spell, want 1", got)
	}
	soldier := tmtB1Card(t, g, tmtB1Named(g, "Soldier")[0])
	if soldier.Power != 2 || soldier.Toughness != 2 || len(soldier.Colors) != 1 || soldier.Colors[0] != "R" ||
		soldier.TypeLine != "Token Creature — Soldier" {
		t.Errorf("Soldier = %d/%d %v %q, want a 2/2 red Soldier", soldier.Power, soldier.Toughness, soldier.Colors, soldier.TypeLine)
	}
	castFromHandForTest(t, g, me, "Green Spell", "Instant", "{G}", "", strict)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Soldier"); got != 1 {
		t.Errorf("%d Soldiers after a green spell, want still 1", got)
	}
}

// --- Bumi, Eclectic Earthbender -------------------------------------------------

func TestAvatarB38BumiEarthbendsOneOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := avB14Forest(g, me.ID)
	avB14Enter(t, g, me, game.Card{Name: "Bumi, Eclectic Earthbender", TypeLine: "Legendary Creature — Human Noble Ally",
		OracleID: avB38BumiOracle, Power: 4, Toughness: 4})
	pickTriggerTarget(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	avB14Earthbent(t, g, land, 1)
}

func TestAvatarB38BumiAttackGrowsEachLandCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bumi := b12Push(g, me.ID, "Bumi, Eclectic Earthbender", "Legendary Creature — Human Noble Ally", avB38BumiOracle, 4, 4)
	arbor := b12Creature(g, me.ID, "Arbor", "Land Creature — Forest Dryad", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	forest := avB14Forest(g, me.ID)
	theirArbor := b12Creature(g, opp.ID, "Their Arbor", "Land Creature — Forest Dryad", 1, 1)

	declareAttack(t, g, opp.ID, bumi)
	passPriorityAroundTable(t, g)
	for name, want := range map[string]struct {
		id uuid.UUID
		n  int
	}{
		"your land creature": {arbor, 2}, "Bumi": {bumi, 0}, "a creature that isn't a land": {bear, 0},
		"a land that isn't a creature": {forest, 0}, "an opponent's land creature": {theirArbor, 0},
	} {
		if got := plusOnes(g, want.id); got != want.n {
			t.Errorf("%s has %d +1/+1 counters, want %d", name, got, want.n)
		}
	}
}

// --- Earthbending Student ------------------------------------------------------

func TestAvatarB38StudentEarthbendsTwoAndGivesLandCreaturesVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := avB14Forest(g, me.ID)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirArbor := b12Creature(g, opp.ID, "Their Arbor", "Land Creature — Forest Dryad", 1, 1)
	student := avB14Enter(t, g, me, game.Card{Name: "Earthbending Student", TypeLine: "Creature — Human Warrior Ally",
		OracleID: avB38StudentOracle, Power: 1, Toughness: 3})
	pickTriggerTarget(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	avB14Earthbent(t, g, land, 2)

	if !hasEffectiveKeyword(t, g, land, "vigilance") {
		t.Error("the earthbent land creature has no vigilance")
	}
	for name, id := range map[string]uuid.UUID{"a creature that isn't a land": bear, "the Student": student, "an opponent's land creature": theirArbor} {
		if hasEffectiveKeyword(t, g, id, "vigilance") {
			t.Errorf("%s was given vigilance", name)
		}
	}

	// The earthbent land has haste and vigilance: it attacks untapped.
	declareAttack(t, g, opp.ID, land)
	if avB1Card(t, g, land).Tapped {
		t.Error("the land creature tapped to attack despite vigilance")
	}
}

// --- Eel-Hounds ------------------------------------------------------------------

func TestAvatarB38EelHoundsPumpsAnotherCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eel := b12Push(g, me.ID, "Eel-Hounds", "Creature — Fish Dog", avB38EelHoundsOracle, 4, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, eel)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the attack trigger")
	}
	if hasID(p.PickTargetCards, eel) || hasID(p.PickTargetCards, theirs) || !hasID(p.PickTargetCards, bear) {
		t.Fatalf("the trigger targets another creature you control: %+v", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if pw, tg := effectivePTOf(t, g, bear); pw != 4 || tg != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", pw, tg)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear did not gain trample")
	}
	if pw, _ := effectivePTOf(t, g, eel); pw != 4 {
		t.Errorf("Eel-Hounds pumped itself to power %d", pw)
	}
}

// --- Match the Odds --------------------------------------------------------------

func TestAvatarB38MatchTheOddsCountsOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Other Bear", "Creature — Bear", 2, 2)
	b12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Match the Odds", "Sorcery — Lesson", avB38MatchOddsOracle, nil)
	passPriorityAroundTable(t, g)
	allies := tmtB1Named(g, "Ally")
	if len(allies) != 1 {
		t.Fatalf("%d Ally tokens, want 1", len(allies))
	}
	ally := tmtB1Card(t, g, allies[0])
	if ally.Controller != me.ID || ally.Power != 1 || ally.Toughness != 1 || len(ally.Colors) != 1 || ally.Colors[0] != "W" {
		t.Errorf("Ally = %d/%d %v, want my white 1/1", ally.Power, ally.Toughness, ally.Colors)
	}
	if got := plusOnes(g, allies[0]); got != 3 {
		t.Errorf("the Ally has %d counters, want 3 (one per opponent's creature)", got)
	}
}

func TestAvatarB38MatchTheOddsWithNoOpposingCreaturesIsAOneOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Match the Odds", "Sorcery — Lesson", avB38MatchOddsOracle, nil)
	passPriorityAroundTable(t, g)
	allies := tmtB1Named(g, "Ally")
	if len(allies) != 1 || plusOnes(g, allies[0]) != 0 {
		t.Fatalf("allies %v; want one with no counters", allies)
	}
}

// --- Seismic Tutelage ------------------------------------------------------------

func TestAvatarB38SeismicTutelageCountersThenDoublesOnEachAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Seismic Tutelage", "Enchantment — Aura", avB38TutelageOracle, cardRefs(host))
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, host); got != 1 {
		t.Fatalf("enters trigger: %d counters, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(host, game.CounterPlusOne, 1) })
	declareAttack(t, g, opp.ID, host, other)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, host); got != 4 {
		t.Errorf("after attacking: %d counters, want 4 (doubled from 2)", got)
	}
	if got := plusOnes(g, other); got != 0 {
		t.Errorf("a creature the Aura isn't on got %d counters", got)
	}
}

// --- Hog-Monkey Rampage ----------------------------------------------------------

func TestAvatarB38HogMonkeyRampageCountersABigCreatureThenFights(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 4, 7)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 6)
	castCatalogSpell(t, g, "Hog-Monkey Rampage", "Instant", avB38RampageOracle, cardRefs(mine, theirs))
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, mine); got != 1 {
		t.Fatalf("a power-4 creature got %d counters, want 1", got)
	}
	if got := damageOn(t, g, theirs); got != 5 {
		t.Errorf("their creature took %d, want 5 (the countered 5/8)", got)
	}
	if got := damageOn(t, g, mine); got != 5 {
		t.Errorf("my creature took %d, want 5", got)
	}
}

func TestAvatarB38HogMonkeyRampageSmallCreatureFightsWithoutACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 5)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 2, 6)
	castCatalogSpell(t, g, "Hog-Monkey Rampage", "Instant", avB38RampageOracle, cardRefs(mine, theirs))
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, mine); got != 0 {
		t.Errorf("a power-3 creature got %d counters", got)
	}
	if got := damageOn(t, g, theirs); got != 3 {
		t.Errorf("their creature took %d, want 3", got)
	}
}

func TestAvatarB38HogMonkeyRampageRefusesTwoOfYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	floatForTest(g, me, "CR")
	if _, err := castFromHandErr(g, me, "Hog-Monkey Rampage", "Instant", "{1}{R/G}", avB38RampageOracle,
		game.CastSpellParams{Strict: true, Targets: cardRefs(a, b)}); err == nil {
		t.Fatal("the second target was a creature you control; it must be an opponent's")
	}
}

// --- Mechanical Glider -------------------------------------------------------------

func TestAvatarB38MechanicalGliderAttachesOnEntryAndGivesFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := seedBear(g, me.ID)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	glider := castCatalogSpell(t, g, "Mechanical Glider", "Artifact — Equipment", avB38GliderOracle, nil)
	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); hasID(p.PickTargetCards, theirs) || !hasID(p.PickTargetCards, bear) {
		t.Fatalf("the entry trigger targets a creature you control: %+v", p)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, glider); host.ID != bear {
		t.Fatalf("the Glider did not attach to the Bear: %+v", host)
	}
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the equipped creature has no flying")
	}
	if pw, tg := effectivePTOf(t, g, bear); pw != 2 || tg != 2 {
		t.Errorf("the Glider changed P/T to %d/%d", pw, tg)
	}

	// Equip {2}: one mana is not enough, two moves it.
	b09AddColorless(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, glider, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(other)}); err == nil {
		t.Fatal("equipped for one mana; the cost is {2}")
	}
	b09AddColorless(me, 1)
	b16Activate(t, g, me.ID, glider, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(other)})
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, glider); host.ID != other {
		t.Fatalf("equip did not move the Glider: %+v", host)
	}
	if hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the Bear kept flying after the Glider moved")
	}
}

// --- Aang, Air Nomad -------------------------------------------------------------

func TestAvatarB38AangGivesOtherCreaturesYouControlVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Aang, Air Nomad", "Legendary Creature — Human Avatar Ally", avB38AangOracle, 5, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("your other creature has no vigilance")
	}
	if hasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Error("an opponent's creature was given vigilance")
	}
	declareAttack(t, g, opp.ID, bear)
	if avB1Card(t, g, bear).Tapped {
		t.Error("the Bear tapped to attack despite vigilance")
	}
}

// --- Aang's Defense --------------------------------------------------------------

func TestAvatarB38AangsDefensePumpsYourBlockerAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	crtBlock(g, blocker, attacker, me.ID)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Aang's Defense", "Instant", avB38AangDefenseOracle, b16TargetCard(blocker))
	passPriorityAroundTable(t, g)
	if pw, tg := effectivePTOf(t, g, blocker); pw != 4 || tg != 4 {
		t.Errorf("blocker is %d/%d, want 4/4", pw, tg)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, got)
	}
}

func TestAvatarB38AangsDefenseRefusesAnOpponentsBlockerOrANonBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "My Attacker", "Creature — Bear", 2, 2)
	theirBlocker := b12Creature(g, opp.ID, "Their Blocker", "Creature — Bear", 2, 2)
	crtBlock(g, theirBlocker, attacker, opp.ID)
	idle := b12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	for name, id := range map[string]uuid.UUID{"an opponent's blocking creature": theirBlocker, "a creature that isn't blocking": idle} {
		floatForTest(g, me, "W")
		if _, err := castFromHandErr(g, me, "Aang's Defense", "Instant", "{W}", avB38AangDefenseOracle,
			game.CastSpellParams{Strict: true, Targets: b16TargetCard(id)}); err == nil {
			t.Errorf("%s was accepted as the target", name)
		}
	}
}

// --- Path to Redemption ----------------------------------------------------------

func avB38PathOn(g *game.Game, owner, host uuid.UUID) uuid.UUID {
	aura := pushCatalogPermanent(g, owner, "Path to Redemption", "Enchantment — Aura", avB38PathOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == aura {
				g.Battlefield.Cards[i].AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
			}
		}
	})
	return aura
}

func TestAvatarB38PathToRedemptionStopsAttackingAndBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	avB38PathOn(g, me.ID, victim)
	assertRestrictions(t, g, victim, game.CantAttackOrBlock)
}

func TestAvatarB38PathToRedemptionExilesAndMakesAnAllyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	aura := avB38PathOn(g, me.ID, victim)

	// Not sorcery speed: any step of your own turn.
	advanceTo(t, g, game.StepBeginCombat)
	b09AddColorless(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for four mana; the cost is {5}")
	}
	b09AddColorless(me, 1)
	b16Activate(t, g, me.ID, aura, 0, game.ActivateAbilityParams{Strict: true})
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(victim) {
		t.Error("the enchanted creature was not exiled")
	}
	allies := tmtB1Named(g, "Ally")
	if len(allies) != 1 {
		t.Fatalf("%d Ally tokens, want 1", len(allies))
	}
	if ally := tmtB1Card(t, g, allies[0]); ally.Controller != me.ID || ally.Power != 1 || ally.Toughness != 1 {
		t.Errorf("Ally = %+v, want the activator's 1/1", ally)
	}
}

func TestAvatarB38PathToRedemptionRefusesOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := b12Creature(g, active.ID, "Ogre", "Creature — Ogre", 3, 3)
	aura := avB38PathOn(g, other.ID, victim)
	b09AddColorless(other, 5)
	if err := g.ActivateCatalogAbility(other.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("Path to Redemption was activated during an opponent's turn")
	}
	if !g.Battlefield.Contains(victim) {
		t.Error("the creature left the battlefield")
	}
}

// --- Razor Rings -----------------------------------------------------------------

func TestAvatarB38RazorRingsGainsTheExcess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	blocker := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	crtBlock(g, blocker, attacker, me.ID)
	life := lifeOf(g, me.ID)
	castCatalogSpell(t, g, "Razor Rings", "Instant", avB38RazorRingsOracle, b16TargetCard(attacker))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the 1/1 survived 4 damage")
	}
	if got := lifeOf(g, me.ID); got != life+3 {
		t.Errorf("life %d -> %d, want +3 (4 damage to a 1/1)", life, got)
	}
}

func TestAvatarB38RazorRingsCountsMarkedDamageAndGainsNothingWithoutExcess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// A 3/3 already dealt 1: lethal is 2, so the excess is 2.
	attacker := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	big := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6)
	crtBlock(g, uuid.Nil, attacker, me.ID)
	crtBlock(g, uuid.Nil, big, me.ID)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == attacker {
				g.Battlefield.Cards[i].DamageMarked = 1
			}
		}
	})
	life := lifeOf(g, me.ID)
	castCatalogSpell(t, g, "Razor Rings", "Instant", avB38RazorRingsOracle, b16TargetCard(attacker))
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != life+2 {
		t.Errorf("life %d -> %d, want +2", life, got)
	}
	castCatalogSpell(t, g, "Razor Rings", "Instant", avB38RazorRingsOracle, b16TargetCard(big))
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != life+2 {
		t.Errorf("4 damage to a 6/6 changed life to %d", got)
	}
	if got := damageOn(t, g, big); got != 4 {
		t.Errorf("the 6/6 took %d, want 4", got)
	}
}

func TestAvatarB38RazorRingsRefusesACreatureNotInCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	idle := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	floatForTest(g, me, "CW")
	if _, err := castFromHandErr(g, me, "Razor Rings", "Instant", "{1}{W}", avB38RazorRingsOracle,
		game.CastSpellParams{Strict: true, Targets: b16TargetCard(idle)}); err == nil {
		t.Fatal("a creature that is neither attacking nor blocking was accepted")
	}
}

// --- Explosive Shot / Zuko's Offense -----------------------------------------------

func TestAvatarB38ExplosiveShotDealsFourToACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	castCatalogSpell(t, g, "Explosive Shot", "Sorcery", avB38ShotOracle, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, theirs); got != 4 {
		t.Errorf("the creature took %d, want 4", got)
	}
}

func TestAvatarB38ExplosiveShotRefusesAPlayerAndInstantSpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	floatForTest(g, me, "CR")
	if _, err := castFromHandErr(g, me, "Explosive Shot", "Sorcery", "{1}{R}", avB38ShotOracle,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("a player was accepted as the target creature")
	}
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	floatForTest(g, me, "CR")
	if _, err := castFromHandErr(g, me, "Explosive Shot", "Sorcery", "{1}{R}", avB38ShotOracle,
		game.CastSpellParams{Strict: true, Targets: b16TargetCard(theirs)}); err == nil {
		t.Error("a sorcery was cast in the beginning of combat step")
	}
}

func TestAvatarB38ZukosOffenseDealsTwoToAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	castCatalogSpell(t, g, "Zuko's Offense", "Sorcery", avB38OffenseOracle, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, theirs); got != 2 {
		t.Errorf("the creature took %d, want 2", got)
	}
	life := lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Zuko's Offense", "Sorcery", avB38OffenseOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-2 {
		t.Errorf("the player's life %d -> %d, want -2", life, got)
	}
}

// --- Fire Nation's Conquest ----------------------------------------------------------

func TestAvatarB38FireNationsConquestPumpsYourCreaturesPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Fire Nation's Conquest", "Enchantment", avB38ConquestOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if pw, tg := effectivePTOf(t, g, bear); pw != 3 || tg != 2 {
		t.Errorf("your creature is %d/%d, want 3/2", pw, tg)
	}
	if pw, tg := effectivePTOf(t, g, theirs); pw != 2 || tg != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", pw, tg)
	}
	late := b12Creature(g, me.ID, "Late Bear", "Creature — Bear", 1, 1)
	if pw, _ := effectivePTOf(t, g, late); pw != 2 {
		t.Errorf("a creature arriving later has power %d, want 2", pw)
	}
}
