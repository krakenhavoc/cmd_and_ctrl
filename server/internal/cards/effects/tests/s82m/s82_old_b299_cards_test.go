package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 299 (Oath of the Gatewatch: activated abilities, counters and
// damage, #5545). Every test plays the real catalog entry through the engine.

const (
	oldB299Collector = "3b0694ff-aa92-4a2e-a366-69fdf5a57c4b"
	oldB299Hedron    = "7cdc9df9-ac6c-49c7-a521-921588696b17"
	oldB299Acolyte   = "c3c40f93-7e10-4139-b9e0-b6dfff8b8038"
	oldB299Pathfind  = "de5a8edb-c41a-4f3a-9231-c54486ba7739"
	oldB299Troll     = "28c44f88-1d47-4894-86ce-4089bd761360"
	oldB299Sire      = "4b1c5278-76a9-4c96-b13a-0737aefb79d9"
	oldB299Wall      = "8771e241-e69b-4e21-9108-223a9488e9c7"
	oldB299Oath      = "53a0923f-7da8-4e31-a231-9b275916e2a2"
	oldB299Assault   = "87038d1b-0fa9-4604-b119-1de56db30144"
	oldB299Devour    = "ea1fc580-a640-4f49-8f29-405c2830eca8"
)

func TestOldB299EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB299Collector: "Kazuul's Toll Collector", oldB299Hedron: "Hedron Alignment",
		oldB299Acolyte: "Stoneforge Acolyte", oldB299Pathfind: "Kozilek's Pathfinder",
		oldB299Troll: "Harvester Troll", oldB299Sire: "Cyclone Sire",
		oldB299Wall: "Wall of Resurgence", oldB299Oath: "Oath of Chandra",
		oldB299Assault: "Pyromancer's Assault", oldB299Devour: "Devour in Flames",
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

func TestOldB299TollCollectorAttachesAnEquipmentAtSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	collector := effectstest.B12Push(g, me.ID, "Kazuul's Toll Collector", "Creature — Ogre Warrior", oldB299Collector, 3, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	sword := effectstest.PushPermanentForTest(g, me.ID, "Sword", "", "Artifact — Equipment")
	g.WithWriteLock(func() { _ = g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: other}) })
	effectstest.ApaActivate(t, g, me, collector, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(sword)})
	if got := effectstest.AttachmentHostOf(t, g, sword); got.ID != collector {
		t.Errorf("the Equipment is attached to %v, want the Collector", got.ID)
	}
	// Not at instant speed.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, collector, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(sword)}); err == nil {
		t.Error("activated outside a main phase")
	}
}

func TestOldB299HedronAlignmentWinsOnlyWithAllFourZones(t *testing.T) {
	for _, tc := range []struct {
		skip string
		win  bool
	}{{"", true}, {"exile", false}, {"hand", false}, {"graveyard", false}, {"battlefield", false}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		mk := func() game.Card {
			return game.Card{InstanceID: uuid.New(), Name: "Hedron Alignment", TypeLine: "Enchantment", OracleID: oldB299Hedron, Owner: me.ID, Controller: me.ID}
		}
		g.WithWriteLock(func() {
			if tc.skip != "battlefield" {
				g.Battlefield.PushTop(mk())
			} else {
				g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Other Hedron", TypeLine: "Enchantment", OracleID: oldB299Hedron, Owner: me.ID, Controller: me.ID})
			}
			if tc.skip != "exile" {
				g.Exile.PushTop(mk())
			}
			if tc.skip != "hand" {
				me.Hand.PushTop(mk())
			}
			if tc.skip != "graveyard" {
				me.Graveyard.PushTop(mk())
			}
		})
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		if won := g.Outcome != nil && g.Outcome.Winner == me.ID; won != tc.win {
			t.Errorf("missing %q: won %v, want %v", tc.skip, won, tc.win)
		}
	}
}

func TestOldB299HedronAlignmentScryAndHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	hedron := effectstest.B12Push(g, me.ID, "Hedron Alignment", "Enchantment", oldB299Hedron, 0, 0)
	effectstest.AssertKeywords(t, g, hedron, "hexproof")
	effectstest.FloatForTest(g, me, "CU")
	if err := g.ActivateCatalogAbility(me.ID, hedron, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("scry: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var scry bool
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceScry {
			scry = true
		}
	}
	if !scry {
		t.Error("no scry prompt")
	}
}

func TestOldB299StoneforgeAcolyteFindsAnEquipmentWithCohort(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	acolyte := effectstest.B12Push(g, me.ID, "Stoneforge Acolyte", "Creature — Kor Artificer Ally", oldB299Acolyte, 1, 2)
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Kor Ally", 1, 1)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Sword", TypeLine: "Artifact — Equipment", ManaCost: "{2}"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}"},
	)
	if err := g.ActivateCatalogAbility(me.ID, acolyte, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("cohort was paid without tapping an Ally")
	}
	effectstest.ApaActivate(t, g, me, acolyte, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{ally}})
	// The prompt offers only the Equipment; take it.
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatal("no choose-cards prompt")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("take the Equipment: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the Equipment did not reach the hand")
	}
}

func TestOldB299KozileksPathfinderStopsOneCreatureBlockingIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, foe := g.Seats[0], g.Seats[1]
	finder := effectstest.B12Push(g, me.ID, "Kozilek's Pathfinder", "Creature — Eldrazi", oldB299Pathfind, 5, 5)
	blocker := effectstest.B12Creature(g, foe.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "C")
	effectstest.ApaActivate(t, g, me, finder, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(blocker)})
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(finder, foe.ID); err != nil {
		t.Fatalf("declare attacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if err := g.DeclareBlocker(blocker, finder); err == nil {
		t.Error("the creature that can't block it blocked it")
	}
}

func TestOldB299HarvesterTrollGrowsOnlyForASacrifice(t *testing.T) {
	for _, sac := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
		troll := effectstest.CastHoldCreature(t, g, "Harvester Troll", "Creature — Troll", oldB299Troll, 2, 3)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) == nil {
			t.Fatal("no sacrifice prompt")
		}
		if sac {
			effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
		} else {
			effectstest.AnswerOwnPermanents(t, g, me.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if sac {
			want = 2
		}
		if got := effectstest.PlusOnes(g, troll); got != want {
			t.Errorf("sacrifice %v: %d counters, want %d", sac, got, want)
		}
		if gone := !effectstest.OnBattlefieldByID(g, fodder); gone != sac {
			t.Errorf("sacrifice %v: fodder gone %v", sac, gone)
		}
	}
}

func TestOldB299WallOfResurgenceAnimatesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	effectstest.CastHoldCreature(t, g, "Wall of Resurgence", "Creature — Wall", oldB299Wall, 0, 6)
	effectstest.AnswerPickTarget(t, g, land)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, land); got != 3 {
		t.Fatalf("land has %d counters, want 3", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, land, "haste") {
		t.Error("the land has no haste")
	}
	c := effectstest.MustBattlefieldCard(t, g, land)
	if !c.IsCreature() || !c.IsLand() || !c.HasSubtype("Elemental") {
		t.Errorf("land is %q, want a land creature — Elemental", c.TypeLine)
	}
	if p := effectstest.RealPowerOf(t, g, land); p != 3 {
		t.Errorf("land has power %d, want 3", p)
	}
}

func TestOldB299CycloneSireDiesIntoALandCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	land := effectstest.PushPermanentForTest(g, me.ID, "Island", "", "Basic Land — Island")
	sire := effectstest.B12Push(g, me.ID, "Cyclone Sire", "Creature — Elemental", oldB299Sire, 3, 4)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, sire); err != nil {
		t.Fatal(err)
	}
	effectstest.AnswerPickTarget(t, g, land)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, land); got != 3 {
		t.Fatalf("land has %d counters, want 3", got)
	}
	if c := effectstest.MustBattlefieldCard(t, g, land); !c.IsCreature() || !c.IsLand() {
		t.Errorf("land is %q, want a land creature", c.TypeLine)
	}
}

func TestOldB299OathOfChandraDamagesACreatureThenPingsAfterAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, foe.ID, "Victim", "Creature — Bear", 2, 3)
	effectstest.CastCatalogSpell(t, g, "Oath of Chandra", "Legendary Enchantment", oldB299Oath, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, victim) {
		t.Error("the 3-toughness creature survived 3 damage")
	}
	// No planeswalker yet: the end step does nothing.
	life := foe.Life
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if foe.Life != life {
		t.Errorf("opponent took %d with no planeswalker this turn", life-foe.Life)
	}
	// A planeswalker enters under my control: the end step pings each opponent.
	// It entered earlier this turn, so go back to a main phase of a fresh turn
	// first: the tally is per turn.
	effectstest.AdvanceToMainOf(t, g, 1)
	walker := game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Chandra", Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(walker) })
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, walker.InstanceID); err != nil {
		t.Fatal(err)
	}
	life = foe.Life
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if foe.Life != life-2 {
		t.Errorf("life %d, want %d after a planeswalker entered under me this turn (on an opponent's end step)", foe.Life, life-2)
	}
}

func TestOldB299PyromancersAssaultShootsOnTheSecondSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me, foe := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Pyromancer's Assault", "Enchantment", oldB299Assault, 0, 0)
	life := foe.Life
	effectstest.CastCatalogSpell(t, g, "Test One", "Instant", "", nil)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("the first spell triggered the Assault")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Test Two", "Instant", "", nil)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the second spell did not trigger the Assault")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: foe.ID}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if foe.Life != life-2 {
		t.Errorf("life %d, want %d", foe.Life, life-2)
	}
}

func TestOldB299DevourInFlamesReturnsALandAndDealsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me, foe := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, foe.ID, "Victim", "Creature — Bear", 2, 5)
	card := game.Card{Name: "Devour in Flames", TypeLine: "Sorcery", ManaCost: "{2}{R}", OracleID: oldB299Devour, Colors: []string{"R"}}
	effectstest.FloatForTest(g, me, "CCR")
	id := effectstest.HandCardFull(me, card.Name, card.TypeLine, card.ManaCost, card.OracleID, card.Colors)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(victim)}); err == nil {
		t.Fatal("cast with no land to return")
	}
	land := effectstest.PushPermanentForTest(g, me.ID, "Mountain", "", "Basic Land — Mountain")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(victim), ReturnIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the land was not returned to hand")
	}
	if effectstest.OnBattlefieldByID(g, victim) {
		t.Error("the 5-toughness creature survived 5 damage")
	}
}
