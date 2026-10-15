package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// flash_sacrifice_test.go — #3712: a permanent that is sacrificed at
// cleanup when it is flashed in is cast at instant speed only when it
// does something first (flash_sacrifice.go).

func TestFlashCleanupSacrificeIsHeldUnlessItMatters(t *testing.T) {
	flashAura := func() protocol.CardView {
		c := creature(cardID(1), 0, "Soar", 0, 0)
		c.TypeLine = "Enchantment — Aura"
		c.ManaCost = "{1}{U}"
		c.Purpose = &protocol.PurposeView{Aura: "benefit", FlashCleanupSacrifice: true}
		return c
	}
	flashEnchantment := func() protocol.CardView {
		c := creature(cardID(1), 0, "Parapet", 0, 0)
		c.TypeLine = "Enchantment"
		c.ManaCost = "{1}{W}"
		c.Purpose = &protocol.PurposeView{FlashCleanupSacrifice: true}
		return c
	}

	// cast values the cast of card, aimed at host (or nothing), in a
	// position on the given turn and step.
	cast := func(t *testing.T, card protocol.CardView, host *protocol.CardView, active int, step string, cfg heuristic.Config, extra ...viewOpt) float64 {
		t.Helper()
		m := castMove(t, 0, card.InstanceID, "cast")
		params := map[string]any{"instance_id": card.InstanceID, "from_zone": "hand"}
		var bf []protocol.CardView
		if host != nil {
			params["targets"] = []map[string]string{cardTarget(host.InstanceID)}
			bf = append(bf, *host)
		}
		m.Params = mustJSON(t, params)
		opts := append([]viewOpt{withBattlefield(bf...), withTurn(9, active, step)}, extra...)
		v := newView([]protocol.PlayerView{newSeat(0, withHand(card)), newSeat(1)}, opts...)
		in := input(0, v, passMove(0), m)
		return rankValue(t, heuristic.NewWithConfig(cfg), in, "cast")
	}
	cfg := heuristic.DefaultConfig()
	idle := creature(cardID(2), 0, "My Bear", 2, 2)
	blocker := creature(cardID(2), 0, "My Bear", 2, 2, blocking(cardID(9)))
	theirSpell := spell(cardID(8), 1, "Shock", "{R}")

	t.Run("an Aura at an idle end step is held", func(t *testing.T) {
		if v := cast(t, flashAura(), &idle, 1, "end", cfg); v >= 0 {
			t.Errorf("value %.3f, want it declined (below zero)", v)
		}
	})
	t.Run("the baseline casts it", func(t *testing.T) {
		if v := cast(t, flashAura(), &idle, 1, "end", heuristic.BaselineConfig()); v < 0 {
			t.Errorf("baseline value %.3f, want an ordinary price", v)
		}
	})
	t.Run("the same Aura in the bot's main phase is cast", func(t *testing.T) {
		if v := cast(t, flashAura(), &idle, 0, "precombat_main", cfg); v <= 0 {
			t.Errorf("value %.3f, want it cast", v)
		}
	})
	t.Run("an Aura on a blocker is cast as a trick", func(t *testing.T) {
		if v := cast(t, flashAura(), &blocker, 1, "declare_blockers", cfg); v <= 0 {
			t.Errorf("value %.3f, want it cast", v)
		}
	})
	t.Run("an Aura on an idle creature in combat is held", func(t *testing.T) {
		if v := cast(t, flashAura(), &idle, 1, "declare_blockers", cfg); v >= 0 {
			t.Errorf("value %.3f, want it declined", v)
		}
	})
	t.Run("an Aura in response to a spell is an answer", func(t *testing.T) {
		if v := cast(t, flashAura(), &idle, 1, "upkeep", cfg, withStack(theirSpell)); v <= 0 {
			t.Errorf("value %.3f, want it cast", v)
		}
	})
	t.Run("a blocker-maker is held at the end step and cast at declare attackers", func(t *testing.T) {
		if v := cast(t, flashEnchantment(), nil, 1, "end", cfg); v >= 0 {
			t.Errorf("end step value %.3f, want it declined", v)
		}
		if v := cast(t, flashEnchantment(), nil, 1, "declare_attackers", cfg); v < 0 {
			t.Errorf("declare attackers value %.3f, want it cast", v)
		}
	})
}
