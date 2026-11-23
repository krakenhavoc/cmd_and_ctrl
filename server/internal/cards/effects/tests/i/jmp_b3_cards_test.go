package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jumpstart deck request #4892 (Dimir), cards batch 3.

const (
	jmpB3SilversmoteGhoulOracle = "67d55b4f-944e-4376-93ed-1a240bca1eae"
	jmpB3SwarmOfBloodfliesOracl = "767e4aa9-527f-425f-bffb-6d842e1fbb74"
	jmpB3TeferisTutelageOracle  = "c6b94e03-61c0-4077-b3da-a70795e97278"
	jmpB3ThoughtCollapseOracle  = "6a265af8-448d-402b-b728-b7ed64521842"
	jmpB3TolarianKrakenOracle   = "846887e4-744f-44ce-9110-d09680b11e95"
	jmpB3TomeAnimaOracle        = "e4a68bb3-5a79-4764-8223-acddd16fb660"
	jmpB3UnsubstantiateOracle   = "c387f46f-ad91-4964-8dc6-acefe537fe3c"
	jmpB3WallOfLostThoughtsOrcl = "49b80368-d38a-4877-b1de-f3b5f3d40d4d"
	jmpB3WightOfPrecinctSixOrcl = "6397c046-4c59-4f0b-9b44-2a804eb95edf"
	jmpB3WitchsCauldronOracle   = "cb65ee08-bfdb-4c18-8967-8bcf31fcabfa"
	jmpB3ZombieInfestationOracl = "bdce1af0-3643-4e77-88f9-320206a191d4"
)

func TestJmpB3EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		jmpB3SilversmoteGhoulOracle: "Silversmote Ghoul",
		jmpB3SwarmOfBloodfliesOracl: "Swarm of Bloodflies",
		jmpB3TeferisTutelageOracle:  "Teferi's Tutelage",
		jmpB3ThoughtCollapseOracle:  "Thought Collapse",
		jmpB3TolarianKrakenOracle:   "Tolarian Kraken",
		jmpB3TomeAnimaOracle:        "Tome Anima",
		jmpB3UnsubstantiateOracle:   "Unsubstantiate",
		jmpB3WallOfLostThoughtsOrcl: "Wall of Lost Thoughts",
		jmpB3WightOfPrecinctSixOrcl: "Wight of Precinct Six",
		jmpB3WitchsCauldronOracle:   "Witch's Cauldron",
		jmpB3ZombieInfestationOracl: "Zombie Infestation",
	}
	for oracle, name := range want {
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

// --- Silversmote Ghoul ----------------------------------------------------

func jmpB3GhoulInGraveyard(owner *game.Player) uuid.UUID {
	id := uuid.New()
	owner.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Silversmote Ghoul", TypeLine: "Creature — Zombie Vampire",
		ManaCost: "{2}{B}", OracleID: jmpB3SilversmoteGhoulOracle, Power: 3, Toughness: 1,
		Owner: owner.ID, Controller: owner.ID,
	})
	return id
}

func TestJmpB3SilversmoteGhoulReturnsTappedAfterYouGainedThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ghoul := jmpB3GhoulInGraveyard(me)
	effectstest.AdvanceToMain(t, g)
	effectstest.JmpB3GainLife(g, me.ID, 1)
	effectstest.JmpB3GainLife(g, me.ID, 2)
	// Losing it back does not undo the gain: the card asks what you gained.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -3) })
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(ghoul) {
		t.Fatal("the Ghoul stayed in the graveyard after you gained 3")
	}
	back := effectstest.FindBattlefieldByName(g, "Silversmote Ghoul")
	if back == uuid.Nil {
		t.Fatal("the Ghoul is not on the battlefield")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, back); !tapped {
		t.Error("the Ghoul returned untapped")
	}
}

func TestJmpB3SilversmoteGhoulStaysWithTwoLifeGainedOrOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghoul := jmpB3GhoulInGraveyard(me)
	effectstest.AdvanceToMain(t, g)
	effectstest.JmpB3GainLife(g, me.ID, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if n := effectstest.TriggersOnStackFrom(g, ghoul); n != 0 {
		t.Fatalf("2 life gained: %d triggers", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(ghoul) {
		t.Fatal("the Ghoul returned with only 2 life gained")
	}
	// The opponent's end step, after you gained 3 on their turn.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.JmpB3GainLife(g, me.ID, 3)
	effectstest.JmpB3GainLife(g, opp.ID, 3)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if n := effectstest.TriggersOnStackFrom(g, ghoul); n != 0 {
		t.Fatalf("an opponent's end step: %d triggers", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(ghoul) {
		t.Error("the Ghoul returned at an opponent's end step")
	}
}

func TestJmpB3SilversmoteGhoulDoesNothingIfItLeftTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ghoul := jmpB3GhoulInGraveyard(me)
	effectstest.AdvanceToMain(t, g)
	effectstest.JmpB3GainLife(g, me.ID, 3)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if n := effectstest.TriggersOnStackFrom(g, ghoul); n != 1 {
		t.Fatalf("%d triggers at your end step, want 1", n)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneExile}, ghoul); err != nil {
		t.Fatalf("exile the Ghoul: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Silversmote Ghoul") != uuid.Nil {
		t.Error("the Ghoul came back from exile")
	}
}

func TestJmpB3SilversmoteGhoulSacrificesToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ghoul := effectstest.B21Push(g, me.ID, "Silversmote Ghoul", "Creature — Zombie Vampire", jmpB3SilversmoteGhoulOracle, 3, 1, "B")
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.B06AddMana(me, "C", "B")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
	if !me.Graveyard.Contains(ghoul) {
		t.Error("the Ghoul was not sacrificed")
	}
}

// --- Swarm of Bloodflies --------------------------------------------------

func TestJmpB3SwarmOfBloodfliesEntersWithTwoAndGrowsOnOtherDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	swarm := effectstest.CastCatalogSpell(t, g, "Swarm of Bloodflies", "Creature — Insect", jmpB3SwarmOfBloodfliesOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, swarm, "+1/+1"); n != 2 {
		t.Fatalf("entered with %d +1/+1 counters, want 2", n)
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, swarm); p != 2 {
		t.Fatalf("entered with power %d, want 2", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, swarm, "flying") {
		t.Error("no flying")
	}
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	effectstest.B25Destroy(g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B25Destroy(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, swarm, "+1/+1"); n != 4 {
		t.Errorf("+1/+1 counters %d, want 4 (two on entry, one per death)", n)
	}
	// A noncreature leaving does nothing.
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	effectstest.B25Destroy(g, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, swarm, "+1/+1"); n != 4 {
		t.Errorf("an artifact died: counters %d, want 4", n)
	}
}

// --- Teferi's Tutelage ----------------------------------------------------

func TestJmpB3TeferisTutelageLootsOnEntryAndEachDrawMillsAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := me.Hand.Size()
	library := opp.Library.Size()
	tutelage := effectstest.CastCatalogSpell(t, g, "Teferi's Tutelage", "Enchantment", jmpB3TeferisTutelageOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Fatalf("hand %d → %d, owes %d; want +1 and 1 discard owed",
			hand, me.Hand.Size(), effectstest.DiscardOwed(g, me.ID))
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	// The loot's draw is a draw: it triggers the mill.
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); effectstest.HasID(p.PickTargetPlayers, me.ID) {
		t.Error("target OPPONENT offered you")
	}
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != library-2 {
		t.Fatalf("opponent library %d → %d, want -2", library, opp.Library.Size())
	}
	// Two draws, two triggers.
	effectstest.B25Draw(g, me.ID, 2)
	for i := 0; i < 2; i++ {
		effectstest.B04WaitForPick(t, g, me.ID)
		effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != library-6 {
		t.Errorf("after two draws the library is %d, want %d", opp.Library.Size(), library-6)
	}
	// An opponent's draw does nothing.
	effectstest.B25Draw(g, opp.ID, 1)
	if n := effectstest.TriggersOnStackFrom(g, tutelage); n != 0 || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Errorf("an opponent's draw triggered the Tutelage")
	}
}

// --- Thought Collapse -----------------------------------------------------

func TestJmpB3ThoughtCollapseCountersAndTheControllerMillsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	library := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Thought Collapse", "Instant", jmpB3ThoughtCollapseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Fatal("the Bolt was not countered")
	}
	if opp.Library.Size() != library-3 {
		t.Errorf("the Bolt's controller milled %d, want 3", library-opp.Library.Size())
	}
}

func TestJmpB3ThoughtCollapseCannotTargetAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Thought Collapse", "Instant", jmpB3ThoughtCollapseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("targeted a creature on the battlefield")
	}
}

// --- Tolarian Kraken ------------------------------------------------------

func TestJmpB3TolarianKrakenPaysOneToTapOrUntapACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B21Push(g, me.ID, "Tolarian Kraken", "Creature — Kraken", jmpB3TolarianKrakenOracle, 4, 6, "U")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B25Draw(g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("your draw asks you to pay {1}")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.MonSettle(t, g)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.MonSettle(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // tap it
	if tapped, _ := effectstest.BattlefieldCardTapped(g, bear); !tapped {
		t.Fatal("the Bear was not tapped")
	}
	// Again, and untap it this time.
	effectstest.B25Draw(g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.MonSettle(t, g)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.MonSettle(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, bear); tapped {
		t.Error("the Bear was not untapped")
	}
}

func TestJmpB3TolarianKrakenDeclinedOrOpponentDrawDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B21Push(g, me.ID, "Tolarian Kraken", "Creature — Kraken", jmpB3TolarianKrakenOracle, 4, 6, "U")
	effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B25Draw(g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.MonSettle(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("declined: a target was still asked")
	}
	effectstest.B25Draw(g, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's draw asked you to pay")
	}
}

// --- Tome Anima -----------------------------------------------------------

func TestJmpB3TomeAnimaIsUnblockableOnceYouHaveDrawnTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attacker, defender := g.Seats[0], g.Seats[1]
	anima := effectstest.B12Push(g, attacker.ID, "Tome Anima", "Creature — Spirit", jmpB3TomeAnimaOracle, 3, 3)
	blocker := effectstest.B12Creature(g, defender.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.BrAttack(t, g, anima)
	drawn := g.TurnTallyFor(attacker.ID).CardsDrawn
	if drawn >= 2 {
		t.Fatalf("the harness drew %d before the test", drawn)
	}
	if err := g.Clone().DeclareBlocker(blocker, anima); err != nil {
		t.Fatalf("%d card(s) drawn: the block was refused: %v", drawn, err)
	}
	effectstest.B25Draw(g, attacker.ID, 2-drawn)
	effectstest.BrRefusal(t, g.Clone().DeclareBlocker(blocker, anima), game.BlockReasonCantBeBlocked)
}

// --- Unsubstantiate -------------------------------------------------------

func TestJmpB3UnsubstantiateReturnsASpellToItsOwnersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Unsubstantiate", "Instant", jmpB3UnsubstantiateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bolt) {
		t.Fatal("the Bolt is not in its owner's hand")
	}
	if me.Life != life {
		t.Errorf("the Bolt resolved anyway: life %d → %d", life, me.Life)
	}
}

func TestJmpB3UnsubstantiateBouncesACreatureButNotANoncreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Unsubstantiate", "Instant", jmpB3UnsubstantiateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: relic}}); err == nil {
		t.Error("targeted a noncreature artifact")
	}
	effectstest.CastCatalogSpell(t, g, "Unsubstantiate", "Instant", jmpB3UnsubstantiateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bear) {
		t.Error("the Bear is not in its owner's hand")
	}
}

// --- Wall of Lost Thoughts ------------------------------------------------

func TestJmpB3WallOfLostThoughtsMillsAnyTargetPlayerFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	library := me.Library.Size()
	wall := effectstest.CastCatalogSpell(t, g, "Wall of Lost Thoughts", "Creature — Wall", jmpB3WallOfLostThoughtsOrcl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); !effectstest.HasID(p.PickTargetPlayers, me.ID) || !effectstest.HasID(p.PickTargetPlayers, g.Seats[1].ID) {
		t.Errorf("target PLAYER: offered %v", p.PickTargetPlayers)
	}
	effectstest.B16PickPlayer(t, g, me.ID, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != library-4 {
		t.Errorf("library %d → %d, want -4", library, me.Library.Size())
	}
	if !effectstest.HasEffectiveKeyword(t, g, wall, "defender") {
		t.Error("no defender")
	}
}

// --- Wight of Precinct Six ------------------------------------------------

func TestJmpB3WightOfPrecinctSixCountsOnlyOpponentsCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wight := effectstest.B21Push(g, me.ID, "Wight of Precinct Six", "Creature — Zombie", jmpB3WightOfPrecinctSixOrcl, 1, 1, "B")
	if p, tough := effectstest.EffectivePTOf(t, g, wight); p != 1 || tough != 1 {
		t.Fatalf("empty graveyards: %d/%d, want 1/1", p, tough)
	}
	effectstest.PushGraveyardCardWithTypeLine(g, me.ID, "Mine", "Creature — Bear")
	effectstest.PushGraveyardCardWithTypeLine(g, g.Seats[1].ID, "Elf", "Creature — Elf")
	effectstest.PushGraveyardCardWithTypeLine(g, g.Seats[2].ID, "Goblin", "Creature — Goblin")
	effectstest.PushGraveyardCardWithTypeLine(g, g.Seats[2].ID, "Golem", "Artifact Creature — Golem")
	effectstest.PushGraveyardCardWithTypeLine(g, g.Seats[3].ID, "Bolt", "Instant")
	if p, tough := effectstest.EffectivePTOf(t, g, wight); p != 4 || tough != 4 {
		t.Errorf("three opposing creature cards: %d/%d, want 4/4", p, tough)
	}
}

// --- Witch's Cauldron -----------------------------------------------------

func TestJmpB3WitchsCauldronSacrificesForALifeAndACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	cauldron := effectstest.PushCatalogPermanent(g, me.ID, "Witch's Cauldron", "Artifact", jmpB3WitchsCauldronOracle, false)
	effectstest.B06AddMana(me, "C", "B")
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no creature to sacrifice")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand, life := me.Hand.Size(), me.Life
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{bear},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d → %d, hand %d → %d; want +1 each", life, me.Life, hand, me.Hand.Size())
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, cauldron); !tapped {
		t.Error("the Cauldron was not tapped")
	}
	// Tapped: no second activation.
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	effectstest.B06AddMana(me, "C", "B")
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{elf},
	}); err == nil {
		t.Error("activated a tapped Cauldron")
	}
}

// --- Zombie Infestation ---------------------------------------------------

func TestJmpB3ZombieInfestationDiscardsTwoForAZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	for _, c := range me.Hand.Cards {
		cID := c.InstanceID
		_ = g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, cID)
	}
	infestation := effectstest.PushCatalogPermanent(g, me.ID, "Zombie Infestation", "Enchantment", jmpB3ZombieInfestationOracl, false)
	d1 := effectstest.SpmB12Hand(me, "One", "Instant")
	if err := g.ActivateCatalogAbility(me.ID, infestation, 0, game.ActivateAbilityParams{
		Strict: true, DiscardIDs: []uuid.UUID{d1},
	}); err == nil {
		t.Fatal("activated with one card in hand")
	}
	d2 := effectstest.SpmB12Hand(me, "Two", "Sorcery")
	if err := g.ActivateCatalogAbility(me.ID, infestation, 0, game.ActivateAbilityParams{
		Strict: true, DiscardIDs: []uuid.UUID{d1, d2},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(d1) || !me.Graveyard.Contains(d2) {
		t.Error("the two cards were not discarded")
	}
	zombie := effectstest.FindBattlefieldByName(g, "Zombie")
	if zombie == uuid.Nil {
		t.Fatal("no Zombie token")
	}
	c, _ := effectstest.BattlefieldCard(g, zombie)
	if c.Power != 2 || c.Toughness != 2 || !c.HasColor("B") || !c.IsToken() {
		t.Errorf("the token is %d/%d colours %v token %v; want a 2/2 black token", c.Power, c.Toughness, c.Colors, c.IsToken())
	}
}
