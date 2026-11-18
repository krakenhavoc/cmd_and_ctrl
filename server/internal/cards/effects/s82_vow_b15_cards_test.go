package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
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
			spec, ok := Lookup(key)
			if !ok {
				t.Errorf("%s is not registered", names[i])
				continue
			}
			if spec.Name != names[i] || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
				t.Errorf("%s: name %q, completeness %s, caveats %v", names[i], spec.Name, spec.Completeness, spec.Caveats)
			}
		}
	}
}

// vowB15Row is a Scryfall record for a transform card: front and back faces.
func vowB15Row(oracle, front, frontType, frontCost, fp, ft, back, backType, bp, bt string, colors []string) cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     front + " // " + back,
		Layout:   "transform",
		TypeLine: frontType + " // " + backType,
		OracleID: uuid.MustParse(oracle),
		CardFaces: []cards.CardFace{
			{Name: front, TypeLine: frontType, ManaCost: frontCost, Colors: colors, Power: fp, Toughness: ft},
			{Name: back, TypeLine: backType, Colors: colors, Power: bp, Toughness: bt},
		},
	}
}

// vowB15Put puts the row's card onto p's battlefield, front face up, not
// summoning sick.
func vowB15Put(t *testing.T, g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	t.Helper()
	id := importToHand(row, p)
	c, ok := cardInZone(p.Hand, id)
	if !ok {
		t.Fatalf("%s did not reach the hand", row.Name)
	}
	p.Hand.Remove(id)
	pushBattlefieldCardWithTimestamp(g, c)
	return id
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
	if err := g.ActivateAbility(p.ID, id, game.AbilityParams{Label: label, Targets: refs}); err != nil {
		t.Fatalf("activate %q: %v", label, err)
	}
	passPriorityAroundTable(t, g)
}

// --- Alluring Suitor // Deadly Dancer ---------------------------------------

func vowB15SuitorRow() cards.Card {
	return vowB15Row(vowB15Suitor, "Alluring Suitor", "Creature — Vampire", "{2}{R}", "2", "3",
		"Deadly Dancer", "Creature — Vampire", "3", "3", []string{"R"})
}

func TestVowB15SuitorTransformsOnlyWhenYouAttackWithExactlyTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	suitor := vowB15Put(t, g, vowB15SuitorRow(), me)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	c := b12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, a, b, c)
	passPriorityAroundTable(t, g)
	if f := iwFace(g, suitor); f != 0 {
		t.Fatalf("three attackers turned the Suitor to face %d", f)
	}

	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	suitor = vowB15Put(t, g, vowB15SuitorRow(), me)
	a = b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	if f := iwFace(g, suitor); f != 0 {
		t.Fatalf("one attacker turned the Suitor to face %d", f)
	}

	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	suitor = vowB15Put(t, g, vowB15SuitorRow(), me)
	a = b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b = b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	if f := iwFace(g, suitor); f != 1 {
		t.Fatalf("two attackers left the Suitor on face %d, want the Dancer", f)
	}
	// The transform trigger gives {R}{R}, kept through the steps and phases.
	if n := me.ManaPool.Total(); n != 2 {
		t.Fatalf("mana after transforming = %d, want {R}{R}", n)
	}
	advanceTo(t, g, game.StepCombatDamage)
	if n := me.ManaPool.Total(); n != 2 {
		t.Errorf("the mana drained as the step ended: %d left, want 2", n)
	}
}

func TestVowB15DeadlyDancerPumpsItselfAndAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dancer := vowB15Put(t, g, vowB15SuitorRow(), me)
	iwTransform(t, g, dancer)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	base := effectivePower(t, g, dancer)
	vowB15Mana(me, "R", "R")
	vowB15Activate(t, g, me, dancer, "{R}{R}: This creature and another target creature each get +1/+0 until end of turn.", bear)
	if p := effectivePower(t, g, dancer); p != base+1 {
		t.Errorf("Dancer power %d, want %d", p, base+1)
	}
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("Bear power %d, want 3", p)
	}
	// "another" target: itself is not a legal target.
	vowB15Mana(me, "R", "R")
	if err := g.ActivateAbility(me.ID, dancer, game.AbilityParams{
		Label:   "{R}{R}: This creature and another target creature each get +1/+0 until end of turn.",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dancer}},
	}); err == nil {
		t.Error("the Dancer targeted itself")
	}
}

// --- Wedding Announcement // Wedding Festivity ------------------------------

func vowB15WeddingRow() cards.Card {
	return vowB15Row(vowB15Wedding, "Wedding Announcement", "Enchantment", "{2}{W}", "", "",
		"Wedding Festivity", "Enchantment", "", "", []string{"W"})
}

func vowB15EndStep(t *testing.T, g *game.Game) {
	t.Helper()
	advanceTo(t, g, game.StepEnd)
	monSettle(t, g)
	passPriorityAroundTable(t, g)
}

func TestVowB15WeddingMakesHumansDrawsAfterAttackingWithTwoAndTransformsAtThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := vowB15Put(t, g, vowB15WeddingRow(), me)
	vowB15EndStep(t, g)
	if n := countersOn(g, w, "invitation"); n != 1 {
		t.Fatalf("%d invitation counters, want 1", n)
	}
	if n := countBattlefieldNamed(g, me.ID, "Human"); n != 1 {
		t.Fatalf("no attack: %d Humans, want 1", n)
	}
	// A turn where two creatures attacked draws instead.
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.Turn.Step = game.StepPrecombatMain })
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	vowB15EndStep(t, g)
	if n := countersOn(g, w, "invitation"); n != 2 {
		t.Fatalf("%d invitation counters, want 2", n)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand grew by %d, want 1", me.Hand.Size()-hand)
	}
	if n := countBattlefieldNamed(g, me.ID, "Human"); n != 1 {
		t.Errorf("an attack turn made a Human: %d", n)
	}
}

func TestVowB15WeddingTransformsOnTheThirdCounterAndPumpsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	w := vowB15Put(t, g, vowB15WeddingRow(), me)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == w {
				g.Battlefield.Cards[i].Counters = map[string]int{"invitation": 2}
			}
		}
	})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	vowB15EndStep(t, g)
	if f := iwFace(g, w); f != 1 {
		t.Fatalf("Wedding Announcement is on face %d with three counters, want Festivity", f)
	}
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("Bear power %d under Wedding Festivity, want 3", p)
	}
	human := findBattlefieldByName(g, "Human")
	if p := effectivePower(t, g, human); p != 2 {
		t.Errorf("Human token power %d under Festivity, want 2", p)
	}
}

// --- Edgar, Charmed Groom // Edgar Markov's Coffin --------------------------

func vowB15EdgarRow() cards.Card {
	return vowB15Row(vowB15Edgar, "Edgar, Charmed Groom", "Legendary Creature — Vampire Noble", "{2}{W}{B}", "4", "4",
		"Edgar Markov's Coffin", "Legendary Artifact", "", "", []string{"W", "B"})
}

func TestVowB15EdgarPumpsOtherVampiresAndReturnsTransformedWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	edgar := vowB15Put(t, g, vowB15EdgarRow(), me)
	vamp := b12Creature(g, me.ID, "Vamp", "Creature — Vampire", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p := effectivePower(t, g, vamp); p != 3 {
		t.Errorf("Vampire power %d, want 3", p)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Vampire got the bonus: %d", p)
	}
	if p := effectivePower(t, g, edgar); p != 4 {
		t.Errorf("Edgar pumped himself: %d", p)
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, edgar, 9); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if f := iwFace(g, edgar); f != 1 {
		t.Fatalf("Edgar is on face %d after dying, want the Coffin", f)
	}
	c, _ := battlefieldCard(g, edgar)
	if c.Controller != me.ID {
		t.Errorf("the Coffin is controlled by %v, want its owner", c.Controller)
	}
}

func TestVowB15CoffinMakesLifelinkVampiresAndTransformsAtThreeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	edgar := vowB15Put(t, g, vowB15EdgarRow(), me)
	iwTransform(t, g, edgar)
	for i := 1; i <= 3; i++ {
		beginUpkeep(t, g, me.ID)
		if n := countBattlefieldNamed(g, me.ID, "Vampire"); n != i {
			t.Fatalf("upkeep %d: %d Vampires, want %d", i, n, i)
		}
		want, wantFace := i, 1
		if i == 3 {
			want, wantFace = 0, 0
		}
		if n := countersOn(g, edgar, "bloodline"); n != want {
			t.Fatalf("upkeep %d: %d bloodline counters, want %d", i, n, want)
		}
		if f := iwFace(g, edgar); f != wantFace {
			t.Fatalf("upkeep %d: face %d, want %d", i, f, wantFace)
		}
	}
	v := findBattlefieldByName(g, "Vampire")
	c, _ := battlefieldCard(g, v)
	if !c.HasColor("W") || !c.HasColor("B") || !c.IsToken() {
		t.Errorf("Vampire token is %+v, want a white and black token", c)
	}
	assertKeywords(t, g, v, "lifelink")
}

// --- Biolume Egg // Biolume Serpent -----------------------------------------

func vowB15EggRow() cards.Card {
	return vowB15Row(vowB15Egg, "Biolume Egg", "Creature — Serpent Egg", "{2}{U}", "0", "4",
		"Biolume Serpent", "Creature — Serpent", "4", "4", []string{"U"})
}

func TestVowB15EggReturnsTransformedAtTheNextEndStepWhenSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	egg := vowB15Put(t, g, vowB15EggRow(), me)
	if err := g.SacrificePermanent(me.ID, egg); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, egg); ok {
		t.Fatal("the Egg came back at once")
	}
	vowB15EndStep(t, g)
	c, ok := battlefieldCard(g, egg)
	if !ok {
		t.Fatal("the Egg never returned at the end step")
	}
	if c.ActiveFace != 1 || c.Controller != me.ID {
		t.Errorf("the Egg returned on face %d under %v, want the Serpent under its owner", c.ActiveFace, c.Controller)
	}
}

func TestVowB15EggDyingDoesNotReturn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	egg := vowB15Put(t, g, vowB15EggRow(), me)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, egg, 9); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	vowB15EndStep(t, g)
	if _, ok := battlefieldCard(g, egg); ok {
		t.Error("an Egg that died (not sacrificed) returned")
	}
}

func TestVowB15SerpentSacrificesTwoIslandsToBecomeUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	egg := vowB15Put(t, g, vowB15EggRow(), me)
	iwTransform(t, g, egg)
	label := "Sacrifice two Islands: This creature can't be blocked this turn."
	i1 := b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	if err := g.ActivateAbility(me.ID, egg, game.AbilityParams{Label: label}); err == nil {
		t.Fatal("activated with one Island")
	}
	i2 := b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	vowB15Activate(t, g, me, egg, label)
	for _, id := range []uuid.UUID{i1, i2} {
		if _, ok := battlefieldCard(g, id); ok {
			t.Error("an Island was not sacrificed")
		}
	}
}
