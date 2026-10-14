package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shield_broker_test.go — #2000: Shield Broker steals a noncommander
// creature for as long as its shield counter lasts, and the counter
// protects it (CR 122.1c).

func TestShieldBrokerStealsUntilTheShieldIsSpent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	commander := b12Creature(g, opp.ID, "Their Commander", "Legendary Creature — Ogre", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == commander {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})

	broker := game.NewCard("Shield Broker", me.ID)
	broker.TypeLine, broker.OracleID = "Creature — Octopus Advisor", "481234e5-fbfd-46e8-8b9a-a4e708dd2195"
	broker.Power, broker.Toughness = 3, 4
	me.Hand.PushTop(broker)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, broker.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: commander}); err == nil {
		t.Error("a commander was accepted as the target")
	}
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)

	if c, ok := battlefieldCard(g, victim); !ok || c.Controller != me.ID || c.Counters[game.CounterShield] != 1 {
		t.Fatalf("after the trigger: %+v, want controlled by me with one shield counter", c)
	}
	b25Destroy(g, victim)
	if _, ok := battlefieldCard(g, victim); !ok {
		t.Fatal("the shield counter didn't stop the destruction")
	}
	if c := layeredCard(t, g, victim); c.Controller != opp.ID {
		t.Errorf("with the shield counter spent, control stays with %v, want it back with its owner", c.Controller)
	}
}
