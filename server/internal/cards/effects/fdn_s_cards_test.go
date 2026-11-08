package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdn_s_cards_test.go — Foundations, the cards from S that the S–Z batch
// (fdn_sz_cards_test.go) left (#4465). Each test plays the real catalog
// entry through the engine and asserts what its text says.

const (
	fdnSSearslicerOracle = "86ec0b56-497f-481c-a03a-c4e64e8e7404"
	fdnSFollyOracle      = "223794f2-67a3-4ab5-81c5-046c63307c6c"
	fdnSDowserOracle     = "dc3073fc-dc23-426e-b867-fee5dd99fe25"
	fdnSSireOracle       = "a15547d0-bbcb-41ee-a9ae-8325a9e7fe1f"
	fdnSArcherOracle     = "3792763c-7aa0-4055-b8a1-3257e0443ef1"
	fdnSBuccaneerOracle  = "5951f734-05d3-41ad-94ee-19bd310077d2"
	fdnSSlagstormOracle  = "2ed1879a-83b8-48df-a6d2-fe28e912b4c5"
	fdnSSowerOracle      = "2ec1860b-89cc-4529-8b8c-fb47f79f2b4b"
	fdnSLagacOracle      = "7215ca87-bbff-4c14-a1f5-de5ddd97c875"
	fdnSStarlightOracle  = "f0691d46-5bf4-4617-8a0a-19d17afb7219"
	fdnSStasisOracle     = "6a83768e-672e-4fec-8931-853f5e96d43d"
	fdnSSpyOracle        = "42081a57-b521-45c2-928d-0a0a4641818d"
	fdnSStrixOracle      = "5a6d01c8-47b6-43e5-a6e9-c48088286007"
)

func TestFdnSEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		fdnSSearslicerOracle: "Searslicer Goblin",
		fdnSFollyOracle:      "Seeker's Folly",
		fdnSDowserOracle:     "Shipwreck Dowser",
		fdnSSireOracle:       "Sire of Seven Deaths",
		fdnSArcherOracle:     "Skeleton Archer",
		fdnSBuccaneerOracle:  "Skyship Buccaneer",
		fdnSSlagstormOracle:  "Slagstorm",
		fdnSSowerOracle:      "Sower of Chaos",
		fdnSLagacOracle:      "Spitfire Lagac",
		fdnSStarlightOracle:  "Starlight Snare",
		fdnSStasisOracle:     "Stasis Snare",
		fdnSSpyOracle:        "Storm Fleet Spy",
		fdnSStrixOracle:      "Strix Lookout",
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

func TestFdnSPrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Sire of Seven Deaths", "Creature — Eldrazi", fdnSSireOracle, 7, 7, []string{"reach", "first strike", "vigilance", "menace", "trample", "lifelink"}},
		{"Skyship Buccaneer", "Creature — Human Pirate", fdnSBuccaneerOracle, 4, 3, []string{"flying"}},
		{"Strix Lookout", "Creature — Bird", fdnSStrixOracle, 1, 2, []string{"flying", "vigilance"}},
		{"Shipwreck Dowser", "Creature — Merfolk Wizard", fdnSDowserOracle, 3, 3, []string{game.KeywordProwess}},
	} {
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		assertKeywords(t, g, id, c.keywords...)
	}
	spec, _ := Lookup(fdnSStasisOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("Stasis Snare keywords %v, want flash", spec.PrintedKeywords)
	}
}

// fdnSCastCreature casts a catalog creature with its printed body from
// the active player's hand in the current main phase, without resolving
// it.
func fdnSCastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// fdnSCastMode casts a catalog sorcery choosing one mode.
func fdnSCastMode(t *testing.T, g *game.Game, name, oracle string, mode int, targets []game.TargetRef) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Sorcery", OracleID: oracle,
		Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{mode}, Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s mode %d: %v", name, mode, err)
	}
}

// --- Raid: Storm Fleet Spy, Skyship Buccaneer, Searslicer Goblin ----------------

func TestFdnSRaidDrawsOnlyAfterAnAttack(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
	}{
		{"Storm Fleet Spy", "Creature — Human Pirate", fdnSSpyOracle, 2, 2},
		{"Skyship Buccaneer", "Creature — Human Pirate", fdnSBuccaneerOracle, 4, 3},
	} {
		for _, attacked := range []bool{false, true} {
			name := c.name + "/no attack"
			if attacked {
				name = c.name + "/attacked"
			}
			t.Run(name, func(t *testing.T) {
				g := newCatalogGame(t)
				me, opp := g.Seats[0], g.Seats[1]
				seedLibrary(me, "First", "Second", "Third")
				bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
				advanceToMain(t, g)
				if attacked {
					declareAttack(t, g, opp.ID, bear)
					passPriorityAroundTable(t, g)
					mkAdvanceTo(t, g, game.StepPostcombatMain)
				}
				before := handSize(me)
				fdnSCastCreature(t, g, c.name, c.typeLine, c.oracle, c.power, c.toughness)
				passPriorityAroundTable(t, g)
				passPriorityAroundTable(t, g)
				want := before
				if attacked {
					want++
				}
				if got := handSize(me); got != want {
					t.Errorf("hand %d, want %d", got, want)
				}
			})
		}
	}
}

func TestFdnSSearslicerGoblinMakesAGoblinAtYourEndStepOnlyAfterAnAttack(t *testing.T) {
	for _, attacked := range []bool{false, true} {
		name := "no attack"
		if attacked {
			name = "attacked"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			b12Push(g, me.ID, "Searslicer Goblin", "Creature — Goblin Warrior", fdnSSearslicerOracle, 2, 1)
			bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
			advanceToMain(t, g)
			if attacked {
				declareAttack(t, g, opp.ID, bear)
				passPriorityAroundTable(t, g)
			}
			advanceToEndStep(t, g)
			passPriorityAroundTable(t, g)
			want := 0
			if attacked {
				want = 1
			}
			if got := countOnBattlefield(g, "Goblin", me.ID); got != want {
				t.Errorf("%d Goblin tokens, want %d", got, want)
			}
		})
	}
}

// --- Skeleton Archer ----------------------------------------------------------

func TestFdnSSkeletonArcherPingsAnyTargetOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	castHoldCreature(t, g, "Skeleton Archer", "Creature — Skeleton Archer", fdnSArcherOracle, 3, 3)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent at %d, want %d", opp.Life, life-1)
	}
}

func TestFdnSSkeletonArcherKillsAOneToughnessCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	castHoldCreature(t, g, "Skeleton Archer", "Creature — Skeleton Archer", fdnSArcherOracle, 3, 3)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, elf)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(elf) {
		t.Error("a 1/1 survived 1 damage")
	}
}

// --- Spitfire Lagac -----------------------------------------------------------

func TestFdnSSpitfireLagacPingsEachOpponentOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Spitfire Lagac", "Creature — Lizard", fdnSLagacOracle, 3, 4)
	before := b27Lives(g)
	playLandFromHand(t, g, "Mountain", "")
	passPriorityAroundTable(t, g)
	for i := 1; i < len(g.Seats); i++ {
		if got := g.Seats[i].Life; got != before[i]-1 {
			t.Errorf("opponent %d: %d -> %d, want -1", i, before[i], got)
		}
	}
	if me.Life != before[0] {
		t.Error("the controller is not an opponent")
	}
}

// --- Strix Lookout ------------------------------------------------------------

func TestFdnSStrixLookoutLootsForOneBlue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	strix := b12Push(g, me.ID, "Strix Lookout", "Creature — Bird", fdnSStrixOracle, 1, 2)
	seedLibrary(me, "Drawn")
	advanceToMain(t, g)
	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, strix, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without blue mana")
	}
	b06AddMana(me, "U")
	hand, grave := handSize(me), me.Graveyard.Size()
	b16Activate(t, g, me.ID, strix, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
	if !isTapped(g, strix) {
		t.Error("the Lookout is not tapped")
	}
	if handSize(me) != hand || me.Graveyard.Size() != grave+1 {
		t.Errorf("hand %d -> %d, graveyard %d -> %d; want the same hand and one more card in the graveyard",
			hand, handSize(me), grave, me.Graveyard.Size())
	}
}

// --- Sower of Chaos -----------------------------------------------------------

func TestFdnSSowerOfChaosStopsATargetCreatureBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sower := b12Push(g, me.ID, "Sower of Chaos", "Creature — Devil", fdnSSowerOracle, 4, 3)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 5)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "R")
	b16Activate(t, g, me.ID, sower, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(wall)})
	passPriorityAroundTable(t, g)
	if r := effectiveOf(t, g, wall).Restrictions; r&game.CantBlock == 0 {
		t.Errorf("restrictions %v, want CantBlock", r)
	}
	if isTapped(g, sower) {
		t.Error("the ability has no {T} in its cost")
	}
}

// --- Starlight Snare ----------------------------------------------------------

func TestFdnSStarlightSnareTapsAndHoldsItsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := pushPermanentForTest(g, me.ID, "Host", "", "Creature — Bear")
	aura := castCatalogSpell(t, g, "Starlight Snare", auraTypeLine, fdnSStarlightOracle, cardRefs(host))
	passPriorityAroundTable(t, g)
	if got := attachmentHostOf(t, g, aura); got.Kind != game.TargetCard || got.ID != host {
		t.Fatalf("Starlight Snare attached to %+v, want host %s", got, host)
	}
	if c, _ := battlefieldCard(g, host); !c.Tapped {
		t.Fatal("the enters trigger did not tap the enchanted creature")
	}
	advanceToUpkeepOf(t, g, 0)
	if c, _ := battlefieldCard(g, host); !c.Tapped {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

// --- Stasis Snare -------------------------------------------------------------

func TestFdnSStasisSnareExilesUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life

	snare := castHoldCreature(t, g, "Stasis Snare", "Enchantment", fdnSStasisOracle, 0, 0)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || onBattlefieldNamed(g, "Their Bear") != 0 {
		t.Fatal("the creature was not exiled")
	}
	if me.Life != life {
		t.Errorf("life %d -> %d: the Snare gains no life", life, me.Life)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(snare) })
	g.RunStateChecksForTest()
	if tmtB1Count(g, opp.ID, "Their Bear") != 1 {
		t.Error("the creature did not return when the Snare left")
	}
}

func TestFdnSStasisSnareOffersOnlyAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Stasis Snare", "Enchantment", fdnSStasisOracle, 0, 0)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
			t.Error("my own creature was accepted as the target")
		}
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own creature was exiled")
	}
}

// --- Slagstorm ----------------------------------------------------------------

func TestFdnSSlagstormDealsThreeToEachCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	big := b12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 4, 4)
	before := b27Lives(g)
	fdnSCastMode(t, g, "Slagstorm", fdnSSlagstormOracle, 0, nil)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(theirs) || g.Battlefield.Contains(mine) {
		t.Error("a creature with toughness 3 or less survived 3 damage")
	}
	if !g.Battlefield.Contains(big) || damageMarkedOn(g, big) != 3 {
		t.Error("the 4/4 was not dealt exactly 3 damage")
	}
	for i, l := range b27Lives(g) {
		if l != before[i] {
			t.Errorf("seat %d life %d, want an unchanged %d", i, l, before[i])
		}
	}
}

func TestFdnSSlagstormDealsThreeToEachPlayer(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	before := b27Lives(g)
	fdnSCastMode(t, g, "Slagstorm", fdnSSlagstormOracle, 1, nil)
	passPriorityAroundTable(t, g)
	for i, l := range b27Lives(g) {
		if l != before[i]-3 {
			t.Errorf("seat %d life %d, want %d", i, l, before[i]-3)
		}
	}
	if !g.Battlefield.Contains(bear) || damageMarkedOn(g, bear) != 0 {
		t.Error("the player mode dealt damage to a creature")
	}
}

// --- Seeker's Folly -----------------------------------------------------------

func TestFdnSSeekersFollyMakesTargetOpponentDiscardTwo(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	for _, n := range []string{"One", "Two", "Three"} {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	}
	hand, grave := handSize(opp), opp.Graveyard.Size()
	fdnSCastMode(t, g, "Seeker's Folly", fdnSFollyOracle, 0, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	c := discardChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("no discard prompt for the target opponent")
	}
	if c.ChooseMax != 2 {
		t.Errorf("discard prompt asks for %d cards, want 2", c.ChooseMax)
	}
	discardFromHand(t, g, opp.ID)
	if handSize(opp) != hand-2 || opp.Graveyard.Size() != grave+2 {
		t.Errorf("opponent hand %d -> %d, graveyard %d -> %d; want two cards moved", hand, handSize(opp), grave, opp.Graveyard.Size())
	}
}

func TestFdnSSeekersFollyShrinksOnlyYourOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	ox := b12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 3, 3)
	mine := b12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	fdnSCastMode(t, g, "Seeker's Folly", fdnSFollyOracle, 1, nil)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(elf) {
		t.Error("an opponent's 1/1 survived -1/-1")
	}
	if p, tough := effectivePower(t, g, ox), effectiveToughness(t, g, ox); p != 2 || tough != 2 {
		t.Errorf("the opponent's 3/3 is %d/%d, want 2/2", p, tough)
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own creature was shrunk")
	}
}

// --- Shipwreck Dowser ---------------------------------------------------------

func TestFdnSShipwreckDowserReturnsAnInstantOrSorceryFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := graveyardCardOf(g, me, "Bolt", "Instant", "{R}", "")
	bear := graveyardCardOf(g, me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	theirs := graveyardCardOf(g, opp, "Their Bolt", "Sorcery", "{R}", "")

	castAndResolveCreature(t, g, "Shipwreck Dowser", "Creature — Merfolk Wizard", fdnSDowserOracle)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target pick")
	}
	if !hasID(prompt.PickTargetCards, bolt) || hasID(prompt.PickTargetCards, bear) || hasID(prompt.PickTargetCards, theirs) {
		t.Errorf("legal set = %v, want only my instant or sorcery", prompt.PickTargetCards)
	}
	pickCard(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) || me.Graveyard.Contains(bolt) {
		t.Error("the chosen card did not return to hand")
	}
}

// --- Sire of Seven Deaths -----------------------------------------------------

func TestFdnSSireOfSevenDeathsWardAsksForSevenLife(t *testing.T) {
	g, _, opp, sire := wardTable(t, "Sire of Seven Deaths", "Creature — Eldrazi", fdnSSireOracle)
	castAtWardedCreature(t, g, opp, sire)
	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if prompt.LifeCost != 7 {
		t.Errorf("prompt LifeCost = %d, want 7", prompt.LifeCost)
	}
}
