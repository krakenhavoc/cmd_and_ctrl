package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tdm_r_cards_test.go — Tarkir: Dragonstorm, cards from R (#4463). Each
// test plays the real catalog entry and asserts what its text says.

const (
	tdmRainveilOracle   = "6ccf72ee-a86a-4b13-9e76-e308d0d28523"
	tdmBargainOracle    = "d87be8c6-7a1f-4db8-ada1-b1a030f03877"
	tdmRallyOracle      = "3f6101a5-3038-45b0-8302-b2e4127b9510"
	tdmRebelliousOracle = "4f3d7102-f8ce-42d4-94eb-f69b695773d1"
	tdmReigningOracle   = "b8c5e481-371c-4999-bb2c-347a2aa33171"
	tdmMerchantOracle   = "62f09a5b-d7fb-44f4-9fab-719185b735ab"
	tdmRescueOracle     = "d70ad04a-b940-4bb1-bcdc-408166e91723"
	tdmRevivalOracle    = "5a848c1f-d11a-42fb-b9ce-663901fc9555"
	tdmRingingOracle    = "334ffdc4-9fe6-4a7f-bf5d-bcd88adab715"
	tdmRiverwalkOracle  = "a123aab8-3f3a-4187-abcb-21a99dcf13bb"
	tdmRoarOracle       = "a3570093-3c16-42e7-a83d-2ff611a891b5"
	tdmRoilingOracle    = "6badc583-fabb-4d63-b70c-e7380acafc23"
	tdmRotCurseOracle   = "9fee2572-8864-475c-bade-a64e36ef69c4"
	tdmRSagaTypeLine    = "Enchantment — Saga"
	tdmRBearTypeLine    = "Creature — Bear"
)

func TestTdmREveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tdmRainveilOracle:   "Rainveil Rejuvenator",
		tdmBargainOracle:    "Rakshasa's Bargain",
		tdmRallyOracle:      "Rally the Monastery",
		tdmRebelliousOracle: "Rebellious Strike",
		tdmReigningOracle:   "Reigning Victor",
		tdmMerchantOracle:   "Reputable Merchant",
		tdmRescueOracle:     "Rescue Leopard",
		tdmRevivalOracle:    "Revival of the Ancestors",
		tdmRingingOracle:    "Ringing Strike Mastery",
		tdmRiverwalkOracle:  "Riverwalk Technique",
		tdmRoarOracle:       "Roar of Endless Song",
		tdmRoilingOracle:    "Roiling Dragonstorm",
		tdmRotCurseOracle:   "Rot-Curse Rakshasa",
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

// --- fixtures ------------------------------------------------------------------

// tdmRSeats is the active seat and the next one.
func tdmRSeats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// tdmRPickTarget answers the newest pick_target prompt for `p` with one
// card and fails the test if there is none or it is refused.
func tdmRPickTarget(t *testing.T, g *game.Game, p uuid.UUID, id uuid.UUID) {
	t.Helper()
	pick := effectstest.LatestPickTarget(g, p)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pick.ID, p, game.TargetRef{Kind: game.TargetCard, ID: id}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// --- Rainveil Rejuvenator ------------------------------------------------------

func TestTdmRainveilRejuvenatorMayMillThreeOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := tdmRSeats(g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	gy, lib := me.Graveyard.Size(), me.Library.Size()
	tdmSTCast(t, g, "Rainveil Rejuvenator", "Creature — Elephant Druid", tdmRainveilOracle, 2, 4)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size() - gy; got != 3 {
		t.Errorf("milled %d, want 3", got)
	}
	if got := lib - me.Library.Size(); got != 3 {
		t.Errorf("library lost %d, want 3", got)
	}
}

func TestTdmRainveilRejuvenatorTapsForGreenEqualToItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := tdmRSeats(g)
	elephant := effectstest.B12Push(g, me.ID, "Rainveil Rejuvenator", "Creature — Elephant Druid", tdmRainveilOracle, 2, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == elephant {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 1}
			}
		}
	})
	if err := g.ActivateManaAbility(me.ID, elephant, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	got := effectstest.Batch01PoolColors(me)
	if len(got) != 3 {
		t.Fatalf("pool %v, want three {G} for power 3", got)
	}
	for _, c := range got {
		if c != "G" {
			t.Errorf("pool %v, want only {G}", got)
		}
	}
}

// --- Rakshasa's Bargain --------------------------------------------------------

func TestTdmRakshasasBargainTakesTwoAndBinsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := tdmRSeats(g)
	ids := effectstest.SamiLibraryTopFirst(me,
		game.Card{Name: "One", TypeLine: "Instant"},
		game.Card{Name: "Two", TypeLine: "Sorcery"},
		game.Card{Name: "Three", TypeLine: "Creature — Bear"},
		game.Card{Name: "Four", TypeLine: "Artifact"},
		game.Card{Name: "Fifth", TypeLine: "Instant"},
	)
	effectstest.CastCatalogSpell(t, g, "Rakshasa's Bargain", "Instant", tdmBargainOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no look-and-take prompt")
	}
	if c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Errorf("prompt asks for %d to %d cards, want exactly two", c.ChooseMin, c.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0], ids[2])
	for _, id := range []uuid.UUID{ids[0], ids[2]} {
		if !me.Hand.Contains(id) {
			t.Error("a chosen card did not reach the hand")
		}
	}
	for _, id := range []uuid.UUID{ids[1], ids[3]} {
		if !me.Graveyard.Contains(id) {
			t.Error("an unchosen card did not go to the graveyard")
		}
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[4] {
		t.Error("the fifth card must now be on top")
	}
}

// --- Rally the Monastery -------------------------------------------------------

func TestTdmRallyTheMonasteryCostsTwoLessAfterAnotherSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := tdmRSeats(g)
	effectstest.ToMainForCost(t, g)
	if got := effectstest.SelfPricedMV(t, g, me, tdmRallyOracle, "Instant", "{3}{W}"); got != 4 {
		t.Errorf("Rally the Monastery with no spell cast: %d, want 4", got)
	}
	effectstest.TdmMPCastFreeSpell(t, g, me, "First")
	if got := effectstest.SelfPricedMV(t, g, me, tdmRallyOracle, "Instant", "{3}{W}"); got != 2 {
		t.Errorf("Rally the Monastery after another spell: %d, want 2", got)
	}
}

func TestTdmRallyTheMonasteryMakesTwoMonks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := tdmRSeats(g)
	effectstest.CastModal(t, g, "Rally the Monastery", "Instant", tdmRallyOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	monks := effectstest.SocB8Named(g, me.ID, "Monk")
	if len(monks) != 2 {
		t.Fatalf("%d Monks, want 2", len(monks))
	}
	for _, m := range monks {
		if m.Power != 1 || m.Toughness != 1 || len(m.Colors) != 1 || m.Colors[0] != "W" || !effectstest.ContainsString(m.Keywords, "prowess") {
			t.Errorf("Monk %d/%d %v %v, want a 1/1 white Monk with prowess", m.Power, m.Toughness, m.Colors, m.Keywords)
		}
	}
}

func TestTdmRallyTheMonasteryPumpsUpToTwoOfYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := tdmRSeats(g)
	a := effectstest.B12Creature(g, me.ID, "A", tdmRBearTypeLine, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", tdmRBearTypeLine, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", tdmRBearTypeLine, 2, 2)
	if _, err := effectstest.CastModalErr(t, g, "Rally the Monastery", "Instant", tdmRallyOracle, []int{1}, effectstest.CardRefs(theirs)); err == nil {
		t.Error("an opponent's creature was accepted for the pump")
	}
	effectstest.CastModal(t, g, "Rally the Monastery", "Instant", tdmRallyOracle, []int{1}, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, a); p != 4 || tough != 4 {
		t.Errorf("A is %d/%d, want 4/4", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, b); p != 3 || tough != 3 {
		t.Errorf("B is %d/%d, want 3/3", p, tough)
	}
}

func TestTdmRallyTheMonasteryDestroysOnlyPowerFourOrGreater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := tdmRSeats(g)
	small := effectstest.B12Creature(g, opp.ID, "Small", tdmRBearTypeLine, 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 4)
	if _, err := effectstest.CastModalErr(t, g, "Rally the Monastery", "Instant", tdmRallyOracle, []int{2}, effectstest.CardRefs(small)); err == nil {
		t.Error("a power 3 creature was accepted for the destroy")
	}
	effectstest.CastModal(t, g, "Rally the Monastery", "Instant", tdmRallyOracle, []int{2}, effectstest.CardRefs(big))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the power 4 creature survived")
	}
}

// --- Rebellious Strike ---------------------------------------------------------

func TestTdmRebelliousStrikePumpsThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := tdmRSeats(g)
	effectstest.SeedLibrary(me, "A", "B")
	bear := effectstest.B12Creature(g, me.ID, "Bear", tdmRBearTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Rebellious Strike", "Instant", tdmRebelliousOracle, effectstest.CardRefs(bear))
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 5 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 5/2", p, tough)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one drawn", hand, me.Hand.Size())
	}
}

// --- Reigning Victor -----------------------------------------------------------

func TestTdmReigningVictorProtectsATargetOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := tdmRSeats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", tdmRBearTypeLine, 2, 2)
	tdmSTCast(t, g, "Reigning Victor", "Creature — Orc Warrior", tdmReigningOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	tdmRPickTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 3/2", p, tough)
	}
	if !effectstest.HasKeywordOf(g, bear, "indestructible") {
		t.Error("the Bear did not gain indestructible")
	}
}

func TestTdmReigningVictorMobilizesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := tdmRSeats(g)
	victor := effectstest.B12Push(g, me.ID, "Reigning Victor", "Creature — Orc Warrior", tdmReigningOracle, 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, victor)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TdmPlaceWarriors(t, g, me.ID)
	warriors := effectstest.SocB8Named(g, me.ID, "Warrior")
	if len(warriors) != 1 {
		t.Fatalf("mobilize 1 made %d Warriors, want 1", len(warriors))
	}
	if !warriors[0].Tapped || warriors[0].AttackingTarget == uuid.Nil {
		t.Error("the Warrior is not tapped and attacking")
	}
}

// --- Reputable Merchant --------------------------------------------------------

func TestTdmReputableMerchantCountersOnEntryAndDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := tdmRSeats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", tdmRBearTypeLine, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", tdmRBearTypeLine, 2, 2)
	merchant := tdmSTCast(t, g, "Reputable Merchant", "Creature — Human Citizen", tdmMerchantOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger did not ask for a target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("an opponent's creature was accepted")
	}
	tdmRPickTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Fatalf("the Bear has %d +1/+1 counters after the entry, want 1", got)
	}
	effectstest.B18Kill(t, g, merchant)
	effectstest.PassPriorityAroundTable(t, g)
	tdmRPickTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the Bear has %d +1/+1 counters after the death, want 2", got)
	}
}

// --- Rescue Leopard ------------------------------------------------------------

func TestTdmRescueLeopardRummagesWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := tdmRSeats(g)
	effectstest.SeedLibrary(me, "A", "B")
	fodder := effectstest.HandCard(me, "Fodder", "Instant")
	leopard := effectstest.B12Push(g, me.ID, "Rescue Leopard", "Creature — Cat", tdmRescueOracle, 4, 2)
	effectstest.DeclareAttack(t, g, opp.ID, leopard)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt when the Leopard became tapped")
	}
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, fodder)
	if !me.Graveyard.Contains(fodder) {
		t.Error("the discarded card is not in the graveyard")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want one discarded and one drawn", hand, me.Hand.Size())
	}
}

func TestTdmRescueLeopardDrawsNothingWithoutADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := tdmRSeats(g)
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.HandCard(me, "Kept", "Instant")
	leopard := effectstest.B12Push(g, me.ID, "Rescue Leopard", "Creature — Cat", tdmRescueOracle, 4, 2)
	effectstest.DeclareAttack(t, g, opp.ID, leopard)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no change when nothing is discarded", hand, me.Hand.Size())
	}
}

// --- Revival of the Ancestors --------------------------------------------------

func TestTdmRevivalOfTheAncestorsChapterOneMakesThreeSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := tdmRSeats(g)
	effectstest.FinS3CastSaga(t, g, "Revival of the Ancestors", tdmRSagaTypeLine, tdmRevivalOracle, "", 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	spirits := effectstest.SocB8Named(g, me.ID, "Spirit")
	if len(spirits) != 3 {
		t.Fatalf("%d Spirits, want 3", len(spirits))
	}
	for _, s := range spirits {
		if s.Power != 1 || s.Toughness != 1 || len(s.Colors) != 1 || s.Colors[0] != "W" || !effects.IsToken(s) || len(s.Keywords) != 0 {
			t.Errorf("Spirit %d/%d %v %v, want a plain 1/1 white Spirit token", s.Power, s.Toughness, s.Colors, s.Keywords)
		}
	}
}

func TestTdmRevivalOfTheAncestorsChapterTwoDistributesThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	a := effectstest.B12Creature(g, me.ID, "A", tdmRBearTypeLine, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", tdmRBearTypeLine, 2, 2)
	effectstest.FinS3Saga(g, me.ID, "Revival of the Ancestors", tdmRSagaTypeLine, tdmRevivalOracle, 0, 0, 1)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{a: 2, b: 1}, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, a, game.CounterPlusOne); got != 2 {
		t.Errorf("A has %d +1/+1 counters, want 2", got)
	}
	if got := effectstest.CountersOn(g, b, game.CounterPlusOne); got != 1 {
		t.Errorf("B has %d +1/+1 counters, want 1", got)
	}
}

func TestTdmRevivalOfTheAncestorsChapterThreeGrantsTrampleAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", tdmRBearTypeLine, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", tdmRBearTypeLine, 2, 2)
	saga := effectstest.FinS3Saga(g, me.ID, "Revival of the Ancestors", tdmRSagaTypeLine, tdmRevivalOracle, 0, 0, 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOf(g, bear, "trample") || !effectstest.HasKeywordOf(g, bear, "lifelink") {
		t.Error("your creature did not gain trample and lifelink")
	}
	if effectstest.HasKeywordOf(g, theirs, "trample") || effectstest.HasKeywordOf(g, theirs, "lifelink") {
		t.Error("an opponent's creature gained the keywords")
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after its last chapter")
	}
}

// --- Ringing Strike Mastery ----------------------------------------------------

func TestTdmRingingStrikeMasteryTapsAndGrantsAnUntapForFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := tdmRSeats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", tdmRBearTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Ringing Strike Mastery", "Enchantment — Aura", tdmRingingOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B13Tapped(t, g, bear) {
		t.Fatal("the enchanted creature was not tapped on entry")
	}
	idx, ref := effectstest.GrantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the enchanted creature has no granted activated ability")
	}
	effectstest.FillPoolColored(me, "U", 5)
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, game.ActivateAbilityParams{Ref: ref}); err != nil {
		t.Fatalf("activate the granted untap: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B13Tapped(t, g, bear) {
		t.Error("{5}: Untap this creature did not untap it")
	}
}

func TestTdmRingingStrikeMasteryKeepsItTappedThroughTheUntapStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", tdmRBearTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Ringing Strike Mastery", "Enchantment — Aura", tdmRingingOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	if !effectstest.B13Tapped(t, g, bear) {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

// --- Riverwalk Technique -------------------------------------------------------

func TestTdmRiverwalkTechniqueLetsTheOwnerTuckANonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := tdmRSeats(g)
	effectstest.SeedLibrary(opp, "X", "Y")
	relic := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	if _, err := effectstest.CastModalErr(t, g, "Riverwalk Technique", "Instant", tdmRiverwalkOracle, []int{0}, effectstest.CardRefs(land)); err == nil {
		t.Error("a land was accepted for the tuck")
	}
	effectstest.CastModal(t, g, "Riverwalk Technique", "Instant", tdmRiverwalkOracle, []int{0}, effectstest.CardRefs(relic))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) != nil {
		t.Fatal("the caster was asked; the owner chooses")
	}
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
	if c == nil {
		t.Fatal("the owner was not asked")
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, []uuid.UUID{relic}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if g.Battlefield.Contains(relic) || opp.Library.Cards[0].InstanceID != relic {
		t.Error("the artifact is not on the bottom of its owner's library")
	}
}

func TestTdmRiverwalkTechniqueCountersANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := tdmRSeats(g)
	effectstest.SeedLibrary(me, "A", "B")
	bear := effectstest.B12Creature(g, me.ID, "Bear", tdmRBearTypeLine, 2, 2)
	strike := effectstest.CastCatalogSpell(t, g, "Rebellious Strike", "Instant", tdmRebelliousOracle, effectstest.CardRefs(bear))
	hand := me.Hand.Size()
	effectstest.CastModal(t, g, "Riverwalk Technique", "Instant", tdmRiverwalkOracle, []int{1}, effectstest.CardRefs(strike))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(strike) {
		t.Error("the countered spell is not in the graveyard")
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("the Bear has power %d, want 2: the countered pump resolved", p)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d: the countered spell drew", hand, me.Hand.Size())
	}
}

// --- Roar of Endless Song ------------------------------------------------------

func TestTdmRoarOfEndlessSongChapterOneMakesAFiveFiveElephant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := tdmRSeats(g)
	effectstest.FinS3CastSaga(t, g, "Roar of Endless Song", tdmRSagaTypeLine, tdmRoarOracle, "", 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	elephants := effectstest.SocB8Named(g, me.ID, "Elephant")
	if len(elephants) != 1 {
		t.Fatalf("%d Elephants, want 1", len(elephants))
	}
	if e := elephants[0]; e.Power != 5 || e.Toughness != 5 || len(e.Colors) != 1 || e.Colors[0] != "G" {
		t.Errorf("Elephant %d/%d %v, want a 5/5 green Elephant", e.Power, e.Toughness, e.Colors)
	}
}

func TestTdmRoarOfEndlessSongChapterThreeDoublesYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", tdmRBearTypeLine, 2, 2)
	glass := effectstest.B12Creature(g, me.ID, "Glass", "Creature — Golem", 3, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", tdmRBearTypeLine, 2, 2)
	effectstest.FinS3Saga(g, me.ID, "Roar of Endless Song", tdmRSagaTypeLine, tdmRoarOracle, 0, 0, 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, glass); p != 6 || tough != 2 {
		t.Errorf("the Golem is %d/%d, want 6/2", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

// --- Roiling Dragonstorm -------------------------------------------------------

func TestTdmRoilingDragonstormDrawsTwoThenDiscardsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := tdmRSeats(g)
	drawn := effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.CastCatalogSpell(t, g, "Roiling Dragonstorm", "Enchantment", tdmRoilingOracle, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DiscardOwed(g, me.ID); got != 1 {
		t.Fatalf("discard owed %d, want 1", got)
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d -> %d, want two drawn", hand, me.Hand.Size())
	}
	var pick uuid.UUID
	for _, id := range drawn {
		if me.Hand.Contains(id) {
			pick = id
			break
		}
	}
	effectstest.AnswerDiscard(t, g, me.ID, pick)
	if me.Hand.Size() != hand+1 || me.Hand.Contains(pick) {
		t.Errorf("hand %d after the discard, want %d without the discarded card", me.Hand.Size(), hand+1)
	}
}

func TestTdmRoilingDragonstormReturnsWhenADragonEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := tdmRSeats(g)
	storm := effectstest.PushCatalogPermanent(g, me.ID, "Roiling Dragonstorm", "Enchantment", tdmRoilingOracle, false)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Bear", TypeLine: tdmRBearTypeLine, Power: 2, Toughness: 2})
	if !g.Battlefield.Contains(storm) {
		t.Fatal("a Bear is not a Dragon")
	}
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Dragon", TypeLine: "Creature — Dragon", Power: 4, Toughness: 4})
	if g.Battlefield.Contains(storm) || !me.Hand.Contains(storm) {
		t.Error("a Dragon you control entering did not return the enchantment to your hand")
	}
}

// --- Rot-Curse Rakshasa --------------------------------------------------------

func TestTdmRotCurseRakshasaHasTrampleAndDecayed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := tdmRSeats(g)
	rak := effectstest.B12Push(g, me.ID, "Rot-Curse Rakshasa", "Creature — Demon", tdmRotCurseOracle, 5, 5)
	if !effectstest.HasKeywordOf(g, rak, "trample") || !effectstest.HasKeywordOf(g, rak, game.KeywordDecayed) {
		t.Error("Rot-Curse Rakshasa lacks trample or decayed")
	}
	if effectstest.AvB13Restrictions(g, rak)&game.CantBlock == 0 {
		t.Error("a decayed creature can block")
	}
}

func TestTdmRotCurseRakshasaRenewsDecayedCountersOnXTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := tdmRSeats(g)
	a := effectstest.B12Creature(g, opp.ID, "A", tdmRBearTypeLine, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", tdmRBearTypeLine, 2, 2)
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Rot-Curse Rakshasa", "Creature — Demon", tdmRotCurseOracle, 5, 5,
		"{B}{B}{B}{B}", game.ActivateAbilityParams{XValue: 2, Targets: effectstest.CardRefs(a, b)})
	if !g.Exile.Contains(id) {
		t.Fatal("renew's cost did not exile the Rakshasa")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []uuid.UUID{a, b} {
		if got := effectstest.CountersOn(g, c, game.CounterDecayed); got != 1 {
			t.Errorf("a target has %d decayed counters, want 1", got)
		}
		if !effectstest.HasKeywordOf(g, c, game.KeywordDecayed) {
			t.Error("a creature with a decayed counter does not have decayed")
		}
	}
}

func TestTdmRotCurseRakshasaRenewWantsExactlyXTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := tdmRSeats(g)
	a := effectstest.B12Creature(g, opp.ID, "A", tdmRBearTypeLine, 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogGraveyardCard(me, "Rot-Curse Rakshasa", "Creature — Demon", tdmRotCurseOracle, 5, 5)
	effectstest.FillPoolColored(me, "B", 4)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{XValue: 2, Targets: effectstest.CardRefs(a)}); err == nil {
		t.Error("renew with X=2 accepted one target")
	}
}
