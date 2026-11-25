package testscr

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// discard_same_name_test.go — #6264: Sphinx of the Chimes' "Discard two
// nonland cards with the same name: Draw four cards.", on a test oracle
// whose ability gains 1 life.
//
//	201.2a Two or more objects have the same name if they have at least
//	       one name in common … An object with no name doesn't have the
//	       same name as any other object.
//	601.2h … Unpayable costs can't be paid.

const testChimesOracle = "test-6264-chimes"

func registerChimes(t *testing.T) {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:     testChimesOracle,
		Name:         "Test Chimes",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:  "Discard two nonland cards with the same name: You gain 1 life.",
			Cost:   effects.DiscardNWithTheSameName(2, "two nonland cards with the same name", func(c game.Card) bool { return !c.IsLand() }),
			Effect: effects.Do(effects.GainLife{Amount: 1}),
		}},
	})
}

func chimesHand(g *game.Game, me *game.Player, names ...string) []uuid.UUID {
	// The harness deals seven same-named fillers; start from an empty hand.
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	var ids []uuid.UUID
	for _, n := range names {
		typ := "Instant"
		if n == "Forest" {
			typ = "Basic Land — Forest"
		}
		c := game.Card{InstanceID: uuid.New(), Name: n, TypeLine: typ, Owner: me.ID, Controller: me.ID}
		me.Hand.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

func TestDiscardTwoCardsWithTheSameName(t *testing.T) {
	registerChimes(t)
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	chimes := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Chimes", TypeLine: "Creature — Sphinx",
		OracleID: testChimesOracle, Power: 5, Toughness: 6, Owner: me.ID, Controller: me.ID})
	h := chimesHand(g, me, "Opt", "Shock", "Opt", "Forest", "Forest")
	opt1, shock, opt2, forest1, forest2 := h[0], h[1], h[2], h[3], h[4]

	activate := func(ids ...uuid.UUID) error {
		return g.ActivateCatalogAbility(me.ID, chimes, 0, game.ActivateAbilityParams{DiscardIDs: ids, Strict: true})
	}
	if err := activate(opt1, shock); err == nil {
		t.Fatal("two cards with different names paid \"with the same name\"")
	}
	if err := activate(forest1, forest2); err == nil {
		t.Fatal("two land cards paid \"two nonland cards\"")
	}

	// The enumerator offers the one payment there is: the two Opts.
	var offered []string
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeActivateAbility {
			continue
		}
		var p struct {
			CardID     string   `json:"source_card_id"`
			DiscardIDs []string `json:"discard_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err == nil && p.CardID == chimes.String() {
			offered = p.DiscardIDs
		}
	}
	if len(offered) != 2 || !((offered[0] == opt1.String() && offered[1] == opt2.String()) || (offered[0] == opt2.String() && offered[1] == opt1.String())) {
		t.Fatalf("enumerated payment %v, want the two Opts", offered)
	}

	life := me.Life
	if err := activate(opt1, opt2); err != nil {
		t.Fatalf("two Opts: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Fatalf("life = %d, want %d", me.Life, life+1)
	}
}

// No nonland name held twice: no activation is offered.
func TestDiscardSameNameIsUnpayableWithoutAPair(t *testing.T) {
	registerChimes(t)
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	chimes := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Chimes", TypeLine: "Creature — Sphinx",
		OracleID: testChimesOracle, Power: 5, Toughness: 6, Owner: me.ID, Controller: me.ID})
	chimesHand(g, me, "Opt", "Shock", "Forest", "Forest")
	for _, m := range legal.EnumerateFor(g, me.ID) {
		var p struct {
			CardID string `json:"source_card_id"`
		}
		if m.Type == legal.TypeActivateAbility && json.Unmarshal(m.Params, &p) == nil && p.CardID == chimes.String() {
			t.Fatalf("an activation was offered with no nonland pair in hand: %s", m.Label)
		}
	}
}
