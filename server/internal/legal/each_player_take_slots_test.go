package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// each_player_take_slots_test.go — #1743, the bot's half of Explore
// the Vastlands. "May reveal a land card and/or an instant or sorcery
// card" is one choose_cards prompt per player with a rule about the
// picked SET (one card per slot) riding the unexported continuation
// frame. The enumerator must offer exactly the sets that rule accepts
// — the singles and the land-plus-spell pairs, never two of one kind —
// and every one it offers must be accepted.

// stampTop retypes the top cards of p's library, top card first, and
// returns them in that order.
func stampTop(t *testing.T, p *game.Player, typeLines ...string) []uuid.UUID {
	t.Helper()
	if len(p.Library.Cards) < len(typeLines) {
		t.Fatalf("setup: library has %d cards, need %d", len(p.Library.Cards), len(typeLines))
	}
	out := make([]uuid.UUID, 0, len(typeLines))
	for i, tl := range typeLines {
		c := &p.Library.Cards[len(p.Library.Cards)-1-i]
		c.TypeLine = tl
		out = append(out, c.InstanceID)
	}
	return out
}

func TestEachPlayerTakeSlotsEnumeratesOnlyAcceptedSets(t *testing.T) {
	g := newTable(t)
	me, them := g.Seats[0], g.Seats[1]
	mine := stampTop(t, me, "Land", "Land", "Instant", "Sorcery", "Creature — Bear")
	landA, landB, instant, sorcery := mine[0], mine[1], mine[2], mine[3]
	stampTop(t, them, "Land", "Land", "Land", "Creature — Bear", "Creature — Bear")

	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Controller: me.ID})
		err := effects.EachPlayerTakesFromLibrary{
			Players: []uuid.UUID{me.ID, them.ID},
			N:       5,
			Take: effects.TakeFromLibraryToHand{
				Slots: []effects.TakeSlot{
					{Label: "a land card", Match: effects.Land(), Max: 1},
					{Label: "an instant or sorcery card", Match: effects.Or(effects.Instant(), effects.Sorcery()), Max: 1},
				},
				Optional: true,
				Reveal:   true,
				Label:    "test — you may reveal a land card and/or an instant or sorcery card",
				Then:     effects.TakeRestOnBottomInRandomOrder,
			},
		}.Apply(ctx)
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	moves := legal.EnumerateFor(g, me.ID)
	// Nothing, four singles, and the four land-plus-spell pairs. Two
	// lands and two spells break the rule.
	if len(moves) != 9 {
		t.Fatalf("enumerated %d answers, want 9: %v", len(moves), labels(moves))
	}
	lands := map[uuid.UUID]bool{landA: true, landB: true}
	spells := map[uuid.UUID]bool{instant: true, sorcery: true}
	pairs := 0
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("a seat owing a choice was offered %q", m.Label)
			continue
		}
		ids := pickedCardIDs(t, m)
		if len(ids) == 0 && !m.AlwaysLegal {
			t.Error("\"choose nothing\" is this kind's AlwaysLegal answer")
		}
		if len(ids) == 2 {
			pairs++
			if lands[ids[0]] == lands[ids[1]] || spells[ids[0]] == spells[ids[1]] {
				t.Errorf("offered two of one kind: %q", m.Label)
			}
		}
	}
	if pairs != 4 {
		t.Errorf("offered %d pairs, want the 4 land-plus-spell pairs", pairs)
	}
	dispatchAll(t, g, me.ID, moves)

	// Three lands and no spell: one slot can be filled, so the other
	// seat's answers are nothing and the three singles.
	theirs := legal.EnumerateFor(g, them.ID)
	if len(theirs) != 4 {
		t.Fatalf("enumerated %d answers for the three-land look, want 4: %v", len(theirs), labels(theirs))
	}
	dispatchAll(t, g, them.ID, theirs)
}
