package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// c17_5533_b1_cards_test.go — Commander 2017 batch 1 (#5533). Every test
// plays the real catalog entry through the engine.

const (
	c17CurseOfVitalityOracle     = "6d2eb412-e09b-4ab3-8653-b458e9f38eff"
	c17CurseOfDisturbanceOracle  = "6cbd36d9-de47-41b0-9ef4-a72ca01adccd"
	c17CurseOfBountyOracle       = "b6689782-08d8-48e1-a05d-cd040dfe85bc"
	c17BalanOracle               = "d12cfd46-73c2-4206-9a10-0d2e7d6ac87e"
	c17BoneyardScourgeOracle     = "eef40ab2-9e17-4e8d-9e58-1a82ab871b5b"
	c17GalecasterColossusOracle  = "d8379d8f-df17-4473-938d-7ac198d17980"
	c17HungryLynxOracle          = "a5fdc358-25de-4a5b-b965-53a01df6019c"
	c17LiciaOracle               = "57170e91-6ebe-4b14-bf56-96d8cf6a4d03"
	c17MagusOfTheMindOracle      = "b39d98c3-0173-4d7c-83d0-f80db82fc9ba"
	c17MirrorOfTheForebearsOracl = "9eaeb54b-73a8-43cc-b4d8-ed8b5eb19646"
	c17NazahnOracle              = "a3a6c024-43de-4f5c-a1d6-c48a0c96a43b"
	c17QasaliSlingersOracle      = "f763d118-3183-4604-a50c-7c3933e2185e"
	c17TaigamOracle              = "677aa518-072b-46d2-b0da-da57c7bc2bba"
	c17VindictiveLichOracle      = "6dc7f837-158f-4ccf-86a5-a07a20891ec5"
	c17WasitoraOracle            = "13fc4168-fcc2-4011-804f-211d5d86b7dd"

	c17CurseType  = "Enchantment — Aura Curse"
	c17NazahnType = "Legendary Creature — Cat Artificer"
	c17LiciaType  = "Legendary Creature — Vampire Soldier"
)

func TestC17B1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		c17CurseOfVitalityOracle:     "Curse of Vitality",
		c17CurseOfDisturbanceOracle:  "Curse of Disturbance",
		c17CurseOfBountyOracle:       "Curse of Bounty",
		c17BalanOracle:               "Balan, Wandering Knight",
		c17BoneyardScourgeOracle:     "Boneyard Scourge",
		c17GalecasterColossusOracle:  "Galecaster Colossus",
		c17HungryLynxOracle:          "Hungry Lynx",
		c17LiciaOracle:               "Licia, Sanguine Tribune",
		c17MagusOfTheMindOracle:      "Magus of the Mind",
		c17MirrorOfTheForebearsOracl: "Mirror of the Forebears",
		c17NazahnOracle:              "Nazahn, Revered Bladesmith",
		c17QasaliSlingersOracle:      "Qasali Slingers",
		c17TaigamOracle:              "Taigam, Sidisi's Hand",
		c17VindictiveLichOracle:      "Vindictive Lich",
		c17WasitoraOracle:            "Wasitora, Nekoru Queen",
	} {
		stdRegisteredFull(t, oracle, name)
	}
}

// c17Count is how many battlefield permanents controlled by `controller`
// match.
func c17Count(g *game.Game, controller uuid.UUID, match func(game.Card) bool) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && match(c) {
			n++
		}
	}
	return n
}

func c17Named(name string) func(game.Card) bool {
	return func(c game.Card) bool { return c.Name == name }
}

// --- the Curses ------------------------------------------------------------

// c17CurseTable is a table with a Curse on the seat after the active one,
// controlled by the seat after that, and two creatures of the active
// player's ready to attack it.
type c17CurseTable struct {
	g                        *game.Game
	active, victim, curseCtl *game.Player
	attackers                []uuid.UUID
}

func c17CurseSetup(t *testing.T, name, oracle string) c17CurseTable {
	t.Helper()
	g := newCatalogGame(t)
	n := len(g.Seats)
	tb := c17CurseTable{
		g:        g,
		active:   g.Seats[g.Turn.ActiveSeat],
		victim:   g.Seats[(g.Turn.ActiveSeat+1)%n],
		curseCtl: g.Seats[(g.Turn.ActiveSeat+2)%n],
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: c17CurseType,
		OracleID: oracle, Owner: tb.curseCtl.ID, Controller: tb.curseCtl.ID,
		AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: tb.victim.ID},
	})
	tb.attackers = []uuid.UUID{
		blbB1Push(g, tb.active.ID, "Bear", "Creature — Bear", "", 2, 2),
		blbB1Push(g, tb.active.ID, "Bear 2", "Creature — Bear", "", 2, 2),
	}
	return tb
}

// attack declares both attackers at the cursed player and resolves the
// Curse's triggers.
func (tb c17CurseTable) attack(t *testing.T) {
	t.Helper()
	g := tb.g
	for g.Turn.Step != game.StepDeclareAttackers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for _, a := range tb.attackers {
		if err := g.DeclareAttacker(a, tb.victim.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
}

func TestC17B1CurseOfVitalityGainsLifeForYouAndTheAttacker(t *testing.T) {
	tb := c17CurseSetup(t, "Curse of Vitality", c17CurseOfVitalityOracle)
	mine, theirs, victim := tb.curseCtl.Life, tb.active.Life, tb.victim.Life
	tb.attack(t)
	if got := tb.curseCtl.Life - mine; got != 2 {
		t.Errorf("the Curse's controller gained %d, want 2", got)
	}
	if got := tb.active.Life - theirs; got != 2 {
		t.Errorf("the attacking opponent gained %d, want 2 (two attackers, one declaration)", got)
	}
	if tb.victim.Life != victim {
		t.Errorf("the enchanted player's life moved to %d", tb.victim.Life)
	}
}

func TestC17B1CurseOfDisturbanceMakesAZombieForYouAndTheAttacker(t *testing.T) {
	tb := c17CurseSetup(t, "Curse of Disturbance", c17CurseOfDisturbanceOracle)
	tb.attack(t)
	if got := c17Count(tb.g, tb.curseCtl.ID, c17Named("Zombie")); got != 1 {
		t.Errorf("the Curse's controller has %d Zombies, want 1", got)
	}
	if got := c17Count(tb.g, tb.active.ID, c17Named("Zombie")); got != 1 {
		t.Errorf("the attacking opponent has %d Zombies, want 1 (two attackers, one declaration)", got)
	}
	if got := c17Count(tb.g, tb.victim.ID, c17Named("Zombie")); got != 0 {
		t.Errorf("the enchanted player has %d Zombies, want 0", got)
	}
}

func TestC17B1CurseOfBountyUntapsNonlandPermanentsForYouAndTheAttacker(t *testing.T) {
	tb := c17CurseSetup(t, "Curse of Bounty", c17CurseOfBountyOracle)
	g := tb.g
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact",
		Owner: tb.curseCtl.ID, Controller: tb.curseCtl.ID, Tapped: true})
	land := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Field", TypeLine: "Land",
		Owner: tb.active.ID, Controller: tb.active.ID, Tapped: true})
	victimRock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact",
		Owner: tb.victim.ID, Controller: tb.victim.ID, Tapped: true})
	tb.attack(t)
	if findBattlefieldCardForTest(g, rock).Tapped {
		t.Error("the Curse's controller's artifact is still tapped")
	}
	for _, a := range tb.attackers {
		if findBattlefieldCardForTest(g, a).Tapped {
			t.Error("an attacking creature is still tapped: the attacker untaps their nonland permanents")
		}
	}
	if !findBattlefieldCardForTest(g, land).Tapped {
		t.Error("a land was untapped")
	}
	if !findBattlefieldCardForTest(g, victimRock).Tapped {
		t.Error("the enchanted player's artifact was untapped")
	}
}

// --- Balan, Wandering Knight ---------------------------------------------------

func c17Equipment(g *game.Game, owner uuid.UUID, name string, host uuid.UUID) uuid.UUID {
	c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Artifact — Equipment", Owner: owner, Controller: owner}
	if host != uuid.Nil {
		c.AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
	}
	return pushBattlefieldCardWithTimestamp(g, c)
}

func TestC17B1BalanHasDoubleStrikeWithTwoEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	balan := b12Push(g, me.ID, "Balan, Wandering Knight", "Legendary Creature — Cat Knight", c17BalanOracle, 3, 3)
	if !hasEffectiveKeyword(t, g, balan, "first strike") {
		t.Error("Balan has no first strike")
	}
	c17Equipment(g, me.ID, "Sword", balan)
	if hasEffectiveKeyword(t, g, balan, "double strike") {
		t.Error("Balan has double strike with one Equipment attached")
	}
	c17Equipment(g, me.ID, "Shield", balan)
	if !hasEffectiveKeyword(t, g, balan, "double strike") {
		t.Error("Balan has no double strike with two Equipment attached")
	}
}

func TestC17B1BalanAttachesAllYourEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	balan := b12Push(g, me.ID, "Balan, Wandering Knight", "Legendary Creature — Cat Knight", c17BalanOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	onBear := c17Equipment(g, me.ID, "Sword", bear)
	loose := c17Equipment(g, me.ID, "Shield", uuid.Nil)
	theirs := c17Equipment(g, opp.ID, "Their Axe", uuid.Nil)
	b06AddMana(me, "W", "C")
	if err := g.ActivateCatalogAbility(me.ID, balan, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{onBear, loose} {
		if !findBattlefieldCardForTest(g, id).IsAttachedTo(balan) {
			t.Error("one of your Equipment is not attached to Balan")
		}
	}
	if findBattlefieldCardForTest(g, theirs).IsAttachedTo(balan) {
		t.Error("an opponent's Equipment was attached to Balan")
	}
	if !hasEffectiveKeyword(t, g, balan, "double strike") {
		t.Error("Balan has no double strike after taking two Equipment")
	}
}

// --- Boneyard Scourge ------------------------------------------------------------

func TestC17B1BoneyardScourgeReturnsWhenADragonDiesAndYouPay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	scourge := pushCatalogGraveyardCard(me, "Boneyard Scourge", "Creature — Zombie Dragon", c17BoneyardScourgeOracle, 4, 3)
	dragon := b12Creature(g, me.ID, "Dragon", "Creature — Dragon", 4, 4)
	advanceToMain(t, g)
	b18Kill(t, g, dragon)
	passUntilTaxed(t, g, me.ID)
	b06AddMana(me, "B", "C")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(scourge) || me.Graveyard.Contains(scourge) {
		t.Error("paying did not return Boneyard Scourge to the battlefield")
	}
}

func TestC17B1BoneyardScourgeIgnoresANonDragon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	scourge := pushCatalogGraveyardCard(me, "Boneyard Scourge", "Creature — Zombie Dragon", c17BoneyardScourgeOracle, 4, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b18Kill(t, g, bear)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("a Bear dying offered to return Boneyard Scourge")
	}
	if !me.Graveyard.Contains(scourge) {
		t.Error("Boneyard Scourge left the graveyard")
	}
}

// --- Galecaster Colossus -----------------------------------------------------------

func TestC17B1GalecasterColossusBouncesByTappingAWizard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	colossus := b12Push(g, me.ID, "Galecaster Colossus", "Creature — Giant Wizard", c17GalecasterColossusOracle, 5, 6)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	if err := g.ActivateCatalogAbility(me.ID, colossus, 0, game.ActivateAbilityParams{
		Strict: true, TapIDs: []uuid.UUID{colossus}, Targets: cardTarget(theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) {
		t.Error("the opponent's artifact did not return to its owner's hand")
	}
	if !findBattlefieldCardForTest(g, colossus).Tapped {
		t.Error("the Wizard that paid the cost is untapped")
	}
}

// --- Hungry Lynx ---------------------------------------------------------------

func TestC17B1HungryLynxProtectsYourCatsFromRats(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lynx := b12Push(g, me.ID, "Hungry Lynx", "Creature — Cat", c17HungryLynxOracle, 2, 2)
	cat := b12Creature(g, me.ID, "Alley Cat", "Creature — Cat", 1, 1)
	theirCat := b12Creature(g, opp.ID, "Their Cat", "Creature — Cat", 1, 1)
	for _, id := range []uuid.UUID{lynx, cat} {
		if !hasEffectiveKeyword(t, g, id, "protection from Rats") {
			t.Error("a Cat you control has no protection from Rats")
		}
	}
	if hasEffectiveKeyword(t, g, theirCat, "protection from Rats") {
		t.Error("an opponent's Cat has protection from Rats")
	}
}

func TestC17B1HungryLynxGrowsYourCatsWhenARatDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	lynx := b12Push(g, me.ID, "Hungry Lynx", "Creature — Cat", c17HungryLynxOracle, 2, 2)
	cat := b12Creature(g, me.ID, "Alley Cat", "Creature — Cat", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirCat := b12Creature(g, opp.ID, "Their Cat", "Creature — Cat", 1, 1)
	rat := b12Creature(g, opp.ID, "Rat", "Creature — Rat", 1, 1)
	b18Kill(t, g, rat)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{lynx, cat} {
		if got := plusOneCounters(g, id); got != 1 {
			t.Errorf("a Cat you control has %d +1/+1 counters, want 1", got)
		}
	}
	if got := plusOneCounters(g, bear); got != 0 {
		t.Errorf("a non-Cat has %d +1/+1 counters", got)
	}
	if got := plusOneCounters(g, theirCat); got != 0 {
		t.Errorf("an opponent's Cat has %d +1/+1 counters", got)
	}
}

func TestC17B1HungryLynxGivesATargetOpponentARatAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Hungry Lynx", "Creature — Cat", c17HungryLynxOracle, 2, 2)
	advanceToEndStepOf(t, g, 0)
	ref := playerTarget(opp.ID)
	assistCardSettle(t, g, me.ID, &ref, true)
	rats := c17Count(g, opp.ID, func(c game.Card) bool {
		return c.Name == "Rat" && slices.Contains(c.Keywords, "deathtouch")
	})
	if rats != 1 {
		t.Errorf("the target opponent has %d Rats with deathtouch, want 1", rats)
	}
	if got := c17Count(g, me.ID, c17Named("Rat")); got != 0 {
		t.Errorf("you have %d Rats, want 0", got)
	}
}

// --- Licia, Sanguine Tribune ---------------------------------------------------

func TestC17B1LiciaCostsOneLessPerLifeGainedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if got := selfPricedMV(t, g, me, c17LiciaOracle, c17LiciaType, "{5}{R}{W}{B}"); got != 8 {
		t.Fatalf("Licia before any life gain: %d, want 8", got)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	if got := selfPricedMV(t, g, me, c17LiciaOracle, c17LiciaType, "{5}{R}{W}{B}"); got != 5 {
		t.Errorf("Licia after gaining 3 life: %d, want 5", got)
	}
}

func TestC17B1LiciaPaysFiveLifeForThreeCountersOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	licia := b12Push(g, me.ID, "Licia, Sanguine Tribune", c17LiciaType, c17LiciaOracle, 4, 4)
	for _, kw := range []string{"first strike", "lifelink"} {
		if !hasEffectiveKeyword(t, g, licia, kw) {
			t.Errorf("Licia has no %s", kw)
		}
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, licia, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life-5 {
		t.Errorf("life %d, want %d", me.Life, life-5)
	}
	if got := plusOneCounters(g, licia); got != 3 {
		t.Errorf("Licia has %d +1/+1 counters, want 3", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, licia, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated a second time in one turn")
	}
}

// --- Magus of the Mind ---------------------------------------------------------

func TestC17B1MagusOfTheMindExilesOnePlusTheSpellsCastForFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C", "D", "E", "F")
	stormFiller(t, g)
	stormFiller(t, g)
	magus := b12Push(g, me.ID, "Magus of the Mind", "Creature — Human Wizard", c17MagusOfTheMindOracle, 4, 5)
	before := map[uuid.UUID]bool{}
	for _, c := range g.Exile.Cards {
		before[c.InstanceID] = true
	}
	b06AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, magus, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(magus) {
		t.Error("Magus of the Mind was not sacrificed")
	}
	free := 0
	for _, c := range g.Exile.Cards {
		if before[c.InstanceID] {
			continue
		}
		if perm := g.CastPermissionOnCardByIDForEffect(c.InstanceID); perm != nil && perm.Player == me.ID && perm.Cost == "{0}" {
			free++
		}
	}
	if free != 3 {
		t.Errorf("%d exiled cards may be played for free, want 3 (one plus two spells cast)", free)
	}
}

// --- Mirror of the Forebears -----------------------------------------------------

func TestC17B1MirrorOfTheForebearsCopiesACreatureOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	mirror := pushNamedTribePermanent(t, g, me.ID, "Mirror of the Forebears", "Artifact", c17MirrorOfTheForebearsOracl, "Cat")
	cat := dcSeed(g, me.ID, "Alley Cat", "Creature — Cat", "", 2, 2)
	bear := dcSeed(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	dcMana(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, mirror, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(bear)}); err == nil {
		t.Error("targeted a creature of another type")
	}
	if err := g.ActivateCatalogAbility(me.ID, mirror, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(cat)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := dcCard(t, g, mirror)
	if c.Name != "Alley Cat" || !c.IsCreature() || !c.IsArtifact() {
		t.Errorf("the Mirror is %q (creature %v, artifact %v), want an artifact copy of the Cat", c.Name, c.IsCreature(), c.IsArtifact())
	}
}

// --- Nazahn, Revered Bladesmith ----------------------------------------------------

func TestC17B1NazahnPutsHammerOfNazahnOntoTheBattlefield(t *testing.T) {
	for _, tc := range []struct {
		name          string
		onBattlefield bool
	}{{"Hammer of Nazahn", true}, {"Bonesplitter", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			gear := pushLibraryCardForTest(me, game.Card{Name: tc.name, TypeLine: "Artifact — Equipment", ManaCost: "{1}"})
			spell := pushLibraryCardForTest(me, game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: "{1}"})
			rfdCast(t, g, "Nazahn, Revered Bladesmith", c17NazahnType, c17NazahnOracle, 5, 4)
			c := searchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no search prompt")
			}
			if hasID(c.SearchCards, spell) {
				t.Error("a sorcery is offered")
			}
			answerSearchByID(t, g, me.ID, gear)
			passPriorityAroundTable(t, g)
			if got := g.Battlefield.Contains(gear); got != tc.onBattlefield {
				t.Errorf("on the battlefield %v, want %v", got, tc.onBattlefield)
			}
			if got := me.Hand.Contains(gear); got == tc.onBattlefield {
				t.Errorf("in hand %v, want %v", got, !tc.onBattlefield)
			}
		})
	}
}

func TestC17B1NazahnTapsACreatureWhenAnEquippedCreatureAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Nazahn, Revered Bladesmith", c17NazahnType, c17NazahnOracle, 5, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	c17Equipment(g, me.ID, "Sword", bear)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	ref := game.TargetRef{Kind: game.TargetCard, ID: theirs}
	assistCardSettle(t, g, me.ID, &ref, true)
	if !findBattlefieldCardForTest(g, theirs).Tapped {
		t.Error("the defending player's creature was not tapped")
	}
}

// --- Qasali Slingers ---------------------------------------------------------------

func TestC17B1QasaliSlingersDestroysAnArtifactWhenACatEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	slingers := b12Push(g, me.ID, "Qasali Slingers", "Creature — Cat Warrior", c17QasaliSlingersOracle, 3, 5)
	if !hasEffectiveKeyword(t, g, slingers, "reach") {
		t.Error("Qasali Slingers has no reach")
	}
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	castHoldCreature(t, g, "Alley Cat", "Creature — Cat", "", 1, 1)
	ref := game.TargetRef{Kind: game.TargetCard, ID: rock}
	assistCardSettle(t, g, me.ID, &ref, true)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived a Cat entering")
	}
}

func TestC17B1QasaliSlingersIgnoresANonCat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, me.ID, "Qasali Slingers", "Creature — Cat Warrior", c17QasaliSlingersOracle, 3, 5)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	castHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	if latestPickTarget(g, me.ID) != nil || !stackFullyEmpty(g) {
		t.Error("a Bear entering triggered Qasali Slingers")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact was destroyed")
	}
}

// --- Taigam, Sidisi's Hand -------------------------------------------------------

func TestC17B1TaigamLooksAtThreeAtUpkeepAndSkipsTheDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Taigam, Sidisi's Hand", "Legendary Creature — Human Wizard", c17TaigamOracle, false)
	advanceToUpkeepOfSeat(t, g, 1)
	top := seedLibrary(me, "Top", "Middle", "Third")
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick among the top three cards at upkeep")
	}
	if len(pick.ChooseCards) != 3 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly one of the top three", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(top[1]) || !me.Graveyard.Contains(top[0]) || !me.Graveyard.Contains(top[2]) {
		t.Error("the chosen card should be in hand and the other two in the graveyard")
	}
	hand := me.Hand.Size()
	toMain(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand went from %d to %d through the draw step, want no draw", hand, me.Hand.Size())
	}
}

func TestC17B1TaigamShrinksByTheCardsExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	taigam := b12Push(g, me.ID, "Taigam, Sidisi's Hand", "Legendary Creature — Human Wizard", c17TaigamOracle, 3, 4)
	one := wthGraveyardCard(me, "Old One", "Sorcery")
	two := wthGraveyardCard(me, "Old Two", "Creature — Bear")
	victim := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	b06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, taigam, 0, game.ActivateAbilityParams{
		Strict: true, ExileIDs: []uuid.UUID{one, two}, Targets: cardTarget(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(one) || !g.Exile.Contains(two) {
		t.Error("the two graveyard cards were not exiled")
	}
	if p, tough := effectivePTOf(t, g, victim); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1 after -2/-2", p, tough)
	}
}

// --- Vindictive Lich ---------------------------------------------------------------

func TestC17B1VindictiveLichModesTargetDifferentOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	n := len(g.Seats)
	opp, other := g.Seats[(g.Turn.ActiveSeat+1)%n], g.Seats[(g.Turn.ActiveSeat+2)%n]
	advanceToMain(t, g)
	lich := b12Push(g, me.ID, "Vindictive Lich", "Creature — Zombie Wizard", c17VindictiveLichOracle, 4, 1)
	oppLife, otherLife := opp.Life, other.Life
	b18Kill(t, g, lich)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the dies trigger did not ask for modes")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	if err := tmcB7TryPickPlayer(t, g, me.ID, opp.ID); err != nil {
		t.Fatalf("sacrifice bullet: %v", err)
	}
	if err := tmcB7TryPickPlayer(t, g, me.ID, opp.ID); err == nil {
		t.Fatal("the life-loss bullet accepted the opponent the sacrifice bullet targets")
	}
	if err := tmcB7TryPickPlayer(t, g, me.ID, other.ID); err != nil {
		t.Fatalf("life-loss bullet: %v", err)
	}
	passPriorityAroundTable(t, g)
	if other.Life != otherLife-5 {
		t.Errorf("the second opponent's life is %d, want %d", other.Life, otherLife-5)
	}
	if opp.Life != oppLife {
		t.Errorf("the first opponent's life moved to %d", opp.Life)
	}
}

// --- Wasitora, Nekoru Queen --------------------------------------------------------

func TestC17B1WasitoraMakesACatDragonWhenThePlayerCantSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	wasitora := b12Push(g, me.ID, "Wasitora, Nekoru Queen", "Legendary Creature — Cat Dragon", c17WasitoraOracle, 5, 4)
	for _, kw := range []string{"flying", "trample"} {
		if !hasEffectiveKeyword(t, g, wasitora, kw) {
			t.Errorf("Wasitora has no %s", kw)
		}
	}
	dealCombatDamageToPlayer(g, wasitora, victim.ID, 5)
	passPriorityAroundTable(t, g)
	dragons := c17Count(g, me.ID, func(c game.Card) bool {
		return c.Name == "Cat Dragon" && c.IsToken() && slices.Contains(c.Keywords, "flying")
	})
	if dragons != 1 {
		t.Errorf("you have %d Cat Dragon tokens, want 1", dragons)
	}
}

func TestC17B1WasitoraMakesThePlayerSacrificeACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	wasitora := b12Push(g, me.ID, "Wasitora, Nekoru Queen", "Legendary Creature — Cat Dragon", c17WasitoraOracle, 5, 4)
	bear := b12Creature(g, victim.ID, "Their Bear", "Creature — Bear", 2, 2)
	dealCombatDamageToPlayer(g, wasitora, victim.ID, 5)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, victim.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the damaged player did not sacrifice their creature")
	}
	if got := c17Count(g, me.ID, c17Named("Cat Dragon")); got != 0 {
		t.Errorf("you have %d Cat Dragons after a sacrifice, want 0", got)
	}
}
