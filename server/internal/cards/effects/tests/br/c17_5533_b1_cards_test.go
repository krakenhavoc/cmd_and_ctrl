package testsbr

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		effectstest.StdRegisteredFull(t, oracle, name)
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
	g := effectstest.NewCatalogGame(t)
	n := len(g.Seats)
	tb := c17CurseTable{
		g:        g,
		active:   g.Seats[g.Turn.ActiveSeat],
		victim:   g.Seats[(g.Turn.ActiveSeat+1)%n],
		curseCtl: g.Seats[(g.Turn.ActiveSeat+2)%n],
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: c17CurseType,
		OracleID: oracle, Owner: tb.curseCtl.ID, Controller: tb.curseCtl.ID,
		AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: tb.victim.ID},
	})
	tb.attackers = []uuid.UUID{
		effectstest.BlbB1Push(g, tb.active.ID, "Bear", "Creature — Bear", "", 2, 2),
		effectstest.BlbB1Push(g, tb.active.ID, "Bear 2", "Creature — Bear", "", 2, 2),
	}
	return tb
}

// attack declares both attackers at the cursed player and resolves the
// Curse's triggers.
func (tb c17CurseTable) Attack(t *testing.T) {
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
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestC17B1CurseOfVitalityGainsLifeForYouAndTheAttacker(t *testing.T) {
	tb := c17CurseSetup(t, "Curse of Vitality", c17CurseOfVitalityOracle)
	mine, theirs, victim := tb.curseCtl.Life, tb.active.Life, tb.victim.Life
	tb.Attack(t)
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
	tb.Attack(t)
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
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact",
		Owner: tb.curseCtl.ID, Controller: tb.curseCtl.ID, Tapped: true})
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Field", TypeLine: "Land",
		Owner: tb.active.ID, Controller: tb.active.ID, Tapped: true})
	victimRock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact",
		Owner: tb.victim.ID, Controller: tb.victim.ID, Tapped: true})
	tb.Attack(t)
	if effectstest.FindBattlefieldCardForTest(g, rock).Tapped {
		t.Error("the Curse's controller's artifact is still tapped")
	}
	for _, a := range tb.attackers {
		if effectstest.FindBattlefieldCardForTest(g, a).Tapped {
			t.Error("an attacking creature is still tapped: the attacker untaps their nonland permanents")
		}
	}
	if !effectstest.FindBattlefieldCardForTest(g, land).Tapped {
		t.Error("a land was untapped")
	}
	if !effectstest.FindBattlefieldCardForTest(g, victimRock).Tapped {
		t.Error("the enchanted player's artifact was untapped")
	}
}

// --- Balan, Wandering Knight ---------------------------------------------------

func c17Equipment(g *game.Game, owner uuid.UUID, name string, host uuid.UUID) uuid.UUID {
	c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Artifact — Equipment", Owner: owner, Controller: owner}
	if host != uuid.Nil {
		c.AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
	}
	return effectstest.PushBattlefieldCardWithTimestamp(g, c)
}

func TestC17B1BalanHasDoubleStrikeWithTwoEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	balan := effectstest.B12Push(g, me.ID, "Balan, Wandering Knight", "Legendary Creature — Cat Knight", c17BalanOracle, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, balan, "first strike") {
		t.Error("Balan has no first strike")
	}
	c17Equipment(g, me.ID, "Sword", balan)
	if effectstest.HasEffectiveKeyword(t, g, balan, "double strike") {
		t.Error("Balan has double strike with one Equipment attached")
	}
	c17Equipment(g, me.ID, "Shield", balan)
	if !effectstest.HasEffectiveKeyword(t, g, balan, "double strike") {
		t.Error("Balan has no double strike with two Equipment attached")
	}
}

func TestC17B1BalanAttachesAllYourEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	balan := effectstest.B12Push(g, me.ID, "Balan, Wandering Knight", "Legendary Creature — Cat Knight", c17BalanOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	onBear := c17Equipment(g, me.ID, "Sword", bear)
	loose := c17Equipment(g, me.ID, "Shield", uuid.Nil)
	theirs := c17Equipment(g, opp.ID, "Their Axe", uuid.Nil)
	effectstest.B06AddMana(me, "W", "C")
	if err := g.ActivateCatalogAbility(me.ID, balan, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{onBear, loose} {
		if !effectstest.FindBattlefieldCardForTest(g, id).IsAttachedTo(balan) {
			t.Error("one of your Equipment is not attached to Balan")
		}
	}
	if effectstest.FindBattlefieldCardForTest(g, theirs).IsAttachedTo(balan) {
		t.Error("an opponent's Equipment was attached to Balan")
	}
	if !effectstest.HasEffectiveKeyword(t, g, balan, "double strike") {
		t.Error("Balan has no double strike after taking two Equipment")
	}
}

// --- Boneyard Scourge ------------------------------------------------------------

func TestC17B1BoneyardScourgeReturnsWhenADragonDiesAndYouPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	scourge := effectstest.PushCatalogGraveyardCard(me, "Boneyard Scourge", "Creature — Zombie Dragon", c17BoneyardScourgeOracle, 4, 3)
	dragon := effectstest.B12Creature(g, me.ID, "Dragon", "Creature — Dragon", 4, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.B18Kill(t, g, dragon)
	effectstest.PassUntilTaxed(t, g, me.ID)
	effectstest.B06AddMana(me, "B", "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(scourge) || me.Graveyard.Contains(scourge) {
		t.Error("paying did not return Boneyard Scourge to the battlefield")
	}
}

func TestC17B1BoneyardScourgeIgnoresANonDragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	scourge := effectstest.PushCatalogGraveyardCard(me, "Boneyard Scourge", "Creature — Zombie Dragon", c17BoneyardScourgeOracle, 4, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B18Kill(t, g, bear)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a Bear dying offered to return Boneyard Scourge")
	}
	if !me.Graveyard.Contains(scourge) {
		t.Error("Boneyard Scourge left the graveyard")
	}
}

// --- Galecaster Colossus -----------------------------------------------------------

func TestC17B1GalecasterColossusBouncesByTappingAWizard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	colossus := effectstest.B12Push(g, me.ID, "Galecaster Colossus", "Creature — Giant Wizard", c17GalecasterColossusOracle, 5, 6)
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	if err := g.ActivateCatalogAbility(me.ID, colossus, 0, game.ActivateAbilityParams{
		Strict: true, TapIDs: []uuid.UUID{colossus}, Targets: effectstest.CardTarget(theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) {
		t.Error("the opponent's artifact did not return to its owner's hand")
	}
	if !effectstest.FindBattlefieldCardForTest(g, colossus).Tapped {
		t.Error("the Wizard that paid the cost is untapped")
	}
}

// --- Hungry Lynx ---------------------------------------------------------------

func TestC17B1HungryLynxProtectsYourCatsFromRats(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lynx := effectstest.B12Push(g, me.ID, "Hungry Lynx", "Creature — Cat", c17HungryLynxOracle, 2, 2)
	cat := effectstest.B12Creature(g, me.ID, "Alley Cat", "Creature — Cat", 1, 1)
	theirCat := effectstest.B12Creature(g, opp.ID, "Their Cat", "Creature — Cat", 1, 1)
	for _, id := range []uuid.UUID{lynx, cat} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "protection from Rats") {
			t.Error("a Cat you control has no protection from Rats")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, theirCat, "protection from Rats") {
		t.Error("an opponent's Cat has protection from Rats")
	}
}

func TestC17B1HungryLynxGrowsYourCatsWhenARatDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	lynx := effectstest.B12Push(g, me.ID, "Hungry Lynx", "Creature — Cat", c17HungryLynxOracle, 2, 2)
	cat := effectstest.B12Creature(g, me.ID, "Alley Cat", "Creature — Cat", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirCat := effectstest.B12Creature(g, opp.ID, "Their Cat", "Creature — Cat", 1, 1)
	rat := effectstest.B12Creature(g, opp.ID, "Rat", "Creature — Rat", 1, 1)
	effectstest.B18Kill(t, g, rat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{lynx, cat} {
		if got := effectstest.PlusOneCounters(g, id); got != 1 {
			t.Errorf("a Cat you control has %d +1/+1 counters, want 1", got)
		}
	}
	if got := effectstest.PlusOneCounters(g, bear); got != 0 {
		t.Errorf("a non-Cat has %d +1/+1 counters", got)
	}
	if got := effectstest.PlusOneCounters(g, theirCat); got != 0 {
		t.Errorf("an opponent's Cat has %d +1/+1 counters", got)
	}
}

func TestC17B1HungryLynxGivesATargetOpponentARatAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Hungry Lynx", "Creature — Cat", c17HungryLynxOracle, 2, 2)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	ref := effectstest.PlayerTarget(opp.ID)
	effectstest.AssistCardSettle(t, g, me.ID, &ref, true)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if got := effectstest.SelfPricedMV(t, g, me, c17LiciaOracle, c17LiciaType, "{5}{R}{W}{B}"); got != 8 {
		t.Fatalf("Licia before any life gain: %d, want 8", got)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	if got := effectstest.SelfPricedMV(t, g, me, c17LiciaOracle, c17LiciaType, "{5}{R}{W}{B}"); got != 5 {
		t.Errorf("Licia after gaining 3 life: %d, want 5", got)
	}
}

func TestC17B1LiciaPaysFiveLifeForThreeCountersOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	licia := effectstest.B12Push(g, me.ID, "Licia, Sanguine Tribune", c17LiciaType, c17LiciaOracle, 4, 4)
	for _, kw := range []string{"first strike", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, licia, kw) {
			t.Errorf("Licia has no %s", kw)
		}
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, licia, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-5 {
		t.Errorf("life %d, want %d", me.Life, life-5)
	}
	if got := effectstest.PlusOneCounters(g, licia); got != 3 {
		t.Errorf("Licia has %d +1/+1 counters, want 3", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, licia, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated a second time in one turn")
	}
}

// --- Magus of the Mind ---------------------------------------------------------

func TestC17B1MagusOfTheMindExilesOnePlusTheSpellsCastForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	effectstest.StormFiller(t, g)
	effectstest.StormFiller(t, g)
	magus := effectstest.B12Push(g, me.ID, "Magus of the Mind", "Creature — Human Wizard", c17MagusOfTheMindOracle, 4, 5)
	before := map[uuid.UUID]bool{}
	for _, c := range g.Exile.Cards {
		before[c.InstanceID] = true
	}
	effectstest.B06AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, magus, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	mirror := effectstest.PushNamedTribePermanent(t, g, me.ID, "Mirror of the Forebears", "Artifact", c17MirrorOfTheForebearsOracl, "Cat")
	cat := effectstest.DcSeed(g, me.ID, "Alley Cat", "Creature — Cat", "", 2, 2)
	bear := effectstest.DcSeed(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.DcMana(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, mirror, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Error("targeted a creature of another type")
	}
	if err := g.ActivateCatalogAbility(me.ID, mirror, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(cat)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DcCard(t, g, mirror)
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			gear := effectstest.PushLibraryCardForTest(me, game.Card{Name: tc.name, TypeLine: "Artifact — Equipment", ManaCost: "{1}"})
			spell := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: "{1}"})
			effectstest.RfdCast(t, g, "Nazahn, Revered Bladesmith", c17NazahnType, c17NazahnOracle, 5, 4)
			c := effectstest.SearchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no search prompt")
			}
			if effectstest.HasID(c.SearchCards, spell) {
				t.Error("a sorcery is offered")
			}
			effectstest.AnswerSearchByID(t, g, me.ID, gear)
			effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Nazahn, Revered Bladesmith", c17NazahnType, c17NazahnOracle, 5, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	c17Equipment(g, me.ID, "Sword", bear)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	ref := game.TargetRef{Kind: game.TargetCard, ID: theirs}
	effectstest.AssistCardSettle(t, g, me.ID, &ref, true)
	if !effectstest.FindBattlefieldCardForTest(g, theirs).Tapped {
		t.Error("the defending player's creature was not tapped")
	}
}

// --- Qasali Slingers ---------------------------------------------------------------

func TestC17B1QasaliSlingersDestroysAnArtifactWhenACatEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	slingers := effectstest.B12Push(g, me.ID, "Qasali Slingers", "Creature — Cat Warrior", c17QasaliSlingersOracle, 3, 5)
	if !effectstest.HasEffectiveKeyword(t, g, slingers, "reach") {
		t.Error("Qasali Slingers has no reach")
	}
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastHoldCreature(t, g, "Alley Cat", "Creature — Cat", "", 1, 1)
	ref := game.TargetRef{Kind: game.TargetCard, ID: rock}
	effectstest.AssistCardSettle(t, g, me.ID, &ref, true)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived a Cat entering")
	}
}

func TestC17B1QasaliSlingersIgnoresANonCat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Qasali Slingers", "Creature — Cat Warrior", c17QasaliSlingersOracle, 3, 5)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	if effectstest.LatestPickTarget(g, me.ID) != nil || !effectstest.StackFullyEmpty(g) {
		t.Error("a Bear entering triggered Qasali Slingers")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact was destroyed")
	}
}

// --- Taigam, Sidisi's Hand -------------------------------------------------------

func TestC17B1TaigamLooksAtThreeAtUpkeepAndSkipsTheDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Taigam, Sidisi's Hand", "Legendary Creature — Human Wizard", c17TaigamOracle, false)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	top := effectstest.SeedLibrary(me, "Top", "Middle", "Third")
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick among the top three cards at upkeep")
	}
	if len(pick.ChooseCards) != 3 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly one of the top three", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top[1]) || !me.Graveyard.Contains(top[0]) || !me.Graveyard.Contains(top[2]) {
		t.Error("the chosen card should be in hand and the other two in the graveyard")
	}
	hand := me.Hand.Size()
	effectstest.ToMain(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand went from %d to %d through the draw step, want no draw", hand, me.Hand.Size())
	}
}

func TestC17B1TaigamShrinksByTheCardsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	taigam := effectstest.B12Push(g, me.ID, "Taigam, Sidisi's Hand", "Legendary Creature — Human Wizard", c17TaigamOracle, 3, 4)
	one := effectstest.WthGraveyardCard(me, "Old One", "Sorcery")
	two := effectstest.WthGraveyardCard(me, "Old Two", "Creature — Bear")
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.B06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, taigam, 0, game.ActivateAbilityParams{
		Strict: true, ExileIDs: []uuid.UUID{one, two}, Targets: effectstest.CardTarget(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(one) || !g.Exile.Contains(two) {
		t.Error("the two graveyard cards were not exiled")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, victim); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1 after -2/-2", p, tough)
	}
}

// --- Vindictive Lich ---------------------------------------------------------------

func TestC17B1VindictiveLichModesTargetDifferentOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	n := len(g.Seats)
	opp, other := g.Seats[(g.Turn.ActiveSeat+1)%n], g.Seats[(g.Turn.ActiveSeat+2)%n]
	effectstest.AdvanceToMain(t, g)
	lich := effectstest.B12Push(g, me.ID, "Vindictive Lich", "Creature — Zombie Wizard", c17VindictiveLichOracle, 4, 1)
	oppLife, otherLife := opp.Life, other.Life
	effectstest.B18Kill(t, g, lich)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the dies trigger did not ask for modes")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	if err := effectstest.TmcB7TryPickPlayer(t, g, me.ID, opp.ID); err != nil {
		t.Fatalf("sacrifice bullet: %v", err)
	}
	if err := effectstest.TmcB7TryPickPlayer(t, g, me.ID, opp.ID); err == nil {
		t.Fatal("the life-loss bullet accepted the opponent the sacrifice bullet targets")
	}
	if err := effectstest.TmcB7TryPickPlayer(t, g, me.ID, other.ID); err != nil {
		t.Fatalf("life-loss bullet: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if other.Life != otherLife-5 {
		t.Errorf("the second opponent's life is %d, want %d", other.Life, otherLife-5)
	}
	if opp.Life != oppLife {
		t.Errorf("the first opponent's life moved to %d", opp.Life)
	}
}

// --- Wasitora, Nekoru Queen --------------------------------------------------------

func TestC17B1WasitoraMakesACatDragonWhenThePlayerCantSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	wasitora := effectstest.B12Push(g, me.ID, "Wasitora, Nekoru Queen", "Legendary Creature — Cat Dragon", c17WasitoraOracle, 5, 4)
	for _, kw := range []string{"flying", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, wasitora, kw) {
			t.Errorf("Wasitora has no %s", kw)
		}
	}
	effectstest.DealCombatDamageToPlayer(g, wasitora, victim.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	dragons := c17Count(g, me.ID, func(c game.Card) bool {
		return c.Name == "Cat Dragon" && c.IsToken() && slices.Contains(c.Keywords, "flying")
	})
	if dragons != 1 {
		t.Errorf("you have %d Cat Dragon tokens, want 1", dragons)
	}
}

func TestC17B1WasitoraMakesThePlayerSacrificeACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	wasitora := effectstest.B12Push(g, me.ID, "Wasitora, Nekoru Queen", "Legendary Creature — Cat Dragon", c17WasitoraOracle, 5, 4)
	bear := effectstest.B12Creature(g, victim.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DealCombatDamageToPlayer(g, wasitora, victim.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, victim.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the damaged player did not sacrifice their creature")
	}
	if got := c17Count(g, me.ID, c17Named("Cat Dragon")); got != 0 {
		t.Errorf("you have %d Cat Dragons after a sacrifice, want 0", got)
	}
}
