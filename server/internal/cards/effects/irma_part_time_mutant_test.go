package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const irmaPartTimeMutantOracle = "deda54da-4fcc-4da3-9337-7768a2db3283"

func irmaPush(g *game.Game, me uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, me, "Irma, Part-Time Mutant", "Legendary Creature — Human Mutant Shapeshifter", irmaPartTimeMutantOracle, false)
}

// irmaToOwnCombat walks to the beginning of combat of seat 0's next
// turn (this turn's, when it is still ahead).
func irmaToOwnCombat(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 120; i++ {
		if g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepBeginCombat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatal("never reached seat 0's beginning of combat")
}

func TestIrmaBecomesACopyKeepsHerNameAndAbilityAndGetsACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	irma := irmaPush(g, me.ID)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(bear, game.CounterPlusOne, 2); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})

	irmaToOwnCombat(t, g)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)

	c := gaLayered(t, g, irma)
	if c.Name != "Irma, Part-Time Mutant" {
		t.Errorf("name %q, want Irma, Part-Time Mutant", c.Name)
	}
	if c.HasSupertype("Legendary") || !c.HasSubtype("Test") {
		t.Errorf("type line %q, want the Bear's (not legendary)", c.TypeLine)
	}
	// The Bear's printed 2/2, not its counters, plus her own counter.
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("Irma is %d/%d, want 3/3 (a 2/2 copy with one +1/+1 counter)", c.CurrentPower(), c.CurrentToughness())
	}

	// The copy lasts, and "she has this ability": next turn she copies
	// again and gets a second counter.
	ogre := pushVanillaCreature(g, me.ID, "Ogre", 3, 3)
	advanceToNextSeatsTurn(t, g)
	irmaToOwnCombat(t, g)
	if c := gaLayered(t, g, irma); c.Name != "Irma, Part-Time Mutant" || c.CurrentPower() != 3 {
		t.Fatalf("before the second trigger Irma is %q %d/%d, want the lasting Bear copy", c.Name, c.CurrentPower(), c.CurrentToughness())
	}
	answerPickTarget(t, g, ogre)
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, irma); c.Name != "Irma, Part-Time Mutant" || c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Errorf("after copying the Ogre Irma is %q %d/%d, want Irma 5/5", c.Name, c.CurrentPower(), c.CurrentToughness())
	}
}

func TestIrmaWithNoTargetStillGetsTheCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	irma := irmaPush(g, me.ID)
	pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	irmaToOwnCombat(t, g)
	b26DeclinePickTarget(t, g, me.ID)
	passPriorityAroundTable(t, g)

	c := gaLayered(t, g, irma)
	if c.Name != "Irma, Part-Time Mutant" || !c.HasSupertype("Legendary") {
		t.Errorf("Irma is %q %q, want herself", c.Name, c.TypeLine)
	}
	if n := counterCount(g, irma, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters %d, want 1", n)
	}
}

// The ruling: a chosen target that is illegal as the ability resolves
// stops the whole ability, the counter included.
func TestIrmaWithAnIllegalTargetDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	irma := irmaPush(g, me.ID)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	irmaToOwnCombat(t, g)
	answerPickTarget(t, g, bear)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, bear); err != nil {
		t.Fatalf("move to graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)

	c := gaLayered(t, g, irma)
	if c.Name != "Irma, Part-Time Mutant" || !c.HasSupertype("Legendary") {
		t.Errorf("Irma is %q %q, want herself", c.Name, c.TypeLine)
	}
	if n := counterCount(g, irma, game.CounterPlusOne); n != 0 {
		t.Errorf("+1/+1 counters %d, want none", n)
	}
}

func TestIrmaCannotTargetHerselfOrAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	irma := irmaPush(g, me.ID)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	irmaToOwnCombat(t, g)
	for _, bad := range []uuid.UUID{theirs, irma} {
		ch := pickTargetPrompt(t, g)
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Errorf("Irma's trigger accepted %s", bad)
		}
	}
	answerPickTarget(t, g, mine)
}

func TestIrmaDoesNotTriggerOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	irma := irmaPush(g, me.ID)
	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	if latestPickTarget(g, me.ID) != nil || triggersOnStackFrom(g, irma) != 0 || len(g.PendingTriggers) != 0 {
		t.Fatal("Irma triggered at the beginning of an opponent's combat")
	}
}
