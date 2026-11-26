package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b165_cards_test.go — Modern Horizons, S82 older sets, batch 165
// (#5519). Each test plays the real catalog entry.

const (
	oldB165Watcher    = "277aafa6-4270-4d24-a888-cf9e8e635d4b"
	oldB165Witch      = "594a57f6-7827-41c7-aa37-00b498720b16"
	oldB165Treefolk   = "9d341ed4-4f9b-40df-9bd9-3bfeaf0fe595"
	oldB165Cancrix    = "9e00188a-0743-4a39-9229-977a0a115659"
	oldB165Igneous    = "4dc8e91a-2c00-4de2-bada-80f323a4ec9d"
	oldB165Headless   = "9aa9c7ca-e160-4bfa-8ba5-87114222fd86"
	oldB165Bladeback  = "a64656b2-0f0c-4239-9dfe-a9ddc7ada733"
	oldB165Swipe      = "c74c87f7-3120-4fd7-8735-1d7a7a429643"
	oldB165Nether     = "194e2563-c0bf-4151-bbb0-0d626130b055"
	oldB165Elk        = "d17a1109-3a7f-4a39-b602-155da819235f"
	oldB165String     = "23985249-c54e-4bfa-b69a-512422a46a77"
	oldB165Prohibit   = "d95f1797-56f9-41be-a5fe-8398961f4b8b"
	oldB165SnowLand   = "Snow Land"
	oldB165BearType   = "Creature — Bear"
	oldB165SliverType = "Creature — Sliver"
)

func TestOldB165EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB165Watcher:   "Watcher for Tomorrow",
		oldB165Witch:     "Warteye Witch",
		oldB165Treefolk:  "Abominable Treefolk",
		oldB165Cancrix:   "Iceberg Cancrix",
		oldB165Igneous:   "Igneous Elemental",
		oldB165Headless:  "Headless Specter",
		oldB165Bladeback: "Bladeback Sliver",
		oldB165Swipe:     "Savage Swipe",
		oldB165Nether:    "Nether Spirit",
		oldB165Elk:       "Bellowing Elk",
		oldB165String:    "String of Disappearances",
		oldB165Prohibit:  "Prohibit",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Warteye Witch ----------------------------------------------------

func TestOldB165WarteyeWitchScriesForYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Warteye Witch", "Creature — Goblin Shaman", oldB165Witch, 3, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB165BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB165BearType, 2, 2)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Fatal("an opponent's creature dying scried")
	}
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	sc := effectstest.ScryChoiceFor(g, me.ID)
	if sc == nil || len(sc.ScryCards) != 1 {
		t.Fatalf("scry 1 after my creature died: %+v", sc)
	}
}

func TestOldB165WarteyeWitchScriesWhenItselfDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	witch := effectstest.B12Push(g, me.ID, "Warteye Witch", "Creature — Goblin Shaman", oldB165Witch, 3, 2)
	effectstest.KillCreature(t, g, me.ID, witch)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("the Witch's own death did not scry")
	}
}

// --- Bladeback Sliver -------------------------------------------------

func TestOldB165BladebackSliverGrantsThePingOnlyWithAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	sliver := effectstest.B12Push(g, me.ID, "Bladeback Sliver", "Creature — Sliver", oldB165Bladeback, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other Sliver", oldB165SliverType, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Sliver", oldB165SliverType, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB165BearType, 2, 2)
	foe := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	// granted returns the activation parameters of the granted ping on id,
	// and false when the creature has no granted ability.
	granted := func(id uuid.UUID) (int, game.ActivateAbilityParams, bool) {
		idx, ref := effectstest.GrantedActivatedIndex(t, g, id)
		return idx, game.ActivateAbilityParams{Ref: ref, Strict: true, Targets: []game.TargetRef{foe}}, idx >= 0
	}
	ping := func(id uuid.UUID) {
		t.Helper()
		idx, params, ok := granted(id)
		if !ok {
			t.Fatal("the creature has no granted ping")
		}
		effectstest.B16Activate(t, g, me.ID, id, idx, params)
		effectstest.PassPriorityAroundTable(t, g)
	}

	if _, _, ok := granted(sliver); ok {
		t.Fatal("the ping is granted with a card in hand")
	}
	oldB165EmptyHand(g, me)
	for _, id := range []uuid.UUID{theirs, bear} {
		if _, _, ok := granted(id); ok {
			t.Errorf("a creature that is not one of my Slivers got the ping")
		}
	}
	opp.Life = 40
	ping(other)
	if opp.Life != 39 {
		t.Errorf("life = %d after another Sliver's ping, want 39", opp.Life)
	}
	ping(sliver)
	if opp.Life != 38 {
		t.Errorf("life = %d after the Bladeback Sliver's own ping, want 38", opp.Life)
	}
	oldB165DrawACard(g, me)
	if _, _, ok := granted(other); ok {
		t.Error("the ping outlived the empty hand")
	}
}

// oldB165EmptyHand discards the seat's hand through the discard event, so the
// layer pass that reads the hand size re-reads it (the hand-size hint).
func oldB165EmptyHand(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() {
		for p.Hand.Size() > 0 {
			c, err := p.Hand.PopTop()
			if err != nil {
				return
			}
			g.EmitEvent(game.Event{Kind: game.EventDiscardCard, CardID: c.InstanceID, OldZone: game.ZoneHand, NewZone: game.ZoneGraveyard})
		}
	})
}

// oldB165DrawACard puts a card in the seat's hand through the draw event.
func oldB165DrawACard(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.NewCard("Drawn", p.ID))
		g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: p.ID, OldZone: game.ZoneLibrary, NewZone: game.ZoneHand})
	})
}

// --- Headless Specter -------------------------------------------------

func TestOldB165HeadlessSpecterDiscardsAtRandomOnlyWhenHellbent(t *testing.T) {
	for _, tc := range []struct {
		name     string
		myHand   int
		wantDisc int
	}{{"hellbent", 0, 1}, {"a card in hand", 1, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := g.Seats[0], g.Seats[1]
			spec := effectstest.B12Push(g, me.ID, "Headless Specter", "Creature — Specter", oldB165Headless, 2, 2)
			me.Hand.Cards = nil
			for i := 0; i < tc.myHand; i++ {
				me.Hand.PushTop(game.NewCard("Held", me.ID))
			}
			opp.Hand.Cards = nil
			for i := 0; i < 3; i++ {
				opp.Hand.PushTop(game.NewCard("Theirs", opp.ID))
			}
			effectstest.DealCombatDamageToPlayer(g, spec, opp.ID, 2)
			effectstest.PassPriorityAroundTable(t, g)
			if got := 3 - opp.Hand.Size(); got != tc.wantDisc {
				t.Errorf("opponent discarded %d, want %d", got, tc.wantDisc)
			}
		})
	}
}

func TestOldB165HeadlessSpecterChecksHellbentAgainAsItResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	spec := effectstest.B12Push(g, me.ID, "Headless Specter", "Creature — Specter", oldB165Headless, 2, 2)
	me.Hand.Cards = nil
	opp.Hand.Cards = nil
	opp.Hand.PushTop(game.NewCard("Theirs", opp.ID))
	effectstest.DealCombatDamageToPlayer(g, spec, opp.ID, 2)
	me.Hand.PushTop(game.NewCard("Drawn in response", me.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != 1 {
		t.Errorf("the opponent discarded though I had a card in hand as the ability resolved")
	}
}

// --- Prohibit ---------------------------------------------------------

func TestOldB165ProhibitCountersBySpellManaValue(t *testing.T) {
	for _, tc := range []struct {
		name    string
		cost    string
		kicked  bool
		counter bool
	}{
		{"mana value 2, unkicked", "{1}{G}", false, true},
		{"mana value 3, unkicked", "{2}{G}", false, false},
		{"mana value 3, kicked", "{2}{G}", true, true},
		{"mana value 4, kicked", "{3}{G}", true, true},
		{"mana value 5, kicked", "{4}{G}", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			// Their spell is a Lightning Bolt aimed at me, priced as the case says:
			// it hurts only if it is not countered.
			bolt := effectstest.B13OpponentCasts(t, g, opp, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, tc.cost, effectstest.B16TargetPlayer(me.ID))
			life := me.Life
			params := game.CastSpellParams{Targets: effectstest.B16TargetCard(bolt)}
			if tc.kicked {
				params.OptionalCosts = []int{0}
			}
			id := effectstest.HandCardFull(me, "Prohibit", "Instant", "{1}{U}", oldB165Prohibit, nil)
			if err := g.CastSpell(me.ID, id, params); err != nil {
				t.Fatalf("Prohibit: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if hurt := me.Life < life; hurt == tc.counter {
				t.Errorf("Bolt dealt damage = %v, want %v (life %d -> %d)", hurt, !tc.counter, life, me.Life)
			}
			if !effectstest.GraveyardHas(opp, bolt) {
				t.Error("the spell is not in its owner's graveyard")
			}
		})
	}
}

// --- Savage Swipe -----------------------------------------------------

func TestOldB165SavageSwipePumpsOnlyAPowerTwoCreatureThenFights(t *testing.T) {
	for _, tc := range []struct {
		name         string
		power        int
		wantPower    int
		wantTheirsOK bool
	}{{"power 2", 2, 4, false}, {"power 3", 3, 3, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mine := effectstest.B12Creature(g, me.ID, "Mine", oldB165BearType, tc.power, 3)
			theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB165BearType, 1, tc.power+0)
			effectstest.CastCatalogSpell(t, g, "Savage Swipe", "Sorcery", oldB165Swipe,
				effectstest.CardTargets(mine, theirs))
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.LayeredCard(t, g, mine).CurrentPower(); got != tc.wantPower {
				t.Errorf("my creature has power %d, want %d", got, tc.wantPower)
			}
			if effectstest.BattlefieldHas(g, theirs) {
				t.Error("the fight did not kill the opposing creature")
			}
		})
	}
}

func TestOldB165SavageSwipeWillNotTargetYourOwnSecondCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "A", oldB165BearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", oldB165BearType, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Savage Swipe", "Sorcery", oldB165Swipe, effectstest.CardTargets(a, b)); err == nil {
		t.Error("fought a creature I control")
	}
}

// --- Nether Spirit ----------------------------------------------------

func TestOldB165NetherSpiritReturnsOnlyAsTheSoleCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spirit := effectstest.PushGraveyardMaterial(me, game.Card{
		Name: "Nether Spirit", TypeLine: "Creature — Spirit", OracleID: oldB165Nether, Power: 2, Toughness: 2,
		InstanceID: uuid.New(), Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.PushGraveyardCreature(g, me.ID, "Bear", "{1}{G}")
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AssistCardTriggerPrompt(g, me.ID) != nil || effectstest.BattlefieldHas(g, spirit) {
		t.Fatal("it returned beside another creature card")
	}
	_ = bear

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	spirit2 := effectstest.PushGraveyardMaterial(me2, game.Card{
		Name: "Nether Spirit", TypeLine: "Creature — Spirit", OracleID: oldB165Nether, Power: 2, Toughness: 2,
		InstanceID: uuid.New(), Owner: me2.ID, Controller: me2.ID,
	})
	effectstest.PushGraveyardCard(g2, me2, "A Sorcery")
	effectstest.BeginUpkeep(t, g2, me2.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g2, me2.ID, false)
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.BattlefieldHas(g2, spirit2) {
		t.Fatal("it returned though I declined")
	}
	effectstest.BeginUpkeep(t, g2, me2.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g2, me2.ID, true)
	effectstest.PassPriorityAroundTable(t, g2)
	if !effectstest.BattlefieldHas(g2, spirit2) {
		t.Fatal("it did not return as the only creature card")
	}
}

// --- Bellowing Elk ----------------------------------------------------

func TestOldB165BellowingElkNeedsAnotherCreatureOfYoursToHaveEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	elk := effectstest.EnterFromHand(t, g, me.ID, "Bellowing Elk", "Creature — Elk", oldB165Elk)
	has := func() (bool, bool) {
		c := effectstest.LayeredCard(t, g, elk)
		return game.HasKeyword(&c, "trample"), game.HasKeyword(&c, "indestructible")
	}
	if tr, in := has(); tr || in {
		t.Fatalf("the Elk's own entry counted as another creature: trample %v, indestructible %v", tr, in)
	}
	effectstest.EnterFromHand(t, g, opp.ID, "Their Bear", oldB165BearType, "")
	if tr, in := has(); tr || in {
		t.Fatalf("an opponent's creature entering counted: trample %v, indestructible %v", tr, in)
	}
	effectstest.EnterFromHand(t, g, me.ID, "My Bear", oldB165BearType, "")
	if tr, in := has(); !tr || !in {
		t.Errorf("after my other creature entered: trample %v, indestructible %v, want both", tr, in)
	}
}

// --- String of Disappearances -----------------------------------------

func TestOldB165StringOfDisappearancesBouncesAndLetsTheControllerPayToCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := effectstest.B12Creature(g, opp.ID, "First", oldB165BearType, 2, 2)
	second := effectstest.B12Creature(g, opp.ID, "Second", oldB165BearType, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(opp, "U", "U")
	id := effectstest.HandCardFull(me, "String of Disappearances", "Instant", "{U}", oldB165String, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(first)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, first) {
		t.Fatal("the target was not returned to its owner's hand")
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("the creature's controller was not asked to pay {U}{U}")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, true)
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	if p := effectstest.LatestPickTarget(g, opp.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, opp.ID, game.TargetRef{Kind: game.TargetCard, ID: second}); err != nil {
			t.Fatalf("new target for the copy: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, second) {
		t.Error("the copy did not return the second creature")
	}
}

func TestOldB165StringOfDisappearancesDeclinedPaymentEndsTheChain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := effectstest.B12Creature(g, opp.ID, "First", oldB165BearType, 2, 2)
	second := effectstest.B12Creature(g, opp.ID, "Second", oldB165BearType, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(opp, "U", "U")
	id := effectstest.HandCardFull(me, "String of Disappearances", "Instant", "{U}", oldB165String, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(first)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, second) {
		t.Error("the second creature was bounced though the payment was declined")
	}
}

// --- Igneous Elemental ------------------------------------------------

func TestOldB165IgneousElementalCostsTwoLessWithALandCardInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "R", "C", "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Igneous Elemental", "Creature — Elemental", "{4}{R}{R}", oldB165Igneous, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for four mana with no land card in the graveyard")
	}
	effectstest.WthGraveyardCard(me, "Forest", "Basic Land — Forest")
	effectstest.WthGraveyardCard(me, "Mountain", "Basic Land — Mountain")
	effectstest.CastFromHandForTest(t, g, me, "Igneous Elemental", "Creature — Elemental", "{4}{R}{R}", oldB165Igneous, game.CastSpellParams{Strict: true})
}

func TestOldB165IgneousElementalDealsTwoWhenYouSayYes(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB165BearType, 2, 2)
		effectstest.CastCatalogSpell(t, g, "Igneous Elemental", "Creature — Elemental", oldB165Igneous, nil)
		pick := game.TargetRef{Kind: game.TargetCard, ID: bear}
		effectstest.AssistCardSettle(t, g, me.ID, &pick, yes)
		if alive := effectstest.BattlefieldHas(g, bear); alive == yes {
			t.Errorf("answered %v: the bear is on the battlefield = %v", yes, alive)
		}
	}
}

// --- Iceberg Cancrix --------------------------------------------------

func TestOldB165IcebergCancrixMillsForAnotherSnowPermanentOfYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.EnterFromHand(t, g, me.ID, "Iceberg Cancrix", "Snow Creature — Crab", oldB165Cancrix)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.AssistCardTriggerPrompt(g, me.ID) != nil {
		t.Fatal("the Cancrix itself entering triggered it")
	}
	effectstest.EnterFromHand(t, g, me.ID, "Plain Land", "Basic Land — Forest", "")
	effectstest.EnterFromHand(t, g, opp.ID, "Their Snow Land", "Basic Snow Land — Island", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.AssistCardTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a nonsnow permanent, or an opponent's snow permanent, triggered it")
	}
	before := len(opp.Library.Cards)
	effectstest.EnterFromHand(t, g, me.ID, "Snow-Covered Island", "Basic Snow Land — Island", "")
	pick := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	effectstest.AssistCardSettle(t, g, me.ID, &pick, true)
	if got := before - len(opp.Library.Cards); got != 2 {
		t.Errorf("the opponent milled %d cards, want 2", got)
	}
}

// --- Abominable Treefolk ----------------------------------------------

func TestOldB165AbominableTreefolkCountsYourSnowPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	folk := effectstest.B12Push(g, me.ID, "Abominable Treefolk", "Snow Creature — Treefolk", oldB165Treefolk, 0, 0)
	if p, tough := effectstest.PtOf(t, g, folk); p != 1 || tough != 1 {
		t.Errorf("alone it is %d/%d, want 1/1 (itself is snow)", p, tough)
	}
	effectstest.B12Permanent(g, me.ID, "Snow-Covered Island", "Basic Snow Land — Island")
	effectstest.B12Permanent(g, me.ID, "Plain Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, opp.ID, "Their Snow Land", "Basic Snow Land — Island")
	if p, tough := effectstest.PtOf(t, g, folk); p != 2 || tough != 2 {
		t.Errorf("with one more snow permanent it is %d/%d, want 2/2", p, tough)
	}
}

func TestOldB165AbominableTreefolkTapsAndFreezesAnOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB165BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Abominable Treefolk", "Snow Creature — Treefolk", oldB165Treefolk, nil)
	pick := game.TargetRef{Kind: game.TargetCard, ID: bear}
	effectstest.AssistCardSettle(t, g, me.ID, &pick, true)
	if !effectstest.HmlTapped(t, g, bear) {
		t.Fatal("the opposing creature is untapped")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if !effectstest.HmlTapped(t, g, bear) {
		t.Error("it untapped during its controller's next untap step")
	}
}

// --- Watcher for Tomorrow ---------------------------------------------

func TestOldB165WatcherForTomorrowHidesACardAndReturnsItWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[0]
	top := effectstest.StackTopLibrary(me, "A", "B", "C", "D", "E")
	watcher := game.NewCard("Watcher for Tomorrow", me.ID)
	watcher.OracleID = oldB165Watcher
	watcher.TypeLine = "Creature — Human Wizard"
	watcher.Power, watcher.Toughness = 2, 1
	effectstest.EnterWithHideaway(t, g, watcher, top[1])
	c, ok := effectstest.CardOnBattlefield(g, watcher.InstanceID)
	if !ok || !c.Tapped {
		t.Fatalf("the Watcher is on the battlefield = %v, tapped = %v; want it to enter tapped", ok, c.Tapped)
	}
	if hidden := effectstest.HiddenCardsIn(g); len(hidden) != 1 || hidden[0].InstanceID != top[1] {
		t.Fatalf("hidden cards = %v, want just the picked one", hidden)
	}
	if effectstest.ZoneOfCard(g, top[1]) != game.ZoneExile {
		t.Fatal("the picked card is not in exile")
	}
	effectstest.KillCreature(t, g, me.ID, watcher.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, top[1]) != game.ZoneHand {
		t.Errorf("the exiled card is in %v, want my hand", effectstest.ZoneOfCard(g, top[1]))
	}
}
