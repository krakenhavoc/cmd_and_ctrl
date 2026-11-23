package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b205_test.go — cards missing from sets since AFR, batch 205 (#5948):
// Double Masters 2022 creatures. Each test plays the real catalog entry.

const (
	gap205Figure     = "cd51cc2d-0d39-4aba-b0ea-70e8118149bc"
	gap205Legionnair = "32fad8d3-5351-49ce-a453-53dee403b5ff"
	gap205Stitchwing = "7e5d2801-873f-4318-9c23-9a73e00e31a4"
	gap205Disciple   = "f06209e6-475e-47e0-b84c-3ec1215633a9"
	gap205Elder      = "dfbd5ad1-0f68-4219-95c4-98f674f321be"
	gap205Kederekt   = "edf459c0-53e4-47e8-8098-105656f9b49c"
	gap205Mistfire   = "794255b1-d94d-4de4-ac0a-87ef743f9102"
	gap205Smuggler   = "f96a2aa6-3259-4bb6-92cd-4d17215d04d0"
)

func TestGap205EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gap205Figure:     "Figure of Destiny",
		gap205Legionnair: "Tenth District Legionnaire",
		gap205Stitchwing: "Advanced Stitchwing",
		gap205Disciple:   "Disciple of the Ring",
		gap205Elder:      "Jeskai Elder",
		gap205Kederekt:   "Kederekt Leviathan",
		gap205Mistfire:   "Mistfire Adept",
		gap205Smuggler:   "Nephalia Smuggler",
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

// --- Figure of Destiny -------------------------------------------------------

func TestGap205FigureOfDestinyGrowsThroughThreeStages(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	fig := effectstest.PushCatalogPermanent(g, me.ID, "Figure of Destiny", "Creature — Kithkin", gap205Figure, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == fig {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 1, 1
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	act := func(idx int, n int) {
		t.Helper()
		mana := make([]string, n)
		for i := range mana {
			mana[i] = "R"
		}
		effectstest.B06AddMana(me, mana...)
		effectstest.B16Activate(t, g, me.ID, fig, idx, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
	}
	act(1, 3) // not a Spirit yet: nothing
	if p, tt := effectstest.PtOf(t, g, fig); p != 1 || tt != 1 {
		t.Fatalf("a non-Spirit stage 2 changed it to %d/%d", p, tt)
	}
	act(2, 6) // not a Warrior yet: nothing
	if p, tt := effectstest.PtOf(t, g, fig); p != 1 || tt != 1 {
		t.Fatalf("a non-Warrior stage 3 changed it to %d/%d", p, tt)
	}
	act(0, 1)
	c, _ := effectstest.BattlefieldCard(g, fig)
	if p, tt := effectstest.PtOf(t, g, fig); p != 2 || tt != 2 || !c.HasSubtype("Spirit") || !c.HasSubtype("Kithkin") {
		t.Fatalf("stage 1: %d/%d %v", p, tt, c.Effective().Subtypes)
	}
	act(1, 3)
	c, _ = effectstest.BattlefieldCard(g, fig)
	if p, tt := effectstest.PtOf(t, g, fig); p != 4 || tt != 4 || !c.HasSubtype("Warrior") || !c.HasSubtype("Spirit") {
		t.Fatalf("stage 2: %d/%d %v", p, tt, c.Effective().Subtypes)
	}
	act(2, 6)
	c, _ = effectstest.BattlefieldCard(g, fig)
	if p, tt := effectstest.PtOf(t, g, fig); p != 8 || tt != 8 || !c.HasSubtype("Avatar") || !c.HasSubtype("Kithkin") {
		t.Fatalf("stage 3: %d/%d %v", p, tt, c.Effective().Subtypes)
	}
	for _, kw := range []string{"flying", "first strike"} {
		if !effectstest.HasEffectiveKeyword(t, g, fig, kw) {
			t.Errorf("no %s at stage 3", kw)
		}
	}
	effectstest.Ecl6NextTurn(t, g)
	if p, _ := effectstest.PtOf(t, g, fig); p != 8 {
		t.Error("the effect ended with the turn")
	}
}

// --- Tenth District Legionnaire ----------------------------------------------

func TestGap205LegionnaireHeroicAddsACounterOnlyWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	leg := effectstest.PushCatalogPermanent(g, me.ID, "Tenth District Legionnaire", "Creature — Human Soldier", gap205Legionnair, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Filler", "Instant", "", []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.SettleProwess(t, g)
	if c, _ := effectstest.BattlefieldCard(g, leg); c.Counters[game.CounterPlusOne] != 0 {
		t.Fatalf("%d counters for a spell at another creature", c.Counters[game.CounterPlusOne])
	}
	effectstest.CastCatalogSpell(t, g, "Filler", "Instant", "", []game.TargetRef{{Kind: game.TargetCard, ID: leg}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry 1 after the counter")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	if c, _ := effectstest.BattlefieldCard(g, leg); c.Counters[game.CounterPlusOne] != 1 {
		t.Fatalf("%d +1/+1 counters after targeting it, want 1", c.Counters[game.CounterPlusOne])
	}
}

// --- Advanced Stitchwing -----------------------------------------------------

func TestGap205StitchwingReturnsTappedFromTheGraveyardForTwoDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Advanced Stitchwing", TypeLine: "Creature — Zombie Horror",
			OracleID: gap205Stitchwing, Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID})
	})
	a := effectstest.PutInHand(me, game.Card{InstanceID: uuid.New(), Name: "A", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	b := effectstest.PutInHand(me, game.Card{InstanceID: uuid.New(), Name: "B", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "C", "C", "U")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{a, b}})
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok || !c.Tapped {
		t.Fatalf("on the battlefield %v, tapped %v; want back and tapped", ok, c.Tapped)
	}
	if effectstest.InHandOf(me, a) || effectstest.InHandOf(me, b) {
		t.Errorf("the two cards were not discarded")
	}
}

// --- Jeskai Elder ------------------------------------------------------------

func TestGap205JeskaiElderHasProwess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	el := effectstest.PushCatalogPermanent(g, me.ID, "Jeskai Elder", "Creature — Human Monk", gap205Elder, false)
	if !effectstest.HasEffectiveKeyword(t, g, el, "prowess") {
		t.Fatal("no prowess")
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, el); p != 2 {
		t.Errorf("power %d once the trigger resolved, want 2", p)
	}
}

// --- Kederekt Leviathan ------------------------------------------------------

func TestGap205KederektBouncesEveryOtherNonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	lev := effectstest.Ecl6Cast(t, g, me, "Kederekt Leviathan", "Creature — Leviathan", gap205Kederekt, 5, 5, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, lev); !ok {
		t.Error("the Leviathan bounced itself")
	}
	if _, ok := effectstest.BattlefieldCard(g, mine); ok {
		t.Error("my Bear stayed")
	}
	if _, ok := effectstest.BattlefieldCard(g, theirs); ok {
		t.Error("their Wolf stayed")
	}
	if _, ok := effectstest.BattlefieldCard(g, land); !ok {
		t.Error("a land was bounced")
	}
}

// --- Mistfire Adept ----------------------------------------------------------

func TestGap205MistfireAdeptGrantsFlyingOnANoncreatureCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Mistfire Adept", "Creature — Human Monk", gap205Mistfire, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.SettleProwess(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target creature has no flying")
	}
}

// --- Nephalia Smuggler -------------------------------------------------------

func TestGap205SmugglerBlinksAnotherCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	sm := effectstest.PushCatalogPermanent(g, me.ID, "Nephalia Smuggler", "Creature — Human Rogue", gap205Smuggler, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 2}
			}
		}
	})
	effectstest.B06AddMana(me, "C", "C", "C", "U")
	effectstest.B16Activate(t, g, me.ID, sm, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, sm) {
		t.Error("the Smuggler did not tap")
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" {
			n++
			if c.Counters[game.CounterPlusOne] != 0 {
				t.Error("the Bear kept its counters, so it never left")
			}
		}
	}
	if n != 1 {
		t.Errorf("%d Bears on the battlefield, want it back once", n)
	}
	// It cannot target itself.
	effectstest.B06AddMana(me, "C", "C", "C", "U")
}

// --- Disciple of the Ring ----------------------------------------------------

func TestGap205DiscipleTapsUntapsAndPumpsByExilingAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	dis := effectstest.PushCatalogPermanent(g, me.ID, "Disciple of the Ring", "Creature — Human Wizard", gap205Disciple, false)
	foe := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	grave := func(typeLine string) uuid.UUID {
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Spent " + typeLine, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		})
		return id
	}
	// A creature card in the graveyard is not a legal cost.
	bad := grave("Creature — Bear")
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, dis, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{bad}, Modes: []int{2}, Targets: effectstest.B16TargetCard(foe)}); err == nil {
		t.Fatal("a creature card paid the exile cost")
	}
	spell := grave("Instant")
	effectstest.B16Activate(t, g, me.ID, dis, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{spell}, Modes: []int{2}, Targets: effectstest.B16TargetCard(foe)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, foe) {
		t.Error("the target creature did not tap")
	}
	if g.Seats[0].Graveyard.Contains(spell) {
		t.Error("the instant card stayed in the graveyard")
	}
	sorc := grave("Sorcery")
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, dis, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{sorc}, Modes: []int{3}, Targets: effectstest.B16TargetCard(foe)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, foe) {
		t.Error("the target creature did not untap")
	}
	sorc2 := grave("Sorcery")
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, dis, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{sorc2}, Modes: []int{1}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, dis); p != 2 || tt != 2 {
		t.Errorf("the Disciple is %d/%d, want 4/5", p, tt)
	}
}

// --- Jeskai Elder: the loot ----------------------------------------------------

func TestGap205JeskaiElderTriggerIsARegisteredCombatDamageLoot(t *testing.T) {
	spec, ok := effects.Lookup(gap205Elder)
	if !ok || len(spec.Triggered) != 1 {
		t.Fatalf("Jeskai Elder has %d triggers, want its one combat damage trigger", len(spec.Triggered))
	}
}
