package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// escape_view_test.go — S29. The escape picker is driven entirely by
// `pay_options` on the graveyard copy of the card, so the three
// things worth pinning are the three the client cannot work out for
// itself:
//
//   - the count. "Exile four other cards" rides as min = max = 4;
//     the client sizes its picker off that and never learns the word
//     "escape".
//   - "OTHER". The spell is sitting in the same graveyard it is
//     being paid out of, so it has to be filtered out of its own
//     picker — the server would reject the cast, but offering the
//     card at all is a trap.
//   - "YOUR graveyard". An opponent's cards must not reach the list,
//     which also makes this an information-leak test.

func TestEscapePayOptionsExcludeTheSpellAndOtherGraveyards(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	const oracle = "test-view-escape"

	prevZones := game.CatalogCastableZones
	game.CatalogCastableZones = func(id string) []game.ZoneKind {
		if id == oracle {
			return []game.ZoneKind{game.ZoneGraveyard}
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogCastableZones = prevZones })

	prevAlts := game.CatalogAlternativeCosts
	game.CatalogAlternativeCosts = func(id string) []game.AlternativeCost {
		if id != oracle {
			return nil
		}
		return []game.AlternativeCost{{
			Key:      "escape",
			Label:    "Escape—{1}{B}, Exile two other cards from your graveyard",
			ManaCost: "{1}{B}",
			FromZone: game.ZoneGraveyard,
			ExileFromGraveyard: &game.TargetSpec{
				Mode:  "card_in_graveyard",
				Label: "two other cards from your graveyard",
				Zones: []game.ZoneKind{game.ZoneGraveyard},
				CardOK: func(_ *game.Game, caster uuid.UUID, c game.Card, _ game.ZoneKind) bool {
					return c.Owner == caster
				},
				Min: 2, Max: 2,
			},
		}}
	}
	t.Cleanup(func() { game.CatalogAlternativeCosts = prevAlts })

	seen := map[uuid.UUID]bool{me.ID: true, them.ID: true}
	push := func(p *game.Player, name, oracleID string) uuid.UUID {
		c := game.NewCard(name, p.ID)
		c.TypeLine = "Instant"
		c.OracleID = oracleID
		c.KnownBy = seen
		p.Graveyard.PushTop(c)
		return c.InstanceID
	}

	spell := push(me, "Test Escape", oracle)
	mine := []uuid.UUID{push(me, "Fodder A", ""), push(me, "Fodder B", "")}
	theirs := push(them, "Their Fodder", "")

	v := ViewOfGameFor(g, me.ID.String())
	var card *CardView
	for i := range v.Seats[0].Graveyard.Cards {
		if v.Seats[0].Graveyard.Cards[i].InstanceID == spell.String() {
			card = &v.Seats[0].Graveyard.Cards[i]
		}
	}
	if card == nil {
		t.Fatalf("graveyard view missing the escape card")
	}
	if len(card.AlternativeCosts) != 1 {
		t.Fatalf("offers = %+v, want escape only", card.AlternativeCosts)
	}
	opts := card.AlternativeCosts[0].PayOptions
	if opts == nil {
		t.Fatalf("the escape offer carries no pay_options — the client has nothing to pick from")
	}
	if opts.Min != 2 || opts.Max != 2 {
		t.Errorf("pay_options count: got %d..%d, want 2..2", opts.Min, opts.Max)
	}
	got := make(map[string]bool, len(opts.Cards))
	for _, id := range opts.Cards {
		got[id] = true
	}
	if got[spell.String()] {
		t.Errorf("the escaping card is offered as payment for its own cost")
	}
	if got[theirs.String()] {
		t.Errorf("an opponent's graveyard card is offered as payment")
	}
	for _, id := range mine {
		if !got[id.String()] {
			t.Errorf("a card in the caster's own graveyard is missing from the picker")
		}
	}
}
