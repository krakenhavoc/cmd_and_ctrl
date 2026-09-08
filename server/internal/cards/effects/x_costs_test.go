package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// x_costs_test.go — S20 sub-PR 3: X spells read the announced X at
// resolution; the cost gate charges it.

const (
	blazeOracle          = "0596920f-9946-42f4-a03b-24aab67f9f1b"
	exsanguinateOracle   = "8164b1e8-3350-465e-8a17-75f57d326344"
	strokeOfGeniusOracle = "0cc6d683-366f-4ae4-be60-20ad9621fdaf"
)

// castXSpell seeds an X card into the active seat's hand and casts
// it with the given X + targets (permissive cost mode — the S15
// gate warns rather than blocks without a pool).
func castXSpell(t *testing.T, g *game.Game, name, typeLine, oracle, manaCost string, x int, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets, XValue: x}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

func TestBlazeDealsXDamage(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life
	id := castXSpell(t, g, "Blaze", "Sorcery", blazeOracle, "{X}{R}", 4,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if item := g.StackMeta[id]; item == nil || item.XValue != 4 {
		t.Fatalf("stack item X = %+v, want 4", item)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("Blaze X=4: life %d -> %d, want -4", before, opp.Life)
	}
}

func TestBlazeXZeroDealsNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life
	castXSpell(t, g, "Blaze", "Sorcery", blazeOracle, "{X}{R}", 0,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("Blaze X=0 changed life: %d -> %d", before, opp.Life)
	}
}

func TestNegativeXRejected(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[0]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Blaze", TypeLine: "Sorcery", OracleID: blazeOracle,
		ManaCost: "{X}{R}", Owner: active.ID, Controller: active.ID})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	err := g.CastSpell(active.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}, XValue: -3,
	})
	if err != game.ErrInvalidParam {
		t.Fatalf("negative X: got %v, want ErrInvalidParam", err)
	}
}

func TestStrictModeChargesX(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[0]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	// Pool: {R} + 2 generic. X=3 needs {R} + 3 → insufficient; X=2 pays.
	active.ManaPool.AddMana(game.ManaToken{Color: "R"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Blaze", TypeLine: "Sorcery", OracleID: blazeOracle,
		ManaCost: "{X}{R}", Owner: active.ID, Controller: active.ID})
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}
	err := g.CastSpell(active.ID, id, game.CastSpellParams{Targets: target, XValue: 3, Strict: true})
	if err == nil {
		t.Fatalf("strict X=3 with {R}{C}{C} should be insufficient")
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Targets: target, XValue: 2, Strict: true}); err != nil {
		t.Fatalf("strict X=2: %v", err)
	}
	if len(active.ManaPool) != 0 {
		t.Errorf("X=2 should have spent the whole pool, %d tokens left", len(active.ManaPool))
	}
}

func TestExsanguinateDrainsEachOpponentAndGains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opps := []*game.Player{g.Seats[1], g.Seats[2], g.Seats[3]}
	for _, o := range opps {
		o.Life = 40
	}
	me.Life = 10
	castXSpell(t, g, "Exsanguinate", "Sorcery", exsanguinateOracle, "{X}{B}{B}", 5, nil)
	passPriorityAroundTable(t, g)
	for i, o := range opps {
		if o.Life != 35 {
			t.Errorf("opponent %d life = %d, want 35", i+1, o.Life)
		}
	}
	if me.Life != 25 {
		t.Errorf("caster life = %d, want 10 + 15", me.Life)
	}
}

func TestStrokeOfGeniusDrawsXForTarget(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Hand.Size()
	castXSpell(t, g, "Stroke of Genius", "Instant", strokeOfGeniusOracle, "{X}{2}{U}", 3,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - before; got != 3 {
		t.Errorf("Stroke X=3: target drew %d, want 3", got)
	}
}
