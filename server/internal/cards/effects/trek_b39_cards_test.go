package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b39_cards_test.go — Star Trek Commander, batch 39 (#4974). Each test
// plays the real catalog entry through the engine and asserts what its text
// says.

const (
	trekB39GorkonOracle   = "24d194d3-f83a-496b-921a-77c09c68b0e1"
	trekB39KyleOracle     = "02583549-b618-4d41-9f46-209a99e30cd7"
	trekB39ScottOracle    = "07f9c121-8a12-4f28-ae89-ec71f073b9ba"
	trekB39TroiOracle     = "9a52cb6b-b0d8-4649-b65a-5f9a9d2a7eb2"
	trekB39OverrideOracle = "75af4154-d873-45d0-9ff2-c977169012a3"
	trekB39RikerOracle    = "cef73cb5-88c2-43cd-8648-1baca085b8a4"
	trekB39CurzonOracle   = "035bee4c-6f79-4123-9b6c-feba190052e8"
	trekB39DeliverOracle  = "edea104c-b563-4fae-9e74-82164a8e090b"
	trekB39HughOracle     = "9964f2fe-b7f4-40f5-8e71-948756d27f0a"
	trekB39JuratiOracle   = "4dc4c8a3-17d6-4c2d-bcbf-9cebf261eda6"
)

func TestTrekB39EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB39GorkonOracle:   "Chancellor Gorkon",
		trekB39KyleOracle:     "Chief Kyle, Transporter Tech",
		trekB39ScottOracle:    "Chief Montgomery Scott",
		trekB39TroiOracle:     "Cmdr. Troi, Outgoing Empath",
		trekB39OverrideOracle: "Collective Override",
		trekB39RikerOracle:    "Commander William Riker",
		trekB39CurzonOracle:   "Curzon Dax, Bound by Blood",
		trekB39DeliverOracle:  "Deliver from Chaos",
		trekB39HughOracle:     "Hugh, First of Many",
		trekB39JuratiOracle:   "Jurati, Cordial Queen",
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

// trekB39Hand puts a card in p's hand with a mana cost and P/T.
func trekB39Hand(g *game.Game, p *game.Player, name, typeLine, cost, oracle string, power, toughness int) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
			KnownBy: map[uuid.UUID]bool{p.ID: true},
		})
	})
	return id
}

// trekB39Borgs is the Borg tokens p controls.
func trekB39Borgs(t *testing.T, g *game.Game, p uuid.UUID) int {
	t.Helper()
	return len(trekB1Borgs(t, g, p))
}

// trekB39Activate activates a catalog ability strictly with `mana` floated.
func trekB39Activate(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, mana string, targets ...game.TargetRef) error {
	t.Helper()
	if mana != "" {
		floatMana(t, g, p, mana)
	}
	return g.ActivateCatalogAbility(p.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: targets})
}

func trekB39Player(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id}
}
func trekB39CardRef(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

// --- Chancellor Gorkon -----------------------------------------------------------

func TestTrekB39GorkonHasVigilanceAndExilesUpToTwoCreaturesWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	gorkon := b27Push(g, me.ID, "Chancellor Gorkon", "Legendary Creature — Klingon Noble", trekB39GorkonOracle, "{3}{W}{B}", 5, 3, "W", "B")
	if !hasEffectiveKeyword(t, g, gorkon, "vigilance") {
		t.Error("Gorkon lacks vigilance")
	}
	a := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, g.Seats[2].ID, "Other Bear", "Creature — Bear", 2, 2)
	c := b12Creature(g, opp.ID, "Spared Bear", "Creature — Bear", 2, 2)
	b27Kill(g, gorkon)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("Gorkon's death asked for no targets")
	}
	if pick.PickTargetMin != 0 || pick.PickTargetMax != 2 {
		t.Errorf("target bounds %d..%d, want 0..2", pick.PickTargetMin, pick.PickTargetMax)
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{trekB39CardRef(a), trekB39CardRef(b)}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) {
		t.Error("the two targeted creatures were not exiled")
	}
	if !g.Battlefield.Contains(c) {
		t.Error("an untargeted creature left the battlefield")
	}
}

func TestTrekB39GorkonMayTargetNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	gorkon := b27Push(g, me.ID, "Chancellor Gorkon", "Legendary Creature — Klingon Noble", trekB39GorkonOracle, "{3}{W}{B}", 5, 3, "W", "B")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b27Kill(g, gorkon)
	passPriorityAroundTable(t, g)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTargets(pick.ID, me.ID, nil); err != nil {
			t.Fatalf("choosing no targets: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature was exiled with no target chosen")
	}
}

// --- Chief Kyle, Transporter Tech ------------------------------------------------

func TestTrekB39KyleBlinksAnotherCreatureYouControlOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	kyle := b27Push(g, me.ID, "Chief Kyle, Transporter Tech", "Legendary Creature — Human Artificer", trekB39KyleOracle, "{2}{W}{U}", 3, 2, "W", "U")
	bear := pushVanillaCreature(g, me.ID, "Kyle Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("landfall asked for no target")
	}
	if hasID(pick.PickTargetCards, kyle) || hasID(pick.PickTargetCards, theirs) || !hasID(pick.PickTargetCards, bear) {
		t.Fatalf("targets offered %v: want only another creature you control", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Fatal("the creature was not exiled")
	}
	advanceTo(t, g, game.StepEnd)
	ringSettle(t, g)
	if n := socB24Named(g, "Kyle Bear"); n != 1 {
		t.Errorf("%d Kyle Bear on the battlefield after the end step, want 1 (returned)", n)
	}
}

func TestTrekB39KyleIgnoresAnOpponentsLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b27Push(g, me.ID, "Chief Kyle, Transporter Tech", "Legendary Creature — Human Artificer", trekB39KyleOracle, "{2}{W}{U}", 3, 2, "W", "U")
	pushVanillaCreature(g, me.ID, "Kyle Bear", 2, 2)
	b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("an opponent's land triggered landfall")
	}
}

// --- Chief Montgomery Scott -------------------------------------------------------

func trekB39ScottTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	scott := b27Push(g, me.ID, "Chief Montgomery Scott", "Legendary Creature — Human Artificer", trekB39ScottOracle, "{2}{G}{U}", 3, 3, "G", "U")
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, g.Seats[1].ID, "Plains", "Basic Land — Plains")
	return g, me, scott
}

func TestTrekB39ScottPutsACreatureWithXPlusOneCounters(t *testing.T) {
	g, me, scott := trekB39ScottTable(t)
	bear := trekB39Hand(g, me, "Big Bear", "Creature — Bear", "{5}{G}", "", 4, 4)
	bolt := trekB39Hand(g, me, "Shock", "Instant", "{R}", "", 0, 0)
	if err := trekB39Activate(t, g, me, scott, "{C}{C}{C}{C}{C}"); err == nil {
		t.Fatal("activated with five mana")
	}
	if err := trekB39Activate(t, g, me, scott, "{C}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick from the hand")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 || hasID(pick.ChooseCards, bolt) || !hasID(pick.ChooseCards, bear) {
		t.Fatalf("pick %d..%d over %v: want an optional creature or Spacecraft card", pick.ChooseMin, pick.ChooseMax, pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature was not put onto the battlefield")
	}
	if n := countersOn(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("+1/+1 counters %d, want 2 (Forest and Island)", n)
	}
	if n := countersOn(g, bear, game.CounterCharge); n != 0 {
		t.Errorf("a creature got %d charge counters", n)
	}
}

func TestTrekB39ScottPutsASpacecraftWithXChargeCounters(t *testing.T) {
	g, me, scott := trekB39ScottTable(t)
	ship := trekB39Hand(g, me, "Shuttle", "Artifact — Spacecraft", "{4}", "", 0, 0)
	if err := trekB39Activate(t, g, me, scott, "{C}{C}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, ship)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ship) {
		t.Fatal("the Spacecraft was not put onto the battlefield")
	}
	if n := countersOn(g, ship, game.CounterCharge); n != 2 {
		t.Errorf("charge counters %d, want 2", n)
	}
	if n := countersOn(g, ship, game.CounterPlusOne); n != 0 {
		t.Errorf("a noncreature Spacecraft got %d +1/+1 counters", n)
	}
}

func TestTrekB39ScottMayPutNothing(t *testing.T) {
	g, me, scott := trekB39ScottTable(t)
	bear := trekB39Hand(g, me, "Big Bear", "Creature — Bear", "{5}{G}", "", 4, 4)
	if err := trekB39Activate(t, g, me, scott, "{C}{C}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the declined card left the hand")
	}
}

// --- Cmdr. Troi, Outgoing Empath --------------------------------------------------

// trekB39TroiTable is seat 1's main phase with my Troi on the battlefield and
// a Divination in their hand.
func trekB39TroiTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	troi := b27Push(g, me.ID, "Cmdr. Troi, Outgoing Empath", "Legendary Creature — Betazoid Human Advisor", trekB39TroiOracle, "{2}{U}{R}", 3, 3, "U", "R")
	powerUpPassTo(t, g, 1)
	div := trekB39Hand(g, opp, "Divination", "Sorcery", "{2}{U}", divinationOracle, 0, 0)
	return g, me, opp, troi, div
}

func TestTrekB39TroiCopiesTheSpellTheOpponentCastsFree(t *testing.T) {
	g, me, opp, troi, div := trekB39TroiTable(t)
	if !hasEffectiveKeyword(t, g, troi, "vigilance") {
		t.Error("Troi lacks vigilance")
	}
	myHand, theirHand := me.Hand.Size(), opp.Hand.Size()
	if err := trekB39Activate(t, g, me, troi, "", trekB39Player(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, opp.ID, game.MayCastKeywordHandCast, div)
	if err := castThroughResolutionCast(t, g, opp.ID, div, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - theirHand; got != 1 {
		t.Errorf("opponent's hand changed by %d, want +1 (cast Divination, drew two)", got)
	}
	if got := me.Hand.Size() - myHand; got != 2 {
		t.Errorf("my hand changed by %d, want +2 (the copy's draw, not Troi's)", got)
	}
}

func TestTrekB39TroiDrawsWhenTheOpponentDeclines(t *testing.T) {
	g, me, opp, troi, div := trekB39TroiTable(t)
	myHand := me.Hand.Size()
	if err := trekB39Activate(t, g, me, troi, "", trekB39Player(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	declineResolutionCast(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(div) {
		t.Error("the declined card left the opponent's hand")
	}
	if got := me.Hand.Size() - myHand; got != 1 {
		t.Errorf("my hand changed by %d, want +1", got)
	}
}

func TestTrekB39TroiTargetsOnlyTheOpponentWhoseTurnItIs(t *testing.T) {
	g, me, _, troi, _ := trekB39TroiTable(t)
	if err := trekB39Activate(t, g, me, troi, "", trekB39Player(g.Seats[2].ID)); err == nil {
		t.Error("Troi targeted an opponent whose turn it is not")
	}
	if err := trekB39Activate(t, g, me, troi, "", trekB39Player(me.ID)); err == nil {
		t.Error("Troi targeted its controller")
	}
}

// --- Collective Override ----------------------------------------------------------

// trekB39OverrideCombat attacks with two of my creatures and leaves the table
// in the declare attackers step with priority.
func trekB39OverrideCombat(t *testing.T, commander bool) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	if commander {
		b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 1, 1)
	}
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	declareAttack(t, g, opp.ID, a, b)
	id := trekB39Hand(g, me, "Collective Override", "Instant", "{2}{W}{U}", trekB39OverrideOracle, 0, 0)
	return g, me, id
}

func TestTrekB39OverrideCreatesABorgPerAttackingCreature(t *testing.T) {
	g, me, id := trekB39OverrideCombat(t, false)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}}); err == nil {
		t.Fatal("both bullets were accepted without a commander")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := trekB39Borgs(t, g, me.ID); n != 2 {
		t.Errorf("%d Borg tokens, want 2", n)
	}
}

func TestTrekB39OverrideWithACommanderDoesBoth(t *testing.T) {
	g, me, id := trekB39OverrideCombat(t, true)
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}}); err != nil {
		t.Fatalf("cast with both bullets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := trekB39Borgs(t, g, me.ID); n != 2 {
		t.Errorf("%d Borg tokens, want 2", n)
	}
	if got := me.Hand.Size() - (hand - 1); got != 2 {
		t.Errorf("drew %d, want 2: the new Borgs are not attacking", got)
	}
}

// --- Commander William Riker ------------------------------------------------------

func TestTrekB39RikerYouAndTargetOpponentEachDrawTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	riker := b27Push(g, me.ID, "Commander William Riker", "Legendary Creature — Human Officer", trekB39RikerOracle, "{1}{R}{W}", 3, 3, "R", "W")
	for _, kw := range []string{"vigilance", "haste"} {
		if !hasEffectiveKeyword(t, g, riker, kw) {
			t.Errorf("Riker lacks %s", kw)
		}
	}
	if err := trekB39Activate(t, g, me, riker, "", trekB39Player(opp.ID)); err == nil {
		t.Fatal("activated without paying {1}")
	}
	if err := trekB39Activate(t, g, me, riker, "", trekB39Player(me.ID)); err == nil {
		t.Fatal("targeted its controller")
	}
	mine, theirs, third := me.Hand.Size(), opp.Hand.Size(), g.Seats[2].Hand.Size()
	if err := trekB39Activate(t, g, me, riker, "{C}", trekB39Player(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size()-mine != 2 || opp.Hand.Size()-theirs != 2 || g.Seats[2].Hand.Size() != third {
		t.Errorf("hands changed by %d / %d / %d, want 2 / 2 / 0",
			me.Hand.Size()-mine, opp.Hand.Size()-theirs, g.Seats[2].Hand.Size()-third)
	}
}

// --- Curzon Dax, Bound by Blood ---------------------------------------------------

func TestTrekB39CurzonPingsEachOpponentWhenAnotherKlingonOrWarriorDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b27Push(g, me.ID, "Curzon Dax, Bound by Blood", "Legendary Creature — Trill Warrior", trekB39CurzonOracle, "{2}{R}{W}", 4, 3, "R", "W")
	klingon := b12Creature(g, me.ID, "Klingon", "Creature — Klingon Soldier", 2, 2)
	warrior := b12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Klingon", "Creature — Klingon Warrior", 2, 2)
	before := lifeOfOpponents(g)
	mine := me.Life
	for _, id := range []uuid.UUID{klingon, warrior, bear, theirs} {
		b27Kill(g, id)
		passPriorityAroundTable(t, g)
	}
	after := lifeOfOpponents(g)
	for i := range after {
		if before[i]-after[i] != 2 {
			t.Errorf("opponent %d lost %d, want 2 (one per Klingon or Warrior of yours)", i, before[i]-after[i])
		}
	}
	if me.Life != mine {
		t.Error("Curzon's controller took damage")
	}
}

func TestTrekB39CurzonDiesAndPutsASmallCreatureFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	curzon := b27Push(g, me.ID, "Curzon Dax, Bound by Blood", "Legendary Creature — Trill Warrior", trekB39CurzonOracle, "{2}{R}{W}", 4, 3, "R", "W")
	small := trekB39Hand(g, me, "Small", "Creature — Bear", "{1}{G}{G}", "", 3, 3)
	big := trekB39Hand(g, me, "Big", "Creature — Bear", "{2}{G}{G}", "", 4, 4)
	opps := lifeOfOpponents(g)
	b27Kill(g, curzon)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("Curzon's death offered no creature")
	}
	if pick.ChooseMin != 0 || hasID(pick.ChooseCards, big) || !hasID(pick.ChooseCards, small) {
		t.Fatalf("pick %d..%d over %v: want an optional creature with mana value 3 or less", pick.ChooseMin, pick.ChooseMax, pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the creature was not put onto the battlefield")
	}
	for i, l := range lifeOfOpponents(g) {
		if l != opps[i] {
			t.Error("Curzon's own death pinged the opponents")
		}
	}
}

// --- Deliver from Chaos -----------------------------------------------------------

func TestTrekB39DeliverExilesAndMakesBorgsEqualToManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	ogre := b27Push(g, opp.ID, "Ogre", "Creature — Ogre", "", "{2}{R}", 3, 3, "R")
	aura := b12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	id := trekB39Hand(g, me, "Deliver from Chaos", "Instant", "{3}{W}{B}", trekB39DeliverOracle, 0, 0)
	floatMana(t, g, me, "{W}{B}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{trekB39CardRef(aura)}}); err == nil {
		t.Fatal("an enchantment was a legal target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{trekB39CardRef(ogre)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(ogre) {
		t.Error("the creature was not exiled")
	}
	if n := trekB39Borgs(t, g, me.ID); n != 3 {
		t.Errorf("%d Borg tokens, want 3", n)
	}
	if n := trekB39Borgs(t, g, opp.ID); n != 0 {
		t.Errorf("the target's controller got %d Borgs", n)
	}
}

func TestTrekB39DeliverOnAZeroArtifactMakesNoBorgs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	id := trekB39Hand(g, me, "Deliver from Chaos", "Instant", "{3}{W}{B}", trekB39DeliverOracle, 0, 0)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{trekB39CardRef(rock)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(rock) {
		t.Error("the artifact was not exiled")
	}
	if n := trekB39Borgs(t, g, me.ID); n != 0 {
		t.Errorf("%d Borg tokens, want 0", n)
	}
}

// --- Hugh, First of Many ----------------------------------------------------------

func trekB39HughTable(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Hugh, First of Many", "Legendary Artifact Creature — Borg", trekB39HughOracle, false)
	return g, me
}

func trekB39Named(g *game.Game, controller uuid.UUID, name string) (cards, tokens int) {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			if IsToken(c) {
				tokens++
			} else {
				cards++
			}
		}
	}
	return cards, tokens
}

func TestTrekB39HughGivesArtifactSpellsReplicate(t *testing.T) {
	g, me := trekB39HughTable(t)
	ring := trekB39Hand(g, me, "Gear", "Artifact", "{1}", "", 0, 0)
	floatMana(t, g, me, "{C}{U}{U}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, ring, game.CastSpellParams{Strict: true, OptionalCosts: []int{0, 0}}); err != nil {
		t.Fatalf("cast with replicate twice: %v", err)
	}
	passPriorityAroundTable(t, g)
	cards, tokens := trekB39Named(g, me.ID, "Gear")
	if cards != 1 || tokens != 2 {
		t.Errorf("%d Gear cards and %d Gear tokens, want 1 and 2", cards, tokens)
	}
}

func TestTrekB39HughWithoutReplicatePaidCopiesNothing(t *testing.T) {
	g, me := trekB39HughTable(t)
	gear := trekB39Hand(g, me, "Gear", "Artifact", "{1}", "", 0, 0)
	floatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, gear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if cards, tokens := trekB39Named(g, me.ID, "Gear"); cards != 1 || tokens != 0 {
		t.Errorf("%d Gear cards and %d tokens, want 1 and 0", cards, tokens)
	}
}

func TestTrekB39HughGivesNoReplicateToANonartifactSpell(t *testing.T) {
	g, me := trekB39HughTable(t)
	bear := trekB39Hand(g, me, "Bear", "Creature — Bear", "{1}", "", 2, 2)
	floatMana(t, g, me, "{C}{U}{C}{C}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err == nil {
		t.Error("a creature spell paid a replicate cost")
	}
}

func TestTrekB39TwoHughsCopyTheSpellOncePerPayment(t *testing.T) {
	g, me := trekB39HughTable(t)
	// A second Hugh that is not legendary (a copy "except it isn't
	// legendary"), so the legend rule keeps both.
	pushCatalogPermanent(g, me.ID, "Hugh, First of Many", "Artifact Creature — Borg", trekB39HughOracle, false)
	gear := trekB39Hand(g, me, "Gear", "Artifact", "{1}", "", 0, 0)
	floatMana(t, g, me, "{C}{U}{C}{C}")
	if err := g.CastSpell(me.ID, gear, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast with replicate once: %v", err)
	}
	b21SettleStack(t, g)
	if cards, tokens := trekB39Named(g, me.ID, "Gear"); cards != 1 || tokens != 1 {
		t.Errorf("%d Gear cards and %d tokens, want 1 and 1", cards, tokens)
	}
}

// --- Jurati, Cordial Queen --------------------------------------------------------

func TestTrekB39JuratiCopiesAnArtifactTokenOncePerPlayerHit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Jurati, Cordial Queen", "Legendary Artifact Creature — Borg Scientist", trekB39JuratiOracle, false)
	a := pushDiesCreatureForTest(g, me.ID, "Drone A", "", "Artifact Creature — Borg", 1, 1)
	b := pushDiesCreatureForTest(g, me.ID, "Drone B", "", "Artifact Creature — Borg", 1, 1)
	var token uuid.UUID
	g.WithWriteLock(func() {
		ids, _ := g.CreateTokensForEffect(me.ID, TokenCard("1/1 colorless Borg artifact"), 1, game.TokenEntryOptions{})
		token = ids[0]
	})
	attackWith(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("Jurati did not trigger")
	}
	pickCard(t, g, me.ID, token)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("Jurati triggered twice for one player")
	}
	if n := trekB39Borgs(t, g, me.ID); n != 2 {
		t.Errorf("%d Borg tokens, want 2", n)
	}
}

func TestTrekB39JuratiIgnoresANonartifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Jurati, Cordial Queen", "Legendary Artifact Creature — Borg Scientist", trekB39JuratiOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		_, _ = g.CreateTokensForEffect(me.ID, TokenCard("1/1 colorless Borg artifact"), 1, game.TokenEntryOptions{})
	})
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a nonartifact creature's damage triggered Jurati")
	}
}
