package testst

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmt_b1_cards_test.go — Teenage Mutant Ninja Turtles, batch 1 (#4457).
// Each test plays the real catalog entry and asserts what its text says.

const (
	tmtB1BishopOracle      = "f33217cc-175f-45e8-bc69-801f41cb4495"
	tmtB1ArmaggonOracle    = "e2a35850-92c5-40ab-8697-2e995819d7ac"
	tmtB1BroadcastOracle   = "e0d221da-16cd-4399-830f-eecc9a0912eb"
	tmtB1ChromeDomeOracle  = "3dc71e16-b935-4ad7-b3f7-945ac2c179a7"
	tmtB1GroundchuckOracle = "da344e46-6bdf-4032-8590-bac850d5713d"
	tmtB1MasterMindOracle  = "52562194-4a97-40d2-aa73-7827f7834f2f"
	tmtB1WarlordOracle     = "2d56a491-b3eb-4f3e-82ab-981d3036c45d"
	tmtB1PetOracle         = "2519ccc0-c206-41fb-aa15-2ef88025ff96"
	tmtB1RobotsOracle      = "45873747-f2a4-4d83-9492-bfc97df7b605"
	tmtB1SallyOracle       = "b529f6b0-c6b2-4eb3-bb48-178e1eda75f3"
	tmtB1SavantiOracle     = "6074632b-4520-4ae8-bcdc-26b1eff71af9"
	tmtB1SlashOracle       = "8abff7b5-fe27-46bd-b951-b5d9f93f3a14"
	tmtB1SouthWindOracle   = "003e5dc6-8b26-4a0a-a50b-5f0806a7bacd"
	tmtB1SuperOracle       = "b7ee76bf-d15a-489e-8f05-414788f8f649"
	tmtB1BovineOracle      = "eaa6b294-ac66-4fe2-b694-97b0a04f7be1"
	tmtB1TriceratonOracle  = "45d40cdb-21cc-4569-a894-83dd3026c947"
	tmtB1TurtlePowerOracle = "15934a19-a29e-4728-816f-20b610eec377"
)

func TestTmtB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmtB1BishopOracle:      "Agent Bishop, Man in Black",
		tmtB1ArmaggonOracle:    "Armaggon, Future Shark",
		tmtB1BroadcastOracle:   "Broadcast Takeover",
		tmtB1ChromeDomeOracle:  "Chrome Dome",
		tmtB1GroundchuckOracle: "Groundchuck & Dirtbag",
		tmtB1MasterMindOracle:  "Krang, Master Mind",
		tmtB1WarlordOracle:     "Krang, Utrom Warlord",
		tmtB1PetOracle:         "Prehistoric Pet",
		tmtB1RobotsOracle:      "Ravenous Robots",
		tmtB1SallyOracle:       "Sally Pride, Lioness Leader",
		tmtB1SavantiOracle:     "Savanti Romero, Time's Exile",
		tmtB1SlashOracle:       "Slash, Reptile Rampager",
		tmtB1SouthWindOracle:   "South Wind Avatar",
		tmtB1SuperOracle:       "Super Shredder",
		tmtB1BovineOracle:      "Transdimensional Bovine",
		tmtB1TriceratonOracle:  "Triceraton Commander",
		tmtB1TurtlePowerOracle: "Turtle Power!",
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

// --- Agent Bishop, Man in Black --------------------------------------------

func TestTmtB1AgentBishopCountersUpToTwoCreaturesAtYourCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bishop := effectstest.B27Push(g, me.ID, "Agent Bishop, Man in Black", "Legendary Creature — Human Soldier", tmtB1BishopOracle, "{2}{W}", 1, 2, "W")
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.AdvanceToStepOf(t, g, 0, game.StepBeginCombat)
	if effectstest.TriggerOnStack(g, bishop) == nil && effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("no Agent Bishop trigger at the beginning of combat")
	}
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(mine, theirs)...)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, mine) != 1 || effectstest.PlusOnes(g, theirs) != 1 || effectstest.PlusOnes(g, bishop) != 0 {
		t.Errorf("counters: mine %d, theirs %d, Bishop %d; want 1, 1, 0", effectstest.PlusOnes(g, mine), effectstest.PlusOnes(g, theirs), effectstest.PlusOnes(g, bishop))
	}
}

func TestTmtB1AgentBishopDoesNotTriggerOnAnOpponentsCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	// Bishop is the opponent's; the combat is seat 0's.
	bishop := effectstest.B27Push(g, opp.ID, "Agent Bishop, Man in Black", "Legendary Creature — Human Soldier", tmtB1BishopOracle, "{2}{W}", 1, 2, "W")
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToStepOf(t, g, 0, game.StepBeginCombat)
	if effectstest.TriggerOnStack(g, bishop) != nil || effectstest.LatestPickTarget(g, opp.ID) != nil {
		t.Error("Agent Bishop triggered at another player's beginning of combat")
	}
}

// --- Armaggon, Future Shark ----------------------------------------------------

func TestTmtB1ArmaggonDestroysUpToThreeCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	d := effectstest.B12Creature(g, opp.ID, "Bear D", "Creature — Bear", 2, 2)

	shark := effectstest.CastCatalogSpell(t, g, "Armaggon, Future Shark", "Legendary Creature — Shark Horror Mutant", tmtB1ArmaggonOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, shark, "flash")
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(a, b, c)...)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if g.Battlefield.Contains(id) {
			t.Errorf("target %s survived", id)
		}
	}
	if !g.Battlefield.Contains(d) || !g.Battlefield.Contains(shark) {
		t.Error("the untargeted creature or Armaggon left the battlefield")
	}
}

// --- Broadcast Takeover ------------------------------------------------------

func TestTmtB1BroadcastTakeoverStealsUntapsAndHastesOpponentsArtifactsForTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	golem := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if id := g.Battlefield.Cards[i].InstanceID; id == rock || id == golem {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})

	effectstest.CastCatalogSpell(t, g, "Broadcast Takeover", "Sorcery", tmtB1BroadcastOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{rock, golem} {
		if got := effectstest.ControllerOf(t, g, id); got != me.ID {
			t.Errorf("%s: controller %s, want the caster", id, got)
		}
		if c := effectstest.TmtB1Card(t, g, id); c.Tapped {
			t.Errorf("%s is still tapped", id)
		}
		if !effectstest.EffectiveAbilitiesContain(t, g, id, "haste") {
			t.Errorf("%s did not gain haste", id)
		}
	}
	if effectstest.ControllerOf(t, g, bear) != opp.ID || effectstest.ControllerOf(t, g, mine) != me.ID {
		t.Error("a non-artifact or the caster's own artifact changed hands")
	}

	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	if effectstest.ControllerOf(t, g, rock) != opp.ID || effectstest.ControllerOf(t, g, golem) != opp.ID {
		t.Error("the artifacts did not return at end of turn")
	}
}

// --- Chrome Dome -------------------------------------------------------------

func TestTmtB1ChromeDomePumpsOtherArtifactCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dome := effectstest.B27Push(g, me.ID, "Chrome Dome", "Artifact Creature — Robot Ninja", tmtB1ChromeDomeOracle, "{2}", 1, 3)
	golem := effectstest.B12Creature(g, me.ID, "My Golem", "Artifact Creature — Golem", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 2, 2)
	if got := effectstest.EffectivePower(t, g, golem); got != 3 {
		t.Errorf("my artifact creature has power %d, want 3", got)
	}
	for id, want := range map[uuid.UUID]int{dome: 1, bear: 2, theirs: 2} {
		if got := effectstest.EffectivePower(t, g, id); got != want {
			t.Errorf("%s has power %d, want %d", id, got, want)
		}
	}
}

func TestTmtB1ChromeDomeCopiesAnotherArtifactWithHasteAndSacrificesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dome := effectstest.B27Push(g, me.ID, "Chrome Dome", "Artifact Creature — Robot Ninja", tmtB1ChromeDomeOracle, "{2}", 1, 3)
	golem := effectstest.B12Creature(g, me.ID, "Steel Golem", "Artifact Creature — Golem", 3, 3)

	if err := effectstest.HobB4Activate(t, g, me.ID, dome, 0, "{C}{C}{C}{C}{C}", effectstest.CardRefs(dome)); err == nil {
		t.Fatal("Chrome Dome copied itself; the target must be another artifact")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, dome, 0, "", effectstest.CardRefs(golem)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, id := range effectstest.TmtB1Named(g, "Steel Golem") {
		if id != golem {
			token = id
		}
	}
	if token == uuid.Nil {
		t.Fatal("no token copy of the Golem")
	}
	c := effectstest.TmtB1Card(t, g, token)
	if !c.IsToken() || !effectstest.EffectiveAbilitiesContain(t, g, token, "haste") {
		t.Errorf("the copy: token %v, abilities %v; want a token with haste", c.IsToken(), effectstest.EffectiveAbilities(t, g, token))
	}
	if effectstest.EffectiveAbilitiesContain(t, g, golem, "haste") {
		t.Error("the original gained haste")
	}
	// Chrome Dome's lord pumps the copy as well: 3 + 1.
	if got := effectstest.EffectivePower(t, g, token); got != 4 {
		t.Errorf("the copy has power %d, want 4", got)
	}

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the token was not sacrificed at the beginning of the end step")
	}
	if !g.Battlefield.Contains(golem) {
		t.Error("the original Golem was sacrificed")
	}
}

// --- Groundchuck & Dirtbag ---------------------------------------------------

func TestTmtB1GroundchuckAddsGreenWhenYouTapALandForMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Groundchuck & Dirtbag", "Legendary Creature — Ox Mole Mutant", tmtB1GroundchuckOracle, "{4}{G}{G}", 8, 8, "G")
	island := effectstest.SeedLand(g, me.ID, "Island", "Basic Land — Island", "")
	effectstest.TapForMana(t, g, me.ID, island)
	if got := effectstest.SortedPool(me); !reflect.DeepEqual(got, []string{"G", "U"}) {
		t.Errorf("pool = %v, want {G}{U}", got)
	}
	theirs := effectstest.SeedLand(g, opp.ID, "Swamp", "Basic Land — Swamp", "")
	effectstest.TapForMana(t, g, opp.ID, theirs)
	if got := effectstest.SortedPool(opp); !reflect.DeepEqual(got, []string{"B"}) {
		t.Errorf("an opponent's land made %v, want only {B}", got)
	}
}

// --- Krang, Master Mind ------------------------------------------------------

func TestTmtB1KrangMasterMindHasAffinityAndCountsOtherArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		effectstest.PushPermanent(g, me.ID, game.Card{Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter"})
	}
	if got := effectstest.SelfPricedMV(t, g, me, tmtB1MasterMindOracle, "Legendary Artifact Creature — Utrom Warrior", "{6}{U}{U}"); got != 5 {
		t.Errorf("with three artifacts Krang costs %d, want 5", got)
	}
	krang := effectstest.B27Push(g, me.ID, "Krang, Master Mind", "Legendary Artifact Creature — Utrom Warrior", tmtB1MasterMindOracle, "{6}{U}{U}", 1, 4, "U")
	if got := effectstest.EffectivePower(t, g, krang); got != 4 {
		t.Errorf("with three other artifacts Krang has power %d, want 4", got)
	}
}

func TestTmtB1KrangMasterMindDrawsUpToFourCardsInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		for me.Hand.Size() > 1 {
			c := me.Hand.Cards[0]
			_, _ = me.Hand.Remove(c.InstanceID)
		}
	})
	effectstest.CastCatalogSpell(t, g, "Krang, Master Mind", "Legendary Artifact Creature — Utrom Warrior", tmtB1MasterMindOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 4 {
		t.Errorf("hand after the trigger = %d, want 4", got)
	}
}

func TestTmtB1KrangMasterMindDrawsNothingWithFourOrMoreCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for me.Hand.Size() < 5 {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	krang := effectstest.CastCatalogSpell(t, g, "Krang, Master Mind", "Legendary Artifact Creature — Utrom Warrior", tmtB1MasterMindOracle, nil)
	before := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TriggerOnStack(g, krang) != nil {
		t.Error("the trigger is waiting with four cards in hand")
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand went %d → %d with four cards in hand, want no draw", before, got)
	}
}

// --- Krang, Utrom Warlord ----------------------------------------------------

func TestTmtB1KrangUtromWarlordGrantsFourKeywordsToOtherArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	krang := effectstest.B27Push(g, me.ID, "Krang, Utrom Warlord", "Legendary Artifact Creature — Utrom Robot", tmtB1WarlordOracle, "{9}", 9, 9)
	golem := effectstest.B12Creature(g, me.ID, "My Golem", "Artifact Creature — Golem", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 2, 2)
	keywords := []string{"flying", "trample", "indestructible", "haste"}
	effectstest.AssertKeywords(t, g, krang, keywords...)
	effectstest.AssertKeywords(t, g, golem, keywords...)
	for _, id := range []uuid.UUID{bear, theirs} {
		for _, kw := range keywords {
			if effectstest.EffectiveAbilitiesContain(t, g, id, kw) {
				t.Errorf("%s gained %s", id, kw)
			}
		}
	}
}

// --- Prehistoric Pet ---------------------------------------------------------

func TestTmtB1PrehistoricPetCantBeBlockedByGreaterPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pet := effectstest.B27Push(g, me.ID, "Prehistoric Pet", "Creature — Dinosaur Ninja", tmtB1PetOracle, "{W}", 1, 2, "W")
	big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 2, 2)
	small := effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(pet, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(big, pet); err == nil {
		t.Error("a creature with greater power blocked Prehistoric Pet")
	}
	if err := g.DeclareBlocker(small, pet); err != nil {
		t.Errorf("a creature with equal power could not block: %v", err)
	}
}

func TestTmtB1PrehistoricPetBouncesAnotherCreatureYouControlOnYourTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	pet := effectstest.B27Push(g, me.ID, "Prehistoric Pet", "Creature — Dinosaur Ninja", tmtB1PetOracle, "{W}", 1, 2, "W")
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, bad := range []uuid.UUID{pet, theirs} {
		if err := effectstest.HobB4Activate(t, g, me.ID, pet, 0, "{C}{W}", effectstest.CardRefs(bad)); err == nil {
			t.Fatalf("Prehistoric Pet bounced %s; it must be another creature you control", bad)
		}
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, pet, 0, "", effectstest.CardRefs(mine)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || !me.Hand.Contains(mine) {
		t.Error("the creature did not return to its owner's hand")
	}

	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	other := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	effectstest.Untap(g, pet)
	if err := effectstest.HobB4Activate(t, g, me.ID, pet, 0, "{C}{W}", effectstest.CardRefs(other)); err == nil {
		t.Error("Prehistoric Pet was activated during an opponent's turn")
	}
}

// --- Ravenous Robots ---------------------------------------------------------

func TestTmtB1RavenousRobotsMakesARobotPerArtifactSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Ravenous Robots", "Artifact Creature — Robot", tmtB1RobotsOracle, "{1}{R}", 2, 1, "R")
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFree(t, g, me, "Bear Cub", "Creature — Bear")
	if got := effectstest.TmtB1Count(g, me.ID, "Robot"); got != 0 {
		t.Fatalf("a creature spell made %d Robots", got)
	}
	effectstest.CastFree(t, g, me, "Rock", "Artifact")
	if got := effectstest.TmtB1Count(g, me.ID, "Robot"); got != 1 {
		t.Fatalf("after my artifact spell: %d Robots, want 1", got)
	}
	if got := effectstest.TmtB1Count(g, opp.ID, "Robot"); got != 0 {
		t.Errorf("the opponent got %d Robots", got)
	}
	robot := effectstest.TmtB1Named(g, "Robot")[0]
	c := effectstest.TmtB1Card(t, g, robot)
	if !c.IsToken() || !c.IsArtifact() || !c.IsCreature() || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 0 {
		t.Errorf("Robot token = %+v, want a 1/1 colorless artifact creature token", c)
	}
}

func TestTmtB1RavenousRobotsHastesCreatureTokensYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	robots := effectstest.B27Push(g, me.ID, "Ravenous Robots", "Artifact Creature — Robot", tmtB1RobotsOracle, "{1}{R}", 2, 1, "R")
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 red Mutant"), 1)
		_ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("2/2 red Mutant"), 1)
	})
	var mine, theirs uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Mutant" && c.Controller == me.ID {
				mine = c.InstanceID
			}
			if c.Name == "Mutant" && c.Controller == opp.ID {
				theirs = c.InstanceID
			}
		}
	})
	if err := effectstest.HobB4Activate(t, g, me.ID, robots, 0, "{R}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, mine, "haste") {
		t.Error("my creature token did not gain haste")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "haste") || effectstest.EffectiveAbilitiesContain(t, g, robots, "haste") {
		t.Error("an opponent's token or the nontoken Robots gained haste")
	}
}

// --- Sally Pride, Lioness Leader ---------------------------------------------

func TestTmtB1SallyPrideMakesAMutantPerNontokenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 black Rat"), 1) })

	effectstest.CastCatalogSpell(t, g, "Sally Pride, Lioness Leader", "Legendary Creature — Cat Mutant Rebel", tmtB1SallyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Two Bears and Sally herself; the Rat is a token and the opponent's Bear is not mine.
	if got := effectstest.TmtB1Count(g, me.ID, "Mutant"); got != 3 {
		t.Errorf("made %d Mutants, want 3", got)
	}
	mutant := effectstest.TmtB1Card(t, g, effectstest.TmtB1Named(g, "Mutant")[0])
	if mutant.Power != 2 || mutant.Toughness != 2 || !reflect.DeepEqual(mutant.Colors, []string{"R"}) {
		t.Errorf("Mutant token = %d/%d %v, want a red 2/2", mutant.Power, mutant.Toughness, mutant.Colors)
	}
}

func TestTmtB1SallyPrideAttackPutsACounterOnEachCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sally := effectstest.B27Push(g, me.ID, "Sally Pride, Lioness Leader", "Legendary Creature — Cat Mutant Rebel", tmtB1SallyOracle, "{3}{W}{W}", 2, 4, "W")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sally, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, sally) != 1 || effectstest.PlusOnes(g, bear) != 1 || effectstest.PlusOnes(g, theirs) != 0 {
		t.Errorf("counters: Sally %d, my Bear %d, theirs %d; want 1, 1, 0", effectstest.PlusOnes(g, sally), effectstest.PlusOnes(g, bear), effectstest.PlusOnes(g, theirs))
	}
}

// --- Savanti Romero, Time's Exile --------------------------------------------

func TestTmtB1SavantiRomeroCountsEveryCounterAfterAddingOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	savanti := effectstest.B27Push(g, me.ID, "Savanti Romero, Time's Exile", "Legendary Creature — Demon Wizard", tmtB1SavantiOracle, "{3}{B}{B}", 4, 4, "B")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(savanti, "time", 2) })
	effectstest.AdvanceToMain(t, g)
	hand, life := me.Hand.Size(), me.Life
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, savanti); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	// Two time counters and the new +1/+1 counter: X = 3.
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
	if got := life - me.Life; got != 3 {
		t.Errorf("lost %d life, want 3", got)
	}
}

// --- Slash, Reptile Rampager -------------------------------------------------

func TestTmtB1SlashAllianceAndAttackMutant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slash := effectstest.B27Push(g, me.ID, "Slash, Reptile Rampager", "Legendary Creature — Mutant Berserker Turtle", tmtB1SlashOracle, "{3}{R}{R}", 7, 5, "R")
	others := append([]*game.Player{}, g.Seats[1:]...)
	start := map[uuid.UUID]int{}
	for _, p := range others {
		start[p.ID] = p.Life
	}
	myLife := me.Life

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("1/1 black Rat"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range others {
		if p.Life != start[p.ID] {
			t.Fatalf("an opponent's creature entering dealt damage to %s", p.ID)
		}
	}

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(slash, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Mutant"); got != 1 {
		t.Fatalf("attack made %d Mutants, want 1", got)
	}
	// The Mutant entering is another creature you control: 2 to each opponent.
	for _, p := range others {
		if got := start[p.ID] - p.Life; got != 2 {
			t.Errorf("opponent %s lost %d life, want 2", p.ID, got)
		}
	}
	if me.Life != myLife {
		t.Errorf("Slash's controller lost %d life", myLife-me.Life)
	}
}

// --- South Wind Avatar -------------------------------------------------------

func TestTmtB1SouthWindAvatarGainsToughnessThenDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	avatar := effectstest.B27Push(g, me.ID, "South Wind Avatar", "Creature — Snake Spirit Avatar", tmtB1SouthWindOracle, "{3}{B}", 3, 4, "B")
	effectstest.AssertKeywords(t, g, avatar, "deathtouch")
	wall := effectstest.B12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 5)
	start := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		start[p.ID] = p.Life
	}
	effectstest.B27Kill(g, wall)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - start[me.ID]; got != 5 {
		t.Errorf("gained %d life, want the Wall's toughness, 5", got)
	}
	for _, p := range g.Seats[1:] {
		if got := start[p.ID] - p.Life; got != 1 {
			t.Errorf("opponent %s lost %d life, want 1", p.ID, got)
		}
	}
}

func TestTmtB1SouthWindAvatarIgnoresItselfAndOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "South Wind Avatar", "Creature — Snake Spirit Avatar", tmtB1SouthWindOracle, "{3}{B}", 3, 4, "B")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 5)
	life := me.Life
	effectstest.B27Kill(g, theirs)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("an opponent's creature dying changed my life by %d", me.Life-life)
	}
}

// --- Super Shredder ----------------------------------------------------------

func TestTmtB1SuperShredderGrowsForEveryOtherPermanentLeaving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shredder := effectstest.B27Push(g, me.ID, "Super Shredder", "Legendary Creature — Mutant Ninja Human", tmtB1SuperOracle, "{1}{B}", 1, 1, "B")
	effectstest.AssertKeywords(t, g, shredder, "menace")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")

	effectstest.B27Kill(g, bear)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(rock) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, shredder); got != 2 {
		t.Errorf("+1/+1 counters = %d after a death and a bounce, want 2", got)
	}
}

// --- Transdimensional Bovine -------------------------------------------------

func TestTmtB1TransdimensionalBovineAddsTwoOfOneColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bovine := effectstest.B27Push(g, me.ID, "Transdimensional Bovine", "Creature — Ox Avatar", tmtB1BovineOracle, "{2}{G}", 0, 4, "G")
	effectstest.AssertKeywords(t, g, bovine, "flying")
	if err := g.ActivateManaAbility(me.ID, bovine, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || pick.ManaAmounts["R"] != 2 {
		t.Fatalf("pick = %+v, want two of any one color", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"R", "R"}) {
		t.Errorf("pool = %v, want [R R]", got)
	}
}

// --- Triceraton Commander ----------------------------------------------------

func TestTmtB1TriceratonCommanderMakesXDinosaurSoldiers(t *testing.T) {
	for _, x := range []int{0, 2} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		cmdr := effectstest.CastXSpell(t, g, "Triceraton Commander", "Creature — Dinosaur Soldier", tmtB1TriceratonOracle, "{X}{X}{W}{W}", x, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(cmdr) {
			t.Fatalf("X=%d: the Commander is not on the battlefield", x)
		}
		if got := effectstest.TmtB1Count(g, me.ID, "Dinosaur Soldier"); got != x {
			t.Errorf("X=%d: %d Dinosaur Soldier tokens, want %d", x, got, x)
		}
	}
}

func TestTmtB1TriceratonCommanderAttackPumpsOtherDinosaurs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdr := effectstest.B27Push(g, me.ID, "Triceraton Commander", "Creature — Dinosaur Soldier", tmtB1TriceratonOracle, "{X}{X}{W}{W}", 2, 2, "W")
	dino := effectstest.B12Creature(g, me.ID, "My Raptor", "Creature — Dinosaur", 3, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Raptor", "Creature — Dinosaur", 3, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(cmdr, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectivePower(t, g, dino) != 4 || effectstest.EffectiveToughness(t, g, dino) != 2 || !effectstest.EffectiveAbilitiesContain(t, g, dino, "flying") {
		t.Errorf("my other Dinosaur: %d/%d, flying %v; want 4/2 with flying",
			effectstest.EffectivePower(t, g, dino), effectstest.EffectiveToughness(t, g, dino), effectstest.EffectiveAbilitiesContain(t, g, dino, "flying"))
	}
	if effectstest.EffectivePower(t, g, cmdr) != 2 || effectstest.EffectivePower(t, g, bear) != 2 || effectstest.EffectivePower(t, g, theirs) != 3 {
		t.Error("the Commander itself, a non-Dinosaur or an opponent's Dinosaur was pumped")
	}
}

// --- Turtle Power! -----------------------------------------------------------

func TestTmtB1TurtlePowerPumpsYourTurtles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	power := effectstest.B27Push(g, me.ID, "Turtle Power!", "Enchantment", tmtB1TurtlePowerOracle, "{2}{G}", 0, 0, "G")
	effectstest.AssertKeywords(t, g, power, "flash")
	turtle := effectstest.B12Creature(g, me.ID, "My Turtle", "Creature — Mutant Ninja Turtle", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Turtle", "Creature — Turtle", 2, 2)
	if effectstest.EffectivePower(t, g, turtle) != 4 || effectstest.EffectiveToughness(t, g, turtle) != 4 {
		t.Errorf("my Turtle is %d/%d, want 4/4", effectstest.EffectivePower(t, g, turtle), effectstest.EffectiveToughness(t, g, turtle))
	}
	if effectstest.EffectivePower(t, g, bear) != 2 || effectstest.EffectivePower(t, g, theirs) != 2 {
		t.Error("a non-Turtle or an opponent's Turtle was pumped")
	}
}
