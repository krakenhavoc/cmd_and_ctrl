package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flashIfHandCard puts a catalog card in `me`'s hand.
func FlashIfHandCard(g *game.Game, me *game.Player, name, typeLine, mana, oracle string) uuid.UUID {
	card := EacHandCard(g, me, name, typeLine, mana)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == card {
				me.Hand.Cards[i].OracleID = oracle
			}
		}
	})
	return card
}

func SsToBeginCombat(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepBeginCombat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

func FlashIfFund(g *game.Game, me *game.Player, colors ...string) {
	g.WithWriteLock(func() {
		me.ManaPool = nil
		for _, c := range colors {
			me.ManaPool.AddMana(game.ManaToken{Color: c})
		}
	})
}
