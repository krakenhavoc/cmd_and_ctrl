package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b456_cards_test.go — Commander 2011, enters and dies triggers, clash
// and join forces, S82 older sets, batch 456 (#5574). Each test plays the real
// catalog entry.

const (
	oldB456Archangel = "19e302f8-6ddb-4a43-98bc-ee71a456ac5a"
	oldB456Hag       = "264e60a3-3191-497d-ab06-1411e668b4aa"
	oldB456Plotter   = "3f85b92d-7212-4fc7-a5e8-6350e1acfdd3"
	oldB456Malfegor  = "52e1546a-63eb-48d7-acc4-81b912aabc90"
	oldB456Pollen    = "1f64d70d-ea38-4419-be91-8b68aab3401e"
	oldB456Stroke    = "314a9604-34ec-4529-bd14-7b63be28669f"
	oldB456LashOut   = "398d57bb-4653-4652-aa42-7cea4317be0a"
	oldB456Whelm     = "7f527852-bf7d-43a3-be0b-784c73ed88aa"
	oldB456Dragon    = "87354a7b-01be-419e-b383-4c28e12a2f2b"
	oldB456Alliance  = "be2a4595-4378-4af8-8368-e3b2bb3cb5bb"
	oldB456Trauma    = "ef9b7588-9bf0-4632-ae5d-1dfebc4d88eb"
	oldB456Bear      = "Creature — Bear"
)

func TestOldB456EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB456Archangel: "Archangel of Strife",
		oldB456Hag:       "Desecrator Hag",
		oldB456Plotter:   "Vedalken Plotter",
		oldB456Malfegor:  "Malfegor",
		oldB456Pollen:    "Pollen Lullaby",
		oldB456Stroke:    "Scattering Stroke",
		oldB456LashOut:   "Lash Out",
		oldB456Whelm:     "Whirlpool Whelm",
		oldB456Dragon:    "Mana-Charged Dragon",
		oldB456Alliance:  "Alliance of Arms",
		oldB456Trauma:    "Shared Trauma",
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

func oldB456Count(g *game.Game, owner uuid.UUID, name string, token bool) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == name && c.IsToken() == token {
			n++
		}
	}
	return n
}

// oldB456AnswerPayments answers every open payment prompt, paying amounts[chooser]
// from each seat that has an entry and nothing from the rest.
func oldB456AnswerPayments(t *testing.T, g *game.Game, amounts map[uuid.UUID]int) {
	t.Helper()
	for i := 0; i < 24; i++ {
		var next *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && (c.Kind == game.PendingChoicePayAmount || c.Kind == game.PendingChoicePayUnless) {
				next = c
				break
			}
		}
		if next == nil {
			return
		}
		if next.Kind == game.PendingChoicePayAmount {
			if err := g.ResolvePayAmount(next.ID, next.Chooser, amounts[next.Chooser]); err != nil {
				t.Fatalf("answer number for %s: %v", next.Chooser, err)
			}
		} else if err := g.ResolvePayUnless(next.ID, next.Chooser, true); err != nil {
			t.Fatalf("pay for %s: %v", next.Chooser, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	t.Fatal("payment prompts never stopped")
}

// oldB456TopCard puts a card of the given mana cost on top of p's library.
func oldB456TopCard(p *game.Player, name, cost string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = cost
	p.Library.PushTop(c)
	return c.InstanceID
}

// oldB456Clash answers a clash with `opponent`: the controller picks them and
// both clashers keep their card on top.
func oldB456Clash(t *testing.T, g *game.Game, me, opponent uuid.UUID) {
	t.Helper()
	answered := false
	for i := 0; i < 8; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceOptionPick {
				open = c
				break
			}
		}
		if open == nil {
			break
		}
		pick := 0
		if !answered {
			for j, o := range open.PickOptions {
				if o.Player == opponent {
					pick = j
				}
			}
			answered = true
		}
		if err := g.ResolveOptionPick(open.ID, open.Chooser, pick); err != nil {
			t.Fatalf("answer the clash question: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Archangel of Strife -------------------------------------------------------

func TestOldB456ArchangelGivesWarPlayersPowerAndPeacePlayersToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	s := g.Seats
	mine := effectstest.B12Creature(g, s[0].ID, "My Bear", oldB456Bear, 2, 2)
	theirs := effectstest.B12Creature(g, s[1].ID, "Their Bear", oldB456Bear, 2, 2)
	other := effectstest.B12Creature(g, s[2].ID, "Third Bear", oldB456Bear, 2, 2)
	effectstest.EnterFromHand(t, g, s[0].ID, "Archangel of Strife", "Creature — Angel", oldB456Archangel)
	choices := map[uuid.UUID]int{s[0].ID: 0, s[1].ID: 1, s[2].ID: 0, s[3].ID: 1}
	for i := 0; i < 6; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceOptionPick {
				open = c
				break
			}
		}
		if open == nil {
			break
		}
		if err := g.ResolveOptionPick(open.ID, open.Chooser, choices[open.Chooser]); err != nil {
			t.Fatalf("choose: %v", err)
		}
	}
	for _, tc := range []struct {
		id   uuid.UUID
		name string
		p, t int
	}{{mine, "war player's Bear", 5, 2}, {theirs, "peace player's Bear", 2, 5}, {other, "war player's second Bear", 5, 2}} {
		c := effectstest.LayeredCard(t, g, tc.id)
		if c.CurrentPower() != tc.p || c.CurrentToughness() != tc.t {
			t.Errorf("%s is %d/%d, want %d/%d", tc.name, c.CurrentPower(), c.CurrentToughness(), tc.p, tc.t)
		}
	}
}

// --- Desecrator Hag --------------------------------------------------------------

func TestOldB456HagReturnsTheGreatestPowerCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	small := effectstest.PushGraveyardMaterial(me, game.Card{Name: "Small", TypeLine: oldB456Bear, Power: 1, Toughness: 1})
	big := effectstest.PushGraveyardMaterial(me, game.Card{Name: "Big", TypeLine: oldB456Bear, Power: 5, Toughness: 5})
	spell := effectstest.PushGraveyardMaterial(me, game.Card{Name: "Spell", TypeLine: "Sorcery"})
	effectstest.EnterFromHand(t, g, me.ID, "Desecrator Hag", "Creature — Hag", oldB456Hag)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(big) || me.Hand.Contains(small) || me.Hand.Contains(spell) {
		t.Errorf("hand has big=%v small=%v spell=%v", me.Hand.Contains(big), me.Hand.Contains(small), me.Hand.Contains(spell))
	}
}

func TestOldB456HagLetsYouChooseAmongTiedCardsAndDoesNothingWithNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushGraveyardMaterial(me, game.Card{Name: "Twin A", TypeLine: oldB456Bear, Power: 4, Toughness: 4})
	b := effectstest.PushGraveyardMaterial(me, game.Card{Name: "Twin B", TypeLine: oldB456Bear, Power: 4, Toughness: 4})
	c := effectstest.PushGraveyardMaterial(me, game.Card{Name: "Lesser", TypeLine: oldB456Bear, Power: 3, Toughness: 3})
	effectstest.EnterFromHand(t, g, me.ID, "Desecrator Hag", "Creature — Hag", oldB456Hag)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(b) || me.Hand.Contains(a) || me.Hand.Contains(c) {
		t.Errorf("hand has b=%v a=%v lesser=%v, want only the chosen twin", me.Hand.Contains(b), me.Hand.Contains(a), me.Hand.Contains(c))
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	hand := len(me2.Hand.Cards)
	effectstest.EnterFromHand(t, g2, me2.ID, "Desecrator Hag", "Creature — Hag", oldB456Hag)
	effectstest.PassPriorityAroundTable(t, g2)
	if len(me2.Hand.Cards) != hand {
		t.Error("the Hag returned something from an empty graveyard")
	}
}

// --- Vedalken Plotter ----------------------------------------------------------------

func TestOldB456PlotterExchangesTwoLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Island", TypeLine: "Basic Land — Island", Owner: opp.ID, Controller: opp.ID})
	effectstest.EnterFromHand(t, g, me.ID, "Vedalken Plotter", "Creature — Vedalken Wizard", oldB456Plotter)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	a, _ := effectstest.BattlefieldCard(g, mine)
	b, _ := effectstest.BattlefieldCard(g, theirs)
	if a.Controller != opp.ID || b.Controller != me.ID {
		t.Errorf("controllers after the exchange: my Forest %s, their Island %s", a.Controller, b.Controller)
	}
}

// --- Malfegor ------------------------------------------------------------------------

func TestOldB456MalfegorDiscardsYourHandAndEachOpponentSacrificesThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	s := g.Seats
	me := s[0]
	for i := 0; i < 2; i++ {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	hand := len(me.Hand.Cards)
	oneA := effectstest.B12Creature(g, s[1].ID, "A1", oldB456Bear, 1, 1)
	oneB := effectstest.B12Creature(g, s[1].ID, "A2", oldB456Bear, 1, 1)
	oneC := effectstest.B12Creature(g, s[1].ID, "A3", oldB456Bear, 1, 1)
	soloB := effectstest.B12Creature(g, s[2].ID, "B1", oldB456Bear, 1, 1)
	mineBear := effectstest.B12Creature(g, me.ID, "Mine", oldB456Bear, 1, 1)
	effectstest.EnterFromHand(t, g, me.ID, "Malfegor", "Legendary Creature — Demon Dragon", oldB456Malfegor)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != 0 {
		t.Errorf("hand has %d cards after discarding, want 0 (had %d)", len(me.Hand.Cards), hand)
	}
	// The hand held the two spares and Malfegor itself left it, so the count is
	// the cards that were really discarded.
	discarded := len(me.Graveyard.Cards)
	if discarded < 2 {
		t.Fatalf("only %d cards were discarded", discarded)
	}
	effectstest.AnswerSacrifice(t, g, s[1].ID, oneA, oneB)
	effectstest.AnswerSacrifice(t, g, s[2].ID, soloB)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, mineBear); !ok {
		t.Error("the controller sacrificed their own creature")
	}
	left := 0
	for _, id := range []uuid.UUID{oneA, oneB, oneC} {
		if _, ok := effectstest.BattlefieldCard(g, id); ok {
			left++
		}
	}
	if want := 3 - discarded; left != max(want, 0) {
		t.Errorf("the first opponent kept %d of 3 creatures after %d discards", left, discarded)
	}
	if _, ok := effectstest.BattlefieldCard(g, soloB); ok {
		t.Error("the lone creature of the second opponent was not sacrificed")
	}
}

func TestOldB456MalfegorWithAnEmptyHandCostsNobodyAnything(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	me := g.Seats[0]
	for len(me.Hand.Cards) > 0 {
		me.Hand.Cards = me.Hand.Cards[:len(me.Hand.Cards)-1]
	}
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB456Bear, 1, 1)
	effectstest.EnterFromHand(t, g, me.ID, "Malfegor", "Legendary Creature — Demon Dragon", oldB456Malfegor)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Error("an opponent sacrificed with nothing discarded")
	}
}

// --- Clash cards -----------------------------------------------------------------------

func TestOldB456PollenLullabyPreventsDamageAndFreezesTheClashedPlayersCreaturesOnAWin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB456TopCard(me, "My Five", "{3}{W}{W}")
	oldB456TopCard(opp, "Their One", "{U}")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB456Bear, 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true })
	effectstest.CastCatalogSpell(t, g, "Pollen Lullaby", "Instant", oldB456Pollen, nil)
	effectstest.PassPriorityAroundTable(t, g)
	oldB456Clash(t, g, me.ID, opp.ID)
	wins := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventClashWon && ev.Actor == me.ID {
			wins = true
		}
	}
	if !wins {
		t.Fatal("the clash was not won")
	}
	// Their next untap step: the Bear stays tapped.
	effectstest.EndTheTurn(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if c, ok := effectstest.BattlefieldCard(g, bear); !ok || !c.Tapped {
		t.Error("their creature untapped in their next untap step")
	}
}

func TestOldB456PollenLullabyDoesNotFreezeOnALoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB456TopCard(me, "My One", "{W}")
	oldB456TopCard(opp, "Their Five", "{3}{U}{U}")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB456Bear, 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true })
	effectstest.CastCatalogSpell(t, g, "Pollen Lullaby", "Instant", oldB456Pollen, nil)
	effectstest.PassPriorityAroundTable(t, g)
	oldB456Clash(t, g, me.ID, opp.ID)
	effectstest.EndTheTurn(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if c, ok := effectstest.BattlefieldCard(g, bear); !ok || c.Tapped {
		t.Error("their creature stayed tapped after a lost clash")
	}
}

func TestOldB456LashOutDealsThreeAndThreeMoreToTheControllerOnAWin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB456TopCard(me, "My Five", "{3}{R}{R}")
	oldB456TopCard(opp, "Their One", "{U}")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB456Bear, 2, 5)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Lash Out", "Instant", oldB456LashOut, effectstest.CardTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	oldB456Clash(t, g, me.ID, opp.ID)
	if got := effectstest.Pr6Marked(g, bear); got != 3 {
		t.Errorf("the creature has %d damage, want 3", got)
	}
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want -3 for the won clash", life, opp.Life)
	}
}

func TestOldB456LashOutHitsOnlyTheCreatureOnALoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB456TopCard(me, "My One", "{R}")
	oldB456TopCard(opp, "Their Five", "{3}{U}{U}")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB456Bear, 2, 5)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Lash Out", "Instant", oldB456LashOut, effectstest.CardTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	oldB456Clash(t, g, me.ID, opp.ID)
	if opp.Life != life {
		t.Errorf("life %d -> %d, want no damage to the player", life, opp.Life)
	}
}

func TestOldB456WhirlpoolWhelmReturnsToHandOnALossAndOffersTheLibraryOnAWin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB456TopCard(me, "My One", "{U}")
	oldB456TopCard(opp, "Their Five", "{3}{U}{U}")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB456Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Whirlpool Whelm", "Instant", oldB456Whelm, effectstest.CardTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	oldB456Clash(t, g, me.ID, opp.ID)
	if !opp.Hand.Contains(bear) {
		t.Error("a lost clash did not return the creature to its owner's hand")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.AdvanceToMain(t, g2)
	oldB456TopCard(me2, "My Five", "{3}{U}{U}")
	oldB456TopCard(opp2, "Their One", "{U}")
	bear2 := effectstest.B12Creature(g2, opp2.ID, "Their Bear", oldB456Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g2, "Whirlpool Whelm", "Instant", oldB456Whelm, effectstest.CardTargets(bear2))
	effectstest.PassPriorityAroundTable(t, g2)
	oldB456Clash(t, g2, me2.ID, opp2.ID)
	effectstest.AnswerMayChoice(t, g2, me2.ID, true)
	effectstest.PassPriorityAroundTable(t, g2)
	if n := opp2.Library.Size(); n == 0 || opp2.Library.Cards[n-1].InstanceID != bear2 {
		t.Error("the won clash did not put the creature on top of its owner's library")
	}
	if opp2.Hand.Contains(bear2) {
		t.Error("the creature was both tucked and returned")
	}
}

func TestOldB456ScatteringStrokeCountersAndRefundsOnAWin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB456TopCard(me, "My Five", "{3}{U}{U}")
	oldB456TopCard(opp, "Their One", "{U}")
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Fireball", "", "{2}{R}", []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.CastFromSeat(t, g, me, "Scattering Stroke", "Instant", oldB456Stroke, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)
	oldB456Clash(t, g, me.ID, opp.ID)
	if !effectstest.P7Countered(g, bolt) {
		t.Fatal("the spell was not countered")
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("mana at the next main phase = %d, want 3 (the countered spell's mana value)", got)
	}
}

func TestOldB456ScatteringStrokeRefundsNothingOnALoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB456TopCard(me, "My One", "{U}")
	oldB456TopCard(opp, "Their Five", "{3}{U}{U}")
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Fireball", "", "{2}{R}", []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.CastFromSeat(t, g, me, "Scattering Stroke", "Instant", oldB456Stroke, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)
	oldB456Clash(t, g, me.ID, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("mana at the next main phase = %d, want none after a lost clash", got)
	}
}

// --- Join forces ------------------------------------------------------------------------

func TestOldB456AllianceOfArmsGivesEveryoneTheTotalInSoldiers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	s := g.Seats
	effectstest.CastCatalogSpell(t, g, "Alliance of Arms", "Sorcery", oldB456Alliance, nil)
	effectstest.FillPool(s[0], 2)
	effectstest.FillPool(s[1], 1)
	effectstest.FillPool(s[2], 4)
	effectstest.PassPriorityAroundTable(t, g)
	oldB456AnswerPayments(t, g, map[uuid.UUID]int{s[0].ID: 2, s[1].ID: 1})
	for i := range s {
		if got := oldB456Count(g, s[i].ID, "Soldier", true); got != 3 {
			t.Errorf("seat %d has %d Soldiers, want 3", i, got)
		}
	}
	if n := s[2].ManaPool.Total(); n != 4 {
		t.Errorf("the seat that declined has %d mana, want its 4 kept", n)
	}
}

func TestOldB456AllianceOfArmsWithNothingPaidMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Alliance of Arms", "Sorcery", oldB456Alliance, nil)
	effectstest.PassPriorityAroundTable(t, g)
	oldB456AnswerPayments(t, g, nil)
	for i, p := range g.Seats {
		if got := oldB456Count(g, p.ID, "Soldier", true); got != 0 {
			t.Errorf("seat %d has %d Soldiers, want none", i, got)
		}
	}
}

func TestOldB456SharedTraumaMillsEveryoneTheTotal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	s := g.Seats
	for _, p := range s {
		effectstest.StockLibraryFor(g, p, 20)
	}
	effectstest.CastCatalogSpell(t, g, "Shared Trauma", "Sorcery", oldB456Trauma, nil)
	effectstest.FillPool(s[0], 1)
	effectstest.FillPool(s[1], 2)
	before := make([]int, len(s))
	for i, p := range s {
		before[i] = p.Library.Size()
	}
	effectstest.PassPriorityAroundTable(t, g)
	oldB456AnswerPayments(t, g, map[uuid.UUID]int{s[0].ID: 1, s[1].ID: 2})
	for i, p := range s {
		if got := before[i] - p.Library.Size(); got != 3 {
			t.Errorf("seat %d milled %d, want 3", i, got)
		}
	}
}

func TestOldB456ManaChargedDragonGetsPlusXWhenAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dragon := effectstest.B12Push(g, me.ID, "Mana-Charged Dragon", "Creature — Dragon", oldB456Dragon, 5, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.FillPool(me, 2)
	effectstest.FillPool(opp, 3)
	if err := g.DeclareAttacker(dragon, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	oldB456AnswerPayments(t, g, map[uuid.UUID]int{me.ID: 2, opp.ID: 3})
	c := effectstest.LayeredCard(t, g, dragon)
	if c.CurrentPower() != 10 || c.CurrentToughness() != 5 {
		t.Errorf("the Dragon is %d/%d, want 10/5", c.CurrentPower(), c.CurrentToughness())
	}
}
