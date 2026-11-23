package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b18_cards_test.go — Streets of New Capenna batch 18 (#5491). Each
// test plays the real catalog entry and asserts what its text says.

const (
	sncB18RacketOracle    = "4b51edb1-37cc-4993-a8a5-2eb5f79fa191"
	sncB18ExampleOracle   = "591edddc-3736-44cf-9719-117c24ff3980"
	sncB18CeremonyOracle  = "18ed0c8a-db8f-4247-87c4-544b833f01bd"
	sncB18XanderOracle    = "d17bfa94-5046-46ae-80af-be5d1d35af13"
	sncB18LarcenistOracle = "abd839ab-8b08-4bc0-900a-3a161c62b06c"
	sncB18CrierOracle     = "d88c2375-65ec-4909-aef3-ee1a658a07cc"
	sncB18AttendantOracle = "94b1e7ed-87f3-44e6-bf34-186dd9a1a0d1"
	sncB18ScrivenerOracle = "a945710e-fab6-4b54-ba40-586f2e54baf4"
	sncB18PlansOracle     = "b2f43df5-ec64-4375-b67f-b198732ab0b1"
)

func TestSncB18EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		sncB18RacketOracle:    "Protection Racket",
		sncB18ExampleOracle:   "Make an Example",
		sncB18CeremonyOracle:  "Master of Ceremonies",
		sncB18XanderOracle:    "Lord Xander, the Collector",
		sncB18LarcenistOracle: "Nimble Larcenist",
		sncB18CrierOracle:     "Sky Crier",
		sncB18AttendantOracle: "Mage's Attendant",
		sncB18ScrivenerOracle: "Vampire Scrivener",
		sncB18PlansOracle:     "Change of Plans",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s is %v with caveats %v, want Full and none", name, spec.Completeness, spec.Caveats)
		}
	}
}

// sncB18Push puts a catalog permanent on the battlefield.
func sncB18Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// --- Nimble Larcenist ---------------------------------------------------------------

func TestSncB18NimbleLarcenistExilesAnArtifactInstantOrSorceryFromTheRevealedHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := effectstest.RevealHand(opp, effectstest.RhForest(), effectstest.RhBolt(),
		game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Ritual", TypeLine: "Sorcery", ManaCost: "{B}"},
		game.Card{Name: "Shrine", TypeLine: "Enchantment", ManaCost: "{2}"})
	effectstest.CastAndResolveCreature(t, g, "Nimble Larcenist", "Creature — Bird Rogue", sncB18LarcenistOracle)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if !effectstest.SameIDs(pick.DiscardOptions, []uuid.UUID{hand[1], hand[2], hand[4]}) {
		t.Fatalf("offered %v, want the Bolt, the Sol Ring and the Ritual only", pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{hand[2]}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(hand[2]) || opp.Hand.Contains(hand[2]) {
		t.Error("the Sol Ring was not exiled from their hand")
	}
	if opp.Graveyard.Contains(hand[2]) {
		t.Error("it was discarded, not exiled")
	}
	if opp.Hand.Size() != len(hand)-1 {
		t.Errorf("their hand is %d cards, want %d", opp.Hand.Size(), len(hand)-1)
	}
}

func TestSncB18NimbleLarcenistWithNoMatchingCardExilesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	hand := effectstest.RevealHand(opp, effectstest.RhForest(), game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"})
	effectstest.CastAndResolveCreature(t, g, "Nimble Larcenist", "Creature — Bird Rogue", sncB18LarcenistOracle)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.NoVariant(t, g)
	if opp.Hand.Size() != len(hand) {
		t.Errorf("their hand is %d cards, want %d", opp.Hand.Size(), len(hand))
	}
}

// --- Sky Crier ----------------------------------------------------------------------

func TestSncB18SkyCrierDrawsForYouAndTheTargetOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crier := sncB18Push(g, me.ID, "Sky Crier", "Creature — Bird Citizen", sncB18CrierOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, crier, "flying") || !effectstest.HasEffectiveKeyword(t, g, crier, "lifelink") {
		t.Fatal("Sky Crier has flying and lifelink")
	}
	effectstest.Std51Pay(g, me, "{3}{W}")
	mine, theirs, third := me.Hand.Size(), opp.Hand.Size(), g.Seats[2].Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, crier, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != mine+1 || opp.Hand.Size() != theirs+1 {
		t.Errorf("hands %d→%d and %d→%d, want one card each", mine, me.Hand.Size(), theirs, opp.Hand.Size())
	}
	if g.Seats[2].Hand.Size() != third {
		t.Error("a third player drew")
	}
}

func TestSncB18SkyCrierCannotTargetYourself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crier := sncB18Push(g, me.ID, "Sky Crier", "Creature — Bird Citizen", sncB18CrierOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.Std51Pay(g, me, "{3}{W}")
	if err := g.ActivateCatalogAbility(me.ID, crier, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err == nil {
		t.Fatal("targeted yourself; the target is an opponent")
	}
}

// --- Vampire Scrivener --------------------------------------------------------------

func TestSncB18VampireScrivenerGrowsOnYourLifeGainAndLossDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	viv := sncB18Push(g, me.ID, "Vampire Scrivener", "Creature — Vampire Warlock", sncB18ScrivenerOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.GainLifeNow(t, g, viv, me, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, viv); n != 1 {
		t.Fatalf("gaining 3 life is one counter, got %d", n)
	}
	effectstest.GainLifeNow(t, g, viv, me, -2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, viv); n != 2 {
		t.Fatalf("losing life is a counter too, got %d", n)
	}
	if p, tough := effectstest.PtOf(t, g, viv); p != 4 || tough != 4 {
		t.Errorf("Scrivener is %d/%d, want 4/4", p, tough)
	}
}

func TestSncB18VampireScrivenerIgnoresAnOpponentsTurnAndOtherPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The Scrivener belongs to the second seat, so seat 0's turn is not its turn.
	viv := sncB18Push(g, opp.ID, "Vampire Scrivener", "Creature — Vampire Warlock", sncB18ScrivenerOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.GainLifeNow(t, g, viv, opp, 2)
	effectstest.GainLifeNow(t, g, viv, opp, -1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, viv); n != 0 {
		t.Errorf("an opponent's turn: %d counters, want 0", n)
	}
	// On its controller's own turn another player's life change is not its.
	mine := sncB18Push(g, me.ID, "Vampire Scrivener", "Creature — Vampire Warlock", sncB18ScrivenerOracle, 2, 2)
	effectstest.GainLifeNow(t, g, mine, opp, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, mine); n != 0 {
		t.Errorf("another player's life gain: %d counters, want 0", n)
	}
}

// --- Mage's Attendant ---------------------------------------------------------------

func TestSncB18MagesAttendantMakesAWizardThatCountersANoncreatureSpellUnlessPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastAndResolveCreature(t, g, "Mage's Attendant", "Creature — Cat Rogue", sncB18AttendantOracle)
	effectstest.PassPriorityAroundTable(t, g)
	var wizard uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Wizard" && c.Controller == me.ID {
			wizard = c.InstanceID
			if !c.IsToken() || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "U" {
				t.Fatalf("the Wizard is %+v, want a 1/1 blue token", c)
			}
		}
	}
	if wizard == uuid.Nil {
		t.Fatal("no Wizard token was created")
	}
	// The opponent casts a noncreature spell; the Wizard answers it.
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", Owner: opp.ID, Controller: opp.ID})
	spell := opp.Hand.Cards[len(opp.Hand.Cards)-1].InstanceID
	opp.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.Std51Pay(g, me, "{1}")
	if err := g.ActivateCatalogAbility(me.ID, wizard, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	}); err != nil {
		t.Fatalf("activate the Wizard: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, wizard); ok {
		t.Error("the Wizard was not sacrificed as a cost")
	}
}

// --- Lord Xander, the Collector -----------------------------------------------------

func TestSncB18LordXanderMakesTheOpponentDiscardHalfTheirHandRoundedDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	hand := effectstest.RevealHand(opp, effectstest.RhForest(), effectstest.RhBolt(), effectstest.RhForest(), effectstest.RhBolt(), effectstest.RhForest())
	effectstest.CastAndResolveCreature(t, g, "Lord Xander, the Collector", "Legendary Creature — Vampire Demon Noble", sncB18XanderOracle)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DiscardOwed(g, opp.ID); n != 2 {
		t.Fatalf("a five-card hand owes %d, want 2 (half, rounded down)", n)
	}
	// The opponent chooses what goes.
	effectstest.AnswerDiscard(t, g, opp.ID, hand[0], hand[2])
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != 3 || !opp.Graveyard.Contains(hand[0]) || !opp.Graveyard.Contains(hand[2]) {
		t.Errorf("hand %d, want 3 with the two chosen lands in the graveyard", opp.Hand.Size())
	}
}

func TestSncB18LordXanderDiscardsNothingFromAOneCardHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.RevealHand(opp, effectstest.RhBolt())
	effectstest.CastAndResolveCreature(t, g, "Lord Xander, the Collector", "Legendary Creature — Vampire Demon Noble", sncB18XanderOracle)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardOwed(g, opp.ID) != 0 || opp.Hand.Size() != 1 {
		t.Errorf("half of one rounds down to none: owed %d, hand %d", effectstest.DiscardOwed(g, opp.ID), opp.Hand.Size())
	}
}

func TestSncB18LordXanderMillsHalfTheDefendersLibraryWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	xander := sncB18Push(g, me.ID, "Lord Xander, the Collector", "Legendary Creature — Vampire Demon Noble", sncB18XanderOracle, 6, 6)
	effectstest.SeedLibraryTop(opp, "Extra", "Creature — Bear", "")
	before := opp.Library.Size()
	third := g.Seats[2].Library.Size()
	effectstest.AttackWith(t, g, opp.ID, xander)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != before-before/2 {
		t.Errorf("library %d → %d, want half (%d) milled", before, got, before/2)
	}
	if g.Seats[2].Library.Size() != third {
		t.Error("a player who was not defending was milled")
	}
}

func TestSncB18LordXanderDyingMakesTheOpponentSacrificeHalfTheirNonlandPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	xander := sncB18Push(g, me.ID, "Lord Xander, the Collector", "Legendary Creature — Vampire Demon Noble", sncB18XanderOracle, 6, 6)
	land := sncB18Push(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	var perms []uuid.UUID
	for _, n := range []string{"A", "B", "C", "D", "E"} {
		perms = append(perms, sncB18Push(g, opp.ID, n, "Creature — Bear", "", 2, 2))
	}
	effectstest.KillCreature(t, g, me.ID, xander)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	// Five nonland permanents: half, rounded down, is two, and they choose which.
	effectstest.AnswerSacrifice(t, g, opp.ID, perms[1], perms[3])
	effectstest.PassPriorityAroundTable(t, g)
	left := 0
	for _, id := range perms {
		if g.Battlefield.Contains(id) {
			left++
		}
	}
	if left != 3 || g.Battlefield.Contains(perms[1]) || g.Battlefield.Contains(perms[3]) {
		t.Errorf("%d of the five are left, want 3 with B and D gone", left)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land is not a nonland permanent and was sacrificed")
	}
}

// --- Make an Example ----------------------------------------------------------------

func TestSncB18MakeAnExampleEveryOpponentSplitsYouChooseAndTheChosenPilesAreSacrificedTogether(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2, o3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	a := effectstest.B12Creature(g, o1.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, o1.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, o1.ID, "C", "Creature — Bear", 2, 2)
	d := effectstest.B12Creature(g, o2.ID, "D", "Creature — Bear", 2, 2)
	e := effectstest.B12Creature(g, o2.ID, "E", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	rock := sncB18Push(g, o1.ID, "Rock", "Artifact", "", 0, 0)
	_ = o3 // controls nothing: asks nothing
	effectstest.CastCatalogSpell(t, g, "Make an Example", "Sorcery", sncB18ExampleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	split := effectstest.LatestRevealPickFor(g, o1.ID)
	if split == nil || len(split.ChooseCards) != 3 {
		t.Fatalf("the first opponent separates their three creatures: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(split.ID, o1.ID, []uuid.UUID{a}); err != nil {
		t.Fatal(err)
	}
	// I choose the pile of B and C (the second pile) for them.
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	for _, id := range []uuid.UUID{a, b, c} {
		if !g.Battlefield.Contains(id) {
			t.Fatal("something was sacrificed before every opponent had split")
		}
	}
	split = effectstest.LatestRevealPickFor(g, o2.ID)
	if split == nil || len(split.ChooseCards) != 2 {
		t.Fatalf("the second opponent separates their two creatures: %+v", g.PendingChoices)
	}
	// An empty pile is legal: they keep nothing in the first pile, and I pick it.
	if err := g.ResolveRevealPick(split.ID, o2.ID, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || g.Battlefield.Contains(b) || g.Battlefield.Contains(c) {
		t.Errorf("first opponent: A kept, B and C sacrificed; got a=%v b=%v c=%v",
			g.Battlefield.Contains(a), g.Battlefield.Contains(b), g.Battlefield.Contains(c))
	}
	if !g.Battlefield.Contains(d) || !g.Battlefield.Contains(e) {
		t.Error("the second opponent's chosen pile was empty, so they keep both creatures")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(rock) {
		t.Error("only opponents' creatures are separated")
	}
	if !o1.Graveyard.Contains(b) {
		t.Error("a sacrificed creature goes to its owner's graveyard")
	}
}

// --- Master of Ceremonies -----------------------------------------------------------

func sncB18Tokens(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && c.IsToken() {
			n++
		}
	}
	return n
}

func TestSncB18MasterOfCeremoniesGivesYouAndEachChooserWhatTheyChose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2, o3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	sncB18Push(g, me.ID, "Master of Ceremonies", "Creature — Rhino Druid", sncB18CeremonyOracle, 3, 4)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	mine, h2, h3 := me.Hand.Size(), o2.Hand.Size(), o3.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, o1.ID, 0) // money
	effectstest.AnswerOptionPick(t, g, o2.ID, 1) // friends
	if sncB18Tokens(g, me.ID, "Treasure")+sncB18Tokens(g, me.ID, "Citizen") != 0 {
		t.Fatal("nothing happens until every opponent has chosen")
	}
	effectstest.AnswerOptionPick(t, g, o3.ID, 2) // secrets
	effectstest.PassPriorityAroundTable(t, g)
	if sncB18Tokens(g, me.ID, "Treasure") != 1 || sncB18Tokens(g, o1.ID, "Treasure") != 1 {
		t.Errorf("money: you and that player each get a Treasure (%d, %d)",
			sncB18Tokens(g, me.ID, "Treasure"), sncB18Tokens(g, o1.ID, "Treasure"))
	}
	if sncB18Tokens(g, me.ID, "Citizen") != 1 || sncB18Tokens(g, o2.ID, "Citizen") != 1 {
		t.Errorf("friends: you and that player each get a Citizen (%d, %d)",
			sncB18Tokens(g, me.ID, "Citizen"), sncB18Tokens(g, o2.ID, "Citizen"))
	}
	if me.Hand.Size() != mine+1 || o3.Hand.Size() != h3+1 {
		t.Errorf("secrets: you and that player each draw (%d→%d, %d→%d)", mine, me.Hand.Size(), h3, o3.Hand.Size())
	}
	if o2.Hand.Size() != h2 || sncB18Tokens(g, o3.ID, "Treasure")+sncB18Tokens(g, o3.ID, "Citizen") != 0 {
		t.Error("a player gets only what they chose")
	}
}

func TestSncB18MasterOfCeremoniesTwoPlayersChoosingTheSameGivesYouOnePerPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2, o3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	sncB18Push(g, me.ID, "Master of Ceremonies", "Creature — Rhino Druid", sncB18CeremonyOracle, 3, 4)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, o1.ID, 0)
	effectstest.AnswerOptionPick(t, g, o2.ID, 0)
	effectstest.AnswerOptionPick(t, g, o3.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := sncB18Tokens(g, me.ID, "Treasure"); n != 3 {
		t.Errorf("three players chose money, so you get three Treasures, got %d", n)
	}
	for _, o := range []*game.Player{o1, o2, o3} {
		if sncB18Tokens(g, o.ID, "Treasure") != 1 {
			t.Errorf("%s did not get a Treasure", o.Name)
		}
	}
}

// --- Protection Racket --------------------------------------------------------------

func TestSncB18ProtectionRacketEachOpponentMayPayLifeToExileTheRevealedCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2, o3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	sncB18Push(g, me.ID, "Protection Racket", "Enchantment", sncB18RacketOracle, 0, 0)
	// Pushed in reverse: the first reveal is the top card.
	cRock := effectstest.SeedLibraryTop(me, "Big Spell", "Sorcery", "{3}{R}{R}")
	bSpell := effectstest.SeedLibraryTop(me, "Mid Spell", "Instant", "{1}{U}")
	aSpell := effectstest.SeedLibraryTop(me, "Top Spell", "Sorcery", "{2}{B}")
	o3.Life = 4
	l1, l2 := o1.Life, o2.Life
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	// First opponent: pays 3 to exile the top card.
	effectstest.AnswerOptionPick(t, g, o1.ID, 1)
	if o1.Life != l1-3 || !g.Exile.Contains(aSpell) {
		t.Fatalf("paid %d life and exile=%v, want 3 life and the card exiled", l1-o1.Life, g.Exile.Contains(aSpell))
	}
	// Second opponent: declines, so the next card goes to my hand.
	effectstest.AnswerOptionPick(t, g, o2.ID, 0)
	if o2.Life != l2 || !me.Hand.Contains(bSpell) {
		t.Fatalf("declined: life %d→%d, in my hand=%v", l2, o2.Life, me.Hand.Contains(bSpell))
	}
	// Third opponent can't pay 5 with 4 life: no question, it goes to my hand.
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(cRock) || o3.Life != 4 {
		t.Errorf("a player who can't pay is not asked: in hand=%v, life %d", me.Hand.Contains(cRock), o3.Life)
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want two cards added", hand, me.Hand.Size())
	}
}

// --- Change of Plans ----------------------------------------------------------------

func TestSncB18ChangeOfPlansConnivesEachTargetThenMayPhaseAnyOfThemOut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	one := effectstest.B12Creature(g, me.ID, "One", "Creature — Bear", 2, 2)
	two := effectstest.B12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	three := effectstest.B12Creature(g, me.ID, "Three", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	// Draws: a nonland, then a land.
	effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	effectstest.SeedLibraryTop(me, "Bolt", "Instant", "{R}")
	hand := me.Hand.Size()
	effectstest.PaCast(t, g, "Change of Plans", "Instant", "{X}{1}{U}", sncB18PlansOracle, game.CastSpellParams{
		XValue:  2,
		Targets: effectstest.CardRefs(one, two),
	})
	effectstest.PassPriorityAroundTable(t, g)
	// "One" connives: draws the Bolt, discards it (a nonland card): a counter.
	if hand2 := me.Hand.Size(); hand2 != hand+1 {
		t.Fatalf("the first connive draws a card before discarding: hand %d → %d", hand, hand2)
	}
	discardFromHandCard(t, g, me, "Bolt")
	if n := effectstest.PlusOneCounters(g, one); n != 1 {
		t.Fatalf("a nonland discard puts a +1/+1 counter: %d", n)
	}
	// "Two" connives next: it draws the Forest and discards it, a land: no counter.
	discardFromHandCard(t, g, me, "Forest")
	if n := effectstest.PlusOneCounters(g, two); n != 0 {
		t.Errorf("a land discard puts no counter: %d", n)
	}
	// Now "any number of them" phase out: both targets are offered, the third creature is not.
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil || !effectstest.SameIDs(pick.ChooseCards, []uuid.UUID{one, two}) {
		t.Fatalf("the phase-out choice offers the targets only: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{two}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !sncB18PhasedOut(g, two) {
		t.Error("the creature chosen did not phase out")
	}
	if sncB18PhasedOut(g, one) || sncB18PhasedOut(g, three) {
		t.Error("only the chosen creature phases out")
	}
}

func sncB18PhasedOut(g *game.Game, id uuid.UUID) bool {
	out := false
	g.ReadSnapshot(func() { out = g.IsPhasedOutForEffect(id) })
	return out
}

// discardFromHandCard answers the open discard prompt with the named card.
func discardFromHandCard(t *testing.T, g *game.Game, p *game.Player, name string) {
	t.Helper()
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			effectstest.AnswerDiscard(t, g, p.ID, c.InstanceID)
			return
		}
	}
	t.Fatalf("no %s in hand", name)
}
