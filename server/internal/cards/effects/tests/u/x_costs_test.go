package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestBlazeDealsXDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life
	id := effectstest.CastXSpell(t, g, "Blaze", "Sorcery", effectstest.BlazeOracle, "{X}{R}", 4,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if item := g.StackMeta[id]; item == nil || item.XValue != 4 {
		t.Fatalf("stack item X = %+v, want 4", item)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("Blaze X=4: life %d -> %d, want -4", before, opp.Life)
	}
}

func TestBlazeXZeroDealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life
	effectstest.CastXSpell(t, g, "Blaze", "Sorcery", effectstest.BlazeOracle, "{X}{R}", 0,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("Blaze X=0 changed life: %d -> %d", before, opp.Life)
	}
}

func TestNegativeXRejected(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[0]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Blaze", TypeLine: "Sorcery", OracleID: effectstest.BlazeOracle,
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
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[0]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	// Pool: {R} + 2 generic. X=3 needs {R} + 3 → insufficient; X=2 pays.
	active.ManaPool.AddMana(game.ManaToken{Color: "R"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Blaze", TypeLine: "Sorcery", OracleID: effectstest.BlazeOracle,
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opps := []*game.Player{g.Seats[1], g.Seats[2], g.Seats[3]}
	for _, o := range opps {
		o.Life = 40
	}
	me.Life = 10
	effectstest.CastXSpell(t, g, "Exsanguinate", "Sorcery", effectstest.ExsanguinateOracle, "{X}{B}{B}", 5, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Hand.Size()
	effectstest.CastXSpell(t, g, "Stroke of Genius", "Instant", effectstest.StrokeOfGeniusOracle, "{X}{2}{U}", 3,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - before; got != 3 {
		t.Errorf("Stroke X=3: target drew %d, want 3", got)
	}
}
