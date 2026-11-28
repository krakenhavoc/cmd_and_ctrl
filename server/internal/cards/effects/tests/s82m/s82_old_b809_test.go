package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b809_test.go — Mirage, batch 809 (#5633): life gain and loss.

const (
	oldB809Roots      = "29ec9b49-fbe5-49ac-8ad6-7159cb88bc66"
	oldB809Purgatory  = "2fb8e100-58a1-4dc6-bc38-ad7f115e3e46"
	oldB809Djinn      = "3e1a7146-75d5-499f-8718-303bc6ea139b"
	oldB809Binding    = "52fc8217-5a14-40b5-a8fc-b6fc45af5f33"
	oldB809Elixir     = "635cc10e-ca25-49a2-af44-3b064263a254"
	oldB809Ancestor   = "82fb9834-de65-41a5-b908-f49668dee3ec"
	oldB809Sirocco    = "a9bdf857-0c6b-4afa-afbb-aff7e4b61b09"
	oldB809Cascade    = "aaf176ff-646f-4182-a546-fe648498587d"
	oldB809Rats       = "bf526a0d-65cc-445f-b2b8-19a2cfdd836b"
	oldB809Auction    = "dea8c725-455e-4668-877a-df388fe22255"
	oldB809Purge      = "df032801-fccf-4bff-aa88-d01650b6c81e"
	oldB809Champion   = "f203671b-76f7-4be0-80a9-826fbb450fe7"
	oldB809BearType   = "Creature — Bear"
	oldB809InstantTyp = "Instant"
)

func TestOldB809EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB809Roots: "Roots of Life", oldB809Purgatory: "Purgatory", oldB809Djinn: "Emberwilde Djinn",
		oldB809Binding: "Spatial Binding", oldB809Elixir: "Elixir of Vitality", oldB809Ancestor: "Auspicious Ancestor",
		oldB809Sirocco: "Sirocco", oldB809Cascade: "Vitalizing Cascade", oldB809Rats: "Sewer Rats",
		oldB809Auction: "Illicit Auction", oldB809Purge: "Phyrexian Purge", oldB809Champion: "Ethereal Champion",
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

func oldB809Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB809Activate(t *testing.T, g *game.Game, who, src uuid.UUID, idx int, p game.ActivateAbilityParams) error {
	t.Helper()
	p.Strict = true
	return g.ActivateCatalogAbility(who, src, idx, p)
}

// oldB809Prompt returns the open pay-unless prompt for the player, or nil.
func oldB809Prompt(g *game.Game, who uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		if c := g.PendingChoices[i]; c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == who {
			return c
		}
	}
	return nil
}

func TestOldB809VitalizingCascadeGainsXPlusThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB809Seats(g)
	effectstest.FloatMana(t, g, me, "{C}{C}{G}{W}")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Vitalizing Cascade", TypeLine: oldB809InstantTyp, ManaCost: "{X}{G}{W}", OracleID: oldB809Cascade, Owner: me.ID, Controller: me.ID})
	before := me.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+5 {
		t.Errorf("life %d→%d, want +5 (X=2 plus 3)", before, me.Life)
	}
}

func TestOldB809ElixirEntersTappedAndSacrificesForFourOrEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB809Seats(g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Elixir of Vitality", TypeLine: "Artifact", ManaCost: "{4}", OracleID: oldB809Elixir, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, id) {
		t.Fatal("the Elixir did not enter tapped")
	}
	if err := oldB809Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a tapped Elixir was sacrificed for life")
	}
	effectstest.Untap(g, id)
	before := me.Life
	// The {8} ability needs the mana.
	if err := oldB809Activate(t, g, me.ID, id, 1, game.ActivateAbilityParams{}); err == nil {
		t.Error("the {8} ability was activated with no mana")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}{C}")
	if err := oldB809Activate(t, g, me.ID, id, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("{8}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+8 || g.Battlefield.Contains(id) {
		t.Errorf("life %d→%d, on battlefield %v: want +8 and sacrificed", before, me.Life, g.Battlefield.Contains(id))
	}

	id2 := effectstest.PushCatalogPermanent(g, me.ID, "Elixir of Vitality", "Artifact", oldB809Elixir, false)
	before = me.Life
	if err := oldB809Activate(t, g, me.ID, id2, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("{T}, sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+4 || g.Battlefield.Contains(id2) {
		t.Errorf("life %d→%d: want +4 and sacrificed", before, me.Life)
	}
}

func TestOldB809SewerRatsPumpsThreeTimesEachTurnForLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB809Seats(g)
	rats := effectstest.B12Push(g, me.ID, "Sewer Rats", "Creature — Rat", oldB809Rats, 1, 1)
	before := me.Life
	for i := 0; i < 3; i++ {
		effectstest.FloatMana(t, g, me, "{B}")
		if err := oldB809Activate(t, g, me.ID, rats, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
	}
	effectstest.FloatMana(t, g, me, "{B}")
	if err := oldB809Activate(t, g, me.ID, rats, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a fourth activation in one turn was allowed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, rats); p != 4 {
		t.Errorf("power %d, want 4 (+3)", p)
	}
	if me.Life != before-3 {
		t.Errorf("life %d→%d, want -3", before, me.Life)
	}
}

func TestOldB809EtherealChampionPaysLifeToPreventOneDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB809Seats(g)
	champ := effectstest.B12Push(g, me.ID, "Ethereal Champion", "Creature — Avatar", oldB809Champion, 3, 4)
	before := me.Life
	if err := oldB809Activate(t, g, me.ID, champ, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-1 {
		t.Errorf("life %d→%d, want -1", before, me.Life)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), champ, 3) })
	if got := effectstest.DamageMarkedOn(g, champ); got != 2 {
		t.Errorf("3 damage into a 1-shield: marked %d, want 2", got)
	}
	me.Life = 0
	if err := oldB809Activate(t, g, me.ID, champ, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("paid 1 life at 0 life")
	}
}

func oldB809Seat(g *game.Game, i int) *game.Player {
	return g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)]
}

// oldB809LetStand answers every open "top the bid" prompt with "let it stand"
// (and tops once as `topper`, to `to`, when set).
func oldB809Bidding(t *testing.T, g *game.Game, topper *game.Player, to int) {
	t.Helper()
	topped := false
	for i := 0; i < 20; i++ {
		var c *game.PendingChoice
		for j := len(g.PendingChoices) - 1; j >= 0; j-- {
			if p := g.PendingChoices[j]; p != nil && (p.Kind == game.PendingChoiceOptionPick || p.Kind == game.PendingChoicePayAmount) {
				c = p
				break
			}
		}
		if c == nil {
			return
		}
		switch {
		case c.Kind == game.PendingChoicePayAmount:
			if err := g.ResolvePayAmount(c.ID, c.Chooser, to); err != nil {
				t.Fatalf("bid: %v", err)
			}
		case topper != nil && c.Chooser == topper.ID && !topped:
			topped = true
			if err := g.ResolveOptionPick(c.ID, c.Chooser, 1); err != nil {
				t.Fatalf("top: %v", err)
			}
		default:
			if err := g.ResolveOptionPick(c.ID, c.Chooser, 0); err != nil {
				t.Fatalf("stand: %v", err)
			}
		}
	}
	t.Fatal("the bidding did not end")
}

func oldB809CastAuction(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Illicit Auction", TypeLine: "Sorcery", OracleID: oldB809Auction, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardTarget(target), Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB809IllicitAuctionHighBidderLosesLifeAndKeepsTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := oldB809Seat(g, 0)
	bidder := oldB809Seat(g, 2)
	owner := oldB809Seat(g, 1)
	bear := effectstest.B12Creature(g, owner.ID, "Bear", oldB809BearType, 2, 2)
	oldB809CastAuction(t, g, me, bear)
	myLife, theirLife := me.Life, bidder.Life
	oldB809Bidding(t, g, bidder, 3)
	if bidder.Life != theirLife-3 || me.Life != myLife {
		t.Errorf("life me %d→%d, bidder %d→%d, want only the bidder to lose 3", myLife, me.Life, theirLife, bidder.Life)
	}
	if got := effectstest.ControllerOf(t, g, bear); got != bidder.ID {
		t.Errorf("the creature is controlled by %s, want the high bidder", got)
	}
	// It lasts: through the cleanup step of this turn.
	effectstest.AdvanceToEndStep(t, g)
	if got := effectstest.ControllerOf(t, g, bear); got != bidder.ID {
		t.Errorf("control reverted at end of turn")
	}
}

func TestOldB809IllicitAuctionWithNoOneTopping(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := oldB809Seat(g, 0)
	owner := oldB809Seat(g, 1)
	bear := effectstest.B12Creature(g, owner.ID, "Bear", oldB809BearType, 2, 2)
	oldB809CastAuction(t, g, me, bear)
	before := me.Life
	oldB809Bidding(t, g, nil, 0)
	if me.Life != before {
		t.Errorf("life %d→%d, want unchanged at a bid of 0", before, me.Life)
	}
	if got := effectstest.ControllerOf(t, g, bear); got != me.ID {
		t.Errorf("the caster did not win the creature at 0")
	}
}

func TestOldB809RootsOfLifeGainsForTheChosenTypeOfAnOpponentsLandOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB809Seats(g)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Roots of Life", TypeLine: "Enchantment", OracleID: oldB809Roots,
		Owner: me.ID, Controller: me.ID, ChosenOption: "Island",
	})
	island := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	swamp := effectstest.B12Permanent(g, opp.ID, "Swamp", "Basic Land — Swamp")
	mine := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	before := me.Life
	for _, id := range []uuid.UUID{swamp, mine} {
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
		effectstest.PassPriorityAroundTable(t, g)
	}
	if me.Life != before {
		t.Fatalf("life %d→%d: a Swamp or my own Island gained life", before, me.Life)
	}
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(island) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d→%d, want +1 for an opponent's tapped Island", before, me.Life)
	}
}

func TestOldB809RootsOfLifeOffersIslandOrSwampAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB809Seats(g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Roots of Life", TypeLine: "Enchantment", OracleID: oldB809Roots, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestOptionPickFor(g, me.ID)
	if c == nil {
		t.Fatal("no Island-or-Swamp choice was offered")
	}
	if len(c.PickOptions) != 2 || c.PickOptions[0].Label != "Island" || c.PickOptions[1].Label != "Swamp" {
		t.Fatalf("options %+v, want Island and Swamp", c.PickOptions)
	}
	if err := g.ResolveOptionPick(c.ID, me.ID, 1); err != nil {
		t.Fatalf("choose Swamp: %v", err)
	}
	if got := g.ChosenOptionOf(id); got != "Swamp" {
		t.Errorf("chosen option %q, want Swamp", got)
	}
}

func TestOldB809EmberwildeDjinnChangesHandsForManaOrLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB809Seats(g)
	djinn := effectstest.B12Push(g, me.ID, "Emberwilde Djinn", "Creature — Djinn", oldB809Djinn, 5, 4)
	effectstest.AdvanceToUpkeepOfSeat(t, g, oldB809SeatIndex(g, opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	prompt := oldB809Prompt(g, opp.ID)
	if prompt == nil {
		t.Fatal("the upkeep player was not offered the payment")
	}
	before := opp.Life
	option := 1 // pay 2 life
	if err := g.ResolvePayUnlessWithOption(prompt.ID, opp.ID, true, &option, nil); err != nil {
		t.Fatalf("pay 2 life: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("life %d→%d, want -2", before, opp.Life)
	}
	if got := effectstest.ControllerOf(t, g, djinn); got != opp.ID {
		t.Errorf("Djinn controlled by %s, want the payer", got)
	}
}

func TestOldB809EmberwildeDjinnStaysPutWhenTheUpkeepPlayerDeclines(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB809Seats(g)
	djinn := effectstest.B12Push(g, me.ID, "Emberwilde Djinn", "Creature — Djinn", oldB809Djinn, 5, 4)
	effectstest.AdvanceToUpkeepOfSeat(t, g, oldB809SeatIndex(g, opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	prompt := oldB809Prompt(g, opp.ID)
	if prompt == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolvePayUnless(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, djinn); got != me.ID {
		t.Error("the Djinn changed hands without a payment")
	}
}

func TestOldB809PurgatoryExilesYourNontokenCreaturesAndReturnsThemForFourAndTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB809Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Purgatory", "Enchantment", oldB809Purgatory, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB809BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB809BearType, 2, 2)
	for _, id := range []uuid.UUID{mine, theirs} {
		g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), id, 5) })
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
	}
	if z := effectstest.B12ZoneOf(g, mine); z != game.ZoneExile {
		t.Errorf("my dead creature is in zone %v, want exile", z)
	}
	if z := effectstest.B12ZoneOf(g, theirs); z != game.ZoneGraveyard {
		t.Errorf("their dead creature is in zone %v, want its graveyard", z)
	}
	// Next turn of mine: the upkeep offer.
	effectstest.AdvanceToUpkeepOfSeat(t, g, oldB809SeatIndex(g, me.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if oldB809Prompt(g, me.ID) == nil {
		t.Fatal("no upkeep offer")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	before := me.Life
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-2 {
		t.Errorf("life %d→%d, want -2", before, me.Life)
	}
	back := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "My Bear" && c.Controller == me.ID {
			back = true
		}
	}
	if !back {
		t.Error("the exiled creature did not return to the battlefield")
	}
}

func TestOldB809SpatialBindingStopsAPermanentPhasingOut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB809Seats(g)
	binding := effectstest.PushCatalogPermanent(g, me.ID, "Spatial Binding", "Enchantment", oldB809Binding, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB809BearType, 2, 2)
	before := me.Life
	if err := oldB809Activate(t, g, me.ID, binding, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-1 {
		t.Errorf("life %d→%d, want -1", before, me.Life)
	}
	if !g.CantPhaseOutForEffect(bear) {
		t.Error("the target can still phase out")
	}
	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, bear) })
	if g.IsPhasedOutForEffect(bear) {
		t.Error("the protected creature phased out")
	}
}

func TestOldB809AuspiciousAncestorGainsOnDeathAndOnWhiteSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB809Seats(g)
	anc := effectstest.B12Push(g, me.ID, "Auspicious Ancestor", "Creature — Human Cleric", oldB809Ancestor, 2, 3)
	// An opponent's white spell offers me the payment.
	white := game.Card{InstanceID: uuid.New(), Name: "White Spell", TypeLine: oldB809InstantTyp, Colors: []string{"W"}, Owner: opp.ID, Controller: opp.ID}
	opp.Hand.PushTop(white)
	before := me.Life
	if err := g.CastSpell(opp.ID, white.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB809Prompt(g, me.ID) == nil {
		t.Fatal("a white spell did not offer the {1}")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d→%d, want +1", before, me.Life)
	}
	// A red spell offers nothing.
	red := game.Card{InstanceID: uuid.New(), Name: "Red Spell", TypeLine: oldB809InstantTyp, Colors: []string{"R"}, Owner: opp.ID, Controller: opp.ID}
	opp.Hand.PushTop(red)
	if err := g.CastSpell(opp.ID, red.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB809Prompt(g, me.ID) != nil {
		t.Error("a red spell offered the payment")
	}
	// Death.
	before = me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), anc, 5) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("life %d→%d, want +3 on death", before, me.Life)
	}
}

func TestOldB809SiroccoRevealsAndMakesBlueInstantsPayOrDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB809Seats(g)
	blue := game.Card{InstanceID: uuid.New(), Name: "Blue Instant", TypeLine: oldB809InstantTyp, Colors: []string{"U"}, Owner: opp.ID, Controller: opp.ID}
	blue2 := game.Card{InstanceID: uuid.New(), Name: "Other Blue Instant", TypeLine: oldB809InstantTyp, Colors: []string{"U"}, Owner: opp.ID, Controller: opp.ID}
	redInstant := game.Card{InstanceID: uuid.New(), Name: "Red Instant", TypeLine: oldB809InstantTyp, Colors: []string{"R"}, Owner: opp.ID, Controller: opp.ID}
	blueSorcery := game.Card{InstanceID: uuid.New(), Name: "Blue Sorcery", TypeLine: "Sorcery", Colors: []string{"U"}, Owner: opp.ID, Controller: opp.ID}
	for _, c := range []game.Card{blue, blue2, redInstant, blueSorcery} {
		opp.Hand.PushTop(c)
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Sirocco", TypeLine: oldB809InstantTyp, OracleID: oldB809Sirocco, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	before := opp.Life
	// First blue instant: pay 4 life; second: decline.
	for i := 0; i < 2; i++ {
		prompt := oldB809Prompt(g, opp.ID)
		if prompt == nil {
			t.Fatalf("question %d not asked", i+1)
		}
		effectstest.AnswerPayUnless(t, g, opp.ID, i == 0)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if opp.Life != before-4 {
		t.Errorf("life %d→%d, want -4 (one payment)", before, opp.Life)
	}
	inHand := 0
	for _, c := range opp.Hand.Cards {
		if c.Name == "Blue Instant" || c.Name == "Other Blue Instant" {
			inHand++
		}
	}
	if inHand != 1 {
		t.Errorf("%d blue instants left in hand, want 1 (the paid one)", inHand)
	}
	if !opp.Hand.Contains(redInstant.InstanceID) || !opp.Hand.Contains(blueSorcery.InstanceID) {
		t.Error("a card that is not a blue instant was discarded")
	}
}

func TestOldB809PhyrexianPurgeCostsThreeLifePerTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB809Seats(g)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", oldB809BearType, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", oldB809BearType, 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Phyrexian Purge", TypeLine: "Sorcery", ManaCost: "{2}{B}{R}", OracleID: oldB809Purge, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{C}{C}{B}{R}")
	me.Life = 5
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(a, b), Strict: true}); err == nil {
		t.Fatal("cast two targets at 5 life (6 life needed)")
	}
	me.Life = 7
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(a, b), Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 1 {
		t.Errorf("life %d, want 1 (7 - 3 per target)", me.Life)
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a targeted creature survived")
	}
}

func oldB809SeatIndex(g *game.Game, id uuid.UUID) int {
	for i, s := range g.Seats {
		if s.ID == id {
			return i
		}
	}
	return -1
}
