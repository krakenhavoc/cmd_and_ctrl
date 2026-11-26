package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b315_cards_test.go — Battle for Zendikar, batch 315 (S82, #5547).
// Each test plays the real catalog entry through the engine.

const (
	oldB315CoilsOracle    = "71af7e54-e22d-4354-a35c-80e4e5f4031c"
	oldB315WarleaderOracl = "784bda96-5bc1-47bc-a69e-e38b047ed914"
	oldB315ConduitOracle  = "cd3d79e9-c72e-4321-bca1-54505d51331a"
)

func TestOldB315EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB315CoilsOracle:    "Tightening Coils",
		oldB315WarleaderOracl: "Veteran Warleader",
		oldB315ConduitOracle:  "Conduit of Ruin",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Tightening Coils ------------------------------------------------------

func TestOldB315TighteningCoilsShrinksPowerAndRemovesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Push(g, opp.ID, "Dragon", "Creature — Dragon", "", 7, 5)
	other := effectstest.B12Push(g, opp.ID, "Other Dragon", "Creature — Dragon", "", 7, 5)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if id := g.Battlefield.Cards[i].InstanceID; id == victim || id == other {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
			}
		}
	})
	if !effectstest.HasEffectiveKeyword(t, g, victim, "flying") {
		t.Fatal("the test creature has no flying to begin with")
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Tightening Coils", TypeLine: "Enchantment — Aura", ManaCost: "{1}{U}", OracleID: oldB315CoilsOracle, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "U", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(victim)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, victim); p != 1 || tough != 5 {
		t.Errorf("enchanted creature is %d/%d, want 1/5", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, victim, "flying") {
		t.Error("enchanted creature kept flying")
	}
	if !effectstest.HasEffectiveKeyword(t, g, other, "flying") {
		t.Error("another creature lost flying")
	}
	effectstest.B14Kill(g, id)
	if p, _ := effectstest.PtOf(t, g, victim); p != 7 {
		t.Errorf("power %d after the Aura left, want 7", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, victim, "flying") {
		t.Error("flying did not come back when the Aura left")
	}
}

// --- Veteran Warleader -----------------------------------------------------

func TestOldB315VeteranWarleaderIsAsBigAsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	w := effectstest.B12Push(g, me.ID, "Veteran Warleader", "Creature — Human Soldier Ally", oldB315WarleaderOracl, 0, 0)
	if p, tough := effectstest.PtOf(t, g, w); p != 1 || tough != 1 {
		t.Errorf("alone it is %d/%d, want 1/1", p, tough)
	}
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear 2", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p, tough := effectstest.PtOf(t, g, w); p != 3 || tough != 3 {
		t.Errorf("with two other creatures it is %d/%d, want 3/3 (an opponent's creature does not count)", p, tough)
	}
}

func TestOldB315VeteranWarleaderTapsAnotherAllyForAChosenKeyword(t *testing.T) {
	for i, kw := range []string{"first strike", "vigilance", "trample"} {
		t.Run(kw, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			w := effectstest.B12Push(g, me.ID, "Veteran Warleader", "Creature — Human Soldier Ally", oldB315WarleaderOracl, 0, 0)
			ally := effectstest.B12Push(g, me.ID, "Other Ally", "Creature — Human Ally", "", 1, 1)
			bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{bear}}); err == nil {
				t.Fatal("a non-Ally creature paid the cost")
			}
			if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{w}}); err == nil {
				t.Fatal("the Warleader paid its own cost (it must be another Ally)")
			}
			effectstest.ApaActivate(t, g, me, w, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{ally}})
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerOptionPick(t, g, me.ID, i)
			if !effectstest.HasEffectiveKeyword(t, g, w, kw) {
				t.Errorf("the Warleader did not gain %s", kw)
			}
			if !effectstest.TappedOnBattlefield(t, g, ally) {
				t.Error("the other Ally was not tapped")
			}
			if effectstest.TappedOnBattlefield(t, g, w) {
				t.Error("the Warleader tapped itself")
			}
			if effectstest.HasEffectiveKeyword(t, g, bear, kw) {
				t.Errorf("another creature gained %s", kw)
			}
		})
	}
}

// --- Conduit of Ruin -------------------------------------------------------

func TestOldB315ConduitOfRuinFindsABigColorlessCreatureAndPutsItOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	big := game.Card{InstanceID: uuid.New(), Name: "Big Eldrazi", TypeLine: "Creature — Eldrazi", ManaCost: "{8}", Power: 8, Toughness: 8}
	small := game.Card{InstanceID: uuid.New(), Name: "Small Eldrazi", TypeLine: "Creature — Eldrazi", ManaCost: "{6}", Power: 6, Toughness: 6}
	green := game.Card{InstanceID: uuid.New(), Name: "Green Fatty", TypeLine: "Creature — Beast", ManaCost: "{6}{G}", Colors: []string{"G"}, Power: 7, Toughness: 7}
	effectstest.LibraryTop(me, green, small, big)
	// Bury the colorless card, so putting it on top is visible.
	effectstest.LibraryTop(me, effectstest.CardNamed("Filler 1", "Sorcery"), effectstest.CardNamed("Filler 2", "Sorcery"))
	effectstest.FloatForTest(g, me, "CCCCCC")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Conduit of Ruin", TypeLine: "Creature — Eldrazi", ManaCost: "{6}", OracleID: oldB315ConduitOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.SearchChoiceFor(g, me.ID)
	if p == nil {
		t.Fatal("no search prompt for the cast trigger")
	}
	if err := g.BeginSearchLibrary(p.ID, me.ID); err != nil {
		t.Fatalf("BeginSearchLibrary: %v", err)
	}
	p = effectstest.SearchChoiceFor(g, me.ID)
	if effectstest.HasID(p.SearchCards, small.InstanceID) || effectstest.HasID(p.SearchCards, green.InstanceID) {
		t.Error("a card that is not a colorless creature with mana value 7 or more is offered")
	}
	if err := g.ResolveSearchLibrary(p.ID, me.ID, []uuid.UUID{big.InstanceID}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if got := effectstest.LibraryTopIDs(me, 1); got[0] != big.InstanceID {
		t.Error("the found card is not on top of the library")
	}
}

func TestOldB315ConduitOfRuinDiscountsOnlyTheFirstCreatureSpellEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Conduit of Ruin", "Creature — Eldrazi", oldB315ConduitOracle, 5, 5)
	cast := func(name, typeLine string) error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{3}", Owner: me.ID, Controller: me.ID, Power: 1, Toughness: 1})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
	}
	effectstest.FloatForTest(g, me, "C")
	if err := cast("Sorcery", "Sorcery"); err == nil {
		t.Fatal("a noncreature spell was discounted")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	effectstest.B12Push(g2, me2.ID, "Conduit of Ruin", "Creature — Eldrazi", oldB315ConduitOracle, 5, 5)
	effectstest.FloatForTest(g2, me2, "C")
	id := uuid.New()
	me2.Hand.PushTop(game.Card{InstanceID: id, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{3}", Owner: me2.ID, Controller: me2.ID, Power: 1, Toughness: 1})
	if err := g2.CastSpell(me2.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the first creature spell did not cost {2} less: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.FloatForTest(g2, me2, "C")
	id2 := uuid.New()
	me2.Hand.PushTop(game.Card{InstanceID: id2, Name: "Bear 2", TypeLine: "Creature — Bear", ManaCost: "{3}", Owner: me2.ID, Controller: me2.ID, Power: 1, Toughness: 1})
	if err := g2.CastSpell(me2.ID, id2, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the second creature spell was discounted too")
	}
}
