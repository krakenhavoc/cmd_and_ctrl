package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tyvar_the_bellicose_test.go — #5644: Tyvar the Bellicose, and the
// engine's EventManaAbilityResolved read by a granted trigger.

const tyvarOracle = "a21d321b-2172-4577-a05b-399a019e3448"

// tyvarTable is a catalog game in the active player's main phase with
// Tyvar on their battlefield.
func tyvarTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	tyvar := b12Push(g, me.ID, "Tyvar the Bellicose", "Legendary Creature — Elf Warrior", tyvarOracle, 5, 4)
	return g, me, tyvar
}

// tyvarTap activates a permanent's first mana ability.
func tyvarTap(t *testing.T, g *game.Game, player, card uuid.UUID) {
	t.Helper()
	if err := g.ActivateManaAbility(player, card, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
}

func TestTyvarTheBellicoseIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(tyvarOracle)
	if !ok {
		t.Fatal("Tyvar the Bellicose is not registered")
	}
	if spec.Name != "Tyvar the Bellicose" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// One trigger for the declaration; the Elves that attacked gain
// deathtouch until end of turn, and the Bear beside them does not.
func TestTyvarGivesTheAttackingElvesDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tyvar := b12Push(g, me.ID, "Tyvar the Bellicose", "Legendary Creature — Elf Warrior", tyvarOracle, 5, 4)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf Scout", 1, 1)
	home := b12Creature(g, me.ID, "Homebody", "Creature — Elf Druid", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, tyvar, elf, bear)
	if n := triggersOnStackFrom(g, tyvar); n != 1 {
		t.Fatalf("%d triggers for one declaration, want 1", n)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{tyvar, elf} {
		if !hasKeywordOnBattlefield(t, g, id, "deathtouch") {
			t.Errorf("an Elf that attacked has no deathtouch")
		}
	}
	if hasKeywordOnBattlefield(t, g, bear, "deathtouch") {
		t.Error("a Bear that attacked gained deathtouch")
	}
	if hasKeywordOnBattlefield(t, g, home, "deathtouch") {
		t.Error("an Elf that did not attack gained deathtouch")
	}
	endTurn(t, g)
	if hasKeywordOnBattlefield(t, g, elf, "deathtouch") {
		t.Error("deathtouch should end with the turn")
	}
}

func TestTyvarIgnoresAnAttackWithNoElf(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tyvar := b12Push(g, me.ID, "Tyvar the Bellicose", "Legendary Creature — Elf Warrior", tyvarOracle, 5, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	declareAttack(t, g, opp.ID, bear)
	if n := triggersOnStackFrom(g, tyvar); n != 0 {
		t.Errorf("%d triggers for a Bear's attack, want 0", n)
	}
	if hasKeywordOnBattlefield(t, g, theirs, "deathtouch") {
		t.Error("an opponent's Elf gained deathtouch")
	}
}

// The granted trigger: a counter per mana, once each turn, and again on
// a later turn.
func TestTyvarCountersACreatureForTheManaItProducedOnceEachTurn(t *testing.T) {
	g, me, tyvar := tyvarTable(t)
	elves := pushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", llanowarElvesOracle, false)
	tyvarTap(t, g, me.ID, elves)
	if me.ManaPool.Total() != 1 {
		t.Fatalf("pool = %v, want the Elves' {G}", me.ManaPool)
	}
	if n := triggersOnStackFrom(g, elves); n != 1 {
		t.Fatalf("%d triggers from the Elves, want its granted one", n)
	}
	if n := triggersOnStackFrom(g, tyvar); n != 0 {
		t.Errorf("%d triggers from Tyvar, want none: the ability is the creature's own", n)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, elves, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d +1/+1 counters, want 1 for one mana", got)
	}

	untapForTest(g, elves)
	tyvarTap(t, g, me.ID, elves)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, elves, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters after a second mana ability this turn, want 1", got)
	}

	endTurn(t, g)
	untapForTest(g, elves)
	tyvarTap(t, g, me.ID, elves)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, elves, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters after a mana ability on a later turn, want 2", got)
	}
}

// CR 106.12b: the amount is what the creature produced after a
// replacement. Nyxbloom Ancient triples it.
func TestTyvarCountsReplacedMana(t *testing.T) {
	g, me, _ := tyvarTable(t)
	pushCatalogPermanent(g, me.ID, "Nyxbloom Ancient", "Enchantment Creature — Elemental", nyxbloomAncientOracle, false)
	elves := pushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", llanowarElvesOracle, false)
	tyvarTap(t, g, me.ID, elves)
	if me.ManaPool.Total() != 3 {
		t.Fatalf("pool = %v, want three {G}", me.ManaPool)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, elves, game.CounterPlusOne); got != 3 {
		t.Errorf("%d +1/+1 counters, want 3 for three mana", got)
	}
}

// A colour that is picked: the ability resolves on the answer, and the
// pick carries that through a restore.
func TestTyvarCountersAfterAColourPickAndARestore(t *testing.T) {
	g, me, _ := tyvarTable(t)
	birds := pushCatalogPermanent(g, me.ID, "Birds of Paradise", "Creature — Bird", birdsOfParadiseOracle, false)
	tyvarTap(t, g, me.ID, birds)
	if n := triggersOnStackFrom(g, birds); n != 0 {
		t.Fatalf("%d triggers before the colour is named, want 0", n)
	}
	g = restoreThroughJSON(t, g)
	me = g.Seats[0]
	pick := latestChoiceOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("the restored game has no mana pick")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, birds, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the Birds, want 1", got)
	}
}

// The trigger carries its amount on the stack through a restore.
func TestTyvarTriggerResolvesAfterARestore(t *testing.T) {
	g, me, _ := tyvarTable(t)
	pushCatalogPermanent(g, me.ID, "Nyxbloom Ancient", "Enchantment Creature — Elemental", nyxbloomAncientOracle, false)
	elves := pushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", llanowarElvesOracle, false)
	tyvarTap(t, g, me.ID, elves)
	if n := triggersOnStackFrom(g, elves); n != 1 {
		t.Fatalf("%d triggers from the Elves, want 1", n)
	}
	g = restoreThroughJSON(t, g)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, elves, game.CounterPlusOne); got != 3 {
		t.Errorf("%d +1/+1 counters after a restore, want 3", got)
	}
}

// The auto-tapper activates the ability too.
func TestTyvarCountersACreatureTheAutoTapperTapped(t *testing.T) {
	g, me, _ := tyvarTable(t)
	elves := pushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", llanowarElvesOracle, false)
	spell := handSpell(me, "Giant Growth-ish", "Instant", "{G}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped cast: %v", err)
	}
	if !isTapped(g, elves) {
		t.Fatal("the auto-tapper did not tap the Elves")
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, elves, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

// CR 605.1b: a creature's triggered mana ability is a mana ability of
// that creature, and counts only its own mana.
func TestTyvarCountersACreatureForItsTriggeredManaAbility(t *testing.T) {
	g, me, _ := tyvarTable(t)
	cub := pushCatalogPermanent(g, me.ID, "Badgermole Cub", "Creature — Badger Mole", avB14CubOracle, false)
	elves := pushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", llanowarElvesOracle, false)
	tyvarTap(t, g, me.ID, elves)
	if me.ManaPool.Total() != 2 {
		t.Fatalf("pool = %v, want the Elves' {G} and the Cub's", me.ManaPool)
	}
	// Two triggers of one player's at once: CR 603.3b asks their order.
	answerOpenOrderingIfAny(t, g, me.ID)
	if len(g.StackMeta) != 2 {
		t.Fatalf("stack = %d items, want the Elves' trigger and the Cub's", len(g.StackMeta))
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, elves, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the Elves, want 1: the Cub's mana is not theirs", got)
	}
	if got := countersOn(g, cub, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the Cub, want 1 for its own mana ability", got)
	}
}

// Only creatures Tyvar's controller controls have the ability, and only
// while Tyvar is there.
func TestTyvarGrantsOnlyToYourCreaturesWhileOnTheBattlefield(t *testing.T) {
	g, me, tyvar := tyvarTable(t)
	opp := g.Seats[1]
	theirs := pushCatalogPermanent(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", llanowarElvesOracle, false)
	forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	tyvarTap(t, g, opp.ID, theirs)
	tyvarTap(t, g, me.ID, forest)
	if len(g.StackMeta) != 0 {
		t.Fatalf("stack = %d items, want none for an opponent's creature and a land", len(g.StackMeta))
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, tyvar); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	g.RunStateChecksForTest()
	mine := pushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", llanowarElvesOracle, false)
	tyvarTap(t, g, me.ID, mine)
	if n := triggersOnStackFrom(g, mine); n != 0 {
		t.Errorf("%d triggers with Tyvar gone, want 0", n)
	}
}
