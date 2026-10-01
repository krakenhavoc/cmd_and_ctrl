package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rooms_adjacent_cards_test.go — ADR 0103 PR 4: the cards that are
// about Rooms without being Rooms. (The eerie creatures are in
// eerie_cards_test.go.)

const (
	ghostlyDancersOracle           = "c53d958a-f660-4d2e-87cd-87702f973b3a"
	keysToTheHouseOracle           = "62320290-ac8e-4f92-bb06-3368f66ae0a9"
	marinaVendrellOracle           = "5aafc6c3-14f2-45b1-bd5b-c27760d791dd"
	ghostlyKeybearerOracle         = "ec86a45f-046c-4aab-9b82-422c6f39142e"
	creepingPeeperOracle           = "acf22340-7ff9-443f-a846-c44bb5a3f4e7"
	rampagingSoulragerOracle       = "f6d8e6cb-0878-4f2c-9a89-ce84cb4f5bbc"
	intrudingSoulragerOracle       = "7673d0db-07d6-4b40-a32e-f2c98d7ea7c1"
	adjacentRoomOracle             = "adjacent-test-room-oracle"
	adjacentRoomLeft, adjRoomRight = "Left Door", "Right Door"
)

// adjacentRoom puts an uncatalogued Room on seat 0's battlefield with
// the given doors unlocked.
func adjacentRoom(g *game.Game, mask game.DoorMask) uuid.UUID {
	return roomsDOnBattlefield(g, g.Seats[0].ID, adjacentRoomOracle, adjacentRoomLeft, adjRoomRight, mask)
}

func roomMask(t *testing.T, g *game.Game, id uuid.UUID) game.DoorMask {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Unlocked
}

// --- Ghostly Dancers ---------------------------------------------------

func TestGhostlyDancersEerieMakesAThreeOneFlyingSpirit(t *testing.T) {
	forEachEerieSource(t, func(t *testing.T, g *game.Game, fire func(eerieAnswers)) {
		eerieBody(g, "Ghostly Dancers", ghostlyDancersOracle, 2, 5, "flying")
		fire(eerieAnswers{})
		spirits := eerieTokens(g, g.Seats[0].ID, "Spirit")
		if len(spirits) != 1 || spirits[0].Power != 3 || spirits[0].Toughness != 1 ||
			spirits[0].Colors[0] != "W" || !contains(spirits[0].Keywords, "flying") {
			t.Fatalf("want one 3/1 white flying Spirit, got %+v", spirits)
		}
	})
}

func castGhostlyDancers(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Ghostly Dancers", "Creature — Spirit", ghostlyDancersOracle, nil)
}

func TestGhostlyDancersReturnsAnEnchantmentFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aura := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: aura, Name: "Old Aura", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID})
	castGhostlyDancers(t, g)
	eerieSettle(t, g, eerieAnswers{})
	if !inZone(me.Hand, aura) || inZone(me.Graveyard, aura) {
		t.Fatal("the enchantment card should be back in hand")
	}
}

func TestGhostlyDancersUnlocksADoorAndThatFullUnlockTriggersEerie(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	room := adjacentRoom(g, game.DoorLeftUnlocked)
	castGhostlyDancers(t, g)
	eerieSettle(t, g, eerieAnswers{})
	if !roomMask(t, g, room).Full() {
		t.Fatalf("the Room should be fully unlocked, mask %v", roomMask(t, g, room))
	}
	if n := len(eerieTokens(g, me.ID, "Spirit")); n != 1 {
		t.Fatalf("the Dancers' own eerie should make a Spirit off the full unlock, got %d", n)
	}
}

func TestGhostlyDancersOffersBothAndTheControllerChooses(t *testing.T) {
	for option, name := range []string{"return the card", "unlock the door"} {
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[0]
			aura := uuid.New()
			me.Graveyard.PushTop(game.Card{InstanceID: aura, Name: "Old Aura", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID})
			room := adjacentRoom(g, game.DoorLeftUnlocked)
			castGhostlyDancers(t, g)
			eerieSettle(t, g, eerieAnswers{Option: option})
			returned := inZone(me.Hand, aura)
			unlocked := roomMask(t, g, room).Full()
			if returned == unlocked || returned != (option == 0) {
				t.Fatalf("option %d: returned=%v unlocked=%v, want exactly the chosen one", option, returned, unlocked)
			}
		})
	}
}

func TestGhostlyDancersWithNothingToDoDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castGhostlyDancers(t, g)
	before := me.Hand.Size()
	eerieSettle(t, g, eerieAnswers{})
	if me.Hand.Size() != before {
		t.Fatal("nothing to return and no Room: the trigger should do nothing")
	}
}

// --- Keys to the House -------------------------------------------------

func pushKeys(g *game.Game) uuid.UUID {
	me := g.Seats[0]
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Keys to the House", TypeLine: "Artifact", ManaCost: "{1}",
		OracleID: keysToTheHouseOracle, Owner: me.ID, Controller: me.ID,
	})
}

func TestKeysToTheHouseFetchesABasicLandAndIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	keys := pushKeys(g)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	if err := g.ActivateCatalogAbility(me.ID, keys, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	eerieSettle(t, g, eerieAnswers{})
	if !inZone(me.Hand, ids[1]) {
		t.Fatal("the Forest should be in hand")
	}
	if onBattlefield(g, keys) {
		t.Fatal("the Keys are sacrificed as a cost")
	}
}

func TestKeysToTheHouseLocksOrUnlocksADoorOnlyAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	keys := pushKeys(g)
	room := adjacentRoom(g, game.DoorLeftUnlocked)

	// Unlock the locked right door.
	if err := g.ActivateCatalogAbility(me.ID, keys, 1, game.ActivateAbilityParams{Targets: cardRefs(room)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	eerieSettle(t, g, eerieAnswers{Option: 1}) // options: Lock Left Door, Unlock Right Door
	if !roomMask(t, g, room).Full() {
		t.Fatalf("the Room should be fully unlocked, mask %v", roomMask(t, g, room))
	}
	if onBattlefield(g, keys) {
		t.Fatal("the Keys are sacrificed as a cost")
	}

	// Lock an unlocked door with a second pair of Keys.
	keys2 := pushKeys(g)
	if err := g.ActivateCatalogAbility(me.ID, keys2, 1, game.ActivateAbilityParams{Targets: cardRefs(room)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	eerieSettle(t, g, eerieAnswers{Option: 0}) // Lock Left Door
	if m := roomMask(t, g, room); m != game.DoorRightUnlocked {
		t.Fatalf("mask %v, want only the right door unlocked", m)
	}

	// Not at instant speed.
	keys3 := pushKeys(g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, keys3, 1, game.ActivateAbilityParams{Targets: cardRefs(room)}); err == nil {
		t.Fatal("the Room ability is sorcery-speed only")
	}
}

func TestKeysToTheHouseCannotTargetAnOpponentsRoom(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	keys := pushKeys(g)
	theirs := roomsDOnBattlefield(g, opp.ID, adjacentRoomOracle, adjacentRoomLeft, adjRoomRight, game.DoorLeftUnlocked)
	err := g.ActivateCatalogAbility(me.ID, keys, 1, game.ActivateAbilityParams{Targets: cardRefs(theirs)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("err = %v, want ErrIllegalTarget", err)
	}
}

// --- Marina Vendrell ---------------------------------------------------

func TestMarinaVendrellTakesEnchantmentsFromTheTopSeven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cards := make([]game.Card, 0, 9)
	for i := 0; i < 9; i++ {
		c := game.Card{Name: "Filler", TypeLine: "Sorcery"}
		if i == 1 || i == 5 || i == 8 { // two inside the top seven, one below it
			c = game.Card{Name: "Aura", TypeLine: "Enchantment"}
		}
		cards = append(cards, c)
	}
	advanceToMain(t, g)
	// seedSearchLibrary pushes each card on the bottom, so cards[0] is the top card.
	ids := seedSearchLibrary(me, cards...)
	top := func(i int) uuid.UUID { return ids[i] }

	before := me.Hand.Size()
	castCatalogSpell(t, g, "Marina Vendrell", "Legendary Creature — Human Warlock", marinaVendrellOracle, nil)
	eerieSettle(t, g, eerieAnswers{})

	if !inZone(me.Hand, top(1)) || !inZone(me.Hand, top(5)) {
		t.Fatal("both enchantments among the top seven should be in hand")
	}
	if inZone(me.Hand, top(8)) {
		t.Fatal("the eighth-and-deeper enchantment was never revealed")
	}
	if me.Hand.Size() != before+2 {
		t.Fatalf("hand %d -> %d, want exactly the two enchantments added", before, me.Hand.Size())
	}
	if me.Library.Size() != 7 {
		t.Fatalf("library has %d cards, want 7", me.Library.Size())
	}
	// The five revealed non-enchantments went to the bottom, so the two
	// unrevealed cards are now on top.
	if c, _ := me.Library.Top(); c.InstanceID != top(7) {
		t.Fatalf("the top of the library should be the first unrevealed card")
	}
}

func TestMarinaVendrellLocksOrUnlocksADoor(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	marina := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Marina Vendrell", TypeLine: "Legendary Creature — Human Warlock",
		OracleID: marinaVendrellOracle, Power: 3, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	room := adjacentRoom(g, game.DoorLeftUnlocked)
	if err := g.ActivateCatalogAbility(me.ID, marina, 0, game.ActivateAbilityParams{Targets: cardRefs(room)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	eerieSettle(t, g, eerieAnswers{Option: 1}) // Unlock Right Door
	if !roomMask(t, g, room).Full() {
		t.Fatal("the Room should be fully unlocked")
	}
	c, _ := g.LookupCardForEffect(marina)
	if !c.Tapped {
		t.Fatal("Marina's ability taps her")
	}
}

// --- Ghostly Keybearer -------------------------------------------------

func TestGhostlyKeybearerUnlocksADoorOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	keybearer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ghostly Keybearer", TypeLine: "Creature — Spirit",
		OracleID: ghostlyKeybearerOracle, Power: 3, Toughness: 3, Keywords: []string{"flying"},
		Owner: me.ID, Controller: me.ID,
	})
	room := adjacentRoom(g, game.DoorLeftUnlocked)
	dealCombatDamageToPlayer(g, keybearer, opp.ID, 3)
	eerieSettle(t, g, eerieAnswers{})
	if !roomMask(t, g, room).Full() {
		t.Fatal("the locked door should be unlocked")
	}
}

func TestGhostlyKeybearerWithNoRoomDoesNothingAndNeverFails(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	keybearer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ghostly Keybearer", TypeLine: "Creature — Spirit",
		OracleID: ghostlyKeybearerOracle, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	dealCombatDamageToPlayer(g, keybearer, opp.ID, 3)
	eerieSettle(t, g, eerieAnswers{})
}

// --- Creeping Peeper ---------------------------------------------------

func peeperMana(t *testing.T, g *game.Game) {
	t.Helper()
	me := g.Seats[0]
	peeper := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Creeping Peeper", TypeLine: "Creature — Eye", ManaCost: "{1}{U}",
		OracleID: creepingPeeperOracle, Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateManaAbility(me.ID, peeper, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "U" {
		t.Fatalf("pool = %+v, want one {U}", me.ManaPool)
	}
}

func peeperCastFromHand(g *game.Game, me *game.Player, name, typeLine, cost string) (uuid.UUID, error) {
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
}

func TestCreepingPeeperManaPaysForAnEnchantmentSpell(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	peeperMana(t, g)
	if _, err := peeperCastFromHand(g, g.Seats[0], "Test Aura", "Enchantment", "{U}"); err != nil {
		t.Fatalf("an enchantment spell should be castable with the Peeper's mana: %v", err)
	}
}

func TestCreepingPeeperManaCannotPayForANonEnchantmentSpell(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	peeperMana(t, g)
	if _, err := peeperCastFromHand(g, g.Seats[0], "Test Bear", "Creature — Bear", "{U}"); err == nil {
		t.Fatal("a creature spell must not be castable with the Peeper's mana")
	}
}

func TestCreepingPeeperManaPaysToUnlockADoor(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	peeperMana(t, g)
	room := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), OracleID: adjacentRoomOracle, Owner: me.ID, Controller: me.ID,
		Name: adjacentRoomLeft, TypeLine: "Enchantment — Room", Layout: game.LayoutSplit, Unlocked: game.DoorLeftUnlocked,
		Faces: []game.Face{
			{Name: adjacentRoomLeft, TypeLine: "Enchantment — Room", ManaCost: "{2}{R}"},
			{Name: adjRoomRight, TypeLine: "Enchantment — Room", ManaCost: "{U}"},
		},
	})
	if err := g.PerformSpecialAction(me.ID, room, game.SpecialActionUnlock, game.SpecialActionParams{Door: game.DoorRight, Strict: true}); err != nil {
		t.Fatalf("unlocking with the Peeper's mana: %v", err)
	}
	if !roomMask(t, g, room).Full() {
		t.Fatal("the door should be unlocked")
	}
}

// --- Rampaging Soulrager -----------------------------------------------

func TestRampagingSoulragerGainsThreePowerAtTwoUnlockedDoors(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rampaging Soulrager", TypeLine: "Creature — Spirit",
		OracleID: rampagingSoulragerOracle, Power: 1, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	if got := b31CurrentPowerOf(g, id); got != 1 {
		t.Fatalf("no doors: power %d, want 1", got)
	}
	room := adjacentRoom(g, game.DoorLeftUnlocked)
	if got := b31CurrentPowerOf(g, id); got != 1 {
		t.Fatalf("one door: power %d, want 1", got)
	}
	if err := g.PerformSpecialAction(me.ID, room, game.SpecialActionUnlock, game.SpecialActionParams{Door: game.DoorRight}); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	if got := b31CurrentPowerOf(g, id); got != 4 {
		t.Fatalf("two doors: power %d, want 4", got)
	}
	g.WithWriteLock(func() {
		if err := g.LockDoorForEffect(room, game.DoorRight, me.ID); err != nil {
			t.Fatalf("lock: %v", err)
		}
	})
	if got := b31CurrentPowerOf(g, id); got != 1 {
		t.Fatalf("locked again: power %d, want 1", got)
	}
}

func TestRampagingSoulragerCountsDoorsAcrossRoomsYouControlOnly(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rampaging Soulrager", TypeLine: "Creature — Spirit",
		OracleID: rampagingSoulragerOracle, Power: 1, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	adjacentRoom(g, game.DoorLeftUnlocked)
	roomsDOnBattlefield(g, opp.ID, adjacentRoomOracle, adjacentRoomLeft, adjRoomRight, game.DoorLeftUnlocked|game.DoorRightUnlocked)
	if got := b31CurrentPowerOf(g, id); got != 1 {
		t.Fatalf("an opponent's doors must not count: power %d, want 1", got)
	}
	roomsDOnBattlefield(g, me.ID, adjacentRoomOracle, adjacentRoomLeft, adjRoomRight, game.DoorRightUnlocked)
	if got := b31CurrentPowerOf(g, id); got != 4 {
		t.Fatalf("one door on each of two Rooms of mine: power %d, want 4", got)
	}
}

// --- Intruding Soulrager -----------------------------------------------

func TestIntrudingSoulragerSacrificesARoomForDamageAndACard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	soulrager := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Intruding Soulrager", TypeLine: "Creature — Spirit",
		OracleID: intrudingSoulragerOracle, Power: 2, Toughness: 2, Keywords: []string{"vigilance"},
		Owner: me.ID, Controller: me.ID,
	})
	room := adjacentRoom(g, game.DoorLeftUnlocked)
	notARoom := pushTypedCard(g, me.ID, "Plain Aura", "Enchantment", "{1}")

	// Nothing to sacrifice but a non-Room: refused.
	if err := g.ActivateCatalogAbility(me.ID, soulrager, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{notARoom}}); err == nil {
		t.Fatal("only a Room may be sacrificed")
	}

	before := b30Lives(g)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, soulrager, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{room}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	eerieSettle(t, g, eerieAnswers{})
	if onBattlefield(g, room) {
		t.Fatal("the Room should have been sacrificed")
	}
	for i, p := range g.Seats {
		want := before[i]
		if i != 0 {
			want -= 2
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d", i, p.Life, want)
		}
	}
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
}
