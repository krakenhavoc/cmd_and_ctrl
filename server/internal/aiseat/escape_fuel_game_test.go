package aiseat_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// escape_fuel_game_test.go is the whole-game half of #1013: a bot with
// an escape creature in its graveyard CHOOSES which cards the cost
// eats, and chooses the lands.
//
// flashback_game_test.go next door proves the chain — the enumerator
// walks the graveyard, the policy prices the cast, the dispatcher
// applies it — and its own fixture had to work AROUND the gap this
// closes: "the escape fuel goes in FIRST: the enumerator's single
// payment takes the oldest cards in the graveyard, and a fixture that
// let it eat its own test subjects would be asserting about whichever
// one survived". That comment is the bug, written down as a fixture
// constraint.
//
// What can only be checked at this level is the whole loop: the
// heuristic prices the graveyard (fuel.go), the runner hands that price
// to the enumerator (CostFuelPricer → Options.OrderCostFuel), the
// enumerator sorts the pool and offers a few payments, and the policy
// picks one — and the one it picks spends the lands.

const oracleUroBot = "ee302659-59ed-4eef-babe-451b9ccf7f14"

// altCostIDsOfCast reads the cards a cast move pays to its alternative
// cost off the wire payload, the way a policy would.
func altCostIDsOfCast(m legal.Move) []string {
	var p struct {
		AltCostIDs []string `json:"alt_cost_ids"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		return nil
	}
	return p.AltCostIDs
}

// TestABotEscapingUroEatsTheLandsNotTheSpells.
//
// The graveyard holds five Forests and three spells, and Uro exiles
// five OTHER cards. There is exactly one payment that spends no spell,
// and before #1013 the enumerator offered whichever five cards were
// oldest — so this fixture puts the SPELLS in first, which is the
// arrangement that used to eat them.
func TestABotEscapingUroEatsTheLandsNotTheSpells(t *testing.T) {
	requireGameTests(t)

	room := newGraveyardRoom(t, 10130)
	g := room.Game
	bot := g.Seats[0]

	// The spells go in FIRST, so zone order names them first and the
	// pre-#1013 enumerator would have exiled three of them.
	spells := map[uuid.UUID]string{}
	for _, name := range []string{"Ancestral Vision", "Snapcaster Bait", "Second Thoughts"} {
		id := seedGraveyard(g, bot, game.Card{Name: name, TypeLine: "Instant", ManaCost: "{1}{U}"})
		spells[id] = name
	}
	lands := map[uuid.UUID]bool{}
	for i := 0; i < 5; i++ {
		lands[seedGraveyard(g, bot, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})] = true
	}
	uro := seedGraveyard(g, bot, game.Card{
		Name: "Uro, Titan of Nature's Wrath", TypeLine: "Legendary Creature — Elder Giant",
		Power: 6, Toughness: 6, ManaCost: "{1}{G}{U}", OracleID: oracleUroBot,
	})
	// Escape—{G}{G}{U}{U}: two green, two blue, and a couple spare so
	// the cast is never gated on the mana.
	for i := 0; i < 3; i++ {
		seedBattlefield(g, bot, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		seedBattlefield(g, bot, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	}

	watch := newCastWatcher(heuristic.New())
	res := playGameIn(t, room, 10130, []aiseat.Policy{watch, passPolicy{}}, 6, 120*time.Second)

	if watch.offeredFor(uro) == 0 {
		t.Fatalf("Uro was never offered out of the graveyard in %d turns — "+
			"the rest of this test proves nothing", res.turns)
	}
	paid := watch.altCostsPaidFor(uro)
	if len(paid) == 0 {
		t.Fatalf("the bot was offered Uro's escape %d times across %d turns and never took it",
			watch.offeredFor(uro), res.turns)
	}
	for _, payment := range paid {
		if len(payment) != 5 {
			t.Errorf("an escape payment named %d cards, the cost is five: %v", len(payment), payment)
		}
		for _, raw := range payment {
			id, err := uuid.Parse(raw)
			if err != nil {
				t.Fatalf("payment names %q, which is not an instance id", raw)
			}
			if name, isSpell := spells[id]; isSpell {
				t.Errorf("the escape ate %s — five Forests in the same graveyard were cheaper "+
					"fuel, and the whole of #1013 is that the bot can tell (%v)", name, payment)
			}
			if !lands[id] && id != uro {
				t.Errorf("the escape named %s, which is neither a Forest in the graveyard "+
					"nor Uro itself", raw)
			}
		}
	}
}
