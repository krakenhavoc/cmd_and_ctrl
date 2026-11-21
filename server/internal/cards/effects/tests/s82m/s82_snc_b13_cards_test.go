package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b13_cards_test.go — S82 Streets of New Capenna, batch 13 (#5491):
// Body Dropper; Forge Boss; Pyre-Sledge Arsonist; Riveteers Ascendancy; Crew
// Captain; Freelance Muscle; Mr. Orfeo, the Boulder; Cutthroat Contender;
// Dusk Mangler; Bellowing Mauler; Incriminate; Industrial Advancement.

const (
	sncB13BodyDropper = "c0c465ea-4242-4737-9c3e-69222bdb8f1b"
	sncB13ForgeBoss   = "c234d2ab-21d5-41a9-9ab8-47a8bd53a686"
	sncB13Arsonist    = "7f42a05c-9e6d-4cbc-92f1-4433cd22f149"
	sncB13Ascendancy  = "71f3168d-9eeb-4570-8437-a3ffc39667e9"
	sncB13CrewCaptain = "0cca60c3-f0be-494d-8c2c-9b135ef83165"
	sncB13Muscle      = "8a749efc-e333-448e-8652-2dc4e6633e6b"
	sncB13Orfeo       = "58447f16-1df9-4e9a-bb22-b9aad042055a"
	sncB13Contender   = "43192553-0db9-48ec-ae88-a2605a6608d8"
	sncB13Mangler     = "74de3356-9779-40f9-9881-fe795a1d5bbd"
	sncB13Mauler      = "fa0a4765-5799-4464-88b6-6d1b15889eb5"
	sncB13Incriminate = "d8b59cc3-766f-4350-971f-edf721fcc5be"
	sncB13Advancement = "7a0aa9ea-610e-4910-b66d-fdac69e88adf"
)

func TestSncB13CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		sncB13BodyDropper: "Body Dropper",
		sncB13ForgeBoss:   "Forge Boss",
		sncB13Arsonist:    "Pyre-Sledge Arsonist",
		sncB13Ascendancy:  "Riveteers Ascendancy",
		sncB13CrewCaptain: "Crew Captain",
		sncB13Muscle:      "Freelance Muscle",
		sncB13Orfeo:       "Mr. Orfeo, the Boulder",
		sncB13Contender:   "Cutthroat Contender",
		sncB13Mangler:     "Dusk Mangler",
		sncB13Mauler:      "Bellowing Mauler",
		sncB13Incriminate: "Incriminate",
		sncB13Advancement: "Industrial Advancement",
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

// sncB13Sacrifice sacrifices a permanent as an effect would.
func sncB13Sacrifice(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(id) })
}

func sncB13Activate(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID, idx int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	if mana != "" {
		effectstest.FillPoolColored(me, "C", 1)
	}
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, src, idx, params)
}

// --- Body Dropper ---------------------------------------------------------

func TestSncB13BodyDropperGrowsOnOtherCreatureSacrificesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dropper := effectstest.B12Push(g, me.ID, "Body Dropper", "Creature — Devil Warrior", sncB13BodyDropper, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 3, 3)
	land := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")

	sncB13Sacrifice(g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Card(t, g, dropper).Counters[game.CounterPlusOne]; n != 1 {
		t.Fatalf("after sacrificing another creature: %d counters, want 1", n)
	}
	sncB13Sacrifice(g, land)
	sncB13Sacrifice(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Card(t, g, dropper).Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("a noncreature or an opponent's sacrifice grew it: %d counters", n)
	}
}

func TestSncB13BodyDropperSacrificingItselfDoesNotTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dropper := effectstest.B12Push(g, me.ID, "Body Dropper", "Creature — Devil Warrior", sncB13BodyDropper, 2, 2)
	sncB13Sacrifice(g, dropper)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(g.PendingTriggers); n != 0 {
		t.Errorf("%d triggers pending after the Dropper itself was sacrificed", n)
	}
}

func TestSncB13BodyDropperActivationGivesMenaceAndGrowsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dropper := effectstest.B12Push(g, me.ID, "Body Dropper", "Creature — Devil Warrior", sncB13BodyDropper, 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := sncB13Activate(t, g, me, dropper, 0, "", game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	effectstest.FillPoolColored(me, "B", 1)
	effectstest.FillPoolColored(me, "R", 1)
	if err := sncB13Activate(t, g, me, dropper, 0, "", game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{dropper}}); err == nil {
		t.Fatal("sacrificed itself for 'another creature'")
	}
	if err := sncB13Activate(t, g, me, dropper, 0, "", game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	d := effectstest.B12Card(t, g, dropper)
	if !effectstest.HasEffectiveKeyword(t, g, dropper, "menace") {
		t.Error("no menace after the ability resolved")
	}
	if n := d.Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("the creature sacrificed to pay did not grow it: %d counters", n)
	}
}

// --- Forge Boss -----------------------------------------------------------

func TestSncB13ForgeBossDamagesOncePerTurnPerBatch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	boss := effectstest.B12Push(g, me.ID, "Forge Boss", "Creature — Human Warrior", sncB13ForgeBoss, 3, 4)
	a := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Cat", "Creature — Cat", 2, 2)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		lives[p.ID] = p.Life
	}
	sncB13Sacrifice(g, a)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if p.Life != lives[p.ID]-2 {
			t.Errorf("seat life %d -> %d, want -2", lives[p.ID], p.Life)
		}
	}
	sncB13Sacrifice(g, b)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != lives[opp.ID]-2 {
		t.Errorf("a second sacrifice the same turn hit again: %d -> %d", lives[opp.ID], opp.Life)
	}
	_ = c
	_ = boss
}

func TestSncB13ForgeBossIgnoresItselfAndNoncreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	boss := effectstest.B12Push(g, me.ID, "Forge Boss", "Creature — Human Warrior", sncB13ForgeBoss, 3, 4)
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	sncB13Sacrifice(g, relic)
	sncB13Sacrifice(g, boss)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Fatalf("a noncreature or the Boss itself triggered it: %d -> %d", life, opp.Life)
	}
	_ = other
}

// --- Pyre-Sledge Arsonist -------------------------------------------------

func TestSncB13ArsonistDealsOneDamagePerPermanentSacrificedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	arsonist := effectstest.B12Push(g, me.ID, "Pyre-Sledge Arsonist", "Creature — Lizard Shaman", sncB13Arsonist, 2, 2)
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 3, 3)
	sncB13Sacrifice(g, relic)
	sncB13Sacrifice(g, bear)
	sncB13Sacrifice(g, theirs) // an opponent's sacrifice is not yours
	life := opp.Life
	if err := sncB13Activate(t, g, me, arsonist, 0, "{1}", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d -> %d, want -2 (two permanents sacrificed)", life, opp.Life)
	}
}

func TestSncB13ArsonistWithNothingSacrificedDealsNothingAndNeedsTheMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	arsonist := effectstest.B12Push(g, me.ID, "Pyre-Sledge Arsonist", "Creature — Lizard Shaman", sncB13Arsonist, 2, 2)
	tgt := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := sncB13Activate(t, g, me, arsonist, 0, "", game.ActivateAbilityParams{Targets: tgt}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	life := opp.Life
	if err := sncB13Activate(t, g, me, arsonist, 0, "{1}", game.ActivateAbilityParams{Targets: tgt}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("life %d -> %d with X = 0", life, opp.Life)
	}
}

// --- Riveteers Ascendancy -------------------------------------------------

// sncB13InGraveyard puts a creature card with the given mana cost in a
// player's graveyard.
func sncB13InGraveyard(p *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Bear", ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}

func sncB13AnswerTrigger(t *testing.T, g *game.Game, p *game.Player, yes bool) {
	t.Helper()
	effectstest.AnswerLatestTriggerPrompt(t, g, p.ID, yes)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSncB13AscendancyReturnsALesserCreatureTappedOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Riveteers Ascendancy", "Enchantment", sncB13Ascendancy, 0, 0)
	fodder := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 4, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == fodder {
				g.Battlefield.Cards[i].ManaCost = "{3}{B}"
			}
		}
	})
	cheap := sncB13InGraveyard(me, "Cheap Bear", "{1}{G}")
	pricey := sncB13InGraveyard(me, "Pricey Bear", "{3}{G}")
	sncB13Sacrifice(g, fodder)
	// The only legal target has a lesser mana value (2 < 4); the 4 is not.
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, cheap)
	}
	sncB13AnswerTrigger(t, g, me, true)
	if !g.Battlefield.Contains(cheap) {
		t.Fatal("the lesser creature did not return")
	}
	if g.Battlefield.Contains(pricey) {
		t.Error("an equal-mana-value creature returned")
	}
	if c := effectstest.B12Card(t, g, cheap); !c.Tapped {
		t.Error("the creature did not enter tapped")
	}
}
