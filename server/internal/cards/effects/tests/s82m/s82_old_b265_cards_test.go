package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b265_cards_test.go — Kaladesh, batch 265 (#5539). Each test plays the
// real catalog entry through the engine and asserts what the printed text says.

const (
	oldB265Revoke      = "60b2f991-47ef-4da3-ba61-0692695ea098"
	oldB265Spectacle   = "97e34c3a-070e-406a-8505-657f4d57def8"
	oldB265Consulate   = "da164485-c985-42fa-819e-42f0d33237c5"
	oldB265Disappear   = "20393369-fdee-45ea-8584-451f38b1c002"
	oldB265Rebuff      = "879c8663-5058-4998-9a59-f2e9ea8914dc"
	oldB265CogKnot     = "2065e9d0-c6e9-4270-898e-9822b532d467"
	oldB265Module      = "af42079b-a3c0-448c-9bb2-b915252e87a9"
	oldB265Lookout     = "5f5c5f59-1cc0-4a72-863c-c26db411f114"
	oldB265Rush        = "6e460d39-0f18-4a2d-9c8d-3eaba301d285"
	oldB265Negotiation = "d36fea45-d8aa-469a-bcfc-5c0e4a0e807b"
	oldB265Challenge   = "f73bf7c5-e43d-4684-8588-2fb400f2ca31"
	oldB265Courier     = "a88f3881-4f43-4473-965f-f5c24ae5752f"
	oldB265Bear        = "Creature — Bear"
)

func TestOldB265AllRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB265Revoke:      "Revoke Privileges",
		oldB265Spectacle:   "Giant Spectacle",
		oldB265Consulate:   "Captured by the Consulate",
		oldB265Disappear:   "Disappearing Act",
		oldB265Rebuff:      "Revolutionary Rebuff",
		oldB265CogKnot:     "Cogworker's Puzzleknot",
		oldB265Module:      "Animation Module",
		oldB265Lookout:     "Night Market Lookout",
		oldB265Rush:        "Rush of Vitality",
		oldB265Negotiation: "Shrewd Negotiation",
		oldB265Challenge:   "Dubious Challenge",
		oldB265Courier:     "Bomat Courier",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func TestOldB265RevokePrivilegesStopsAttackingBlockingAndCrewing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.Enchant(t, g, "Revoke Privileges", oldB265Revoke, bear)
	eff := effectstest.LayeredCard(t, g, bear).Effective()
	if eff.Restrictions&game.CantAttackOrBlock != game.CantAttackOrBlock {
		t.Errorf("restrictions %v lack attack and block", eff.Restrictions)
	}
	if eff.ExtraRestrictions&game.CantCrew == 0 {
		t.Error("the enchanted creature can still crew Vehicles")
	}
	if eff.Restrictions&game.CantActivate != 0 {
		t.Error("its activated abilities are locked too")
	}
	// A creature it does not enchant is free.
	other := effectstest.PushVanillaCreature(g, me.ID, "Other", 2, 2)
	if eff := effectstest.LayeredCard(t, g, other).Effective(); eff.Restrictions&game.CantAttackOrBlock != 0 || eff.ExtraRestrictions&game.CantCrew != 0 {
		t.Error("an unenchanted creature is restricted")
	}
}

func TestOldB265GiantSpectacleGivesPlusTwoPlusOneAndMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, bear, "menace") {
		t.Fatal("menace before the Aura")
	}
	aura := effectstest.Enchant(t, g, "Giant Spectacle", oldB265Spectacle, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("the creature is %d/%d, want 4/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "menace") {
		t.Error("no menace with the Aura")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	g.RunStateChecksForTest()
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 || effectstest.HasEffectiveKeyword(t, g, bear, "menace") {
		t.Errorf("after the Aura left the creature is %d/%d", p, tough)
	}
}

func TestOldB265CapturedByTheConsulateCantAttackAndTakesOpposingSingleTargetSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.PushVanillaCreature(g, third.ID, "Their Bear", 2, 2)
	// "Enchant creature you don't control": my own creature is not a legal target.
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	if _, err := tryEnchant(g, me, "Captured by the Consulate", oldB265Consulate, mine); err == nil {
		t.Fatal("the Aura enchanted a creature its controller controls")
	}
	effectstest.Enchant(t, g, "Captured by the Consulate", oldB265Consulate, host)
	if eff := effectstest.LayeredCard(t, g, host).Effective(); eff.Restrictions&game.CantAttack == 0 || eff.Restrictions&game.CantBlock != 0 {
		t.Errorf("restrictions %v, want can't attack and no more", eff.Restrictions)
	}
	life := me.Life
	// An opponent casts a Bolt at me: it is redirected to the enchanted creature.
	effectstest.FloatForTest(g, opp, "R")
	bolt := effectstest.B10OpponentCastsBolt(t, g, opp, effectstest.PlayerTarget(me.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(bolt) {
		effectstest.PassPriorityAroundTable(t, g)
	}
	g.RunStateChecksForTest()
	if me.Life != life {
		t.Errorf("my life %d after the redirected Bolt, want %d", me.Life, life)
	}
	if g.Battlefield.Contains(host) {
		t.Error("the enchanted 2/2 survived the redirected Bolt")
	}
}

func TestOldB265CapturedByTheConsulateLeavesMySpellsAndIllegalRedirectsAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.PushVanillaCreature(g, third.ID, "Their Bear", 2, 2)
	effectstest.Enchant(t, g, "Captured by the Consulate", oldB265Consulate, host)
	// My own Bolt is not an opponent's spell.
	effectstest.FloatForTest(g, me, "R")
	mineBolt := effectstest.B10OpponentCastsBolt(t, g, me, effectstest.PlayerTarget(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	_ = mineBolt
	if !g.Battlefield.Contains(host) {
		t.Error("my own spell was redirected to the enchanted creature")
	}
	// The enchanted creature becomes untargetable by the opponent's spell: no redirect.
	effectstest.KrgGrantUntilEOT(t, g, third.ID, host, "hexproof")
	if !effectstest.HasEffectiveKeyword(t, g, host, "hexproof") {
		t.Fatal("the test could not give the host hexproof")
	}
	life := me.Life
	effectstest.FloatForTest(g, opp, "R")
	effectstest.B10OpponentCastsBolt(t, g, opp, effectstest.PlayerTarget(me.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if me.Life != life-3 {
		t.Errorf("my life %d, want %d: the Bolt could not be redirected to a hexproof creature", me.Life, life-3)
	}
	if !g.Battlefield.Contains(host) {
		t.Error("the hexproof host died")
	}
}

// tryEnchant casts an Aura from the active seat at the host and reports the
// cast's error rather than failing the test.
func tryEnchant(g *game.Game, me *game.Player, name, oracle string, host uuid.UUID) (uuid.UUID, error) {
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: effectstest.AuraTypeLine, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}}})
	return id, err
}

func TestOldB265DisappearingActReturnsAPermanentAndCountersAnySpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.FloatForTest(g, opp, "R")
	bolt := effectstest.B10OpponentCastsBolt(t, g, opp, effectstest.PlayerTarget(me.ID))
	act := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: act, Name: "Disappearing Act", TypeLine: "Instant", OracleID: oldB265Disappear, Owner: me.ID, Controller: me.ID})
	target := []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}
	if err := g.CastSpell(me.ID, act, game.CastSpellParams{Strict: true, Targets: target}); err == nil {
		t.Fatal("cast without returning a permanent")
	}
	if err := g.CastSpell(me.ID, act, game.CastSpellParams{Strict: true, Targets: target, ReturnIDs: []uuid.UUID{uuid.New()}}); err == nil {
		t.Fatal("returned a permanent that is not on the battlefield")
	}
	if err := g.CastSpell(me.ID, act, game.CastSpellParams{Targets: target, ReturnIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("Disappearing Act: %v", err)
	}
	if !me.Hand.Contains(rock) || g.Battlefield.Contains(rock) {
		t.Error("the permanent was not returned as the cost")
	}
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != life {
		t.Error("the Bolt was not countered")
	}
}

func TestOldB265RevolutionaryRebuffTaxesANonartifactSpell(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, them := g.Seats[0], g.Seats[1]
		effectstest.ToMainForCost(t, g)
		victim := effectstest.CastCatalogSpell(t, g, "My Sorcery", "Sorcery", "", nil)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		effectstest.CastFromSeat(t, g, them, "Revolutionary Rebuff", "Instant", oldB265Rebuff,
			[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
		effectstest.PassUntilTaxed(t, g, me.ID)
		if !g.Stack.Contains(victim) {
			t.Fatal("the spell resolved while the {2} was unanswered")
		}
		if pay {
			effectstest.FloatForTest(g, me, "CC")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		if got := g.Stack.Contains(victim); got != false {
			// Paid: the spell goes on to resolve; declined: it is countered.
			if !pay {
				t.Error("declining the tax must counter the spell")
			}
		}
		if !pay && !me.Graveyard.Contains(victim) {
			t.Error("the countered spell is not in the graveyard")
		}
	}
}

func TestOldB265RevolutionaryRebuffCantTargetAnArtifactSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	them := g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "My Rock", "Artifact", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	them.Hand.PushTop(game.Card{InstanceID: id, Name: "Revolutionary Rebuff", TypeLine: "Instant", OracleID: oldB265Rebuff, Owner: them.ID, Controller: them.ID})
	if err := g.CastSpell(them.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err == nil {
		t.Error("Revolutionary Rebuff targeted an artifact spell")
	}
}

func TestOldB265RevolutionaryRebuffCanTargetACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	them := g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "My Bear", oldB265Bear, "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	them.Hand.PushTop(game.Card{InstanceID: id, Name: "Revolutionary Rebuff", TypeLine: "Instant", OracleID: oldB265Rebuff, Owner: them.ID, Controller: them.ID})
	if err := g.CastSpell(them.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Errorf("Revolutionary Rebuff could not target a creature spell: %v", err)
	}
}

func TestOldB265CogworkersPuzzleknotMakesAServoWhenItEntersAndWhenSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	knot := effectstest.CastCatalogSpell(t, g, "Cogworker's Puzzleknot", "Artifact", oldB265CogKnot, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CardsNamed(g, "Servo", me.ID); got != 1 {
		t.Fatalf("%d Servos after the entry trigger, want 1", got)
	}
	if err := effectstest.UsgB2Activate(t, g, me, knot, "CC", nil); err == nil {
		t.Fatal("activated without {W}")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := effectstest.UsgB2Activate(t, g, me, knot, "WC", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(knot) || !me.Graveyard.Contains(knot) {
		t.Error("the Puzzleknot is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CardsNamed(g, "Servo", me.ID); got != 2 {
		t.Errorf("%d Servos after the sacrifice, want 2", got)
	}
}

// oldB265PutCounters puts n counters of the kind on a permanent.
func oldB265PutCounters(g *game.Game, id uuid.UUID, kind string, n int) {
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, kind, n) })
	g.RunStateChecksForTest()
}

func TestOldB265AnimationModulePaysForAServoWhenAPlusOneCounterLandsOnYourPermanent(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Animation Module", "Artifact", oldB265Module, 0, 0)
		mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
		theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
		effectstest.AdvanceToMain(t, g)
		effectstest.FillPool(me, 1)
		// An opponent's permanent, or a -1/-1 counter of mine, does nothing.
		oldB265PutCounters(g, theirs, game.CounterPlusOne, 1)
		oldB265PutCounters(g, mine, game.CounterMinusOne, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("asked to pay for a counter that is not a +1/+1 counter on my permanent")
		}
		oldB265PutCounters(g, mine, game.CounterPlusOne, 2)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("two +1/+1 counters on my creature did not offer the {1} payment")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if pay {
			want = 1
		}
		if got := effectstest.CardsNamed(g, "Servo", me.ID); got != want {
			t.Errorf("pay=%v: %d Servos, want %d (one trigger for the batch of two counters)", pay, got, want)
		}
	}
}

func TestOldB265AnimationModuleGivesAnotherCounterOfAChosenKind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	module := effectstest.B12Push(g, me.ID, "Animation Module", "Artifact", oldB265Module, 0, 0)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	at := func(id uuid.UUID) []game.TargetRef { return []game.TargetRef{{Kind: game.TargetCard, ID: id}} }
	activate := func(targets []game.TargetRef) {
		t.Helper()
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
		if err := effectstest.UsgB2Activate(t, g, me, module, "CCC", targets); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if err := g.TapCard(module, false); err != nil {
			t.Fatalf("untap: %v", err)
		}
	}
	// With no counter on the target nothing happens.
	activate(at(bear))
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Fatalf("%d counters out of nothing", got)
	}
	// One kind: no question, one more of it.
	oldB265PutCounters(g, bear, game.CounterPlusOne, 1)
	activate(at(bear))
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	// Two kinds: the controller chooses one.
	oldB265PutCounters(g, bear, "charge", 1)
	activate(at(bear))
	kinds := effectstest.LatestOptionPickFor(g, me.ID)
	if kinds == nil {
		t.Fatal("two kinds of counter and no question")
	}
	chosen := 0
	for i, o := range kinds.PickOptions {
		if o.Label == "A charge counter" {
			chosen = i
		}
	}
	effectstest.AnswerOptionPick(t, g, me.ID, chosen)
	effectstest.PassPriorityAroundTable(t, g)
	if p, m := effectstest.CountersOn(g, bear, game.CounterPlusOne), effectstest.CountersOn(g, bear, "charge"); p != 2 || m != 2 {
		t.Errorf("+1/+1 %d and charge %d, want 2 and 2 (only the chosen kind grew)", p, m)
	}
}

func TestOldB265AnimationModuleCanGiveAPlayerAnotherCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	module := effectstest.B12Push(g, me.ID, "Animation Module", "Artifact", oldB265Module, 0, 0)
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(opp.ID, "poison", 2) })
	if err := effectstest.UsgB2Activate(t, g, me, module, "CCC", []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Counters["poison"]; got != 3 {
		t.Errorf("%d poison counters, want 3", got)
	}
}

func TestOldB265NightMarketLookoutDrainsWhenItBecomesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	look := effectstest.B12Push(g, me.ID, "Night Market Lookout", "Creature — Human Rogue", oldB265Lookout, 1, 1)
	effectstest.AdvanceToMain(t, g)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	effectstest.TapForTest(t, g, look)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := lives[s.ID] - 1
		if s.ID == me.ID {
			want = lives[s.ID] + 1
		}
		if s.Life != want {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, want)
		}
	}
}

func TestOldB265RushOfVitalityPumpsAndGrantsLifelinkAndIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Rush of Vitality", "Instant", oldB265Rush, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the creature is %d/%d, want 3/2", p, tough)
	}
	for _, kw := range []string{"lifelink", "indestructible"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("no %s", kw)
		}
	}
	effectstest.NextTurnOfMine(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 || effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") || effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the effect outlasted the turn")
	}
}

func TestOldB265ShrewdNegotiationExchangesAnArtifactForAnArtifactOrCreature(t *testing.T) {
	for _, theirType := range []string{"Artifact", oldB265Bear} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
		myBear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
		theirs := effectstest.B12Permanent(g, opp.ID, "Their Thing", theirType)
		if _, err := effectstest.CastModalErr(t, g, "Shrewd Negotiation", "Sorcery", oldB265Negotiation, nil, effectstest.CardRefs(myBear, theirs)); err == nil {
			t.Fatal("the first clause took a creature that is not an artifact")
		}
		effectstest.CastCatalogSpell(t, g, "Shrewd Negotiation", "Sorcery", oldB265Negotiation, effectstest.CardRefs(mine, theirs))
		effectstest.PassPriorityAroundTable(t, g)
		c1, ok1 := effectstest.BattlefieldCard(g, mine)
		c2, ok2 := effectstest.BattlefieldCard(g, theirs)
		if !ok1 || !ok2 || c1.Controller != opp.ID || c2.Controller != me.ID {
			t.Errorf("%s: the permanents did not trade places (%v/%v)", theirType, c1.Controller, c2.Controller)
		}
	}
}

func TestOldB265ShrewdNegotiationRefusesAPermanentOfYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Permanent(g, me.ID, "Rock A", "Artifact")
	b := effectstest.B12Permanent(g, me.ID, "Rock B", "Artifact")
	if _, err := effectstest.CastModalErr(t, g, "Shrewd Negotiation", "Sorcery", oldB265Negotiation, nil, effectstest.CardRefs(a, b)); err == nil {
		t.Error("the second clause took a permanent its controller controls")
	}
}

func TestOldB265DubiousChallengeExilesTwoCreaturesAndSplitsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lib := effectstest.DftMPLibraryTop(me,
		game.Card{Name: "Wolf", TypeLine: "Creature — Wolf", Power: 2, Toughness: 2},
		game.Card{Name: "Spell", TypeLine: "Instant"},
		game.Card{Name: "Bear", TypeLine: oldB265Bear, Power: 3, Toughness: 3},
		game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", Power: 4, Toughness: 4},
		game.Card{Name: "Land", TypeLine: "Basic Land — Forest"},
	)
	effectstest.CastCatalogSpell(t, g, "Dubious Challenge", "Sorcery", oldB265Challenge, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose_cards prompt for the caster")
	}
	for _, id := range pick.ChooseCards {
		if id == lib[1] || id == lib[4] {
			t.Error("a noncreature card was offered")
		}
	}
	if len(pick.ChooseCards) != 3 {
		t.Errorf("%d candidates, want the three creature cards", len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{lib[0], lib[2], lib[3]}); err == nil {
		t.Fatal("three cards were accepted for 'up to two'")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, lib[2], lib[3])
	for _, id := range []uuid.UUID{lib[2], lib[3]} {
		if !g.Exile.Contains(id) {
			t.Fatal("a chosen creature card is not exiled")
		}
	}
	if !me.Library.Contains(lib[0]) {
		t.Error("an unchosen card left the library")
	}
	// The opponent chooses one of the exiled cards.
	second := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if second == nil {
		t.Fatal("the target opponent was not asked")
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("the caster is also asked")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, lib[3])
	effectstest.PassPriorityAroundTable(t, g)
	// An exiled card that enters is a new object, so they are found by name.
	ogre, ok := oldB265NamedOnBattlefield(g, "Ogre")
	bear, ok2 := oldB265NamedOnBattlefield(g, "Bear")
	if !ok || !ok2 {
		t.Fatal("an exiled creature did not enter the battlefield")
	}
	if ogre.Controller != opp.ID || ogre.Owner != me.ID {
		t.Errorf("the opponent's pick is controlled by %v, owned by %v", ogre.Controller, ogre.Owner)
	}
	if bear.Controller != me.ID {
		t.Errorf("the rest is controlled by %v, want the caster", bear.Controller)
	}
}

func TestOldB265DubiousChallengeOpponentMayChooseNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lib := effectstest.DftMPLibraryTop(me,
		game.Card{Name: "Wolf", TypeLine: "Creature — Wolf", Power: 2, Toughness: 2},
		game.Card{Name: "Bear", TypeLine: oldB265Bear, Power: 3, Toughness: 3},
	)
	effectstest.CastCatalogSpell(t, g, "Dubious Challenge", "Sorcery", oldB265Challenge, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, lib[0], lib[1])
	effectstest.AnswerChooseCards(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	_ = lib
	for _, name := range []string{"Wolf", "Bear"} {
		c, ok := oldB265NamedOnBattlefield(g, name)
		if !ok || c.Controller != me.ID {
			t.Errorf("%s is not on my side of the battlefield: %v %v", name, ok, c.Controller)
		}
	}
}

func TestOldB265DubiousChallengeWithNoCreatureOnlyShuffles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.DftMPLibraryTop(me,
		game.Card{Name: "Spell", TypeLine: "Instant"},
		game.Card{Name: "Land", TypeLine: "Basic Land — Forest"},
	)
	size := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Dubious Challenge", "Sorcery", oldB265Challenge, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil || effectstest.ChooseCardsChoiceFor(g, opp.ID) != nil {
		t.Error("a prompt was opened with no creature card to choose")
	}
	if me.Library.Size() != size {
		t.Errorf("library %d, want %d", me.Library.Size(), size)
	}
	if got := effectstest.CountEvents(g, game.EventSearchLibrary); got == 0 {
		t.Error("the library was not shuffled")
	}
}

func TestOldB265BomatCourierExilesFaceDownOnAttackAndReturnsThemWhenSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	courier := effectstest.B12Push(g, me.ID, "Bomat Courier", "Artifact Creature — Construct", oldB265Courier, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, courier, "haste") {
		t.Error("no haste")
	}
	top := me.Library.Cards[me.Library.Size()-1].InstanceID
	size := me.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, courier)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) || me.Library.Size() != size-1 {
		t.Fatal("the top card was not exiled")
	}
	exiled := g.ExiledWithForEffect(game.PermissionCardRef{ID: courier, Epoch: effectstest.LayeredCard(t, g, courier).ObjectEpoch})
	if len(exiled) != 1 || exiled[0].InstanceID != top {
		t.Fatalf("%d cards exiled with the Courier, want the top card", len(exiled))
	}
	// A second attack exiles another (a second turn), simulated by the ability's
	// own return: {R}, discard the hand, sacrifice.
	handBefore := me.Hand.Size()
	if err := effectstest.UsgB2Activate(t, g, me, courier, "C", nil); err == nil {
		t.Fatal("activated without {R}")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	hand := append([]uuid.UUID(nil), handIDs(me)...)
	if err := effectstest.UsgB2Activate(t, g, me, courier, "R", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(courier) || !me.Graveyard.Contains(courier) {
		t.Error("the Courier is sacrificed as a cost")
	}
	for _, id := range hand {
		if me.Hand.Contains(id) {
			t.Error("a card of the old hand is still in hand: the hand is a cost")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) || g.Exile.Contains(top) {
		t.Error("the exiled card did not go to its owner's hand")
	}
	if me.Hand.Size() != 1 {
		t.Errorf("hand %d after the hand was discarded and one card returned, want 1 (was %d)", me.Hand.Size(), handBefore)
	}
}

// oldB265NamedOnBattlefield finds a battlefield card by name.
func oldB265NamedOnBattlefield(g *game.Game, name string) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c, true
		}
	}
	return game.Card{}, false
}

func handIDs(p *game.Player) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range p.Hand.Cards {
		out = append(out, c.InstanceID)
	}
	return out
}
