package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kuja_a_kefka_test.go — Kefka, Dancing Mad (#3213).

func KefkaGame(t *testing.T) (*game.Game, uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	me := g.Seats[0]
	return g, B27Push(g, me.ID, "Kefka, Dancing Mad", "Legendary Creature — Human Wizard", kefkaOracle, "{5}{B}{R}", 6, 6, "B", "R")
}

func GraveyardSpell(owner *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	owner.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", ManaCost: cost,
		OracleID: SpellControlDivinationOracle, Owner: owner.ID, Controller: owner.ID})
	return id
}
