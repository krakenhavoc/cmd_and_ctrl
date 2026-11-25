package testsbr

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// choose_x_modes_test.go — #6125: "Choose X. You may choose the same
// mode more than once." (Doomsday Confluence {X}{X}{B}). The modes and X
// are announced together (CR 601.2b), the count of modes is X, and a mode
// may repeat (CR 700.2d).

const chooseXOracle = "00000000-6125-4000-8000-000000000003"

func registerChooseX(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:     chooseXOracle,
		Name:         "Test Confluence",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Modes: effects.ChooseXRepeating(
			effects.ModeDoing("You gain 1 life.", nil, func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.GainLife{Amount: 1}.Apply(ctx)
			}),
			effects.ModeDoing("You gain 2 life.", nil, func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.GainLife{Amount: 2}.Apply(ctx)
			}),
		),
	})
}

func chooseXHand(g *game.Game, me *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Confluence", OracleID: chooseXOracle, TypeLine: "Sorcery",
			ManaCost: "{X}{X}{B}", Owner: me.ID, Controller: me.ID})
	})
	return id
}

// X = 3 with the first mode twice and the second once: three modes, the
// same one repeated, for {X}{X}{B} = seven mana.
func TestChooseXTakesExactlyXModesWithRepeats(t *testing.T) {
	registerChooseX(t)
	g, me, _ := effectstest.P7Table(t)
	spell := chooseXHand(g, me)
	effectstest.FloatMana(t, g, me, "{B}{C}{C}{C}{C}{C}{C}")
	life := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 3, Modes: []int{0, 0, 1}}); err != nil {
		t.Fatalf("X=3, modes [0 0 1]: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Fatalf("life %d -> %d, want +4 (1 + 1 + 2)", life, me.Life)
	}
}

// The count of modes must be the announced X; X = 0 chooses none.
func TestChooseXRefusesAnyOtherCount(t *testing.T) {
	registerChooseX(t)
	g, me, _ := effectstest.P7Table(t)
	spell := chooseXHand(g, me)
	effectstest.FloatMana(t, g, me, "{B}{C}{C}{C}{C}")
	for name, p := range map[string]game.CastSpellParams{
		"one mode for X=2":  {XValue: 2, Modes: []int{0}},
		"two modes for X=1": {XValue: 1, Modes: []int{0, 1}},
		"a mode for X=0":    {XValue: 0, Modes: []int{0}},
	} {
		p.Strict = true
		if err := g.CastSpell(me.ID, spell, p); err == nil {
			t.Fatalf("%s: accepted", name)
		}
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 0}); err != nil {
		t.Fatalf("X=0 with no mode: %v", err)
	}
}

// The enumerator's move announces as many modes as its X.
func TestChooseXMoveAnnouncesAsManyModesAsX(t *testing.T) {
	registerChooseX(t)
	g, me, _ := effectstest.P7Table(t)
	spell := chooseXHand(g, me)
	effectstest.FloatMana(t, g, me, "{B}{C}{C}")
	n := 0
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeCastSpell || m.Source != spell {
			continue
		}
		var p struct {
			XValue int   `json:"x_value"`
			Modes  []int `json:"modes"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		n++
		if len(p.Modes) != p.XValue || p.XValue < 1 {
			t.Errorf("move announces X=%d with %d modes", p.XValue, len(p.Modes))
		}
	}
	if n == 0 {
		t.Fatal("no cast move offered")
	}
}
