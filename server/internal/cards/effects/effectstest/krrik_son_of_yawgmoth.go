package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// krrik_son_of_yawgmoth_test.go — ADR 0131 (#2531), PRs 1 and 2: the third
// ability, "for each {B} in a cost, you may pay 2 life rather than pay
// that mana", for casts, activations, attack taxes and a mana ability, on the real
// catalog card. The engine's cases are game/life_for_mana_test.go.

const (
	KrrikOracle = "cbe3a4e7-5dbe-4f58-8ee6-a1762b65acfd"
)

// krrikTable is a four-seat catalog game on the active seat's main
// phase with K'rrik on that seat's side, a creature to kill on the next
// seat's, and `life` on the active seat.
func KrrikTable(t *testing.T) (g *game.Game, me *game.Player, krrik, victim uuid.UUID) {
	t.Helper()
	g = NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	krrik = PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "K'rrik, Son of Yawgmoth", OracleID: KrrikOracle,
		TypeLine: "Legendary Creature — Phyrexian Horror Minion", ManaCost: "{4}{B/P}{B/P}{B/P}",
		Colors: []string{"B"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim = WpCreature(g, opp.ID, "Three Drop", "{2}{B}")
	return g, me, krrik, victim
}
