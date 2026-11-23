package testsst

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Rope ------------------------------------------------------------------------

func TestStdB41RopePumpsGivesReachLimitsBlockersAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rope := effectstest.B12Push(g, me.ID, "Rope", "Artifact — Clue Equipment", stdB41RopeOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	w1 := effectstest.B12Creature(g, opp.ID, "Wall One", "Creature — Wall", 0, 5)
	w2 := effectstest.B12Creature(g, opp.ID, "Wall Two", "Creature — Wall", 0, 5)

	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.P7Activate(t, g, me, rope, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.CardRef(bear)}})
	if !effectstest.AttachedTo(t, g, rope, bear) {
		t.Fatal("Rope is not attached")
	}
	if p, tough := effectstest.MkmMPPT(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("the equipped creature is %d/%d, want 3/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "reach") {
		t.Error("the equipped creature has no reach")
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: w1, Attacker: bear}, {Blocker: w2, Attacker: bear}}); err == nil {
		t.Fatal("two creatures blocked the equipped creature")
	}
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: w1, Attacker: bear}}); err != nil {
		t.Fatalf("one creature blocking it: %v", err)
	}
}

func TestStdB41RopeCracksForACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	rope := effectstest.B12Push(g, me.ID, "Rope", "Artifact — Clue Equipment", stdB41RopeOracle, 0, 0)
	effectstest.SeedLibrary(me, "One")
	hand := me.Hand.Size()
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.P7Activate(t, g, me, rope, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(rope) {
		t.Error("Rope was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// --- Red Herring -----------------------------------------------------------------

func TestStdB41RedHerringMustAttackEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	herring := effectstest.B12Push(g, me.ID, "Red Herring", "Artifact Creature — Clue Fish", stdB41HerringOracle, 2, 2)
	effectstest.AssertKeywords(t, g, herring, "haste")
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	herring := effectstest.B12Push(g, me.ID, "Red Herring", "Artifact Creature — Clue Fish", stdB41HerringOracle, 2, 2)
	effectstest.SeedLibrary(me, "One")
	hand := me.Hand.Size()
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.P7Activate(t, g, me, herring, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(herring) {
		t.Error("Red Herring was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// --- Vengeful Tracker ------------------------------------------------------------

func TestStdB41VengefulTrackerPunishesAnOpponentsArtifactSacrificeOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Vengeful Tracker", "Creature — Human Detective", stdB41TrackerOracle, 2, 2)
	theirClue := effectstest.PushToken(g, opp.ID, effects.ClueToken())
	myClue := effectstest.PushToken(g, me.ID, effects.ClueToken())
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// Their Clue, sacrificed to its own ability: 2 damage to them.
	life, mine := opp.Life, me.Life
	effectstest.SeedLibrary(opp, "One")
	effectstest.FloatMana(t, g, opp, "{C}{C}")
	if err := g.ActivateCatalogAbility(opp.ID, theirClue, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("crack their Clue: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-2)
	}
	// A creature that is not an artifact: nothing.
	life = opp.Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirBear) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("sacrificing a non-artifact cost %d life", life-opp.Life)
	}
	// My own artifact: nothing.
	effectstest.SeedLibrary(me, "One")
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, myClue, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("crack my Clue: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine || opp.Life != life {
		t.Errorf("sacrificing my own artifact moved life: me %d (was %d), them %d (was %d)", me.Life, mine, opp.Life, life)
	}
}

// --- Reckless Detective ----------------------------------------------------------

func TestStdB41RecklessDetectiveSacrificesAnArtifactToDrawAndPump(t *testing.T) {
	g, me, opp := effectstest.EoTable(t)
	det := effectstest.B12Push(g, me.ID, "Reckless Detective", "Creature — Devil Detective", stdB41DetectiveOracle, 0, 3)
	clue := effectstest.PushToken(g, me.ID, effects.ClueToken())
	effectstest.SeedLibrary(me, "One")
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, det)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EoOfferedOptions(t, g, me.ID); !got[0] || got[1] {
		t.Fatalf("offered %v, want the sacrifice only (the hand is empty)", got)
	}
	effectstest.EoPay(t, g, me.ID, 0, clue)
	if got := effectstest.EffectivePower(t, g, det); got != 2 {
		t.Errorf("power %d, want 2", got)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	if effects.OnBattlefield(g, clue) {
		t.Error("the Clue was not sacrificed")
	}
}

func TestStdB41RecklessDetectiveDiscardsACardInstead(t *testing.T) {
	g, me, opp := effectstest.EoTable(t)
	det := effectstest.B12Push(g, me.ID, "Reckless Detective", "Creature — Devil Detective", stdB41DetectiveOracle, 0, 3)
	card := effectstest.EoHand(me, "Spare", "Sorcery")
	effectstest.SeedLibrary(me, "One")
	effectstest.DeclareAttack(t, g, opp.ID, det)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EoOfferedOptions(t, g, me.ID); got[0] || !got[1] {
		t.Fatalf("offered %v, want the discard only (no artifact)", got)
	}
	effectstest.EoPay(t, g, me.ID, 1, card)
	if got := effectstest.EffectivePower(t, g, det); got != 2 {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	effectstest.CastHoldCreature(t, g, "Rubblebelt Maverick", "Creature — Human Detective", stdB41MaverickOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("surveil prompt %+v, want 2 cards", c)
	}
}

func TestStdB41RubblebeltMaverickExilesItselfFromTheGraveyardAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	grave := effectstest.GyCard(me, "Rubblebelt Maverick", "Creature — Human Detective", "{G}", stdB41MaverickOracle)

	effectstest.FloatMana(t, g, me, "{G}")
	effectstest.P7Activate(t, g, me, grave, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.CardRef(bear)}})
	if !effectstest.InExile(g, grave) {
		t.Error("the Maverick was not exiled from the graveyard")
	}
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("the bear has %d +1/+1 counters, want 1", n)
	}

	// Not at instant speed: with a spell on the stack it is refused.
	grave2 := effectstest.GyCard(me, "Rubblebelt Maverick", "Creature — Human Detective", "{G}", stdB41MaverickOracle)
	effectstest.CastCatalogSpell(t, g, "Spell", "Sorcery", "", nil)
	effectstest.FloatMana(t, g, me, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, grave2, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.CardRef(bear)}}); err == nil {
		t.Error("activated with a spell on the stack")
	}
}

// --- Sharp-Eyed Rookie -----------------------------------------------------------

func TestStdB41SharpEyedRookieGrowsAndInvestigatesForABiggerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rookie := effectstest.B12Push(g, me.ID, "Sharp-Eyed Rookie", "Creature — Human Detective", stdB41RookieOracle, 2, 2)
	effectstest.AssertKeywords(t, g, rookie, "vigilance")

	// 2/2 enters: neither stat is greater, so nothing triggers.
	effectstest.CastHoldCreature(t, g, "Same Size", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, rookie, game.CounterPlusOne); n != 0 || effectstest.TmtB1Count(g, me.ID, "Clue") != 0 {
		t.Fatalf("an equal creature gave %d counters and %d Clues", n, effectstest.TmtB1Count(g, me.ID, "Clue"))
	}
	// A 1/3: its toughness is greater.
	effectstest.CastHoldCreature(t, g, "Tall", "Creature — Wall", "", 1, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, rookie, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after a taller creature, want 1", n)
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

func TestStdB41SharpEyedRookieRechecksAsItResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rookie := effectstest.B12Push(g, me.ID, "Sharp-Eyed Rookie", "Creature — Human Detective", stdB41RookieOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
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
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, rookie, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters, want 2 (the trigger should do nothing)", n)
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 0 {
		t.Errorf("%d Clues, want 0", got)
	}
}

// --- Lamplight Phoenix -----------------------------------------------------------

func TestStdB41LamplightPhoenixReturnsTappedForFourManaValueOfEvidence(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	phoenix := effectstest.B12Push(g, me.ID, "Lamplight Phoenix", "Creature — Phoenix", stdB41PhoenixOracle, 3, 3)
	effectstest.AssertKeywords(t, g, phoenix, "flying")
	gy := effectstest.MkmCEEvidence(me, "{1}", "{3}")
	effectstest.KillCreature(t, g, me.ID, phoenix)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
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
	effectstest.AnswerChooseCards(t, g, me.ID, gy...)
	effectstest.PassPriorityAroundTable(t, g)
	// It returns as a new object (CR 400.7).
	phoenix = effectstest.FindBattlefieldByName(g, "Lamplight Phoenix")
	if phoenix == uuid.Nil {
		t.Fatal("the Phoenix did not return")
	}
	if !effectstest.B16Tapped(t, g, phoenix) {
		t.Error("the Phoenix returned untapped")
	}
	if !effectstest.InExile(g, gy[0]) || !effectstest.InExile(g, gy[1]) {
		t.Error("the evidence was not exiled")
	}
}

func TestStdB41LamplightPhoenixStaysPutWhenDeclinedOrTheGraveyardIsShort(t *testing.T) {
	t.Run("declined", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		phoenix := effectstest.B12Push(g, me.ID, "Lamplight Phoenix", "Creature — Phoenix", stdB41PhoenixOracle, 3, 3)
		gy := effectstest.MkmCEEvidence(me, "{5}")
		effectstest.KillCreature(t, g, me.ID, phoenix)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if effects.OnBattlefield(g, phoenix) || !me.Graveyard.Contains(phoenix) || !me.Graveyard.Contains(gy[0]) {
			t.Error("declining moved something")
		}
	})
	t.Run("only the Phoenix itself would make four", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		// The Phoenix is {1}{R}{R}: with a {1} beside it the total is 4, but
		// the Phoenix can't pay for itself, so the offer is not made.
		phoenix := effectstest.B12Push(g, me.ID, "Lamplight Phoenix", "Creature — Phoenix", stdB41PhoenixOracle, 3, 3)
		effectstest.MkmCEEvidence(me, "{1}")
		effectstest.KillCreature(t, g, me.ID, phoenix)
		effectstest.PassPriorityAroundTable(t, g)
		if pick := effectstest.ChooseCardsChoiceFor(g, me.ID); pick != nil {
			t.Errorf("asked with only {1} besides the Phoenix: %+v", pick)
		}
	})
}

// --- Krenko's Buzzcrusher --------------------------------------------------------

func stdB41Land(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.B12Permanent(g, owner, name, "Land")
}

func TestStdB41BuzzcrusherDestroysNonbasicLandsAndTheirControllersSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := stdB41Land(g, me.ID, "My Nonbasic")
	theirs := stdB41Land(g, opp.ID, "Their Nonbasic")
	basic := effectstest.B12Permanent(g, opp.ID, "Plains", "Basic Land — Plains")
	myForest := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	theirPlains := effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})

	effectstest.CastHoldCreature(t, g, "Krenko's Buzzcrusher", "Artifact Creature — Insect Thopter", stdB41BuzzcrusherOracle, 4, 4)
	// The controller names a land for each player; a basic land isn't offered.
	for range 2 {
		pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
		if pick == nil {
			t.Fatal("no destroy prompt")
		}
		for _, id := range pick.ChooseCards {
			if id == basic {
				t.Fatal("a basic land was offered")
			}
		}
		effectstest.AnswerChooseCards(t, g, me.ID, pick.ChooseCards[0])
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, mine) || effects.OnBattlefield(g, theirs) {
		t.Fatal("a chosen nonbasic land survived")
	}
	if !effects.OnBattlefield(g, basic) {
		t.Error("the basic land was destroyed")
	}
	// Each land's controller may search; the found land enters tapped.
	for _, who := range []struct {
		p    *game.Player
		land uuid.UUID
	}{{me, myForest}, {opp, theirPlains}} {
		search := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, who.p.ID)
		if search == nil {
			t.Fatalf("no search for %s", who.p.Name)
		}
		effectstest.BeginSearchIfAsked(t, g, search)
		if err := g.ResolveSearchLibrary(search.ID, who.p.ID, []uuid.UUID{who.land}); err != nil {
			t.Fatalf("search: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effects.OnBattlefield(g, who.land) || !effectstest.B16Tapped(t, g, who.land) {
			t.Errorf("%s's basic land is not on the battlefield tapped", who.p.Name)
		}
	}
}

func TestStdB41BuzzcrusherChoosingNoLandDestroysNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := stdB41Land(g, opp.ID, "Their Nonbasic")
	effectstest.CastHoldCreature(t, g, "Krenko's Buzzcrusher", "Artifact Creature — Insect Thopter", stdB41BuzzcrusherOracle, 4, 4)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, theirs) {
		t.Error("a land was destroyed though none was chosen")
	}
	if search := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, opp.ID); search != nil {
		t.Error("a search was offered though nothing was destroyed")
	}
}

// --- Case of the Locked Hothouse -------------------------------------------------

func TestStdB41HothouseGivesAnExtraLandAndSolvesAtSevenLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.SeedPermanentWithOracle(g, me.ID, "Case of the Locked Hothouse", "Enchantment — Case", stdB41HothouseOracle)
	if err := effectstest.B02cPlayLandRaw(t, g, "Forest"); err != nil {
		t.Fatalf("the turn's ordinary land drop: %v", err)
	}
	if err := effectstest.B02cPlayLandRaw(t, g, "Island"); err != nil {
		t.Fatalf("the extra land drop was refused: %v", err)
	}
	if err := effectstest.B02cPlayLandRaw(t, g, "Swamp"); err == nil {
		t.Error("a third land in one turn was allowed")
	}
	// Two lands are on the battlefield now; make it six, then seven.
	for range 4 {
		stdB41Land(g, me.ID, "Forest")
	}
	if effectstest.MkmAC2EndStepSolves(t, g, id) {
		t.Fatal("solved with six lands")
	}
	stdB41Land(g, me.ID, "Forest")
	for i := 0; i < 400 && !effectstest.MkmB1Solved(g, id); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !effectstest.MkmB1Solved(g, id) {
		t.Fatal("not solved with seven lands")
	}
}

func TestStdB41HothouseTopOfLibraryOnlyOnceSolved(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.ToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Case of the Locked Hothouse", "Enchantment — Case", stdB41HothouseOracle, 0, 0)
	land := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"})
	if effectstest.TopVisibleTo(g, me.ID, me.ID) {
		t.Error("the top card is visible before the Case is solved")
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("played a land off the top before the Case was solved")
	}
	effectstest.MkmAC2Solve(t, g, id)
	if !effectstest.TopVisibleTo(g, me.ID, me.ID) || effectstest.TopVisibleTo(g, me.ID, opp.ID) {
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
		card := effectstest.SeedTopOfLibrary(me, game.Card{Name: c.name, TypeLine: c.typeLine, ManaCost: "{0}"})
		if got := effectstest.GrantedPermissionOn(g, me.ID, card, game.ZoneLibrary).Granted(); got != (c.ok || c.name == "Top Aura") {
			t.Errorf("%s: cast permission %v", c.name, got)
		}
	}
	ench := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Enchantment", TypeLine: "Enchantment", ManaCost: "{0}"})
	if err := g.CastSpell(me.ID, ench, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast an enchantment off the top: %v", err)
	}
}

// --- A Killer Among Us -----------------------------------------------------------

func stdB41KillerEnters(t *testing.T, g *game.Game, me *game.Player, choice int) uuid.UUID {
	t.Helper()
	id := effectstest.CastHoldCreature(t, g, "A Killer Among Us", "Enchantment", stdB41KillerOracle, 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for _, name := range []string{"Human", "Merfolk", "Goblin"} {
		if effectstest.TmtB1Count(g, me.ID, name) != 1 {
			t.Fatalf("no %s token", name)
		}
	}
	effectstest.AnswerOptionPick(t, g, me.ID, choice)
	return id
}

func TestStdB41AKillerAmongUsChoiceIsSecretUntilRevealed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
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
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			killer := stdB41KillerEnters(t, g, me, tc.choice)
			tokens := effectstest.TmtB1Named(g, tc.token)
			if len(tokens) != 1 {
				t.Fatalf("%d %s tokens", len(tokens), tc.token)
			}
			// Tokens that entered this turn can't attack; clear sickness.
			for i := range g.Battlefield.Cards {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
			effectstest.DeclareAttack(t, g, opp.ID, tokens[0])
			effectstest.PassPriorityAroundTable(t, g)
			if err := g.ActivateCatalogAbility(me.ID, killer, 0, game.ActivateAbilityParams{
				Strict: true, Targets: []game.TargetRef{effectstest.CardRef(tokens[0])}}); err != nil {
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
			if effects.OnBattlefield(g, killer) {
				t.Error("the enchantment was not sacrificed")
			}
			effectstest.PassPriorityAroundTable(t, g)
			n := effectstest.CountersOn(g, tokens[0], game.CounterPlusOne)
			if tc.hit && (n != 3 || !effectstest.HasEffectiveKeyword(t, g, tokens[0], "deathtouch")) {
				t.Errorf("chosen type: %d counters, deathtouch %v; want 3 and deathtouch", n, effectstest.HasEffectiveKeyword(t, g, tokens[0], "deathtouch"))
			}
			if !tc.hit && (n != 0 || effectstest.HasEffectiveKeyword(t, g, tokens[0], "deathtouch")) {
				t.Errorf("other type got %d counters", n)
			}
		})
	}
}

func TestStdB41AKillerAmongUsOnlyTheChooserMayReveal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	killer := stdB41KillerEnters(t, g, me, 0)
	tokens := effectstest.TmtB1Named(g, "Human")
	// Another player gains control: they can't pay the reveal.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == killer {
				g.Battlefield.Cards[i].Controller = opp.ID
			}
		}
	})
	if err := g.ActivateCatalogAbility(opp.ID, killer, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{effectstest.CardRef(tokens[0])}}); err == nil {
		t.Error("a player who did not choose revealed the type")
	}
}
