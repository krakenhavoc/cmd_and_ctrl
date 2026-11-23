package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_12_cards_test.go — Marvel Super Heroes Commander, twelfth batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	msc12InfinityFormulaOracle = "83669321-46cd-4805-b4ce-441c5db0f338"
	msc12IconicShieldOracle    = "1c4fdd52-dc99-4942-bdaa-93f74dd4ebdc"
	msc12TrickShotOracle       = "ead2aee3-4f95-4ff1-8be9-ea9619e2386e"
	msc12GiantManOracle        = "effa84e6-0aee-422f-81e9-3feeca38eda2"
	msc12GravitonOracle        = "eab391d2-ccdd-401c-8f10-69aac4619213"
	msc12LadySpiderOracle      = "59442ef7-5276-4f38-83c8-9a849388a711"
	msc12HeroesForHireOracle   = "45e5135b-274c-46b4-8249-7e05e37aaff7"
	msc12IngeniousSmithOracle  = "f3988b96-9e00-445c-b3e1-6b2df9b13dac"
	msc12IronSuitcaseOracle    = "d2ba20a0-7c08-49ed-b960-4b3154a8c025"
	msc12InspiredFireOracle    = "7f43b15b-b5d4-4ec6-bd5b-9cf56b2a6bde"
	msc12LuckyOracle           = "3a429a1e-62c0-4f01-b09e-2bbf63e56a98"
)

func TestMsc12EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		msc12InfinityFormulaOracle: "Infinity Formula",
		msc12IconicShieldOracle:    "Iconic Shield",
		msc12TrickShotOracle:       "Hawkeye, Trick Shot",
		msc12GiantManOracle:        "Giant-Man, Gargantuan Genius",
		msc12GravitonOracle:        "Graviton, Fundamental Force",
		msc12LadySpiderOracle:      "Lady Spider, Maybelle Reilly",
		msc12HeroesForHireOracle:   "Heroes for Hire",
		msc12IngeniousSmithOracle:  "Ingenious Smith",
		msc12IronSuitcaseOracle:    "Iron Suitcase",
		msc12InspiredFireOracle:    "Inspired Fire",
		msc12LuckyOracle:           "Lucky the Pizza Dog",
	}
	for oracle, name := range want {
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

// --- Infinity Formula ---

func TestMsc12InfinityFormulaAttachesOnEntryAndTheHostGainsTwoLifePerAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PowerUpMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	formula := effectstest.Msc12Cast(t, g, "Infinity Formula", "Artifact — Equipment", "{2}{W}", "{C}{C}{W}", msc12InfinityFormulaOracle, 0)

	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, formula); host.ID != bear {
		t.Fatalf("the Equipment did not attach to the Bear: %+v", host)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("equipped P/T %d/%d, want 3/4", p, tough)
	}
	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2 from the granted attack trigger", life, me.Life)
	}
}

func TestMsc12InfinityFormulaGrantsNothingToACreatureItIsNotAttachedTo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PowerUpMain(t, g)
	first := effectstest.SeedBear(g, me.ID)
	second := effectstest.SeedBear(g, me.ID)
	formula := effectstest.SeedEquipment(g, me.ID, "Infinity Formula", msc12InfinityFormulaOracle)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.EquipTo(t, g, me.ID, formula, first)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.EquipTo(t, g, me.ID, formula, second)

	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, first)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("the old host's attack gained life: %d -> %d", life, me.Life)
	}
}

// --- Iconic Shield ---

func TestMsc12IconicShieldGivesAnotherAttackerIndestructibleNotItsHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PowerUpMain(t, g)
	host := effectstest.SeedBear(g, me.ID)
	friend := effectstest.SeedBear(g, me.ID)
	shield := effectstest.SeedEquipment(g, me.ID, "Iconic Shield", msc12IconicShieldOracle)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.EquipTo(t, g, me.ID, shield, host)
	if p, tough := effectstest.EffectivePTOf(t, g, host); p != 3 || tough != 4 {
		t.Fatalf("equipped P/T %d/%d, want 3/4", p, tough)
	}

	effectstest.DeclareAttack(t, g, opp.ID, host, friend)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if effectstest.HasID(pick.PickTargetCards, host) || !effectstest.HasID(pick.PickTargetCards, friend) {
			t.Fatalf("only the OTHER attacker is a legal target: %+v", pick.PickTargetCards)
		}
		effectstest.PickCard(t, g, me.ID, friend)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, friend, "indestructible") {
		t.Error("the other attacker did not gain indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, host, "indestructible") {
		t.Error("the equipped creature gained indestructible: the text says another")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, friend, "indestructible") {
		t.Error("indestructible outlived the turn")
	}
}

func TestMsc12IconicShieldHostAttackingAloneHasNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PowerUpMain(t, g)
	host := effectstest.SeedBear(g, me.ID)
	shield := effectstest.SeedEquipment(g, me.ID, "Iconic Shield", msc12IconicShieldOracle)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.EquipTo(t, g, me.ID, shield, host)

	effectstest.DeclareAttack(t, g, opp.ID, host)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		t.Fatalf("a lone attacker found a target: %+v", pick.PickTargetCards)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, host, "indestructible") {
		t.Error("the lone attacker gained indestructible")
	}
}

// --- Hawkeye, Trick Shot ---

func TestMsc12TrickShotEveryHeroEnteringDealsDamageEqualToYourHeroes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PowerUpMain(t, g)
	effectstest.B12Push(g, me.ID, "Hawkeye, Trick Shot", "Legendary Creature — Human Archer Hero", msc12TrickShotOracle, 3, 4)
	effectstest.B12Creature(g, me.ID, "Other Hero", "Creature — Human Hero", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 2, 2)
	life := opp.Life

	effectstest.Msc12Cast(t, g, "New Hero", "Creature — Human Hero", "{1}", "{C}", "", 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.Msc12PickPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want -3 (Hawkeye, Other Hero and New Hero; theirs do not count)", life, opp.Life)
	}
}

func TestMsc12TrickShotHawkeyeHimselfAndNonHeroesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PowerUpMain(t, g)
	life := opp.Life

	// A non-Hero entering does nothing even with a Hawkeye out.
	hawkeye := effectstest.Msc12Cast(t, g, "Hawkeye, Trick Shot", "Legendary Creature — Human Archer Hero", "{3}{R}", "{C}{C}{C}{R}", msc12TrickShotOracle, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.Msc12PickPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Fatalf("Hawkeye entering alone: opponent life %d -> %d, want -1", life, opp.Life)
	}
	if !effectstest.HasEffectiveKeyword(t, g, hawkeye, "reach") || !effectstest.HasEffectiveKeyword(t, g, hawkeye, "first strike") {
		t.Error("Hawkeye lacks reach or first strike")
	}

	effectstest.Msc12Cast(t, g, "Plain Bear", "Creature — Bear", "{1}", "{C}", "", 2)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatal("a Bear entering triggered Hawkeye")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want it unchanged by a Bear", opp.Life)
	}

	// An opponent's Hero entering does not trigger it either.
	effectstest.B12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 2, 2)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatal("an opponent's Hero triggered Hawkeye")
	}
}

func TestMsc12GiantManAddsGreenForEachCreatureWithPowerFourOrMoreAtTheFirstMain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Giant-Man, Gargantuan Genius", "Legendary Creature — Human Scientist Hero", msc12GiantManOracle, 6, 6)
	effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 4, 4)
	effectstest.B12Creature(g, me.ID, "Small Bear", "Creature — Bear", 3, 3)
	effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 7, 7)

	effectstest.PowerUpMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Msc12PoolOf(me, "G"); got != 2 {
		t.Errorf("{G} in the pool = %d, want 2 (Giant-Man and the 4/4; not the 3/3 or their 7/7)", got)
	}
	if got := effectstest.Msc12PoolOf(opp, "G"); got != 0 {
		t.Errorf("the opponent got %d {G}", got)
	}
}

func TestMsc12GiantManAddsNothingOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Giant-Man, Gargantuan Genius", "Legendary Creature — Human Scientist Hero", msc12GiantManOracle, 6, 6)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.PowerUpMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Msc12PoolOf(me, "G"); got != 0 {
		t.Errorf("{G} in the pool on another player's turn = %d, want 0", got)
	}
}

// --- Graviton, Fundamental Force ---

func msc12Draw(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p.ID, 1) })
}

func TestMsc12GravitonSecondDrawOffersFlyingOrTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	effectstest.B12Push(g, me.ID, "Graviton, Fundamental Force", "Legendary Creature — Human Villain", msc12GravitonOracle, 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	msc12Draw(t, g, me)
	if c := effectstest.ModePickChoiceFor(g, me.ID); c != nil {
		t.Fatal("the first draw triggered Graviton")
	}
	msc12Draw(t, g, me)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the second draw did not offer a mode")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 0)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target creature did not gain flying")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, bear); tapped {
		t.Error("the flying mode tapped the creature")
	}
	msc12Draw(t, g, me)
	if c := effectstest.ModePickChoiceFor(g, me.ID); c != nil {
		t.Error("the third draw triggered Graviton again")
	}
}

func TestMsc12GravitonTapMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	effectstest.B12Push(g, me.ID, "Graviton, Fundamental Force", "Legendary Creature — Human Villain", msc12GravitonOracle, 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	msc12Draw(t, g, me)
	msc12Draw(t, g, me)
	effectstest.AnswerMode(t, g, effectstest.ModePickChoiceFor(g, me.ID), me.ID, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, bear); !tapped {
		t.Error("the target creature was not tapped")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the tap mode gave flying")
	}
}

func TestMsc12GravitonIgnoresAnOpponentsSecondDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, me.ID, "Graviton, Fundamental Force", "Legendary Creature — Human Villain", msc12GravitonOracle, 3, 3)
	msc12Draw(t, g, opp)
	msc12Draw(t, g, opp)
	if c := effectstest.ModePickChoiceFor(g, me.ID); c != nil {
		t.Error("an opponent's second draw triggered Graviton")
	}
}

// --- Lady Spider, Maybelle Reilly ---

func TestMsc12LadySpiderCountersASpiderForEachDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lady := effectstest.B12Push(g, me.ID, "Lady Spider, Maybelle Reilly", "Legendary Creature — Spider Human Hero", msc12LadySpiderOracle, 1, 3)
	other := effectstest.B12Creature(g, me.ID, "Other Spider", "Creature — Spider", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, lady) || !effectstest.HasID(p.PickTargetCards, other) || effectstest.HasID(p.PickTargetCards, bear) {
		t.Fatalf("only Spiders are legal targets: %+v", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, other) != 1 || effectstest.PlusOnes(g, lady) != 0 || effectstest.PlusOnes(g, bear) != 0 {
		t.Errorf("counters: other %d, Lady %d, Bear %d; want 1, 0, 0", effectstest.PlusOnes(g, other), effectstest.PlusOnes(g, lady), effectstest.PlusOnes(g, bear))
	}
}

func TestMsc12LadySpiderIgnoresAnOpponentsDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Lady Spider, Maybelle Reilly", "Legendary Creature — Spider Human Hero", msc12LadySpiderOracle, 1, 3)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Error("an opponent's discard triggered Lady Spider")
	}
}

// --- Heroes for Hire ---

func TestMsc12HeroesForHireMakesThreeTreasuresAndSacrificingOneExilesTheTopCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PowerUpMain(t, g)
	hire := effectstest.Msc12Cast(t, g, "Heroes for Hire", "Enchantment", "{3}{R}{R}", "{C}{C}{C}{R}{R}", msc12HeroesForHireOracle, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTreasuresControlledBy(g, me.ID); got != 3 {
		t.Fatalf("Treasures after entry = %d, want 3", got)
	}

	var treasure uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.Name == "Treasure" {
				treasure = c.InstanceID
				return
			}
		}
	})
	top := me.Library.Cards[me.Library.Size()-1].InstanceID
	if err := g.ActivateCatalogAbility(me.ID, hire, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{treasure}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Error("the top card was not exiled")
	}
	if got := effectstest.CountTreasuresControlledBy(g, me.ID); got != 2 {
		t.Errorf("Treasures after the sacrifice = %d, want 2", got)
	}
}

func TestMsc12HeroesForHireRefusesAnythingButATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hire := effectstest.B12Push(g, me.ID, "Heroes for Hire", "Enchantment", msc12HeroesForHireOracle, 0, 0)
	bear := effectstest.SeedBear(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, hire, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("a Bear was accepted for 'Sacrifice a Treasure'")
	}
}

// --- Ingenious Smith ---

func TestMsc12IngeniousSmithTakesAnArtifactFromTheTopFourAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PowerUpMain(t, g)
	ids := effectstest.SamiLibraryTopFirst(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Rock", TypeLine: "Artifact"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Fourth", TypeLine: "Instant"},
		game.Card{Name: "Fifth", TypeLine: "Artifact"},
	)
	lib := me.Library.Size()
	effectstest.Msc12Cast(t, g, "Ingenious Smith", "Creature — Human Artificer", "{1}{W}", "{C}{W}", msc12IngeniousSmithOracle, 1)
	effectstest.PassPriorityAroundTable(t, g)
	// Only artifact cards among the top four are offered: Rock, not the fifth.
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	if !me.Hand.Contains(ids[1]) {
		t.Fatal("the artifact did not reach the hand")
	}
	if me.Library.Size() != lib-1 {
		t.Errorf("library %d -> %d, want one card taken", lib, me.Library.Size())
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[4] {
		t.Error("the fifth card must now be on top; the other three went to the bottom")
	}
}

func TestMsc12IngeniousSmithGrowsOncePerTurnWhenArtifactsEnter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PowerUpMain(t, g)
	smith := effectstest.B12Push(g, me.ID, "Ingenious Smith", "Creature — Human Artificer", msc12IngeniousSmithOracle, 1, 1)

	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, smith) != 0 {
		t.Fatal("an opponent's artifact grew Smith")
	}
	effectstest.Msc12Cast(t, g, "Rock", "Artifact", "{1}", "{C}", "", 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, smith) != 1 {
		t.Fatalf("counters after the first artifact = %d, want 1", effectstest.PlusOnes(g, smith))
	}
	effectstest.Msc12Cast(t, g, "Second Rock", "Artifact", "{1}", "{C}", "", 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, smith) != 1 {
		t.Errorf("counters after a second artifact = %d, want it to stay 1 this turn", effectstest.PlusOnes(g, smith))
	}
}

// --- Iron Suitcase ---

func TestMsc12IronSuitcaseScriesTwoAndAnimatesForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PowerUpMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	suitcase := effectstest.Msc12Cast(t, g, "Iron Suitcase", "Artifact", "{1}", "{C}", msc12IronSuitcaseOracle, 0)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("entering did not offer a scry")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scry looks at %d cards, want 2", len(c.ScryCards))
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)

	if effectstest.ContainsString(effectstest.EffectiveTypes(t, g, suitcase), "Creature") {
		t.Fatal("the Suitcase is a creature before it is animated")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, suitcase, 0, "{C}{C}", nil); err == nil {
		t.Fatal("animated for {2}")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, suitcase, 0, "{C}{C}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ContainsString(effectstest.EffectiveTypes(t, g, suitcase), "Creature") {
		t.Fatal("the Suitcase did not become a creature")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, suitcase); p != 3 || tough != 3 {
		t.Errorf("P/T %d/%d, want 3/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, suitcase, "flying") {
		t.Error("no flying")
	}
	types := effectstest.EffectiveSubtypes(t, g, suitcase)
	if !effectstest.ContainsString(types, "Construct") {
		t.Errorf("subtypes %v lack Construct", types)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.ContainsString(effectstest.EffectiveTypes(t, g, suitcase), "Creature") {
		t.Error("the Suitcase stayed a creature past the turn")
	}
}

// --- Inspired Fire ---

func TestMsc12InspiredFireCostsOneLessForEachCardDrawnThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fire := game.Card{InstanceID: uuid.New(), Name: "Inspired Fire", TypeLine: "Instant",
		ManaCost: "{3}{R}", OracleID: msc12InspiredFireOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(fire)
	price := func() int {
		t.Helper()
		p, err := g.PriceCast(me.ID, fire, game.CastSpellParams{})
		if err != nil {
			t.Fatalf("PriceCast: %v", err)
		}
		return p.Total.ManaValue()
	}
	tally := g.TurnTallyFor(me.ID).CardsDrawn
	if got := price(); got != 4-tally {
		t.Fatalf("with %d cards already drawn this turn: mana value %d, want %d", tally, got, 4-tally)
	}
	for drawn := tally; drawn < 5; {
		msc12Draw(t, g, me)
		drawn++
		want := 4 - drawn
		if want < 1 {
			want = 1 // generic mana only: {R} always remains
		}
		if got := price(); got != want {
			t.Fatalf("after %d draws: mana value %d, want %d", drawn, got, want)
		}
	}
}

func TestMsc12InspiredFireDealsFourDamageToTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	small := effectstest.B12Creature(g, opp.ID, "Four-Toughness", "Creature — Bear", 2, 4)
	big := effectstest.B12Creature(g, opp.ID, "Five-Toughness", "Creature — Bear", 2, 5)
	effectstest.CastCatalogSpell(t, g, "Inspired Fire", "Instant", msc12InspiredFireOracle, effectstest.CardRefs(small))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(small) == false {
		t.Error("4 damage did not kill the 4-toughness creature")
	}
	effectstest.CastCatalogSpell(t, g, "Inspired Fire", "Instant", msc12InspiredFireOracle, effectstest.CardRefs(big))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(big) {
		t.Error("4 damage killed a 5-toughness creature")
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Inspired Fire", "Instant", msc12InspiredFireOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Error("Inspired Fire accepted a player as its target")
	}
}

// --- Lucky the Pizza Dog ---

func TestMsc12LuckyMakesFoodForCatDogAndHeroSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PowerUpMain(t, g)
	effectstest.B12Push(g, me.ID, "Lucky the Pizza Dog", "Legendary Creature — Dog", msc12LuckyOracle, 2, 2)
	foods := func() int { return effectstest.BattlefieldNamed(g, me.ID, "Food") }

	for _, tc := range []struct {
		name, typeLine string
		want           int
	}{
		{"Bear Cub", "Creature — Bear", 0},
		{"Alley Cat", "Creature — Cat", 1},
		{"Stray Dog", "Creature — Dog", 1},
		{"Hero Spell", "Creature — Human Hero", 1},
	} {
		before := foods()
		effectstest.Msc12CastAndSettle(t, g, tc.name, tc.typeLine)
		if got := foods() - before; got != tc.want {
			t.Errorf("casting %s (%s): %d Food, want %d", tc.name, tc.typeLine, got, tc.want)
		}
	}
}

func TestMsc12LuckyGrowsAtEachEndStepOnlyIfYouGainedLifeThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PowerUpMain(t, g)
	lucky := effectstest.B12Push(g, me.ID, "Lucky the Pizza Dog", "Legendary Creature — Dog", msc12LuckyOracle, 2, 2)

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, lucky) != 0 {
		t.Fatalf("grew with no life gained: %d", effectstest.PlusOnes(g, lucky))
	}

	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.PowerUpPassTo(t, g, 0)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, lucky) != 1 {
		t.Errorf("counters after gaining life = %d, want 1", effectstest.PlusOnes(g, lucky))
	}
}
