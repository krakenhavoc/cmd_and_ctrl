package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3172 cav-h2: Timely Ward's "as though it had flash if it targets a
// commander" (FlashIfTargets).

const timelyWardOracle = "8e589dca-d31e-4d2f-81b1-18e5f9198795"

func timelyWardTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	g, me, _ := p7Table(t)
	ward := eacHandCard(g, me, "Timely Ward", "Enchantment — Aura", "{2}{W}")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == ward {
				me.Hand.Cards[i].OracleID = timelyWardOracle
			}
		}
	})
	commander := b12Creature(g, me.ID, "My Commander", "Legendary Creature — Human", 2, 2)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == commander {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
	return g, me, ward, commander, bear
}

func timelyWardFund(g *game.Game, me *game.Player) {
	g.WithWriteLock(func() {
		me.ManaPool = nil
		for _, c := range []string{"W", "C", "C"} {
			me.ManaPool.AddMana(game.ManaToken{Color: c})
		}
	})
}

func TestTimelyWardFlashesOntoACommanderOnly(t *testing.T) {
	g, me, ward, commander, bear := timelyWardTable(t)
	ssToBeginCombat(t, g)
	timelyWardFund(g, me)
	target := func(id uuid.UUID) game.CastSpellParams {
		return game.CastSpellParams{Strict: true, AlternativeCost: "flash-if-targets",
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}}}
	}
	// The printed cast is sorcery speed whatever it targets.
	if err := g.CastSpell(me.ID, ward, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: commander}}}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("printed cost at instant speed: err = %v, want ErrSorcerySpeedRequired", err)
	}
	// Claiming the flash offer on a creature that is not a commander is refused.
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
	g, me, ward, _, bear := timelyWardTable(t)
	advanceToMain(t, g)
	timelyWardFund(g, me)
	if err := g.CastSpell(me.ID, ward, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("sorcery-speed cast onto a plain creature: %v", err)
	}
}
