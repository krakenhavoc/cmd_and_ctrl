package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b231_cards_test.go — OLD batch 231 (Ixalan, #5532): nine cards (Trove
// of Temptation, the tenth, is already on develop). Each test plays the real
// catalog entry through the engine.

const (
	oldB231ScoundrelsOracle  = "1c63c4ca-5d5a-4ab9-af0d-ed1254b581cc"
	oldB231SkinshifterOracle = "23df9146-9fce-4769-b017-87aa36c79300"
	oldB231CeratopsOracle    = "34f5383b-21fe-45f5-9598-7382ee4c5b12"
	oldB231TotemOracle       = "5432bd43-e92c-464b-9ad2-3515f3042799"
	oldB231BindingOracle     = "5a2d0df9-950e-4c3d-8a08-275f179e33c7"
	oldB231ChaliceOracle     = "52732f93-b951-43c3-bb02-c254996606e5"
	oldB231DryadOracle       = "53929852-6359-4e7a-8f2c-a5954c76ffcc"
	oldB231BladeOracle       = "5fe86083-1009-4de2-b186-f817ad1cc4e4"
	oldB231DesireOracle      = "caa0c20c-594b-4ffd-87de-60bf15d8c3be"
)

func TestOldB231EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB231ScoundrelsOracle:  "Wanted Scoundrels",
		oldB231SkinshifterOracle: "Tilonalli's Skinshifter",
		oldB231CeratopsOracle:    "Goring Ceratops",
		oldB231TotemOracle:       "Sentinel Totem",
		oldB231BindingOracle:     "Ixalan's Binding",
		oldB231ChaliceOracle:     "Hierophant's Chalice",
		oldB231DryadOracle:       "Blossom Dryad",
		oldB231BladeOracle:       "Prying Blade",
		oldB231DesireOracle:      "Depths of Desire",
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

func oldB231Cast(t *testing.T, g *game.Game, p *game.Player, c game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Hand.PushTop(c)
	if err := g.CastSpell(p.ID, c.InstanceID, params); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	return c.InstanceID
}

func oldB231Zone(g *game.Game, id uuid.UUID, kind game.ZoneKind) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == kind
}

func TestOldB231WantedScoundrelsGivesTheTargetOpponentTwoTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	s := effectstest.DftABPush(g, me.ID, "Wanted Scoundrels", "Creature — Human Pirate", oldB231ScoundrelsOracle, 4, 3)
	effectstest.KillCreature(t, g, me.ID, s)
	g.RunStateChecksForTest()
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TmtB1Count(g, opp.ID, "Treasure"); n != 2 {
		t.Errorf("the opponent has %d Treasures, want 2", n)
	}
	if n := effectstest.TmtB1Count(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("I have %d Treasures, want 0", n)
	}
}

func TestOldB231SkinshifterCopiesAnotherNonlegendaryAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	skin := effectstest.DftABPush(g, me.ID, "Tilonalli's Skinshifter", "Creature — Human Shaman", oldB231SkinshifterOracle, 0, 1)
	ogre := effectstest.DftABPush(g, me.ID, "Hill Ogre", "Creature — Ogre", "", 4, 4)
	king := effectstest.DftABPush(g, me.ID, "Legend King", "Legendary Creature — Human", "", 3, 3)
	idle := effectstest.DftABPush(g, me.ID, "Idle Bear", "Creature — Bear", "", 2, 2)

	effectstest.DeclareAttack(t, g, opp.ID, skin, ogre, king)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt")
	}
	if !effectstest.HasID(prompt.PickTargetCards, ogre) {
		t.Error("another nonlegendary attacker is not a legal target")
	}
	for name, id := range map[string]uuid.UUID{"itself": skin, "a legendary attacker": king, "a creature that is not attacking": idle} {
		if effectstest.HasID(prompt.PickTargetCards, id) {
			t.Errorf("%s is offered", name)
		}
	}
	effectstest.PickCard(t, g, me.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.B12Card(t, g, skin)
	if c.Name != "Hill Ogre" {
		t.Errorf("the Skinshifter is %q, want a copy of Hill Ogre", c.Name)
	}
	if p := effectstest.EffectivePower(t, g, skin); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	effectstest.AdvancePastCleanupOf(t, g, 0)
	if c := effectstest.B12Card(t, g, skin); c.Name != "Tilonalli's Skinshifter" {
		t.Errorf("after the turn the Skinshifter is %q", c.Name)
	}
}

func TestOldB231GoringCeratopsGivesTheOthersDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cer := effectstest.DftABPush(g, me.ID, "Goring Ceratops", "Creature — Dinosaur", oldB231CeratopsOracle, 3, 3)
	mine := effectstest.DftABPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.DftABPush(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, cer)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, mine, "double strike") {
		t.Error("my other creature did not gain double strike")
	}
	if effectstest.P1g2HasKeyword(g, theirs, "double strike") {
		t.Error("an opponent's creature gained double strike")
	}
	if !effectstest.P1g2HasKeyword(g, cer, "double strike") {
		t.Error("the Ceratops lost its printed double strike")
	}
}

func TestOldB231SentinelTotemScriesAndExilesEveryGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.PushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	theirs := effectstest.PushGraveyardPermanent(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	totem := effectstest.DftABCast(t, g, me, game.Card{Name: "Sentinel Totem", TypeLine: "Artifact", OracleID: oldB231TotemOracle, ManaCost: "{1}"})
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) == nil {
		t.Fatal("no scry prompt as the totem enters")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, totem, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"mine": mine, "theirs": theirs} {
		if !oldB231Zone(g, id, game.ZoneExile) {
			t.Errorf("the card in %s graveyard was not exiled", name)
		}
	}
	if oldB231Zone(g, totem, game.ZoneBattlefield) {
		t.Error("the totem stayed on the battlefield")
	}
}

func TestOldB231IxalansBindingExilesAndBarsTheNameUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	elk := effectstest.DftABPush(g, opp.ID, "Test Elk", "Creature — Elk", "", 3, 3)
	land := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	binding := oldB231Cast(t, g, me, game.Card{Name: "Ixalan's Binding", TypeLine: "Enchantment", OracleID: oldB231BindingOracle}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(prompt.PickTargetCards, land) {
		t.Error("a land is a legal target")
	}
	effectstest.PickCard(t, g, me.ID, elk)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB231Zone(g, elk, game.ZoneExile) {
		t.Fatal("the Elk was not exiled")
	}
	// An opponent can't cast a spell with the exiled card's name; its owner's other cards are fine.
	same := effectstest.HandCardFull(opp, "Test Elk", "Instant", "", "", nil)
	other := effectstest.HandCardFull(opp, "Other Elk", "Instant", "", "", nil)
	if err := g.CastSpell(opp.ID, same, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("an opponent cast a spell with the exiled card's name")
	}
	if err := g.CastSpell(opp.ID, other, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("an unrelated spell was refused: %v", err)
	}
	// The binding's controller is not barred.
	mineSame := effectstest.HandCardFull(me, "Test Elk", "Instant", "", "", nil)
	if err := g.CastSpell(me.ID, mineSame, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("the controller was barred from the name: %v", err)
	}
	// It leaves, and the card comes back.
	effectstest.KillCreature(t, g, me.ID, binding)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	back := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Test Elk" && c.Controller == opp.ID {
				back = true
			}
		}
	})
	if !back {
		t.Error("the Elk did not return when the Binding left")
	}
}

func TestOldB231HierophantsChaliceDrainsOneAndMakesColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	life, oppLife := me.Life, opp.Life
	chalice := effectstest.DftABCast(t, g, me, game.Card{Name: "Hierophant's Chalice", TypeLine: "Artifact", OracleID: oldB231ChaliceOracle, ManaCost: "{3}"})
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if me.Life != life+1 || opp.Life != oppLife-1 {
		t.Errorf("me %d → %d, opponent %d → %d; want +1 and -1", life, me.Life, oppLife, opp.Life)
	}
	before := len(me.ManaPool.Units())
	if err := g.ActivateManaAbility(me.ID, chalice, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if len(me.ManaPool.Units()) != before+1 {
		t.Errorf("tapping added %d mana, want 1", len(me.ManaPool.Units())-before)
	}
}

func TestOldB231BlossomDryadUntapsATargetLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dryad := effectstest.DftABPush(g, me.ID, "Blossom Dryad", "Creature — Dryad", oldB231DryadOracle, 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == land {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, dryad, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.DftABCardRef(land)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, land); tapped {
		t.Error("the land is still tapped")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, dryad); !tapped {
		t.Error("the Dryad did not tap")
	}
}

func TestOldB231PryingBladePumpsAndMakesATreasureOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	blade := effectstest.DftABPush(g, me.ID, "Prying Blade", "Artifact — Equipment", oldB231BladeOracle, 0, 0)
	bear := effectstest.DftABPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.EquipTo(t, g, me.ID, blade, bear)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("equipped power %d, want 3", p)
	}
	if tough := effectstest.B12Card(t, g, bear).Toughness; tough != 2 {
		t.Errorf("equipped toughness %d, want 2", tough)
	}
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d → %d, want -3", life, opp.Life)
	}
	if n := effectstest.TmtB1Count(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
}

func TestOldB231DepthsOfDesireBouncesAndMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	elk := effectstest.DftABPush(g, opp.ID, "Test Elk", "Creature — Elk", "", 3, 3)
	oldB231Cast(t, g, me, game.Card{Name: "Depths of Desire", TypeLine: "Instant", OracleID: oldB231DesireOracle},
		game.CastSpellParams{Targets: effectstest.DftABCardRef(elk)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(opp, elk) {
		t.Error("the Elk is not back in its owner's hand")
	}
	if n := effectstest.TmtB1Count(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures for me, want 1", n)
	}
}
