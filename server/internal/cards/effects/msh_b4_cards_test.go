package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_b4_cards_test.go — Marvel Super Heroes, half B, fourth batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshB4GorillaOracle   = "6cb65b68-91b2-45ba-867b-3b31d6dd5643"
	mshB4RobotOracle     = "a2558476-5506-4a09-93fc-b8ab33c98985"
	mshB4TroopersOracle  = "6de67fad-0d01-443d-9ae0-4126ae903069"
	mshB4HellcatOracle   = "4c27ba61-6812-4dd7-8efc-65435a8df7b5"
	mshB4HeroTrainOracle = "87f1a840-021c-4699-a2c4-f2926daf98eb"
	mshB4HexMagicOracle  = "2283729c-11db-471c-9f63-46092c97bd00"
	mshB4HireCrewOracle  = "6eb7a870-b626-4326-9e4c-45193f8ce138"
	mshB4TorchOracle     = "5b6f2770-e774-4a77-801f-a3fda093c3dd"
	mshB4IronManOracle   = "deb16a9a-2fa9-4ad3-bb20-7c16066e3d14"
	mshB4IronFistOracle  = "bbf873c2-94c1-46b8-912e-e29e48e0212c"
)

func TestMshB4EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mshB4GorillaOracle:   "Guerrilla Gorilla",
		mshB4RobotOracle:     "HYDRA Assault Robot",
		mshB4TroopersOracle:  "HYDRA Troopers",
		mshB4HellcatOracle:   "Hellcat, Undying Vigilante",
		mshB4HeroTrainOracle: "Hero in Training",
		mshB4HexMagicOracle:  "Hex Magic",
		mshB4HireCrewOracle:  "Hire a Crew",
		mshB4TorchOracle:     "Human Torch, Johnny Storm",
		mshB4IronManOracle:   "I Am Iron Man",
		mshB4IronFistOracle:  "Iron Fist, Living Weapon",
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

// mshB4PickOpponent answers the target prompt a trigger opened for "target
// opponent", if it opened one.
func mshB4PickOpponent(t *testing.T, g *game.Game, chooser, opponent uuid.UUID) {
	t.Helper()
	if latestPickTarget(g, chooser) != nil {
		pickPlayer(t, g, chooser, opponent)
	}
}

// --- Guerrilla Gorilla --------------------------------------------------------

func TestMshB4GorillaSacrificesToDestroyANoncreatureArtifactOrEnchantment(t *testing.T) {
	for _, tc := range []struct{ name, typeLine string }{
		{"Test Rock", "Artifact"},
		{"Test Rune", "Enchantment"},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		powerUpMain(t, g)
		gorilla := pushCatalogPermanent(g, me.ID, "Guerrilla Gorilla", "Creature — Ape Soldier Hero", mshB4GorillaOracle, false)
		target := b12Permanent(g, opp.ID, tc.name, tc.typeLine)

		if err := hobB4Activate(t, g, me.ID, gorilla, 0, "", cardRefs(target)); err != nil {
			t.Fatalf("%s: activate: %v", tc.typeLine, err)
		}
		if g.Battlefield.Contains(gorilla) {
			t.Errorf("%s: the Gorilla was not sacrificed as a cost", tc.typeLine)
		}
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(target) {
			t.Errorf("%s: the %s survived", tc.typeLine, tc.name)
		}
		if got := graveyardNames(opp); len(got) != 1 || got[0] != tc.name {
			t.Errorf("%s: opponent's graveyard = %v", tc.typeLine, got)
		}
	}
}

func TestMshB4GorillaCannotTargetACreatureOrALandOrActAtInstantSpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	gorilla := pushCatalogPermanent(g, me.ID, "Guerrilla Gorilla", "Creature — Ape Soldier Hero", mshB4GorillaOracle, false)
	artifactCreature := b12Creature(g, opp.ID, "Test Golem", "Artifact Creature — Golem", 2, 2)
	enchantmentCreature := b12Creature(g, opp.ID, "Test Spirit", "Enchantment Creature — Spirit", 2, 2)
	plainCreature := b12Creature(g, opp.ID, "Test Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Test Rock", "Artifact")

	for _, bad := range []uuid.UUID{artifactCreature, enchantmentCreature, plainCreature} {
		if err := hobB4Activate(t, g, me.ID, gorilla, 0, "", cardRefs(bad)); err == nil {
			t.Fatal("the Gorilla targeted a creature")
		}
		if !g.Battlefield.Contains(gorilla) {
			t.Fatal("a refused activation still sacrificed the Gorilla")
		}
	}

	advanceTo(t, g, game.StepBeginCombat)
	if err := hobB4Activate(t, g, me.ID, gorilla, 0, "", cardRefs(rock)); err == nil {
		t.Error("the Gorilla activated outside a main phase")
	}
	if !g.Battlefield.Contains(gorilla) || !g.Battlefield.Contains(rock) {
		t.Error("the refused activation changed the board")
	}
}

// --- HYDRA Assault Robot ------------------------------------------------------

func TestMshB4RobotPingsForAnotherVillainOrArtifactYouControl(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		pings          bool
	}{
		{"Test Villain", "Creature — Human Villain", true},
		{"Test Trinket", "Artifact", true},
		{"Test Both", "Artifact Creature — Robot Villain", true},
		{"Test Hero", "Creature — Human Hero", false},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		powerUpMain(t, g)
		pushCatalogPermanent(g, me.ID, "HYDRA Assault Robot", "Artifact Creature — Robot Villain", mshB4RobotOracle, false)

		castPowerUpCreature(t, g, me, tc.name, tc.typeLine, "{1}", "{C}", "", 2)
		mshB4PickOpponent(t, g, me.ID, opp.ID)
		passPriorityAroundTable(t, g)

		want := 40
		if tc.pings {
			want = 39
		}
		if opp.Life != want {
			t.Errorf("%s: opponent life %d, want %d", tc.name, opp.Life, want)
		}
	}
}

func TestMshB4RobotTriggersOnceForAnArtifactVillainAndNotForItselfOrOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	pushCatalogPermanent(g, me.ID, "HYDRA Assault Robot", "Artifact Creature — Robot Villain", mshB4RobotOracle, false)

	castPowerUpCreature(t, g, me, "Both", "Artifact Creature — Robot Villain", "{1}", "{C}", "", 2)
	mshB4PickOpponent(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != 39 {
		t.Errorf("an artifact Villain pinged for %d, want exactly 1", 40-opp.Life)
	}

	// Another Robot of the same name entering is "another", so it counts;
	// an opponent's Villain does not.
	b12Creature(g, opp.ID, "Their Villain", "Creature — Human Villain", 2, 2)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || opp.Life != 39 {
		t.Errorf("an opponent's Villain triggered the Robot (life %d)", opp.Life)
	}
}

// --- HYDRA Troopers -----------------------------------------------------------

func mshB4CastTroopers(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	castPowerUpCreature(t, g, me, "HYDRA Troopers", "Creature — Human Soldier Villain", "{2}{B}", "{C}{C}{B}", mshB4TroopersOracle, 3)
	passPriorityAroundTable(t, g)
}

func TestMshB4TroopersMakeATappedVillainWithTwoCreatureCardsInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	seedLibrary(me, "Lib A", "Lib B", "Lib C")
	for _, name := range []string{"Dead One", "Dead Two"} {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	library := me.Library.Size()

	mshB4CastTroopers(t, g, me)

	if got := mshB1Tokens(g, me.ID, "Villain"); got != 1 {
		t.Fatalf("Villain tokens = %d, want 1", got)
	}
	token := mshB1Token(t, g, me.ID, "Villain")
	if !token.Tapped || !game.HasKeyword(&token, "menace") {
		t.Errorf("token tapped=%v menace=%v, want a tapped token with menace", token.Tapped, game.HasKeyword(&token, "menace"))
	}
	if p, tough := effectivePTOf(t, g, token.InstanceID); p != 2 || tough != 1 {
		t.Errorf("token is %d/%d, want 2/1", p, tough)
	}
	if me.Library.Size() != library {
		t.Errorf("the library shrank from %d to %d: the controller milled with the token branch", library, me.Library.Size())
	}
}

func TestMshB4TroopersMillTwoWithFewerThanTwoCreatureCardsInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	seedLibrary(me, "Lib A", "Lib B", "Lib C")
	// One creature card and one noncreature card: still not two creatures.
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead One", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	library := me.Library.Size()

	mshB4CastTroopers(t, g, me)

	if got := mshB1Tokens(g, me.ID, "Villain"); got != 0 {
		t.Errorf("Villain tokens = %d, want none", got)
	}
	if me.Library.Size() != library-2 {
		t.Errorf("library %d -> %d, want two cards milled", library, me.Library.Size())
	}
	if me.Graveyard.Size() != 4 {
		t.Errorf("graveyard has %d cards, want 4", me.Graveyard.Size())
	}
}

// --- Hellcat, Undying Vigilante ----------------------------------------------

func TestMshB4HellcatReturnsOnceWithACounterAndOnlyHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	hellcat := b12Push(g, me.ID, "Hellcat, Undying Vigilante", "Legendary Creature — Human Hero", mshB4HellcatOracle, 2, 2)

	killCreature(t, g, me.ID, hellcat)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(hellcat) {
		t.Fatalf("Hellcat did not return; zone %v", b12ZoneOf(g, hellcat))
	}
	if got := b12Counter(t, g, hellcat, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if !hasKeywordOnBattlefield(t, g, hellcat, "haste") {
		t.Error("the returned Hellcat has no haste")
	}

	// She lost her dies trigger, so a second death is final.
	killCreature(t, g, me.ID, hellcat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hellcat) || b12ZoneOf(g, hellcat) != game.ZoneGraveyard {
		t.Errorf("Hellcat came back a second time (zone %v)", b12ZoneOf(g, hellcat))
	}
}

func TestMshB4HellcatDoesNotReturnFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	hellcat := b12Push(g, me.ID, "Hellcat, Undying Vigilante", "Legendary Creature — Human Hero", mshB4HellcatOracle, 2, 2)

	exileCreature(t, g, hellcat)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(hellcat) || b12ZoneOf(g, hellcat) != game.ZoneExile {
		t.Errorf("an exiled Hellcat is in zone %v, want exile", b12ZoneOf(g, hellcat))
	}
}

// --- Hero in Training ---------------------------------------------------------

func TestMshB4HeroInTrainingDrawsAndGainsLifeOnlyWithAnotherHero(t *testing.T) {
	for _, withHero := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		powerUpMain(t, g)
		seedLibrary(me, "Lib A", "Lib B")
		if withHero {
			b12Creature(g, me.ID, "Test Hero", "Creature — Human Hero", 1, 1)
		}
		hand := me.Hand.Size()

		castPowerUpCreature(t, g, me, "Hero in Training", "Creature — Human Hero", "{2}{W}", "{C}{C}{W}", mshB4HeroTrainOracle, 2)
		passPriorityAroundTable(t, g)

		if got := me.Hand.Size(); got != hand+1 {
			t.Errorf("withHero=%v: hand %d -> %d, want one card drawn", withHero, hand, got)
		}
		want := 40
		if withHero {
			want = 42
		}
		if me.Life != want {
			t.Errorf("withHero=%v: life %d, want %d (Hero in Training is not 'another')", withHero, me.Life, want)
		}
	}
}

func TestMshB4HeroInTrainingIgnoresAnOpponentsHero(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	seedLibrary(me, "Lib A")
	b12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 1, 1)

	castPowerUpCreature(t, g, me, "Hero in Training", "Creature — Human Hero", "{2}{W}", "{C}{C}{W}", mshB4HeroTrainOracle, 2)
	passPriorityAroundTable(t, g)

	if me.Life != 40 {
		t.Errorf("life %d, want 40: an opponent's Hero is not yours", me.Life)
	}
}

// --- Hex Magic ----------------------------------------------------------------

func TestMshB4HexMagicExilesTheHandDrawsThatManyAndLetsYouPlayThemUntilYourNextTurnEnds(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	var exiled []uuid.UUID
	exiled = append(exiled, handCardFull(me, "Old Spell", "Sorcery", "{1}", "", nil))
	exiled = append(exiled, handCardFull(me, "Old Land", "Land", "", "", nil))
	exiled = append(exiled, handCardFull(me, "Old Bear", "Creature — Bear", "{1}", "", nil))
	seedLibrary(me, "New A", "New B", "New C", "New D")

	castCatalogSpell(t, g, "Hex Magic", "Sorcery — Arcane", mshB4HexMagicOracle, nil)
	passPriorityAroundTable(t, g)

	for _, id := range exiled {
		if b12ZoneOf(g, id) != game.ZoneExile {
			t.Errorf("%s is in %v, want exile", b12Card(t, g, id).Name, b12ZoneOf(g, id))
		}
		if !b12ExiledPlayableBy(t, g, id, me.ID) {
			t.Errorf("%s is not playable by its owner", b12Card(t, g, id).Name)
		}
	}
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand has %d cards, want the 3 drawn", got)
	}
	for _, c := range me.Hand.Cards {
		if c.Name == "Old Spell" || c.Name == "Old Land" || c.Name == "Old Bear" {
			t.Errorf("%s was drawn: it was exiled", c.Name)
		}
	}

	// Playable through the opponents' turns and my next turn, not after.
	opponentsTurn := func(seat int) {
		powerUpPassTo(t, g, seat)
		for _, id := range exiled {
			if !b12ExiledPlayableBy(t, g, id, me.ID) {
				t.Errorf("seat %d's turn: %s lost its permission early", seat, b12Card(t, g, id).Name)
			}
		}
	}
	opponentsTurn(1)
	opponentsTurn(2)
	opponentsTurn(3)
	opponentsTurn(0)
	powerUpPassTo(t, g, 1)
	for _, id := range exiled {
		if b12ExiledPlayableBy(t, g, id, me.ID) {
			t.Errorf("%s is still playable after the end of my next turn", b12Card(t, g, id).Name)
		}
	}
}

func TestMshB4HexMagicWithAnEmptyHandDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	seedLibrary(me, "New A", "New B")
	library := me.Library.Size()

	castCatalogSpell(t, g, "Hex Magic", "Sorcery — Arcane", mshB4HexMagicOracle, nil)
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != 0 || me.Library.Size() != library {
		t.Errorf("hand %d, library %d -> %d: an empty hand draws nothing", me.Hand.Size(), library, me.Library.Size())
	}
}

// --- Hire a Crew --------------------------------------------------------------

func TestMshB4HireACrewMakesAVillainThenPumpsEveryCreatureYouControlIncludingIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Hire a Crew", "Instant", mshB4HireCrewOracle, nil)
	passPriorityAroundTable(t, g)

	token := mshB1Token(t, g, me.ID, "Villain")
	if !game.HasKeyword(&token, "menace") || token.Tapped {
		t.Errorf("token menace=%v tapped=%v, want an untapped token with menace", game.HasKeyword(&token, "menace"), token.Tapped)
	}
	if p, tough := effectivePTOf(t, g, token.InstanceID); p != 3 || tough != 1 {
		t.Errorf("token is %d/%d, want 3/1 (2/1 pumped by its own spell)", p, tough)
	}
	if p, tough := effectivePTOf(t, g, mine); p != 3 || tough != 2 {
		t.Errorf("my creature is %d/%d, want 3/2", p, tough)
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their creature is %d/%d, want 2/2", p, tough)
	}

	// A creature that enters afterwards is not pumped.
	late := b12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
	if p, _ := effectivePTOf(t, g, late); p != 2 {
		t.Errorf("a creature that entered later has power %d, want 2", p)
	}
}

// --- Human Torch, Johnny Storm ------------------------------------------------

func mshB4PushTorch(g *game.Game, me *game.Player) uuid.UUID {
	return pushCatalogPermanent(g, me.ID, "Human Torch, Johnny Storm", "Legendary Creature — Human Hero", mshB4TorchOracle, false)
}

func TestMshB4TorchPingsAnOpponentForEachCardYouDrawWhileYouControlAnotherHero(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	seedLibrary(me, "Lib A", "Lib B", "Lib C")
	mshB4PushTorch(g, me)
	b12Creature(g, me.ID, "Test Hero", "Creature — Human Hero", 1, 1)

	for i := 1; i <= 2; i++ {
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatalf("DrawCard: %v", err)
		}
		mshB4PickOpponent(t, g, me.ID, opp.ID)
		passPriorityAroundTable(t, g)
		if opp.Life != 40-i {
			t.Errorf("after %d draws the opponent is at %d, want %d", i, opp.Life, 40-i)
		}
	}
}

func TestMshB4TorchNeedsAnotherHeroAndIgnoresAnOpponentsDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	seedLibrary(me, "Lib A", "Lib B")
	seedLibrary(opp, "Opp A")
	mshB4PushTorch(g, me)

	// Alone: Human Torch is not "another" Hero.
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || opp.Life != 40 {
		t.Fatalf("Human Torch pinged with no other Hero (life %d)", opp.Life)
	}

	// An opponent's Hero does not count, nor does an opponent drawing.
	b12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 1, 1)
	b12Creature(g, me.ID, "Test Hero", "Creature — Human Hero", 1, 1)
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || opp.Life != 40 {
		t.Errorf("an opponent's draw pinged (life %d)", opp.Life)
	}
}

func TestMshB4TorchDoesNothingIfTheOtherHeroLeavesInResponse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	seedLibrary(me, "Lib A")
	mshB4PushTorch(g, me)
	hero := b12Creature(g, me.ID, "Test Hero", "Creature — Human Hero", 1, 1)

	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	mshB4PickOpponent(t, g, me.ID, opp.ID)
	killCreature(t, g, me.ID, hero)
	passPriorityAroundTable(t, g)

	if opp.Life != 40 {
		t.Errorf("opponent life %d, want 40: the intervening 'if' is checked again on resolution", opp.Life)
	}
}

func TestMshB4TorchPowerUpCostsLessTheTurnItEntersAndPutsThreeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	torch := castPowerUpCreature(t, g, me, "Human Torch, Johnny Storm", "Legendary Creature — Human Hero", "{2}{R}", "{C}{C}{R}", mshB4TorchOracle, 2)

	if got := chargedPowerUp(t, g, me.ID, torch); got != "{4}" {
		t.Fatalf("the turn it entered the power-up costs %q, want {4}", got)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, torch, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("power-up: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, torch, game.CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", got)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, torch, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the power-up was activated a second time")
	}
}

// --- I Am Iron Man ------------------------------------------------------------

func TestMshB4IronManMakesAnArtifactOrCreatureAFourByFourFlyingArtifactCreatureAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	seedLibrary(me, "Lib A", "Lib B")
	rock := b12Permanent(g, opp.ID, "Test Rock", "Artifact — Treasure")
	hand := me.Hand.Size()

	castCatalogSpell(t, g, "I Am Iron Man", "Instant", mshB4IronManOracle, cardRefs(rock))
	passPriorityAroundTable(t, g)

	c := gaLayered(t, g, rock)
	if !c.IsArtifact() || !c.IsCreature() {
		t.Errorf("the Treasure is %q, want an artifact creature", c.TypeLine)
	}
	if !c.HasSubtype("Treasure") {
		t.Error("the Treasure lost its subtype: it is still an artifact, so its artifact type stays (CR 205.1a)")
	}
	if p, tough := effectivePTOf(t, g, rock); p != 4 || tough != 4 {
		t.Errorf("it is %d/%d, want 4/4", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, rock, "flying") {
		t.Error("it has no flying")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, got)
	}

	// Until end of turn only.
	powerUpPassTo(t, g, 1)
	if gaLayered(t, g, rock).IsCreature() {
		t.Error("the Treasure is still a creature on the next turn")
	}
}

func TestMshB4IronManSetsTheBaseSizeOfACreatureAndAddsTheArtifactType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	seedLibrary(me, "Lib A")
	goblin := b12Creature(g, me.ID, "Test Goblin", "Creature — Goblin", 1, 1)

	castCatalogSpell(t, g, "I Am Iron Man", "Instant", mshB4IronManOracle, cardRefs(goblin))
	passPriorityAroundTable(t, g)

	if p, tough := effectivePTOf(t, g, goblin); p != 4 || tough != 4 {
		t.Errorf("a 1/1 is %d/%d, want 4/4", p, tough)
	}
	c := gaLayered(t, g, goblin)
	if !c.IsArtifact() || !c.IsCreature() || !c.HasSubtype("Goblin") {
		t.Errorf("the Goblin is %q, want an artifact creature that is still a Goblin", c.TypeLine)
	}
	if !hasKeywordOnBattlefield(t, g, goblin, "flying") {
		t.Error("the Goblin has no flying")
	}
}

// --- Iron Fist, Living Weapon -------------------------------------------------

func TestMshB4IronFistGainsATapAbilityForASpellTargetingYourCreatureAndDealsHisPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	fist := b12Push(g, me.ID, "Iron Fist, Living Weapon", "Legendary Creature — Human Warrior Hero", mshB4IronFistOracle, 4, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if idx, _ := grantedActivatedIndex(t, g, fist); idx >= 0 {
		t.Fatal("Iron Fist has the ability before any spell")
	}

	castCatalogSpell(t, g, "Djeru's Resolve", "Instant", pr7aDjerusResolve, []game.TargetRef{pr7aCard(bear)})
	passPriorityAroundTable(t, g)

	idx, ref := grantedActivatedIndex(t, g, fist)
	if idx < 0 {
		t.Fatal("Iron Fist did not gain the tap ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, fist, idx, game.ActivateAbilityParams{Ref: ref, Targets: cardRefs(fist)}); err == nil {
		t.Error("Iron Fist targeted himself: it is 'any other target'")
	}
	b16Activate(t, g, me.ID, fist, idx, game.ActivateAbilityParams{Ref: ref, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	passPriorityAroundTable(t, g)
	if opp.Life != 36 {
		t.Errorf("opponent life %d, want 36 (his power is 4)", opp.Life)
	}
	if !b16Tapped(t, g, fist) {
		t.Error("the ability's {T} cost did not tap Iron Fist")
	}

	// It is gone at the end of the turn.
	advancePastCleanupForTest(t, g)
	if idx, _ := grantedActivatedIndex(t, g, fist); idx >= 0 {
		t.Error("Iron Fist kept the ability into the next turn")
	}
}

func TestMshB4IronFistDamagesACreatureForItsCurrentPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	fist := b12Push(g, me.ID, "Iron Fist, Living Weapon", "Legendary Creature — Human Warrior Hero", mshB4IronFistOracle, 2, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Djeru's Resolve", "Instant", pr7aDjerusResolve, []game.TargetRef{pr7aCard(bear)})
	passPriorityAroundTable(t, g)
	idx, ref := grantedActivatedIndex(t, g, fist)
	if idx < 0 {
		t.Fatal("Iron Fist did not gain the tap ability")
	}
	b16Activate(t, g, me.ID, fist, idx, game.ActivateAbilityParams{Ref: ref, Targets: cardRefs(victim)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("a 2-power Iron Fist did not kill a 2/2")
	}
}

func TestMshB4IronFistIgnoresASpellThatTargetsOnlyAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	fist := b12Push(g, me.ID, "Iron Fist, Living Weapon", "Legendary Creature — Human Warrior Hero", mshB4IronFistOracle, 2, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Djeru's Resolve", "Instant", pr7aDjerusResolve, []game.TargetRef{pr7aCard(theirs)})
	passPriorityAroundTable(t, g)

	if idx, _ := grantedActivatedIndex(t, g, fist); idx >= 0 {
		t.Error("a spell on an opponent's creature gave Iron Fist the ability")
	}
}

func TestMshB4IronManCannotTargetANonartifactNoncreaturePermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	rune := b12Permanent(g, me.ID, "Test Rune", "Enchantment")

	for _, bad := range []uuid.UUID{forest, rune} {
		if err := castCatalogSpellErr(t, g, "I Am Iron Man", "Instant", mshB4IronManOracle, cardRefs(bad)); err == nil {
			t.Error("I Am Iron Man targeted a land or an enchantment")
		}
	}
}
