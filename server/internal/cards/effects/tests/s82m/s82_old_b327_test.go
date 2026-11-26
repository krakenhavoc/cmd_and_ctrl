package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b327_test.go — Dragons of Tarkir, batch 327 (#5551): the enters
// and dies triggers. Each test plays the real catalog entry through the
// engine and asserts what its text says.

const (
	oldB327Regent    = "0cf18002-1bbe-40b3-ae15-fa51e61e9391"
	oldB327Custodian = "0d060d19-9c8b-4831-a9be-aa2c98b6858a"
	oldB327Sibsig    = "4880c976-b417-48e4-98a2-9e5a8da0d758"
	oldB327Marauder  = "1a0162b9-7a7a-4ed9-9a70-189433932449"
	oldB327Foremost  = "581c2fb6-fcce-42c1-b198-1f8d62e862b1"
	oldB327Sovereign = "cb58055f-8190-4bc7-ad32-844ca66fd2b0"
	oldB327Warkite   = "0df3444c-ae29-4fcc-bf46-4edf52b90e30"
)

func TestOldB327CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB327Regent: "Icefall Regent", oldB327Custodian: "Custodian of the Trove", oldB327Sibsig: "Sibsig Icebreakers",
		oldB327Marauder: "Boltwing Marauder", oldB327Foremost: "Arashin Foremost", oldB327Sovereign: "Arashin Sovereign",
		oldB327Warkite: "Swift Warkite",
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

func TestOldB327PrintedKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, typeLine, oracle string
		kws                    []string
	}{
		{"Icefall Regent", "Creature — Dragon", oldB327Regent, []string{"flying"}},
		{"Boltwing Marauder", "Creature — Dragon", oldB327Marauder, []string{"flying"}},
		{"Arashin Foremost", "Creature — Human Warrior", oldB327Foremost, []string{"double strike"}},
		{"Arashin Sovereign", "Creature — Dragon", oldB327Sovereign, []string{"flying"}},
		{"Swift Warkite", "Creature — Dragon", oldB327Warkite, []string{"flying"}},
	} {
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
		for _, kw := range tc.kws {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
}

func TestOldB327IcefallRegentTapsAndHoldsAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Target")
	mine := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Mine")
	regent := effectstest.CastHoldCreature(t, g, "Icefall Regent", "Creature — Dragon", oldB327Regent, 4, 3)
	opts := effectstest.PickTargetPrompt(t, g).PickTargetCards
	if len(opts) != 1 || opts[0] != target {
		t.Fatalf("the Regent's targets are %v, want only the opponent's creature %v (not %v)", opts, target, mine)
	}
	effectstest.AnswerPickTarget(t, g, target)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, target) || effectstest.HoldCount(t, g, target) != 1 {
		t.Fatalf("after the trigger: tapped %v holds %d, want tapped with one hold", effectstest.TappedForTest(t, g, target), effectstest.HoldCount(t, g, target))
	}
	for round := 0; round < 2; round++ {
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		if !effectstest.TappedForTest(t, g, target) {
			t.Fatalf("round %d: the held creature untapped while the Regent is under your control", round+1)
		}
		effectstest.AdvanceToUpkeepOf(t, g, 0)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, regent); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TappedForTest(t, g, target) {
		t.Error("the creature stayed tapped after the Regent left the battlefield")
	}
}

func TestOldB327IcefallRegentTaxesOpponentsSpellsThatTargetIt(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	regent := effectstest.B12Push(g, me.ID, "Icefall Regent", "Creature — Dragon", oldB327Regent, 4, 3)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < len(g.Seats); i++ {
		if g.Turn.ActiveSeat == 1 {
			break
		}
		effectstest.EndTurn(t, g)
	}
	effectstest.AdvanceToMain(t, g)
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: effectstest.LightningBoltOracle, Owner: opp.ID, Controller: opp.ID})
	effectstest.FlashIfFund(g, opp, "R")
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(regent)}); err == nil {
		t.Fatal("cast a {R} spell at the Regent with only {R}")
	}
	effectstest.FlashIfFund(g, opp, "R", "C", "C")
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(regent)}); err != nil {
		t.Fatalf("with {R}{C}{C}: %v", err)
	}
}

func TestOldB327CustodianOfTheTroveEntersTappedWithDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	card := game.Card{InstanceID: uuid.New(), Name: "Custodian of the Trove", TypeLine: "Artifact Creature — Golem", OracleID: oldB327Custodian, Power: 2, Toughness: 5, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(card.InstanceID) {
		t.Fatal("the Custodian did not enter")
	}
	if !effectstest.TappedForTest(t, g, card.InstanceID) {
		t.Error("the Custodian did not enter tapped")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, card.InstanceID, "defender") {
		t.Error("the Custodian has no defender")
	}
}

func TestOldB327SibsigIcebreakersMakesEachPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mineHand := effectstest.PutInHand(me, game.Card{Name: "My Spare", TypeLine: "Sorcery"})
	theirHand := effectstest.PutInHand(opp, game.Card{Name: "Their Spare", TypeLine: "Sorcery"})
	effectstest.CastHoldCreature(t, g, "Sibsig Icebreakers", "Creature — Zombie", oldB327Sibsig, 2, 3)
	for _, tc := range []struct {
		p    *game.Player
		card uuid.UUID
	}{{me, mineHand}, {opp, theirHand}} {
		if c := effectstest.DiscardChoiceFor(g, tc.p.ID); c != nil {
			effectstest.AnswerDiscard(t, g, tc.p.ID, tc.card)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(mineHand) == false || opp.Graveyard.Contains(theirHand) == false {
		t.Errorf("each player should have discarded: me %v, opponent %v", me.Graveyard.Contains(mineHand), opp.Graveyard.Contains(theirHand))
	}
}

func TestOldB327BoltwingMarauderPumpsATargetWhenAnotherCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Boltwing Marauder", "Creature — Dragon", oldB327Marauder, 5, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreatureNoResolve(t, g, "Newcomer", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PtOf(t, g, theirs); p != 4 || tou != 2 {
		t.Errorf("the targeted creature is %d/%d, want 4/2", p, tou)
	}
}

func TestOldB327BoltwingMarauderIgnoresItsOwnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Boltwing Marauder", "Creature — Dragon", oldB327Marauder, 5, 4)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID); c != nil {
		t.Error("the Marauder asked for a target as it entered itself")
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Errorf("the Bear has %d power, want 2", p)
	}
}

func TestOldB327ArashinForemostGivesAnotherWarriorDoubleStrikeWhenItEntersAndAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	warrior := effectstest.B12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	foremost := effectstest.CastHoldCreatureNoResolve(t, g, "Arashin Foremost", "Creature — Human Warrior", oldB327Foremost, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	opts := effectstest.PickTargetPrompt(t, g).PickTargetCards
	if len(opts) != 1 || opts[0] != warrior {
		t.Fatalf("targets %v, want only the other Warrior %v (not the Bear %v or the Foremost %v)", opts, warrior, bear, foremost)
	}
	effectstest.AnswerPickTarget(t, g, warrior)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, warrior, "double strike") {
		t.Error("the Warrior did not gain double strike on entry")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, warrior, "double strike") {
		t.Error("double strike lasted past end of turn")
	}
	_ = opp
}

func TestOldB327ArashinForemostTriggersOnAttackToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	foremost := effectstest.B12Push(g, me.ID, "Arashin Foremost", "Creature — Human Warrior", oldB327Foremost, 2, 2)
	warrior := effectstest.B12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(foremost, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerPickTarget(t, g, warrior)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, warrior, "double strike") {
		t.Error("the Warrior did not gain double strike when the Foremost attacked")
	}
}

func TestOldB327ArashinSovereignMayGoToTheTopOrBottomOfItsOwnersLibrary(t *testing.T) {
	for _, tc := range []string{"decline", "top", "bottom"} {
		t.Run(tc, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.SeedLibrary(me, "A", "B", "C")
			sov := effectstest.B12Push(g, me.ID, "Arashin Sovereign", "Creature — Dragon", oldB327Sovereign, 6, 6)
			effectstest.B27Kill(g, sov)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, tc != "decline")
			if tc != "decline" {
				ask := effectstest.PutInLibraryChoiceFor(g, me.ID)
				if ask == nil {
					t.Fatal("the top-or-bottom choice was not asked")
				}
				var err error
				if tc == "top" {
					err = g.ResolvePutInLibrary(ask.ID, me.ID, nil, []uuid.UUID{sov})
				} else {
					err = g.ResolvePutInLibrary(ask.ID, me.ID, []uuid.UUID{sov}, nil)
				}
				if err != nil {
					t.Fatalf("ResolvePutInLibrary: %v", err)
				}
			}
			effectstest.PassPriorityAroundTable(t, g)
			cards := me.Library.Cards
			switch tc {
			case "decline":
				if !me.Graveyard.Contains(sov) {
					t.Error("declining should leave the Sovereign in the graveyard")
				}
			case "top":
				if len(cards) == 0 || cards[len(cards)-1].InstanceID != sov || me.Graveyard.Contains(sov) {
					t.Error("the Sovereign is not on top of its owner's library")
				}
			case "bottom":
				if len(cards) == 0 || cards[0].InstanceID != sov || me.Graveyard.Contains(sov) {
					t.Error("the Sovereign is not on the bottom of its owner's library")
				}
			}
		})
	}
}

// oldB327Warkite casts the Warkite and resolves the spell, leaving its
// trigger's pick open when there is a legal card to put.
func oldB327CastWarkite(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.CastHoldCreatureNoResolve(t, g, "Swift Warkite", "Creature — Dragon", oldB327Warkite, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB327SwiftWarkitePutsACheapCreatureFromHandWithHasteAndReturnsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cheap := effectstest.PutInHand(me, game.Card{Name: "Cheap", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2})
	pricey := effectstest.PutInHand(me, game.Card{Name: "Pricey", TypeLine: "Creature — Bear", ManaCost: "{3}{G}", Power: 4, Toughness: 4})
	oldB327CastWarkite(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the Warkite asked for no card")
	}
	for _, id := range pick.ChooseCards {
		if id == pricey {
			t.Error("a mana value 4 card was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Fatal("the card did not enter")
	}
	if !effectstest.HasEffectiveKeyword(t, g, cheap, "haste") {
		t.Error("the creature has no haste")
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) || !me.Hand.Contains(cheap) {
		t.Errorf("at the end step the creature should return to hand: on battlefield %v, in hand %v", g.Battlefield.Contains(cheap), me.Hand.Contains(cheap))
	}
}

func TestOldB327SwiftWarkiteTakesACreatureFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dead := effectstest.PushGraveyardPermanent(me, "Dead", "Creature — Bear", "{1}")
	oldB327CastWarkite(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) || !effectstest.HasEffectiveKeyword(t, g, dead, "haste") {
		t.Fatal("the graveyard creature did not enter with haste")
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dead) || !me.Hand.Contains(dead) {
		t.Error("the graveyard creature should end up in hand at the end step")
	}
}

func TestOldB327SwiftWarkiteMayDeclineAndIgnoresPriceyCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cheap := effectstest.PutInHand(me, game.Card{Name: "Cheap", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2})
	oldB327CastWarkite(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) || !me.Hand.Contains(cheap) {
		t.Error("declining should leave the creature in hand")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.PutInHand(me2, game.Card{Name: "Pricey", TypeLine: "Creature — Bear", ManaCost: "{3}{G}", Power: 4, Toughness: 4})
	oldB327CastWarkite(t, g2)
	if effectstest.ChooseCardsChoiceFor(g2, me2.ID) != nil {
		t.Error("the Warkite asked although no creature card has mana value 3 or less")
	}
}
