package effectstest

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// explore_the_vastlands_test.go — #1743 item 3. Explore the Vastlands
// is the back face of Wandering Archaic: "Each player looks at the top
// five cards of their library and may reveal a land card and/or an
// instant or sorcery card from among them. Each player puts the cards
// they revealed this way into their hand and the rest on the bottom of
// their library in a random order. Each player gains 3 life."
//
// Two shapes are under test: TakeFromLibraryToHand.Slots (two
// predicates, one look, one card per slot) and
// EachPlayerTakesFromLibrary (that question asked of every player as
// one instruction). The card is the proof of both.

type VastCard struct{ Name, TypeLine string }

// vastFive: a land, an instant, a sorcery and two creatures — two of
// the five fit the land slot or the spell slot each way round.
var VastFive = []VastCard{
	{"Forest", "Basic Land — Forest"},
	{"Opt", "Instant"},
	{"Grizzly Bears", "Creature — Bear"},
	{"Preordain", "Sorcery"},
	{"Llanowar Elves", "Creature — Elf Druid"},
}

// stackTop puts `cards` on top of p's library, the first one on top,
// and returns their IDs in that (top-first) order.
func StackTop(p *game.Player, cards []VastCard) []uuid.UUID {
	pairs := make([]struct{ Name, TypeLine string }, len(cards))
	for i, c := range cards {
		pairs[i] = struct{ Name, TypeLine string }{c.Name, c.TypeLine}
	}
	return HornTopFive(p, pairs)
}

// castExplore puts Wandering Archaic // Explore the Vastlands into the
// active player's hand, casts its BACK face, and resolves it.
func CastExplore(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	card := MdfcCard(me.ID, WanderingArchaicOracle, game.LayoutModalDFC,
		game.Face{Name: "Wandering Archaic", TypeLine: "Creature — Avatar", ManaCost: "{5}", Power: 4, Toughness: 4},
		game.Face{Name: "Explore the Vastlands", TypeLine: "Sorcery", ManaCost: "{3}"},
	)
	me.Hand.PushTop(card)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Explore the Vastlands: %v", err)
	}
	PassPriorityAroundTable(t, g)
	return card.InstanceID
}

// apnapSeats is the table from the active player round.
func ApnapSeats(g *game.Game) []*game.Player {
	out := make([]*game.Player, 0, len(g.Seats))
	for i := range g.Seats {
		out = append(out, g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)])
	}
	return out
}

func VastRevealed(g *game.Game, id uuid.UUID) bool { return HornRevealed(g, id) }

// libraryBottom returns the bottom n cards of p's library, bottom card
// first (Zone.Cards is bottom-first).
func LibraryBottom(p *game.Player, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for _, c := range p.Library.Cards[:n] {
		out = append(out, c.InstanceID)
	}
	return out
}

func SameSet(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	for _, id := range a {
		if !slices.Contains(b, id) {
			return false
		}
	}
	return true
}

func Without(ids []uuid.UUID, drop ...uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range ids {
		if !slices.Contains(drop, id) {
			out = append(out, id)
		}
	}
	return out
}

// testSlotsTake is the slot rule of TestTakeSlotsACardMatchingBothFillsOneSlot
// as a registered take (#4253 retired the unkeyed take, whose pick held a
// closure).
var TestSlotsTake = effects.RegisterTake("test/4253-take-creature-and-or-artifact", effects.TakeDef{
	Reveal: true,
	Slots: []effects.TakeSlot{
		{Label: "a creature card", Match: effects.Creature(), Max: 1},
		{Label: "an artifact card", Match: effects.Artifact(), Max: 1},
	},
	Then: effects.TakeRestOnBottomInRandomOrder,
})
