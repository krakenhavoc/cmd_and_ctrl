package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #4977, whole-engine: the real Daring Discovery through the real
// enumerator, view and policy, with PriceCantBlock on (it ships off).
// The bot's two 3/3s face an opponent's two 4/4s in its first main phase,
// with five Mountains and the sorcery in hand. Its target entry reaches the wire as `cant_block`, the bot casts
// it at both 4/4s before attacks (one alone opens nothing: the other
// still blocks), and the two 3/3s then connect for six, damage the 4/4s
// would have blocked at the bot's loss. Two against two keeps the pair
// inside the enumerator's twelve target sets (legal.MaxExpansionPerSource).
func TestHeuristicCastsDaringDiscoveryBeforeItsAttack(t *testing.T) {
	const oracle = "a9564593-b5a6-4a83-be1e-2af8caf647d7"
	g := newSettledTable(t, 4977)
	advanceToStep(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	everyone := seatIDs(g)
	dd := uuid.New()
	walls := map[uuid.UUID]bool{}
	g.WithWriteLock(func() {
		for _, s := range g.Seats {
			s.Hand.Cards = nil
			s.Command.Cards = nil
		}
		me.Hand.Cards = []game.Card{{InstanceID: dd, Name: "Daring Discovery", OracleID: oracle,
			TypeLine: "Sorcery", ManaCost: "{4}{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
			KnownBy: map[uuid.UUID]bool{me.ID: true}}}
		push := func(c game.Card) {
			c.AddKnowersAll(everyone)
			g.Battlefield.PushTop(c)
		}
		for i := 0; i < 5; i++ {
			push(game.Card{InstanceID: uuid.New(), Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
		}
		for i := 0; i < 2; i++ {
			push(game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{R}",
				Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
			w := game.Card{InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall", ManaCost: "{3}{G}",
				Power: 4, Toughness: 4, Owner: them.ID, Controller: them.ID}
			walls[w.InstanceID] = true
			push(w)
		}
	})
	g.BumpLayerVersionForTest()

	view := protocol.ViewOfGameFor(g, me.ID.String())
	declared := false
	for _, c := range view.Seats[g.Turn.ActiveSeat].Hand.Cards {
		if c.InstanceID == dd.String() && c.Purpose != nil && c.Purpose.Targets != nil {
			for _, e := range *c.Purpose.Targets {
				declared = declared || (e.UntilEndOfTurn == "harm" && e.Pump != nil && e.Pump.CantBlock)
			}
		}
	}
	if !declared {
		t.Fatal("Daring Discovery's target entry does not reach the bot's view as a can't-block harm")
	}

	life := them.Life
	cfg := heuristic.DefaultConfig()
	cfg.PriceCantBlock = true // ships off (#4977); this is the price itself
	pol := heuristic.NewWithConfig(cfg)
	made := playHeliodMainPhase(t, g, me, pol)
	cast := false
	for _, m := range made {
		if m.move.Kind != legal.KindCast || m.move.Source != dd {
			continue
		}
		cast = true
		ids := targetIDs(t, m.move)
		if len(ids) != 2 {
			t.Errorf("Daring Discovery was cast at %d creatures, want both walls", len(ids))
		}
		for _, id := range ids {
			if !walls[id] {
				t.Errorf("Daring Discovery was cast at %s, which is not one of the opponent's walls", id)
			}
		}
	}
	if !cast {
		t.Fatalf("the bot did not cast Daring Discovery before attacks; it made %v", labelsOf(made))
	}
	// The discover may leave the bot a choice; answer it, then play the
	// turn out.
	for i := 0; i < 4 && owesAnyChoice(g); i++ {
		driveChoices(t, g, pol, me.ID, 10)
		if g.Turn.Step == game.StepPrecombatMain {
			playHeliodMainPhase(t, g, me, pol)
		}
	}
	playHeliodTurn(t, g, me, pol)
	if lost := life - them.Life; lost < 6 {
		t.Errorf("the opponent lost %d life, want the two 3/3s' six past the walls that could not block", lost)
	}
}
