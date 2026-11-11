package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s31_cards_test.go — batch 31, Murders at Karlov Manor Commander
// "Investigate: making Clues" (#5368). Each test plays the real catalog
// entry and asserts what its text says.

const (
	cmdr2S31ConfirmOracle  = "a30d076b-c942-45e4-b943-3749ce955291"
	cmdr2S31BodiesOracle   = "2d6a0d89-7ebd-420e-9892-be1506ab3714"
	cmdr2S31WaveOracle     = "d1ec1e4e-faa6-4afc-9e00-a3327105c11b"
	cmdr2S31EtherealOracle = "24038629-7c12-45ef-8ec4-0d2bb3e2265c"
	cmdr2S31UlvenOracle    = "6588ea9a-658f-40d9-a64b-3a418ab81183"
	cmdr2S31OngoingOracle  = "68227832-2916-49b9-92bb-c09b8a0c13f5"
	cmdr2S31MoleOracle     = "5b2c6b44-99b3-496d-93ca-d5211222038f"
	cmdr2S31MorskaOracle   = "80ea3e65-946d-42c3-a26f-3d74e81d70b4"
	cmdr2S31DisorderOracle = "5e7586dd-d42e-4ac6-adb8-d46b6cdef923"
	cmdr2S31TailOracle     = "3d28e89d-d5e2-46b1-8dd2-d808b78f08fe"
)

func TestCmdr2S31EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		cmdr2S31ConfirmOracle:  "Confirm Suspicions",
		cmdr2S31BodiesOracle:   "Follow the Bodies",
		cmdr2S31WaveOracle:     "Wavesifter",
		cmdr2S31EtherealOracle: "Ethereal Investigator",
		cmdr2S31UlvenOracle:    "Ulvenwald Mysteries",
		cmdr2S31OngoingOracle:  "Ongoing Investigation",
		cmdr2S31MoleOracle:     "Graf Mole",
		cmdr2S31MorskaOracle:   "Morska, Undersea Sleuth",
		cmdr2S31DisorderOracle: "Disorder in the Court",
		cmdr2S31TailOracle:     "Unshakable Tail",
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

// cmdr2S31Clue puts a real Clue under a player's control.
func cmdr2S31Clue(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushPermanent(g, owner, ClueToken())
}

func cmdr2S31Sacrifice(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(id); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
}

// --- Confirm Suspicions ------------------------------------------------------

func TestCmdr2S31ConfirmSuspicionsCountersAndInvestigatesThreeTimes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	shock := castFromSeat(t, g, opp, "Shock", "Instant", "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	life := me.Life
	castFromSeat(t, g, me, "Confirm Suspicions", "Instant", cmdr2S31ConfirmOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: shock}})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(shock) {
		t.Error("the Shock was not countered")
	}
	if me.Life != life {
		t.Errorf("life %d, want %d: the countered Shock still resolved", me.Life, life)
	}
	if got := len(mkmMP2Clues(g, me.ID)); got != 3 {
		t.Errorf("%d Clues, want 3", got)
	}
}

func TestCmdr2S31ConfirmSuspicionsNeedsASpellTarget(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	err := castCatalogSpellErr(t, g, "Confirm Suspicions", "Instant", cmdr2S31ConfirmOracle, nil)
	if err == nil {
		t.Fatal("cast with no target on the stack")
	}
}

// --- Follow the Bodies -------------------------------------------------------

func TestCmdr2S31FollowTheBodiesCopiesForEachPermanentPutIntoAGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	killCreature(t, g, me.ID, b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2))
	killCreature(t, g, opp.ID, b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2))
	castCatalogSpell(t, g, "Follow the Bodies", "Sorcery", cmdr2S31BodiesOracle, nil)
	passPriorityAroundTable(t, g)
	if got := len(mkmMP2Clues(g, me.ID)); got != 3 {
		t.Errorf("%d Clues, want 3 (the spell and two copies)", got)
	}
}

func TestCmdr2S31FollowTheBodiesWithNothingDeadIsOneClue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Follow the Bodies", "Sorcery", cmdr2S31BodiesOracle, nil)
	passPriorityAroundTable(t, g)
	if got := len(mkmMP2Clues(g, me.ID)); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- Wavesifter --------------------------------------------------------------

func TestCmdr2S31WavesifterInvestigatesTwiceWhenCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := castHoldCreature(t, g, "Wavesifter", "Creature — Elemental", cmdr2S31WaveOracle, 3, 2)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Wavesifter is not on the battlefield")
	}
	assertKeywords(t, g, id, "flying")
	if got := len(mkmMP2Clues(g, me.ID)); got != 2 {
		t.Errorf("%d Clues, want 2", got)
	}
}

func TestCmdr2S31WavesifterEvokedStillInvestigatesThenIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if ac := game.AlternativeCostByKey(cmdr2S31WaveOracle, "evoke"); ac == nil || ac.ManaCost != "{G}{U}" {
		t.Fatalf("Wavesifter's evoke is %+v", ac)
	}
	id := castWithAltCost(t, g, "Wavesifter", "Creature — Elemental", cmdr2S31WaveOracle, "evoke")
	if !answerAnyTriggerOrderPrompt(t, g, me.ID) {
		t.Fatal("evoked Wavesifter should queue two triggers and ask for an order")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
		t.Error("the evoked Wavesifter was not sacrificed")
	}
	if got := len(mkmMP2Clues(g, me.ID)); got != 2 {
		t.Errorf("%d Clues, want 2", got)
	}
}

// --- Graf Mole ---------------------------------------------------------------

func TestCmdr2S31GrafMoleGainsThreeWhenYouSacrificeAClue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Graf Mole", "Creature — Mole Beast", cmdr2S31MoleOracle, false)
	theirs := cmdr2S31Clue(g, opp.ID)
	cmdr2S31Sacrifice(t, g, theirs)
	fodder := b12Permanent(g, me.ID, "Fodder", "Artifact")
	cmdr2S31Sacrifice(t, g, fodder)
	life := me.Life
	cmdr2S31Sacrifice(t, g, cmdr2S31Clue(g, me.ID))
	if got := me.Life - life; got != 3 {
		t.Errorf("gained %d life for my Clue, want 3 (another player's Clue and a non-Clue gain nothing)", got)
	}
}

// --- Ethereal Investigator ---------------------------------------------------

func TestCmdr2S31EtherealInvestigatorInvestigatesOncePerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := castHoldCreature(t, g, "Ethereal Investigator", "Creature — Spirit", cmdr2S31EtherealOracle, 2, 3)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("not on the battlefield")
	}
	assertKeywords(t, g, id, "flying")
	if got, want := len(mkmMP2Clues(g, me.ID)), len(g.Seats)-1; got != want {
		t.Errorf("%d Clues, want %d (one per opponent)", got, want)
	}
}

func TestCmdr2S31EtherealInvestigatorMakesASpiritOnTheSecondDrawOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Ethereal Investigator", "Creature — Spirit", cmdr2S31EtherealOracle, false)
	hobA5Draw(t, g, other, nil)
	hobA5Draw(t, g, other, nil)
	if n := countBattlefieldNamed(g, me.ID, "Spirit"); n != 0 {
		t.Fatalf("%d Spirits after another player drew twice", n)
	}
	hobA5Draw(t, g, me, nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Spirit"); n != 0 {
		t.Fatalf("%d Spirits after the first draw", n)
	}
	hobA5Draw(t, g, me, nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Spirit"); n != 1 {
		t.Fatalf("%d Spirits after the second draw, want 1", n)
	}
	hobA5Draw(t, g, me, nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Spirit"); n != 1 {
		t.Errorf("%d Spirits after the third draw, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && (c.Power != 1 || c.Toughness != 1 || !hasKeywordOnBattlefield(t, g, c.InstanceID, "flying")) {
			t.Errorf("Spirit is %d/%d, want a 1/1 flier", c.Power, c.Toughness)
		}
	}
}

// --- Ulvenwald Mysteries -----------------------------------------------------

func TestCmdr2S31UlvenwaldMysteriesInvestigatesOnANontokenCreatureDying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Ulvenwald Mysteries", "Enchantment", cmdr2S31UlvenOracle, false)
	killCreature(t, g, opp.ID, b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2))
	passPriorityAroundTable(t, g)
	if got := len(mkmMP2Clues(g, me.ID)); got != 0 {
		t.Fatalf("%d Clues after an opponent's creature died", got)
	}
	tok := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Token", TypeLine: "Token Creature — Bear",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	killCreature(t, g, me.ID, tok)
	passPriorityAroundTable(t, g)
	if got := len(mkmMP2Clues(g, me.ID)); got != 0 {
		t.Fatalf("%d Clues after a token died", got)
	}
	killCreature(t, g, me.ID, b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2))
	passPriorityAroundTable(t, g)
	if got := len(mkmMP2Clues(g, me.ID)); got != 1 {
		t.Errorf("%d Clues after my nontoken creature died, want 1", got)
	}
}

func TestCmdr2S31UlvenwaldMysteriesMakesASoldierWhenYouSacrificeAClue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Ulvenwald Mysteries", "Enchantment", cmdr2S31UlvenOracle, false)
	cmdr2S31Sacrifice(t, g, cmdr2S31Clue(g, opp.ID))
	cmdr2S31Sacrifice(t, g, b12Permanent(g, me.ID, "Fodder", "Artifact"))
	if n := countBattlefieldNamed(g, me.ID, "Human Soldier"); n != 0 {
		t.Fatalf("%d Soldiers before I sacrificed a Clue", n)
	}
	cmdr2S31Sacrifice(t, g, cmdr2S31Clue(g, me.ID))
	if n := countBattlefieldNamed(g, me.ID, "Human Soldier"); n != 1 {
		t.Fatalf("%d Soldiers, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Human Soldier" && (c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "W") {
			t.Errorf("Soldier is %d/%d %v, want a 1/1 white", c.Power, c.Toughness, c.Colors)
		}
	}
}

// --- Ongoing Investigation ---------------------------------------------------

func TestCmdr2S31OngoingInvestigationInvestigatesWhenCreaturesConnect(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Ongoing Investigation", "Enchantment", cmdr2S31OngoingOracle, false)
	a := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Elk", "Creature — Elk", 3, 3)
	attackWith(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	if got := len(mkmMP2Clues(g, me.ID)); got != 1 {
		t.Errorf("%d Clues after two creatures hit one player, want 1", got)
	}
}

func TestCmdr2S31OngoingInvestigationExilesACreatureCardForAClueAndLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	ench := pushCatalogPermanent(g, me.ID, "Ongoing Investigation", "Enchantment", cmdr2S31OngoingOracle, false)
	bear := mkmMP2Grave(me, "Dead Bear", "Creature — Bear")
	spell := mkmMP2Grave(me, "Dead Bolt", "Instant")
	fillPoolColored(me, "G", 1)
	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, ench, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("a noncreature card paid the exile cost")
	}
	life := me.Life
	b16Activate(t, g, me.ID, ench, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{bear}})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) || !me.Graveyard.Contains(spell) {
		t.Error("the wrong card was exiled")
	}
	if got := len(mkmMP2Clues(g, me.ID)); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}
}

// --- Morska, Undersea Sleuth -------------------------------------------------

func TestCmdr2S31MorskaHasNoMaximumHandSize(t *testing.T) {
	spec, _ := Lookup(cmdr2S31MorskaOracle)
	if !spec.NoMaxHandSize {
		t.Error("Morska must declare NoMaxHandSize")
	}
}

func TestCmdr2S31MorskaInvestigatesAtYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Morska, Undersea Sleuth", "Legendary Creature — Vedalken Fish Detective", cmdr2S31MorskaOracle, false)
	advanceToNextSeatsTurn(t, g)
	advanceToUpkeepOfSeat(t, g, seatIndexOf(g, me.ID))
	passPriorityAroundTable(t, g)
	if got := len(mkmMP2Clues(g, me.ID)); got != 1 {
		t.Errorf("%d Clues after my upkeep, want 1", got)
	}
}

func TestCmdr2S31MorskaGrowsByTwoOnTheSecondDrawOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	id := pushCatalogPermanent(g, me.ID, "Morska, Undersea Sleuth", "Legendary Creature — Vedalken Fish Detective", cmdr2S31MorskaOracle, false)
	hobA5Draw(t, g, other, nil)
	hobA5Draw(t, g, other, nil)
	hobA5Draw(t, g, me, nil)
	if n := hobA5Counters(t, g, id); n != 0 {
		t.Fatalf("%d counters after the first draw", n)
	}
	hobA5Draw(t, g, me, nil)
	if n := hobA5Counters(t, g, id); n != 2 {
		t.Fatalf("%d counters after the second draw, want 2", n)
	}
	hobA5Draw(t, g, me, nil)
	if n := hobA5Counters(t, g, id); n != 2 {
		t.Errorf("%d counters after the third draw, want 2", n)
	}
}

// --- Disorder in the Court ---------------------------------------------------

func TestCmdr2S31DisorderExilesXCreaturesInvestigatesAndReturnsThemTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castXSpell(t, g, "Disorder in the Court", "Instant", cmdr2S31DisorderOracle, "{X}{W}{U}", 2,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if !exileHas(g, mine) || !exileHas(g, theirs) {
		t.Fatal("the targets were not exiled")
	}
	if got := len(mkmMP2Clues(g, me.ID)); got != 2 {
		t.Fatalf("%d Clues, want 2", got)
	}
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	for _, c := range []struct {
		owner *game.Player
		name  string
	}{{me, "Mine"}, {opp, "Theirs"}} {
		var found *game.Card
		for _, bc := range g.Battlefield.Cards {
			if bc.Name == c.name {
				bc := bc
				found = &bc
			}
		}
		if found == nil {
			t.Errorf("%s did not return", c.name)
			continue
		}
		if found.Controller != c.owner.ID {
			t.Errorf("%s returned under %s, want its owner", c.name, found.Controller)
		}
		if !found.Tapped {
			t.Errorf("%s returned untapped", c.name)
		}
	}
}

func TestCmdr2S31DisorderWithXZeroDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	castXSpell(t, g, "Disorder in the Court", "Instant", cmdr2S31DisorderOracle, "{X}{W}{U}", 0, nil)
	passPriorityAroundTable(t, g)
	if got := len(mkmMP2Clues(g, me.ID)); got != 0 {
		t.Errorf("%d Clues for X=0", got)
	}
}

// --- Unshakable Tail ---------------------------------------------------------

func cmdr2S31TailOnBattlefield(g *game.Game, me *game.Player) uuid.UUID {
	return pushCatalogPermanent(g, me.ID, "Unshakable Tail", "Creature — Zombie Detective", cmdr2S31TailOracle, false)
}

func TestCmdr2S31UnshakableTailSurveilsWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Top", "Next")
	castHoldCreature(t, g, "Unshakable Tail", "Creature — Zombie Detective", cmdr2S31TailOracle, 3, 2)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("surveil choice %+v, want one card", c)
	}
}

func TestCmdr2S31UnshakableTailSurveilsAtYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Top", "Next")
	cmdr2S31TailOnBattlefield(g, me)
	advanceToNextSeatsTurn(t, g)
	advanceToUpkeepOfSeat(t, g, seatIndexOf(g, me.ID))
	passPriorityAroundTable(t, g)
	if c := surveilChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("surveil choice %+v, want one card", c)
	}
}

func TestCmdr2S31UnshakableTailInvestigatesWhenCreatureCardsAreMilled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdr2S31TailOnBattlefield(g, me)
	seedLibraryTop(me, "Land", "Basic Land — Forest", "")
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if got := len(mkmMP2Clues(g, me.ID)); got != 0 {
		t.Fatalf("%d Clues after milling a land", got)
	}
	seedLibraryTop(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if got := len(mkmMP2Clues(g, me.ID)); got != 0 {
		t.Fatalf("%d Clues after an opponent milled a creature", got)
	}
	seedLibraryTop(me, "Bear", "Creature — Bear", "{1}{G}")
	seedLibraryTop(me, "Elk", "Creature — Elk", "{1}{G}")
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 2) })
	passPriorityAroundTable(t, g)
	if got := len(mkmMP2Clues(g, me.ID)); got != 1 {
		t.Errorf("%d Clues after milling two creature cards at once, want 1", got)
	}
}

func TestCmdr2S31UnshakableTailReturnsFromTheGraveyardForAClue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tail := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: tail, Name: "Unshakable Tail", TypeLine: "Creature — Zombie Detective",
		OracleID: cmdr2S31TailOracle, Owner: me.ID, Controller: me.ID})
	fodder := b12Permanent(g, me.ID, "Rock", "Artifact")
	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, tail, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Fatal("a non-Clue paid the cost")
	}
	clue := cmdr2S31Clue(g, me.ID)
	b16Activate(t, g, me.ID, tail, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{clue}})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(tail) {
		t.Error("Unshakable Tail did not return to hand")
	}
	if g.Battlefield.Contains(clue) {
		t.Error("the Clue was not sacrificed")
	}
}
