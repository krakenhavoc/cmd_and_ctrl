package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// token_abilities_view_test.go is the wire half of #521.
//
// stampActivatedAbilities used to skip any card with no oracle ID,
// and a token has none — so every token's intrinsic activated
// abilities were dropped on the way to the client. Food, Clue, Blood
// and the Lander carried their whole printed text in that slot, so
// they reached the table inert: the client had nothing to render and
// nothing to fire. Treasure escaped only because its ability is a
// MANA ability, and viewOfManaAbilities is stamped unconditionally.
//
// With a token's abilities in the catalog under its own key, the
// projection asks the same accessor it asks for a printed card.

func pushViewToken(g *game.Game, controller uuid.UUID, tmpl game.Card) uuid.UUID {
	tmpl.InstanceID = uuid.New()
	tmpl.Owner, tmpl.Controller = controller, controller
	g.Battlefield.PushTop(tmpl)
	return tmpl.InstanceID
}

func TestTokenActivatedAbilitiesReachTheClient(t *testing.T) {
	cases := []struct {
		name  string
		tmpl  game.Card
		label string
	}{
		{"Food", effects.FoodToken(), "{2}, {T}, Sacrifice this artifact: You gain 3 life"},
		{"Clue", effects.ClueToken(), "{2}, Sacrifice this artifact: Draw a card"},
		{"Blood", effects.BloodToken(), "{1}, {T}, Sacrifice this artifact: Draw a card (discard cost not yet modelled)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := buildActiveGame(t)
			id := pushViewToken(g, g.Seats[0].ID, tc.tmpl)

			c := conditionCardView(t, g, id)
			if len(c.ActivatedAbilities) != 1 {
				t.Fatalf("%s shipped %d activated abilities, want 1 — the client cannot crack it",
					tc.name, len(c.ActivatedAbilities))
			}
			if got := c.ActivatedAbilities[0].Label; got != tc.label {
				t.Errorf("label = %q, want %q", got, tc.label)
			}
			if !c.ActivatedAbilities[0].SacrificeSelf {
				t.Errorf("%s's ability reached the wire without its sacrifice cost", tc.name)
			}
		})
	}
}

// A Treasure's mana row was already shipped; what was missing is the
// per-row stamping that the same guard skipped.
func TestTreasureManaRowStillShips(t *testing.T) {
	g := buildActiveGame(t)
	id := pushViewToken(g, g.Seats[0].ID, effects.TreasureToken())

	c := conditionCardView(t, g, id)
	if len(c.ManaAbilities) != 1 {
		t.Fatalf("Treasure shipped %d mana abilities, want 1", len(c.ManaAbilities))
	}
	if !c.ManaAbilities[0].SacrificeCost || !c.ManaAbilities[0].TapCost {
		t.Errorf("Treasure mana row = %+v, want {T} + sacrifice", c.ManaAbilities[0])
	}
	if len(c.ActivatedAbilities) != 0 {
		t.Errorf("a Treasure has no stack-using ability: %+v", c.ActivatedAbilities)
	}
}

// --- ADR 0083: the token's printed text ---------------------------
//
// An activated ability reaches the client as a row in the right-click
// menu, which is both the text and the control. A TRIGGERED or STATIC
// ability has no control, so without `token_text` a Dragon Egg would
// reach the board as the word "Dragon Egg" and its dies-trigger would
// be a surprise. A token has no scryfall_id to fetch oracle text from
// (ADR 0078's art is still Proposed), so the server says what the
// token prints, the way it already does for an emblem.

func TestATokensPrintedTextReachesTheClient(t *testing.T) {
	cases := []struct {
		name string
		tmpl game.Card
		want string
	}{
		{
			"a trigger only",
			effects.PestToken(),
			"When this token dies, you gain 1 life.",
		},
		{
			"a keyword and an activated ability",
			effects.FoodToken(),
			"{2}, {T}, Sacrifice this token: You gain 3 life.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := buildActiveGame(t)
			id := pushViewToken(g, g.Seats[0].ID, tc.tmpl)
			if got := conditionCardView(t, g, id).TokenText; got != tc.want {
				t.Errorf("token_text = %q, want %q", got, tc.want)
			}
		})
	}
}

// A printed card says nothing here — the client renders its oracle
// text from the printing, and a second copy on the wire would be a
// second source of truth to drift.
func TestAPrintedCardShipsNoTokenText(t *testing.T) {
	g := buildActiveGame(t)
	c := game.NewCard("Grizzly Bears", g.Seats[0].ID)
	c.TypeLine = "Creature — Bear"
	c.OracleID = "oracle-bears"
	c.ScryfallID = "scryfall-bears"
	id := pushViewToken(g, g.Seats[0].ID, c)

	if got := conditionCardView(t, g, id).TokenText; got != "" {
		t.Errorf("a printed card shipped token_text %q", got)
	}
}

// A vanilla token says nothing either: it has no catalog entry at all,
// which is the whole reason tokens_table.go is a data table.
func TestAVanillaTokenShipsNoTokenText(t *testing.T) {
	g := buildActiveGame(t)
	id := pushViewToken(g, g.Seats[0].ID, effects.TokenCard("1/1 white Soldier"))

	if got := conditionCardView(t, g, id).TokenText; got != "" {
		t.Errorf("a 1/1 Soldier shipped token_text %q", got)
	}
}
