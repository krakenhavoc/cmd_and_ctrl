package effects

import (
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// msc12Cast casts a permanent from the active player's hand and resolves it,
// leaving any entry trigger's prompt open.
func msc12Cast(t *testing.T, g *game.Game, name, typeLine, cost, pay, oracle string, pt int) uuid.UUID {
	t.Helper()
	return castPowerUpCreature(t, g, g.Seats[g.Turn.ActiveSeat], name, typeLine, cost, pay, oracle, pt)
}

// msc12CastAndSettle casts a 1/1 creature spell for {1} and lets the stack
// empty, so a cast trigger and the spell both resolve.
func msc12CastAndSettle(t *testing.T, g *game.Game, name, typeLine string) {
	t.Helper()
	p := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{1}",
			Owner: p.ID, Controller: p.ID, Power: 1, Toughness: 1})
	})
	floatMana(t, g, p, "{C}")
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatalf("%s did not resolve onto the battlefield", name)
	}
}

// msc12PickPlayer answers the open pick_target prompt with a player.
func msc12PickPlayer(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// --- Infinity Formula ---

func TestMsc12InfinityFormulaAttachesOnEntryAndTheHostGainsTwoLifePerAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	bear := seedBear(g, me.ID)
	formula := msc12Cast(t, g, "Infinity Formula", "Artifact — Equipment", "{2}{W}", "{C}{C}{W}", msc12InfinityFormulaOracle, 0)

	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, formula); host.ID != bear {
		t.Fatalf("the Equipment did not attach to the Bear: %+v", host)
	}
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("equipped P/T %d/%d, want 3/4", p, tough)
	}
	life := me.Life
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2 from the granted attack trigger", life, me.Life)
	}
}

func TestMsc12InfinityFormulaGrantsNothingToACreatureItIsNotAttachedTo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	first := seedBear(g, me.ID)
	second := seedBear(g, me.ID)
	formula := seedEquipment(g, me.ID, "Infinity Formula", msc12InfinityFormulaOracle)
	floatMana(t, g, me, "{C}{C}")
	equipTo(t, g, me.ID, formula, first)
	floatMana(t, g, me, "{C}{C}")
	equipTo(t, g, me.ID, formula, second)

	life := me.Life
	declareAttack(t, g, opp.ID, first)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("the old host's attack gained life: %d -> %d", life, me.Life)
	}
}

// --- Iconic Shield ---

func TestMsc12IconicShieldGivesAnotherAttackerIndestructibleNotItsHost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	host := seedBear(g, me.ID)
	friend := seedBear(g, me.ID)
	shield := seedEquipment(g, me.ID, "Iconic Shield", msc12IconicShieldOracle)
	floatMana(t, g, me, "{C}{C}{C}")
	equipTo(t, g, me.ID, shield, host)
	if p, tough := effectivePTOf(t, g, host); p != 3 || tough != 4 {
		t.Fatalf("equipped P/T %d/%d, want 3/4", p, tough)
	}

	declareAttack(t, g, opp.ID, host, friend)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		if hasID(pick.PickTargetCards, host) || !hasID(pick.PickTargetCards, friend) {
			t.Fatalf("only the OTHER attacker is a legal target: %+v", pick.PickTargetCards)
		}
		pickCard(t, g, me.ID, friend)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, friend, "indestructible") {
		t.Error("the other attacker did not gain indestructible")
	}
	if hasEffectiveKeyword(t, g, host, "indestructible") {
		t.Error("the equipped creature gained indestructible: the text says another")
	}
	advanceToNextSeatsTurn(t, g)
	if hasEffectiveKeyword(t, g, friend, "indestructible") {
		t.Error("indestructible outlived the turn")
	}
}

func TestMsc12IconicShieldHostAttackingAloneHasNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	host := seedBear(g, me.ID)
	shield := seedEquipment(g, me.ID, "Iconic Shield", msc12IconicShieldOracle)
	floatMana(t, g, me, "{C}{C}{C}")
	equipTo(t, g, me.ID, shield, host)

	declareAttack(t, g, opp.ID, host)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		t.Fatalf("a lone attacker found a target: %+v", pick.PickTargetCards)
	}
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, host, "indestructible") {
		t.Error("the lone attacker gained indestructible")
	}
}

// --- Hawkeye, Trick Shot ---

func TestMsc12TrickShotEveryHeroEnteringDealsDamageEqualToYourHeroes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	b12Push(g, me.ID, "Hawkeye, Trick Shot", "Legendary Creature — Human Archer Hero", msc12TrickShotOracle, 3, 4)
	b12Creature(g, me.ID, "Other Hero", "Creature — Human Hero", 2, 2)
	b12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 2, 2)
	life := opp.Life

	msc12Cast(t, g, "New Hero", "Creature — Human Hero", "{1}", "{C}", "", 2)
	b04WaitForPick(t, g, me.ID)
	msc12PickPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want -3 (Hawkeye, Other Hero and New Hero; theirs do not count)", life, opp.Life)
	}
}

func TestMsc12TrickShotHawkeyeHimselfAndNonHeroesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	life := opp.Life

	// A non-Hero entering does nothing even with a Hawkeye out.
	hawkeye := msc12Cast(t, g, "Hawkeye, Trick Shot", "Legendary Creature — Human Archer Hero", "{3}{R}", "{C}{C}{C}{R}", msc12TrickShotOracle, 3)
	b04WaitForPick(t, g, me.ID)
	msc12PickPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Fatalf("Hawkeye entering alone: opponent life %d -> %d, want -1", life, opp.Life)
	}
	if !hasEffectiveKeyword(t, g, hawkeye, "reach") || !hasEffectiveKeyword(t, g, hawkeye, "first strike") {
		t.Error("Hawkeye lacks reach or first strike")
	}

	msc12Cast(t, g, "Plain Bear", "Creature — Bear", "{1}", "{C}", "", 2)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("a Bear entering triggered Hawkeye")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want it unchanged by a Bear", opp.Life)
	}

	// An opponent's Hero entering does not trigger it either.
	b12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 2, 2)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("an opponent's Hero triggered Hawkeye")
	}
}

// --- Giant-Man, Gargantuan Genius ---

func msc12PoolOf(p *game.Player, color string) int {
	n := 0
	for _, tok := range p.ManaPool.Units() {
		if tok.Color == color {
			n++
		}
	}
	return n
}

func TestMsc12GiantManAddsGreenForEachCreatureWithPowerFourOrMoreAtTheFirstMain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Giant-Man, Gargantuan Genius", "Legendary Creature — Human Scientist Hero", msc12GiantManOracle, 6, 6)
	b12Creature(g, me.ID, "Big Bear", "Creature — Bear", 4, 4)
	b12Creature(g, me.ID, "Small Bear", "Creature — Bear", 3, 3)
	b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 7, 7)

	powerUpMain(t, g)
	passPriorityAroundTable(t, g)
	if got := msc12PoolOf(me, "G"); got != 2 {
		t.Errorf("{G} in the pool = %d, want 2 (Giant-Man and the 4/4; not the 3/3 or their 7/7)", got)
	}
	if got := msc12PoolOf(opp, "G"); got != 0 {
		t.Errorf("the opponent got %d {G}", got)
	}
}

func TestMsc12GiantManAddsNothingOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Giant-Man, Gargantuan Genius", "Legendary Creature — Human Scientist Hero", msc12GiantManOracle, 6, 6)
	advanceToNextSeatsTurn(t, g)
	powerUpMain(t, g)
	passPriorityAroundTable(t, g)
	if got := msc12PoolOf(me, "G"); got != 0 {
		t.Errorf("{G} in the pool on another player's turn = %d, want 0", got)
	}
}

// --- Graviton, Fundamental Force ---

func msc12Draw(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p.ID, 1) })
}

func TestMsc12GravitonSecondDrawOffersFlyingOrTap(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	b12Push(g, me.ID, "Graviton, Fundamental Force", "Legendary Creature — Human Villain", msc12GravitonOracle, 3, 3)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	msc12Draw(t, g, me)
	if c := modePickChoiceFor(g, me.ID); c != nil {
		t.Fatal("the first draw triggered Graviton")
	}
	msc12Draw(t, g, me)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the second draw did not offer a mode")
	}
	answerMode(t, g, c, me.ID, 0)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target creature did not gain flying")
	}
	if tapped, _ := battlefieldCardTapped(g, bear); tapped {
		t.Error("the flying mode tapped the creature")
	}
	msc12Draw(t, g, me)
	if c := modePickChoiceFor(g, me.ID); c != nil {
		t.Error("the third draw triggered Graviton again")
	}
}

func TestMsc12GravitonTapMode(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	b12Push(g, me.ID, "Graviton, Fundamental Force", "Legendary Creature — Human Villain", msc12GravitonOracle, 3, 3)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	msc12Draw(t, g, me)
	msc12Draw(t, g, me)
	answerMode(t, g, modePickChoiceFor(g, me.ID), me.ID, 1)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if tapped, _ := battlefieldCardTapped(g, bear); !tapped {
		t.Error("the target creature was not tapped")
	}
	if hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the tap mode gave flying")
	}
}

func TestMsc12GravitonIgnoresAnOpponentsSecondDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	b12Push(g, me.ID, "Graviton, Fundamental Force", "Legendary Creature — Human Villain", msc12GravitonOracle, 3, 3)
	msc12Draw(t, g, opp)
	msc12Draw(t, g, opp)
	if c := modePickChoiceFor(g, me.ID); c != nil {
		t.Error("an opponent's second draw triggered Graviton")
	}
}

// --- Lady Spider, Maybelle Reilly ---

func TestMsc12LadySpiderCountersASpiderForEachDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lady := b12Push(g, me.ID, "Lady Spider, Maybelle Reilly", "Legendary Creature — Spider Human Hero", msc12LadySpiderOracle, 1, 3)
	other := b12Creature(g, me.ID, "Other Spider", "Creature — Spider", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, lady) || !hasID(p.PickTargetCards, other) || hasID(p.PickTargetCards, bear) {
		t.Fatalf("only Spiders are legal targets: %+v", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if plusOnes(g, other) != 1 || plusOnes(g, lady) != 0 || plusOnes(g, bear) != 0 {
		t.Errorf("counters: other %d, Lady %d, Bear %d; want 1, 0, 0", plusOnes(g, other), plusOnes(g, lady), plusOnes(g, bear))
	}
}

func TestMsc12LadySpiderIgnoresAnOpponentsDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Lady Spider, Maybelle Reilly", "Legendary Creature — Spider Human Hero", msc12LadySpiderOracle, 1, 3)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Error("an opponent's discard triggered Lady Spider")
	}
}

// --- Heroes for Hire ---

func TestMsc12HeroesForHireMakesThreeTreasuresAndSacrificingOneExilesTheTopCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	hire := msc12Cast(t, g, "Heroes for Hire", "Enchantment", "{3}{R}{R}", "{C}{C}{C}{R}{R}", msc12HeroesForHireOracle, 0)
	passPriorityAroundTable(t, g)
	if got := countTreasuresControlledBy(g, me.ID); got != 3 {
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
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Error("the top card was not exiled")
	}
	if got := countTreasuresControlledBy(g, me.ID); got != 2 {
		t.Errorf("Treasures after the sacrifice = %d, want 2", got)
	}
}

func TestMsc12HeroesForHireRefusesAnythingButATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hire := b12Push(g, me.ID, "Heroes for Hire", "Enchantment", msc12HeroesForHireOracle, 0, 0)
	bear := seedBear(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, hire, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("a Bear was accepted for 'Sacrifice a Treasure'")
	}
}

// --- Ingenious Smith ---

func TestMsc12IngeniousSmithTakesAnArtifactFromTheTopFourAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	ids := samiLibraryTopFirst(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Rock", TypeLine: "Artifact"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Fourth", TypeLine: "Instant"},
		game.Card{Name: "Fifth", TypeLine: "Artifact"},
	)
	lib := me.Library.Size()
	msc12Cast(t, g, "Ingenious Smith", "Creature — Human Artificer", "{1}{W}", "{C}{W}", msc12IngeniousSmithOracle, 1)
	passPriorityAroundTable(t, g)
	// Only artifact cards among the top four are offered: Rock, not the fifth.
	answerChooseCards(t, g, me.ID, ids[1])
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	smith := b12Push(g, me.ID, "Ingenious Smith", "Creature — Human Artificer", msc12IngeniousSmithOracle, 1, 1)

	b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	passPriorityAroundTable(t, g)
	if plusOnes(g, smith) != 0 {
		t.Fatal("an opponent's artifact grew Smith")
	}
	msc12Cast(t, g, "Rock", "Artifact", "{1}", "{C}", "", 0)
	passPriorityAroundTable(t, g)
	if plusOnes(g, smith) != 1 {
		t.Fatalf("counters after the first artifact = %d, want 1", plusOnes(g, smith))
	}
	msc12Cast(t, g, "Second Rock", "Artifact", "{1}", "{C}", "", 0)
	passPriorityAroundTable(t, g)
	if plusOnes(g, smith) != 1 {
		t.Errorf("counters after a second artifact = %d, want it to stay 1 this turn", plusOnes(g, smith))
	}
}

// --- Iron Suitcase ---

func TestMsc12IronSuitcaseScriesTwoAndAnimatesForThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	seedLibrary(me, "A", "B", "C")
	suitcase := msc12Cast(t, g, "Iron Suitcase", "Artifact", "{1}", "{C}", msc12IronSuitcaseOracle, 0)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("entering did not offer a scry")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scry looks at %d cards, want 2", len(c.ScryCards))
	}
	answerScryKeepAll(t, g, me.ID)

	if containsString(effectiveTypes(t, g, suitcase), "Creature") {
		t.Fatal("the Suitcase is a creature before it is animated")
	}
	if err := hobB4Activate(t, g, me.ID, suitcase, 0, "{C}{C}", nil); err == nil {
		t.Fatal("animated for {2}")
	}
	if err := hobB4Activate(t, g, me.ID, suitcase, 0, "{C}{C}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !containsString(effectiveTypes(t, g, suitcase), "Creature") {
		t.Fatal("the Suitcase did not become a creature")
	}
	if p, tough := effectivePTOf(t, g, suitcase); p != 3 || tough != 3 {
		t.Errorf("P/T %d/%d, want 3/3", p, tough)
	}
	if !hasEffectiveKeyword(t, g, suitcase, "flying") {
		t.Error("no flying")
	}
	types := effectiveSubtypes(t, g, suitcase)
	if !containsString(types, "Construct") {
		t.Errorf("subtypes %v lack Construct", types)
	}
	advanceToNextSeatsTurn(t, g)
	if containsString(effectiveTypes(t, g, suitcase), "Creature") {
		t.Error("the Suitcase stayed a creature past the turn")
	}
}

// --- Inspired Fire ---

func TestMsc12InspiredFireCostsOneLessForEachCardDrawnThisTurn(t *testing.T) {
	g := newCatalogGame(t)
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
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := b12Creature(g, opp.ID, "Four-Toughness", "Creature — Bear", 2, 4)
	big := b12Creature(g, opp.ID, "Five-Toughness", "Creature — Bear", 2, 5)
	castCatalogSpell(t, g, "Inspired Fire", "Instant", msc12InspiredFireOracle, cardRefs(small))
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(small) == false {
		t.Error("4 damage did not kill the 4-toughness creature")
	}
	castCatalogSpell(t, g, "Inspired Fire", "Instant", msc12InspiredFireOracle, cardRefs(big))
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(big) {
		t.Error("4 damage killed a 5-toughness creature")
	}
	if err := castCatalogSpellErr(t, g, "Inspired Fire", "Instant", msc12InspiredFireOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Error("Inspired Fire accepted a player as its target")
	}
}

// --- Lucky the Pizza Dog ---

func TestMsc12LuckyMakesFoodForCatDogAndHeroSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	b12Push(g, me.ID, "Lucky the Pizza Dog", "Legendary Creature — Dog", msc12LuckyOracle, 2, 2)
	foods := func() int { return battlefieldNamed(g, me.ID, "Food") }

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
		msc12CastAndSettle(t, g, tc.name, tc.typeLine)
		if got := foods() - before; got != tc.want {
			t.Errorf("casting %s (%s): %d Food, want %d", tc.name, tc.typeLine, got, tc.want)
		}
	}
}

func TestMsc12LuckyGrowsAtEachEndStepOnlyIfYouGainedLifeThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	lucky := b12Push(g, me.ID, "Lucky the Pizza Dog", "Legendary Creature — Dog", msc12LuckyOracle, 2, 2)

	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if plusOnes(g, lucky) != 0 {
		t.Fatalf("grew with no life gained: %d", plusOnes(g, lucky))
	}

	advanceToNextSeatsTurn(t, g)
	powerUpPassTo(t, g, 0)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if plusOnes(g, lucky) != 1 {
		t.Errorf("counters after gaining life = %d, want 1", plusOnes(g, lucky))
	}
}
