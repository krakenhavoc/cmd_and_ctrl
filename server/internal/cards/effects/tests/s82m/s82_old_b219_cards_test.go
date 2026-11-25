package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b219_cards_test.go — S82 OLD batch 219 (#5531, Rivals of Ixalan:
// counters, life, planeswalkers, enters and dies triggers, combat tricks).
// Each test plays the real catalog entry.

const (
	oldB219ForeHeralds = "56e2bdcf-d1d8-4c56-af3a-388cd42353e6"
	oldB219Tetzimoc    = "7565c101-a479-4812-9d18-63918084bfca"
	oldB219Paladin     = "f44dc71d-d08a-45ed-ac7c-f6d1ef811d96"
	oldB219Boarder     = "cffcd887-660c-4646-9333-bcc8549f6609"
	oldB219Conquist    = "d3839a3f-3018-4a8d-b298-3f5b5fc70cd1"
	oldB219Arterial    = "f5e2fb92-4678-410c-8c83-42aead9d29d3"
	oldB219Angrath     = "2737a78c-77ed-4bd0-ac0d-39ca2d816e83"
	oldB219Vraska      = "83045a9c-57da-411f-80aa-56178d14856b"
	oldB219Guardian    = "58afb897-4d57-4b53-a5c3-b532cb3d5180"
	oldB219Poisoner    = "8150cc84-3f9f-470d-a0f9-ee6d9dab13da"
	oldB219Ambusher    = "85791f11-44ab-4405-b284-1d91c7e0afbb"
	oldB219Bravado     = "e85421f3-37f6-49ee-9fc4-4424fd4f5236"
	oldB219Bear        = "Creature — Bear"
)

func TestOldB219EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB219ForeHeralds:     "Forerunner of the Heralds",
		oldB219Tetzimoc:        "Tetzimoc, Primal Death",
		oldB219Paladin:         "Paladin of Atonement",
		oldB219Boarder:         "Fathom Fleet Boarder",
		oldB219Conquist:        "Vraska's Conquistador",
		oldB219Arterial:        "Arterial Flow",
		oldB219Angrath:         "Angrath, Minotaur Pirate",
		oldB219Vraska:          "Vraska, Scheming Gorgon",
		oldB219Guardian:        "Golden Guardian",
		oldB219Guardian + "#1": "Gold-Forge Garrison",
		oldB219Poisoner:        "Dire Fleet Poisoner",
		oldB219Ambusher:        "Angrath's Ambusher",
		oldB219Bravado:         "Buccaneer's Bravado",
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

func oldB219Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB219Walker(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, loyalty int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Counters: map[string]int{game.CounterLoyalty: loyalty}, Owner: owner, Controller: owner})
}

// --- Forerunner of the Heralds ------------------------------------------------

func TestOldB219ForerunnerOfTheHeraldsTutorsAMerfolkAndGrowsForAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB219Seats(g)
	match := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Match", TypeLine: "Creature — Merfolk Wizard"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Other", TypeLine: "Creature — Goblin"})
	heralds := effectstest.CastHoldCreature(t, g, "Forerunner of the Heralds", "Creature — Merfolk Scout", oldB219ForeHeralds, 3, 2)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt after saying yes")
	}
	if effectstest.SearchOptionNamed(g, c, "Other") != uuid.Nil {
		t.Error("a non-Merfolk was offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, match)
	effectstest.PassPriorityAroundTable(t, g)
	if top, err := me.Library.Top(); err != nil || top.InstanceID != match {
		t.Fatal("the Merfolk is not on top of the library")
	} else if !top.IsKnownTo(opp.ID) {
		t.Error("the tutored card was not revealed")
	}
	if n := effectstest.AangCounterCount(g, heralds, game.CounterPlusOne); n != 0 {
		t.Fatalf("the Forerunner put a counter on itself for entering: %d", n)
	}
	effectstest.CastHoldCreature(t, g, "Merrow", "Creature — Merfolk", "", 1, 1)
	if n := effectstest.AangCounterCount(g, heralds, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after another Merfolk entered, want 1", n)
	}
	effectstest.CastHoldCreature(t, g, "Goblin", "Creature — Goblin", "", 1, 1)
	if n := effectstest.AangCounterCount(g, heralds, game.CounterPlusOne); n != 1 {
		t.Errorf("a Goblin entering changed the counters to %d", n)
	}
}

// --- Tetzimoc, Primal Death -----------------------------------------------------

func TestOldB219TetzimocPutsPreyCountersFromHandOnYourTurnAndDestroysTheirs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB219Seats(g)
	marked := effectstest.B12Creature(g, opp.ID, "Marked", oldB219Bear, 2, 2)
	unmarked := effectstest.B12Creature(g, opp.ID, "Unmarked", oldB219Bear, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB219Bear, 2, 2)
	tet := effectstest.HandCardForTest(me, "Tetzimoc, Primal Death", "Legendary Creature — Elder Dinosaur", oldB219Tetzimoc)
	effectstest.AdvanceToMain(t, g)
	effectstest.AddPool(t, g, me.ID, "{B}{B}{B}")
	effectstest.B16Activate(t, g, me.ID, tet, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(marked)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16Activate(t, g, me.ID, tet, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(mine)})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.AangCounterCount(g, marked, "prey"); n != 1 {
		t.Fatalf("%d prey counters on the target, want 1", n)
	}
	if z, _ := effectstest.CardWhere(g, me, tet); z != "hand" {
		t.Errorf("revealing left the card in %s, want the hand", z)
	}
	// Cast it: only the opponent's creature with a prey counter dies.
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	effectstest.CastCatalogSpell(t, g, "Tetzimoc, Primal Death", "Legendary Creature — Elder Dinosaur", oldB219Tetzimoc, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(marked) {
		t.Error("the opponent's creature with a prey counter survived")
	}
	if !g.Battlefield.Contains(unmarked) {
		t.Error("an opponent's creature without a prey counter was destroyed")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own creature with a prey counter was destroyed")
	}
}

func TestOldB219TetzimocCantBeRevealedOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB219Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB219Bear, 2, 2)
	tet := effectstest.HandCardForTest(me, "Tetzimoc, Primal Death", "Legendary Creature — Elder Dinosaur", oldB219Tetzimoc)
	oppSeat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	for i := 0; i < 400 && !(g.Turn.ActiveSeat == oppSeat && g.Turn.Step == game.StepPrecombatMain); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AddPool(t, g, me.ID, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, tet, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Error("the prey ability was activated on an opponent's turn")
	}
}

// --- Paladin of Atonement -------------------------------------------------------

func TestOldB219PaladinGrowsAtEachUpkeepAfterYouLostLifeAndGainsItsToughnessWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB219Seats(g)
	pal := effectstest.B12Push(g, me.ID, "Paladin of Atonement", "Creature — Vampire Knight", oldB219Paladin, 1, 1)
	// No life lost last turn: an upkeep changes nothing.
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: g.Seats[1].ID}) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.AangCounterCount(g, pal, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters without a life loss last turn", n)
	}
	// Walk through a turn in which I lose life, then the next upkeep (anyone's).
	effectstest.ToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -3); err != nil {
			t.Fatal(err)
		}
	})
	for i := 0; i < 200 && effectstest.AangCounterCount(g, pal, game.CounterPlusOne) == 0; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.AangCounterCount(g, pal, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d counters after the next upkeep, want 1", n)
	}
	life := me.Life
	effectstest.P1Destroy(t, g, pal)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d (its toughness was 2 with the counter)", me.Life, life+2)
	}
}

// --- Fathom Fleet Boarder -------------------------------------------------------

func TestOldB219FathomFleetBoarderCostsTwoLifeUnlessAnotherPirate(t *testing.T) {
	for _, tc := range []struct {
		name     string
		pirate   bool
		wantLoss int
	}{{"alone", false, 2}, {"with another Pirate", true, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, _ := oldB219Seats(g)
			if tc.pirate {
				effectstest.B12Push(g, me.ID, "Pirate", "Creature — Human Pirate", "", 2, 2)
			}
			life := me.Life
			effectstest.CastHoldCreature(t, g, "Fathom Fleet Boarder", "Creature — Orc Pirate", oldB219Boarder, 3, 3)
			effectstest.PassPriorityAroundTable(t, g)
			if me.Life != life-tc.wantLoss {
				t.Errorf("life %d, want %d", me.Life, life-tc.wantLoss)
			}
		})
	}
}

// --- Vraska's Conquistador ------------------------------------------------------

func TestOldB219VraskasConquistadorDrainsOnAttackOnlyWithAVraskaWalker(t *testing.T) {
	for _, withVraska := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := oldB219Seats(g)
		con := effectstest.B12Push(g, me.ID, "Vraska's Conquistador", "Creature — Vampire Soldier", oldB219Conquist, 2, 1)
		if withVraska {
			oldB219Walker(g, me.ID, "Vraska, Test", "Legendary Planeswalker — Vraska", "", 5)
		} else {
			oldB219Walker(g, me.ID, "Other Walker", "Legendary Planeswalker — Angrath", "", 5)
		}
		oppLife, myLife := opp.Life, me.Life
		effectstest.DeclareAttack(t, g, opp.ID, con)
		effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if withVraska {
			if opp.Life != oppLife-2 || me.Life != myLife+2 {
				t.Errorf("with a Vraska: opponent %d (was %d), me %d (was %d); want -2 and +2", opp.Life, oppLife, me.Life, myLife)
			}
		} else if opp.Life != oppLife || me.Life != myLife {
			t.Errorf("without a Vraska the Conquistador drained: opponent %d (was %d), me %d (was %d)", opp.Life, oppLife, me.Life, myLife)
		}
	}
}

func TestOldB219VraskasConquistadorDrainsWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB219Seats(g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB219Bear, 2, 2)
	con := effectstest.B12Push(g, opp.ID, "Vraska's Conquistador", "Creature — Vampire Soldier", oldB219Conquist, 2, 1)
	oldB219Walker(g, opp.ID, "Vraska, Test", "Legendary Planeswalker — Vraska", "", 5)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(con, attacker); err != nil {
		t.Fatal(err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 40-2 {
		t.Errorf("the attacking player has %d life, want 38", me.Life)
	}
}

// --- Arterial Flow --------------------------------------------------------------

func TestOldB219ArterialFlowDiscardsTwoAndDrainsWithAVampire(t *testing.T) {
	for _, vampire := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, _ := oldB219Seats(g)
		if vampire {
			effectstest.B12Push(g, me.ID, "Bat", "Creature — Vampire", "", 1, 1)
		}
		before := map[uuid.UUID]int{}
		for _, p := range g.Seats {
			p.Hand.Cards = nil
			for i := 0; i < 3; i++ {
				p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Held", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
			}
			before[p.ID] = p.Life
		}
		effectstest.CastCatalogSpell(t, g, "Arterial Flow", "Sorcery", oldB219Arterial, nil)
		effectstest.PassPriorityAroundTable(t, g)
		for i := 0; i < 8; i++ {
			for _, p := range g.Seats {
				if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, p.ID); c != nil && len(c.ChooseCards) >= 2 {
					if err := g.ResolveChooseCards(c.ID, p.ID, c.ChooseCards[:2]); err != nil {
						t.Fatal(err)
					}
				}
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
		for _, p := range g.Seats {
			if p.ID == me.ID {
				continue
			}
			if got := p.Hand.Size(); got != 1 {
				t.Errorf("an opponent has %d cards in hand, want 1", got)
			}
			want := before[p.ID]
			if vampire {
				want -= 2
			}
			if p.Life != want {
				t.Errorf("vampire=%v: an opponent has %d life, want %d", vampire, p.Life, want)
			}
		}
		wantMine := before[me.ID]
		if vampire {
			wantMine += 2
		}
		if me.Life != wantMine {
			t.Errorf("vampire=%v: my life %d, want %d", vampire, me.Life, wantMine)
		}
	}
}

// --- Angrath, Minotaur Pirate ---------------------------------------------------

func TestOldB219AngrathPlusTwoDamagesTheOpponentAndTheirCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB219Seats(g)
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	angrath := oldB219Walker(g, me.ID, "Angrath, Minotaur Pirate", "Legendary Planeswalker — Angrath", oldB219Angrath, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB219Bear, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB219Bear, 3, 3)
	other := effectstest.B12Creature(g, third.ID, "Other", oldB219Bear, 3, 3)
	before, thirdLife := opp.Life, third.Life
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, angrath, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 || third.Life != thirdLife {
		t.Errorf("life: target %d (was %d), other opponent %d (was %d)", opp.Life, before, third.Life, thirdLife)
	}
	if got := effectstest.DamageOn(t, g, theirs); got != 1 {
		t.Errorf("the target's creature took %d, want 1", got)
	}
	if effectstest.DamageOn(t, g, mine) != 0 || effectstest.DamageOn(t, g, other) != 0 {
		t.Error("a creature the target does not control was damaged")
	}
	if n := effectstest.LoyaltyOf(g, angrath); n != 7 {
		t.Errorf("loyalty %d, want 7", n)
	}
}

func TestOldB219AngrathPlusTwoOnAPlaneswalkerHitsItsControllersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB219Seats(g)
	angrath := oldB219Walker(g, me.ID, "Angrath, Minotaur Pirate", "Legendary Planeswalker — Angrath", oldB219Angrath, 5)
	walker := oldB219Walker(g, opp.ID, "Their Walker", "Legendary Planeswalker — Test", "", 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB219Bear, 3, 3)
	before := opp.Life
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, angrath, 0, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(walker)})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.LoyaltyOf(g, walker); n != 3 {
		t.Errorf("walker loyalty %d, want 3", n)
	}
	if got := effectstest.DamageOn(t, g, theirs); got != 1 {
		t.Errorf("the walker controller's creature took %d, want 1", got)
	}
	if opp.Life != before {
		t.Errorf("the walker's controller lost life: %d (was %d)", opp.Life, before)
	}
}

func TestOldB219AngrathMinusThreeReturnsAPirateCardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB219Seats(g)
	angrath := oldB219Walker(g, me.ID, "Angrath, Minotaur Pirate", "Legendary Planeswalker — Angrath", oldB219Angrath, 5)
	pirate := effectstest.PushGraveyardPermanent(me, "Pirate", "Creature — Human Pirate", "{2}")
	bear := effectstest.PushGraveyardPermanent(me, "Bear", "Creature — Bear", "{2}")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, angrath, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(bear)}); err == nil {
		t.Fatal("a non-Pirate card was a legal target")
	}
	effectstest.S82Activate(t, g, me.ID, angrath, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(pirate)})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(pirate) {
		t.Error("the Pirate did not return to the battlefield")
	}
	if n := effectstest.LoyaltyOf(g, angrath); n != 2 {
		t.Errorf("loyalty %d, want 2", n)
	}
}

func TestOldB219AngrathUltimateDestroysTheirCreaturesAndDealsTheirTotalPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB219Seats(g)
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	angrath := oldB219Walker(g, me.ID, "Angrath, Minotaur Pirate", "Legendary Planeswalker — Angrath", oldB219Angrath, 11)
	a := effectstest.B12Creature(g, opp.ID, "A", oldB219Bear, 3, 3)
	b := effectstest.B12Creature(g, opp.ID, "B", oldB219Bear, 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB219Bear, 5, 5)
	other := effectstest.B12Creature(g, third.ID, "Other", oldB219Bear, 2, 2)
	before := opp.Life
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, angrath, 2, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("an opponent's creature survived")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(other) {
		t.Error("a creature the target does not control was destroyed")
	}
	if opp.Life != before-7 {
		t.Errorf("opponent life %d, want %d (3 + 4 power)", opp.Life, before-7)
	}
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(angrath) {
		t.Error("Angrath survived paying all his loyalty")
	}
}

// --- Vraska, Scheming Gorgon ----------------------------------------------------

func TestOldB219VraskaPlusTwoPumpsYourCreaturesAndMinusThreeDestroys(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB219Seats(g)
	vraska := oldB219Walker(g, me.ID, "Vraska, Scheming Gorgon", "Legendary Planeswalker — Vraska", oldB219Vraska, 5)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB219Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB219Bear, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, vraska, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, mine); p != 3 {
		t.Errorf("my creature has power %d, want 3", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("their creature has power %d, want 2", p)
	}
	if n := effectstest.LoyaltyOf(g, vraska); n != 7 {
		t.Fatalf("loyalty %d, want 7", n)
	}
	// Next activation needs another turn; use a second walker for −3.
	v2 := oldB219Walker(g, me.ID, "Vraska, Second Copy", "Legendary Planeswalker — Vraska", oldB219Vraska, 5)
	effectstest.S82Activate(t, g, me.ID, v2, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the target creature survived the -3")
	}
	if n := effectstest.LoyaltyOf(g, v2); n != 2 {
		t.Errorf("loyalty %d, want 2", n)
	}
}

func TestOldB219VraskaUltimateGivesDeathtouchAndALoseTheGameTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB219Seats(g)
	vraska := oldB219Walker(g, me.ID, "Vraska, Scheming Gorgon", "Legendary Planeswalker — Vraska", oldB219Vraska, 10)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB219Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB219Bear, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, vraska, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, mine, "deathtouch") {
		t.Error("my creature did not gain deathtouch")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "deathtouch") {
		t.Error("their creature gained deathtouch")
	}
	// Noncombat damage to an opponent counts too.
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(mine, opp.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Eliminated {
		t.Error("the damaged opponent did not lose the game")
	}
}

// --- Golden Guardian // Gold-Forge Garrison ----------------------------------------

func TestOldB219GoldenGuardianFightsAndReturnsTransformedWhenItDiesThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB219Seats(g)
	guard := effectstest.VowB15Put(t, g, effectstest.VowB15Row(oldB219Guardian, "Golden Guardian", "Artifact Creature — Golem", "{4}", "4", "4",
		"Gold-Forge Garrison", "Land", "", "", nil), me)
	big := effectstest.B12Creature(g, me.ID, "Big", oldB219Bear, 6, 6)
	small := effectstest.B12Creature(g, me.ID, "Small", oldB219Bear, 1, 1)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	foe := effectstest.B12Creature(g, opp.ID, "Foe", oldB219Bear, 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.AddPool(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, guard, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(foe)}); err == nil {
		t.Fatal("it fought a creature its controller does not control")
	}
	if err := g.ActivateCatalogAbility(me.ID, guard, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(guard)}); err == nil {
		t.Fatal("it fought itself")
	}
	effectstest.B16Activate(t, g, me.ID, guard, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(big)})
	effectstest.PassPriorityAroundTable(t, g)
	// 4 damage to Big, 6 damage to the Guardian: it dies this turn and returns.
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, guard)
	if !ok {
		t.Fatalf("the Guardian did not return to the battlefield; it is in %s", func() string { z, _ := effectstest.CardWhere(g, me, guard); return z }())
	}
	if card.Controller != me.ID || !card.IsLand() {
		t.Errorf("it returned as %q controlled by %v, want the Garrison land under my control", card.Name, card.Controller)
	}
	_ = small
}

func TestOldB219GoldForgeGarrisonMakesTwoManaOfOneColorAndAGolem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB219Seats(g)
	effectstest.AdvanceToMain(t, g)
	garrison := effectstest.B12Push(g, me.ID, "Gold-Forge Garrison", "Land", oldB219Guardian+"#1", 0, 0)
	if err := g.ActivateManaAbility(me.ID, garrison, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if me.ManaPool.CountColor("R") != 2 || me.ManaPool.Total() != 2 {
		t.Errorf("pool %v, want {R}{R}", me.ManaPool)
	}
	garrison2 := effectstest.B12Push(g, me.ID, "Gold-Forge Garrison", "Land", oldB219Guardian+"#1", 0, 0)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, garrison2, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	golems := effectstest.S46Tokens(g, me.ID, "Golem")
	if len(golems) != 1 || golems[0].Power != 4 || golems[0].Toughness != 4 || len(golems[0].Colors) != 0 {
		t.Errorf("tokens %+v, want one colorless 4/4 Golem", golems)
	}
}

// --- Dire Fleet Poisoner, Angrath's Ambusher, Buccaneer's Bravado ------------------

func TestOldB219DireFleetPoisonerPumpsAnAttackingPirate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB219Seats(g)
	pirate := effectstest.B12Creature(g, me.ID, "Pirate", "Creature — Human Pirate", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB219Bear, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, pirate, bear)
	poisoner := effectstest.HandCardForTest(me, "Dire Fleet Poisoner", "Creature — Human Pirate", oldB219Poisoner)
	if err := g.CastSpell(me.ID, poisoner, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, pirate)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, pirate); p != 3 || tough != 3 {
		t.Errorf("the Pirate is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, pirate, "deathtouch") {
		t.Error("the Pirate did not gain deathtouch")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 {
		t.Errorf("the non-Pirate Bear changed: power %d", p)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, poisoner, "flash") {
		t.Error("no flash")
	}
}

func TestOldB219AngrathsAmbusherGetsPlusTwoWhileYouControlAnAngrath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB219Seats(g)
	amb := effectstest.B12Push(g, me.ID, "Angrath's Ambusher", "Creature — Orc Pirate", oldB219Ambusher, 2, 3)
	if p, _ := effectstest.EffectivePTOf(t, g, amb); p != 2 {
		t.Fatalf("power %d without an Angrath, want 2", p)
	}
	oldB219Walker(g, opp.ID, "Their Angrath", "Legendary Planeswalker — Angrath", "", 5)
	if p, _ := effectstest.EffectivePTOf(t, g, amb); p != 2 {
		t.Errorf("power %d with only an opponent's Angrath, want 2", p)
	}
	oldB219Walker(g, me.ID, "My Angrath", "Legendary Planeswalker — Angrath", "", 5)
	if p, tough := effectstest.EffectivePTOf(t, g, amb); p != 4 || tough != 3 {
		t.Errorf("it is %d/%d with my Angrath, want 4/3", p, tough)
	}
}

func TestOldB219BuccaneersBravadoModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB219Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB219Bear, 2, 2)
	pirate := effectstest.B12Creature(g, me.ID, "Pirate", "Creature — Human Pirate", 2, 2)
	effectstest.CastCatalogSpellWithModesTargets(t, g, oldB219Bravado, []int{0}, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "first strike") {
		t.Error("the Bear did not gain first strike")
	}
	effectstest.CastCatalogSpellWithModesTargets(t, g, oldB219Bravado, []int{1}, effectstest.CardRefs(pirate))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, pirate); p != 3 || tough != 3 {
		t.Errorf("the Pirate is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, pirate, "double strike") {
		t.Error("the Pirate did not gain double strike")
	}
}
