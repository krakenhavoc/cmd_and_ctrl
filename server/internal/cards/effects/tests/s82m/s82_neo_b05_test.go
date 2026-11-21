package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82Neo05SoulripperOracle   = "060d109f-2969-4f9e-a723-535d662a571a"
	s82Neo05SurgehackerOracle  = "bf9421d2-54e2-44d8-8b39-2e41192a6517"
	s82Neo05BankbusterOracle   = "b6be94dc-7162-4887-b959-3bda709724eb"
	s82Neo05SurveyorOracle     = "92c9b2fd-95d6-4937-aee2-67b4bb8ca29d"
	s82Neo05RecoveryOracle     = "18ad0dd1-2d0e-46b4-8ab1-ece07180267f"
	s82Neo05AceOracle          = "1dbadf75-28cd-4525-8f51-9c53049b9b54"
	s82Neo05TerrariumOracle    = "04b33a43-57f1-46e9-b2c2-1dd3f8a4935c"
	s82Neo05TerminalOracle     = "a0eb1d5b-2c54-44fa-b7ad-bdd4b1147929"
	s82Neo05ArtificerOracle    = "600e1093-5619-49b3-b193-bd263e4fde33"
	s82Neo05ScroungerOracle    = "2750451f-348e-452f-88d0-bc921c46e8cc"
	s82Neo05SteelbreakerOracle = "172e7250-94a4-466f-9438-159174925012"
	s82Neo05SmelterOracle      = "dd645241-2992-4c4e-a6a8-c2215d2340ae"
	s82Neo05AnvilOracle        = "c018f072-8f6d-4835-a717-ed4ecba70d97"
)

func TestS82NeoB05EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82Neo05SoulripperOracle:   "Mukotai Soulripper",
		s82Neo05SurgehackerOracle:  "Surgehacker Mech",
		s82Neo05BankbusterOracle:   "Reckoner Bankbuster",
		s82Neo05SurveyorOracle:     "Aerial Surveyor",
		s82Neo05RecoveryOracle:     "Imperial Recovery Unit",
		s82Neo05AceOracle:          "Kitsune Ace",
		s82Neo05TerrariumOracle:    "Ecologist's Terrarium",
		s82Neo05TerminalOracle:     "Network Terminal",
		s82Neo05ArtificerOracle:    "Automated Artificer",
		s82Neo05ScroungerOracle:    "Undercity Scrounger",
		s82Neo05SteelbreakerOracle: "Scrapyard Steelbreaker",
		s82Neo05SmelterOracle:      "Sokenzan Smelter",
		s82Neo05AnvilOracle:        "Oni-Cult Anvil",
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

func s82Neo05Counters(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	return effectstest.LayeredCard(t, g, id).Counters[kind]
}

// --- Mukotai Soulripper ---

func TestS82NeoB05SoulripperSacrificesForACounterAndMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	v := effectstest.B12Push(g, me.ID, "Mukotai Soulripper", "Artifact Creature — Vehicle", s82Neo05SoulripperOracle, 4, 3)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.DeclareAttack(t, g, opp.ID, v)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCardsOrPick(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if n := s82Neo05Counters(t, g, v, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters %d, want 1", n)
	}
	if !effectstest.HasKeywordOf(g, v, "menace") {
		t.Error("no menace until end of turn")
	}
}

func TestS82NeoB05SoulripperAsksNothingWithNothingToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	v := effectstest.B12Push(g, me.ID, "Mukotai Soulripper", "Artifact Creature — Vehicle", s82Neo05SoulripperOracle, 4, 3)
	effectstest.DeclareAttack(t, g, opp.ID, v)
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82Neo05Counters(t, g, v, game.CounterPlusOne); n != 0 {
		t.Errorf("+1/+1 counters %d, want 0", n)
	}
	if effectstest.HasKeywordOf(g, v, "menace") {
		t.Error("menace without a sacrifice")
	}
}

// --- Surgehacker Mech ---

func TestS82NeoB05SurgehackerDealsTwiceTheVehiclesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Permanent(g, me.ID, "Cart", "Artifact — Vehicle")
	effectstest.B12Permanent(g, opp.ID, "Their Cart", "Artifact — Vehicle")
	small := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	big := effectstest.B12Creature(g, opp.ID, "Hydra", "Creature — Hydra", 7, 7)
	effectstest.RfdCast(t, g, "Surgehacker Mech", "Artifact — Vehicle", s82Neo05SurgehackerOracle, 5, 5)
	effectstest.AnswerPickTargetIfOpen(t, g, small)
	effectstest.PassPriorityAroundTable(t, g)
	// Two Vehicles of mine: 4 damage.
	if g.Battlefield.Contains(small) {
		t.Error("the 4/4 survived 4 damage")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the 7/7 should be untouched")
	}
}

// --- Reckoner Bankbuster ---

func TestS82NeoB05BankbusterEntersWithChargeCountersAndPaysOutOnTheLast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	v := effectstest.RfdCast(t, g, "Reckoner Bankbuster", "Artifact — Vehicle", s82Neo05BankbusterOracle, 4, 4)
	if n := s82Neo05Counters(t, g, v, game.CounterCharge); n != 3 {
		t.Fatalf("charge counters on entry %d, want 3", n)
	}
	handBefore := len(me.Hand.Cards)
	for i := 1; i <= 3; i++ {
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(v) })
		effectstest.FloatMana(t, g, me, "{C}{C}")
		if err := g.ActivateCatalogAbility(me.ID, v, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if i < 3 && (effectstest.CountNamed(g, "Treasure") != 0 || effectstest.CountNamed(g, "Pilot") != 0) {
			t.Fatalf("tokens before the last counter (activation %d)", i)
		}
	}
	if n := s82Neo05Counters(t, g, v, game.CounterCharge); n != 0 {
		t.Errorf("charge counters left %d, want 0", n)
	}
	if effectstest.CountNamed(g, "Treasure") != 1 || effectstest.CountNamed(g, "Pilot") != 1 {
		t.Errorf("Treasures %d, Pilots %d, want 1 and 1", effectstest.CountNamed(g, "Treasure"), effectstest.CountNamed(g, "Pilot"))
	}
	if n := len(me.Hand.Cards) - handBefore; n != 3 {
		t.Errorf("cards drawn %d, want 3", n)
	}
	// With no counter left the ability can't be paid.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(v) })
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, v, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no charge counter to remove")
	}
}

// --- Aerial Surveyor ---

func TestS82NeoB05SurveyorFetchesAPlainsOnlyWhenTheDefenderHasMoreLands(t *testing.T) {
	for _, tc := range []struct {
		name         string
		mine, theirs int
		want         int
	}{{"defender has more", 1, 3, 1}, {"equal", 2, 2, 0}, {"defender has fewer", 3, 1, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			v := effectstest.B12Push(g, me.ID, "Aerial Surveyor", "Artifact Creature — Vehicle", s82Neo05SurveyorOracle, 3, 4)
			for i := 0; i < tc.mine; i++ {
				effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
			}
			for i := 0; i < tc.theirs; i++ {
				effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
			}
			plains := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
			effectstest.PushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
			effectstest.DeclareAttack(t, g, opp.ID, v)
			if tc.want == 0 {
				if effectstest.TriggersOnStackFrom(g, v) != 0 {
					t.Fatal("the trigger went on the stack")
				}
				return
			}
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.SearchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no search prompt")
			}
			effectstest.BeginSearchIfAsked(t, g, c)
			effectstest.AnswerSearchByID(t, g, me.ID, plains)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(plains) {
				t.Fatal("the Plains did not arrive")
			}
			if !effectstest.LayeredCard(t, g, plains).Tapped {
				t.Error("the Plains did not enter tapped")
			}
		})
	}
}

// --- Imperial Recovery Unit ---

func TestS82NeoB05RecoveryUnitReturnsACheapCreatureOrVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	v := effectstest.B12Push(g, me.ID, "Imperial Recovery Unit", "Artifact Creature — Vehicle", s82Neo05RecoveryOracle, 3, 4)
	cheap := effectstest.PushGraveyardPermanent(me, "Cheap Cart", "Artifact — Vehicle", "{2}")
	dear := effectstest.PushGraveyardPermanent(me, "Dear Bear", "Creature — Bear", "{3}")
	rock := effectstest.PushGraveyardPermanent(me, "Cheap Rock", "Artifact", "{1}")
	effectstest.DeclareAttack(t, g, opp.ID, v)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		for _, bad := range []uuid.UUID{dear, rock} {
			for _, id := range p.PickTargetCards {
				if id == bad {
					t.Errorf("%s is offered as a target", bad)
				}
			}
		}
	}
	effectstest.AnswerPickTargetIfOpen(t, g, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(cheap) {
		t.Error("the cheap Vehicle did not return to hand")
	}
	if !me.Graveyard.Contains(dear) || !me.Graveyard.Contains(rock) {
		t.Error("an illegal card left the graveyard")
	}
}

// --- Kitsune Ace ---

func TestS82NeoB05KitsuneAceGivesTheVehicleFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Kitsune Ace", "Creature — Fox Pilot", s82Neo05AceOracle, 2, 2)
	v := effectstest.B12Push(g, me.ID, "Cart", "Artifact Creature — Vehicle", "", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, v)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1}, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOf(g, v, "first strike") {
		t.Error("the Vehicle has no first strike")
	}
}

func TestS82NeoB05KitsuneAceUntapsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ace := effectstest.B12Push(g, me.ID, "Kitsune Ace", "Creature — Fox Pilot", s82Neo05AceOracle, 2, 2)
	v := effectstest.B12Push(g, me.ID, "Cart", "Artifact Creature — Vehicle", "", 3, 3)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(ace) })
	effectstest.DeclareAttack(t, g, opp.ID, v)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1}, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LayeredCard(t, g, ace).Tapped {
		t.Error("the Ace is still tapped")
	}
	if effectstest.HasKeywordOf(g, v, "first strike") {
		t.Error("the Vehicle got first strike from the other bullet")
	}
}

func TestS82NeoB05KitsuneAceIgnoresACreatureThatIsNotAVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Kitsune Ace", "Creature — Fox Pilot", s82Neo05AceOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.NoModePick(t, g, me.ID, "a Bear attacked")
}

// --- Ecologist's Terrarium ---

func TestS82NeoB05TerrariumFetchesABasicLandAndLaterGrowsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.RfdCast(t, g, "Ecologist's Terrarium", "Artifact", s82Neo05TerrariumOracle, 0, 0)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(forest) {
		t.Fatal("the basic land did not reach the hand")
	}
	terr := uuid.Nil
	for _, card := range g.Battlefield.Cards {
		if card.Name == "Ecologist's Terrarium" {
			terr = card.InstanceID
		}
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// Not at sorcery speed: refused during combat.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, terr, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
}

func TestS82NeoB05TerrariumSacrificeGivesACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	terr := effectstest.B12Push(g, me.ID, "Ecologist's Terrarium", "Artifact", s82Neo05TerrariumOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, terr, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(terr) {
		t.Error("the Terrarium was not sacrificed")
	}
	if n := s82Neo05Counters(t, g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters %d, want 1", n)
	}
}

// --- Network Terminal ---

func TestS82NeoB05NetworkTerminalLootsByTappingAnotherArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	term := effectstest.B12Push(g, me.ID, "Network Terminal", "Artifact", s82Neo05TerminalOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, term, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no other artifact to tap")
	}
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.SeedLibrary(me, "Drawn")
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, term, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.LayeredCard(t, g, rock).Tapped || !effectstest.LayeredCard(t, g, term).Tapped {
		t.Error("the Terminal and the other artifact should both be tapped")
	}
}

func TestS82NeoB05NetworkTerminalMakesAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	term := effectstest.B12Push(g, me.ID, "Network Terminal", "Artifact", s82Neo05TerminalOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, term, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	cost, _ := game.ParseCost("{R}")
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(game.Card{Name: "Bolt", TypeLine: "Instant"})) {
		t.Error("the mana is not red")
	}
}

// --- Automated Artificer ---

func TestS82NeoB05ArtificerManaPaysArtifactSpellsAndAbilitiesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	art := effectstest.PushCatalogPermanent(g, me.ID, "Automated Artificer", "Artifact Creature — Artificer", s82Neo05ArtificerOracle, false)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, art, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Artificer: %v", err)
	}
	cost, _ := game.ParseCost("{C}")
	rock := game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}"}
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(rock)) {
		t.Error("refused an artifact spell")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("refused an ability")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) {
		t.Error("paid for a nonartifact spell")
	}
}

// --- Undercity Scrounger ---

func TestS82NeoB05ScroungerNeedsADeathThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	s := effectstest.B12Push(g, me.ID, "Undercity Scrounger", "Artifact Creature — Human Rogue", s82Neo05ScroungerOracle, 1, 4)
	victim := effectstest.B12Creature(g, me.ID, "Victim", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated before any creature died")
	}
	if effectstest.CountNamed(g, "Treasure") != 0 {
		t.Fatal("a Treasure appeared")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(victim); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate after a death: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountNamed(g, "Treasure") != 1 {
		t.Errorf("Treasures %d, want 1", effectstest.CountNamed(g, "Treasure"))
	}
}

// --- Scrapyard Steelbreaker ---

func TestS82NeoB05SteelbreakerSacrificesAnotherArtifactForPlusTwoPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	s := effectstest.B12Push(g, me.ID, "Scrapyard Steelbreaker", "Artifact Creature — Human Warrior", s82Neo05SteelbreakerOracle, 3, 4)
	r1 := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	r2 := effectstest.B12Permanent(g, me.ID, "Rock Two", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "Wastes", "Basic Land")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("sacrificed a land")
	}
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{s}}); err == nil {
		t.Fatal("sacrificed itself")
	}
	for _, r := range []uuid.UUID{r1, r2} {
		if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{r}}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tg := s82One13PT(t, g, s); p != 7 || tg != 6 {
		t.Errorf("after two activations: %d/%d, want 7/6", p, tg)
	}
}

// --- Sokenzan Smelter ---

func s82Neo05ToCombatOffer(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestS82NeoB05SmelterPaysAndSacrificesAnArtifactForAConstruct(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Sokenzan Smelter", "Creature — Goblin Artificer", s82Neo05SmelterOracle, 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	s82Neo05ToCombatOffer(t, g, me)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the beginning of combat offers the optional {1}")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.AnswerChooseCardsOrPick(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	var construct *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Construct" {
			construct = &g.Battlefield.Cards[i]
		}
	}
	if construct == nil {
		t.Fatal("no Construct token")
	}
	if p, tg := s82One13PT(t, g, construct.InstanceID); p != 3 || tg != 1 {
		t.Errorf("Construct is %d/%d, want 3/1", p, tg)
	}
	if !effectstest.HasKeywordOf(g, construct.InstanceID, "haste") || !effectstest.LayeredCard(t, g, construct.InstanceID).IsArtifact() {
		t.Error("the Construct should be an artifact creature with haste")
	}
}

func TestS82NeoB05SmelterDeclinedKeepsTheArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Sokenzan Smelter", "Creature — Goblin Artificer", s82Neo05SmelterOracle, 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	s82Neo05ToCombatOffer(t, g, me)
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) || effectstest.CountNamed(g, "Construct") != 0 {
		t.Error("declining should change nothing")
	}
}

func TestS82NeoB05SmelterAsksNothingWithoutAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Sokenzan Smelter", "Creature — Goblin Artificer", s82Neo05SmelterOracle, 2, 2)
	s82Neo05ToCombatOffer(t, g, me)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("offered a payment with no artifact to sacrifice")
	}
}

// --- Oni-Cult Anvil ---

func TestS82NeoB05AnvilMakesOneConstructPerTurnWhenArtifactsLeaveOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	anvil := effectstest.B12Push(g, me.ID, "Oni-Cult Anvil", "Artifact", s82Neo05AnvilOracle, 0, 0)
	r1 := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	r2 := effectstest.B12Permanent(g, me.ID, "Rock Two", "Artifact")
	r3 := effectstest.B12Permanent(g, me.ID, "Rock Three", "Artifact")
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		for _, r := range []uuid.UUID{r1, r2} {
			if err := g.DestroyPermanentForEffect(r); err != nil {
				t.Fatalf("destroy: %v", err)
			}
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountNamed(g, "Construct"); n != 1 {
		t.Fatalf("Constructs after a batch of two: %d, want 1", n)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(r3); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountNamed(g, "Construct"); n != 1 {
		t.Errorf("a second trigger in the same turn: %d Constructs, want 1", n)
	}
	_ = anvil
}

func TestS82NeoB05AnvilIgnoresAnOpponentsTurnAndNonartifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Oni-Cult Anvil", "Artifact", s82Neo05AnvilOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountNamed(g, "Construct") != 0 {
		t.Fatal("a creature leaving triggered it")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(rock); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountNamed(g, "Construct") != 0 {
		t.Error("an artifact leaving on an opponent's turn triggered it")
	}
}

func TestS82NeoB05AnvilSacrificeDrainsEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	anvil := effectstest.B12Push(g, me.ID, "Oni-Cult Anvil", "Artifact", s82Neo05AnvilOracle, 0, 0)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	before := make([]int, len(g.Seats))
	for i, s := range g.Seats {
		before[i] = s.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, anvil, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, s := range g.Seats {
		want := before[i] - 1
		if i == 0 {
			want = before[i] + 1
		}
		if s.Life != want {
			t.Errorf("seat %d life %d, want %d", i, s.Life, want)
		}
	}
	// The Rock left on my turn: one Construct.
	if n := effectstest.CountNamed(g, "Construct"); n != 1 {
		t.Errorf("Constructs %d, want 1", n)
	}
}
