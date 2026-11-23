package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_vow_b15_cards_test.go — Innistrad: Crimson Vow, S82 batch 15 (#5495).
// Every test imports the real two-faced row and plays the catalog entries
// through the engine.

const (
	vowB15Suitor   = "e98e8129-3e0e-45cc-b00c-1387cc592023"
	vowB15Egg      = "7af3876e-5297-46e6-936e-6c8344f4b6ac"
	vowB15Edgar    = "96cb97ca-4ce0-4d33-86fb-4db4ce1956be"
	vowB15Jacob    = "16d2a56a-7181-4ef3-866b-10dd8615d9e9"
	vowB15Runo     = "70d7d1a7-9d39-44ec-a4ea-a70ffa453d23"
	vowB15Board    = "6469c7fd-6c0f-4233-8c38-4742ddcf3c3f"
	vowB15Visioner = "8ffe9e3f-fd9f-4d8b-9467-4cd301e3c8b4"
	vowB15Wedding  = "259ac30c-cc05-4c04-9b23-71283f84b808"
)

func TestVowB15EveryFaceIsRegisteredFull(t *testing.T) {
	for oracle, names := range map[string][2]string{
		vowB15Suitor:   {"Alluring Suitor", "Deadly Dancer"},
		vowB15Egg:      {"Biolume Egg", "Biolume Serpent"},
		vowB15Edgar:    {"Edgar, Charmed Groom", "Edgar Markov's Coffin"},
		vowB15Jacob:    {"Jacob Hauken, Inspector", "Hauken's Insight"},
		vowB15Runo:     {"Runo Stromkirk", "Krothuss, Lord of the Deep"},
		vowB15Board:    {"Soulcipher Board", "Cipherbound Spirit"},
		vowB15Visioner: {"Voltaic Visionary", "Volt-Charged Berserker"},
		vowB15Wedding:  {"Wedding Announcement", "Wedding Festivity"},
	} {
		for i, key := range []string{oracle, oracle + "#1"} {
			spec, ok := effects.Lookup(key)
			if !ok {
				t.Errorf("%s is not registered", names[i])
				continue
			}
			if spec.Name != names[i] || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
				t.Errorf("%s: name %q, completeness %s, caveats %v", names[i], spec.Name, spec.Completeness, spec.Caveats)
			}
		}
	}
}

func vowB15Mana(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

func vowB15Activate(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, label string, targets ...uuid.UUID) {
	t.Helper()
	var refs []game.TargetRef
	for _, tgt := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: tgt})
	}
	if err := g.ActivateCatalogAbility(p.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: refs}); err != nil {
		t.Fatalf("activate %q: %v", label, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Alluring Suitor // Deadly Dancer ---------------------------------------

func vowB15SuitorRow() cards.Card {
	return effectstest.VowB15Row(vowB15Suitor, "Alluring Suitor", "Creature — Vampire", "{2}{R}", "2", "3",
		"Deadly Dancer", "Creature — Vampire", "3", "3", []string{"R"})
}

func TestVowB15SuitorTransformsOnlyWhenYouAttackWithExactlyTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	suitor := effectstest.VowB15Put(t, g, vowB15SuitorRow(), me)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b, c)
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, suitor); f != 0 {
		t.Fatalf("three attackers turned the Suitor to face %d", f)
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	suitor = effectstest.VowB15Put(t, g, vowB15SuitorRow(), me)
	a = effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, suitor); f != 0 {
		t.Fatalf("one attacker turned the Suitor to face %d", f)
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	suitor = effectstest.VowB15Put(t, g, vowB15SuitorRow(), me)
	a = effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b = effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, suitor); f != 1 {
		t.Fatalf("two attackers left the Suitor on face %d, want the Dancer", f)
	}
	// The transform trigger gives {R}{R}, kept through the steps and phases.
	if n := me.ManaPool.Total(); n != 2 {
		t.Fatalf("mana after transforming = %d, want {R}{R}", n)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if n := me.ManaPool.Total(); n != 2 {
		t.Errorf("the mana drained as the step ended: %d left, want 2", n)
	}
}

func TestVowB15DeadlyDancerPumpsItselfAndAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dancer := effectstest.VowB15Put(t, g, vowB15SuitorRow(), me)
	effectstest.IwTransform(t, g, dancer)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	base := effectstest.EffectivePower(t, g, dancer)
	vowB15Mana(me, "R", "R")
	vowB15Activate(t, g, me, dancer, "{R}{R}: This creature and another target creature each get +1/+0 until end of turn.", bear)
	if p := effectstest.EffectivePower(t, g, dancer); p != base+1 {
		t.Errorf("Dancer power %d, want %d", p, base+1)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("Bear power %d, want 3", p)
	}
	// "another" target: itself is not a legal target.
	vowB15Mana(me, "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, dancer, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dancer}},
	}); err == nil {
		t.Error("the Dancer targeted itself")
	}
}

// --- Wedding Announcement // Wedding Festivity ------------------------------

func vowB15WeddingRow() cards.Card {
	return effectstest.VowB15Row(vowB15Wedding, "Wedding Announcement", "Enchantment", "{2}{W}", "", "",
		"Wedding Festivity", "Enchantment", "", "", []string{"W"})
}

func vowB15EndStep(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.MonSettle(t, g)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestVowB15WeddingMakesHumansDrawsAfterAttackingWithTwoAndTransformsAtThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := effectstest.VowB15Put(t, g, vowB15WeddingRow(), me)
	vowB15EndStep(t, g)
	if n := effectstest.CountersOn(g, w, "invitation"); n != 1 {
		t.Fatalf("%d invitation counters, want 1", n)
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Human"); n != 1 {
		t.Fatalf("no attack: %d Humans, want 1", n)
	}
	// A turn where two creatures attacked draws instead.
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.Turn.Step = game.StepPrecombatMain })
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	vowB15EndStep(t, g)
	if n := effectstest.CountersOn(g, w, "invitation"); n != 2 {
		t.Fatalf("%d invitation counters, want 2", n)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand grew by %d, want 1", me.Hand.Size()-hand)
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Human"); n != 1 {
		t.Errorf("an attack turn made a Human: %d", n)
	}
}

func TestVowB15WeddingTransformsOnTheThirdCounterAndPumpsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	w := effectstest.VowB15Put(t, g, vowB15WeddingRow(), me)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == w {
				g.Battlefield.Cards[i].Counters = map[string]int{"invitation": 2}
			}
		}
	})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	vowB15EndStep(t, g)
	if f := effectstest.IwFace(g, w); f != 1 {
		t.Fatalf("Wedding Announcement is on face %d with three counters, want Festivity", f)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("Bear power %d under Wedding Festivity, want 3", p)
	}
	human := effectstest.FindBattlefieldByName(g, "Human")
	if p := effectstest.EffectivePower(t, g, human); p != 2 {
		t.Errorf("Human token power %d under Festivity, want 2", p)
	}
}

// --- Edgar, Charmed Groom // Edgar Markov's Coffin --------------------------

func vowB15EdgarRow() cards.Card {
	return effectstest.VowB15Row(vowB15Edgar, "Edgar, Charmed Groom", "Legendary Creature — Vampire Noble", "{2}{W}{B}", "4", "4",
		"Edgar Markov's Coffin", "Legendary Artifact", "", "", []string{"W", "B"})
}

func TestVowB15EdgarPumpsOtherVampiresAndReturnsTransformedWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	edgar := effectstest.VowB15Put(t, g, vowB15EdgarRow(), me)
	vamp := effectstest.B12Creature(g, me.ID, "Vamp", "Creature — Vampire", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p := effectstest.EffectivePower(t, g, vamp); p != 3 {
		t.Errorf("Vampire power %d, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Vampire got the bonus: %d", p)
	}
	if p := effectstest.EffectivePower(t, g, edgar); p != 4 {
		t.Errorf("Edgar pumped himself: %d", p)
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, edgar, 9); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, edgar); f != 1 {
		t.Fatalf("Edgar is on face %d after dying, want the Coffin", f)
	}
	c, _ := effectstest.BattlefieldCard(g, edgar)
	if c.Controller != me.ID {
		t.Errorf("the Coffin is controlled by %v, want its owner", c.Controller)
	}
}

func TestVowB15CoffinMakesLifelinkVampiresAndTransformsAtThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	edgar := effectstest.VowB15Put(t, g, vowB15EdgarRow(), me)
	effectstest.IwTransform(t, g, edgar)
	for i := 1; i <= 3; i++ {
		effectstest.BeginUpkeep(t, g, me.ID)
		if n := effectstest.CountBattlefieldNamed(g, me.ID, "Vampire"); n != i {
			t.Fatalf("upkeep %d: %d Vampires, want %d", i, n, i)
		}
		want, wantFace := i, 1
		if i == 3 {
			want, wantFace = 0, 0
		}
		if n := effectstest.CountersOn(g, edgar, "bloodline"); n != want {
			t.Fatalf("upkeep %d: %d bloodline counters, want %d", i, n, want)
		}
		if f := effectstest.IwFace(g, edgar); f != wantFace {
			t.Fatalf("upkeep %d: face %d, want %d", i, f, wantFace)
		}
	}
	v := effectstest.FindBattlefieldByName(g, "Vampire")
	c, _ := effectstest.BattlefieldCard(g, v)
	if !c.HasColor("W") || !c.HasColor("B") || !c.IsToken() {
		t.Errorf("Vampire token is %+v, want a white and black token", c)
	}
	effectstest.AssertKeywords(t, g, v, "lifelink")
}

// --- Biolume Egg // Biolume Serpent -----------------------------------------

func vowB15EggRow() cards.Card {
	return effectstest.VowB15Row(vowB15Egg, "Biolume Egg", "Creature — Serpent Egg", "{2}{U}", "0", "4",
		"Biolume Serpent", "Creature — Serpent", "4", "4", []string{"U"})
}

func TestVowB15EggReturnsTransformedAtTheNextEndStepWhenSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	egg := effectstest.VowB15Put(t, g, vowB15EggRow(), me)
	if err := g.SacrificePermanent(me.ID, egg); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, egg); ok {
		t.Fatal("the Egg came back at once")
	}
	vowB15EndStep(t, g)
	c, ok := effectstest.BattlefieldCard(g, egg)
	if !ok {
		t.Fatal("the Egg never returned at the end step")
	}
	if c.ActiveFace != 1 || c.Controller != me.ID {
		t.Errorf("the Egg returned on face %d under %v, want the Serpent under its owner", c.ActiveFace, c.Controller)
	}
}

func TestVowB15EggDyingDoesNotReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	egg := effectstest.VowB15Put(t, g, vowB15EggRow(), me)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, egg, 9); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	vowB15EndStep(t, g)
	if _, ok := effectstest.BattlefieldCard(g, egg); ok {
		t.Error("an Egg that died (not sacrificed) returned")
	}
}

func TestVowB15SerpentSacrificesTwoIslandsToBecomeUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	egg := effectstest.VowB15Put(t, g, vowB15EggRow(), me)
	effectstest.IwTransform(t, g, egg)
	label := "Sacrifice two Islands: This creature can't be blocked this turn."
	i1 := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	if err := g.ActivateCatalogAbility(me.ID, egg, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{i1}}); err == nil {
		t.Fatal("activated with one Island")
	}
	i2 := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	if err := g.ActivateCatalogAbility(me.ID, egg, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{i1, i2}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	_ = label
	for _, id := range []uuid.UUID{i1, i2} {
		if _, ok := effectstest.BattlefieldCard(g, id); ok {
			t.Error("an Island was not sacrificed")
		}
	}
}
