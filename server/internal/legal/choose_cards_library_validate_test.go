package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// choose_cards_library_validate_test.go — #998, the bot's half.
//
// A library-top pick can now carry a rule about the picked SET ("any
// number of nonland permanent cards with total mana value 4 or less",
// Ao, the Dawn Sky). The rule rides the unexported continuation frame,
// so an enumerator that offered every subset within the bounds would
// offer answers the resolver refuses — the #544 wedge. It does not:
// every set goes through ChooseCardsPickLegalLocked before it is
// offered, and this test is that path reached through the CATALOG
// primitive rather than through a hand-built prompt, which is the wire
// #998 added.

// libraryTopPermanents stamps the top `costs` cards of a seat's
// library as artifacts with those mana costs and returns them in
// top-first order — the order a look reports them in.
func libraryTopPermanents(t *testing.T, p *game.Player, costs ...string) []uuid.UUID {
	t.Helper()
	if len(p.Library.Cards) < len(costs) {
		t.Fatalf("setup: library has %d cards, need %d", len(p.Library.Cards), len(costs))
	}
	out := make([]uuid.UUID, 0, len(costs))
	for i, cost := range costs {
		c := &p.Library.Cards[len(p.Library.Cards)-1-i]
		c.TypeLine = "Artifact"
		c.ManaCost = cost
		out = append(out, c.InstanceID)
	}
	return out
}

func TestChooseCardsOverLibraryCardsHonoursASetRule(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	top := libraryTopPermanents(t, me, "{1}", "{2}", "{4}")
	one, two, four := top[0], top[1], top[2]

	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Controller: me.ID})
		err := effects.PutFromLibraryOntoBattlefield{
			Player:   me.ID,
			Cards:    g.LookAtTopOfLibraryForEffect(me.ID, 3),
			Max:      0,
			Optional: true,
			Validate: func(picked []game.Card) bool {
				total := 0
				for _, c := range picked {
					total += c.ManaValue()
				}
				return total <= 4
			},
			Label: "test — put any number with total mana value 4 or less onto the battlefield",
			Then:  effects.PutRestOnBottomInRandomOrder,
		}.Apply(ctx)
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	moves := legal.EnumerateFor(g, me.ID)

	// "Choose nothing", the three singles, and the one legal pair
	// ({1} + {2}). {1}+{4}, {2}+{4} and the triple all break the rule.
	if len(moves) != 5 {
		t.Fatalf("enumerated %d answers, want 5: %v", len(moves), labels(moves))
	}
	mv := map[uuid.UUID]int{one: 1, two: 2, four: 4}
	sawEmpty, sawPair := false, false
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("a seat owing a choice was offered %q", m.Label)
			continue
		}
		ids := pickedCardIDs(t, m)
		if len(ids) == 0 {
			sawEmpty = true
			if !m.AlwaysLegal {
				t.Error("\"choose nothing\" is this kind's AlwaysLegal answer")
			}
			continue
		}
		total := 0
		for _, id := range ids {
			total += mv[id]
		}
		if total > 4 {
			t.Errorf("offered a set the resolver would refuse: %q (total %d)", m.Label, total)
		}
		if len(ids) == 2 {
			sawPair = true
			if !(ids[0] == one && ids[1] == two) && !(ids[0] == two && ids[1] == one) {
				t.Errorf("the only legal pair is {1}+{2}, got %q", m.Label)
			}
		}
	}
	if !sawEmpty {
		t.Error("\"choose nothing\" was not offered for a zero-floor prompt")
	}
	if !sawPair {
		t.Error("the legal pair was not reached — filteredCombinations must spread the budget across sizes")
	}

	// Soundness: every move the enumerator offers is one the dispatcher
	// accepts, continuation included.
	dispatchAll(t, g, me.ID, moves)
}
