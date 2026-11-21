package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tyvar_the_bellicose_test.go — #5644: Tyvar the Bellicose, and the
// engine's EventManaAbilityResolved read by a granted trigger.

const tyvarOracle = "a21d321b-2172-4577-a05b-399a019e3448"

// tyvarTable is a catalog game in the active player's main phase with
// Tyvar on their battlefield.
func tyvarTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	tyvar := effectstest.B12Push(g, me.ID, "Tyvar the Bellicose", "Legendary Creature — Elf Warrior", tyvarOracle, 5, 4)
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
	spec, ok := effects.Lookup(tyvarOracle)
	if !ok {
		t.Fatal("Tyvar the Bellicose is not registered")
	}
	if spec.Name != "Tyvar the Bellicose" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// One trigger for the declaration; the Elves that attacked gain
// deathtouch until end of turn, and the Bear beside them does not.
func TestTyvarGivesTheAttackingElvesDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tyvar := effectstest.B12Push(g, me.ID, "Tyvar the Bellicose", "Legendary Creature — Elf Warrior", tyvarOracle, 5, 4)
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf Scout", 1, 1)
	home := effectstest.B12Creature(g, me.ID, "Homebody", "Creature — Elf Druid", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, tyvar, elf, bear)
	if n := effectstest.TriggersOnStackFrom(g, tyvar); n != 1 {
		t.Fatalf("%d triggers for one declaration, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{tyvar, elf} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "deathtouch") {
			t.Errorf("an Elf that attacked has no deathtouch")
		}
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "deathtouch") {
		t.Error("a Bear that attacked gained deathtouch")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, home, "deathtouch") {
		t.Error("an Elf that did not attack gained deathtouch")
	}
	effectstest.EndTurn(t, g)
	if effectstest.HasKeywordOnBattlefield(t, g, elf, "deathtouch") {
		t.Error("deathtouch should end with the turn")
	}
}

func TestTyvarIgnoresAnAttackWithNoElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tyvar := effectstest.B12Push(g, me.ID, "Tyvar the Bellicose", "Legendary Creature — Elf Warrior", tyvarOracle, 5, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if n := effectstest.TriggersOnStackFrom(g, tyvar); n != 0 {
		t.Errorf("%d triggers for a Bear's attack, want 0", n)
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "deathtouch") {
		t.Error("an opponent's Elf gained deathtouch")
	}
}

// The granted trigger: a counter per mana, once each turn, and again on
// a later turn.
func TestTyvarCountersACreatureForTheManaItProducedOnceEachTurn(t *testing.T) {
	g, me, tyvar := tyvarTable(t)
	elves := effectstest.PushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", effectstest.LlanowarElvesOracle, false)
	tyvarTap(t, g, me.ID, elves)
	if me.ManaPool.Total() != 1 {
		t.Fatalf("pool = %v, want the Elves' {G}", me.ManaPool)
	}
	if n := effectstest.TriggersOnStackFrom(g, elves); n != 1 {
		t.Fatalf("%d triggers from the Elves, want its granted one", n)
	}
	if n := effectstest.TriggersOnStackFrom(g, tyvar); n != 0 {
		t.Errorf("%d triggers from Tyvar, want none: the ability is the creature's own", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, elves, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d +1/+1 counters, want 1 for one mana", got)
	}

	effectstest.UntapForTest(g, elves)
	tyvarTap(t, g, me.ID, elves)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, elves, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters after a second mana ability this turn, want 1", got)
	}

	effectstest.EndTurn(t, g)
	effectstest.UntapForTest(g, elves)
	tyvarTap(t, g, me.ID, elves)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, elves, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters after a mana ability on a later turn, want 2", got)
	}
}

// CR 106.12b: the amount is what the creature produced after a
// replacement. Nyxbloom Ancient triples it.
func TestTyvarCountsReplacedMana(t *testing.T) {
	g, me, _ := tyvarTable(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Nyxbloom Ancient", "Enchantment Creature — Elemental", effectstest.NyxbloomAncientOracle, false)
	elves := effectstest.PushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", effectstest.LlanowarElvesOracle, false)
	tyvarTap(t, g, me.ID, elves)
	if me.ManaPool.Total() != 3 {
		t.Fatalf("pool = %v, want three {G}", me.ManaPool)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, elves, game.CounterPlusOne); got != 3 {
		t.Errorf("%d +1/+1 counters, want 3 for three mana", got)
	}
}

// A colour that is picked: the ability resolves on the answer, and the
// pick carries that through a restore.
func TestTyvarCountersAfterAColourPickAndARestore(t *testing.T) {
	g, me, _ := tyvarTable(t)
	birds := effectstest.PushCatalogPermanent(g, me.ID, "Birds of Paradise", "Creature — Bird", effectstest.BirdsOfParadiseOracle, false)
	tyvarTap(t, g, me.ID, birds)
	if n := effectstest.TriggersOnStackFrom(g, birds); n != 0 {
		t.Fatalf("%d triggers before the colour is named, want 0", n)
	}
	g = effectstest.RestoreThroughJSON(t, g)
	me = g.Seats[0]
	pick := effectstest.LatestChoiceOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("the restored game has no mana pick")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, birds, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the Birds, want 1", got)
	}
}

// The trigger carries its amount on the stack through a restore.
func TestTyvarTriggerResolvesAfterARestore(t *testing.T) {
	g, me, _ := tyvarTable(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Nyxbloom Ancient", "Enchantment Creature — Elemental", effectstest.NyxbloomAncientOracle, false)
	elves := effectstest.PushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", effectstest.LlanowarElvesOracle, false)
	tyvarTap(t, g, me.ID, elves)
	if n := effectstest.TriggersOnStackFrom(g, elves); n != 1 {
		t.Fatalf("%d triggers from the Elves, want 1", n)
	}
	g = effectstest.RestoreThroughJSON(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, elves, game.CounterPlusOne); got != 3 {
		t.Errorf("%d +1/+1 counters after a restore, want 3", got)
	}
}

// The auto-tapper activates the ability too.
func TestTyvarCountersACreatureTheAutoTapperTapped(t *testing.T) {
	g, me, _ := tyvarTable(t)
	elves := effectstest.PushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", effectstest.LlanowarElvesOracle, false)
	spell := effectstest.HandSpell(me, "Giant Growth-ish", "Instant", "{G}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped cast: %v", err)
	}
	if !effectstest.IsTapped(g, elves) {
		t.Fatal("the auto-tapper did not tap the Elves")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, elves, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

// CR 605.1b: a creature's triggered mana ability is a mana ability of
// that creature, and counts only its own mana.
func TestTyvarCountersACreatureForItsTriggeredManaAbility(t *testing.T) {
	g, me, _ := tyvarTable(t)
	cub := effectstest.PushCatalogPermanent(g, me.ID, "Badgermole Cub", "Creature — Badger Mole", effectstest.AvB14CubOracle, false)
	elves := effectstest.PushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", effectstest.LlanowarElvesOracle, false)
	tyvarTap(t, g, me.ID, elves)
	if me.ManaPool.Total() != 2 {
		t.Fatalf("pool = %v, want the Elves' {G} and the Cub's", me.ManaPool)
	}
	// Two triggers of one player's at once: CR 603.3b asks their order.
	effectstest.AnswerOpenOrderingIfAny(t, g, me.ID)
	if len(g.StackMeta) != 2 {
		t.Fatalf("stack = %d items, want the Elves' trigger and the Cub's", len(g.StackMeta))
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, elves, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the Elves, want 1: the Cub's mana is not theirs", got)
	}
	if got := effectstest.CountersOn(g, cub, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the Cub, want 1 for its own mana ability", got)
	}
}

// Only creatures Tyvar's controller controls have the ability, and only
// while Tyvar is there.
func TestTyvarGrantsOnlyToYourCreaturesWhileOnTheBattlefield(t *testing.T) {
	g, me, tyvar := tyvarTable(t)
	opp := g.Seats[1]
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", effectstest.LlanowarElvesOracle, false)
	forest := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	tyvarTap(t, g, opp.ID, theirs)
	tyvarTap(t, g, me.ID, forest)
	if len(g.StackMeta) != 0 {
		t.Fatalf("stack = %d items, want none for an opponent's creature and a land", len(g.StackMeta))
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, tyvar); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	g.RunStateChecksForTest()
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", effectstest.LlanowarElvesOracle, false)
	tyvarTap(t, g, me.ID, mine)
	if n := effectstest.TriggersOnStackFrom(g, mine); n != 0 {
		t.Errorf("%d triggers with Tyvar gone, want 0", n)
	}
}
