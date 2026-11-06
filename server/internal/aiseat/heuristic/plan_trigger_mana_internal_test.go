package heuristic

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// plan_trigger_mana_internal_test.go is #4770: with Config.PlanTriggerMana
// the plan's mana model counts what a row's trigger_adds names, the mana
// triggered mana abilities add when the source is tapped (Crypt Ghast's
// {B} on a Swamp), which CR 106.7 keeps out of `produced`.

// ghastSwamp is a Swamp as the view stamps it under Crypt Ghast.
func ghastSwamp() protocol.CardView {
	return planLand("{B}", func(c *protocol.CardView) { c.ManaAbilities[0].TriggerAdds = "{B}" })
}

func TestPlanTriggerManaIsOffByDefault(t *testing.T) {
	if DefaultConfig().PlanTriggerMana {
		t.Error("DefaultConfig: PlanTriggerMana is on")
	}
	if BaselineConfig().PlanTriggerMana {
		t.Error("BaselineConfig: PlanTriggerMana is on")
	}
}

func TestManaModelCountsTriggerManaWithTheSwitch(t *testing.T) {
	me := uuid.New()
	bf := []protocol.CardView{ghastSwamp(), ghastSwamp(), planLand("{G}")}
	v := planSeatView(me, bf)

	off := New()
	st := off.newState(aiseat.Input{Seat: me, View: v})
	if got := st.manaAvailable(off.manaModel(st)); len(got) != 3 {
		t.Errorf("off: %d units %#x, want the three lands' own mana", len(got), got)
	}

	cfg := DefaultConfig()
	cfg.PlanTriggerMana = true
	on := NewWithConfig(cfg)
	st = on.newState(aiseat.Input{Seat: me, View: v})
	units := st.manaAvailable(on.manaModel(st))
	if len(units) != 5 {
		t.Fatalf("on: %d units %#x, want five: {B}{B} per Swamp and the Forest's {G}", len(units), units)
	}
	for _, c := range []struct {
		cost string
		ok   bool
	}{
		{"{B}{B}{B}{B}{G}", true},
		{"{3}{B}{B}", true},
		{"{B}{B}{B}{B}{B}", false},
		{"{G}{G}", false},
	} {
		if _, ok := payMana(append([]manaUnit(nil), units...), parseManaCost(c.cost, 0), nil); ok != c.ok {
			t.Errorf("on: %s paid = %v, want %v", c.cost, ok, c.ok)
		}
	}
}

// A land with a {C} row and an {R} row under Mirari's Wake makes {C}{C}
// or {R}{R}, never {C}{R}. Without trigger mana the two rows merge into
// one mana of either; with it they are not merged, because the mask
// would then pay {C}{R}, which no tap makes. The source is the first row.
func TestManaModelDoesNotMergeRowsWithDifferentTriggerMana(t *testing.T) {
	me := uuid.New()
	land := protocol.CardView{Name: "Two Rows", TypeLine: "Land", ManaAbilities: []protocol.ManaAbilityView{
		{TapCost: true, Produced: "{C}", TriggerAdds: "{C}"},
		{TapCost: true, Produced: "{R}", TriggerAdds: "{R}"},
	}}
	v := planSeatView(me, []protocol.CardView{land})

	off := New()
	st := off.newState(aiseat.Input{Seat: me, View: v})
	if got := st.manaAvailable(off.manaModel(st)); len(got) != 1 || got[0] != manaC|manaR {
		t.Errorf("off: units %#x, want one mana of {C} or {R}", got)
	}

	cfg := DefaultConfig()
	cfg.PlanTriggerMana = true
	p := NewWithConfig(cfg)
	st = p.newState(aiseat.Input{Seat: me, View: v})
	units := st.manaAvailable(p.manaModel(st))
	if len(units) != 2 {
		t.Fatalf("on: units %#x, want one row's two mana", units)
	}
	if _, ok := payMana(append([]manaUnit(nil), units...), parseManaCost("{C}{R}", 0), nil); ok {
		t.Errorf("on: units %#x pay {C}{R}, which no tap of the land makes", units)
	}
	if _, ok := payMana(append([]manaUnit(nil), units...), parseManaCost("{2}", 0), nil); !ok {
		t.Errorf("on: units %#x do not pay {2}", units)
	}
}

func TestProducedUnitsReadsATriggersPick(t *testing.T) {
	got := producedUnits("{W|U|B|R|G}", nil)
	if len(got) != 1 || got[0] != manaAnyColor {
		t.Errorf("Fertile Ground's any colour = %#x, want one unit of any colour", got)
	}
	if got := producedUnits("{G}{G}", nil); len(got) != 2 || got[0] != manaG || got[1] != manaG {
		t.Errorf("{G}{G} = %#x, want two {G}", got)
	}
}
