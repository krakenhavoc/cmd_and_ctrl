package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// vivi_autotap_test.go — #1621's auto-tap half (owner decision
// 2026-09-30): the planner may use a mana ability that costs its
// source nothing, bounded by "Activate only once each turn" — Vivi
// Ornitier's {0} — but only as its LAST resort, after every ordinary
// source, every frozen source and the sacrifice tier. The preview, the
// strict cast gate and the bot enumerator all read the one planner, so
// every test here asks at least two of them.

// viviAtMain seats a Vivi Ornitier with `power` +1/+1 counters under
// the active player, on that player's precombat main phase.
func viviAtMain(t *testing.T, power int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	vivi := b12Push(g, me.ID, "Vivi Ornitier", "Legendary Creature — Wizard", viviOrnitierOracle, 0, 3)
	if power > 0 {
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(vivi, game.CounterPlusOne, power) })
	}
	return g, me, vivi
}

// seatSpell puts a vanilla spell costing `cost` into `p`'s hand.
func seatSpell(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		ManaCost: cost, Owner: p.ID, Controller: p.ID,
	})
	return id
}

func mustCost(t *testing.T, s string) game.ParsedCost {
	t.Helper()
	c, err := game.ParseCost(s)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", s, err)
	}
	return c
}

// viviActivations counts Vivi's mana-ability activations so far.
func viviActivations(g *game.Game, vivi uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventManaAbilityActivated && ev.Source == vivi {
			n++
		}
	}
	return n
}

func previewFor(t *testing.T, g *game.Game, who uuid.UUID, cost string) ([]game.AutoTapPlanEntry, bool) {
	t.Helper()
	return g.AutoTapPlanPreferringExcluding(who, mustCost(t, cost), 0, nil, 0)
}

func planHas(plan []game.AutoTapPlanEntry, id uuid.UUID) bool {
	for _, e := range plan {
		if e.CardID == id {
			return true
		}
	}
	return false
}

func castOffered(g *game.Game, seat, spell uuid.UUID) bool {
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type == legal.TypeCastSpell && m.Source == spell {
			return true
		}
	}
	return false
}

// Lands that can pay leave Vivi alone: the one activation this turn is
// not spent to save a land from tapping.
func TestAutoTapDoesNotUseViviWhenTheLandsSuffice(t *testing.T) {
	g, me, vivi := viviAtMain(t, 3)
	i1 := seatToken(g, me.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	i2 := seatToken(g, me.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	spell := seatSpell(me, "Two-Drop Drake", "Creature — Drake", "{1}{U}")

	plan, ok := previewFor(t, g, me.ID, "{1}{U}")
	if !ok || len(plan) != 2 || planHas(plan, vivi) {
		t.Fatalf("preview ok=%v plan=%v, want the two Islands and not Vivi", ok, plan)
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("strict auto-tap cast: %v", err)
	}
	for _, id := range []uuid.UUID{i1, i2} {
		if c, _ := battlefieldCard(g, id); !c.Tapped {
			t.Errorf("Island %v untapped — the lands pay first", id)
		}
	}
	if n := viviActivations(g, vivi); n != 0 {
		t.Errorf("Vivi activated %d time(s) for a cast the lands paid", n)
	}
}

// Lands fall short: the planner spends the Treasure first and reaches
// for Vivi only for what is left, the preview names exactly the
// payment the cast makes, and the bot enumerator offers the cast
// because the same planner says it is payable.
func TestAutoTapReachesForViviAfterTheTreasure(t *testing.T) {
	g, me, vivi := viviAtMain(t, 2)
	island := seatToken(g, me.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	treasure := seatToken(g, me.ID, TreasureToken())

	// {1}{U}: the Island and the Treasure pay it, and Vivi is left be —
	// the sacrifice tier comes before the costless one.
	if plan, ok := previewFor(t, g, me.ID, "{1}{U}"); !ok || planHas(plan, vivi) || !planHas(plan, treasure) {
		t.Fatalf("{1}{U}: preview ok=%v plan=%v, want the Island and the Treasure and not Vivi", ok, plan)
	}

	// {3}{U} needs all four mana: Island, Treasure, and Vivi's two.
	spell := seatSpell(me, "Four-Drop Sphinx", "Creature — Sphinx", "{3}{U}")
	plan, ok := previewFor(t, g, me.ID, "{3}{U}")
	if !ok || len(plan) != 3 {
		t.Fatalf("{3}{U}: preview ok=%v plan=%v, want Island, Treasure and Vivi", ok, plan)
	}
	if plan[len(plan)-1].CardID != vivi {
		t.Errorf("Vivi is not the last source recruited: plan=%v", plan)
	}
	for _, e := range plan {
		if e.CardID != vivi {
			continue
		}
		if e.Taps || e.Sacrifices || e.Exiles || !e.OncePerTurn {
			t.Errorf("Vivi's entry %+v, want once-per-turn and nothing tapped, sacrificed or exiled", e)
		}
	}
	if !castOffered(g, me.ID, spell) {
		t.Error("the enumerator does not offer a cast the planner says is payable")
	}

	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("strict auto-tap cast: %v", err)
	}
	// The cast spent exactly what the preview named.
	if c, _ := battlefieldCard(g, island); !c.Tapped {
		t.Error("the Island was not tapped")
	}
	if _, ok := battlefieldCard(g, treasure); ok {
		t.Error("the Treasure survived a cast it paid for")
	}
	if n := viviActivations(g, vivi); n != 1 {
		t.Errorf("Vivi activated %d time(s), want 1", n)
	}
	if c, _ := battlefieldCard(g, vivi); c.Tapped {
		t.Error("Vivi was tapped — its {0} taps nothing")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool has %d tokens after an exact payment", len(me.ManaPool))
	}
}

// A Vivi bigger than the gap still pays only what is owed; the rest
// floats, as any surplus does (CR 106.4).
func TestAutoTapViviSurplusFloats(t *testing.T) {
	g, me, vivi := viviAtMain(t, 3)
	_ = seatToken(g, me.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	spell := seatSpell(me, "Three-Drop Djinn", "Creature — Djinn", "{2}{U}")

	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("strict auto-tap cast: %v", err)
	}
	if n := viviActivations(g, vivi); n != 1 {
		t.Errorf("Vivi activated %d time(s), want 1", n)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool has %d tokens, want Vivi's one surplus mana floating", len(me.ManaPool))
	}
}

// "Activate only … once each turn": a Vivi the auto-tapper already
// spent this turn is not a source again, to the preview, the strict
// gate or the enumerator.
func TestAutoTapDoesNotPlanViviTwiceInOneTurn(t *testing.T) {
	g, me, vivi := viviAtMain(t, 2)
	island := seatToken(g, me.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	first := seatSpell(me, "Three-Drop Djinn", "Creature — Djinn", "{2}{U}")
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("first cast: %v", err)
	}
	if n := viviActivations(g, vivi); n != 1 {
		t.Fatalf("Vivi activated %d time(s) on the first cast, want 1", n)
	}
	passPriorityAroundTable(t, g)

	// Untap the Island by hand so it alone is short of the second cast.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == island {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	second := seatSpell(me, "Two-Drop Drake", "Creature — Drake", "{1}{U}")
	if plan, ok := previewFor(t, g, me.ID, "{1}{U}"); ok {
		t.Errorf("preview ok with plan %v — Vivi's once-each-turn ability is spent", plan)
	}
	if castOffered(g, me.ID, second) {
		t.Error("the enumerator offers a cast only a spent Vivi could pay")
	}
	err := g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true, AutoTap: true})
	var short *game.InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("second cast = %v, want an InsufficientManaError", err)
	}
	if n := viviActivations(g, vivi); n != 1 {
		t.Errorf("Vivi activated %d time(s), want still 1", n)
	}
}

// "Activate only during your turn": an opponent's Vivi is not a source
// for its controller while somebody else has the turn.
func TestAutoTapDoesNotPlanViviOnAnotherPlayersTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vivi := b12Push(g, them.ID, "Vivi Ornitier", "Legendary Creature — Wizard", viviOrnitierOracle, 0, 3)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(vivi, game.CounterPlusOne, 2) })

	if plan, ok := previewFor(t, g, them.ID, "{1}{U}"); ok || planHas(plan, vivi) {
		t.Errorf("preview ok=%v plan=%v on someone else's turn, want no plan", ok, plan)
	}
	if ids, ok := g.AutoTapForCost(them.ID, mustCost(t, "{U}"), 0); ok {
		t.Errorf("AutoTapForCost planned %v off an opponent's-turn Vivi", ids)
	}
}

// A power-0 Vivi — every Vivi the turn it lands — prices at nothing
// and is not a source.
func TestAutoTapDoesNotPlanAPowerZeroVivi(t *testing.T) {
	g, me, vivi := viviAtMain(t, 0)
	if plan, ok := previewFor(t, g, me.ID, "{U}"); ok || planHas(plan, vivi) {
		t.Errorf("preview ok=%v plan=%v off a power-0 Vivi, want no plan", ok, plan)
	}
}

// Ramos, Dragon Engine is once-each-turn too, but its payout removes
// five +1/+1 counters — a cost, so it stays out of every plan.
func TestAutoTapNeverPlansRamosCounterPayout(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	ramos := b12Push(g, me.ID, "Ramos, Dragon Engine", "Legendary Artifact Creature — Dragon", ramosOracle, 4, 4)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ramos, game.CounterPlusOne, 5) })
	if plan, ok := previewFor(t, g, me.ID, "{W}"); ok || planHas(plan, ramos) {
		t.Errorf("preview ok=%v plan=%v planned Ramos's counter payout", ok, plan)
	}
}
