package aiseat_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// temporary_body_test.go — #3802 and #3776 through the real engine.
//
// #3802: a creature cast for evoke, warp, dash or blitz is a body that
// leaves, and the bot prices it so: it hard-casts the body it wants when
// it can pay for it, and takes the alternative cost when that is the
// point (an enters effect, a lethal haste attack).
//
// #3776: turning Unstable Hulk face up pumps it +6/+6 with trample for the
// turn, and the bot takes the flip when that is lethal.
//
// Not behind AISEAT_GAME_TESTS: one move list and one decision each.

const (
	oracleMulldrifterTB  = "24d0f5e7-0d9e-4b76-900e-a7274e80312d"
	oracleWarchiefTB     = "0695af85-7405-4512-bb97-42517a532ada"
	oracleRagavanTB      = "37108cd4-bbab-4ce3-9ed6-f60e8422e703"
	oracleTimelineTB     = "3c307122-bc18-41f1-87d5-ae6b3ca9b8b9"
	oracleUnstableHulkTB = "8d01c85c-5f83-4514-b81c-4165051d5722"
)

// tbTable is a settled two-seat table at the active seat's precombat main
// with an empty hand and `pool` in its mana pool.
func tbTable(t *testing.T, seed uint64, pool string) (*game.Game, *game.Player) {
	t.Helper()
	g := newSettledTable(t, seed)
	advanceToStep(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	if pool != "" {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, pool); err != nil {
			t.Fatal(err)
		}
	}
	return g, me
}

// tbHandCard puts one catalog creature in the seat's hand.
func tbHandCard(g *game.Game, me *game.Player, c game.Card) uuid.UUID {
	var id uuid.UUID
	g.WithWriteLock(func() {
		c.Owner, c.Controller = me.ID, me.ID
		c.AddKnowersAll(seatIDs(g))
		me.Hand.PushTop(c)
		id = c.InstanceID
	})
	return id
}

// tbDecide asks the default heuristic for a decision and returns the move.
func tbDecide(t *testing.T, g *game.Game, me *game.Player, cfg heuristic.Config) legal.Move {
	t.Helper()
	moves := legal.EnumerateFor(g, me.ID)
	in := aiseat.Input{View: protocol.ViewOfGameFor(g, me.ID.String()), Seat: me.ID, Moves: moves}
	d, err := heuristic.NewWithConfig(cfg).Decide(context.Background(), in)
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if d.Index < 0 || d.Index >= len(moves) {
		t.Fatalf("the bot made no move (index %d of %d: %v)", d.Index, len(moves), attackLabels(moves))
	}
	return moves[d.Index]
}

// tbAltCost is the alternative cost a cast move claims, "" for none.
func tbAltCost(t *testing.T, m legal.Move) string {
	t.Helper()
	var p struct {
		AlternativeCost string `json:"alternative_cost"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("decode %q: %v", m.Label, err)
	}
	return p.AlternativeCost
}

func mulldrifter() game.Card {
	c := game.NewCard("Mulldrifter", uuid.Nil)
	c.TypeLine, c.ManaCost, c.OracleID = "Creature — Elemental", "{4}{U}", oracleMulldrifterTB
	c.Power, c.Toughness = 2, 2
	return c
}

func warchief() game.Card {
	c := game.NewCard("Workshop Warchief", uuid.Nil)
	c.TypeLine, c.ManaCost, c.OracleID = "Creature — Rhino Warrior", "{3}{G}{G}", oracleWarchiefTB
	c.Power, c.Toughness = 5, 3
	return c
}

func ragavan() game.Card {
	c := game.NewCard("Ragavan, Nimble Pilferer", uuid.Nil)
	c.TypeLine, c.ManaCost, c.OracleID = "Legendary Creature — Monkey Pirate", "{R}", oracleRagavanTB
	c.Power, c.Toughness = 2, 1
	return c
}

func timelineCuller() game.Card {
	c := game.NewCard("Timeline Culler", uuid.Nil)
	c.TypeLine, c.ManaCost, c.OracleID = "Creature — Drix Warlock", "{B}{B}", oracleTimelineTB
	c.Power, c.Toughness = 2, 2
	return c
}

// With the mana for the printed cost the bot hard-casts Mulldrifter: the
// body AND the two cards, where an evoke is the two cards alone.
func TestHeuristicHardCastsMulldrifterWhenItCanPayForTheBody(t *testing.T) {
	g, me := tbTable(t, 3802, "{U}{U}{U}{U}{U}")
	tbHandCard(g, me, mulldrifter())
	m := tbDecide(t, g, me, heuristic.DefaultConfig())
	if m.Kind != legal.KindCast || tbAltCost(t, m) != "" {
		t.Fatalf("the bot chose %q (alt %q); want the hard cast, the 2/2 flier that stays", m.Label, tbAltCost(t, m))
	}
}

// With only the evoke mana the cards are worth the cast.
func TestHeuristicEvokesMulldrifterWhenThatIsAllItCanPay(t *testing.T) {
	g, me := tbTable(t, 3803, "{U}{U}{U}")
	tbHandCard(g, me, mulldrifter())
	m := tbDecide(t, g, me, heuristic.DefaultConfig())
	if m.Kind != legal.KindCast || tbAltCost(t, m) != "evoke" {
		t.Fatalf("the bot chose %q (alt %q); want the evoke for the two cards", m.Label, tbAltCost(t, m))
	}
}

// Seven lands pay for the Warchief's printed {3}{G}{G}; blitzing it for
// {4}{G}{G} spends a card's worth of body to swing once.
func TestHeuristicHardCastsWorkshopWarchiefInsteadOfBlitzingIt(t *testing.T) {
	g, me := tbTable(t, 3804, "{G}{G}{G}{G}{G}{G}{G}")
	tbHandCard(g, me, warchief())
	m := tbDecide(t, g, me, heuristic.DefaultConfig())
	if m.Kind != legal.KindCast || tbAltCost(t, m) != "" {
		t.Fatalf("the bot chose %q (alt %q); want the hard cast of the 5/3 that stays", m.Label, tbAltCost(t, m))
	}
}

// A lethal haste attack is the point of a dash: the opponent is at 2 with
// nothing to block, Ragavan hard-cast is summoning-sick and does nothing
// this turn, dashed it connects.
func TestHeuristicDashesForLethal(t *testing.T) {
	g, me := tbTable(t, 3805, "{R}{R}")
	tbHandCard(g, me, ragavan())
	g.WithWriteLock(func() { g.Seats[1-g.Turn.ActiveSeat].Life = 2 })
	m := tbDecide(t, g, me, heuristic.DefaultConfig())
	if m.Kind != legal.KindCast || tbAltCost(t, m) != "dash" {
		t.Fatalf("the bot chose %q (alt %q); want the dash for the lethal swing", m.Label, tbAltCost(t, m))
	}
}

// Blitz is the same shape: haste, and a card back when it dies.
func TestHeuristicBlitzesForLethal(t *testing.T) {
	g, me := tbTable(t, 3806, "{G}{G}{G}{G}{G}{G}")
	tbHandCard(g, me, warchief())
	g.WithWriteLock(func() { g.Seats[1-g.Turn.ActiveSeat].Life = 5 })
	m := tbDecide(t, g, me, heuristic.DefaultConfig())
	if m.Kind != legal.KindCast || tbAltCost(t, m) != "blitz" {
		t.Fatalf("the bot chose %q (alt %q); want the blitz for the lethal 5-power swing", m.Label, tbAltCost(t, m))
	}
}

// Warp is the temporary body without the haste attack of dash and blitz:
// Timeline Culler has haste, so a warped one swings, but at an opponent
// on a healthy life total with the mana for the printed cost the bot keeps
// the body.
func TestHeuristicHardCastsTimelineCullerInsteadOfWarpingIt(t *testing.T) {
	g, me := tbTable(t, 3807, "{B}{B}{B}{B}")
	tbHandCard(g, me, timelineCuller())
	m := tbDecide(t, g, me, heuristic.DefaultConfig())
	if m.Kind != legal.KindCast || strings.HasPrefix(tbAltCost(t, m), "warp") {
		t.Fatalf("the bot chose %q (alt %q); want the hard cast of the body that stays", m.Label, tbAltCost(t, m))
	}
}

// --- Unstable Hulk ------------------------------------------------------

// tbFaceDownHulk casts Unstable Hulk face down, resolves it, and returns
// it as a permanent its controller has held since the start of the turn.
func tbFaceDownHulk(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	hulk := game.NewCard("Unstable Hulk", me.ID)
	hulk.TypeLine, hulk.ManaCost, hulk.OracleID = "Creature — Goblin Mutant", "{1}{R}{R}", oracleUnstableHulkTB
	hulk.Power, hulk.Toughness = 2, 2
	id := tbHandCard(g, me, hulk)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "morph"}); err != nil {
		t.Fatalf("cast face down: %v", err)
	}
	for i := 0; i < 16 && len(g.StackMeta) > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority %d: %v", i, err)
		}
	}
	if len(g.StackMeta) != 0 {
		t.Fatal("the face-down Hulk did not resolve")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	return id
}

func tbTurnsFaceUp(t *testing.T, m legal.Move) bool {
	t.Helper()
	var p struct {
		Kind string `json:"kind"`
	}
	_ = json.Unmarshal(m.Params, &p)
	return m.Type == legal.TypeSpecialAction && p.Kind == "turn_face_up"
}

// The opponent is at 8 with an empty board: the face-down 2/2 deals 2,
// turned up it is an 8/8 trampler and the swing is lethal. The flip costs
// the Hulk's controller a turn and the bot takes it anyway.
func TestHeuristicTurnsUnstableHulkUpForLethal(t *testing.T) {
	g, me := tbTable(t, 3776, "")
	id := tbFaceDownHulk(t, g, me)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{R}{R}{R}{R}"); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { g.Seats[1-g.Turn.ActiveSeat].Life = 8 })
	m := tbDecide(t, g, me, heuristic.DefaultConfig())
	if !tbTurnsFaceUp(t, m) {
		t.Fatalf("the bot chose %q; want to turn the Hulk (%s) face up for the lethal 8/8", m.Label, id)
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: me.ID, Params: m.Params}); err != nil {
		t.Fatalf("dispatch %q: %v", m.Label, err)
	}
}

// A healthy opponent: the flip is a skipped turn for six more damage,
// which is not worth it, and the bot does not take it.
func TestHeuristicDoesNotTurnUnstableHulkUpWithoutAKill(t *testing.T) {
	g, me := tbTable(t, 3777, "")
	tbFaceDownHulk(t, g, me)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{R}{R}{R}{R}"); err != nil {
		t.Fatal(err)
	}
	m := tbDecide(t, g, me, heuristic.DefaultConfig())
	if tbTurnsFaceUp(t, m) {
		t.Fatalf("the bot chose %q; a flip that kills nobody is a skipped turn for nothing", m.Label)
	}
}

// After combat the pump is worth nothing, even against a seat at 8.
func TestHeuristicDoesNotTurnUnstableHulkUpAfterCombat(t *testing.T) {
	g, me := tbTable(t, 3778, "")
	tbFaceDownHulk(t, g, me)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{R}{R}{R}{R}"); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { g.Seats[1-g.Turn.ActiveSeat].Life = 8 })
	advanceToStep(t, g, game.StepPostcombatMain)
	m := tbDecide(t, g, me, heuristic.DefaultConfig())
	if tbTurnsFaceUp(t, m) {
		t.Fatalf("the bot chose %q in the second main phase; the pump no longer attacks", m.Label)
	}
}
