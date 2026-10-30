package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_b5_cards_test.go — Marvel Super Heroes, half B, fifth batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshB5IronManArmorOracle = "958d81c2-ca09-4275-890e-3bb595317325"
	mshB5JusticeOracle      = "184dba74-63ed-4820-9174-68adf26bfabf"
	mshB5KaZarOracle        = "88eeb519-5782-4732-bd34-81d1ee6dbc8f"
	mshB5KillmongerOracle   = "1c3ba54a-a8ef-4838-b3aa-cb6df827a619"
	mshB5EnforcersOracle    = "342df67f-44d8-4b0a-b1d7-4241966d0598"
	mshB5KnightOracle       = "39145960-a212-46b4-addb-4b187f6bf6fd"
	mshB5SentinelOracle     = "3c2b9bb2-8201-47e8-84d5-f3518cac03cb"
	mshB5LokiOracle         = "9c78778a-6335-4949-8ade-11d0f085cb2b"
	mshB5LukeOracle         = "c95d1558-09a4-4dbc-b08a-fe21b47bc5d2"
	mshB5MachinesmithOracle = "945e8443-9647-4eab-b091-ee8e7bc347f5"
	mshB5MasqueOracle       = "b5b0e85f-6651-4667-829c-2253ff8b130d"
)

func TestMshB5EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mshB5IronManArmorOracle:      "Iron Man Armor",
		jenniferWaltersOracle:        "Jennifer Walters",
		jenniferWaltersOracle + "#1": "The Sensational She-Hulk",
		mshB5JusticeOracle:           "Justice, Vance Astrovik",
		mshB5KaZarOracle:             "Ka-Zar of the Savage Land",
		mshB5KillmongerOracle:        "Killmonger, Scourge of Wakanda",
		mshB5EnforcersOracle:         "Kingpin's Enforcers",
		mshB5KnightOracle:            "Knight of Wundagore",
		mshB5SentinelOracle:          "Kree Sentinel",
		mshB5LokiOracle:              "Loki Laufeyson",
		mshB5LukeOracle:              "Luke Cage, Power Man",
		mshB5MachinesmithOracle:      "Machinesmith Automaton",
		mshB5MasqueOracle:            "Madame Masque",
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

// mshB5PickPlayerIfAsked answers an open target pick with a player.
func mshB5PickPlayerIfAsked(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	if p := latestPickTarget(g, chooser); p != nil {
		if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetPlayer, ID: target}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
}

// --- Luke Cage, Power Man ---------------------------------------------------

func TestMshB5LukeCageGainsPowerAndIndestructibleAttackingAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	luke := b12Push(g, me.ID, "Luke Cage, Power Man", "Legendary Creature — Human Hero", mshB5LukeOracle, 2, 5)

	declareAttack(t, g, opp.ID, luke)
	passPriorityAroundTable(t, g)

	if p, _ := effectivePTOf(t, g, luke); p != 4 {
		t.Errorf("Luke Cage's power = %d, want 4 (+2/+0)", p)
	}
	if !hasKeywordOnBattlefield(t, g, luke, "indestructible") {
		t.Error("Luke Cage is not indestructible")
	}
}

func TestMshB5LukeCageDoesNothingWhenAnotherCreatureAttacksToo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	luke := b12Push(g, me.ID, "Luke Cage, Power Man", "Legendary Creature — Human Hero", mshB5LukeOracle, 2, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, luke, bear)
	passPriorityAroundTable(t, g)

	if p, _ := effectivePTOf(t, g, luke); p != 2 {
		t.Errorf("Luke Cage's power = %d, want 2 (not alone)", p)
	}
	if hasKeywordOnBattlefield(t, g, luke, "indestructible") {
		t.Error("Luke Cage is indestructible though he did not attack alone")
	}
}

func TestMshB5LukeCageIgnoresAnotherCreatureAttackingAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	luke := b12Push(g, me.ID, "Luke Cage, Power Man", "Legendary Creature — Human Hero", mshB5LukeOracle, 2, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)

	if p, _ := effectivePTOf(t, g, luke); p != 2 {
		t.Errorf("Luke Cage's power = %d, want 2: he did not attack", p)
	}
}

// --- Machinesmith Automaton -------------------------------------------------

func TestMshB5MachinesmithGrowsForAnotherArtifactOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	smith := castPowerUpCreature(t, g, me, "Machinesmith Automaton", "Artifact Creature — Robot Villain", "{2}{R}", "{C}{C}{R}", mshB5MachinesmithOracle, 2)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, smith); got != 0 {
		t.Fatalf("counters after it entered = %d, want 0: it is not another artifact", got)
	}
	if !hasKeywordOnBattlefield(t, g, smith, "trample") {
		t.Error("Machinesmith Automaton lacks trample")
	}

	castPowerUpCreature(t, g, me, "Test Rock", "Artifact", "{1}", "{C}", "", 0)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, smith); got != 1 {
		t.Errorf("counters after another artifact entered = %d, want 1", got)
	}

	castPowerUpCreature(t, g, me, "Test Bear", "Creature — Bear", "{1}", "{C}", "", 2)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, smith); got != 1 {
		t.Errorf("counters after a nonartifact entered = %d, want it to stay 1", got)
	}
}

// --- Kree Sentinel ----------------------------------------------------------

func TestMshB5KreeSentinelHasReachAndBasicLandcycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me, searchTestLand("Forest", "Basic Land — Forest"), game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	forest := ids[0]

	sentinel, _ := cycleFromHand(t, g, "Kree Sentinel", "Artifact Creature — Kree Robot Villain", mshB5SentinelOracle, "{C}{C}")
	if !me.Graveyard.Contains(sentinel) {
		t.Fatal("the landcycled Sentinel is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("basic landcycling did not ask for a basic land")
	}
	if len(c.SearchCards) != 1 || c.SearchCards[0] != forest {
		t.Errorf("search offers %v, want only the basic Forest", c.SearchCards)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(forest) {
		t.Error("the basic land did not reach the hand")
	}

	g2 := newCatalogGame(t)
	onField := pushCatalogPermanent(g2, g2.Seats[0].ID, "Kree Sentinel", "Artifact Creature — Kree Robot Villain", mshB5SentinelOracle, false)
	if !hasKeywordOnBattlefield(t, g2, onField, "reach") {
		t.Error("Kree Sentinel lacks reach")
	}
}

// --- Kingpin's Enforcers ----------------------------------------------------

func TestMshB5EnforcersSacrificeAnArtifactOrCreatureToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	enforcers := b12Push(g, me.ID, "Kingpin's Enforcers", "Creature — Human Villain", mshB5EnforcersOracle, 2, 3)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	land := seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	advanceToMain(t, g)
	if !hasKeywordOnBattlefield(t, g, enforcers, "lifelink") {
		t.Error("Kingpin's Enforcers lack lifelink")
	}

	floatMana(t, g, me, "{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, enforcers, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("a land is neither an artifact nor a creature")
	}
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, enforcers, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(rock) {
		t.Fatal("the artifact is sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}

	floatMana(t, g, me, "{C}{C}{B}")
	b16Activate(t, g, me.ID, enforcers, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{enforcers}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(enforcers) || me.Hand.Size() != hand+2 {
		t.Error("sacrificing the Enforcers to their own ability still draws")
	}
}

// --- Knight of Wundagore ----------------------------------------------------

func TestMshB5KnightGrowsOnceEachTurnWhenYouPutACounterOnAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	knight := b12Push(g, me.ID, "Knight of Wundagore", "Creature — Cat Knight Villain", mshB5KnightOracle, 2, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !hasKeywordOnBattlefield(t, g, knight, "trample") {
		t.Error("Knight of Wundagore lacks trample")
	}

	// A counter on the Knight itself, or one an opponent puts, does nothing.
	if err := g.AddCounterBy(me.ID, knight, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	if err := g.AddCounterBy(opp.ID, bear, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, knight); got != 1 {
		t.Fatalf("Knight counters = %d, want 1 (only the one put on itself)", got)
	}

	// You put one on another creature, even an opponent's: it grows.
	if err := g.AddCounterBy(me.ID, theirs, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, knight); got != 2 {
		t.Fatalf("Knight counters = %d, want 2", got)
	}

	// Only once each turn.
	if err := g.AddCounterBy(me.ID, bear, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, knight); got != 2 {
		t.Errorf("Knight counters = %d, want it to stay 2 this turn", got)
	}
}

// --- Madame Masque ----------------------------------------------------------

func TestMshB5MasqueConnivesWhenSheEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	spell := handCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)
	seedLibrary(me, "Drawn")

	masque := castPowerUpCreature(t, g, me, "Madame Masque", "Legendary Creature — Human Villain", "{4}{B}", "{C}{C}{C}{C}{B}", mshB5MasqueOracle, 3)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("Madame Masque did not connive: no discard prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{spell}); err != nil {
		t.Fatalf("discard: %v", err)
	}
	if got := plusOnes(g, masque); got != 1 {
		t.Errorf("Masque has %d +1/+1 counters, want 1 (she discarded a nonland card)", got)
	}
}

func TestMshB5MasqueMakesAVillainOnTheSecondDrawOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Madame Masque", "Legendary Creature — Human Villain", mshB5MasqueOracle, false)

	hobA5Draw(t, g, other, nil)
	hobA5Draw(t, g, other, nil)
	hobA5Draw(t, g, me, nil)
	if n := mshB1Tokens(g, me.ID, "Villain"); n != 0 {
		t.Fatalf("%d Villain tokens before the second draw, want 0", n)
	}
	hobA5Draw(t, g, me, nil)
	if n := mshB1Tokens(g, me.ID, "Villain"); n != 1 {
		t.Fatalf("%d Villain tokens after the second draw, want 1", n)
	}
	tok := mshB1Token(t, g, me.ID, "Villain")
	if p, tough := effectivePTOf(t, g, tok.InstanceID); p != 2 || tough != 1 || !hasKeywordOnBattlefield(t, g, tok.InstanceID, "menace") {
		t.Errorf("token is %d/%d, want a 2/1 with menace", p, tough)
	}
	hobA5Draw(t, g, me, nil)
	if n := mshB1Tokens(g, me.ID, "Villain"); n != 1 {
		t.Errorf("%d Villain tokens after the third draw, want it to stay 1", n)
	}
}

// --- Justice, Vance Astrovik ------------------------------------------------

func TestMshB5JusticeBouncesATargetWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	justice := castPowerUpCreature(t, g, me, "Justice, Vance Astrovik", "Legendary Creature — Mutant Hero", "{2}{U}", "{C}{C}{U}", mshB5JusticeOracle, 2)
	mshB2PickIfAsked(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)

	if !opp.Hand.Contains(theirs) {
		t.Error("the target was not returned to its owner's hand")
	}
	if got := plusOnes(g, justice); got != 0 {
		t.Errorf("Justice has %d counters, want 0: the bounced permanent was not his", got)
	}
}

func TestMshB5JusticeGrowsWhenAnotherNonlandPermanentOfYoursIsReturned(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	justice := b12Push(g, me.ID, "Justice, Vance Astrovik", "Legendary Creature — Mutant Hero", mshB5JusticeOracle, 2, 2)
	if !hasKeywordOnBattlefield(t, g, justice, "flying") {
		t.Error("Justice lacks flying")
	}
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	bounce := func(owner, id uuid.UUID) {
		t.Helper()
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneHand, Owner: owner}, id); err != nil {
			t.Fatalf("bounce: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	bounce(me.ID, land)
	bounce(opp.ID, theirs)
	if got := plusOnes(g, justice); got != 0 {
		t.Fatalf("Justice has %d counters after a land and an opponent's creature went home, want 0", got)
	}
	bounce(me.ID, bear)
	if got := plusOnes(g, justice); got != 1 {
		t.Errorf("Justice has %d counters after my Bear returned, want 1", got)
	}
	killCreature(t, g, me.ID, b12Creature(g, me.ID, "Doomed", "Creature — Bear", 1, 1))
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, justice); got != 1 {
		t.Errorf("Justice has %d counters after a creature died, want it to stay 1", got)
	}
}

// --- Ka-Zar of the Savage Land ----------------------------------------------

func TestMshB5KaZarCreatesZabuAndLandfallGrowsHim(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	castCatalogSpell(t, g, "Ka-Zar of the Savage Land", "Legendary Creature — Human Barbarian Hero", mshB5KaZarOracle, nil)
	passPriorityAroundTable(t, g)

	zabu := mshB1Token(t, g, me.ID, "Cat")
	if zabu.Name != "Zabu" || !mshB1HasType(zabu, "Legendary") {
		t.Fatalf("token is %q (%s), want the legendary Zabu", zabu.Name, zabu.TypeLine)
	}
	if p, tough := effectivePTOf(t, g, zabu.InstanceID); p != 2 || tough != 2 {
		t.Errorf("Zabu is %d/%d, want 2/2", p, tough)
	}
	if len(zabu.Colors) != 1 || zabu.Colors[0] != "G" {
		t.Errorf("Zabu's colors = %v, want green", zabu.Colors)
	}

	// Landfall: a land played off the top of the library grows Zabu.
	top := seedTopOfLibrary(me, game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"})
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("play the land off the top: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, zabu.InstanceID); got != 1 {
		t.Errorf("Zabu has %d +1/+1 counters, want 1", got)
	}
}

func TestMshB5KaZarPlaysLandsFromTheTopButNotOtherCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushCatalogPermanent(g, me.ID, "Ka-Zar of the Savage Land", "Legendary Creature — Human Barbarian Hero", mshB5KaZarOracle, false)
	if spec, _ := Lookup(mshB5KaZarOracle); spec.LibraryTopVisible != game.LibraryTopOwner {
		t.Errorf("Ka-Zar's library-top look = %v, want the owner's private look", spec.LibraryTopVisible)
	}

	bear := seedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{G}"})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "library"}); err == nil {
		t.Error("a nonland card was cast from the top of the library")
	}
	land := seedTopOfLibrary(me, game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"})
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Errorf("a land from the top was refused: %v", err)
	}
}

// --- Killmonger, Scourge of Wakanda -----------------------------------------

func TestMshB5KillmongerGetsBonusWithTwoCreatureCardsInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	killmonger := b12Push(g, me.ID, "Killmonger, Scourge of Wakanda", "Legendary Creature — Human Mercenary Villain", mshB5KillmongerOracle, 3, 3)
	killCreature(t, g, me.ID, b12Creature(g, me.ID, "Dead Bear", "Creature — Bear", 2, 2))
	killCreature(t, g, me.ID, b12Permanent(g, me.ID, "Dead Rock", "Artifact"))
	if p, tough := effectivePTOf(t, g, killmonger); p != 3 || tough != 3 {
		t.Fatalf("Killmonger is %d/%d with one creature card in the graveyard, want 3/3", p, tough)
	}
	killCreature(t, g, me.ID, b12Creature(g, me.ID, "Dead Wolf", "Creature — Wolf", 2, 2))
	if p, tough := effectivePTOf(t, g, killmonger); p != 5 || tough != 4 {
		t.Errorf("Killmonger is %d/%d with two creature cards in the graveyard, want 5/4", p, tough)
	}
}

func TestMshB5KillmongerSacrificesAnotherCreatureToDestroyANonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	forest := seedLand(g, opp.ID, "Forest", "Basic Land — Forest", "")

	killmonger := castCatalogSpell(t, g, "Killmonger, Scourge of Wakanda", "Legendary Creature — Human Mercenary Villain", mshB5KillmongerOracle, nil)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	sac := sacrificeChoiceFor(g, me.ID)
	if sac == nil {
		t.Fatal("Killmonger did not ask which creature to sacrifice")
	}
	if len(sac.SacrificeOptions) != 1 || sac.SacrificeOptions[0] != bear {
		t.Errorf("sacrifice offered %v, want only the other creature %v", sac.SacrificeOptions, bear)
	}
	answerSacrifice(t, g, me.ID, bear)
	mshB2PickIfAsked(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the opponent's nonland permanent was not destroyed")
	}
	if !g.Battlefield.Contains(forest) || !g.Battlefield.Contains(killmonger) {
		t.Error("the land or Killmonger left the battlefield")
	}
}

func TestMshB5KillmongerDecliningSacrificesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")

	castCatalogSpell(t, g, "Killmonger, Scourge of Wakanda", "Legendary Creature — Human Mercenary Villain", mshB5KillmongerOracle, nil)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(rock) {
		t.Error("declining still sacrificed or destroyed something")
	}
}

func TestMshB5KillmongerAloneAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Killmonger, Scourge of Wakanda", "Legendary Creature — Human Mercenary Villain", mshB5KillmongerOracle, nil)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("Killmonger offered a sacrifice with no other creature")
	}
}

// --- Iron Man Armor ---------------------------------------------------------

func TestMshB5IronManArmorAttachesWhenItEntersAndEquips(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Iron Man Armor", "Artifact — Equipment", mshB5IronManArmorOracle, nil)
	passPriorityAroundTable(t, g)
	mshB2PickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("the equipped Bear is %d/%d, want 4/3", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("the equipped Bear has no flying")
	}
}

func TestMshB5IronManArmorBecomesAnArtifactCreatureThatCountsArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	armor := b12Push(g, me.ID, "Iron Man Armor", "Artifact — Equipment", mshB5IronManArmorOracle, 0, 0)

	floatMana(t, g, me, "{C}{C}")
	b16Activate(t, g, me.ID, armor, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	card := gaLayered(t, g, armor)
	if !card.IsArtifact() || !card.IsCreature() || !card.HasSubtype("Construct") || !card.HasSubtype("Hero") || !card.HasSubtype("Equipment") {
		t.Errorf("the animated Armor is not an Equipment Construct Hero artifact creature: %+v", card.Effective())
	}
	if !hasKeywordOnBattlefield(t, g, armor, "flying") {
		t.Error("the animated Armor has no flying")
	}
	if p, tough := effectivePTOf(t, g, armor); p != 1 || tough != 1 {
		t.Errorf("the Armor is %d/%d, want 1/1 (0/0 plus one artifact: itself)", p, tough)
	}
	b12Permanent(g, me.ID, "My Rock", "Artifact")
	if p, tough := effectivePTOf(t, g, armor); p != 2 || tough != 2 {
		t.Errorf("the Armor is %d/%d with a second artifact, want 2/2", p, tough)
	}

	advancePastCleanupForTest(t, g)
	if gaLayered(t, g, armor).IsCreature() {
		t.Error("the Armor is still a creature after the turn ended")
	}
}

// --- Loki Laufeyson ---------------------------------------------------------

// mshB5CastBolt casts a Lightning Bolt of the given cost from `caster`'s hand
// at `victim`, keeps every copy's target, and settles the stack.
func mshB5CastBolt(t *testing.T, g *game.Game, caster *game.Player, victim uuid.UUID, cost, pay string, wantCopies int) {
	t.Helper()
	floatMana(t, g, caster, pay)
	castFromHandForTest(t, g, caster, "Lightning Bolt", "Instant", cost, lightningBoltOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: victim}}})
	answered := 0
	for i := 0; i < 48; i++ {
		if p := triggerOrderPromptFor(g, caster.ID); p != nil {
			if err := g.ResolveTriggerOrder(p.ID, caster.ID, append([]uuid.UUID(nil), p.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := latestPickTarget(g, caster.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, caster.ID, game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			answered++
			continue
		}
		if stackFullyEmpty(g) {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if answered != wantCopies {
		t.Fatalf("answered %d copy-target prompts, want %d", answered, wantCopies)
	}
	passPriorityAroundTable(t, g)
}

func TestMshB5LokiCopiesTheNextCheapEnoughInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	loki := b12Push(g, me.ID, "Loki Laufeyson", "Legendary Creature — God Sorcerer Villain", mshB5LokiOracle, 2, 1)

	if err := hobB4Activate(t, g, me.ID, loki, 0, "{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, loki) {
		t.Error("Loki did not tap")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("%d delayed triggers, want 1", len(g.DelayedTriggers))
	}

	// A mana value 3 spell is above Loki's power 2: it is not copied and
	// the trigger keeps waiting.
	before := opp.Life
	mshB5CastBolt(t, g, me, opp.ID, "{2}{R}", "{C}{C}{R}", 0)
	if opp.Life != before-3 {
		t.Errorf("life %d -> %d, want one Bolt only", before, opp.Life)
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("the trigger was used up by a spell that was too expensive: %d left", len(g.DelayedTriggers))
	}

	// A creature spell is not an instant or sorcery.
	castPowerUpCreature(t, g, me, "Test Bear", "Creature — Bear", "{1}", "{C}", "", 2)
	passPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("a creature spell used the trigger up: %d left", len(g.DelayedTriggers))
	}

	// Mana value 1 is within the power: copied.
	before = opp.Life
	mshB5CastBolt(t, g, me, opp.ID, "{R}", "{R}", 1)
	if opp.Life != before-6 {
		t.Errorf("life %d -> %d, want the Bolt and its copy", before, opp.Life)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers left, want it used up", len(g.DelayedTriggers))
	}
}

func TestMshB5LokiReadsHisPowerWhenTheSpellIsCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	loki := b12Push(g, me.ID, "Loki Laufeyson", "Legendary Creature — God Sorcerer Villain", mshB5LokiOracle, 2, 1)
	if err := hobB4Activate(t, g, me.ID, loki, 0, "{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	// Two counters make him a 4/3: a mana value 3 spell now counts.
	if err := g.AddCounterBy(me.ID, loki, game.CounterPlusOne, 2); err != nil {
		t.Fatal(err)
	}
	before := opp.Life
	mshB5CastBolt(t, g, me, opp.ID, "{2}{R}", "{C}{C}{R}", 1)
	if opp.Life != before-6 {
		t.Errorf("life %d -> %d, want the Bolt and its copy", before, opp.Life)
	}
}

func TestMshB5LokiPowerUpPutsTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	loki := castPowerUpCreature(t, g, me, "Loki Laufeyson", "Legendary Creature — God Sorcerer Villain", "{1}{R}", "{C}{R}", mshB5LokiOracle, 2)

	price, err := g.PriceActivation(me.ID, loki, 1, nil)
	if err != nil {
		t.Fatalf("PriceActivation: %v", err)
	}
	if got := price.Total.String(); got != "{3}" {
		t.Fatalf("the turn it entered: %q, want {3} ({4}{R} less {1}{R})", got)
	}
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, loki, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	if got := plusOnes(g, loki); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
}

// --- Jennifer Walters // The Sensational She-Hulk ---------------------------

// mshB5JenniferCard is the card as the deck importer makes it: both faces,
// face 0 materialised.
func mshB5JenniferCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   jenniferWaltersOracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Jennifer Walters", TypeLine: "Legendary Creature — Human Advisor Hero", ManaCost: "{2}{G}{W}", Colors: []string{"G", "W"}, Power: 3, Toughness: 4},
			{Name: "The Sensational She-Hulk", TypeLine: "Legendary Creature — Gamma Hero", Colors: []string{"G", "W"}, Power: 6, Toughness: 6},
		},
	}
	c.SetFace(0)
	return c
}

func TestMshB5JenniferStopsOpponentsCastingOnYourTurnOnBothFaces(t *testing.T) {
	for face := 0; face < 2; face++ {
		g := newCatalogGame(t)
		active := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		jen := mshB5JenniferCard(active.ID)
		jen.SetFace(face)
		pushBattlefieldCardWithTimestamp(g, jen)
		advanceToMain(t, g)
		bolt := handCardFull(opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle, []string{"R"})
		err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: active.ID}},
		})
		if err == nil {
			t.Errorf("face %d: an opponent cast a spell during its controller's turn", face)
		}
		if _, err := castWithOptionalCosts(t, g, "Burst Lightning", "Instant", burstLightningOracle,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, nil, nil); err != nil {
			t.Errorf("face %d: the controller was refused: %v", face, err)
		}
	}
}

func TestMshB5JenniferTransformsAtSorcerySpeedOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jen := mshB5JenniferCard(me.ID)
	pushBattlefieldCardWithTimestamp(g, jen)
	advanceToMain(t, g)

	if err := hobB4Activate(t, g, me.ID, jen.InstanceID, 0, "{C}{C}{C}{G}{W}{W}", nil); err != nil {
		t.Fatalf("transform: %v", err)
	}
	passPriorityAroundTable(t, g)
	card := battlefieldCardFor(g, jen.InstanceID)
	if card == nil || card.ActiveFace != 1 || card.Name != "The Sensational She-Hulk" {
		t.Fatalf("Jennifer did not transform: %+v", card)
	}
	for _, kw := range []string{"reach", "trample"} {
		if !hasKeywordOnBattlefield(t, g, jen.InstanceID, kw) {
			t.Errorf("She-Hulk lacks %s", kw)
		}
	}

	// Not at instant speed: with a spell on the stack the ability is refused.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	jen2 := mshB5JenniferCard(me2.ID)
	pushBattlefieldCardWithTimestamp(g2, jen2)
	advanceToMain(t, g2)
	castCatalogSpell(t, g2, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: g2.Seats[1].ID}})
	if err := hobB4Activate(t, g2, me2.ID, jen2.InstanceID, 0, "{C}{C}{C}{G}{W}{W}", nil); err == nil {
		t.Error("Jennifer transformed with a spell on the stack")
	}
}

// mshB5ReflectDamage deals 3 to `mine` with a Bolt, picks `victim` for the
// reflection, and answers the "you may" with `yes`.
func mshB5ReflectDamage(t *testing.T, g *game.Game, me *game.Player, mine, victim uuid.UUID, yes bool) {
	t.Helper()
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)
	mshB5PickPlayerIfAsked(t, g, me.ID, victim)
	if latestTriggerPrompt(g, me.ID) != nil || stackHasTriggerFor(g, me.ID) {
		answerLatestTriggerPrompt(t, g, me.ID, yes)
	}
	passPriorityAroundTable(t, g)
}

// stackHasTriggerFor reports whether a triggered ability controlled by
// `who` is still waiting.
func stackHasTriggerFor(g *game.Game, who uuid.UUID) bool {
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.Controller == who {
			return true
		}
	}
	return len(g.PendingTriggers) > 0
}

func TestMshB5SheHulkReflectsDamageDealtToYourCreatureOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	jen := mshB5JenniferCard(me.ID)
	jen.SetFace(1)
	pushBattlefieldCardWithTimestamp(g, jen)
	ox := b12Creature(g, me.ID, "Ox", "Creature — Ox", 1, 9)

	before := opp.Life
	mshB5ReflectDamage(t, g, me, ox, opp.ID, true)
	if opp.Life != before-3 {
		t.Fatalf("opponent life %d -> %d, want 3 reflected", before, opp.Life)
	}

	// Only once each turn: a second hit reflects nothing.
	mshB5ReflectDamage(t, g, me, ox, opp.ID, true)
	if opp.Life != before-3 {
		t.Errorf("opponent life %d -> %d, want no second reflection this turn", before, opp.Life)
	}
}

func TestMshB5SheHulkDecliningDoesNotUseTheTurnUp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	jen := mshB5JenniferCard(me.ID)
	jen.SetFace(1)
	pushBattlefieldCardWithTimestamp(g, jen)
	ox := b12Creature(g, me.ID, "Ox", "Creature — Ox", 1, 9)

	before := opp.Life
	mshB5ReflectDamage(t, g, me, ox, opp.ID, false)
	if opp.Life != before {
		t.Fatalf("declined, yet the opponent lost %d", before-opp.Life)
	}
	mshB5ReflectDamage(t, g, me, ox, opp.ID, true)
	if opp.Life != before-3 {
		t.Errorf("opponent life %d -> %d, want 3 reflected after declining the first", before, opp.Life)
	}
}

func TestMshB5SheHulkIgnoresDamageToAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	jen := mshB5JenniferCard(me.ID)
	jen.SetFace(1)
	pushBattlefieldCardWithTimestamp(g, jen)
	theirs := b12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 1, 9)

	before := opp.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("She-Hulk reacted to damage dealt to an opponent's creature")
	}
	if opp.Life != before {
		t.Errorf("opponent life %d -> %d, want unchanged", before, opp.Life)
	}
}
