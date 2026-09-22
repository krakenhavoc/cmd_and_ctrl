package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// exhaust_test.go — the enumerator half of #1181, and of #1183. An
// exhaust ability this object has already activated is not a move: the
// engine refuses it with ErrAbilityExhausted, so a bot seat must never
// be offered it (#544's rule, and the reason the enumerator and the
// engine read the same game.AbilityExhausted rather than two copies of
// the rule).
//
// #1183 adds the MANA half, which is the same sentence through the
// sibling reader (game.ManaAbilityExhausted) on the other move kind —
// `manaMoves` rather than `activatedMoves`.

const (
	oracleProwcatcherSpecialist = "594132ac-32a7-41d5-b7f0-3692bbed7f6f"
	oracleGreenbeltGuardian     = "2d8aa053-289d-40d9-baa7-9bd1c5b8e957"
	oracleLootThePathfinder     = "68c7e459-0932-4644-a3c0-9a1eae1db7a3"
)

func TestASpentExhaustAbilityIsNotEnumerated(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	goblin := battlefieldCard(g, active, game.Card{
		Name: "Prowcatcher Specialist", TypeLine: "Creature — Goblin Warrior",
		ManaCost: "{1}{R}", Power: 2, Toughness: 2, OracleID: oracleProwcatcherSpecialist,
	})
	basics(g, active, "Mountain", 4)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, goblin)
	if len(acts) != 1 {
		t.Fatalf("an unspent exhaust ability: want 1 activation offered, got %v", labels(acts))
	}
	// Every offer is one the dispatcher accepts.
	dispatchAll(t, g, active.ID, acts)

	if err := g.ActivateCatalogAbility(active.ID, goblin, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("activate the exhaust ability: %v", err)
	}
	basics(g, active, "Mountain", 4)

	if acts := activationsOf(legal.EnumerateFor(g, active.ID), goblin); len(acts) != 0 {
		t.Errorf("a spent exhaust ability: want no activation offered, got %v — the engine "+
			"refuses it, so nothing may propose it", labels(acts))
	}
}

// TestSpendingAnExhaustAbilityLeavesTheOtherOnesEnumerated is the
// per-ABILITY half on the printed card that has both: Greenbelt
// Guardian's repeatable "{G}: target creature gains trample" is still
// a move after its exhaust ability is gone.
func TestSpendingAnExhaustAbilityLeavesTheOtherOnesEnumerated(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	elf := battlefieldCard(g, active, game.Card{
		Name: "Greenbelt Guardian", TypeLine: "Creature — Elf Ranger",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, OracleID: oracleGreenbeltGuardian,
	})
	basics(g, active, "Forest", 6)
	advanceTo(t, g, game.StepPrecombatMain)

	if acts := activationsOf(legal.EnumerateFor(g, active.ID), elf); len(acts) < 2 {
		t.Fatalf("both abilities should be offered, got %v", labels(acts))
	}
	// Index 1 is the exhaust ability; index 0 is the repeatable pump.
	if err := g.ActivateCatalogAbility(active.ID, elf, 1, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("activate the exhaust ability: %v", err)
	}
	basics(g, active, "Forest", 6)

	acts := activationsOf(legal.EnumerateFor(g, active.ID), elf)
	if len(acts) == 0 {
		t.Fatal("the repeatable ability stopped being a move when the exhaust ability was spent")
	}
	for _, m := range acts {
		if m.Label == "Exhaust — {3}{G}: Put three +1/+1 counters on this creature." {
			t.Errorf("the spent exhaust ability is still offered: %v", labels(acts))
		}
	}
	dispatchAll(t, g, active.ID, acts)
}

// TestASpentExhaustManaAbilityIsNotEnumerated is #1183's half, on
// Loot, the Pathfinder: the "Exhaust — {G}, {T}: Add three mana of any
// one color" ability is a KindMana move until it is spent, and then it
// is no move at all. Enumerating it after the fact would hand a bot a
// tap the engine refuses with ErrAbilityExhausted.
func TestASpentExhaustManaAbilityIsNotEnumerated(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	loot := battlefieldCard(g, active, game.Card{
		Name: "Loot, the Pathfinder", TypeLine: "Legendary Creature — Beast Noble",
		ManaCost: "{2}{G}{U}{R}", Power: 2, Toughness: 4, OracleID: oracleLootThePathfinder,
	})
	basics(g, active, "Forest", 4)
	// One more Forest, by ID: the CONTROL for the assertion at the
	// bottom. An ordinary mana source on the same board must still be
	// enumerated after Loot's exhaust is spent, so "no mana move" is a
	// statement about the ability rather than about the move list
	// having gone empty for some other reason (an unanswered prompt,
	// lost priority).
	control := battlefieldCard(g, active, basic("Forest", "Forest"))
	advanceTo(t, g, game.StepPrecombatMain)

	// Non-vacuity: the {G} in the cost has to be floatable, so the
	// enumerator offers the tap only with a Forest available for it.
	if got := movesFrom(legal.EnumerateFor(g, active.ID), loot, legal.KindMana); len(got) != 0 {
		t.Fatalf("with an empty pool the mana ability is not a move yet, got %v", labels(got))
	}
	active.ManaPool.AddMana(game.ManaToken{Color: "G"})
	before := movesFrom(legal.EnumerateFor(g, active.ID), loot, legal.KindMana)
	if len(before) != 1 {
		t.Fatalf("an unspent exhaust mana ability: want 1 mana move offered, got %v", labels(before))
	}

	if err := g.ActivateManaAbility(active.ID, loot, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate the exhaust mana ability: %v", err)
	}
	// "Any one color" opens a mana_pick, and an open pending choice
	// suppresses the whole move list — answering it is what keeps the
	// assertion below about the exhaust record rather than about the
	// prompt.
	answerManaPick(t, g, active.ID)
	// Untapped and paid for again, so the absence below is the record
	// and not the {T}.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == loot {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	active.ManaPool = nil
	active.ManaPool.AddMana(game.ManaToken{Color: "G"})
	// The control: an ordinary Forest on the same board is still a
	// mana move, so the emptiness below is about Loot's ability and
	// not about the move list.
	after := legal.EnumerateFor(g, active.ID)
	if got := movesFrom(after, control, legal.KindMana); len(got) == 0 {
		t.Fatalf("the control Forest stopped being a mana move too; the fixture is no longer " +
			"isolating the exhaust ability")
	}

	if got := movesFrom(after, loot, legal.KindMana); len(got) != 0 {
		t.Errorf("a spent exhaust mana ability: want no mana move offered, got %v — the engine "+
			"refuses it, so nothing may propose it", labels(got))
	}
}

// answerManaPick resolves an open "add mana of any one color" prompt
// with its first offered colour. An unanswered pending choice empties
// the move list, which would make any "this is no longer a move"
// assertion vacuous.
func answerManaPick(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	var pick *game.PendingChoice
	g.ReadSnapshot(func() {
		for i := range g.PendingChoices {
			if g.PendingChoices[i].Kind == game.PendingChoiceMana {
				pick = g.PendingChoices[i]
			}
		}
	})
	if pick == nil {
		t.Fatal("no mana_pick prompt to answer")
	}
	if err := g.ResolveManaChoice(pick.ID, chooser, pick.ColorOptions[0]); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
}

// basics seeds n untapped basics of one type for a seat — the
// coloured sibling of `mana`, which seeds Islands and so cannot pay a
// {R} or a {G}.
func basics(g *game.Game, p *game.Player, name string, n int) {
	for i := 0; i < n; i++ {
		battlefieldCard(g, p, basic(name, name))
	}
}
