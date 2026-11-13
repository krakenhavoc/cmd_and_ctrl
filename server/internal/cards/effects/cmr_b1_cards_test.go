package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmr_b1_cards_test.go — Commander Legends, batch 1 (#5507). Each test
// plays the real catalog entry and asserts what its text says.

const (
	cmrB1FalthisOracle     = "6a0b1b7e-1a93-4c00-be94-8fa56af57468"
	cmrB1KelethOracle      = "6df7f4b4-a3a8-44d0-845f-84304bf6a3f8"
	cmrB1PravaOracle       = "4daa7b7d-e410-4bcf-b3e2-2751774bef47"
	cmrB1TuyaOracle        = "63e8ac70-0ec3-4bef-832f-e3f4c8a692ad"
	cmrB1KangeeOracle      = "6a344d2e-fbd7-4ff9-9e2a-97ae69d60c66"
	cmrB1SianiOracle       = "ecf808fd-4828-497e-b84b-1cb018a9924b"
	cmrB1DawngladeOracle   = "bb1c6266-5e25-4a5b-b15a-faeda3ce04dc"
	cmrB1PrideOracle       = "83e2b251-bd60-4c5c-b9e2-32c22b28fb8b"
	cmrB1HuntmasterOracle  = "393ef069-ea6b-4969-b878-d942281cc918"
	cmrB1TroveOracle       = "f27972da-ad2f-4df5-82d3-c909cce8cc91"
	cmrB1CommodoreOracle   = "d5ac319a-c870-4890-87c0-0e02f122f234"
	cmrB1ThroneguardOracle = "c8131fb3-3146-4434-9e4b-ac88ab3cbe37"
)

func TestCmrB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		cmrB1FalthisOracle:     "Falthis, Shadowcat Familiar",
		cmrB1KelethOracle:      "Keleth, Sunmane Familiar",
		cmrB1PravaOracle:       "Prava of the Steel Legion",
		cmrB1TuyaOracle:        "Tuya Bearclaw",
		cmrB1KangeeOracle:      "Kangee, Sky Warden",
		cmrB1SianiOracle:       "Siani, Eye of the Storm",
		cmrB1DawngladeOracle:   "Dawnglade Regent",
		cmrB1PrideOracle:       "Pride of the Perfect",
		cmrB1HuntmasterOracle:  "Exquisite Huntmaster",
		cmrB1TroveOracle:       "Trove Tracker",
		cmrB1CommodoreOracle:   "Crimson Fleet Commodore",
		cmrB1ThroneguardOracle: "Staunch Throneguard",
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

// cmrB1Push seeds a creature on the battlefield, catalog row or not,
// untapped and able to attack.
func cmrB1Push(g *game.Game, owner uuid.UUID, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner, owner
	return pushBattlefieldCardWithTimestamp(g, c)
}

// cmrB1Flyer is a vanilla creature with flying.
func cmrB1Flyer(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return cmrB1Push(g, owner, game.Card{Name: name, TypeLine: "Creature — Bird", Power: power, Toughness: toughness, Keywords: []string{"flying"}})
}

// cmrB1Commander is a vanilla commander creature.
func cmrB1Commander(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return cmrB1Push(g, owner, game.Card{Name: name, TypeLine: "Legendary Creature — Human", Power: 3, Toughness: 3, IsCommander: true})
}

// cmrB1WantPT asserts a creature's current power and toughness.
func cmrB1WantPT(t *testing.T, g *game.Game, id uuid.UUID, what string, power, toughness int) {
	t.Helper()
	if p, tg := ptOf(t, g, id); p != power || tg != toughness {
		t.Errorf("%s is %d/%d, want %d/%d", what, p, tg, power, toughness)
	}
}

// --- Falthis, Shadowcat Familiar -------------------------------------------

func TestCmrB1FalthisGivesYourCommandersMenaceAndDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	falthis := cmrB1Push(g, me.ID, game.Card{Name: "Falthis, Shadowcat Familiar", TypeLine: "Legendary Creature — Nightmare Cat",
		OracleID: cmrB1FalthisOracle, Power: 2, Toughness: 2, IsCommander: true})
	partner := cmrB1Commander(g, me.ID, "My Other Commander")
	theirs := cmrB1Commander(g, opp.ID, "Their Commander")
	plain := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	for _, tc := range []struct {
		what string
		id   uuid.UUID
		want bool
	}{
		{"Falthis as one of my commanders", falthis, true},
		{"my other commander", partner, true},
		{"an opponent's commander", theirs, false},
		{"my non-commander creature", plain, false},
	} {
		for _, kw := range []string{"menace", "deathtouch"} {
			if got := hasEffectiveKeyword(t, g, tc.id, kw); got != tc.want {
				t.Errorf("%s has %s = %v, want %v", tc.what, kw, got, tc.want)
			}
		}
	}
}

// --- Keleth, Sunmane Familiar ----------------------------------------------

func TestCmrB1KelethPutsACounterOnEachAttackingCommanderOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmrB1Push(g, me.ID, game.Card{Name: "Keleth, Sunmane Familiar", TypeLine: "Legendary Creature — Horse",
		OracleID: cmrB1KelethOracle, Power: 1, Toughness: 1})
	commander := cmrB1Commander(g, me.ID, "My Commander")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, commander, bear)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, commander); got != 1 {
		t.Errorf("the attacking commander has %d +1/+1 counters, want 1", got)
	}
	if got := plusOnes(g, bear); got != 0 {
		t.Errorf("a non-commander attacker has %d +1/+1 counters, want 0", got)
	}
}

// --- Prava of the Steel Legion ----------------------------------------------

func TestCmrB1PravaPumpsYourCreatureTokensOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if g.Seats[g.Turn.ActiveSeat] != me {
		t.Fatal("setup: seat 0 is not the active player")
	}
	cmrB1Push(g, me.ID, game.Card{Name: "Prava of the Steel Legion", TypeLine: "Legendary Creature — Cat Soldier",
		OracleID: cmrB1PravaOracle, Power: 1, Toughness: 4})
	myToken := cmrB1Push(g, me.ID, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1})
	myBear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirToken := cmrB1Push(g, opp.ID, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1})
	cmrB1WantPT(t, g, myToken, "my token on my turn", 2, 5)
	cmrB1WantPT(t, g, myBear, "my nontoken creature", 2, 2)
	cmrB1WantPT(t, g, theirToken, "an opponent's token", 1, 1)

	// Prava under a player whose turn it is not pumps nothing.
	g = newCatalogGame(t)
	opp = g.Seats[1]
	cmrB1Push(g, opp.ID, game.Card{Name: "Prava of the Steel Legion", TypeLine: "Legendary Creature — Cat Soldier",
		OracleID: cmrB1PravaOracle, Power: 1, Toughness: 4})
	theirToken = cmrB1Push(g, opp.ID, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1})
	cmrB1WantPT(t, g, theirToken, "a token on its controller's opponent's turn", 1, 1)
}

func TestCmrB1PravaMakesASoldierForFourMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	prava := cmrB1Push(g, me.ID, game.Card{Name: "Prava of the Steel Legion", TypeLine: "Legendary Creature — Cat Soldier",
		OracleID: cmrB1PravaOracle, Power: 1, Toughness: 4})
	if err := hobB4Activate(t, g, me.ID, prava, 0, "{C}{C}{W}", nil); err == nil {
		t.Fatal("activated for three mana; the cost is {3}{W}")
	}
	if err := hobB4Activate(t, g, me.ID, prava, 0, "{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	toks := s46Tokens(g, me.ID, "Soldier")
	if len(toks) != 1 {
		t.Fatalf("%d Soldier tokens, want 1", len(toks))
	}
	if !toks[0].HasColor("W") {
		t.Error("the Soldier is not white")
	}
	cmrB1WantPT(t, g, toks[0].InstanceID, "the new Soldier on my turn", 2, 5)
}

// --- Tuya Bearclaw ------------------------------------------------------------

func TestCmrB1TuyaGetsTheGreatestPowerAmongYourOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tuya := cmrB1Push(g, me.ID, game.Card{Name: "Tuya Bearclaw", TypeLine: "Legendary Creature — Human Warrior",
		OracleID: cmrB1TuyaOracle, Power: 2, Toughness: 2})
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "My Giant", "Creature — Giant", 4, 4)
	b12Creature(g, opp.ID, "Their Titan", "Creature — Giant", 9, 9)
	declareAttack(t, g, opp.ID, tuya)
	passPriorityAroundTable(t, g)
	cmrB1WantPT(t, g, tuya, "Tuya beside a 4-power creature", 6, 6)
}

func TestCmrB1TuyaAloneGetsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tuya := cmrB1Push(g, me.ID, game.Card{Name: "Tuya Bearclaw", TypeLine: "Legendary Creature — Human Warrior",
		OracleID: cmrB1TuyaOracle, Power: 2, Toughness: 2})
	b12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, tuya)
	passPriorityAroundTable(t, g)
	cmrB1WantPT(t, g, tuya, "Tuya beside a 0-power creature", 2, 2)
}

// --- Kangee, Sky Warden ------------------------------------------------------

func TestCmrB1KangeeAttackingPumpsAttackingFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kangee := cmrB1Push(g, me.ID, game.Card{Name: "Kangee, Sky Warden", TypeLine: "Legendary Creature — Bird Wizard",
		OracleID: cmrB1KangeeOracle, Power: 3, Toughness: 3})
	flyer := cmrB1Flyer(g, me.ID, "My Bird", 1, 1)
	home := cmrB1Flyer(g, me.ID, "Bird at Home", 1, 1)
	ground := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, kangee, flyer, ground)
	passPriorityAroundTable(t, g)
	cmrB1WantPT(t, g, kangee, "Kangee", 5, 3)
	cmrB1WantPT(t, g, flyer, "an attacking flyer", 3, 1)
	cmrB1WantPT(t, g, ground, "an attacker without flying", 2, 2)
	cmrB1WantPT(t, g, home, "a flyer that did not attack", 1, 1)
	if !hasEffectiveKeyword(t, g, kangee, "vigilance") {
		t.Error("Kangee has no vigilance")
	}
}

func TestCmrB1KangeeBlockingPumpsBlockingFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kangee := cmrB1Push(g, opp.ID, game.Card{Name: "Kangee, Sky Warden", TypeLine: "Legendary Creature — Bird Wizard",
		OracleID: cmrB1KangeeOracle, Power: 3, Toughness: 3})
	bird := cmrB1Flyer(g, opp.ID, "Their Bird", 1, 1)
	home := cmrB1Flyer(g, opp.ID, "Bird at Home", 1, 1)
	first := b12Creature(g, me.ID, "First Bear", "Creature — Bear", 1, 1)
	second := b12Creature(g, me.ID, "Second Bear", "Creature — Bear", 1, 1)
	declareAttack(t, g, opp.ID, first, second)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(kangee, first); err != nil {
		t.Fatalf("DeclareBlocker Kangee: %v", err)
	}
	if err := g.DeclareBlocker(bird, second); err != nil {
		t.Fatalf("DeclareBlocker bird: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	cmrB1WantPT(t, g, kangee, "Kangee", 3, 5)
	cmrB1WantPT(t, g, bird, "a blocking flyer", 1, 3)
	cmrB1WantPT(t, g, home, "a flyer that did not block", 1, 1)
	cmrB1WantPT(t, g, first, "the attacker Kangee blocks", 1, 1)
}

// --- Siani, Eye of the Storm -------------------------------------------------

func TestCmrB1SianiScriesOneForEachAttackingFlyer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "One", "Two", "Three", "Four")
	siani := cmrB1Push(g, me.ID, game.Card{Name: "Siani, Eye of the Storm", TypeLine: "Legendary Creature — Djinn Monk",
		OracleID: cmrB1SianiOracle, Power: 3, Toughness: 2})
	flyer := cmrB1Flyer(g, me.ID, "My Bird", 1, 1)
	cmrB1Flyer(g, me.ID, "Bird at Home", 1, 1)
	ground := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, siani, flyer, ground)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("attacking did not scry")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scry %d, want 2: Siani and the other attacking flyer", len(c.ScryCards))
	}
}

// --- Dawnglade Regent ---------------------------------------------------------

func TestCmrB1DawngladeRegentCrownsYouAndHexproofsYourPermanentsWhileYouWearIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, me.ID, "My Forest", "Basic Land — Forest")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if hasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Fatal("setup: my creature has hexproof already")
	}
	regent := castHoldCreature(t, g, "Dawnglade Regent", "Creature — Elk", cmrB1DawngladeOracle, 8, 8)
	passPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Fatal("the enters trigger did not make me the monarch")
	}
	for _, tc := range []struct {
		what string
		id   uuid.UUID
		want bool
	}{
		{"the Regent", regent, true},
		{"my creature", mine, true},
		{"my land", land, true},
		{"an opponent's creature", theirs, false},
	} {
		if got := hasEffectiveKeyword(t, g, tc.id, "hexproof"); got != tc.want {
			t.Errorf("%s has hexproof = %v, want %v", tc.what, got, tc.want)
		}
	}
	if err := g.SetMonarch(opp.ID); err != nil {
		t.Fatalf("SetMonarch: %v", err)
	}
	if hasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("my creature kept hexproof after the crown moved")
	}
}

// --- Pride of the Perfect ------------------------------------------------------

func TestCmrB1PrideOfThePerfectPumpsYourElves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Pride of the Perfect", "Enchantment", cmrB1PrideOracle, false)
	elf := b12Creature(g, me.ID, "My Elf", "Creature — Elf Warrior", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirElf := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	cmrB1WantPT(t, g, elf, "my Elf", 3, 1)
	cmrB1WantPT(t, g, bear, "my non-Elf", 2, 2)
	cmrB1WantPT(t, g, theirElf, "an opponent's Elf", 1, 1)
}

// --- Exquisite Huntmaster -------------------------------------------------------

func TestCmrB1ExquisiteHuntmasterDiesIntoAnElfWarrior(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hunt := cmrB1Push(g, me.ID, game.Card{Name: "Exquisite Huntmaster", TypeLine: "Creature — Elf Warrior",
		OracleID: cmrB1HuntmasterOracle, Power: 4, Toughness: 2})
	trekB18BDestroy(t, g, hunt)
	passPriorityAroundTable(t, g)
	toks := s46Tokens(g, me.ID, "Elf Warrior")
	if len(toks) != 1 {
		t.Fatalf("%d Elf Warrior tokens, want 1", len(toks))
	}
	cmrB1WantPT(t, g, toks[0].InstanceID, "the Elf Warrior", 1, 1)
	if !toks[0].HasColor("G") {
		t.Error("the Elf Warrior is not green")
	}
}

func TestCmrB1ExquisiteHuntmasterEncoreMakesATokenPerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	_, me := s46Encore(t, g, "Exquisite Huntmaster", "Creature — Elf Warrior", cmrB1HuntmasterOracle, 4, 2, "{C}{C}{C}{C}{B}")
	if n := len(s46Tokens(g, me.ID, "Exquisite Huntmaster")); n != 3 {
		t.Fatalf("%d tokens, want one per opponent", n)
	}
}

// --- Trove Tracker ---------------------------------------------------------------

func TestCmrB1TroveTrackerDrawsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	tracker := cmrB1Push(g, me.ID, game.Card{Name: "Trove Tracker", TypeLine: "Creature — Human Pirate",
		OracleID: cmrB1TroveOracle, Power: 2, Toughness: 2})
	before := handSize(me)
	trekB18BDestroy(t, g, tracker)
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}

func TestCmrB1TroveTrackerEncoreMakesATokenPerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	_, me := s46Encore(t, g, "Trove Tracker", "Creature — Human Pirate", cmrB1TroveOracle, 2, 2, "{C}{C}{C}{C}{C}{U}{U}")
	if n := len(s46Tokens(g, me.ID, "Trove Tracker")); n != 3 {
		t.Fatalf("%d tokens, want one per opponent", n)
	}
}

// --- Crimson Fleet Commodore, Staunch Throneguard -----------------------------

func TestCmrB1MonarchCreaturesCrownYouAndKeepTheirKeyword(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
		power, toughness                int
	}{
		{"Crimson Fleet Commodore", "Creature — Ogre Pirate", cmrB1CommodoreOracle, "trample", 5, 2},
		{"Staunch Throneguard", "Artifact Creature — Construct", cmrB1ThroneguardOracle, "vigilance", 2, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			passPriorityAroundTable(t, g)
			if g.Monarch != me.ID {
				t.Error("the enters trigger did not make me the monarch")
			}
			if !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("%s has no %s", tc.name, tc.keyword)
			}
		})
	}
}
