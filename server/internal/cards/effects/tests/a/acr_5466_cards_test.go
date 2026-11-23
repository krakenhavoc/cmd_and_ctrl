package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// acr_5466_cards_test.go — Assassin's Creed, batch 1 (#5466, S82):
// fifteen cards built from seams the engine already has. Each test plays
// the real catalog entry.

const (
	acrb1DetainedOracle    = "d59488b5-1118-4276-b862-11e860c7c32a"
	acrb1BlacksmithOracle  = "f98d0f8a-c8b5-4fb8-b598-e70ffab3c0ff"
	acrb1PatriarchOracle   = "0c766a86-9c70-465e-874b-9d9f3b4eb090"
	acrb1ViewpointOracle   = "46c404c3-636a-4945-a61c-62f013a9c50f"
	acrb1HookbladeOracle   = "bad2385b-6981-46fa-a012-d3992c3bc118"
	acrb1HaystackOracle    = "d64d69bd-1815-4647-90bc-767fad5c7da7"
	acrb1MisthiosOracle    = "e99c088e-1326-4844-b932-fb51802656f5"
	acrb1AmbushersOracle   = "aea65f96-3c76-4761-9a7a-2d5c10d75ca0"
	acrb1AssassinDenOracle = "826274fd-9831-4122-a5d9-0d1caa3d3a85"
	acrb1EzioOracle        = "2127a245-d664-45cb-8ed1-0024de1a10df"
	acrb1SpartanOracle     = "b255d1ec-49ac-49a2-9af8-21915bbcd224"
	acrb1ShaoJunOracle     = "126aab4d-df6e-4840-8cdc-49e890b1557f"
	acrb1TranquilizeOracle = "1133633a-b717-42cb-b477-4e18584d876f"
	acrb1EagleVisionOracle = "f96f1b4c-f06a-4374-8d8f-aaa06c6108c3"
	acrb1RavenOracle       = "e42450c7-a9de-412f-b439-76f2f9b6031b"
)

func TestAcrb1CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		acrb1DetainedOracle:    "Detained by Legionnaires",
		acrb1BlacksmithOracle:  "Settlement Blacksmith",
		acrb1PatriarchOracle:   "Brotherhood Patriarch",
		acrb1ViewpointOracle:   "Towering Viewpoint",
		acrb1HookbladeOracle:   "Hookblade Veteran",
		acrb1HaystackOracle:    "Haystack",
		acrb1MisthiosOracle:    "Misthios's Fury",
		acrb1AmbushersOracle:   "Brotherhood Ambushers",
		acrb1AssassinDenOracle: "Assassin Den",
		acrb1EzioOracle:        "Ezio, Blade of Vengeance",
		acrb1SpartanOracle:     "Spartan Veteran",
		acrb1ShaoJunOracle:     "Shao Jun",
		acrb1TranquilizeOracle: "Tranquilize",
		acrb1EagleVisionOracle: "Eagle Vision",
		acrb1RavenOracle:       "Keen-Eyed Raven",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

// --- Detained by Legionnaires -------------------------------------------------

func TestAcrb1DetainedByLegionnairesStopsAttacksAndBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Detained by Legionnaires", "Enchantment — Aura", acrb1DetainedOracle, effectstest.CardRefs(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, ogre, game.CantAttackOrBlock)
}

// --- Settlement Blacksmith ----------------------------------------------------

func TestAcrb1SettlementBlacksmithDrawsOnlyWithAnEquipment(t *testing.T) {
	for _, withEquipment := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		if withEquipment {
			effectstest.B12Permanent(g, me.ID, "Test Sword", "Artifact — Equipment")
		}
		effectstest.AdvanceToMain(t, g)
		before := me.Hand.Size()
		effectstest.CastHoldCreature(t, g, "Settlement Blacksmith", "Creature — Human Artificer", acrb1BlacksmithOracle, 3, 3)
		effectstest.PassPriorityAroundTable(t, g)
		want := before
		if withEquipment {
			want = before + 1
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("equipment=%v: hand %d -> %d, want %d", withEquipment, before, got, want)
		}
	}
}

// --- Brotherhood Patriarch ----------------------------------------------------

func TestAcrb1BrotherhoodPatriarchDrainsEachOpponentWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	patriarch := effectstest.B12Push(g, me.ID, "Brotherhood Patriarch", "Creature — Human Assassin", acrb1PatriarchOracle, 4, 1)
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(patriarch) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != lives[0]+2 {
		t.Errorf("your life %d -> %d, want +2", lives[0], me.Life)
	}
	for i := 1; i < len(g.Seats); i++ {
		if g.Seats[i].Life != lives[i]-2 {
			t.Errorf("opponent %d life %d -> %d, want -2", i, lives[i], g.Seats[i].Life)
		}
	}
}

// --- Towering Viewpoint -------------------------------------------------------

func TestAcrb1ToweringViewpointGivesATargetFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wall := effectstest.B12Push(g, me.ID, "Towering Viewpoint", "Artifact Creature — Wall", acrb1ViewpointOracle, 0, 4)
	for _, kw := range []string{"defender", "reach"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, wall, kw) {
			t.Errorf("Towering Viewpoint has no %s", kw)
		}
	}
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.Acrb1AddMana(t, g, me.ID, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, wall, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "flying") {
		t.Error("the target did not gain flying")
	}
	effectstest.DftMPPassTurn(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "flying") {
		t.Error("the flying outlasted the turn")
	}
}

// --- Hookblade Veteran --------------------------------------------------------

func TestAcrb1HookbladeVeteranFliesOnlyDuringItsControllersTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Push(g, me.ID, "Hookblade Veteran", "Creature — Human Assassin", acrb1HookbladeOracle, 1, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Hookblade Veteran", "Creature — Human Assassin", acrb1HookbladeOracle, 1, 2)
	if !effectstest.EffectiveAbilitiesContain(t, g, mine, "flying") || effectstest.EffectiveAbilitiesContain(t, g, theirs, "flying") {
		t.Error("on seat 0's turn only seat 0's Veteran should fly")
	}
	effectstest.DftMPPassTurn(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, mine, "flying") || !effectstest.EffectiveAbilitiesContain(t, g, theirs, "flying") {
		t.Error("on seat 1's turn only seat 1's Veteran should fly")
	}
}

// --- Haystack -----------------------------------------------------------------

func TestAcrb1HaystackPhasesOutYourCreatureUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	haystack := effectstest.PushCatalogPermanent(g, me.ID, "Haystack", "Artifact", acrb1HaystackOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.Acrb1AddMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, haystack, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs), Strict: true}); err == nil {
		t.Fatal("Haystack targeted an opponent's creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, haystack, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if c, _ := effectstest.BattlefieldCard(g, haystack); !c.Tapped {
		t.Error("Haystack did not tap for its cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.PhasedOutInCatalogGame(g, bear) || g.Battlefield.Contains(bear) {
		t.Fatal("the creature did not phase out")
	}
	effectstest.PassTurnsTo(t, g, 0)
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature did not phase in on your next turn")
	}
}

// --- Misthios's Fury ----------------------------------------------------------

func TestAcrb1MisthiossFuryHitsTheControllerOnlyWithAnEquipment(t *testing.T) {
	for _, withEquipment := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if withEquipment {
			effectstest.B12Permanent(g, me.ID, "Test Sword", "Artifact — Equipment")
		}
		ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 4)
		life := opp.Life
		effectstest.CastCatalogSpell(t, g, "Misthios's Fury", "Instant", acrb1MisthiosOracle, effectstest.CardRefs(ogre))
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CardByID(g, ogre).DamageMarked; got != 3 {
			t.Errorf("equipment=%v: %d damage on the creature, want 3", withEquipment, got)
		}
		want := life
		if withEquipment {
			want = life - 2
		}
		if opp.Life != want {
			t.Errorf("equipment=%v: controller life %d -> %d, want %d", withEquipment, life, opp.Life, want)
		}
	}
}

// --- Brotherhood Ambushers and Eagle Vision (freerunning) ---------------------

func TestAcrb1BrotherhoodAmbushersFreerunsAfterAnAssassinConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Acrb1ConnectThenPostcombat(t, g, "Creature — Human Assassin")
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Brotherhood Ambushers", TypeLine: "Creature — Human Assassin", ManaCost: "{4}{B}",
		OracleID: acrb1AmbushersOracle, Colors: []string{"B"}, Power: 6, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	effectstest.FloatForTest(g, me, "CCCB")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "freerunning"}); err != nil {
		t.Fatalf("freerunning {3}{B} was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("Brotherhood Ambushers did not enter")
	}
}

func TestAcrb1EagleVisionFreerunsOnlyAfterAnAssassinConnects(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		want     bool
	}{
		{"Creature — Human Assassin", true},
		{"Creature — Human Warrior", false},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.Acrb1ConnectThenPostcombat(t, g, tc.typeLine)
		spell := effectstest.HandCardFull(me, "Eagle Vision", "Sorcery", "{4}{U}", acrb1EagleVisionOracle, []string{"U"})
		effectstest.FloatForTest(g, me, "CU")
		before := me.Hand.Size()
		err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: "freerunning"})
		if !tc.want {
			if err == nil {
				t.Errorf("%s: freerunning was accepted without an Assassin connecting", tc.typeLine)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: freerunning was refused: %v", tc.typeLine, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Size(); got != before-1+3 {
			t.Errorf("hand %d -> %d, want the spell gone and three cards drawn", before, got)
		}
	}
}

// --- Assassin Den -------------------------------------------------------------

func TestAcrb1AssassinDenCountersAndUnblocksAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	den := effectstest.B12Push(g, me.ID, "Assassin Den", "Creature — Wall", acrb1AssassinDenOracle, 2, 4)
	if !effectstest.EffectiveAbilitiesContain(t, g, den, "defender") {
		t.Error("Assassin Den has no defender")
	}
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.Acrb1AddMana(t, g, me.ID, "{C}{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, den, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Fatal("Assassin Den activated outside a main phase")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.Acrb1AddMana(t, g, me.ID, "{C}{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, den, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs), Strict: true}); err == nil {
		t.Fatal("Assassin Den targeted an opponent's creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, den, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	effectstest.AssertRestrictions(t, g, bear, game.CantBeBlocked)
	effectstest.AssertRestrictions(t, g, theirs, 0)
}

// --- Ezio, Blade of Vengeance -------------------------------------------------

func TestAcrb1EzioDrawsForEachAssassinThatConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ezio := effectstest.B12Push(g, me.ID, "Ezio, Blade of Vengeance", "Legendary Creature — Human Assassin", acrb1EzioOracle, 5, 5)
	if !effectstest.EffectiveAbilitiesContain(t, g, ezio, "deathtouch") {
		t.Error("Ezio has no deathtouch")
	}
	assassin := effectstest.B12Creature(g, me.ID, "My Assassin", "Creature — Human Assassin", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	before := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, assassin, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d, want one card for the Assassin and none for the Bear", before, got)
	}
}

// --- Spartan Veteran ----------------------------------------------------------

func TestAcrb1SpartanVeteranFirstStrikeOnYourTurnAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Push(g, me.ID, "Spartan Veteran", "Creature — Human Soldier", acrb1SpartanOracle, 1, 1)
	theirs := effectstest.B12Push(g, opp.ID, "Spartan Veteran", "Creature — Human Soldier", acrb1SpartanOracle, 1, 1)
	if !effectstest.EffectiveAbilitiesContain(t, g, mine, "first strike") || effectstest.EffectiveAbilitiesContain(t, g, theirs, "first strike") {
		t.Error("on seat 0's turn only seat 0's Veteran should have first strike")
	}
	effectstest.Acrb1AddMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, mine, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, mine); got != 2 {
		t.Errorf("power %d, want 2", got)
	}
	effectstest.DftMPPassTurn(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, mine, "first strike") || !effectstest.EffectiveAbilitiesContain(t, g, theirs, "first strike") {
		t.Error("on seat 1's turn only seat 1's Veteran should have first strike")
	}
	if got := effectstest.EffectivePower(t, g, mine); got != 1 {
		t.Errorf("the pump outlasted the turn: power %d", got)
	}
}

// --- Shao Jun -----------------------------------------------------------------

func TestAcrb1ShaoJunLeapStrikeOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shao := effectstest.B12Push(g, me.ID, "Shao Jun", "Legendary Creature — Human Assassin", acrb1ShaoJunOracle, 3, 3)
	for _, kw := range []string{"flying", "first strike"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, shao, kw) {
			t.Errorf("Shao Jun has no %s on your turn", kw)
		}
	}
	effectstest.DftMPPassTurn(t, g)
	for _, kw := range []string{"flying", "first strike"} {
		if effectstest.EffectiveAbilitiesContain(t, g, shao, kw) {
			t.Errorf("Shao Jun has %s on an opponent's turn", kw)
		}
	}
}

func TestAcrb1ShaoJunRopeDartTapsTwoArtifactsToPingEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shao := effectstest.B12Push(g, me.ID, "Shao Jun", "Legendary Creature — Human Assassin", acrb1ShaoJunOracle, 3, 3)
	a1 := effectstest.B12Permanent(g, me.ID, "Trinket A", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, shao, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a1}, Strict: true}); err == nil {
		t.Fatal("Rope Dart activated with one artifact")
	}
	a2 := effectstest.B12Permanent(g, me.ID, "Trinket B", "Artifact")
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, shao, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a1, a2}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, id := range []uuid.UUID{a1, a2} {
		if c, _ := effectstest.BattlefieldCard(g, id); !c.Tapped {
			t.Error("an artifact was not tapped for the cost")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != lives[0] {
		t.Errorf("your life changed: %d -> %d", lives[0], me.Life)
	}
	for i := 1; i < len(g.Seats); i++ {
		if g.Seats[i].Life != lives[i]-1 {
			t.Errorf("opponent %d life %d -> %d, want -1", i, lives[i], g.Seats[i].Life)
		}
	}
}

// --- Tranquilize --------------------------------------------------------------

func TestAcrb1TranquilizeTapsAndStunsAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Tranquilize", "Sorcery", acrb1TranquilizeOracle, effectstest.CardRefs(mine)); err == nil {
		t.Fatal("Tranquilize targeted your own creature")
	}
	effectstest.CastCatalogSpell(t, g, "Tranquilize", "Sorcery", acrb1TranquilizeOracle, effectstest.CardRefs(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, ogre); !c.Tapped {
		t.Error("the creature was not tapped")
	}
	if n := effectstest.CountersOn(g, ogre, game.CounterStun); n != 3 {
		t.Errorf("%d stun counters, want 3", n)
	}
}

// --- Keen-Eyed Raven ----------------------------------------------------------

func TestAcrb1KeenEyedRavenCountersAnotherCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	raven := effectstest.CastHoldCreature(t, g, "Keen-Eyed Raven", "Creature — Bird", acrb1RavenOracle, 2, 2)
	if !effectstest.EffectiveAbilitiesContain(t, g, raven, "flying") {
		t.Error("Keen-Eyed Raven has no flying")
	}
	ch := effectstest.PickTargetPrompt(t, g)
	if effectstest.HasID(ch.PickTargetCards, raven) {
		t.Error("the Raven offered itself as the \"another\" target")
	}
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the target, want 1", got)
	}
	if got := effectstest.CounterOn(g, raven, game.CounterPlusOne); got != 0 {
		t.Errorf("the Raven put a counter on itself: %d", got)
	}
}
