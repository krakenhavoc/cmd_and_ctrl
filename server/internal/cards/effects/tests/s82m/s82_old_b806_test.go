package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b806_test.go — Mirage creatures with keywords, S82 OLD batch 806,
// played through the real catalog entries.

const (
	oldB806Melesse  = "1c4bd85b-03c2-4235-aea1-76c48520adf3"
	oldB806Elephant = "3b2ce431-7101-4256-827d-a14de9f867fd"
	oldB806Djinn    = "66fc64db-7dc4-43fa-8087-281a72286c35"
	oldB806Night    = "73016220-2436-4072-96dd-65c0cb05fff4"
	oldB806Wyvern   = "7a7f37a3-9e99-4d3d-b705-b074fb3e9b3e"
	oldB806Griffin  = "852b6e19-4594-41ea-9267-0911df643af5"
	oldB806Meerkat  = "871549e6-24c0-4fbd-8b71-ba8639b0ee47"
	oldB806Barrier  = "8783533d-dd98-4b80-b348-29051e1b5c28"
)

func TestOldB806CardsAreFullWithNoCaveats(t *testing.T) {
	for id, name := range map[string]string{
		oldB806Melesse: "Melesse Spirit", oldB806Elephant: "Wild Elephant",
		oldB806Djinn: "Vaporous Djinn", oldB806Night: "Spirit of the Night",
		oldB806Wyvern: "Cerulean Wyvern", oldB806Griffin: "Ekundu Griffin",
		oldB806Meerkat: "Karoo Meerkat", oldB806Barrier: "Blistering Barrier",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB806PrintedKeywordsAreOnTheBattlefieldCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
		have, lack             []string
	}{
		{"Melesse Spirit", "Creature — Angel Spirit", oldB806Melesse, 3, 3, []string{"flying", game.ProtectionFromColor("B")}, []string{"trample", game.ProtectionFromColor("W")}},
		{"Wild Elephant", "Creature — Elephant", oldB806Elephant, 3, 3, []string{"trample"}, []string{"flying"}},
		{"Spirit of the Night", "Legendary Creature — Demon Spirit", oldB806Night, 6, 5, []string{"flying", "trample", "haste", game.ProtectionFromColor("B")}, []string{"first strike"}},
		{"Cerulean Wyvern", "Creature — Drake", oldB806Wyvern, 3, 3, []string{"flying", game.ProtectionFromColor("G")}, []string{game.ProtectionFromColor("B")}},
		{"Ekundu Griffin", "Creature — Griffin", oldB806Griffin, 2, 2, []string{"flying", "first strike"}, []string{"trample"}},
		{"Karoo Meerkat", "Creature — Mongoose", oldB806Meerkat, 2, 1, []string{game.ProtectionFromColor("U")}, []string{"flying", game.ProtectionFromColor("R")}},
		{"Blistering Barrier", "Creature — Wall", oldB806Barrier, 5, 2, []string{"defender"}, []string{"flying"}},
	} {
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
		for _, kw := range tc.have {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("%s lacks %q", tc.name, kw)
			}
		}
		for _, kw := range tc.lack {
			if effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("%s has %q, which it does not print", tc.name, kw)
			}
		}
	}
}

func TestOldB806BlisteringBarrierCannotAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wall := effectstest.B12Push(g, me.ID, "Blistering Barrier", "Creature — Wall", oldB806Barrier, 5, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackersWith([]game.AttackDeclaration{{Attacker: wall, Target: g.Seats[1].ID}}, game.DeclareAttackersParams{}); err == nil {
		t.Error("a creature with defender attacked")
	}
}

func TestOldB806SpiritOfTheNightHasFirstStrikeOnlyWhileAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spirit := effectstest.B12Push(g, me.ID, "Spirit of the Night", "Legendary Creature — Demon Spirit", oldB806Night, 6, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if effectstest.HasEffectiveKeyword(t, g, spirit, "first strike") {
		t.Fatal("first strike before it attacks")
	}
	if _, err := g.DeclareAttackersWith([]game.AttackDeclaration{{Attacker: spirit, Target: g.Seats[1].ID}}, game.DeclareAttackersParams{}); err != nil {
		t.Fatalf("declare attackers: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, spirit, "first strike") {
		t.Error("no first strike while attacking")
	}
}

func TestOldB806VaporousDjinnPaysOrPhasesOut(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[1]
		djinn := effectstest.B12Push(g, me.ID, "Vaporous Djinn", "Creature — Djinn", oldB806Djinn, 3, 4)
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		if pay {
			effectstest.B06AddMana(me, "U", "U")
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		phased := effectstest.PhasedOutInCatalogGame(g, djinn)
		if phased == pay {
			t.Errorf("pay=%v: phased out = %v", pay, phased)
		}
	}
}
