package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b324_cards_test.go — Dragons of Tarkir, batch 324 (S82 older sets,
// #5551): the formidable creatures. Every test plays the real catalog entry
// through the engine.

const (
	oldB324Outrider   = "2a98f634-d739-47a7-98cb-b7f935cd65ef"
	oldB324Bear       = "2f6f49a4-f2ea-4147-a6d3-69cf875c302e"
	oldB324Pummeler   = "4b2831f5-70e3-4f4f-8393-4cd141c9da34"
	oldB324Whisperer  = "54884040-b426-43da-9bb1-016621e1446d"
	oldB324Arynx      = "57303267-7e69-4bd3-8e1f-a528d41121a1"
	oldB324Crater     = "8ea607cd-c14b-4bb1-aba1-c5cdc397a825"
	oldB324Shaman     = "ca405696-8c30-4c01-bf90-11cdf8837942"
	oldB324Watcher    = "f696b727-083e-4825-80d9-4bac35289958"
	oldB324Circle     = "f720f80a-74bc-4681-ad4f-d33152ff988b"
	oldB324ElkHerd    = "fc1020b7-086e-41d6-a9ae-1c6c72cf820c"
	oldB324Beastbreak = "fdd4ef61-607a-4090-9380-dabbbb448519"
	oldB324Creature   = "Creature — Bear"
	oldB324Bolt       = "4457ed35-7c10-48c8-9776-456485fdf070"
)

func TestOldB324EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB324Outrider, oldB324Bear, oldB324Pummeler, oldB324Whisperer, oldB324Arynx, oldB324Crater,
		oldB324Shaman, oldB324Watcher, oldB324Circle, oldB324ElkHerd, oldB324Beastbreak,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// oldB324Activate activates ability `idx` of a permanent, strictly.
func oldB324Activate(g *game.Game, me *game.Player, id uuid.UUID, idx int, params game.ActivateAbilityParams) error {
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, id, idx, params)
}

// oldB324Float adds mana of one letter to the pool.
func oldB324Float(g *game.Game, me *game.Player, letter string, n int) {
	effectstest.FloatForTest(g, me, strings.Repeat(letter, n))
}

func oldB324Power(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	p, _ := effectstest.EffectivePTOf(t, g, id)
	return p
}

func TestOldB324OutriderGainsFirstStrikeOnlyWithEightPower(t *testing.T) {
	for _, formidable := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		rider := effectstest.B12Push(g, me.ID, "Sabertooth Outrider", "Creature — Human Warrior", oldB324Outrider, 4, 2)
		if formidable {
			effectstest.B12Creature(g, me.ID, "Big Bear", oldB324Creature, 4, 4)
		} else {
			effectstest.B12Creature(g, me.ID, "Small Bear", oldB324Creature, 3, 3)
		}
		effectstest.AttackWithTrigger(t, g, rider, opp.ID, uuid.Nil)
		got := effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, rider), "first strike")
		if got != formidable {
			t.Errorf("formidable=%v: first strike = %v", formidable, got)
		}
	}
}

func TestOldB324ElkHerdGrantsTrampleToEveryCreatureYouControl(t *testing.T) {
	for _, formidable := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		herd := effectstest.B12Push(g, me.ID, "Stampeding Elk Herd", "Creature — Elk", oldB324ElkHerd, 5, 5)
		idle := effectstest.B12Creature(g, me.ID, "Idle Bear", oldB324Creature, 1, 1)
		if formidable {
			effectstest.B12Creature(g, me.ID, "Extra Bear", oldB324Creature, 2, 2)
		}
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB324Creature, 2, 2)
		effectstest.AttackWithTrigger(t, g, herd, opp.ID, uuid.Nil)
		for name, id := range map[string]uuid.UUID{"herd": herd, "idle": idle} {
			if got := effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, id), "trample"); got != formidable {
				t.Errorf("formidable=%v: %s trample = %v", formidable, name, got)
			}
		}
		if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, theirs), "trample") {
			t.Error("an opponent's creature gained trample")
		}
	}
}

func TestOldB324BearRegeneratesOnlyWithEightPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Push(g, me.ID, "Dragon-Scarred Bear", "Creature — Bear", oldB324Bear, 3, 2)
	oldB324Float(g, me, "G", 2)
	if err := oldB324Activate(g, me, bear, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with only 3 total power")
	}
	big := effectstest.B12Creature(g, me.ID, "Big Bear", oldB324Creature, 5, 5)
	if err := oldB324Activate(g, me, bear, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with 8 power: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	g.RunStateChecksForTest()
	if effectstest.ZoneOf(g, bear) != game.ZoneBattlefield {
		t.Error("the Bear did not regenerate")
	}
	_ = big
}

func TestOldB324PummelerGivesYourCreaturesMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	pummeler := effectstest.B12Push(g, me.ID, "Atarka Pummeler", "Creature — Ogre Warrior", oldB324Pummeler, 4, 5)
	ally := effectstest.B12Creature(g, me.ID, "Ally Bear", oldB324Creature, 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB324Creature, 2, 2)
	oldB324Float(g, me, "R", 5)
	if err := oldB324Activate(g, me, pummeler, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"pummeler": pummeler, "ally": ally} {
		if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, id), "menace") {
			t.Errorf("%s lacks menace", name)
		}
	}
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, theirs), "menace") {
		t.Error("an opponent's creature gained menace")
	}
}

func TestOldB324WhispererPumpsFliesAndMakesADragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	whisperer := effectstest.B12Push(g, me.ID, "Dragon Whisperer", "Creature — Human Shaman", oldB324Whisperer, 2, 2)
	oldB324Float(g, me, "R", 12)
	if err := oldB324Activate(g, me, whisperer, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("flying: %v", err)
	}
	if err := oldB324Activate(g, me, whisperer, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, whisperer), "flying") || oldB324Power(t, g, whisperer) != 3 {
		t.Errorf("the Whisperer is %d power flying=%v, want 3 with flying", oldB324Power(t, g, whisperer),
			effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, whisperer), "flying"))
	}
	if err := oldB324Activate(g, me, whisperer, 2, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the Dragon ability worked with 3 total power")
	}
	effectstest.B12Creature(g, me.ID, "Big Bear", oldB324Creature, 5, 5)
	if err := oldB324Activate(g, me, whisperer, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Dragon ability with 8 power: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.IsToken() && c.Name == "Dragon" {
			n++
			if p, tough := effectstest.EffectivePTOf(t, g, c.InstanceID); p != 4 || tough != 4 || !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, c.InstanceID), "flying") {
				t.Errorf("the Dragon is %d/%d, want a 4/4 flier", p, tough)
			}
		}
	}
	if n != 1 {
		t.Errorf("%d Dragon tokens, want 1", n)
	}
}

func TestOldB324ArynxNeedsEightPowerToForceABlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	arynx := effectstest.B12Push(g, me.ID, "Lurking Arynx", "Creature — Cat Beast", oldB324Arynx, 3, 5)
	foe := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB324Creature, 2, 2)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: foe}}
	oldB324Float(g, me, "G", 3)
	if err := oldB324Activate(g, me, arynx, 0, game.ActivateAbilityParams{Targets: target}); err == nil {
		t.Fatal("activated with 3 total power")
	}
	effectstest.B12Creature(g, me.ID, "Big Bear", oldB324Creature, 5, 5)
	if err := oldB324Activate(g, me, arynx, 0, game.ActivateAbilityParams{Targets: target}); err != nil {
		t.Fatalf("activate with 8 power: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB324CraterElementalSacrificesToKillAndSetsBasePower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	crater := effectstest.B12Push(g, me.ID, "Crater Elemental", "Creature — Elemental", oldB324Crater, 0, 6)
	oldB324Float(g, me, "R", 4)
	if err := oldB324Activate(g, me, crater, 1, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the base power ability worked with no power on the table")
	}
	effectstest.B12Creature(g, me.ID, "Big Bear", oldB324Creature, 8, 8)
	if err := oldB324Activate(g, me, crater, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("base power with 8 power: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := oldB324Power(t, g, crater); p != 8 {
		t.Errorf("the Elemental has %d power, want 8", p)
	}
	foe := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB324Creature, 2, 4)
	if err := oldB324Activate(g, me, crater, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: foe}}}); err != nil {
		t.Fatalf("sacrifice ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, crater) == game.ZoneBattlefield {
		t.Error("the Elemental was not sacrificed")
	}
	if effectstest.ZoneOf(g, foe) == game.ZoneBattlefield {
		t.Error("the target survived 4 damage")
	}
}

func TestOldB324ShamanMakesTwoCreatureOnlyMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	shaman := effectstest.PushCatalogPermanent(g, me.ID, "Shaman of Forgotten Ways", "Creature — Human Shaman", oldB324Shaman, false)
	if err := g.ActivateManaAbility(me.ID, shaman, 0, game.ManaAbilityParams{Colors: []string{"G", "R"}}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if me.ManaPool.Total() != 2 {
		t.Fatalf("pool %v, want two mana", me.ManaPool.Units())
	}
	spell := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}, OracleID: oldB324Bolt, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}}); err == nil {
		t.Error("the restricted mana paid for a noncreature spell")
	}
}

func TestOldB324ShamanSetsEveryLifeTotalToItsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	shaman := effectstest.B12Push(g, me.ID, "Shaman of Forgotten Ways", "Creature — Human Shaman", oldB324Shaman, 2, 3)
	effectstest.B12Creature(g, me.ID, "Big Bear", oldB324Creature, 6, 6)
	effectstest.B12Creature(g, opp.ID, "Their Bear", oldB324Creature, 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Other Bear", oldB324Creature, 1, 1)
	oldB324Float(g, me, "G", 11)
	if err := oldB324Activate(g, me, shaman, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 2 || opp.Life != 2 {
		t.Errorf("life totals %d and %d, want 2 and 2", me.Life, opp.Life)
	}
	if g.Seats[2].Life != 0 || g.Seats[3].Life != 0 {
		t.Errorf("creatureless players are at %d and %d, want 0", g.Seats[2].Life, g.Seats[3].Life)
	}
}

func TestOldB324GladeWatcherCanAttackOnlyAfterTheFormidableWaiver(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	watcher := effectstest.B12Push(g, me.ID, "Glade Watcher", "Creature — Elemental", oldB324Watcher, 3, 3)
	effectstest.B12Creature(g, me.ID, "Big Bear", oldB324Creature, 5, 5)
	oldB324Float(g, me, "G", 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(watcher, opp.ID); err == nil {
		t.Fatal("it attacked with defender")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := effectstest.SeatsForCombat(g2)
	effectstest.AdvanceToMain(t, g2)
	watcher2 := effectstest.B12Push(g2, me2.ID, "Glade Watcher", "Creature — Elemental", oldB324Watcher, 3, 3)
	big := effectstest.B12Creature(g2, me2.ID, "Big Bear", oldB324Creature, 5, 5)
	oldB324Float(g2, me2, "G", 1)
	if err := oldB324Activate(g2, me2, watcher2, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("waiver with 8 power: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(watcher2, opp2.ID); err != nil {
		t.Errorf("it could not attack after the waiver: %v", err)
	}
	_ = big
}

func TestOldB324GladeWatcherWaiverNeedsEightPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	watcher := effectstest.B12Push(g, me.ID, "Glade Watcher", "Creature — Elemental", oldB324Watcher, 3, 3)
	oldB324Float(g, me, "G", 1)
	if err := oldB324Activate(g, me, watcher, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the waiver worked with 3 total power")
	}
}

func TestOldB324CircleOfEldersTapsForThreeOnlyWithEightPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	circle := effectstest.PushCatalogPermanent(g, me.ID, "Circle of Elders", "Creature — Human Shaman", oldB324Circle, false)
	if err := g.ActivateManaAbility(me.ID, circle, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("it tapped for mana with 1 total power")
	}
	big := effectstest.B12Creature(g, me.ID, "Big Bear", oldB324Creature, 7, 7)
	if err := g.ActivateManaAbility(me.ID, circle, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap with 8 power: %v", err)
	}
	if me.ManaPool.CountColor("C") != 3 {
		t.Errorf("pool %v, want CCC", me.ManaPool.Units())
	}
	_ = big
}

func TestOldB324BeastbreakerGetsPlusFourOnlyWithEightPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	breaker := effectstest.B12Push(g, me.ID, "Atarka Beastbreaker", "Creature — Human Warrior", oldB324Beastbreak, 2, 2)
	oldB324Float(g, me, "G", 10)
	if err := oldB324Activate(g, me, breaker, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("it pumped with 2 total power")
	}
	effectstest.B12Creature(g, me.ID, "Big Bear", oldB324Creature, 6, 6)
	if err := oldB324Activate(g, me, breaker, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("pump with 8 power: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, breaker); p != 6 || tough != 6 {
		t.Errorf("the Beastbreaker is %d/%d, want 6/6", p, tough)
	}
}
