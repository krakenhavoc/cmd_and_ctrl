package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc12Cast casts a permanent from the active player's hand and resolves it,
// leaving any entry trigger's prompt open.
func Msc12Cast(t *testing.T, g *game.Game, name, typeLine, cost, pay, oracle string, pt int) uuid.UUID {
	t.Helper()
	return CastPowerUpCreature(t, g, g.Seats[g.Turn.ActiveSeat], name, typeLine, cost, pay, oracle, pt)
}

// msc12CastAndSettle casts a 1/1 creature spell for {1} and lets the stack
// empty, so a cast trigger and the spell both resolve.
func Msc12CastAndSettle(t *testing.T, g *game.Game, name, typeLine string) {
	t.Helper()
	p := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{1}",
			Owner: p.ID, Controller: p.ID, Power: 1, Toughness: 1})
	})
	FloatMana(t, g, p, "{C}")
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatalf("%s did not resolve onto the battlefield", name)
	}
}

// msc12PickPlayer answers the open pick_target prompt with a player.
func Msc12PickPlayer(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	ch := PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// --- Giant-Man, Gargantuan Genius ---

func Msc12PoolOf(p *game.Player, color string) int {
	n := 0
	for _, tok := range p.ManaPool.Units() {
		if tok.Color == color {
			n++
		}
	}
	return n
}
