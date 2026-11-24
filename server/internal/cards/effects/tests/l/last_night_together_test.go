package testsl

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// last_night_together_test.go — the two targets are untapped and pumped,
// a combat is added after this main phase, and only they can attack in
// it (#6087): not another creature, not one that arrives later. The
// turn's other combat is unrestricted.

const lastNightTogetherOracle = "0ad48edf-49e2-4f3a-b50e-11231a616752"

func TestLastNightTogetherOnlyTheChosenAttackInItsCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(a) })

	effectstest.CastCatalogSpell(t, g, "Last Night Together", "Sorcery", lastNightTogetherOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	main := g.Turn.PhaseID

	if effectstest.TappedNow(g, a) {
		t.Error("the tapped target was not untapped")
	}
	for _, id := range []uuid.UUID{a, b} {
		if n := effectstest.PlusOneCounters(g, id); n != 2 {
			t.Errorf("a target has %d +1/+1 counters, want 2", n)
		}
		for _, kw := range []string{"vigilance", "indestructible", "haste"} {
			if !effectstest.EffectiveAbilitiesContain(t, g, id, kw) {
				t.Errorf("a target lacks %s", kw)
			}
		}
	}
	if n := effectstest.PlusOneCounters(g, other); n != 0 {
		t.Errorf("the other creature got %d counters", n)
	}
	if len(g.PhaseEffects) != 1 {
		t.Fatalf("%d waiting rules, want 1", len(g.PhaseEffects))
	}

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	added := g.Turn.PhaseID
	if added == main {
		t.Fatal("no combat was added")
	}
	late := effectstest.B12Creature(g, opp.ID, "Late Arrival", "Creature — Bear", 2, 2)
	if effectstest.RestrictionsOf(t, g, late)&game.CantAttack == 0 {
		t.Error("a creature that arrived later may attack in the added combat")
	}
	if err := g.DeclareAttacker(other, opp.ID); !errors.Is(err, game.ErrCantAttack) {
		t.Errorf("another creature attacked in the added combat: err = %v", err)
	}
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Errorf("a chosen creature could not attack: %v", err)
	}
	if err := g.DeclareAttacker(b, opp.ID); err != nil {
		t.Errorf("the other chosen creature could not attack: %v", err)
	}

	// The turn's own combat follows, unrestricted.
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if g.Turn.PhaseID == added {
		t.Fatal("never left the added combat")
	}
	if effectstest.RestrictionsOf(t, g, other)&game.CantAttack != 0 {
		t.Error("the restriction outlived its combat")
	}
}

// TestLastNightTogetherNeedsTwoTargets: one target is not "two target
// creatures", so a strict cast with the mana for it is refused.
func TestLastNightTogetherNeedsTwoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Last Night Together", TypeLine: "Sorcery",
		OracleID: lastNightTogetherOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.AddPool(t, g, me.ID, "{R}{G}{C}{C}{C}")
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}}})
	if err == nil {
		t.Fatal("cast with one target was accepted")
	}
	if !me.Hand.Contains(id) {
		t.Fatal("the refused cast moved the card")
	}
	// The same pool pays for it with two targets.
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}}); err != nil {
		t.Fatalf("cast with two targets: %v", err)
	}
}
