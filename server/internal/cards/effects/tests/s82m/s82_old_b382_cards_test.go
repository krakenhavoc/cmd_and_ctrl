package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b382_cards_test.go — OLD batch 382 (Commander 2013, #5560): seven
// removal cards. Each test plays the real catalog entry through the engine.

const (
	oldB382Ashes     = "3e229329-65e4-4240-959a-b97b26908c0e"
	oldB382Druid     = "560e1e81-6675-4d08-8e82-cdf1abd4b3d0"
	oldB382Shambler  = "7aba1039-f108-423e-8aa4-d9d8b4da2414"
	oldB382Charm     = "a1a49639-bcc4-4522-8862-c9fb13d40880"
	oldB382Authority = "c12b8e31-5702-4d3c-80c9-56c87e50679d"
	oldB382Deepfire  = "c8119ebe-aedd-4bdb-8f7f-368674a049fd"
	oldB382Vent      = "f5bac25d-72e9-4655-8a04-3646fc10be27"
)

func TestOldB382EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB382Ashes: "From the Ashes", oldB382Druid: "Quagmire Druid", oldB382Shambler: "Mold Shambler",
		oldB382Charm: "Selesnya Charm", oldB382Authority: "Act of Authority", oldB382Deepfire: "Deepfire Elemental",
		oldB382Vent: "Fissure Vent",
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

func oldB382Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) error {
	t.Helper()
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	return g.CastSpell(me.ID, id, params)
}

// oldB382TakeTheMay answers Act of Authority's two prompts, the target pick
// and the "you may", whichever order they come in.
func oldB382TakeTheMay(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	for i := 0; i < 6; i++ {
		switch {
		case effectstest.LatestPickTarget(g, me.ID) != nil:
			effectstest.PickCard(t, g, me.ID, target)
		case effectstest.ConfirmForTriggerPrompt(g, me.ID) != nil:
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		case effectstest.LatestConfirmFor(g, me.ID) != nil:
			effectstest.AnswerMayChoice(t, g, me.ID, true)
		default:
			effectstest.PassPriorityAroundTable(t, g)
		}
	}
}

// --- Fissure Vent -----------------------------------------------------------

func TestOldB382FissureVentDestroysBothTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	shock := effectstest.B12Permanent(g, opp.ID, "Shock Land", "Land — Mountain Forest")
	basic := effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	if err := oldB382Cast(t, g, "Fissure Vent", "Sorcery", oldB382Vent, game.CastSpellParams{Strict: true, Modes: []int{0, 1}, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: relic, Mode: 0}, {Kind: game.TargetCard, ID: basic, Mode: 1}}}); err == nil {
		t.Fatal("a basic land was a legal target of the nonbasic bullet")
	}
	if err := oldB382Cast(t, g, "Fissure Vent", "Sorcery", oldB382Vent, game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: relic, Mode: 0}, {Kind: game.TargetCard, ID: shock, Mode: 1}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) || g.Battlefield.Contains(shock) {
		t.Error("a chosen bullet's target survived")
	}
	if !g.Battlefield.Contains(basic) {
		t.Error("the basic land was destroyed")
	}
}

func TestOldB382FissureVentCanDoOnlyOneBullet(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	shock := effectstest.B12Permanent(g, opp.ID, "Shock Land", "Land — Mountain Forest")
	if err := oldB382Cast(t, g, "Fissure Vent", "Sorcery", oldB382Vent, game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: shock, Mode: 0}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(shock) {
		t.Error("the nonbasic land survived")
	}
	if !g.Battlefield.Contains(relic) {
		t.Error("the unchosen bullet destroyed the artifact")
	}
}

// --- Selesnya Charm ---------------------------------------------------------

func TestOldB382SelesnyaCharmModes(t *testing.T) {
	t.Run("pump and trample", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		if err := oldB382Cast(t, g, "Selesnya Charm", "Instant", oldB382Charm, game.CastSpellParams{Modes: []int{0},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear, Mode: 0}}}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if p := effectstest.EffectivePower(t, g, bear); p != 4 {
			t.Errorf("power %d, want 4", p)
		}
		if !effectstest.P1g2HasKeyword(g, bear, "trample") {
			t.Error("the Bear lacks trample")
		}
	})
	t.Run("exile a big creature", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 4, 4)
		big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 5, 5)
		if err := oldB382Cast(t, g, "Selesnya Charm", "Instant", oldB382Charm, game.CastSpellParams{Strict: true, Modes: []int{1},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small, Mode: 0}}}); err == nil {
			t.Fatal("a 4-power creature was a legal target")
		}
		if err := oldB382Cast(t, g, "Selesnya Charm", "Instant", oldB382Charm, game.CastSpellParams{Modes: []int{1},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big, Mode: 0}}}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if z := effectstest.B12ZoneOf(g, big); z != game.ZoneExile {
			t.Errorf("the Ogre is in %s, want exile", z)
		}
	})
	t.Run("a Knight token", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		if err := oldB382Cast(t, g, "Selesnya Charm", "Instant", oldB382Charm, game.CastSpellParams{Modes: []int{2}}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.CountOnBattlefield(g, "Knight", me.ID); n != 1 {
			t.Errorf("%d Knights, want 1", n)
		}
	})
}

// --- Mold Shambler ----------------------------------------------------------

func TestOldB382MoldShamblerDestroysOnlyWhenKicked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Mold Shambler", "Creature — Fungus Beast", oldB382Shambler, nil, nil, nil); err != nil {
		t.Fatalf("unkicked cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(relic) {
		t.Fatal("an unkicked Mold Shambler destroyed the artifact")
	}
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Mold Shambler", "Creature — Fungus Beast", oldB382Shambler, nil, []int{0}, nil); err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("a creature was accepted as the target")
	}
	effectstest.PickCard(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the kicked Mold Shambler did not destroy the artifact")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature was destroyed")
	}
}

// --- Quagmire Druid ---------------------------------------------------------

func TestOldB382QuagmireDruidSacrificesACreatureToDestroyAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	druid := effectstest.B12Push(g, me.ID, "Quagmire Druid", "Creature — Zombie Druid", oldB382Druid, 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	ench := effectstest.B12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	effectstest.AddPool(t, g, me.ID, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, druid, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(fodder)}); err == nil {
		t.Fatal("a creature was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, druid, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(ench), SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if !g.Battlefield.Contains(druid) {
		t.Error("the Druid was sacrificed")
	}
}

// --- Deepfire Elemental -----------------------------------------------------

func TestOldB382DeepfireElementalDestroysByManaValueX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	elem := effectstest.B12Push(g, me.ID, "Deepfire Elemental", "Creature — Elemental", oldB382Deepfire, 4, 4)
	two := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: two, Name: "Two Drop", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	three := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: three, Name: "Three Drop", TypeLine: "Artifact", ManaCost: "{3}", Owner: opp.ID, Controller: opp.ID})
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, elem, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, Targets: effectstest.CardRefs(three)}); err == nil {
		t.Fatal("a mana value 3 artifact was a legal target for X = 2")
	}
	if err := g.ActivateCatalogAbility(me.ID, elem, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, Targets: effectstest.CardRefs(two)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(two) {
		t.Error("the mana value 2 creature survived")
	}
	if !g.Battlefield.Contains(three) {
		t.Error("the mana value 3 artifact was destroyed")
	}
}

// --- From the Ashes ---------------------------------------------------------

func TestOldB382FromTheAshesDestroysNonbasicLandsAndFetchesBasics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Permanent(g, me.ID, "Shock A", "Land — Mountain Forest")
	b := effectstest.B12Permanent(g, opp.ID, "Shock B", "Land — Mountain Forest")
	c := effectstest.B12Permanent(g, opp.ID, "Shock C", "Land — Plains Island")
	basic := effectstest.B12Permanent(g, opp.ID, "Plains", "Basic Land — Plains")
	var mine, theirs []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
		mine = append(mine, id)
		id = uuid.New()
		opp.Library.PushTop(game.Card{InstanceID: id, Name: "Island", TypeLine: "Basic Land — Island", Owner: opp.ID, Controller: opp.ID})
		theirs = append(theirs, id)
	}
	if err := oldB382Cast(t, g, "From the Ashes", "Sorcery", oldB382Ashes, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if g.Battlefield.Contains(id) {
			t.Errorf("nonbasic land %s survived", effectstest.B12Card(t, g, id).Name)
		}
	}
	if !g.Battlefield.Contains(basic) {
		t.Error("the basic land was destroyed")
	}
	// The caster lost one land and may fetch one; the opponent lost two and
	// may fetch up to two.
	effectstest.AnswerSearchByID(t, g, me.ID, mine[1])
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, opp.ID, theirs[0], theirs[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine[1]) {
		t.Error("the caster's basic land did not enter")
	}
	if !g.Battlefield.Contains(theirs[0]) || !g.Battlefield.Contains(theirs[1]) {
		t.Error("the opponent's two basic lands did not both enter")
	}
}

// --- Act of Authority -------------------------------------------------------

func TestOldB382ActOfAuthorityEntersAndMayExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	if err := oldB382Cast(t, g, "Act of Authority", "Enchantment", oldB382Authority, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	oldB382TakeTheMay(t, g, me, relic)
	if z := effectstest.B12ZoneOf(g, relic); z != game.ZoneExile {
		t.Errorf("the artifact is in %s, want exile", z)
	}
}

func TestOldB382ActOfAuthorityUpkeepExilesAndHandsItOver(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	act := effectstest.PushCatalogPermanent(g, me.ID, "Act of Authority", "Enchantment", oldB382Authority, false)
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	effectstest.WalkToUpkeepOfSeat0(t, g)
	oldB382TakeTheMay(t, g, me, relic)
	if z := effectstest.B12ZoneOf(g, relic); z != game.ZoneExile {
		t.Fatalf("the artifact is in %s, want exile", z)
	}
	if c := effectstest.B12Card(t, g, act); c.Controller != opp.ID {
		t.Errorf("Act of Authority is controlled by %s, want the artifact's controller", c.Controller)
	}
}
