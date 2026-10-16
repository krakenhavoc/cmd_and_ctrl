package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cumulative_upkeep_grant_test.go — #3869, CR 702.24: asking whether a
// permanent has cumulative upkeep, and giving it cumulative upkeep.

const (
	cugBalduvianShaman = "3354c3d3-5713-4139-beb2-4fa17f6b47d9"
	cugManaChains      = "340d27c6-2e8e-472a-85f5-9b36d9cc3e42"
	cugBreathOfDreams  = "477f5011-aaeb-4a5f-8a65-5fb45cb06389"
	cugDreamsOfTheDead = "3b6cf935-3b99-4371-a797-6f40b89adeeb"
	cugDecomposition   = "62817ef5-9b8d-4bca-bdd4-ab3bc4e063d0"
)

// cugPut puts a catalog or vanilla permanent onto the battlefield with
// its colours, so the layer pass and the target walk see them.
func cugPut(g *game.Game, owner uuid.UUID, name, oracle, typeLine string, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, OracleID: oracle, TypeLine: typeLine,
		Power: 2, Toughness: 2, Colors: colors, Owner: owner, Controller: owner,
	})
}

func cugHas(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return game.CumulativeUpkeepInstances(gaLayered(t, g, id))
}

// cugTriggersFrom counts the triggered abilities from `id` on the stack.
func cugTriggersFrom(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == id {
			n++
		}
	}
	return n
}

// cugUpkeep moves to seat 1's upkeep and resolves the stack up to the
// first cumulative-upkeep prompt.
func cugUpkeep(t *testing.T, g *game.Game) {
	t.Helper()
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
}

// Mana Chains: the enchanted creature HAS cumulative upkeep {1}, its
// controller is asked, the age counter goes on the creature, and a
// decline sacrifices the creature, not the Aura's controller's anything.
func TestManaChainsGivesTheEnchantedCreatureCumulativeUpkeep(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		auraOwner, victim := g.Seats[0], g.Seats[1]
		bear := cugPut(g, victim.ID, "Grizzly Bears", "", "Creature — Bear", "G")
		chains := cugPut(g, auraOwner.ID, "Mana Chains", cugManaChains, "Enchantment — Aura", "U")
		ncdAttach(t, g, chains, bear)
		if got := cugHas(t, g, bear); got != 1 {
			t.Fatalf("the enchanted creature has %d instances of cumulative upkeep, want 1", got)
		}
		if got := cugHas(t, g, chains); got != 0 {
			t.Fatalf("Mana Chains itself has %d instances of cumulative upkeep, want 0", got)
		}

		advanceToUpkeepOf(t, g, 1)
		if cugTriggersFrom(g, bear) != 1 {
			t.Fatal("the granted cumulative upkeep did not trigger from the creature at its controller's upkeep")
		}
		passPriorityAroundTable(t, g)
		if got := ageCountersOn(t, g, bear); got != 1 {
			t.Fatalf("age counters on the creature = %d, want 1", got)
		}
		choice := remoraUpkeepPrompt(t, g, victim.ID)
		if choice.PayCost != "{1}" {
			t.Fatalf("cost = %q, want {1}", choice.PayCost)
		}
		if pay {
			victim.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
		if err := g.ResolvePayUnless(choice.ID, victim.ID, pay); err != nil {
			t.Fatalf("answer the upkeep: %v", err)
		}
		_, alive := battlefieldCard(g, bear)
		if alive != pay {
			t.Fatalf("paid=%v: the creature on the battlefield = %v", pay, alive)
		}
		if _, ok := battlefieldCard(g, chains); ok && !pay {
			// The Aura falls off as a state-based action once its
			// creature is gone; it was never the thing sacrificed.
			passPriorityAroundTable(t, g)
		}
	}
}

// CR 702.24b: two instances trigger separately, and each one counts
// every age counter on the permanent as it resolves. A green creature
// under Breath of Dreams and Mana Chains pays {1}, then {1}{1}.
func TestTwoInstancesOfCumulativeUpkeepEachTriggerAndCountEveryCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	bear := cugPut(g, victim.ID, "Grizzly Bears", "", "Creature — Bear", "G")
	cugPut(g, me.ID, "Breath of Dreams", cugBreathOfDreams, "Enchantment", "U")
	chains := cugPut(g, me.ID, "Mana Chains", cugManaChains, "Enchantment — Aura", "U")
	ncdAttach(t, g, chains, bear)
	if got := cugHas(t, g, bear); got != 2 {
		t.Fatalf("the creature has %d instances of cumulative upkeep, want 2", got)
	}

	advanceToUpkeepOf(t, g, 1)
	if got := cugTriggersFrom(g, bear); got != 2 {
		t.Fatalf("%d cumulative upkeep triggers from the creature, want 2", got)
	}
	for want := 1; want <= 2; want++ {
		passPriorityAroundTable(t, g)
		if got := ageCountersOn(t, g, bear); got != want {
			t.Fatalf("instance %d: age counters = %d, want %d", want, got, want)
		}
		choice := remoraUpkeepPrompt(t, g, victim.ID)
		parsed, err := game.ParseCost(choice.PayCost)
		if err != nil || parsed.Generic != want {
			t.Fatalf("instance %d: cost %q, want {%d} (err %v)", want, choice.PayCost, want, err)
		}
		for i := 0; i < want; i++ {
			victim.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
		if err := g.ResolvePayUnless(choice.ID, victim.ID, true); err != nil {
			t.Fatalf("instance %d: pay: %v", want, err)
		}
	}
	if _, ok := battlefieldCard(g, bear); !ok {
		t.Fatal("the creature was sacrificed though both instances were paid")
	}
}

// Breath of Dreams reaches every green creature, whoever controls it,
// and no other creature; it has its own cumulative upkeep {U}.
func TestBreathOfDreamsGivesGreenCreaturesCumulativeUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	breath := cugPut(g, me.ID, "Breath of Dreams", cugBreathOfDreams, "Enchantment", "U")
	mine := cugPut(g, me.ID, "Llanowar Elves", "", "Creature — Elf Druid", "G")
	theirs := cugPut(g, opp.ID, "Grizzly Bears", "", "Creature — Bear", "G")
	red := cugPut(g, opp.ID, "Goblin Piker", "", "Creature — Goblin", "R")
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{
		{"Breath of Dreams", breath, 1},
		{"your green creature", mine, 1},
		{"an opponent's green creature", theirs, 1},
		{"a red creature", red, 0},
	} {
		if got := cugHas(t, g, tc.id); got != tc.want {
			t.Errorf("%s: %d instances of cumulative upkeep, want %d", tc.name, got, tc.want)
		}
	}
}

// Balduvian Shaman: only a white enchantment you control that doesn't
// have cumulative upkeep is a legal target, and the ability changes a
// colour word and gives it cumulative upkeep {1} for good, which then
// makes it an illegal target for the next activation.
func TestBalduvianShamanChangesAWordAndGrantsCumulativeUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shaman := cugPut(g, me.ID, "Balduvian Shaman", cugBalduvianShaman, "Creature — Human Cleric Shaman", "U")
	white := cugPut(g, me.ID, "Circle of Protection: Black", "", "Enchantment", "W")
	blue := cugPut(g, me.ID, "Propaganda", "", "Enchantment", "U")
	theirs := cugPut(g, opp.ID, "Circle of Protection: Red", "", "Enchantment", "W")

	for _, bad := range []struct {
		name string
		id   uuid.UUID
	}{{"a blue enchantment", blue}, {"an opponent's white enchantment", theirs}} {
		if err := activateOn(t, g, me.ID, shaman, 0, bad.id); err == nil {
			t.Fatalf("Balduvian Shaman targeted %s", bad.name)
		}
	}
	if err := activateOn(t, g, me.ID, shaman, 0, white); err != nil {
		t.Fatalf("activate on a white enchantment: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, twBlack)
	answerOptionPick(t, g, me.ID, twWhite) // W, U, R, G after black is removed: white is 0
	passPriorityAroundTable(t, g)

	if got := twTextChanges(t, g, white); len(got) != 1 || got[0] != "black → white" {
		t.Fatalf("text changes = %v, want [black → white]", got)
	}
	if got := cugHas(t, g, white); got != 1 {
		t.Fatalf("the enchantment has %d instances of cumulative upkeep, want 1", got)
	}

	// It now has cumulative upkeep, so it is no longer a legal target.
	untapForTest(g, shaman)
	if err := activateOn(t, g, me.ID, shaman, 0, white); err == nil {
		t.Fatal("Balduvian Shaman targeted an enchantment that already has cumulative upkeep")
	}

	// The grant lasts past the turn and charges at its controller's upkeep.
	for g.Turn.ActiveSeat == 0 {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("advance: %v", err)
		}
	}
	advanceToUpkeepOf(t, g, 0)
	if cugTriggersFrom(g, white) != 1 {
		t.Fatal("the granted cumulative upkeep did not trigger at its controller's next upkeep")
	}
}

// The grant is made even when no word changed: the ruling lets the
// Shaman target an enchantment with no colour words at all, and the
// gain does not depend on the change.
func TestBalduvianShamanGrantsEvenWithNoWordToChange(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shaman := cugPut(g, me.ID, "Balduvian Shaman", cugBalduvianShaman, "Creature — Human Cleric Shaman", "U")
	white := cugPut(g, me.ID, "Plain White Enchantment", "", "Enchantment", "W")
	if err := activateOn(t, g, me.ID, shaman, 0, white); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, twRed)
	answerOptionPick(t, g, me.ID, twBlue)
	passPriorityAroundTable(t, g)
	if got := cugHas(t, g, white); got != 1 {
		t.Fatalf("the enchantment has %d instances of cumulative upkeep, want 1", got)
	}
}

// A printed cumulative upkeep counts too: Breath of Dreams is a white
// enchantment once the Shaman's own target clause is irrelevant, so use
// the predicate directly on a printed instance and on an object that
// lost its abilities.
func TestDoesNotHaveCumulativeUpkeepSeesPrintedInstances(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	remora := cugPut(g, me.ID, "Mystic Remora", oracleMysticRemora, "Enchantment", "U")
	plain := cugPut(g, me.ID, "Plain Enchantment", "", "Enchantment", "W")
	pred := DoesNotHaveCumulativeUpkeep()
	if pred(g, me.ID, gaLayered(t, g, remora)) {
		t.Error("Mystic Remora's printed cumulative upkeep was not seen")
	}
	if !pred(g, me.ID, gaLayered(t, g, plain)) {
		t.Error("an enchantment with no cumulative upkeep was refused")
	}
}

// Dreams of the Dead returns a white or black creature card with
// cumulative upkeep {2}; an unpaid upkeep sacrifices it, and the
// replacement exiles it instead of putting it into the graveyard.
func TestDreamsOfTheDeadReturnsWithCumulativeUpkeepAndExilesOnLeaving(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	dreams := cugPut(g, me.ID, "Dreams of the Dead", cugDreamsOfTheDead, "Enchantment", "U")
	black := pushGraveyardMaterial(me, game.Card{Name: "Black Knight", TypeLine: "Creature — Human Knight", Colors: []string{"B"}, Power: 2, Toughness: 2})
	green := pushGraveyardMaterial(me, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2})

	advanceToUpkeepOf(t, g, 1)
	for i := 0; i < 2; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	}
	if err := activateOn(t, g, me.ID, dreams, 0, green); err == nil {
		t.Fatal("Dreams of the Dead targeted a green creature card")
	}
	if err := activateOn(t, g, me.ID, dreams, 0, black); err != nil {
		t.Fatalf("activate on a black creature card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, black); !ok {
		t.Fatal("the black creature card was not returned to the battlefield")
	}
	if got := cugHas(t, g, black); got != 1 {
		t.Fatalf("the returned creature has %d instances of cumulative upkeep, want 1", got)
	}

	// Its controller's next upkeep: {2} for one age counter, declined.
	for g.Turn.ActiveSeat == 1 {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("advance: %v", err)
		}
	}
	cugUpkeep(t, g)
	choice := remoraUpkeepPrompt(t, g, me.ID)
	if choice.PayCost != "{2}" {
		t.Fatalf("cost = %q, want {2}", choice.PayCost)
	}
	if err := g.ResolvePayUnless(choice.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if _, ok := battlefieldCard(g, black); ok {
		t.Fatal("the creature survived an unpaid cumulative upkeep")
	}
	if !exileHas(g, black) {
		t.Fatal("the sacrificed creature was not exiled instead")
	}
	if _, ok := graveyardCard(me, black); ok {
		t.Fatal("the sacrificed creature reached the graveyard")
	}
}

// Decomposition: the enchanted black creature's controller pays one life
// per age counter, or sacrifices it; when it dies its controller loses 2
// life.
func TestDecompositionChargesLifePerAgeCounter(t *testing.T) {
	g := newCatalogGame(t)
	auraOwner, victim := g.Seats[0], g.Seats[1]
	knight := cugPut(g, victim.ID, "Black Knight", "", "Creature — Human Knight", "B")
	aura := cugPut(g, auraOwner.ID, "Decomposition", cugDecomposition, "Enchantment — Aura", "G")
	ncdAttach(t, g, aura, knight)

	// Upkeep 1 pays 1 life; upkeep 2 pays 2.
	for turn := 1; turn <= 2; turn++ {
		if turn > 1 {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("leave the upkeep: %v", err)
			}
		}
		cugUpkeep(t, g)
		life := victim.Life
		choice := remoraUpkeepPrompt(t, g, victim.ID)
		if err := g.ResolvePayUnless(choice.ID, victim.ID, true); err != nil {
			t.Fatalf("turn %d: pay: %v", turn, err)
		}
		if got := life - victim.Life; got != turn {
			t.Fatalf("turn %d: paid %d life, want %d", turn, got, turn)
		}
	}

	// Upkeep 3: declined, the creature is sacrificed, and the Aura's
	// dies trigger takes 2 life from the creature's controller.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("leave the upkeep: %v", err)
	}
	cugUpkeep(t, g)
	life := victim.Life
	choice := remoraUpkeepPrompt(t, g, victim.ID)
	if err := g.ResolvePayUnless(choice.ID, victim.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, knight); ok {
		t.Fatal("the creature survived an unpaid cumulative upkeep")
	}
	if got := life - victim.Life; got != 2 {
		t.Fatalf("the dead creature's controller lost %d life, want 2", got)
	}
}
