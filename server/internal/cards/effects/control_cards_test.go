package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// control_cards_test.go pins the four gain-and-exchange-of-control
// cards ADR 0063 shipped (#756). The duration model's own card, Mass
// Diminish, is in mass_diminish_test.go.

const (
	actOfTreasonOracle      = "9d08af23-9f4a-4097-9abc-3b17475ab744"
	agentOfTreacheryOracle  = "4cd82c7a-86df-4d29-b578-c009208c0d4b"
	sowerOfTemptationOracle = "5881aac2-4f31-45cb-bdb5-64a29ec23316"
	switcherooOracle        = "68227969-39cf-42cf-b3b8-cf8a04647d7e"
)

// --- Act of Treason (until end of turn) -------------------------

func TestActOfTreasonStealsUntapsAndHastes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := ctrlPushCreature(g, opp.ID, "Bear")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == victim {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})

	castCatalogSpell(t, g, "Act of Treason", "Sorcery", actOfTreasonOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)

	if got := controllerOf(t, g, victim); got != me.ID {
		t.Errorf("controller %s, want the caster %s", got, me.ID)
	}
	var tapped, sick bool
	var abilities []string
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == victim {
				tapped = c.Tapped
				sick = c.SummonedThisTurn
				abilities = c.Effective().Abilities
			}
		}
	})
	if tapped {
		t.Error("Act of Treason did not untap the creature it stole")
	}
	if !eotHasAbility(abilities, "haste") {
		t.Errorf("effective abilities %v do not include haste", abilities)
	}
	// CR 302.6 — the steal DOES cost a turn of sickness (the flag is
	// set), and the haste grant is what makes the creature usable
	// anyway. Both halves have to be true or the card is a lie in one
	// direction or the other.
	if !sick {
		t.Error("a stolen creature should be marked summoning-sick under its new controller")
	}
	if summoningSickOf(t, g, victim) {
		t.Error("the haste grant did not lift the sickness — Act of Treason cannot attack")
	}
}

func TestActOfTreasonGivesTheCreatureBackAtCleanup(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := g.Seats[0], g.Seats[1]
	victim := ctrlPushCreature(g, opp.ID, "Bear")

	castCatalogSpell(t, g, "Act of Treason", "Sorcery", actOfTreasonOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, victim); got == opp.ID {
		t.Fatalf("setup: the theft did not happen")
	}

	advancePastCleanupForTest(t, g)

	if got := controllerOf(t, g, victim); got != opp.ID {
		t.Errorf("after cleanup: controller %s, want the owner %s", got, opp.ID)
	}
	var abilities []string
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == victim {
				abilities = c.Effective().Abilities
			}
		}
	})
	if eotHasAbility(abilities, "haste") {
		t.Error("the haste grant outlived the turn")
	}
}

// --- Agent of Treachery (no stated duration) --------------------

func TestAgentOfTreacheryKeepsThePermanentAfterItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := ctrlPushCreature(g, opp.ID, "Bear")

	agent := castAndResolveCreature(t, g, "Agent of Treachery",
		"Creature — Human Rogue", agentOfTreacheryOracle)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)

	if got := controllerOf(t, g, victim); got != me.ID {
		t.Fatalf("controller %s after the Agent entered, want %s", got, me.ID)
	}

	// CR 611.2a: no stated duration means until the game ends. The
	// Agent dying changes nothing — that is the whole difference
	// between this and Sower of Temptation.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(agent) })
	if got := controllerOf(t, g, victim); got != me.ID {
		t.Errorf("the Agent died and the permanent went home: controller %s, want %s", got, me.ID)
	}

	advancePastCleanupForTest(t, g)
	advancePastCleanupForTest(t, g)
	if got := controllerOf(t, g, victim); got != me.ID {
		t.Errorf("the theft expired on a turn boundary: controller %s, want %s", got, me.ID)
	}
}

// TestAgentOfTreacheryDrawsOnThreeStolenPermanents pins the card's own
// payoff: an intervening-if (CR 603.4) counting the permanents you
// control but do not own, read off the board rather than off a record
// of what any effect took.
func TestAgentOfTreacheryDrawsOnThreeStolenPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stolen := []uuid.UUID{
		ctrlPushCreature(g, opp.ID, "Bear One"),
		ctrlPushCreature(g, opp.ID, "Bear Two"),
		ctrlPushCreature(g, opp.ID, "Bear Three"),
	}
	castAndResolveCreature(t, g, "Agent of Treachery",
		"Creature — Human Rogue", agentOfTreacheryOracle)
	pickCard(t, g, me.ID, stolen[0])
	passPriorityAroundTable(t, g)

	// The Agent takes one; the other two come across by hand, which is
	// the point — the trigger counts the board, not its own work.
	g.WithWriteLock(func() {
		for _, id := range stolen[1:] {
			g.GainControlForEffect(uuid.New(), id, me.ID,
				game.IndefiniteDuration(), "test — second-hand theft")
		}
	})
	if got := controllerOf(t, g, stolen[2]); got != me.ID {
		t.Fatalf("setup: controller %s, want %s", got, me.ID)
	}

	before := me.Hand.Size()
	for i := 0; i < 40 && g.Turn.Step != game.StepEnd; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 3 {
		t.Errorf("drew %d cards at the end step, want 3", got)
	}
}

// --- Sower of Temptation (for as long as ~ remains) -------------

func TestSowerOfTemptationGivesTheCreatureBackWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := ctrlPushCreature(g, opp.ID, "Bear")

	sower := castAndResolveCreature(t, g, "Sower of Temptation",
		"Creature — Faerie Wizard", sowerOfTemptationOracle)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)

	if got := controllerOf(t, g, victim); got != me.ID {
		t.Fatalf("controller %s after the Sower entered, want %s", got, me.ID)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(sower) })
	if got := controllerOf(t, g, victim); got != opp.ID {
		t.Errorf("the Sower died and the creature stayed stolen: controller %s, want %s", got, opp.ID)
	}
}

func TestSowerOfTemptationSurvivesTheTurnItEnteredOn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := ctrlPushCreature(g, opp.ID, "Bear")

	castAndResolveCreature(t, g, "Sower of Temptation",
		"Creature — Faerie Wizard", sowerOfTemptationOracle)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, victim); got != me.ID {
		t.Fatalf("setup: controller %s, want %s", got, me.ID)
	}

	// "For as long as" is not "until end of turn": the cleanup sweep
	// must leave it alone.
	advancePastCleanupForTest(t, g)
	if got := controllerOf(t, g, victim); got != me.ID {
		t.Errorf("the cleanup step ended a for-as-long-as effect: controller %s, want %s", got, me.ID)
	}
}

// --- Switcheroo (exchange) --------------------------------------

func TestSwitcherooExchangesControlBothWays(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := ctrlPushCreature(g, me.ID, "My Bear")
	theirs := ctrlPushCreature(g, opp.ID, "Their Ox")

	castCatalogSpell(t, g, "Switcheroo", "Sorcery", switcherooOracle,
		[]game.TargetRef{
			{Kind: game.TargetCard, ID: mine},
			{Kind: game.TargetCard, ID: theirs},
		})
	passPriorityAroundTable(t, g)

	if got := controllerOf(t, g, mine); got != opp.ID {
		t.Errorf("my creature: controller %s, want %s", got, opp.ID)
	}
	if got := controllerOf(t, g, theirs); got != me.ID {
		t.Errorf("their creature: controller %s, want %s", got, me.ID)
	}
	// Ownership is untouched (CR 108.3).
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == mine && c.Owner != me.ID {
				t.Errorf("owner of my creature changed to %s", c.Owner)
			}
			if c.InstanceID == theirs && c.Owner != opp.ID {
				t.Errorf("owner of their creature changed to %s", c.Owner)
			}
		}
	})
}

func TestSwitcherooDoesNothingWhenOneCreatureIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := ctrlPushCreature(g, me.ID, "My Bear")
	theirs := ctrlPushCreature(g, opp.ID, "Their Ox")

	castCatalogSpell(t, g, "Switcheroo", "Sorcery", switcherooOracle,
		[]game.TargetRef{
			{Kind: game.TargetCard, ID: mine},
			{Kind: game.TargetCard, ID: theirs},
		})
	// In response, their creature dies. CR 701.12b: all or nothing.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	passPriorityAroundTable(t, g)

	if got := controllerOf(t, g, mine); got != me.ID {
		t.Errorf("my creature changed hands in a half-exchange: controller %s, want %s", got, me.ID)
	}
}
