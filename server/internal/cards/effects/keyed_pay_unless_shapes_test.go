package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyed_pay_unless_shapes_test.go — #4264: the pay-unless shapes #4256
// left on the closure form are keyed restore points. Each test takes a
// STRICT restore (restoreRoundTrip with strict) while the prompt is open
// and answers it on the restored table; the unknown-key tests are the
// rollback case (refuseUnknownKey and its own for the any-player
// registry).

// payUnlessFor is the open pay-unless `chooser` owes, nil when none.
func payUnlessFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// --- the upkeep shape ------------------------------------------------

// Stasis: the upkeep prompt keeps its key and its step anchor across a
// restore, so the restored table is still held in the upkeep, and the
// answer still sacrifices Stasis (or keeps it when the {U} is paid).
func TestStasisUpkeepPromptIsARestorePoint(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := newCatalogGame(t)
		p0 := g.Seats[0].ID
		stasis := seedReplacementPermanent(g, stasisOracle, "Stasis", p0)
		walkToUpkeepOfSeat0(t, g)
		passPriorityAroundTable(t, g)
		if c := upkeepPromptFor(t, g, p0); c.PickThen != upkeepSacrificeSource.Key() {
			t.Fatalf("Stasis's prompt key is %q, want the keyed decline", c.PickThen)
		}
		g = restoreRoundTrip(t, g, true)
		me := g.Seats[0]
		prompt := upkeepPromptFor(t, g, p0)
		if prompt.PickThen != upkeepSacrificeSource.Key() || prompt.PayCost != "{U}" {
			t.Fatalf("restored prompt: key %q cost %q", prompt.PickThen, prompt.PayCost)
		}
		if prompt.OwedInStep.Step != game.StepUpkeep || prompt.OwedInStep.Turn != g.Turn.Seq {
			t.Errorf("restored prompt is anchored to %v, want this turn's upkeep", prompt.OwedInStep)
		}
		if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
			t.Errorf("a pass while the restored upkeep prompt was open = %v, want ErrChoicePending", err)
		}
		if pay {
			me.ManaPool.AddMana(game.ManaToken{Color: "U"})
		}
		answerPayUnless(t, g, p0, pay)
		if kept := g.Battlefield.Contains(stasis); kept != pay {
			t.Errorf("pay=%v: Stasis kept = %v", pay, kept)
		}
	}
}

// Pact of Negation: the debt rides a key, so the restored table asks the
// same question and an unpaid Pact loses the game.
func TestPactOfNegationDebtIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	pact := handCardFull(me, "Pact of Negation", "Instant", "{0}", pactOfNegationOracle, []string{"U"})
	if err := g.CastSpell(me.ID, pact, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("casting Pact of Negation: %v", err)
	}
	passPriorityAroundTable(t, g)
	walkToUpkeepOfSeat0(t, g)
	passPriorityAroundTable(t, g)
	if c := upkeepPromptFor(t, g, me.ID); c.PickThen != pactLoseTheGame.Key() {
		t.Fatalf("the Pact's prompt key is %q, want the keyed decline", c.PickThen)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	if prompt := upkeepPromptFor(t, g, me.ID); prompt.PickThen != pactLoseTheGame.Key() {
		t.Fatalf("restored prompt key %q", prompt.PickThen)
	}
	answerPayUnless(t, g, me.ID, false)
	if !me.Eliminated {
		t.Error("an unpaid Pact of Negation did not lose its controller the game on the restored table")
	}
}

// Koskun Falls: a non-mana payment (tap an untapped creature) is plain
// data on the prompt, so the restored prompt still offers the creatures
// and a tap still keeps Koskun Falls.
func TestKoskunFallsTapPaymentIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	falls := pushPermanentForTest(g, me.ID, "Koskun Falls", koskunFallsOracle, "World Enchantment")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	choice := koskunPrompt(t, g, 1)
	if choice.PayActionClause == nil || choice.PayActionClause.Kind != game.PayActionTap {
		t.Fatalf("Koskun Falls's prompt carries payment %+v, want a tap", choice.PayActionClause)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[1]
	choice = remoraUpkeepPrompt(t, g, me.ID)
	if choice.PickThen != upkeepSacrificeSource.Key() || choice.PayAction() == nil || choice.PayAction().Kind != game.PayActionTap {
		t.Fatalf("restored prompt: key %q payment %+v", choice.PickThen, choice.PayAction())
	}
	var opts []uuid.UUID
	g.ReadSnapshot(func() { opts = g.PayActionOptionsForEffect(me.ID, choice.PayAction()) })
	if len(opts) != 1 || opts[0] != bear {
		t.Fatalf("the restored payment offers %v, want only the Bear %s", opts, bear)
	}
	if err := g.ResolvePayUnlessWithCards(choice.ID, me.ID, true, []uuid.UUID{bear}); err != nil {
		t.Fatalf("tap the Bear on the restored table: %v", err)
	}
	if c, ok := battlefieldCard(g, bear); !ok || !c.Tapped {
		t.Error("the Bear was not tapped as the payment")
	}
	if _, ok := battlefieldCard(g, falls); !ok {
		t.Error("Koskun Falls was sacrificed although the payment was made")
	}
}

// A cumulative upkeep (Mystic Remora): the age counters' worth of cost is
// the prompt's own words, so the restored prompt charges the same.
func TestCumulativeUpkeepPromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	remora := pushPermanentForTest(g, owner.ID, "Mystic Remora", oracleMysticRemora, "Enchantment")
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	before := remoraUpkeepPrompt(t, g, owner.ID)
	cost := before.PayCost
	g = restoreRoundTrip(t, g, true)
	owner = g.Seats[1]
	choice := remoraUpkeepPrompt(t, g, owner.ID)
	if choice.PickThen != upkeepSacrificeSource.Key() || choice.PayCost != cost {
		t.Fatalf("restored prompt: key %q cost %q, want %q", choice.PickThen, choice.PayCost, cost)
	}
	if err := g.ResolvePayUnless(choice.ID, owner.ID, false); err != nil {
		t.Fatal(err)
	}
	if _, ok := battlefieldCard(g, remora); ok {
		t.Error("an unpaid cumulative upkeep did not sacrifice the permanent on the restored table")
	}
}

func TestUpkeepPromptWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0].ID
	seedReplacementPermanent(g, stasisOracle, "Stasis", p0)
	walkToUpkeepOfSeat0(t, g)
	passPriorityAroundTable(t, g)
	refuseUnknownKey(t, g, upkeepSacrificeSource.Key())
}

// --- the may-pay shape -----------------------------------------------

// Extort: "you may pay {W/B}. If you do, ..." rides a key, so the restored
// prompt drains only when the mana is really paid.
func TestExtortMayPayIsARestorePoint(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me, _, _ := awbPontiffBoard(t, g, 0)
		lives := b17Life(g)
		e2CastSpellOfColor(t, g, "Some Spell", "Sorcery", "U")
		for i := 0; i < 16 && !hasPayUnlessFor(g, me.ID); i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		if c := payUnlessFor(g, me.ID); c == nil || c.PickThen != extortDrain.Key() {
			t.Fatalf("the extort prompt is %+v, want the keyed may-pay", c)
		}
		g = restoreRoundTrip(t, g, true)
		me = g.Seats[0]
		c := payUnlessFor(g, me.ID)
		if c == nil || c.PickThen != extortDrain.Key() {
			t.Fatalf("restored: %+v, want the keyed may-pay", c)
		}
		if pay {
			me.ManaPool.AddMana(game.ManaToken{Color: "B"})
		}
		answerPayUnless(t, g, me.ID, pay)
		passPriorityAroundTable(t, g)
		drained := 0
		for i, p := range g.Seats {
			if i != 0 && p.Life < lives[i] {
				drained += lives[i] - p.Life
			}
		}
		want := 0
		if pay {
			want = len(g.Seats) - 1
		}
		if drained != want {
			t.Errorf("pay=%v: opponents lost %d life on the restored table, want %d", pay, drained, want)
		}
	}
}

// Hellkite Charger: a may-pay that holds its step. The restored prompt is
// still owed in that step and still gives the extra combat when paid.
func TestHellkiteChargerMayPayIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hellkite := pushDiesCreatureForTest(g, me.ID, "Hellkite Charger", hellkiteChargerOracle, "Creature — Dragon", 5, 5)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(hellkite, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no pay prompt after Hellkite Charger's trigger")
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	c := payUnlessFor(g, me.ID)
	if c == nil || c.PickThen != hellkiteChargerExtraCombat.Key() || c.OwedInStep.Step != game.StepDeclareAttackers {
		t.Fatalf("restored: %+v, want the keyed may-pay owed in the declare-attackers step", c)
	}
	if _, err := g.AdvanceStep(); err == nil && g.Turn.Step != game.StepDeclareAttackers {
		t.Fatalf("the restored table left the step with the payment unanswered: %+v", g.Turn)
	}
	fillPool(me, 5)
	fillPoolColored(me, "R", 2)
	answerPayUnless(t, g, me.ID, true)
	if isTapped(g, hellkite) {
		t.Error("the paid half did not untap the attackers on the restored table")
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Errorf("after combat %+v, want the additional combat", next)
	}
}

// --- the guarded shape -----------------------------------------------

// Divert: the guarded spell rides the prompt (Guards), so the restored
// prompt still holds the spell on the stack and the decline still
// retargets it.
func TestDivertGuardedPayUnlessIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	victim, other := g.Seats[2].ID, g.Seats[3].ID
	bolt, _ := boltThenDivert(t, g)
	passUntilTaxed(t, g, g.Seats[0].ID)
	g = restoreRoundTrip(t, g, true)
	payer, caster := g.Seats[0], g.Seats[1]
	c := payUnlessFor(g, payer.ID)
	if c == nil || c.PickThen != divertChangeTarget.Key() || c.GuardsStackItem != bolt {
		t.Fatalf("restored: %+v, want the keyed prompt guarding the Bolt", c)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("a pass while the restored tax was open = %v, want ErrChoicePending", err)
	}
	answerPayUnless(t, g, payer.ID, false)
	prompt := latestRetarget(g, caster.ID)
	if prompt == nil {
		t.Fatal("a declined Divert opened no retarget prompt on the restored table")
	}
	if err := g.ResolveRetarget(prompt.ID, caster.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: other}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if lifeOf(g, victim) != 40 || lifeOf(g, other) != 37 {
		t.Errorf("damage landed wrong: old=%d new=%d", lifeOf(g, victim), lifeOf(g, other))
	}
}

// Daze: the plain "counter that spell unless its controller pays" is
// keyed by the engine itself, with no card in the way.
func TestDazeTaxIsARestorePoint(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := newCatalogGame(t)
		me, them := g.Seats[0], g.Seats[1]
		toMainForCost(t, g)
		victim := castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
		daze := handCardFull(them, "Daze", "Instant", "{1}{U}", dazeOracle, []string{"U"})
		island := freeSpellPermanent(g, them.ID, "Island", "Basic Land — Island")
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority to the responder: %v", err)
		}
		if err := g.CastSpell(them.ID, daze, game.CastSpellParams{
			Strict:          true,
			AlternativeCost: "return",
			AltCostIDs:      []uuid.UUID{island},
			Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
		}); err != nil {
			t.Fatalf("casting Daze: %v", err)
		}
		passUntilTaxed(t, g, me.ID)
		g = restoreRoundTrip(t, g, true)
		me = g.Seats[0]
		c := payUnlessFor(g, me.ID)
		if c == nil || c.PickThen != "pay-unless/counter-guarded" || c.GuardsStackItem != victim {
			t.Fatalf("restored: %+v, want the engine's keyed counter-unless-pays guarding the spell", c)
		}
		if err := theThirdSeatPasses(t, g); !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("a pass while the restored tax was open = %v, want ErrChoicePending", err)
		}
		if pay {
			me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
		answerPayUnless(t, g, me.ID, pay)
		if countered := !g.Stack.Contains(victim); countered == pay {
			t.Errorf("pay=%v: the Dazed spell countered = %v", pay, countered)
		}
	}
}

// --- the any-player shape --------------------------------------------

// Dash Hopes: every link of the offer is a keyed prompt. Restored at the
// first link and again at the second (with the first player's decline
// recorded), the offer asks the same players in the same order and the
// first to pay counters the spell.
func TestAnyPlayerOfferIsARestorePointAtEveryLink(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMainOf(t, g, 0)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := appCastAsSeat(t, g, opp, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	dash := appCastAsSeat(t, g, me, "Dash Hopes", "Instant", appDashHopes,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	passPriorityAroundTable(t, g)
	link := appOpenLink(g)
	if link == nil || link.PickThen != anyPlayerCounterThisSpell.Key() || link.GuardsStackItem != dash {
		t.Fatalf("the first link is %+v, want the keyed offer guarding Dash Hopes", link)
	}

	// First link, restored: seat 0 declines.
	g = restoreRoundTrip(t, g, true)
	me, opp = g.Seats[0], g.Seats[1]
	link = appOpenLink(g)
	if link == nil || link.Chooser != me.ID || link.PickThen != anyPlayerCounterThisSpell.Key() ||
		link.PayAction() == nil || link.PayAction().Kind != game.PayActionLife {
		t.Fatalf("restored first link: %+v", link)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("a pass while the restored offer was open = %v, want ErrChoicePending", err)
	}
	if err := g.ResolvePayUnlessWithCards(link.ID, me.ID, false, nil); err != nil {
		t.Fatalf("decline the first link: %v", err)
	}

	// Second link, restored: the first decline is on the offer.
	link = appOpenLink(g)
	if link == nil || link.Chooser != opp.ID || len(link.AnyPlayerOffer.Declined) != 1 || link.AnyPlayerOffer.Declined[0] != me.ID {
		t.Fatalf("second link: %+v, want seat 1 with seat 0's decline recorded", link)
	}
	g = restoreRoundTrip(t, g, true)
	opp = g.Seats[1]
	link = appOpenLink(g)
	if link == nil || link.Chooser != opp.ID || len(link.AnyPlayerOffer.Declined) != 1 {
		t.Fatalf("restored second link: %+v", link)
	}
	oppLife := opp.Life
	if err := g.ResolvePayUnlessWithCards(link.ID, opp.ID, true, nil); err != nil {
		t.Fatalf("pay the second link: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-5 || !appInZone(g, dash, game.ZoneGraveyard) {
		t.Errorf("payer at %d life (want %d), Dash Hopes in graveyard = %v", opp.Life, oppLife-5, appInZone(g, dash, game.ZoneGraveyard))
	}
	if !appInZone(g, bolt, game.ZoneGraveyard) && g.Stack.Contains(bolt) {
		t.Error("the Bolt is still on the stack")
	}
}

// Temporal Extortion: the per-payer price is a registered function of the
// game and the player, so a restored offer asks each player for half
// their OWN life.
func TestAnyPlayerPerPayerPriceIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Life, opp.Life = 40, 21
	castCatalogSpell(t, g, "Temporal Extortion", "Sorcery", appTemporalExtortion, nil)
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me, opp = g.Seats[0], g.Seats[1]
	link := appOpenLink(g)
	if link == nil || link.PickThen != anyPlayerCounterThisSpellHalfLife.Key() || link.PayCost != "20 life" {
		t.Fatalf("restored first link: %+v, want 20 life", link)
	}
	if err := g.ResolvePayUnlessWithCards(link.ID, me.ID, false, nil); err != nil {
		t.Fatal(err)
	}
	link = appOpenLink(g)
	if link == nil || link.Chooser != opp.ID || link.PayCost != "11 life" {
		t.Fatalf("second link: %+v, want seat 1's 11 life", link)
	}
}

func TestAnyPlayerOfferWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMainOf(t, g, 0)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := appCastAsSeat(t, g, opp, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	appCastAsSeat(t, g, me, "Dash Hopes", "Instant", appDashHopes,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	passPriorityAroundTable(t, g)
	refuseUnknownKey(t, g, anyPlayerCounterThisSpell.Key())
}

// --- the may-pay chain: a paid branch that opens a keyed prompt -------

// Banon's paid half opens a keyed discard, so the table is a restore
// point at the payment and again at the discard.
func TestBanonMayPayThenDiscardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	banon := pushCatalogPermanent(g, me.ID, "Banon, the Returners' Leader", "Legendary Creature — Human Cleric Rebel", banonOracle, false)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(banon, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no pay prompt after Banon's attack trigger")
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := payUnlessFor(g, me.ID)
	if c == nil || c.PickThen != banonPaid.Key() {
		t.Fatalf("restored: %+v, want the keyed may-pay", c)
	}
	fillPool(me, 1)
	answerPayUnless(t, g, me.ID, true)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	d := discardChoiceFor(g, me.ID)
	if d == nil || d.PickThen != banonDraw.Key() {
		t.Fatalf("the paid half's discard is %+v, want the keyed prompt", d)
	}
}
