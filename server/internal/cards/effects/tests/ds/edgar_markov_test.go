package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// edgar_markov_test.go — #2802: an eminence TRIGGER, proved on Edgar
// Markov. "Whenever you cast another Vampire spell, if Edgar is in the
// command zone or on the battlefield, create a 1/1 black Vampire
// creature token."

const edgarMarkovOracle = "41e2790d-49f5-4e98-b8d9-04179f47f13a"

func edgarCard(p *game.Player) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: "Edgar Markov", TypeLine: "Legendary Creature — Vampire Knight",
		ManaCost: "{3}{R}{W}{B}", OracleID: edgarMarkovOracle, IsCommander: true,
		Owner: p.ID, Controller: p.ID, Power: 4, Toughness: 4,
	}
}

func vampireTokens(g *game.Game) int {
	return len(effectstest.BattlefieldIDsNamed(g, "Vampire"))
}

func TestEdgarMarkovIsRegisteredWithAnEminenceTrigger(t *testing.T) {
	spec, ok := effects.Lookup(edgarMarkovOracle)
	if !ok || spec.Completeness != effects.CompletenessFull {
		t.Fatal("Edgar Markov is registered complete")
	}
	z := game.TriggerZones(spec.Triggered[0])
	if len(z) != 2 || !game.TriggerWatchesFromZone(spec.Triggered[0], game.ZoneCommand) ||
		!game.TriggerWatchesFromZone(spec.Triggered[0], game.ZoneBattlefield) {
		t.Errorf("the eminence trigger watches from %v, want the battlefield and the command zone", z)
	}
	if game.TriggerWatchesFromZone(spec.Triggered[1], game.ZoneCommand) {
		t.Error("the attack trigger works from the command zone")
	}
}

// From the command zone: another Vampire spell its owner casts makes a
// token; a non-Vampire spell does not.
func TestEdgarMarkovMakesAVampireFromTheCommandZone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Command.PushTop(edgarCard(me))

	effectstest.CastCatalogSpell(t, g, "Vampire Bat", "Creature — Vampire Bat", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := vampireTokens(g); n != 1 {
		t.Fatalf("%d Vampire tokens after casting a Vampire, want 1", n)
	}
	if id := effectstest.BattlefieldIDsNamed(g, "Vampire")[0]; effectstest.ControllerOf(t, g, id) != me.ID {
		t.Error("the token is not Edgar's owner's")
	}

	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := vampireTokens(g); n != 1 {
		t.Errorf("a non-Vampire spell made a token: %d", n)
	}
}

// "You": Edgar in an opponent's command zone does nothing for my
// Vampire. Edgar in a graveyard does nothing at all (CR 113.6).
func TestEdgarMarkovWorksOnlyForItsOwnerAndFromItsZones(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp.Command.PushTop(edgarCard(opp))
	me.Graveyard.PushTop(edgarCard(me))

	effectstest.CastCatalogSpell(t, g, "Vampire Bat", "Creature — Vampire Bat", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := vampireTokens(g); n != 0 {
		t.Errorf("%d Vampire tokens, want none", n)
	}
}

// CR 603.4: the "if" is checked again on resolution. Edgar leaving the
// command zone with the trigger on the stack makes it do nothing.
func TestEdgarMarkovChecksItsZoneOnResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	edgar := edgarCard(me)
	me.Command.PushTop(edgar)

	effectstest.CastCatalogSpell(t, g, "Vampire Bat", "Creature — Vampire Bat", "", nil)
	queued := 0
	g.ReadSnapshot(func() {
		for _, it := range g.PendingTriggers {
			if it != nil && it.SourceCardID == edgar.InstanceID {
				queued++
			}
		}
		for _, it := range g.StackMeta {
			if it != nil && it.SourceCardID == edgar.InstanceID {
				queued++
			}
		}
	})
	if queued != 1 {
		t.Fatalf("Edgar's trigger queued %d times, want 1", queued)
	}
	g.WithWriteLock(func() {
		me.Command.Cards = nil
		me.Graveyard.PushTop(edgar)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := vampireTokens(g); n != 0 {
		t.Errorf("%d Vampire tokens after Edgar left the command zone, want none", n)
	}
}

// On the battlefield the eminence still works, and Edgar's attack puts
// a +1/+1 counter on each Vampire its controller controls.
func TestEdgarMarkovOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	edgar := effectstest.B12Push(g, me.ID, "Edgar Markov", "Legendary Creature — Vampire Knight", edgarMarkovOracle, 4, 4)
	vamp := effectstest.B12Push(g, me.ID, "Vampire Bat", "Creature — Vampire Bat", "", 1, 1)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Vampire", "Creature — Vampire", "", 1, 1)

	effectstest.DeclareAttack(t, g, opp.ID, edgar)
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{edgar: 1, vamp: 1, bear: 0, theirs: 0} {
		c, _ := effectstest.BattlefieldCard(g, id)
		if got := c.Counters[game.CounterPlusOne]; got != want {
			t.Errorf("%s has %d +1/+1 counters, want %d", c.Name, got, want)
		}
	}
}
