package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestBlbForageEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.BlbFgCamelliaOracle:    "Camellia, the Seedmiser",
		effectstest.BlbFgCorpseberryOracle: "Corpseberry Cultivator",
		effectstest.BlbFgCuriousOracle:     "Curious Forager",
		effectstest.BlbFgFeedOracle:        "Feed the Cycle",
		effectstest.BlbFgTreetopOracle:     "Treetop Sentries",
		effectstest.BlbFgBushyOracle:       "Bushy Bodyguard",
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

// fgCamellia is Camellia under me, untapped and able to activate, beside
// another Squirrel.
func fgCamellia(g *game.Game, me *game.Player) (camellia, squirrel uuid.UUID) {
	camellia = effectstest.PushCatalogPermanent(g, me.ID, "Camellia, the Seedmiser", "Legendary Creature — Squirrel Warlock", effectstest.BlbFgCamelliaOracle, false)
	squirrel = effectstest.PushCatalogPermanent(g, me.ID, "Squirrel Friend", "Creature — Squirrel", "", false)
	return camellia, squirrel
}

func fgActivateCamellia(g *game.Game, me *game.Player, camellia uuid.UUID, branch *int, exile, sac []uuid.UUID) error {
	return g.ActivateCatalogAbility(me.ID, camellia, 0, game.ActivateAbilityParams{
		CostBranch: branch, ExileIDs: exile, SacrificeIDs: sac,
	})
}

// --- Camellia, the Seedmiser: forage as an activated ability's cost -----

// Paying by exiling three cards from the graveyard: the three are in
// exile, the stack item says it foraged, the event fires once, and the
// other Squirrel gets its counter.
func TestBlbForageCamelliaPaysByExilingThree(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	camellia, squirrel := fgCamellia(g, me)
	gy := effectstest.FgGraveyard(me, 4)
	effectstest.FillPool(me, 2)
	zero := 0
	if err := fgActivateCamellia(g, me, camellia, &zero, gy[:3], nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	it := effectstest.EoTopAbility(g)
	if it == nil || !it.Paid.Foraged || it.Paid.CostBranch != 1 || len(it.Paid.Exiled) != 3 {
		t.Fatalf("stack item %+v: want foraged, branch 1, three exiled", it)
	}
	for _, id := range gy[:3] {
		if !effectstest.Exiled(g, id) {
			t.Errorf("graveyard card %s was not exiled", id)
		}
	}
	if !effectstest.InGraveyard(me, gy[3]) {
		t.Error("the fourth graveyard card left; forage exiles exactly three")
	}
	if n := effectstest.FgForages(g, me.ID); n != 1 {
		t.Errorf("%d forage events, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.FgPlusOnes(t, g, squirrel); n != 1 {
		t.Errorf("the other Squirrel has %d +1/+1 counters, want 1", n)
	}
	if n := effectstest.FgPlusOnes(t, g, camellia); n != 0 {
		t.Errorf("Camellia has %d +1/+1 counters; \"each other\" leaves her out", n)
	}
}

// Paying by sacrificing a Food: a Food token and a Food artifact
// creature each pay, and the sacrifice also triggers Camellia's own
// "whenever you sacrifice one or more Foods".
func TestBlbForageCamelliaPaysBySacrificingAFood(t *testing.T) {
	for _, tc := range []struct {
		name string
		food func(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID
	}{
		{"a Food token", effectstest.FgFoodToken},
		{"a Food artifact creature", func(_ *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
			return effectstest.PushCatalogPermanent(g, owner, "Gingerbread Golem", "Artifact Creature — Food Golem", "", false)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := effectstest.EoTable(t)
			camellia, squirrel := fgCamellia(g, me)
			food := tc.food(t, g, me.ID)
			effectstest.FgGraveyard(me, 2)
			effectstest.FillPool(me, 2)
			if got := effectstest.EoActivationBranches(t, g, me.ID, camellia); !got[1] || got[0] || got[-1] {
				t.Fatalf("offered branches %v, want the Food only", got)
			}
			one := 1
			if err := fgActivateCamellia(g, me, camellia, &one, nil, []uuid.UUID{food}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if effects.OnBattlefield(g, food) {
				t.Fatal("the Food was not sacrificed")
			}
			if it := effectstest.EoTopAbility(g); it == nil || !it.Paid.Foraged || len(it.Paid.SacrificedObjects) != 1 {
				t.Fatalf("stack item %+v: want foraged with the Food recorded", it)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if n, _ := effectstest.CountTokensNamed(g, me.ID, "Squirrel"); n != 1 {
				t.Errorf("%d Squirrel tokens, want 1 from sacrificing a Food", n)
			}
			if n := effectstest.FgPlusOnes(t, g, squirrel); n != 1 {
				t.Errorf("the other Squirrel has %d +1/+1 counters, want 1", n)
			}
		})
	}
}

// With two cards in the graveyard and no Food, forage can't be paid
// either way: the enumerator offers no activation and the engine refuses
// every attempt with nothing paid (CR 118.3).
func TestBlbForageCamelliaRefusedWithTwoCardsAndNoFood(t *testing.T) {
	g, me, opp := effectstest.EoTable(t)
	camellia, _ := fgCamellia(g, me)
	gy := effectstest.FgGraveyard(me, 2)
	oppFood := effectstest.FgFoodToken(t, g, opp.ID)
	effectstest.FillPool(me, 2)
	if got := effectstest.EoActivationBranches(t, g, me.ID, camellia); len(got) != 0 {
		t.Fatalf("offered %v, want nothing", got)
	}
	zero, one := 0, 1
	for name, err := range map[string]error{
		"two cards":          fgActivateCamellia(g, me, camellia, &zero, gy, nil),
		"no branch":          fgActivateCamellia(g, me, camellia, nil, gy, nil),
		"an opponent's Food": fgActivateCamellia(g, me, camellia, &one, nil, []uuid.UUID{oppFood}),
	} {
		if err == nil {
			t.Errorf("%s: activation accepted", name)
		}
	}
	if effectstest.EoTopAbility(g) != nil || !effectstest.InGraveyard(me, gy[0]) || !effectstest.InGraveyard(me, gy[1]) || !effects.OnBattlefield(g, oppFood) {
		t.Error("a refused activation paid something")
	}
	if n := effectstest.FgForages(g, me.ID); n != 0 {
		t.Errorf("%d forage events after refusals", n)
	}
}

// Sacrificing two Foods at once is one "one or more Foods" trigger.
func TestBlbForageCamelliaMakesOneSquirrelPerBatchOfFoods(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	fgCamellia(g, me)
	a, b := effectstest.FgFoodToken(t, g, me.ID), effectstest.FgFoodToken(t, g, me.ID)
	g.WithWriteLock(func() {
		if err := g.SacrificeAllKeyedForEffect(uuid.Nil, []uuid.UUID{a, b}, game.KeyedTail{}); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Squirrel"); n != 1 {
		t.Errorf("%d Squirrel tokens, want 1 for one batch of two Foods", n)
	}
}

// The forage record survives a restore with the ability on the stack.
func TestBlbForageCamelliaForagedSurvivesARestore(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	camellia, squirrel := fgCamellia(g, me)
	gy := effectstest.FgGraveyard(me, 3)
	effectstest.FillPool(me, 2)
	zero := 0
	if err := fgActivateCamellia(g, me, camellia, &zero, gy, nil); err != nil {
		t.Fatal(err)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	if it := effectstest.EoTopAbility(g); it == nil || !it.Paid.Foraged {
		t.Fatalf("restored stack item %+v, want Foraged", it)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.FgPlusOnes(t, g, squirrel); n != 1 {
		t.Errorf("after the restore the Squirrel has %d counters, want 1", n)
	}
}

// --- Feed the Cycle: forage as a branch of an additional cost ----------

func fgCastFeed(t *testing.T, g *game.Game, branch int, target uuid.UUID, exile, sac []uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	return effectstest.CastWithTapParams(t, g, "Feed the Cycle", "Instant", "{1}{B}", effectstest.BlbFgFeedOracle, game.CastSpellParams{
		CostBranch: &branch, ExileIDs: exile, SacrificeIDs: sac,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
}

func TestBlbForageFeedTheCycleDestroysForEachWayToPay(t *testing.T) {
	for _, tc := range []struct {
		name    string
		branch  int
		foraged bool
	}{{"exile three", 0, true}, {"sacrifice a Food", 1, true}, {"pay {B}", 2, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := effectstest.EoTable(t)
			bear := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
			gy := effectstest.FgGraveyard(me, 3)
			food := effectstest.FgFoodToken(t, g, me.ID)
			effectstest.FillPoolColored(me, "B", 3)
			var exile, sac []uuid.UUID
			switch tc.branch {
			case 0:
				exile = gy
			case 1:
				sac = []uuid.UUID{food}
			}
			spell, err := fgCastFeed(t, g, tc.branch, bear, exile, sac)
			if err != nil {
				t.Fatalf("cast: %v", err)
			}
			if it := g.StackMeta[spell]; it == nil || it.Paid.Foraged != tc.foraged {
				t.Fatalf("stack item %+v: Foraged want %v", it, tc.foraged)
			}
			wantForages := 0
			if tc.foraged {
				wantForages = 1
			}
			if n := effectstest.FgForages(g, me.ID); n != wantForages {
				t.Errorf("%d forage events, want %d", n, wantForages)
			}
			if effectstest.Exiled(g, gy[0]) != (tc.branch == 0) || effects.OnBattlefield(g, food) == (tc.branch == 1) {
				t.Errorf("branch %d paid the wrong way", tc.branch)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if effects.OnBattlefield(g, bear) {
				t.Error("the creature survived")
			}
		})
	}
}

// With two cards in the graveyard and no Food, both forage branches are
// unpayable and refused; paying {B} still casts it.
func TestBlbForageFeedTheCycleForageBranchesNeedThreeCardsOrAFood(t *testing.T) {
	g, me, opp := effectstest.EoTable(t)
	bear := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	gy := effectstest.FgGraveyard(me, 2)
	effectstest.FillPoolColored(me, "B", 3)
	card := game.Card{OracleID: effectstest.BlbFgFeedOracle, Name: "Feed the Cycle", Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() {
		for b, want := range []bool{false, false, true} {
			if got := g.AdditionalCostBranchPayableLocked(me.ID, card, b); got != want {
				t.Errorf("branch %d payable %v, want %v", b, got, want)
			}
		}
	})
	if _, err := fgCastFeed(t, g, 0, bear, gy, nil); err == nil {
		t.Fatal("cast by exiling two cards was accepted")
	}
	if !effectstest.InGraveyard(me, gy[0]) {
		t.Fatal("a refused cast exiled a card")
	}
	if _, err := fgCastFeed(t, g, 2, bear, nil, nil); err != nil {
		t.Fatalf("cast paying {B}: %v", err)
	}
}

// --- Treetop Sentries: "you may forage. If you do" ---------------------

func TestBlbForageTreetopSentriesDrawsOnlyWhenItForages(t *testing.T) {
	for _, tc := range []struct {
		name   string
		cards  int
		food   bool
		option int // -1 declines
		offer  []int
	}{
		{"exile three", 3, false, 0, []int{0}},
		{"sacrifice a Food", 0, true, 1, []int{1}},
		{"decline", 3, true, -1, []int{0, 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := effectstest.EoTable(t)
			gy := effectstest.FgGraveyard(me, tc.cards)
			var food uuid.UUID
			if tc.food {
				food = effectstest.FgFoodToken(t, g, me.ID)
			}
			effectstest.EnterFromHand(t, g, me.ID, "Treetop Sentries", "Creature — Squirrel Archer", effectstest.BlbFgTreetopOracle)
			effectstest.PassPriorityAroundTable(t, g)
			got := effectstest.EoOfferedOptions(t, g, me.ID)
			if len(got) != len(tc.offer) {
				t.Fatalf("offered %v, want %v", got, tc.offer)
			}
			for _, o := range tc.offer {
				if !got[o] {
					t.Fatalf("offered %v, want %v", got, tc.offer)
				}
			}
			hand := me.Hand.Size()
			switch tc.option {
			case 0:
				effectstest.EoPay(t, g, me.ID, 0, gy...)
			case 1:
				effectstest.EoPay(t, g, me.ID, 1, food)
			default:
				effectstest.AnswerPayUnless(t, g, me.ID, false)
			}
			effectstest.PassPriorityAroundTable(t, g)
			want, forages := hand, 0
			if tc.option >= 0 {
				want, forages = hand+1, 1
			}
			if got := me.Hand.Size(); got != want {
				t.Errorf("hand %d, want %d", got, want)
			}
			if n := effectstest.FgForages(g, me.ID); n != forages {
				t.Errorf("%d forage events, want %d", n, forages)
			}
		})
	}
}

// With two cards in the graveyard and no Food, nobody is asked and
// nothing is drawn (CR 118.3).
func TestBlbForageTreetopSentriesAsksNothingWhenItCantForage(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	effectstest.FgGraveyard(me, 2)
	effectstest.EnterFromHand(t, g, me.ID, "Treetop Sentries", "Creature — Squirrel Archer", effectstest.BlbFgTreetopOracle)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EoPrompt(g, me.ID) != nil {
		t.Fatal("asked to forage with two cards and no Food")
	}
	if me.Hand.Size() != hand {
		t.Error("drew without foraging")
	}
}

// The question survives a restore, and the answer still forages.
func TestBlbForageTreetopSentriesQuestionSurvivesARestore(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	gy := effectstest.FgGraveyard(me, 3)
	effectstest.EnterFromHand(t, g, me.ID, "Treetop Sentries", "Creature — Squirrel Archer", effectstest.BlbFgTreetopOracle)
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = effectstest.PlayerByIDForTest(g, me.ID)
	hand := me.Hand.Size()
	effectstest.EoPay(t, g, me.ID, 0, gy...)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || effectstest.FgForages(g, me.ID) != 1 {
		t.Errorf("after a restore: hand %d (want %d), %d forage events", me.Hand.Size(), hand+1, effectstest.FgForages(g, me.ID))
	}
}

// --- Curious Forager: "you may forage. When you do" --------------------

// The reflexive trigger targets after the forage, so a card the forage
// exiled can't be chosen, and a permanent card left in the graveyard
// comes back.
func TestBlbForageCuriousForagerReturnsAPermanentCardAfterForaging(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	gy := effectstest.FgGraveyard(me, 3)
	land := effectstest.PushGraveyardPermanent(me, "Forest", "Basic Land — Forest", "")
	effectstest.EnterFromHand(t, g, me.ID, "Curious Forager", "Creature — Squirrel Druid", effectstest.BlbFgCuriousOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EoPay(t, g, me.ID, 0, gy...)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: gy[0]}); err == nil {
		t.Fatal("a card the forage exiled was accepted as the target")
	}
	effectstest.AnswerPickTarget(t, g, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the permanent card did not return to hand")
	}
}

func TestBlbForageCuriousForagerDoesNothingWhenDeclined(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	effectstest.FgGraveyard(me, 3)
	land := effectstest.PushGraveyardPermanent(me, "Forest", "Basic Land — Forest", "")
	effectstest.EnterFromHand(t, g, me.ID, "Curious Forager", "Creature — Squirrel Druid", effectstest.BlbFgCuriousOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InGraveyard(me, land) || effectstest.FgForages(g, me.ID) != 0 {
		t.Error("declining still foraged or returned a card")
	}
}

// --- Corpseberry Cultivator: "you may forage" and "whenever you forage" --

// Its own beginning-of-combat forage pays off its own "whenever you
// forage".
func TestBlbForageCorpseberryGrowsWhenItForagesAtCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	cult := effectstest.BlbB1Push(g, me.ID, "Corpseberry Cultivator", "Creature — Squirrel Warlock", effectstest.BlbFgCorpseberryOracle, 2, 3)
	food := effectstest.FgFoodToken(t, g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EoPay(t, g, me.ID, 1, food)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, food) {
		t.Error("the Food was not sacrificed")
	}
	if n := effectstest.FgPlusOnes(t, g, cult); n != 1 {
		t.Errorf("the Cultivator has %d +1/+1 counters, want 1", n)
	}
}

// "Whenever you forage" sees every way to forage: another card's forage
// cost and a spell's forage branch each give a counter, and paying the
// spell's {B} does not.
func TestBlbForageCorpseberrySeesEveryForage(t *testing.T) {
	g, me, opp := effectstest.EoTable(t)
	cult := effectstest.BlbB1Push(g, me.ID, "Corpseberry Cultivator", "Creature — Squirrel Warlock", effectstest.BlbFgCorpseberryOracle, 2, 3)
	camellia, _ := fgCamellia(g, me)
	gy := effectstest.FgGraveyard(me, 6)
	effectstest.FillPool(me, 2)
	zero := 0
	if err := fgActivateCamellia(g, me, camellia, &zero, gy[:3], nil); err != nil {
		t.Fatalf("activate Camellia: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.FgPlusOnes(t, g, cult); n != 2 {
		// Camellia's counter on each other Squirrel, and the forage trigger.
		t.Fatalf("after Camellia's forage the Cultivator has %d counters, want 2", n)
	}
	bear := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	effectstest.FillPoolColored(me, "B", 3)
	if _, err := fgCastFeed(t, g, 0, bear, gy[3:], nil); err != nil {
		t.Fatalf("cast Feed the Cycle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.FgPlusOnes(t, g, cult); n != 3 {
		t.Fatalf("after Feed the Cycle's forage the Cultivator has %d counters, want 3", n)
	}
	other := effectstest.PushCatalogPermanent(g, opp.ID, "Other Bear", "Creature — Bear", "", false)
	effectstest.FillPoolColored(me, "B", 3)
	if _, err := fgCastFeed(t, g, 2, other, nil, nil); err != nil {
		t.Fatalf("cast Feed the Cycle for {B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.FgPlusOnes(t, g, cult); n != 3 {
		t.Errorf("paying {B} gave the Cultivator a counter (%d)", n)
	}
}

// --- Bushy Bodyguard: "you may forage. If you do, put two counters on it" --

func TestBlbForageBushyBodyguardGrowsWhenItForages(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	food := effectstest.FgFoodToken(t, g, me.ID)
	bushy := effectstest.EnterFromHand(t, g, me.ID, "Bushy Bodyguard", "Creature — Squirrel Warrior", effectstest.BlbFgBushyOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EoOfferedOptions(t, g, me.ID); !got[1] || got[0] {
		t.Fatalf("offered %v, want the Food only (the graveyard is empty)", got)
	}
	effectstest.EoPay(t, g, me.ID, 1, food)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.FgPlusOnes(t, g, bushy); n != 2 {
		t.Errorf("the Bodyguard has %d +1/+1 counters, want 2", n)
	}
	if effects.OnBattlefield(g, food) {
		t.Error("the Food was not sacrificed")
	}
}

func TestBlbForageBushyBodyguardStaysSmallWithNothingToForage(t *testing.T) {
	g, me, _ := effectstest.EoTable(t)
	effectstest.FgGraveyard(me, 2)
	bushy := effectstest.EnterFromHand(t, g, me.ID, "Bushy Bodyguard", "Creature — Squirrel Warrior", effectstest.BlbFgBushyOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EoPrompt(g, me.ID) != nil {
		t.Fatal("asked to forage with two cards and no Food")
	}
	if n := effectstest.FgPlusOnes(t, g, bushy); n != 0 {
		t.Errorf("the Bodyguard has %d +1/+1 counters without foraging", n)
	}
}
