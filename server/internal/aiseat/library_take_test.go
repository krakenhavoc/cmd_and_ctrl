package aiseat_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_take_test.go — #1831 through the real card. Horn of the Mark
// looks at the top five and asks "you may reveal a creature card from
// among them and put it into your hand": a choose_cards prompt over the
// bot's own library with "take nothing" as its first offer. The
// heuristic used to score every answer the same and so took nothing,
// every time. The unit tests in heuristic/library_take_test.go pin the
// scoring; this pins that the filtered view a real seat receives
// carries what the scoring reads (the candidates as cards of the seat's
// own library) and that the engine accepts the answer.
//
// Not behind AISEAT_GAME_TESTS: one trigger, one prompt, milliseconds.

const botHornOfTheMarkOracle = "9b836c32-84f0-41ae-b7d9-67b92c743c60"

func TestHeuristicSeatTakesTheBestCreatureFromHornOfTheMark(t *testing.T) {
	g := newSettledTable(t, 1831)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	everyone := seatIDs(g)

	public := func(name, typeLine, oracle string, power, tough int) uuid.UUID {
		c := game.Card{
			InstanceID: uuid.New(),
			Name:       name, TypeLine: typeLine, OracleID: oracle,
			Power: power, Toughness: tough,
			Owner: me.ID, Controller: me.ID,
		}
		c.AddKnowersAll(everyone)
		g.Battlefield.PushTop(c)
		return c.InstanceID
	}
	public("Horn of the Mark", "Legendary Artifact", botHornOfTheMarkOracle, 0, 0)
	attackers := []uuid.UUID{
		public("Rider", "Creature — Human Knight", "", 2, 2),
		public("Rider", "Creature — Human Knight", "", 2, 2),
	}

	// The top five, top card last (PushTop). Two creature cards, the
	// Dragon listed after the Bear so that the enumerator's order alone
	// would not pick it.
	top := []game.Card{
		{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"},
		{Name: "Dragon", TypeLine: "Creature — Dragon", ManaCost: "{4}{R}{R}", Power: 6, Toughness: 6},
		{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"},
	}
	var dragon, bear uuid.UUID
	for i := len(top) - 1; i >= 0; i-- {
		c := top[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = me.ID, me.ID
		me.Library.PushTop(c)
		switch c.Name {
		case "Dragon":
			dragon = c.InstanceID
		case "Bear":
			bear = c.InstanceID
		}
	}

	advanceToStep(t, g, game.StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	// Lock the declaration in and resolve the trigger: pass until the
	// Horn's prompt is owed.
	for i := 0; i < 32 && !owesChoice(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				break
			}
			t.Fatalf("PassPriority %d: %v", i, err)
		}
	}
	if !owesChoice(g, me.ID) {
		t.Fatal("two attackers raised no Horn of the Mark prompt")
	}

	taken := driveChoices(t, g, heuristic.New(), me.ID, 8)
	t.Logf("took %v", taken)

	if !me.Hand.Contains(dragon) {
		t.Errorf("the Dragon is not in hand: the heuristic seat should take the best creature card of the five (took %v)", taken)
	}
	if me.Hand.Contains(bear) {
		t.Error("the Bear is in hand; the Horn takes one card and it should be the Dragon")
	}
}
