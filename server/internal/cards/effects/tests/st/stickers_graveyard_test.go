package testsst

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stickers_graveyard_test.go — stickers on cards in a graveyard, sticker
// kicker, "all abilities of ability stickers", and Done for the Day (ADR
// 0208, #4307).

const (
	scampireOracle         = "64dbed80-b3a2-4a0f-afe4-c3d968173e50"
	roxiOracle             = "5e24a352-7c90-458b-9bb2-bd4bc22bc748"
	ambassadorOracle       = "183a9488-e21f-448b-8b69-e53b56fa0a56"
	unlawfulEntryOracle    = "898ee071-dd37-4a02-895d-a41b568d7eef"
	aGoodDayToPieOracle    = "92de7cf0-ac3e-46ae-86b2-7d8ebb0276e4"
	doneForTheDayOracle    = "2a248d6a-aa83-4e59-848b-46de8204b538"
	wickerPickerOracle     = "15a71d11-da00-4c1a-bf64-ec3f627ed6a3"
	pinCollectionOracle    = "69fedfbb-5d88-4fe0-81e1-1e97a56c6aca"
	clandestineChamOracle  = "886e00ab-c694-4839-b3fb-7a72048db4b2"
	stkCreatureTypeLine    = "Creature — Bear"
	stkEmployeeTypeLine    = "Creature — Human Employee"
	stkPerformerTypeLine   = "Creature — Human Performer"
	stkRobotTypeLine       = "Artifact Creature — Robot"
	stkEnchantmentTypeLine = "Enchantment"
)

// stkGraveyardCard puts a card with an oracle in a graveyard.
func stkGraveyardCard(owner *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	owner.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: owner.ID, Controller: owner.ID,
	})
	return id
}

// Scampire: it gets a ticket, then a sticker goes on a creature card in the
// graveyard, where its name applies.
func TestStkScampirePutsAStickerOnACreatureCardInTheGraveyard(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster")
	dead := stkGraveyardCard(me, "Dead Bear", stkCreatureTypeLine, "")
	land := stkGraveyardCard(me, "Dead Land", "Land", "")
	effectstest.CastAndResolveCreature(t, g, "Scampire", "Creature — Vampire Employee", scampireOracle)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("Scampire asked no sticker question")
	}
	for _, id := range pick.ChooseCards {
		if id == land {
			t.Error("a land card in the graveyard is offered")
		}
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != dead {
		t.Fatalf("candidates = %v, want just the Dead Bear", pick.ChooseCards)
	}
	effectstest.StkPutOn(t, g, me, dead, "“Night”", 0)
	if got := stkTickets(g, me); got != 1 {
		t.Errorf("tickets = %d, want 1 (Scampire gave one; the name sticker is free)", got)
	}
	c := effectstest.StkCard(t, g, dead)
	if len(c.Stickers) != 1 {
		t.Fatalf("stickers = %+v, want one", c.Stickers)
	}
	if got := stkNames(c); got != "Night Dead Bear" {
		t.Errorf("name in the graveyard = %q, want %q (CR 123.6)", got, "Night Dead Bear")
	}
}

// A card with a sticker is "stickered" in a graveyard (CR 123.4): its
// ability sticker's keyword applies there (CR 123.7).
func TestStkAbilityStickerOnAGraveyardCardGivesItsKeyword(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 2, "Night Brushwagg Ringmaster")
	dead := stkGraveyardCard(me, "Dead Bear", stkCreatureTypeLine, "")
	effectstest.StkGive(t, g, me, dead, "Menace")
	c := effectstest.StkCard(t, g, dead)
	has := false
	for _, a := range c.Effective().Abilities {
		has = has || a == game.CounterMenace
	}
	if !has {
		t.Errorf("the menace sticker gave a card in a graveyard no menace: %v", c.Effective().Abilities)
	}
}

// Scampire's second ability returns a creature card with a sticker from
// the graveyard, gives it haste and exiles it at the end step.
func TestStkScampireReturnsAStickeredCreatureCard(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster")
	stickered := stkGraveyardCard(me, "Dead Bear", stkCreatureTypeLine, "")
	plain := stkGraveyardCard(me, "Plain Bear", stkCreatureTypeLine, "")
	effectstest.StkGive(t, g, me, stickered, "“Night”")
	scampire := effectstest.PushCatalogPermanent(g, me.ID, "Scampire", "Creature — Vampire Employee", scampireOracle, false)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.P2Activate(t, g, me, scampire, 0, "BBBB", effectstest.P2Card(plain)); err == nil {
		t.Fatal("a creature card with no sticker was a legal target")
	}
	if err := effectstest.P2Activate(t, g, me, scampire, 0, "BBBB", effectstest.P2Card(stickered)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, stickered) {
		t.Fatal("the creature card did not return")
	}
	if !effectstest.P2Keyword(g, stickered, "haste") {
		t.Error("the returned creature has no haste")
	}
	// The sticker came with it (CR 123.5).
	if got := stkNames(effectstest.StkCard(t, g, stickered)); got != "Night Dead Bear" {
		t.Errorf("name on the battlefield = %q, want %q", got, "Night Dead Bear")
	}
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, stickered) {
		t.Error("the returned creature was not exiled at the end step")
	}
}

// Roxi: power is the permanents with an art sticker plus the graveyard
// cards with one, counted once each.
func TestStkRoxiCountsArtStickeredPermanentsAndGraveyardCards(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster", "Phyrexian Midway Bamboozle")
	roxi := effectstest.P2Push(g, me.ID, "Roxi, Publicist to the Stars", "Legendary Creature — Human Employee", roxiOracle, 0, 4)
	bear := effectstest.P2Push(g, me.ID, "Bear", stkCreatureTypeLine, "", 2, 2)
	dead := stkGraveyardCard(me, "Dead Bear", stkCreatureTypeLine, "")
	opp := g.Seats[1]
	theirs := effectstest.P2Push(g, opp.ID, "Their Bear", stkCreatureTypeLine, "", 2, 2)
	_ = theirs
	if p, _ := effectstest.P2PT(g, roxi); p != 0 {
		t.Fatalf("power with no stickers = %d, want 0", p)
	}
	effectstest.StkGive(t, g, me, bear, "Art sticker 1 — Night Brushwagg Ringmaster")
	if p, _ := effectstest.P2PT(g, roxi); p != 1 {
		t.Errorf("power with one art-stickered permanent = %d, want 1", p)
	}
	effectstest.StkGive(t, g, me, bear, "Art sticker 2 — Night Brushwagg Ringmaster")
	if p, _ := effectstest.P2PT(g, roxi); p != 1 {
		t.Errorf("power with two art stickers on one permanent = %d, want 1 (permanents, not stickers)", p)
	}
	effectstest.StkGive(t, g, me, dead, "Art sticker 3 — Night Brushwagg Ringmaster")
	if p, tough := effectstest.P2PT(g, roxi); p != 2 || tough != 4 {
		t.Errorf("P/T with a graveyard card too = %d/%d, want 2/4", p, tough)
	}
	// The card leaves the graveyard for a hand: the sticker comes off.
	g.WithWriteLock(func() { _ = g.ReturnFromGraveyardForEffect(dead, game.ZoneHand) })
	if p, _ := effectstest.P2PT(g, roxi); p != 1 {
		t.Errorf("power after the stickered card left the graveyard = %d, want 1", p)
	}
	// A name sticker is not an art sticker.
	effectstest.StkGive(t, g, me, roxi, "“Night”")
	if p, _ := effectstest.P2PT(g, roxi); p != 1 {
		t.Errorf("power after a name sticker = %d, want 1", p)
	}
}

// Roxi's enters trigger: up to two art stickers among nonland permanents.
func TestStkRoxiDistributesUpToTwoArtStickers(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster")
	bear := effectstest.P2Push(g, me.ID, "Bear", stkCreatureTypeLine, "", 2, 2)
	// Cast from the hand with its printed toughness, as the deck has it.
	effectstest.AdvanceToMain(t, g)
	roxi := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: roxi, Name: "Roxi, Publicist to the Stars", TypeLine: "Legendary Creature — Human Employee",
			OracleID: roxiOracle, Toughness: 4, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, roxi, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	which := effectstest.LatestChooseCardsFor(g, me.ID)
	if which == nil {
		t.Fatal("no sticker question")
	}
	kinds := func() {
		w := effectstest.LatestOptionPickFor(g, me.ID)
		for _, o := range w.PickOptions {
			if o.Label != game.DeclineStickerLabel && !strings.HasPrefix(o.Label, "Art sticker") {
				t.Errorf("offered %q, which is not an art sticker", o.Label)
			}
		}
	}
	if err := g.ResolveChooseCards(which.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("first permanent: %v", err)
	}
	kinds()
	w := effectstest.LatestOptionPickFor(g, me.ID)
	if err := g.ResolveOptionPick(w.ID, me.ID, effectstest.StkOption(t, w, "Art sticker 1")); err != nil {
		t.Fatalf("first sticker: %v", err)
	}
	which = effectstest.LatestChooseCardsFor(g, me.ID)
	if which == nil {
		t.Fatal("no second sticker question")
	}
	if err := g.ResolveChooseCards(which.ID, me.ID, []uuid.UUID{roxi}); err != nil {
		t.Fatalf("second permanent: %v", err)
	}
	w = effectstest.LatestOptionPickFor(g, me.ID)
	if err := g.ResolveOptionPick(w.ID, me.ID, effectstest.StkOption(t, w, "Art sticker 2")); err != nil {
		t.Fatalf("second sticker: %v", err)
	}
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("a third question is open")
	}
	// Two permanents with an art sticker, Roxi herself one: power 2.
	if p, _ := effectstest.P2PT(g, roxi); p != 2 {
		t.Errorf("Roxi's power = %d, want 2", p)
	}
}

// Ambassador Blorpityblorpboop: base power and toughness become the totals
// of the power and toughness stickers on permanents you control.
func TestStkAmbassadorTakesTheStickersTotals(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 8, "Night Brushwagg Ringmaster")
	amb := effectstest.P2Push(g, me.ID, "Ambassador Blorpityblorpboop", "Legendary Creature — Alien Advisor Guest", ambassadorOracle, 3, 3)
	bear := effectstest.P2Push(g, me.ID, "Bear", stkCreatureTypeLine, "", 2, 2)
	other := effectstest.P2Push(g, me.ID, "Other", stkCreatureTypeLine, "", 1, 1)
	effectstest.StkGive(t, g, me, bear, "sticker 2/3")
	effectstest.StkGive(t, g, me, other, "sticker 10/10")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	// 2/3 on the Bear and 10/10 on the other: totals 12 and 13. The bear's
	// own sticker keeps it 2/3.
	if p, tough := effectstest.P2PT(g, amb); p != 12 || tough != 13 {
		t.Errorf("Ambassador P/T = %d/%d, want 12/13", p, tough)
	}
}

// Declined, the base stays; accepted again later it is read afresh.
func TestStkAmbassadorMayDeclineAndIsAskedEachCombat(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 2, "Night Brushwagg Ringmaster")
	amb := effectstest.P2Push(g, me.ID, "Ambassador Blorpityblorpboop", "Legendary Creature — Alien Advisor Guest", ambassadorOracle, 3, 3)
	bear := effectstest.P2Push(g, me.ID, "Bear", stkCreatureTypeLine, "", 2, 2)
	effectstest.StkGive(t, g, me, bear, "sticker 2/3")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.P2PT(g, amb); p != 3 || tough != 3 {
		t.Errorf("Ambassador P/T after declining = %d/%d, want 3/3", p, tough)
	}
}

// With no stickers the totals are 0/0 and it dies, as printed.
func TestStkAmbassadorWithNoStickersBecomesZeroZero(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster")
	amb := effectstest.P2Push(g, me.ID, "Ambassador Blorpityblorpboop", "Legendary Creature — Alien Advisor Guest", ambassadorOracle, 3, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SettleStateBasedActions(t, g)
	if effects.OnBattlefield(g, amb) {
		t.Error("a 0/0 Ambassador is still on the battlefield")
	}
}

// Unlawful Entry: the spell, and the graveyard ability on an art sticker on
// a creature.
func TestStkUnlawfulEntryReturnsFromTheGraveyardOnAnArtStickerOnACreature(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster")
	entry := stkGraveyardCard(me, "Unlawful Entry", "Instant", unlawfulEntryOracle)
	bear := effectstest.P2Push(g, me.ID, "Bear", stkCreatureTypeLine, "", 2, 2)
	// A name sticker does not trigger it.
	effectstest.StkGive(t, g, me, bear, "“Night”")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(entry) {
		t.Fatal("a name sticker returned Unlawful Entry")
	}
	effectstest.StkGive(t, g, me, bear, "Art sticker 1 — Night Brushwagg Ringmaster")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(entry) {
		t.Error("an art sticker on a creature did not return Unlawful Entry to hand")
	}
}

// "On a creature": an art sticker on a creature CARD in a graveyard, or on a
// noncreature permanent, does not trigger it.
func TestStkUnlawfulEntryNeedsACreature(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster")
	entry := stkGraveyardCard(me, "Unlawful Entry", "Instant", unlawfulEntryOracle)
	dead := stkGraveyardCard(me, "Dead Bear", stkCreatureTypeLine, "")
	rock := effectstest.P2Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
	effectstest.StkGive(t, g, me, dead, "Art sticker 1 — Night Brushwagg Ringmaster")
	effectstest.StkGive(t, g, me, rock, "Art sticker 2 — Night Brushwagg Ringmaster")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil || !me.Graveyard.Contains(entry) {
		t.Error("Unlawful Entry triggered for something that is not a creature")
	}
}

// Unlawful Entry's spell half.
func TestStkUnlawfulEntryGivesPlusOneAndFlying(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0)
	bear := effectstest.P2Push(g, me.ID, "Bear", stkCreatureTypeLine, "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Unlawful Entry", "Instant", unlawfulEntryOracle, []game.TargetRef{effectstest.P2Card(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.P2PT(g, bear); p != 3 {
		t.Errorf("power = %d, want 3", p)
	}
	if !effectstest.P2Keyword(g, bear, "flying") {
		t.Error("the Bear has no flying")
	}
}

// A Good Day to Pie: a name sticker on a creature returns it; an art
// sticker does not.
func TestStkAGoodDayToPieReturnsOnANameStickerOnACreature(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster")
	pie := stkGraveyardCard(me, "A Good Day to Pie", "Instant", aGoodDayToPieOracle)
	bear := effectstest.P2Push(g, me.ID, "Bear", stkCreatureTypeLine, "", 2, 2)
	effectstest.StkGive(t, g, me, bear, "Art sticker 1 — Night Brushwagg Ringmaster")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pie) {
		t.Fatal("an art sticker returned A Good Day to Pie")
	}
	effectstest.StkGive(t, g, me, bear, "“Night”")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(pie) {
		t.Error("a name sticker on a creature did not return A Good Day to Pie")
	}
}

// Declining leaves it in the graveyard.
func TestStkAGoodDayToPieMayBeLeftInTheGraveyard(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster")
	pie := stkGraveyardCard(me, "A Good Day to Pie", "Instant", aGoodDayToPieOracle)
	bear := effectstest.P2Push(g, me.ID, "Bear", stkCreatureTypeLine, "", 2, 2)
	effectstest.StkGive(t, g, me, bear, "“Night”")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pie) {
		t.Error("it left the graveyard though the player declined")
	}
}

// The spell half: tap up to two target creatures.
func TestStkAGoodDayToPieTapsUpToTwo(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0)
	opp := g.Seats[1]
	a := effectstest.P2Push(g, opp.ID, "Bear A", stkCreatureTypeLine, "", 2, 2)
	b := effectstest.P2Push(g, opp.ID, "Bear B", stkCreatureTypeLine, "", 2, 2)
	c := effectstest.P2Push(g, opp.ID, "Bear C", stkCreatureTypeLine, "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "A Good Day to Pie", "Instant", aGoodDayToPieOracle, []game.TargetRef{effectstest.P2Card(a), effectstest.P2Card(b)})
	effectstest.PassPriorityAroundTable(t, g)
	if !stkTapped(g, a) || !stkTapped(g, b) || stkTapped(g, c) {
		t.Errorf("tapped a=%v b=%v c=%v, want a and b only", stkTapped(g, a), stkTapped(g, b), stkTapped(g, c))
	}
	_ = me
}

// Done for the Day: with one of the three types you may get a ticket or a
// Treasure, and with all three a sticker besides.
func TestStkDoneForTheDayGivesATicket(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster")
	effectstest.PushCatalogPermanent(g, me.ID, "Done for the Day", stkEnchantmentTypeLine, doneForTheDayOracle, false)
	effectstest.P2Push(g, me.ID, "Intern", stkEmployeeTypeLine, "", 1, 1)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("the end step trigger asked nothing")
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, effectstest.StkOption(t, pick, "Get {TK}")); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := stkTickets(g, me); got != 1 {
		t.Errorf("tickets = %d, want 1", got)
	}
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("a sticker question is open with only an Employee")
	}
}

func TestStkDoneForTheDayMakesATreasure(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0)
	effectstest.PushCatalogPermanent(g, me.ID, "Done for the Day", stkEnchantmentTypeLine, doneForTheDayOracle, false)
	effectstest.P2Push(g, me.ID, "Bot", stkRobotTypeLine, "", 1, 1)
	before := len(g.Battlefield.Cards)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if err := g.ResolveOptionPick(pick.ID, me.ID, effectstest.StkOption(t, pick, "Treasure")); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(g.Battlefield.Cards); got != before+1 {
		t.Errorf("battlefield = %d permanents, want %d", got, before+1)
	}
	if got := stkTickets(g, me); got != 0 {
		t.Errorf("tickets = %d, want 0", got)
	}
}

func TestStkDoneForTheDayWithAllThreePutsAStickerOn(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster")
	effectstest.PushCatalogPermanent(g, me.ID, "Done for the Day", stkEnchantmentTypeLine, doneForTheDayOracle, false)
	emp := effectstest.P2Push(g, me.ID, "Intern", stkEmployeeTypeLine, "", 1, 1)
	effectstest.P2Push(g, me.ID, "Dancer", stkPerformerTypeLine, "", 1, 1)
	effectstest.P2Push(g, me.ID, "Bot", stkRobotTypeLine, "", 1, 1)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if err := g.ResolveOptionPick(pick.ID, me.ID, effectstest.StkOption(t, pick, "Get {TK}")); err != nil {
		t.Fatalf("pick: %v", err)
	}
	// The sticker question waits for the first answer, and the ticket is in
	// hand when it is asked.
	effectstest.StkPutOn(t, g, me, emp, "“Night”", 0)
	if got := stkNames(effectstest.StkCard(t, g, emp)); got != "Night Intern" {
		t.Errorf("name = %q, want %q", got, "Night Intern")
	}
}

// Neither is a legal answer, and without any of the three it never triggers.
func TestStkDoneForTheDayNeedsOneOfTheThree(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0)
	effectstest.PushCatalogPermanent(g, me.ID, "Done for the Day", stkEnchantmentTypeLine, doneForTheDayOracle, false)
	effectstest.P2Push(g, me.ID, "Bear", stkCreatureTypeLine, "", 1, 1)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Error("it triggered without an Employee, a Performer or a Robot")
	}
}

// Wicker Picker: sticker kicker {1} on a creature spell gives a ticket and
// a sticker question, and the sticker stays on the permanent.
func TestStkWickerPickerStickerKickerPutsAStickerOnTheSpell(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 1, "Night Brushwagg Ringmaster")
	effectstest.PushCatalogPermanent(g, me.ID, "Wicker Picker", "Artifact Creature — Scarecrow Guest", wickerPickerOracle, false)
	effectstest.AdvanceToMain(t, g)
	bear := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Grizzly Bears", TypeLine: stkCreatureTypeLine,
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast with sticker kicker: %v", err)
	}
	if got := stkTickets(g, me); got != 2 {
		t.Fatalf("tickets after kicking = %d, want 2 (1 + the one sticker kicker gives)", got)
	}
	effectstest.StkPutOn(t, g, me, bear, "Menace", 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, bear) {
		t.Fatal("the creature did not enter")
	}
	c := effectstest.StkCard(t, g, bear)
	if len(c.Stickers) != 1 || !game.HasKeyword(&c, game.CounterMenace) {
		t.Errorf("the creature kept no menace sticker: %+v", c.Stickers)
	}
	if got := stkTickets(g, me); got != 0 {
		t.Errorf("tickets after the sticker = %d, want 0", got)
	}
}

// Sticker kicker is optional, and an unkicked cast asks nothing.
func TestStkWickerPickerUnkickedAsksNothing(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 1, "Night Brushwagg Ringmaster")
	effectstest.PushCatalogPermanent(g, me.ID, "Wicker Picker", "Artifact Creature — Scarecrow Guest", wickerPickerOracle, false)
	effectstest.AdvanceToMain(t, g)
	bear := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Grizzly Bears", TypeLine: stkCreatureTypeLine,
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("an unkicked cast asked a sticker question")
	}
	if got := stkTickets(g, me); got != 1 {
		t.Errorf("tickets = %d, want 1", got)
	}
}

// The grant reaches creature spells only, and only while Wicker Picker is
// on the battlefield under the caster.
func TestStkWickerPickerOnlyForCreatureSpells(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 1, "Night Brushwagg Ringmaster")
	effectstest.AdvanceToMain(t, g)
	bear := uuid.New()
	rock := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Grizzly Bears", TypeLine: stkCreatureTypeLine, Owner: me.ID, Controller: me.ID})
		me.Hand.PushTop(game.Card{InstanceID: rock, Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{OptionalCosts: []int{0}}); err == nil {
		t.Fatal("sticker kicker was accepted with no Wicker Picker")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Wicker Picker", "Artifact Creature — Scarecrow Guest", wickerPickerOracle, false)
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{OptionalCosts: []int{0}}); err == nil {
		t.Fatal("sticker kicker was accepted on an artifact spell")
	}
}

// A kicker the creature prints is separate: both can be paid, and the
// printed one is what "was kicked" counts (CR 702.33e).
func TestStkStickerKickerIsNotTheCardsOwnKicker(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster")
	effectstest.PushCatalogPermanent(g, me.ID, "Wicker Picker", "Artifact Creature — Scarecrow Guest", wickerPickerOracle, false)
	effectstest.AdvanceToMain(t, g)
	var offered []game.AdditionalCost
	g.WithWriteLock(func() {
		offered = g.OptionalCostsForCastLocked(me.ID, game.Card{
			Name: "Kicked Bear", TypeLine: stkCreatureTypeLine, Owner: me.ID, Controller: me.ID,
		})
	})
	if len(offered) != 1 || offered[0].Key != game.StickerKickerKey || offered[0].ManaCost != "{1}" || !offered[0].Granted {
		t.Errorf("offered = %+v, want one granted sticker kicker {1}", offered)
	}
	if game.KickedTimesPaid(game.Card{}, nil) != 0 {
		t.Error("kicked count of nothing")
	}
}

// Pin Collection: an ability sticker with ticket cost X or less goes on
// for free, and the equipped creature has its ability.
func TestStkPinCollectionPutsAFreeAbilityStickerOnItAndLendsIt(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster")
	effectstest.AdvanceToMain(t, g)
	pin := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: pin, Name: "Pin Collection", TypeLine: "Artifact — Equipment",
			OracleID: pinCollectionOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, pin, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no sticker question for the Equipment")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{pin}); err != nil {
		t.Fatalf("choose the Equipment: %v", err)
	}
	which := effectstest.LatestOptionPickFor(g, me.ID)
	for _, o := range which.PickOptions {
		switch {
		case o.Label == game.DeclineStickerLabel:
		case strings.Contains(o.Label, "Persist"):
			t.Errorf("offered %q, which costs 3 tickets (X is 2)", o.Label)
		case !strings.HasPrefix(o.Label, "Ability sticker"):
			t.Errorf("offered %q, which is not an ability sticker", o.Label)
		}
	}
	if err := g.ResolveOptionPick(which.ID, me.ID, effectstest.StkOption(t, which, "Menace")); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if got := stkTickets(g, me); got != 0 {
		t.Errorf("tickets = %d, want 0 (none paid, none had)", got)
	}
	bear := effectstest.P2Push(g, me.ID, "Bear", stkCreatureTypeLine, "", 2, 2)
	effectstest.FloatForTest(g, me, "WW")
	if err := g.ActivateCatalogAbility(me.ID, pin, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.P2Card(bear)}}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P2Keyword(g, bear, game.CounterMenace) {
		t.Error("the equipped creature has no menace from the Equipment's sticker")
	}
	if p, tough := effectstest.P2PT(g, bear); p != 3 || tough != 3 {
		t.Errorf("equipped Bear = %d/%d, want 3/3", p, tough)
	}
}

// The cost cap is X: with X = 3 persist is offered too.
func TestStkPinCollectionCapIsX(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster")
	effectstest.AdvanceToMain(t, g)
	pin := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: pin, Name: "Pin Collection", TypeLine: "Artifact — Equipment",
			OracleID: pinCollectionOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, pin, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{pin}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	which := effectstest.LatestOptionPickFor(g, me.ID)
	found := false
	for _, o := range which.PickOptions {
		if strings.Contains(o.Label, "Persist") {
			found = true
		}
		if strings.Contains(o.Label, "10/10") {
			t.Errorf("offered %q beyond X = 3", o.Label)
		}
	}
	if !found {
		t.Error("persist (3 tickets) is not offered with X = 3")
	}
}

// The free limit survives a restore taken at the sticker question.
func TestStkPinCollectionFreeStickerSurvivesARestore(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster")
	effectstest.AdvanceToMain(t, g)
	pin := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: pin, Name: "Pin Collection", TypeLine: "Artifact — Equipment",
			OracleID: pinCollectionOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, pin, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the restored table lost the object question")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{pin}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	which := effectstest.LatestOptionPickFor(g, me.ID)
	if which == nil {
		t.Fatal("the restored table lost the sticker question")
	}
	if err := g.ResolveOptionPick(which.ID, me.ID, effectstest.StkOption(t, which, "Menace")); err != nil {
		t.Fatalf("pick (free, no tickets): %v", err)
	}
	if c := effectstest.StkCard(t, g, pin); len(c.Stickers) != 1 {
		t.Errorf("stickers = %+v, want the free menace sticker", c.Stickers)
	}
}

// Clandestine Chameleon: all abilities of ability stickers on other
// permanents it owns and cards in its owner's graveyard.
func TestStkClandestineChameleonHasTheAbilitiesOfOtherStickers(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 12, "Night Brushwagg Ringmaster", "Phyrexian Midway Bamboozle")
	cham := effectstest.P2Push(g, me.ID, "Clandestine Chameleon", "Creature — Lizard Guest", clandestineChamOracle, 4, 3)
	bear := effectstest.P2Push(g, me.ID, "Bear", stkCreatureTypeLine, "", 2, 2)
	dead := stkGraveyardCard(me, "Dead Bear", stkCreatureTypeLine, "")
	opp := g.Seats[1]
	theirs := effectstest.P2Push(g, opp.ID, "Their Bear", stkCreatureTypeLine, "", 2, 2)
	_ = theirs
	if effectstest.P2Keyword(g, cham, game.CounterMenace) {
		t.Fatal("setup: the Chameleon already has menace")
	}
	effectstest.StkGive(t, g, me, bear, "Menace")
	if !effectstest.P2Keyword(g, cham, game.CounterMenace) {
		t.Error("no menace from the sticker on another permanent")
	}
	effectstest.StkGive(t, g, me, dead, "Persist")
	if !effectstest.P2Keyword(g, cham, game.KeywordPersist) {
		t.Error("no persist from the sticker on a card in the graveyard")
	}
	// A name sticker on a lender gives nothing.
	effectstest.StkGive(t, g, me, bear, "“Night”")
	// A bundle: the attack trigger comes with it (Bamboozle's ticket).
	effectstest.StkGive(t, g, me, bear, "Whenever this creature attacks")
	c := effectstest.StkCard(t, g, cham)
	if len(c.Effective().GrantedAbilities) != 1 {
		t.Errorf("granted bundles = %+v, want the attack trigger's", c.Effective().GrantedAbilities)
	}
	// The bear's own keywords are its own; the Chameleon lends nothing back.
	// A sticker on the Chameleon itself is its own sticker's (not "other").
	effectstest.StkGive(t, g, me, cham, "Undying")
	if !effectstest.P2Keyword(g, cham, game.KeywordUndying) {
		t.Error("the Chameleon's own sticker did not give undying")
	}
}

// The borrowed attack trigger fires for the Chameleon with it as "this
// creature".
func TestStkClandestineChameleonBorrowedTriggerFires(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 3, "Phyrexian Midway Bamboozle")
	cham := effectstest.P2Push(g, me.ID, "Clandestine Chameleon", "Creature — Lizard Guest", clandestineChamOracle, 4, 3)
	bear := effectstest.P2Push(g, me.ID, "Bear", stkCreatureTypeLine, "", 2, 2)
	effectstest.StkGive(t, g, me, bear, "Whenever this creature attacks")
	before := stkTickets(g, me)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(cham, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := stkTickets(g, me); got != before+1 {
		t.Errorf("tickets = %d, want %d: the borrowed attack trigger did not fire", got, before+1)
	}
}

// Chameleon's enters trigger: two tickets, then a sticker.
func TestStkClandestineChameleonEntersWithTicketsAndASticker(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster")
	cham := effectstest.CastAndResolveCreature(t, g, "Clandestine Chameleon", "Creature — Lizard Guest", clandestineChamOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.StkPutOn(t, g, me, cham, "Menace", 0)
	if got := stkTickets(g, me); got != 0 {
		t.Errorf("tickets = %d, want 0 (two given, two paid)", got)
	}
	if !effectstest.P2Keyword(g, cham, game.CounterMenace) {
		t.Error("the Chameleon's own menace sticker did not give menace")
	}
}
