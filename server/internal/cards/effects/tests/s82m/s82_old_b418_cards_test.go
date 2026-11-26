package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b418_cards_test.go — Return to Ravnica, batch 418 (S82 older sets,
// #5566): eight cards of pump and combat tricks. Every test plays the real
// catalog entry through the engine.

const (
	oldB418Chorus   = "0765c18b-25a0-4f79-aa1e-7adba7f5c91e"
	oldB418Krovod   = "1de561c7-4ff4-4eb9-8cbe-a29f568e786a"
	oldB418Steed    = "3517abe3-eb18-49af-9f1c-b372d2039d99"
	oldB418Shadow   = "3ba3405d-8c9a-4a80-811d-529cd3cb657f"
	oldB418Drake    = "4fee183e-973e-4be7-bb68-3e1c0c6bfb5b"
	oldB418Instinct = "92d2df7e-9f5c-4c28-a891-b69010bc0001"
	oldB418Weird    = "bf9b24e1-60d4-4739-808a-535336e191c2"
	oldB418Rites    = "ed1a1c04-1b2b-4616-abb2-8bd19c47b589"
)

func TestOldB418EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB418Chorus: "Chorus of Might", oldB418Krovod: "Bazaar Krovod", oldB418Steed: "Aquus Steed",
		oldB418Shadow: "Perilous Shadow", oldB418Drake: "Tower Drake", oldB418Instinct: "Traitorous Instinct",
		oldB418Weird: "Blistercoil Weird", oldB418Rites: "Rites of Reaping",
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

func oldB418Activate(t *testing.T, g *game.Game, me *game.Player, source uuid.UUID, mana string, targets []game.TargetRef) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, mana)
	if err := g.ActivateCatalogAbility(me.ID, source, 0, game.ActivateAbilityParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB418TowerDrakeFliesAndGrowsInToughnessForW(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	drake := effectstest.B12Push(g, me.ID, "Tower Drake", "Creature — Drake", oldB418Drake, 2, 1)
	if !effectstest.HasEffectiveKeyword(t, g, drake, "flying") {
		t.Error("the Drake has no flying")
	}
	oldB418Activate(t, g, me, drake, "{W}", nil)
	if p, tough := effectstest.PtOf(t, g, drake); p != 2 || tough != 2 {
		t.Errorf("the Drake is %d/%d, want 2/2", p, tough)
	}
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, drake, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the pump was paid with blue mana")
	}
}

func TestOldB418PerilousShadowStacksItsPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	shade := effectstest.B12Push(g, me.ID, "Perilous Shadow", "Creature — Insect Shade", oldB418Shadow, 0, 4)
	oldB418Activate(t, g, me, shade, "{C}{B}", nil)
	oldB418Activate(t, g, me, shade, "{C}{B}", nil)
	if p, tough := effectstest.PtOf(t, g, shade); p != 4 || tough != 8 {
		t.Errorf("the Shadow is %d/%d, want 4/8", p, tough)
	}
}

func TestOldB418AquusSteedShrinksAnyCreatureAndTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	steed := effectstest.B12Push(g, me.ID, "Aquus Steed", "Creature — Beast", oldB418Steed, 1, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	oldB418Activate(t, g, me, steed, "{C}{C}{U}", effectstest.CardTarget(victim))
	if p, tough := effectstest.PtOf(t, g, victim); p != 1 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 1/3", p, tough)
	}
	if !effectstest.B16Tapped(t, g, steed) {
		t.Error("the Steed did not tap")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, steed, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim)}); err == nil {
		t.Error("a tapped Steed activated again")
	}
}

func TestOldB418BlistercoilWeirdGrowsAndUntapsOnInstantsAndSorceriesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	weird := effectstest.B12Push(g, me.ID, "Blistercoil Weird", "Creature — Weird", oldB418Weird, 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.TapForTest(t, g, weird)
	cast := func(name, typeLine string) {
		id := effectstest.HandCardFull(me, name, typeLine, "", "", nil)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	cast("Bear", "Creature — Bear")
	if p, _ := effectstest.PtOf(t, g, weird); p != 1 || !effectstest.B16Tapped(t, g, weird) {
		t.Errorf("a creature spell moved the Weird: power %d, tapped %v", p, effectstest.B16Tapped(t, g, weird))
	}
	cast("Shock", "Instant")
	if p, tough := effectstest.PtOf(t, g, weird); p != 2 || tough != 2 || effectstest.B16Tapped(t, g, weird) {
		t.Errorf("after an instant the Weird is %d/%d tapped %v, want 2/2 untapped", p, tough, effectstest.B16Tapped(t, g, weird))
	}
	cast("Divination", "Sorcery")
	if p, tough := effectstest.PtOf(t, g, weird); p != 3 || tough != 3 {
		t.Errorf("after a sorcery the Weird is %d/%d, want 3/3", p, tough)
	}
}

func TestOldB418BazaarKrovodPumpsAndUntapsAnotherAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	krovod := effectstest.B12Push(g, me.ID, "Bazaar Krovod", "Creature — Beast", oldB418Krovod, 2, 5)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, krovod, bear)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack-trigger target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, krovod) {
		t.Error("the Krovod offered itself as a target")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 2/4", p, tough)
	}
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("the Bear is still tapped")
	}
}

func TestOldB418ChorusOfMightCountsYourCreaturesAndGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Chorus of Might", "Instant", oldB418Chorus, effectstest.CardTarget(a))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, a); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4 (two creatures you control)", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, a, "trample") {
		t.Error("the target has no trample")
	}
}

func TestOldB418ChorusOfMightCanPumpAnOpponentsCreatureByYourCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Chorus of Might", "Instant", oldB418Chorus, effectstest.CardTarget(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("the target is %d/%d, want 2/2 (one creature you control)", p, tough)
	}
}

func TestOldB418TraitorousInstinctStealsUntapsPumpsAndGivesHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	victim := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, victim)
	effectstest.CastCatalogSpell(t, g, "Traitorous Instinct", "Sorcery", oldB418Instinct, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, victim) != me.ID {
		t.Error("the creature was not stolen")
	}
	if effectstest.B16Tapped(t, g, victim) {
		t.Error("the creature is still tapped")
	}
	if p, _ := effectstest.PtOf(t, g, victim); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, victim, "haste") {
		t.Error("the creature has no haste")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB418RitesOfReapingPumpsOneAndShrinksAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Rites of Reaping", "Sorcery", oldB418Rites, effectstest.CardRefs(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mine); p != 5 || tough != 5 {
		t.Errorf("the first target is %d/%d, want 5/5", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("the second target is %d/%d, want 1/1", p, tough)
	}
}

func TestOldB418RitesOfReapingRefusesTheSameCreatureTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	id := effectstest.HandCardFull(me, "Rites of Reaping", "Sorcery", "{4}{B}{G}", oldB418Rites, []string{"B", "G"})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(mine, mine)}); err == nil {
		t.Error("Rites of Reaping took one creature for both targets")
	}
}
