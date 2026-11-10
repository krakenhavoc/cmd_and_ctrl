package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b41_cards_test.go — Standard set batch 41 (#4452): ten Murders at
// Karlov Manor cards that needed a seam. Each test plays the real catalog
// entry and asserts what its text says.

const (
	stdB41BuzzcrusherOracle = "79a63064-f25b-4c34-867e-a9d32da97b14"
	stdB41PhoenixOracle     = "c6f7efad-381c-45ba-9976-159b7e78f7ed"
	stdB41DetectiveOracle   = "4debd0e4-3bb6-4dd1-addc-08e4bda69d11"
	stdB41HerringOracle     = "f0ee45b5-697b-48f8-b1a6-cb3ae75e4a29"
	stdB41TrackerOracle     = "fe540a23-0357-44fa-8e1c-0a5c2a04cb01"
	stdB41HothouseOracle    = "2380fd04-6d19-4f5a-8a9e-c17d9ae09fb6"
	stdB41KillerOracle      = "79d577a8-53c0-4220-bc0e-70764d72d5fd"
	stdB41RopeOracle        = "a3f09898-6c77-4d68-8d10-8440c75f884c"
	stdB41MaverickOracle    = "1ba59349-a162-4444-b3e0-cddfc01ef03b"
	stdB41RookieOracle      = "0642fb0d-73cb-4989-852c-3233fe36c3b4"
)

func TestStdB41EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		stdB41BuzzcrusherOracle: "Krenko's Buzzcrusher",
		stdB41PhoenixOracle:     "Lamplight Phoenix",
		stdB41DetectiveOracle:   "Reckless Detective",
		stdB41HerringOracle:     "Red Herring",
		stdB41TrackerOracle:     "Vengeful Tracker",
		stdB41HothouseOracle:    "Case of the Locked Hothouse",
		stdB41KillerOracle:      "A Killer Among Us",
		stdB41RopeOracle:        "Rope",
		stdB41MaverickOracle:    "Rubblebelt Maverick",
		stdB41RookieOracle:      "Sharp-Eyed Rookie",
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

// --- Rope ------------------------------------------------------------------------

func TestStdB41RopePumpsGivesReachLimitsBlockersAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	rope := b12Push(g, me.ID, "Rope", "Artifact — Clue Equipment", stdB41RopeOracle, 0, 0)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	w1 := b12Creature(g, opp.ID, "Wall One", "Creature — Wall", 0, 5)
	w2 := b12Creature(g, opp.ID, "Wall Two", "Creature — Wall", 0, 5)

	floatMana(t, g, me, "{C}{C}{C}")
	p7Activate(t, g, me, rope, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{cardRef(bear)}})
	if !attachedTo(t, g, rope, bear) {
		t.Fatal("Rope is not attached")
	}
	if p, tough := mkmMPPT(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("the equipped creature is %d/%d, want 3/4", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "reach") {
		t.Error("the equipped creature has no reach")
	}
	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: w1, Attacker: bear}, {Blocker: w2, Attacker: bear}}); err == nil {
		t.Fatal("two creatures blocked the equipped creature")
	}
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: w1, Attacker: bear}}); err != nil {
		t.Fatalf("one creature blocking it: %v", err)
	}
}

func TestStdB41RopeCracksForACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rope := b12Push(g, me.ID, "Rope", "Artifact — Clue Equipment", stdB41RopeOracle, 0, 0)
	seedLibrary(me, "One")
	hand := me.Hand.Size()
	floatMana(t, g, me, "{C}{C}")
	p7Activate(t, g, me, rope, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(rope) {
		t.Error("Rope was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// --- Red Herring -----------------------------------------------------------------

func TestStdB41RedHerringMustAttackEachCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	herring := b12Push(g, me.ID, "Red Herring", "Artifact Creature — Clue Fish", stdB41HerringOracle, 2, 2)
	assertKeywords(t, g, herring, "haste")
	advanceToDeclareAttackersOf(t, g, 0)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != herring {
		t.Fatalf("pass with Red Herring home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(herring, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with Red Herring attacking: %v", err)
	}
}

func TestStdB41RedHerringSacrificesItselfToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	herring := b12Push(g, me.ID, "Red Herring", "Artifact Creature — Clue Fish", stdB41HerringOracle, 2, 2)
	seedLibrary(me, "One")
	hand := me.Hand.Size()
	floatMana(t, g, me, "{C}{C}")
	p7Activate(t, g, me, herring, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(herring) {
		t.Error("Red Herring was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// --- Vengeful Tracker ------------------------------------------------------------

func TestStdB41VengefulTrackerPunishesAnOpponentsArtifactSacrificeOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Vengeful Tracker", "Creature — Human Detective", stdB41TrackerOracle, 2, 2)
	theirClue := pushToken(g, opp.ID, ClueToken())
	myClue := pushToken(g, me.ID, ClueToken())
	theirBear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// Their Clue, sacrificed to its own ability: 2 damage to them.
	life, mine := opp.Life, me.Life
	seedLibrary(opp, "One")
	floatMana(t, g, opp, "{C}{C}")
	if err := g.ActivateCatalogAbility(opp.ID, theirClue, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("crack their Clue: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-2)
	}
	// A creature that is not an artifact: nothing.
	life = opp.Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirBear) })
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("sacrificing a non-artifact cost %d life", life-opp.Life)
	}
	// My own artifact: nothing.
	seedLibrary(me, "One")
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, myClue, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("crack my Clue: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != mine || opp.Life != life {
		t.Errorf("sacrificing my own artifact moved life: me %d (was %d), them %d (was %d)", me.Life, mine, opp.Life, life)
	}
}

// --- Reckless Detective ----------------------------------------------------------

func TestStdB41RecklessDetectiveSacrificesAnArtifactToDrawAndPump(t *testing.T) {
	g, me, opp := eoTable(t)
	det := b12Push(g, me.ID, "Reckless Detective", "Creature — Devil Detective", stdB41DetectiveOracle, 0, 3)
	clue := pushToken(g, me.ID, ClueToken())
	seedLibrary(me, "One")
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, det)
	passPriorityAroundTable(t, g)
	if got := eoOfferedOptions(t, g, me.ID); !got[0] || got[1] {
		t.Fatalf("offered %v, want the sacrifice only (the hand is empty)", got)
	}
	eoPay(t, g, me.ID, 0, clue)
	if got := effectivePower(t, g, det); got != 2 {
		t.Errorf("power %d, want 2", got)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	if onBattlefield(g, clue) {
		t.Error("the Clue was not sacrificed")
	}
}

func TestStdB41RecklessDetectiveDiscardsACardInstead(t *testing.T) {
	g, me, opp := eoTable(t)
	det := b12Push(g, me.ID, "Reckless Detective", "Creature — Devil Detective", stdB41DetectiveOracle, 0, 3)
	card := eoHand(me, "Spare", "Sorcery")
	seedLibrary(me, "One")
	declareAttack(t, g, opp.ID, det)
	passPriorityAroundTable(t, g)
	if got := eoOfferedOptions(t, g, me.ID); got[0] || !got[1] {
		t.Fatalf("offered %v, want the discard only (no artifact)", got)
	}
	eoPay(t, g, me.ID, 1, card)
	if got := effectivePower(t, g, det); got != 2 {
		t.Errorf("power %d, want 2", got)
	}
	if !me.Graveyard.Contains(card) {
		t.Error("the card was not discarded")
	}
	if me.Hand.Size() != 1 { // the discarded card left, one was drawn
		t.Errorf("hand %d, want 1", me.Hand.Size())
	}
}

// --- Rubblebelt Maverick ---------------------------------------------------------

func TestStdB41RubblebeltMaverickSurveilsOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "One", "Two", "Three")
	castHoldCreature(t, g, "Rubblebelt Maverick", "Creature — Human Detective", stdB41MaverickOracle, 1, 1)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("surveil prompt %+v, want 2 cards", c)
	}
}

func TestStdB41RubblebeltMaverickExilesItselfFromTheGraveyardAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	grave := gyCard(me, "Rubblebelt Maverick", "Creature — Human Detective", "{G}", stdB41MaverickOracle)

	floatMana(t, g, me, "{G}")
	p7Activate(t, g, me, grave, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{cardRef(bear)}})
	if !inExile(g, grave) {
		t.Error("the Maverick was not exiled from the graveyard")
	}
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("the bear has %d +1/+1 counters, want 1", n)
	}

	// Not at instant speed: with a spell on the stack it is refused.
	grave2 := gyCard(me, "Rubblebelt Maverick", "Creature — Human Detective", "{G}", stdB41MaverickOracle)
	castCatalogSpell(t, g, "Spell", "Sorcery", "", nil)
	floatMana(t, g, me, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, grave2, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{cardRef(bear)}}); err == nil {
		t.Error("activated with a spell on the stack")
	}
}

// --- Sharp-Eyed Rookie -----------------------------------------------------------

func TestStdB41SharpEyedRookieGrowsAndInvestigatesForABiggerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rookie := b12Push(g, me.ID, "Sharp-Eyed Rookie", "Creature — Human Detective", stdB41RookieOracle, 2, 2)
	assertKeywords(t, g, rookie, "vigilance")

	// 2/2 enters: neither stat is greater, so nothing triggers.
	castHoldCreature(t, g, "Same Size", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, rookie, game.CounterPlusOne); n != 0 || tmtB1Count(g, me.ID, "Clue") != 0 {
		t.Fatalf("an equal creature gave %d counters and %d Clues", n, tmtB1Count(g, me.ID, "Clue"))
	}
	// A 1/3: its toughness is greater.
	castHoldCreature(t, g, "Tall", "Creature — Wall", "", 1, 3)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, rookie, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after a taller creature, want 1", n)
	}
	if got := tmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

func TestStdB41SharpEyedRookieRechecksAsItResolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rookie := b12Push(g, me.ID, "Sharp-Eyed Rookie", "Creature — Human Detective", stdB41RookieOracle, 2, 2)
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Big", TypeLine: "Creature — Beast", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	// Resolve the creature spell only: the Rookie's trigger is then on the stack.
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	// In response, the Rookie grows past the entered creature.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(rookie, game.CounterPlusOne, 2) })
	passPriorityAroundTable(t, g)
	if n := countersOn(g, rookie, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters, want 2 (the trigger should do nothing)", n)
	}
	if got := tmtB1Count(g, me.ID, "Clue"); got != 0 {
		t.Errorf("%d Clues, want 0", got)
	}
}

// --- Lamplight Phoenix -----------------------------------------------------------

func TestStdB41LamplightPhoenixReturnsTappedForFourManaValueOfEvidence(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	phoenix := b12Push(g, me.ID, "Lamplight Phoenix", "Creature — Phoenix", stdB41PhoenixOracle, 3, 3)
	assertKeywords(t, g, phoenix, "flying")
	gy := mkmCEEvidence(me, "{1}", "{3}")
	killCreature(t, g, me.ID, phoenix)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no collect evidence prompt")
	}
	for _, id := range pick.ChooseCards {
		if id == phoenix {
			t.Fatal("the Phoenix was offered to pay for its own collect evidence")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, gy[:1]); err == nil {
		t.Fatal("collect evidence 4 with a total of 1 was accepted")
	}
	answerChooseCards(t, g, me.ID, gy...)
	passPriorityAroundTable(t, g)
	// It returns as a new object (CR 400.7).
	phoenix = findBattlefieldByName(g, "Lamplight Phoenix")
	if phoenix == uuid.Nil {
		t.Fatal("the Phoenix did not return")
	}
	if !b16Tapped(t, g, phoenix) {
		t.Error("the Phoenix returned untapped")
	}
	if !inExile(g, gy[0]) || !inExile(g, gy[1]) {
		t.Error("the evidence was not exiled")
	}
}

func TestStdB41LamplightPhoenixStaysPutWhenDeclinedOrTheGraveyardIsShort(t *testing.T) {
	t.Run("declined", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		phoenix := b12Push(g, me.ID, "Lamplight Phoenix", "Creature — Phoenix", stdB41PhoenixOracle, 3, 3)
		gy := mkmCEEvidence(me, "{5}")
		killCreature(t, g, me.ID, phoenix)
		passPriorityAroundTable(t, g)
		answerChooseCards(t, g, me.ID)
		passPriorityAroundTable(t, g)
		if onBattlefield(g, phoenix) || !me.Graveyard.Contains(phoenix) || !me.Graveyard.Contains(gy[0]) {
			t.Error("declining moved something")
		}
	})
	t.Run("only the Phoenix itself would make four", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		// The Phoenix is {1}{R}{R}: with a {1} beside it the total is 4, but
		// the Phoenix can't pay for itself, so the offer is not made.
		phoenix := b12Push(g, me.ID, "Lamplight Phoenix", "Creature — Phoenix", stdB41PhoenixOracle, 3, 3)
		mkmCEEvidence(me, "{1}")
		killCreature(t, g, me.ID, phoenix)
		passPriorityAroundTable(t, g)
		if pick := chooseCardsChoiceFor(g, me.ID); pick != nil {
			t.Errorf("asked with only {1} besides the Phoenix: %+v", pick)
		}
	})
}

// --- Krenko's Buzzcrusher --------------------------------------------------------

func stdB41Land(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return b12Permanent(g, owner, name, "Land")
}

func TestStdB41BuzzcrusherDestroysNonbasicLandsAndTheirControllersSearch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := stdB41Land(g, me.ID, "My Nonbasic")
	theirs := stdB41Land(g, opp.ID, "Their Nonbasic")
	basic := b12Permanent(g, opp.ID, "Plains", "Basic Land — Plains")
	myForest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	theirPlains := pushLibraryCardForTest(opp, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})

	castHoldCreature(t, g, "Krenko's Buzzcrusher", "Artifact Creature — Insect Thopter", stdB41BuzzcrusherOracle, 4, 4)
	// The controller names a land for each player; a basic land isn't offered.
	for range 2 {
		pick := chooseCardsChoiceFor(g, me.ID)
		if pick == nil {
			t.Fatal("no destroy prompt")
		}
		for _, id := range pick.ChooseCards {
			if id == basic {
				t.Fatal("a basic land was offered")
			}
		}
		answerChooseCards(t, g, me.ID, pick.ChooseCards[0])
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, mine) || onBattlefield(g, theirs) {
		t.Fatal("a chosen nonbasic land survived")
	}
	if !onBattlefield(g, basic) {
		t.Error("the basic land was destroyed")
	}
	// Each land's controller may search; the found land enters tapped.
	for _, who := range []struct {
		p    *game.Player
		land uuid.UUID
	}{{me, myForest}, {opp, theirPlains}} {
		search := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, who.p.ID)
		if search == nil {
			t.Fatalf("no search for %s", who.p.Name)
		}
		beginSearchIfAsked(t, g, search)
		if err := g.ResolveSearchLibrary(search.ID, who.p.ID, []uuid.UUID{who.land}); err != nil {
			t.Fatalf("search: %v", err)
		}
		passPriorityAroundTable(t, g)
		if !onBattlefield(g, who.land) || !b16Tapped(t, g, who.land) {
			t.Errorf("%s's basic land is not on the battlefield tapped", who.p.Name)
		}
	}
}

func TestStdB41BuzzcrusherChoosingNoLandDestroysNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := stdB41Land(g, opp.ID, "Their Nonbasic")
	castHoldCreature(t, g, "Krenko's Buzzcrusher", "Artifact Creature — Insect Thopter", stdB41BuzzcrusherOracle, 4, 4)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, theirs) {
		t.Error("a land was destroyed though none was chosen")
	}
	if search := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, opp.ID); search != nil {
		t.Error("a search was offered though nothing was destroyed")
	}
}

// --- Case of the Locked Hothouse -------------------------------------------------

func TestStdB41HothouseGivesAnExtraLandAndSolvesAtSevenLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := seedPermanentWithOracle(g, me.ID, "Case of the Locked Hothouse", "Enchantment — Case", stdB41HothouseOracle)
	if err := b02cPlayLandRaw(t, g, "Forest"); err != nil {
		t.Fatalf("the turn's ordinary land drop: %v", err)
	}
	if err := b02cPlayLandRaw(t, g, "Island"); err != nil {
		t.Fatalf("the extra land drop was refused: %v", err)
	}
	if err := b02cPlayLandRaw(t, g, "Swamp"); err == nil {
		t.Error("a third land in one turn was allowed")
	}
	// Two lands are on the battlefield now; make it six, then seven.
	for range 4 {
		stdB41Land(g, me.ID, "Forest")
	}
	if mkmAC2EndStepSolves(t, g, id) {
		t.Fatal("solved with six lands")
	}
	stdB41Land(g, me.ID, "Forest")
	for i := 0; i < 400 && !mkmB1Solved(g, id); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if !mkmB1Solved(g, id) {
		t.Fatal("not solved with seven lands")
	}
}

func TestStdB41HothouseTopOfLibraryOnlyOnceSolved(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	id := b12Push(g, me.ID, "Case of the Locked Hothouse", "Enchantment — Case", stdB41HothouseOracle, 0, 0)
	land := seedTopOfLibrary(me, game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"})
	if topVisibleTo(g, me.ID, me.ID) {
		t.Error("the top card is visible before the Case is solved")
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("played a land off the top before the Case was solved")
	}
	mkmAC2Solve(t, g, id)
	if !topVisibleTo(g, me.ID, me.ID) || topVisibleTo(g, me.ID, opp.ID) {
		t.Error("the look should be private to the Case's controller once solved")
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("play a land off the top: %v", err)
	}
	for _, c := range []struct {
		name, typeLine string
		ok             bool
	}{
		{"Top Bear", "Creature — Bear", true},
		{"Top Aura", "Enchantment — Aura", false}, // needs a target; just check the permission
		{"Top Sorcery", "Sorcery", false},
		{"Top Artifact", "Artifact", false},
	} {
		card := seedTopOfLibrary(me, game.Card{Name: c.name, TypeLine: c.typeLine, ManaCost: "{0}"})
		if got := grantedPermissionOn(g, me.ID, card, game.ZoneLibrary).Granted(); got != (c.ok || c.name == "Top Aura") {
			t.Errorf("%s: cast permission %v", c.name, got)
		}
	}
	ench := seedTopOfLibrary(me, game.Card{Name: "Top Enchantment", TypeLine: "Enchantment", ManaCost: "{0}"})
	if err := g.CastSpell(me.ID, ench, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast an enchantment off the top: %v", err)
	}
}

// --- A Killer Among Us -----------------------------------------------------------

func stdB41KillerEnters(t *testing.T, g *game.Game, me *game.Player, choice int) uuid.UUID {
	t.Helper()
	id := castHoldCreature(t, g, "A Killer Among Us", "Enchantment", stdB41KillerOracle, 0, 0)
	passPriorityAroundTable(t, g)
	for _, name := range []string{"Human", "Merfolk", "Goblin"} {
		if tmtB1Count(g, me.ID, name) != 1 {
			t.Fatalf("no %s token", name)
		}
	}
	answerOptionPick(t, g, me.ID, choice)
	return id
}

func TestStdB41AKillerAmongUsChoiceIsSecretUntilRevealed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	killer := stdB41KillerEnters(t, g, me, 2) // Goblin
	var word string
	var by uuid.UUID
	g.WithWriteLock(func() { word, by = g.SecretChoiceOf(killer) })
	if word != "Goblin" || by != me.ID {
		t.Fatalf("stored choice %q by %v, want Goblin by me", word, by)
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventCreatureTypeChosen || ev.Kind == game.EventOptionChosen {
			t.Fatalf("the choice was announced: %+v", ev)
		}
	}
	if opp.ID == by {
		t.Fatal("test setup")
	}
}

func TestStdB41AKillerAmongUsRevealsAndPumpsOnlyTheChosenType(t *testing.T) {
	for _, tc := range []struct {
		name   string
		choice int
		token  string
		hit    bool
	}{
		{"the chosen type", 2, "Goblin", true},
		{"another type", 2, "Human", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			killer := stdB41KillerEnters(t, g, me, tc.choice)
			tokens := tmtB1Named(g, tc.token)
			if len(tokens) != 1 {
				t.Fatalf("%d %s tokens", len(tokens), tc.token)
			}
			// Tokens that entered this turn can't attack; clear sickness.
			for i := range g.Battlefield.Cards {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
			declareAttack(t, g, opp.ID, tokens[0])
			passPriorityAroundTable(t, g)
			if err := g.ActivateCatalogAbility(me.ID, killer, 0, game.ActivateAbilityParams{
				Strict: true, Targets: []game.TargetRef{cardRef(tokens[0])}}); err != nil {
				t.Fatalf("reveal: %v", err)
			}
			revealed := false
			for _, ev := range g.Events {
				if ev.Kind == game.EventCreatureTypeChosen && ev.CardID == killer && ev.Label == "Goblin" {
					revealed = true
				}
			}
			if !revealed {
				t.Error("the type was not revealed as the cost was paid")
			}
			if onBattlefield(g, killer) {
				t.Error("the enchantment was not sacrificed")
			}
			passPriorityAroundTable(t, g)
			n := countersOn(g, tokens[0], game.CounterPlusOne)
			if tc.hit && (n != 3 || !hasEffectiveKeyword(t, g, tokens[0], "deathtouch")) {
				t.Errorf("chosen type: %d counters, deathtouch %v; want 3 and deathtouch", n, hasEffectiveKeyword(t, g, tokens[0], "deathtouch"))
			}
			if !tc.hit && (n != 0 || hasEffectiveKeyword(t, g, tokens[0], "deathtouch")) {
				t.Errorf("other type got %d counters", n)
			}
		})
	}
}

func TestStdB41AKillerAmongUsOnlyTheChooserMayReveal(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	killer := stdB41KillerEnters(t, g, me, 0)
	tokens := tmtB1Named(g, "Human")
	// Another player gains control: they can't pay the reveal.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == killer {
				g.Battlefield.Cards[i].Controller = opp.ID
			}
		}
	})
	if err := g.ActivateCatalogAbility(opp.ID, killer, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{cardRef(tokens[0])}}); err == nil {
		t.Error("a player who did not choose revealed the type")
	}
}
