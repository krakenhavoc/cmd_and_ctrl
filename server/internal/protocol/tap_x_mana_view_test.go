package protocol

import (
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Hazel of the Rootbloom
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tap_x_mana_view_test.go — #5447: "Tap X untapped tokens you control:
// Add X mana" ships its tap-others options with min 0 and max every
// option, and no up-front colour list, since the number of colours picked
// is the number of tokens the activator names.

const viewHazelOracle = "845b40b1-dc09-44fa-9910-22544c1424ba"

func TestTapXManaAbilityShipsAnOpenCount(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	hazel := uuid.New()
	tokens := []uuid.UUID{uuid.New(), uuid.New()}
	g.WithWriteLock(func() {
		known := map[uuid.UUID]bool{}
		for _, p := range g.Seats {
			known[p.ID] = true
		}
		g.Battlefield.PushTop(game.Card{
			InstanceID: hazel, Name: "Hazel of the Rootbloom", TypeLine: "Legendary Creature — Squirrel Druid",
			OracleID: viewHazelOracle, Power: 3, Toughness: 5, Owner: me.ID, Controller: me.ID, KnownBy: known,
		})
		for _, id := range tokens {
			g.Battlefield.PushTop(game.Card{
				InstanceID: id, Name: "Squirrel", TypeLine: "Token Creature — Squirrel",
				Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, KnownBy: known,
			})
		}
		g.Battlefield.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, KnownBy: known,
		})
	})
	card := battlefieldCardView(t, ViewOfGameFor(g, me.ID.String()), hazel)
	if len(card.ManaAbilities) != 1 {
		t.Fatalf("Hazel has %d mana abilities on the view, want 1", len(card.ManaAbilities))
	}
	row := card.ManaAbilities[0]
	o := row.TapOthersOptions
	if o == nil || o.Min != 0 || o.Max != 2 || len(o.Cards) != 2 || o.CountFromX {
		t.Fatalf("tap_others_options = %+v, want the two tokens with min 0 and max 2", o)
	}
	if len(row.ColorOptions) != 0 {
		t.Errorf("color_options = %v, want none before the payment is named", row.ColorOptions)
	}
}
