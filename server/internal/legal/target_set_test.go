package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// target_set_test.go — #1559, the enumerator's half of a rule over the
// chosen SET of targets (CR 601.2c): it never builds a set two of
// whose picks share the rule's key, never offers a card above an
// X-bounded clause's announced X, offers nothing when no legal set
// exists — and every move it offers, the engine accepts. The view
// ships the same keys the enumerator prunes on.

const (
	oracleAgadeem     = "562d71b9-1646-474e-9293-55da6947a758"
	oracleRunAwayPair = "290faa28-450e-4797-9a8f-642d8af3f82a"
)

func castTargetIDs(t *testing.T, m legal.Move) (x int, ids []uuid.UUID) {
	t.Helper()
	var p struct {
		XValue  int `json:"x_value"`
		Targets []struct {
			ID string `json:"id"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", string(m.Params), err)
	}
	for _, tw := range p.Targets {
		ids = append(ids, uuid.MustParse(tw.ID))
	}
	return p.XValue, ids
}

func graveyardCreature(p *game.Player, name, cost string) uuid.UUID {
	c := creature(name, cost, 2, 2)
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

// Agadeem's Awakening off five Swamps is X=2: the enumerator offers
// only sets of DIFFERENT mana values, each 2 or less, and the engine
// accepts every one of them.
func TestAgadeemEnumeratesOnlyLegalSets(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	card := handCard(active, game.Card{
		Name: "Agadeem's Awakening", TypeLine: "Sorcery",
		ManaCost: "{X}{B}{B}{B}", OracleID: oracleAgadeem,
	})
	for i := 0; i < 5; i++ {
		battlefieldCard(g, active, basic("Swamp", "Swamp"))
	}
	mv := map[uuid.UUID]int{
		graveyardCreature(active, "One", "{B}"):      1,
		graveyardCreature(active, "Two A", "{1}{B}"): 2,
		graveyardCreature(active, "Two B", "{B}{B}"): 2,
		graveyardCreature(active, "Three", "{2}{B}"): 3,
	}
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, card)
	if len(casts) == 0 {
		t.Fatalf("Agadeem's Awakening is castable off five Swamps: %v", labels(moves))
	}
	sawPair := false
	for _, m := range casts {
		x, ids := castTargetIDs(t, m)
		seen := map[int]bool{}
		for _, id := range ids {
			if mv[id] > x {
				t.Errorf("%q offers mana value %d at X=%d", m.Label, mv[id], x)
			}
			if seen[mv[id]] {
				t.Errorf("%q offers two cards of mana value %d", m.Label, mv[id])
			}
			seen[mv[id]] = true
		}
		if len(ids) == 2 {
			sawPair = true
		}
	}
	if !sawPair {
		t.Error("a mana value 1 and a mana value 2 card together are a legal set and should be offered")
	}
	dispatchAll(t, g, active.ID, moves)
}

// Run Away Together: with every creature under one controller there is
// no legal PAIR, so there is no cast — the enumerator counts keys, not
// candidates. With a second controller's creature, every offered pair
// straddles two controllers.
func TestRunAwayTogetherEnumeratesOnlyCrossControllerPairs(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	card := handCard(active, game.Card{
		Name: "Run Away Together", TypeLine: "Instant",
		ManaCost: "{1}{U}", OracleID: oracleRunAwayPair,
	})
	for i := 0; i < 2; i++ {
		battlefieldCard(g, active, basic("Island", "Island"))
	}
	battlefieldCard(g, opp, creature("Bear A", "{1}{G}", 2, 2))
	battlefieldCard(g, opp, creature("Bear B", "{1}{G}", 2, 2))
	advanceTo(t, g, game.StepPrecombatMain)

	if casts := castMovesFor(legal.EnumerateFor(g, active.ID), card); len(casts) != 0 {
		t.Fatalf("no legal pair, no cast; offered %v", labels(casts))
	}

	battlefieldCard(g, active, creature("My Bear", "{1}{G}", 2, 2))
	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, card)
	if len(casts) == 0 {
		t.Fatal("a pair across two controllers is castable")
	}
	for _, m := range casts {
		_, ids := castTargetIDs(t, m)
		controllers := map[uuid.UUID]bool{}
		g.WithWriteLock(func() {
			for _, id := range ids {
				c, _ := g.LookupCardForEffect(id)
				controllers[c.Controller] = true
			}
		})
		if len(controllers) != 2 {
			t.Errorf("%q names two creatures of one controller", m.Label)
		}
	}
	dispatchAll(t, g, active.ID, moves)
}

// The view ships the set rule and the X bound with the SAME keys the
// enumerator prunes on: two cards the enumerator never pairs carry one
// key on the wire, and every card carries its mana value for the
// client to apply X.
func TestSetRuleViewAgreesWithTheEnumerator(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	card := handCard(active, game.Card{
		Name: "Agadeem's Awakening", TypeLine: "Sorcery",
		ManaCost: "{X}{B}{B}{B}", OracleID: oracleAgadeem,
	})
	twoA := graveyardCreature(active, "Two A", "{1}{B}")
	twoB := graveyardCreature(active, "Two B", "{B}{B}")
	three := graveyardCreature(active, "Three", "{2}{B}")
	advanceTo(t, g, game.StepPrecombatMain)

	v := protocol.ViewOfGameFor(g, active.ID.String())
	var lt *protocol.LegalTargetsView
	for _, s := range v.Seats {
		for _, c := range s.Hand.Cards {
			if c.InstanceID == card.String() {
				lt = c.LegalTargets
			}
		}
	}
	if lt == nil || lt.Different == nil {
		t.Fatalf("the hand card ships its set rule: %+v", lt)
	}
	if lt.Different.Label == "" {
		t.Error("the rule carries its printed label for the banner")
	}
	keys := lt.Different.Keys
	if keys[twoA.String()] == "" || keys[twoA.String()] != keys[twoB.String()] || keys[twoA.String()] == keys[three.String()] {
		t.Errorf("keys = %v: the two mana-value-2 cards share one, the 3 differs", keys)
	}
	if !lt.ManaValueAtMostX || lt.ManaValues[three.String()] != 3 || lt.ManaValues[twoA.String()] != 2 {
		t.Errorf("the X bound ships with each card's mana value: %+v", lt)
	}
}
