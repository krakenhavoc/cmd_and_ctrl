package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// phyrexian_view_test.go — the wire half of CR 107.4's "or 2 life"
// (#916, #917).
//
// The client decides locally whether to open the stepper, and what
// its ceiling is. Both answers come from the server as a COUNT
// (`phyrexian_symbols`) rather than from a second parse of the mana
// string on the client, which is the whole point: #787 was a symbol
// family one parser had never heard of, and a client re-deriving the
// count would be a second parser to keep in step.

func TestHandCardShipsThePhyrexianSymbolCount(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	seen := map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true}

	for _, tc := range []struct {
		name string
		cost string
		want int
	}{
		{"Gitaxian Probe", "{U/P}", 1},
		{"Dismember", "{1}{B/P}{B/P}", 2},
		{"Compleated Sage", "{2}{G}{G/U/P}{U}", 1},
		{"Lightning Bolt", "{R}", 0},
		{"Land", "", 0},
	} {
		c := game.NewCard(tc.name, me.ID)
		c.TypeLine = "Sorcery"
		c.ManaCost = tc.cost
		c.KnownBy = seen
		me.Hand.PushTop(c)

		v := ViewOfGameFor(g, me.ID.String())
		var got *CardView
		for i := range v.Seats[0].Hand.Cards {
			if v.Seats[0].Hand.Cards[i].InstanceID == c.InstanceID.String() {
				got = &v.Seats[0].Hand.Cards[i]
			}
		}
		if got == nil {
			t.Fatalf("%s: hand view missing the seeded card", tc.name)
		}
		if got.PhyrexianSymbols != tc.want {
			t.Errorf("%s (%s): phyrexian_symbols = %d, want %d",
				tc.name, tc.cost, got.PhyrexianSymbols, tc.want)
		}
	}
}

// TestActivatedAbilityViewShipsThePhyrexianSymbolCount — Birthing
// Pod's "{1}{G/P}" is one, Solphim's "{1}{R/P}{R/P}" is two, and an
// ordinary cost is none. The client's stepper is bounded by this and
// by the activator's life total, and sends the answer back as
// `phyrexian_life`.
func TestActivatedAbilityViewShipsThePhyrexianSymbolCount(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Test Pod",
		TypeLine:   "Artifact",
		OracleID:   "00000000-0000-0000-0000-0000000000cc",
		Owner:      owner,
		Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{
			{Label: "{1}{G/P}: pod", Cost: game.AbilityCost{Mana: "{1}{G/P}"}},
			{Label: "{1}{R/P}{R/P}: double", Cost: game.AbilityCost{Mana: "{1}{R/P}{R/P}"}},
			{Label: "{2}: plain", Cost: game.AbilityCost{Mana: "{2}"}},
			{Label: "{T}: free", Cost: game.AbilityCost{Tap: true}},
		},
	})

	var c CardView
	for _, v := range ViewOfGame(g).Battlefield.Cards {
		if v.InstanceID == id.String() {
			c = v
		}
	}
	if len(c.ActivatedAbilities) != 4 {
		t.Fatalf("got %d abilities on the wire, want 4", len(c.ActivatedAbilities))
	}
	want := []int{1, 2, 0, 0}
	for i, n := range want {
		if got := c.ActivatedAbilities[i].PhyrexianSymbols; got != n {
			t.Errorf("ability %d (%q): phyrexian_symbols = %d, want %d",
				i, c.ActivatedAbilities[i].Label, got, n)
		}
	}
}

// TestAlternativeCostShipsItsOwnPhyrexianSymbolCount — claiming an
// offer replaces the mana cost, so it replaces the ceiling on the
// claim. A client that read the printed count while paying an
// alternative cost would offer a stepper the announce gate rejects.
func TestAlternativeCostShipsItsOwnPhyrexianSymbolCount(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-view-phyrexian-alt"

	prevAlts := game.CatalogAlternativeCosts
	game.CatalogAlternativeCosts = func(id string) []game.AlternativeCost {
		if id != oracle {
			return nil
		}
		return []game.AlternativeCost{
			{Key: "compleated", Label: "Pay {W/P}{W/P}", ManaCost: "{W/P}{W/P}"},
			{Key: "flat", Label: "Pay {R}", ManaCost: "{R}"},
		}
	}
	t.Cleanup(func() { game.CatalogAlternativeCosts = prevAlts })

	c := game.NewCard("Test Offer", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{U/P}"
	c.OracleID = oracle
	c.KnownBy = map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true}
	me.Hand.PushTop(c)

	v := ViewOfGameFor(g, me.ID.String())
	var card *CardView
	for i := range v.Seats[0].Hand.Cards {
		if v.Seats[0].Hand.Cards[i].InstanceID == c.InstanceID.String() {
			card = &v.Seats[0].Hand.Cards[i]
		}
	}
	if card == nil {
		t.Fatalf("hand view missing the seeded card")
	}
	if card.PhyrexianSymbols != 1 {
		t.Errorf("printed cost: phyrexian_symbols = %d, want 1", card.PhyrexianSymbols)
	}
	want := map[string]int{"compleated": 2, "flat": 0}
	for _, offer := range card.AlternativeCosts {
		if got := offer.PhyrexianSymbols; got != want[offer.Key] {
			t.Errorf("offer %q: phyrexian_symbols = %d, want %d", offer.Key, got, want[offer.Key])
		}
	}
}

// TestRevealedOpponentHandCardDropsThePhyrexianSymbolCount — the
// count is stamped for the OWNER's cost prompts, with the other cast
// clauses, and is stripped with them from an opponent's revealed hand
// card (keepKnownInHandZone). Nobody but the caster opens that
// stepper, so nobody but the caster is told its ceiling. The
// hidden-card path is covered by the redaction allowlist in
// face_down_view_test.go.
func TestRevealedOpponentHandCardDropsThePhyrexianSymbolCount(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	c := game.NewCard("Dismember", them.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{1}{B/P}{B/P}"
	c.KnownBy = map[uuid.UUID]bool{them.ID: true, me.ID: true}
	them.Hand.PushTop(c)

	// Their own view keeps it: it is their prompt.
	mine := ViewOfGameFor(g, them.ID.String())
	var own *CardView
	for i := range mine.Seats[1].Hand.Cards {
		if mine.Seats[1].Hand.Cards[i].InstanceID == c.InstanceID.String() {
			own = &mine.Seats[1].Hand.Cards[i]
		}
	}
	if own == nil {
		t.Fatalf("the owner's own hand view is missing the seeded card")
	}
	if own.PhyrexianSymbols != 2 {
		t.Fatalf("the owner's copy: phyrexian_symbols = %d, want 2", own.PhyrexianSymbols)
	}

	v := ViewOfGameFor(g, me.ID.String())
	seen := false
	for _, card := range v.Seats[1].Hand.Cards {
		if card.InstanceID != c.InstanceID.String() {
			continue
		}
		seen = true
		if card.PhyrexianSymbols != 0 {
			t.Errorf("an opponent's revealed hand card kept phyrexian_symbols = %d", card.PhyrexianSymbols)
		}
	}
	if !seen {
		t.Fatalf("the revealed card is not in the opponent's view; this test would pass vacuously")
	}
}
