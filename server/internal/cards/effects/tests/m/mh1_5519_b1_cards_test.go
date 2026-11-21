package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mh1_5519_b1_cards_test.go — Modern Horizons, batch 1 (#5519). Every
// card is registered Full, and each one with rules text is played
// through the engine.

const (
	mh1ImpostorOracle         = "a2b8185a-9ff2-4d65-9bdd-b060b7ce69bb"
	mh1AutomatonOracle        = "d50b1d2e-1825-432a-ab73-068fdc356f50"
	mh1VenomousOracle         = "f1929f1f-6aa3-4308-bbff-cab03883ac68"
	mh1GoblinChampionOracle   = "414446e7-44fa-455c-8cf4-8afa170b6586"
	mh1PutridGoblinOracle     = "27faa94b-c7a7-45e7-8630-80126fab74ba"
	mh1SlugOracle             = "68fbc13c-00b3-49bb-b343-4f2f1e8f5a7c"
	mh1NewMoonOracle          = "bf44b9b7-180c-4d8c-a410-9561e4c8fb9b"
	mh1LancerOracle           = "8ca33b31-4707-40d0-b4ed-cbe625793122"
	mh1CleavingOracle         = "9b33059c-c862-4452-8609-14a0e2551fb3"
	mh1ChosenOracle           = "f664596a-9ad1-4db3-95b0-457698ad2f66"
	mh1HollowheadOracle       = "739ad367-a4ff-4779-8888-3c27f0f60253"
	mh1ScuttlingOracle        = "7c3dcfea-f42c-4515-9b7d-6e9d8cb4d7f9"
	mh1FeasterOracle          = "3e3d4f28-93a0-4313-a9a0-48fd06540cf8"
	mh1NagaOracle             = "08109241-690f-4d09-ad2f-1cddab9b3c64"
	mh1MobOracle              = "5307a61f-2142-4e12-8329-ceb729f6eb79"
	mh1MoxOracle              = "cc319819-68f9-4c1e-8d60-25a5e3e50669"
	mh1NaturesChantOracle     = "46e4ace2-e8bf-4b72-81de-286215886a10"
	mh1OreScaleOracle         = "c4224ae5-4807-45cb-bd4b-70f2c0d9d9f4"
	mh1RainOracle             = "7d5dfa6b-f685-4925-b93a-ae5815a3c0a7"
	mh1RansackOracle          = "9b97a7fe-c5d9-4698-a49c-183b3f1ce4db"
	mh1ExtinctionOracle       = "3937a96c-3be1-465d-98b6-1db67d2215e6"
	mh1RhoxOracle             = "a681c6ea-b10a-4061-abf4-7b45e06ccebd"
	mh1ScourOracle            = "3fb9f0d9-6514-4005-b65a-7929f7c4df06"
	mh1UmezawasCharmOracle    = "7cfd7606-1dab-4627-ba42-37d6e375f9f1"
	mh1SandsprinterOracle     = "6560517d-41c0-4c9c-b9d5-3ecfcb2b99fd"
	mh1WeatherOracle          = "3a091f7b-efc9-4aaa-9da8-157ef8099a36"
	mh1EyekiteOracle          = "fb3f2677-a31d-49ad-a9d1-7324733938d6"
	mh1SpinehornOracle        = "db46b8f4-22bb-4042-abe8-dd83b2426142"
	mh1TwinSilkOracle         = "a2193302-402a-4cf0-8014-5facecad647b"
	mh1SlingGangOracle        = "4e93b23c-a7f7-4bdd-bbca-0dc48bb5c223"
	mh1SmitingHelixOracle     = "19d19e32-af13-4bb3-925f-8ed54cce2fcb"
	mh1ShenanigansOracle      = "1d9b94af-3bfa-4257-8347-675468f0b0e5"
	mh1RecruitOracle          = "e25e2a3f-4add-4744-b1f8-2e880998a51f"
	mh1MurasaOracle           = "0dfdeaa0-fa12-49cf-b361-efa49d3ffcb2"
	mh1WallOracle             = "b97341a7-fc96-4281-b875-fb81065e0157"
	mh1VengefulDevilOracle    = "ba771caa-2cfb-4373-bbd9-6160b0543010"
	mh1SkelementalOracle      = "7c3f2b3d-1517-4818-9675-189949175623"
	mh1MotherBearOracle       = "fe742976-ef13-4c25-972f-254b1ed436de"
	mh1AzraOracle             = "6f6e0abb-18d0-4a15-9b21-0a72542998ac"
	mh1MindRakeOracle         = "1d8e0d40-ab2c-46e2-80c2-69dfdd47aefc"
	mh1StirringAddressOracle  = "45a60f56-35e4-46e5-838f-8b4407a67cd6"
	mh1RuinationRioterOracle  = "8b3cec74-0d32-4762-b3e0-3f4e5eb6f179"
	mh1BearType               = "Creature — Bear"
	mh1SliverType             = "Creature — Sliver"
	mh1ShapeshifterType       = "Creature — Shapeshifter"
	mh1BasicForestType        = "Basic Land — Forest"
	mh1SkelementalTypeLine    = "Creature — Elemental Skeleton"
	mh1SandsprinterTypeLine   = "Creature — Lizard Warrior"
	mh1LightningSkelementalNm = "Lightning Skelemental"
)

// mh1Sliver seeds a plain Sliver creature with no rules text.
func mh1Sliver(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.B12Creature(g, owner, name, mh1SliverType, 1, 1)
}

// mh1Tokens counts the tokens named `name` that `controller` controls.
func mh1Tokens(g *game.Game, name string, controller uuid.UUID) int {
	n := 0
	for _, c := range effectstest.TokensNamed(g, name) {
		if c.Controller == controller {
			n++
		}
	}
	return n
}

// mh1AnyDiscardOwed reports whether any seat has a discard prompt open.
func mh1AnyDiscardOwed(g *game.Game) bool {
	for _, s := range g.Seats {
		if effectstest.DiscardOwed(g, s.ID) > 0 {
			return true
		}
	}
	return false
}

// mh1SettleAnswering empties the stack, ordering the simultaneous
// triggers of "who" as offered and answering any target prompt with
// "target".
func mh1SettleAnswering(t *testing.T, g *game.Game, who, target uuid.UUID) {
	t.Helper()
	for i := 0; i < 48; i++ {
		if p := effectstest.TriggerOrderPromptFor(g, who); p != nil {
			if err := g.ResolveTriggerOrder(p.ID, who, append([]uuid.UUID(nil), p.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := effectstest.LatestPickTarget(g, who); p != nil {
			if err := g.ResolvePickTarget(p.ID, who, game.TargetRef{Kind: game.TargetCard, ID: target}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			return
		}
		effectstest.FinishBlockDeclarations(t, g)
		if effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack did not empty")
}

// --- every card is in the catalog -------------------------------------------

func TestMh1B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		mh1ImpostorOracle, mh1AutomatonOracle, mh1VenomousOracle, mh1GoblinChampionOracle,
		mh1PutridGoblinOracle, mh1SlugOracle, mh1NewMoonOracle, mh1LancerOracle,
		mh1CleavingOracle, mh1ChosenOracle, mh1HollowheadOracle, mh1ScuttlingOracle,
		mh1FeasterOracle, mh1NagaOracle, mh1MobOracle, mh1MoxOracle,
		mh1NaturesChantOracle, mh1OreScaleOracle, mh1RainOracle, mh1RansackOracle,
		mh1ExtinctionOracle, mh1RhoxOracle, mh1ScourOracle, mh1UmezawasCharmOracle,
		mh1SandsprinterOracle, mh1WeatherOracle, mh1EyekiteOracle, mh1SpinehornOracle,
		mh1TwinSilkOracle, mh1SlingGangOracle, mh1SmitingHelixOracle, mh1ShenanigansOracle,
		mh1RecruitOracle, mh1MurasaOracle, mh1WallOracle, mh1VengefulDevilOracle,
		mh1SkelementalOracle, mh1MotherBearOracle, mh1AzraOracle, mh1MindRakeOracle,
		mh1StirringAddressOracle, mh1RuinationRioterOracle,
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

// --- printed keywords ---------------------------------------------------------

func TestMh1B1PrintedKeywords(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Impostor of the Sixth Pride", mh1ShapeshifterType, mh1ImpostorOracle, []string{game.KeywordChangeling}},
		{"Universal Automaton", "Artifact Creature — Shapeshifter", mh1AutomatonOracle, []string{game.KeywordChangeling}},
		{"Venomous Changeling", mh1ShapeshifterType, mh1VenomousOracle, []string{game.KeywordChangeling, "deathtouch"}},
		{"Goblin Champion", "Creature — Goblin Warrior", mh1GoblinChampionOracle, []string{"haste", game.KeywordExalted}},
		{"Putrid Goblin", "Creature — Zombie Goblin", mh1PutridGoblinOracle, []string{game.KeywordPersist}},
		{"Gluttonous Slug", "Creature — Slug Horror", mh1SlugOracle, []string{"menace", game.KeywordEvolve}},
		{"Feaster of Fools", "Creature — Demon", mh1FeasterOracle, []string{"flying"}},
		{"Ore-Scale Guardian", "Creature — Dragon", mh1OreScaleOracle, []string{"flying", "haste"}},
		{"Viashino Sandsprinter", mh1SandsprinterTypeLine, mh1SandsprinterOracle, []string{"trample", "haste"}},
		{"Eyekite", "Creature — Drake", mh1EyekiteOracle, []string{"flying"}},
		{"Twin-Silk Spider", "Creature — Spider", mh1TwinSilkOracle, []string{"reach"}},
		{"Murasa Behemoth", "Creature — Beast", mh1MurasaOracle, []string{"trample"}},
		{"Wall of One Thousand Cuts", "Creature — Wall", mh1WallOracle, []string{"defender", "flying"}},
		{"Vengeful Devil", "Creature — Devil", mh1VengefulDevilOracle, []string{"haste"}},
		{mh1LightningSkelementalNm, mh1SkelementalTypeLine, mh1SkelementalOracle, []string{"trample", "haste"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			abilities := effectstest.EffectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !effectstest.HasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// A changeling is every creature type: a Sliver lord pumps the Impostor.
func TestMh1B1ChangelingIsASliver(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	impostor := effectstest.B12Push(g, me.ID, "Impostor of the Sixth Pride", mh1ShapeshifterType, mh1ImpostorOracle, 3, 1)
	effectstest.B12Push(g, me.ID, "Cleaving Sliver", mh1SliverType, mh1CleavingOracle, 2, 2)
	if got := effectstest.EffectivePower(t, g, impostor); got != 5 {
		t.Errorf("the Impostor has power %d under Cleaving Sliver, want 5", got)
	}
}

// --- exalted, persist, evolve ---------------------------------------------------

func TestMh1B1GoblinChampionExaltsALoneAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Goblin Champion", "Creature — Goblin Warrior", mh1GoblinChampionOracle, 0, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", mh1BearType, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the lone attacker is %d/%d, want 3/3", p, tough)
	}
}

func TestMh1B1PutridGoblinPersists(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Putrid Goblin", "Creature — Zombie Goblin", mh1PutridGoblinOracle, 2, 2)
	effectstest.B25Destroy(g, id)
	effectstest.SettleProwess(t, g)
	back, ok := effectstest.VisBattlefieldNamed(g, me.ID, "Putrid Goblin")
	if !ok {
		t.Fatal("Putrid Goblin did not return with persist")
	}
	if n := back.Counters[game.CounterMinusOne]; n != 1 {
		t.Errorf("it returned with %d -1/-1 counters, want 1", n)
	}
}

func TestMh1B1GluttonousSlugEvolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	slug := effectstest.B12Push(g, me.ID, "Gluttonous Slug", "Creature — Slug Horror", mh1SlugOracle, 0, 3)
	effectstest.EnterCreature(t, g, me.ID, "Wall", 0, 2)
	effectstest.SettleProwess(t, g)
	if n := effectstest.PlusOneCounters(g, slug); n != 0 {
		t.Fatalf("a 0/2 grew the 0/3 Slug: %d counters", n)
	}
	// A 1/1 has greater power than the 0/3: the Slug evolves to 1/4.
	effectstest.EnterCreature(t, g, me.ID, "Grizzly", 1, 1)
	effectstest.SettleProwess(t, g)
	if n := effectstest.PlusOneCounters(g, slug); n != 1 {
		t.Fatalf("the Slug has %d +1/+1 counters after a 1/1 entered, want 1", n)
	}
	// Another 1/1 is greater in neither against the 1/4.
	effectstest.EnterCreature(t, g, me.ID, "Other Grizzly", 1, 1)
	effectstest.SettleProwess(t, g)
	if n := effectstest.PlusOneCounters(g, slug); n != 1 {
		t.Errorf("the Slug has %d +1/+1 counters after a second 1/1 entered, want 1", n)
	}
}

// --- ninjutsu --------------------------------------------------------------------

func TestMh1B1NinjasArriveWithNinjutsu(t *testing.T) {
	for _, tc := range []struct{ name, oracle, mana string }{
		{"Ninja of the New Moon", mh1NewMoonOracle, "{B}{B}{B}{B}"},
		{"Mist-Syndicate Naga", mh1NagaOracle, "{U}{U}{U}"},
		{"Azra Smokeshaper", mh1AzraOracle, "{B}{B}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			ninja, attacker, me, _ := effectstest.NinjutsuBoard(t, g, tc.name, tc.oracle, tc.mana)
			if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{ReturnIDs: []uuid.UUID{attacker}}); err != nil {
				t.Fatalf("activate ninjutsu: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerPickTargetIfOpen(t, g, ninja)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(ninja) {
				t.Fatal("the ninja is not on the battlefield")
			}
			if !me.Hand.Contains(attacker) {
				t.Error("the attacker was not returned to hand")
			}
		})
	}
}

func TestMh1B1MistSyndicateNagaCopiesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	naga := effectstest.B12Push(g, me.ID, "Mist-Syndicate Naga", "Creature — Snake Ninja", mh1NagaOracle, 3, 1)
	effectstest.DealCombatDamage(g, me.ID, naga, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := mh1Tokens(g, "Mist-Syndicate Naga", me.ID); n != 1 {
		t.Errorf("%d token copies, want 1", n)
	}
}

func TestMh1B1AzraSmokeshaperGivesIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", mh1BearType, 2, 2)
	effectstest.CastHoldCreature(t, g, "Azra Smokeshaper", "Creature — Azra Ninja", mh1AzraOracle, 3, 3)
	mh1SettleAnswering(t, g, me.ID, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the Bear did not gain indestructible")
	}
}

// --- Slivers ------------------------------------------------------------------

func TestMh1B1LancerSliverGivesYourSliversFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := mh1Sliver(g, me.ID, "My Sliver")
	theirs := mh1Sliver(g, opp.ID, "Their Sliver")
	lancer := effectstest.B12Push(g, me.ID, "Lancer Sliver", mh1SliverType, mh1LancerOracle, 2, 2)
	if !effectstest.HasKeywordOf(g, mine, "first strike") || !effectstest.HasKeywordOf(g, lancer, "first strike") {
		t.Error("your Slivers, the Lancer included, should have first strike")
	}
	if effectstest.HasKeywordOf(g, theirs, "first strike") {
		t.Error("an opponent's Sliver has first strike")
	}
}

func TestMh1B1CleavingSliverPumpsYourSlivers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := mh1Sliver(g, me.ID, "My Sliver")
	theirs := mh1Sliver(g, opp.ID, "Their Sliver")
	bear := effectstest.B12Creature(g, me.ID, "Bear", mh1BearType, 2, 2)
	effectstest.B12Push(g, me.ID, "Cleaving Sliver", mh1SliverType, mh1CleavingOracle, 2, 2)
	if p, tough := effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 3 || tough != 1 {
		t.Errorf("your Sliver is %d/%d, want 3/1", p, tough)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 1 {
		t.Errorf("an opponent's Sliver has power %d, want 1", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("the Bear has power %d, want 2", p)
	}
}

// Each Sliver you control has exalted, so a lone Sliver attacker gets
// +1/+1 for the Chosen and +1/+1 for itself.
func TestMh1B1FirstSliversChosenGivesEachSliverExalted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "First Sliver's Chosen", mh1SliverType, mh1ChosenOracle, 3, 3)
	attacker := mh1Sliver(g, me.ID, "My Sliver")
	theirs := mh1Sliver(g, opp.ID, "Their Sliver")
	if effectstest.HasKeywordOf(g, theirs, game.KeywordExalted) {
		t.Error("an opponent's Sliver has exalted")
	}
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, attacker); p != 3 || tough != 3 {
		t.Errorf("the lone Sliver is %d/%d, want 3/3", p, tough)
	}
}

func TestMh1B1HollowheadSliverGrantsARummage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := mh1Sliver(g, me.ID, "My Sliver")
	theirs := mh1Sliver(g, opp.ID, "Their Sliver")
	effectstest.B12Push(g, me.ID, "Hollowhead Sliver", mh1SliverType, mh1HollowheadOracle, 2, 2)
	if idx, _ := effectstest.GrantedActivatedIndex(t, g, theirs); idx >= 0 {
		t.Error("an opponent's Sliver has the granted ability")
	}
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.D2dSetHand(me, 2)
	pitched := me.Hand.Cards[0].InstanceID
	idx, ref := effectstest.GrantedActivatedIndex(t, g, mine)
	if idx < 0 {
		t.Fatal("your Sliver has no granted ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, mine, idx, game.ActivateAbilityParams{Ref: ref, DiscardIDs: []uuid.UUID{pitched}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitched) {
		t.Error("the discarded card is not in the graveyard")
	}
	if me.Hand.Size() != 2 {
		t.Errorf("hand %d, want 2 after discarding one and drawing one", me.Hand.Size())
	}
	if !effectstest.B13Tapped(t, g, mine) {
		t.Error("the Sliver did not tap")
	}
}

func TestMh1B1ScuttlingSliverUntapsASliver(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := mh1Sliver(g, me.ID, "My Sliver")
	effectstest.B12Push(g, me.ID, "Scuttling Sliver", mh1SliverType, mh1ScuttlingOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(mine) })
	if !effectstest.B13Tapped(t, g, mine) {
		t.Fatal("the Sliver did not tap")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.TspB1ActivateGranted(t, g, me.ID, mine, nil)
	if effectstest.B13Tapped(t, g, mine) {
		t.Error("the Sliver is still tapped")
	}
}

// --- spells -----------------------------------------------------------------------

func TestMh1B1MobConvokesToDestroyACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var tappers []uuid.UUID
	for i := 0; i < 4; i++ {
		tappers = append(tappers, effectstest.PushTapCostPermanent(g, me.ID, "Soldier", "Creature — Soldier", "W"))
	}
	tappers = append(tappers, effectstest.PushTapCostPermanent(g, me.ID, "Cutpurse", "Creature — Human Rogue", "B"))
	victim := effectstest.B12Creature(g, opp.ID, "Victim", mh1BearType, 2, 2)
	if _, err := effectstest.CastWithTapParams(t, g, "Mob", "Instant", "{4}{B}", mh1MobOracle,
		game.CastSpellParams{Strict: true, TapIDs: tappers, Targets: effectstest.CardTarget(victim)}); err != nil {
		t.Fatalf("convoke cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the target creature survived")
	}
}

func TestMh1B1ArtifactRemoval(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		hitsEnchantment        bool
	}{
		{"Nature's Chant", "Instant", mh1NaturesChantOracle, true},
		{"Shenanigans", "Sorcery", mh1ShenanigansOracle, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			opp := effectstest.Top100Opponent(g)
			rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
			aura := effectstest.B12Permanent(g, opp.ID, "Charm", "Enchantment")
			err := effectstest.CastCatalogSpellErr(t, g, tc.name, tc.typeLine, tc.oracle, effectstest.CardTarget(aura))
			if tc.hitsEnchantment == (err != nil) {
				t.Fatalf("targeting an enchantment: err = %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, effectstest.CardTarget(rock))
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(rock) {
				t.Error("the artifact survived")
			}
		})
	}
}

func TestMh1B1RainOfRevelationDrawsThreeThenDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Rain of Revelation", "Instant", mh1RainOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d before the discard, want %d", got, hand+3)
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 1 {
		t.Errorf("owes %d discards, want 1", n)
	}
}

func TestMh1B1RansackTheLabTakesOneAndBinsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.FinB3LibraryCards(me, "Fourth", "Third", "Second", "First") // First is on top
	effectstest.CastCatalogSpell(t, g, "Ransack the Lab", "Sorcery", mh1RansackOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no take prompt")
	}
	if len(pick.ChooseCards) != 3 {
		t.Errorf("offered %d cards, want 3", len(pick.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[2])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[2]) {
		t.Error("the chosen card did not come to hand")
	}
	if !me.Graveyard.Contains(ids[1]) || !me.Graveyard.Contains(ids[3]) {
		t.Error("the rest did not go to the graveyard")
	}
	if !me.Library.Contains(ids[0]) {
		t.Error("the fourth card left the library")
	}
}

func TestMh1B1ReturnFromExtinctionReturnsACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.WthGraveyardCard(me, "Bear", mh1BearType)
	effectstest.CastModal(t, g, "Return from Extinction", "Sorcery", mh1ExtinctionOracle, []int{0}, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the creature card did not return to hand")
	}
}

func TestMh1B1ReturnFromExtinctionNeedsASharedType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.WthGraveyardCard(me, "Bear", mh1BearType)
	elf := effectstest.WthGraveyardCard(me, "Elf", "Creature — Elf")
	other := effectstest.WthGraveyardCard(me, "Other Bear", mh1BearType)
	if err := effectstest.WthCastWithModes(t, g, "Return from Extinction", "Sorcery", mh1ExtinctionOracle, []int{1}, effectstest.CardRefs(bear, elf)); err == nil {
		t.Fatal("two creature cards that share no type were targeted")
	}
	effectstest.CastModal(t, g, "Return from Extinction", "Sorcery", mh1ExtinctionOracle, []int{1}, effectstest.CardRefs(bear, other))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || !me.Hand.Contains(other) {
		t.Error("the two Bears did not return to hand")
	}
}

func TestMh1B1ScourAllPossibilitiesScriesThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedLibrary(me, "Junk", "Gas", "Third")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Scour All Possibilities", "Sorcery", mh1ScourOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("looked at %d cards, want 2", len(c.ScryCards))
	}
	if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{ids[0]}, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("it did not draw the card left on top")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand delta %d, want 1", got)
	}
}

func TestMh1B1UmezawasCharmModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", mh1BearType, 2, 2)
	effectstest.CastModal(t, g, "Umezawa's Charm", "Instant", mh1UmezawasCharmOracle, []int{0}, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d after +2/+2, want 4/4", p, tough)
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", mh1BearType, 2, 2)
	effectstest.CastModal(t, g, "Umezawa's Charm", "Instant", mh1UmezawasCharmOracle, []int{1}, effectstest.CardTarget(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("their Bear is %d/%d after -1/-1, want 1/1", p, tough)
	}
	life := me.Life
	effectstest.CastModal(t, g, "Umezawa's Charm", "Instant", mh1UmezawasCharmOracle, []int{2}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestMh1B1WeatherTheStormGainsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Weather the Storm", "Instant", mh1WeatherOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestMh1B1SmitingHelixDealsThreeAndGainsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	mine, theirs := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Smiting Helix", "Sorcery", mh1SmitingHelixOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-3 || me.Life != mine+3 {
		t.Errorf("opponent %d -> %d, you %d -> %d; want -3 and +3", theirs, opp.Life, mine, me.Life)
	}
}

func TestMh1B1SmitingHelixFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	id := effectstest.SeedGraveyardCard(t, g, "Smiting Helix", "Sorcery", mh1SmitingHelixOracle)
	theirs := opp.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-3 {
		t.Errorf("opponent %d, want %d", opp.Life, theirs-3)
	}
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back spell is not in exile")
	}
}

func TestMh1B1RecruitTheWorthyMakesASoldier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.CastCatalogSpell(t, g, "Recruit the Worthy", "Instant", mh1RecruitOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := mh1Tokens(g, "Soldier", me.ID); n != 1 {
		t.Errorf("%d Soldier tokens, want 1", n)
	}
}

func TestMh1B1MindRakeTargetsOrOverloads(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.D2dSetHand(opp, 3)
	effectstest.CastCatalogSpell(t, g, "Mind Rake", "Sorcery", mh1MindRakeOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DiscardOwed(g, opp.ID); n != 2 {
		t.Errorf("the target owes %d discards, want 2", n)
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 0 {
		t.Errorf("the caster owes %d discards, want 0", n)
	}

	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g2)
	me2 := effectstest.Top100ActiveSeat(g2)
	for _, s := range g2.Seats {
		effectstest.D2dSetHand(s, 3)
	}
	id := effectstest.B13HandCard(me2, "Mind Rake", "Sorcery", mh1MindRakeOracle)
	if err := g2.CastSpell(me2.ID, id, game.CastSpellParams{AlternativeCost: "overload"}); err != nil {
		t.Fatalf("overload cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if !mh1AnyDiscardOwed(g2) {
		t.Error("the overloaded Mind Rake asked nobody to discard")
	}
}

func TestMh1B1StirringAddressTargetsOrOverloads(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", mh1BearType, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other Bear", mh1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", mh1BearType, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Stirring Address", "Instant", mh1StirringAddressOracle, effectstest.CardTarget(theirs)); err == nil {
		t.Fatal("Stirring Address targeted a creature you don't control")
	}
	effectstest.CastCatalogSpell(t, g, "Stirring Address", "Instant", mh1StirringAddressOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 4 {
		t.Errorf("the target has power %d, want 4", p)
	}
	if p, _ := effectstest.PtOf(t, g, other); p != 2 {
		t.Errorf("the other Bear has power %d, want 2", p)
	}
	id := effectstest.B13HandCard(me, "Stirring Address", "Instant", mh1StirringAddressOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "overload"}); err != nil {
		t.Fatalf("overload cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 6 {
		t.Errorf("the first Bear has power %d after the overload, want 6", p)
	}
	if p, _ := effectstest.PtOf(t, g, other); p != 4 {
		t.Errorf("the other Bear has power %d after the overload, want 4", p)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 {
		t.Errorf("their Bear has power %d, want 2", p)
	}
}

// --- creatures with abilities -------------------------------------------------------

func TestMh1B1MoxTantaliteSuspendsAndTapsForAnyColor(t *testing.T) {
	sa := effectstest.SuspendOf(mh1MoxOracle)
	if sa == nil {
		t.Fatal("Mox Tantalite declares no suspend")
	}
	if sa.Counters != 3 || sa.Cost != "{0}" {
		t.Errorf("suspend %d—%q, want 3—{0}", sa.Counters, sa.Cost)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mox := effectstest.PushCatalogPermanent(g, me.ID, "Mox Tantalite", "Artifact", mh1MoxOracle, false)
	effectstest.B28TapForMana(t, g, me.ID, mox, "B")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want [B]", got)
	}
}

func TestMh1B1FeasterOfFoolsConvokesAndDevours(t *testing.T) {
	spec, ok := effects.Lookup(mh1FeasterOracle)
	if !ok {
		t.Fatal("Feaster of Fools is not registered")
	}
	if spec.TapCost == nil {
		t.Error("no convoke")
	}
	if len(spec.Replacements) != 1 {
		t.Errorf("%d replacements, want devour's one", len(spec.Replacements))
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fodder := effectstest.PushDevourFodder(g, me.ID, "Goblin A")
	id := effectstest.CastDevourer(t, g, "Feaster of Fools", mh1FeasterOracle)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, fodder)
	if got := effectstest.DevourPlusOnes(g, id); got != 2 {
		t.Errorf("counters = %d, want 2 (devour 2 x 1)", got)
	}
	if !me.Graveyard.Contains(fodder) {
		t.Error("the devoured creature is not in the graveyard")
	}
}

func TestMh1B1OreScaleGuardianCostsLessPerLandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	for i := 0; i < 5; i++ {
		effectstest.WthGraveyardCard(me, "Forest", mh1BasicForestType)
	}
	effectstest.FloatMana(t, g, me, "{R}{R}")
	if _, err := effectstest.CastWithTapParams(t, g, "Ore-Scale Guardian", "Creature — Dragon", "{5}{R}{R}", mh1OreScaleOracle,
		game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {R}{R} with five land cards in the graveyard: %v", err)
	}
}

func TestMh1B1DrawnTwoStatics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "Draw A", "Draw B", "Draw C")
	kite := effectstest.B12Push(g, me.ID, "Eyekite", "Creature — Drake", mh1EyekiteOracle, 1, 2)
	minotaur := effectstest.B12Push(g, me.ID, "Spinehorn Minotaur", "Creature — Minotaur Berserker", mh1SpinehornOracle, 2, 3)
	for g.TurnTallyFor(me.ID).CardsDrawn < 2 {
		effectstest.AvB9Draw(t, g, me.ID, 1)
	}
	if got := effectstest.EffectivePower(t, g, kite); got != 3 {
		t.Errorf("Eyekite has power %d after two draws, want 3", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, minotaur, "double strike") {
		t.Error("Spinehorn Minotaur has no double strike after two draws")
	}
}

func TestMh1B1DrawnTwoStaticsAreOffBeforeTwoDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kite := effectstest.B12Push(g, me.ID, "Eyekite", "Creature — Drake", mh1EyekiteOracle, 1, 2)
	minotaur := effectstest.B12Push(g, me.ID, "Spinehorn Minotaur", "Creature — Minotaur Berserker", mh1SpinehornOracle, 2, 3)
	if g.TurnTallyFor(me.ID).CardsDrawn >= 2 {
		t.Skip("two cards already drawn this turn")
	}
	if got := effectstest.EffectivePower(t, g, kite); got != 1 {
		t.Errorf("Eyekite has power %d, want 1", got)
	}
	if effectstest.HasEffectiveKeyword(t, g, minotaur, "double strike") {
		t.Error("Spinehorn Minotaur has double strike")
	}
}

func TestMh1B1MurasaBehemothGrowsWithALandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beast := effectstest.B12Push(g, me.ID, "Murasa Behemoth", "Creature — Beast", mh1MurasaOracle, 5, 5)
	effectstest.WthGraveyardCard(me, "Bear", mh1BearType)
	g.BumpLayerVersionForTest()
	if p, tough := effectstest.PtOf(t, g, beast); p != 5 || tough != 5 {
		t.Errorf("%d/%d with no land card in the graveyard, want 5/5", p, tough)
	}
	effectstest.WthGraveyardCard(me, "Forest", mh1BasicForestType)
	g.BumpLayerVersionForTest()
	if p, tough := effectstest.PtOf(t, g, beast); p != 8 || tough != 8 {
		t.Errorf("%d/%d with a land card in the graveyard, want 8/8", p, tough)
	}
}

func TestMh1B1TwinSilkSpiderMakesASpider(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Twin-Silk Spider", "Creature — Spider", mh1TwinSilkOracle, 1, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := mh1Tokens(g, "Spider", me.ID); n != 1 {
		t.Errorf("%d Spider tokens, want 1", n)
	}
}

func TestMh1B1SlingGangLieutenantMakesGoblinsAndDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	lt := effectstest.CastHoldCreature(t, g, "Sling-Gang Lieutenant", "Creature — Goblin", mh1SlingGangOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := mh1Tokens(g, "Goblin", me.ID); n != 2 {
		t.Fatalf("%d Goblin tokens, want 2", n)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", mh1BearType, 2, 2)
	mine, theirs := me.Life, opp.Life
	if err := g.ActivateCatalogAbility(me.ID, lt, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a Bear paid the Goblin sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, lt, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, SacrificeIDs: []uuid.UUID{lt}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lt) || opp.Life != theirs-1 || me.Life != mine+1 {
		t.Errorf("Lieutenant on battlefield %v, opponent %d -> %d, you %d -> %d",
			g.Battlefield.Contains(lt), theirs, opp.Life, mine, me.Life)
	}
}

func TestMh1B1WallOfOneThousandCutsCanAttackThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	wall := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Wall of One Thousand Cuts", OracleID: mh1WallOracle,
		TypeLine: "Creature — Wall", Power: 3, Toughness: 5})
	effectstest.AdvanceToMainOf(t, g, 0)
	before := g.Clone()
	effectstest.AdvanceTo(t, before, game.StepDeclareAttackers)
	effectstest.WantDefenderRefused(t, before, wall, b.ID)
	effectstest.FloatMana(t, g, me, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, wall, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := effectstest.DeclareResult(g, wall, b.ID); err != nil {
		t.Errorf("attack after paying {W}: %v", err)
	}
}

func TestMh1B1VengefulDevilPingsOnlyAfterADeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	devil := effectstest.B12Push(g, me.ID, "Vengeful Devil", "Creature — Devil", mh1VengefulDevilOracle, 1, 1)
	target := []game.TargetRef{effectstest.PlayerTarget(opp.ID)}
	if err := g.ActivateCatalogAbility(me.ID, devil, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err == nil {
		t.Fatal("activated with no creature dead this turn")
	}
	victim := effectstest.B12Creature(g, opp.ID, "Victim", mh1BearType, 2, 2)
	effectstest.B27Kill(g, victim)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, devil, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err != nil {
		t.Fatalf("activate after a death: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent %d, want %d", opp.Life, life-1)
	}
}

func TestMh1B1LightningSkelementalMakesThemDiscardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skel := effectstest.B12Push(g, me.ID, mh1LightningSkelementalNm, mh1SkelementalTypeLine, mh1SkelementalOracle, 6, 1)
	effectstest.D2dSetHand(opp, 3)
	effectstest.DealCombatDamage(g, me.ID, skel, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DiscardOwed(g, opp.ID); n != 2 {
		t.Errorf("the damaged player owes %d discards, want 2", n)
	}
}

func TestMh1B1LightningSkelementalIsSacrificedAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogPermanent(g, me.ID, mh1LightningSkelementalNm, mh1SkelementalTypeLine, mh1SkelementalOracle, true)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("Lightning Skelemental was not sacrificed at the end step")
	}
}

func TestMh1B1ViashinoSandsprinterReturnsAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Viashino Sandsprinter", mh1SandsprinterTypeLine, mh1SandsprinterOracle, true)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) || !me.Hand.Contains(id) {
		t.Error("Viashino Sandsprinter did not return to hand at the end step")
	}
}

func TestMh1B1MotherBearMakesTwoBearsFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	id := effectstest.SeedGraveyardCard(t, g, "Mother Bear", mh1BearType, mh1MotherBearOracle)
	effectstest.FloatMana(t, g, me, "{G}{G}{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Error("Mother Bear was not exiled")
	}
	if n := mh1Tokens(g, "Bear", me.ID); n != 2 {
		t.Errorf("%d Bear tokens, want 2", n)
	}
}

func TestMh1B1RhoxVeteranTapsACreatureWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rhox := effectstest.B12Push(g, me.ID, "Rhox Veteran", "Creature — Rhino Soldier", mh1RhoxOracle, 2, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", mh1BearType, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", mh1BearType, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, rhox, bear)
	mh1SettleAnswering(t, g, me.ID, blocker)
	if !effectstest.HmlTapped(t, g, blocker) {
		t.Error("the opponent's creature is untapped")
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("the other attacker has power %d after battle cry, want 3", p)
	}
}

func TestMh1B1RuinationRioterCountsLandCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Ruination Rioter", "Creature — Human Berserker", mh1RuinationRioterOracle, false)
	effectstest.WthGraveyardCard(me, "Forest", mh1BasicForestType)
	effectstest.WthGraveyardCard(me, "Mountain", "Basic Land — Mountain")
	before := foe.Life
	effectstest.KillCreature(t, g, me.ID, id)
	if pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: foe.ID}); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if foe.Life != before-2 {
		t.Errorf("foe life %d, want %d", foe.Life, before-2)
	}
}
