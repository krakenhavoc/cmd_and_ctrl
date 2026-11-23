package testsi

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsb "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/b"
	cardsdi "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/di"
	cardsh "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/h"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- the upkeep shape ------------------------------------------------

// Stasis: the upkeep prompt keeps its key and its step anchor across a
// restore, so the restored table is still held in the upkeep, and the
// answer still sacrifices Stasis (or keeps it when the {U} is paid).
func TestStasisUpkeepPromptIsARestorePoint(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		p0 := g.Seats[0].ID
		stasis := effectstest.SeedReplacementPermanent(g, effectstest.StasisOracle, "Stasis", p0)
		effectstest.WalkToUpkeepOfSeat0(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if c := effectstest.UpkeepPromptFor(t, g, p0); c.PickThen != effects.UpkeepSacrificeSource.Key() {
			t.Fatalf("Stasis's prompt key is %q, want the keyed decline", c.PickThen)
		}
		g = effectstest.RestoreRoundTrip(t, g, true)
		me := g.Seats[0]
		prompt := effectstest.UpkeepPromptFor(t, g, p0)
		if prompt.PickThen != effects.UpkeepSacrificeSource.Key() || prompt.PayCost != "{U}" {
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
		effectstest.AnswerPayUnless(t, g, p0, pay)
		if kept := g.Battlefield.Contains(stasis); kept != pay {
			t.Errorf("pay=%v: Stasis kept = %v", pay, kept)
		}
	}
}

// Pact of Negation: the debt rides a key, so the restored table asks the
// same question and an unpaid Pact loses the game.
func TestPactOfNegationDebtIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	pact := effectstest.HandCardFull(me, "Pact of Negation", "Instant", "{0}", effectstest.PactOfNegationOracle, []string{"U"})
	if err := g.CastSpell(me.ID, pact, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("casting Pact of Negation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.UpkeepPromptFor(t, g, me.ID); c.PickThen != effects.PactLoseTheGame.Key() {
		t.Fatalf("the Pact's prompt key is %q, want the keyed decline", c.PickThen)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	if prompt := effectstest.UpkeepPromptFor(t, g, me.ID); prompt.PickThen != effects.PactLoseTheGame.Key() {
		t.Fatalf("restored prompt key %q", prompt.PickThen)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if !me.Eliminated {
		t.Error("an unpaid Pact of Negation did not lose its controller the game on the restored table")
	}
}

// Koskun Falls: a non-mana payment (tap an untapped creature) is plain
// data on the prompt, so the restored prompt still offers the creatures
// and a tap still keeps Koskun Falls.
func TestKoskunFallsTapPaymentIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	falls := effectstest.PushPermanentForTest(g, me.ID, "Koskun Falls", effectstest.KoskunFallsOracle, "World Enchantment")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	choice := effectstest.KoskunPrompt(t, g, 1)
	if choice.PayActionClause == nil || choice.PayActionClause.Kind != game.PayActionTap {
		t.Fatalf("Koskun Falls's prompt carries payment %+v, want a tap", choice.PayActionClause)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[1]
	choice = effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if choice.PickThen != effects.UpkeepSacrificeSource.Key() || choice.PayAction() == nil || choice.PayAction().Kind != game.PayActionTap {
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
	if c, ok := effectstest.BattlefieldCard(g, bear); !ok || !c.Tapped {
		t.Error("the Bear was not tapped as the payment")
	}
	if _, ok := effectstest.BattlefieldCard(g, falls); !ok {
		t.Error("Koskun Falls was sacrificed although the payment was made")
	}
}

// A cumulative upkeep (Mystic Remora): the age counters' worth of cost is
// the prompt's own words, so the restored prompt charges the same.
func TestCumulativeUpkeepPromptIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	remora := effectstest.PushPermanentForTest(g, owner.ID, "Mystic Remora", effectstest.OracleMysticRemora, "Enchantment")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	before := effectstest.RemoraUpkeepPrompt(t, g, owner.ID)
	cost := before.PayCost
	g = effectstest.RestoreRoundTrip(t, g, true)
	owner = g.Seats[1]
	choice := effectstest.RemoraUpkeepPrompt(t, g, owner.ID)
	if choice.PickThen != effects.UpkeepSacrificeSource.Key() || choice.PayCost != cost {
		t.Fatalf("restored prompt: key %q cost %q, want %q", choice.PickThen, choice.PayCost, cost)
	}
	if err := g.ResolvePayUnless(choice.ID, owner.ID, false); err != nil {
		t.Fatal(err)
	}
	if _, ok := effectstest.BattlefieldCard(g, remora); ok {
		t.Error("an unpaid cumulative upkeep did not sacrifice the permanent on the restored table")
	}
}

func TestUpkeepPromptWithAnUnknownKeyIsRefused(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p0 := g.Seats[0].ID
	effectstest.SeedReplacementPermanent(g, effectstest.StasisOracle, "Stasis", p0)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RefuseUnknownKey(t, g, effects.UpkeepSacrificeSource.Key())
}

// --- the may-pay shape -----------------------------------------------

// Extort: "you may pay {W/B}. If you do, ..." rides a key, so the restored
// prompt drains only when the mana is really paid.
func TestExtortMayPayIsARestorePoint(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, _, _ := effectstest.AwbPontiffBoard(t, g, 0)
		lives := effectstest.B17Life(g)
		effectstest.E2CastSpellOfColor(t, g, "Some Spell", "Sorcery", "U")
		for i := 0; i < 16 && !effectstest.HasPayUnlessFor(g, me.ID); i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		if c := effectstest.PayUnlessFor(g, me.ID); c == nil || c.PickThen != effects.ExtortDrain.Key() {
			t.Fatalf("the extort prompt is %+v, want the keyed may-pay", c)
		}
		g = effectstest.RestoreRoundTrip(t, g, true)
		me = g.Seats[0]
		c := effectstest.PayUnlessFor(g, me.ID)
		if c == nil || c.PickThen != effects.ExtortDrain.Key() {
			t.Fatalf("restored: %+v, want the keyed may-pay", c)
		}
		if pay {
			me.ManaPool.AddMana(game.ManaToken{Color: "B"})
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hellkite := effectstest.PushDiesCreatureForTest(g, me.ID, "Hellkite Charger", effectstest.HellkiteChargerOracle, "Creature — Dragon", 5, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(hellkite, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no pay prompt after Hellkite Charger's trigger")
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	c := effectstest.PayUnlessFor(g, me.ID)
	if c == nil || c.PickThen != cardsh.HellkiteChargerExtraCombat.Key() || c.OwedInStep.Step != game.StepDeclareAttackers {
		t.Fatalf("restored: %+v, want the keyed may-pay owed in the declare-attackers step", c)
	}
	if _, err := g.AdvanceStep(); err == nil && g.Turn.Step != game.StepDeclareAttackers {
		t.Fatalf("the restored table left the step with the payment unanswered: %+v", g.Turn)
	}
	effectstest.FillPool(me, 5)
	effectstest.FillPoolColored(me, "R", 2)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if effectstest.IsTapped(g, hellkite) {
		t.Error("the paid half did not untap the attackers on the restored table")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Errorf("after combat %+v, want the additional combat", next)
	}
}

// --- the guarded shape -----------------------------------------------

// Divert: the guarded spell rides the prompt (Guards), so the restored
// prompt still holds the spell on the stack and the decline still
// retargets it.
func TestDivertGuardedPayUnlessIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim, other := g.Seats[2].ID, g.Seats[3].ID
	bolt, _ := effectstest.BoltThenDivert(t, g)
	effectstest.PassUntilTaxed(t, g, g.Seats[0].ID)
	g = effectstest.RestoreRoundTrip(t, g, true)
	payer, caster := g.Seats[0], g.Seats[1]
	c := effectstest.PayUnlessFor(g, payer.ID)
	if c == nil || c.PickThen != cardsdi.DivertChangeTarget.Key() || c.GuardsStackItem != bolt {
		t.Fatalf("restored: %+v, want the keyed prompt guarding the Bolt", c)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("a pass while the restored tax was open = %v, want ErrChoicePending", err)
	}
	effectstest.AnswerPayUnless(t, g, payer.ID, false)
	prompt := effectstest.LatestRetarget(g, caster.ID)
	if prompt == nil {
		t.Fatal("a declined Divert opened no retarget prompt on the restored table")
	}
	if err := g.ResolveRetarget(prompt.ID, caster.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: other}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, victim) != 40 || effectstest.LifeOf(g, other) != 37 {
		t.Errorf("damage landed wrong: old=%d new=%d", effectstest.LifeOf(g, victim), effectstest.LifeOf(g, other))
	}
}

// Daze: the plain "counter that spell unless its controller pays" is
// keyed by the engine itself, with no card in the way.
func TestDazeTaxIsARestorePoint(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, them := g.Seats[0], g.Seats[1]
		effectstest.ToMainForCost(t, g)
		victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
		daze := effectstest.HandCardFull(them, "Daze", "Instant", "{1}{U}", effectstest.DazeOracle, []string{"U"})
		island := effectstest.FreeSpellPermanent(g, them.ID, "Island", "Basic Land — Island")
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
		effectstest.PassUntilTaxed(t, g, me.ID)
		g = effectstest.RestoreRoundTrip(t, g, true)
		me = g.Seats[0]
		c := effectstest.PayUnlessFor(g, me.ID)
		if c == nil || c.PickThen != "pay-unless/counter-guarded" || c.GuardsStackItem != victim {
			t.Fatalf("restored: %+v, want the engine's keyed counter-unless-pays guarding the spell", c)
		}
		if err := effectstest.TheThirdSeatPasses(t, g); !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("a pass while the restored tax was open = %v, want ErrChoicePending", err)
		}
		if pay {
			me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
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
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMainOf(t, g, 0)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.AppCastAsSeat(t, g, opp, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	dash := effectstest.AppCastAsSeat(t, g, me, "Dash Hopes", "Instant", effectstest.AppDashHopes,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)
	link := effectstest.AppOpenLink(g)
	if link == nil || link.PickThen != effects.AnyPlayerCounterThisSpell.Key() || link.GuardsStackItem != dash {
		t.Fatalf("the first link is %+v, want the keyed offer guarding Dash Hopes", link)
	}

	// First link, restored: seat 0 declines.
	g = effectstest.RestoreRoundTrip(t, g, true)
	me, opp = g.Seats[0], g.Seats[1]
	link = effectstest.AppOpenLink(g)
	if link == nil || link.Chooser != me.ID || link.PickThen != effects.AnyPlayerCounterThisSpell.Key() ||
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
	link = effectstest.AppOpenLink(g)
	if link == nil || link.Chooser != opp.ID || len(link.AnyPlayerOffer.Declined) != 1 || link.AnyPlayerOffer.Declined[0] != me.ID {
		t.Fatalf("second link: %+v, want seat 1 with seat 0's decline recorded", link)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	opp = g.Seats[1]
	link = effectstest.AppOpenLink(g)
	if link == nil || link.Chooser != opp.ID || len(link.AnyPlayerOffer.Declined) != 1 {
		t.Fatalf("restored second link: %+v", link)
	}
	oppLife := opp.Life
	if err := g.ResolvePayUnlessWithCards(link.ID, opp.ID, true, nil); err != nil {
		t.Fatalf("pay the second link: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-5 || !effectstest.AppInZone(g, dash, game.ZoneGraveyard) {
		t.Errorf("payer at %d life (want %d), Dash Hopes in graveyard = %v", opp.Life, oppLife-5, effectstest.AppInZone(g, dash, game.ZoneGraveyard))
	}
	if !effectstest.AppInZone(g, bolt, game.ZoneGraveyard) && g.Stack.Contains(bolt) {
		t.Error("the Bolt is still on the stack")
	}
}

// Temporal Extortion: the per-payer price is a registered function of the
// game and the player, so a restored offer asks each player for half
// their OWN life.
func TestAnyPlayerPerPayerPriceIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Life, opp.Life = 40, 21
	effectstest.CastCatalogSpell(t, g, "Temporal Extortion", "Sorcery", effectstest.AppTemporalExtortion, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me, opp = g.Seats[0], g.Seats[1]
	link := effectstest.AppOpenLink(g)
	if link == nil || link.PickThen != effects.AnyPlayerCounterThisSpellHalfLife.Key() || link.PayCost != "20 life" {
		t.Fatalf("restored first link: %+v, want 20 life", link)
	}
	if err := g.ResolvePayUnlessWithCards(link.ID, me.ID, false, nil); err != nil {
		t.Fatal(err)
	}
	link = effectstest.AppOpenLink(g)
	if link == nil || link.Chooser != opp.ID || link.PayCost != "11 life" {
		t.Fatalf("second link: %+v, want seat 1's 11 life", link)
	}
}

func TestAnyPlayerOfferWithAnUnknownKeyIsRefused(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMainOf(t, g, 0)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.AppCastAsSeat(t, g, opp, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.AppCastAsSeat(t, g, me, "Dash Hopes", "Instant", effectstest.AppDashHopes,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RefuseUnknownKey(t, g, effects.AnyPlayerCounterThisSpell.Key())
}

// --- the may-pay chain: a paid branch that opens a keyed prompt -------

// Banon's paid half opens a keyed discard, so the table is a restore
// point at the payment and again at the discard.
func TestBanonMayPayThenDiscardIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	banon := effectstest.PushCatalogPermanent(g, me.ID, "Banon, the Returners' Leader", "Legendary Creature — Human Cleric Rebel", effectstest.BanonOracle, false)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(banon, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no pay prompt after Banon's attack trigger")
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := effectstest.PayUnlessFor(g, me.ID)
	if c == nil || c.PickThen != cardsb.BanonPaid.Key() {
		t.Fatalf("restored: %+v, want the keyed may-pay", c)
	}
	effectstest.FillPool(me, 1)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	d := effectstest.DiscardChoiceFor(g, me.ID)
	if d == nil || d.PickThen != cardsb.BanonDraw.Key() {
		t.Fatalf("the paid half's discard is %+v, want the keyed prompt", d)
	}
}
