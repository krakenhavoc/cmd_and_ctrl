package effects

import (
	"testing"

	"github.com/google/uuid"

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

func jmpB3GainLife(g *game.Game, p uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, p, n) })
}

func TestJmpB3SilversmoteGhoulReturnsTappedAfterYouGainedThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ghoul := jmpB3GhoulInGraveyard(me)
	advanceToMain(t, g)
	jmpB3GainLife(g, me.ID, 1)
	jmpB3GainLife(g, me.ID, 2)
	// Losing it back does not undo the gain: the card asks what you gained.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -3) })
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(ghoul) {
		t.Fatal("the Ghoul stayed in the graveyard after you gained 3")
	}
	back := findBattlefieldByName(g, "Silversmote Ghoul")
	if back == uuid.Nil {
		t.Fatal("the Ghoul is not on the battlefield")
	}
	if tapped, _ := battlefieldCardTapped(g, back); !tapped {
		t.Error("the Ghoul returned untapped")
	}
}

func TestJmpB3SilversmoteGhoulStaysWithTwoLifeGainedOrOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghoul := jmpB3GhoulInGraveyard(me)
	advanceToMain(t, g)
	jmpB3GainLife(g, me.ID, 2)
	advanceTo(t, g, game.StepEnd)
	if n := triggersOnStackFrom(g, ghoul); n != 0 {
		t.Fatalf("2 life gained: %d triggers", n)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(ghoul) {
		t.Fatal("the Ghoul returned with only 2 life gained")
	}
	// The opponent's end step, after you gained 3 on their turn.
	advanceToMainOf(t, g, 1)
	jmpB3GainLife(g, me.ID, 3)
	jmpB3GainLife(g, opp.ID, 3)
	advanceTo(t, g, game.StepEnd)
	if n := triggersOnStackFrom(g, ghoul); n != 0 {
		t.Fatalf("an opponent's end step: %d triggers", n)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(ghoul) {
		t.Error("the Ghoul returned at an opponent's end step")
	}
}

func TestJmpB3SilversmoteGhoulDoesNothingIfItLeftTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ghoul := jmpB3GhoulInGraveyard(me)
	advanceToMain(t, g)
	jmpB3GainLife(g, me.ID, 3)
	advanceTo(t, g, game.StepEnd)
	if n := triggersOnStackFrom(g, ghoul); n != 1 {
		t.Fatalf("%d triggers at your end step, want 1", n)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneExile}, ghoul); err != nil {
		t.Fatalf("exile the Ghoul: %v", err)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Silversmote Ghoul") != uuid.Nil {
		t.Error("the Ghoul came back from exile")
	}
}

func TestJmpB3SilversmoteGhoulSacrificesToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ghoul := b21Push(g, me.ID, "Silversmote Ghoul", "Creature — Zombie Vampire", jmpB3SilversmoteGhoulOracle, 3, 1, "B")
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	b06AddMana(me, "C", "B")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
	if !me.Graveyard.Contains(ghoul) {
		t.Error("the Ghoul was not sacrificed")
	}
}

// --- Swarm of Bloodflies --------------------------------------------------

func TestJmpB3SwarmOfBloodfliesEntersWithTwoAndGrowsOnOtherDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	swarm := castCatalogSpell(t, g, "Swarm of Bloodflies", "Creature — Insect", jmpB3SwarmOfBloodfliesOracl, nil)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, swarm, "+1/+1"); n != 2 {
		t.Fatalf("entered with %d +1/+1 counters, want 2", n)
	}
	if p := effectivePowerWithCounters(t, g, swarm); p != 2 {
		t.Fatalf("entered with power %d, want 2", p)
	}
	if !hasEffectiveKeyword(t, g, swarm, "flying") {
		t.Error("no flying")
	}
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	b25Destroy(g, mine)
	passPriorityAroundTable(t, g)
	b25Destroy(g, theirs)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, swarm, "+1/+1"); n != 4 {
		t.Errorf("+1/+1 counters %d, want 4 (two on entry, one per death)", n)
	}
	// A noncreature leaving does nothing.
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	b25Destroy(g, relic)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, swarm, "+1/+1"); n != 4 {
		t.Errorf("an artifact died: counters %d, want 4", n)
	}
}

// --- Teferi's Tutelage ----------------------------------------------------

func TestJmpB3TeferisTutelageLootsOnEntryAndEachDrawMillsAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := me.Hand.Size()
	library := opp.Library.Size()
	tutelage := castCatalogSpell(t, g, "Teferi's Tutelage", "Enchantment", jmpB3TeferisTutelageOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || discardOwed(g, me.ID) != 1 {
		t.Fatalf("hand %d → %d, owes %d; want +1 and 1 discard owed",
			hand, me.Hand.Size(), discardOwed(g, me.ID))
	}
	discardFromHand(t, g, me.ID)
	// The loot's draw is a draw: it triggers the mill.
	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); hasID(p.PickTargetPlayers, me.ID) {
		t.Error("target OPPONENT offered you")
	}
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Library.Size() != library-2 {
		t.Fatalf("opponent library %d → %d, want -2", library, opp.Library.Size())
	}
	// Two draws, two triggers.
	b25Draw(g, me.ID, 2)
	for i := 0; i < 2; i++ {
		b04WaitForPick(t, g, me.ID)
		b16PickPlayer(t, g, me.ID, opp.ID)
	}
	passPriorityAroundTable(t, g)
	if opp.Library.Size() != library-6 {
		t.Errorf("after two draws the library is %d, want %d", opp.Library.Size(), library-6)
	}
	// An opponent's draw does nothing.
	b25Draw(g, opp.ID, 1)
	if n := triggersOnStackFrom(g, tutelage); n != 0 || latestPickTarget(g, me.ID) != nil {
		t.Errorf("an opponent's draw triggered the Tutelage")
	}
}

// --- Thought Collapse -----------------------------------------------------

func TestJmpB3ThoughtCollapseCountersAndTheControllerMillsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	library := opp.Library.Size()
	castCatalogSpell(t, g, "Thought Collapse", "Instant", jmpB3ThoughtCollapseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Fatal("the Bolt was not countered")
	}
	if opp.Library.Size() != library-3 {
		t.Errorf("the Bolt's controller milled %d, want 3", library-opp.Library.Size())
	}
}

func TestJmpB3ThoughtCollapseCannotTargetAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Thought Collapse", "Instant", jmpB3ThoughtCollapseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("targeted a creature on the battlefield")
	}
}

// --- Tolarian Kraken ------------------------------------------------------

func TestJmpB3TolarianKrakenPaysOneToTapOrUntapACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b21Push(g, me.ID, "Tolarian Kraken", "Creature — Kraken", jmpB3TolarianKrakenOracle, 4, 6, "U")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	b25Draw(g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("your draw asks you to pay {1}")
	}
	b06AddMana(me, "C")
	answerPayUnless(t, g, me.ID, true)
	monSettle(t, g)
	answerPickTarget(t, g, bear)
	monSettle(t, g)
	answerOptionPick(t, g, me.ID, 0) // tap it
	if tapped, _ := battlefieldCardTapped(g, bear); !tapped {
		t.Fatal("the Bear was not tapped")
	}
	// Again, and untap it this time.
	b25Draw(g, me.ID, 1)
	passPriorityAroundTable(t, g)
	b06AddMana(me, "C")
	answerPayUnless(t, g, me.ID, true)
	monSettle(t, g)
	answerPickTarget(t, g, bear)
	monSettle(t, g)
	answerOptionPick(t, g, me.ID, 1)
	if tapped, _ := battlefieldCardTapped(g, bear); tapped {
		t.Error("the Bear was not untapped")
	}
}

func TestJmpB3TolarianKrakenDeclinedOrOpponentDrawDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b21Push(g, me.ID, "Tolarian Kraken", "Creature — Kraken", jmpB3TolarianKrakenOracle, 4, 6, "U")
	b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	b25Draw(g, me.ID, 1)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	monSettle(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("declined: a target was still asked")
	}
	b25Draw(g, opp.ID, 1)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's draw asked you to pay")
	}
}

// --- Tome Anima -----------------------------------------------------------

func TestJmpB3TomeAnimaIsUnblockableOnceYouHaveDrawnTwo(t *testing.T) {
	g := newCatalogGame(t)
	attacker, defender := g.Seats[0], g.Seats[1]
	anima := b12Push(g, attacker.ID, "Tome Anima", "Creature — Spirit", jmpB3TomeAnimaOracle, 3, 3)
	blocker := b12Creature(g, defender.ID, "Wall", "Creature — Wall", 0, 4)
	brAttack(t, g, anima)
	drawn := g.TurnTallyFor(attacker.ID).CardsDrawn
	if drawn >= 2 {
		t.Fatalf("the harness drew %d before the test", drawn)
	}
	if err := g.Clone().DeclareBlocker(blocker, anima); err != nil {
		t.Fatalf("%d card(s) drawn: the block was refused: %v", drawn, err)
	}
	b25Draw(g, attacker.ID, 2-drawn)
	brRefusal(t, g.Clone().DeclareBlocker(blocker, anima), game.BlockReasonCantBeBlocked)
}

// --- Unsubstantiate -------------------------------------------------------

func TestJmpB3UnsubstantiateReturnsASpellToItsOwnersHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	life := me.Life
	castCatalogSpell(t, g, "Unsubstantiate", "Instant", jmpB3UnsubstantiateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bolt) {
		t.Fatal("the Bolt is not in its owner's hand")
	}
	if me.Life != life {
		t.Errorf("the Bolt resolved anyway: life %d → %d", life, me.Life)
	}
}

func TestJmpB3UnsubstantiateBouncesACreatureButNotANoncreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	if err := castCatalogSpellErr(t, g, "Unsubstantiate", "Instant", jmpB3UnsubstantiateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: relic}}); err == nil {
		t.Error("targeted a noncreature artifact")
	}
	castCatalogSpell(t, g, "Unsubstantiate", "Instant", jmpB3UnsubstantiateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bear) {
		t.Error("the Bear is not in its owner's hand")
	}
}

// --- Wall of Lost Thoughts ------------------------------------------------

func TestJmpB3WallOfLostThoughtsMillsAnyTargetPlayerFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	library := me.Library.Size()
	wall := castCatalogSpell(t, g, "Wall of Lost Thoughts", "Creature — Wall", jmpB3WallOfLostThoughtsOrcl, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); !hasID(p.PickTargetPlayers, me.ID) || !hasID(p.PickTargetPlayers, g.Seats[1].ID) {
		t.Errorf("target PLAYER: offered %v", p.PickTargetPlayers)
	}
	b16PickPlayer(t, g, me.ID, me.ID)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != library-4 {
		t.Errorf("library %d → %d, want -4", library, me.Library.Size())
	}
	if !hasEffectiveKeyword(t, g, wall, "defender") {
		t.Error("no defender")
	}
}

// --- Wight of Precinct Six ------------------------------------------------

func TestJmpB3WightOfPrecinctSixCountsOnlyOpponentsCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wight := b21Push(g, me.ID, "Wight of Precinct Six", "Creature — Zombie", jmpB3WightOfPrecinctSixOrcl, 1, 1, "B")
	if p, tough := effectivePTOf(t, g, wight); p != 1 || tough != 1 {
		t.Fatalf("empty graveyards: %d/%d, want 1/1", p, tough)
	}
	pushGraveyardCardWithTypeLine(g, me.ID, "Mine", "Creature — Bear")
	pushGraveyardCardWithTypeLine(g, g.Seats[1].ID, "Elf", "Creature — Elf")
	pushGraveyardCardWithTypeLine(g, g.Seats[2].ID, "Goblin", "Creature — Goblin")
	pushGraveyardCardWithTypeLine(g, g.Seats[2].ID, "Golem", "Artifact Creature — Golem")
	pushGraveyardCardWithTypeLine(g, g.Seats[3].ID, "Bolt", "Instant")
	if p, tough := effectivePTOf(t, g, wight); p != 4 || tough != 4 {
		t.Errorf("three opposing creature cards: %d/%d, want 4/4", p, tough)
	}
}

// --- Witch's Cauldron -----------------------------------------------------

func TestJmpB3WitchsCauldronSacrificesForALifeAndACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	cauldron := pushCatalogPermanent(g, me.ID, "Witch's Cauldron", "Artifact", jmpB3WitchsCauldronOracle, false)
	b06AddMana(me, "C", "B")
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no creature to sacrifice")
	}
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand, life := me.Hand.Size(), me.Life
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{bear},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+1 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d → %d, hand %d → %d; want +1 each", life, me.Life, hand, me.Hand.Size())
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if tapped, _ := battlefieldCardTapped(g, cauldron); !tapped {
		t.Error("the Cauldron was not tapped")
	}
	// Tapped: no second activation.
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	b06AddMana(me, "C", "B")
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{elf},
	}); err == nil {
		t.Error("activated a tapped Cauldron")
	}
}

// --- Zombie Infestation ---------------------------------------------------

func TestJmpB3ZombieInfestationDiscardsTwoForAZombie(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	for _, c := range me.Hand.Cards {
		cID := c.InstanceID
		_ = g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, cID)
	}
	infestation := pushCatalogPermanent(g, me.ID, "Zombie Infestation", "Enchantment", jmpB3ZombieInfestationOracl, false)
	d1 := spmB12Hand(me, "One", "Instant")
	if err := g.ActivateCatalogAbility(me.ID, infestation, 0, game.ActivateAbilityParams{
		Strict: true, DiscardIDs: []uuid.UUID{d1},
	}); err == nil {
		t.Fatal("activated with one card in hand")
	}
	d2 := spmB12Hand(me, "Two", "Sorcery")
	if err := g.ActivateCatalogAbility(me.ID, infestation, 0, game.ActivateAbilityParams{
		Strict: true, DiscardIDs: []uuid.UUID{d1, d2},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(d1) || !me.Graveyard.Contains(d2) {
		t.Error("the two cards were not discarded")
	}
	zombie := findBattlefieldByName(g, "Zombie")
	if zombie == uuid.Nil {
		t.Fatal("no Zombie token")
	}
	c, _ := battlefieldCard(g, zombie)
	if c.Power != 2 || c.Toughness != 2 || !c.HasColor("B") || !c.IsToken() {
		t.Errorf("the token is %d/%d colours %v token %v; want a 2/2 black token", c.Power, c.Toughness, c.Colors, c.IsToken())
	}
}
