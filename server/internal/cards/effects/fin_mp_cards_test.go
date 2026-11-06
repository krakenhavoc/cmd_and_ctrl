package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_mp_cards_test.go — Final Fantasy (FIN), cards M–P (#4462). Each
// test plays the real catalog entry and asserts what its text says.

const (
	finMPArsenalOracle  = "e7fff34e-9b80-4a67-a917-fd9efddb5a1f"
	finMPDamperOracle   = "ada82f93-dc93-4a95-a13f-42b5942b3413"
	finMPPotOracle      = "caed9999-74b8-43e3-a319-fa505c04c71c"
	finMPArmorOracle    = "0e9d8e12-b9da-4452-a758-5caffd8a752e"
	finMPInfantryOracle = "563992ff-6fdf-45a0-afc2-7c73c2ec6f74"
	finMPMalboroOracle  = "6804e329-800b-4aa4-9a00-b9a1f4b9537b"
	finMPMatoyaOracle   = "ef6c0054-3a89-4fac-b822-37c7d2514489"
	finMPMinwuOracle    = "31ca9e12-8ad0-472e-8b2a-b7df26370d8c"
	finMPFistOracle     = "1b524149-340e-4f94-9520-6a5258fabb00"
	finMPNamazuOracle   = "3d01c351-6dbd-4d63-9843-3aa1a3c039a6"
	finMPPaladinOracle  = "22076484-e9d5-4f33-8cf3-b9b427436b05"
	finMPTrainOracle    = "a0c4eb78-bee0-42c6-ab1e-64bacde87960"
	finMPPromptoOracle  = "e13d7600-0924-42eb-9d65-aa18548624a7"

	finMPOptOracle      = "713332c1-5bd8-400f-bfff-c1ca0697a043"
	finMPConsiderOracle = "4c9bcba6-87b5-4fb3-97ee-6fe5b739337d"
)

func TestFinMPEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		finMPArsenalOracle:  "Machinist's Arsenal",
		finMPDamperOracle:   "Magic Damper",
		finMPPotOracle:      "Magic Pot",
		finMPArmorOracle:    "Magitek Armor",
		finMPInfantryOracle: "Magitek Infantry",
		finMPMalboroOracle:  "Malboro",
		finMPMatoyaOracle:   "Matoya, Archon Elder",
		finMPMinwuOracle:    "Minwu, White Mage",
		finMPFistOracle:     "Monk's Fist",
		finMPNamazuOracle:   "Namazu Trader",
		finMPPaladinOracle:  "Paladin's Arms",
		finMPTrainOracle:    "Phantom Train",
		finMPPromptoOracle:  "Prompto Argentum",
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

// finMPHeroPT asserts the job-select Hero's power, toughness and added
// creature type.
func finMPHeroPT(t *testing.T, g *game.Game, hero uuid.UUID, power, toughness int, subtype string) {
	t.Helper()
	v := s58p6View(t, g, hero)
	if v.CurrentPower() != power || v.CurrentToughness() != toughness || !v.HasSubtype(subtype) || !v.HasSubtype("Hero") {
		t.Errorf("Hero is %d/%d %v, want a %d/%d Hero %s", v.CurrentPower(), v.CurrentToughness(), v.Effective().Subtypes,
			power, toughness, subtype)
	}
}

// finMPEquipLabel asserts the Equipment's one activated row is equip at
// `cost`.
func finMPEquipLabel(t *testing.T, oracle, cost string) {
	t.Helper()
	spec, _ := Lookup(oracle)
	if len(spec.Activated) != 1 || spec.Activated[0].Label != EquipAbility(cost).Label {
		t.Errorf("activated rows %+v, want one equip %s", spec.Activated, cost)
	}
}

// --- Machinist's Arsenal ------------------------------------------------------------

func TestFinMPMachinistsArsenalGrowsWithYourArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arsenal := castCatalogSpell(t, g, "Machinist's Arsenal", "Artifact — Equipment", finMPArsenalOracle, nil)
	passPriorityAroundTable(t, g)
	hero := ysBHero(t, g, arsenal)
	finMPHeroPT(t, g, hero, 3, 3, "Artificer") // the Arsenal itself is one artifact
	b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	finMPHeroPT(t, g, hero, 3, 3, "Artificer")
	b12Permanent(g, me.ID, "Relic", "Artifact")
	finMPHeroPT(t, g, hero, 5, 5, "Artificer")
	finMPEquipLabel(t, finMPArsenalOracle, "{4}")
}

// --- Magic Damper -------------------------------------------------------------------

func TestFinMPMagicDamperPumpsShieldsAndUntapsYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Magic Damper", "Instant", finMPDamperOracle, finABCard(theirs)); err == nil {
		t.Fatal("Magic Damper targeted a creature you don't control")
	}
	b16Tap(g, bear)
	castCatalogSpell(t, g, "Magic Damper", "Instant", finMPDamperOracle, finABCard(bear))
	passPriorityAroundTable(t, g)
	if v := s58p6View(t, g, bear); v.CurrentPower() != 3 || v.CurrentToughness() != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", v.CurrentPower(), v.CurrentToughness())
	}
	if !hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the Bear has no hexproof")
	}
	if isTapped(g, bear) {
		t.Error("the Bear was not untapped")
	}
}

// --- Magic Pot ----------------------------------------------------------------------

func TestFinMPMagicPotExilesAGraveyardCardAndLeavesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pot := b12Push(g, me.ID, "Magic Pot", "Artifact Creature — Goblin Construct", finMPPotOracle, 1, 4)
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, pot, 0, game.ActivateAbilityParams{Strict: true, Targets: finABCard(dead)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(dead) || !g.Exile.Contains(dead) {
		t.Error("the graveyard card was not exiled")
	}
	if !isTapped(g, pot) {
		t.Error("the Pot did not tap")
	}
	treasures := countTreasures(g, me.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(pot) })
	passPriorityAroundTable(t, g)
	if got := countTreasures(g, me.ID); got != treasures+1 {
		t.Errorf("%d Treasures after the Pot died, want %d", got, treasures+1)
	}
}

// --- Magitek Armor ------------------------------------------------------------------

func TestFinMPMagitekArmorMakesAHeroThatCrewsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	armor := castCatalogSpell(t, g, "Magitek Armor", "Artifact — Vehicle", finMPArmorOracle, nil)
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Hero"); n != 1 {
		t.Fatalf("%d Hero tokens, want 1", n)
	}
	var hero uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Hero" && c.Controller == me.ID {
			hero = c.InstanceID
		}
	}
	if v := s58p6View(t, g, hero); v.CurrentPower() != 1 || v.CurrentToughness() != 1 {
		t.Errorf("the Hero is %d/%d, want 1/1", v.CurrentPower(), v.CurrentToughness())
	}
	if c, ok := battlefieldCard(g, armor); !ok || c.IsCreature() {
		t.Fatal("the uncrewed Armor should be a noncreature Vehicle")
	}
	if err := g.ActivateCatalogAbility(me.ID, armor, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{hero}}); err != nil {
		t.Fatalf("crew 1 with the Hero: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, armor); !ok || !c.IsCreature() {
		t.Error("the crewed Armor is not a creature")
	}
}

// --- Magitek Infantry ---------------------------------------------------------------

func TestFinMPMagitekInfantryGetsPowerWithAnotherArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	infantry := b12Push(g, me.ID, "Magitek Infantry", "Artifact Creature — Robot Soldier", finMPInfantryOracle, 1, 1)
	b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	if p, tt := ptOf(t, g, infantry); p != 1 || tt != 1 {
		t.Fatalf("alone it is %d/%d, want 1/1", p, tt)
	}
	b12Permanent(g, me.ID, "Relic", "Artifact")
	if p, tt := ptOf(t, g, infantry); p != 2 || tt != 1 {
		t.Errorf("with another artifact it is %d/%d, want 2/1", p, tt)
	}
}

func TestFinMPMagitekInfantryFetchesAnotherInfantryTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	infantry := b12Push(g, me.ID, "Magitek Infantry", "Artifact Creature — Robot Soldier", finMPInfantryOracle, 1, 1)
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	other := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Magitek Infantry", TypeLine: "Artifact Creature — Robot Soldier",
		OracleID: finMPInfantryOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	floatForTest(g, me, "WCC")
	if err := g.ActivateCatalogAbility(me.ID, infantry, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Grizzly Bears") != uuid.Nil {
		t.Error("a card with another name is offered")
	}
	answerSearchByID(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, other)
	if !ok || !card.Tapped {
		t.Fatalf("the second Infantry is on the battlefield tapped: present %v, tapped %v", ok, card.Tapped)
	}
	// Each is now "another artifact" for the other.
	if p, _ := ptOf(t, g, infantry); p != 2 {
		t.Errorf("the first Infantry has power %d, want 2", p)
	}
}

// --- Malboro ------------------------------------------------------------------------

func TestFinMPMalboroBadBreathHitsEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	type before struct{ life, hand, library int }
	was := map[uuid.UUID]before{}
	for _, p := range g.Seats[1:] {
		was[p.ID] = before{p.Life, p.Hand.Size(), p.Library.Size()}
	}
	exiled := g.Exile.Size()
	myLife := me.Life
	castRingCreature(t, g, "Malboro", "Creature — Plant Horror", finMPMalboroOracle, 4, 4)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		discardFromHand(t, g, p.ID)
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		w := was[p.ID]
		if p.Life != w.life-2 || p.Hand.Size() != w.hand-1 || p.Library.Size() != w.library-3 {
			t.Errorf("opponent: life %d → %d, hand %d → %d, library %d → %d; want -2 life, one discard, three off the library",
				w.life, p.Life, w.hand, p.Hand.Size(), w.library, p.Library.Size())
		}
	}
	if got := g.Exile.Size() - exiled; got != 9 {
		t.Errorf("%d cards exiled, want three from each of three opponents", got)
	}
	if me.Life != myLife {
		t.Errorf("your life %d → %d, want unchanged", myLife, me.Life)
	}
	spec, _ := Lookup(finMPMalboroOracle)
	if len(spec.Activated) != 1 || spec.Activated[0].Label != Typecycling("Swampcycling", "{2}", "a Swamp card", IsLandWithSubtype("Swamp")).Label {
		t.Errorf("activated rows %+v, want swampcycling {2}", spec.Activated)
	}
}

// --- Matoya, Archon Elder -----------------------------------------------------------

func TestFinMPMatoyaDrawsAfterEachScryAndSurveil(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Matoya, Archon Elder", "Legendary Creature — Human Warlock", finMPMatoyaOracle, 1, 4)
	seedLibrary(me, "A", "B", "C", "D", "E", "F")

	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Opt", "Instant", finMPOptOracle, nil)
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("after Opt: hand %d → %d, want Opt's draw and Matoya's", hand, me.Hand.Size())
	}

	hand = me.Hand.Size()
	castCatalogSpell(t, g, "Consider", "Instant", finMPConsiderOracle, nil)
	passPriorityAroundTable(t, g)
	fraBKeepSurveil(t, g, me)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("after Consider: hand %d → %d, want Consider's draw and Matoya's", hand, me.Hand.Size())
	}

	// An opponent's scry draws you nothing.
	opp := g.Seats[1]
	hand = me.Hand.Size()
	opt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: opt, Name: "Opt", TypeLine: "Instant", OracleID: finMPOptOracle, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, opt, game.CastSpellParams{}); err != nil {
		t.Fatalf("the opponent's Opt: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("an opponent's scry drew you a card: hand %d → %d", hand, me.Hand.Size())
	}
}

// --- Minwu, White Mage --------------------------------------------------------------

func TestFinMPMinwuGrowsEachClericOnLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	minwu := b12Push(g, me.ID, "Minwu, White Mage", "Legendary Creature — Human Cleric", finMPMinwuOracle, 3, 3)
	cleric := b12Creature(g, me.ID, "Healer", "Creature — Elf Cleric", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Healer", "Creature — Human Cleric", 1, 1)
	for _, kw := range []string{"vigilance", "lifelink"} {
		if !hasEffectiveKeyword(t, g, minwu, kw) {
			t.Errorf("Minwu has no %s", kw)
		}
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	passPriorityAroundTable(t, g)
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"Minwu", minwu, 1}, {"your Cleric", cleric, 1}, {"your Bear", bear, 0}, {"their Cleric", theirs, 0}} {
		if got := countersOn(g, tc.id, game.CounterPlusOne); got != tc.want {
			t.Errorf("%s has %d +1/+1 counters, want %d", tc.name, got, tc.want)
		}
	}
}

// --- Monk's Fist --------------------------------------------------------------------

func TestFinMPMonksFistMakesAMonkHero(t *testing.T) {
	g := newCatalogGame(t)
	fist := castCatalogSpell(t, g, "Monk's Fist", "Artifact — Equipment", finMPFistOracle, nil)
	passPriorityAroundTable(t, g)
	finMPHeroPT(t, g, ysBHero(t, g, fist), 2, 1, "Monk")
	finMPEquipLabel(t, finMPFistOracle, "{2}")
}

// --- Namazu Trader ------------------------------------------------------------------

func TestFinMPNamazuTraderLosesOneForATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life, treasures := me.Life, countTreasures(g, me.ID)
	castRingCreature(t, g, "Namazu Trader", "Creature — Fish Citizen", finMPNamazuOracle, 3, 4)
	passPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, me.Life)
	}
	if got := countTreasures(g, me.ID); got != treasures+1 {
		t.Errorf("%d Treasures, want %d", got, treasures+1)
	}
}

func TestFinMPNamazuTraderSacrificesToSurveilTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trader := b12Push(g, me.ID, "Namazu Trader", "Creature — Fish Citizen", finMPNamazuOracle, 3, 4)
	relic := b12Permanent(g, me.ID, "Relic", "Artifact")
	land := b12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	seedLibrary(me, "A", "B", "C")
	declareAttack(t, g, opp.ID, trader)
	passPriorityAroundTable(t, g)
	prompt := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || !hasID(prompt.ChooseCards, relic) ||
		hasID(prompt.ChooseCards, trader) || hasID(prompt.ChooseCards, land) {
		t.Fatalf("want a may-sacrifice over another creature or artifact: %+v", prompt)
	}
	answerOwnPermanents(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact was not sacrificed")
	}
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("want a surveil 2 prompt, got %+v", c)
	}
}

func TestFinMPNamazuTraderDecliningSurveilsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trader := b12Push(g, me.ID, "Namazu Trader", "Creature — Fish Citizen", finMPNamazuOracle, 3, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, trader)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a declined sacrifice sacrificed the creature")
	}
	if c := surveilChoiceFor(g, me.ID); c != nil {
		t.Errorf("declining still surveilled: %+v", c)
	}
}

// --- Paladin's Arms -----------------------------------------------------------------

func TestFinMPPaladinsArmsMakesAWardedKnightHero(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	arms := castCatalogSpell(t, g, "Paladin's Arms", "Artifact — Equipment", finMPPaladinOracle, nil)
	passPriorityAroundTable(t, g)
	hero := ysBHero(t, g, arms)
	finMPHeroPT(t, g, hero, 3, 2, "Knight")
	finMPEquipLabel(t, finMPPaladinOracle, "{4}")

	castAtWardedCreature(t, g, opp, hero)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {1} prompt for the opponent's spell")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hero) {
		t.Error("a declined ward must counter the spell; the Hero should live")
	}
}

// --- Phantom Train ------------------------------------------------------------------

func TestFinMPPhantomTrainSacrificesToGrowAndBecomeASpirit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	train := b12Push(g, me.ID, "Phantom Train", "Artifact — Vehicle", finMPTrainOracle, 4, 4)
	relic := b12Permanent(g, me.ID, "Relic", "Artifact")
	if c, _ := battlefieldCard(g, train); c.IsCreature() {
		t.Fatal("the Train is a creature before its ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, train, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{train}}); err == nil {
		t.Fatal("the Train sacrificed itself to its own ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, train, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{relic}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(relic) {
		t.Fatal("the artifact was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	v := s58p6View(t, g, train)
	if !v.IsCreature() || !v.IsArtifact() || !v.HasSubtype("Spirit") || !v.HasSubtype("Vehicle") {
		t.Errorf("the Train is %v %v, want a Spirit Vehicle artifact creature", v.Effective().Types, v.Effective().Subtypes)
	}
	if v.CurrentPower() != 5 || v.CurrentToughness() != 5 || countersOn(g, train, game.CounterPlusOne) != 1 {
		t.Errorf("the Train is %d/%d, want 5/5 with one +1/+1 counter", v.CurrentPower(), v.CurrentToughness())
	}
	if !hasEffectiveKeyword(t, g, train, "trample") {
		t.Error("the Train has no trample")
	}
}

// --- Prompto Argentum ---------------------------------------------------------------

func TestFinMPPromptoMakesATreasureOnlyOnFourManaNoncreatureSpells(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost, pool string
		want                       int
	}{
		{"three mana", "Sorcery", "{2}{R}", "RRR", 0},
		{"four mana", "Sorcery", "{3}{R}", "RRRR", 1},
		{"four mana creature", "Creature — Bear", "{3}{G}", "GGGG", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			prompto := pushCatalogPermanent(g, me.ID, "Prompto Argentum", "Legendary Creature — Human Scout", finMPPromptoOracle, false)
			if !hasEffectiveKeyword(t, g, prompto, "haste") {
				t.Error("Prompto has no haste")
			}
			advanceToMain(t, g)
			treasures := countTreasures(g, me.ID)
			id := socB17Hand(me, "Test Spell", tc.typeLine, tc.cost, socB17PlainSorceryOracle)
			floatForTest(g, me, tc.pool)
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
				t.Fatalf("cast %s: %v", tc.cost, err)
			}
			passPriorityAroundTable(t, g)
			if got := countTreasures(g, me.ID) - treasures; got != tc.want {
				t.Errorf("%d Treasures made, want %d", got, tc.want)
			}
		})
	}
}
