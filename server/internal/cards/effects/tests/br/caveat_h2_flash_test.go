package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func timelyWardFund(g *game.Game, me *game.Player) {
	g.WithWriteLock(func() {
		me.ManaPool = nil
		for _, c := range []string{"W", "C", "C"} {
			me.ManaPool.AddMana(game.ManaToken{Color: c})
		}
	})
}

func TestTimelyWardFlashesOntoACommanderOnly(t *testing.T) {
	g, me, ward, commander, bear := effectstest.TimelyWardTable(t)
	effectstest.SsToBeginCombat(t, g)
	timelyWardFund(g, me)
	target := func(id uuid.UUID) game.CastSpellParams {
		return game.CastSpellParams{Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}}}
	}
	// #5207: the flash is the card's own timing permission, so the
	// instant-speed cast pays the printed cost. On a creature that is not
	// a commander it is refused.
	if err := g.CastSpell(me.ID, ward, target(bear)); err == nil {
		t.Fatal("flash-cast onto a creature that is not a commander")
	}
	if g.Stack.Contains(ward) {
		t.Fatal("a refused cast reached the stack")
	}
	if err := g.CastSpell(me.ID, ward, target(commander)); err != nil {
		t.Fatalf("flash-cast onto a commander: %v", err)
	}
	if !g.Stack.Contains(ward) {
		t.Error("the Aura is not on the stack")
	}
}

func TestTimelyWardCastsAtSorcerySpeedOnAnyCreature(t *testing.T) {
	g, me, ward, _, bear := effectstest.TimelyWardTable(t)
	effectstest.AdvanceToMain(t, g)
	timelyWardFund(g, me)
	if err := g.CastSpell(me.ID, ward, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("sorcery-speed cast onto a plain creature: %v", err)
	}
}
