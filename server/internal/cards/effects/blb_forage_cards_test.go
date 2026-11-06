package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blb_forage_cards_test.go — #5184: the forage keyword action (CR
// 701.61a) as an activated ability's cost (Camellia, the Seedmiser), as
// an either/or additional cost (Feed the Cycle), as an effect's "you may
// forage" (Treetop Sentries, Bushy Bodyguard, Curious Forager, Corpseberry
// Cultivator),
// and "whenever you forage" (Corpseberry Cultivator). Bloomburrow, #4467.
// Every test drives the real catalog card.

const (
	blbFgCamelliaOracle    = "312c7a7b-2084-42b0-9b13-0a6d3e43a73d"
	blbFgCorpseberryOracle = "5ded56df-c7d5-4043-85df-410f47219faa"
	blbFgCuriousOracle     = "30af656a-99ab-4e0a-81fe-3cdb0ea6a0cc"
	blbFgFeedOracle        = "c338eedb-9b70-4b77-b5d6-8ea74b7bea3e"
	blbFgTreetopOracle     = "5bf5739a-6322-4e6e-b41e-e7bdbb11b1ce"
	blbFgBushyOracle       = "5436bb0d-b5c7-43a9-8b3f-67cdc0dd0381"
)

func TestBlbForageEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		blbFgCamelliaOracle:    "Camellia, the Seedmiser",
		blbFgCorpseberryOracle: "Corpseberry Cultivator",
		blbFgCuriousOracle:     "Curious Forager",
		blbFgFeedOracle:        "Feed the Cycle",
		blbFgTreetopOracle:     "Treetop Sentries",
		blbFgBushyOracle:       "Bushy Bodyguard",
	} {
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

// fgGraveyard puts n cards in p's graveyard and returns them, oldest
// first.
func fgGraveyard(p *game.Player, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		c := game.NewCard("Old Card", p.ID)
		c.TypeLine = "Instant"
		p.Graveyard.PushTop(c)
		out = append(out, c.InstanceID)
	}
	return out
}

// fgFoodToken creates a Food token under owner and returns it.
func fgFoodToken(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	before := map[uuid.UUID]bool{}
	for _, c := range g.Battlefield.Cards {
		before[c.InstanceID] = true
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(owner, FoodToken(), 1); err != nil {
			t.Fatal(err)
		}
	})
	for _, c := range g.Battlefield.Cards {
		if !before[c.InstanceID] && c.Name == "Food" {
			return c.InstanceID
		}
	}
	t.Fatal("no Food token was created")
	return uuid.Nil
}

// fgForages counts the EventForages player has made.
func fgForages(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventForage && ev.Actor == player {
			n++
		}
	}
	return n
}

// fgCamellia is Camellia under me, untapped and able to activate, beside
// another Squirrel.
func fgCamellia(g *game.Game, me *game.Player) (camellia, squirrel uuid.UUID) {
	camellia = pushCatalogPermanent(g, me.ID, "Camellia, the Seedmiser", "Legendary Creature — Squirrel Warlock", blbFgCamelliaOracle, false)
	squirrel = pushCatalogPermanent(g, me.ID, "Squirrel Friend", "Creature — Squirrel", "", false)
	return camellia, squirrel
}

func fgActivateCamellia(g *game.Game, me *game.Player, camellia uuid.UUID, branch *int, exile, sac []uuid.UUID) error {
	return g.ActivateCatalogAbility(me.ID, camellia, 0, game.ActivateAbilityParams{
		CostBranch: branch, ExileIDs: exile, SacrificeIDs: sac,
	})
}

func fgPlusOnes(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c.Counters[game.CounterPlusOne]
}

// --- Camellia, the Seedmiser: forage as an activated ability's cost -----

// Paying by exiling three cards from the graveyard: the three are in
// exile, the stack item says it foraged, the event fires once, and the
// other Squirrel gets its counter.
func TestBlbForageCamelliaPaysByExilingThree(t *testing.T) {
	g, me, _ := eoTable(t)
	camellia, squirrel := fgCamellia(g, me)
	gy := fgGraveyard(me, 4)
	fillPool(me, 2)
	zero := 0
	if err := fgActivateCamellia(g, me, camellia, &zero, gy[:3], nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	it := eoTopAbility(g)
	if it == nil || !it.Paid.Foraged || it.Paid.CostBranch != 1 || len(it.Paid.Exiled) != 3 {
		t.Fatalf("stack item %+v: want foraged, branch 1, three exiled", it)
	}
	for _, id := range gy[:3] {
		if !exiled(g, id) {
			t.Errorf("graveyard card %s was not exiled", id)
		}
	}
	if !inGraveyard(me, gy[3]) {
		t.Error("the fourth graveyard card left; forage exiles exactly three")
	}
	if n := fgForages(g, me.ID); n != 1 {
		t.Errorf("%d forage events, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if n := fgPlusOnes(t, g, squirrel); n != 1 {
		t.Errorf("the other Squirrel has %d +1/+1 counters, want 1", n)
	}
	if n := fgPlusOnes(t, g, camellia); n != 0 {
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
		{"a Food token", fgFoodToken},
		{"a Food artifact creature", func(_ *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
			return pushCatalogPermanent(g, owner, "Gingerbread Golem", "Artifact Creature — Food Golem", "", false)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := eoTable(t)
			camellia, squirrel := fgCamellia(g, me)
			food := tc.food(t, g, me.ID)
			fgGraveyard(me, 2)
			fillPool(me, 2)
			if got := eoActivationBranches(t, g, me.ID, camellia); !got[1] || got[0] || got[-1] {
				t.Fatalf("offered branches %v, want the Food only", got)
			}
			one := 1
			if err := fgActivateCamellia(g, me, camellia, &one, nil, []uuid.UUID{food}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if onBattlefield(g, food) {
				t.Fatal("the Food was not sacrificed")
			}
			if it := eoTopAbility(g); it == nil || !it.Paid.Foraged || len(it.Paid.SacrificedObjects) != 1 {
				t.Fatalf("stack item %+v: want foraged with the Food recorded", it)
			}
			passPriorityAroundTable(t, g)
			if n, _ := countTokensNamed(g, me.ID, "Squirrel"); n != 1 {
				t.Errorf("%d Squirrel tokens, want 1 from sacrificing a Food", n)
			}
			if n := fgPlusOnes(t, g, squirrel); n != 1 {
				t.Errorf("the other Squirrel has %d +1/+1 counters, want 1", n)
			}
		})
	}
}

// With two cards in the graveyard and no Food, forage can't be paid
// either way: the enumerator offers no activation and the engine refuses
// every attempt with nothing paid (CR 118.3).
func TestBlbForageCamelliaRefusedWithTwoCardsAndNoFood(t *testing.T) {
	g, me, opp := eoTable(t)
	camellia, _ := fgCamellia(g, me)
	gy := fgGraveyard(me, 2)
	oppFood := fgFoodToken(t, g, opp.ID)
	fillPool(me, 2)
	if got := eoActivationBranches(t, g, me.ID, camellia); len(got) != 0 {
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
	if eoTopAbility(g) != nil || !inGraveyard(me, gy[0]) || !inGraveyard(me, gy[1]) || !onBattlefield(g, oppFood) {
		t.Error("a refused activation paid something")
	}
	if n := fgForages(g, me.ID); n != 0 {
		t.Errorf("%d forage events after refusals", n)
	}
}

// Sacrificing two Foods at once is one "one or more Foods" trigger.
func TestBlbForageCamelliaMakesOneSquirrelPerBatchOfFoods(t *testing.T) {
	g, me, _ := eoTable(t)
	fgCamellia(g, me)
	a, b := fgFoodToken(t, g, me.ID), fgFoodToken(t, g, me.ID)
	g.WithWriteLock(func() {
		if err := g.SacrificeAllKeyedForEffect(uuid.Nil, []uuid.UUID{a, b}, game.KeyedTail{}); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Squirrel"); n != 1 {
		t.Errorf("%d Squirrel tokens, want 1 for one batch of two Foods", n)
	}
}

// The forage record survives a restore with the ability on the stack.
func TestBlbForageCamelliaForagedSurvivesARestore(t *testing.T) {
	g, me, _ := eoTable(t)
	camellia, squirrel := fgCamellia(g, me)
	gy := fgGraveyard(me, 3)
	fillPool(me, 2)
	zero := 0
	if err := fgActivateCamellia(g, me, camellia, &zero, gy, nil); err != nil {
		t.Fatal(err)
	}
	g = restoreRoundTrip(t, g, true)
	if it := eoTopAbility(g); it == nil || !it.Paid.Foraged {
		t.Fatalf("restored stack item %+v, want Foraged", it)
	}
	passPriorityAroundTable(t, g)
	if n := fgPlusOnes(t, g, squirrel); n != 1 {
		t.Errorf("after the restore the Squirrel has %d counters, want 1", n)
	}
}

// --- Feed the Cycle: forage as a branch of an additional cost ----------

func fgCastFeed(t *testing.T, g *game.Game, branch int, target uuid.UUID, exile, sac []uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	return castWithTapParams(t, g, "Feed the Cycle", "Instant", "{1}{B}", blbFgFeedOracle, game.CastSpellParams{
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
			g, me, opp := eoTable(t)
			bear := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
			gy := fgGraveyard(me, 3)
			food := fgFoodToken(t, g, me.ID)
			fillPoolColored(me, "B", 3)
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
			if n := fgForages(g, me.ID); n != wantForages {
				t.Errorf("%d forage events, want %d", n, wantForages)
			}
			if exiled(g, gy[0]) != (tc.branch == 0) || onBattlefield(g, food) == (tc.branch == 1) {
				t.Errorf("branch %d paid the wrong way", tc.branch)
			}
			passPriorityAroundTable(t, g)
			if onBattlefield(g, bear) {
				t.Error("the creature survived")
			}
		})
	}
}

// With two cards in the graveyard and no Food, both forage branches are
// unpayable and refused; paying {B} still casts it.
func TestBlbForageFeedTheCycleForageBranchesNeedThreeCardsOrAFood(t *testing.T) {
	g, me, opp := eoTable(t)
	bear := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	gy := fgGraveyard(me, 2)
	fillPoolColored(me, "B", 3)
	card := game.Card{OracleID: blbFgFeedOracle, Name: "Feed the Cycle", Owner: me.ID, Controller: me.ID}
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
	if !inGraveyard(me, gy[0]) {
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
			g, me, _ := eoTable(t)
			gy := fgGraveyard(me, tc.cards)
			var food uuid.UUID
			if tc.food {
				food = fgFoodToken(t, g, me.ID)
			}
			enterFromHand(t, g, me.ID, "Treetop Sentries", "Creature — Squirrel Archer", blbFgTreetopOracle)
			passPriorityAroundTable(t, g)
			got := eoOfferedOptions(t, g, me.ID)
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
				eoPay(t, g, me.ID, 0, gy...)
			case 1:
				eoPay(t, g, me.ID, 1, food)
			default:
				answerPayUnless(t, g, me.ID, false)
			}
			passPriorityAroundTable(t, g)
			want, forages := hand, 0
			if tc.option >= 0 {
				want, forages = hand+1, 1
			}
			if got := me.Hand.Size(); got != want {
				t.Errorf("hand %d, want %d", got, want)
			}
			if n := fgForages(g, me.ID); n != forages {
				t.Errorf("%d forage events, want %d", n, forages)
			}
		})
	}
}

// With two cards in the graveyard and no Food, nobody is asked and
// nothing is drawn (CR 118.3).
func TestBlbForageTreetopSentriesAsksNothingWhenItCantForage(t *testing.T) {
	g, me, _ := eoTable(t)
	fgGraveyard(me, 2)
	enterFromHand(t, g, me.ID, "Treetop Sentries", "Creature — Squirrel Archer", blbFgTreetopOracle)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if eoPrompt(g, me.ID) != nil {
		t.Fatal("asked to forage with two cards and no Food")
	}
	if me.Hand.Size() != hand {
		t.Error("drew without foraging")
	}
}

// The question survives a restore, and the answer still forages.
func TestBlbForageTreetopSentriesQuestionSurvivesARestore(t *testing.T) {
	g, me, _ := eoTable(t)
	gy := fgGraveyard(me, 3)
	enterFromHand(t, g, me.ID, "Treetop Sentries", "Creature — Squirrel Archer", blbFgTreetopOracle)
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me = playerByIDForTest(g, me.ID)
	hand := me.Hand.Size()
	eoPay(t, g, me.ID, 0, gy...)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || fgForages(g, me.ID) != 1 {
		t.Errorf("after a restore: hand %d (want %d), %d forage events", me.Hand.Size(), hand+1, fgForages(g, me.ID))
	}
}

// --- Curious Forager: "you may forage. When you do" --------------------

// The reflexive trigger targets after the forage, so a card the forage
// exiled can't be chosen, and a permanent card left in the graveyard
// comes back.
func TestBlbForageCuriousForagerReturnsAPermanentCardAfterForaging(t *testing.T) {
	g, me, _ := eoTable(t)
	gy := fgGraveyard(me, 3)
	land := pushGraveyardPermanent(me, "Forest", "Basic Land — Forest", "")
	enterFromHand(t, g, me.ID, "Curious Forager", "Creature — Squirrel Druid", blbFgCuriousOracle)
	passPriorityAroundTable(t, g)
	eoPay(t, g, me.ID, 0, gy...)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: gy[0]}); err == nil {
		t.Fatal("a card the forage exiled was accepted as the target")
	}
	answerPickTarget(t, g, land)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the permanent card did not return to hand")
	}
}

func TestBlbForageCuriousForagerDoesNothingWhenDeclined(t *testing.T) {
	g, me, _ := eoTable(t)
	fgGraveyard(me, 3)
	land := pushGraveyardPermanent(me, "Forest", "Basic Land — Forest", "")
	enterFromHand(t, g, me.ID, "Curious Forager", "Creature — Squirrel Druid", blbFgCuriousOracle)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !inGraveyard(me, land) || fgForages(g, me.ID) != 0 {
		t.Error("declining still foraged or returned a card")
	}
}

// --- Corpseberry Cultivator: "you may forage" and "whenever you forage" --

// Its own beginning-of-combat forage pays off its own "whenever you
// forage".
func TestBlbForageCorpseberryGrowsWhenItForagesAtCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	cult := blbB1Push(g, me.ID, "Corpseberry Cultivator", "Creature — Squirrel Warlock", blbFgCorpseberryOracle, 2, 3)
	food := fgFoodToken(t, g, me.ID)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	eoPay(t, g, me.ID, 1, food)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, food) {
		t.Error("the Food was not sacrificed")
	}
	if n := fgPlusOnes(t, g, cult); n != 1 {
		t.Errorf("the Cultivator has %d +1/+1 counters, want 1", n)
	}
}

// "Whenever you forage" sees every way to forage: another card's forage
// cost and a spell's forage branch each give a counter, and paying the
// spell's {B} does not.
func TestBlbForageCorpseberrySeesEveryForage(t *testing.T) {
	g, me, opp := eoTable(t)
	cult := blbB1Push(g, me.ID, "Corpseberry Cultivator", "Creature — Squirrel Warlock", blbFgCorpseberryOracle, 2, 3)
	camellia, _ := fgCamellia(g, me)
	gy := fgGraveyard(me, 6)
	fillPool(me, 2)
	zero := 0
	if err := fgActivateCamellia(g, me, camellia, &zero, gy[:3], nil); err != nil {
		t.Fatalf("activate Camellia: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := fgPlusOnes(t, g, cult); n != 2 {
		// Camellia's counter on each other Squirrel, and the forage trigger.
		t.Fatalf("after Camellia's forage the Cultivator has %d counters, want 2", n)
	}
	bear := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	fillPoolColored(me, "B", 3)
	if _, err := fgCastFeed(t, g, 0, bear, gy[3:], nil); err != nil {
		t.Fatalf("cast Feed the Cycle: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := fgPlusOnes(t, g, cult); n != 3 {
		t.Fatalf("after Feed the Cycle's forage the Cultivator has %d counters, want 3", n)
	}
	other := pushCatalogPermanent(g, opp.ID, "Other Bear", "Creature — Bear", "", false)
	fillPoolColored(me, "B", 3)
	if _, err := fgCastFeed(t, g, 2, other, nil, nil); err != nil {
		t.Fatalf("cast Feed the Cycle for {B}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := fgPlusOnes(t, g, cult); n != 3 {
		t.Errorf("paying {B} gave the Cultivator a counter (%d)", n)
	}
}

// --- Bushy Bodyguard: "you may forage. If you do, put two counters on it" --

func TestBlbForageBushyBodyguardGrowsWhenItForages(t *testing.T) {
	g, me, _ := eoTable(t)
	food := fgFoodToken(t, g, me.ID)
	bushy := enterFromHand(t, g, me.ID, "Bushy Bodyguard", "Creature — Squirrel Warrior", blbFgBushyOracle)
	passPriorityAroundTable(t, g)
	if got := eoOfferedOptions(t, g, me.ID); !got[1] || got[0] {
		t.Fatalf("offered %v, want the Food only (the graveyard is empty)", got)
	}
	eoPay(t, g, me.ID, 1, food)
	passPriorityAroundTable(t, g)
	if n := fgPlusOnes(t, g, bushy); n != 2 {
		t.Errorf("the Bodyguard has %d +1/+1 counters, want 2", n)
	}
	if onBattlefield(g, food) {
		t.Error("the Food was not sacrificed")
	}
}

func TestBlbForageBushyBodyguardStaysSmallWithNothingToForage(t *testing.T) {
	g, me, _ := eoTable(t)
	fgGraveyard(me, 2)
	bushy := enterFromHand(t, g, me.ID, "Bushy Bodyguard", "Creature — Squirrel Warrior", blbFgBushyOracle)
	passPriorityAroundTable(t, g)
	if eoPrompt(g, me.ID) != nil {
		t.Fatal("asked to forage with two cards and no Food")
	}
	if n := fgPlusOnes(t, g, bushy); n != 0 {
		t.Errorf("the Bodyguard has %d +1/+1 counters without foraging", n)
	}
}
