package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// taii_wakeen_perfect_shot_test.go — #4468: Taii's draw on noncombat
// damage equal to the creature's toughness, and its {X}, {T} turn-long
// "that much damage plus X" for noncombat damage only.

const (
	taiiOracle      = "17327d50-be12-423d-9d37-d6bae519f50c"
	taiiBoltOracle  = "4457ed35-7c10-48c8-9776-456485fdf070"
	taiiTypeLine    = "Legendary Creature — Human Mercenary"
	taiiActivateRow = 0
)

func taiiPush(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushDiesCreatureForTest(g, owner, "Taii Wakeen, Perfect Shot", taiiOracle, taiiTypeLine, 2, 3)
}

func taiiBolt(t *testing.T, g *game.Game, target game.TargetRef) {
	t.Helper()
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", taiiBoltOracle, []game.TargetRef{target})
	passPriorityAroundTable(t, g)
}

func taiiCard(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func taiiPlayer(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetPlayer, ID: id} }

func taiiDamageOn(g *game.Game, id uuid.UUID) int {
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return -1
	}
	return c.DamageMarked
}

// Exactly the creature's toughness from a source you control draws;
// more, less, combat damage or an opponent's source does not. Damage
// already marked does not matter (the first ruling).
func TestTaiiDrawsOnNoncombatDamageEqualToToughness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	taiiPush(g, me.ID)
	three := pushVanillaCreature(g, opp.ID, "Three", 1, 3)
	four := pushVanillaCreature(g, opp.ID, "Four", 1, 4)
	two := pushVanillaCreature(g, opp.ID, "Two", 1, 2)
	hurt := pushVanillaCreature(g, opp.ID, "Hurt", 1, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == hurt {
				g.Battlefield.Cards[i].DamageMarked = 1
			}
		}
	})

	for _, tc := range []struct {
		name   string
		target uuid.UUID
		draws  int
	}{
		{"3 to a 3-toughness creature", three, 1},
		{"3 to a 4-toughness creature", four, 0},
		{"3 to a 2-toughness creature", two, 0},
		{"3 to a 3-toughness creature already holding 1 damage", hurt, 1},
	} {
		hand := me.Hand.Size()
		taiiBolt(t, g, taiiCard(tc.target))
		if got := me.Hand.Size() - hand; got != tc.draws {
			t.Errorf("%s: drew %d, want %d", tc.name, got, tc.draws)
		}
	}
	hand := me.Hand.Size()
	taiiBolt(t, g, taiiPlayer(opp.ID))
	if me.Hand.Size() != hand {
		t.Errorf("damage to a player drew a card")
	}
}

// Combat damage equal to the toughness does not draw.
func TestTaiiDoesNotDrawOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	taiiPush(g, me.ID)
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 3, 3)
	blocker := pushVanillaCreature(g, opp.ID, "Blocker", 1, 3)
	hand := me.Hand.Size()
	pr7aBlockedCombat(t, g, opp.ID, attacker, blocker)
	advanceTo(t, g, game.StepEndCombat)
	if findBattlefieldByName(g, "Blocker") != uuid.Nil {
		t.Fatalf("setup: the blocker survived; it was not dealt 3 combat damage")
	}
	if me.Hand.Size() != hand {
		t.Errorf("combat damage equal to toughness drew a card")
	}
}

// {X}, {T}: your noncombat damage gets +X this turn, to players and
// permanents; combat damage does not; an opponent's source does not.
// Lethal-by-toughness after the bonus draws.
func TestTaiiAddsXToNoncombatDamageThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	taii := taiiPush(g, me.ID)
	five := pushVanillaCreature(g, opp.ID, "Five", 1, 9)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(me.ID, taii, taiiActivateRow, game.ActivateAbilityParams{XValue: 2}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !isTapped(g, taii) {
		t.Errorf("Taii did not tap for its cost")
	}

	life := opp.Life
	taiiBolt(t, g, taiiPlayer(opp.ID))
	if got := life - opp.Life; got != 5 {
		t.Errorf("Lightning Bolt dealt %d to a player, want 5 (3 plus 2)", got)
	}
	hand := me.Hand.Size()
	taiiBolt(t, g, taiiCard(five))
	if got := taiiDamageOn(g, five); got != 5 {
		t.Errorf("Lightning Bolt marked %d on a creature, want 5", got)
	}
	if me.Hand.Size() != hand {
		t.Errorf("5 damage to a 9-toughness creature drew")
	}

	// A 5-toughness creature dealt 3+2: exactly its toughness, so Taii draws.
	exact := pushVanillaCreature(g, opp.ID, "Exact", 1, 5)
	hand = me.Hand.Size()
	taiiBolt(t, g, taiiCard(exact))
	if me.Hand.Size() != hand+1 {
		t.Errorf("3 plus 2 to a 5-toughness creature: drew %d, want 1", me.Hand.Size()-hand)
	}

	attacker := pushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	life = opp.Life
	attackWith(t, g, opp.ID, attacker)
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 2 {
		t.Errorf("combat damage dealt %d, want 2: only noncombat damage is increased", got)
	}

	// An opponent's source is not "a source you control".
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	myLife := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirs, me.ID, 1) })
	if got := myLife - me.Life; got != 1 {
		t.Errorf("an opponent's source dealt %d, want 1", got)
	}

	// "This turn": gone when the turn ends.
	advanceToNextSeatsTurn(t, g)
	var lines []string
	g.ReadSnapshot(func() { lines = g.DamageMultiplierLines() })
	if len(lines) != 0 {
		t.Errorf("the increase outlived the turn: %q", lines)
	}
}

// X = 0 registers nothing (plus 0 changes nothing).
func TestTaiiWithXZeroAddsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	taii := taiiPush(g, me.ID)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(me.ID, taii, taiiActivateRow, game.ActivateAbilityParams{XValue: 0}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	var lines []string
	g.ReadSnapshot(func() { lines = g.DamageMultiplierLines() })
	if len(lines) != 0 {
		t.Errorf("X = 0 registered %q", lines)
	}
}
