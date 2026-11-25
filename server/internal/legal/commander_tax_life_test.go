package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// commander_tax_life_test.go — the enumerator half of #6151: Liesa,
// Shroud of Dusk pays her commander tax (CR 903.8) in life, 2 per previous
// cast. The move is priced at her printed mana, and it is not offered when
// the life total can't pay the tax and survive.

const oracleLiesa = "2270daaf-252d-40c4-bed5-bd46ae15f041"

func liesaInCommandZone(t *testing.T, g *game.Game, p *game.Player, casts int) uuid.UUID {
	t.Helper()
	basicLands(g, p, 3, "Plains")
	basicLands(g, p, 2, "Swamp")
	id := uuid.New()
	p.Command.PushTop(game.Card{
		InstanceID: id, Name: "Liesa, Shroud of Dusk", TypeLine: "Legendary Creature — Angel",
		ManaCost: "{2}{W}{W}{B}", Power: 5, Toughness: 5, OracleID: oracleLiesa,
		Owner: p.ID, Controller: p.ID, IsCommander: true,
		KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	p.CommanderCasts[id] = casts
	return id
}

func TestCommanderTaxInLifeIsOfferedAtThePrintedMana(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	id := liesaInCommandZone(t, g, active, 2)
	advanceTo(t, g, game.StepPrecombatMain)

	for _, m := range wantCastMana(t, g, active.ID, id, "{2}{W}{W}{B}", 5) {
		if m.Cost == nil || m.Cost.Life != 4 {
			t.Errorf("%q states cost %+v, want 4 life", m.Label, m.Cost)
		}
	}
	dispatchAll(t, g, active.ID, legal.EnumerateFor(g, active.ID))
}

func TestCommanderTaxInLifeIsNotOfferedWithoutTheLife(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	id := liesaInCommandZone(t, g, active, 2)
	active.Life = 4
	advanceTo(t, g, game.StepPrecombatMain)

	if casts := castMovesFor(legal.EnumerateFor(g, active.ID), id); len(casts) != 0 {
		t.Errorf("offered %d casts paying 4 life at 4 life: %v", len(casts), labels(casts))
	}
}
