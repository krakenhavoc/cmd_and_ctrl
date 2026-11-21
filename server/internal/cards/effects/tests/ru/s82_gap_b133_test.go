package testsru

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const s82GapBanishFromEdorasOracle = "e4508afd-7cc3-4176-8291-cfc2c155ac11"

func TestS82GapBanishFromEdorasCostsTwoLessAgainstATappedCreatureAndExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	untapped := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	tapped := effectstest.B12Creature(g, opp.ID, "Tired Ox", "Creature — Ox", 2, 2)
	g.WithWriteLock(func() { effectstest.B16Tap(g, tapped) })
	effectstest.AdvanceToMain(t, g)
	card := game.Card{Name: "Banish from Edoras", TypeLine: "Sorcery", ManaCost: "{4}{W}", OracleID: s82GapBanishFromEdorasOracle}

	if err := effectstest.SocB19CastWithPool(g, me, card, "W", 3, game.CastSpellParams{Targets: effectstest.CardRefs(untapped)}); err == nil {
		t.Fatal("cast for three mana against an untapped creature")
	}
	if err := effectstest.SocB19CastWithPool(g, me, card, "W", 3, game.CastSpellParams{Targets: effectstest.CardRefs(tapped)}); err != nil {
		t.Fatalf("cast for {2}{W} against a tapped creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(tapped) {
		t.Error("the tapped creature was not exiled")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature left the battlefield")
	}
}

func TestS82GapBanishFromEdorasExilesAnUntappedCreatureAtFullPrice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ox := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	effectstest.AdvanceToMain(t, g)
	card := game.Card{Name: "Banish from Edoras", TypeLine: "Sorcery", ManaCost: "{4}{W}", OracleID: s82GapBanishFromEdorasOracle}
	if err := effectstest.SocB19CastWithPool(g, me, card, "W", 5, game.CastSpellParams{Targets: effectstest.CardRefs(ox)}); err != nil {
		t.Fatalf("cast for {4}{W}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ox) {
		t.Error("the creature was not exiled")
	}
}
