package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hob_a8_cards_test.go — The Hobbit / The Hobbit Eternal, half A, eighth
// batch (#4454): Thorin, Aragorn and Arwen and the Necklace of Girion, then
// the plain cards from the NEXT list. Each test plays the real catalog
// entry and asserts what its text says.

const (
	hobA8ThorinOracle     = "7926b4c1-9a93-4193-ae6f-e75bbcd762ec"
	hobA8AragornOracle    = "b8a3440a-03c5-444c-a1de-252524c3db82"
	hobA8NecklaceOracle   = "9f970385-e7c6-40bf-999e-7015693c2250"
	hobA8HunterOracle     = "fdf7f144-56e4-4f88-b81a-b85473922355"
	hobA8BatCloudOracle   = "3cc64ce6-45e2-4dc3-9b28-d7206a131873"
	hobA8RollOracle       = "aa913409-7a7a-4e74-82ff-7ff8195585ad"
	hobA8GandalfOracle    = "6760d7d1-098e-4131-8824-53e5427d047b"
	hobA8BolgOracle       = "eab096fd-a7e2-4b96-99ec-1039d3a552f3"
	hobA8NotaryOracle     = "dc20b29b-856e-4d45-99cd-5e12cfc2324d"
	hobA8BlacksmithOracle = "9febdfa8-8af2-4aba-b2e6-cb2baad7e02b"
)

func TestHobA8EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		hobA8ThorinOracle:   "Thorin, Company's Leader",
		hobA8AragornOracle:  "Aragorn and Arwen, Wed",
		hobA8NecklaceOracle: "Necklace of Girion",
		hobA8HunterOracle:   "Snowslope Hunter",
		hobA8BatCloudOracle: "Dreaded Bat-Cloud",
		hobA8RollOracle:     "Roll-Roll-Roll-Roll",
		hobA8GandalfOracle:  "Gandalf, Shadow's Foe",
		hobA8BolgOracle:     "Bolg, Erebor's Reckoning",
		hobA8NotaryOracle:   "The Notary Hobbits",

		hobA8BlacksmithOracle: "Iron Hills Blacksmith",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
	}
}

func hobA8Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return ringBFCard(t, g, id).Counters[game.CounterPlusOne]
}

// --- Thorin, Company's Leader --------------------------------------------

const hobA8DwarfTypeLine = "Creature — Dwarf Warrior"

func hobA8Thorin(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := pushCatalogPermanent(g, owner, "Thorin, Company's Leader", "Legendary Creature — Dwarf Warrior", hobA8ThorinOracle, false)
	return id
}

func TestHobA8ThorinMakesTwoTreasuresPerDwarfThatConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	hobA8Thorin(g, me.ID)
	dwarf := b12Creature(g, me.ID, "Dwarf A", hobA8DwarfTypeLine, 2, 2)
	other := b12Creature(g, me.ID, "Dwarf B", hobA8DwarfTypeLine, 2, 2)

	before := countTreasuresControlledBy(g, me.ID)
	dealCombatDamageToPlayer(g, dwarf, opp.ID, 2)
	dealCombatDamageToPlayer(g, other, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if got := countTreasuresControlledBy(g, me.ID) - before; got != 4 {
		t.Errorf("two Dwarves connected and made %d Treasures, want 4 (two each, not once per batch)", got)
	}
}

func TestHobA8ThorinCountsHimself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	thorin := hobA8Thorin(g, me.ID)

	before := countTreasuresControlledBy(g, me.ID)
	dealCombatDamageToPlayer(g, thorin, opp.ID, 4)
	passPriorityAroundTable(t, g)
	if got := countTreasuresControlledBy(g, me.ID) - before; got != 2 {
		t.Errorf("Thorin connected and made %d Treasures, want 2", got)
	}
}

func TestHobA8ThorinIgnoresNonDwarvesOpposingDwarvesAndOtherDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	hobA8Thorin(g, me.ID)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf Warrior", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Dwarf", hobA8DwarfTypeLine, 2, 2)
	dwarf := b12Creature(g, me.ID, "Dwarf", hobA8DwarfTypeLine, 2, 2)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 5, 5)

	before := countTreasuresControlledBy(g, me.ID)
	dealCombatDamageToPlayer(g, elf, opp.ID, 2)
	dealCombatDamageToPlayer(g, theirs, me.ID, 2)
	// A Dwarf's damage to a creature, and noncombat damage to a player.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: dwarf, Target: victim, Amount: 2, Combat: true})
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: dwarf, Target: opp.ID, Amount: 2})
	})
	passPriorityAroundTable(t, g)
	if got := countTreasuresControlledBy(g, me.ID) - before; got != 0 {
		t.Errorf("made %d Treasures, want 0", got)
	}
	if got := countTreasuresControlledBy(g, opp.ID); got != 0 {
		t.Errorf("the opponent's Dwarf made %d Treasures for them, want 0", got)
	}
}

func TestHobA8ThorinTriggersOnCombatDamageToABattle(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	hobA8Thorin(g, me.ID)
	dwarf := b12Creature(g, me.ID, "Dwarf", hobA8DwarfTypeLine, 2, 2)
	battle := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Siege", TypeLine: "Battle — Siege",
		Owner: opp.ID, Controller: opp.ID, ProtectorPlayerID: opp.ID,
		Counters: map[string]int{game.CounterDefense: 5},
	})

	before := countTreasuresControlledBy(g, me.ID)
	dealCombatDamageToPlayer(g, dwarf, battle, 2)
	passPriorityAroundTable(t, g)
	if got := countTreasuresControlledBy(g, me.ID) - before; got != 2 {
		t.Errorf("a Dwarf damaged a battle and made %d Treasures, want 2", got)
	}
}

func TestHobA8ThorinGivesYourCreaturesDoubleStrikeForTen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	thorin := hobA8Thorin(g, me.ID)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	err := g.ActivateCatalogAbility(me.ID, thorin, 0, game.ActivateAbilityParams{Strict: true})
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("{10} with an empty pool: %v, want insufficient mana", err)
	}
	for i := 0; i < 10; i++ {
		b06AddMana(me, "C")
	}
	b16Activate(t, g, me.ID, thorin, 0, game.ActivateAbilityParams{Strict: true})

	for _, id := range []uuid.UUID{mine, thorin} {
		if !hasKeywordOnBattlefield(t, g, id, "double strike") {
			t.Errorf("%s did not gain double strike", id)
		}
	}
	if hasKeywordOnBattlefield(t, g, theirs, "double strike") {
		t.Error("an opponent's creature gained double strike")
	}
	advanceToNextSeatsTurn(t, g)
	if hasKeywordOnBattlefield(t, g, mine, "double strike") {
		t.Error("double strike lasted past the turn")
	}
}

// --- Aragorn and Arwen, Wed ----------------------------------------------

func hobA8CastAragorn(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return castHoldCreature(t, g, "Aragorn and Arwen, Wed", "Legendary Creature — Human Elf Noble", hobA8AragornOracle, 3, 6)
}

func TestHobA8AragornCountersAndLifeOnEnteringCountOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life

	aragorn := hobA8CastAragorn(t, g)
	assistCardSettle(t, g, me.ID, nil, true)

	for _, id := range []uuid.UUID{a, b} {
		if n := hobA8Counters(t, g, id); n != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", id, n)
		}
	}
	if n := hobA8Counters(t, g, aragorn); n != 0 {
		t.Errorf("Aragorn and Arwen put %d counters on herself, want 0 (each other creature)", n)
	}
	if n := hobA8Counters(t, g, theirs); n != 0 {
		t.Errorf("an opponent's creature got %d counters, want 0", n)
	}
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2 (one for each other creature you control)", got)
	}
}

func TestHobA8AragornTriggersAgainWhenSheAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	aragorn := pushSizedCreature(g, me.ID, "Aragorn and Arwen, Wed", 3, 6)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == aragorn {
				g.Battlefield.Cards[i].OracleID = hobA8AragornOracle
				g.Battlefield.Cards[i].TypeLine = "Legendary Creature — Human Elf Noble"
			}
		}
	})
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	life := me.Life

	declareAttack(t, g, opp.ID, aragorn)
	passPriorityAroundTable(t, g)

	if n := hobA8Counters(t, g, a); n != 1 {
		t.Errorf("a Bear has %d counters after the attack trigger, want 1", n)
	}
	if n := hobA8Counters(t, g, aragorn); n != 0 {
		t.Errorf("Aragorn and Arwen put %d counters on herself, want 0", n)
	}
	if got := me.Life - life; got != 3 {
		t.Errorf("gained %d life, want 3", got)
	}
	if !hasKeywordOnBattlefield(t, g, aragorn, "vigilance") {
		t.Error("no vigilance")
	}
}

// --- Necklace of Girion ----------------------------------------------------

func hobA8Necklace(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Necklace of Girion", "Legendary Artifact", hobA8NecklaceOracle, false)
}

// hobA8CastColoredCreature casts a 2/2 creature spell of one colour from the
// active player's hand.
func hobA8CastColoredCreature(t *testing.T, g *game.Game, color string) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Test " + color + " Bear", TypeLine: "Creature — Bear", Colors: []string{color},
		Power: 2, Toughness: 2, Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
}

func TestHobA8NecklacePutsACounterOnTargetCreatureWhenYouCastAGreenSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	hobA8Necklace(g, me.ID)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	pick := &game.TargetRef{Kind: game.TargetCard, ID: bear}

	hobA8CastColoredCreature(t, g, "R")
	assistCardSettle(t, g, me.ID, pick, true)
	if n := hobA8Counters(t, g, bear); n != 0 {
		t.Fatalf("a red spell put %d counters, want 0", n)
	}
	hobA8CastColoredCreature(t, g, "G")
	assistCardSettle(t, g, me.ID, pick, true)
	if n := hobA8Counters(t, g, bear); n != 1 {
		t.Errorf("a green spell put %d counters, want 1", n)
	}
}

func TestHobA8NecklacePutsACounterWhenAForestYouControlEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	hobA8Necklace(g, me.ID)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	pick := &game.TargetRef{Kind: game.TargetCard, ID: bear}

	b12PlayFromHand(t, g, "Mountain", "Basic Land — Mountain", "", game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, pick, true)
	if n := hobA8Counters(t, g, bear); n != 0 {
		t.Fatalf("a Mountain put %d counters, want 0", n)
	}
	g.Seats[0].LandDropsPerTurn++
	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, pick, true)
	if n := hobA8Counters(t, g, bear); n != 1 {
		t.Errorf("a Forest put %d counters, want 1", n)
	}
}

func TestHobA8NecklaceIgnoresAnOpponentsForest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	hobA8Necklace(g, me.ID)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	forest := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: forest, NewZone: game.ZoneBattlefield})
	})
	passPriorityAroundTable(t, g)
	if n := hobA8Counters(t, g, bear); n != 0 {
		t.Errorf("an opponent's Forest put %d counters, want 0", n)
	}
}

func TestHobA8NecklaceTapsForGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	necklace := hobA8Necklace(g, me.ID)
	if err := g.ActivateManaAbility(me.ID, necklace, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want [G]", got)
	}
	if !b16Tapped(t, g, necklace) {
		t.Error("the Necklace did not tap")
	}
}

// --- Snowslope Hunter ------------------------------------------------------

func hobA8Hunter(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Snowslope Hunter", "Creature — Goblin Ranger", hobA8HunterOracle, false)
}

func TestHobA8SnowslopeHunterSacrificesAnotherCreatureOrArtifactToImpulseOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	hunter := hobA8Hunter(g, me.ID)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Mana Rock", "Artifact")
	shrine := b12Permanent(g, me.ID, "Shrine", "Enchantment")
	ids := seedSearchLibrary(me,
		game.Card{Name: "Top Land", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)

	if err := g.ActivateCatalogAbility(me.ID, hunter, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{hunter}}); err == nil {
		t.Fatal("Snowslope Hunter can't sacrifice itself: the cost says another")
	}
	if err := g.ActivateCatalogAbility(me.ID, hunter, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{shrine}}); err == nil {
		t.Fatal("an enchantment is neither a creature nor an artifact")
	}
	b16Activate(t, g, me.ID, hunter, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	if g.Battlefield.Contains(bear) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if !g.Exile.Contains(ids[0]) || !me.Library.Contains(ids[1]) {
		t.Fatal("the top card of the library should be exiled and the next one left alone")
	}
	perm := exiledPermission(g, ids[0])
	if perm.Player != me.ID || perm.CastOnly || !permissionLive(g, perm, me.ID) {
		t.Errorf("%+v: the controller may PLAY the card, a land included", perm)
	}

	if err := g.ActivateCatalogAbility(me.ID, hunter, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Fatal("a second activation in the same turn must be refused")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the refused activation still sacrificed the artifact")
	}

	// Through the other seats' turns and my next turn the grant holds, then it lapses.
	advanceToMainOf(t, g, 2)
	if !permissionLive(g, exiledPermission(g, ids[0]), me.ID) {
		t.Error("the grant survives the opponents' turns")
	}
	advanceToMainOf(t, g, 0)
	if !permissionLive(g, exiledPermission(g, ids[0]), me.ID) {
		t.Error("on my next turn the grant is still live")
	}
	b16Activate(t, g, me.ID, hunter, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(rock) {
		t.Error("an artifact is a legal sacrifice on the next turn")
	}
	advanceToMainOf(t, g, 1)
	if permissionLive(g, exiledPermission(g, ids[0]), me.ID) {
		t.Error("the grant ends with my next turn")
	}
}

func TestHobA8SnowslopeHunterOnlyActivatesDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMainOf(t, g, 1)
	hunter := hobA8Hunter(g, me.ID)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	seedSearchLibrary(me, game.Card{Name: "Top Card", TypeLine: "Sorcery"})

	if err := g.ActivateCatalogAbility(me.ID, hunter, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("activating on an opponent's turn must be refused")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the refused activation sacrificed the creature")
	}
}

// --- Dreaded Bat-Cloud -----------------------------------------------------

func TestHobA8DreadedBatCloudCostsThreeLessAfterACreatureDied(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	if mv := selfPricedMV(t, g, me, hobA8BatCloudOracle, "Creature — Bat", "{4}{B}"); mv != 5 {
		t.Fatalf("with no creature dead this turn the Bat-Cloud costs %d, want 5", mv)
	}
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b15Destroy(t, g, victim)
	if mv := selfPricedMV(t, g, me, hobA8BatCloudOracle, "Creature — Bat", "{4}{B}"); mv != 2 {
		t.Errorf("after any creature died the Bat-Cloud costs %d, want 2 ({1}{B})", mv)
	}
	advanceToNextSeatsTurn(t, g)
	if mv := selfPricedMV(t, g, g.Seats[g.Turn.ActiveSeat], hobA8BatCloudOracle, "Creature — Bat", "{4}{B}"); mv != 5 {
		t.Errorf("on a later turn the Bat-Cloud costs %d, want 5", mv)
	}
}

func TestHobA8DreadedBatCloudCastsForTwoAndFliesWithDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b15Destroy(t, g, b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2))
	b06AddMana(me, "B", "B")
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Dreaded Bat-Cloud", TypeLine: "Creature — Bat", ManaCost: "{4}{B}", OracleID: hobA8BatCloudOracle,
		Colors: []string{"B"}, Power: 4, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting for {1}{B} with two mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Bat-Cloud did not resolve")
	}
	for _, kw := range []string{"flying", "deathtouch"} {
		if !hasKeywordOnBattlefield(t, g, id, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

// --- Roll-Roll-Roll-Roll ---------------------------------------------------

func hobA8CastRollSaga(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Roll-Roll-Roll-Roll", "Enchantment — Saga", hobA8RollOracle, nil)
	passPriorityAroundTable(t, g)
	return id
}

func TestHobA8RollRollRollRollBlinksACreatureAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	saga := hobA8CastRollSaga(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("chapter I opened no target prompt")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("an opponent's creature was accepted as the target")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, bear) {
		t.Fatal("the creature was not exiled")
	}
	if got := loreCountersOn(g, saga); got != 1 {
		t.Errorf("lore counters = %d, want 1", got)
	}

	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	back := roonaOnly(t, g, "Bear")
	if back.Controller != me.ID || back.InstanceID == bear {
		t.Errorf("the Bear came back as %v under %v, want a new object under its owner", back.InstanceID, back.Controller)
	}
}

func TestHobA8RollRollRollRollBlinksALandAndLaterChaptersMayChooseNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	saga := hobA8CastRollSaga(t, g)
	pickCard(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, land) {
		t.Fatal("the land was not exiled")
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Forest") == uuid.Nil {
		t.Fatal("the Forest did not come back at the end step")
	}

	advanceToPrecombatMainOf(t, g, 0)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("chapter II opened no target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("choosing no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefieldByID(g, bear) {
		t.Error("choosing nothing still exiled the creature")
	}
	if got := loreCountersOn(g, saga); got != 2 {
		t.Errorf("lore counters = %d, want 2", got)
	}
}

// --- Gandalf, Shadow's Foe -------------------------------------------------

func hobA8Forests(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, b12Permanent(g, owner, "Forest", "Basic Land — Forest"))
	}
	return ids
}

func hobA8TappedForests(g *game.Game, owner uuid.UUID) (tapped, untapped int) {
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Forest" || c.Controller != owner {
			continue
		}
		if c.Tapped {
			tapped++
		} else {
			untapped++
		}
	}
	return tapped, untapped
}

func TestHobA8GandalfBlinksUpToThreeLandsTappedAndEachLandfallsHim(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	forests := hobA8Forests(g, me.ID, 4)
	gandalf := castHoldCreature(t, g, "Gandalf, Shadow's Foe", "Legendary Creature — Avatar Wizard", hobA8GandalfOracle, 3, 4)
	hand := len(me.Hand.Cards)

	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger opened no target prompt")
	}
	var refs []game.TargetRef
	for _, id := range forests[:3] {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, refs); err != nil {
		t.Fatalf("choosing three lands: %v", err)
	}
	passPriorityAroundTable(t, g)

	for _, id := range forests[:3] {
		if onBattlefieldByID(g, id) {
			t.Errorf("land %s kept its identity: it should return as a new object", id)
		}
	}
	if !onBattlefieldByID(g, forests[3]) {
		t.Error("the land that was not chosen left the battlefield")
	}
	tapped, untapped := hobA8TappedForests(g, me.ID)
	if tapped != 3 || untapped != 1 {
		t.Errorf("%d tapped and %d untapped Forests, want 3 and 1", tapped, untapped)
	}
	if got := len(me.Hand.Cards) - hand; got != 3 {
		t.Errorf("drew %d cards from three landfall triggers, want 3", got)
	}
	if n := hobA8Counters(t, g, gandalf); n != 3 {
		t.Errorf("Gandalf has %d +1/+1 counters, want 3", n)
	}
	if !hasKeywordOnBattlefield(t, g, gandalf, "vigilance") {
		t.Error("no vigilance")
	}
}

func TestHobA8GandalfMayChooseNoLandsAndLandfallIsYoursOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	forests := hobA8Forests(g, me.ID, 2)
	gandalf := castHoldCreature(t, g, "Gandalf, Shadow's Foe", "Legendary Creature — Avatar Wizard", hobA8GandalfOracle, 3, 4)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger opened no target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("choosing no lands: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range forests {
		if !onBattlefieldByID(g, id) {
			t.Error("a land was exiled though none was chosen")
		}
	}
	hand := len(me.Hand.Cards)

	theirs := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: theirs, NewZone: game.ZoneBattlefield})
	})
	passPriorityAroundTable(t, g)
	if n := hobA8Counters(t, g, gandalf); n != 0 || len(me.Hand.Cards) != hand {
		t.Errorf("an opponent's land gave %d counters and %d cards, want 0 and 0", n, len(me.Hand.Cards)-hand)
	}

	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, nil, true)
	if n := hobA8Counters(t, g, gandalf); n != 1 {
		t.Errorf("a land you play gave %d counters, want 1", n)
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand = %d, want %d (the test land came and went, one card drawn)", len(me.Hand.Cards), hand+1)
	}
}

// --- Bolg, Erebor's Reckoning ----------------------------------------------

func hobA8Bolg(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Bolg, Erebor's Reckoning", "Legendary Creature — Goblin Soldier", hobA8BolgOracle, 6, 6)
}

func TestHobA8BolgPumpsOtherGoblinsAndOrcsAndShrinksOpposingCreaturesEachCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bolg := hobA8Bolg(g, me.ID)
	goblin := b12Creature(g, me.ID, "Goblin", "Creature — Goblin Warrior", 2, 2)
	orc := b12Creature(g, me.ID, "Orc", "Creature — Orc Soldier", 3, 3)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf Scout", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin Warrior", 3, 3)

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	for _, c := range []struct {
		name string
		id   uuid.UUID
		p, t int
	}{
		{"Bolg himself", bolg, 6, 6},
		{"my Goblin", goblin, 4, 4},
		{"my Orc", orc, 5, 5},
		{"my Elf", elf, 2, 2},
		{"their Goblin", theirs, 2, 2},
	} {
		if p, th := rfPT(t, g, c.id); p != c.p || th != c.t {
			t.Errorf("%s is %d/%d, want %d/%d", c.name, p, th, c.p, c.t)
		}
	}

	// The next player's combat triggers it again: the first pump is gone.
	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if p, th := rfPT(t, g, goblin); p != 4 || th != 4 {
		t.Errorf("on the next turn's combat my Goblin is %d/%d, want 4/4 (once, not stacked)", p, th)
	}
	if p, th := rfPT(t, g, theirs); p != 2 || th != 2 {
		t.Errorf("on the next turn's combat their Goblin is %d/%d, want 2/2", p, th)
	}
}

func TestHobA8BolgTramplesAndKillsOneToughnessOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bolg := hobA8Bolg(g, me.ID)
	weak := b12Creature(g, opp.ID, "Their Mouse", "Creature — Mouse", 1, 1)
	if !hasKeywordOnBattlefield(t, g, bolg, "trample") {
		t.Error("no trample")
	}
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, weak) {
		t.Error("a 1/1 opposing creature should die to -1/-1")
	}
}

// --- The Notary Hobbits ----------------------------------------------------

func TestHobA8NotaryHobbitsMakeTwoNonlegendaryCopiesOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	original := castHoldCreature(t, g, "The Notary Hobbits", "Legendary Creature — Halfling Advisor", hobA8NotaryOracle, 1, 1)
	assistCardSettle(t, g, me.ID, nil, true)

	copies := s58Named(g, me.ID, "The Notary Hobbits")
	if len(copies) != 3 {
		t.Fatalf("%d Notary Hobbits on the battlefield, want the original and two tokens", len(copies))
	}
	tokens := 0
	for _, c := range copies {
		if c.InstanceID == original {
			if !c.IsLegendary() || c.IsToken() {
				t.Errorf("the original should stay a legendary card (legendary=%v token=%v)", c.IsLegendary(), c.IsToken())
			}
			continue
		}
		tokens++
		if !c.IsToken() || c.IsLegendary() {
			t.Errorf("a copy is token=%v legendary=%v, want a nonlegendary token", c.IsToken(), c.IsLegendary())
		}
		if c.OracleID != hobA8NotaryOracle {
			t.Errorf("the copy carries oracle id %q, want the Hobbits' (a real copy)", c.OracleID)
		}
	}
	if tokens != 2 {
		t.Errorf("%d tokens, want 2", tokens)
	}
}

func TestHobA8NotaryHobbitsDoNotCopyThemselvesWhenTheyEnterAsAToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	original := pushCatalogPermanent(g, me.ID, "The Notary Hobbits", "Legendary Creature — Halfling Advisor", hobA8NotaryOracle, false)
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: original})
		if err := (CreateTokenCopy{Controller: me.ID, Copy: original, N: 1}).Apply(ctx); err != nil {
			t.Fatalf("CreateTokenCopy: %v", err)
		}
	})
	assistCardSettle(t, g, me.ID, nil, true)
	if got := len(s58Named(g, me.ID, "The Notary Hobbits")); got != 2 {
		t.Errorf("%d Notary Hobbits, want 2: a token entering must not copy itself", got)
	}
}

func TestHobA8NotaryHobbitsTapForOneColorlessPerHalflingYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	hobbits := pushCatalogPermanent(g, me.ID, "The Notary Hobbits", "Legendary Creature — Halfling Advisor", hobA8NotaryOracle, false)
	b12Creature(g, me.ID, "Hobbit", "Creature — Halfling Rogue", 1, 1)
	b12Creature(g, opp.ID, "Their Hobbit", "Creature — Halfling Rogue", 1, 1)
	b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)

	if err := g.ActivateManaAbility(me.ID, hobbits, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 2 || got[0] != "C" || got[1] != "C" {
		t.Errorf("pool = %v, want [C C] (the Notary Hobbits and my Hobbit, not theirs)", got)
	}
	if !b16Tapped(t, g, hobbits) {
		t.Error("the Notary Hobbits did not tap")
	}
}

// --- Iron Hills Blacksmith -------------------------------------------------

func TestHobA8IronHillsBlacksmithMakesAnAxeThatEquipsForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	smith := castHoldCreature(t, g, "Iron Hills Blacksmith", "Creature — Dwarf Artificer", hobA8BlacksmithOracle, 1, 1)
	assistCardSettle(t, g, me.ID, nil, true)

	if !hasKeywordOnBattlefield(t, g, smith, "double strike") {
		t.Error("the Blacksmith has no double strike")
	}
	axes := s58Named(g, me.ID, "Axe")
	if len(axes) != 1 {
		t.Fatalf("%d Axes, want 1", len(axes))
	}
	axe := axes[0]
	if !axe.IsToken() || !axe.IsArtifact() || !axe.HasSubtype("Equipment") || !axe.IsColorless() {
		t.Errorf("the Axe is %q: want a colorless Equipment artifact token", axe.TypeLine)
	}
	if p, _ := rfPT(t, g, bear); p != 2 {
		t.Fatalf("the Bear is %d power before equipping, want 2", p)
	}

	err := g.ActivateCatalogAbility(me.ID, axe.InstanceID, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("equip {2} with an empty pool: %v, want insufficient mana", err)
	}
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, axe.InstanceID, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("equip {2}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, th := rfPT(t, g, bear); p != 3 || th != 2 {
		t.Errorf("the equipped Bear is %d/%d, want 3/2", p, th)
	}
	if p, _ := rfPT(t, g, smith); p != 1 {
		t.Errorf("the Blacksmith is %d power, want 1 (the Axe is on the Bear)", p)
	}
}
