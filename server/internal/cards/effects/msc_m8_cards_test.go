package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m8_cards_test.go — Marvel Super Heroes Commander, M–Z batch 8 (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mscM8DetentionOracle = "ce95e51d-4ef3-44a8-ae47-cf1bf71ebd3f"
	mscM8RoboticsOracle  = "855318cb-9382-4183-87c7-93bdfac86431"
	mscM8HydraOracle     = "7274f6ed-310d-4b65-8083-eedb8e885a33"
	mscM8MinotaurOracle  = "f81207e8-d5a2-4a5e-8a81-803ac563fe76"
	mscM8GrovesOracle    = "3c87ea85-ca29-45a7-b5b2-758c62898b0a"
	mscM8PhotonOracle    = "2d0aa801-1227-4226-8c27-3bdd441857aa"
	mscM8SheHulkOracle   = "d62b8609-78a5-43ef-8e6e-dbfe8c7453ca"
	mscM8RedGhostOracle  = "9984d950-29dc-4d6b-8750-8d1bf86ea98f"
	mscM8ModokOracle     = "59f75a19-95cd-4ac8-94f1-27c1f915c34a"
	mscM8Mach1Oracle     = "75810ce7-328c-4512-8e36-fd3badbd3997"
	mscM8MethodsOracle   = "45f91d6a-eaab-4179-ac1f-5dc9e647a3da"
	mscM8ShuriOracle     = "1ea7b298-be87-4032-9f45-37f9b39ce3fa"
)

func TestMscM8EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mscM8DetentionOracle: "Secure Detention",
		mscM8RoboticsOracle:  "Robotics Mastery",
		mscM8HydraOracle:     "Madame Hydra, Reanimated",
		mscM8MinotaurOracle:  "Minotaur, Roxxon CEO",
		mscM8GrovesOracle:    "Scattered Groves",
		mscM8PhotonOracle:    "Photon, Lady of Light",
		mscM8SheHulkOracle:   "She-Hulk, Jennifer Walters",
		mscM8RedGhostOracle:  "Red Ghost, Intangible Genius",
		mscM8ModokOracle:     "M.O.D.O.K., Evil Intellect",
		mscM8Mach1Oracle:     "MACH-1, Swooping Scoundrel",
		mscM8MethodsOracle:   "Methods of the Mighty",
		mscM8ShuriOracle:     "Shuri, the Black Panther",
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

// --- Secure Detention ------------------------------------------------------------

const mscM8DetentionBundle = game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana

func TestMscM8SecureDetentionLocksACreatureAndMakesASoldier(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	feeder := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Carrion Feeder", TypeLine: "Creature — Zombie",
		Power: 1, Toughness: 1, OracleID: carrionFeederOracle, Owner: me.ID, Controller: me.ID,
	})

	aura := castCatalogSpell(t, g, "Secure Detention", "Enchantment — Aura", mscM8DetentionOracle, b16TargetCard(feeder))
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)

	assertRestrictions(t, g, feeder, mscM8DetentionBundle)
	if err := g.ActivateCatalogAbility(me.ID, feeder, 0, game.ActivateAbilityParams{}); err != game.ErrCantActivate {
		t.Errorf("a detained sac outlet activated: %v", err)
	}
	if got := mshB1Tokens(g, me.ID, "Soldier"); got != 1 {
		t.Fatalf("%d Soldier tokens, want 1", got)
	}
	soldier := mshB1Token(t, g, me.ID, "Soldier")
	if soldier.Power != 1 || soldier.Toughness != 1 || len(soldier.Colors) != 1 || soldier.Colors[0] != "W" {
		t.Errorf("Soldier is %d/%d %v, want a 1/1 white token", soldier.Power, soldier.Toughness, soldier.Colors)
	}

	// Losing the Aura gives the creature back; the Soldier stays.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(aura); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if got := restrictionsOf(t, g, feeder); got&mscM8DetentionBundle&^game.CantAttackOrBlock != 0 {
		t.Errorf("the activation lock outlived the Aura: %v", got.Names())
	}
	if got := mshB1Tokens(g, me.ID, "Soldier"); got != 1 {
		t.Errorf("%d Soldier tokens after the Aura left, want 1", got)
	}
}

func TestMscM8SecureDetentionEnchantsAnArtifactButNotALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := b12Push(g, me.ID, "Mana Rock", "Artifact", "", 0, 0)
	land := b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	advanceToMain(t, g)

	if err := castCatalogSpellErr(t, g, "Secure Detention", "Enchantment — Aura", mscM8DetentionOracle, b16TargetCard(land)); err == nil {
		t.Fatal("Secure Detention was cast on a land")
	}
	castCatalogSpell(t, g, "Secure Detention", "Enchantment — Aura", mscM8DetentionOracle, b16TargetCard(rock))
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, rock, mscM8DetentionBundle)
}

// --- Robotics Mastery ------------------------------------------------------------

func TestMscM8RoboticsMasteryHasFlashPumpsAndMakesTwoFlyingRobots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)

	aura := handCardFull(me, "Robotics Mastery", "Enchantment — Aura", "{4}{U}", mscM8RoboticsOracle, []string{"U"})
	floatMana(t, g, me, "{U}")
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{Strict: true, Targets: b16TargetCard(bear)}); err == nil {
		t.Fatal("Robotics Mastery was cast for one mana")
	}
	me.ManaPool = nil
	floatMana(t, g, me, "{U}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{Strict: true, Targets: b16TargetCard(bear)}); err != nil {
		t.Fatalf("flash Robotics Mastery at the beginning of combat: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)

	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("enchanted power %d, want 4", got)
	}
	if got := mshB1Tokens(g, me.ID, "Robot"); got != 2 {
		t.Fatalf("%d Robot tokens, want 2", got)
	}
	robot := mshB1Token(t, g, me.ID, "Robot")
	if robot.Power != 1 || robot.Toughness != 1 || !robot.IsArtifact() || len(robot.Colors) != 0 {
		t.Errorf("Robot is %q %d/%d %v, want a colorless 1/1 artifact creature", robot.TypeLine, robot.Power, robot.Toughness, robot.Colors)
	}
	if !hasEffectiveKeyword(t, g, robot.InstanceID, "flying") {
		t.Error("a Robot token lacks flying")
	}

	// The Robots are the Aura's trigger, not the host's: they outlive it.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(aura); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if got := mshB1Tokens(g, me.ID, "Robot"); got != 2 {
		t.Errorf("%d Robots after the Aura left, want 2", got)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("power %d after the Aura left, want 2", got)
	}
}

func TestMscM8SecureDetentionIsNotFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	aura := handCardFull(me, "Secure Detention", "Enchantment — Aura", "{3}{W}", mscM8DetentionOracle, []string{"W"})
	floatMana(t, g, me, "{W}{C}{C}{C}")
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{Strict: true, Targets: b16TargetCard(bear)}); err == nil {
		t.Error("a non-flash Aura was cast at the beginning of combat")
	}
}

// --- Madame Hydra, Reanimated -----------------------------------------------------

func TestMscM8MadameHydraGrowsWithTheGraveyardAndMillsOnEntryAndAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	me.Library.Cards = nil
	plTop(me, "Mill Bear", "Creature — Bear", "")
	plTop(me, "Mill Wolf", "Creature — Wolf", "")

	hydra := castHoldCreature(t, g, "Madame Hydra, Reanimated", "Legendary Creature — Human Horror Villain", mscM8HydraOracle, 2, 2)
	assertKeywords(t, g, hydra, "menace")
	if got := me.Graveyard.Size(); got != 2 {
		t.Fatalf("%d cards milled on entry, want 2", got)
	}
	if got := hobA4Power(t, g, hydra); got != 4 {
		t.Errorf("power %d with two creature cards in the graveyard, want 4", got)
	}

	battlefieldCardPtr(t, g, hydra).SummonedThisTurn = false
	plTop(me, "Milled Land", "Basic Land — Forest", "")
	plTop(me, "Mill Ape", "Creature — Ape", "")
	attackWith(t, g, opp.ID, hydra)
	passPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != 4 {
		t.Errorf("%d cards in the graveyard after the attack mill, want 4", got)
	}
	if got := hobA4Power(t, g, hydra); got != 5 {
		t.Errorf("power %d with three creature cards (the land does not count), want 5", got)
	}
}

// --- Minotaur, Roxxon CEO ---------------------------------------------------------

func TestMscM8MinotaurMakesAVillainWhenAnyNontokenCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	minotaur := b12Push(g, me.ID, "Minotaur, Roxxon CEO", "Legendary Creature — Minotaur Villain", mscM8MinotaurOracle, 4, 4)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if got := mshB1Tokens(g, me.ID, "Villain"); got != 1 {
		t.Fatalf("%d Villain tokens after an opposing creature died, want 1", got)
	}
	if got := mshB1Tokens(g, opp.ID, "Villain"); got != 0 {
		t.Errorf("the opponent got %d Villains, want 0 (you create the token)", got)
	}
	villain := mshB1Token(t, g, me.ID, "Villain")
	if villain.Power != 2 || villain.Toughness != 1 || len(villain.Colors) != 1 || villain.Colors[0] != "B" {
		t.Errorf("Villain is %d/%d %v, want a 2/1 black token", villain.Power, villain.Toughness, villain.Colors)
	}
	assertKeywords(t, g, villain.InstanceID, "menace")

	// A token dying makes nothing.
	killCreature(t, g, me.ID, villain.InstanceID)
	passPriorityAroundTable(t, g)
	if got := mshB1Tokens(g, me.ID, "Villain"); got != 0 {
		t.Errorf("%d Villains after a token died, want 0", got)
	}

	// Minotaur's own death counts.
	killCreature(t, g, me.ID, minotaur)
	passPriorityAroundTable(t, g)
	if got := mshB1Tokens(g, me.ID, "Villain"); got != 1 {
		t.Errorf("%d Villains after Minotaur died, want 1", got)
	}
}

// --- Scattered Groves ---------------------------------------------------------------

func TestMscM8ScatteredGrovesEntersTappedTapsForGreenOrWhiteAndCycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := playLandFromHand(t, g, "Scattered Groves", mscM8GrovesOracle)
	card, ok := battlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatal("Scattered Groves must enter tapped")
	}
	for _, color := range []string{"G", "W"} {
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
		me.ManaPool = nil
		b28TapForMana(t, g, me.ID, id, color)
		if got := poolColors(me); len(got) != 1 || got[0] != color {
			t.Errorf("pool = %v, want {%s}", got, color)
		}
	}
}

func TestMscM8ScatteredGrovesCycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := me.Hand.Size()
	id, _ := cycleFromHand(t, g, "Scattered Groves", "Land — Forest Plains", mscM8GrovesOracle, "{C}{C}")
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand size %d after cycling, want %d (discard one, draw one)", got, hand)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled land is not in the graveyard")
	}
}

// --- Photon, Lady of Light --------------------------------------------------------

func TestMscM8PhotonBlinksAnotherCreatureAndTakesItOutOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	photon := b12Push(g, me.ID, "Photon, Lady of Light", "Legendary Creature — Human Hero", mscM8PhotonOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Blinked Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounterBy(me.ID, bear, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}

	declareAttack(t, g, opp.ID, photon, bear)
	// "Other": Photon is not an eligible target for her own trigger.
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt for Photon's attack trigger")
	}
	if err := g.ResolvePickTarget(prompt.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: photon}); err == nil {
		t.Fatal("Photon accepted herself as the other creature")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Error("the Bear kept its old identity: it was not exiled and returned")
	}
	back := findBattlefieldByName(g, "Blinked Bear")
	if back == uuid.Nil {
		t.Fatal("the Bear did not return to the battlefield")
	}
	if got := plusOnes(g, back); got != 0 {
		t.Errorf("the returned Bear has %d +1/+1 counters, want 0 (a new object)", got)
	}
	if c := battlefieldCardPtr(t, g, back); c.AttackingTarget != uuid.Nil {
		t.Error("the returned Bear is still attacking")
	}
	if !g.Battlefield.Contains(photon) {
		t.Fatal("Photon left the battlefield")
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != 40-3 {
		t.Errorf("opponent life %d, want 37 (only Photon still attacks)", opp.Life)
	}
}

func TestMscM8PhotonBlinksNothingWhenNoTargetIsChosen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	photon := b12Push(g, me.ID, "Photon, Lady of Light", "Legendary Creature — Human Hero", mscM8PhotonOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Kept Bear", "Creature — Bear", 2, 2)
	assertKeywords(t, g, photon, "flying")

	declareAttack(t, g, opp.ID, photon)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt for Photon's attack trigger")
	}
	if err := g.ResolvePickTargets(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolvePickTargets with no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear was blinked although no target was chosen")
	}
}

// --- She-Hulk, Jennifer Walters ---------------------------------------------------

func mscM8SheHulkBoard(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	hulk := b12Push(g, me.ID, "She-Hulk, Jennifer Walters", "Legendary Creature — Gamma Berserker Hero", mscM8SheHulkOracle, 4, 4)
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	return g, me, hulk, land
}

func TestMscM8SheHulkSacrificesALandToDrawAndGrow(t *testing.T) {
	g, me, hulk, land := mscM8SheHulkBoard(t)
	assertKeywords(t, g, hulk, "trample")
	hand := me.Hand.Size()

	floatMana(t, g, me, "{R}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, hulk, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{land},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	if got := hobA4Power(t, g, hulk); got != 5 {
		t.Errorf("power %d, want 5 (one +1/+1 counter)", got)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left floating, want 0", me.ManaPool.Total())
	}
}

func TestMscM8SheHulkNeedsTheManaAndALand(t *testing.T) {
	g, me, hulk, land := mscM8SheHulkBoard(t)

	floatMana(t, g, me, "{R}{C}")
	if err := g.ActivateCatalogAbility(me.ID, hulk, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{land},
	}); err == nil {
		t.Error("She-Hulk's ability was activated for two mana")
	}
	me.ManaPool = nil
	floatMana(t, g, me, "{R}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, hulk, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("She-Hulk's ability was activated without sacrificing a land")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a refused activation still took the land")
	}
}

func TestMscM8SheHulkStillDrawsWhenSheLeavesBeforeItResolves(t *testing.T) {
	g, me, hulk, land := mscM8SheHulkBoard(t)
	hand := me.Hand.Size()

	floatMana(t, g, me, "{R}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, hulk, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{land},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	killCreature(t, g, me.ID, hulk)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1 (the draw does not need She-Hulk)", got)
	}
	if g.Battlefield.Contains(hulk) {
		t.Error("She-Hulk came back")
	}
}

// --- Red Ghost, Intangible Genius -------------------------------------------------

func TestMscM8RedGhostMakesAnApeOnlyOnTheSecondDrawOfATurn(t *testing.T) {
	g := newCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Red Ghost, Intangible Genius", "Legendary Creature — Human Scientist Villain", mscM8RedGhostOracle, false)

	hobA5Draw(t, g, other, nil)
	hobA5Draw(t, g, other, nil)
	if got := mshB1Tokens(g, me.ID, "Ape"); got != 0 {
		t.Fatalf("%d Apes after another player drew twice, want 0", got)
	}
	hobA5Draw(t, g, me, nil)
	if got := mshB1Tokens(g, me.ID, "Ape"); got != 0 {
		t.Fatalf("%d Apes after the first draw, want 0", got)
	}
	hobA5Draw(t, g, me, nil)
	if got := mshB1Tokens(g, me.ID, "Ape"); got != 1 {
		t.Fatalf("%d Apes after the second draw, want 1", got)
	}
	hobA5Draw(t, g, me, nil)
	if got := mshB1Tokens(g, me.ID, "Ape"); got != 1 {
		t.Errorf("%d Apes after the third draw, want still 1", got)
	}
	ape := mshB1Token(t, g, me.ID, "Ape")
	if ape.Power != 3 || ape.Toughness != 3 || len(ape.Colors) != 1 || ape.Colors[0] != "R" || !mshB1HasType(ape, "Villain") {
		t.Errorf("Ape is %q %d/%d %v, want a red 3/3 Ape Villain", ape.TypeLine, ape.Power, ape.Toughness, ape.Colors)
	}
	assertKeywords(t, g, ape.InstanceID, "haste")
}

func TestMscM8RedGhostCannotBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	ghost := b12Push(g, me.ID, "Red Ghost, Intangible Genius", "Legendary Creature — Human Scientist Villain", mscM8RedGhostOracle, 2, 3)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ghost, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(plain, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, ghost); err == nil {
		t.Error("Red Ghost was blocked")
	}
	if err := g.DeclareBlocker(wall, plain); err != nil {
		t.Errorf("an ordinary attacker could not be blocked either: %v", err)
	}
}

func TestMscM8RedGhostHasWardTwo(t *testing.T) {
	g, _, opp, warded := wardTable(t, "Red Ghost, Intangible Genius", "Legendary Creature — Human Scientist Villain", mscM8RedGhostOracle)
	blade := castAtWardedCreature(t, g, opp, warded)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(warded) {
		t.Error("Red Ghost died to a spell whose controller declined the ward {2}")
	}
	if !opp.Graveyard.Contains(blade) {
		t.Error("the countered spell is not in its owner's graveyard")
	}
}

// --- M.O.D.O.K., Evil Intellect -----------------------------------------------------

func mscM8ModokTable(t *testing.T) (*game.Game, *game.Player, *game.Player, game.TargetRef) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "M.O.D.O.K., Evil Intellect", "Legendary Artifact Creature — Villain", mscM8ModokOracle, false)
	return g, me, opp, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
}

func TestMscM8ModokMakesAnOpponentSacrificeANontokenCreatureOnTheSecondDraw(t *testing.T) {
	g, me, opp, pick := mscM8ModokTable(t)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.CreateTokenForEffect(opp.ID, TokenCard("1/1 white Soldier"), 1); err != nil {
		t.Fatal(err)
	}

	hobA5Draw(t, g, me, &pick)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the first draw made the opponent sacrifice")
	}
	hobA5Draw(t, g, me, &pick)
	answerSacrifice(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the opponent kept their only nontoken creature")
	}
	if got := mshB1Tokens(g, opp.ID, "Soldier"); got != 1 {
		t.Errorf("the opponent has %d Soldier tokens, want 1 (a token is never the sacrifice)", got)
	}
}

func TestMscM8ModokLetsTheOpponentChooseAndSkipsTokenOnlyBoards(t *testing.T) {
	g, me, opp, pick := mscM8ModokTable(t)
	keep := b12Creature(g, opp.ID, "Kept Bear", "Creature — Bear", 2, 2)
	lose := b12Creature(g, opp.ID, "Lost Bear", "Creature — Bear", 2, 2)

	hobA5Draw(t, g, me, &pick)
	hobA5Draw(t, g, me, &pick)
	answerSacrifice(t, g, opp.ID, lose)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lose) || !g.Battlefield.Contains(keep) {
		t.Errorf("kept %v lost %v, want the opponent's choice to go", g.Battlefield.Contains(keep), g.Battlefield.Contains(lose))
	}

	// A third draw does not trigger, and a token-only board loses nothing.
	g2, me2, opp2, pick2 := mscM8ModokTable(t)
	if err := g2.CreateTokenForEffect(opp2.ID, TokenCard("1/1 white Soldier"), 1); err != nil {
		t.Fatal(err)
	}
	hobA5Draw(t, g2, me2, &pick2)
	hobA5Draw(t, g2, me2, &pick2)
	passPriorityAroundTable(t, g2)
	if got := mshB1Tokens(g2, opp2.ID, "Soldier"); got != 1 {
		t.Errorf("a token was sacrificed to M.O.D.O.K. (%d left)", got)
	}
}

// --- MACH-1, Swooping Scoundrel -------------------------------------------------------

func mscM8KeepSurveil(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	ch := surveilChoiceFor(g, p.ID)
	if ch == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(ch.ID, p.ID, ch.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
}

func mscM8GainLife(g *game.Game, source uuid.UUID, p *game.Player, n int) {
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(source, p.ID, n) })
}

func TestMscM8Mach1SurveilsOnEntryAndThenNotAgainOnALifeGainThatTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C", "D", "E", "F")

	mach := castHoldCreature(t, g, "MACH-1, Swooping Scoundrel", "Legendary Artifact Creature — Human Villain", mscM8Mach1Oracle, 1, 3)
	assertKeywords(t, g, mach, "flying")
	mscM8KeepSurveil(t, g, me)
	passPriorityAroundTable(t, g)

	mscM8GainLife(g, mach, me, 2)
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) != nil {
		t.Fatal("a life gain after the enters trigger surveilled again in the same turn")
	}
}

func TestMscM8Mach1SurveilsOnTheFirstLifeGainEachTurnOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C", "D", "E", "F")
	mach := b12Push(g, me.ID, "MACH-1, Swooping Scoundrel", "Legendary Artifact Creature — Human Villain", mscM8Mach1Oracle, 1, 3)

	mscM8GainLife(g, mach, opp, 3)
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) != nil {
		t.Fatal("an opponent's life gain surveilled")
	}
	mscM8GainLife(g, mach, me, 2)
	passPriorityAroundTable(t, g)
	mscM8KeepSurveil(t, g, me)
	passPriorityAroundTable(t, g)

	mscM8GainLife(g, mach, me, 2)
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) != nil {
		t.Fatal("a second life gain in one turn surveilled again")
	}

	advanceToUpkeepOf(t, g, 1)
	mscM8GainLife(g, mach, me, 1)
	passPriorityAroundTable(t, g)
	mscM8KeepSurveil(t, g, me)
}

// --- Methods of the Mighty ----------------------------------------------------------

func mscM8MethodsBoard(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	rock := b12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	tapped := b12Creature(g, opp.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	untapped := b12Creature(g, opp.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	battlefieldCardPtr(t, g, tapped).Tapped = true
	return g, me, rock, tapped, untapped, mine
}

func TestMscM8MethodsOfTheMightyDoesAllThreeThings(t *testing.T) {
	g, me, rock, tapped, untapped, mine := mscM8MethodsBoard(t)
	other := b12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)

	castModal(t, g, "Methods of the Mighty", "Instant", mscM8MethodsOracle, []int{0, 1, 2}, []game.TargetRef{
		modeRef(game.TargetCard, rock, 0, 0),
		modeRef(game.TargetCard, tapped, 1, 0),
	})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature survived")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature was destroyed")
	}
	if plusOnes(g, mine) != 1 || plusOnes(g, other) != 1 {
		t.Errorf("counters on my creatures: %d and %d, want 1 and 1", plusOnes(g, mine), plusOnes(g, other))
	}
	if plusOnes(g, untapped) != 0 {
		t.Error("an opposing creature got a counter")
	}
}

func TestMscM8MethodsOfTheMightyCounterModeAloneNeedsNoTarget(t *testing.T) {
	g, _, rock, tapped, _, mine := mscM8MethodsBoard(t)
	castModal(t, g, "Methods of the Mighty", "Instant", mscM8MethodsOracle, []int{2}, nil)
	passPriorityAroundTable(t, g)
	if plusOnes(g, mine) != 1 {
		t.Errorf("%d counters on my creature, want 1", plusOnes(g, mine))
	}
	if !g.Battlefield.Contains(rock) || !g.Battlefield.Contains(tapped) {
		t.Error("an unchosen bullet destroyed something")
	}
}

func TestMscM8MethodsOfTheMightyRefusesIllegalTargets(t *testing.T) {
	g, me, rock, _, untapped, mine := mscM8MethodsBoard(t)
	active := g.Seats[g.Turn.ActiveSeat]
	if active.ID != me.ID {
		t.Fatal("the test expects seat 0 to be active")
	}
	try := func(modes []int, targets []game.TargetRef) error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Methods of the Mighty", TypeLine: "Instant",
			OracleID: mscM8MethodsOracle, Owner: me.ID, Controller: me.ID})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
	}
	if err := try([]int{1}, []game.TargetRef{modeRef(game.TargetCard, untapped, 0, 0)}); err == nil {
		t.Error("the destroy-tapped-creature bullet accepted an untapped creature")
	}
	if err := try([]int{0}, []game.TargetRef{modeRef(game.TargetCard, mine, 0, 0)}); err == nil {
		t.Error("the destroy-artifact bullet accepted a creature")
	}
	if err := try([]int{1}, []game.TargetRef{modeRef(game.TargetCard, rock, 0, 0)}); err == nil {
		t.Error("the destroy-tapped-creature bullet accepted an artifact")
	}
	if err := try(nil, nil); err == nil {
		t.Error("the spell was cast with no mode chosen")
	}
}

// --- Shuri, the Black Panther ----------------------------------------------------------

func mscM8ShuriAttack(t *testing.T, artifacts int) (hand int, shuriPower, bearPower int) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	shuri := b12Push(g, me.ID, "Shuri, the Black Panther", "Legendary Creature — Human Noble Hero", mscM8ShuriOracle, 2, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	assertKeywords(t, g, shuri, "lifelink")
	for i := 0; i < artifacts; i++ {
		b12Permanent(g, me.ID, "Gadget", "Artifact")
	}
	before := me.Hand.Size()
	declareAttack(t, g, opp.ID, shuri)
	passPriorityAroundTable(t, g)
	return me.Hand.Size() - before, effectivePower(t, g, shuri), effectivePower(t, g, bear)
}

func TestMscM8ShuriDrawsWithThreeArtifactsAndPumpsWithSix(t *testing.T) {
	for _, tc := range []struct {
		artifacts, draws, shuriPower, bearPower int
	}{
		{0, 0, 2, 2},
		{2, 0, 2, 2},
		{3, 1, 2, 2},
		{5, 1, 2, 2},
		{6, 1, 4, 4},
		{8, 1, 4, 4},
	} {
		draws, shuriPower, bearPower := mscM8ShuriAttack(t, tc.artifacts)
		if draws != tc.draws || shuriPower != tc.shuriPower || bearPower != tc.bearPower {
			t.Errorf("%d artifacts: drew %d, Shuri power %d, Bear power %d; want %d, %d, %d",
				tc.artifacts, draws, shuriPower, bearPower, tc.draws, tc.shuriPower, tc.bearPower)
		}
	}
}
