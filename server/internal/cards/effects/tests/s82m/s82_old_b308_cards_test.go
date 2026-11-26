package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b308_cards_test.go — Battle for Zendikar Devoid cards, batch 308
// (#5547). Each test plays the real catalog entry and asserts what its text
// says.

const (
	oldB308Crumble   = "c13da0f4-917f-4e07-b484-21cb4bec4da6"
	oldB308Nettle    = "cc4e8b8e-f615-4842-84c5-d6154986b666"
	oldB308Sludge    = "dd05294b-2c5b-48c2-a9f6-8a528baa3dff"
	oldB308Herald    = "e1890cbc-b15b-44bc-b1c0-ca3d9fd6028a"
	oldB308Mist      = "e4e9b48b-92f9-46ac-b55a-573bc328b11a"
	oldB308Assault   = "e773258a-e59b-4e5f-b4e1-4f95cb11a39f"
	oldB308Tyrant    = "f3b70d61-e33e-4243-aac9-e7bd9ba40acb"
	oldB308Skitter   = "ff3f5b1e-9d53-4258-84ee-cfecf685dd05"
	oldB308GreenBear = "Creature — Bear"
)

func TestOldB308CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB308Crumble: "Crumble to Dust", oldB308Nettle: "Nettle Drone", oldB308Sludge: "Sludge Crawler",
		oldB308Herald: "Herald of Kozilek", oldB308Mist: "Mist Intruder", oldB308Assault: "Processor Assault",
		oldB308Tyrant: "Barrage Tyrant", oldB308Skitter: "Skitterskin",
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

func oldB308Colored(g *game.Game, owner uuid.UUID, name string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: oldB308GreenBear, ManaCost: "{G}", Colors: []string{"G"},
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func TestOldB308CrumbleToDustExilesANonbasicLandAndLetsYouTakeAnyNumberOfCopies(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Command Tower", "Land")
	x.Permanent("tower", "Command Tower", "Land")
	x.Permanent("forest", "Forest", "Basic Land — Forest")
	if err := effectstest.CastCatalogSpellErr(t, x.G, "Crumble to Dust", "Sorcery", oldB308Crumble, x.Target("forest")); err == nil {
		t.Fatal("a basic land was a legal target")
	}
	effectstest.CastCatalogSpell(t, x.G, "Crumble to Dust", "Sorcery", oldB308Crumble, x.Target("tower"))
	effectstest.PassPriorityAroundTable(t, x.G)
	x.Exiled(t, "tower")
	// "any number": the graveyard's copies are offered too, and finding none is legal
	effectstest.WantKeys(t, x.Candidates(t), "gy", "gy 2", "hand", "lib")
	if c := effectstest.SearchChoiceFor(x.G, x.Me.ID); x.G.SearchFloorLocked(c) != 0 {
		t.Error("any number: finding nothing must be legal")
	}
	effectstest.AnswerSearchByID(t, x.G, x.Me.ID, x.Cards["gy 2"], x.Cards["hand"])
	x.Exiled(t, "gy 2", "hand")
	x.NotExiled(t, "gy", "lib", "forest", "gy forest")
}

func TestOldB308NettleDronePingsAndUntapsOnAColorlessSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drone := effectstest.B12Push(g, me.ID, "Nettle Drone", "Creature — Eldrazi Drone", oldB308Nettle, 3, 1)
	effectstest.AdvanceToMain(t, g)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats[1:] {
		lives[s.ID] = s.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, drone, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats[1:] {
		if s.Life != lives[s.ID]-1 {
			t.Errorf("an opponent's life %d -> %d, want 1 damage", lives[s.ID], s.Life)
		}
	}
	if !effectstest.IsTapped(g, drone) {
		t.Fatal("the drone did not tap for its cost")
	}
	effectstest.CastCatalogSpell(t, g, "Grey Spell", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, drone) {
		t.Error("a colorless spell did not untap the drone")
	}
}

func TestOldB308NettleDroneIgnoresAColoredSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drone := effectstest.B12Push(g, me.ID, "Nettle Drone", "Creature — Eldrazi Drone", oldB308Nettle, 3, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, drone, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	spell := effectstest.SocB17Hand(me, "Green Sorcery", "Sorcery", "{G}", "")
	effectstest.FloatForTest(g, me, "G")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, drone) {
		t.Error("a green spell untapped the drone")
	}
}

func TestOldB308SludgeCrawlerPumpsAndIngests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	crawler := effectstest.B12Push(g, me.ID, "Sludge Crawler", "Creature — Eldrazi Drone", oldB308Sludge, 1, 1)
	effectstest.SeedLibrary(opp, "a", "b", "c")
	effectstest.FloatForTest(g, me, "CCCC")
	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, crawler, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tough := effectstest.EffectivePower(t, g, crawler), effectstest.EffectiveToughness(t, g, crawler); p != 3 || tough != 3 {
		t.Errorf("after two pumps: %d/%d, want 3/3", p, tough)
	}
	lib := opp.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, crawler)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != lib-1 {
		t.Errorf("library %d -> %d, want one card exiled", lib, opp.Library.Size())
	}
}

func TestOldB308MistIntruderFliesAndIngests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	mist := effectstest.B12Push(g, me.ID, "Mist Intruder", "Creature — Eldrazi Drone", oldB308Mist, 1, 2)
	if !effectstest.HasEffectiveKeyword(t, g, mist, "flying") {
		t.Error("no flying")
	}
	effectstest.SeedLibrary(opp, "a", "b", "c")
	lib := opp.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, mist)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != lib-1 {
		t.Errorf("library %d -> %d, want one card exiled", lib, opp.Library.Size())
	}
}

func TestOldB308HeraldOfKozilekDiscountsOnlyYourColorlessSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Herald of Kozilek", "Creature — Eldrazi Drone", oldB308Herald, 2, 4)
	effectstest.AdvanceToMain(t, g)
	grey := effectstest.PutInHand(me, game.Card{Name: "Grey Rock", TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, grey, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a {2} colorless spell was not castable for {1}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	green := effectstest.PutInHand(me, game.Card{Name: "Green Rock", TypeLine: "Sorcery", ManaCost: "{1}{G}", Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "G")
	if err := g.CastSpell(me.ID, green, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a green {1}{G} spell was discounted")
	}
}

func TestOldB308ProcessorAssaultNeedsAnOpponentsExiledCardAndDealsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	card := effectstest.PutInHand(me, game.Card{Name: "Processor Assault", TypeLine: "Sorcery", ManaCost: "{1}{R}", OracleID: oldB308Assault, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "CR")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(victim)}); err == nil {
		t.Fatal("cast with no card to process")
	}
	exiled := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: exiled, Name: "Exiled Card", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(victim), ExileIDs: []uuid.UUID{exiled}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !opp.Graveyard.Contains(exiled) {
		t.Error("the processed card did not go to its owner's graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("5 damage did not kill a 5/5")
	}
}

func TestOldB308BarrageTyrantFlingsADrainedCreaturesPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tyrant := effectstest.B12Push(g, me.ID, "Barrage Tyrant", "Creature — Eldrazi", oldB308Tyrant, 5, 3)
	golem := effectstest.B12Creature(g, me.ID, "Grey Golem", "Artifact Creature — Golem", 4, 4)
	green := oldB308Colored(g, me.ID, "Green Bear", 6, 6)
	life := opp.Life
	effectstest.FloatForTest(g, me, "CCR")
	if err := g.ActivateCatalogAbility(me.ID, tyrant, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, SacrificeIDs: []uuid.UUID{green}, Strict: true}); err == nil {
		t.Fatal("a green creature was sacrificed to Barrage Tyrant")
	}
	if err := g.ActivateCatalogAbility(me.ID, tyrant, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, SacrificeIDs: []uuid.UUID{golem}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(golem) {
		t.Error("the Golem was not sacrificed")
	}
	if opp.Life != life-4 {
		t.Errorf("opponent life %d -> %d, want 4 damage (the Golem's power)", life, opp.Life)
	}
}

func TestOldB308SkitterskinCantBlockAndRegeneratesOnlyWithAnotherColorlessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	skit := effectstest.B12Push(g, me.ID, "Skitterskin", "Creature — Eldrazi Drone", oldB308Skitter, 4, 3)
	effectstest.FloatForTest(g, me, "CB")
	if err := g.ActivateCatalogAbility(me.ID, skit, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("regenerated with no other colorless creature")
	}
	effectstest.B12Creature(g, me.ID, "Grey Golem", "Artifact Creature — Golem", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, skit, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with a Golem out: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB308SkitterskinCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	skit := effectstest.B12Push(g, opp.ID, "Skitterskin", "Creature — Eldrazi Drone", oldB308Skitter, 4, 3)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(skit, attacker); err == nil {
		t.Error("Skitterskin blocked")
	}
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Errorf("an ordinary Wall could not block: %v", err)
	}
}
