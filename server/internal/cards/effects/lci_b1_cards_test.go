package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lci_b1_cards_test.go — The Lost Caverns of Ixalan, batch 1 (#4470).
// Each test plays the real catalog entry and asserts what its text says.

const (
	lciB1GorestalkerOracle   = "465cda20-c8f3-4a7c-a250-875877bce21d"
	lciB1AcolyteOracle       = "324d2aa8-6dea-4b0d-898d-1e52ed187565"
	lciB1LeapOracle          = "1347224f-421e-4270-8e50-cf25c875909c"
	lciB1ReminiscenceOracle  = "dbfa2743-34ff-40e3-81f7-740fc08cd18a"
	lciB1SunscribeOracle     = "0f85714b-e731-45a3-aa65-8cd59cd48280"
	lciB1BlademasterOracle   = "43c92618-0aae-41f8-aa39-393e5d6b5ff7"
	lciB1CompanionOracle     = "460c0f0b-bbf4-4307-9b44-350175d83bfe"
	lciB1StomperOracle       = "94657522-9977-4d2b-8eb6-f1e64e82cd0f"
	lciB1ScoutOracle         = "ec3911c0-686f-4c78-838d-a683eed6c01c"
	lciB1WrestlerOracle      = "db27e3f0-5750-4c4b-9da9-70c3ab627f9c"
	lciB1ColossadactylOracle = "34ff281a-f294-4351-9665-863a5b2e882b"
	lciB1CosmiumBlastOracle  = "52c96b5a-ce47-4bde-b530-0033b9bfd808"
)

func TestLciB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		lciB1GorestalkerOracle:   "Abyssal Gorestalker",
		lciB1AcolyteOracle:       "Acolyte of Aclazotz",
		lciB1LeapOracle:          "Acrobatic Leap",
		lciB1ReminiscenceOracle:  "Ancestral Reminiscence",
		lciB1SunscribeOracle:     "Attentive Sunscribe",
		lciB1BlademasterOracle:   "Brazen Blademaster",
		lciB1CompanionOracle:     "Cartographer's Companion",
		lciB1StomperOracle:       "Cavern Stomper",
		lciB1ScoutOracle:         "Cenote Scout",
		lciB1WrestlerOracle:      "Cogwork Wrestler",
		lciB1ColossadactylOracle: "Colossadactyl",
		lciB1CosmiumBlastOracle:  "Cosmium Blast",
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

// --- fixtures ------------------------------------------------------------------

// lciB1Attack declares the attacker against defender and locks the
// declaration in, so "whenever this attacks" triggers are harvested, then
// resolves what they put on the stack.
func lciB1Attack(t *testing.T, g *game.Game, attacker, defender uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
}

// lciB1MapsControlledBy counts the Map tokens a player controls.
func lciB1MapsControlledBy(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Map" && c.Controller == player {
			n++
		}
	}
	return n
}

// --- Abyssal Gorestalker ---------------------------------------------------------

func TestLciB1GorestalkerMakesEachPlayerSacrificeTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	mineKept := b12Creature(g, me.ID, "My Kept Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	x := b12Creature(g, opp1.ID, "Theirs X", "Creature — Bear", 2, 2)
	y := b12Creature(g, opp1.ID, "Theirs Y", "Creature — Bear", 2, 2)
	kept := b12Creature(g, opp1.ID, "Theirs Kept", "Creature — Bear", 2, 2)
	lone := b12Creature(g, opp2.ID, "Lone Z", "Creature — Bear", 2, 2)
	land := b12Permanent(g, opp3.ID, "Their Land", "Basic Land — Swamp")

	stalker, err := msc17Cast(t, g, me, game.Card{Name: "Abyssal Gorestalker", TypeLine: "Creature — Horror", OracleID: lciB1GorestalkerOracle, Power: 6, Toughness: 6}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast Abyssal Gorestalker: %v", err)
	}
	passPriorityAroundTable(t, g)

	c := sacrificeChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the Gorestalker's controller is not asked to sacrifice")
	}
	if !edictOffers(sacrificeOptions(c), stalker) {
		t.Error("the Gorestalker itself is not offered to its own trigger")
	}
	if edictOffers(sacrificeOptions(c), rock) {
		t.Error("a noncreature is offered")
	}
	answerSacrifice(t, g, me.ID, stalker, mine)
	answerSacrifice(t, g, opp1.ID, x, y)
	// A player with one creature owes it without being asked (ADR 0175).
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{stalker, mine, x, y, lone} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", id)
		}
	}
	for _, id := range []uuid.UUID{mineKept, kept, rock, land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s left the battlefield but was not a chosen creature", id)
		}
	}
	if sacrificeChoiceFor(g, opp3.ID) != nil {
		t.Error("a player with no creature is asked")
	}
}

// --- Acolyte of Aclazotz ---------------------------------------------------------

func TestLciB1AcolyteSacrificesAnotherCreatureOrArtifactToDrain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	acolyte := b12Push(g, me.ID, "Acolyte of Aclazotz", "Creature — Vampire Cleric", lciB1AcolyteOracle, 1, 4)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	land := b12Permanent(g, me.ID, "My Land", "Basic Land — Swamp")
	advanceToMain(t, g)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}

	if err := g.ActivateCatalogAbility(me.ID, acolyte, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{acolyte}}); err == nil {
		t.Fatal("the Acolyte paid with itself; the cost is another creature or artifact")
	}
	if err := g.ActivateCatalogAbility(me.ID, acolyte, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("a land paid the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, acolyte, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("activate with an artifact: %v", err)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed as the cost")
	}
	if !tappedNow(g, acolyte) {
		t.Error("the Acolyte did not tap")
	}
	passPriorityAroundTable(t, g)

	if me.Life != lives[me.ID]+1 {
		t.Errorf("controller life %d, want %d", me.Life, lives[me.ID]+1)
	}
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		if p.Life != lives[p.ID]-1 {
			t.Errorf("%s at %d, want %d", p.Name, p.Life, lives[p.ID]-1)
		}
	}
}

// --- Acrobatic Leap --------------------------------------------------------------

func TestLciB1AcrobaticLeapPumpsGrantsFlyingAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	c := seedCreature(g, "Bear", g.Seats[0].ID)
	avB4SetTapped(g, c)
	castCatalogSpell(t, g, "Acrobatic Leap", "Instant", lciB1LeapOracle, rfCardTarget(c))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, c), effectiveToughness(t, g, c); p != 3 || tough != 5 {
		t.Errorf("P/T = %d/%d, want 3/5 (a 2/2 gets +1/+3)", p, tough)
	}
	if !hasEffectiveKeyword(t, g, c, "flying") {
		t.Error("the creature did not gain flying")
	}
	if tappedNow(g, c) {
		t.Error("the creature was not untapped")
	}
}

// --- Ancestral Reminiscence ------------------------------------------------------

func TestLciB1AncestralReminiscenceDrawsThreeThenDiscardsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Ancestral Reminiscence", "Sorcery", lciB1ReminiscenceOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Fatalf("drew %d cards before the discard, want 3", got)
	}
	if got := discardOwed(g, me.ID); got != 1 {
		t.Fatalf("the discard prompt asks for %d cards, want 1", got)
	}
	pitched := me.Hand.Cards[0].InstanceID
	answerDiscard(t, g, me.ID, pitched)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitched) {
		t.Error("the chosen card was not discarded")
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand changed by %d, want +2 (three drawn, one discarded)", got)
	}
}

// --- Attentive Sunscribe ---------------------------------------------------------

func TestLciB1AttentiveSunscribeScriesWhenItBecomesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scribe := b12Push(g, me.ID, "Attentive Sunscribe", "Artifact Creature — Gnome", lciB1SunscribeOracle, 2, 2)
	advanceToMain(t, g)
	if scryChoiceFor(g, me.ID) != nil {
		t.Fatal("a scry is owed before the Sunscribe has tapped")
	}
	lciB1Attack(t, g, scribe, opp.ID)
	sc := scryChoiceFor(g, me.ID)
	if sc == nil {
		t.Fatal("attacking tapped the Sunscribe: a scry 1 is owed")
	}
	if len(sc.ScryCards) != 1 {
		t.Errorf("scry looks at %d cards, want 1", len(sc.ScryCards))
	}
}

// --- Brazen Blademaster ----------------------------------------------------------

func TestLciB1BlademasterGetsBonusOnlyWithTwoArtifacts(t *testing.T) {
	setup := func(t *testing.T, artifacts int) (*game.Game, uuid.UUID) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		blade := b12Push(g, me.ID, "Brazen Blademaster", "Creature — Orc Pirate", lciB1BlademasterOracle, 2, 3)
		for i := 0; i < artifacts; i++ {
			b12Permanent(g, me.ID, "Rock", "Artifact")
		}
		b12Permanent(g, opp.ID, "Their Rock", "Artifact")
		lciB1Attack(t, g, blade, opp.ID)
		return g, blade
	}
	t.Run("two artifacts", func(t *testing.T) {
		g, blade := setup(t, 2)
		if p, tough := effectivePower(t, g, blade), effectiveToughness(t, g, blade); p != 4 || tough != 4 {
			t.Errorf("P/T = %d/%d, want 4/4", p, tough)
		}
	})
	t.Run("one artifact", func(t *testing.T) {
		g, blade := setup(t, 1)
		if p, tough := effectivePower(t, g, blade), effectiveToughness(t, g, blade); p != 2 || tough != 3 {
			t.Errorf("P/T = %d/%d with one artifact, want 2/3 (an opponent's artifact does not count)", p, tough)
		}
	})
}

// --- Cartographer's Companion ----------------------------------------------------

func TestLciB1CartographersCompanionMakesAMap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc17Play(t, g, me, game.Card{Name: "Cartographer's Companion", TypeLine: "Artifact Creature — Gnome", OracleID: lciB1CompanionOracle, Power: 2, Toughness: 1})
	passPriorityAroundTable(t, g)
	if got := lciB1MapsControlledBy(g, me.ID); got != 1 {
		t.Errorf("controller has %d Map tokens, want 1", got)
	}
	for _, p := range g.Seats[1:] {
		if got := lciB1MapsControlledBy(g, p.ID); got != 0 {
			t.Errorf("%s has %d Map tokens, want 0", p.Name, got)
		}
	}
}

// --- Cavern Stomper --------------------------------------------------------------

func TestLciB1CavernStomperScriesTwoOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc17Play(t, g, me, game.Card{Name: "Cavern Stomper", TypeLine: "Creature — Dinosaur", OracleID: lciB1StomperOracle, Power: 7, Toughness: 7})
	passPriorityAroundTable(t, g)
	sc := scryChoiceFor(g, me.ID)
	if sc == nil {
		t.Fatal("no scry when Cavern Stomper entered")
	}
	if len(sc.ScryCards) != 2 {
		t.Errorf("scry looks at %d cards, want 2", len(sc.ScryCards))
	}
}

func TestLciB1CavernStomperBarsLowPowerBlockersThisTurn(t *testing.T) {
	setup := func(t *testing.T, activate bool) (g *game.Game, stomper, small, big uuid.UUID) {
		g = newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		stomper = b12Push(g, me.ID, "Cavern Stomper", "Creature — Dinosaur", lciB1StomperOracle, 7, 7)
		small = pushVanillaCreature(g, opp.ID, "Power Two", 2, 2)
		big = pushVanillaCreature(g, opp.ID, "Power Three", 3, 3)
		advanceToMain(t, g)
		if activate {
			fillPoolColored(me, "G", 3)
			if err := g.ActivateCatalogAbility(me.ID, stomper, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatal("three mana paid {3}{G}")
			}
			fillPoolColored(me, "G", 1)
			if err := g.ActivateCatalogAbility(me.ID, stomper, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
		}
		lciB1Attack(t, g, stomper, opp.ID)
		advanceTo(t, g, game.StepDeclareBlockers)
		return
	}
	t.Run("activated", func(t *testing.T) {
		g, stomper, small, big := setup(t, true)
		brRefusal(t, g.DeclareBlocker(small, stomper), game.BlockReasonCantBeBlockedBy)
		if err := g.DeclareBlocker(big, stomper); err != nil {
			t.Fatalf("a power 3 creature blocks: %v", err)
		}
	})
	t.Run("not activated", func(t *testing.T) {
		g, stomper, small, _ := setup(t, false)
		if err := g.DeclareBlocker(small, stomper); err != nil {
			t.Fatalf("without the ability a power 2 creature blocks: %v", err)
		}
	})
}

// --- Cenote Scout ----------------------------------------------------------------

func TestLciB1CenoteScoutExploresOnEntry(t *testing.T) {
	t.Run("a land goes to the hand", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		top := putOnTopForTest(me, "Forest", "Basic Land — Forest")
		scout := msc17Play(t, g, me, game.Card{Name: "Cenote Scout", TypeLine: "Creature — Merfolk Scout", OracleID: lciB1ScoutOracle, Power: 1, Toughness: 1})
		passPriorityAroundTable(t, g)
		if !me.Hand.Contains(top) {
			t.Error("the revealed land did not go to the hand")
		}
		if got := countersOn(g, scout, game.CounterPlusOne); got != 0 {
			t.Errorf("the Scout has %d +1/+1 counters after revealing a land, want 0", got)
		}
	})
	t.Run("a nonland gives a counter and may be binned", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		top := putOnTopForTest(me, "Some Sorcery", "Sorcery")
		scout := msc17Play(t, g, me, game.Card{Name: "Cenote Scout", TypeLine: "Creature — Merfolk Scout", OracleID: lciB1ScoutOracle, Power: 1, Toughness: 1})
		passPriorityAroundTable(t, g)
		if got := countersOn(g, scout, game.CounterPlusOne); got != 1 {
			t.Errorf("the Scout has %d +1/+1 counters, want 1", got)
		}
		answerExploreQuestion(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
		if !me.Graveyard.Contains(top) {
			t.Error("the revealed card was not put into the graveyard")
		}
	})
}

// --- Cogwork Wrestler ------------------------------------------------------------

func TestLciB1CogworkWrestlerShrinksAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 3, 3)
	bystander := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if _, err := msc17Cast(t, g, me, game.Card{Name: "Cogwork Wrestler", TypeLine: "Artifact Creature — Gnome", OracleID: lciB1WrestlerOracle, Power: 1, Toughness: 2}, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Cogwork Wrestler: %v", err)
	}
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target pick for the enters trigger with two opposing creatures")
	}
	for _, c := range p.PickTargetCards {
		if c == mine {
			t.Error("the controller's own creature is offered")
		}
	}
	pickIfAsked(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, victim), effectiveToughness(t, g, victim); p != 1 || tough != 3 {
		t.Errorf("the target is %d/%d, want 1/3 (a 3/3 gets -2/-0)", p, tough)
	}
	if got := effectivePower(t, g, bystander); got != 2 {
		t.Errorf("an untargeted creature's power %d, want 2", got)
	}
	if got := effectivePower(t, g, mine); got != 2 {
		t.Errorf("my creature's power %d, want 2", got)
	}
}

func TestLciB1CogworkWrestlerHasFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepBeginCombat)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Cogwork Wrestler", TypeLine: "Artifact Creature — Gnome", OracleID: lciB1WrestlerOracle, Power: 1, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast at the beginning of combat: %v", err)
	}
}

// --- Colossadactyl ---------------------------------------------------------------

func TestLciB1ColossadactylHasReachAndTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Colossadactyl", "Creature — Dinosaur", lciB1ColossadactylOracle, 4, 5)
	advanceToMain(t, g)
	for _, kw := range []string{"reach", "trample"} {
		if !hasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("Colossadactyl has no %s", kw)
		}
	}
}

// --- Cosmium Blast ---------------------------------------------------------------

func TestLciB1CosmiumBlastDealsFourToAnAttackerOrBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, opp.ID, "Their Attacker", "Creature — Bear", 4, 4)
	if err := castCatalogSpellErr(t, g, "Cosmium Blast", "Instant", lciB1CosmiumBlastOracle, pr6Card(attacker)); err == nil {
		t.Fatal("Cosmium Blast targeted a creature that is neither attacking nor blocking")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == attacker {
				g.Battlefield.Cards[i].AttackingTarget = me.ID
			}
		}
	})
	castCatalogSpell(t, g, "Cosmium Blast", "Instant", lciB1CosmiumBlastOracle, pr6Card(attacker))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("4 damage did not destroy a 4/4 attacker")
	}
	if !opp.Graveyard.Contains(attacker) {
		t.Error("the attacker is not in its owner's graveyard")
	}
}
