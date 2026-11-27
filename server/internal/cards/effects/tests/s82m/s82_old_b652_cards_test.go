package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b652_cards_test.go — OLD batch 652 (Invasion, #5616): ten cards,
// the Knights, Djinns and self-pumpers. Each test plays the real catalog entry
// through the engine.

const (
	oldB652Battleflies = "58ed1294-ba7d-43a1-b988-8d78d9fa0e3e"
	oldB652Crusading   = "59385809-3223-4bc2-a824-9bab257c917a"
	oldB652Shambler    = "7947a210-32d0-4c99-82b4-4b5969447832"
	oldB652Marauding   = "8745b061-d5c6-46c1-a3fc-4d71d717f19e"
	oldB652Vanguard    = "9123086e-ae67-46c5-aed5-6366e2fe5a8c"
	oldB652Firescream  = "b86c0c23-bb29-4af0-bbd4-50b3ae634375"
	oldB652Golem       = "c54879da-0f1d-4ad0-b082-95f0e4d35685"
	oldB652Goham       = "e9db02fb-78e2-44b3-9e1b-ef0148234337"
	oldB652Tek         = "f85aef9b-b6ad-4716-b48b-5e26b474a59d"
	oldB652Halam       = "fb85d8d1-6938-48bf-9a6d-d4616b05b4f6"
)

func TestOldB652EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB652Battleflies: "Phyrexian Battleflies", oldB652Crusading: "Crusading Knight", oldB652Shambler: "Urborg Shambler",
		oldB652Marauding: "Marauding Knight", oldB652Vanguard: "Llanowar Vanguard", oldB652Firescream: "Firescreamer",
		oldB652Golem: "Sparring Golem", oldB652Goham: "Goham Djinn", oldB652Tek: "Tek", oldB652Halam: "Halam Djinn",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name=%q completeness=%v caveats=%v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB652Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return effectstest.ApaPush(g, owner, owner, game.Card{
		Name: name, TypeLine: typeLine, OracleID: oracle, Colors: colors, Power: power, Toughness: toughness,
	})
}

func oldB652Recompute(g *game.Game) {
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
}

func TestOldB652DjinnsShrinkWhileTheirColorIsMostCommon(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, color, kw string
		power, toughness        int
	}{
		{"Goham Djinn", oldB652Goham, "B", "", 5, 5},
		{"Halam Djinn", oldB652Halam, "R", "haste", 6, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := effectstest.P7Table(t)
			djinn := oldB652Push(g, me.ID, tc.name, "Creature — Djinn", tc.oracle, tc.power, tc.toughness, tc.color)
			if p, tough := effectstest.EffectivePTOf(t, g, djinn); p != tc.power-2 || tough != tc.toughness-2 {
				t.Fatalf("alone its colour leads: %d/%d, want %d/%d", p, tough, tc.power-2, tc.toughness-2)
			}
			oldB652Push(g, opp.ID, "Other One", "Creature — Test", "", 2, 2, "W")
			oldB652Push(g, opp.ID, "Other Two", "Creature — Test", "", 2, 2, "W")
			oldB652Recompute(g)
			if p, tough := effectstest.EffectivePTOf(t, g, djinn); p != tc.power || tough != tc.toughness {
				t.Fatalf("two white permanents lead: %d/%d, want %d/%d", p, tough, tc.power, tc.toughness)
			}
			oldB652Push(g, me.ID, "Tie", "Creature — Test", "", 2, 2, tc.color)
			oldB652Recompute(g)
			if p, tough := effectstest.EffectivePTOf(t, g, djinn); p != tc.power-2 || tough != tc.toughness-2 {
				t.Fatalf("tied for most common: %d/%d, want %d/%d", p, tough, tc.power-2, tc.toughness-2)
			}
			if tc.kw != "" && !effectstest.HasEffectiveKeyword(t, g, djinn, tc.kw) {
				t.Errorf("the Djinn lacks %s", tc.kw)
			}
		})
	}
}

func TestOldB652GohamDjinnRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	djinn := oldB652Push(g, me.ID, "Goham Djinn", "Creature — Djinn", oldB652Goham, 5, 5, "B")
	effectstest.FillPoolColored(me, "B", 1)
	effectstest.FillPool(me, 1)
	if err := effectstest.ActivateOn(t, g, me.ID, djinn, 0, uuid.Nil); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(djinn) })
	if !g.Battlefield.Contains(djinn) {
		t.Error("the regeneration shield did not save the Djinn")
	}
}

func TestOldB652KnightsCountOnlyOpponentsLands(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, land, landType, pro string
	}{
		{"Crusading Knight", oldB652Crusading, "Swamp", "Basic Land — Swamp", "protection from black"},
		{"Marauding Knight", oldB652Marauding, "Plains", "Basic Land — Plains", "protection from white"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := effectstest.P7Table(t)
			knight := oldB652Push(g, me.ID, tc.name, "Creature — Knight", tc.oracle, 2, 2, "W")
			if p, tough := effectstest.EffectivePTOf(t, g, knight); p != 2 || tough != 2 {
				t.Fatalf("no lands: %d/%d, want 2/2", p, tough)
			}
			oldB652Push(g, me.ID, "My "+tc.land, tc.landType, "", 0, 0)
			oldB652Push(g, opp.ID, "Their "+tc.land, tc.landType, "", 0, 0)
			oldB652Push(g, opp.ID, "Their Other "+tc.land, tc.landType, "", 0, 0)
			oldB652Push(g, opp.ID, "Their Mountain", "Basic Land — Mountain", "", 0, 0)
			oldB652Recompute(g)
			if p, tough := effectstest.EffectivePTOf(t, g, knight); p != 4 || tough != 4 {
				t.Errorf("two opposing %ss: %d/%d, want 4/4", tc.land, p, tough)
			}
			if !effectstest.HasEffectiveKeyword(t, g, knight, tc.pro) {
				t.Errorf("the Knight lacks %s", tc.pro)
			}
		})
	}
}

func TestOldB652UrborgShamblerShrinksOtherBlackCreatures(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	shambler := oldB652Push(g, me.ID, "Urborg Shambler", "Creature — Horror", oldB652Shambler, 4, 3, "B")
	mine := oldB652Push(g, me.ID, "My Black", "Creature — Test", "", 3, 3, "B")
	theirs := oldB652Push(g, opp.ID, "Their Black", "Creature — Test", "", 3, 3, "B")
	green := oldB652Push(g, opp.ID, "Green", "Creature — Test", "", 3, 3, "G")
	gold := oldB652Push(g, opp.ID, "Gold", "Creature — Test", "", 3, 3, "B", "G")
	oldB652Push(g, opp.ID, "Black Artifact", "Artifact", "", 0, 0, "B")
	oldB652Recompute(g)
	for _, tc := range []struct {
		id   uuid.UUID
		what string
		want int
	}{{shambler, "the Shambler", 4}, {mine, "your black creature", 2}, {theirs, "their black creature", 2}, {green, "a green creature", 3}, {gold, "a black-green creature", 2}} {
		if p, _ := effectstest.EffectivePTOf(t, g, tc.id); p != tc.want {
			t.Errorf("%s has power %d, want %d", tc.what, p, tc.want)
		}
	}
}

func TestOldB652BattlefliesPumpAtMostTwicePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	flies := oldB652Push(g, me.ID, "Phyrexian Battleflies", "Creature — Phyrexian Insect", oldB652Battleflies, 0, 1, "B")
	effectstest.FillPoolColored(me, "B", 3)
	for i := 0; i < 2; i++ {
		if err := effectstest.ActivateOn(t, g, me.ID, flies, 0, uuid.Nil); err != nil {
			t.Fatalf("pump %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if err := effectstest.ActivateOn(t, g, me.ID, flies, 0, uuid.Nil); err == nil {
		t.Fatal("a third activation in one turn was allowed")
	}
	if p, _ := effectstest.PtOf(t, g, flies); p != 2 {
		t.Errorf("power %d, want 2", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, flies, "flying") {
		t.Error("the Battleflies lack flying")
	}
}

func TestOldB652FirescreamerAndVanguardPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fire := oldB652Push(g, me.ID, "Firescreamer", "Creature — Kavu", oldB652Firescream, 2, 2, "B")
	vanguard := oldB652Push(g, me.ID, "Llanowar Vanguard", "Creature — Dryad", oldB652Vanguard, 1, 1, "G")
	effectstest.FillPoolColored(me, "R", 3)
	for i := 0; i < 3; i++ {
		if err := effectstest.ActivateOn(t, g, me.ID, fire, 0, uuid.Nil); err != nil {
			t.Fatalf("firebreathing %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tough := effectstest.PtOf(t, g, fire); p != 5 || tough != 2 {
		t.Errorf("Firescreamer %d/%d, want 5/2", p, tough)
	}
	if err := effectstest.ActivateOn(t, g, me.ID, vanguard, 0, uuid.Nil); err != nil {
		t.Fatalf("Vanguard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, vanguard); p != 1 || tough != 5 {
		t.Errorf("Llanowar Vanguard %d/%d, want 1/5", p, tough)
	}
	if err := effectstest.ActivateOn(t, g, me.ID, vanguard, 0, uuid.Nil); err == nil {
		t.Error("the Vanguard activated while tapped")
	}
}

func TestOldB652SparringGolemGrowsPerBlocker(t *testing.T) {
	for _, blockers := range []int{1, 2} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		golem := effectstest.B12Push(g, me.ID, "Sparring Golem", "Artifact Creature — Golem", oldB652Golem, 2, 2)
		var decls []game.BlockDeclaration
		for i := 0; i < blockers; i++ {
			b := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
			decls = append(decls, game.BlockDeclaration{Blocker: b, Attacker: golem})
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(golem, opp.ID); err != nil {
			t.Fatal(err)
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if err := g.DeclareBlockers(decls); err != nil {
			t.Fatal(err)
		}
		effectstest.LockInBlocks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, golem); p != 2+blockers || tough != 2+blockers {
			t.Errorf("%d blocker(s): %d/%d, want %d/%d", blockers, p, tough, 2+blockers, 2+blockers)
		}
	}
}

func TestOldB652TekGainsEachAbilityFromItsLand(t *testing.T) {
	for _, tc := range []struct {
		landType string
		power    int
		tough    int
		keyword  string
	}{
		{"Plains", 2, 4, ""},
		{"Island", 2, 2, "flying"},
		{"Swamp", 4, 2, ""},
		{"Mountain", 2, 2, "first strike"},
		{"Forest", 2, 2, "trample"},
	} {
		t.Run(tc.landType, func(t *testing.T) {
			g, me, opp := effectstest.P7Table(t)
			tek := oldB652Push(g, me.ID, "Tek", "Artifact Creature — Dragon", oldB652Tek, 2, 2)
			oldB652Recompute(g)
			for _, kw := range []string{"flying", "first strike", "trample"} {
				if effectstest.HasEffectiveKeyword(t, g, tek, kw) {
					t.Fatalf("a bare Tek has %s", kw)
				}
			}
			oldB652Push(g, opp.ID, "Their "+tc.landType, "Basic Land — "+tc.landType, "", 0, 0)
			oldB652Recompute(g)
			if p, tough := effectstest.EffectivePTOf(t, g, tek); p != 2 || tough != 2 {
				t.Fatalf("an opponent's %s counted: %d/%d", tc.landType, p, tough)
			}
			oldB652Push(g, me.ID, "My "+tc.landType, "Basic Land — "+tc.landType, "", 0, 0)
			oldB652Recompute(g)
			if p, tough := effectstest.EffectivePTOf(t, g, tek); p != tc.power || tough != tc.tough {
				t.Errorf("%d/%d, want %d/%d", p, tough, tc.power, tc.tough)
			}
			if tc.keyword != "" && !effectstest.HasEffectiveKeyword(t, g, tek, tc.keyword) {
				t.Errorf("Tek lacks %s", tc.keyword)
			}
		})
	}
}
