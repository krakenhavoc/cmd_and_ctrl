package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b651_cards_test.go — OLD batch 651 (Invasion, #5616): ten cards,
// from the colour-count Djinns to the Weavers. Each test plays the real
// catalog entry through the engine.

const (
	oldB651Sulam     = "01e311d2-8fe6-4e64-9575-3ef266985f0b"
	oldB651Zanam     = "049fc248-4e46-43de-986d-fa023a3e8792"
	oldB651Leech     = "09bfc3b1-cc40-4cb9-87e4-86b0018caa75"
	oldB651Tears     = "0decd8ea-a3d6-436b-81e9-bdd836537a1c"
	oldB651Ruham     = "16ad3b64-a87f-47e8-8e7b-b97bf084a3d2"
	oldB651RogueKavu = "17df95b2-49ef-43ca-b3c8-66a270f457a7"
	oldB651HateWeav  = "2a41f61b-a419-4b79-a25f-2506eaf05f55"
	oldB651Sunscape  = "4ab1411e-b255-4a6e-b81a-5de7b2d4ea12"
	oldB651SpiritWea = "4e6e28f1-111b-4bac-b814-487f6c206485"
	oldB651Skittish  = "55722e42-a729-4289-895f-be913398c434"
)

func TestOldB651EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB651Sulam: "Sulam Djinn", oldB651Zanam: "Zanam Djinn", oldB651Leech: "Andradite Leech",
		oldB651Tears: "Defiling Tears", oldB651Ruham: "Ruham Djinn", oldB651RogueKavu: "Rogue Kavu",
		oldB651HateWeav: "Hate Weaver", oldB651Sunscape: "Sunscape Apprentice", oldB651SpiritWea: "Spirit Weaver",
		oldB651Skittish: "Skittish Kavu",
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

// oldB651Push puts a coloured catalog permanent on the battlefield.
func oldB651Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return effectstest.ApaPush(g, owner, owner, game.Card{
		Name: name, TypeLine: typeLine, OracleID: oracle, Colors: colors, Power: power, Toughness: toughness,
	})
}

func oldB651Recompute(g *game.Game) {
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
}

func TestOldB651DjinnsShrinkWhileTheirColorIsMostCommon(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, color, kw string
		power, toughness        int
	}{
		{"Sulam Djinn", oldB651Sulam, "G", "trample", 6, 6},
		{"Zanam Djinn", oldB651Zanam, "U", "flying", 5, 6},
		{"Ruham Djinn", oldB651Ruham, "W", "first strike", 5, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := effectstest.P7Table(t)
			djinn := oldB651Push(g, me.ID, tc.name, "Creature — Djinn", tc.oracle, tc.power, tc.toughness, tc.color)
			if p, tough := effectstest.EffectivePTOf(t, g, djinn); p != tc.power-2 || tough != tc.toughness-2 {
				t.Fatalf("alone its colour leads: %d/%d, want %d/%d", p, tough, tc.power-2, tc.toughness-2)
			}
			other := "R"
			oldB651Push(g, opp.ID, "Other One", "Creature — Test", "", 2, 2, other)
			oldB651Push(g, opp.ID, "Other Two", "Creature — Test", "", 2, 2, other)
			oldB651Recompute(g)
			if p, tough := effectstest.EffectivePTOf(t, g, djinn); p != tc.power || tough != tc.toughness {
				t.Fatalf("two red permanents lead: %d/%d, want %d/%d", p, tough, tc.power, tc.toughness)
			}
			oldB651Push(g, me.ID, "Tie", "Creature — Test", "", 2, 2, tc.color)
			oldB651Recompute(g)
			if p, tough := effectstest.EffectivePTOf(t, g, djinn); p != tc.power-2 || tough != tc.toughness-2 {
				t.Fatalf("tied for most common: %d/%d, want %d/%d", p, tough, tc.power-2, tc.toughness-2)
			}
			if !effectstest.HasEffectiveKeyword(t, g, djinn, tc.kw) {
				t.Errorf("the Djinn lacks %s", tc.kw)
			}
		})
	}
}

func TestOldB651SkittishKavuNeedsNoWhiteOrBlueOpposition(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	kavu := oldB651Push(g, me.ID, "Skittish Kavu", "Creature — Kavu", oldB651Skittish, 1, 1, "R")
	if p, tough := effectstest.EffectivePTOf(t, g, kavu); p != 2 || tough != 2 {
		t.Fatalf("no opposition: %d/%d, want 2/2", p, tough)
	}
	oldB651Push(g, opp.ID, "Their Green", "Creature — Test", "", 2, 2, "G")
	oldB651Push(g, opp.ID, "Their White Land-Not-Creature", "Artifact", "", 0, 0, "W")
	oldB651Push(g, me.ID, "My White", "Creature — Test", "", 2, 2, "W")
	oldB651Recompute(g)
	if p, _ := effectstest.EffectivePTOf(t, g, kavu); p != 2 {
		t.Fatalf("green, a noncreature and your own white creature don't count: power %d, want 2", p)
	}
	blue := oldB651Push(g, opp.ID, "Their Blue", "Creature — Test", "", 2, 2, "U")
	oldB651Recompute(g)
	if p, tough := effectstest.EffectivePTOf(t, g, kavu); p != 1 || tough != 1 {
		t.Fatalf("an opposing blue creature: %d/%d, want 1/1", p, tough)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(blue) })
	oldB651Recompute(g)
	if p, _ := effectstest.EffectivePTOf(t, g, kavu); p != 2 {
		t.Fatalf("the blue creature left: power %d, want 2", p)
	}
	oldB651Push(g, opp.ID, "Their Gold", "Creature — Test", "", 2, 2, "W", "G")
	oldB651Recompute(g)
	if p, _ := effectstest.EffectivePTOf(t, g, kavu); p != 1 {
		t.Fatalf("a white-green opposing creature: power %d, want 1", p)
	}
}

func TestOldB651RogueKavuPumpsOnlyWhenAttackingAlone(t *testing.T) {
	for _, alone := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
		kavu := oldB651Push(g, me.ID, "Rogue Kavu", "Creature — Kavu", oldB651RogueKavu, 1, 1, "R")
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
		if err := g.DeclareAttacker(kavu, opp.ID); err != nil {
			t.Fatalf("attack: %v", err)
		}
		if !alone {
			if err := g.DeclareAttacker(bear, opp.ID); err != nil {
				t.Fatalf("second attacker: %v", err)
			}
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		want := 3
		if !alone {
			want = 1
		}
		if p, _ := effectstest.PtOf(t, g, kavu); p != want {
			t.Errorf("alone=%v: power %d, want %d", alone, p, want)
		}
	}
}

func TestOldB651WeaversPumpOnlyTheirColors(t *testing.T) {
	for _, tc := range []struct {
		name, oracle     string
		legal, illegal   string
		power, toughness int
		secondLegalColor string
	}{
		{"Hate Weaver", oldB651HateWeav, "U", "G", 1, 0, "R"},
		{"Spirit Weaver", oldB651SpiritWea, "G", "R", 0, 1, "U"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			weaver := oldB651Push(g, me.ID, tc.name, "Creature — Zombie Wizard", tc.oracle, 2, 1, "B")
			bad := oldB651Push(g, me.ID, "Wrong", "Creature — Test", "", 2, 2, tc.illegal)
			good := oldB651Push(g, me.ID, "Right", "Creature — Test", "", 2, 2, tc.legal)
			good2 := oldB651Push(g, me.ID, "Right Too", "Creature — Test", "", 2, 2, tc.secondLegalColor)
			effectstest.FillPool(me, 4)
			if err := effectstest.ActivateOn(t, g, me.ID, weaver, 0, bad); err == nil {
				t.Fatal("a creature of the wrong colour was a legal target")
			}
			for _, id := range []uuid.UUID{good, good2} {
				if err := effectstest.ActivateOn(t, g, me.ID, weaver, 0, id); err != nil {
					t.Fatalf("activate: %v", err)
				}
				effectstest.PassPriorityAroundTable(t, g)
				if p, tough := effectstest.PtOf(t, g, id); p != 2+tc.power || tough != 2+tc.toughness {
					t.Errorf("%d/%d, want %d/%d", p, tough, 2+tc.power, 2+tc.toughness)
				}
			}
			if p, tough := effectstest.PtOf(t, g, bad); p != 2 || tough != 2 {
				t.Errorf("the wrong-colour creature became %d/%d", p, tough)
			}
		})
	}
}

func TestOldB651SunscapeApprenticeAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	apprentice := oldB651Push(g, me.ID, "Sunscape Apprentice", "Creature — Human Wizard", oldB651Sunscape, 1, 1, "W")
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.FillPoolColored(me, "G", 1)
	if err := effectstest.ActivateOn(t, g, me.ID, apprentice, 0, theirs); err != nil {
		t.Fatalf("pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, theirs); p != 3 || tough != 3 {
		t.Errorf("the pumped creature is %d/%d, want 3/3", p, tough)
	}
	// The ability taps the Apprentice, so a second activation is refused.
	effectstest.FillPoolColored(me, "U", 1)
	if err := effectstest.ActivateOn(t, g, me.ID, apprentice, 1, mine); err == nil {
		t.Fatal("the Apprentice activated while tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(apprentice) })
	if err := effectstest.ActivateOn(t, g, me.ID, apprentice, 1, theirs); err == nil {
		t.Fatal("an opponent's creature was a legal target of the tuck")
	}
	if err := effectstest.ActivateOn(t, g, me.ID, apprentice, 1, mine); err != nil {
		t.Fatalf("tuck: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, mine); z != game.ZoneLibrary {
		t.Fatalf("the creature is in %s, want the library", z)
	}
	if n := len(me.Library.Cards); n == 0 || me.Library.Cards[n-1].InstanceID != mine {
		t.Error("the creature is not on top of its owner's library")
	}
}

func TestOldB651AndraditeLeechTaxesYourBlackSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Andradite Leech", "Creature — Leech", oldB651Leech, false)
	if got := effectstest.PriceInHand(t, g, me, "My Black", "Instant", "{B}"); got != 2 {
		t.Errorf("your black spell: %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "My Blue", "Instant", "{U}"); got != 1 {
		t.Errorf("your blue spell: %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Their Black", "Instant", "{B}"); got != 1 {
		t.Errorf("an opponent's black spell: %d, want 1", got)
	}
}

func TestOldB651AndraditeLeechPumpsForBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	leech := oldB651Push(g, me.ID, "Andradite Leech", "Creature — Leech", oldB651Leech, 2, 2, "B")
	effectstest.FillPoolColored(me, "B", 2)
	for i := 0; i < 2; i++ {
		if err := effectstest.ActivateOn(t, g, me.ID, leech, 0, uuid.Nil); err != nil {
			t.Fatalf("pump %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tough := effectstest.PtOf(t, g, leech); p != 4 || tough != 4 {
		t.Errorf("%d/%d after two pumps, want 4/4", p, tough)
	}
}

func TestOldB651DefilingTears(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := oldB651Push(g, me.ID, "Green Bear", "Creature — Bear", "", 2, 2, "G")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Defiling Tears", TypeLine: "Instant", OracleID: oldB651Tears, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 1 {
		t.Errorf("%d/%d, want 3/1", p, tough)
	}
	colors := effectstest.GaLayered(t, g, bear).EffectiveColors()
	if len(colors) != 1 || colors[0] != "B" {
		t.Errorf("colours %v, want only black", colors)
	}
	// The grant: {B}: Regenerate this creature.
	effectstest.FillPoolColored(me, "B", 1)
	if err := effectstest.ActivateOn(t, g, me.ID, bear, 0, uuid.Nil); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if !g.Battlefield.Contains(bear) {
		t.Error("the regeneration shield did not save the creature")
	}
}

func TestOldB651DefilingTearsKillsAToughnessOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elf := oldB651Push(g, me.ID, "Elf", "Creature — Elf", "", 1, 1, "G")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Defiling Tears", TypeLine: "Instant", OracleID: oldB651Tears, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: elf}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(elf) {
		t.Error("a 1/1 given -1 toughness survived")
	}
}
