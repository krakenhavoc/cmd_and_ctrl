package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// warp_test.go — S29, #324. Warp is three shipped mechanisms in a
// trench coat (an alternative cost, a delayed trigger, an exile-play
// grant), so the tests are about the seams between them rather than
// about any one of them:
//
//  1. The discount is real — the creature enters for the warp cost.
//  2. The take-back is real — it is gone at the end step, and gone
//     to EXILE rather than to a graveyard, so it can come back.
//  3. "On a later turn" is real. This is the one a naive
//     implementation gets wrong, because an unbounded WhileExiled
//     grant is live the instant it is stamped — and it is stamped
//     during the end step of the turn the creature was warped in.

// weftstalkerWarpOracle is the same key pirates_batch2_test.go
// already names; spelled again here so this file reads on its own.
const weftstalkerWarpOracle = "926d52a5-4db1-46ce-9567-17c28bf56ae7"

// warpCreature seeds Weftstalker Ardent in the active seat's hand at
// a main phase and casts it for its warp cost.
func warpCreature(t *testing.T, g *game.Game) (uuid.UUID, *game.Player) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Weftstalker Ardent",
		TypeLine:   "Creature — Drix Artificer",
		ManaCost:   "{2}{R}",
		OracleID:   weftstalkerWarpOracle,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{AlternativeCost: "warp"}); err != nil {
		t.Fatalf("warp cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	return id, active
}

// advanceThroughEndStep walks the cursor to the end step, which is
// where the warp trigger fires, and then lets it resolve.
func advanceThroughEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
}

// The discount, and the cost it replaces. A warped Weftstalker
// Ardent costs {R}, not {2}{R} and not both.
func TestWarpChargesTheWarpCostNotThePrinted(t *testing.T) {
	t.Run("one red is enough", func(t *testing.T) {
		g := newCatalogGame(t)
		active := g.Seats[g.Turn.ActiveSeat]
		id := uuid.New()
		active.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Weftstalker Ardent",
			TypeLine: "Creature — Drix Artificer", ManaCost: "{2}{R}",
			OracleID: weftstalkerWarpOracle, Owner: active.ID, Controller: active.ID,
		})
		for g.Turn.Step != game.StepPrecombatMain {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
		active.ManaPool.AddMana(game.ManaToken{Color: "R"})

		if err := g.CastSpell(active.ID, id, game.CastSpellParams{
			Strict: true, AlternativeCost: "warp",
		}); err != nil {
			t.Fatalf("warp cast on one red: %v", err)
		}
		if len(active.ManaPool) != 0 {
			t.Errorf("warp cost not charged: pool %+v", active.ManaPool)
		}
	})

	t.Run("one red is not enough for the printed cost", func(t *testing.T) {
		g := newCatalogGame(t)
		active := g.Seats[g.Turn.ActiveSeat]
		id := uuid.New()
		active.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Weftstalker Ardent",
			TypeLine: "Creature — Drix Artificer", ManaCost: "{2}{R}",
			OracleID: weftstalkerWarpOracle, Owner: active.ID, Controller: active.ID,
		})
		for g.Turn.Step != game.StepPrecombatMain {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
		active.ManaPool.AddMana(game.ManaToken{Color: "R"})

		if err := g.CastSpell(active.ID, id, game.CastSpellParams{Strict: true}); err == nil {
			t.Fatalf("printed cost paid off one red mana")
		}
	})
}

// The take-back. The creature really enters — so its trigger is live
// while it is there — and at the beginning of the next end step it
// is exiled, not destroyed and not sacrificed.
func TestWarpExilesAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	id, active := warpCreature(t, g)

	if !g.Battlefield.Contains(id) {
		t.Fatalf("the warped creature did not enter the battlefield")
	}
	if len(g.DelayedTriggers) == 0 {
		t.Errorf("warp scheduled no delayed trigger")
	}

	advanceThroughEndStep(t, g)

	if g.Battlefield.Contains(id) {
		t.Errorf("the warped creature is still on the battlefield at the end step")
	}
	if active.Graveyard.Contains(id) {
		t.Errorf("the warped creature went to the graveyard — it must be exiled")
	}
	if !g.Exile.Contains(id) {
		t.Fatalf("the warped creature is not in exile")
	}
}

// "On a later turn", and the reason ExilePlayPermission grew a
// floor. The grant is stamped during the end step of the turn the
// creature was warped in, so an unbounded grant with no floor would
// be live immediately.
func TestWarpGrantIsDarkUntilTheNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	id, active := warpCreature(t, g)
	warpTurn := g.Turn.Number
	advanceThroughEndStep(t, g)

	var grant game.ExilePlayPermission
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			grant = c.ExilePlay
		}
	}
	if !grant.Granted() {
		t.Fatalf("no exile-play grant on the warped creature")
	}
	if grant.Player != active.ID {
		t.Errorf("grant names %v, want the warping player %v", grant.Player, active.ID)
	}
	if !grant.WhileExiled {
		t.Errorf("warp's grant is not unbounded — it lasts as long as the card is exiled")
	}
	if grant.NotBeforeTurn != warpTurn+1 {
		t.Errorf("NotBeforeTurn: got %d, want %d", grant.NotBeforeTurn, warpTurn+1)
	}
	if grant.Active(active.ID, warpTurn) {
		t.Errorf("the grant is live on the turn the creature was warped")
	}
	if !grant.Active(active.ID, warpTurn+1) {
		t.Errorf("the grant is not live on the next turn")
	}
}

// And the payoff: on a later turn the creature is cast out of exile
// for its PRINTED cost, through the ordinary exile cast path.
func TestWarpedCreatureIsRecastableFromExileLater(t *testing.T) {
	g := newCatalogGame(t)
	id, active := warpCreature(t, g)
	advanceThroughEndStep(t, g)

	// Walk to this seat's next turn.
	for i := 0; i < 64 && !(g.Turn.Step == game.StepPrecombatMain && g.Seats[g.Turn.ActiveSeat].ID == active.ID); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Seats[g.Turn.ActiveSeat].ID != active.ID || g.Turn.Step != game.StepPrecombatMain {
		t.Fatalf("never reached the warping seat's next main phase")
	}
	active.ManaPool.AddMana(game.ManaToken{Color: "R"})
	active.ManaPool.AddMana(game.ManaToken{Color: "R"})
	active.ManaPool.AddMana(game.ManaToken{Color: "R"})

	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("recast from exile: %v", err)
	}
	if len(active.ManaPool) != 0 {
		t.Errorf("the recast did not charge the printed {2}{R}: pool %+v", active.ManaPool)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Errorf("the recast creature is not on the battlefield")
	}
}

// A Weftstalker Ardent cast for its printed cost is an ordinary
// creature: no exile, no grant, no delayed trigger. The clause
// belongs to the cost that was paid.
func TestPrintedCostCastOfAWarpCardStaysOnTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Weftstalker Ardent",
		TypeLine: "Creature — Drix Artificer", ManaCost: "{2}{R}",
		OracleID: weftstalkerWarpOracle, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("printed-cost cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("a printed-cost cast scheduled a warp trigger: %+v", g.DelayedTriggers)
	}
	advanceThroughEndStep(t, g)
	if !g.Battlefield.Contains(id) {
		t.Errorf("a printed-cost Weftstalker Ardent left the battlefield at the end step")
	}
}
