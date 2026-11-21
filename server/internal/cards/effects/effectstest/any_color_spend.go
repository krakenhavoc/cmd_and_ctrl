package effectstest

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// any_color_spend_test.go — #1600's proof cards: Chromatic Orrery,
// Mycosynth Lattice and Oath of Nissa, the three printed forms of a
// player's "you may spend mana as though it were mana of any color"
// (CR 609.4b). The engine half is game/spend_any_color_test.go.

// mycosynthLatticeOracle is declared in lattice_lord_test.go.
const (
	OathOfNissaOracle = "c4efdbab-711d-4269-9b24-b05d36f7e5c7"
	PestilenceOracle  = "dafe63ef-f3d6-45e7-877a-573da92ba85e"
)

// spendTable is a four-seat catalog game in seat 0's main phase.
func SpendTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceTo(t, g, game.StepPrecombatMain)
	return g, g.Seats[0], g.Seats[1]
}

// handSpell puts a non-catalog spell with `cost` into p's hand.
func HandSpell(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

func RefusedForMana(t *testing.T, err error, what string) {
	t.Helper()
	var short *game.InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("%s: %v, want insufficient mana", what, err)
	}
}
